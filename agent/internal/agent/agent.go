package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jumpfrp/agent/internal/frps"
	"github.com/jumpfrp/agent/internal/monitor"
	agentapi "github.com/jumpfrp/agent/internal/api"
)

type Config struct {
	NodeID    string
	Token     string
	MasterURL string
	FrpsPort  int
	AgentPort int
}

type Agent struct {
	cfg        *Config
	frpsMgr    *frps.Manager
	monitor    *monitor.Monitor
	httpServer *http.Server
	stopCh     chan struct{}
}

func New(cfg *Config) *Agent {
	return &Agent{
		cfg:     cfg,
		stopCh:  make(chan struct{}),
		monitor: monitor.New(),
	}
}

func (a *Agent) Start() error {
	log.Printf("JumpFrp Agent 启动 | 节点: %s | 主控: %s", a.cfg.NodeID, a.cfg.MasterURL)

	// 初始化 frps 管理器
	a.frpsMgr = frps.NewManager(a.cfg.FrpsPort)

	// 启动 frps
	if err := a.frpsMgr.Start(); err != nil {
		log.Printf("[警告] frps 启动失败: %v (可能未安装)", err)
	}

	// 启动 HTTP API 服务
	router := agentapi.NewRouter(a.cfg.Token, a.frpsMgr, a.monitor)
	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.cfg.AgentPort),
		Handler: router,
	}
	go func() {
		log.Printf("Agent API 监听 :%d", a.cfg.AgentPort)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Agent API 错误: %v", err)
		}
	}()

	// 向主控注册
	go a.register()

	// 启动心跳
	go a.heartbeatLoop()

	return nil
}

func (a *Agent) Stop() {
	close(a.stopCh)
	if a.httpServer != nil {
		a.httpServer.Close()
	}
	if a.frpsMgr != nil {
		a.frpsMgr.Stop()
	}
}

// 向主控注册节点
func (a *Agent) register() {
	for i := 0; i < 5; i++ {
		err := a.callMaster("POST", "/api/agent/register", map[string]interface{}{
			"node_id": a.cfg.NodeID,
			"token":   a.cfg.Token,
		})
		if err == nil {
			log.Println("已向主控注册成功")
			return
		}
		log.Printf("注册失败 (第%d次): %v", i+1, err)
		time.Sleep(5 * time.Second)
	}
	log.Println("[警告] 注册失败，将在心跳中重试")
}

// 心跳循环，每 30 秒上报一次
func (a *Agent) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 立即发一次
	a.sendHeartbeat()

	for {
		select {
		case <-ticker.C:
			a.sendHeartbeat()
		case <-a.stopCh:
			return
		}
	}
}

func (a *Agent) sendHeartbeat() {
	stats := a.monitor.Collect()
	conns := 0
	if a.frpsMgr != nil {
		conns = a.frpsMgr.ConnectionCount()
	}

	payload := map[string]interface{}{
		"node_id":      a.cfg.NodeID,
		"token":        a.cfg.Token,
		"cpu_usage":    stats.CPUUsage,
		"memory_usage": stats.MemoryUsage,
		"current_conns": conns,
		"version":      "1.0.0",
	}

	if err := a.callMaster("POST", "/api/agent/heartbeat", payload); err != nil {
		log.Printf("心跳上报失败: %v", err)
	}
}

func (a *Agent) callMaster(method, path string, payload interface{}) error {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(method, a.cfg.MasterURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Token", a.cfg.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("主控返回 %d", resp.StatusCode)
	}
	return nil
}

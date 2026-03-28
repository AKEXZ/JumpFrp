package agent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/jumpfrp/agent/internal/frps"
	"github.com/jumpfrp/agent/internal/monitor"
	"github.com/jumpfrp/agent/internal/tc"
	agentapi "github.com/jumpfrp/agent/internal/api"
)

const (
	frpsConfigPath = "/opt/jumpfrp/frps.toml"
)

type Config struct {
	NodeID    string
	Token     string
	MasterURL string
	FrpsPort  int
	AgentPort int
}

type Agent struct {
	cfg           *Config
	frpsMgr       *frps.Manager
	monitor       *monitor.Monitor
	httpServer    *http.Server
	stopCh        chan struct{}
	configVersion int
	tcCtrl        *tc.TrafficControl
}

func New(cfg *Config) *Agent {
	return &Agent{
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
}

func (a *Agent) Start() error {
	log.Printf("JumpFrp Agent 启动 | 节点: %s | 主控: %s", a.cfg.NodeID, a.cfg.MasterURL)

	// 初始化流量控制
	if err := a.initTrafficControl(); err != nil {
		log.Printf("[警告] 流量控制初始化失败: %v", err)
	}

	// 初始化 frps 管理器
	a.frpsMgr = frps.NewManager(a.cfg.FrpsPort)

	// 尝试从本地配置文件启动
	if err := a.frpsMgr.Start(); err != nil {
		log.Printf("[警告] frps 启动失败: %v", err)
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

	// 启动 frps 日志监控
	go a.monitorFrpsLogs()

	return nil
}

// initTrafficControl 初始化流量控制
func (a *Agent) initTrafficControl() error {
	iface := os.Getenv("NETWORK_IFACE")
	if iface == "" {
		iface = detectNetworkInterface()
	}
	if iface == "" {
		iface = "eth0"
	}

	a.tcCtrl = tc.New(iface)
	return a.tcCtrl.Init()
}

func detectNetworkInterface() string {
	ifaces := []string{"eth0", "ens33", "enp0s3", "ens5", "bond0"}
	for _, iface := range ifaces {
		if _, err := os.Stat(fmt.Sprintf("/sys/class/net/%s", iface)); err == nil {
			return iface
		}
	}
	output, _ := exec.Command("ip", "route", "show", "default").Output()
	parts := strings.Fields(string(output))
	for i, part := range parts {
		if part == "dev" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "eth0"
}

func (a *Agent) Stop() {
	close(a.stopCh)
	if a.httpServer != nil {
		a.httpServer.Close()
	}
	if a.frpsMgr != nil {
		a.frpsMgr.Stop()
	}
	if a.tcCtrl != nil {
		a.tcCtrl.Cleanup()
	}
}

// monitorFrpsLogs 监控 frps 日志，检测连接变化
func (a *Agent) monitorFrpsLogs() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.parseFrpsLogs()
		case <-a.stopCh:
			return
		}
	}
}

// parseFrpsLogs 解析 frps 日志，提取连接信息
func (a *Agent) parseFrpsLogs() {
	if a.tcCtrl == nil {
		return
	}

	logPath := "/var/log/frps.log"
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		logPath = "/opt/jumpfrp/frps.log"
	}

	file, err := os.Open(logPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	loginRegex := regexp.MustCompile(`login from (\S+):\d+, auth token: (\S+)`)

	seen := make(map[string]bool)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := loginRegex.FindStringSubmatch(line); len(matches) == 3 {
			ip := matches[1]
			token := matches[2]
			seen[token] = true

			// 获取用户 VIP 等级并添加连接
			go func(ip, token string) {
				vipLevel := a.getUserVIPLevel(token)
				a.tcCtrl.AddConnection(token, ip, vipLevel)
			}(ip, token)
		}
	}
}

// getUserVIPLevel 从主控获取用户的 VIP 等级
func (a *Agent) getUserVIPLevel(token string) int {
	payload := map[string]interface{}{
		"token": token,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", a.cfg.MasterURL+"/api/agent/get-user-vip", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	var result struct {
		Code     int `json:"code"`
		VIPLevel int `json:"vip_level"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.Code != 0 {
		return 0
	}
	return result.VIPLevel
}

// 向主控注册节点
func (a *Agent) register() {
	for i := 0; i < 5; i++ {
		resp, err := a.callMasterWithResponse("POST", "/api/agent/register", map[string]interface{}{
			"node_id": a.cfg.NodeID,
			"token":   a.cfg.Token,
		})
		if err == nil && resp != nil {
			if resp.FrpsConfig != "" {
				a.saveAndApplyFrpsConfig(resp.FrpsConfig)
			}
			log.Println("已向主控注册成功")
			return
		}
		log.Printf("注册失败 (第%d次): %v", i+1, err)
		time.Sleep(5 * time.Second)
	}
	log.Println("[警告] 注册失败，将在心跳中重试")
}

// 心跳循环
func (a *Agent) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
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
		"node_id":       a.cfg.NodeID,
		"token":         a.cfg.Token,
		"cpu_usage":     stats.CPUUsage,
		"memory_usage":  stats.MemoryUsage,
		"current_conns": conns,
		"version":       "1.0.0",
	}

	resp, err := a.callMasterWithResponse("POST", "/api/agent/heartbeat", payload)
	if err != nil {
		log.Printf("心跳上报失败: %v", err)
		return
	}

	if resp != nil && resp.FrpsConfig != "" {
		a.saveAndApplyFrpsConfig(resp.FrpsConfig)
	}
}

// saveAndApplyFrpsConfig 保存并应用 frps.toml 配置
func (a *Agent) saveAndApplyFrpsConfig(config string) bool {
	if err := os.WriteFile(frpsConfigPath, []byte(config), 0644); err != nil {
		log.Printf("保存 frps.toml 失败: %v", err)
		return false
	}

	log.Println("frps.toml 配置已更新，正在重启 frps...")
	a.frpsMgr.Restart()
	log.Println("frps 重启完成")
	return true
}

type masterResponse struct {
	Code         int    `json:"code"`
	Msg          string `json:"msg"`
	FrpsConfig   string `json:"frps_config"`
	ConfigVersion int   `json:"config_version"`
}

func (a *Agent) callMasterWithResponse(method, path string, payload interface{}) (*masterResponse, error) {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(method, a.cfg.MasterURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Token", a.cfg.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("主控返回 %d", resp.StatusCode)
	}

	var result masterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (a *Agent) callMaster(method, path string, payload interface{}) error {
	_, err := a.callMasterWithResponse(method, path, payload)
	return err
}

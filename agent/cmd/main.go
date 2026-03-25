package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jumpfrp/agent/internal/agent"
)

func main() {
	nodeID    := flag.String("node-id", "", "节点标识 (slug)")
	token     := flag.String("token", "", "Agent Token")
	masterURL := flag.String("master-url", "https://api.jumpfrp.top", "主控地址")
	frpsPort  := flag.Int("frps-port", 7000, "frps 监听端口")
	agentPort := flag.Int("agent-port", 7500, "Agent HTTP 监听端口")
	flag.Parse()

	// 也支持环境变量
	if *nodeID == "" {
		*nodeID = os.Getenv("AGENT_NODE_ID")
	}
	if *token == "" {
		*token = os.Getenv("AGENT_TOKEN")
	}
	if *masterURL == "https://api.jumpfrp.top" && os.Getenv("AGENT_MASTER_URL") != "" {
		*masterURL = os.Getenv("AGENT_MASTER_URL")
	}

	if *nodeID == "" || *token == "" {
		log.Fatal("必须指定 --node-id 和 --token")
	}

	cfg := &agent.Config{
		NodeID:    *nodeID,
		Token:     *token,
		MasterURL: *masterURL,
		FrpsPort:  *frpsPort,
		AgentPort: *agentPort,
	}

	a := agent.New(cfg)
	if err := a.Start(); err != nil {
		log.Fatalf("Agent 启动失败: %v", err)
	}

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Agent 正在关闭...")
	a.Stop()
}

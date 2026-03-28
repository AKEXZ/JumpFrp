package frps

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"text/template"
	"time"
)

const (
	frpsConfigPath = "/opt/jumpfrp/frps.toml"
	frpsBinaryPath = "/opt/jumpfrp/frps"
)

type Manager struct {
	port     int
	cmd      *exec.Cmd
	connCnt  int64
	configPath string
}

func NewManager(port int) *Manager {
	return &Manager{
		port:       port,
		configPath: frpsConfigPath,
	}
}

func (m *Manager) Start() error {
	// 查找 frps 二进制
	frpsBin, err := findFrps()
	if err != nil {
		return fmt.Errorf("frps 未找到: %w", err)
	}

	// 确保配置目录存在
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 如果没有配置文件，使用默认配置
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		m.saveDefaultConfig()
	}

	// 启动 frps
	m.cmd = exec.Command(frpsBin, "-c", m.configPath)
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("启动 frps 失败: %w", err)
	}

	log.Printf("frps 已启动 (pid=%d, config=%s)", m.cmd.Process.Pid, m.configPath)
	return nil
}

// 保存默认配置文件
func (m *Manager) saveDefaultConfig() {
	token := os.Getenv("FRPS_TOKEN")
	if token == "" {
		token = "default-token"
	}

	cfg := fmt.Sprintf(`
# frps.toml - JumpFrp 服务端配置
bindPort = %d
auth.method = "token"
auth.token = "%s"

[transport]
max_pool_count = 100
pool_count = 10

[log]
to = "/var/log/frps.log"
level = "info"
max_days = 3
`, m.port, token)

	os.WriteFile(m.configPath, []byte(cfg), 0644)
}

// Restart 重启 frps
func (m *Manager) Restart() {
	// 停止当前进程
	if m.cmd != nil && m.cmd.Process != nil {
		log.Println("停止 frps...")
		m.cmd.Process.Kill()
		m.cmd.Wait()
	}

	// 等待一下确保进程完全退出
	time.Sleep(2 * time.Second)

	// 重新启动
	frpsBin, err := findFrps()
	if err != nil {
		log.Printf("重启 frps 失败: %v", err)
		return
	}

	m.cmd = exec.Command(frpsBin, "-c", m.configPath)
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	if err := m.cmd.Start(); err != nil {
		log.Printf("重启 frps 失败: %v", err)
		return
	}

	log.Printf("frps 已重启 (pid=%d)", m.cmd.Process.Pid)
}

func (m *Manager) Stop() {
	if m.cmd != nil && m.cmd.Process != nil {
		m.cmd.Process.Kill()
		log.Println("frps 已停止")
	}
}

func (m *Manager) ConnectionCount() int {
	return int(atomic.LoadInt64(&m.connCnt))
}

func (m *Manager) IncrConn() {
	atomic.AddInt64(&m.connCnt, 1)
}

func (m *Manager) DecrConn() {
	atomic.AddInt64(&m.connCnt, -1)
}

func findFrps() (string, error) {
	// 按优先级查找 frps
	candidates := []string{
		frpsBinaryPath,
		"/usr/local/bin/frps",
		"/usr/bin/frps",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	// 尝试 PATH
	if path, err := exec.LookPath("frps"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("在常见路径中未找到 frps")
}

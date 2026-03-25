package frps

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"text/template"
	"bytes"
)

const frpsConfigTpl = `
[common]
bind_port = {{.Port}}
token = {{.Token}}
dashboard_port = {{.DashPort}}
dashboard_user = admin
dashboard_pwd = {{.DashPwd}}
log_file = /var/log/frps.log
log_level = info
log_max_days = 3
`

type Manager struct {
	port    int
	cmd     *exec.Cmd
	connCnt int64
}

func NewManager(port int) *Manager {
	return &Manager{port: port}
}

func (m *Manager) Start() error {
	// 查找 frps 二进制
	frpsBin, err := findFrps()
	if err != nil {
		return fmt.Errorf("frps 未找到: %w", err)
	}

	// 生成配置文件
	cfgPath := "/etc/frps/frps.ini"
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		cfgPath = "/tmp/frps.ini"
	}

	cfg := struct {
		Port     int
		Token    string
		DashPort int
		DashPwd  string
	}{
		Port:     m.port,
		Token:    os.Getenv("FRPS_TOKEN"),
		DashPort: m.port + 1,
		DashPwd:  "jumpfrp-dashboard",
	}

	var buf bytes.Buffer
	tpl := template.Must(template.New("frps").Parse(frpsConfigTpl))
	tpl.Execute(&buf, cfg)

	if err := os.WriteFile(cfgPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("写入 frps 配置失败: %w", err)
	}

	// 启动 frps
	m.cmd = exec.Command(frpsBin, "-c", cfgPath)
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("启动 frps 失败: %w", err)
	}

	log.Printf("frps 已启动 (pid=%d, port=%d)", m.cmd.Process.Pid, m.port)
	return nil
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
		"/usr/local/bin/frps",
		"/usr/bin/frps",
		"/opt/jumpfrp/frps",
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

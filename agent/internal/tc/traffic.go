package tc

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TrafficControl 管理流量控制
type TrafficControl struct {
	mu          sync.RWMutex
	iface       string                    // 网卡名，如 eth0
	connections map[string]*ConnectionInfo // token -> 连接信息
	vipLimits   map[int]int               // VIP等级 -> 带宽限制(Mbps)
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	Token     string
	VIPLevel  int
	Mark      int32 // iptables MARK 值
	IP        string
	MarkedAt  time.Time
}

// MARK 分配：每连接一个唯一 MARK (1-65535)
var nextMark int32 = 1

// MARK 基础值（避免与已有规则冲突）
const markBase = 1000

func New(iface string) *TrafficControl {
	return &TrafficControl{
		iface:       iface,
		connections: make(map[string]*ConnectionInfo),
		vipLimits: map[int]int{
			0: 1,  // Free: 1Mbps
			1: 5,  // Basic: 5Mbps
			2: 20, // Pro: 20Mbps
			3: 100, // Ultimate: 100Mbps
		},
	}
}

// Init 初始化流量控制
func (tc *TrafficControl) Init() error {
	// 1. 清理旧规则
	if err := tc.cleanup(); err != nil {
		log.Printf("[TC] 清理旧规则: %v", err)
	}

	// 2. 创建根队列
	if err := tc.createRootQdisc(); err != nil {
		return fmt.Errorf("创建根队列失败: %w", err)
	}

	// 3. 创建 VIP 等级分类
	if err := tc.createVIPClasses(); err != nil {
		return fmt.Errorf("创建VIP分类失败: %w", err)
	}

	// 4. 初始化 iptables MARK 链
	if err := tc.initIPTables(); err != nil {
		return fmt.Errorf("初始化iptables失败: %w", err)
	}

	log.Printf("[TC] 流量控制初始化完成 (网卡: %s)", tc.iface)
	return nil
}

// createRootQdisc 创建根 HTB 队列
func (tc *TrafficControl) createRootQdisc() error {
	// 删除已存在的队列
	exec.Command("tc", "qdisc", "del", "dev", tc.iface, "root").Run()

	// 添加根 HTB 队列
	cmd := exec.Command("tc", "qdisc", "add", "dev", tc.iface, "root", "handle", "1:", "htb", "default", "9999")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("添加HTB队列: %v", err)
	}

	// 创建根类（总带宽 1Gbps）
	cmd = exec.Command("tc", "class", "add", "dev", tc.iface, "parent", "1:", "classid", "1:9999", "htb", "rate", "1000mbit", "ceil", "1000mbit")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("创建根类: %v", err)
	}

	return nil
}

// createVIPClasses 为每个VIP等级创建分类
func (tc *TrafficControl) createVIPClasses() error {
	// 先清理旧分类
	for vip := 0; vip <= 3; vip++ {
		exec.Command("tc", "class", "del", "dev", tc.iface, "parent", "1:", "classid", fmt.Sprintf("1:%d", vip*10+1)).Run()
	}

	// 为每个VIP等级创建分类
	for vip, rate := range tc.vipLimits {
		classID := fmt.Sprintf("1:%d", vip*10+1) // 1:1, 1:11, 1:21, 1:31
		cmd := exec.Command("tc", "class", "add", "dev", tc.iface, "parent", "1:", "classid", classID, "htb",
			"rate", fmt.Sprintf("%dmbit", rate),
			"ceil", fmt.Sprintf("%dmbit", rate*2),
			"burst", fmt.Sprintf("%dm", rate/10+1))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("创建VIP%d分类(%dmbit): %v", vip, rate, err)
		}
		log.Printf("[TC] 创建VIP%d分类: %dmbit", vip, rate)
	}

	return nil
}

// initIPTables 初始化 iptables MARK 规则
func (tc *TrafficControl) initIPTables() error {
	// 清理旧的 MARK 规则
	exec.Command("iptables", "-t", "mangle", "-F", "OUTPUT").Run()
	exec.Command("iptables", "-t", "mangle", "-F", "PREROUTING").Run()

	// 创建 MARK 链（如果没有）
	exec.Command("iptables", "-t", "mangle", "-N", "TC_MARK").Run()

	// OUTPUT 链：本地发出的流量标记
	cmd := exec.Command("iptables", "-t", "mangle", "-A", "OUTPUT", "-m", "conntrack", "--ctstate", "ESTABLISHED,RELATED", "-j", "TC_MARK")
	cmd.Run()

	return nil
}

// AddConnection 添加一个连接
func (tc *TrafficControl) AddConnection(token string, ip string, vipLevel int) (int32, error) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// 检查是否已存在
	if info, ok := tc.connections[token]; ok {
		info.MarkedAt = time.Now()
		return info.Mark, nil
	}

	// 分配新 MARK
	mark := markBase + nextMark
	nextMark++

	info := &ConnectionInfo{
		Token:    token,
		VIPLevel: vipLevel,
		Mark:     mark,
		IP:       ip,
		MarkedAt: time.Now(),
	}
	tc.connections[token] = info

	// 获取该 VIP 等级的 class ID
	classID := fmt.Sprintf("1:%d", vipLevel*10+1)

	// 添加 iptables MARK 规则
	// 根据源 IP 标记（frpc 连接的源 IP）
	cmd := exec.Command("iptables", "-t", "mangle", "-A", "TC_MARK",
		"-s", ip,
		"-j", "MARK",
		"--set-mark", strconv.FormatInt(int64(mark), 10))
	if err := cmd.Run(); err != nil {
		log.Printf("[TC] 添加MARK规则失败: %v", err)
	}

	// 添加 tc filter 规则
	cmd = exec.Command("tc", "filter", "add", "dev", tc.iface, "parent", "1:",
		"protocol", "ip", "fw", "handle", strconv.FormatInt(int64(mark), 10),
		"classid", classID)
	if err := cmd.Run(); err != nil {
		log.Printf("[TC] 添加filter规则失败: %v", err)
	}

	log.Printf("[TC] 添加连接: token=%s, ip=%s, vip=%d, mark=%d, class=%s", token, ip, vipLevel, mark, classID)
	return mark, nil
}

// RemoveConnection 移除一个连接
func (tc *TrafficControl) RemoveConnection(token string) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	info, ok := tc.connections[token]
	if !ok {
		return nil
	}

	// 删除 iptables MARK 规则
	exec.Command("iptables", "-t", "mangle", "-D", "TC_MARK",
		"-s", info.IP,
		"-j", "MARK",
		"--set-mark", strconv.FormatInt(int64(info.Mark), 10)).Run()

	// 删除 tc filter 规则
	exec.Command("tc", "filter", "del", "dev", tc.iface, "parent", "1:",
		"protocol", "ip", "fw",
		"handle", strconv.FormatInt(int64(info.Mark), 10)).Run()

	delete(tc.connections, token)
	log.Printf("[TC] 移除连接: token=%s, ip=%s", token, info.IP)
	return nil
}

// ParseFrpsLogs 解析 frps 日志，提取连接信息
// 日志格式示例: "login from 192.168.1.100:12345, auth token: user_token_xxx"
func (tc *TrafficControl) ParseFrpsLogs(logPath string) error {
	file, err := os.Open(logPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	loginRegex := regexp.MustCompile(`login from (\S+):\d+, auth token: (\S+)`)
	disconnectRegex := regexp.MustCompile(`(\S+) disconnected`)

	for scanner.Scan() {
		line := scanner.Text()

		// 解析登录
		if matches := loginRegex.FindStringSubmatch(line); len(matches) == 3 {
			ip := matches[1]
			token := matches[2]
			// 估算 VIP 等级（默认 0）
			vipLevel := tc.estimateVIPLevel(token)
			tc.AddConnection(token, ip, vipLevel)
		}

		// 解析断开
		if matches := disconnectRegex.FindStringSubmatch(line); len(matches) == 2 {
			token := matches[1]
			tc.RemoveConnection(token)
		}
	}

	return nil
}

// estimateVIPLevel 估算 VIP 等级（通过 token 或其他方式）
// 这里需要根据实际架构调整，可能需要查询主控获取
func (tc *TrafficControl) estimateVIPLevel(token string) int {
	// 临时实现：默认免费
	// TODO: 通过主控 API 获取用户 VIP 等级
	return 0
}

// SetVIPLimit 设置某 VIP 等级的带宽限制
func (tc *TrafficControl) SetVIPLimit(vipLevel int, rateMbps int) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.vipLimits[vipLevel] = rateMbps
	classID := fmt.Sprintf("1:%d", vipLevel*10+1)

	// 更新 tc 分类
	cmd := exec.Command("tc", "class", "change", "dev", tc.iface, "parent", "1:", "classid", classID, "htb",
		"rate", fmt.Sprintf("%dmbit", rateMbps),
		"ceil", fmt.Sprintf("%dmbit", rateMbps*2))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("更新带宽限制: %v", err)
	}

	log.Printf("[TC] 更新VIP%d带宽限制: %dmbit", vipLevel, rateMbps)
	return nil
}

// Status 获取流量控制状态
func (tc *TrafficControl) Status() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	var buf bytes.Buffer
	buf.WriteString("=== 流量控制状态 ===\n")
	buf.WriteString(fmt.Sprintf("网卡: %s\n", tc.iface))
	buf.WriteString(fmt.Sprintf("活跃连接数: %d\n", len(tc.connections)))
	buf.WriteString("\nVIP 带宽限制:\n")
	for vip, rate := range tc.vipLimits {
		buf.WriteString(fmt.Sprintf("  VIP%d: %dmbit\n", vip, rate))
	}
	buf.WriteString("\n活跃连接:\n")
	for token, info := range tc.connections {
		buf.WriteString(fmt.Sprintf("  %s -> %s (VIP%d, mark=%d)\n", token[:min(20, len(token))], info.IP, info.VIPLevel, info.Mark))
	}
	return buf.String()
}

// Cleanup 清理所有规则
func (tc *TrafficControl) Cleanup() error {
	return tc.cleanup()
}

func (tc *TrafficControl) cleanup() error {
	// 清理 tc
	exec.Command("tc", "qdisc", "del", "dev", tc.iface, "root").Run()

	// 清理 iptables
	exec.Command("iptables", "-t", "mangle", "-F", "TC_MARK").Run()
	exec.Command("iptables", "-t", "mangle", "-X", "TC_MARK").Run()

	tc.mu.Lock()
	tc.connections = make(map[string]*ConnectionInfo)
	tc.mu.Unlock()

	return nil
}

// GetStats 获取流量统计
func (tc *TrafficControl) GetStats() (map[string]string, error) {
	stats := make(map[string]string)

	// 获取各分类的统计
	for vip := 0; vip <= 3; vip++ {
		classID := fmt.Sprintf("1:%d", vip*10+1)
		cmd := exec.Command("tc", "-s", "class", "show", "dev", tc.iface)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		// 解析输出找对应 class
		lines := strings.Split(string(output), "\n")
		for i, line := range lines {
			if strings.Contains(line, classID) && i+1 < len(lines) {
				stats[fmt.Sprintf("vip%d", vip)] = strings.TrimSpace(lines[i+1])
			}
		}
	}

	return stats, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

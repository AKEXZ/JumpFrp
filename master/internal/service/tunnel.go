package service

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/jumpfrp/master/internal/model"
	"gorm.io/gorm"
)

type TunnelService struct {
	db *gorm.DB
}

func NewTunnelService(db *gorm.DB) *TunnelService {
	return &TunnelService{db: db}
}

type CreateTunnelInput struct {
	NodeID    uint   `json:"node_id" binding:"required"`
	Name      string `json:"name" binding:"required,min=1,max=50"`
	Protocol  string `json:"protocol" binding:"required,oneof=tcp udp http https"`
	LocalIP   string `json:"local_ip"`
	LocalPort int    `json:"local_port" binding:"required,min=1,max=65535"`
	Subdomain string `json:"subdomain"`
}

// 更新隧道输入
type UpdateTunnelInput struct {
	NodeID    uint   `json:"node_id"`
	Protocol  string `json:"protocol"`
	LocalIP   string `json:"local_ip"`
	LocalPort int    `json:"local_port"`
}

// 创建隧道
func (s *TunnelService) Create(userID uint, input CreateTunnelInput) (*model.Tunnel, error) {
	// 获取用户信息
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	quota := user.GetQuota()

	// 检查隧道数量限制（只统计已开启的隧道）
	var enabledCount int64
	s.db.Model(&model.Tunnel{}).Where("user_id = ? AND enabled = ?", userID, true).Count(&enabledCount)
	if int(enabledCount) >= quota.MaxTunnels {
		return nil, fmt.Errorf("已达到隧道数量上限 (%d)，请升级 VIP 或关闭其他隧道", quota.MaxTunnels)
	}

	// 检查协议权限
	allowed := false
	for _, p := range quota.Protocols {
		if p == input.Protocol {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("当前 VIP 等级不支持 %s 协议，请升级", input.Protocol)
	}

	// 获取节点信息
	var node model.Node
	if err := s.db.First(&node, input.NodeID).Error; err != nil {
		return nil, errors.New("节点不存在")
	}
	if node.Status == model.NodeStatusOffline {
		return nil, errors.New("该节点当前离线，请选择其他节点")
	}
	if node.MinVIPLevel > user.VIPLevel {
		return nil, fmt.Errorf("该节点需要 VIP %d 及以上才能使用", node.MinVIPLevel)
	}

	// 检查隧道名称唯一性（同用户下）
	var nameCount int64
	s.db.Model(&model.Tunnel{}).Where("user_id = ? AND name = ?", userID, input.Name).Count(&nameCount)
	if nameCount > 0 {
		return nil, errors.New("隧道名称已存在")
	}

	// 检查本地端口是否已被其他隧道占用（同一用户）
	var localPortCount int64
	s.db.Model(&model.Tunnel{}).
		Where("user_id = ? AND local_ip = ? AND local_port = ?", userID, input.LocalIP, input.LocalPort).
		Count(&localPortCount)
	if localPortCount > 0 {
		return nil, errors.New("本地端口已被其他隧道占用")
	}

	// 分配远程端口
	remotePort, err := s.allocatePort(node, userID)
	if err != nil {
		return nil, err
	}

	// 处理子域名（HTTP/HTTPS）
	subdomain := ""
	if input.Protocol == "http" || input.Protocol == "https" {
		if input.Subdomain != "" {
			if !quota.CanSubdomain {
				return nil, errors.New("当前 VIP 等级不支持自定义子域名，请升级至 Pro 或以上")
			}
			// 检查子域名是否已被占用
			var subCount int64
			s.db.Model(&model.Tunnel{}).Where("subdomain = ?", input.Subdomain).Count(&subCount)
			if subCount > 0 {
				return nil, errors.New("子域名已被占用")
			}
			subdomain = input.Subdomain
		}
	}

	localIP := input.LocalIP
	if localIP == "" {
		localIP = "127.0.0.1"
	}

	tunnel := &model.Tunnel{
		UserID:         userID,
		NodeID:         input.NodeID,
		Name:           input.Name,
		Protocol:       input.Protocol,
		LocalIP:        localIP,
		LocalPort:      input.LocalPort,
		RemotePort:     remotePort,
		Subdomain:      subdomain,
		BandwidthLimit: quota.MaxBandwidth,
		Status:         model.TunnelStatusInactive,
	}

	if err := s.db.Create(tunnel).Error; err != nil {
		return nil, err
	}

	// 预加载关联
	s.db.Preload("Node").First(tunnel, tunnel.ID)
	return tunnel, nil
}

// 更新隧道
func (s *TunnelService) Update(userID uint, tunnelID uint, input UpdateTunnelInput) (*model.Tunnel, error) {
	var tunnel model.Tunnel
	if err := s.db.Where("id = ? AND user_id = ?", tunnelID, userID).First(&tunnel).Error; err != nil {
		return nil, errors.New("隧道不存在")
	}

	// 获取用户信息
	var user model.User
	s.db.First(&user, userID)
	quota := user.GetQuota()

	// 如果更换了节点，检查新节点
	if input.NodeID != 0 && input.NodeID != tunnel.NodeID {
		var node model.Node
		if err := s.db.First(&node, input.NodeID).Error; err != nil {
			return nil, errors.New("节点不存在")
		}
		if node.Status == model.NodeStatusOffline {
			return nil, errors.New("该节点当前离线")
		}
		if node.MinVIPLevel > user.VIPLevel {
			return nil, fmt.Errorf("该节点需要 VIP %d 及以上", node.MinVIPLevel)
		}
		tunnel.NodeID = input.NodeID
	}

	// 如果更换了协议
	if input.Protocol != "" && input.Protocol != tunnel.Protocol {
		allowed := false
		for _, p := range quota.Protocols {
			if p == input.Protocol {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("当前 VIP 不支持 %s 协议", input.Protocol)
		}
		tunnel.Protocol = input.Protocol
	}

	// 如果更换了本地IP
	if input.LocalIP != "" {
		tunnel.LocalIP = input.LocalIP
	}

	// 如果更换了本地端口
	if input.LocalPort != 0 {
		// 检查本地端口是否冲突（同用户）
		var count int64
		s.db.Model(&model.Tunnel{}).
			Where("user_id = ? AND id != ? AND local_ip = ? AND local_port = ?",
				userID, tunnelID, tunnel.LocalIP, input.LocalPort).
			Count(&count)
		if count > 0 {
			return nil, errors.New("本地端口已被其他隧道占用")
		}
		tunnel.LocalPort = input.LocalPort
	}

	s.db.Save(&tunnel)
	s.db.Preload("Node").First(&tunnel, tunnelID)
	return &tunnel, nil
}

// 从节点端口池随机分配一个未使用的端口
func (s *TunnelService) allocatePort(node model.Node, userID uint) (int, error) {
	if node.PortRangeStart == 0 || node.PortRangeEnd == 0 {
		return 0, errors.New("节点端口池未配置")
	}

	// 获取该节点已使用的端口
	var usedPorts []int
	s.db.Model(&model.Tunnel{}).
		Where("node_id = ?", node.ID).
		Pluck("remote_port", &usedPorts)

	usedSet := make(map[int]bool)
	for _, p := range usedPorts {
		usedSet[p] = true
	}

	// 解析排除端口
	excludeSet := make(map[int]bool)
	if node.PortExcludes != "" {
		for _, ps := range strings.Split(node.PortExcludes, ",") {
			ps = strings.TrimSpace(ps)
			if p, err := strconv.Atoi(ps); err == nil {
				excludeSet[p] = true
			}
		}
	}

	// 检查用户已分配的端口数量
	var userPortCount int64
	s.db.Model(&model.Tunnel{}).Where("user_id = ? AND node_id = ?", userID, node.ID).Count(&userPortCount)

	var user model.User
	s.db.First(&user, userID)
	quota := user.GetQuota()
	if int(userPortCount) >= quota.MaxPorts {
		return 0, fmt.Errorf("已达到端口数量上限 (%d)，请升级 VIP", quota.MaxPorts)
	}

	// 随机打乱端口范围，找一个可用的
	portRange := node.PortRangeEnd - node.PortRangeStart + 1
	indices := rand.New(rand.NewSource(time.Now().UnixNano())).Perm(portRange)

	for _, idx := range indices {
		port := node.PortRangeStart + idx
		if !usedSet[port] && !excludeSet[port] {
			return port, nil
		}
	}

	return 0, errors.New("节点端口池已耗尽，请联系管理员")
}

// 获取用户隧道列表
func (s *TunnelService) ListByUser(userID uint) ([]model.Tunnel, error) {
	var tunnels []model.Tunnel
	err := s.db.Preload("Node").Where("user_id = ?", userID).Find(&tunnels).Error
	return tunnels, err
}

// 删除隧道
func (s *TunnelService) Delete(userID, tunnelID uint) error {
	var tunnel model.Tunnel
	if err := s.db.Where("id = ? AND user_id = ?", tunnelID, userID).First(&tunnel).Error; err != nil {
		return errors.New("隧道不存在或无权限")
	}
	return s.db.Delete(&tunnel).Error
}

// 生成 frpc 配置文件内容 (TOML 格式，frp 0.61.0+)
func (s *TunnelService) GenFrpcConfig(tunnelID uint) (string, error) {
	var tunnel model.Tunnel
	if err := s.db.Preload("Node").Preload("User").First(&tunnel, tunnelID).Error; err != nil {
		return "", errors.New("隧道不存在")
	}

	// 获取 frps token（从节点配置或使用默认值）
	frpsToken := "default-token"
	if tunnel.Node.Token != "" {
		frpsToken = tunnel.Node.Token
	}

	var cfg strings.Builder

	// common 部分
	cfg.WriteString("[common]\n")
	cfg.WriteString(fmt.Sprintf("server_addr = \"%s\"\n", tunnel.Node.IP))
	cfg.WriteString(fmt.Sprintf("server_port = %d\n", tunnel.Node.FrpsPort))
	cfg.WriteString(fmt.Sprintf("auth.method = \"token\"\n"))
	cfg.WriteString(fmt.Sprintf("auth.token = \"%s\"\n", frpsToken))
	cfg.WriteString(fmt.Sprintf("pool_count = 10\n"))
	cfg.WriteString(fmt.Sprintf("transport.tcp_mux = true\n"))
	cfg.WriteString(fmt.Sprintf("transport.protocol = \"%s\"\n\n", tunnel.Protocol))

	// 隧道部分
	sectionName := tunnel.Name
	if len(sectionName) > 50 {
		sectionName = sectionName[:50]
	}
	cfg.WriteString(fmt.Sprintf("[[proxies]]\n"))
	cfg.WriteString(fmt.Sprintf("name = \"%s\"\n", sectionName))
	cfg.WriteString(fmt.Sprintf("type = \"%s\"\n", tunnel.Protocol))
	cfg.WriteString(fmt.Sprintf("local_ip = \"%s\"\n", tunnel.LocalIP))
	cfg.WriteString(fmt.Sprintf("local_port = %d\n", tunnel.LocalPort))

	switch tunnel.Protocol {
	case "tcp", "udp":
		cfg.WriteString(fmt.Sprintf("remote_port = %d\n", tunnel.RemotePort))
	case "http", "https":
		if tunnel.Subdomain != "" {
			cfg.WriteString(fmt.Sprintf("subdomain = \"%s\"\n", tunnel.Subdomain))
		} else {
			cfg.WriteString(fmt.Sprintf("remote_port = %d\n", tunnel.RemotePort))
		}
	}

	// 带宽限制
	if tunnel.BandwidthLimit > 0 {
		cfg.WriteString(fmt.Sprintf("bandwidth_limit = \"%dMB\"\n", tunnel.BandwidthLimit))
	}

	return cfg.String(), nil
}

package model

import (
	"time"

	"gorm.io/gorm"
)

// 节点状态
const (
	NodeStatusOnline  = "online"
	NodeStatusOffline = "offline"
	NodeStatusMaintain = "maintain"
)

type Node struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	Name             string         `gorm:"size:100" json:"name"`
	Slug             string         `gorm:"uniqueIndex;size:50" json:"slug"`
	IP               string         `gorm:"size:50" json:"ip"`
	Region           string         `gorm:"size:100" json:"region"`
	FrpsPort         int            `gorm:"default:7000" json:"frps_port"`
	AgentPort        int            `gorm:"default:7500" json:"agent_port"`
	AgentToken       string         `gorm:"size:64" json:"-"`
	PortRangeStart   int            `json:"port_range_start"`
	PortRangeEnd     int            `json:"port_range_end"`
	PortExcludes     string         `gorm:"size:500" json:"port_excludes"` // comma separated
	MinVIPLevel      int            `gorm:"column:min_vip_level;default:0" json:"min_vip_level"`
	BandwidthLimit   int            `gorm:"column:bandwidth_limit;default:0" json:"bandwidth_limit"` // Mbps, 0=不限速
	MaxConnections   int            `json:"max_connections"`
	Status           string         `gorm:"size:20;default:'offline'" json:"status"`
	Version          string         `gorm:"size:20" json:"version"`
	LastHeartbeat    *time.Time     `json:"last_heartbeat"`
	CPUUsage         float64        `json:"cpu_usage"`
	MemoryUsage      float64        `json:"memory_usage"`
	CurrentConns     int            `json:"current_conns"`
	Instaled         bool           `json:"installed"`
	Remark           string         `gorm:"size:500" json:"remark"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// 用户隧道
const (
	TunnelStatusActive   = "active"
	TunnelStatusInactive = "inactive"
	TunnelStatusOffline  = "offline"
)

type Tunnel struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	UserID           uint           `gorm:"index" json:"user_id"`
	NodeID           uint           `gorm:"index" json:"node_id"`
	Name             string         `gorm:"size:100" json:"name"`
	Protocol         string         `gorm:"size:10" json:"protocol"` // tcp/udp/http/https
	LocalIP          string         `gorm:"size:50;default:'127.0.0.1'" json:"local_ip"`
	LocalPort        int            `json:"local_port"`
	RemotePort       int            `json:"remote_port"`
	Subdomain        string         `gorm:"size:100" json:"subdomain"`
	BandwidthLimit   int            `gorm:"column:bandwidth_limit" json:"bandwidth_limit"` // Mbps
	Enabled          bool           `gorm:"default:true" json:"enabled"` // 隧道开关，VIP过期关闭
	Status           string         `gorm:"size:20;default:'inactive'" json:"status"`
	LastConnectedAt  *time.Time     `json:"last_connected_at"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	User            User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Node            Node           `gorm:"foreignKey:NodeID" json:"node,omitempty"`
}

// VIP 订单
type VIPOrder struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	UserID       uint           `gorm:"index" json:"user_id"`
	VIPLevel     int            `gorm:"column:vip_level" json:"vip_level"`
	DurationDays int            `json:"duration_days"`
	Price        float64        `json:"price"`
	Status       string         `gorm:"size:20" json:"status"` // pending/paid/cancelled
	ExpireAt     *time.Time     `json:"expire_at"`
	CreatedAt    time.Time      `json:"created_at"`

	User        User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// 管理员操作日志
type AdminLog struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	AdminID     uint           `gorm:"index" json:"admin_id"`
	Action      string         `gorm:"size:50" json:"action"`
	TargetType  string         `gorm:"size:50" json:"target_type"`
	TargetID    uint           `json:"target_id"`
	Detail      string         `gorm:"type:text" json:"detail"`
	IP          string         `gorm:"size:50" json:"ip"`
	CreatedAt   time.Time      `json:"created_at"`

	Admin       User           `gorm:"foreignKey:AdminID" json:"admin,omitempty"`
}

// 子域名申请
type Subdomain struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	UserID       uint           `gorm:"index" json:"user_id"`
	TunnelID     uint           `json:"tunnel_id"`
	Subdomain    string         `gorm:"uniqueIndex;size:100" json:"subdomain"`
	Status       string         `gorm:"size:20" json:"status"` // pending/approved/rejected
	CreatedAt    time.Time      `json:"created_at"`

	User        User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

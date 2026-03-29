package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/jumpfrp/master/internal/model"
	"gorm.io/gorm"
)

// SMTPConfig 邮件配置结构
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	SSL      bool   `json:"ssl"`
	Enabled  bool   `json:"enabled"`
}

// SiteConfig 站点基础配置
type SiteConfig struct {
	SiteName    string `json:"site_name"`
	SiteURL     string `json:"site_url"`
	RegisterOpen bool  `json:"register_open"`
	ICP         string `json:"icp"`
}

type SystemService struct {
	db    *gorm.DB
	mu    sync.RWMutex
	cache map[string]string
}

func NewSystemService(db *gorm.DB) *SystemService {
	svc := &SystemService{db: db, cache: make(map[string]string)}
	svc.loadCache()
	return svc
}

func (s *SystemService) loadCache() {
	var configs []model.SystemConfig
	s.db.Find(&configs)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range configs {
		s.cache[c.Key] = c.Value
	}
}

func (s *SystemService) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache[key]
}

func (s *SystemService) Set(key, value, remark string) error {
	result := s.db.Where(model.SystemConfig{Key: key}).
		Assign(model.SystemConfig{Value: value, Remark: remark}).
		FirstOrCreate(&model.SystemConfig{})
	if result.Error != nil {
		return result.Error
	}
	// 更新已存在记录的 value
	s.db.Model(&model.SystemConfig{}).Where("key = ?", key).Update("value", value)

	s.mu.Lock()
	s.cache[key] = value
	s.mu.Unlock()
	
	// 如果是用户相关的更新，递增配置版本
	if key == "api_token" || strings.HasPrefix(key, "user_") {
		s.IncrementConfigVersion()
	}
	
	return nil
}

// GetConfigVersion 获取当前配置版本号
func (s *SystemService) GetConfigVersion() int {
	versionStr := s.Get("config_version")
	if versionStr == "" {
		return 0
	}
	var version int
	fmt.Sscanf(versionStr, "%d", &version)
	return version
}

// IncrementConfigVersion 递增配置版本号（新增用户时调用）
func (s *SystemService) IncrementConfigVersion() {
	version := s.GetConfigVersion() + 1
	s.Set("config_version", fmt.Sprintf("%d", version), "frps 配置版本号")
}

// GetSMTPConfig 获取 SMTP 配置
func (s *SystemService) GetSMTPConfig() SMTPConfig {
	raw := s.Get("smtp")
	cfg := SMTPConfig{
		Port:    587,
		From:    "noreply@jumpfrp.top",
		Enabled: false,
	}
	if raw != "" {
		json.Unmarshal([]byte(raw), &cfg)
	}
	return cfg
}

// SetSMTPConfig 保存 SMTP 配置
func (s *SystemService) SetSMTPConfig(cfg SMTPConfig) error {
	b, _ := json.Marshal(cfg)
	return s.Set("smtp", string(b), "SMTP 邮件配置")
}

// GenerateFrpsConfig 生成节点的 frps.toml 配置文件（带所有用户 token）
func (s *SystemService) GenerateFrpsConfig(node *model.Node) string {
	var cfg strings.Builder

	cfg.WriteString("# frps.toml - JumpFrp 服务端配置\n")
	cfg.WriteString("# 由主控自动生成，请勿手动修改\n\n")

	cfg.WriteString(fmt.Sprintf("bindPort = %d\n", node.FrpsPort))
	cfg.WriteString(fmt.Sprintf("auth.method = \"token\"\n"))
	cfg.WriteString("\n")

	// 获取所有用户的 API Token
	var users []model.User
	s.db.Where("api_token != '' AND api_token IS NOT NULL").Find(&users)

	// 添加所有用户的 token
	for _, user := range users {
		cfg.WriteString(fmt.Sprintf("[[auth.tokens]]\n"))
		cfg.WriteString(fmt.Sprintf("token = \"%s\"\n", user.APIToken))
		cfg.WriteString("\n")
	}

	// 传输配置
	cfg.WriteString("[transport]\n")
	cfg.WriteString("max_pool_count = 100\n")
	cfg.WriteString("pool_count = 10\n")
	cfg.WriteString("tcp_mux = true\n")
	cfg.WriteString("transport.tcp_mux = true\n")

	// 带宽限制（服务端强制限速）
	if node.BandwidthLimit > 0 {
		cfg.WriteString(fmt.Sprintf("transport.max_bandwidth_per_client = \"%dMB\"\n", node.BandwidthLimit))
	}

	cfg.WriteString("\n")

	// HTTP/HTTPS 配置
	cfg.WriteString("[[vhost.httpRoutes]]\n")
	cfg.WriteString("custom_domains = [\"*.jumpfrp.top\"]\n")
	cfg.WriteString("\n")

	cfg.WriteString("[[vhost.httpsRoutes]]\n")
	cfg.WriteString("custom_domains = [\"*.jumpfrp.top\"]\n")
	cfg.WriteString("\n")

	// 日志配置
	cfg.WriteString("[log]\n")
	cfg.WriteString("to = \"/var/log/frps.log\"\n")
	cfg.WriteString("level = \"info\"\n")
	cfg.WriteString("max_days = 3\n")

	return cfg.String()
}

// GetSiteConfig 获取站点配置
func (s *SystemService) GetSiteConfig() SiteConfig {
	raw := s.Get("site")
	cfg := SiteConfig{
		SiteName:     "JumpFrp",
		SiteURL:      "https://jumpfrp.top",
		RegisterOpen: true,
	}
	if raw != "" {
		json.Unmarshal([]byte(raw), &cfg)
	}
	return cfg
}

// SetSiteConfig 保存站点配置
func (s *SystemService) SetSiteConfig(cfg SiteConfig) error {
	b, _ := json.Marshal(cfg)
	return s.Set("site", string(b), "站点基础配置")
}

package service

import (
	"encoding/json"
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
	return nil
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

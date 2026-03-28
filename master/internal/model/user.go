package model

import (
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// VIP 等级
const (
	VIPFree     = 0
	VIPBasic    = 1
	VIPPro      = 2
	VIPUltimate = 3
)

// 用户状态
const (
	UserStatusActive  = "active"
	UserStatusBanned  = "banned"
)

type User struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:50" json:"username"`
	Email        string         `gorm:"uniqueIndex;size:100" json:"email"`
	PasswordHash string         `gorm:"size:255" json:"-"`
	VIPLevel     int            `gorm:"column:vip_level;default:0" json:"vip_level"`
	VIPExpireAt  *time.Time     `json:"vip_expire_at"`
	APIToken     string         `gorm:"uniqueIndex;size:64" json:"api_token"`
	Status       string         `gorm:"size:20;default:'active'" json:"status"`
	EmailVerified bool          `gorm:"default:false" json:"email_verified"`
	VerifyCode   string         `gorm:"size:10" json:"-"`
	VerifyExpire *time.Time     `json:"-"`
	ResetToken   string         `gorm:"size:64" json:"-"`
	ResetExpire  *time.Time     `json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// VIP 权益配置
type VIPQuota struct {
	MaxTunnels   int
	MaxPorts     int
	MaxBandwidth int // Mbps
	Protocols    []string
	CanSubdomain bool
	CanFixedPort bool
}

var VIPQuotas = map[int]VIPQuota{
	VIPFree: {
		MaxTunnels:   1,
		MaxPorts:     3,
		MaxBandwidth: 1,
		Protocols:    []string{"tcp"},
		CanSubdomain: false,
		CanFixedPort: false,
	},
	VIPBasic: {
		MaxTunnels:   5,
		MaxPorts:     10,
		MaxBandwidth: 5,
		Protocols:    []string{"tcp", "udp"},
		CanSubdomain: false,
		CanFixedPort: false,
	},
	VIPPro: {
		MaxTunnels:   20,
		MaxPorts:     50,
		MaxBandwidth: 20,
		Protocols:    []string{"tcp", "udp", "http", "https"},
		CanSubdomain: true,
		CanFixedPort: false,
	},
	VIPUltimate: {
		MaxTunnels:   9999,
		MaxPorts:     200,
		MaxBandwidth: 100,
		Protocols:    []string{"tcp", "udp", "http", "https"},
		CanSubdomain: true,
		CanFixedPort: true,
	},
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) GetQuota() VIPQuota {
	// 检查 VIP 是否过期
	if u.VIPLevel > VIPFree && u.VIPExpireAt != nil && u.VIPExpireAt.Before(time.Now()) {
		return VIPQuotas[VIPFree]
	}
	if q, ok := VIPQuotas[u.VIPLevel]; ok {
		return q
	}
	return VIPQuotas[VIPFree]
}

func CreateDefaultAdmin(db *gorm.DB) {
	var count int64
	db.Model(&User{}).Where("username = ?", "admin").Count(&count)
	if count > 0 {
		return
	}
	admin := &User{
		Username:      "admin",
		Email:         "admin@jumpfrp.top",
		VIPLevel:      VIPUltimate,
		Status:        UserStatusActive,
		EmailVerified: true,
		APIToken:      "admin-token-change-in-production",
	}
	if err := admin.SetPassword("admin123456"); err != nil {
		log.Printf("failed to set admin password: %v", err)
		return
	}
	db.Create(admin)
	log.Println("Default admin created: admin / admin123456 (please change password!)")
}

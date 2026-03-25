package model

import "time"

// 系统配置（KV 存储）
type SystemConfig struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Key       string    `gorm:"uniqueIndex;size:100" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	Remark    string    `gorm:"size:200" json:"remark"`
	UpdatedAt time.Time `json:"updated_at"`
}

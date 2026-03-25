package model

import "time"

// 流量日志（按天聚合）
type TrafficLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	TunnelID  uint      `gorm:"index" json:"tunnel_id"`
	NodeID    uint      `gorm:"index" json:"node_id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	BytesIn   int64     `json:"bytes_in"`
	BytesOut  int64     `json:"bytes_out"`
	Date      string    `gorm:"index;size:10" json:"date"` // YYYY-MM-DD
	CreatedAt time.Time `json:"created_at"`
}

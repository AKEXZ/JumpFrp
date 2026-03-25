package admin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jumpfrp/master/internal/model"
	"gorm.io/gorm"
)

// Agent 注册
func agentRegister(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			NodeID string `json:"node_id" binding:"required"`
			Token  string `json:"token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}

		var node model.Node
		if err := db.Where("slug = ? AND agent_token = ?", req.NodeID, req.Token).First(&node).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "节点认证失败"})
			return
		}

		now := time.Now()
		db.Model(&node).Updates(map[string]interface{}{
			"status":        model.NodeStatusOnline,
			"installed":     true,
			"last_heartbeat": now,
		})

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "注册成功"})
	}
}

// Agent 心跳
func agentHeartbeat(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			NodeID        string  `json:"node_id" binding:"required"`
			Token         string  `json:"token" binding:"required"`
			CPUUsage      float64 `json:"cpu_usage"`
			MemoryUsage   float64 `json:"memory_usage"`
			CurrentConns  int     `json:"current_conns"`
			Version       string  `json:"version"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}

		var node model.Node
		if err := db.Where("slug = ? AND agent_token = ?", req.NodeID, req.Token).First(&node).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "节点认证失败"})
			return
		}

		now := time.Now()
		db.Model(&node).Updates(map[string]interface{}{
			"status":         model.NodeStatusOnline,
			"cpu_usage":      req.CPUUsage,
			"memory_usage":   req.MemoryUsage,
			"current_conns":  req.CurrentConns,
			"last_heartbeat": now,
			"version":        req.Version,
		})

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "心跳已接收"})
	}
}

// 公开路由（不需要 JWT，但需要 Agent Token）
func RegisterAgentRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	rg.POST("/register", agentRegister(db))
	rg.POST("/heartbeat", agentHeartbeat(db))
}

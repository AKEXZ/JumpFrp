package admin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jumpfrp/master/internal/model"
	"github.com/jumpfrp/master/internal/service"
	"gorm.io/gorm"
)

// Agent 注册
func agentRegister(db *gorm.DB, sysSvc *service.SystemService) gin.HandlerFunc {
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

		// 返回 frps.toml 配置
		frpsConfig := sysSvc.GenerateFrpsConfig(&node)

		c.JSON(http.StatusOK, gin.H{
			"code":        0,
			"msg":         "注册成功",
			"frps_config": frpsConfig,
		})
	}
}

// Agent 心跳
func agentHeartbeat(db *gorm.DB, sysSvc *service.SystemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			NodeID        string  `json:"node_id" binding:"required"`
			Token         string  `json:"token" binding:"required"`
			CPUUsage      float64 `json:"cpu_usage"`
			MemoryUsage   float64 `json:"memory_usage"`
			CurrentConns  int     `json:"current_conns"`
			Version       string  `json:"version"`
			ConfigVersion int     `json:"config_version"` // 配置版本，变化时需要重新加载
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

		// 返回心跳响应，包含是否需要更新配置
		c.JSON(http.StatusOK, gin.H{
			"code":           0,
			"msg":            "心跳已接收",
			"frps_config":    sysSvc.GenerateFrpsConfig(&node),
			"config_version": 1, // 配置版本号
		})
	}
}

// 公开路由（不需要 JWT，但需要 Agent Token）
func RegisterAgentRoutes(rg *gin.RouterGroup, db *gorm.DB, sysSvc *service.SystemService) {
	rg.POST("/register", agentRegister(db, sysSvc))
	rg.POST("/heartbeat", agentHeartbeat(db, sysSvc))
	rg.POST("/get-user-vip", getUserVIPLevel(db))
	rg.POST("/get-all-tokens", getAllUserTokens(db))
}

// getUserVIPLevel 获取用户的 VIP 等级（Agent 查询用）
func getUserVIPLevel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}

		// 通过用户的 API Token 查找用户
		var user model.User
		if err := db.Where("api_token = ?", req.Token).First(&user).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "vip_level": 0})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "vip_level": user.VIPLevel})
	}
}

// 获取所有用户的 API Token（供 Agent 更新 frps 配置）
func getAllUserTokens(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			NodeID string `json:"node_id" binding:"required"`
			Token  string `json:"token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}

		// 验证 Agent 身份
		var node model.Node
		if err := db.Where("slug = ? AND agent_token = ?", req.NodeID, req.Token).First(&node).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "节点认证失败"})
			return
		}

		// 获取所有用户的 API Token
		var users []model.User
		db.Where("api_token != '' AND api_token IS NOT NULL").Pluck("api_token", &users)

		var tokens []string
		for _, user := range users {
			tokens = append(tokens, user.APIToken)
		}

		c.JSON(http.StatusOK, gin.H{
			"code":   0,
			"tokens": tokens,
		})
	}
}

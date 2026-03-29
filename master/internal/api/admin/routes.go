package admin

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jumpfrp/master/config"
	"github.com/jumpfrp/master/internal/middleware"
	"github.com/jumpfrp/master/internal/model"
	"github.com/jumpfrp/master/internal/service"
	"gorm.io/gorm"
)

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config, sysSvc *service.SystemService) {
	auth := rg.Group("", middleware.JWTAuth(cfg), middleware.AdminRequired())
	{
		// 仪表盘统计
		auth.GET("/dashboard", func(c *gin.Context) {
			var userCount, nodeCount, tunnelCount, onlineNodeCount int64
			var vipCounts [4]int64

			db.Model(&model.User{}).Count(&userCount)
			db.Model(&model.Node{}).Count(&nodeCount)
			db.Model(&model.Node{}).Where("status = ?", model.NodeStatusOnline).Count(&onlineNodeCount)
			db.Model(&model.Tunnel{}).Count(&tunnelCount)

			for i := 0; i <= 3; i++ {
				db.Model(&model.User{}).Where("vip_level = ?", i).Count(&vipCounts[i])
			}

			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"data": gin.H{
					"users":        userCount,
					"nodes":        nodeCount,
					"online_nodes": onlineNodeCount,
					"tunnels":      tunnelCount,
					"vip_dist": gin.H{
						"free":     vipCounts[0],
						"basic":    vipCounts[1],
						"pro":      vipCounts[2],
						"ultimate": vipCounts[3],
					},
				},
			})
		})

		// 用户管理
		auth.GET("/users", listUsers(db))
		auth.POST("/users", createUser(db))
		auth.PUT("/users/:id/vip", setUserVIP(db))
		auth.PUT("/users/:id/ban", banUser(db))
		auth.PUT("/users/:id/password", resetUserPassword(db))
		auth.DELETE("/users/:id", deleteUser(db))

		// 节点管理
		auth.GET("/nodes", listNodes(db))
		auth.POST("/nodes", createNode(db))
		auth.PUT("/nodes/:id", updateNode(db))
		auth.DELETE("/nodes/:id", deleteNode(db))
		auth.GET("/nodes/:id/install-cmd", getInstallCmd(db, cfg))

		// 隧道管理
		auth.GET("/tunnels", listAllTunnels(db))
		auth.DELETE("/tunnels/:id", forceDeleteTunnel(db))

		// VIP 管理
		auth.GET("/orders", listOrders(db))
		auth.POST("/vip/grant", adminGrantVIP(db))

		// 域名管理
		auth.GET("/subdomains", listSubdomains(db))
		auth.POST("/subdomains", createSubdomain(db))
		auth.PUT("/subdomains/:id/approve", approveSubdomain(db))
		auth.DELETE("/subdomains/:id", deleteSubdomain(db))

		// 系统管理
		auth.POST("/system/force-update-config", func(c *gin.Context) {
			sysSvc.IncrementConfigVersion()
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "已通知所有 Agent 更新配置，请等待 30 秒"})
		})

		// 系统设置
		RegisterSettingsRoutes(auth, db, sysSvc)
	}

	// 公开路由（不需要认证）
	// 节点自动注册（安装脚本使用）
	rg.POST("/node/auto-register", func(c *gin.Context) {
		var req struct {
			Name   string `json:"name" binding:"required"`
			IP     string `json:"ip" binding:"required"`
			Region string `json:"region" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}

		// 创建节点
		node := model.Node{
			Name:       req.Name,
			IP:         req.IP,
			Region:     req.Region,
			FrpsPort:   7000,
			AgentPort:  7500,
			Status:     model.NodeStatusOffline,
			AgentToken: generateToken(32),
			Slug:       fmt.Sprintf("node-%s", generateToken(6)),
		}

		if err := db.Create(&node).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建节点失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "节点创建成功",
			"data": gin.H{
				"node_id": node.Slug,
				"token":   node.AgentToken,
				"name":    node.Name,
				"ip":      node.IP,
				"region":  node.Region,
			},
		})
	})
}

func generateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}

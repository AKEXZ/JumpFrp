package admin

import (
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

		// 系统设置
		RegisterSettingsRoutes(auth, db, sysSvc)
	}
}

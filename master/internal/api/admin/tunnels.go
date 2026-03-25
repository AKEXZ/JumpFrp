package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jumpfrp/master/internal/model"
	"gorm.io/gorm"
)

func listAllTunnels(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tunnels []model.Tunnel
		db.Preload("User").Preload("Node").Find(&tunnels)
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": tunnels})
	}
}

func forceDeleteTunnel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&model.Tunnel{}, id)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "已强制删除"})
	}
}

package admin

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jumpfrp/master/config"
	"github.com/jumpfrp/master/internal/model"
	"gorm.io/gorm"
)

func listUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []model.User
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
		keyword := c.Query("keyword")

		query := db.Model(&model.User{})
		if keyword != "" {
			query = query.Where("username LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}

		var total int64
		query.Count(&total)
		query.Offset((page - 1) * size).Limit(size).Find(&users)

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{"list": users, "total": total},
		})
	}
}

func setUserVIP(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			VIPLevel int `json:"vip_level" binding:"min=0,max=3"`
			Days     int `json:"days" binding:"required,min=1"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}

		expire := time.Now().AddDate(0, 0, req.Days)
		result := db.Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{
			"vip_level":     req.VIPLevel,
			"vip_expire_at": expire,
		})
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": result.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "VIP 已设置"})
	}
}

func banUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Ban bool `json:"ban"`
		}
		c.ShouldBindJSON(&req)

		status := model.UserStatusActive
		if req.Ban {
			status = model.UserStatusBanned
		}
		db.Model(&model.User{}).Where("id = ?", id).Update("status", status)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
	}
}

func listNodes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var nodes []model.Node
		db.Find(&nodes)
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": nodes})
	}
}

func createNode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var node model.Node
		if err := c.ShouldBindJSON(&node); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		node.AgentToken = newToken(32)
		node.Status = model.NodeStatusOffline

		if err := db.Create(&node).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "节点创建成功", "data": node})
	}
}

func updateNode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var node model.Node
		if err := db.First(&node, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "节点不存在"})
			return
		}
		if err := c.ShouldBindJSON(&node); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		db.Save(&node)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功", "data": node})
	}
}

func deleteNode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&model.Node{}, id)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	}
}

func getInstallCmd(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var node model.Node
		if err := db.First(&node, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "节点不存在"})
			return
		}

		masterURL := "https://api.jumpfrp.top"
		cmd := fmt.Sprintf(
			"bash <(wget -qO- %s/install.sh) --node-id %s --token %s --master-url %s --frps-port %d --agent-port %d",
			masterURL, node.Slug, node.AgentToken, masterURL, node.FrpsPort, node.AgentPort,
		)

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"command":     cmd,
				"node_id":     node.Slug,
				"agent_token": node.AgentToken,
			},
		})
	}
}

func newToken(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

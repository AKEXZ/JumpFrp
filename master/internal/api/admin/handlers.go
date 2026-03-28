package admin

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
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

		var user model.User
		if err := db.First(&user, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "用户不存在"})
			return
		}

		oldLevel := user.VIPLevel

		// 如果设置为 Free（Days=0 表示永久Free，即到期）
		var expire *time.Time
		if req.Days > 0 {
			e := time.Now().AddDate(0, 0, req.Days)
			expire = &e
		}

		user.VIPLevel = req.VIPLevel
		user.VIPExpireAt = expire

		// 使用 GORM 的 Select 指定要更新的字段（使用模型字段名）
		result := db.Select("VIPLevel", "VIPExpireAt").Updates(&user)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": result.Error.Error()})
			return
		}

		// 如果是降级/到期（VIPLevel = 0 或 Days = 0），关闭所有隧道
		if req.VIPLevel == 0 || req.Days == 0 {
			db.Model(&model.Tunnel{}).Where("user_id = ?", id).Updates(map[string]interface{}{
				"Enabled":       false,
				"BandwidthLimit": 1, // Free 带宽 1Mbps
			})
			log.Printf("[VIP] 用户 %s 被设置为 Free，已关闭所有隧道", user.Username)
		} else if req.VIPLevel > oldLevel {
			// 如果升级 VIP，自动调整现有隧道的带宽限制并开启隧道
			upgradeTunnelBandwidthAndEnable(db, id, req.VIPLevel)
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "VIP 已设置"})
	}
}

// upgradeTunnelBandwidthAndEnable 升级用户所有隧道的带宽限制并开启
func upgradeTunnelBandwidthAndEnable(db *gorm.DB, userID interface{}, newLevel int) {
	// 获取新 VIP 等级的带宽限制
	quota := getVIPQuota(newLevel)
	newBandwidth := quota.MaxBandwidth

	// 更新该用户所有隧道的带宽限制并开启
	result := db.Model(&model.Tunnel{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"BandwidthLimit": newBandwidth,
		"Enabled":        true,
	})

	log.Printf("[VIP] 用户 %v 升级至 VIP%d，批量更新 %d 条隧道的带宽限制为 %d Mbps 并开启",
		userID, newLevel, result.RowsAffected, newBandwidth)
}

// getVIPQuota 获取 VIP 等级配置
func getVIPQuota(level int) model.VIPQuota {
	quotas := map[int]model.VIPQuota{
		0: {MaxBandwidth: 1},   // Free: 1Mbps
		1: {MaxBandwidth: 5},   // Basic: 5Mbps
		2: {MaxBandwidth: 20},   // Pro: 20Mbps
		3: {MaxBandwidth: 100}, // Ultimate: 100Mbps
	}
	if q, ok := quotas[level]; ok {
		return q
	}
	return quotas[0]
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

		// 验证必填字段
		if node.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "节点名称不能为空"})
			return
		}
		if node.IP == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "IP 地址不能为空"})
			return
		}
		if node.Region == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "地区不能为空"})
			return
		}

		node.AgentToken = newToken(32)
		node.Status = model.NodeStatusOffline

		// 如果没有提供 slug，自动生成
		if node.Slug == "" {
			node.Slug = fmt.Sprintf("node-%s", newToken(6))
		}

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

func deleteUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		// 不允许删除管理员
		var user model.User
		if err := db.First(&user, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "用户不存在"})
			return
		}
		if user.Username == "admin" {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "不能删除管理员账号"})
			return
		}
		// 删除用户的所有隧道
		db.Where("user_id = ?", id).Delete(&model.Tunnel{})
		// 删除用户
		db.Delete(&model.User{}, id)
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

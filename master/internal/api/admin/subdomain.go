package admin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jumpfrp/master/internal/model"
	"gorm.io/gorm"
)

// 列出所有域名申请
func listSubdomains(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.Query("status") // pending/approved/rejected/all
		var subdomains []model.Subdomain

		query := db.Preload("User").Preload("Tunnel")
		if status != "" && status != "all" {
			query = query.Where("status = ?", status)
		}
		query.Order("created_at DESC").Find(&subdomains)

		c.JSON(http.StatusOK, gin.H{"code": 0, "data": subdomains})
	}
}

// 审批域名申请
func approveSubdomain(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Approve bool `json:"approve"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}

		var subdomain model.Subdomain
		if err := db.First(&subdomain, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "域名申请不存在"})
			return
		}

		status := "approved"
		if !req.Approve {
			status = "rejected"
		}

		db.Model(&subdomain).Update("status", status)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "审批成功"})
	}
}

// 删除域名
func deleteSubdomain(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&model.Subdomain{}, id)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	}
}

// 手动添加域名（管理员直接为用户添加）
func createSubdomain(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID    uint   `json:"user_id" binding:"required"`
			TunnelID  uint   `json:"tunnel_id"`
			Subdomain string `json:"subdomain" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}

		// 检查域名是否已存在
		var count int64
		db.Model(&model.Subdomain{}).Where("subdomain = ?", req.Subdomain).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "域名已被占用"})
			return
		}

		subdomain := model.Subdomain{
			UserID:    req.UserID,
			TunnelID:  req.TunnelID,
			Subdomain: req.Subdomain,
			Status:    "approved",
			CreatedAt: time.Now(),
		}
		if err := db.Create(&subdomain).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "添加成功", "data": subdomain})
	}
}
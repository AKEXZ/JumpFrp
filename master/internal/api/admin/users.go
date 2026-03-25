package admin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jumpfrp/master/internal/model"
	"gorm.io/gorm"
)

func createUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username      string `json:"username" binding:"required,min=3,max=20"`
			Email         string `json:"email" binding:"required,email"`
			Password      string `json:"password" binding:"required,min=8"`
			VIPLevel      int    `json:"vip_level"`
			VIPDays       int    `json:"vip_days"`
			EmailVerified bool   `json:"email_verified"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}

		// 检查用户名/邮箱唯一性
		var count int64
		db.Model(&model.User{}).Where("username = ? OR email = ?", req.Username, req.Email).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "用户名或邮箱已存在"})
			return
		}

		user := &model.User{
			Username:      req.Username,
			Email:         req.Email,
			VIPLevel:      req.VIPLevel,
			Status:        model.UserStatusActive,
			EmailVerified: req.EmailVerified,
			APIToken:      newToken(32),
		}
		if err := user.SetPassword(req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}

		// 设置 VIP 到期时间
		if req.VIPLevel > 0 && req.VIPDays > 0 {
			expire := time.Now().AddDate(0, 0, req.VIPDays)
			user.VIPExpireAt = &expire
		}

		if err := db.Create(user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "用户创建成功", "data": user})
	}
}

func resetUserPassword(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Password string `json:"password" binding:"required,min=8"`
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
		if err := user.SetPassword(req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}
		db.Save(&user)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "密码已重置"})
	}
}

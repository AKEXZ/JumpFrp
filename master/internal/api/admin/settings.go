package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jumpfrp/master/internal/service"
	"gorm.io/gorm"
)

func getSettings(sysSvc *service.SystemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"smtp": sysSvc.GetSMTPConfig(),
				"site": sysSvc.GetSiteConfig(),
			},
		})
	}
}

func saveSMTP(sysSvc *service.SystemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cfg service.SMTPConfig
		if err := c.ShouldBindJSON(&cfg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		if err := sysSvc.SetSMTPConfig(cfg); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "SMTP 配置已保存"})
	}
}

func saveSite(sysSvc *service.SystemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cfg service.SiteConfig
		if err := c.ShouldBindJSON(&cfg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		if err := sysSvc.SetSiteConfig(cfg); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "站点配置已保存"})
	}
}

func testSMTP(sysSvc *service.SystemService, db interface{}) gin.HandlerFunc {
	mailSvc := service.NewMailService(sysSvc)
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		// 直接发送测试邮件，不调用 SendVerifyCode（避免创建用户）
		err := mailSvc.Send(req.Email, "【JumpFrp】SMTP 测试邮件", service.MailData{
			Username: "测试用户",
			Content:  "这是一封测试邮件，用于验证 SMTP 配置是否正确。如果您收到此邮件，说明配置成功！",
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "发送失败: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "测试邮件已发送，请检查收件箱"})
	}
}

// RegisterSettingsRoutes 注册系统设置路由（在 RegisterRoutes 中调用）
func RegisterSettingsRoutes(auth *gin.RouterGroup, db *gorm.DB, sysSvc *service.SystemService) {
	auth.GET("/settings", getSettings(sysSvc))
	auth.POST("/settings/smtp", saveSMTP(sysSvc))
	auth.POST("/settings/site", saveSite(sysSvc))
	auth.POST("/settings/smtp/test", testSMTP(sysSvc, db))
}

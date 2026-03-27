package user

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jumpfrp/master/config"
	"github.com/jumpfrp/master/internal/middleware"
	"github.com/jumpfrp/master/internal/model"
	"github.com/jumpfrp/master/internal/service"
	"gorm.io/gorm"
)
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config, sysSvc *service.SystemService) {
	authSvc := service.NewAuthService(db, cfg, sysSvc)
	tunnelSvc := service.NewTunnelService(db)
	vipSvc := service.NewVIPService(db)

	// ── 公开路由 ──────────────────────────────────────────
	rg.POST("/auth/send-code", middleware.AuthRateLimit(), func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		if err := authSvc.SendVerifyCode(req.Email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "验证码已发送"})
	})

	rg.POST("/auth/register", func(c *gin.Context) {
		var input service.RegisterInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		user, err := authSvc.Register(input)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "注册成功", "data": user})
	})

	rg.POST("/auth/login", middleware.AuthRateLimit(), func(c *gin.Context) {
		var input service.LoginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		token, user, err := authSvc.Login(input)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 0, "msg": "登录成功",
			"data": gin.H{"token": token, "user": user},
		})
	})

	// 可用节点列表（公开，用于前台展示）
	rg.GET("/nodes", func(c *gin.Context) {
		nodes := make([]model.Node, 0) // 初始化为空切片，避免 JSON 序列化为 null
		// 返回所有节点（包括离线的），便于测试
		db.Select("id,name,slug,ip,region,frps_port,min_vip_level,status,current_conns,max_connections").
			Order("id ASC").
			Find(&nodes)
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": nodes})
	})

	// ── 需要登录的路由 ─────────────────────────────────────
	auth := rg.Group("", middleware.JWTAuth(cfg))

	// 个人信息
	auth.GET("/profile", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var user model.User
		db.First(&user, userID)
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": user})
	})

	// 修改密码
	auth.PUT("/password", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var req struct {
			OldPassword string `json:"old_password" binding:"required"`
			NewPassword string `json:"new_password" binding:"required,min=8"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		var user model.User
		db.First(&user, userID)
		if !user.CheckPassword(req.OldPassword) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "原密码错误"})
			return
		}
		user.SetPassword(req.NewPassword)
		db.Save(&user)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "密码修改成功"})
	})

	// ── VIP ───────────────────────────────────────────────
	// 套餐列表（公开）
	rg.GET("/vip/plans", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": vipSvc.GetPlans()})
	})

	// 我的 VIP 信息
	auth.GET("/vip/info", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		user, err := vipSvc.GetUserVIP(userID.(uint))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}
		quota := user.GetQuota()
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"user":  user,
				"quota": quota,
			},
		})
	})

	// 我的订单
	auth.GET("/vip/orders", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		orders, err := vipSvc.GetUserOrders(userID.(uint))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": orders})
	})

	// ── 隧道管理 ──────────────────────────────────────────	// 隧道列表
	auth.GET("/tunnels", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		tunnels, err := tunnelSvc.ListByUser(userID.(uint))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": tunnels})
	})

	// 创建隧道
	auth.POST("/tunnels", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var input service.CreateTunnelInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		tunnel, err := tunnelSvc.Create(userID.(uint), input)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "隧道创建成功", "data": tunnel})
	})

	// 删除隧道
	auth.DELETE("/tunnels/:id", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		tunnelID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := tunnelSvc.Delete(userID.(uint), uint(tunnelID)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	})

	// 获取 frpc 配置文件
	auth.GET("/tunnels/:id/frpc-config", func(c *gin.Context) {
		tunnelID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		cfg, err := tunnelSvc.GenFrpcConfig(uint(tunnelID))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		c.Header("Content-Type", "text/plain")
		c.Header("Content-Disposition", "attachment; filename=frpc.ini")
		c.String(http.StatusOK, cfg)
	})

	// ── 域名管理 ──────────────────────────────────────────
	// 我的域名列表
	auth.GET("/subdomains", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var subdomains []model.Subdomain
		db.Where("user_id = ?", userID).Order("created_at DESC").Find(&subdomains)
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": subdomains})
	})

	// 申请域名
	auth.POST("/subdomains", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var req struct {
			TunnelID  uint   `json:"tunnel_id" binding:"required"`
			Subdomain string `json:"subdomain" binding:"required,min=3,max=50"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}

		// 检查用户 VIP 等级（Pro+ 才能申请自定义域名）
		var user model.User
		db.First(&user, userID)
		if user.VIPLevel < 2 {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "需要 Pro 或以上 VIP 等级才能申请自定义域名"})
			return
		}

		// 检查隧道是否属于当前用户
		var tunnel model.Tunnel
		if err := db.First(&tunnel, req.TunnelID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "隧道不存在"})
			return
		}
		if tunnel.UserID != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作此隧道"})
			return
		}

		// 检查协议是否为 http/https
		if tunnel.Protocol != "http" && tunnel.Protocol != "https" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "仅支持 HTTP/HTTPS 隧道绑定域名"})
			return
		}

		// 检查域名是否已存在
		var count int64
		db.Model(&model.Subdomain{}).Where("subdomain = ?", req.Subdomain).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "域名已被占用"})
			return
		}

		// 创建申请（自动审批）
		subdomain := model.Subdomain{
			UserID:    userID.(uint),
			TunnelID:  req.TunnelID,
			Subdomain: req.Subdomain,
			Status:    "approved",
			CreatedAt: time.Now(),
		}
		if err := db.Create(&subdomain).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}

		// 更新隧道的子域名
		db.Model(&tunnel).Update("subdomain", req.Subdomain)

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "域名绑定成功", "data": subdomain})
	})

	// 删除域名
	auth.DELETE("/subdomains/:id", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		subdomainID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

		var subdomain model.Subdomain
		if err := db.First(&subdomain, subdomainID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "域名不存在"})
			return
		}
		if subdomain.UserID != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作"})
			return
		}

		// 清除隧道的子域名
		db.Model(&model.Tunnel{}).Where("id = ?", subdomain.TunnelID).Update("subdomain", "")

		db.Delete(&subdomain)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
	})
}

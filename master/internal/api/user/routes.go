package user

import (
	"net/http"
	"strconv"

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
		var nodes []model.Node
		// 返回在线和维护状态的节点（不返回离线节点）
		db.Where("status IN ?", []string{model.NodeStatusOnline, model.NodeStatusMaintain}).
			Select("id,name,slug,ip,region,frps_port,min_vip_level,status,current_conns,max_connections").
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
}

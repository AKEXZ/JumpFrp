package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jumpfrp/master/config"
	"github.com/jumpfrp/master/internal/api/admin"
	userApi "github.com/jumpfrp/master/internal/api/user"
	"github.com/jumpfrp/master/internal/middleware"
	"github.com/jumpfrp/master/internal/model"
	"github.com/jumpfrp/master/internal/scheduler"
	"github.com/jumpfrp/master/internal/service"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db, err := gorm.Open(sqlite.Open(cfg.Database.Path), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(
		&model.User{},
		&model.Node{},
		&model.Tunnel{},
		&model.VIPOrder{},
		&model.AdminLog{},
		&model.Subdomain{},
		&model.TrafficLog{},
		&model.SystemConfig{},
	)
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// 创建默认管理员
	model.CreateDefaultAdmin(db)

	// 初始化系统配置服务（全局单例，供邮件/设置共享）
	sysSvc := service.NewSystemService(db)

	// 启动定时任务
	sched := scheduler.New(db, cfg, sysSvc)
	sched.Start()

	// 初始化 Gin
	if cfg.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// 中间件
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit())
	r.Use(middleware.SecurityHeaders())

	// API 路由组
	api := r.Group("/api")
	{
		// 用户端路由
		userApi.RegisterRoutes(api.Group("/user"), db, cfg, sysSvc)

		// 管理员路由
		admin.RegisterRoutes(api.Group("/admin"), db, cfg, sysSvc)

		// Agent 路由（节点心跳/注册，无需 JWT）
		admin.RegisterAgentRoutes(api.Group("/agent"), db)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 安装/卸载脚本下载
	r.GET("/install.sh", func(c *gin.Context) {
		c.File("./scripts/install.sh")
	})
	r.GET("/uninstall.sh", func(c *gin.Context) {
		c.File("./scripts/uninstall.sh")
	})

	// Agent 二进制下载（压缩后的 .gz 文件）
	r.GET("/download/agent-linux-amd64", func(c *gin.Context) {
		c.File("./agent/jumpfrp-agent-linux-amd64.gz")
	})
	r.GET("/download/agent-linux-arm64", func(c *gin.Context) {
		c.File("./agent/jumpfrp-agent-linux-arm64.gz")
	})

	// 前端静态文件（生产环境）
	r.Static("/assets", "./web/assets")
	r.StaticFile("/favicon.ico", "./web/favicon.ico")
	r.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "接口不存在"})
			return
		}
		c.File("./web/index.html")
	})

	// 启动服务
	log.Printf("JumpFrp Master starting on %s", cfg.Server.Addr)
	if err := r.Run(cfg.Server.Addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

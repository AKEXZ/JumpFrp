package main

import (
	"crypto/rand"
	"encoding/hex"
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

	// 修复可能缺失的列名（兼容旧数据库）
	fixMissingColumns(db)

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
		admin.RegisterAgentRoutes(api.Group("/agent"), db, sysSvc)
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
	
	// 为所有没有 APIToken 的用户生成 token
	ensureAllUsersHaveToken(db)
	
	if err := r.Run(cfg.Server.Addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

// fixMissingColumns 修复可能缺失的列（兼容旧数据库）
func fixMissingColumns(db *gorm.DB) {
	// Users 表需要添加的列
	userColumns := []struct {
		column  string
		colType string
	}{
		{"vip_level", "INTEGER DEFAULT 0"},
		{"vip_expire_at", "TEXT"},
		{"email_verified", "INTEGER DEFAULT 0"},
		{"api_token", "TEXT"},
		{"verify_code", "TEXT"},
		{"verify_expire", "TEXT"},
		{"reset_token", "TEXT"},
		{"reset_expire", "TEXT"},
	}
	for _, col := range userColumns {
		// 检查列是否存在
		var count int64
		db.Raw("SELECT COUNT(*) FROM pragma_table_info('users') WHERE name = ?", col.column).Scan(&count)
		if count == 0 {
			db.Exec("ALTER TABLE users ADD COLUMN " + col.column + " " + col.colType)
			log.Printf("添加缺失列: users.%s", col.column)
		}
	}

	// Nodes 表需要添加的列
	nodeColumns := []struct {
		column  string
		colType string
	}{
		{"bandwidth_limit", "INTEGER DEFAULT 0"},
		{"token", "TEXT"},
	}
	for _, col := range nodeColumns {
		var count int64
		db.Raw("SELECT COUNT(*) FROM pragma_table_info('nodes') WHERE name = ?", col.column).Scan(&count)
		if count == 0 {
			db.Exec("ALTER TABLE nodes ADD COLUMN " + col.column + " " + col.colType)
			log.Printf("添加缺失列: nodes.%s", col.column)
		}
	}

	// Tunnels 表需要添加的列
	tunnelColumns := []struct {
		column  string
		colType string
	}{
		{"bandwidth_limit", "INTEGER DEFAULT 0"},
		{"enabled", "INTEGER DEFAULT 1"},
	}
	for _, col := range tunnelColumns {
		var count int64
		db.Raw("SELECT COUNT(*) FROM pragma_table_info('tunnels') WHERE name = ?", col.column).Scan(&count)
		if count == 0 {
			db.Exec("ALTER TABLE tunnels ADD COLUMN " + col.column + " " + col.colType)
			log.Printf("添加缺失列: tunnels.%s", col.column)
		}
	}

	log.Println("数据库列修复完成")
}

// generateToken 生成随机 token
func generateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ensureAllUsersHaveToken 为所有没有 APIToken 的用户生成 token
func ensureAllUsersHaveToken(db *gorm.DB) {
	var users []model.User
	db.Where("api_token = '' OR api_token IS NULL").Find(&users)
	
	if len(users) == 0 {
		return
	}
	
	for _, user := range users {
		user.APIToken = generateToken(32)
		db.Model(&user).Update("api_token", user.APIToken)
		log.Printf("[Token] 为用户 %s 生成 APIToken", user.Username)
	}
	
	log.Printf("[Token] 共为 %d 个用户生成了 APIToken", len(users))
}


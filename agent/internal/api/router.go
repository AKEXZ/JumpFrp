package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jumpfrp/agent/internal/frps"
	"github.com/jumpfrp/agent/internal/monitor"
)

func NewRouter(token string, frpsMgr *frps.Manager, mon *monitor.Monitor) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Token 鉴权中间件
	r.Use(func(c *gin.Context) {
		if c.GetHeader("X-Agent-Token") != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "unauthorized"})
			return
		}
		c.Next()
	})

	// 状态查询（主控心跳检测用）
	r.GET("/status", func(c *gin.Context) {
		stats := mon.Collect()
		conns := 0
		if frpsMgr != nil {
			conns = frpsMgr.ConnectionCount()
		}
		c.JSON(http.StatusOK, gin.H{
			"status":        "ok",
			"cpu_usage":     stats.CPUUsage,
			"memory_usage":  stats.MemoryUsage,
			"current_conns": conns,
		})
	})

	// 重启 frps
	r.POST("/frps/restart", func(c *gin.Context) {
		if frpsMgr == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"msg": "frps 未初始化"})
			return
		}
		frpsMgr.Stop()
		if err := frpsMgr.Start(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"msg": "frps 已重启"})
	})

	return r
}

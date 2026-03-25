package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ── IP 限流 ──────────────────────────────────────────────
type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	// 定期清理过期记录
	go func() {
		for range time.Tick(time.Minute) {
			rl.cleanup()
		}
	}()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// 过滤过期记录
	reqs := rl.requests[ip]
	valid := reqs[:0]
	for _, t := range reqs {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	rl.requests[ip] = valid

	if len(valid) >= rl.limit {
		return false
	}
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-rl.window)
	for ip, reqs := range rl.requests {
		valid := reqs[:0]
		for _, t := range reqs {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = valid
		}
	}
}

// 全局限流器
var (
	globalLimiter = newRateLimiter(120, time.Minute)  // 全局：120次/分钟
	authLimiter   = newRateLimiter(10, time.Minute)   // 登录：10次/分钟
)

// RateLimit 全局限流中间件
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !globalLimiter.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": 429,
				"msg":  "请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}

// AuthRateLimit 登录/注册接口限流（防暴力破解）
func AuthRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !authLimiter.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": 429,
				"msg":  "操作过于频繁，请 1 分钟后再试",
			})
			return
		}
		c.Next()
	}
}

// ── 安全响应头 ────────────────────────────────────────────
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

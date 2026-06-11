package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

// rateLimitEntry 速率限制条目
type rateLimitEntry struct {
	count    int
	resetAt  time.Time
}

// ipRateLimiter 基于IP的速率限制器
type ipRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*rateLimitEntry
	limit    int           // 时间窗口内允许的最大请求数
	window   time.Duration // 时间窗口大小
}

var shareLimiter = &ipRateLimiter{
	visitors: make(map[string]*rateLimitEntry),
	limit:    20,              // 每个时间窗口最多20次请求
	window:   1 * time.Minute, // 1分钟时间窗口
}

// ShareRateLimit 分享接口的IP速率限制中间件
// 防止对 /share/info 和 /share/download 接口的暴力破解攻击
func ShareRateLimit() gin.HandlerFunc {
	// 启动后台清理协程，定期清除过期条目
	go shareLimiter.cleanup()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !shareLimiter.allow(ip) {
			c.JSON(http.StatusTooManyRequests, models.NewJsonResponse(429, fmt.Sprintf("请求过于频繁，请%d秒后再试", int(shareLimiter.window.Seconds())), nil))
			c.Abort()
			return
		}
		c.Next()
	}
}

// allow 检查指定IP是否允许访问
func (l *ipRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	entry, exists := l.visitors[ip]

	if !exists || now.After(entry.resetAt) {
		// 新IP或时间窗口已过期，重置计数
		l.visitors[ip] = &rateLimitEntry{
			count:   1,
			resetAt: now.Add(l.window),
		}
		return true
	}

	if entry.count >= l.limit {
		return false
	}

	entry.count++
	return true
}

// cleanup 定期清理过期的速率限制条目，防止内存泄漏
func (l *ipRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for ip, entry := range l.visitors {
			if now.After(entry.resetAt) {
				delete(l.visitors, ip)
			}
		}
		l.mu.Unlock()
	}
}

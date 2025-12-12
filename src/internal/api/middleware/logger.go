package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// GinLogger Gin 日志中间件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		go func() {
			since := time.Since(start)
			attrs := []slog.Attr{
				slog.String("query", query),
				slog.String("ip", c.ClientIP()),
			}
			mes := fmt.Sprintf("method: %s path:%s status:%v latency:%v", c.Request.Method, path, c.Writer.Status(), since)
			if len(c.Errors) > 0 {
				slog.ErrorContext(c.Request.Context(), c.Errors.String(), attrs)
			} else {
				slog.InfoContext(c.Request.Context(), mes, attrs)
			}
		}()
		c.Next()
	}
}

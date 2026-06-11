package middleware

import (
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

// sensitiveParamRe 匹配 query 中可能包含敏感信息的参数
var sensitiveParamRe = regexp.MustCompile(`(?i)((?:token|password|secret|authorization|api_key|apikey)=)[^&]*`)

// GinLogger Gin 日志中间件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		since := time.Since(start)
		// 脱敏处理 query 中的敏感参数
		sanitizedQuery := sensitiveParamRe.ReplaceAllString(query, "${1}****")
		attrs := []slog.Attr{
			slog.String("query", sanitizedQuery),
			slog.String("ip", c.ClientIP()),
		}
		mes := fmt.Sprintf("method: %s path:%s status:%v latency:%v", c.Request.Method, path, c.Writer.Status(), since)
		if len(c.Errors) > 0 {
			slog.ErrorContext(c.Request.Context(), c.Errors.String(), attrs)
		} else {
			slog.InfoContext(c.Request.Context(), mes, attrs)
		}
	}
}

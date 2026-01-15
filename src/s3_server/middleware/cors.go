package middleware

import (
	"encoding/json"
	"fmt"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/logger"
	"myobj/src/s3_server/types"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware S3 CORS中间件
// 根据Bucket的CORS配置自动添加CORS响应头
func CORSMiddleware(factory *impl.RepositoryFactory) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只处理S3请求（Bucket相关请求）
		bucketName := c.Param("bucket")
		if bucketName == "" {
			c.Next()
			return
		}

		// 获取用户ID（从上下文）
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.Next()
			return
		}

		// 获取Bucket的CORS配置
		corsRepo := factory.S3BucketCORS()
		cors, err := corsRepo.GetByBucket(c.Request.Context(), bucketName, userIDStr)
		if err != nil {
			// CORS配置不存在，跳过CORS处理
			c.Next()
			return
		}

		// 解析CORS配置
		var corsConfig types.CORSConfiguration
		if err := json.Unmarshal([]byte(cors.CORSConfig), &corsConfig); err != nil {
			logger.LOG.Warn("Failed to unmarshal CORS config", "error", err)
			c.Next()
			return
		}

		// 获取请求的Origin
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		// 查找匹配的CORS规则
		var matchedRule *types.CORSRule
		for i := range corsConfig.CORSRules {
			rule := &corsConfig.CORSRules[i]
			if matchOrigin(origin, rule.AllowedOrigins) {
				matchedRule = rule
				break
			}
		}

		if matchedRule == nil {
			// 没有匹配的规则，跳过CORS处理
			c.Next()
			return
		}

		// 处理OPTIONS预检请求
		if c.Request.Method == "OPTIONS" {
			requestMethod := c.GetHeader("Access-Control-Request-Method")
			requestHeaders := c.GetHeader("Access-Control-Request-Headers")

			// 检查请求方法是否允许
			if !matchMethod(requestMethod, matchedRule.AllowedMethods) {
				c.Status(http.StatusForbidden)
				c.Abort()
				return
			}

			// 检查请求头是否允许
			if requestHeaders != "" && len(matchedRule.AllowedHeaders) > 0 {
				headers := strings.Split(requestHeaders, ",")
				for _, header := range headers {
					header = strings.TrimSpace(header)
					if !matchHeader(header, matchedRule.AllowedHeaders) {
						c.Status(http.StatusForbidden)
						c.Abort()
						return
					}
				}
			}

			// 设置CORS响应头
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", strings.Join(matchedRule.AllowedMethods, ", "))
			if len(matchedRule.AllowedHeaders) > 0 {
				c.Header("Access-Control-Allow-Headers", strings.Join(matchedRule.AllowedHeaders, ", "))
			}
			if len(matchedRule.ExposeHeaders) > 0 {
				c.Header("Access-Control-Expose-Headers", strings.Join(matchedRule.ExposeHeaders, ", "))
			}
			if matchedRule.MaxAgeSeconds > 0 {
				c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", matchedRule.MaxAgeSeconds))
			} else {
				c.Header("Access-Control-Max-Age", "86400") // 默认24小时
			}
			c.Header("Access-Control-Allow-Credentials", "true")

			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}

		// 对于非OPTIONS请求，添加CORS响应头
		c.Header("Access-Control-Allow-Origin", origin)
		if len(matchedRule.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(matchedRule.ExposeHeaders, ", "))
		}
		c.Header("Access-Control-Allow-Credentials", "true")

		c.Next()
	}
}

// matchOrigin 匹配Origin
func matchOrigin(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
		// 支持通配符匹配（如：https://*.example.com）
		if strings.Contains(allowed, "*") {
			pattern := strings.ReplaceAll(allowed, ".", "\\.")
			pattern = strings.ReplaceAll(pattern, "*", ".*")
			matched, _ := regexp.MatchString("^"+pattern+"$", origin)
			if matched {
				return true
			}
		}
	}
	return false
}

// matchMethod 匹配HTTP方法
func matchMethod(method string, allowedMethods []string) bool {
	for _, allowed := range allowedMethods {
		if allowed == "*" || strings.EqualFold(allowed, method) {
			return true
		}
	}
	return false
}

// matchHeader 匹配请求头
func matchHeader(header string, allowedHeaders []string) bool {
	headerLower := strings.ToLower(header)
	for _, allowed := range allowedHeaders {
		allowedLower := strings.ToLower(allowed)
		if allowed == "*" || allowedLower == headerLower {
			return true
		}
	}
	return false
}

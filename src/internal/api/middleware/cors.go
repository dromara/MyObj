package middleware

import (
	"myobj/src/config"

	"github.com/gin-gonic/gin"
)

// CORS 跨域资源共享中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.CONFIG.Cors.Enable {
			c.Next()
			return
		} else {
			method := c.Request.Method
			origin := c.Request.Header.Get("Origin")
			// 允许所有来源,生产环境应该配置具体域名
			if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
			} else {
				c.Header("Access-Control-Allow-Origin", config.CONFIG.Cors.AllowOrigins)
			}

			// 允许的请求头
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Signature, X-Timestamp, X-Nonce, X-Requested-With"+","+config.CONFIG.Cors.AllowHeaders)

			// 允许的请求方法
			c.Header("Access-Control-Allow-Methods", config.CONFIG.Cors.AllowMethods)

			// 允许浏览器访问的响应头
			c.Header("Access-Control-Expose-Headers", config.CONFIG.Cors.ExposeHeaders)
			// 允许发送凭证(cookies)
			if config.CONFIG.Cors.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
			// 预检请求缓存时间(秒)
			c.Header("Access-Control-Max-Age", "86400")
			// 处理OPTIONS预检请求
			if method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
			c.Next()
		}
	}
}

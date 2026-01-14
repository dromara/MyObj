package middleware

import (
	"context"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/logger"
	"myobj/src/s3_server/auth"
	"myobj/src/s3_server/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// S3AuthMiddleware S3签名认证中间件
func S3AuthMiddleware(factory *impl.RepositoryFactory, region string) gin.HandlerFunc {
	verifier := auth.NewSignatureV4(region, "s3")

	return func(c *gin.Context) {
		// 生成请求ID
		requestID := uuid.NewString()
		c.Header("X-Amz-Request-Id", requestID)
		c.Set("request_id", requestID)

		// 1. 提取Access Key
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.LOG.Warn("S3 request missing Authorization header",
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
			)
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrAccessDenied, "")
			c.Abort()
			return
		}

		accessKeyID := auth.ExtractAccessKeyID(authHeader)
		if accessKeyID == "" {
			logger.LOG.Warn("S3 invalid Authorization header format",
				"auth_header", authHeader,
			)
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidAccessKeyId, "")
			c.Abort()
			return
		}

		// 2. 查询API Key和Secret Key
		ctx := context.Background()
		apiKey, err := factory.ApiKey().GetByKey(ctx, accessKeyID)
		if err != nil {
			logger.LOG.Warn("S3 API Key not found",
				"access_key", accessKeyID,
				"error", err,
			)
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidAccessKeyId, "")
			c.Abort()
			return
		}

		// 3. 验证签名
		if err := verifier.VerifyRequest(c.Request, apiKey.PrivateKey); err != nil {
			logger.LOG.Warn("S3 signature verification failed",
				"access_key", accessKeyID,
				"user_id", apiKey.UserID,
				"error", err,
			)
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrSignatureDoesNotMatch, "")
			c.Abort()
			return
		}

		// 4. 获取用户信息
		user, err := factory.User().GetByID(ctx, apiKey.UserID)
		if err != nil {
			logger.LOG.Warn("S3 user not found",
				"user_id", apiKey.UserID,
				"access_key", accessKeyID,
			)
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrAccessDenied, "")
			c.Abort()
			return
		}

		// 5. 检查用户状态（0正常 1禁用）
		if user.State != 0 {
			logger.LOG.Warn("S3 user is not active",
				"user_id", user.ID,
				"state", user.State,
			)
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrAccessDenied, "")
			c.Abort()
			return
		}

		// 6. 设置用户上下文
		c.Set("user_id", user.ID)
		c.Set("user", user)
		c.Set("api_key_id", apiKey.ID)
		c.Set("access_key", accessKeyID)

		logger.LOG.Info("S3 request authenticated",
			"user_id", user.ID,
			"username", user.UserName,
			"access_key", accessKeyID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"request_id", requestID,
		)

		c.Next()
	}
}

// S3LoggerMiddleware S3请求日志中间件
func S3LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求信息
		logger.LOG.Debug("S3 request received",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"remote_addr", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		c.Next()

		// 记录响应信息
		logger.LOG.Debug("S3 request completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"request_id", c.GetString("request_id"),
		)
	}
}

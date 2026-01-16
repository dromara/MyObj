package middleware

import (
	"context"
	"encoding/json"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/s3_server/auth"
	"myobj/src/s3_server/types"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// S3AuthMiddleware S3签名认证中间件
func S3AuthMiddleware(factory *impl.RepositoryFactory, region string) gin.HandlerFunc {
	verifier := auth.NewSignatureV4(region, "s3")

	return func(c *gin.Context) {
		// 生成请求ID
		requestID := uuid.NewString()
		c.Header("X-Amz-Request-Id", requestID)
		c.Set("request_id", requestID)

		ctx := context.Background()
		var accessKeyID string
		var apiKey *models.ApiKey
		var err error

		// 1. 检查是否是预签名URL请求
		query := c.Request.URL.Query()
		if query.Get("X-Amz-Signature") != "" {
			// 预签名URL验证
			credential := query.Get("X-Amz-Credential")
			if credential == "" {
				types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "missing X-Amz-Credential")
				c.Abort()
				return
			}

			// 提取Access Key ID（第一个斜杠前的部分）
			if idx := strings.Index(credential, "/"); idx > 0 {
				accessKeyID = credential[:idx]
			} else {
				types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "invalid X-Amz-Credential")
				c.Abort()
				return
			}

			// 查询API Key
			apiKey, err = factory.ApiKey().GetByKey(ctx, accessKeyID)
			if err != nil {
				logger.LOG.Warn("S3 API Key not found (presigned URL)",
					"access_key", accessKeyID,
					"error", err,
				)
				types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidAccessKeyId, "")
				c.Abort()
				return
			}

			// 验证预签名URL
			if err := verifier.VerifyPresignedURL(c.Request, apiKey.S3SecretKey); err != nil {
				logger.LOG.Warn("S3 presigned URL verification failed",
					"access_key", accessKeyID,
					"user_id", apiKey.UserID,
					"error", err,
				)
				types.WriteErrorResponse(c.Writer, c.Request, types.ErrSignatureDoesNotMatch, "")
				c.Abort()
				return
			}
		} else {
			// 常规Authorization头验证
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				// 对于 GET 请求，检查是否允许公开访问（基于 ACL）
				if c.Request.Method == "GET" || c.Request.Method == "HEAD" {
					if allowPublicAccess(ctx, factory, c.Request.URL.Path) {
						// 允许公开访问，设置匿名用户上下文
						c.Set("user_id", "")
						c.Set("is_public_access", true)
						logger.LOG.Debug("S3 public access allowed",
							"path", c.Request.URL.Path,
							"method", c.Request.Method,
						)
						c.Next()
						return
					}
				}

				logger.LOG.Warn("S3 request missing Authorization header",
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)
				types.WriteErrorResponse(c.Writer, c.Request, types.ErrAccessDenied, "")
				c.Abort()
				return
			}

			accessKeyID = auth.ExtractAccessKeyID(authHeader)
			if accessKeyID == "" {
				logger.LOG.Warn("S3 invalid Authorization header format",
					"auth_header", authHeader,
				)
				types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidAccessKeyId, "")
				c.Abort()
				return
			}

			// 查询API Key
			apiKey, err = factory.ApiKey().GetByKey(ctx, accessKeyID)
			if err != nil {
				logger.LOG.Warn("S3 API Key not found",
					"access_key", accessKeyID,
					"error", err,
				)
				types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidAccessKeyId, "")
				c.Abort()
				return
			}

			// 验证签名
			// 使用 S3 Secret Key（专门用于 S3 服务的 HMAC-SHA256 签名）
			if err := verifier.VerifyRequest(c.Request, apiKey.S3SecretKey); err != nil {
				// 记录详细的调试信息
				logger.LOG.Warn("S3 signature verification failed",
					"access_key", accessKeyID,
					"user_id", apiKey.UserID,
					"error", err,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"host", c.Request.Host,
					"authorization", c.Request.Header.Get("Authorization"),
					"x_amz_date", c.Request.Header.Get("X-Amz-Date"),
					"x_amz_content_sha256", c.Request.Header.Get("X-Amz-Content-Sha256"),
				)
				types.WriteErrorResponse(c.Writer, c.Request, types.ErrSignatureDoesNotMatch, "")
				c.Abort()
				return
			}
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

// allowPublicAccess 检查是否允许公开访问（基于 ACL）
func allowPublicAccess(ctx context.Context, factory *impl.RepositoryFactory, path string) bool {
	// 解析路径：/:bucket/*key
	// 例如：/my-obj/2026/01/15/file.jpg
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return false // 根路径需要认证
	}

	parts := strings.SplitN(path, "/", 2)
	bucketName := parts[0]
	objectKey := ""
	if len(parts) > 1 {
		objectKey = "/" + parts[1]
	}

	// 1. 查找 Bucket（通过数据库直接查询，不限制 userID）
	var bucket models.S3Bucket
	err := factory.DB().WithContext(ctx).Where("bucket_name = ?", bucketName).First(&bucket).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false
		}
		return false
	}

	// 2. 检查 Bucket ACL
	bucketACL, err := factory.S3ACL().GetBucketACL(ctx, bucketName, bucket.UserID)
	bucketHasPublicRead := false
	if err == nil && bucketACL != nil {
		var aclPolicy types.AccessControlPolicy
		if err := json.Unmarshal([]byte(bucketACL.ACLConfig), &aclPolicy); err == nil {
			if hasPublicReadAccess(&aclPolicy) {
				bucketHasPublicRead = true
				// 如果只是访问 Bucket（没有 Object Key），允许
				if objectKey == "" {
					return true
				}
			}
		}
	}

	// 3. 如果有 Object Key，检查 Object ACL
	if objectKey != "" {
		objectACL, err := factory.S3ACL().GetObjectACL(ctx, bucketName, objectKey, "", bucket.UserID)
		if err == nil && objectACL != nil {
			var aclPolicy types.AccessControlPolicy
			if err := json.Unmarshal([]byte(objectACL.ACLConfig), &aclPolicy); err == nil {
				if hasPublicReadAccess(&aclPolicy) {
					return true
				}
			}
		}
		// 如果没有 Object ACL，继承 Bucket ACL
		if bucketHasPublicRead {
			return true
		}
	}

	return false
}

// hasPublicReadAccess 检查 ACL 是否允许 AllUsers 读取
func hasPublicReadAccess(acl *types.AccessControlPolicy) bool {
	if acl == nil || acl.AccessControlList.Grants == nil {
		return false
	}

	allUsersURI := "http://acs.amazonaws.com/groups/global/AllUsers"

	for _, grant := range acl.AccessControlList.Grants {
		// 检查是否是 AllUsers
		if grant.Grantee.URI == allUsersURI || grant.Grantee.Type == "Group" {
			// 检查权限：READ 或 FULL_CONTROL
			if grant.Permission == "READ" || grant.Permission == "FULL_CONTROL" {
				return true
			}
		}
	}

	return false
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

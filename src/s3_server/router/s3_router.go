package router

import (
	"myobj/src/config"
	coreService "myobj/src/core/service"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/logger"
	"myobj/src/s3_server/handler"
	"myobj/src/s3_server/middleware"

	"github.com/gin-gonic/gin"
)

// SetupS3Router 配置S3路由
func SetupS3Router(router *gin.Engine, factory *impl.RepositoryFactory, fileService *coreService.FileService) {
	// 获取S3配置
	region := "us-east-1" // 默认区域
	if config.CONFIG.S3.Region != "" {
		region = config.CONFIG.S3.Region
	}

	// 获取路径前缀（仅在共用端口模式下生效）
	pathPrefix := "/"
	if config.CONFIG.S3.SharePort && config.CONFIG.S3.PathPrefix != "" {
		pathPrefix = config.CONFIG.S3.PathPrefix
		logger.LOG.Warn("S3 路由使用路径前缀（不推荐）",
			"path_prefix", pathPrefix,
			"warning", "使用路径前缀会导致与标准S3客户端SDK不兼容，建议使用独立端口",
		)
	}

	logger.LOG.Info("S3 路由配置",
		"path_prefix", pathPrefix,
		"region", region,
		"share_port", config.CONFIG.S3.SharePort,
	)

	// 创建S3处理器
	s3Handler := handler.NewS3Handler(factory, fileService)

	// S3 API路由组 - 使用配置的路径前缀
	s3Group := router.Group(pathPrefix)

	// 应用S3中间件
	s3Group.Use(middleware.S3LoggerMiddleware())
	s3Group.Use(middleware.S3AuthMiddleware(factory, region))
	// CORS中间件需要在认证之后，因为需要获取user_id
	s3Group.Use(middleware.CORSMiddleware(factory))

	// Service API - 列出所有Bucket
	s3Group.GET("/", s3Handler.ListBuckets)

	// Bucket操作
	s3Group.PUT("/:bucket", handleBucketRequest(s3Handler))
	s3Group.HEAD("/:bucket", handleBucketRequest(s3Handler))
	s3Group.DELETE("/:bucket", handleBucketRequest(s3Handler))
	s3Group.GET("/:bucket", handleBucketRequest(s3Handler))

	// Object操作
	s3Group.PUT("/:bucket/*key", handleObjectRequest(s3Handler))
	s3Group.GET("/:bucket/*key", handleObjectRequest(s3Handler))
	s3Group.HEAD("/:bucket/*key", handleObjectRequest(s3Handler))
	s3Group.DELETE("/:bucket/*key", handleObjectRequest(s3Handler))

	logger.LOG.Info("S3 routes configured successfully", "path_prefix", pathPrefix)
}

// handleBucketRequest 处理Bucket级别的请求
func handleBucketRequest(h *handler.S3Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否是 ListMultipartUploads 操作
		_, hasUploads := c.GetQuery("uploads")
		_, hasDelete := c.GetQuery("delete")
		_, hasVersioning := c.GetQuery("versioning")
		_, hasVersions := c.GetQuery("versions")
		_, hasCORS := c.GetQuery("cors")
		_, hasACL := c.GetQuery("acl")
		_, hasPolicy := c.GetQuery("policy")
		_, hasLifecycle := c.GetQuery("lifecycle")

		if hasLifecycle {
			// PUT /:bucket?lifecycle - 设置Bucket Lifecycle
			// GET /:bucket?lifecycle - 获取Bucket Lifecycle
			// DELETE /:bucket?lifecycle - 删除Bucket Lifecycle
			if c.Request.Method == "PUT" {
				h.PutBucketLifecycle(c)
			} else if c.Request.Method == "GET" {
				h.GetBucketLifecycle(c)
			} else if c.Request.Method == "DELETE" {
				h.DeleteBucketLifecycle(c)
			} else {
				c.Status(405) // Method Not Allowed
			}
			return
		}

		if hasPolicy {
			// PUT /:bucket?policy - 设置Bucket Policy
			// GET /:bucket?policy - 获取Bucket Policy
			// DELETE /:bucket?policy - 删除Bucket Policy
			if c.Request.Method == "PUT" {
				h.PutBucketPolicy(c)
			} else if c.Request.Method == "GET" {
				h.GetBucketPolicy(c)
			} else if c.Request.Method == "DELETE" {
				h.DeleteBucketPolicy(c)
			} else {
				c.Status(405) // Method Not Allowed
			}
			return
		}

		if hasACL {
			// PUT /:bucket?acl - 设置Bucket ACL
			// GET /:bucket?acl - 获取Bucket ACL
			if c.Request.Method == "PUT" {
				h.PutBucketACL(c)
			} else if c.Request.Method == "GET" {
				h.GetBucketACL(c)
			} else {
				c.Status(405) // Method Not Allowed
			}
			return
		}

		if hasCORS {
			// PUT /:bucket?cors - 设置CORS配置
			// GET /:bucket?cors - 获取CORS配置
			// DELETE /:bucket?cors - 删除CORS配置
			if c.Request.Method == "PUT" {
				h.PutBucketCORS(c)
			} else if c.Request.Method == "GET" {
				h.GetBucketCORS(c)
			} else if c.Request.Method == "DELETE" {
				h.DeleteBucketCORS(c)
			} else {
				c.Status(405) // Method Not Allowed
			}
			return
		}

		if hasVersions && c.Request.Method == "GET" {
			// GET /:bucket?versions - 列出对象版本
			h.ListObjectVersions(c)
			return
		}

		if hasVersioning {
			// PUT /:bucket?versioning - 设置版本控制
			// GET /:bucket?versioning - 获取版本控制状态
			if c.Request.Method == "PUT" {
				h.PutBucketVersioning(c)
			} else if c.Request.Method == "GET" {
				h.GetBucketVersioning(c)
			} else {
				c.Status(405) // Method Not Allowed
			}
			return
		}

		if hasUploads && c.Request.Method == "GET" {
			// GET /:bucket?uploads - 列出分片上传会话
			h.ListMultipartUploads(c)
			return
		}

		if hasDelete && c.Request.Method == "POST" {
			// POST /:bucket?delete - 批量删除对象
			h.DeleteObjects(c)
			return
		}

		// 根据查询参数判断操作类型
		listType := c.Query("list-type")

		switch c.Request.Method {
		case "PUT":
			h.CreateBucket(c)
		case "HEAD":
			h.HeadBucket(c)
		case "DELETE":
			h.DeleteBucket(c)
		case "GET":
			// 根据list-type判断使用V1还是V2
			if listType == "2" {
				h.ListObjectsV2(c)
			} else {
				h.ListObjects(c)
			}
		default:
			c.Status(405) // Method Not Allowed
		}
	}
}

// handleObjectRequest 处理Object级别的请求
func handleObjectRequest(h *handler.S3Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否是Multipart Upload操作
		_, hasUploadID := c.GetQuery("uploadId")
		_, hasUploads := c.GetQuery("uploads")
		_, hasTagging := c.GetQuery("tagging")
		_, hasACL := c.GetQuery("acl")

		// Object ACL操作
		if hasACL {
			// PUT /:bucket/:key?acl - 设置对象ACL
			// GET /:bucket/:key?acl - 获取对象ACL
			if c.Request.Method == "PUT" {
				h.PutObjectACL(c)
			} else if c.Request.Method == "GET" {
				h.GetObjectACL(c)
			} else {
				c.Status(405) // Method Not Allowed
			}
			return
		}

		// Object Tagging操作
		if hasTagging {
			// PUT /:bucket/:key?tagging - 设置对象标签
			// GET /:bucket/:key?tagging - 获取对象标签
			// DELETE /:bucket/:key?tagging - 删除对象标签
			if c.Request.Method == "PUT" {
				h.PutObjectTagging(c)
			} else if c.Request.Method == "GET" {
				h.GetObjectTagging(c)
			} else if c.Request.Method == "DELETE" {
				h.DeleteObjectTagging(c)
			} else {
				c.Status(405) // Method Not Allowed
			}
			return
		}

		// Multipart Upload操作
		if hasUploadID {
			// 有 uploadId，是分片上传相关操作
			switch c.Request.Method {
			case "PUT":
				// PUT /:bucket/:key?uploadId=xxx&partNumber=xxx - 上传分片
				h.UploadPart(c)
			case "POST":
				// POST /:bucket/:key?uploadId=xxx - 完成分片上传
				h.CompleteMultipartUpload(c)
			case "DELETE":
				// DELETE /:bucket/:key?uploadId=xxx - 取消分片上传
				h.AbortMultipartUpload(c)
			case "GET":
				// GET /:bucket/:key?uploadId=xxx - 列出分片
				h.ListParts(c)
			default:
				c.Status(405) // Method Not Allowed
			}
			return
		}

		if hasUploads {
			// POST /:bucket/:key?uploads - 初始化分片上传
			// 注意：S3规范中，初始化分片上传使用 POST /:bucket/:key?uploads
			if c.Request.Method == "POST" {
				h.InitiateMultipartUpload(c)
			} else {
				c.Status(405) // Method Not Allowed
			}
			return
		}

		// 检查是否是生成预签名URL请求
		_, hasPresign := c.GetQuery("presign")
		if hasPresign && c.Request.Method == "GET" {
			// GET /:bucket/:key?presign=true&expires=3600&method=GET - 生成预签名URL
			h.GeneratePresignedURL(c)
			return
		}

		// 标准Object操作
		switch c.Request.Method {
		case "PUT":
			// 检查是否是 CopyObject（有 x-amz-copy-source 头）
			if c.GetHeader("x-amz-copy-source") != "" {
				h.CopyObject(c)
			} else {
				h.PutObject(c)
			}
		case "GET":
			h.GetObject(c)
		case "HEAD":
			h.HeadObject(c)
		case "DELETE":
			h.DeleteObject(c)
		default:
			c.Status(405) // Method Not Allowed
		}
	}
}

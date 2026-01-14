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

	// 创建S3处理器
	s3Handler := handler.NewS3Handler(factory, fileService)

	// S3 API路由组
	s3Group := router.Group("/")

	// 应用S3中间件
	s3Group.Use(middleware.S3LoggerMiddleware())
	s3Group.Use(middleware.S3AuthMiddleware(factory, region))

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

	logger.LOG.Info("S3 routes configured successfully")
}

// handleBucketRequest 处理Bucket级别的请求
func handleBucketRequest(h *handler.S3Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		// TODO: 实现Multipart Upload路由判断
		if hasUploadID || hasUploads {
			// Multipart Upload操作
			// 暂时返回未实现
			c.Status(501) // Not Implemented
			return
		}

		// 标准Object操作
		switch c.Request.Method {
		case "PUT":
			h.PutObject(c)
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

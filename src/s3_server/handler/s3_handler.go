package handler

import (
	"encoding/xml"
	"fmt"
	"io"
	coreService "myobj/src/core/service"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/logger"
	"myobj/src/s3_server/service"
	"myobj/src/s3_server/types"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// S3Handler S3 API处理器
type S3Handler struct {
	bucketService *service.S3BucketService
	objectService *service.S3ObjectService
	factory       *impl.RepositoryFactory
}

// NewS3Handler 创建S3处理器
func NewS3Handler(factory *impl.RepositoryFactory, fileService *coreService.FileService) *S3Handler {
	return &S3Handler{
		bucketService: service.NewS3BucketService(factory),
		objectService: service.NewS3ObjectService(factory, fileService),
		factory:       factory,
	}
}

// ListBuckets 列出所有Bucket
// GET /
func (h *S3Handler) ListBuckets(c *gin.Context) {
	userID := c.GetString("user_id")

	buckets, err := h.bucketService.ListBuckets(c.Request.Context(), userID)
	if err != nil {
		logger.LOG.Error("List buckets failed",
			"user_id", userID,
			"error", err,
		)
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, "")
		return
	}

	// 构建XML响应
	response := types.ListAllMyBucketsResult{
		Owner: types.Owner{
			ID:          userID,
			DisplayName: userID,
		},
		Buckets: types.Buckets{
			Bucket: make([]types.Bucket, 0, len(buckets)),
		},
	}

	for _, b := range buckets {
		response.Buckets.Bucket = append(response.Buckets.Bucket, types.Bucket{
			Name:         b.BucketName,
			CreationDate: b.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		})
	}

	// 写入XML响应
	c.Header("Content-Type", "application/xml")
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(response, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("List buckets success",
		"user_id", userID,
		"bucket_count", len(buckets),
	)
}

// CreateBucket 创建Bucket
// PUT /:bucket
func (h *S3Handler) CreateBucket(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 获取区域（从请求头或使用默认值）
	region := c.GetHeader("x-amz-bucket-region")
	if region == "" {
		region = "us-east-1"
	}

	// 创建Bucket
	err := h.bucketService.CreateBucket(c.Request.Context(), bucketName, userID, region)
	if err != nil {
		logger.LOG.Error("Create bucket failed",
			"bucket_name", bucketName,
			"user_id", userID,
			"error", err,
		)

		// 根据错误类型返回不同的错误码
		if err.Error() == "bucket already exists" {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrBucketAlreadyExists, bucketName)
		} else if err.Error() == "bucket name must be between 3 and 63 characters long" ||
			err.Error() == "bucket name must consist of lowercase letters, numbers, dots and hyphens" ||
			err.Error() == "bucket name cannot contain consecutive dots" ||
			err.Error() == "bucket name cannot be formatted as IP address" {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidBucketName, bucketName)
		} else {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, bucketName)
		}
		return
	}

	// 返回成功
	c.Header("Location", "/"+bucketName)
	c.Status(http.StatusOK)

	logger.LOG.Info("Create bucket success",
		"bucket_name", bucketName,
		"user_id", userID,
		"region", region,
	)
}

// HeadBucket 检查Bucket是否存在
// HEAD /:bucket
func (h *S3Handler) HeadBucket(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	exists, err := h.bucketService.HeadBucket(c.Request.Context(), bucketName, userID)
	if err != nil {
		logger.LOG.Error("Head bucket failed",
			"bucket_name", bucketName,
			"user_id", userID,
			"error", err,
		)
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, bucketName)
		return
	}

	if !exists {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		return
	}

	// Bucket存在，返回200
	c.Status(http.StatusOK)

	logger.LOG.Debug("Head bucket success",
		"bucket_name", bucketName,
		"user_id", userID,
	)
}

// DeleteBucket 删除Bucket
// DELETE /:bucket
func (h *S3Handler) DeleteBucket(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	err := h.bucketService.DeleteBucket(c.Request.Context(), bucketName, userID)
	if err != nil {
		logger.LOG.Error("Delete bucket failed",
			"bucket_name", bucketName,
			"user_id", userID,
			"error", err,
		)

		// 根据错误类型返回不同的错误码
		if err.Error() == "bucket not found" {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else if err.Error() == "bucket is not empty" {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrBucketNotEmpty, bucketName)
		} else {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, bucketName)
		}
		return
	}

	// 返回成功（204 No Content）
	c.Status(http.StatusNoContent)

	logger.LOG.Info("Delete bucket success",
		"bucket_name", bucketName,
		"user_id", userID,
	)
}

// ListObjects 列出Bucket中的对象
// GET /:bucket
func (h *S3Handler) ListObjects(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 解析查询参数
	prefix := c.Query("prefix")
	delimiter := c.Query("delimiter")
	marker := c.Query("marker")
	maxKeysStr := c.DefaultQuery("max-keys", "1000")

	// TODO: 实现ListObjects逻辑
	// 当前返回空列表
	response := types.ListBucketResult{
		Name:        bucketName,
		Prefix:      prefix,
		Marker:      marker,
		MaxKeys:     1000,
		Delimiter:   delimiter,
		IsTruncated: false,
		Contents:    []types.Contents{},
	}

	// 写入XML响应
	c.Header("Content-Type", "application/xml")
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(response, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("List objects success",
		"bucket_name", bucketName,
		"user_id", userID,
		"prefix", prefix,
		"max_keys", maxKeysStr,
	)
}

// ListObjectsV2 列出Bucket中的对象（V2版本）
// GET /:bucket?list-type=2
func (h *S3Handler) ListObjectsV2(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 解析查询参数
	prefix := c.Query("prefix")
	delimiter := c.Query("delimiter")
	continuationToken := c.Query("continuation-token")
	maxKeysStr := c.DefaultQuery("max-keys", "1000")

	// TODO: 实现ListObjectsV2逻辑
	// 当前返回空列表
	response := types.ListBucketResultV2{
		Name:              bucketName,
		Prefix:            prefix,
		KeyCount:          0,
		MaxKeys:           1000,
		Delimiter:         delimiter,
		IsTruncated:       false,
		ContinuationToken: continuationToken,
		Contents:          []types.Contents{},
	}

	// 写入XML响应
	c.Header("Content-Type", "application/xml")
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(response, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("List objects v2 success",
		"bucket_name", bucketName,
		"user_id", userID,
		"prefix", prefix,
		"max_keys", maxKeysStr,
	)
}

// PutObject 上传对象
// PUT /:bucket/:key+
func (h *S3Handler) PutObject(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")

	// 解析请求头
	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	contentMD5 := c.GetHeader("Content-MD5")
	storageClass := c.GetHeader("x-amz-storage-class")

	// 解析用户元数据
	userMetadata := make(map[string]string)
	for key, values := range c.Request.Header {
		if strings.HasPrefix(strings.ToLower(key), "x-amz-meta-") {
			metaKey := key[len("x-amz-meta-"):]
			if len(values) > 0 {
				userMetadata[metaKey] = values[0]
			}
		}
	}

	// 调用服务层
	output, err := h.objectService.PutObject(c.Request.Context(), &service.PutObjectInput{
		BucketName:   bucketName,
		ObjectKey:    objectKey,
		UserID:       userID,
		Body:         c.Request.Body,
		ContentType:  contentType,
		ContentMD5:   contentMD5,
		UserMetadata: userMetadata,
		StorageClass: storageClass,
	})

	if err != nil {
		logger.LOG.Error("Put object failed",
			"bucket", bucketName,
			"key", objectKey,
			"error", err,
		)

		if strings.Contains(err.Error(), "bucket not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else if strings.Contains(err.Error(), "MD5 mismatch") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "Content-MD5 mismatch")
		} else if strings.Contains(err.Error(), "insufficient user space") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrEntityTooLarge, "Insufficient storage space")
		} else {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, "")
		}
		return
	}

	// 返回ETag
	c.Header("ETag", fmt.Sprintf("\"%s\"", output.ETag))
	c.Header("x-amz-version-id", output.VersionID)
	c.Status(http.StatusOK)

	logger.LOG.Info("Put object success",
		"bucket", bucketName,
		"key", objectKey,
		"etag", output.ETag,
	)
}

// GetObject 下载对象
// GET /:bucket/:key+
func (h *S3Handler) GetObject(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")

	// 解析Range请求
	var rangeStart, rangeEnd int64
	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		// 解析 "bytes=0-1023" 格式
		if strings.HasPrefix(rangeHeader, "bytes=") {
			rangeStr := rangeHeader[6:]
			parts := strings.Split(rangeStr, "-")
			if len(parts) == 2 {
				if parts[0] != "" {
					rangeStart, _ = strconv.ParseInt(parts[0], 10, 64)
				}
				if parts[1] != "" {
					rangeEnd, _ = strconv.ParseInt(parts[1], 10, 64)
				}
			}
		}
	}

	// 调用服务层
	output, err := h.objectService.GetObject(c.Request.Context(), &service.GetObjectInput{
		BucketName: bucketName,
		ObjectKey:  objectKey,
		UserID:     userID,
		RangeStart: rangeStart,
		RangeEnd:   rangeEnd,
	})

	if err != nil {
		logger.LOG.Error("Get object failed",
			"bucket", bucketName,
			"key", objectKey,
			"error", err,
		)

		if strings.Contains(err.Error(), "not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchKey, bucketName+"/"+objectKey)
		} else {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, "")
		}
		return
	}
	defer output.Body.Close()

	// 设置响应头
	c.Header("Content-Type", output.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", output.ContentLength))
	c.Header("ETag", fmt.Sprintf("\"%s\"", output.ETag))
	c.Header("Last-Modified", output.LastModified.Format(http.TimeFormat))
	c.Header("x-amz-version-id", output.VersionID)

	// 设置用户元数据
	for key, value := range output.UserMetadata {
		c.Header("x-amz-meta-"+key, value)
	}

	// 如果是Range请求，返回206
	if rangeHeader != "" {
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d",
			output.ActualStart, output.ActualEnd, output.ContentLength))
		c.Status(http.StatusPartialContent)
	} else {
		c.Status(http.StatusOK)
	}

	// 流式输出文件内容
	io.Copy(c.Writer, output.Body)

	logger.LOG.Info("Get object success",
		"bucket", bucketName,
		"key", objectKey,
		"size", output.ContentLength,
	)
}

// HeadObject 获取对象元数据
// HEAD /:bucket/:key+
func (h *S3Handler) HeadObject(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")

	// 调用服务层
	output, err := h.objectService.HeadObject(c.Request.Context(), bucketName, objectKey, userID)

	if err != nil {
		logger.LOG.Error("Head object failed",
			"bucket", bucketName,
			"key", objectKey,
			"error", err,
		)

		if strings.Contains(err.Error(), "not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchKey, bucketName+"/"+objectKey)
		} else {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, "")
		}
		return
	}

	// 设置响应头
	c.Header("Content-Type", output.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", output.ContentLength))
	c.Header("ETag", fmt.Sprintf("\"%s\"", output.ETag))
	c.Header("Last-Modified", output.LastModified.Format(http.TimeFormat))
	c.Header("x-amz-version-id", output.VersionID)

	// 设置用户元数据
	for key, value := range output.UserMetadata {
		c.Header("x-amz-meta-"+key, value)
	}

	c.Status(http.StatusOK)

	logger.LOG.Info("Head object success",
		"bucket", bucketName,
		"key", objectKey,
	)
}

// DeleteObject 删除对象
// DELETE /:bucket/:key+
func (h *S3Handler) DeleteObject(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")

	// 调用服务层
	err := h.objectService.DeleteObject(c.Request.Context(), bucketName, objectKey, userID)

	if err != nil {
		logger.LOG.Error("Delete object failed",
			"bucket", bucketName,
			"key", objectKey,
			"error", err,
		)
		// S3规范：即使删除失败，也返回204
	}

	// S3规范：即使对象不存在，删除也返回204
	c.Status(http.StatusNoContent)

	logger.LOG.Info("Delete object success",
		"bucket", bucketName,
		"key", objectKey,
	)
}

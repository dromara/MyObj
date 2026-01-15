package handler

import (
	"encoding/xml"
	"fmt"
	"io"
	coreService "myobj/src/core/service"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/s3_server/service"
	"myobj/src/s3_server/types"
	"net/http"
	"strconv"
	"strings"
	"time"

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
		s3Err := types.MapErrorToS3Error(err)
		types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
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
		} else {
			s3Err := types.MapErrorToS3Error(err)
			types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
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

// PutBucketVersioning 设置Bucket版本控制
// PUT /:bucket?versioning
func (h *S3Handler) PutBucketVersioning(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 解析XML请求体
	var versioningConfig types.VersioningConfiguration
	if err := xml.NewDecoder(c.Request.Body).Decode(&versioningConfig); err != nil {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "invalid XML")
		return
	}

	// 验证状态值
	status := "Disabled"
	if versioningConfig.Status == "Enabled" {
		status = "Enabled"
	} else if versioningConfig.Status == "Suspended" {
		status = "Suspended"
	}

	// 调用服务层
	err := h.bucketService.PutBucketVersioning(c.Request.Context(), &service.PutBucketVersioningInput{
		BucketName: bucketName,
		UserID:     userID,
		Status:     status,
	})

	if err != nil {
		logger.LOG.Error("Put bucket versioning failed",
			"bucket", bucketName,
			"user_id", userID,
			"status", status,
			"error", err,
		)

		s3Err := types.MapErrorToS3Error(err)
		types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
		return
	}

	// 返回204 No Content
	c.Status(http.StatusNoContent)

	logger.LOG.Info("Put bucket versioning success",
		"bucket", bucketName,
		"user_id", userID,
		"status", status,
	)
}

// GetBucketVersioning 获取Bucket版本控制状态
// GET /:bucket?versioning
func (h *S3Handler) GetBucketVersioning(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 调用服务层
	output, err := h.bucketService.GetBucketVersioning(c.Request.Context(), &service.GetBucketVersioningInput{
		BucketName: bucketName,
		UserID:     userID,
	})

	if err != nil {
		logger.LOG.Error("Get bucket versioning failed",
			"bucket", bucketName,
			"user_id", userID,
			"error", err,
		)

		s3Err := types.MapErrorToS3Error(err)
		types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
		return
	}

	// 构建XML响应
	response := types.VersioningConfiguration{
		XMLName: xml.Name{Local: "VersioningConfiguration"},
		Status:  output.Status,
	}

	c.Header("Content-Type", "application/xml")
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(response, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("Get bucket versioning success",
		"bucket", bucketName,
		"user_id", userID,
		"status", output.Status,
	)
}

// PutBucketCORS 设置Bucket CORS配置
// PUT /:bucket?cors
func (h *S3Handler) PutBucketCORS(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 解析XML请求体
	var corsConfig types.CORSConfiguration
	if err := xml.NewDecoder(c.Request.Body).Decode(&corsConfig); err != nil {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "invalid XML")
		return
	}

	// 调用服务层
	err := h.bucketService.PutBucketCORS(c.Request.Context(), &service.PutBucketCORSInput{
		BucketName: bucketName,
		UserID:     userID,
		CORSConfig: &corsConfig,
	})

	if err != nil {
		logger.LOG.Error("Put bucket CORS failed",
			"bucket", bucketName,
			"user_id", userID,
			"error", err,
		)

		if strings.Contains(err.Error(), "bucket not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else {
			s3Err := types.MapErrorToS3Error(err)
			types.WriteErrorResponse(c.Writer, c.Request, s3Err, "")
		}
		return
	}

	// 返回204 No Content
	c.Status(http.StatusNoContent)

	logger.LOG.Info("Put bucket CORS success",
		"bucket", bucketName,
		"user_id", userID,
		"rules_count", len(corsConfig.CORSRules),
	)
}

// GetBucketCORS 获取Bucket CORS配置
// GET /:bucket?cors
func (h *S3Handler) GetBucketCORS(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 调用服务层
	output, err := h.bucketService.GetBucketCORS(c.Request.Context(), &service.GetBucketCORSInput{
		BucketName: bucketName,
		UserID:     userID,
	})

	if err != nil {
		logger.LOG.Error("Get bucket CORS failed",
			"bucket", bucketName,
			"user_id", userID,
			"error", err,
		)

		if strings.Contains(err.Error(), "bucket not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else {
			s3Err := types.MapErrorToS3Error(err)
			types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
		}
		return
	}

	// 构建XML响应
	c.Header("Content-Type", "application/xml")
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(output.CORSConfig, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("Get bucket CORS success",
		"bucket", bucketName,
		"user_id", userID,
		"rules_count", len(output.CORSConfig.CORSRules),
	)
}

// DeleteBucketCORS 删除Bucket CORS配置
// DELETE /:bucket?cors
func (h *S3Handler) DeleteBucketCORS(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 调用服务层
	err := h.bucketService.DeleteBucketCORS(c.Request.Context(), &service.DeleteBucketCORSInput{
		BucketName: bucketName,
		UserID:     userID,
	})

	if err != nil {
		logger.LOG.Error("Delete bucket CORS failed",
			"bucket", bucketName,
			"user_id", userID,
			"error", err,
		)

		s3Err := types.MapErrorToS3Error(err)
		types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
		return
	}

	// 返回204 No Content
	c.Status(http.StatusNoContent)

	logger.LOG.Info("Delete bucket CORS success",
		"bucket", bucketName,
		"user_id", userID,
	)
}

// ListObjectVersions 列出对象的所有版本
// GET /:bucket?versions
func (h *S3Handler) ListObjectVersions(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 解析查询参数
	prefix := c.Query("prefix")
	delimiter := c.Query("delimiter")
	keyMarker := c.Query("key-marker")
	versionIDMarker := c.Query("version-id-marker")
	maxKeysStr := c.Query("max-keys")

	maxKeys := 1000
	if maxKeysStr != "" {
		if parsed, err := strconv.Atoi(maxKeysStr); err == nil && parsed > 0 && parsed <= 1000 {
			maxKeys = parsed
		}
	}

	// 调用服务层
	output, err := h.objectService.ListObjectVersions(c.Request.Context(), &service.ListObjectVersionsInput{
		BucketName:      bucketName,
		UserID:          userID,
		Prefix:          prefix,
		Delimiter:       delimiter,
		KeyMarker:       keyMarker,
		VersionIDMarker: versionIDMarker,
		MaxKeys:         maxKeys,
	})

	if err != nil {
		logger.LOG.Error("List object versions failed",
			"bucket", bucketName,
			"user_id", userID,
			"error", err,
		)

		s3Err := types.MapErrorToS3Error(err)
		types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
		return
	}

	// 构建XML响应
	response := types.ListVersionsResult{
		XMLName:             xml.Name{Local: "ListVersionsResult"},
		Name:                bucketName,
		Prefix:              prefix,
		KeyMarker:           keyMarker,
		VersionIDMarker:     versionIDMarker,
		MaxKeys:             maxKeys,
		Delimiter:           delimiter,
		IsTruncated:         output.IsTruncated,
		NextKeyMarker:       output.NextKeyMarker,
		NextVersionIDMarker: output.NextVersionIDMarker,
		Versions:            make([]types.Version, 0, len(output.Versions)),
		DeleteMarkers:       make([]types.DeleteMarkerEntry, 0, len(output.DeleteMarkers)),
		CommonPrefixes:      make([]types.CommonPrefix, 0, len(output.CommonPrefixes)),
	}

	// 转换版本信息
	for _, v := range output.Versions {
		response.Versions = append(response.Versions, types.Version{
			Key:          v.Key,
			VersionID:    v.VersionID,
			IsLatest:     v.IsLatest,
			LastModified: v.LastModified.Format(time.RFC3339),
			ETag:         v.ETag,
			Size:         v.Size,
			StorageClass: v.StorageClass,
			Owner: types.Owner{
				ID:          v.Owner,
				DisplayName: v.Owner,
			},
		})
	}

	// 转换删除标记
	for _, dm := range output.DeleteMarkers {
		response.DeleteMarkers = append(response.DeleteMarkers, types.DeleteMarkerEntry{
			Key:          dm.Key,
			VersionID:    dm.VersionID,
			IsLatest:     dm.IsLatest,
			LastModified: dm.LastModified.Format(time.RFC3339),
			Owner: types.Owner{
				ID:          dm.Owner,
				DisplayName: dm.Owner,
			},
		})
	}

	// 转换公共前缀
	for _, cp := range output.CommonPrefixes {
		response.CommonPrefixes = append(response.CommonPrefixes, types.CommonPrefix{
			Prefix: cp,
		})
	}

	c.Header("Content-Type", "application/xml")
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(response, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("List object versions success",
		"bucket", bucketName,
		"user_id", userID,
		"versions_count", len(output.Versions),
		"delete_markers_count", len(output.DeleteMarkers),
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
	maxKeys, _ := strconv.Atoi(maxKeysStr)
	if maxKeys <= 0 {
		maxKeys = 1000
	}

	// 调用服务层
	output, err := h.objectService.ListObjects(c.Request.Context(), &service.ListObjectsInput{
		BucketName: bucketName,
		UserID:     userID,
		Prefix:     prefix,
		Delimiter:  delimiter,
		Marker:     marker,
		MaxKeys:    maxKeys,
	})

	if err != nil {
		logger.LOG.Error("List objects failed",
			"bucket_name", bucketName,
			"user_id", userID,
			"error", err,
		)

		s3Err := types.MapErrorToS3Error(err)
		types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
		return
	}

	// 构建响应
	response := types.ListBucketResult{
		Name:           bucketName,
		Prefix:         prefix,
		Marker:         marker,
		MaxKeys:        maxKeys,
		Delimiter:      delimiter,
		IsTruncated:    output.IsTruncated,
		Contents:       make([]types.Contents, 0, len(output.Objects)),
		CommonPrefixes: make([]types.CommonPrefix, 0, len(output.CommonPrefixes)),
	}

	// 获取文件信息以填充 Size 和 LastModified
	for _, obj := range output.Objects {
		fileInfo, err := h.factory.FileInfo().GetByID(c.Request.Context(), obj.FileID)
		if err != nil {
			logger.LOG.Warn("Get file info failed", "file_id", obj.FileID, "error", err)
			continue
		}

		contents := types.Contents{
			Key:          obj.ObjectKey,
			LastModified: fileInfo.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
			ETag:         fmt.Sprintf("\"%s\"", obj.ETag),
			Size:         int64(fileInfo.Size),
			StorageClass: obj.StorageClass,
		}
		response.Contents = append(response.Contents, contents)
	}

	// 添加公共前缀
	for _, commonPrefix := range output.CommonPrefixes {
		response.CommonPrefixes = append(response.CommonPrefixes, types.CommonPrefix{
			Prefix: commonPrefix,
		})
	}

	// 设置 NextMarker
	if output.IsTruncated && output.NextMarker != "" {
		response.NextMarker = output.NextMarker
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
		"max_keys", maxKeys,
		"object_count", len(output.Objects),
		"common_prefix_count", len(output.CommonPrefixes),
		"is_truncated", output.IsTruncated,
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
	continuationToken := c.Query("continuation-token") // V2 使用 continuation-token 而不是 marker
	maxKeysStr := c.DefaultQuery("max-keys", "1000")
	maxKeys, _ := strconv.Atoi(maxKeysStr)
	if maxKeys <= 0 {
		maxKeys = 1000
	}

	// 调用服务层（V2 使用 continuation-token 作为 marker）
	output, err := h.objectService.ListObjects(c.Request.Context(), &service.ListObjectsInput{
		BucketName: bucketName,
		UserID:     userID,
		Prefix:     prefix,
		Delimiter:  delimiter,
		Marker:     continuationToken, // V2 中 continuation-token 等同于 V1 的 marker
		MaxKeys:    maxKeys,
	})

	if err != nil {
		logger.LOG.Error("List objects v2 failed",
			"bucket_name", bucketName,
			"user_id", userID,
			"error", err,
		)

		s3Err := types.MapErrorToS3Error(err)
		types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
		return
	}

	// 构建响应
	response := types.ListBucketResultV2{
		Name:              bucketName,
		Prefix:            prefix,
		KeyCount:          len(output.Objects) + len(output.CommonPrefixes),
		MaxKeys:           maxKeys,
		Delimiter:         delimiter,
		IsTruncated:       output.IsTruncated,
		ContinuationToken: continuationToken,
		Contents:          make([]types.Contents, 0, len(output.Objects)),
		CommonPrefixes:    make([]types.CommonPrefix, 0, len(output.CommonPrefixes)),
	}

	// 获取文件信息以填充 Size 和 LastModified
	for _, obj := range output.Objects {
		fileInfo, err := h.factory.FileInfo().GetByID(c.Request.Context(), obj.FileID)
		if err != nil {
			logger.LOG.Warn("Get file info failed", "file_id", obj.FileID, "error", err)
			continue
		}

		contents := types.Contents{
			Key:          obj.ObjectKey,
			LastModified: fileInfo.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
			ETag:         fmt.Sprintf("\"%s\"", obj.ETag),
			Size:         int64(fileInfo.Size),
			StorageClass: obj.StorageClass,
		}
		response.Contents = append(response.Contents, contents)
	}

	// 添加公共前缀
	for _, commonPrefix := range output.CommonPrefixes {
		response.CommonPrefixes = append(response.CommonPrefixes, types.CommonPrefix{
			Prefix: commonPrefix,
		})
	}

	// 设置 NextContinuationToken
	if output.IsTruncated && output.NextMarker != "" {
		response.NextContinuationToken = output.NextMarker
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
		"max_keys", maxKeys,
		"key_count", response.KeyCount,
		"object_count", len(output.Objects),
		"common_prefix_count", len(output.CommonPrefixes),
		"is_truncated", output.IsTruncated,
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
	// 读取服务端加密头
	sseAlgorithm := c.GetHeader("x-amz-server-side-encryption")
	sseKMSKeyID := c.GetHeader("x-amz-server-side-encryption-aws-kms-key-id")
	sseCustomerKey := c.GetHeader("x-amz-server-side-encryption-customer-key")
	sseCustomerMD5 := c.GetHeader("x-amz-server-side-encryption-customer-key-MD5")
	
	// 读取 ACL header（支持预定义 ACL）
	acl := c.GetHeader("x-amz-acl")
	if acl != "" {
		logger.LOG.Debug("PutObject with ACL",
			"bucket", bucketName,
			"key", objectKey,
			"acl", acl,
		)
	}

	output, err := h.objectService.PutObject(c.Request.Context(), &service.PutObjectInput{
		BucketName:     bucketName,
		ObjectKey:      objectKey,
		UserID:         userID,
		Body:           c.Request.Body,
		ContentType:    contentType,
		ContentMD5:     contentMD5,
		UserMetadata:   userMetadata,
		StorageClass:   storageClass,
		SSEAlgorithm:   sseAlgorithm,
		SSEKMSKeyID:    sseKMSKeyID,
		SSECustomerKey: sseCustomerKey,
		SSECustomerMD5: sseCustomerMD5,
		ACL:            acl,
	})

	if err != nil {
		logger.LOG.Error("Put object failed",
			"bucket", bucketName,
			"key", objectKey,
			"error", err,
		)

		if strings.Contains(err.Error(), "bucket not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else {
			s3Err := types.MapErrorToS3Error(err)
			types.WriteErrorResponse(c.Writer, c.Request, s3Err, "")
		}
		return
	}

	// 返回ETag
	c.Header("ETag", fmt.Sprintf("\"%s\"", output.ETag))
	c.Header("x-amz-version-id", output.VersionID)

	// 如果启用了服务端加密，返回加密头
	if sseAlgorithm != "" {
		c.Header("x-amz-server-side-encryption", sseAlgorithm)
		if sseKMSKeyID != "" {
			c.Header("x-amz-server-side-encryption-aws-kms-key-id", sseKMSKeyID)
		}
	}

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
	
	// 如果是公开访问，需要从 Bucket 获取 userID
	if userID == "" && c.GetBool("is_public_access") {
		bucket, err := h.factory.S3Bucket().GetByName(c.Request.Context(), bucketName, "")
		if err == nil {
			userID = bucket.UserID
		}
	}

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

	// 解析版本ID（可选）
	versionID := c.Query("versionId")

	// 调用服务层
	output, err := h.objectService.GetObject(c.Request.Context(), &service.GetObjectInput{
		BucketName: bucketName,
		ObjectKey:  objectKey,
		UserID:     userID,
		VersionID:  versionID,
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
// HEAD /:bucket/:key+?versionId=xxx
func (h *S3Handler) HeadObject(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")
	versionID := c.Query("versionId") // 可选：版本ID

	// 调用服务层
	output, err := h.objectService.HeadObject(c.Request.Context(), &service.HeadObjectInput{
		BucketName: bucketName,
		ObjectKey:  objectKey,
		UserID:     userID,
		VersionID:  versionID,
	})

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
	versionID := c.Query("versionId") // 可选：版本ID

	// 调用服务层
	err := h.objectService.DeleteObject(c.Request.Context(), bucketName, objectKey, userID, versionID)

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

// ==================== Multipart Upload 相关方法 ====================

// InitiateMultipartUpload 初始化分片上传
// POST /:bucket/:key?uploads
func (h *S3Handler) InitiateMultipartUpload(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")

	// 解析请求头
	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
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
	output, err := h.objectService.InitiateMultipartUpload(c.Request.Context(), &service.InitiateMultipartUploadInput{
		BucketName:   bucketName,
		ObjectKey:    objectKey,
		UserID:       userID,
		ContentType:  contentType,
		UserMetadata: userMetadata,
		StorageClass: storageClass,
	})

	if err != nil {
		logger.LOG.Error("Initiate multipart upload failed",
			"bucket", bucketName,
			"key", objectKey,
			"error", err,
		)

		s3Err := types.MapErrorToS3Error(err)
		types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
		return
	}

	// 构建XML响应
	response := types.InitiateMultipartUploadResult{
		XMLName:  "InitiateMultipartUploadResult",
		Bucket:   output.BucketName,
		Key:      output.ObjectKey,
		UploadID: output.UploadID,
	}

	c.Header("Content-Type", "application/xml")
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(response, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("Initiate multipart upload success",
		"bucket", bucketName,
		"key", objectKey,
		"upload_id", output.UploadID,
	)
}

// UploadPart 上传分片
// PUT /:bucket/:key?uploadId=xxx&partNumber=xxx
func (h *S3Handler) UploadPart(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")
	uploadID := c.Query("uploadId")
	partNumberStr := c.Query("partNumber")

	if uploadID == "" || partNumberStr == "" {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "uploadId and partNumber are required")
		return
	}

	partNumber, err := strconv.Atoi(partNumberStr)
	if err != nil || partNumber < 1 || partNumber > 10000 {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "invalid partNumber")
		return
	}

	// 调用服务层
	output, err := h.objectService.UploadPart(c.Request.Context(), &service.UploadPartInput{
		BucketName: bucketName,
		ObjectKey:  objectKey,
		UploadID:   uploadID,
		PartNumber: partNumber,
		Body:       c.Request.Body,
		UserID:     userID,
	})

	if err != nil {
		logger.LOG.Error("Upload part failed",
			"bucket", bucketName,
			"key", objectKey,
			"upload_id", uploadID,
			"part_number", partNumber,
			"error", err,
		)

		if strings.Contains(err.Error(), "upload not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchUpload, uploadID)
		} else if strings.Contains(err.Error(), "access denied") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrAccessDenied, "")
		} else {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, "")
		}
		return
	}

	// 返回ETag
	c.Header("ETag", fmt.Sprintf("\"%s\"", output.ETag))
	c.Status(http.StatusOK)

	logger.LOG.Info("Upload part success",
		"bucket", bucketName,
		"key", objectKey,
		"upload_id", uploadID,
		"part_number", partNumber,
		"etag", output.ETag,
	)
}

// CompleteMultipartUpload 完成分片上传
// POST /:bucket/:key?uploadId=xxx
func (h *S3Handler) CompleteMultipartUpload(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")
	uploadID := c.Query("uploadId")

	if uploadID == "" {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "uploadId is required")
		return
	}

	// 解析XML请求体
	var completeRequest types.CompleteMultipartUpload
	if err := xml.NewDecoder(c.Request.Body).Decode(&completeRequest); err != nil {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "invalid XML")
		return
	}

	// 转换为服务层格式
	parts := make([]service.PartInfo, 0, len(completeRequest.Part))
	for _, part := range completeRequest.Part {
		parts = append(parts, service.PartInfo{
			PartNumber: part.PartNumber,
			ETag:       strings.Trim(part.ETag, "\""), // 移除引号
		})
	}

	// 调用服务层
	output, err := h.objectService.CompleteMultipartUpload(c.Request.Context(), &service.CompleteMultipartUploadInput{
		BucketName: bucketName,
		ObjectKey:  objectKey,
		UploadID:   uploadID,
		Parts:      parts,
		UserID:     userID,
	})

	if err != nil {
		logger.LOG.Error("Complete multipart upload failed",
			"bucket", bucketName,
			"key", objectKey,
			"upload_id", uploadID,
			"error", err,
		)

		if strings.Contains(err.Error(), "upload not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchUpload, uploadID)
		} else {
			s3Err := types.MapErrorToS3Error(err)
			types.WriteErrorResponse(c.Writer, c.Request, s3Err, "")
		}
		return
	}

	// 构建XML响应
	response := types.CompleteMultipartUploadResult{
		XMLName:  "CompleteMultipartUploadResult",
		Location: output.Location,
		Bucket:   output.BucketName,
		Key:      output.ObjectKey,
		ETag:     fmt.Sprintf("\"%s\"", output.ETag),
	}

	c.Header("Content-Type", "application/xml")
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(response, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("Complete multipart upload success",
		"bucket", bucketName,
		"key", objectKey,
		"upload_id", uploadID,
		"etag", output.ETag,
	)
}

// AbortMultipartUpload 取消分片上传
// DELETE /:bucket/:key?uploadId=xxx
func (h *S3Handler) AbortMultipartUpload(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")
	uploadID := c.Query("uploadId")

	if uploadID == "" {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "uploadId is required")
		return
	}

	// 调用服务层
	err := h.objectService.AbortMultipartUpload(c.Request.Context(), bucketName, objectKey, uploadID, userID)

	if err != nil {
		logger.LOG.Error("Abort multipart upload failed",
			"bucket", bucketName,
			"key", objectKey,
			"upload_id", uploadID,
			"error", err,
		)

		if strings.Contains(err.Error(), "upload not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchUpload, uploadID)
		} else if strings.Contains(err.Error(), "access denied") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrAccessDenied, "")
		} else {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, "")
		}
		return
	}

	// 返回204
	c.Status(http.StatusNoContent)

	logger.LOG.Info("Abort multipart upload success",
		"bucket", bucketName,
		"key", objectKey,
		"upload_id", uploadID,
	)
}

// ListParts 列出分片
// GET /:bucket/:key?uploadId=xxx
func (h *S3Handler) ListParts(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")
	uploadID := c.Query("uploadId")

	if uploadID == "" {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "uploadId is required")
		return
	}

	maxPartsStr := c.DefaultQuery("max-parts", "1000")
	maxParts, _ := strconv.Atoi(maxPartsStr)
	partNumberMarkerStr := c.DefaultQuery("part-number-marker", "0")
	partNumberMarker, _ := strconv.Atoi(partNumberMarkerStr)

	// 调用服务层
	output, err := h.objectService.ListParts(c.Request.Context(), &service.ListPartsInput{
		BucketName:       bucketName,
		ObjectKey:        objectKey,
		UploadID:         uploadID,
		UserID:           userID,
		MaxParts:         maxParts,
		PartNumberMarker: partNumberMarker,
	})

	if err != nil {
		logger.LOG.Error("List parts failed",
			"bucket", bucketName,
			"key", objectKey,
			"upload_id", uploadID,
			"error", err,
		)

		if strings.Contains(err.Error(), "upload not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchUpload, uploadID)
		} else if strings.Contains(err.Error(), "access denied") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrAccessDenied, "")
		} else {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, "")
		}
		return
	}

	// 构建标准 XML 响应
	parts := make([]types.PartDetail, 0, len(output.Parts))
	for _, part := range output.Parts {
		parts = append(parts, types.PartDetail{
			PartNumber:   part.PartNumber,
			ETag:         fmt.Sprintf("\"%s\"", part.ETag),
			Size:         part.Size,
			LastModified: part.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		})
	}

	response := types.ListPartsResult{
		XMLName:          xml.Name{Local: "ListPartsResult"},
		Bucket:           output.BucketName,
		Key:              output.ObjectKey,
		UploadID:         output.UploadID,
		PartNumberMarker: output.PartNumberMarker,
		MaxParts:         output.MaxParts,
		IsTruncated:      output.IsTruncated,
		Parts:            parts,
	}

	if output.IsTruncated && output.NextPartNumberMarker > 0 {
		response.NextPartNumberMarker = output.NextPartNumberMarker
	}

	// 设置 Owner 信息
	response.Owner = types.Owner{
		ID:          userID,
		DisplayName: userID,
	}
	response.Initiator = response.Owner

	c.Header("Content-Type", "application/xml")
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(response, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("List parts success",
		"bucket", bucketName,
		"key", objectKey,
		"upload_id", uploadID,
		"part_count", len(output.Parts),
	)
}

// ListMultipartUploads 列出分片上传会话
// GET /:bucket?uploads
func (h *S3Handler) ListMultipartUploads(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 解析查询参数
	prefix := c.Query("prefix")
	delimiter := c.Query("delimiter")
	keyMarker := c.Query("key-marker")
	uploadIDMarker := c.Query("upload-id-marker")
	maxUploadsStr := c.DefaultQuery("max-uploads", "1000")
	maxUploads, _ := strconv.Atoi(maxUploadsStr)
	if maxUploads <= 0 {
		maxUploads = 1000
	}

	// 调用服务层
	output, err := h.objectService.ListMultipartUploads(c.Request.Context(), &service.ListMultipartUploadsInput{
		BucketName:     bucketName,
		UserID:         userID,
		Prefix:         prefix,
		Delimiter:      delimiter,
		KeyMarker:      keyMarker,
		UploadIDMarker: uploadIDMarker,
		MaxUploads:     maxUploads,
	})

	if err != nil {
		logger.LOG.Error("List multipart uploads failed",
			"bucket_name", bucketName,
			"user_id", userID,
			"error", err,
		)

		s3Err := types.MapErrorToS3Error(err)
		types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
		return
	}

	// 构建响应
	uploads := make([]types.Upload, 0, len(output.Uploads))
	for _, upload := range output.Uploads {
		uploads = append(uploads, types.Upload{
			Key:          upload.ObjectKey,
			UploadID:     upload.UploadID,
			StorageClass: "STANDARD",
			Initiated:    upload.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
			Initiator: types.Owner{
				ID:          upload.UserID,
				DisplayName: upload.UserID,
			},
			Owner: types.Owner{
				ID:          upload.UserID,
				DisplayName: upload.UserID,
			},
		})
	}

	commonPrefixes := make([]types.CommonPrefix, 0, len(output.CommonPrefixes))
	for _, prefix := range output.CommonPrefixes {
		commonPrefixes = append(commonPrefixes, types.CommonPrefix{
			Prefix: prefix,
		})
	}

	response := types.ListMultipartUploadsResult{
		XMLName:            xml.Name{Local: "ListMultipartUploadsResult"},
		Bucket:             output.BucketName,
		Prefix:             output.Prefix,
		Delimiter:          output.Delimiter,
		KeyMarker:          output.KeyMarker,
		UploadIDMarker:     output.UploadIDMarker,
		NextKeyMarker:      output.NextKeyMarker,
		NextUploadIDMarker: output.NextUploadIDMarker,
		MaxUploads:         output.MaxUploads,
		IsTruncated:        output.IsTruncated,
		Uploads:            uploads,
		CommonPrefixes:     commonPrefixes,
	}

	// 写入XML响应
	c.Header("Content-Type", "application/xml")
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(response, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("List multipart uploads success",
		"bucket_name", bucketName,
		"user_id", userID,
		"prefix", prefix,
		"max_uploads", maxUploads,
		"upload_count", len(output.Uploads),
		"common_prefix_count", len(output.CommonPrefixes),
		"is_truncated", output.IsTruncated,
	)
}

// CopyObject 复制对象
// PUT /:bucket/:key with x-amz-copy-source header
func (h *S3Handler) CopyObject(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")

	// 解析 x-amz-copy-source 头（格式：/bucket/key 或 bucket/key）
	copySource := c.GetHeader("x-amz-copy-source")
	if copySource == "" {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "x-amz-copy-source header is required")
		return
	}

	// 移除开头的斜杠
	if strings.HasPrefix(copySource, "/") {
		copySource = copySource[1:]
	}

	// 解析源 bucket 和 key
	parts := strings.SplitN(copySource, "/", 2)
	if len(parts) != 2 {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "invalid x-amz-copy-source format")
		return
	}
	sourceBucket := parts[0]
	sourceKey := parts[1]

	// 解析可选的元数据指令
	copyMetadataDirective := c.GetHeader("x-amz-metadata-directive") // COPY 或 REPLACE
	storageClass := c.GetHeader("x-amz-storage-class")

	// 解析用户元数据（如果 metadata-directive 是 REPLACE）
	userMetadata := make(map[string]string)
	if copyMetadataDirective == "REPLACE" {
		for key, values := range c.Request.Header {
			if strings.HasPrefix(strings.ToLower(key), "x-amz-meta-") {
				metaKey := key[len("x-amz-meta-"):]
				if len(values) > 0 {
					userMetadata[metaKey] = values[0]
				}
			}
		}
	}

	// 调用服务层
	output, err := h.objectService.CopyObject(c.Request.Context(), &service.CopyObjectInput{
		SourceBucket:      sourceBucket,
		SourceKey:         sourceKey,
		DestinationBucket: bucketName,
		DestinationKey:    objectKey,
		UserID:            userID,
		MetadataDirective: copyMetadataDirective,
		UserMetadata:      userMetadata,
		StorageClass:      storageClass,
	})

	if err != nil {
		logger.LOG.Error("Copy object failed",
			"source_bucket", sourceBucket,
			"source_key", sourceKey,
			"dest_bucket", bucketName,
			"dest_key", objectKey,
			"error", err,
		)

		if strings.Contains(err.Error(), "bucket not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else {
			s3Err := types.MapErrorToS3Error(err)
			types.WriteErrorResponse(c.Writer, c.Request, s3Err, sourceBucket+"/"+sourceKey)
		}
		return
	}

	// 构建XML响应
	response := types.CopyObjectResult{
		XMLName:      "CopyObjectResult",
		LastModified: output.LastModified.Format("2006-01-02T15:04:05.000Z"),
		ETag:         fmt.Sprintf("\"%s\"", output.ETag),
	}

	c.Header("Content-Type", "application/xml")
	c.Header("x-amz-copy-source-version-id", output.SourceVersionID)
	c.Header("x-amz-version-id", output.VersionID)
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(response, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("Copy object success",
		"source_bucket", sourceBucket,
		"source_key", sourceKey,
		"dest_bucket", bucketName,
		"dest_key", objectKey,
		"etag", output.ETag,
	)
}

// DeleteObjects 批量删除对象
// POST /:bucket?delete
func (h *S3Handler) DeleteObjects(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 解析XML请求体
	var deleteRequest types.DeleteRequest
	if err := xml.NewDecoder(c.Request.Body).Decode(&deleteRequest); err != nil {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "invalid XML")
		return
	}

	// 转换为服务层格式
	objects := make([]service.ObjectToDelete, 0, len(deleteRequest.Object))
	for _, obj := range deleteRequest.Object {
		objects = append(objects, service.ObjectToDelete{
			Key:       obj.Key,
			VersionID: obj.VersionID,
		})
	}

	// 调用服务层
	output, err := h.objectService.DeleteObjects(c.Request.Context(), &service.DeleteObjectsInput{
		BucketName: bucketName,
		UserID:     userID,
		Objects:    objects,
		Quiet:      deleteRequest.Quiet,
	})

	if err != nil {
		logger.LOG.Error("Delete objects failed",
			"bucket", bucketName,
			"user_id", userID,
			"error", err,
		)

		s3Err := types.MapErrorToS3Error(err)
		types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
		return
	}

	// 构建XML响应
	response := types.DeleteResult{
		XMLName: "DeleteResult",
		Deleted: make([]types.DeletedObject, 0, len(output.Deleted)),
		Error:   make([]types.DeleteError, 0, len(output.Errors)),
	}

	for _, deleted := range output.Deleted {
		response.Deleted = append(response.Deleted, types.DeletedObject{
			Key:       deleted.Key,
			VersionID: deleted.VersionID,
		})
	}

	for _, errInfo := range output.Errors {
		response.Error = append(response.Error, types.DeleteError{
			Key:       errInfo.Key,
			Code:      errInfo.Code,
			Message:   errInfo.Message,
			VersionID: errInfo.VersionID,
		})
	}

	c.Header("Content-Type", "application/xml")
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(response, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("Delete objects success",
		"bucket", bucketName,
		"user_id", userID,
		"total", len(objects),
		"deleted", len(output.Deleted),
		"errors", len(output.Errors),
	)
}

// GeneratePresignedURL 生成预签名URL
// GET /:bucket/:key?X-Amz-Expires=3600&X-Amz-Algorithm=AWS4-HMAC-SHA256
// 或者通过查询参数：?presign=true&expires=3600&method=GET
func (h *S3Handler) GeneratePresignedURL(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")

	// 获取API Key信息
	accessKeyID := c.GetString("access_key")
	apiKeyInterface, exists := c.Get("api_key")
	if !exists {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidAccessKeyId, "")
		return
	}
	apiKey := apiKeyInterface.(*models.ApiKey)

	// 解析查询参数
	expiresStr := c.Query("X-Amz-Expires")
	if expiresStr == "" {
		expiresStr = c.Query("expires")
	}
	if expiresStr == "" {
		expiresStr = "3600" // 默认1小时
	}

	method := c.Query("method")
	if method == "" {
		method = c.Request.Method
	}
	if method == "" {
		method = "GET"
	}

	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil || expires <= 0 {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "invalid expires parameter")
		return
	}

	// 限制过期时间（最长7天）
	if expires > 604800 {
		expires = 604800
	}

	// 调用服务层
	output, err := h.objectService.GeneratePresignedURL(c.Request.Context(), &service.PresignedURLInput{
		BucketName:  bucketName,
		ObjectKey:   objectKey,
		Method:      method,
		Expires:     expires,
		UserID:      userID,
		AccessKeyID: accessKeyID,
		SecretKey:   apiKey.S3SecretKey, // 使用 S3 Secret Key
		Headers:     nil,                 // 可以扩展支持自定义请求头
	})

	if err != nil {
		logger.LOG.Error("Generate presigned URL failed",
			"bucket", bucketName,
			"key", objectKey,
			"user_id", userID,
			"error", err,
		)

		s3Err := types.MapErrorToS3Error(err)
		types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
		return
	}

	// 返回JSON响应
	c.JSON(http.StatusOK, gin.H{
		"url":     output.URL,
		"expires": output.Expires,
	})

	logger.LOG.Info("Generate presigned URL success",
		"bucket", bucketName,
		"key", objectKey,
		"user_id", userID,
		"method", method,
		"expires", expires,
	)
}

// PutObjectTagging 设置对象标签
// PUT /:bucket/:key+?tagging
func (h *S3Handler) PutObjectTagging(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")
	versionID := c.Query("versionId")

	// 解析XML请求体
	var tagging types.Tagging
	if err := xml.NewDecoder(c.Request.Body).Decode(&tagging); err != nil {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "invalid XML")
		return
	}

	// 转换标签格式
	tags := make(map[string]string)
	for _, tag := range tagging.TagSet.Tags {
		tags[tag.Key] = tag.Value
	}

	// 调用服务层
	err := h.objectService.PutObjectTagging(c.Request.Context(), &service.PutObjectTaggingInput{
		BucketName: bucketName,
		ObjectKey:  objectKey,
		UserID:     userID,
		VersionID:  versionID,
		Tags:       tags,
	})

	if err != nil {
		logger.LOG.Error("Put object tagging failed",
			"bucket", bucketName,
			"key", objectKey,
			"version_id", versionID,
			"error", err,
		)

		if strings.Contains(err.Error(), "bucket not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else if strings.Contains(err.Error(), "object not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchKey, bucketName+"/"+objectKey)
		} else if strings.Contains(err.Error(), "too many") || strings.Contains(err.Error(), "length must") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, err.Error())
		} else {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, "")
		}
		return
	}

	// 返回204 No Content
	c.Status(http.StatusNoContent)

	logger.LOG.Info("Put object tagging success",
		"bucket", bucketName,
		"key", objectKey,
		"version_id", versionID,
		"tags_count", len(tags),
	)
}

// GetObjectTagging 获取对象标签
// GET /:bucket/:key+?tagging
func (h *S3Handler) GetObjectTagging(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")
	versionID := c.Query("versionId")

	// 调用服务层
	output, err := h.objectService.GetObjectTagging(c.Request.Context(), &service.GetObjectTaggingInput{
		BucketName: bucketName,
		ObjectKey:  objectKey,
		UserID:     userID,
		VersionID:  versionID,
	})

	if err != nil {
		logger.LOG.Error("Get object tagging failed",
			"bucket", bucketName,
			"key", objectKey,
			"version_id", versionID,
			"error", err,
		)

		if strings.Contains(err.Error(), "bucket not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else if strings.Contains(err.Error(), "object not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchKey, bucketName+"/"+objectKey)
		} else {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, "")
		}
		return
	}

	// 构建XML响应
	response := types.Tagging{
		XMLName: xml.Name{Local: "Tagging"},
		TagSet: types.TagSet{
			Tags: make([]types.Tag, 0, len(output.Tags)),
		},
	}

	for key, value := range output.Tags {
		response.TagSet.Tags = append(response.TagSet.Tags, types.Tag{
			Key:   key,
			Value: value,
		})
	}

	c.Header("Content-Type", "application/xml")
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(response, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("Get object tagging success",
		"bucket", bucketName,
		"key", objectKey,
		"version_id", versionID,
		"tags_count", len(output.Tags),
	)
}

// DeleteObjectTagging 删除对象标签
// DELETE /:bucket/:key+?tagging
func (h *S3Handler) DeleteObjectTagging(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")
	versionID := c.Query("versionId")

	// 调用服务层
	err := h.objectService.DeleteObjectTagging(c.Request.Context(), &service.DeleteObjectTaggingInput{
		BucketName: bucketName,
		ObjectKey:  objectKey,
		UserID:     userID,
		VersionID:  versionID,
	})

	if err != nil {
		logger.LOG.Error("Delete object tagging failed",
			"bucket", bucketName,
			"key", objectKey,
			"version_id", versionID,
			"error", err,
		)

		if strings.Contains(err.Error(), "bucket not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else if strings.Contains(err.Error(), "object not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchKey, bucketName+"/"+objectKey)
		} else {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, "")
		}
		return
	}

	// 返回204 No Content
	c.Status(http.StatusNoContent)

	logger.LOG.Info("Delete object tagging success",
		"bucket", bucketName,
		"key", objectKey,
		"version_id", versionID,
	)
}

// PutBucketACL 设置Bucket ACL
// PUT /:bucket?acl
func (h *S3Handler) PutBucketACL(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 解析XML请求体
	var aclPolicy types.AccessControlPolicy
	if err := xml.NewDecoder(c.Request.Body).Decode(&aclPolicy); err != nil {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "invalid XML")
		return
	}

	// 调用服务层
	err := h.bucketService.PutBucketACL(c.Request.Context(), &service.PutBucketACLInput{
		BucketName: bucketName,
		UserID:     userID,
		ACL:        &aclPolicy,
	})

	if err != nil {
		logger.LOG.Error("Put bucket ACL failed",
			"bucket", bucketName,
			"user_id", userID,
			"error", err,
		)

		if strings.Contains(err.Error(), "bucket not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else {
			s3Err := types.MapErrorToS3Error(err)
			types.WriteErrorResponse(c.Writer, c.Request, s3Err, "")
		}
		return
	}

	// 返回204 No Content
	c.Status(http.StatusNoContent)

	logger.LOG.Info("Put bucket ACL success",
		"bucket", bucketName,
		"user_id", userID,
		"grants_count", len(aclPolicy.AccessControlList.Grants),
	)
}

// GetBucketACL 获取Bucket ACL
// GET /:bucket?acl
func (h *S3Handler) GetBucketACL(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 调用服务层
	output, err := h.bucketService.GetBucketACL(c.Request.Context(), &service.GetBucketACLInput{
		BucketName: bucketName,
		UserID:     userID,
	})

	if err != nil {
		logger.LOG.Error("Get bucket ACL failed",
			"bucket", bucketName,
			"user_id", userID,
			"error", err,
		)

		s3Err := types.MapErrorToS3Error(err)
		types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
		return
	}

	// 构建XML响应
	c.Header("Content-Type", "application/xml")
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(output.ACL, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("Get bucket ACL success",
		"bucket", bucketName,
		"user_id", userID,
		"grants_count", len(output.ACL.AccessControlList.Grants),
	)
}

// PutObjectACL 设置对象ACL
// PUT /:bucket/:key+?acl
func (h *S3Handler) PutObjectACL(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")
	versionID := c.Query("versionId")

	// 解析XML请求体
	var aclPolicy types.AccessControlPolicy
	if err := xml.NewDecoder(c.Request.Body).Decode(&aclPolicy); err != nil {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "invalid XML")
		return
	}

	// 调用服务层
	err := h.objectService.PutObjectACL(c.Request.Context(), &service.PutObjectACLInput{
		BucketName: bucketName,
		ObjectKey:  objectKey,
		UserID:     userID,
		VersionID:  versionID,
		ACL:        &aclPolicy,
	})

	if err != nil {
		logger.LOG.Error("Put object ACL failed",
			"bucket", bucketName,
			"key", objectKey,
			"version_id", versionID,
			"error", err,
		)

		if strings.Contains(err.Error(), "bucket not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else if strings.Contains(err.Error(), "object not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchKey, bucketName+"/"+objectKey)
		} else {
			s3Err := types.MapErrorToS3Error(err)
			types.WriteErrorResponse(c.Writer, c.Request, s3Err, "")
		}
		return
	}

	// 返回204 No Content
	c.Status(http.StatusNoContent)

	logger.LOG.Info("Put object ACL success",
		"bucket", bucketName,
		"key", objectKey,
		"version_id", versionID,
		"grants_count", len(aclPolicy.AccessControlList.Grants),
	)
}

// GetObjectACL 获取对象ACL
// GET /:bucket/:key+?acl
func (h *S3Handler) GetObjectACL(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")
	objectKey := c.Param("key")
	versionID := c.Query("versionId")

	// 调用服务层
	output, err := h.objectService.GetObjectACL(c.Request.Context(), &service.GetObjectACLInput{
		BucketName: bucketName,
		ObjectKey:  objectKey,
		UserID:     userID,
		VersionID:  versionID,
	})

	if err != nil {
		logger.LOG.Error("Get object ACL failed",
			"bucket", bucketName,
			"key", objectKey,
			"version_id", versionID,
			"error", err,
		)

		if strings.Contains(err.Error(), "bucket not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else if strings.Contains(err.Error(), "object not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchKey, bucketName+"/"+objectKey)
		} else {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, "")
		}
		return
	}

	// 构建XML响应
	c.Header("Content-Type", "application/xml")
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(output.ACL, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("Get object ACL success",
		"bucket", bucketName,
		"key", objectKey,
		"version_id", versionID,
		"grants_count", len(output.ACL.AccessControlList.Grants),
	)
}

// PutBucketPolicy 设置Bucket Policy
// PUT /:bucket?policy
func (h *S3Handler) PutBucketPolicy(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 解析JSON请求体
	var policy types.BucketPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "invalid JSON")
		return
	}

	// 调用服务层
	err := h.bucketService.PutBucketPolicy(c.Request.Context(), &service.PutBucketPolicyInput{
		BucketName: bucketName,
		UserID:     userID,
		Policy:     &policy,
	})

	if err != nil {
		logger.LOG.Error("Put bucket policy failed",
			"bucket", bucketName,
			"user_id", userID,
			"error", err,
		)

		if strings.Contains(err.Error(), "bucket not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else {
			s3Err := types.MapErrorToS3Error(err)
			types.WriteErrorResponse(c.Writer, c.Request, s3Err, "")
		}
		return
	}

	// 返回204 No Content
	c.Status(http.StatusNoContent)

	logger.LOG.Info("Put bucket policy success",
		"bucket", bucketName,
		"user_id", userID,
		"statements_count", len(policy.Statement),
	)
}

// GetBucketPolicy 获取Bucket Policy
// GET /:bucket?policy
func (h *S3Handler) GetBucketPolicy(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 调用服务层
	output, err := h.bucketService.GetBucketPolicy(c.Request.Context(), &service.GetBucketPolicyInput{
		BucketName: bucketName,
		UserID:     userID,
	})

	if err != nil {
		logger.LOG.Error("Get bucket policy failed",
			"bucket", bucketName,
			"user_id", userID,
			"error", err,
		)

		if strings.Contains(err.Error(), "bucket not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else if strings.Contains(err.Error(), "policy not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, "")
		}
		return
	}

	// 返回JSON响应
	c.JSON(http.StatusOK, output.Policy)

	logger.LOG.Info("Get bucket policy success",
		"bucket", bucketName,
		"user_id", userID,
		"statements_count", len(output.Policy.Statement),
	)
}

// DeleteBucketPolicy 删除Bucket Policy
// DELETE /:bucket?policy
func (h *S3Handler) DeleteBucketPolicy(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 调用服务层
	err := h.bucketService.DeleteBucketPolicy(c.Request.Context(), &service.DeleteBucketPolicyInput{
		BucketName: bucketName,
		UserID:     userID,
	})

	if err != nil {
		logger.LOG.Error("Delete bucket policy failed",
			"bucket", bucketName,
			"user_id", userID,
			"error", err,
		)

		s3Err := types.MapErrorToS3Error(err)
		types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
		return
	}

	// 返回204 No Content
	c.Status(http.StatusNoContent)

	logger.LOG.Info("Delete bucket policy success",
		"bucket", bucketName,
		"user_id", userID,
	)
}

// PutBucketLifecycle 设置Bucket Lifecycle
// PUT /:bucket?lifecycle
func (h *S3Handler) PutBucketLifecycle(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 解析XML请求体（S3 Lifecycle使用XML格式）
	var lifecycleConfig types.LifecycleConfiguration
	if err := xml.NewDecoder(c.Request.Body).Decode(&lifecycleConfig); err != nil {
		types.WriteErrorResponse(c.Writer, c.Request, types.ErrInvalidArgument, "invalid XML")
		return
	}

	// 调用服务层
	err := h.bucketService.PutBucketLifecycle(c.Request.Context(), &service.PutBucketLifecycleInput{
		BucketName: bucketName,
		UserID:     userID,
		Lifecycle:  &lifecycleConfig,
	})

	if err != nil {
		logger.LOG.Error("Put bucket lifecycle failed",
			"bucket", bucketName,
			"user_id", userID,
			"error", err,
		)

		if strings.Contains(err.Error(), "bucket not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else {
			s3Err := types.MapErrorToS3Error(err)
			types.WriteErrorResponse(c.Writer, c.Request, s3Err, "")
		}
		return
	}

	// 返回204 No Content
	c.Status(http.StatusNoContent)

	logger.LOG.Info("Put bucket lifecycle success",
		"bucket", bucketName,
		"user_id", userID,
		"rules_count", len(lifecycleConfig.Rules),
	)
}

// GetBucketLifecycle 获取Bucket Lifecycle
// GET /:bucket?lifecycle
func (h *S3Handler) GetBucketLifecycle(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 调用服务层
	output, err := h.bucketService.GetBucketLifecycle(c.Request.Context(), &service.GetBucketLifecycleInput{
		BucketName: bucketName,
		UserID:     userID,
	})

	if err != nil {
		logger.LOG.Error("Get bucket lifecycle failed",
			"bucket", bucketName,
			"user_id", userID,
			"error", err,
		)

		if strings.Contains(err.Error(), "bucket not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else if strings.Contains(err.Error(), "lifecycle configuration not found") {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrNoSuchBucket, bucketName)
		} else {
			types.WriteErrorResponse(c.Writer, c.Request, types.ErrInternalError, "")
		}
		return
	}

	// 构建XML响应
	c.Header("Content-Type", "application/xml")
	c.Status(http.StatusOK)

	xmlData, _ := xml.MarshalIndent(output.Lifecycle, "", "  ")
	c.Writer.Write([]byte(xml.Header))
	c.Writer.Write(xmlData)

	logger.LOG.Info("Get bucket lifecycle success",
		"bucket", bucketName,
		"user_id", userID,
		"rules_count", len(output.Lifecycle.Rules),
	)
}

// DeleteBucketLifecycle 删除Bucket Lifecycle
// DELETE /:bucket?lifecycle
func (h *S3Handler) DeleteBucketLifecycle(c *gin.Context) {
	userID := c.GetString("user_id")
	bucketName := c.Param("bucket")

	// 调用服务层
	err := h.bucketService.DeleteBucketLifecycle(c.Request.Context(), &service.DeleteBucketLifecycleInput{
		BucketName: bucketName,
		UserID:     userID,
	})

	if err != nil {
		logger.LOG.Error("Delete bucket lifecycle failed",
			"bucket", bucketName,
			"user_id", userID,
			"error", err,
		)

		s3Err := types.MapErrorToS3Error(err)
		types.WriteErrorResponse(c.Writer, c.Request, s3Err, bucketName)
		return
	}

	// 返回204 No Content
	c.Status(http.StatusNoContent)

	logger.LOG.Info("Delete bucket lifecycle success",
		"bucket", bucketName,
		"user_id", userID,
	)
}

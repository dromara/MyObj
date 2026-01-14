package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"myobj/src/core/service"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/hash"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// S3ObjectService S3对象服务
type S3ObjectService struct {
	objectMetadataRepo repository.S3ObjectMetadataRepository
	bucketRepo         repository.S3BucketRepository
	fileService        *service.FileService
	factory            *impl.RepositoryFactory
}

// NewS3ObjectService 创建S3对象服务
func NewS3ObjectService(factory *impl.RepositoryFactory, fileService *service.FileService) *S3ObjectService {
	return &S3ObjectService{
		objectMetadataRepo: factory.S3ObjectMetadata(),
		bucketRepo:         factory.S3Bucket(),
		fileService:        fileService,
		factory:            factory,
	}
}

// PutObjectInput PutObject输入参数
type PutObjectInput struct {
	BucketName   string
	ObjectKey    string
	UserID       string
	Body         io.Reader
	ContentType  string
	ContentMD5   string
	UserMetadata map[string]string
	StorageClass string
}

// PutObjectOutput PutObject输出
type PutObjectOutput struct {
	ETag      string
	VersionID string
}

// PutObject 上传对象
func (s *S3ObjectService) PutObject(ctx context.Context, input *PutObjectInput) (*PutObjectOutput, error) {
	// 1. 验证Bucket是否存在
	bucket, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("bucket not found")
		}
		logger.LOG.Error("Get bucket failed",
			"bucket_name", input.BucketName,
			"error", err,
		)
		return nil, err
	}

	// 2. 检查用户空间是否足够
	user, err := s.factory.User().GetByID(ctx, input.UserID)
	if err != nil {
		logger.LOG.Error("Get user failed", "user_id", input.UserID, "error", err)
		return nil, fmt.Errorf("get user failed: %w", err)
	}

	// 3. 选择存储磁盘
	disks, err := s.factory.Disk().List(ctx, 0, 1000)
	if err != nil || len(disks) == 0 {
		logger.LOG.Error("No available disk", "error", err)
		return nil, fmt.Errorf("no available disk")
	}

	// 选择剩余空间最大的磁盘
	var bestDisk *models.Disk
	var maxFreeSpace int64 = -1
	for _, disk := range disks {
		freeSpaceBytes := int64(disk.Size) * 1024 * 1024 * 1024
		if freeSpaceBytes > maxFreeSpace {
			maxFreeSpace = freeSpaceBytes
			bestDisk = disk
		}
	}
	if bestDisk == nil {
		return nil, fmt.Errorf("no available disk")
	}

	// 4. 创建临时目录并保存文件
	tempDir := filepath.Join(bestDisk.DataPath, "temp", "s3_upload_"+uuid.New().String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		logger.LOG.Error("Create temp dir failed", "error", err, "path", tempDir)
		return nil, fmt.Errorf("create temp dir failed: %w", err)
	}

	tempFilePath := filepath.Join(tempDir, "upload.tmp")
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("create temp file failed: %w", err)
	}

	// 5. 读取文件内容并计算哈希
	hasher := md5.New()
	multiWriter := io.MultiWriter(tempFile, hasher)
	fileSize, err := io.Copy(multiWriter, input.Body)
	if err != nil {
		tempFile.Close()
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("save file failed: %w", err)
	}
	tempFile.Close()

	// 计算MD5作为ETag
	etag := hex.EncodeToString(hasher.Sum(nil))

	// 6. 如果提供了Content-MD5，验证完整性
	if input.ContentMD5 != "" && input.ContentMD5 != etag {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("content MD5 mismatch")
	}

	// 7. 检查用户空间（如果不是无限空间）
	if user.Space > 0 && user.FreeSpace < fileSize {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("insufficient user space")
	}

	// 8. 计算文件的BLAKE3哈希（用于去重）
	blake3Hasher := hash.NewFastBlake3Hasher()
	blake3Hash, _, err := blake3Hasher.ComputeFileHash(tempFilePath)
	if err != nil {
		logger.LOG.Warn("Calculate BLAKE3 hash failed, use MD5", "error", err)
		blake3Hash = etag
	}

	// 9. 检查是否已存在相同文件（秒传）
	existingFile, err := s.factory.FileInfo().GetByHash(ctx, blake3Hash)
	var fileInfo *models.FileInfo
	var isNewFile bool = true

	if err == nil && existingFile != nil {
		// 文件已存在，秒传
		fileInfo = existingFile
		isNewFile = false
		logger.LOG.Info("File already exists, instant upload",
			"file_hash", blake3Hash,
			"file_id", fileInfo.ID,
		)
		// 删除临时文件
		os.RemoveAll(tempDir)
	} else {
		// 新文件，需要移动到最终存储位置
		finalFileName := fmt.Sprintf("%s_%s", blake3Hash, filepath.Base(input.ObjectKey))
		finalPath := filepath.Join(bestDisk.DataPath, "files", finalFileName)

		// 确保目标目录存在
		if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
			os.RemoveAll(tempDir)
			return nil, fmt.Errorf("create target dir failed: %w", err)
		}

		// 移动文件
		if err := os.Rename(tempFilePath, finalPath); err != nil {
			// 如果跨分区移动失败，尝试复制
			if err := copyFile(tempFilePath, finalPath); err != nil {
				os.RemoveAll(tempDir)
				return nil, fmt.Errorf("move file failed: %w", err)
			}
			os.Remove(tempFilePath)
		}
		os.RemoveAll(tempDir)

		// 创建FileInfo记录
		fileInfo = &models.FileInfo{
			ID:              uuid.New().String(),
			FileHash:        blake3Hash,
			ChunkSignature:  blake3Hash[:16], // 使用前16位作为签名
			FirstChunkHash:  "",
			SecondChunkHash: "",
			ThirdChunkHash:  "",
			Size:            int(fileSize),
			Mime:            input.ContentType,
			Path:            finalPath,
			IsEnc:           false,
			ThumbnailImg:    "",
			CreatedAt:       custom_type.Now(),
		}

		if err := s.factory.FileInfo().Create(ctx, fileInfo); err != nil {
			os.Remove(finalPath)
			logger.LOG.Error("Create file info failed", "error", err)
			return nil, fmt.Errorf("create file info failed: %w", err)
		}

		// 更新用户空间
		if user.Space > 0 {
			user.FreeSpace -= fileSize
			if err := s.factory.User().Update(ctx, user); err != nil {
				logger.LOG.Error("Update user space failed", "error", err)
				// 不回滚，因为文件已经保存
			}
		}
	}

	// 10. 检查是否已存在相同对象元数据（覆盖上传）
	existingMetadata, err := s.objectMetadataRepo.GetByKey(ctx, input.BucketName, input.ObjectKey, input.UserID)
	if err == nil && existingMetadata != nil {
		// 标记旧版本为非最新
		if err := s.objectMetadataRepo.MarkOldVersions(ctx, input.BucketName, input.ObjectKey, input.UserID); err != nil {
			logger.LOG.Warn("Mark old versions failed", "error", err)
		}
	}

	// 11. 创建对象元数据
	versionID := uuid.New().String()
	userMetadataJSON := ""
	if len(input.UserMetadata) > 0 {
		// 简单JSON序列化
		metaPairs := []string{}
		for k, v := range input.UserMetadata {
			metaPairs = append(metaPairs, fmt.Sprintf(`"%s":"%s"`, k, v))
		}
		userMetadataJSON = "{" + strings.Join(metaPairs, ",") + "}"
	}

	storageClass := input.StorageClass
	if storageClass == "" {
		storageClass = "STANDARD"
	}

	objectMetadata := &models.S3ObjectMetadata{
		FileID:       fileInfo.ID,
		BucketName:   input.BucketName,
		ObjectKey:    input.ObjectKey,
		UserID:       input.UserID,
		ETag:         etag,
		StorageClass: storageClass,
		ContentType:  input.ContentType,
		UserMetadata: userMetadataJSON,
		VersionID:    versionID,
		IsLatest:     true,
		CreatedAt:    custom_type.Now(),
		UpdatedAt:    custom_type.Now(),
	}

	if err := s.objectMetadataRepo.Create(ctx, objectMetadata); err != nil {
		logger.LOG.Error("Create object metadata failed", "error", err)
		// 如果是新文件，需要清理
		if isNewFile {
			s.factory.FileInfo().Delete(ctx, fileInfo.ID)
			os.Remove(fileInfo.Path)
		}
		return nil, fmt.Errorf("create object metadata failed: %w", err)
	}

	// 12. 创建UserFiles关联（将对象关联到Bucket的虚拟路径）
	userFile := &models.UserFiles{
		UserID:      input.UserID,
		FileID:      fileInfo.ID,
		FileName:    filepath.Base(input.ObjectKey),
		VirtualPath: fmt.Sprintf("%d", bucket.VirtualPathID),
		IsPublic:    false,
		CreatedAt:   custom_type.Now(),
		UfID:        uuid.NewString(),
	}

	if err := s.factory.UserFiles().Create(ctx, userFile); err != nil {
		logger.LOG.Error("Create user file failed", "error", err)
		// 不回滚，对象元数据已创建
	}

	logger.LOG.Info("PutObject success",
		"bucket", input.BucketName,
		"key", input.ObjectKey,
		"size", fileSize,
		"etag", etag,
		"is_new_file", isNewFile,
	)

	return &PutObjectOutput{
		ETag:      etag,
		VersionID: versionID,
	}, nil
}

// GetObjectInput GetObject输入参数
type GetObjectInput struct {
	BucketName string
	ObjectKey  string
	UserID     string
	VersionID  string // 可选：版本ID
	RangeStart int64  // Range请求起始位置
	RangeEnd   int64  // Range请求结束位置（0表示到文件末尾）
}

// GetObjectOutput GetObject输出
type GetObjectOutput struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
	ETag          string
	LastModified  time.Time
	VersionID     string
	UserMetadata  map[string]string
	ActualStart   int64 // 实际返回的起始位置
	ActualEnd     int64 // 实际返回的结束位置
}

// GetObject 获取对象
func (s *S3ObjectService) GetObject(ctx context.Context, input *GetObjectInput) (*GetObjectOutput, error) {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("bucket not found")
		}
		return nil, err
	}

	// 2. 获取对象元数据
	objectMetadata, err := s.objectMetadataRepo.GetByKey(ctx, input.BucketName, input.ObjectKey, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("object not found")
		}
		logger.LOG.Error("Get object metadata failed",
			"bucket", input.BucketName,
			"key", input.ObjectKey,
			"error", err,
		)
		return nil, err
	}

	// 3. 获取文件信息
	fileInfo, err := s.factory.FileInfo().GetByID(ctx, objectMetadata.FileID)
	if err != nil {
		logger.LOG.Error("Get file info failed",
			"file_id", objectMetadata.FileID,
			"error", err,
		)
		return nil, fmt.Errorf("file info not found")
	}

	// 4. 打开文件
	file, err := os.Open(fileInfo.Path)
	if err != nil {
		logger.LOG.Error("Open file failed",
			"path", fileInfo.Path,
			"error", err,
		)
		return nil, fmt.Errorf("open file failed: %w", err)
	}

	// 5. 处理Range请求
	actualStart := input.RangeStart
	actualEnd := input.RangeEnd
	contentLength := int64(fileInfo.Size)

	if input.RangeStart > 0 || input.RangeEnd > 0 {
		// Range请求
		if actualEnd == 0 || actualEnd >= int64(fileInfo.Size) {
			actualEnd = int64(fileInfo.Size) - 1
		}

		if actualStart < 0 {
			actualStart = 0
		}

		if actualStart > actualEnd {
			file.Close()
			return nil, fmt.Errorf("invalid range")
		}

		// Seek到起始位置
		if _, err := file.Seek(actualStart, 0); err != nil {
			file.Close()
			return nil, fmt.Errorf("seek file failed: %w", err)
		}

		contentLength = actualEnd - actualStart + 1
	}

	// 6. 解析用户元数据
	userMetadata := make(map[string]string)
	if objectMetadata.UserMetadata != "" {
		// 简单JSON解析（实际应使用json.Unmarshal）
		metaStr := strings.Trim(objectMetadata.UserMetadata, "{}")
		if metaStr != "" {
			pairs := strings.Split(metaStr, ",")
			for _, pair := range pairs {
				kv := strings.SplitN(pair, ":", 2)
				if len(kv) == 2 {
					key := strings.Trim(kv[0], `"`)
					value := strings.Trim(kv[1], `"`)
					userMetadata[key] = value
				}
			}
		}
	}

	logger.LOG.Info("GetObject success",
		"bucket", input.BucketName,
		"key", input.ObjectKey,
		"size", contentLength,
		"range", fmt.Sprintf("%d-%d", actualStart, actualEnd),
	)

	return &GetObjectOutput{
		Body:          file,
		ContentType:   objectMetadata.ContentType,
		ContentLength: contentLength,
		ETag:          objectMetadata.ETag,
		LastModified:  time.Time(objectMetadata.UpdatedAt),
		VersionID:     objectMetadata.VersionID,
		UserMetadata:  userMetadata,
		ActualStart:   actualStart,
		ActualEnd:     actualEnd,
	}, nil
}

// DeleteObject 删除对象
func (s *S3ObjectService) DeleteObject(ctx context.Context, bucketName, objectKey, userID string) error {
	// 1. 获取对象元数据
	objectMetadata, err := s.objectMetadataRepo.GetByKey(ctx, bucketName, objectKey, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// S3协议：删除不存在的对象不报错
			return nil
		}
		return err
	}

	// 2. 删除对象元数据
	if err := s.objectMetadataRepo.Delete(ctx, bucketName, objectKey, userID); err != nil {
		logger.LOG.Error("Delete object metadata failed", "error", err)
		return err
	}

	// 3. 检查是否还有其他引用
	// 注意：FileInfo可能被多个对象引用（去重），所以不立即删除物理文件
	// TODO: 实现引用计数或定期清理未引用的文件

	logger.LOG.Info("DeleteObject success",
		"bucket", bucketName,
		"key", objectKey,
		"file_id", objectMetadata.FileID,
	)

	return nil
}

// HeadObject 获取对象元数据（不返回Body）
func (s *S3ObjectService) HeadObject(ctx context.Context, bucketName, objectKey, userID string) (*GetObjectOutput, error) {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, bucketName, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("bucket not found")
		}
		return nil, err
	}

	// 2. 获取对象元数据
	objectMetadata, err := s.objectMetadataRepo.GetByKey(ctx, bucketName, objectKey, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("object not found")
		}
		return nil, err
	}

	// 3. 获取文件信息
	fileInfo, err := s.factory.FileInfo().GetByID(ctx, objectMetadata.FileID)
	if err != nil {
		return nil, fmt.Errorf("file info not found")
	}

	// 4. 解析用户元数据
	userMetadata := make(map[string]string)
	if objectMetadata.UserMetadata != "" {
		metaStr := strings.Trim(objectMetadata.UserMetadata, "{}")
		if metaStr != "" {
			pairs := strings.Split(metaStr, ",")
			for _, pair := range pairs {
				kv := strings.SplitN(pair, ":", 2)
				if len(kv) == 2 {
					key := strings.Trim(kv[0], `"`)
					value := strings.Trim(kv[1], `"`)
					userMetadata[key] = value
				}
			}
		}
	}

	return &GetObjectOutput{
		Body:          nil, // HeadObject不返回Body
		ContentType:   objectMetadata.ContentType,
		ContentLength: int64(fileInfo.Size),
		ETag:          objectMetadata.ETag,
		LastModified:  time.Time(objectMetadata.UpdatedAt),
		VersionID:     objectMetadata.VersionID,
		UserMetadata:  userMetadata,
	}, nil
}

// ListObjects 列出Bucket中的对象
func (s *S3ObjectService) ListObjects(ctx context.Context, bucketName, userID, prefix, marker string, maxKeys int) ([]*models.S3ObjectMetadata, error) {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, bucketName, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("bucket not found")
		}
		return nil, err
	}

	// 2. 列出对象
	objects, err := s.objectMetadataRepo.ListByBucket(ctx, bucketName, userID, prefix, maxKeys, marker)
	if err != nil {
		logger.LOG.Error("List objects failed", "error", err)
		return nil, err
	}

	return objects, nil
}

// copyFile 复制文件（跨分区移动时使用）
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

package service

import (
	"context"
	"crypto/aes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"myobj/src/config"
	"myobj/src/core/service"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/hash"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/preview"
	"myobj/src/pkg/repository"
	"myobj/src/s3_server/auth"
	"myobj/src/s3_server/types"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// getOperationTimeout 获取操作超时时间
func getOperationTimeout() time.Duration {
	timeout := 30 * time.Second // 默认30秒
	if config.CONFIG != nil && config.CONFIG.S3.OperationTimeout > 0 {
		timeout = time.Duration(config.CONFIG.S3.OperationTimeout) * time.Second
	}
	return timeout
}

// decryptedFileReader 包装ReadCloser，在关闭时清理临时解密文件
type decryptedFileReader struct {
	io.ReadCloser
	filePath string
}

func (r *decryptedFileReader) Close() error {
	err := r.ReadCloser.Close()
	// 清理临时解密文件
	if r.filePath != "" {
		if removeErr := os.Remove(r.filePath); removeErr != nil && !os.IsNotExist(removeErr) {
			logger.LOG.Warn("Failed to remove decrypted temp file",
				"path", r.filePath,
				"error", removeErr,
			)
		}
	}
	return err
}

// S3ObjectService S3对象服务
type S3ObjectService struct {
	objectMetadataRepo repository.S3ObjectMetadataRepository
	bucketRepo         repository.S3BucketRepository
	bucketService      *S3BucketService
	fileService        *service.FileService
	factory            *impl.RepositoryFactory
	encryptionService  *EncryptionService
}

// NewS3ObjectService 创建S3对象服务
func NewS3ObjectService(factory *impl.RepositoryFactory, fileService *service.FileService) *S3ObjectService {
	// 从配置读取主密钥，支持环境变量
	masterKey := config.CONFIG.S3.EncryptionMasterKey
	if masterKey == "" {
		// 尝试从环境变量读取
		if envKey := os.Getenv("S3_ENCRYPTION_MASTER_KEY"); envKey != "" {
			masterKey = envKey
		}
	}
	encryptionService := NewEncryptionService(masterKey)
	return &S3ObjectService{
		objectMetadataRepo: factory.S3ObjectMetadata(),
		bucketRepo:         factory.S3Bucket(),
		bucketService:      NewS3BucketService(factory),
		fileService:        fileService,
		factory:            factory,
		encryptionService:  encryptionService,
	}
}

// PutObjectInput PutObject输入参数
type PutObjectInput struct {
	BucketName     string
	ObjectKey      string
	UserID         string
	Body           io.Reader
	ContentType    string
	ContentMD5     string
	UserMetadata   map[string]string
	StorageClass   string
	SSEAlgorithm   string // x-amz-server-side-encryption: AES256, aws:kms
	SSEKMSKeyID    string // x-amz-server-side-encryption-aws-kms-key-id
	SSECustomerKey string // x-amz-server-side-encryption-customer-key (SSE-C)
	SSECustomerMD5 string // x-amz-server-side-encryption-customer-key-MD5 (SSE-C)
	ACL            string // x-amz-acl: private, public-read, public-read-write, authenticated-read, etc.
}

// PutObjectOutput PutObject输出
type PutObjectOutput struct {
	ETag      string
	VersionID string
}

// PutObject 上传对象
func (s *S3ObjectService) PutObject(ctx context.Context, input *PutObjectInput) (*PutObjectOutput, error) {
	// 添加超时控制
	timeout := getOperationTimeout()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 1. 验证Bucket是否存在，如果不存在则自动创建
	bucket, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Bucket不存在，自动创建
			logger.LOG.Info("Bucket not found, auto-creating",
				"bucket_name", input.BucketName,
				"user_id", input.UserID,
			)
			
			// 获取区域（从配置读取）
			region := config.CONFIG.S3.Region
			if region == "" {
				region = "us-east-1"
			}
			
			// 自动创建Bucket
			if err := s.bucketService.CreateBucket(ctx, input.BucketName, input.UserID, region); err != nil {
				// 如果创建失败（例如名称已存在），尝试再次获取
				if err != types.ErrBucketAlreadyExistsError {
					logger.LOG.Error("Auto-create bucket failed",
						"bucket_name", input.BucketName,
						"user_id", input.UserID,
						"error", err,
					)
					return nil, fmt.Errorf("auto-create bucket failed: %w", err)
				}
			}
			
			// 重新获取Bucket
			bucket, err = s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
			if err != nil {
				logger.LOG.Error("Get bucket failed after auto-create",
					"bucket_name", input.BucketName,
					"error", err,
				)
				return nil, fmt.Errorf("get bucket failed after auto-create: %w", err)
			}
			
			logger.LOG.Info("Bucket auto-created successfully",
				"bucket_name", input.BucketName,
				"user_id", input.UserID,
			)
		} else {
			logger.LOG.Error("Get bucket failed",
				"bucket_name", input.BucketName,
				"error", err,
			)
			return nil, err
		}
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
		return nil, types.ErrNoAvailableDiskError
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
		return nil, types.ErrNoAvailableDiskError
	}

	// 4. 创建临时目录并保存文件
	tempDir := filepath.Join(bestDisk.DataPath, "temp", "s3_upload_"+uuid.New().String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		logger.LOG.Error("Create temp dir failed", "error", err, "path", tempDir)
		return nil, fmt.Errorf("create temp dir failed: %w", err)
	}
	// 确保临时目录在函数退出时被清理
	defer func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	}()

	tempFilePath := filepath.Join(tempDir, "upload.tmp")
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("create temp file failed: %w", err)
	}
	defer tempFile.Close()

	// 5. 读取文件内容并计算哈希
	hasher := md5.New()
	multiWriter := io.MultiWriter(tempFile, hasher)
	fileSize, err := io.Copy(multiWriter, input.Body)
	if err != nil {
		return nil, fmt.Errorf("save file failed: %w", err)
	}
	// 确保文件已写入磁盘
	if err := tempFile.Sync(); err != nil {
		return nil, fmt.Errorf("sync temp file failed: %w", err)
	}

	// 计算MD5作为ETag
	etag := hex.EncodeToString(hasher.Sum(nil))

	// 6. 如果提供了Content-MD5，验证完整性
	if input.ContentMD5 != "" && input.ContentMD5 != etag {
		return nil, types.ErrContentMD5MismatchError
	}

	// 7. 检查用户空间（如果不是无限空间）
	if user.Space > 0 && user.FreeSpace < fileSize {
		return nil, types.ErrInsufficientUserSpaceError
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
	var finalPath string

	if err == nil && existingFile != nil {
		// 文件已存在，秒传
		fileInfo = existingFile
		isNewFile = false
		finalPath = existingFile.Path
		logger.LOG.Info("File already exists, instant upload",
			"file_hash", blake3Hash,
			"file_id", fileInfo.ID,
		)
		// 临时文件会在defer中清理
	} else {
		// 新文件，需要移动到最终存储位置
		finalFileName := fmt.Sprintf("%s_%s", blake3Hash, filepath.Base(input.ObjectKey))
		finalPath = filepath.Join(bestDisk.DataPath, "files", finalFileName)

		// 确保目标目录存在
		if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
			return nil, fmt.Errorf("create target dir failed: %w", err)
		}

		// 移动文件
		if err := os.Rename(tempFilePath, finalPath); err != nil {
			// 如果跨分区移动失败，尝试复制
			if err := copyFile(tempFilePath, finalPath); err != nil {
				return nil, fmt.Errorf("move file failed: %w", err)
			}
			os.Remove(tempFilePath)
		}
		// 移动成功后，取消临时目录的清理（文件已不在临时目录）
		tempDir = ""

		// 生成缩略图（如果是图片且配置启用）
		var thumbnailPath string
		if config.CONFIG.File.Thumbnail && isImage(input.ContentType) {
			thumbnailFileName := fmt.Sprintf("%s_%s.jpg", blake3Hash, filepath.Base(input.ObjectKey))
			thumbnailPath = filepath.Join(bestDisk.DataPath, "files", thumbnailFileName)
			
			if err := preview.GenerateImageThumbnail(finalPath, thumbnailPath, 300); err != nil {
				logger.LOG.Warn("Generate thumbnail failed",
					"file", finalPath,
					"error", err,
				)
				// 缩略图生成失败不影响主流程
				thumbnailPath = ""
			} else {
				logger.LOG.Debug("Thumbnail generated successfully",
					"file", finalPath,
					"thumbnail", thumbnailPath,
				)
			}
		}

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
			ThumbnailImg:    thumbnailPath,
			CreatedAt:       custom_type.Now(),
		}

		if err := s.factory.FileInfo().Create(ctx, fileInfo); err != nil {
			os.Remove(finalPath)
			logger.LOG.Error("Create file info failed", "error", err)
			return nil, fmt.Errorf("create file info failed: %w", err)
		}
		// 注意：文件清理在事务失败时处理，这里不需要defer

		// 更新用户空间
		if user.Space > 0 {
			user.FreeSpace -= fileSize
			if err := s.factory.User().Update(ctx, user); err != nil {
				logger.LOG.Error("Update user space failed", "error", err)
				// 回滚：删除已创建的文件和元数据
				s.factory.FileInfo().Delete(ctx, fileInfo.ID)
				os.Remove(fileInfo.Path)
				return nil, fmt.Errorf("update user space failed: %w", err)
			}
		}
	}

	// 10. 处理服务端加密（如果启用）
	var encryptionMetadata *models.S3ObjectEncryption
	if input.SSEAlgorithm != "" {
		encryptionMetadata, err = s.handleServerSideEncryption(ctx, input, fileInfo, isNewFile)
		if err != nil {
			logger.LOG.Error("Server-side encryption failed", "error", err)
			// 如果加密失败，清理文件
			if isNewFile {
				s.factory.FileInfo().Delete(ctx, fileInfo.ID)
				os.Remove(fileInfo.Path)
			}
			return nil, fmt.Errorf("server-side encryption failed: %w", err)
		}
		logger.LOG.Info("Server-side encryption applied",
			"bucket", input.BucketName,
			"key", input.ObjectKey,
			"algorithm", input.SSEAlgorithm)
	}

	// 11. 检查Bucket版本控制状态
	versioningEnabled := bucket.Versioning == "Enabled"

	// 12. 检查是否已存在相同对象元数据
	existingMetadata, err := s.objectMetadataRepo.GetByKey(ctx, input.BucketName, input.ObjectKey, input.UserID)
	if err == nil && existingMetadata != nil {
		if versioningEnabled {
			// 版本控制已启用：标记旧版本为非最新，创建新版本
			if err := s.objectMetadataRepo.MarkOldVersions(ctx, input.BucketName, input.ObjectKey, input.UserID); err != nil {
				logger.LOG.Warn("Mark old versions failed", "error", err)
			}
		} else {
			// 版本控制未启用：删除旧版本，直接覆盖
			if err := s.objectMetadataRepo.Delete(ctx, input.BucketName, input.ObjectKey, input.UserID); err != nil {
				logger.LOG.Warn("Delete old metadata failed", "error", err)
			}
		}
	}

	// 13. 创建对象元数据
	var versionID string
	if versioningEnabled {
		// 版本控制已启用：生成新的版本ID
		versionID = uuid.New().String()
	} else {
		// 版本控制未启用：使用空字符串（表示null版本）
		versionID = ""
	}
	// 序列化用户元数据为JSON
	userMetadataJSON := ""
	if len(input.UserMetadata) > 0 {
		metadataBytes, err := json.Marshal(input.UserMetadata)
		if err != nil {
			return nil, fmt.Errorf("marshal user metadata failed: %w", err)
		}
		userMetadataJSON = string(metadataBytes)
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
		Tags:         "", // 初始为空，可通过PutObjectTagging设置
		VersionID:    versionID,
		IsLatest:     true,
		CreatedAt:    custom_type.Now(),
		UpdatedAt:    custom_type.Now(),
	}

	// 使用事务保护：创建对象元数据、UserFiles关联、加密元数据等
	db := s.factory.DB()
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txFactory := s.factory.WithTx(tx)
		txObjectMetadataRepo := txFactory.S3ObjectMetadata()
		txUserFilesRepo := txFactory.UserFiles()
		txEncryptionRepo := txFactory.S3ObjectEncryption()

		// 创建对象元数据
		if err := txObjectMetadataRepo.Create(ctx, objectMetadata); err != nil {
			logger.LOG.Error("Create object metadata failed", "error", err)
			return fmt.Errorf("create object metadata failed: %w", err)
		}

		// 创建UserFiles关联（将对象关联到Bucket的虚拟路径）
		userFile := &models.UserFiles{
			UserID:      input.UserID,
			FileID:      fileInfo.ID,
			FileName:    filepath.Base(input.ObjectKey),
			VirtualPath: fmt.Sprintf("%d", bucket.VirtualPathID),
			IsPublic:    false,
			CreatedAt:   custom_type.Now(),
			UfID:        uuid.NewString(),
		}

		if err := txUserFilesRepo.Create(ctx, userFile); err != nil {
			logger.LOG.Error("Create user file failed", "error", err)
			return fmt.Errorf("create user file failed: %w", err)
		}

		// 创建加密元数据（如果启用）
		if encryptionMetadata != nil {
			// 设置版本ID（如果版本控制已启用）
			encryptionMetadata.VersionID = objectMetadata.VersionID
			if err := txEncryptionRepo.Create(ctx, encryptionMetadata); err != nil {
				logger.LOG.Error("Create encryption metadata failed", "error", err)
				return fmt.Errorf("create encryption metadata failed: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		// 事务失败，清理文件
		if isNewFile && fileInfo != nil {
			s.factory.FileInfo().Delete(ctx, fileInfo.ID)
			os.Remove(fileInfo.Path)
		}
		return nil, err
	}

	// UserFiles关联已在事务中创建，这里不需要重复创建

	// 如果提供了 ACL header，自动设置对象 ACL
	if input.ACL != "" {
		logger.LOG.Debug("Processing ACL for PutObject",
			"bucket", input.BucketName,
			"key", input.ObjectKey,
			"acl", input.ACL,
		)
		aclPolicy := s.convertCannedACLToPolicy(input.ACL, input.UserID)
		if aclPolicy != nil {
			err = s.PutObjectACL(ctx, &PutObjectACLInput{
				BucketName: input.BucketName,
				ObjectKey:  input.ObjectKey,
				UserID:     input.UserID,
				VersionID:  versionID,
				ACL:        aclPolicy,
			})
			if err != nil {
				logger.LOG.Warn("Set object ACL failed",
					"bucket", input.BucketName,
					"key", input.ObjectKey,
					"acl", input.ACL,
					"error", err,
				)
				// ACL 设置失败不影响上传成功，只记录警告
			} else {
				logger.LOG.Info("Set object ACL success",
					"bucket", input.BucketName,
					"key", input.ObjectKey,
					"acl", input.ACL,
				)
			}
		}
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

// HeadObjectInput HeadObject输入参数（与GetObjectInput相同，但不包含Range）
type HeadObjectInput struct {
	BucketName string
	ObjectKey  string
	UserID     string
	VersionID  string // 可选：版本ID
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
	// 1. 验证Bucket是否存在（如果是公开访问，userID 可能为空）
	var bucket *models.S3Bucket
	var err error
	if input.UserID != "" {
		bucket, err = s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	} else {
		// 公开访问：直接通过数据库查询 Bucket（不限制 userID）
		var buckets []models.S3Bucket
		err = s.factory.DB().WithContext(ctx).Where("bucket_name = ?", input.BucketName).Find(&buckets).Error
		if err == nil && len(buckets) > 0 {
			bucket = &buckets[0]
			input.UserID = bucket.UserID // 设置 userID 以便后续使用
		} else if err == nil {
			err = gorm.ErrRecordNotFound
		}
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrBucketNotFoundError
		}
		return nil, err
	}

	// 2. 获取对象元数据（支持版本ID）
	var objectMetadata *models.S3ObjectMetadata
	if input.VersionID != "" {
		// 通过版本ID获取特定版本
		objectMetadata, err = s.objectMetadataRepo.GetByKeyAndVersion(ctx, input.BucketName, input.ObjectKey, input.VersionID, input.UserID)
	} else {
		// 获取最新版本
		objectMetadata, err = s.objectMetadataRepo.GetByKey(ctx, input.BucketName, input.ObjectKey, input.UserID)
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrObjectNotFoundError
		}
		logger.LOG.Error("Get object metadata failed",
			"bucket", input.BucketName,
			"key", input.ObjectKey,
			"version_id", input.VersionID,
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

	// 4. 检查对象是否加密
	encryptionMetadata, err := s.factory.S3ObjectEncryption().GetByObject(ctx, input.BucketName, input.ObjectKey, input.VersionID, input.UserID)
	var dataKey []byte
	var iv []byte
	if err == nil && encryptionMetadata != nil {
		// 对象已加密，需要解密
		// 获取数据密钥
		if encryptionMetadata.EncryptionType == "SSE-S3" && encryptionMetadata.KeyID != "" {
			// SSE-S3：从数据库获取加密密钥并解密
			encryptionKey, err := s.factory.S3EncryptionKey().GetByKeyID(ctx, encryptionMetadata.KeyID)
			if err != nil {
				return nil, fmt.Errorf("获取加密密钥失败: %w", err)
			}
			dataKey, err = s.encryptionService.DecryptDataKey(encryptionKey.KeyData)
			if err != nil {
				return nil, fmt.Errorf("解密数据密钥失败: %w", err)
			}
		} else if encryptionMetadata.EncryptionType == "SSE-C" {
			// SSE-C：使用客户提供的密钥（需要从请求头获取，这里简化处理）
			return nil, fmt.Errorf("SSE-C not implemented yet")
		}

		// 解码IV
		if encryptionMetadata.IV != "" {
			iv, err = base64.StdEncoding.DecodeString(encryptionMetadata.IV)
			if err != nil {
				return nil, fmt.Errorf("解码IV失败: %w", err)
			}
		}
	}

	// 5. 打开文件
	file, err := os.Open(fileInfo.Path)
	if err != nil {
		logger.LOG.Error("Open file failed",
			"path", fileInfo.Path,
			"error", err,
		)
		return nil, fmt.Errorf("open file failed: %w", err)
	}

	// 如果文件已加密，创建解密流
	var reader io.ReadSeeker = file
	var decryptedPath string                       // 用于记录解密文件路径，以便后续清理
	var contentLength int64 = int64(fileInfo.Size) // 初始文件大小
	if encryptionMetadata != nil && dataKey != nil {
		// 创建临时解密文件（用于Range请求）
		decryptedPath = s.encryptionService.GetDecryptedFilePath(fileInfo.Path)
		if err := s.encryptionService.DecryptFile(fileInfo.Path, decryptedPath, dataKey, iv); err != nil {
			file.Close()
			return nil, fmt.Errorf("解密文件失败: %w", err)
		}
		file.Close()

		// 打开解密后的文件
		decryptedFile, err := os.Open(decryptedPath)
		if err != nil {
			// 清理解密文件
			os.Remove(decryptedPath)
			return nil, fmt.Errorf("打开解密文件失败: %w", err)
		}
		reader = decryptedFile

		// 注意：解密后的文件大小可能不同（因为包含IV），需要重新计算
		decryptedInfo, err := decryptedFile.Stat()
		if err == nil {
			contentLength = decryptedInfo.Size()
		}
	}

	// 6. 处理Range请求
	actualStart := input.RangeStart
	actualEnd := input.RangeEnd

	if input.RangeStart > 0 || input.RangeEnd > 0 {
		// Range请求
		if actualEnd == 0 || actualEnd >= contentLength {
			actualEnd = contentLength - 1
		}

		if actualStart < 0 {
			actualStart = 0
		}

		if actualStart > actualEnd {
			if closer, ok := reader.(io.Closer); ok {
				closer.Close()
			}
			return nil, types.ErrInvalidRangeError
		}

		// Seek到起始位置
		if _, err := reader.Seek(actualStart, 0); err != nil {
			if closer, ok := reader.(io.Closer); ok {
				closer.Close()
			}
			return nil, fmt.Errorf("seek file failed: %w", err)
		}

		contentLength = actualEnd - actualStart + 1
	}

	// 7. 解析用户元数据
	userMetadata := make(map[string]string)
	if objectMetadata.UserMetadata != "" {
		// 使用json.Unmarshal解析用户元数据
		if err := json.Unmarshal([]byte(objectMetadata.UserMetadata), &userMetadata); err != nil {
			logger.LOG.Warn("Unmarshal user metadata failed, using empty metadata", "error", err)
			userMetadata = make(map[string]string)
		}
	}

	logger.LOG.Info("GetObject success",
		"bucket", input.BucketName,
		"key", input.ObjectKey,
		"size", contentLength,
		"range", fmt.Sprintf("%d-%d", actualStart, actualEnd),
	)

	// 如果使用了解密文件，需要返回可关闭的Reader，并在关闭时清理临时文件
	var body io.ReadCloser
	if decryptedPath != "" {
		// 创建包装的ReadCloser，在关闭时清理临时文件
		if readCloser, ok := reader.(io.ReadCloser); ok {
			body = &decryptedFileReader{
				ReadCloser: readCloser,
				filePath:   decryptedPath,
			}
		} else {
			// 如果reader不是ReadCloser，包装它
			body = &decryptedFileReader{
				ReadCloser: io.NopCloser(reader),
				filePath:   decryptedPath,
			}
		}
	} else {
		// 未加密文件，直接包装
		if readCloser, ok := reader.(io.ReadCloser); ok {
			body = readCloser
		} else {
			body = io.NopCloser(reader)
		}
	}

	return &GetObjectOutput{
		Body:          body,
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

// DeleteObject 删除对象（支持版本ID）
func (s *S3ObjectService) DeleteObject(ctx context.Context, bucketName, objectKey, userID, versionID string) error {
	// 1. 获取Bucket信息（检查版本控制状态）
	bucket, err := s.bucketRepo.GetByName(ctx, bucketName, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// S3协议：删除不存在的Bucket中的对象不报错
			return nil
		}
		return err
	}

	// 2. 获取对象元数据（支持版本ID）
	var objectMetadata *models.S3ObjectMetadata
	if versionID != "" {
		// 通过版本ID获取特定版本
		objectMetadata, err = s.objectMetadataRepo.GetByKeyAndVersion(ctx, bucketName, objectKey, versionID, userID)
	} else {
		// 获取最新版本
		objectMetadata, err = s.objectMetadataRepo.GetByKey(ctx, bucketName, objectKey, userID)
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// S3协议：删除不存在的对象不报错
			return nil
		}
		return err
	}

	fileID := objectMetadata.FileID

	// 3. 检查版本控制状态
	versioningEnabled := bucket.Versioning == "Enabled"

	if versioningEnabled && versionID == "" {
		// 版本控制已启用且未指定版本ID：创建DeleteMarker（软删除）
		// 标记旧版本为非最新，但不删除元数据
		if err := s.objectMetadataRepo.MarkOldVersions(ctx, bucketName, objectKey, userID); err != nil {
			logger.LOG.Warn("Mark old versions failed", "error", err)
		}
		logger.LOG.Info("DeleteObject: Created delete marker (versioning enabled)",
			"bucket", bucketName,
			"key", objectKey,
		)
		// 不删除文件，只标记为非最新
		return nil
	}

	// 4. 删除对象元数据（指定版本或版本控制未启用）
	if versionID != "" {
		// 删除特定版本
		if err := s.objectMetadataRepo.DeleteByVersion(ctx, bucketName, objectKey, versionID, userID); err != nil {
			logger.LOG.Error("Delete object metadata by version failed", "error", err)
			return err
		}
	} else {
		// 删除最新版本
		if err := s.objectMetadataRepo.Delete(ctx, bucketName, objectKey, userID); err != nil {
			logger.LOG.Error("Delete object metadata failed", "error", err)
			return err
		}
	}

	// 5. 删除 UserFiles 关联（仅在版本控制未启用或删除所有版本时）
	// 如果版本控制已启用且只删除特定版本，不删除 UserFiles（其他版本可能还在使用）
	if !versioningEnabled || versionID == "" {
		// 查找对应的 UserFiles 记录
		bucket, err := s.bucketRepo.GetByName(ctx, bucketName, userID)
		if err == nil {
			virtualPathID := fmt.Sprintf("%d", bucket.VirtualPathID)
			// 查找该虚拟路径下的所有 UserFiles 记录
			userFiles, err := s.factory.UserFiles().ListByVirtualPath(ctx, userID, virtualPathID, 0, 10000)
			if err == nil {
				// 查找匹配的 UserFiles 记录（通过 file_id 和文件名匹配）
				objectFileName := filepath.Base(objectKey)
				for _, uf := range userFiles {
					if uf.FileID == fileID && uf.FileName == objectFileName {
						// 删除 UserFiles 记录
						if err := s.factory.UserFiles().Delete(ctx, userID, uf.UfID); err != nil {
							logger.LOG.Warn("Delete user file failed", "uf_id", uf.UfID, "error", err)
						}
						break
					}
				}
			}
		}
	}

	// 6. 检查文件引用计数并清理（仅在版本控制未启用或删除所有版本时）
	// 如果版本控制已启用且只删除特定版本，不清理文件（其他版本可能还在使用）
	if !versioningEnabled || versionID == "" {
		if err := s.cleanupFileIfUnreferenced(ctx, fileID); err != nil {
			logger.LOG.Warn("Cleanup file failed", "file_id", fileID, "error", err)
			// 不返回错误，因为对象已删除成功
		}
	}

	logger.LOG.Info("DeleteObject success",
		"bucket", bucketName,
		"key", objectKey,
		"file_id", fileID,
	)

	return nil
}

// cleanupFileIfUnreferenced 检查文件引用计数，如果无引用则清理物理文件
func (s *S3ObjectService) cleanupFileIfUnreferenced(ctx context.Context, fileID string) error {
	// 1. 统计 S3 对象引用
	s3RefCount, err := s.objectMetadataRepo.CountByFileID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("count S3 object references failed: %w", err)
	}

	// 2. 统计 UserFiles 引用（直接统计 UserFiles 表中的引用）
	var userFilesRefCount int64
	err = s.factory.DB().WithContext(ctx).Model(&models.UserFiles{}).
		Where("file_id = ?", fileID).
		Count(&userFilesRefCount).Error
	if err != nil {
		return fmt.Errorf("count UserFiles references failed: %w", err)
	}

	totalRefCount := s3RefCount + userFilesRefCount

	// 3. 如果还有引用，不删除物理文件
	if totalRefCount > 0 {
		logger.LOG.Debug("File still has references, skip cleanup",
			"file_id", fileID,
			"s3_ref_count", s3RefCount,
			"user_files_ref_count", userFilesRefCount,
			"total_ref_count", totalRefCount)
		return nil
	}

	// 4. 无引用，删除物理文件
	fileInfo, err := s.factory.FileInfo().GetByID(ctx, fileID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 文件信息已不存在，无需清理
			return nil
		}
		return fmt.Errorf("get file info failed: %w", err)
	}

	// 5. 删除物理文件
	if fileInfo.Path != "" {
		if err := os.Remove(fileInfo.Path); err != nil && !os.IsNotExist(err) {
			logger.LOG.Warn("Remove physical file failed", "path", fileInfo.Path, "error", err)
		} else {
			logger.LOG.Info("Cleaned up unreferenced file", "file_id", fileID, "path", fileInfo.Path)
		}
	}

	// 6. 删除缩略图
	if fileInfo.ThumbnailImg != "" {
		os.Remove(fileInfo.ThumbnailImg)
	}

	// 7. 删除 FileInfo 记录
	if err := s.factory.FileInfo().Delete(ctx, fileID); err != nil {
		logger.LOG.Warn("Delete file info failed", "file_id", fileID, "error", err)
	}

	return nil
}

// HeadObject 获取对象元数据（不返回Body）
func (s *S3ObjectService) HeadObject(ctx context.Context, input *HeadObjectInput) (*GetObjectOutput, error) {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrBucketNotFoundError
		}
		return nil, err
	}

	// 2. 获取对象元数据（支持版本ID）
	var objectMetadata *models.S3ObjectMetadata
	if input.VersionID != "" {
		// 通过版本ID获取特定版本
		objectMetadata, err = s.objectMetadataRepo.GetByKeyAndVersion(ctx, input.BucketName, input.ObjectKey, input.VersionID, input.UserID)
	} else {
		// 获取最新版本
		objectMetadata, err = s.objectMetadataRepo.GetByKey(ctx, input.BucketName, input.ObjectKey, input.UserID)
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrObjectNotFoundError
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

// ListObjectsInput ListObjects输入参数
type ListObjectsInput struct {
	BucketName string
	UserID     string
	Prefix     string
	Delimiter  string
	Marker     string
	MaxKeys    int
}

// ListObjectsOutput ListObjects输出
type ListObjectsOutput struct {
	Objects        []*models.S3ObjectMetadata
	CommonPrefixes []string // 公共前缀（目录）
	IsTruncated    bool
	NextMarker     string
}

// ListObjects 列出Bucket中的对象
func (s *S3ObjectService) ListObjects(ctx context.Context, input *ListObjectsInput) (*ListObjectsOutput, error) {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrBucketNotFoundError
		}
		return nil, err
	}

	// 2. 设置默认值
	maxKeys := input.MaxKeys
	if maxKeys <= 0 || maxKeys > 1000 {
		maxKeys = 1000
	}

	// 3. 列出对象（查询比 maxKeys 多一个，用于判断是否还有更多）
	objects, err := s.objectMetadataRepo.ListByBucket(ctx, input.BucketName, input.UserID, input.Prefix, maxKeys+1, input.Marker)
	if err != nil {
		logger.LOG.Error("List objects failed", "error", err)
		return nil, err
	}

	// 4. 处理 delimiter（目录分隔符）
	var resultObjects []*models.S3ObjectMetadata
	var commonPrefixes []string
	prefixSet := make(map[string]bool)

	isTruncated := false
	nextMarker := ""

	if input.Delimiter != "" {
		// 有 delimiter，需要区分文件和目录
		for _, obj := range objects {
			// 检查是否超过 maxKeys
			if len(resultObjects)+len(commonPrefixes) >= maxKeys {
				isTruncated = true
				nextMarker = obj.ObjectKey
				break
			}

			// 移除 prefix 部分
			relativeKey := obj.ObjectKey
			if input.Prefix != "" {
				if !strings.HasPrefix(obj.ObjectKey, input.Prefix) {
					continue
				}
				relativeKey = obj.ObjectKey[len(input.Prefix):]
			}

			// 检查是否包含 delimiter
			delimiterIndex := strings.Index(relativeKey, input.Delimiter)
			if delimiterIndex >= 0 {
				// 这是一个目录前缀
				commonPrefix := input.Prefix + relativeKey[:delimiterIndex+len(input.Delimiter)]
				if !prefixSet[commonPrefix] {
					commonPrefixes = append(commonPrefixes, commonPrefix)
					prefixSet[commonPrefix] = true
				}
			} else {
				// 这是一个文件
				resultObjects = append(resultObjects, obj)
			}
		}
	} else {
		// 没有 delimiter，直接返回所有对象
		if len(objects) > maxKeys {
			isTruncated = true
			nextMarker = objects[maxKeys].ObjectKey
			resultObjects = objects[:maxKeys]
		} else {
			resultObjects = objects
		}
	}

	return &ListObjectsOutput{
		Objects:        resultObjects,
		CommonPrefixes: commonPrefixes,
		IsTruncated:    isTruncated,
		NextMarker:     nextMarker,
	}, nil
}

// ==================== ListObjectVersions 相关方法 ====================

// ListObjectVersionsInput ListObjectVersions输入参数
type ListObjectVersionsInput struct {
	BucketName      string
	UserID          string
	Prefix          string
	Delimiter       string
	KeyMarker       string
	VersionIDMarker string
	MaxKeys         int
}

// ObjectVersion 对象版本信息
type ObjectVersion struct {
	Key          string
	VersionID    string
	IsLatest     bool
	LastModified time.Time
	ETag         string
	Size         int64
	StorageClass string
	Owner        string
}

// DeleteMarker 删除标记
type DeleteMarker struct {
	Key          string
	VersionID    string
	IsLatest     bool
	LastModified time.Time
	Owner        string
}

// ListObjectVersionsOutput ListObjectVersions输出
type ListObjectVersionsOutput struct {
	Versions            []ObjectVersion
	DeleteMarkers       []DeleteMarker
	CommonPrefixes      []string
	IsTruncated         bool
	NextKeyMarker       string
	NextVersionIDMarker string
}

// ListObjectVersions 列出对象的所有版本
func (s *S3ObjectService) ListObjectVersions(ctx context.Context, input *ListObjectVersionsInput) (*ListObjectVersionsOutput, error) {
	// 1. 验证Bucket是否存在
	bucket, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrBucketNotFoundError
		}
		return nil, err
	}

	// 2. 检查版本控制是否启用
	if bucket.Versioning != "Enabled" && bucket.Versioning != "Suspended" {
		// 版本控制未启用，返回空结果
		return &ListObjectVersionsOutput{
			Versions:       []ObjectVersion{},
			DeleteMarkers:  []DeleteMarker{},
			CommonPrefixes: []string{},
			IsTruncated:    false,
		}, nil
	}

	// 3. 设置默认值
	maxKeys := input.MaxKeys
	if maxKeys <= 0 || maxKeys > 1000 {
		maxKeys = 1000
	}

	// 4. 列出所有版本（包括旧版本）
	objects, err := s.objectMetadataRepo.ListVersionsByBucket(ctx, input.BucketName, input.UserID, input.Prefix, input.KeyMarker, input.VersionIDMarker, maxKeys+1)
	if err != nil {
		logger.LOG.Error("List object versions failed", "error", err)
		return nil, err
	}

	// 5. 处理结果
	var versions []ObjectVersion
	var deleteMarkers []DeleteMarker
	var commonPrefixes []string
	prefixSet := make(map[string]bool)

	isTruncated := false
	nextKeyMarker := ""
	nextVersionIDMarker := ""

	// 检查是否超过 maxKeys
	if len(objects) > maxKeys {
		isTruncated = true
		nextKeyMarker = objects[maxKeys].ObjectKey
		nextVersionIDMarker = objects[maxKeys].VersionID
		objects = objects[:maxKeys]
	}

	// 处理 delimiter
	if input.Delimiter != "" {
		for _, obj := range objects {
			// 移除 prefix 部分
			relativeKey := obj.ObjectKey
			if input.Prefix != "" {
				if !strings.HasPrefix(obj.ObjectKey, input.Prefix) {
					continue
				}
				relativeKey = obj.ObjectKey[len(input.Prefix):]
			}

			// 检查是否包含 delimiter
			delimiterIndex := strings.Index(relativeKey, input.Delimiter)
			if delimiterIndex >= 0 {
				// 这是一个目录前缀
				commonPrefix := input.Prefix + relativeKey[:delimiterIndex+len(input.Delimiter)]
				if !prefixSet[commonPrefix] {
					commonPrefixes = append(commonPrefixes, commonPrefix)
					prefixSet[commonPrefix] = true
				}
			} else {
				// 这是一个文件版本或DeleteMarker
				if obj.IsDeleteMarker {
					// DeleteMarker
					deleteMarkers = append(deleteMarkers, DeleteMarker{
						Key:          obj.ObjectKey,
						VersionID:    obj.VersionID,
						IsLatest:     obj.IsLatest,
						LastModified: time.Time(obj.UpdatedAt),
						Owner:        obj.UserID,
					})
				} else {
					// 正常版本
					fileInfo, err := s.factory.FileInfo().GetByID(ctx, obj.FileID)
					if err == nil && fileInfo != nil {
						versions = append(versions, ObjectVersion{
							Key:          obj.ObjectKey,
							VersionID:    obj.VersionID,
							IsLatest:     obj.IsLatest,
							LastModified: time.Time(obj.UpdatedAt),
							ETag:         obj.ETag,
							Size:         int64(fileInfo.Size),
							StorageClass: obj.StorageClass,
							Owner:        obj.UserID,
						})
					}
				}
			}
		}
	} else {
		// 没有 delimiter，直接返回所有版本
		for _, obj := range objects {
			if obj.IsDeleteMarker {
				// DeleteMarker
				deleteMarkers = append(deleteMarkers, DeleteMarker{
					Key:          obj.ObjectKey,
					VersionID:    obj.VersionID,
					IsLatest:     obj.IsLatest,
					LastModified: time.Time(obj.UpdatedAt),
					Owner:        obj.UserID,
				})
			} else {
				// 正常版本
				fileInfo, err := s.factory.FileInfo().GetByID(ctx, obj.FileID)
				if err == nil && fileInfo != nil {
					versions = append(versions, ObjectVersion{
						Key:          obj.ObjectKey,
						VersionID:    obj.VersionID,
						IsLatest:     obj.IsLatest,
						LastModified: time.Time(obj.UpdatedAt),
						ETag:         obj.ETag,
						Size:         int64(fileInfo.Size),
						StorageClass: obj.StorageClass,
						Owner:        obj.UserID,
					})
				}
			}
		}
	}

	logger.LOG.Info("List object versions success",
		"bucket", input.BucketName,
		"user_id", input.UserID,
		"versions_count", len(versions),
		"delete_markers_count", len(deleteMarkers),
	)

	return &ListObjectVersionsOutput{
		Versions:            versions,
		DeleteMarkers:       deleteMarkers,
		CommonPrefixes:      commonPrefixes,
		IsTruncated:         isTruncated,
		NextKeyMarker:       nextKeyMarker,
		NextVersionIDMarker: nextVersionIDMarker,
	}, nil
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

// ==================== Multipart Upload 相关方法 ====================

// InitiateMultipartUploadInput 初始化分片上传输入参数
type InitiateMultipartUploadInput struct {
	BucketName   string
	ObjectKey    string
	UserID       string
	ContentType  string
	UserMetadata map[string]string
	StorageClass string
}

// InitiateMultipartUploadOutput 初始化分片上传输出
type InitiateMultipartUploadOutput struct {
	BucketName string
	ObjectKey  string
	UploadID   string
}

// InitiateMultipartUpload 初始化分片上传
func (s *S3ObjectService) InitiateMultipartUpload(ctx context.Context, input *InitiateMultipartUploadInput) (*InitiateMultipartUploadOutput, error) {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrBucketNotFoundError
		}
		return nil, err
	}

	// 2. 生成 UploadID
	uploadID := uuid.New().String()

	// 3. 序列化元数据
	metadataJSON := ""
	if len(input.UserMetadata) > 0 {
		metadataBytes, err := json.Marshal(input.UserMetadata)
		if err != nil {
			return nil, fmt.Errorf("marshal user metadata failed: %w", err)
		}
		metadataJSON = string(metadataBytes)
	}

	storageClass := input.StorageClass
	if storageClass == "" {
		storageClass = "STANDARD"
	}

	// 4. 创建分片上传会话
	upload := &models.S3MultipartUpload{
		UploadID:   uploadID,
		BucketName: input.BucketName,
		ObjectKey:  input.ObjectKey,
		UserID:     input.UserID,
		Metadata:   metadataJSON,
		Status:     "in-progress",
		CreatedAt:  custom_type.Now(),
		UpdatedAt:  custom_type.Now(),
	}

	multipartRepo := s.factory.S3Multipart()
	if err := multipartRepo.CreateUpload(ctx, upload); err != nil {
		logger.LOG.Error("Create multipart upload failed", "error", err)
		return nil, fmt.Errorf("create multipart upload failed: %w", err)
	}

	logger.LOG.Info("Initiate multipart upload success",
		"bucket", input.BucketName,
		"key", input.ObjectKey,
		"upload_id", uploadID,
	)

	return &InitiateMultipartUploadOutput{
		BucketName: input.BucketName,
		ObjectKey:  input.ObjectKey,
		UploadID:   uploadID,
	}, nil
}

// UploadPartInput 上传分片输入参数
type UploadPartInput struct {
	BucketName string
	ObjectKey  string
	UploadID   string
	PartNumber int
	Body       io.Reader
	UserID     string
}

// UploadPartOutput 上传分片输出
type UploadPartOutput struct {
	ETag string
}

// UploadPart 上传分片
func (s *S3ObjectService) UploadPart(ctx context.Context, input *UploadPartInput) (*UploadPartOutput, error) {
	// 添加超时控制（分片上传可能需要更长时间）
	timeout := getOperationTimeout() * 2 // 分片上传允许更长时间
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 1. 验证上传会话是否存在
	multipartRepo := s.factory.S3Multipart()
	upload, err := multipartRepo.GetUpload(ctx, input.UploadID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrUploadNotFoundError
		}
		return nil, err
	}

	// 2. 验证用户权限
	if upload.UserID != input.UserID {
		return nil, types.ErrAccessDeniedError
	}

	// 3. 验证 Bucket 和 Key 是否匹配
	if upload.BucketName != input.BucketName || upload.ObjectKey != input.ObjectKey {
		return nil, fmt.Errorf("bucket or key mismatch")
	}

	// 4. 验证状态
	if upload.Status != "in-progress" {
		return nil, fmt.Errorf("upload is not in progress")
	}

	// 5. 验证 PartNumber（1-10000）
	if input.PartNumber < 1 || input.PartNumber > 10000 {
		return nil, fmt.Errorf("invalid part number")
	}

	// 6. 检查分片是否已存在
	existingPart, err := multipartRepo.GetPart(ctx, input.UploadID, input.PartNumber)
	if err == nil && existingPart != nil {
		// 分片已存在，删除旧分片
		if err := multipartRepo.DeletePart(ctx, existingPart.ID); err != nil {
			logger.LOG.Warn("Delete existing part failed", "error", err)
		}
		// 删除旧分片文件
		if existingPart.ChunkPath != "" {
			os.Remove(existingPart.ChunkPath)
		}
	}

	// 7. 选择存储磁盘
	disks, err := s.factory.Disk().List(ctx, 0, 1000)
	if err != nil || len(disks) == 0 {
		return nil, types.ErrNoAvailableDiskError
	}

	var bestDisk *models.Disk
	var maxFreeSpace int64 = -1
	for _, disk := range disks {
		freeSpaceBytes := int64(disk.Size) * 1024 * 1024 * 1024
		if freeSpaceBytes > maxFreeSpace {
			maxFreeSpace = freeSpaceBytes
			bestDisk = disk
		}
	}

	// 8. 保存分片到临时文件
	tempDir := filepath.Join(bestDisk.DataPath, "temp", "s3_multipart_"+uuid.New().String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("create temp dir failed: %w", err)
	}
	// 确保临时目录在失败时被清理
	defer func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	}()

	chunkPath := filepath.Join(tempDir, fmt.Sprintf("part_%d", input.PartNumber))
	chunkFile, err := os.Create(chunkPath)
	if err != nil {
		return nil, fmt.Errorf("create chunk file failed: %w", err)
	}
	defer func() {
		if err := chunkFile.Close(); err != nil {
			logger.LOG.Warn("Close chunk file failed", "error", err)
		}
	}()

	// 9. 计算 MD5
	hasher := md5.New()
	multiWriter := io.MultiWriter(chunkFile, hasher)

	written, err := io.Copy(multiWriter, input.Body)
	if err != nil {
		return nil, fmt.Errorf("write chunk failed: %w", err)
	}
	// 确保数据已写入磁盘
	if err := chunkFile.Sync(); err != nil {
		return nil, fmt.Errorf("sync chunk file failed: %w", err)
	}

	etag := hex.EncodeToString(hasher.Sum(nil))

	// 10. 保存分片信息
	part := &models.S3MultipartPart{
		UploadID:   input.UploadID,
		PartNumber: input.PartNumber,
		ETag:       etag,
		Size:       written,
		ChunkPath:  chunkPath,
		CreatedAt:  custom_type.Now(),
	}

	if err := multipartRepo.CreatePart(ctx, part); err != nil {
		return nil, fmt.Errorf("create part record failed: %w", err)
	}
	// 分片保存成功，取消临时目录清理（分片需要保留直到完成上传）
	tempDir = ""

	logger.LOG.Info("Upload part success",
		"upload_id", input.UploadID,
		"part_number", input.PartNumber,
		"size", written,
		"etag", etag,
	)

	return &UploadPartOutput{
		ETag: etag,
	}, nil
}

// CompleteMultipartUploadInput 完成分片上传输入参数
type CompleteMultipartUploadInput struct {
	BucketName string
	ObjectKey  string
	UploadID   string
	Parts      []PartInfo // 分片信息（PartNumber 和 ETag）
	UserID     string
}

// PartInfo 分片信息
type PartInfo struct {
	PartNumber int
	ETag       string
}

// CompleteMultipartUploadOutput 完成分片上传输出
type CompleteMultipartUploadOutput struct {
	Location   string
	BucketName string
	ObjectKey  string
	ETag       string
}

// CompleteMultipartUpload 完成分片上传
func (s *S3ObjectService) CompleteMultipartUpload(ctx context.Context, input *CompleteMultipartUploadInput) (*CompleteMultipartUploadOutput, error) {
	// 1. 验证上传会话
	multipartRepo := s.factory.S3Multipart()
	upload, err := multipartRepo.GetUpload(ctx, input.UploadID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrUploadNotFoundError
		}
		return nil, err
	}

	// 2. 验证用户权限
	if upload.UserID != input.UserID {
		return nil, types.ErrAccessDeniedError
	}

	// 3. 验证状态
	if upload.Status != "in-progress" {
		return nil, fmt.Errorf("upload is not in progress")
	}

	// 4. 验证分片数量
	if len(input.Parts) == 0 {
		return nil, types.ErrNoPartsProvidedError
	}

	// 5. 验证分片顺序（必须按 PartNumber 升序）
	for i := 1; i < len(input.Parts); i++ {
		if input.Parts[i].PartNumber <= input.Parts[i-1].PartNumber {
			return nil, fmt.Errorf("parts must be in ascending order")
		}
	}

	// 6. 获取所有分片
	allParts, err := multipartRepo.ListParts(ctx, input.UploadID)
	if err != nil {
		return nil, fmt.Errorf("list parts failed: %w", err)
	}

	// 7. 验证所有分片都存在且 ETag 匹配
	partMap := make(map[int]*models.S3MultipartPart)
	for _, part := range allParts {
		partMap[part.PartNumber] = part
	}

	for _, reqPart := range input.Parts {
		dbPart, exists := partMap[reqPart.PartNumber]
		if !exists {
			return nil, fmt.Errorf("part %d not found", reqPart.PartNumber)
		}
		if dbPart.ETag != reqPart.ETag {
			return nil, fmt.Errorf("part %d etag mismatch", reqPart.PartNumber)
		}
	}

	// 8. 选择存储磁盘
	disks, err := s.factory.Disk().List(ctx, 0, 1000)
	if err != nil || len(disks) == 0 {
		return nil, types.ErrNoAvailableDiskError
	}

	var bestDisk *models.Disk
	var maxFreeSpace int64 = -1
	for _, disk := range disks {
		freeSpaceBytes := int64(disk.Size) * 1024 * 1024 * 1024
		if freeSpaceBytes > maxFreeSpace {
			maxFreeSpace = freeSpaceBytes
			bestDisk = disk
		}
	}

	// 创建最终文件路径
	bucket, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("bucket not found")
	}

	finalDir := filepath.Join(bestDisk.DataPath, "files", input.UserID)
	if err := os.MkdirAll(finalDir, 0755); err != nil {
		return nil, fmt.Errorf("create final dir failed: %w", err)
	}

	// 计算文件总大小
	var totalSize int64
	for _, part := range allParts {
		if _, exists := partMap[part.PartNumber]; exists {
			totalSize += part.Size
		}
	}

	// 合并文件
	finalPath := filepath.Join(finalDir, uuid.New().String())

	// 创建最终文件
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return nil, fmt.Errorf("create final file failed: %w", err)
	}
	defer func() {
		if err := finalFile.Close(); err != nil {
			logger.LOG.Warn("Close final file failed", "error", err)
		}
	}()

	// 按 PartNumber 顺序合并
	hasher := md5.New()
	multiWriter := io.MultiWriter(finalFile, hasher)

	for _, reqPart := range input.Parts {
		dbPart := partMap[reqPart.PartNumber]
		partFile, err := os.Open(dbPart.ChunkPath)
		if err != nil {
			os.Remove(finalPath)
			return nil, fmt.Errorf("open part file failed: %w", err)
		}

		_, err = io.Copy(multiWriter, partFile)
		if err != nil {
			partFile.Close()
			os.Remove(finalPath)
			return nil, fmt.Errorf("copy part failed: %w", err)
		}
		if err := partFile.Close(); err != nil {
			logger.LOG.Warn("Close part file failed", "error", err)
		}
	}

	// 确保数据已写入磁盘
	if err := finalFile.Sync(); err != nil {
		os.Remove(finalPath)
		return nil, fmt.Errorf("sync final file failed: %w", err)
	}
	finalETag := hex.EncodeToString(hasher.Sum(nil))

	// 9. 创建 FileInfo
	// 从上传会话的元数据中获取 ContentType
	contentType := "application/octet-stream"
	if upload.Metadata != "" {
		var metadata map[string]string
		if err := json.Unmarshal([]byte(upload.Metadata), &metadata); err == nil {
			if ct, ok := metadata["Content-Type"]; ok {
				contentType = ct
			}
		}
	}

	fileInfo := &models.FileInfo{
		ID:             uuid.NewString(),
		Name:           filepath.Base(input.ObjectKey),
		RandomName:     filepath.Base(finalPath),
		Path:           finalPath,
		Size:           int(totalSize),
		Mime:           contentType,
		FileHash:       finalETag,
		ChunkSignature: finalETag[:16], // 使用前16位作为签名
		IsEnc:          false,
		IsChunk:        false,
		ThumbnailImg:   "",
		CreatedAt:      custom_type.Now(),
		UpdatedAt:      custom_type.Now(),
	}

	if err := s.factory.FileInfo().Create(ctx, fileInfo); err != nil {
		os.Remove(finalPath)
		return nil, fmt.Errorf("create file info failed: %w", err)
	}
	// 确保文件在失败时被清理
	defer func() {
		if err != nil {
			s.factory.FileInfo().Delete(ctx, fileInfo.ID)
			os.Remove(finalPath)
		}
	}()

	// 10. 创建对象元数据
	versionID := uuid.New().String()
	objectMetadata := &models.S3ObjectMetadata{
		FileID:       fileInfo.ID,
		BucketName:   input.BucketName,
		ObjectKey:    input.ObjectKey,
		UserID:       input.UserID,
		ETag:         finalETag,
		StorageClass: "STANDARD",
		ContentType:  "application/octet-stream",
		UserMetadata: upload.Metadata,
		VersionID:    versionID,
		IsLatest:     true,
		CreatedAt:    custom_type.Now(),
		UpdatedAt:    custom_type.Now(),
	}

	if err := s.objectMetadataRepo.Create(ctx, objectMetadata); err != nil {
		s.factory.FileInfo().Delete(ctx, fileInfo.ID)
		os.Remove(finalPath)
		return nil, fmt.Errorf("create object metadata failed: %w", err)
	}

	// 11. 创建 UserFiles 关联
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
		logger.LOG.Warn("Create user file failed", "error", err)
	}

	// 12. 更新上传状态
	if err := multipartRepo.UpdateUploadStatus(ctx, input.UploadID, "completed"); err != nil {
		logger.LOG.Warn("Update upload status failed", "error", err)
	}

	// 13. 清理临时分片文件
	for _, part := range allParts {
		if part.ChunkPath != "" {
			os.Remove(part.ChunkPath)
			// 尝试删除父目录
			partDir := filepath.Dir(part.ChunkPath)
			os.Remove(partDir)
		}
	}

	logger.LOG.Info("Complete multipart upload success",
		"bucket", input.BucketName,
		"key", input.ObjectKey,
		"upload_id", input.UploadID,
		"total_size", totalSize,
		"etag", finalETag,
	)

	return &CompleteMultipartUploadOutput{
		Location:   fmt.Sprintf("/%s/%s", input.BucketName, input.ObjectKey),
		BucketName: input.BucketName,
		ObjectKey:  input.ObjectKey,
		ETag:       finalETag,
	}, nil
}

// AbortMultipartUpload 取消分片上传
func (s *S3ObjectService) AbortMultipartUpload(ctx context.Context, bucketName, objectKey, uploadID, userID string) error {
	// 1. 验证上传会话
	multipartRepo := s.factory.S3Multipart()
	upload, err := multipartRepo.GetUpload(ctx, uploadID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return types.ErrUploadNotFoundError
		}
		return err
	}

	// 2. 验证用户权限
	if upload.UserID != userID {
		return types.ErrAccessDeniedError
	}

	// 3. 获取所有分片并删除文件
	parts, err := multipartRepo.ListParts(ctx, uploadID)
	if err == nil {
		for _, part := range parts {
			if part.ChunkPath != "" {
				os.Remove(part.ChunkPath)
				// 尝试删除父目录
				partDir := filepath.Dir(part.ChunkPath)
				os.Remove(partDir)
			}
		}
		// 删除分片记录
		multipartRepo.DeletePartsByUploadID(ctx, uploadID)
	}

	// 4. 更新状态并删除上传会话
	multipartRepo.UpdateUploadStatus(ctx, uploadID, "aborted")
	multipartRepo.DeleteUpload(ctx, uploadID)

	logger.LOG.Info("Abort multipart upload success",
		"bucket", bucketName,
		"key", objectKey,
		"upload_id", uploadID,
	)

	return nil
}

// ListPartsInput 列出分片输入参数
type ListPartsInput struct {
	BucketName       string
	ObjectKey        string
	UploadID         string
	UserID           string
	MaxParts         int
	PartNumberMarker int
}

// ListPartsOutput 列出分片输出
type ListPartsOutput struct {
	BucketName           string
	ObjectKey            string
	UploadID             string
	PartNumberMarker     int
	NextPartNumberMarker int
	MaxParts             int
	IsTruncated          bool
	Parts                []*models.S3MultipartPart
}

// ListParts 列出分片
func (s *S3ObjectService) ListParts(ctx context.Context, input *ListPartsInput) (*ListPartsOutput, error) {
	// 1. 验证上传会话
	multipartRepo := s.factory.S3Multipart()
	upload, err := multipartRepo.GetUpload(ctx, input.UploadID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrUploadNotFoundError
		}
		return nil, err
	}

	// 2. 验证用户权限
	if upload.UserID != input.UserID {
		return nil, types.ErrAccessDeniedError
	}

	// 3. 获取所有分片
	allParts, err := multipartRepo.ListParts(ctx, input.UploadID)
	if err != nil {
		return nil, fmt.Errorf("list parts failed: %w", err)
	}

	// 4. 过滤和分页
	maxParts := input.MaxParts
	if maxParts <= 0 || maxParts > 1000 {
		maxParts = 1000
	}

	var parts []*models.S3MultipartPart
	startIndex := 0
	for i, part := range allParts {
		if part.PartNumber > input.PartNumberMarker {
			startIndex = i
			break
		}
	}

	endIndex := startIndex + maxParts
	if endIndex > len(allParts) {
		endIndex = len(allParts)
	}

	parts = allParts[startIndex:endIndex]
	isTruncated := endIndex < len(allParts)
	nextPartNumberMarker := 0
	if isTruncated && len(parts) > 0 {
		nextPartNumberMarker = parts[len(parts)-1].PartNumber
	}

	return &ListPartsOutput{
		BucketName:           input.BucketName,
		ObjectKey:            input.ObjectKey,
		UploadID:             input.UploadID,
		PartNumberMarker:     input.PartNumberMarker,
		NextPartNumberMarker: nextPartNumberMarker,
		MaxParts:             maxParts,
		IsTruncated:          isTruncated,
		Parts:                parts,
	}, nil
}

// ListMultipartUploadsInput 列出分片上传会话输入参数
type ListMultipartUploadsInput struct {
	BucketName     string
	UserID         string
	Prefix         string
	Delimiter      string
	KeyMarker      string
	UploadIDMarker string
	MaxUploads     int
}

// ListMultipartUploadsOutput 列出分片上传会话输出
type ListMultipartUploadsOutput struct {
	BucketName         string
	Prefix             string
	Delimiter          string
	KeyMarker          string
	UploadIDMarker     string
	NextKeyMarker      string
	NextUploadIDMarker string
	MaxUploads         int
	IsTruncated        bool
	Uploads            []*models.S3MultipartUpload
	CommonPrefixes     []string
}

// ListMultipartUploads 列出分片上传会话
func (s *S3ObjectService) ListMultipartUploads(ctx context.Context, input *ListMultipartUploadsInput) (*ListMultipartUploadsOutput, error) {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrBucketNotFoundError
		}
		return nil, err
	}

	// 2. 设置默认值
	maxUploads := input.MaxUploads
	if maxUploads <= 0 || maxUploads > 1000 {
		maxUploads = 1000
	}

	// 3. 列出上传会话（查询比 maxUploads 多一个，用于判断是否还有更多）
	multipartRepo := s.factory.S3Multipart()
	allUploads, err := multipartRepo.ListUploads(ctx, input.BucketName, input.UserID, input.Prefix, input.KeyMarker, input.UploadIDMarker, maxUploads+1)
	if err != nil {
		logger.LOG.Error("List multipart uploads failed", "error", err)
		return nil, err
	}

	// 4. 处理 delimiter（目录分隔符）
	var resultUploads []*models.S3MultipartUpload
	var commonPrefixes []string
	prefixSet := make(map[string]bool)

	isTruncated := false
	nextKeyMarker := ""
	nextUploadIDMarker := ""

	if input.Delimiter != "" {
		// 有 delimiter，需要区分文件和目录
		for _, upload := range allUploads {
			// 检查是否超过 maxUploads
			if len(resultUploads)+len(commonPrefixes) >= maxUploads {
				isTruncated = true
				nextKeyMarker = upload.ObjectKey
				nextUploadIDMarker = upload.UploadID
				break
			}

			// 移除 prefix 部分
			relativeKey := upload.ObjectKey
			if input.Prefix != "" {
				if !strings.HasPrefix(upload.ObjectKey, input.Prefix) {
					continue
				}
				relativeKey = upload.ObjectKey[len(input.Prefix):]
			}

			// 检查是否包含 delimiter
			delimiterIndex := strings.Index(relativeKey, input.Delimiter)
			if delimiterIndex >= 0 {
				// 这是一个目录前缀
				commonPrefix := input.Prefix + relativeKey[:delimiterIndex+len(input.Delimiter)]
				if !prefixSet[commonPrefix] {
					commonPrefixes = append(commonPrefixes, commonPrefix)
					prefixSet[commonPrefix] = true
				}
			} else {
				// 这是一个文件
				resultUploads = append(resultUploads, upload)
			}
		}
	} else {
		// 没有 delimiter，直接返回所有上传会话
		if len(allUploads) > maxUploads {
			isTruncated = true
			nextKeyMarker = allUploads[maxUploads].ObjectKey
			nextUploadIDMarker = allUploads[maxUploads].UploadID
			resultUploads = allUploads[:maxUploads]
		} else {
			resultUploads = allUploads
		}
	}

	return &ListMultipartUploadsOutput{
		BucketName:         input.BucketName,
		Prefix:             input.Prefix,
		Delimiter:          input.Delimiter,
		KeyMarker:          input.KeyMarker,
		UploadIDMarker:     input.UploadIDMarker,
		NextKeyMarker:      nextKeyMarker,
		NextUploadIDMarker: nextUploadIDMarker,
		MaxUploads:         maxUploads,
		IsTruncated:        isTruncated,
		Uploads:            resultUploads,
		CommonPrefixes:     commonPrefixes,
	}, nil
}

// CopyObjectInput 复制对象输入参数
type CopyObjectInput struct {
	SourceBucket      string
	SourceKey         string
	DestinationBucket string
	DestinationKey    string
	UserID            string
	MetadataDirective string // COPY 或 REPLACE
	UserMetadata      map[string]string
	StorageClass      string
}

// CopyObjectOutput 复制对象输出
type CopyObjectOutput struct {
	ETag            string
	LastModified    time.Time
	VersionID       string
	SourceVersionID string
}

// CopyObject 复制对象
func (s *S3ObjectService) CopyObject(ctx context.Context, input *CopyObjectInput) (*CopyObjectOutput, error) {
	// 1. 验证源和目标 Bucket 是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.SourceBucket, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("source bucket not found")
		}
		return nil, err
	}

	destBucket, err := s.bucketRepo.GetByName(ctx, input.DestinationBucket, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("destination bucket not found")
		}
		return nil, err
	}

	// 2. 获取源对象元数据
	sourceMetadata, err := s.objectMetadataRepo.GetByKey(ctx, input.SourceBucket, input.SourceKey, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("source object not found")
		}
		return nil, err
	}

	// 3. 获取源文件信息
	sourceFileInfo, err := s.factory.FileInfo().GetByID(ctx, sourceMetadata.FileID)
	if err != nil {
		return nil, fmt.Errorf("source file not found: %w", err)
	}

	// 4. 检查目标对象是否已存在（如果存在，标记旧版本）
	existingMetadata, err := s.objectMetadataRepo.GetByKey(ctx, input.DestinationBucket, input.DestinationKey, input.UserID)
	if err == nil && existingMetadata != nil {
		// 标记旧版本为非最新
		if err := s.objectMetadataRepo.MarkOldVersions(ctx, input.DestinationBucket, input.DestinationKey, input.UserID); err != nil {
			logger.LOG.Warn("Mark old versions failed", "error", err)
		}
	}

	// 5. 处理元数据
	userMetadataJSON := ""
	contentType := sourceMetadata.ContentType
	storageClass := sourceMetadata.StorageClass

	if input.MetadataDirective == "REPLACE" {
		// 使用新元数据
		if len(input.UserMetadata) > 0 {
			metadataBytes, err := json.Marshal(input.UserMetadata)
			if err != nil {
				return nil, fmt.Errorf("marshal user metadata failed: %w", err)
			}
			userMetadataJSON = string(metadataBytes)
		}
		// 可以从请求头获取新的 ContentType（如果提供）
		// 这里简化处理，保持原 ContentType
	} else {
		// COPY：使用源对象的元数据
		userMetadataJSON = sourceMetadata.UserMetadata
	}

	if input.StorageClass != "" {
		storageClass = input.StorageClass
	}

	// 6. 复制文件（如果源和目标在同一磁盘，可以创建硬链接；否则复制文件）
	// 查找源文件所在的磁盘
	sourceDisk, err := s.factory.Disk().GetByPath(ctx, filepath.Dir(sourceFileInfo.Path))
	if err != nil {
		logger.LOG.Warn("Get source disk failed, will copy file", "error", err)
		sourceDisk = nil
	}

	// 选择目标磁盘
	disks, err := s.factory.Disk().List(ctx, 0, 1000)
	if err != nil || len(disks) == 0 {
		return nil, types.ErrNoAvailableDiskError
	}

	var bestDisk *models.Disk
	var maxFreeSpace int64 = -1
	for _, disk := range disks {
		freeSpaceBytes := int64(disk.Size) * 1024 * 1024 * 1024
		if freeSpaceBytes > maxFreeSpace {
			maxFreeSpace = freeSpaceBytes
			bestDisk = disk
		}
	}

	// 创建目标文件路径
	finalFileName := fmt.Sprintf("%s_%s", sourceFileInfo.FileHash, filepath.Base(input.DestinationKey))
	finalPath := filepath.Join(bestDisk.DataPath, "files", finalFileName)

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
		return nil, fmt.Errorf("create target dir failed: %w", err)
	}

	// 如果源文件和目标文件在同一磁盘，尝试创建硬链接
	if sourceDisk != nil && sourceDisk.ID == bestDisk.ID {
		// 同磁盘，尝试创建硬链接
		if err := os.Link(sourceFileInfo.Path, finalPath); err != nil {
			// 硬链接失败（可能是跨文件系统），回退到复制
			logger.LOG.Debug("Hard link failed, fallback to copy", "error", err)
			if err := copyFile(sourceFileInfo.Path, finalPath); err != nil {
				return nil, fmt.Errorf("copy file failed: %w", err)
			}
		} else {
			logger.LOG.Debug("Created hard link for copy", "source", sourceFileInfo.Path, "dest", finalPath)
		}
	} else {
		// 不同磁盘，复制文件
		if err := copyFile(sourceFileInfo.Path, finalPath); err != nil {
			return nil, fmt.Errorf("copy file failed: %w", err)
		}
	}

	// 7. 创建目标 FileInfo（如果文件已存在，可以复用）
	var destFileInfo *models.FileInfo
	var isNewFile bool
	existingFile, err := s.factory.FileInfo().GetByHash(ctx, sourceFileInfo.FileHash)
	if err == nil && existingFile != nil {
		// 文件已存在，复用
		destFileInfo = existingFile
		isNewFile = false
		// 删除刚复制的文件（因为已存在）
		os.Remove(finalPath)
	} else {
		// 创建新的 FileInfo
		isNewFile = true
		destFileInfo = &models.FileInfo{
			ID:              uuid.New().String(),
			FileHash:        sourceFileInfo.FileHash,
			ChunkSignature:  sourceFileInfo.ChunkSignature,
			FirstChunkHash:  sourceFileInfo.FirstChunkHash,
			SecondChunkHash: sourceFileInfo.SecondChunkHash,
			ThirdChunkHash:  sourceFileInfo.ThirdChunkHash,
			Size:            sourceFileInfo.Size,
			Mime:            contentType,
			Path:            finalPath,
			IsEnc:           sourceFileInfo.IsEnc,
			ThumbnailImg:    sourceFileInfo.ThumbnailImg,
			CreatedAt:       custom_type.Now(),
		}

		if err := s.factory.FileInfo().Create(ctx, destFileInfo); err != nil {
			os.Remove(finalPath)
			return nil, fmt.Errorf("create file info failed: %w", err)
		}
	}

	// 8. 创建目标对象元数据
	versionID := uuid.New().String()
	destObjectMetadata := &models.S3ObjectMetadata{
		FileID:       destFileInfo.ID,
		BucketName:   input.DestinationBucket,
		ObjectKey:    input.DestinationKey,
		UserID:       input.UserID,
		ETag:         sourceMetadata.ETag,
		StorageClass: storageClass,
		ContentType:  contentType,
		UserMetadata: userMetadataJSON,
		VersionID:    versionID,
		IsLatest:     true,
		CreatedAt:    custom_type.Now(),
		UpdatedAt:    custom_type.Now(),
	}

	if err := s.objectMetadataRepo.Create(ctx, destObjectMetadata); err != nil {
		// 如果是新文件，需要清理
		if isNewFile {
			s.factory.FileInfo().Delete(ctx, destFileInfo.ID)
			os.Remove(finalPath)
		}
		return nil, fmt.Errorf("create object metadata failed: %w", err)
	}

	// 9. 创建 UserFiles 关联
	userFile := &models.UserFiles{
		UserID:      input.UserID,
		FileID:      destFileInfo.ID,
		FileName:    filepath.Base(input.DestinationKey),
		VirtualPath: fmt.Sprintf("%d", destBucket.VirtualPathID),
		IsPublic:    false,
		CreatedAt:   custom_type.Now(),
		UfID:        uuid.NewString(),
	}

	if err := s.factory.UserFiles().Create(ctx, userFile); err != nil {
		logger.LOG.Warn("Create user file failed", "error", err)
	}

	logger.LOG.Info("Copy object success",
		"source_bucket", input.SourceBucket,
		"source_key", input.SourceKey,
		"dest_bucket", input.DestinationBucket,
		"dest_key", input.DestinationKey,
		"etag", sourceMetadata.ETag,
	)

	return &CopyObjectOutput{
		ETag:            sourceMetadata.ETag,
		LastModified:    time.Time(destObjectMetadata.CreatedAt),
		VersionID:       versionID,
		SourceVersionID: sourceMetadata.VersionID,
	}, nil
}

// ==================== 批量删除相关方法 ====================

// DeleteObjectsInput 批量删除输入参数
type DeleteObjectsInput struct {
	BucketName string
	UserID     string
	Objects    []ObjectToDelete // 要删除的对象列表
	Quiet      bool             // 是否静默模式（只返回错误）
}

// ObjectToDelete 待删除对象
type ObjectToDelete struct {
	Key       string
	VersionID string // 可选：版本ID
}

// DeleteObjectsOutput 批量删除输出
type DeleteObjectsOutput struct {
	Deleted []DeletedObjectInfo
	Errors  []DeleteErrorInfo
}

// DeletedObjectInfo 已删除对象信息
type DeletedObjectInfo struct {
	Key       string
	VersionID string
}

// DeleteErrorInfo 删除错误信息
type DeleteErrorInfo struct {
	Key       string
	Code      string
	Message   string
	VersionID string
}

// DeleteObjects 批量删除对象
func (s *S3ObjectService) DeleteObjects(ctx context.Context, input *DeleteObjectsInput) (*DeleteObjectsOutput, error) {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrBucketNotFoundError
		}
		return nil, err
	}

	// 2. 批量删除对象
	var deleted []DeletedObjectInfo
	var errors []DeleteErrorInfo

	for _, obj := range input.Objects {
		// 删除单个对象（版本ID为空字符串表示删除最新版本）
		versionID := ""
		if obj.VersionID != "" {
			versionID = obj.VersionID
		}
		err := s.DeleteObject(ctx, input.BucketName, obj.Key, input.UserID, versionID)
		if err != nil {
			// 记录错误
			if !input.Quiet {
				errors = append(errors, DeleteErrorInfo{
					Key:       obj.Key,
					Code:      "InternalError",
					Message:   err.Error(),
					VersionID: obj.VersionID,
				})
			}
		} else {
			// 记录成功删除
			if !input.Quiet {
				deleted = append(deleted, DeletedObjectInfo{
					Key:       obj.Key,
					VersionID: obj.VersionID,
				})
			}
		}
	}

	logger.LOG.Info("DeleteObjects completed",
		"bucket", input.BucketName,
		"total", len(input.Objects),
		"deleted", len(deleted),
		"errors", len(errors),
	)

	return &DeleteObjectsOutput{
		Deleted: deleted,
		Errors:  errors,
	}, nil
}

// ==================== 预签名 URL 相关方法 ====================

// PresignedURLInput 生成预签名URL输入参数
type PresignedURLInput struct {
	BucketName  string
	ObjectKey   string
	Method      string            // HTTP方法（GET, PUT等）
	Expires     int64             // 过期时间（秒，默认3600）
	UserID      string            // 用户ID
	AccessKeyID string            // Access Key ID
	SecretKey   string            // Secret Key
	Headers     map[string]string // 额外的请求头（可选）
}

// PresignedURLOutput 生成预签名URL输出
type PresignedURLOutput struct {
	URL     string // 预签名URL
	Expires int64  // 过期时间（Unix时间戳）
}

// GeneratePresignedURL 生成预签名URL
func (s *S3ObjectService) GeneratePresignedURL(ctx context.Context, input *PresignedURLInput) (*PresignedURLOutput, error) {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrBucketNotFoundError
		}
		return nil, err
	}

	// 2. 设置默认值
	if input.Method == "" {
		input.Method = "GET"
	}
	if input.Expires <= 0 {
		input.Expires = 3600 // 默认1小时
	}
	if input.Expires > 604800 {
		// S3限制：预签名URL最长7天
		input.Expires = 604800
	}

	// 3. 构建基础URL（从配置读取）
	protocol := "http"
	if config.CONFIG.Server.SSL {
		protocol = "https"
	}
	baseURL := fmt.Sprintf("%s://%s:%d", protocol, config.CONFIG.Server.Host, config.CONFIG.Server.Port)
	if config.CONFIG.Server.Host == "0.0.0.0" {
		// 如果监听所有接口，使用 localhost
		baseURL = fmt.Sprintf("%s://localhost:%d", protocol, config.CONFIG.Server.Port)
	}

	// 4. 获取区域（从配置读取）
	region := config.CONFIG.S3.Region
	if region == "" {
		region = "us-east-1"
	}

	// 5. 生成预签名URL
	signer := auth.NewSignatureV4(region, "s3")
	url, err := signer.GeneratePresignedURL(baseURL, auth.PresignedURLParams{
		Method:      input.Method,
		Bucket:      input.BucketName,
		Key:         input.ObjectKey,
		Expires:     input.Expires,
		AccessKeyID: input.AccessKeyID,
		SecretKey:   input.SecretKey,
		Headers:     input.Headers,
	})
	if err != nil {
		return nil, fmt.Errorf("generate presigned URL failed: %w", err)
	}

	// 5. 计算过期时间戳
	expiresAt := time.Now().Add(time.Duration(input.Expires) * time.Second).Unix()

	logger.LOG.Info("Generate presigned URL success",
		"bucket", input.BucketName,
		"key", input.ObjectKey,
		"method", input.Method,
		"expires", input.Expires,
	)

	return &PresignedURLOutput{
		URL:     url,
		Expires: expiresAt,
	}, nil
}

// ==================== Object ACL 相关方法 ====================

// PutObjectACLInput 设置对象ACL输入参数
type PutObjectACLInput struct {
	BucketName string
	ObjectKey  string
	UserID     string
	VersionID  string // 可选：版本ID
	ACL        *types.AccessControlPolicy
}

// PutObjectACL 设置对象ACL配置
func (s *S3ObjectService) PutObjectACL(ctx context.Context, input *PutObjectACLInput) error {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrBucketNotFoundError
		}
		return err
	}

	// 2. 验证对象是否存在（支持版本ID）
	if input.VersionID != "" {
		_, err = s.objectMetadataRepo.GetByKeyAndVersion(ctx, input.BucketName, input.ObjectKey, input.VersionID, input.UserID)
	} else {
		_, err = s.objectMetadataRepo.GetByKey(ctx, input.BucketName, input.ObjectKey, input.UserID)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrObjectNotFoundError
		}
		return err
	}

	// 3. 验证ACL配置
	if input.ACL == nil {
		return fmt.Errorf("ACL configuration is required")
	}

	// 4. 验证Owner（必须是对象所有者）
	if input.ACL.Owner.ID == "" || input.ACL.Owner.ID != input.UserID {
		return fmt.Errorf("ACL owner must be the object owner")
	}

	// 5. 验证Grants
	for i, grant := range input.ACL.AccessControlList.Grants {
		if grant.Permission == "" {
			return fmt.Errorf("grant %d: permission is required", i)
		}
		validPermissions := map[string]bool{
			"READ":         true,
			"WRITE":        true,
			"READ_ACP":     true,
			"WRITE_ACP":    true,
			"FULL_CONTROL": true,
		}
		if !validPermissions[grant.Permission] {
			return fmt.Errorf("grant %d: invalid permission: %s", i, grant.Permission)
		}
	}

	// 6. 序列化ACL配置为JSON
	aclJSON, err := json.Marshal(input.ACL)
	if err != nil {
		logger.LOG.Error("Marshal ACL config failed", "error", err)
		return fmt.Errorf("marshal ACL config failed: %w", err)
	}

	// 7. 创建或更新ACL配置
	aclConfig := &models.S3ObjectACL{
		BucketName: input.BucketName,
		ObjectKey:  input.ObjectKey,
		VersionID:  input.VersionID,
		UserID:     input.UserID,
		ACLConfig:  string(aclJSON),
		UpdatedAt:  custom_type.Now(),
	}

	if err := s.factory.S3ACL().CreateOrUpdateObjectACL(ctx, aclConfig); err != nil {
		logger.LOG.Error("Create or update object ACL failed", "error", err)
		return err
	}

	logger.LOG.Info("Put object ACL success",
		"bucket", input.BucketName,
		"key", input.ObjectKey,
		"version_id", input.VersionID,
		"grants_count", len(input.ACL.AccessControlList.Grants),
	)

	return nil
}

// GetObjectACLInput 获取对象ACL输入参数
type GetObjectACLInput struct {
	BucketName string
	ObjectKey  string
	UserID     string
	VersionID  string // 可选：版本ID
}

// GetObjectACLOutput 获取对象ACL输出
type GetObjectACLOutput struct {
	ACL *types.AccessControlPolicy
}

// GetObjectACL 获取对象ACL配置
func (s *S3ObjectService) GetObjectACL(ctx context.Context, input *GetObjectACLInput) (*GetObjectACLOutput, error) {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrBucketNotFoundError
		}
		return nil, err
	}

	// 2. 验证对象是否存在（支持版本ID）
	var _ *models.S3ObjectMetadata
	if input.VersionID != "" {
		_, err = s.objectMetadataRepo.GetByKeyAndVersion(ctx, input.BucketName, input.ObjectKey, input.VersionID, input.UserID)
	} else {
		_, err = s.objectMetadataRepo.GetByKey(ctx, input.BucketName, input.ObjectKey, input.UserID)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrObjectNotFoundError
		}
		return nil, err
	}

	// 3. 获取ACL配置
	acl, err := s.factory.S3ACL().GetObjectACL(ctx, input.BucketName, input.ObjectKey, input.VersionID, input.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果没有ACL配置，返回默认ACL（私有，只有所有者有权限）
			defaultACL := &types.AccessControlPolicy{
				Owner: types.Owner{
					ID:          input.UserID,
					DisplayName: input.UserID,
				},
				AccessControlList: types.AccessControlList{
					Owner: types.Owner{
						ID:          input.UserID,
						DisplayName: input.UserID,
					},
					Grants: []types.Grant{
						{
							Grantee: types.Grantee{
								Type:        "CanonicalUser",
								ID:          input.UserID,
								DisplayName: input.UserID,
							},
							Permission: "FULL_CONTROL",
						},
					},
				},
			}
			return &GetObjectACLOutput{ACL: defaultACL}, nil
		}
		return nil, err
	}

	// 4. 反序列化ACL配置
	var aclPolicy types.AccessControlPolicy
	if err := json.Unmarshal([]byte(acl.ACLConfig), &aclPolicy); err != nil {
		logger.LOG.Error("Unmarshal ACL config failed", "error", err)
		return nil, fmt.Errorf("unmarshal ACL config failed: %w", err)
	}

	return &GetObjectACLOutput{
		ACL: &aclPolicy,
	}, nil
}

// convertCannedACLToPolicy 将预定义 ACL（如 "public-read"）转换为完整的 ACL 策略
func (s *S3ObjectService) convertCannedACLToPolicy(cannedACL string, userID string) *types.AccessControlPolicy {
	// 预定义 ACL 映射
	allUsersURI := "http://acs.amazonaws.com/groups/global/AllUsers"
	authenticatedUsersURI := "http://acs.amazonaws.com/groups/global/AuthenticatedUsers"

	aclPolicy := &types.AccessControlPolicy{
		Owner: types.Owner{
			ID:          userID,
			DisplayName: userID,
		},
		AccessControlList: types.AccessControlList{
			Owner: types.Owner{
				ID:          userID,
				DisplayName: userID,
			},
			Grants: []types.Grant{
				{
					Grantee: types.Grantee{
						Type:        "CanonicalUser",
						ID:          userID,
						DisplayName: userID,
					},
					Permission: "FULL_CONTROL",
				},
			},
		},
	}

	switch strings.ToLower(cannedACL) {
	case "private":
		// 私有：只有所有者有权限（默认）
		break
	case "public-read":
		// 公开读取：所有人可以读取
		aclPolicy.AccessControlList.Grants = append(aclPolicy.AccessControlList.Grants, types.Grant{
			Grantee: types.Grantee{
				Type: "Group",
				URI:  allUsersURI,
			},
			Permission: "READ",
		})
	case "public-read-write":
		// 公开读写：所有人可以读写
		aclPolicy.AccessControlList.Grants = append(aclPolicy.AccessControlList.Grants, types.Grant{
			Grantee: types.Grantee{
				Type: "Group",
				URI:  allUsersURI,
			},
			Permission: "READ",
		}, types.Grant{
			Grantee: types.Grantee{
				Type: "Group",
				URI:  allUsersURI,
			},
			Permission: "WRITE",
		})
	case "authenticated-read":
		// 认证用户读取：所有 AWS 认证用户可以读取
		aclPolicy.AccessControlList.Grants = append(aclPolicy.AccessControlList.Grants, types.Grant{
			Grantee: types.Grantee{
				Type: "Group",
				URI:  authenticatedUsersURI,
			},
			Permission: "READ",
		})
	case "bucket-owner-read":
		// Bucket 所有者读取（需要 Bucket 信息，这里简化处理）
		// 注意：这个 ACL 需要 Bucket 所有者信息，当前实现简化处理
		break
	case "bucket-owner-full-control":
		// Bucket 所有者完全控制（需要 Bucket 信息，这里简化处理）
		break
	default:
		// 未知的 ACL，返回 nil
		logger.LOG.Warn("Unknown canned ACL", "acl", cannedACL)
		return nil
	}

	return aclPolicy
}

// isImage 判断MIME类型是否为图片
func isImage(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

// handleServerSideEncryption 处理服务端加密
func (s *S3ObjectService) handleServerSideEncryption(ctx context.Context, input *PutObjectInput, fileInfo *models.FileInfo, isNewFile bool) (*models.S3ObjectEncryption, error) {
	// 只支持 SSE-S3（使用S3管理的密钥）
	// 注意：aws:kms 需要KMS服务支持，这里暂时只实现 SSE-S3
	if input.SSEAlgorithm != "AES256" {
		if input.SSEAlgorithm == "aws:kms" {
			return nil, fmt.Errorf("SSE-KMS not implemented yet, only SSE-S3 (AES256) is supported")
		}
		return nil, fmt.Errorf("unsupported encryption algorithm: %s (only AES256 is supported)", input.SSEAlgorithm)
	}

	// SSE-C（客户提供的密钥）暂不支持
	if input.SSECustomerKey != "" {
		return nil, fmt.Errorf("SSE-C not implemented yet, only SSE-S3 (AES256) is supported")
	}

	// 生成数据密钥
	keyID, dataKey, err := s.encryptionService.GenerateDataKey()
	if err != nil {
		return nil, fmt.Errorf("generate data key failed: %w", err)
	}

	// 加密数据密钥（使用主密钥）
	encryptedKeyData, err := s.encryptionService.EncryptDataKey(dataKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt data key failed: %w", err)
	}

	var iv []byte
	var encryptedPath string

	// 如果文件是新文件，需要加密文件
	if isNewFile {
		encryptedPath = s.encryptionService.GetEncryptedFilePath(fileInfo.Path)
		// EncryptFile 会将IV写入文件开头，并返回这个IV
		iv, err = s.encryptionService.EncryptFile(fileInfo.Path, encryptedPath, dataKey)
		if err != nil {
			return nil, fmt.Errorf("encrypt file failed: %w", err)
		}

		// 删除原文件，使用加密文件
		if err := os.Remove(fileInfo.Path); err != nil {
			logger.LOG.Warn("Remove original file failed", "error", err)
			// 如果删除失败，尝试清理加密文件
			os.Remove(encryptedPath)
			return nil, fmt.Errorf("remove original file failed: %w", err)
		}
		fileInfo.Path = encryptedPath

		// 更新文件大小（加密后增加了IV的大小，即16字节）
		if stat, err := os.Stat(encryptedPath); err == nil {
			fileInfo.Size = int(stat.Size())
		}
	} else {
		// 对于秒传的文件，检查是否已加密
		// 如果文件路径包含 .encrypted 后缀，说明已加密
		if strings.HasSuffix(fileInfo.Path, ".encrypted") {
			// 文件已加密，需要从加密文件中读取IV
			// DecryptFile 会从文件开头读取IV，但这里我们只需要读取IV，不解密
			file, err := os.Open(fileInfo.Path)
			if err != nil {
				return nil, fmt.Errorf("open encrypted file failed: %w", err)
			}
			defer file.Close()

			// 读取文件开头的IV（16字节）
			iv = make([]byte, aes.BlockSize)
			if _, err := io.ReadFull(file, iv); err != nil {
				return nil, fmt.Errorf("read IV from encrypted file failed: %w", err)
			}
		} else {
			// 文件未加密，需要加密
			encryptedPath = s.encryptionService.GetEncryptedFilePath(fileInfo.Path)
			iv, err = s.encryptionService.EncryptFile(fileInfo.Path, encryptedPath, dataKey)
			if err != nil {
				return nil, fmt.Errorf("encrypt existing file failed: %w", err)
			}

			// 删除原文件，使用加密文件
			if err := os.Remove(fileInfo.Path); err != nil {
				logger.LOG.Warn("Remove original file failed", "error", err)
				os.Remove(encryptedPath)
				return nil, fmt.Errorf("remove original file failed: %w", err)
			}
			fileInfo.Path = encryptedPath

			// 更新文件大小
			if stat, err := os.Stat(encryptedPath); err == nil {
				fileInfo.Size = int(stat.Size())
			}

			// 更新FileInfo记录
			if err := s.factory.FileInfo().Update(ctx, fileInfo); err != nil {
				logger.LOG.Error("Update file info path failed", "error", err)
				// 回滚：恢复原文件路径
				os.Remove(encryptedPath)
				return nil, fmt.Errorf("update file info failed: %w", err)
			}
		}
	}

	// 保存加密密钥到数据库
	encryptionKey := &models.S3EncryptionKey{
		KeyID:     keyID,
		Algorithm: "AES256",
		KeyData:   encryptedKeyData, // EncryptDataKey 已经返回 base64 编码的字符串
		CreatedAt: custom_type.Now(),
		UpdatedAt: custom_type.Now(),
	}

	if err := s.factory.S3EncryptionKey().Create(ctx, encryptionKey); err != nil {
		// 如果保存密钥失败，需要清理加密文件（如果是新加密的）
		if !isNewFile && encryptedPath != "" {
			os.Remove(encryptedPath)
		}
		return nil, fmt.Errorf("save encryption key failed: %w", err)
	}

	// 创建加密元数据
	// 注意：VersionID 将在创建对象元数据后设置（在事务中）
	encryptionMetadata := &models.S3ObjectEncryption{
		BucketName:     input.BucketName,
		ObjectKey:      input.ObjectKey,
		VersionID:      "", // 将在事务中创建对象元数据后设置
		UserID:         input.UserID,
		EncryptionType: "SSE-S3",
		KeyID:          keyID,
		IV:             base64.StdEncoding.EncodeToString(iv), // 使用加密时生成的IV
		CreatedAt:      custom_type.Now(),
	}

	return encryptionMetadata, nil
}

// ==================== Object Tagging 相关方法 ====================

// PutObjectTaggingInput 设置对象标签输入参数
type PutObjectTaggingInput struct {
	BucketName string
	ObjectKey  string
	UserID     string
	VersionID  string // 可选：版本ID
	Tags       map[string]string
}

// PutObjectTagging 设置对象标签
func (s *S3ObjectService) PutObjectTagging(ctx context.Context, input *PutObjectTaggingInput) error {
	// 添加超时控制
	timeout := getOperationTimeout()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrBucketNotFoundError
		}
		return err
	}

	// 2. 验证对象是否存在（支持版本ID）
	var objectMetadata *models.S3ObjectMetadata
	if input.VersionID != "" {
		objectMetadata, err = s.objectMetadataRepo.GetByKeyAndVersion(ctx, input.BucketName, input.ObjectKey, input.VersionID, input.UserID)
	} else {
		objectMetadata, err = s.objectMetadataRepo.GetByKey(ctx, input.BucketName, input.ObjectKey, input.UserID)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrObjectNotFoundError
		}
		return err
	}

	// 3. 验证标签数量（S3限制：最多10个标签）
	if len(input.Tags) > 10 {
		return fmt.Errorf("too many tags: maximum 10 tags allowed")
	}

	// 4. 验证标签键值长度（S3限制：键和值长度不超过256字符）
	for key, value := range input.Tags {
		if len(key) > 256 {
			return fmt.Errorf("tag key length must be <= 256 characters")
		}
		if len(value) > 256 {
			return fmt.Errorf("tag value length must be <= 256 characters")
		}
	}

	// 5. 序列化标签为JSON
	tagsJSON := ""
	if len(input.Tags) > 0 {
		tagsBytes, err := json.Marshal(input.Tags)
		if err != nil {
			logger.LOG.Error("Marshal tags failed", "error", err)
			return fmt.Errorf("marshal tags failed: %w", err)
		}
		tagsJSON = string(tagsBytes)
	}

	// 6. 更新对象元数据的标签字段
	objectMetadata.Tags = tagsJSON
	objectMetadata.UpdatedAt = custom_type.Now()

	if err := s.objectMetadataRepo.Update(ctx, objectMetadata); err != nil {
		logger.LOG.Error("Update object tags failed", "error", err)
		return err
	}

	logger.LOG.Info("Put object tagging success",
		"bucket", input.BucketName,
		"key", input.ObjectKey,
		"version_id", input.VersionID,
		"tags_count", len(input.Tags),
	)

	return nil
}

// GetObjectTaggingInput 获取对象标签输入参数
type GetObjectTaggingInput struct {
	BucketName string
	ObjectKey  string
	UserID     string
	VersionID  string // 可选：版本ID
}

// GetObjectTaggingOutput 获取对象标签输出
type GetObjectTaggingOutput struct {
	Tags map[string]string
}

// GetObjectTagging 获取对象标签
func (s *S3ObjectService) GetObjectTagging(ctx context.Context, input *GetObjectTaggingInput) (*GetObjectTaggingOutput, error) {
	// 添加超时控制
	timeout := getOperationTimeout()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrBucketNotFoundError
		}
		return nil, err
	}

	// 2. 验证对象是否存在（支持版本ID）
	var objectMetadata *models.S3ObjectMetadata
	if input.VersionID != "" {
		objectMetadata, err = s.objectMetadataRepo.GetByKeyAndVersion(ctx, input.BucketName, input.ObjectKey, input.VersionID, input.UserID)
	} else {
		objectMetadata, err = s.objectMetadataRepo.GetByKey(ctx, input.BucketName, input.ObjectKey, input.UserID)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrObjectNotFoundError
		}
		return nil, err
	}

	// 3. 解析标签JSON
	tags := make(map[string]string)
	if objectMetadata.Tags != "" {
		if err := json.Unmarshal([]byte(objectMetadata.Tags), &tags); err != nil {
			logger.LOG.Warn("Unmarshal tags failed, using empty tags", "error", err)
			tags = make(map[string]string)
		}
	}

	return &GetObjectTaggingOutput{
		Tags: tags,
	}, nil
}

// DeleteObjectTaggingInput 删除对象标签输入参数
type DeleteObjectTaggingInput struct {
	BucketName string
	ObjectKey  string
	UserID     string
	VersionID  string // 可选：版本ID
}

// DeleteObjectTagging 删除对象标签
func (s *S3ObjectService) DeleteObjectTagging(ctx context.Context, input *DeleteObjectTaggingInput) error {
	// 添加超时控制
	timeout := getOperationTimeout()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrBucketNotFoundError
		}
		return err
	}

	// 2. 验证对象是否存在（支持版本ID）
	var objectMetadata *models.S3ObjectMetadata
	if input.VersionID != "" {
		objectMetadata, err = s.objectMetadataRepo.GetByKeyAndVersion(ctx, input.BucketName, input.ObjectKey, input.VersionID, input.UserID)
	} else {
		objectMetadata, err = s.objectMetadataRepo.GetByKey(ctx, input.BucketName, input.ObjectKey, input.UserID)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrObjectNotFoundError
		}
		return err
	}

	// 3. 清空标签字段
	objectMetadata.Tags = ""
	objectMetadata.UpdatedAt = custom_type.Now()

	if err := s.objectMetadataRepo.Update(ctx, objectMetadata); err != nil {
		logger.LOG.Error("Delete object tags failed", "error", err)
		return err
	}

	logger.LOG.Info("Delete object tagging success",
		"bucket", input.BucketName,
		"key", input.ObjectKey,
		"version_id", input.VersionID,
	)

	return nil
}

package service

import (
	"context"
	"errors"
	"fmt"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// S3BucketService S3 Bucket服务
type S3BucketService struct {
	bucketRepo      repository.S3BucketRepository
	virtualPathRepo repository.VirtualPathRepository
	factory         *impl.RepositoryFactory
}

// NewS3BucketService 创建S3 Bucket服务
func NewS3BucketService(factory *impl.RepositoryFactory) *S3BucketService {
	return &S3BucketService{
		bucketRepo:      factory.S3Bucket(),
		virtualPathRepo: factory.VirtualPath(),
		factory:         factory,
	}
}

// ValidateBucketName 验证Bucket名称（符合S3规范）
func ValidateBucketName(bucketName string) error {
	// S3 bucket命名规范：
	// 1. 长度在3-63个字符之间
	// 2. 只能包含小写字母、数字、点(.)和连字符(-)
	// 3. 必须以字母或数字开头和结尾
	// 4. 不能包含连续的点
	// 5. 不能是IP地址格式

	if len(bucketName) < 3 || len(bucketName) > 63 {
		return fmt.Errorf("bucket name must be between 3 and 63 characters long")
	}

	// 检查字符合法性
	matched, _ := regexp.MatchString(`^[a-z0-9][a-z0-9.-]*[a-z0-9]$`, bucketName)
	if !matched {
		return fmt.Errorf("bucket name must consist of lowercase letters, numbers, dots and hyphens")
	}

	// 不能包含连续的点
	if strings.Contains(bucketName, "..") {
		return fmt.Errorf("bucket name cannot contain consecutive dots")
	}

	// 不能是IP地址格式
	ipPattern := regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)
	if ipPattern.MatchString(bucketName) {
		return fmt.Errorf("bucket name cannot be formatted as IP address")
	}

	return nil
}

// ListBuckets 列出用户的所有Bucket
func (s *S3BucketService) ListBuckets(ctx context.Context, userID string) ([]*models.S3Bucket, error) {
	buckets, err := s.bucketRepo.ListByUserID(ctx, userID)
	if err != nil {
		logger.LOG.Error("List S3 buckets failed",
			"user_id", userID,
			"error", err,
		)
		return nil, err
	}

	logger.LOG.Info("List S3 buckets success",
		"user_id", userID,
		"bucket_count", len(buckets),
	)

	return buckets, nil
}

// CreateBucket 创建Bucket（对应创建虚拟目录）
func (s *S3BucketService) CreateBucket(ctx context.Context, bucketName, userID, region string) error {
	// 1. 验证Bucket名称
	if err := ValidateBucketName(bucketName); err != nil {
		logger.LOG.Warn("Invalid bucket name",
			"bucket_name", bucketName,
			"error", err,
		)
		return err
	}

	// 2. 检查是否已存在
	exists, err := s.bucketRepo.Exists(ctx, bucketName, userID)
	if err != nil {
		logger.LOG.Error("Check bucket existence failed",
			"bucket_name", bucketName,
			"user_id", userID,
			"error", err,
		)
		return err
	}

	if exists {
		return fmt.Errorf("bucket already exists")
	}

	// 3. 创建对应的虚拟路径
	// 获取用户根目录
	rootPath, err := s.virtualPathRepo.GetRootPath(ctx, userID)
	if err != nil {
		logger.LOG.Error("Get user root path failed",
			"user_id", userID,
			"error", err,
		)
		return fmt.Errorf("get user root path failed: %w", err)
	}

	// 在根目录下创建bucket对应的虚拟目录
	virtualPath := &models.VirtualPath{
		UserID:      userID,
		Path:        "/" + bucketName,
		ParentLevel: fmt.Sprintf("%d", rootPath.ID),
		IsDir:       true,
		CreatedTime: custom_type.Now(),
		UpdateTime:  custom_type.Now(),
	}

	if err := s.virtualPathRepo.Create(ctx, virtualPath); err != nil {
		logger.LOG.Error("Create virtual path failed",
			"bucket_name", bucketName,
			"user_id", userID,
			"error", err,
		)
		return fmt.Errorf("create virtual path failed: %w", err)
	}

	// 4. 创建Bucket记录
	bucket := &models.S3Bucket{
		BucketName:    bucketName,
		UserID:        userID,
		Region:        region,
		VirtualPathID: virtualPath.ID,
		CreatedAt:     custom_type.Now(),
		UpdatedAt:     custom_type.Now(),
	}

	if err := s.bucketRepo.Create(ctx, bucket); err != nil {
		// 回滚：删除虚拟路径
		s.virtualPathRepo.Delete(ctx, virtualPath.ID)
		logger.LOG.Error("Create bucket failed",
			"bucket_name", bucketName,
			"user_id", userID,
			"error", err,
		)
		return fmt.Errorf("create bucket failed: %w", err)
	}

	logger.LOG.Info("Create bucket success",
		"bucket_name", bucketName,
		"user_id", userID,
		"virtual_path_id", virtualPath.ID,
	)

	return nil
}

// HeadBucket 检查Bucket是否存在
func (s *S3BucketService) HeadBucket(ctx context.Context, bucketName, userID string) (bool, error) {
	exists, err := s.bucketRepo.Exists(ctx, bucketName, userID)
	if err != nil {
		logger.LOG.Error("Check bucket existence failed",
			"bucket_name", bucketName,
			"user_id", userID,
			"error", err,
		)
		return false, err
	}

	return exists, nil
}

// DeleteBucket 删除Bucket
func (s *S3BucketService) DeleteBucket(ctx context.Context, bucketName, userID string) error {
	// 1. 获取Bucket
	bucket, err := s.bucketRepo.GetByName(ctx, bucketName, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("bucket not found")
		}
		logger.LOG.Error("Get bucket failed",
			"bucket_name", bucketName,
			"user_id", userID,
			"error", err,
		)
		return err
	}

	// 2. 检查Bucket是否为空（检查虚拟路径下是否有文件）
	fileCount, err := s.factory.FileInfo().CountByVirtualPath(ctx, userID, fmt.Sprintf("%d", bucket.VirtualPathID))
	if err != nil {
		logger.LOG.Error("Count files in bucket failed",
			"bucket_name", bucketName,
			"error", err,
		)
		return err
	}

	if fileCount > 0 {
		return fmt.Errorf("bucket is not empty")
	}

	// 3. 删除Bucket
	if err := s.bucketRepo.Delete(ctx, bucket.ID); err != nil {
		logger.LOG.Error("Delete bucket failed",
			"bucket_name", bucketName,
			"user_id", userID,
			"error", err,
		)
		return err
	}

	// 4. 删除虚拟路径
	if err := s.virtualPathRepo.Delete(ctx, bucket.VirtualPathID); err != nil {
		logger.LOG.Error("Delete virtual path failed",
			"virtual_path_id", bucket.VirtualPathID,
			"error", err,
		)
		// 不返回错误，因为Bucket已删除
	}

	logger.LOG.Info("Delete bucket success",
		"bucket_name", bucketName,
		"user_id", userID,
	)

	return nil
}

// GetBucket 获取Bucket信息
func (s *S3BucketService) GetBucket(ctx context.Context, bucketName, userID string) (*models.S3Bucket, error) {
	bucket, err := s.bucketRepo.GetByName(ctx, bucketName, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("bucket not found")
		}
		logger.LOG.Error("Get bucket failed",
			"bucket_name", bucketName,
			"user_id", userID,
			"error", err,
		)
		return nil, err
	}

	return bucket, nil
}

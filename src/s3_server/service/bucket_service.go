package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"
	"myobj/src/s3_server/types"
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
	// 添加超时控制
	timeout := getOperationTimeout()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 1. 验证Bucket名称
	if err := ValidateBucketName(bucketName); err != nil {
		logger.LOG.Warn("Invalid bucket name",
			"bucket_name", bucketName,
			"error", err,
		)
		return fmt.Errorf("%w: %w", types.ErrInvalidBucketNameError, err)
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
		return types.ErrBucketAlreadyExistsError
	}

	// 3. 使用事务创建Bucket和虚拟路径
	db := s.factory.DB()
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txFactory := s.factory.WithTx(tx)
		txVirtualPathRepo := txFactory.VirtualPath()
		txBucketRepo := txFactory.S3Bucket()

		// 获取用户根目录
		rootPath, err := txVirtualPathRepo.GetRootPath(ctx, userID)
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

		if err := txVirtualPathRepo.Create(ctx, virtualPath); err != nil {
			logger.LOG.Error("Create virtual path failed",
				"bucket_name", bucketName,
				"user_id", userID,
				"error", err,
			)
			return fmt.Errorf("create virtual path failed: %w", err)
		}

		// 创建Bucket记录
		bucket := &models.S3Bucket{
			BucketName:    bucketName,
			UserID:        userID,
			Region:        region,
			VirtualPathID: virtualPath.ID,
			CreatedAt:     custom_type.Now(),
			UpdatedAt:     custom_type.Now(),
		}

		if err := txBucketRepo.Create(ctx, bucket); err != nil {
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
	})
}

// HeadBucket 检查Bucket是否存在
func (s *S3BucketService) HeadBucket(ctx context.Context, bucketName, userID string) (bool, error) {
	// 添加超时控制
	timeout := getOperationTimeout()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

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
	// 添加超时控制
	timeout := getOperationTimeout()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 1. 获取Bucket
	bucket, err := s.bucketRepo.GetByName(ctx, bucketName, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrBucketNotFoundError
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
		return types.ErrBucketNotEmptyError
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
			return nil, types.ErrBucketNotFoundError
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

// ==================== 版本控制相关方法 ====================

// PutBucketVersioningInput 设置Bucket版本控制输入参数
type PutBucketVersioningInput struct {
	BucketName string
	UserID     string
	Status     string // Enabled/Suspended/Disabled
}

// PutBucketVersioning 设置Bucket版本控制状态
func (s *S3BucketService) PutBucketVersioning(ctx context.Context, input *PutBucketVersioningInput) error {
	// 1. 验证Bucket是否存在
	bucket, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return types.ErrBucketNotFoundError
		}
		return err
	}

	// 2. 验证状态值
	validStatuses := map[string]bool{
		"Enabled":   true,
		"Suspended": true,
		"Disabled":  true,
	}
	if !validStatuses[input.Status] {
		return fmt.Errorf("invalid versioning status: %s", input.Status)
	}

	// 3. 更新版本控制状态
	bucket.Versioning = input.Status
	bucket.UpdatedAt = custom_type.Now()

	if err := s.bucketRepo.Update(ctx, bucket); err != nil {
		logger.LOG.Error("Update bucket versioning failed",
			"bucket_name", input.BucketName,
			"user_id", input.UserID,
			"status", input.Status,
			"error", err,
		)
		return err
	}

	logger.LOG.Info("Put bucket versioning success",
		"bucket_name", input.BucketName,
		"user_id", input.UserID,
		"status", input.Status,
	)

	return nil
}

// GetBucketVersioningInput 获取Bucket版本控制状态输入参数
type GetBucketVersioningInput struct {
	BucketName string
	UserID     string
}

// GetBucketVersioningOutput 获取Bucket版本控制状态输出
type GetBucketVersioningOutput struct {
	Status string // Enabled/Suspended/Disabled
}

// GetBucketVersioning 获取Bucket版本控制状态
func (s *S3BucketService) GetBucketVersioning(ctx context.Context, input *GetBucketVersioningInput) (*GetBucketVersioningOutput, error) {
	// 1. 验证Bucket是否存在
	bucket, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrBucketNotFoundError
		}
		return nil, err
	}

	// 2. 返回版本控制状态
	status := bucket.Versioning
	if status == "" {
		status = "Disabled" // 默认值
	}

	return &GetBucketVersioningOutput{
		Status: status,
	}, nil
}

// ==================== CORS相关方法 ====================

// PutBucketCORSInput 设置Bucket CORS配置输入参数
type PutBucketCORSInput struct {
	BucketName string
	UserID     string
	CORSConfig *types.CORSConfiguration
}

// PutBucketCORS 设置Bucket CORS配置
func (s *S3BucketService) PutBucketCORS(ctx context.Context, input *PutBucketCORSInput) error {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return types.ErrBucketNotFoundError
		}
		return err
	}

	// 2. 验证CORS规则
	if input.CORSConfig == nil || len(input.CORSConfig.CORSRules) == 0 {
		return fmt.Errorf("CORS configuration is required")
	}

	// 验证每个规则
	for i, rule := range input.CORSConfig.CORSRules {
		if len(rule.AllowedOrigins) == 0 {
			return fmt.Errorf("CORS rule %d: AllowedOrigin is required", i)
		}
		if len(rule.AllowedMethods) == 0 {
			return fmt.Errorf("CORS rule %d: AllowedMethod is required", i)
		}
	}

	// 3. 序列化CORS配置为JSON
	corsJSON, err := json.Marshal(input.CORSConfig)
	if err != nil {
		logger.LOG.Error("Marshal CORS config failed", "error", err)
		return fmt.Errorf("marshal CORS config failed: %w", err)
	}

	// 4. 创建或更新CORS配置
	corsConfig := &models.S3BucketCORS{
		BucketName: input.BucketName,
		UserID:     input.UserID,
		CORSConfig: string(corsJSON),
		UpdatedAt:  custom_type.Now(),
	}

	if err := s.factory.S3BucketCORS().CreateOrUpdate(ctx, corsConfig); err != nil {
		logger.LOG.Error("Create or update CORS config failed",
			"bucket_name", input.BucketName,
			"user_id", input.UserID,
			"error", err,
		)
		return err
	}

	logger.LOG.Info("Put bucket CORS success",
		"bucket_name", input.BucketName,
		"user_id", input.UserID,
		"rules_count", len(input.CORSConfig.CORSRules),
	)

	return nil
}

// GetBucketCORSInput 获取Bucket CORS配置输入参数
type GetBucketCORSInput struct {
	BucketName string
	UserID     string
}

// GetBucketCORSOutput 获取Bucket CORS配置输出
type GetBucketCORSOutput struct {
	CORSConfig *types.CORSConfiguration
}

// GetBucketCORS 获取Bucket CORS配置
func (s *S3BucketService) GetBucketCORS(ctx context.Context, input *GetBucketCORSInput) (*GetBucketCORSOutput, error) {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrBucketNotFoundError
		}
		return nil, err
	}

	// 2. 获取CORS配置
	cors, err := s.factory.S3BucketCORS().GetByBucket(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrCORSNotFoundError
		}
		return nil, err
	}

	// 3. 反序列化CORS配置
	var corsConfig types.CORSConfiguration
	if err := json.Unmarshal([]byte(cors.CORSConfig), &corsConfig); err != nil {
		logger.LOG.Error("Unmarshal CORS config failed", "error", err)
		return nil, fmt.Errorf("unmarshal CORS config failed: %w", err)
	}

	return &GetBucketCORSOutput{
		CORSConfig: &corsConfig,
	}, nil
}

// DeleteBucketCORSInput 删除Bucket CORS配置输入参数
type DeleteBucketCORSInput struct {
	BucketName string
	UserID     string
}

// DeleteBucketCORS 删除Bucket CORS配置
func (s *S3BucketService) DeleteBucketCORS(ctx context.Context, input *DeleteBucketCORSInput) error {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return types.ErrBucketNotFoundError
		}
		return err
	}

	// 2. 删除CORS配置
	if err := s.factory.S3BucketCORS().Delete(ctx, input.BucketName, input.UserID); err != nil {
		logger.LOG.Error("Delete CORS config failed",
			"bucket_name", input.BucketName,
			"user_id", input.UserID,
			"error", err,
		)
		return err
	}

	logger.LOG.Info("Delete bucket CORS success",
		"bucket_name", input.BucketName,
		"user_id", input.UserID,
	)

	return nil
}

// ==================== ACL相关方法 ====================

// PutBucketACLInput 设置Bucket ACL输入参数
type PutBucketACLInput struct {
	BucketName string
	UserID     string
	ACL        *types.AccessControlPolicy
}

// PutBucketACL 设置Bucket ACL配置
func (s *S3BucketService) PutBucketACL(ctx context.Context, input *PutBucketACLInput) error {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return types.ErrBucketNotFoundError
		}
		return err
	}

	// 2. 验证ACL配置
	if input.ACL == nil {
		return fmt.Errorf("ACL configuration is required")
	}

	// 3. 验证Owner（必须是Bucket所有者）
	if input.ACL.Owner.ID == "" || input.ACL.Owner.ID != input.UserID {
		return fmt.Errorf("ACL owner must be the bucket owner")
	}

	// 4. 验证Grants
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

	// 5. 序列化ACL配置为JSON
	aclJSON, err := json.Marshal(input.ACL)
	if err != nil {
		logger.LOG.Error("Marshal ACL config failed", "error", err)
		return fmt.Errorf("marshal ACL config failed: %w", err)
	}

	// 6. 创建或更新ACL配置
	aclConfig := &models.S3BucketACL{
		BucketName: input.BucketName,
		UserID:     input.UserID,
		ACLConfig:  string(aclJSON),
		UpdatedAt:  custom_type.Now(),
	}

	if err := s.factory.S3ACL().CreateOrUpdateBucketACL(ctx, aclConfig); err != nil {
		logger.LOG.Error("Create or update bucket ACL failed",
			"bucket_name", input.BucketName,
			"user_id", input.UserID,
			"error", err,
		)
		return err
	}

	logger.LOG.Info("Put bucket ACL success",
		"bucket_name", input.BucketName,
		"user_id", input.UserID,
		"grants_count", len(input.ACL.AccessControlList.Grants),
	)

	return nil
}

// GetBucketACLInput 获取Bucket ACL输入参数
type GetBucketACLInput struct {
	BucketName string
	UserID     string
}

// GetBucketACLOutput 获取Bucket ACL输出
type GetBucketACLOutput struct {
	ACL *types.AccessControlPolicy
}

// GetBucketACL 获取Bucket ACL配置
func (s *S3BucketService) GetBucketACL(ctx context.Context, input *GetBucketACLInput) (*GetBucketACLOutput, error) {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrBucketNotFoundError
		}
		return nil, err
	}

	// 2. 获取ACL配置
	acl, err := s.factory.S3ACL().GetBucketACL(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有ACL配置，返回默认ACL（私有，只有所有者有权限）
			defaultACL := &types.AccessControlPolicy{
				Owner: types.Owner{
					ID:          input.UserID,
					DisplayName: input.UserID, // 可以使用用户名，这里简化使用UserID
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
			return &GetBucketACLOutput{ACL: defaultACL}, nil
		}
		return nil, err
	}

	// 3. 反序列化ACL配置
	var aclPolicy types.AccessControlPolicy
	if err := json.Unmarshal([]byte(acl.ACLConfig), &aclPolicy); err != nil {
		logger.LOG.Error("Unmarshal ACL config failed", "error", err)
		return nil, fmt.Errorf("unmarshal ACL config failed: %w", err)
	}

	return &GetBucketACLOutput{
		ACL: &aclPolicy,
	}, nil
}

// ==================== Bucket Policy相关方法 ====================

// PutBucketPolicyInput 设置Bucket Policy输入参数
type PutBucketPolicyInput struct {
	BucketName string
	UserID     string
	Policy     *types.BucketPolicy
}

// PutBucketPolicy 设置Bucket Policy配置
func (s *S3BucketService) PutBucketPolicy(ctx context.Context, input *PutBucketPolicyInput) error {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return types.ErrBucketNotFoundError
		}
		return err
	}

	// 2. 验证Policy配置
	if input.Policy == nil {
		return fmt.Errorf("policy configuration is required")
	}

	// 3. 验证Policy JSON格式
	if err := types.ValidateBucketPolicy(input.Policy); err != nil {
		logger.LOG.Error("Invalid bucket policy", "error", err)
		return fmt.Errorf("invalid policy: %w", err)
	}

	// 4. 序列化Policy配置为JSON
	policyJSON, err := json.Marshal(input.Policy)
	if err != nil {
		logger.LOG.Error("Marshal policy config failed", "error", err)
		return fmt.Errorf("marshal policy config failed: %w", err)
	}

	// 5. 创建或更新Policy配置
	policyConfig := &models.S3BucketPolicy{
		BucketName: input.BucketName,
		UserID:     input.UserID,
		PolicyJSON: string(policyJSON),
		UpdatedAt:  custom_type.Now(),
	}

	if err := s.factory.S3BucketPolicy().CreateOrUpdate(ctx, policyConfig); err != nil {
		logger.LOG.Error("Create or update bucket policy failed",
			"bucket_name", input.BucketName,
			"user_id", input.UserID,
			"error", err,
		)
		return err
	}

	logger.LOG.Info("Put bucket policy success",
		"bucket_name", input.BucketName,
		"user_id", input.UserID,
		"statements_count", len(input.Policy.Statement),
	)

	return nil
}

// GetBucketPolicyInput 获取Bucket Policy输入参数
type GetBucketPolicyInput struct {
	BucketName string
	UserID     string
}

// GetBucketPolicyOutput 获取Bucket Policy输出
type GetBucketPolicyOutput struct {
	Policy *types.BucketPolicy
}

// GetBucketPolicy 获取Bucket Policy配置
func (s *S3BucketService) GetBucketPolicy(ctx context.Context, input *GetBucketPolicyInput) (*GetBucketPolicyOutput, error) {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrBucketNotFoundError
		}
		return nil, err
	}

	// 2. 获取Policy配置
	policy, err := s.factory.S3BucketPolicy().GetByBucket(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrPolicyNotFoundError
		}
		return nil, err
	}

	// 3. 反序列化Policy配置
	var bucketPolicy types.BucketPolicy
	if err := json.Unmarshal([]byte(policy.PolicyJSON), &bucketPolicy); err != nil {
		logger.LOG.Error("Unmarshal policy config failed", "error", err)
		return nil, fmt.Errorf("unmarshal policy config failed: %w", err)
	}

	return &GetBucketPolicyOutput{
		Policy: &bucketPolicy,
	}, nil
}

// DeleteBucketPolicyInput 删除Bucket Policy输入参数
type DeleteBucketPolicyInput struct {
	BucketName string
	UserID     string
}

// DeleteBucketPolicy 删除Bucket Policy配置
func (s *S3BucketService) DeleteBucketPolicy(ctx context.Context, input *DeleteBucketPolicyInput) error {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return types.ErrBucketNotFoundError
		}
		return err
	}

	// 2. 删除Policy配置
	if err := s.factory.S3BucketPolicy().Delete(ctx, input.BucketName, input.UserID); err != nil {
		logger.LOG.Error("Delete bucket policy failed",
			"bucket_name", input.BucketName,
			"user_id", input.UserID,
			"error", err,
		)
		return err
	}

	logger.LOG.Info("Delete bucket policy success",
		"bucket_name", input.BucketName,
		"user_id", input.UserID,
	)

	return nil
}

// ==================== Lifecycle相关方法 ====================

// PutBucketLifecycleInput 设置Bucket Lifecycle输入参数
type PutBucketLifecycleInput struct {
	BucketName string
	UserID     string
	Lifecycle  *types.LifecycleConfiguration
}

// PutBucketLifecycle 设置Bucket Lifecycle配置
func (s *S3BucketService) PutBucketLifecycle(ctx context.Context, input *PutBucketLifecycleInput) error {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return types.ErrBucketNotFoundError
		}
		return err
	}

	// 2. 验证Lifecycle配置
	if input.Lifecycle == nil {
		return fmt.Errorf("lifecycle configuration is required")
	}

	// 3. 验证Lifecycle JSON格式
	if err := types.ValidateLifecycleConfiguration(input.Lifecycle); err != nil {
		logger.LOG.Error("Invalid lifecycle configuration", "error", err)
		return fmt.Errorf("invalid lifecycle: %w", err)
	}

	// 4. 序列化Lifecycle配置为JSON
	lifecycleJSON, err := json.Marshal(input.Lifecycle)
	if err != nil {
		logger.LOG.Error("Marshal lifecycle config failed", "error", err)
		return fmt.Errorf("marshal lifecycle config failed: %w", err)
	}

	// 5. 创建或更新Lifecycle配置
	lifecycleConfig := &models.S3BucketLifecycle{
		BucketName:    input.BucketName,
		UserID:        input.UserID,
		LifecycleJSON: string(lifecycleJSON),
		UpdatedAt:     custom_type.Now(),
	}

	if err := s.factory.S3BucketLifecycle().CreateOrUpdate(ctx, lifecycleConfig); err != nil {
		logger.LOG.Error("Create or update bucket lifecycle failed",
			"bucket_name", input.BucketName,
			"user_id", input.UserID,
			"error", err,
		)
		return err
	}

	logger.LOG.Info("Put bucket lifecycle success",
		"bucket_name", input.BucketName,
		"user_id", input.UserID,
		"rules_count", len(input.Lifecycle.Rules),
	)

	return nil
}

// GetBucketLifecycleInput 获取Bucket Lifecycle输入参数
type GetBucketLifecycleInput struct {
	BucketName string
	UserID     string
}

// GetBucketLifecycleOutput 获取Bucket Lifecycle输出
type GetBucketLifecycleOutput struct {
	Lifecycle *types.LifecycleConfiguration
}

// GetBucketLifecycle 获取Bucket Lifecycle配置
func (s *S3BucketService) GetBucketLifecycle(ctx context.Context, input *GetBucketLifecycleInput) (*GetBucketLifecycleOutput, error) {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrBucketNotFoundError
		}
		return nil, err
	}

	// 2. 获取Lifecycle配置
	lifecycle, err := s.factory.S3BucketLifecycle().GetByBucket(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrLifecycleNotFoundError
		}
		return nil, err
	}

	// 3. 反序列化Lifecycle配置
	var lifecycleConfig types.LifecycleConfiguration
	if err := json.Unmarshal([]byte(lifecycle.LifecycleJSON), &lifecycleConfig); err != nil {
		logger.LOG.Error("Unmarshal lifecycle config failed", "error", err)
		return nil, fmt.Errorf("unmarshal lifecycle config failed: %w", err)
	}

	return &GetBucketLifecycleOutput{
		Lifecycle: &lifecycleConfig,
	}, nil
}

// DeleteBucketLifecycleInput 删除Bucket Lifecycle输入参数
type DeleteBucketLifecycleInput struct {
	BucketName string
	UserID     string
}

// DeleteBucketLifecycle 删除Bucket Lifecycle配置
func (s *S3BucketService) DeleteBucketLifecycle(ctx context.Context, input *DeleteBucketLifecycleInput) error {
	// 1. 验证Bucket是否存在
	_, err := s.bucketRepo.GetByName(ctx, input.BucketName, input.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return types.ErrBucketNotFoundError
		}
		return err
	}

	// 2. 删除Lifecycle配置
	if err := s.factory.S3BucketLifecycle().Delete(ctx, input.BucketName, input.UserID); err != nil {
		logger.LOG.Error("Delete bucket lifecycle failed",
			"bucket_name", input.BucketName,
			"user_id", input.UserID,
			"error", err,
		)
		return err
	}

	logger.LOG.Info("Delete bucket lifecycle success",
		"bucket_name", input.BucketName,
		"user_id", input.UserID,
	)

	return nil
}

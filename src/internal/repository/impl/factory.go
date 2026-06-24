package impl

import (
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

// RepositoryFactory 仓储工厂
type RepositoryFactory struct {
	db *gorm.DB

	userRepo         repository.UserRepository
	fileInfoRepo     repository.FileInfoRepository
	groupRepo        repository.GroupRepository
	shareRepo        repository.ShareRepository
	diskRepo         repository.DiskRepository
	apiKeyRepo       repository.ApiKeyRepository
	fileChunkRepo    repository.FileChunkRepository
	powerRepo        repository.PowerRepository
	groupPowerRepo   repository.GroupPowerRepository
	userFilesRepo    repository.UserFilesRepository
	virtualPathRepo  repository.VirtualPathRepository
	recycledRepo     repository.RecycledRepository
	downloadTaskRepo repository.DownloadTaskRepository
	sysConfigRepo    repository.SysConfigRepository
	uploadChunkRepo  repository.UploadChunkRepository
	uploadTaskRepo   repository.UploadTaskRepository

	// 云盘任务仓储
	cloudTaskRepo     *CloudTaskRepository
	cloudTaskFileRepo *CloudTaskFileRepository

	// S3相关仓储
	s3BucketRepo         repository.S3BucketRepository
	s3ObjectMetadataRepo repository.S3ObjectMetadataRepository
	s3MultipartRepo      repository.S3MultipartRepository
	// 以下S3仓储使用具体类型而非接口，因为它们仅在S3模块内部使用，
	// 暂未抽取公共接口（S3模块独立于主业务，接口隔离成本高于收益）。
	s3BucketCORSRepo      *S3BucketCORSRepositoryImpl
	s3ACLRepo             *S3ACLRepositoryImpl
	s3BucketPolicyRepo    *S3BucketPolicyRepositoryImpl
	s3BucketLifecycleRepo *S3BucketLifecycleRepositoryImpl
	s3EncryptionKeyRepo   *S3EncryptionKeyRepositoryImpl
	s3ObjectEncryptionRepo *S3ObjectEncryptionRepositoryImpl
}

// NewRepositoryFactory 创建仓储工厂实例（立即初始化所有仓储，确保并发安全）
func NewRepositoryFactory(db *gorm.DB) *RepositoryFactory {
	return &RepositoryFactory{
		db: db,

		userRepo:         NewUserRepository(db),
		fileInfoRepo:     NewFileInfoRepository(db),
		groupRepo:        NewGroupRepository(db),
		shareRepo:        NewShareRepository(db),
		diskRepo:         NewDiskRepository(db),
		apiKeyRepo:       NewApiKeyRepository(db),
		fileChunkRepo:    NewFileChunkRepository(db),
		powerRepo:        NewPowerRepository(db),
		groupPowerRepo:   NewGroupPowerRepository(db),
		userFilesRepo:    NewUserFilesRepository(db),
		virtualPathRepo:  NewVirtualPathRepository(db),
		recycledRepo:     NewRecycledRepository(db),
		downloadTaskRepo: NewDownloadTaskRepository(db),
		sysConfigRepo:    NewSysConfigRepository(db),
		uploadChunkRepo:  NewUploadChunkRepository(db),
		uploadTaskRepo:   NewUploadTaskRepository(db),

		// 云盘任务仓储
		cloudTaskRepo:     NewCloudTaskRepository(db),
		cloudTaskFileRepo: NewCloudTaskFileRepository(db),

		s3BucketRepo:          NewS3BucketRepository(db),
		s3ObjectMetadataRepo:  NewS3ObjectMetadataRepository(db),
		s3MultipartRepo:       NewS3MultipartRepository(db),
		s3BucketCORSRepo:      NewS3BucketCORSRepository(db),
		s3ACLRepo:             NewS3ACLRepository(db),
		s3BucketPolicyRepo:    NewS3BucketPolicyRepository(db),
		s3BucketLifecycleRepo: NewS3BucketLifecycleRepository(db),
		s3EncryptionKeyRepo:   NewS3EncryptionKeyRepository(db),
		s3ObjectEncryptionRepo: NewS3ObjectEncryptionRepository(db),
	}
}

// CloudTask 获取云盘任务仓储
func (f *RepositoryFactory) CloudTask() *CloudTaskRepository {
	return f.cloudTaskRepo
}

// CloudTaskFile 获取云盘任务文件仓储
func (f *RepositoryFactory) CloudTaskFile() *CloudTaskFileRepository {
	return f.cloudTaskFileRepo
}

// User 获取用户仓储
func (f *RepositoryFactory) User() repository.UserRepository {
	return f.userRepo
}

// FileInfo 获取文件信息仓储
func (f *RepositoryFactory) FileInfo() repository.FileInfoRepository {
	return f.fileInfoRepo
}

// Group 获取组仓储
func (f *RepositoryFactory) Group() repository.GroupRepository {
	return f.groupRepo
}

// Share 获取分享仓储
func (f *RepositoryFactory) Share() repository.ShareRepository {
	return f.shareRepo
}

// Disk 获取磁盘仓储
func (f *RepositoryFactory) Disk() repository.DiskRepository {
	return f.diskRepo
}

// ApiKey 获取API密钥仓储
func (f *RepositoryFactory) ApiKey() repository.ApiKeyRepository {
	return f.apiKeyRepo
}

// FileChunk 获取文件分片仓储
func (f *RepositoryFactory) FileChunk() repository.FileChunkRepository {
	return f.fileChunkRepo
}

// Power 获取权限仓储
func (f *RepositoryFactory) Power() repository.PowerRepository {
	return f.powerRepo
}

// GroupPower 获取组权限关联仓储
func (f *RepositoryFactory) GroupPower() repository.GroupPowerRepository {
	return f.groupPowerRepo
}

// UserFiles 获取用户文件关联仓储
func (f *RepositoryFactory) UserFiles() repository.UserFilesRepository {
	return f.userFilesRepo
}

// VirtualPath 获取虚拟路径仓储
func (f *RepositoryFactory) VirtualPath() repository.VirtualPathRepository {
	return f.virtualPathRepo
}

// Recycled 获取回收站仓储
func (f *RepositoryFactory) Recycled() repository.RecycledRepository {
	return f.recycledRepo
}

// DownloadTask 获取下载任务仓储
func (f *RepositoryFactory) DownloadTask() repository.DownloadTaskRepository {
	return f.downloadTaskRepo
}

// SysConfig 获取系统配置仓储
func (f *RepositoryFactory) SysConfig() repository.SysConfigRepository {
	return f.sysConfigRepo
}

// UploadChunk 获取上传分片信息仓储
func (f *RepositoryFactory) UploadChunk() repository.UploadChunkRepository {
	return f.uploadChunkRepo
}

// UploadTask 获取上传任务仓储
func (f *RepositoryFactory) UploadTask() repository.UploadTaskRepository {
	return f.uploadTaskRepo
}

// S3Bucket 获取S3 Bucket仓储
func (f *RepositoryFactory) S3Bucket() repository.S3BucketRepository {
	return f.s3BucketRepo
}

// S3ObjectMetadata 获取S3对象元数据仓储
func (f *RepositoryFactory) S3ObjectMetadata() repository.S3ObjectMetadataRepository {
	return f.s3ObjectMetadataRepo
}

// S3Multipart 获取S3分片上传仓储
func (f *RepositoryFactory) S3Multipart() repository.S3MultipartRepository {
	return f.s3MultipartRepo
}

// S3BucketCORS 获取S3 Bucket CORS配置仓储
func (f *RepositoryFactory) S3BucketCORS() *S3BucketCORSRepositoryImpl {
	return f.s3BucketCORSRepo
}

// S3ACL 获取S3 ACL配置仓储
func (f *RepositoryFactory) S3ACL() *S3ACLRepositoryImpl {
	return f.s3ACLRepo
}

// S3BucketPolicy 获取S3 Bucket Policy配置仓储
func (f *RepositoryFactory) S3BucketPolicy() *S3BucketPolicyRepositoryImpl {
	return f.s3BucketPolicyRepo
}

// S3BucketLifecycle 获取S3 Bucket Lifecycle配置仓储
func (f *RepositoryFactory) S3BucketLifecycle() *S3BucketLifecycleRepositoryImpl {
	return f.s3BucketLifecycleRepo
}

// S3EncryptionKey 获取S3加密密钥仓储
func (f *RepositoryFactory) S3EncryptionKey() *S3EncryptionKeyRepositoryImpl {
	return f.s3EncryptionKeyRepo
}

// S3ObjectEncryption 获取S3对象加密元数据仓储
func (f *RepositoryFactory) S3ObjectEncryption() *S3ObjectEncryptionRepositoryImpl {
	return f.s3ObjectEncryptionRepo
}

// DB 获取数据库实例（用于事务操作）
func (f *RepositoryFactory) DB() *gorm.DB {
	return f.db
}

// WithTx 创建基于事务的新工厂实例
func (f *RepositoryFactory) WithTx(tx *gorm.DB) *RepositoryFactory {
	return NewRepositoryFactory(tx)
}

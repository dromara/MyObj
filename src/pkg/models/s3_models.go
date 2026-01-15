package models

import "myobj/src/pkg/custom_type"

// S3Bucket S3存储桶模型（对应用户虚拟目录）
type S3Bucket struct {
	ID            int                  `gorm:"primaryKey;autoIncrement" json:"id"`
	BucketName    string               `gorm:"uniqueIndex:idx_bucket_user;size:63;not null" json:"bucket_name"` // Bucket名称（符合S3命名规范）
	UserID        string               `gorm:"uniqueIndex:idx_bucket_user;size:36;not null;index" json:"user_id"`
	Region        string               `gorm:"size:32;default:'us-east-1'" json:"region"`
	VirtualPathID int                  `gorm:"index;not null" json:"virtual_path_id"` // 关联到虚拟路径ID
	Versioning    string               `gorm:"size:16;default:'Disabled'" json:"versioning"` // 版本控制状态：Enabled/Suspended/Disabled
	CreatedAt     custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt     custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"updated_at"`
}

func (S3Bucket) TableName() string {
	return "s3_buckets"
}

// S3ObjectMetadata S3对象元数据（扩展FileInfo）
type S3ObjectMetadata struct {
	ID           int                  `gorm:"primaryKey;autoIncrement" json:"id"`
	FileID       string               `gorm:"size:36;index" json:"file_id"` // 关联FileInfo.ID（DeleteMarker时为空）
	BucketName   string               `gorm:"index:idx_bucket_key;size:63;not null" json:"bucket_name"`
	ObjectKey    string               `gorm:"index:idx_bucket_key;size:1024;not null" json:"object_key"` // S3对象键名
	UserID       string               `gorm:"size:36;not null;index" json:"user_id"`
	ETag         string               `gorm:"size:64" json:"etag"` // MD5或BLAKE3哈希（DeleteMarker时为空）
	StorageClass string               `gorm:"size:32;default:'STANDARD'" json:"storage_class"`
	ContentType  string               `gorm:"size:256" json:"content_type"`
	UserMetadata string               `gorm:"type:text" json:"user_metadata"`  // JSON格式存储x-amz-meta-*
	Tags         string               `gorm:"type:text" json:"tags"` // JSON格式存储对象标签
	VersionID    string               `gorm:"index;size:36" json:"version_id"` // 版本控制ID
	IsLatest     bool                 `gorm:"default:true;index" json:"is_latest"`
	IsDeleteMarker bool               `gorm:"default:false;index" json:"is_delete_marker"` // 是否为删除标记
	CreatedAt    custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt    custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"updated_at"`
}

func (S3ObjectMetadata) TableName() string {
	return "s3_object_metadata"
}

// S3MultipartUpload 分片上传会话
type S3MultipartUpload struct {
	UploadID   string               `gorm:"primaryKey;size:64" json:"upload_id"`
	BucketName string               `gorm:"index;size:63;not null" json:"bucket_name"`
	ObjectKey  string               `gorm:"index;size:1024;not null" json:"object_key"`
	UserID     string               `gorm:"index;size:36;not null" json:"user_id"`
	Metadata   string               `gorm:"type:text" json:"metadata"`                   // JSON格式元数据
	Status     string               `gorm:"size:32;default:'in-progress'" json:"status"` // in-progress/completed/aborted
	CreatedAt  custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt  custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"updated_at"`
}

func (S3MultipartUpload) TableName() string {
	return "s3_multipart_uploads"
}

// S3MultipartPart 分片信息
type S3MultipartPart struct {
	ID         int                  `gorm:"primaryKey;autoIncrement" json:"id"`
	UploadID   string               `gorm:"index:idx_upload_part;size:64;not null" json:"upload_id"`
	PartNumber int                  `gorm:"index:idx_upload_part;not null" json:"part_number"`
	ETag       string               `gorm:"size:64;not null" json:"etag"`
	Size       int64                `gorm:"not null" json:"size"`
	ChunkPath  string               `gorm:"size:512" json:"chunk_path"` // 临时分片路径
	CreatedAt  custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
}

func (S3MultipartPart) TableName() string {
	return "s3_multipart_parts"
}

// S3BucketCORS Bucket的CORS配置
type S3BucketCORS struct {
	ID         int                  `gorm:"primaryKey;autoIncrement" json:"id"`
	BucketName string               `gorm:"uniqueIndex:idx_bucket_cors;size:63;not null" json:"bucket_name"`
	UserID     string               `gorm:"uniqueIndex:idx_bucket_cors;size:36;not null;index" json:"user_id"`
	CORSConfig string               `gorm:"type:text;not null" json:"cors_config"` // JSON格式存储CORS规则
	CreatedAt  custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt  custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"updated_at"`
}

func (S3BucketCORS) TableName() string {
	return "s3_bucket_cors"
}

// S3BucketACL Bucket的ACL配置
type S3BucketACL struct {
	ID         int                  `gorm:"primaryKey;autoIncrement" json:"id"`
	BucketName string               `gorm:"uniqueIndex:idx_bucket_acl;size:63;not null" json:"bucket_name"`
	UserID     string               `gorm:"uniqueIndex:idx_bucket_acl;size:36;not null;index" json:"user_id"`
	ACLConfig  string               `gorm:"type:text;not null" json:"acl_config"` // JSON格式存储ACL配置
	CreatedAt  custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt  custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"updated_at"`
}

func (S3BucketACL) TableName() string {
	return "s3_bucket_acl"
}

// S3ObjectACL Object的ACL配置
type S3ObjectACL struct {
	ID         int                  `gorm:"primaryKey;autoIncrement" json:"id"`
	BucketName string               `gorm:"index:idx_object_acl;size:63;not null" json:"bucket_name"`
	ObjectKey  string               `gorm:"index:idx_object_acl;size:1024;not null" json:"object_key"`
	VersionID  string               `gorm:"index:idx_object_acl;size:36" json:"version_id"` // 版本ID（支持版本控制）
	UserID     string               `gorm:"index:idx_object_acl;size:36;not null;index" json:"user_id"`
	ACLConfig  string               `gorm:"type:text;not null" json:"acl_config"` // JSON格式存储ACL配置
	CreatedAt  custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt  custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"updated_at"`
}

func (S3ObjectACL) TableName() string {
	return "s3_object_acl"
}

// S3BucketPolicy Bucket的策略配置
type S3BucketPolicy struct {
	ID         int                  `gorm:"primaryKey;autoIncrement" json:"id"`
	BucketName string               `gorm:"uniqueIndex:idx_bucket_policy;size:63;not null" json:"bucket_name"`
	UserID     string               `gorm:"uniqueIndex:idx_bucket_policy;size:36;not null;index" json:"user_id"`
	PolicyJSON string               `gorm:"type:text;not null" json:"policy_json"` // JSON格式存储Bucket Policy
	CreatedAt  custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt  custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"updated_at"`
}

func (S3BucketPolicy) TableName() string {
	return "s3_bucket_policy"
}

// S3BucketLifecycle Bucket的生命周期配置
type S3BucketLifecycle struct {
	ID            int                  `gorm:"primaryKey;autoIncrement" json:"id"`
	BucketName    string               `gorm:"uniqueIndex:idx_bucket_lifecycle;size:63;not null" json:"bucket_name"`
	UserID        string               `gorm:"uniqueIndex:idx_bucket_lifecycle;size:36;not null;index" json:"user_id"`
	LifecycleJSON string               `gorm:"type:text;not null" json:"lifecycle_json"` // JSON格式存储Lifecycle规则
	CreatedAt     custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt     custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"updated_at"`
}

func (S3BucketLifecycle) TableName() string {
	return "s3_bucket_lifecycle"
}

// S3EncryptionKey S3加密密钥（用于SSE-S3）
type S3EncryptionKey struct {
	ID        int                  `gorm:"primaryKey;autoIncrement" json:"id"`
	KeyID     string               `gorm:"uniqueIndex:idx_key_id;size:64;not null" json:"key_id"` // 密钥ID（用于标识）
	KeyData   string               `gorm:"type:text;not null" json:"key_data"`                   // 加密后的密钥数据（base64）
	Algorithm string               `gorm:"size:32;default:'AES256'" json:"algorithm"`            // 加密算法（AES256等）
	CreatedAt custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"updated_at"`
}

func (S3EncryptionKey) TableName() string {
	return "s3_encryption_keys"
}

// S3ObjectEncryption 对象加密元数据
type S3ObjectEncryption struct {
	ID              int                  `gorm:"primaryKey;autoIncrement" json:"id"`
	BucketName      string               `gorm:"index:idx_bucket_key;size:63;not null" json:"bucket_name"`
	ObjectKey       string               `gorm:"index:idx_bucket_key;size:1024;not null" json:"object_key"`
	VersionID       string               `gorm:"index:idx_bucket_key;size:36" json:"version_id"` // 版本ID（支持版本控制）
	UserID          string               `gorm:"index;size:36;not null" json:"user_id"`
	EncryptionType  string               `gorm:"size:32;not null" json:"encryption_type"` // SSE-S3, SSE-C, SSE-KMS
	Algorithm       string               `gorm:"size:32;default:'AES256'" json:"algorithm"` // AES256等
	KeyID           string               `gorm:"size:64" json:"key_id"`                    // 密钥ID（SSE-S3或SSE-KMS）
	EncryptedKey    string               `gorm:"type:text" json:"encrypted_key"`            // 加密的密钥（SSE-C时使用）
	IV              string               `gorm:"size:64" json:"iv"`                        // 初始化向量（base64）
	CreatedAt       custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt       custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"updated_at"`
}

func (S3ObjectEncryption) TableName() string {
	return "s3_object_encryption"
}

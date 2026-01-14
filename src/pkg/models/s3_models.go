package models

import "myobj/src/pkg/custom_type"

// S3Bucket S3存储桶模型（对应用户虚拟目录）
type S3Bucket struct {
	ID            int                  `gorm:"primaryKey;autoIncrement" json:"id"`
	BucketName    string               `gorm:"uniqueIndex:idx_bucket_user;size:63;not null" json:"bucket_name"` // Bucket名称（符合S3命名规范）
	UserID        string               `gorm:"uniqueIndex:idx_bucket_user;size:36;not null;index" json:"user_id"`
	Region        string               `gorm:"size:32;default:'us-east-1'" json:"region"`
	VirtualPathID int                  `gorm:"index;not null" json:"virtual_path_id"` // 关联到虚拟路径ID
	CreatedAt     custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt     custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"updated_at"`
}

func (S3Bucket) TableName() string {
	return "s3_buckets"
}

// S3ObjectMetadata S3对象元数据（扩展FileInfo）
type S3ObjectMetadata struct {
	ID           int                  `gorm:"primaryKey;autoIncrement" json:"id"`
	FileID       string               `gorm:"size:36;not null;index" json:"file_id"` // 关联FileInfo.ID
	BucketName   string               `gorm:"index:idx_bucket_key;size:63;not null" json:"bucket_name"`
	ObjectKey    string               `gorm:"index:idx_bucket_key;size:1024;not null" json:"object_key"` // S3对象键名
	UserID       string               `gorm:"size:36;not null;index" json:"user_id"`
	ETag         string               `gorm:"size:64" json:"etag"` // MD5或BLAKE3哈希
	StorageClass string               `gorm:"size:32;default:'STANDARD'" json:"storage_class"`
	ContentType  string               `gorm:"size:256" json:"content_type"`
	UserMetadata string               `gorm:"type:text" json:"user_metadata"`  // JSON格式存储x-amz-meta-*
	VersionID    string               `gorm:"index;size:36" json:"version_id"` // 版本控制ID
	IsLatest     bool                 `gorm:"default:true;index" json:"is_latest"`
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

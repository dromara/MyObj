package models

import (
	"myobj/src/pkg/custom_type"
)

// UploadTask 上传任务表（用于持久化上传任务，支持断点续传）
type UploadTask struct {
	// 任务ID（使用 precheck_id 作为主键）
	ID string `gorm:"column:id;type:text;primaryKey" json:"id"`
	// 用户ID
	UserID string `gorm:"column:user_id;type:text;index" json:"user_id"`
	// 文件名
	FileName string `gorm:"column:file_name;type:text;not null" json:"file_name"`
	// 文件大小（字节）
	FileSize int64 `gorm:"column:file_size;type:integer;not null" json:"file_size"`
	// 分片大小（字节，默认5MB）
	ChunkSize int64 `gorm:"column:chunk_size;type:integer;not null;default:5242880" json:"chunk_size"`
	// 总分片数
	TotalChunks int `gorm:"column:total_chunks;type:integer;not null" json:"total_chunks"`
	// 已上传分片数
	UploadedChunks int `gorm:"column:uploaded_chunks;type:integer;default:0" json:"uploaded_chunks"`
	// 文件hash签名（用于秒传检测）
	ChunkSignature string `gorm:"column:chunk_signature;type:text" json:"chunk_signature"`
	// 路径ID
	PathID string `gorm:"column:path_id;type:text" json:"path_id"`
	// 临时目录路径
	TempDir string `gorm:"column:temp_dir;type:text" json:"temp_dir"`
	// 任务状态（pending/uploading/completed/failed/aborted）
	Status string `gorm:"column:status;type:text;default:'pending'" json:"status"`
	// 错误信息
	ErrorMessage string `gorm:"column:error_message;type:text" json:"error_message"`
	// 创建时间
	CreateTime custom_type.JsonTime `gorm:"column:create_time;type:datetime" json:"create_time"`
	// 更新时间
	UpdateTime custom_type.JsonTime `gorm:"column:update_time;type:datetime" json:"update_time"`
	// 过期时间（7天后自动清理）
	ExpireTime custom_type.JsonTime `gorm:"column:expire_time;type:datetime" json:"expire_time"`
}

func (UploadTask) TableName() string {
	return "upload_task"
}


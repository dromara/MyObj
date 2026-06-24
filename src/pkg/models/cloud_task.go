package models

import (
	"myobj/src/pkg/custom_type"
)

// CloudTask 云盘任务
type CloudTask struct {
	ID          int                  `gorm:"type:INTEGER;not null;primaryKey;unique;autoIncrement" json:"id"`
	UserID      string               `gorm:"type:VARCHAR;not null;index" json:"user_id"`
	TaskType    string               `gorm:"type:VARCHAR;not null" json:"task_type"` // parse, download, save
	Provider    string               `gorm:"type:VARCHAR;not null" json:"provider"`
	ShareID     string               `gorm:"type:VARCHAR" json:"share_id"`
	ShareURL    string               `gorm:"type:TEXT" json:"share_url"`
	SharePwd    string               `gorm:"type:VARCHAR" json:"-"`
	Status      int                  `gorm:"type:INTEGER;not null;default:0" json:"status"`
	FileCount   int                  `gorm:"type:INTEGER" json:"file_count"`
	TotalSize   int64                `gorm:"type:BIGINT" json:"total_size"`
	SuccessCount int                 `gorm:"type:INTEGER;default:0" json:"success_count"`
	FailedCount  int                 `gorm:"type:INTEGER;default:0" json:"failed_count"`
	TargetPath  string               `gorm:"type:VARCHAR" json:"target_path"`
	ErrorMsg    string               `gorm:"type:TEXT" json:"error_msg"`
	CreatedAt   custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt   custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"updated_at"`
	CompletedAt *custom_type.JsonTime `gorm:"type:DATETIME" json:"completed_at"`
}

func (CloudTask) TableName() string {
	return "cloud_tasks"
}

// CloudTaskFile 云盘任务文件
type CloudTaskFile struct {
	ID         int                  `gorm:"type:INTEGER;not null;primaryKey;unique;autoIncrement" json:"id"`
	TaskID     int                  `gorm:"type:INTEGER;not null;index" json:"task_id"`
	FileID     string               `gorm:"type:VARCHAR;not null" json:"file_id"`
	FileName   string               `gorm:"type:VARCHAR;not null" json:"file_name"`
	FileSize   int64                `gorm:"type:BIGINT" json:"file_size"`
	IsDir      bool                 `gorm:"type:BOOLEAN;default:false" json:"is_dir"`
	FileType   string               `gorm:"type:VARCHAR" json:"file_type"`
	Status     int                  `gorm:"type:INTEGER;not null;default:0" json:"status"`
	LocalPath  string               `gorm:"type:VARCHAR" json:"local_path"`
	LocalFileID string              `gorm:"type:VARCHAR" json:"local_file_id"`
	ErrorMsg   string               `gorm:"type:TEXT" json:"error_msg"`
	CreatedAt  custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt  custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"updated_at"`
}

func (CloudTaskFile) TableName() string {
	return "cloud_task_files"
}

// 任务状态常量
const (
	CloudTaskStatusPending    = 0 // 待处理
	CloudTaskStatusProcessing = 1 // 处理中
	CloudTaskStatusCompleted  = 2 // 完成
	CloudTaskStatusFailed     = 3 // 失败
	CloudTaskStatusCancelled  = 4 // 取消
)

// 文件状态常量
const (
	CloudFileStatusPending    = 0 // 待处理
	CloudFileStatusDownloading = 1 // 下载中
	CloudFileStatusCompleted  = 2 // 完成
	CloudFileStatusFailed     = 3 // 失败
)

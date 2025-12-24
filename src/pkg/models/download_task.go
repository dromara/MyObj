package models

import (
	"myobj/src/pkg/custom_type"
)

// DownloadTask 下载任务表
type DownloadTask struct {
	// 任务 ID
	ID string `gorm:"column:id;type:text;primaryKey"`
	// 用户 ID
	UserID string `gorm:"column:user_id;type:text;index"`
	// 文件 ID
	FileID string `gorm:"column:file_id;type:text"`
	// 文件名
	FileName string `gorm:"column:file_name;type:text"`
	// 文件大小
	FileSize int64 `gorm:"column:file_size;type:integer"`
	// 已下载大小
	DownloadedSize int64 `gorm:"column:downloaded_size;type:integer;default:0"`
	// 下载进度 (0-100)
	Progress int `gorm:"column:progress;type:integer;default:0"`
	// 下载速度 (字节/秒)
	Speed int64 `gorm:"column:speed;type:integer;default:0"`
	// 任务类型
	Type int `gorm:"column:type;type:integer;not null"`
	// 下载URL
	URL string `gorm:"column:url;type:text"`
	// 下载路径
	Path string `gorm:"column:path;type:text"`
	// 虚拟路径
	VirtualPath string `gorm:"column:virtual_path;type:text"`
	// 任务状态
	State int `gorm:"column:state;type:integer"`
	// 错误信息
	ErrorMsg string `gorm:"column:error_msg;type:text"`
	// 目标临时目录
	TargetDir string `gorm:"column:target_dir;type:text"`
	// 是否支持断点续传
	SupportRange bool `gorm:"column:support_range;type:boolean;default:false"`
	// 是否加密存储
	EnableEncryption bool `gorm:"column:enable_encryption;type:boolean;default:false"`
	// 种子InfoHash（BT/磁力链任务）
	InfoHash string `gorm:"column:info_hash;type:text;index"`
	// 种子内文件索引（BT/磁力链任务）
	FileIndex int `gorm:"column:file_index;type:integer"`
	// 种子名称（BT/磁力链任务）
	TorrentName string `gorm:"column:torrent_name;type:text"`
	// 创建时间
	CreateTime custom_type.JsonTime `gorm:"column:create_time;type:datetime"`
	// 更新时间
	UpdateTime custom_type.JsonTime `gorm:"column:update_time;type:datetime"`
	// 完成时间
	FinishTime custom_type.JsonTime `gorm:"column:finish_time;type:datetime"`
}

func (DownloadTask) TableName() string {
	return "download_task"
}

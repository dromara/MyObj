package models

import (
	"myobj/src/pkg/custom_type"
)

// Share 分享记录
type Share struct {
	ID            int                  `gorm:"type:INTEGER;not null;primaryKey;unique" json:"id"` // 分享记录ID，主键且唯一
	UserID        string               `gorm:"type:VARCHAR;not null" json:"user_id"`              // 用户ID
	FileID        string               `gorm:"type:VARCHAR;not null" json:"file_id"`              // 文件ID
	Token         string               `gorm:"type:TEXT;not null" json:"token"`                   // 分享令牌
	ExpiresAt     custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"expires_at"`          // 分享过期时间
	PasswordHash  string               `gorm:"type:TEXT;not null" json:"password_hash"`           // 访问密码哈希
	DownloadCount int                  `gorm:"type:INTEGER;not null" json:"download_count"`       // 下载次数统计
	CreatedAt     custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`          // 分享创建时间
}

func (Share) TableName() string {
	return "shares"
}

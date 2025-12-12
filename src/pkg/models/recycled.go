package models

import "myobj/src/pkg/custom_type"

// Recycled 回收站表
type Recycled struct {
	// 回收站ID
	ID string `gorm:"type:VARCHAR;not null;primaryKey;unique" json:"id"`
	// 文件ID
	FileID string `gorm:"type:VARCHAR;not null" json:"file_id"`
	// 用户ID
	UserID string `gorm:"type:VARCHAR;not null" json:"user_id"`
	// 删除时间
	CreatedAt custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
}

func (Recycled) TableName() string {
	return "recycled"
}

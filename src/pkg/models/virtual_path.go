package models

import (
	"myobj/src/pkg/custom_type"
)

// VirtualPath 虚拟路径
type VirtualPath struct {
	ID          int                  `gorm:"type:INTEGER;not null;primaryKey;unique" json:"id"`  // 主键ID，唯一标识
	UserID      string               `gorm:"type:VARCHAR;not null" json:"user_id"`               // 用户ID，关联用户表
	Path        string               `gorm:"type:TEXT;not null" json:"path"`                     // 虚拟路径
	IsFile      bool                 `gorm:"type:BOOLEAN;default:false;not null" json:"is_file"` // 是否为文件，默认false
	IsDir       bool                 `gorm:"type:BOOLEAN;default:true;not null" json:"is_dir"`   // 是否为目录，默认true
	ParentLevel string               `gorm:"type:TEXT" json:"parent_level"`                      // 父级层级信息
	CreatedTime custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_time"`         // 创建时间
	UpdateTime  custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"update_time"`          // 更新时间
}

func (VirtualPath) TableName() string {
	return "virtual_path"
}

package models

import (
	"myobj/src/pkg/custom_type"
)

// Group 组
type Group struct {
	ID           int                  `gorm:"type:INTEGER;not null;primaryKey;unique" json:"id"` // 组ID，主键且唯一
	Name         string               `gorm:"type:VARCHAR;not null" json:"name"`                 // 组名称
	GroupDefault int                  `gorm:"type:INTEGER;not null" json:"group_default"`        // 是否为默认组 0-否 1-是
	CreatedAt    custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`          // 创建时间
	Space        int64                `gorm:"type:INTEGER" json:"space"`                         // 组默认可用存储空间
}

func (Group) TableName() string {
	return "groups"
}

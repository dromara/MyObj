package models

import (
	"myobj/src/pkg/custom_type"
)

// Power 权限表
type Power struct {
	ID             int                  `gorm:"type:INTEGER;not null;primaryKey;unique" json:"id"` // 权限ID，主键且唯一
	Name           string               `gorm:"type:VARCHAR;not null" json:"name"`                 // 权限名称
	Description    string               `gorm:"type:TEXT;not null" json:"description"`             // 权限描述
	Characteristic string               `gorm:"type:TEXT;not null" json:"characteristic"`          // 权限特征
	CreatedAt      custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`          // 创建时间
}

func (Power) TableName() string {
	return "power"
}

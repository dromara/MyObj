package models

import (
	"myobj/src/pkg/custom_type"
)

// EnterpriseRole 企业角色
type EnterpriseRole struct {
	ID           string               `gorm:"type:VARCHAR;not null;primaryKey;unique" json:"id"`
	EnterpriseID string               `gorm:"type:VARCHAR;not null;index:idx_er_enterprise_id" json:"enterprise_id"`
	Name         string               `gorm:"type:VARCHAR;not null" json:"name"`
	IsDefault    int                  `gorm:"type:INTEGER;not null;default:0" json:"is_default"`
	IsAdmin      int                  `gorm:"type:INTEGER;not null;default:0" json:"is_admin"`
	CreatedAt    custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
}

func (EnterpriseRole) TableName() string {
	return "enterprise_role"
}

// EnterpriseRolePower 企业角色权限关联
type EnterpriseRolePower struct {
	RoleID  string `gorm:"type:VARCHAR;not null;primaryKey" json:"role_id"`
	PowerID int    `gorm:"type:INTEGER;not null;primaryKey" json:"power_id"`
}

func (EnterpriseRolePower) TableName() string {
	return "enterprise_role_power"
}

package models

// GroupPower 组权限
type GroupPower struct {
	GroupID int `gorm:"type:INTEGER;not null;primaryKey" json:"group_id"` // 组ID
	PowerID int `gorm:"type:INTEGER;not null" json:"power_id"`            // 权限ID
}

func (GroupPower) TableName() string {
	return "group_power"
}

package models

import (
	"myobj/src/pkg/custom_type"
)

// UserInfo 用户信息
type UserInfo struct {
	//用户id
	ID string `gorm:"type:VARCHAR;not null;primaryKey;unique" json:"id"`
	//用户昵称
	Name string `gorm:"type:VARCHAR;not null" json:"name"`
	//用户名
	UserName string `gorm:"type:VARCHAR;not null" json:"user_name"`
	//用户密码
	Password string `gorm:"type:TEXT;not null" json:"password"`
	//用户邮箱
	Email string `gorm:"type:TEXT;not null" json:"email"`
	//用户手机号
	Phone string `gorm:"type:VARCHAR;not null" json:"phone"`
	//用户组id
	GroupID int `gorm:"type:INTEGER;not null" json:"group_id"`
	//用户创建时间
	CreatedAt custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	//用户可用存储空间
	Space int64 `gorm:"type:integer" json:"space"`
	//用户文件密码
	FilePassword string `gorm:"type:text" json:"file_password"`
	//用户剩余存储空间
	FreeSpace int64 `gorm:"type:free_space" json:"free_space"`
	//用户状态 0正常 1禁用
	State int `gorm:"type:INTEGER;not null" json:"state"`
}

func (UserInfo) TableName() string {
	return "user_info"
}

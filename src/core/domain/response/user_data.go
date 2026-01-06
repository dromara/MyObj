package response

import (
	"myobj/src/pkg/models"
)

// UserLoginResponse 用户登录响应结构体
type UserLoginResponse struct {
	Token string           `json:"token"`
	User  *models.UserInfo `json:"user_info"`
	Power []*models.Power  `json:"power"`
}

type UserInfoResponse struct {
	//用户id
	ID string `json:"id"`
	//用户昵称
	Name string `json:"name"`
	//用户名
	UserName string `json:"user_name"`
	//用户邮箱
	Email string `json:"email"`
	//用户手机号
	Phone string `json:"phone"`
	//用户可用存储空间
	Space int64 `json:"space"`
	//用户剩余存储空间
	FreeSpace int64 `json:"free_space"`
	//用户状态 0正常 1禁用
	State int `json:"state"`
}

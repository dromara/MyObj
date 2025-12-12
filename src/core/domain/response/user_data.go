package response

import "myobj/src/pkg/models"

// UserLoginResponse 用户登录响应结构体
type UserLoginResponse struct {
	Token string           `json:"token"`
	User  *models.UserInfo `json:"user_info"`
	Power []*models.Power  `json:"power"`
}

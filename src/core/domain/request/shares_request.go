package request

import "myobj/src/pkg/custom_type"

type CreateShareRequest struct {
	// 文件ID
	FileID string `json:"file_id"`
	// 过期时间
	Expire custom_type.JsonTime `json:"expire"`
	// 密码
	Password string `json:"password"`
}

type DeleteShareRequest struct {
	// 分享ID
	ID int `json:"id"`
}

type UpdateSharePasswordRequest struct {
	// 分享ID
	ID int `json:"id"`
	// 新密码
	Password string `json:"password"`
}

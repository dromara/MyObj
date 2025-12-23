package request

// UserLoginRequest 用户登录请求结构体
type UserLoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Challenge string `json:"challenge"`
}

// UserRegisterRequest 用户注册请求结构体
type UserRegisterRequest struct {
	Nickname  string `json:"nickname"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Challenge string `json:"challenge"` //挑战ID
}

// UserUpdateRequest 用户更新请求结构体
type UserUpdateRequest struct {
	ID string `json:"id"`
	// 用户昵称
	Nickname string `json:"nickname"`
	// 用户手机号
	Phone string `json:"phone"`
	// 用户邮箱
	Email string `json:"email"`
	// 用户名
	Username string `json:"username"`
}

// UserUpdatePasswordRequest 用户更新密码请求结构体
type UserUpdatePasswordRequest struct {
	ID        string `json:"id"`
	OldPasswd string `json:"old_passwd"`
	NewPasswd string `json:"new_passwd"`
	Challenge string `json:"challenge"` //挑战ID
}

// UserSetFilePasswordRequest 用户设置文件密码请求结构体
type UserSetFilePasswordRequest struct {
	ID        string `json:"id"`
	Passwd    string `json:"passwd"`
	Challenge string `json:"challenge"` //挑战ID
}

// GenerateApiKeyRequest 生成API Key请求结构体
type GenerateApiKeyRequest struct {
	ExpiresDays int `json:"expires_days"` // 过期天数，0表示永不过期
}

// DeleteApiKeyRequest 删除API Key请求结构体
type DeleteApiKeyRequest struct {
	ApiKeyID int `json:"api_key_id" binding:"required"` // API Key ID
}
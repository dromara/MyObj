package response

import "myobj/src/pkg/models"

// AdminUserListResponse 管理员用户列表响应
type AdminUserListResponse struct {
	Users    []*AdminUserInfo `json:"users"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// AdminUserInfo 管理员用户信息（包含组名）
type AdminUserInfo struct {
	models.UserInfo
	GroupName string `json:"group_name,omitempty"`
}

// AdminGroupListResponse 管理员组列表响应
type AdminGroupListResponse struct {
	Groups []*models.Group `json:"groups"`
	Total  int64           `json:"total"`
}

// AdminPowerListResponse 管理员权限列表响应
type AdminPowerListResponse struct {
	Powers []*models.Power `json:"powers"`
	Total  int64           `json:"total"`
}

// AdminGroupPowersResponse 组的权限列表响应
type AdminGroupPowersResponse struct {
	PowerIDs []int `json:"power_ids"`
}

// AdminDiskListResponse 管理员磁盘列表响应
type AdminDiskListResponse struct {
	Disks []*models.Disk `json:"disks"`
	Total int64          `json:"total"`
}

// AdminSystemConfigResponse 系统配置响应
type AdminSystemConfigResponse struct {
	AllowRegister bool   `json:"allow_register"`
	WebdavEnabled bool   `json:"webdav_enabled"`
	Version       string `json:"version"`
	TotalUsers    int64  `json:"total_users"`
	TotalFiles    int64  `json:"total_files"`
	Uptime        string `json:"uptime,omitempty"`
}

// PackageCreateResponse 创建打包下载响应
type PackageCreateResponse struct {
	PackageID   string `json:"package_id"`
	PackageName string `json:"package_name"`
	Status      string `json:"status"` // creating, ready, failed
	Progress    int    `json:"progress"` // 0-100
	TotalSize   int64  `json:"total_size"`
}

// PackageProgressResponse 打包进度响应
type PackageProgressResponse struct {
	PackageID   string `json:"package_id"`
	Status      string `json:"status"` // creating, ready, failed
	Progress    int    `json:"progress"` // 0-100
	TotalSize   int64  `json:"total_size"`
	CreatedSize int64  `json:"created_size"`
	ErrorMsg    string `json:"error_msg,omitempty"`
}


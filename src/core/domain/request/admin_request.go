package request

// AdminUserListRequest 管理员用户列表请求
type AdminUserListRequest struct {
	Page     int    `json:"page" form:"page" binding:"required,min=1"`
	PageSize int    `json:"pageSize" form:"pageSize" binding:"required,min=1,max=100"`
	Keyword  string `json:"keyword" form:"keyword"`
	GroupID  int    `json:"group_id" form:"group_id"`
	State    *int   `json:"state" form:"state"` // nil或-1-全部 0-正常 1-禁用（使用指针类型以区分未传递和传递了0）
}

// AdminCreateUserRequest 管理员创建用户请求
type AdminCreateUserRequest struct {
	UserName string `json:"user_name" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	GroupID  int    `json:"group_id" binding:"required"`
	Space    int64  `json:"space"` // 存储空间（字节），0表示无限
}

// AdminUpdateUserRequest 管理员更新用户请求
type AdminUpdateUserRequest struct {
	ID      string `json:"id" binding:"required"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	GroupID int    `json:"group_id"`
	Space   int64  `json:"space"`
	State   int    `json:"state"` // 0-正常 1-禁用
}

// AdminDeleteUserRequest 管理员删除用户请求
type AdminDeleteUserRequest struct {
	ID string `json:"id" binding:"required"`
}

// AdminToggleUserStateRequest 管理员启用/禁用用户请求
type AdminToggleUserStateRequest struct {
	ID    string `json:"id" binding:"required"`
	State int    `json:"state" binding:"oneof=0 1"` // 0-正常 1-禁用（oneof 已包含必填验证）
}

// AdminGroupListRequest 管理员组列表请求（暂无参数）
type AdminGroupListRequest struct{}

// AdminCreateGroupRequest 管理员创建组请求
type AdminCreateGroupRequest struct {
	Name         string `json:"name" binding:"required"`
	Space        int64  `json:"space"`         // 存储空间（字节），0表示无限
	GroupDefault int    `json:"group_default"` // 0-否 1-是
}

// AdminUpdateGroupRequest 管理员更新组请求
type AdminUpdateGroupRequest struct {
	ID           int    `json:"id" binding:"required"`
	Name         string `json:"name"`
	Space        int64  `json:"space"`
	GroupDefault int    `json:"group_default"` // 0-否 1-是
}

// AdminDeleteGroupRequest 管理员删除组请求
type AdminDeleteGroupRequest struct {
	ID int `json:"id" binding:"required"`
}

// AdminPowerListRequest 管理员权限列表请求
type AdminPowerListRequest struct {
	Page     int `json:"page" form:"page" binding:"omitempty,min=1"`
	PageSize int `json:"pageSize" form:"pageSize" binding:"omitempty,min=1,max=1000"` // 允许最大1000，用于组管理分配权限时获取所有权限
}

// AdminCreatePowerRequest 管理员创建权限请求
type AdminCreatePowerRequest struct {
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description" binding:"required"`
	Characteristic string `json:"characteristic" binding:"required"`
}

// AdminUpdatePowerRequest 管理员更新权限请求
type AdminUpdatePowerRequest struct {
	ID             int    `json:"id" binding:"required"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Characteristic string `json:"characteristic"`
}

// AdminDeletePowerRequest 管理员删除权限请求
type AdminDeletePowerRequest struct {
	ID int `json:"id" binding:"required"`
}

// AdminBatchDeletePowerRequest 管理员批量删除权限请求
type AdminBatchDeletePowerRequest struct {
	IDs []int `json:"ids" binding:"required,min=1"`
}

// AdminAssignPowerRequest 管理员分配权限请求
type AdminAssignPowerRequest struct {
	GroupID  int   `json:"group_id" binding:"required"`
	PowerIDs []int `json:"power_ids" binding:"required"`
}

// AdminGetGroupPowersRequest 获取组的权限列表请求
type AdminGetGroupPowersRequest struct {
	GroupID int `json:"group_id" form:"group_id" binding:"required"`
}

// AdminDiskListRequest 管理员磁盘列表请求（暂无参数）
type AdminDiskListRequest struct{}

// AdminCreateDiskRequest 管理员创建磁盘请求
type AdminCreateDiskRequest struct {
	DiskPath string `json:"disk_path" binding:"required"`
	DataPath string `json:"data_path" binding:"required"`
	Size     int64  `json:"size" binding:"required,min=0"` // 大小（字节）
}

// AdminUpdateDiskRequest 管理员更新磁盘请求
type AdminUpdateDiskRequest struct {
	ID       string `json:"id" binding:"required"`
	DiskPath string `json:"disk_path"`
	DataPath string `json:"data_path"`
	Size     int64  `json:"size"` // 大小（字节）
}

// AdminDeleteDiskRequest 管理员删除磁盘请求
type AdminDeleteDiskRequest struct {
	ID string `json:"id" binding:"required"`
}

// AdminGetSystemConfigRequest 获取系统配置请求（暂无参数）
type AdminGetSystemConfigRequest struct{}

// AdminUpdateSystemConfigRequest 更新系统配置请求
type AdminUpdateSystemConfigRequest struct {
	AllowRegister bool `json:"allow_register"`
	WebdavEnabled bool `json:"webdav_enabled"`
}

// PackageCreateRequest 创建打包下载请求
type PackageCreateRequest struct {
	FileIDs     []string `json:"file_ids" binding:"required,min=1"`
	PackageName string   `json:"package_name"`
}

// PackageProgressRequest 获取打包进度请求
type PackageProgressRequest struct {
	PackageID string `json:"package_id" form:"package_id" binding:"required"`
}

// PackageDownloadRequest 下载打包文件请求
type PackageDownloadRequest struct {
	PackageID string `json:"package_id" form:"package_id" binding:"required"`
}

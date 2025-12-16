package request

// CreateOfflineDownloadRequest 创建离线下载任务请求
type CreateOfflineDownloadRequest struct {
	// 下载URL
	URL string `json:"url" binding:"required"`
	// 保存的虚拟路径（可选，默认为/离线下载/）
	VirtualPath string `json:"virtual_path"`
	// 是否加密存储
	EnableEncryption bool `json:"enable_encryption"`
}

// DownloadTaskListRequest 下载任务列表请求
type DownloadTaskListRequest struct {
	// 任务状态（可选，0=初始化,1=下载中,2=暂停,3=完成,4=失败）
	State int `form:"state"`
	// 页码
	Page int `form:"page" binding:"required,min=1"`
	// 每页数量
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

// TaskOperationRequest 任务操作请求（暂停、恢复、取消）
type TaskOperationRequest struct {
	// 任务ID
	TaskID string `json:"task_id" binding:"required"`
}

// DeleteTaskRequest 删除任务请求
type DeleteTaskRequest struct {
	// 任务ID
	TaskID string `json:"task_id" binding:"required"`
}

// CreateLocalFileDownloadRequest 创建网盘文件下载任务请求
type CreateLocalFileDownloadRequest struct {
	// 文件ID
	FileID string `json:"file_id" binding:"required"`
	// 文件解密密码（加密文件必需）
	FilePassword string `json:"file_password"`
}

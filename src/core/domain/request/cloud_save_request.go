package request

// SaveShareFilesRequest 保存分享文件请求
type SaveShareFilesRequest struct {
	// 云盘提供者: aliyun, baidu, xunlei, 115, quark, caiyun, tianyi, uc, wopan
	Provider string `json:"provider" binding:"required"`
	// 分享ID
	ShareID string `json:"share_id" binding:"required"`
	// 保存类型: single(单个文件), multiple(多个文件), all(全部文件), directory(目录)
	SaveType string `json:"save_type" binding:"required"`
	// 文件ID列表（single和multiple类型必需）
	FileIDs []string `json:"file_ids"`
	// 目录名称（directory类型必需）
	DirName string `json:"dir_name"`
	// 目标路径（可选，默认为用户目录）
	TargetPath string `json:"target_path"`
}

// GetShareFileTreeRequest 获取分享文件树请求
type GetShareFileTreeRequest struct {
	// 云盘提供者
	Provider string `json:"provider" binding:"required"`
	// 分享ID
	ShareID string `json:"share_id" binding:"required"`
	// 父目录ID（为空则获取根目录）
	ParentFileID string `json:"parent_file_id"`
	// 是否递归获取子目录
	Recursive bool `json:"recursive"`
	// 最大递归深度
	MaxDepth int `json:"max_depth"`
}

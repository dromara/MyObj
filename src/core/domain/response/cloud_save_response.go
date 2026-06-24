package response

// SaveShareFilesResponse 保存分享文件响应
type SaveShareFilesResponse struct {
	SuccessCount int              `json:"success_count"`
	FailedCount  int              `json:"failed_count"`
	SavedFiles   []SavedFileInfo  `json:"saved_files"`
	FailedFiles  []FailedFileInfo `json:"failed_files,omitempty"`
}

// SavedFileInfo 已保存文件信息
type SavedFileInfo struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	Size     int64  `json:"size"`
}

// FailedFileInfo 失败文件信息
type FailedFileInfo struct {
	FileID string `json:"file_id"`
	Error  string `json:"error"`
}

// ShareFileTreeResponse 分享文件树响应
type ShareFileTreeResponse struct {
	ShareID string           `json:"share_id"`
	Files   []ShareFileNode  `json:"files"`
}

// ShareFileNode 分享文件节点
type ShareFileNode struct {
	FileID   string          `json:"file_id"`
	Name     string          `json:"name"`
	Size     int64           `json:"size"`
	IsDir    bool            `json:"is_dir"`
	FileType string          `json:"file_type"`
	FileExt  string          `json:"file_ext"`
	Children []ShareFileNode `json:"children,omitempty"`
}

// ShareTransferStatusResponse 转存状态响应
type ShareTransferStatusResponse struct {
	TaskID       string `json:"task_id"`
	Status       string `json:"status"` // pending, downloading, completed, failed
	TotalFiles   int    `json:"total_files"`
	SuccessFiles int    `json:"success_files"`
	FailedFiles  int    `json:"failed_files"`
	Progress     int    `json:"progress"` // 0-100
	Message      string `json:"message"`
}

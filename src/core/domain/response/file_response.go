package response

import "myobj/src/pkg/custom_type"

// FileListResponse 文件列表响应结构体
type FileListResponse struct {
	// 面包屑路径
	Breadcrumbs []Breadcrumb `json:"breadcrumbs"`
	// 当前路径
	CurrentPath string `json:"current_path"`
	// 目录列表
	Folders []*FolderItem `json:"folders"`
	// 文件列表
	Files []*FileItem `json:"files"`
	// 总数（目录+文件）
	Total int64 `json:"total"`
	// 当前页
	Page int `json:"page"`
	// 每页数量
	PageSize int `json:"page_size"`
}

// Breadcrumb 面包屑项
type Breadcrumb struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// FolderItem 目录项
type FolderItem struct {
	ID          int                  `json:"id"`
	Name        string               `json:"name"`
	Path        string               `json:"path"`
	CreatedTime custom_type.JsonTime `json:"created_time"`
}

// FileItem 文件项
type FileItem struct {
	FileID       string               `json:"file_id"`
	UfID         string               `json:"uf_id"` // 用户文件ID
	FileName     string               `json:"file_name"`
	FileSize     int                  `json:"file_size"`
	MimeType     string               `json:"mime_type"`
	IsEnc        bool                 `json:"is_enc"`
	HasThumbnail bool                 `json:"has_thumbnail"` // 是否有缩略图
	CreatedAt    custom_type.JsonTime `json:"created_at"`
}

// FileDir 文件目录结构体
type FileDir struct {
	//路径ID
	ID int `json:"id"`
	// 路径
	Path string `json:"path"`
	// 子路径
	Subpath []struct {
		ID   int    `json:"id"`
		Path string `json:"path"`
	} `json:"subpath"`
	// 父级路径id
	ParentID string `json:"parent_id"`
	// 文件夹创建时间
	CreatedTime custom_type.JsonTime `json:"created_time"`
}

// FileInfoData 文件信息结构体
type FileInfoData struct {
	// 文件ID
	FileID string `json:"file_id"`
	// 文件名
	FileName string `json:"file_name"`
	// 文件大小
	FileSize int64 `json:"file_size"`
	// 文件hash
	FileHash string `json:"file_hash"`
	// 是否加密
	IsEnc bool `json:"is_enc"`
	// 文件类型
	MimeType string `json:"mime_type"`
	// 文件虚拟路径
	VirtualPath string `json:"virtual_path"`
	// 文件上传时间
	CreatedAt string `json:"created_at"`
	// 文件缩略图 base64
	Thumbnail string `json:"thumbnail"`
}

// ShareListItem 分享列表项
type ShareListItem struct {
	ID            int                  `json:"id"`
	UserID        string               `json:"user_id"`
	FileID        string               `json:"file_id"`
	FileName      string               `json:"file_name"` // 用户文件名
	Token         string               `json:"token"`
	ExpiresAt     custom_type.JsonTime `json:"expires_at"`
	PasswordHash  string               `json:"password_hash"`
	DownloadCount int                  `json:"download_count"`
	CreatedAt     custom_type.JsonTime `json:"created_at"`
}

// FilePrecheckResponse 文件预检查响应结构体
type FilePrecheckResponse struct {
	PrecheckID string   `json:"precheck_id"`
	Md5        []string `json:"md5"`
}

// VideoPlayTokenResponse 视频播放 Token 响应
type VideoPlayTokenResponse struct {
	// 播放 Token（24小时有效）
	PlayToken string `json:"play_token"`
	// 文件信息
	FileInfo VideoFileInfo `json:"file_info"`
}

// VideoFileInfo 视频文件信息
type VideoFileInfo struct {
	// 文件ID
	FileID string `json:"file_id"`
	// 文件名
	FileName string `json:"file_name"`
	// 文件大小（字节）
	FileSize int64 `json:"file_size"`
	// 是否加密
	IsEnc bool `json:"is_enc"`
	// MIME 类型
	MimeType string `json:"mime_type"`
}

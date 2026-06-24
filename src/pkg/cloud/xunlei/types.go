package xunlei

// ShareLinkInfoRequest 获取分享链接信息请求
type ShareLinkInfoRequest struct {
	ShareID string `json:"share_id"`
	Pwd     string `json:"pwd,omitempty"`
}

// ShareLinkInfoResponse 获取分享链接信息响应
type ShareLinkInfoResponse struct {
	ShareID    string `json:"share_id"`
	Title      string `json:"title"`
	ShareTitle string `json:"share_title"`
	FileCount  int    `json:"file_count"`
	CreateTime int64  `json:"create_time"`
	ExpireTime int64  `json:"expire_time"`
	Creator    struct {
		Nickname string `json:"nickname"`
		UserID   string `json:"user_id"`
	} `json:"creator"`
}

// ShareFileListRequest 获取分享文件列表请求
type ShareFileListRequest struct {
	ShareID      string `json:"share_id"`
	ParentFileID string `json:"parent_file_id,omitempty"`
	Pwd          string `json:"pwd,omitempty"`
	PageToken    string `json:"page_token,omitempty"`
	PageSize     int    `json:"page_size,omitempty"`
}

// ShareFileListResponse 获取分享文件列表响应
type ShareFileListResponse struct {
	Files       []ShareFileItem `json:"files"`
	NextPageToken string        `json:"next_page_token"`
	TotalCount  int             `json:"total_count"`
}

// ShareFileItem 分享文件项
type ShareFileItem struct {
	FileID       string `json:"file_id"`
	ParentFileID string `json:"parent_file_id"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	IsDir        bool   `json:"is_dir"`
	FileType     string `json:"file_type"`
	FileExt      string `json:"file_ext"`
	Thumbnail    string `json:"thumbnail"`
	CreateTime   int64  `json:"create_time"`
	UpdateTime   int64  `json:"update_time"`
	Hash         string `json:"hash"`
}

// ShareDownloadRequest 获取下载链接请求
type ShareDownloadRequest struct {
	ShareID  string `json:"share_id"`
	FileID   string `json:"file_id"`
	Pwd      string `json:"pwd,omitempty"`
}

// ShareDownloadResponse 获取下载链接响应
type ShareDownloadResponse struct {
	URL        string `json:"url"`
	Size       int64  `json:"size"`
	ExpireTime int64  `json:"expire_time"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

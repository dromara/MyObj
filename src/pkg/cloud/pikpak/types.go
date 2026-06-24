package pikpak

// ShareTokenRequest 获取分享token请求
type ShareTokenRequest struct {
	ShareID string `json:"share_id"`
}

// ShareTokenResponse 获取分享token响应
type ShareTokenResponse struct {
	ShareToken string `json:"share_token"`
}

// ShareDetailRequest 获取分享详情请求
type ShareDetailRequest struct {
	ShareID    string `json:"share_id"`
	ShareToken string `json:"share_token"`
	ParentID   string `json:"parent_id,omitempty"`
	PageToken  string `json:"page_token,omitempty"`
	PageSize   int    `json:"page_size,omitempty"`
	ThumbnailSize string `json:"thumbnail_size,omitempty"`
}

// ShareDetailResponse 获取分享详情响应
type ShareDetailResponse struct {
	ShareID    string          `json:"share_id"`
	Title      string          `json:"title"`
	ShareToken string          `json:"share_token"`
	Files      []ShareFileItem `json:"files"`
	NextPageToken string       `json:"next_page_token"`
}

// ShareFileItem 文件项
type ShareFileItem struct {
	FileID       string `json:"file_id"`
	ParentID     string `json:"parent_id"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	Kind         string `json:"kind"`
	MimeType     string `json:"mime_type"`
	ThumbnailLink string `json:"thumbnail_link"`
	WebContentLink string `json:"web_content_link"`
	CreatedTime  string `json:"created_time"`
	ModifiedTime string `json:"modified_time"`
	Hash         string `json:"hash"`
	FolderType   string `json:"folder_type"`
}

// DownloadURLRequest 获取下载链接请求
type DownloadURLRequest struct {
	ShareToken string `json:"share_token"`
	FileID     string `json:"file_id"`
}

// DownloadURLResponse 下载链接响应
type DownloadURLResponse struct {
	WebContentLink string `json:"web_content_link"`
	Link           string `json:"link"`
	Size           int64  `json:"size"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

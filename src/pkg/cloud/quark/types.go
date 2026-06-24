package quark

// ShareTokenRequest 获取分享token请求
type ShareTokenRequest struct {
	ShareID   string `json:"share_id"`
	SharePwd  string `json:"share_pwd,omitempty"`
	Passcode  string `json:"passcode,omitempty"`
}

// ShareTokenResponse 获取分享token响应
type ShareTokenResponse struct {
	ShareToken  string `json:"share_token"`
	HasPassword bool   `json:"has_password"`
	Stoken      string `json:"stoken"`
}

// ShareDetailRequest 获取分享详情请求
type ShareDetailRequest struct {
	ShareID    string `json:"share_id"`
	ShareToken string `json:"share_token"`
	Password   string `json:"password,omitempty"`
	Force      int    `json:"force,omitempty"`
	Page       int    `json:"page,omitempty"`
	Size       int    `json:"size,omitempty"`
	Pr         string `json:"pr,omitempty"`
	Category   int    `json:"category,omitempty"`
}

// ShareDetailResponse 获取分享详情响应
type ShareDetailResponse struct {
	ShareID    string          `json:"share_id"`
	Title      string          `json:"title"`
	Subtitle   string          `json:"subtitle"`
	ShareName  string          `json:"share_name"`
	CreateTime int64           `json:"create_time"`
	ExpireTime int64           `json:"expire_time"`
	Expired    bool            `json:"expired"`
	FileCount  int             `json:"file_count"`
	SaveCount  int             `json:"save_count"`
	Count      int             `json:"count"`
	List       []ShareFileItem `json:"list"`
}

// ShareFileItem 分享文件项
type ShareFileItem struct {
	FileID       string `json:"file_id"`
	ParentFileID string `json:"parent_file_id"`
	Name         string `json:"file_name"`
	Size         int64  `json:"size"`
	FormatType   string `json:"format_type"`
	Category     int    `json:"category"`
	FileType     int    `json:"file_type"`
	UpdatedAt    int64  `json:"updated_at"`
	CreatedAt    int64  `json:"created_at"`
	Thumbnail    string `json:"thumbnail"`
	Duration     int    `json:"duration"`
	// 0=file, 1=folder
	Dir bool `json:"dir"`
}

// ShareFileTokenRequest 获取文件下载token请求
type ShareFileTokenRequest struct {
	ShareID    string   `json:"share_id"`
	ShareToken string   `json:"share_token"`
	FileIDs    []string `json:"file_id_list"`
}

// ShareFileTokenResponse 获取文件下载token响应
type ShareFileTokenResponse struct {
	DownloadToken string            `json:"download_token"`
	FileTokens    []FileTokenItem   `json:"file_tokens"`
}

// FileTokenItem 文件token项
type FileTokenItem struct {
	FileID      string `json:"file_id"`
	DownloadURL string `json:"download_url"`
}

// DownloadRequest 获取下载链接请求
type DownloadRequest struct {
	ShareID       string   `json:"share_id"`
	ShareToken    string   `json:"share_token"`
	DownloadToken string   `json:"download_token"`
	FileID        string   `json:"file_id"`
	FormatType    string   `json:"format_type,omitempty"`
}

// DownloadResponse 获取下载链接响应
type DownloadResponse struct {
	DownloadURL string `json:"download_url"`
	FileID      string `json:"file_id"`
	Size        int64  `json:"size"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Status    int    `json:"status"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      any    `json:"data,omitempty"`
}

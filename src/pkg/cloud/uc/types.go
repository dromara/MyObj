package uc

// ShareTokenRequest 获取分享token请求
type ShareTokenRequest struct {
	ShareID string `json:"share_id"`
	Pwd     string `json:"pwd,omitempty"`
}

// ShareTokenResponse 获取分享token响应
type ShareTokenResponse struct {
	St string `json:"st"`
}

// ShareDetailRequest 获取分享详情请求
type ShareDetailRequest struct {
	ShareID    string `json:"share_id"`
	Pwd        string `json:"pwd,omitempty"`
	ShareToken string `json:"shareToken"`
	Page       int    `json:"page"`
	Pr         int    `json:"pr"`
	Scene       string `json:"s_type"`
}

// ShareDetailResponse 获取分享详情响应
type ShareDetailResponse struct {
	Status int    `json:"status"`
	ErrNo  int    `json:"errno"`
	Msg    string `json:"msg"`
	Data   struct {
		ShareID     string `json:"share_id"`
		Title       string `json:"title"`
		FileName    string `json:"file_name"`
		ShareName   string `json:"share_name"`
		FileList    []FileItem `json:"list"`
		Pagetoken   string `json:"page_token"`
		HasMore     bool   `json:"has_more"`
		TotalCount  int    `json:"total_count"`
		Expired     bool   `json:"expired"`
		ExpireTime  string `json:"expire_time"`
		Creator     string `json:"creator"`
	} `json:"data"`
}

// FileItem 文件项
type FileItem struct {
	FileID      string `json:"file_id"`
	ParentID    string `json:"parent_id"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	IsDir       bool   `json:"is_dir"`
	FileType    int    `json:"file_type"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
	Thumbnail   string `json:"thumbnail"`
	FileExtension string `json:"file_extension"`
}

// DownloadURLRequest 获取下载链接请求
type DownloadURLRequest struct {
	ShareToken string   `json:"shareToken"`
	FileIDs    []string `json:"file_ids"`
	ShareID    string   `json:"share_id"`
}

// DownloadURLResponse 获取下载链接响应
type DownloadURLResponse struct {
	Status int    `json:"status"`
	ErrNo  int    `json:"errno"`
	Msg    string `json:"msg"`
	Data   []DownloadURLItem `json:"data"`
}

// DownloadURLItem 下载链接项
type DownloadURLItem struct {
	FileID  string `json:"file_id"`
	URL     string `json:"url"`
	ErrNo   int    `json:"errno"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Status int    `json:"status"`
	ErrNo  int    `json:"errno"`
	Msg    string `json:"msg"`
}

package aliyun

// ShareTokenRequest 获取分享token请求
type ShareTokenRequest struct {
	ShareID string `json:"share_id"`
	SharePwd string `json:"share_pwd,omitempty"`
}

// ShareTokenResponse 获取分享token响应
type ShareTokenResponse struct {
	ExpireTime  string `json:"expire_time"`
	ShareToken  string `json:"share_token"`
	HasPassword bool   `json:"has_password"`
}

// ShareLinkInfoRequest 获取分享链接信息请求
type ShareLinkInfoRequest struct {
	ShareID string `json:"share_id"`
}

// ShareLinkInfoResponse 获取分享链接信息响应
type ShareLinkInfoResponse struct {
	Creator struct {
		Avatar    string `json:"avatar"`
		CreatedAt string `json:"created_at"`
		Email     string `json:"email"`
		NickName  string `json:"nick_name"`
		UpdatedAt string `json:"updated_at"`
		UserID    string `json:"user_id"`
	} `json:"creator"`
	ShareID     string `json:"share_id"`
	ShareName   string `json:"share_name"`
	ShareTitle  string `json:"share_title"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
	Expiration  string `json:"expiration"`
	Expired     bool   `json:"expired"`
	LikeCount   int64  `json:"like_count"`
	PreviewCount int64 `json:"preview_count"`
	SaveCount   int64  `json:"save_count"`
	Count       struct {
		Files   int `json:"files"`
		Folders int `json:"folders"`
	} `json:"count"`
	Permission struct {
		CanCopy   bool `json:"can_copy"`
		CanSave   bool `json:"can_save"`
		CanShare  bool `json:"can_share"`
		CanPreview bool `json:"can_preview"`
	} `json:"permission"`
	SharePwd string `json:"share_pwd"`
}

// ShareFileListRequest 获取分享文件列表请求
type ShareFileListRequest struct {
	ShareID      string `json:"share_id"`
	ShareToken   string `json:"share_token"`
	ParentFileID string `json:"parent_file_id"`
	Marker       string `json:"marker,omitempty"`
	Limit        int    `json:"limit,omitempty"`
	OrderBy      string `json:"order_by,omitempty"`
	OrderDirection string `json:"order_direction,omitempty"`
}

// ShareFileListResponse 获取分享文件列表响应
type ShareFileListResponse struct {
	Items      []ShareFileItem `json:"items"`
	NextMarker string          `json:"next_marker"`
}

// ShareFileItem 分享文件项
type ShareFileItem struct {
	DriveID      string `json:"drive_id"`
	FileID       string `json:"file_id"`
	ParentFileID string `json:"parent_file_id"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	FileIDExtension string `json:"file_id_extension"`
	Type         string `json:"type"` // file 或 folder
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	Thumbnail    string `json:"thumbnail"`
	Category     string `json:"category"` // image, video, doc, audio, other
	FileExtension string `json:"file_extension"`
	URL          string `json:"url"`
	HashInfo     struct {
		Sha1 string `json:"sha1"`
	} `json:"hash_info"`
}

// GetShareLinkDownloadURLRequest 获取分享下载链接请求
type GetShareLinkDownloadURLRequest struct {
	ShareID    string `json:"share_id"`
	FileID     string `json:"file_id"`
	ShareToken string `json:"share_token"`
}

// GetShareLinkDownloadURLResponse 获取分享下载链接响应
type GetShareLinkDownloadURLResponse struct {
	URL             string `json:"url"`
	Expiration      string `json:"expiration"`
	InternalURL     string `json:"internal_url,omitempty"`
	Size            int64  `json:"size"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	RequestID string `json:"request_id"`
}

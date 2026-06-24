package wopan

// GetShareInfoRequest 获取分享信息请求
type GetShareInfoRequest struct {
	ShareID string `json:"shareId"`
}

// GetShareInfoResponse 获取分享信息响应
type GetShareInfoResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ShareID     string `json:"shareId"`
		ShareName   string `json:"shareName"`
		ShareTitle  string `json:"shareTitle"`
		HasPassword bool   `json:"hasPassword"`
		Expired     bool   `json:"expired"`
		Expiration  string `json:"expiration"`
		FileCount   int    `json:"fileCount"`
		FolderCount int    `json:"folderCount"`
		TotalSize   int64  `json:"totalSize"`
		Creator     struct {
			NickName string `json:"nickName"`
			UserID   string `json:"userId"`
		} `json:"creator"`
	} `json:"data"`
}

// VerifySharePwdRequest 验证分享密码请求
type VerifySharePwdRequest struct {
	ShareID string `json:"shareId"`
	Pwd     string `json:"pwd"`
}

// VerifySharePwdResponse 验证分享密码响应
type VerifySharePwdResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ShareToken string `json:"shareToken"`
		Valid      bool   `json:"valid"`
	} `json:"data"`
}

// GetShareFileListRequest 获取分享文件列表请求
type GetShareFileListRequest struct {
	ShareID      string `json:"shareId"`
	ShareToken   string `json:"shareToken"`
	ParentFileID string `json:"parentFileId"`
	PageNum      int    `json:"pageNum"`
	PageSize     int    `json:"pageSize"`
}

// GetShareFileListResponse 获取分享文件列表响应
type GetShareFileListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items    []ShareFileItem `json:"items"`
		Total    int             `json:"total"`
		PageNum  int             `json:"pageNum"`
		PageSize int             `json:"pageSize"`
	} `json:"data"`
}

// ShareFileItem 分享文件项
type ShareFileItem struct {
	FileID       string `json:"fileId"`
	ParentFileID string `json:"parentFileId"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	Type         int    `json:"type"` // 1=文件, 2=目录
	FileType     string `json:"fileType"`
	FileExt      string `json:"fileExt"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
	Thumbnail    string `json:"thumbnail"`
}

// GetDownloadURLRequest 获取下载链接请求
type GetDownloadURLRequest struct {
	ShareID    string `json:"shareId"`
	ShareToken string `json:"shareToken"`
	FileID     string `json:"fileId"`
}

// GetDownloadURLResponse 获取下载链接响应
type GetDownloadURLResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		URL        string `json:"url"`
		Size       int64  `json:"size"`
		Expiration int64  `json:"expiration"` // 秒
	} `json:"data"`
}

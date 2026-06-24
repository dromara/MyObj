package caiyun

type GetShareInfoRequest struct {
	ShareID string `json:"shareId"`
}

type GetShareInfoResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ShareID     string `json:"shareId"`
		ShareTitle  string `json:"shareTitle"`
		ShareName   string `json:"shareName"`
		ShareType   int    `json:"shareType"`
		SharePwd    string `json:"sharePwd"`
		ExpireTime  string `json:"expireTime"`
		FileCount   int    `json:"fileCount"`
		FolderCount int    `json:"folderCount"`
		CreatorName string `json:"creatorName"`
		ContentOK   int    `json:"contentOK"`
		AccessCode  string `json:"accessCode"`
	} `json:"data"`
}

type GetShareFileListRequest struct {
	ShareID      string `json:"shareId"`
	AccessCode   string `json:"accessCode,omitempty"`
	CatalogID    string `json:"catalogID,omitempty"`
	PageNum      int    `json:"pageNum,omitempty"`
	PageSize     int    `json:"pageSize,omitempty"`
}

type GetShareFileListResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TotalCount int                `json:"totalCount"`
		FileList   []ShareFileItem    `json:"fileList"`
		CatalogList []ShareCatalogItem `json:"catalogList"`
	} `json:"data"`
}

type ShareFileItem struct {
	FileID      string `json:"fileId"`
	FileName    string `json:"fileName"`
	FileSize    int64  `json:"fileSize"`
	FileType    string `json:"fileType"`
	UpdateTime  string `json:"updateTime"`
	CreateTime  string `json:"createTime"`
	ThumbnailURL string `json:"thumbnailURL"`
	PanFileID   string `json:"panFileID"`
}

type ShareCatalogItem struct {
	CatalogID   string `json:"catalogID"`
	CatalogName string `json:"catalogName"`
	UpdateTime  string `json:"updateTime"`
	CreateTime  string `json:"createTime"`
}

type GetShareDownloadURLRequest struct {
	ShareID    string `json:"shareId"`
	AccessCode string `json:"accessCode,omitempty"`
	FileID     string `json:"fileID"`
}

type GetShareDownloadURLResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		DownloadURL string `json:"downloadURL"`
		FileSize    int64  `json:"fileSize"`
	} `json:"data"`
}

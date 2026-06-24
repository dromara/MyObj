package tianyi

type ShareInfoResponse struct {
	ShareID      string `json:"shareId"`
	ShareName    string `json:"shareName"`
	ShareTitle   string `json:"shareTitle"`
	ShareDesc    string `json:"shareDesc"`
	ShareMode    int    `json:"shareMode"`
	CreateTime   string `json:"createTime"`
	ExpireTime   string `json:"expireTime"`
	Expired      bool   `json:"expired"`
	FileID       string `json:"fileId"`
	FileName     string `json:"fileName"`
	FileSize     int64  `json:"fileSize"`
	IsFolder     bool   `json:"isFolder"`
	CreatorName  string `json:"creatorName"`
	AccessCode   string `json:"accessCode"`
	NeedAccessCode bool `json:"needAccessCode"`
}

type ShareFileListResponse struct {
	ShareID    string          `json:"shareId"`
	FileListAO ShareFileListAO `json:"fileListAO"`
}

type ShareFileListAO struct {
	Count    int            `json:"count"`
	FileList []ShareFileDTO `json:"fileList"`
}

type ShareFileDTO struct {
	FileID        string `json:"fileId"`
	FileName      string `json:"fileName"`
	FileSize      int64  `json:"fileSize"`
	IsFolder      bool   `json:"isFolder"`
	CreateDate    string `json:"createDate"`
	LastOpTime    string `json:"lastOpTime"`
	FileType      string `json:"fileType"`
	IconURL       string `json:"iconUrl"`
	SmallIconURL  string `json:"smallIconUrl"`
	ParentFolderID string `json:"parentFolderId"`
	StarLabel     int    `json:"starLabel"`
}

type DownloadURLResponse struct {
	FileID      string `json:"fileId"`
	FileName    string `json:"fileName"`
	FileSize    int64  `json:"fileSize"`
	DownloadURL string `json:"downloadUrl"`
}

type APIError struct {
	ErrorCode string `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
}

type LoginShareResponse struct {
	ShareID       string `json:"shareId"`
	ValidationStatus int `json:"validationStatus"`
}

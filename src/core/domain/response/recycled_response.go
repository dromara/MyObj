package response

import "myobj/src/pkg/custom_type"

// RecycledItem 回收站文件项
type RecycledItem struct {
	RecycledID   string               `json:"recycled_id"`
	FileID       string               `json:"file_id"`
	FileName     string               `json:"file_name"`
	FileSize     int64                `json:"file_size"`
	MimeType     string               `json:"mime_type"`
	IsEnc        bool                 `json:"is_enc"`
	HasThumbnail bool                 `json:"has_thumbnail"`
	DeletedAt    custom_type.JsonTime `json:"deleted_at"`
}

// RecycledListResponse 回收站列表响应
type RecycledListResponse struct {
	Items    []*RecycledItem `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

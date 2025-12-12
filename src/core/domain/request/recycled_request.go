package request

// RecycledListRequest 回收站列表请求
type RecycledListRequest struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

// RestoreFileRequest 还原文件请求
type RestoreFileRequest struct {
	RecycledID string `json:"recycled_id" binding:"required"`
}

// DeleteRecycledRequest 永久删除文件请求
type DeleteRecycledRequest struct {
	RecycledID string `json:"recycled_id" binding:"required"`
}

// EmptyRecycledRequest 清空回收站请求
type EmptyRecycledRequest struct {
	// 可以添加确认字段
}

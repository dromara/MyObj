package response

// PageResponse 分页响应结构体
type PageResponse struct {
	// 数据总数
	Total int64 `json:"total"`
	// 当前页
	Page int `json:"page"`
	// 每页数量
	PageSize int `json:"page_size"`
	// 数据
	Data any `json:"data"`
}

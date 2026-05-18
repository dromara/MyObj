package request

// AuditLogListRequest 审计日志查询请求
type AuditLogListRequest struct {
	Page      int    `form:"page" binding:"required,min=1"`
	PageSize  int    `form:"pageSize" binding:"required,min=1,max=100"`
	UserID    string `form:"user_id"`
	Action    string `form:"action"`
	Keyword   string `form:"keyword"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}

// AuditLogExportRequest 审计日志导出请求
type AuditLogExportRequest struct {
	UserID    string `form:"user_id"`
	Action    string `form:"action"`
	Keyword   string `form:"keyword"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}

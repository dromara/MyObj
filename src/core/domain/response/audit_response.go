package response

import "myobj/src/pkg/custom_type"

// AuditLogResponse 审计日志响应
type AuditLogResponse struct {
	ID         string               `json:"id"`
	UserID     string               `json:"user_id"`
	UserName   string               `json:"user_name"`
	Action     string               `json:"action"`
	TargetType string               `json:"target_type"`
	TargetPath string               `json:"target_path"`
	TargetName string               `json:"target_name"`
	Detail     string               `json:"detail"`
	IP         string               `json:"ip"`
	CreatedAt  custom_type.JsonTime `json:"created_at"`
}

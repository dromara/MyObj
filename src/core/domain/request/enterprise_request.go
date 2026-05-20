package request

// CreateEnterpriseRequest 创建企业请求
type CreateEnterpriseRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
}

// UpdateEnterpriseRequest 更新企业信息请求
type UpdateEnterpriseRequest struct {
	EnterpriseID string `json:"enterprise_id" binding:"required"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Logo         string `json:"logo"`
}

// SwitchEnterpriseRequest 切换企业请求
type SwitchEnterpriseRequest struct {
	EnterpriseID string `json:"enterprise_id"` // 空=切换到个人空间
}

// InviteMemberRequest 直接邀请成员请求
type InviteMemberRequest struct {
	EnterpriseID string `json:"enterprise_id" binding:"required"`
	UserName     string `json:"user_name" binding:"required"`
}

// JoinEnterpriseRequest 通过邀请码加入企业请求
type JoinEnterpriseRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

// UpdateMemberRoleRequest 修改成员角色请求
type UpdateMemberRoleRequest struct {
	EnterpriseID string `json:"enterprise_id" binding:"required"`
	MemberID     string `json:"member_id" binding:"required"`
	RoleID       string `json:"role_id" binding:"required"`
}

// RemoveMemberRequest 移除成员请求
type RemoveMemberRequest struct {
	EnterpriseID string `json:"enterprise_id" binding:"required"`
	MemberID     string `json:"member_id" binding:"required"`
}

// LeaveEnterpriseRequest 退出企业请求
type LeaveEnterpriseRequest struct {
	EnterpriseID string `json:"enterprise_id" binding:"required"`
}

// CreateEnterpriseRoleRequest 创建企业角色请求
type CreateEnterpriseRoleRequest struct {
	EnterpriseID string `json:"enterprise_id" binding:"required"`
	Name         string `json:"name" binding:"required"`
	PowerIDs     []int  `json:"power_ids"`
}

// UpdateEnterpriseRoleRequest 更新企业角色请求
type UpdateEnterpriseRoleRequest struct {
	EnterpriseID string `json:"enterprise_id" binding:"required"`
	RoleID       string `json:"role_id" binding:"required"`
	Name         string `json:"name"`
	PowerIDs     []int  `json:"power_ids"`
}

// DeleteEnterpriseRoleRequest 删除企业角色请求
type DeleteEnterpriseRoleRequest struct {
	EnterpriseID string `json:"enterprise_id" binding:"required"`
	RoleID       string `json:"role_id" binding:"required"`
}

// EnterpriseListRequest 企业列表请求（分页查询成员等）
type EnterpriseListRequest struct {
	EnterpriseID string `form:"enterprise_id" binding:"required"`
	Page         int    `form:"page" binding:"required,min=1"`
	PageSize     int    `form:"pageSize" binding:"required,min=1,max=100"`
}

// TransferOwnershipRequest 转让所有权请求
type TransferOwnershipRequest struct {
	EnterpriseID string `json:"enterprise_id" binding:"required"`
	NewOwnerID   string `json:"new_owner_id"`
	NewOwnerName string `json:"new_owner_name"` // 通过用户名查找（与NewOwnerID二选一）
}

// DissolveEnterpriseRequest 解散企业请求
type DissolveEnterpriseRequest struct {
	EnterpriseID string `json:"enterprise_id" binding:"required"`
}

// ToggleEnterpriseStateRequest 启用/禁用企业请求
type ToggleEnterpriseStateRequest struct {
	EnterpriseID string `json:"enterprise_id" binding:"required"`
	State        int    `json:"state" binding:"required,oneof=0 1"` // 0=正常, 1=禁用
}

// SetEnterpriseQuotaRequest 设置企业存储配额请求
type SetEnterpriseQuotaRequest struct {
	EnterpriseID   string `json:"enterprise_id" binding:"required"`
	Space          int64  `json:"space" binding:"min=0"` // 总配额（字节）
	SpaceUnlimited bool   `json:"space_unlimited"`        // true=不限制空间
}

// EnterpriseAuditListRequest 企业审计日志查询请求
type EnterpriseAuditListRequest struct {
	EnterpriseID string `form:"enterprise_id" binding:"required"`
	Action       string `form:"action"`
	Keyword      string `form:"keyword"`
	StartTime    string `form:"start_time"`
	EndTime      string `form:"end_time"`
	Page         int    `form:"page" binding:"required,min=1"`
	PageSize     int    `form:"pageSize" binding:"required,min=1,max=100"`
}

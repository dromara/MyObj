package response

import "myobj/src/pkg/custom_type"

// EnterpriseResponse 企业信息响应
type EnterpriseResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Logo        string               `json:"logo"`
	Description string               `json:"description"`
	CreatorID   string               `json:"creator_id"`
	Space       int64                `json:"space"`
	FreeSpace   int64                `json:"free_space"`
	InviteCode  string               `json:"invite_code"`
	State       int                  `json:"state"`
	CreatedAt   custom_type.JsonTime `json:"created_at"`
	MemberCount int64                `json:"member_count"`
	Role        string               `json:"role"` // 当前用户在该企业的角色名
}

// EnterpriseMemberResponse 企业成员响应
type EnterpriseMemberResponse struct {
	ID         string               `json:"id"`
	UserID     string               `json:"user_id"`
	UserName   string               `json:"user_name"`
	UserAvatar string               `json:"user_avatar"`
	RoleID     string               `json:"role_id"`
	RoleName   string               `json:"role_name"`
	IsAdmin    int                  `json:"is_admin"`
	Status     int                  `json:"status"`
	JoinedAt   custom_type.JsonTime `json:"joined_at"`
}

// EnterpriseRoleResponse 企业角色响应
type EnterpriseRoleResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault int    `json:"is_default"`
	IsAdmin   int    `json:"is_admin"`
	PowerIDs  []int  `json:"power_ids"`
}

// EnterpriseInviteResponse 企业邀请响应
type EnterpriseInviteResponse struct {
	ID             string               `json:"id"`
	EnterpriseID   string               `json:"enterprise_id"`
	EnterpriseName string               `json:"enterprise_name"`
	InviterID      string               `json:"inviter_id"`
	InviterName    string               `json:"inviter_name"`
	Status         int                  `json:"status"`
	CreatedAt      custom_type.JsonTime `json:"created_at"`
}

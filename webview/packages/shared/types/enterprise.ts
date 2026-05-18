// 企业信息
export interface Enterprise {
  id: string
  name: string
  logo: string
  description: string
  creator_id: string
  space: number
  free_space: number
  invite_code: string
  invite_link: string
  state: number
  created_at: string
  member_count?: number
  role?: string
}

// 企业成员
export interface EnterpriseMember {
  id: string
  user_id: string
  user_name: string
  user_avatar: string
  role_id: string
  role_name: string
  is_admin: number
  status: number
  joined_at: string
}

// 企业角色
export interface EnterpriseRole {
  id: string
  name: string
  is_default: number
  is_admin: number
  power_ids: number[]
}

// 企业审计日志
export interface EnterpriseAuditLog {
  id: string
  user_id: string
  user_name: string
  enterprise_id: string
  action: string
  target_type: string
  target_path: string
  target_name: string
  detail: string
  ip: string
  created_at: string
}

// 共享空间文件
export interface SharedFileEntry {
  id: string
  file_id: string
  file_name: string
  path_id: number
  uploader_id: string
  size: number
  created_at: string
}

// 共享空间目录
export interface SharedDirEntry {
  id: number
  enterprise_id: string
  name: string
  parent_id: number
  created_at: string
}

// 空间使用情况
export interface SpaceUsage {
  total_space: number
  free_space: number
  used_space: number
  file_count: number
}

// 权限信息
export interface EnterprisePower {
  id: number
  name: string
  description: string
  characteristic: string
}

// ========== 请求类型 ==========

export interface CreateEnterpriseRequest {
  name: string
  description?: string
  logo?: string
}

export interface UpdateEnterpriseRequest {
  enterprise_id: string
  name?: string
  description?: string
  logo?: string
}

export interface SwitchEnterpriseRequest {
  enterprise_id?: string
}

export interface InviteMemberRequest {
  enterprise_id: string
  user_name: string
}

export interface JoinEnterpriseRequest {
  invite_code: string
}

export interface UpdateMemberRoleRequest {
  enterprise_id: string
  member_id: string
  role_id: string
}

export interface RemoveMemberRequest {
  enterprise_id: string
  member_id: string
}

export interface LeaveEnterpriseRequest {
  enterprise_id: string
}

export interface CreateRoleRequest {
  enterprise_id: string
  name: string
  power_ids?: number[]
}

export interface UpdateRoleRequest {
  role_id: string
  name?: string
  power_ids?: number[]
}

export interface DeleteRoleRequest {
  role_id: string
}

export interface TransferOwnershipRequest {
  enterprise_id: string
  new_owner_id?: string
  new_owner_name?: string
}

export interface DissolveEnterpriseRequest {
  enterprise_id: string
}

export interface ToggleEnterpriseStateRequest {
  enterprise_id: string
  state: number
}

export interface SetEnterpriseQuotaRequest {
  enterprise_id: string
  space: number
}

export interface EnterpriseListRequest {
  enterprise_id: string
  page: number
  pageSize: number
}

export interface EnterpriseAuditListRequest {
  enterprise_id: string
  action?: string
  keyword?: string
  start_time?: string
  end_time?: string
  page: number
  pageSize: number
}

export interface SharedFileListRequest {
  enterprise_id: string
  path_id?: number
  page: number
  pageSize: number
}

export interface CreateSharedDirRequest {
  enterprise_id: string
  name: string
  parent_id?: number
}

export interface DeleteSharedFileRequest {
  id: string
}

export interface SharedUploadPrecheckRequest {
  enterprise_id: string
  file_name: string
  file_size: number
  chunk_signature?: string
  path_id?: number
}

// ========== 响应类型 ==========

export interface EnterpriseListResponse {
  list: Enterprise[]
}

export interface EnterpriseMemberListResponse {
  list: EnterpriseMember[]
  total: number
}

export interface EnterpriseRoleListResponse {
  list: EnterpriseRole[]
}

export interface EnterpriseAuditListResponse {
  list: EnterpriseAuditLog[]
  total: number
  page: number
  pageSize: number
}

export interface SharedFileListResponse {
  dirs: SharedDirEntry[]
  files: SharedFileEntry[]
  total: number
  page: number
  pageSize: number
}

export interface InviteCodeResponse {
  invite_code: string
  invite_link: string
}

export interface PendingInviteResponse {
  id: string
  enterprise_id: string
  enterprise_name: string
  inviter_id: string
  inviter_name: string
  status: number
  created_at: string
}

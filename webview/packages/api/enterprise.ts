import { get, post, put, del, download } from '@myobj/http'
import { filterParams, API_BASE_URL, API_ENDPOINTS } from '@myobj/shared'
import type {
  ApiResponse,
  Enterprise,
  EnterpriseListResponse,
  EnterpriseMemberListResponse,
  EnterpriseRoleListResponse,
  EnterpriseAuditListResponse,
  SharedFileListResponse,
  SpaceUsage,
  EnterprisePower,
  InviteCodeResponse,
  PendingInviteResponse,
  CreateEnterpriseRequest,
  UpdateEnterpriseRequest,
  SwitchEnterpriseRequest,
  InviteMemberRequest,
  JoinEnterpriseRequest,
  UpdateMemberRoleRequest,
  RemoveMemberRequest,
  LeaveEnterpriseRequest,
  CreateRoleRequest,
  UpdateRoleRequest,
  DeleteRoleRequest,
  TransferOwnershipRequest,
  DissolveEnterpriseRequest,
  ToggleEnterpriseStateRequest,
  SetEnterpriseQuotaRequest,
  EnterpriseListRequest,
  EnterpriseAuditListRequest,
  SharedFileListRequest,
  CreateSharedDirRequest,
  DeleteSharedFileRequest,
  SharedUploadPrecheckRequest
} from '@myobj/shared'

// ========== 企业管理 API ==========

export const createEnterprise = (data: CreateEnterpriseRequest) => {
  return post<ApiResponse<{ enterprise_id: string; invite_code: string }>>(API_ENDPOINTS.ENTERPRISE.CREATE, data)
}

export const getMyEnterprises = () => {
  return get<ApiResponse<EnterpriseListResponse>>(API_ENDPOINTS.ENTERPRISE.LIST)
}

export const getEnterpriseInfo = (enterprise_id: string) => {
  return get<ApiResponse<Enterprise>>(API_ENDPOINTS.ENTERPRISE.INFO, { enterprise_id })
}

export const updateEnterprise = (data: UpdateEnterpriseRequest) => {
  return put<ApiResponse>(API_ENDPOINTS.ENTERPRISE.UPDATE, data)
}

export const switchEnterprise = (data: SwitchEnterpriseRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SWITCH, data)
}

export const transferOwnership = (data: TransferOwnershipRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.TRANSFER, data)
}

export const dissolveEnterprise = (data: DissolveEnterpriseRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.DISSOLVE, data)
}

export const toggleEnterpriseState = (data: ToggleEnterpriseStateRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.TOGGLE_STATE, data)
}

export const setEnterpriseQuota = (data: SetEnterpriseQuotaRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SET_QUOTA, data)
}

// ========== 成员管理 API ==========

export const getMemberList = (params: EnterpriseListRequest) => {
  return get<ApiResponse<EnterpriseMemberListResponse>>(API_ENDPOINTS.ENTERPRISE.MEMBER.LIST, filterParams(params))
}

export const inviteMember = (data: InviteMemberRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.MEMBER.INVITE, data)
}

export const removeMember = (data: RemoveMemberRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.MEMBER.REMOVE, data)
}

export const updateMemberRole = (data: UpdateMemberRoleRequest) => {
  return put<ApiResponse>(API_ENDPOINTS.ENTERPRISE.MEMBER.UPDATE_ROLE, data)
}

export const leaveEnterprise = (data: LeaveEnterpriseRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.MEMBER.LEAVE, data)
}

export const joinEnterprise = (data: JoinEnterpriseRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.MEMBER.JOIN, data)
}

export const acceptInvite = (invite_id: string) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.MEMBER.ACCEPT, {}, { params: { invite_id } })
}

export const getInviteCode = (enterprise_id: string) => {
  return get<ApiResponse<InviteCodeResponse>>(API_ENDPOINTS.ENTERPRISE.MEMBER.INVITE_CODE, { enterprise_id })
}

export const refreshInviteCode = (enterprise_id: string) => {
  return post<ApiResponse<{ invite_code: string }>>(API_ENDPOINTS.ENTERPRISE.MEMBER.REFRESH_CODE, {}, { params: { enterprise_id } })
}

export const getPendingInvites = () => {
  return get<ApiResponse<PendingInviteResponse[]>>(API_ENDPOINTS.ENTERPRISE.MEMBER.PENDING)
}

// ========== 角色管理 API ==========

export const getRoleList = (enterprise_id: string) => {
  return get<ApiResponse<EnterpriseRoleListResponse>>(API_ENDPOINTS.ENTERPRISE.ROLE.LIST, { enterprise_id })
}

export const createRole = (data: CreateRoleRequest) => {
  return post<ApiResponse<{ role_id: string }>>(API_ENDPOINTS.ENTERPRISE.ROLE.CREATE, data)
}

export const updateRole = (data: UpdateRoleRequest) => {
  return put<ApiResponse>(API_ENDPOINTS.ENTERPRISE.ROLE.UPDATE, data)
}

export const deleteRole = (data: DeleteRoleRequest) => {
  return del<ApiResponse>(API_ENDPOINTS.ENTERPRISE.ROLE.DELETE, { data })
}

export const getAllPowers = () => {
  return get<ApiResponse<EnterprisePower[]>>(API_ENDPOINTS.ENTERPRISE.POWERS)
}

// ========== 共享空间 API ==========

export const createSharedDir = (data: CreateSharedDirRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SPACE.MKDIR, data)
}

export const getSharedFileList = (params: SharedFileListRequest) => {
  return get<ApiResponse<SharedFileListResponse>>(API_ENDPOINTS.ENTERPRISE.SPACE.LIST, filterParams(params))
}

export const sharedUploadPrecheck = (data: SharedUploadPrecheckRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SPACE.UPLOAD_PRECHECK, data)
}

export const deleteSharedFile = (data: DeleteSharedFileRequest, enterprise_id?: string) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SPACE.DELETE, data, enterprise_id ? { params: { enterprise_id } } : {})
}

export const downloadSharedFile = (id: string, enterprise_id?: string) => {
  let url = API_ENDPOINTS.ENTERPRISE.SPACE.DOWNLOAD + '?id=' + id
  if (enterprise_id) url += '&enterprise_id=' + enterprise_id
  return download(url, id)
}

export const getSpaceUsage = (enterprise_id: string) => {
  return get<ApiResponse<SpaceUsage>>(API_ENDPOINTS.ENTERPRISE.SPACE.USAGE, { enterprise_id })
}

export const deleteSharedDir = (id: number, enterprise_id?: string) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SPACE.DELETE_DIR, { id }, enterprise_id ? { params: { enterprise_id } } : {})
}

export const renameSharedFile = (id: string, name: string, enterprise_id?: string) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SPACE.RENAME, { id, name }, enterprise_id ? { params: { enterprise_id } } : {})
}

export const renameSharedDir = (id: number, name: string, enterprise_id?: string) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SPACE.RENAME_DIR, { id, name }, enterprise_id ? { params: { enterprise_id } } : {})
}

// 预览（返回 URL）
export const previewSharedFile = (id: string) => {
  return API_BASE_URL + API_ENDPOINTS.ENTERPRISE.SPACE.PREVIEW + '?id=' + id
}

// 缩略图（返回 URL）
export const getSharedFileThumbnail = (fileId: string) => {
  return API_BASE_URL + API_ENDPOINTS.ENTERPRISE.SPACE.THUMBNAIL + '/' + fileId
}

// 搜索文件
export const searchSharedFiles = (params: { enterprise_id: string; keyword: string; page: number; pageSize: number }) => {
  return get<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SPACE.SEARCH, filterParams(params))
}

// 获取目录树
export const getSharedPathTree = (enterprise_id: string) => {
  return get<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SPACE.PATH_TREE, { enterprise_id })
}

// 移动文件
export const moveSharedFile = (data: { enterprise_id: string; file_id: string; target_path_id: number }) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SPACE.MOVE, data)
}

// 创建打包下载任务
export const createPackage = (data: { enterprise_id: string; file_ids: string[]; package_name?: string }) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SPACE.PACKAGE, data)
}

// 查询打包进度
export const getPackageProgress = (package_id: string) => {
  return get<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SPACE.PACKAGE_PROGRESS, { package_id })
}

// 下载打包文件（返回 URL）
export const downloadPackage = (package_id: string) => {
  return API_ENDPOINTS.ENTERPRISE.SPACE.PACKAGE_DOWNLOAD + '?package_id=' + package_id
}

// 解压冲突检测
export const extractCheck = (data: { enterprise_id: string; file_id: string; target_path_id?: number; file_password?: string }) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SPACE.EXTRACT_CHECK, data)
}

// 开始解压
export const extractStart = (data: { enterprise_id: string; file_id: string; target_path_id?: number; file_password?: string; conflict_strategy?: string }) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SPACE.EXTRACT, data)
}

// 查询解压进度
export const getExtractProgress = (task_id: string) => {
  return get<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SPACE.EXTRACT_PROGRESS, { task_id })
}

// 创建文件分享
export const createShare = (data: { enterprise_id: string; file_id: string; expire?: number; password?: string }) => {
  return post<ApiResponse>(API_ENDPOINTS.ENTERPRISE.SPACE.SHARE, data)
}

// ========== 审计日志 API ==========

export const getEnterpriseAuditLogs = (params: EnterpriseAuditListRequest) => {
  return get<ApiResponse<EnterpriseAuditListResponse>>(API_ENDPOINTS.ENTERPRISE.AUDIT.LIST, filterParams(params))
}

export const exportEnterpriseAuditLogs = (params: Omit<EnterpriseAuditListRequest, 'page' | 'pageSize'>) => {
  const query = new URLSearchParams()
  query.set('enterprise_id', params.enterprise_id)
  if (params.action) query.set('action', params.action)
  if (params.keyword) query.set('keyword', params.keyword)
  if (params.start_time) query.set('start_time', params.start_time)
  if (params.end_time) query.set('end_time', params.end_time)
  const qs = query.toString()
  const url = API_ENDPOINTS.ENTERPRISE.AUDIT.EXPORT + (qs ? '?' + qs : '')
  const filename = `enterprise_audit_${new Date().toISOString().slice(0, 19).replace(/[-:T]/g, '')}.csv`
  return download(url, filename)
}

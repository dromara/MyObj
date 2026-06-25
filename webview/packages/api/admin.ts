import { get, post } from '@myobj/http'
import { filterParams } from '@myobj/shared'
import { API_ENDPOINTS } from '@myobj/shared'
import type { ApiResponse, AdminUser, CreateUserRequest, UpdateUserRequest, UserListRequest, UserListResponse, AdminGroup, CreateGroupRequest, UpdateGroupRequest, GroupListResponse, AdminPower, PowerListResponse, CreatePowerRequest, UpdatePowerRequest, BatchDeletePowerRequest, AssignPowerRequest, AdminDisk, CreateDiskRequest, UpdateDiskRequest, DiskListResponse, ScannedDiskInfo, SystemConfig, UpdateSystemConfigRequest } from '@myobj/shared'
// ========== 用户管理 API ==========

/**
 * 获取用户列表
 */
export const getAdminUserList = (params: UserListRequest) => {
  return get<ApiResponse<UserListResponse>>(API_ENDPOINTS.ADMIN.USER.LIST, filterParams(params))
}

/**
 * 创建用户
 */
export const createAdminUser = (data: CreateUserRequest) => {
  return post<ApiResponse<AdminUser>>(API_ENDPOINTS.ADMIN.USER.CREATE, data)
}

/**
 * 更新用户
 */
export const updateAdminUser = (data: UpdateUserRequest) => {
  return post<ApiResponse<AdminUser>>(API_ENDPOINTS.ADMIN.USER.UPDATE, data)
}

/**
 * 删除用户
 */
export const deleteAdminUser = (id: string) => {
  return post<ApiResponse>(API_ENDPOINTS.ADMIN.USER.DELETE, { id })
}

/**
 * 启用/禁用用户
 */
export const toggleUserState = (id: string, state: number) => {
  return post<ApiResponse>(API_ENDPOINTS.ADMIN.USER.TOGGLE_STATE, { id, state })
}

// ========== 组管理 API ==========

/**
 * 获取组列表
 */
export const getAdminGroupList = () => {
  return get<ApiResponse<GroupListResponse>>(API_ENDPOINTS.ADMIN.GROUP.LIST)
}

/**
 * 创建组
 */
export const createAdminGroup = (data: CreateGroupRequest) => {
  return post<ApiResponse<AdminGroup>>(API_ENDPOINTS.ADMIN.GROUP.CREATE, data)
}

/**
 * 更新组
 */
export const updateAdminGroup = (data: UpdateGroupRequest) => {
  return post<ApiResponse<AdminGroup>>(API_ENDPOINTS.ADMIN.GROUP.UPDATE, data)
}

/**
 * 删除组
 */
export const deleteAdminGroup = (id: number) => {
  return post<ApiResponse>(API_ENDPOINTS.ADMIN.GROUP.DELETE, { id })
}

// ========== 权限管理 API ==========

/**
 * 获取权限列表
 */
export const getAdminPowerList = (params?: { page?: number; pageSize?: number }) => {
  return get<ApiResponse<PowerListResponse>>(API_ENDPOINTS.ADMIN.POWER.LIST, params)
}

/**
 * 创建权限
 */
export const createAdminPower = (data: CreatePowerRequest) => {
  return post<ApiResponse<AdminPower>>(API_ENDPOINTS.ADMIN.POWER.CREATE, data)
}

/**
 * 更新权限
 */
export const updateAdminPower = (data: UpdatePowerRequest) => {
  return post<ApiResponse<AdminPower>>(API_ENDPOINTS.ADMIN.POWER.UPDATE, data)
}

/**
 * 删除权限
 */
export const deleteAdminPower = (id: number) => {
  return post<ApiResponse>(API_ENDPOINTS.ADMIN.POWER.DELETE, { id })
}

/**
 * 批量删除权限
 */
export const batchDeleteAdminPower = (data: BatchDeletePowerRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.ADMIN.POWER.BATCH_DELETE, data)
}

/**
 * 为组分配权限
 */
export const assignPowerToGroup = (data: AssignPowerRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.ADMIN.POWER.ASSIGN, data)
}

/**
 * 获取组的权限列表
 */
export const getGroupPowers = (group_id: number) => {
  return get<ApiResponse<{ power_ids: number[] }>>(API_ENDPOINTS.ADMIN.POWER.GROUP_POWERS, filterParams({ group_id }))
}

// ========== 磁盘管理 API ==========

/**
 * 获取磁盘列表
 */
export const getAdminDiskList = () => {
  return get<ApiResponse<DiskListResponse>>(API_ENDPOINTS.ADMIN.DISK.LIST)
}

/**
 * 创建磁盘
 */
export const createAdminDisk = (data: CreateDiskRequest) => {
  return post<ApiResponse<AdminDisk>>(API_ENDPOINTS.ADMIN.DISK.CREATE, data)
}

/**
 * 更新磁盘
 */
export const updateAdminDisk = (data: UpdateDiskRequest) => {
  return post<ApiResponse<AdminDisk>>(API_ENDPOINTS.ADMIN.DISK.UPDATE, data)
}

/**
 * 删除磁盘
 */
export const deleteAdminDisk = (id: string) => {
  return post<ApiResponse>(API_ENDPOINTS.ADMIN.DISK.DELETE, { id })
}

/**
 * 扫描磁盘
 */
export const scanDisks = () => {
  return get<ApiResponse<ScannedDiskInfo[]>>(API_ENDPOINTS.ADMIN.DISK.SCAN)
}

// ========== 系统配置 API ==========

/**
 * 获取系统配置
 */
export const getSystemConfig = () => {
  return get<ApiResponse<SystemConfig>>(API_ENDPOINTS.ADMIN.SYSTEM.CONFIG)
}

/**
 * 更新系统配置
 */
export const updateSystemConfig = (data: UpdateSystemConfigRequest) => {
  return post<ApiResponse<SystemConfig>>(API_ENDPOINTS.ADMIN.SYSTEM.UPDATE_CONFIG, data)
}

// ========== 系统升级 API ==========

export interface UpdateInfo {
  has_update: boolean
  current_version: string
  latest_version: string
  release_note: string
  download_url: string
  file_size: number
}

/**
 * 检查更新
 */
export const checkUpdate = () => {
  return get<ApiResponse<UpdateInfo>>(API_ENDPOINTS.ADMIN.SYSTEM.CHECK_UPDATE)
}

/**
 * 执行升级
 */
export const performUpgrade = (downloadUrl: string) => {
  return post<ApiResponse>(API_ENDPOINTS.ADMIN.SYSTEM.UPGRADE, { download_url: downloadUrl })
}

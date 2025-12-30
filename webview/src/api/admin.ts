import { get, post } from '@/utils/request'
import { filterParams } from '@/utils/params'
import { API_ENDPOINTS } from '@/config/api'
import type { ApiResponse } from '@/types'

// 用户管理相关类型
export interface AdminUser {
  id: string
  name: string
  user_name: string
  email: string
  phone: string
  group_id: number
  group_name?: string
  space: number
  free_space: number
  state: number
  created_at: string
}

export interface CreateUserRequest {
  user_name: string
  password: string
  name: string
  email: string
  phone: string
  group_id: number
  space: number
}

export interface UpdateUserRequest {
  id: string
  name?: string
  email?: string
  phone?: string
  group_id?: number
  space?: number
  state?: number
}

export interface UserListRequest {
  page: number
  pageSize: number
  keyword?: string
  group_id?: number
  state?: number
}

export interface UserListResponse {
  users: AdminUser[]
  total: number
  page: number
  page_size: number
}

// 组管理相关类型
export interface AdminGroup {
  id: number
  name: string
  group_default: number
  space: number
  created_at: string
}

export interface CreateGroupRequest {
  name: string
  space: number
  group_default?: number
}

export interface UpdateGroupRequest {
  id: number
  name?: string
  space?: number
  group_default?: number
}

export interface GroupListResponse {
  groups: AdminGroup[]
  total: number
}

// 权限管理相关类型
export interface AdminPower {
  id: number
  name: string
  description: string
  characteristic: string
  created_at: string
}

export interface PowerListResponse {
  powers: AdminPower[]
  total: number
}

export interface CreatePowerRequest {
  name: string
  description: string
  characteristic: string
}

export interface UpdatePowerRequest {
  id: number
  name?: string
  description?: string
  characteristic?: string
}

export interface BatchDeletePowerRequest {
  ids: number[]
}

export interface AssignPowerRequest {
  group_id: number
  power_ids: number[]
}

// 磁盘管理相关类型
export interface AdminDisk {
  id: string
  size: number
  disk_path: string
  data_path: string
}

export interface CreateDiskRequest {
  disk_path: string
  data_path: string
  size: number
}

export interface UpdateDiskRequest {
  id: string
  disk_path?: string
  data_path?: string
  size?: number
}

export interface DiskListResponse {
  disks: AdminDisk[]
  total: number
}

// 系统配置相关类型
export interface SystemConfig {
  allow_register: boolean
  webdav_enabled: boolean
  version: string
  total_users: number
  total_files: number
  [key: string]: any
}

export interface UpdateSystemConfigRequest {
  allow_register?: boolean
  webdav_enabled?: boolean
  [key: string]: any
}

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
export const getAdminPowerList = () => {
  return get<ApiResponse<PowerListResponse>>(API_ENDPOINTS.ADMIN.POWER.LIST)
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


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

// 扫描磁盘信息类型（对应后端 DiskInfo）
export interface ScannedDiskInfo {
  mount: string
  total: number
  used: number
  free: number
  avail: number
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

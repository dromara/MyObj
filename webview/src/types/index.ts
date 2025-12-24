// API配置类型定义
export interface ApiConfig {
  baseURL: string
  timeout: number
}

// 用户信息类型
export interface UserInfo {
  id: string
  name: string
  user_name: string
  password: string
  email: string
  phone: string
  group_id: number
  created_at: string
  space: number
  file_password: string
  free_space: number
  state: number
}

// 存储信息类型
export interface StorageInfo {
  used: number
  total: number
  percentage: number
}

// 登录请求参数
export interface LoginRequest {
  username: string
  password: string
  challenge: string
}

// 注册请求参数
export interface RegisterRequest {
  username: string
  password: string
  email: string
  challenge: string
}

// 权限信息类型
export interface PowerInfo {
  id: number
  name: string
  description: string
  characteristic: string
  created_at: string
}

// 登录数据类型
export interface LoginData {
  token: string
  user_info: UserInfo
  power: PowerInfo[]
}

// 登录响应
export interface LoginResponse {
  code: number
  message: string
  data: LoginData
}

// 文件信息类型
export interface FileInfo {
  id: number
  name: string
  type: 'folder' | 'file'
  size: string | number
  modified: string
  icon?: string
  path?: string
  isPublic?: boolean
  shareUrl?: string
}

// 面包屑项
export interface Breadcrumb {
  id: number
  name: string
  path: string
}

// 目录项
export interface FolderItem {
  id: number
  name: string
  path: string
  created_time: string
}

// 文件项
export interface FileItem {
  file_id: string
  file_name: string
  file_size: number
  mime_type: string
  is_enc: boolean
  has_thumbnail: boolean
  public: boolean
  created_at: string
}

// 文件列表请求
export interface FileListRequest {
  virtualPath?: string
  type?: string
  sortBy?: string
  page: number
  pageSize: number
}

// 文件列表响应
export interface FileListResponse {
  breadcrumbs: Breadcrumb[]
  current_path: string
  folders: FolderItem[]
  files: FileItem[]
  total: number
  page: number
  page_size: number
}

// 分享信息类型
export interface ShareInfo {
  id: number
  user_id: string
  file_id: string
  file_name: string // 用户文件名
  token: string
  expires_at: string
  password_hash: string
  download_count: number
  created_at: string
}

// 创建分享请求
export interface CreateShareRequest {
  file_id: string
  expire: string
  password: string
}

// 创建分享响应
export interface CreateShareResponse {
  code: number
  message: string
  data: string // 分享链接
}

// 离线下载任务
export interface OfflineTask {
  id: number
  url: string
  fileName: string
  status: 'pending' | 'downloading' | 'completed' | 'failed'
  progress: number
  speed?: string
  createdAt: string
}

// API响应基础类型
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 路由菜单项类型
export interface MenuItem {
  id: number
  name: string
  icon: string
  type: string
  active: boolean
}

// 密码挑战响应
export interface ChallengeResponse {
  code: number
  message: string
  data: {
    publicKey: string
    id: string
  }
}

// 修改密码请求
export interface UpdatePasswordRequest {
  old_passwd: string
  new_passwd: string
  challenge: string
}

// 设置文件密码请求
export interface SetFilePasswordRequest {
  passwd: string
  challenge: string
}

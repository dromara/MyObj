import { get, post } from '@/utils/request'
import { filterParams } from '@/utils/params'
import { API_ENDPOINTS, API_BASE_URL } from '@/config/api'
import type { ApiResponse } from '@/types'
import cache from '@/plugins/cache'

// 离线下载任务类型
export interface OfflineDownloadTask {
  id: string
  url: string
  file_name: string
  file_size: number
  downloaded_size: number
  progress: number
  speed: number
  type: number
  type_text: string
  state: number
  state_text: string
  virtual_path: string
  support_range: boolean
  error_msg: string
  file_id: string
  create_time: string
  update_time: string
  finish_time: string
}

// 创建离线下载任务请求
export interface CreateOfflineDownloadRequest {
  url: string
  virtual_path?: string
  enable_encryption?: boolean
  file_password?: string
}

// 下载任务列表响应
export interface DownloadTaskListResponse {
  tasks: OfflineDownloadTask[]
  total: number
  page: number
  page_size: number
}

/**
 * 获取下载任务列表
 */
export const getDownloadTaskList = (params: { 
  page: number
  pageSize: number
  state?: number
  type?: number
}) => {
  const filteredParams = filterParams(params)
  return get<ApiResponse<DownloadTaskListResponse>>(
    API_ENDPOINTS.DOWNLOAD.LIST,
    filteredParams
  )
}

/**
 * 创建离线下载任务
 */
export const createOfflineDownload = (data: CreateOfflineDownloadRequest) => {
  return post<ApiResponse<OfflineDownloadTask>>(
    API_ENDPOINTS.DOWNLOAD.CREATE_OFFLINE,
    data
  )
}

/**
 * 暂停下载任务
 */
export const pauseDownload = (taskId: string) => {
  return post<ApiResponse>(API_ENDPOINTS.DOWNLOAD.PAUSE, { task_id: taskId })
}

/**
 * 恢复下载任务
 */
export const resumeDownload = (taskId: string) => {
  return post<ApiResponse>(API_ENDPOINTS.DOWNLOAD.RESUME, { task_id: taskId })
}

/**
 * 取消下载任务
 */
export const cancelDownload = (taskId: string) => {
  return post<ApiResponse>(API_ENDPOINTS.DOWNLOAD.CANCEL, { task_id: taskId })
}

// 删除下载任务请求
export interface DeleteDownloadRequest {
  task_id: string
}

/**
 * 删除下载任务
 */
export const deleteDownload = (taskId: string) => {
  return post<ApiResponse>(API_ENDPOINTS.DOWNLOAD.DELETE, { task_id: taskId })
}

// 创建网盘文件下载任务请求
export interface CreateLocalFileDownloadRequest {
  file_id: string
  file_password?: string
}

/**
 * 创建网盘文件下载任务
 */
export const createLocalFileDownload = (data: CreateLocalFileDownloadRequest) => {
  return post<ApiResponse<{ task_id: string; file_name: string; file_size: number }>>(
    API_ENDPOINTS.DOWNLOAD.LOCAL_CREATE,
    data
  )
}

/**
 * 获取网盘文件下载链接
 */
export const getLocalFileDownloadUrl = (taskId: string) => {
  const token = cache.local.get('token')
  return `${API_BASE_URL}/download/local/file/${taskId}?token=${token}`
}

// 种子文件信息
export interface TorrentFileInfo {
  index: number
  name: string
  size: number
  path: string
}

// 解析种子响应
export interface ParseTorrentResponse {
  name: string
  info_hash: string
  files: TorrentFileInfo[]
  total_size: number
}

// 解析种子请求
export interface ParseTorrentRequest {
  content: string // 种子文件内容（Base64编码）或磁力链接（magnet:开头）
}

// 开始种子下载请求
export interface StartTorrentDownloadRequest {
  content: string // 种子文件内容（Base64编码）或磁力链接
  file_indexes: number[] // 要下载的文件索引列表
  virtual_path?: string // 保存的虚拟路径（可选，默认为/离线下载/）
  enable_encryption?: boolean // 是否加密存储
  file_password?: string // 文件密码（加密文件必需）
}

// 开始种子下载响应
export interface StartTorrentDownloadResponse {
  task_ids: string[] // 创建的任务ID列表
  torrent_name: string // 种子名称
  task_count: number // 创建的任务数量
}

/**
 * 解析种子/磁力链
 */
export const parseTorrent = (data: ParseTorrentRequest) => {
  return post<ApiResponse<ParseTorrentResponse>>(
    API_ENDPOINTS.DOWNLOAD.TORRENT_PARSE,
    data
  )
}

/**
 * 开始种子/磁力链下载
 */
export const startTorrentDownload = (data: StartTorrentDownloadRequest) => {
  return post<ApiResponse<StartTorrentDownloadResponse>>(
    API_ENDPOINTS.DOWNLOAD.TORRENT_START,
    data
  )
}

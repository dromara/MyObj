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

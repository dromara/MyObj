import { get, post, upload } from '@/utils/request'
import { API_ENDPOINTS, API_BASE_URL } from '@/config/api'
import type { FileListRequest, FileListResponse, ApiResponse } from '@/types'
import logger from '@/plugins/logger'

// 文件搜索请求参数
export interface FileSearchParams {
  keyword: string
  type?: string
  sortBy?: string
  page?: number
  pageSize?: number
}

// 文件信息
export interface FileInfo {
  id: string
  name: string
  type: string
  size: number
  mime: string
  ownerName?: string
  viewCount?: number
  downloadCount?: number
  createdAt: string
  updatedAt: string
}

// 搜索响应
export interface SearchResponse {
  code: number
  message: string
  data: {
    files: FileInfo[]
    total: number
  }
}

/**
 * 获取文件列表
 */
export const getFileList = (params: FileListRequest) => {
  return get<ApiResponse<FileListResponse>>(API_ENDPOINTS.FILE.LIST, params)
}

/**
 * 获取文件缩略图（带鉴权）
 */
export const getThumbnail = async (fileId: string): Promise<string> => {
  try {
    const url = `${API_BASE_URL}${API_ENDPOINTS.FILE.THUMBNAIL}/${fileId}`
    
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}` || ''
      }
    })

    if (!response.ok) {
      throw new Error('Failed to fetch thumbnail')
    }

    const blob = await response.blob()
    return URL.createObjectURL(blob)
  } catch (error) {
    logger.error('Error fetching thumbnail:', error)
    return '' // 返回空字符串表示加载失败
  }
}

/**
 * 获取文件缩略图URL
 */
export const getThumbnailUrl = (fileId: string) => {
  return `${API_ENDPOINTS.FILE.THUMBNAIL}/${fileId}`
}

/**
 * 搜索当前用户的文件
 */
export const searchUserFiles = (params: FileSearchParams) => {
  return get<SearchResponse>(API_ENDPOINTS.FILE.SEARCH_USER, params)
}

/**
 * 搜索广场公开文件
 */
export const searchPublicFiles = (params: FileSearchParams) => {
  return get<SearchResponse>(API_ENDPOINTS.FILE.SEARCH_PUBLIC, params)
}

/**
 * 下载文件
 */
export const downloadFile = (fileId: string) => {
  return get(`${API_ENDPOINTS.FILE.DOWNLOAD}/${fileId}`).then(() => {
    // 处理下载逻辑
  })
}

/**
 * 移动文件请求参数
 */
export interface MoveFileRequest {
  file_id: string
  source_path: string
  target_path: string
}

/**
 * 移动文件
 */
export const moveFile = (data: MoveFileRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.FILE.MOVE, data)
}

/**
 * 获取虚拟路径树
 */
export const getVirtualPathTree = () => {
  return get<ApiResponse>(API_ENDPOINTS.FILE.LIST.replace('/list', '/virtualPath'))
}

/**
 * 删除文件请求参数
 */
export interface DeleteFileRequest {
  file_ids: string[]
}

/**
 * 删除文件（移动到回收站）
 */
export const deleteFiles = (data: DeleteFileRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.FILE.DELETE, data)
}

/**
 * 文件重命名请求参数
 */
export interface RenameFileRequest {
  file_id: string
  new_file_name: string
}

/**
 * 重命名文件
 */
export const renameFile = (data: RenameFileRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.FILE.RENAME, data)
}

// 上传文件请求参数
export interface uploadPrecheckParams {
  chunk_signature: string,
  file_name: string,
  file_size: number,
  files_md5: string[],
  path_id: string
}

/**
 * 上传文件预检
 */
export const uploadPrecheck = (data: uploadPrecheckParams) => {
  return post<ApiResponse>(API_ENDPOINTS.FILE.PRECHECK, data)
}

// 上传进度响应
export interface UploadProgressResponse {
  precheck_id: string
  file_name: string
  file_size: number
  uploaded: number  // 已上传分片数
  total: number     // 总分片数
  progress: number  // 进度百分比 (0-100)
  md5: string[]     // 已上传分片的MD5列表
  is_complete: boolean // 是否已完成
}

/**
 * 查询上传进度
 */
export const getUploadProgress = (precheckId: string) => {
  return get<ApiResponse<UploadProgressResponse>>(API_ENDPOINTS.FILE.PROGRESS, { precheck_id: precheckId })
}

// 上传请求参数
export interface uploadParams {
  precheck_id: string,
  file: File,
  chunk_index: number,
  total_chunks: number,
  chunk_md5: string,
  is_enc: boolean,
}

/**
 * 上传
 */
export const uploadFile = (
  data: uploadParams, 
  onProgress?: (percent: number, loaded?: number, total?: number) => void,
  options?: { onCancel?: (cancel: () => void) => void }
) => {
  const formData = new FormData();
  formData.append('precheck_id', data.precheck_id);
  formData.append('chunk_index', data.chunk_index.toString());
  formData.append('total_chunks', data.total_chunks.toString());
  formData.append('chunk_md5', data.chunk_md5);
  formData.append('is_enc', data.is_enc.toString());
  return upload(API_ENDPOINTS.FILE.UPLOAD, data.file, formData, onProgress, options)
}

// 公开文件列表请求参数
export interface PublicFileListParams {
  type?: string
  sortBy?: string
  page: number
  pageSize: number
}

// 公开文件列表项
export interface PublicFileItem {
  uf_id: string
  file_name: string
  file_size: number
  mime_type: string
  owner_name: string
  has_thumbnail: boolean
  created_at: string
}

// 公开文件列表响应
export interface PublicFileListResponse {
  files: PublicFileItem[]
  total: number
  page: number
  page_size: number
}

/**
 * 获取公开文件列表（文件广场）
 */
export const getPublicFileList = (params: PublicFileListParams) => {
  return get<ApiResponse<PublicFileListResponse>>(API_ENDPOINTS.FILE.PUBLIC_LIST, params)
}

// 未完成的上传任务项
export interface UncompletedUploadTask {
  id: string
  file_name: string
  file_size: number
  chunk_size: number
  total_chunks: number
  uploaded_chunks: number
  progress: number
  status: string
  error_message?: string
  path_id: string
  create_time: string
  update_time: string
  expire_time: string
}

/**
 * 查询未完成的上传任务列表
 */
export const listUncompletedUploads = () => {
  return get<ApiResponse<UncompletedUploadTask[]>>(API_ENDPOINTS.FILE.UNCOMPLETED)
}

/**
 * 删除上传任务请求参数
 */
export interface DeleteUploadTaskRequest {
  task_id: string
}

/**
 * 删除上传任务
 */
export const deleteUploadTask = (taskId: string) => {
  return post<ApiResponse>(API_ENDPOINTS.FILE.DELETE_UPLOAD_TASK, {
    task_id: taskId
  })
}

/**
 * 清理过期的上传任务
 */
export const cleanExpiredUploads = () => {
  return post<ApiResponse<{ cleaned_count: number }>>(API_ENDPOINTS.FILE.CLEAN_EXPIRED)
}


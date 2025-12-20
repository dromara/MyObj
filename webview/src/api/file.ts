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
export const uploadFile = (data: uploadParams) => {
  const formData = new FormData();
  formData.append('precheck_id', data.precheck_id);
  formData.append('chunk_index', data.chunk_index.toString());
  formData.append('total_chunks', data.total_chunks.toString());
  formData.append('chunk_md5', data.chunk_md5);
  formData.append('is_enc', data.is_enc.toString());
  return upload(API_ENDPOINTS.FILE.UPLOAD, data.file, formData)
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


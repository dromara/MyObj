import { post } from '@/utils/request'
import { API_ENDPOINTS } from '@/config/api'
import type { ApiResponse } from '@/types'

// 创建文件夹请求参数
export interface CreateFolderRequest {
  parent_level: string
  dir_path: string
}

/**
 * 创建文件夹
 */
export const createFolder = (data: CreateFolderRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.FOLDER.CREATE, data)
}

/**
 * 删除目录请求参数
 */
export interface DeleteDirRequest {
  dir_id: number
}

/**
 * 删除文件夹（会递归删除目录下的所有文件和子目录）
 */
export const deleteFolder = (data: DeleteDirRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.FOLDER.DELETE, data)
}

/**
 * 目录重命名请求参数
 */
export interface RenameDirRequest {
  dir_id: number
  new_dir_name: string
}

/**
 * 重命名目录
 */
export const renameDir = (data: RenameDirRequest) => {
  return post<ApiResponse>('/file/renameDir', data)
}

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
 * 删除文件夹
 */
export const deleteFolder = (folderId: string) => {
  return post<ApiResponse>(API_ENDPOINTS.FOLDER.DELETE, { folder_id: folderId })
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

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
 * 重命名文件夹
 */
export const renameFolder = (folderId: string, newName: string) => {
  return post<ApiResponse>(API_ENDPOINTS.FOLDER.RENAME, {
    folder_id: folderId,
    new_name: newName
  })
}

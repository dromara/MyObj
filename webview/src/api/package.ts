import { get, post } from '@/utils/network/request'
import { filterParams } from '@/utils/common/params'
import { API_ENDPOINTS, API_BASE_URL } from '@/config/api'
import type { ApiResponse } from '@/types'

// 打包下载相关类型
export interface CreatePackageRequest {
  file_ids: string[]
  package_name?: string
}

export interface CreatePackageResponse {
  package_id: string
  package_name: string
  status: 'creating' | 'ready' | 'failed'
  progress: number
  total_size: number
}

export interface PackageProgressResponse {
  package_id: string
  status: 'creating' | 'ready' | 'failed'
  progress: number
  total_size: number
  created_size: number
  error_msg?: string
}

/**
 * 创建打包下载任务
 */
export const createPackage = (data: CreatePackageRequest) => {
  return post<ApiResponse<CreatePackageResponse>>(API_ENDPOINTS.PACKAGE.CREATE, data)
}

/**
 * 获取打包进度
 */
export const getPackageProgress = (packageId: string) => {
  return get<ApiResponse<PackageProgressResponse>>(
    API_ENDPOINTS.PACKAGE.PROGRESS,
    filterParams({ package_id: packageId })
  )
}

/**
 * 下载打包文件
 */
export const downloadPackage = (packageId: string) => {
  return `${API_BASE_URL}${API_ENDPOINTS.PACKAGE.DOWNLOAD}?package_id=${packageId}`
}

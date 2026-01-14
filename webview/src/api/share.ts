import { post, get } from '@/utils/network/request'
import { API_ENDPOINTS, API_BASE_URL } from '@/config/api'
import type { CreateShareRequest, CreateShareResponse, ApiResponse, ShareInfo } from '@/types'

/**
 * 创建文件分享
 */
export const createShare = (data: CreateShareRequest) => {
  return post<CreateShareResponse>(API_ENDPOINTS.SHARE.CREATE, data)
}

/**
 * 获取我的分享列表
 */
export const getShareList = () => {
  return get<ApiResponse<ShareInfo[]>>(API_ENDPOINTS.SHARE.LIST)
}

/**
 * 删除分享
 */
export const deleteShare = (id: number) => {
  return post<ApiResponse>(API_ENDPOINTS.SHARE.DELETE, { id })
}

/**
 * 修改分享密码
 */
export const updateSharePassword = (id: number, password: string) => {
  return post<ApiResponse>(API_ENDPOINTS.SHARE.UPDATE_PASSWORD, { id, password })
}

/**
 * 获取分享信息（不触发下载）
 * @param token 分享token
 * @param password 分享密码（如果有密码则必需）
 */
export const getShareInfo = (token: string, password?: string) => {
  const params: any = { token }
  if (password) {
    params.password = password
  }
  return get<ApiResponse<ShareInfoResponse>>(API_ENDPOINTS.SHARE.INFO, params)
}

/**
 * 获取分享下载URL（GET请求，直接触发浏览器下载）
 */
export const getShareDownloadUrl = (token: string, password?: string): string => {
  const params = new URLSearchParams({ token })
  if (password) {
    params.append('password', password)
  }
  return `${API_BASE_URL}${API_ENDPOINTS.SHARE.DOWNLOAD}?${params.toString()}`
}

// 分享信息响应类型
export interface ShareInfoResponse {
  file_id: string
  file_name: string
  file_size: number
  mime_type: string
  has_password: boolean
  expires_at: string
  download_count: number
  is_expired: boolean
}

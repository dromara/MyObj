import { post, get } from '@/utils/request'
import { API_ENDPOINTS } from '@/config/api'
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

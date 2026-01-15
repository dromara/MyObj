import { post, get } from '@/utils/network/request'
import { API_ENDPOINTS } from '@/config/api'
import type { ApiResponse } from '@/types'

// 更新用户信息请求参数
export interface UpdateUserRequest {
  nickname?: string
  phone?: string
  email?: string
  username?: string
}

// 修改密码请求参数
export interface UpdatePasswordRequest {
  old_passwd: string
  new_passwd: string
  challenge: string
}

// 生成 API Key 请求参数
export interface GenerateApiKeyRequest {
  expires_days?: number // 过期天数，0或不传表示永不过期
}

// 删除 API Key 请求参数
export interface DeleteApiKeyRequest {
  api_key_id: number
}

// API Key 信息
export interface ApiKeyInfo {
  id: number
  key: string // 已掩码的 key
  expires_at: string | null
  created_at: string
  is_expired: boolean
}

// 生成 API Key 响应
export interface GenerateApiKeyResponse {
  id: number
  key: string // 完整的 key（只在生成时返回一次）
  public_key: string // RSA 公钥（用于加密/解密）
  s3_secret_key: string // S3 Secret Key（用于 S3 服务签名，只在生成时返回一次）
  expires_at: string | null
  created_at: string
}

/**
 * 获取用户信息
 */
export const getUserInfo = () => {
  return get<ApiResponse>(API_ENDPOINTS.USER.INFO)
}

/**
 * 更新用户信息
 */
export const updateUser = (data: UpdateUserRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.USER.UPDATE, data)
}

/**
 * 修改密码
 */
export const updatePassword = (data: UpdatePasswordRequest) => {
  return post<ApiResponse>(API_ENDPOINTS.USER.CHANGE_PASSWORD, data)
}

/**
 * 生成 API Key
 */
export const generateApiKey = (data: GenerateApiKeyRequest) => {
  return post<ApiResponse<GenerateApiKeyResponse>>('/user/apiKey/generate', data)
}

/**
 * 获取 API Key 列表
 */
export const listApiKeys = () => {
  return get<ApiResponse<ApiKeyInfo[]>>('/user/apiKey/list')
}

/**
 * 删除 API Key
 */
export const deleteApiKey = (data: DeleteApiKeyRequest) => {
  return post<ApiResponse>('/user/apiKey/delete', data)
}

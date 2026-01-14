// 认证相关API
import { post, get } from '@/utils/network/request'
import { API_ENDPOINTS } from '@/config/api'
import type {
  LoginRequest,
  RegisterRequest,
  LoginResponse,
  ChallengeResponse,
  UpdatePasswordRequest,
  SetFilePasswordRequest,
  ApiResponse
} from '@/types'

/**
 * 获取密码挑战秘钥
 */
export const getChallenge = (): Promise<ChallengeResponse> => {
  return get(API_ENDPOINTS.AUTH.CHALLENGE)
}

/**
 * 用户登录
 */
export const login = (data: LoginRequest): Promise<LoginResponse> => {
  return post(API_ENDPOINTS.AUTH.LOGIN, data)
}

/**
 * 用户注册
 */
export const register = (data: RegisterRequest): Promise<ApiResponse> => {
  return post(API_ENDPOINTS.AUTH.REGISTER, data)
}

/**
 * 用户登出
 */
export const logout = (): Promise<void> => {
  return post(API_ENDPOINTS.AUTH.LOGOUT)
}

/**
 * 刷新token
 */
export const refreshToken = (): Promise<{ token: string }> => {
  return post(API_ENDPOINTS.AUTH.REFRESH)
}

/**
 * 修改密码
 */
export const updatePassword = (data: UpdatePasswordRequest): Promise<ApiResponse> => {
  return post(API_ENDPOINTS.USER.CHANGE_PASSWORD, data)
}

/**
 * 设置文件密码
 */
export const setFilePassword = (data: SetFilePasswordRequest): Promise<ApiResponse> => {
  return post(API_ENDPOINTS.USER.SET_FILE_PASSWORD, data)
}

/**
 * 修改文件密码
 */
export const updateFilePassword = (data: UpdatePasswordRequest): Promise<ApiResponse> => {
  return post(API_ENDPOINTS.USER.UPDATE_FILE_PASSWORD, data)
}

/**
 * 获取系统信息（判断是否首次使用和注册配置）
 */
export const getSysInfo = (): Promise<ApiResponse<{ is_first_use: boolean; allow_register: boolean }>> => {
  return get(API_ENDPOINTS.USER.SYS_INFO)
}

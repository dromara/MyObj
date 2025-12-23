/**
 * 分享相关工具函数
 */

/**
 * 获取分享链接
 * @param token 分享token
 * @returns 完整的分享链接
 */
import { API_BASE_URL } from '@/config/api'

export const getShareUrl = (token: string): string => {
  return `${window.location.origin}${API_BASE_URL}/share/download?token=${token}`
}


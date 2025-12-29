import { post } from '@/utils/request'
import type { ApiResponse } from '@/types'
import { API_BASE_URL } from '@/config/api'

/**
 * 视频播放 Token 响应
 */
export interface VideoPlayTokenResponse {
  play_token: string
  file_info: {
    file_id: string
    file_name: string
    file_size: number
    is_enc: boolean
    mime_type: string
  }
}

/**
 * 创建视频播放预检（获取播放 Token）
 * @param fileId 文件ID
 * @param sharePassword 文件密码（加密文件必需）
 */
export const createVideoPlayPrecheck = (fileId: string, sharePassword?: string) => {
  return post<ApiResponse<VideoPlayTokenResponse>>('/video/play/precheck', {
    file_id: fileId,
    share_password: sharePassword || ''
  })
}

/**
 * 获取视频流 URL（支持 Range 请求，每次最大 2MB）
 * @param token 播放 Token
 * @param jwtToken JWT Token（可选，如果提供会添加到 URL 参数中）
 * @returns 视频流 URL
 */
export const getVideoStreamUrl = (token: string, jwtToken?: string): string => {
  let url = `${API_BASE_URL}/video/stream?token=${token}`
  if (jwtToken) {
    url += `&jwt=${encodeURIComponent(jwtToken)}`
  }
  return url
}


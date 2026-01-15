/**
 * 分享相关工具函数
 */

/**
 * 获取分享链接（新的分享页面）
 * @param token 分享token
 * @returns 完整的分享链接
 */
export const getShareUrl = (token: string): string => {
  return `${window.location.origin}/share/${token}`
}

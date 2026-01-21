/**
 * 链接类型检测工具
 * 用于识别剪贴板中的下载链接类型
 */

export type LinkType = 'http' | 'magnet' | 'torrent' | 'unknown'

export interface DetectedLink {
  type: LinkType
  content: string
  url?: string // HTTP/HTTPS 链接的 URL
  magnet?: string // 磁力链接
  torrentBase64?: string // 种子文件的 Base64 编码
}

/**
 * HTTP/HTTPS 链接正则表达式
 */
const HTTP_URL_REGEX = /^https?:\/\/.+$/i

/**
 * 磁力链接正则表达式
 */
const MAGNET_REGEX = /^magnet:\?xt=urn:btih:.+$/i

/**
 * 检测文本是否为 HTTP/HTTPS 链接
 */
export function isHttpUrl(text: string): boolean {
  return HTTP_URL_REGEX.test(text.trim())
}

/**
 * 检测文本是否为磁力链接
 */
export function isMagnetLink(text: string): boolean {
  return MAGNET_REGEX.test(text.trim())
}

/**
 * 检测文本是否为种子文件内容（Base64 编码）
 * 注意：这是一个简单的检测，实际应该通过文件扩展名或 MIME 类型判断
 */
export function isTorrentBase64(text: string): boolean {
  // Base64 编码的种子文件通常以 d8:announce 开头（Bencode 格式）
  try {
    const decoded = atob(text.trim())
    // 检查是否包含 Bencode 特征
    return decoded.includes('announce') && decoded.includes('info')
  } catch {
    return false
  }
}

/**
 * 检测链接类型
 * @param content 剪贴板内容（文本或 Base64 编码的种子文件）
 * @returns 检测到的链接信息，如果无法识别则返回 null
 */
export function detectLinkType(content: string): DetectedLink | null {
  if (!content || typeof content !== 'string') {
    return null
  }

  const trimmed = content.trim()

  // 检测磁力链接
  if (isMagnetLink(trimmed)) {
    return {
      type: 'magnet',
      content: trimmed,
      magnet: trimmed
    }
  }

  // 检测 HTTP/HTTPS 链接
  if (isHttpUrl(trimmed)) {
    return {
      type: 'http',
      content: trimmed,
      url: trimmed
    }
  }

  // 检测种子文件（Base64 编码）
  if (isTorrentBase64(trimmed)) {
    return {
      type: 'torrent',
      content: trimmed,
      torrentBase64: trimmed
    }
  }

  return null
}

/**
 * 格式化链接显示名称
 */
export function formatLinkDisplayName(link: DetectedLink): string {
  switch (link.type) {
    case 'http':
      try {
        const url = new URL(link.url || '')
        // 显示完整 URL，但限制长度
        const fullUrl = url.href
        if (fullUrl.length > 60) {
          return fullUrl.substring(0, 57) + '...'
        }
        return fullUrl
      } catch {
        // 如果 URL 解析失败，显示原始内容（限制长度）
        const content = link.url || link.content || 'HTTP 链接'
        if (content.length > 60) {
          return content.substring(0, 57) + '...'
        }
        return content
      }
    case 'magnet':
      // 尝试从磁力链接中提取文件名
      const nameMatch = link.magnet?.match(/dn=([^&]+)/)
      if (nameMatch) {
        try {
          const fileName = decodeURIComponent(nameMatch[1])
          // 如果文件名太长，截断
          if (fileName.length > 50) {
            return fileName.substring(0, 47) + '...'
          }
          return fileName
        } catch {
          // 如果解码失败，显示简化的磁力链接
          const hashMatch = link.magnet?.match(/btih:([^&]+)/)
          if (hashMatch) {
            return `磁力链接 (${hashMatch[1].substring(0, 8)}...)`
          }
          return '磁力链接'
        }
      }
      // 如果没有文件名，显示 hash
      const hashMatch = link.magnet?.match(/btih:([^&]+)/)
      if (hashMatch) {
        return `磁力链接 (${hashMatch[1].substring(0, 8)}...)`
      }
      return '磁力链接'
    case 'torrent':
      return '种子文件'
    default:
      return '未知链接'
  }
}

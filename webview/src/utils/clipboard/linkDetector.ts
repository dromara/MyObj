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
  fileName?: string // 从 URL 中提取的文件名
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
 * 可下载文件扩展名列表
 * 参考: https://support.microsoft.com/en-us/windows/common-file-name-extensions-in-windows
 */
const DOWNLOADABLE_EXTENSIONS: Set<string> = new Set([
  // 视频格式
  'mp4',
  'mkv',
  'avi',
  'mov',
  'wmv',
  'flv',
  'webm',
  'mpeg',
  'mpg',
  'vob',
  'rmvb',
  'rm',
  'm4v',
  '3gp',
  'ts',
  // 音频格式
  'mp3',
  'wav',
  'aac',
  'wma',
  'flac',
  'ogg',
  'm4a',
  'ape',
  'aiff',
  'opus',
  // 压缩包/归档格式
  'zip',
  'rar',
  '7z',
  'tar',
  'gz',
  'bz2',
  'xz',
  'lz',
  'lzma',
  // 磁盘镜像/安装包
  'iso',
  'img',
  'dmg',
  'exe',
  'msi',
  'deb',
  'rpm',
  'apk',
  'ipa',
  'appimage',
  // 文档格式
  'pdf',
  'doc',
  'docx',
  'xls',
  'xlsx',
  'ppt',
  'pptx',
  'odt',
  'ods',
  'odp',
  'rtf',
  'epub',
  'mobi',
  // 种子文件
  'torrent',
  // 其他二进制/数据文件
  'bin',
  'dat',
  'pkg',
  'bundle'
])

/**
 * 网页/脚本扩展名（不应被识别为下载链接）
 */
const WEBPAGE_EXTENSIONS: Set<string> = new Set([
  'html',
  'htm',
  'php',
  'asp',
  'aspx',
  'jsp',
  'jspx',
  'cgi',
  'pl',
  'py',
  'rb',
  'shtml',
  'xhtml'
])

/**
 * 从 URL 中提取文件名和扩展名
 * @param url URL 字符串
 * @returns { fileName: string | null, extension: string | null }
 */
export function extractFileInfo(
  url: string
): { fileName: string | null; extension: string | null } {
  try {
    const urlObj = new URL(url)
    const pathname = urlObj.pathname

    // 获取路径的最后一部分
    const segments = pathname.split('/').filter(Boolean)
    if (segments.length === 0) {
      return { fileName: null, extension: null }
    }

    const lastSegment = decodeURIComponent(segments[segments.length - 1])

    // 检查是否有文件扩展名
    const dotIndex = lastSegment.lastIndexOf('.')
    if (dotIndex === -1 || dotIndex === 0 || dotIndex === lastSegment.length - 1) {
      return { fileName: null, extension: null }
    }

    const fileName = lastSegment
    const extension = lastSegment.substring(dotIndex + 1).toLowerCase()

    return { fileName, extension }
  } catch {
    return { fileName: null, extension: null }
  }
}

/**
 * 检测文本是否为 HTTP/HTTPS 链接
 */
export function isHttpUrl(text: string): boolean {
  return HTTP_URL_REGEX.test(text.trim())
}

/**
 * 检测 URL 是否为可下载文件链接
 * @param url HTTP/HTTPS URL
 * @returns true 如果 URL 指向可下载文件
 */
export function isDownloadableUrl(url: string): boolean {
  const { extension } = extractFileInfo(url)

  if (!extension) {
    return false
  }

  // 排除网页扩展名
  if (WEBPAGE_EXTENSIONS.has(extension)) {
    return false
  }

  // 检查是否为已知的可下载扩展名
  return DOWNLOADABLE_EXTENSIONS.has(extension)
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

  // 检测磁力链接（磁力链接始终是下载链接）
  if (isMagnetLink(trimmed)) {
    return {
      type: 'magnet',
      content: trimmed,
      magnet: trimmed
    }
  }

  // 检测 HTTP/HTTPS 链接
  // 注意：只有包含可下载文件扩展名的链接才会被识别为下载链接
  // 普通网页链接（如 https://www.baidu.com/）不会被识别
  if (isHttpUrl(trimmed) && isDownloadableUrl(trimmed)) {
    const { fileName } = extractFileInfo(trimmed)
    return {
      type: 'http',
      content: trimmed,
      url: trimmed,
      fileName: fileName || undefined
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
      // 优先显示文件名
      if (link.fileName) {
        if (link.fileName.length > 50) {
          return link.fileName.substring(0, 47) + '...'
        }
        return link.fileName
      }
      // 备选：显示 URL
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

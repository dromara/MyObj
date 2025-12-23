/**
 * 文件预览工具函数
 */

import type { FileItem } from '@/types'
import type { PreviewType } from '@/types/preview'
import { API_BASE_URL } from '@/config/api'
import { API_ENDPOINTS } from '@/config/api'

/**
 * 检测文件类型
 */
export const detectFileType = (file: FileItem): PreviewType => {
  if (!file.mime_type) {
    return 'unsupported'
  }

  const mimeType = file.mime_type.toLowerCase()

  // 图片类型
  if (mimeType.startsWith('image/')) {
    return 'image'
  }

  // 视频类型
  if (mimeType.startsWith('video/')) {
    return 'video'
  }

  // 音频类型
  if (mimeType.startsWith('audio/')) {
    return 'audio'
  }

  // PDF 类型
  if (mimeType === 'application/pdf') {
    return 'pdf'
  }

  // 文本/代码类型
  if (mimeType.startsWith('text/')) {
    // 代码文件扩展名
    const codeExts = [
      'js', 'ts', 'jsx', 'tsx', 'vue', 'html', 'css', 'scss', 'less',
      'java', 'py', 'cpp', 'c', 'go', 'rs', 'php', 'rb', 'swift',
      'sql', 'sh', 'bash', 'zsh', 'ps1'
    ]
    
    // 从文件名获取扩展名
    const fileName = file.file_name || ''
    const ext = fileName.split('.').pop()?.toLowerCase() || ''
    
    if (codeExts.includes(ext)) {
      return 'code'
    }
    
    return 'text'
  }

  // JSON、XML 等
  if (mimeType === 'application/json' || mimeType === 'application/xml') {
    return 'text'
  }

  return 'unsupported'
}

/**
 * 获取文件预览 URL（返回blob URL，带认证）
 * @param fileId 文件ID
 * @returns Promise<string> blob URL
 */
export const getFilePreviewUrl = async (fileId: string): Promise<string> => {
  try {
    const token = localStorage.getItem('token')
    const url = `${API_BASE_URL}/download/preview?file_id=${fileId}`
    
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Authorization': token ? `Bearer ${token}` : ''
      }
    })
    
    if (!response.ok) {
      throw new Error('获取文件失败: ' + response.status)
    }
    
    const blob = await response.blob()
    return window.URL.createObjectURL(blob)
  } catch (error) {
    throw new Error('获取文件预览失败: ' + (error instanceof Error ? error.message : '未知错误'))
  }
}

/**
 * 获取文件下载 URL
 * @param fileId 文件ID
 * @returns 下载URL
 */
export const getFileDownloadUrl = (fileId: string): string => {
  // 使用预览接口（预览和下载使用同一个接口）
  return `${API_BASE_URL}/download/preview?file_id=${fileId}`
}

/**
 * 获取文件缩略图 URL
 * @param fileId 文件ID
 * @returns 缩略图URL
 */
export const getThumbnailUrl = (fileId: string): string => {
  return `${API_BASE_URL}${API_ENDPOINTS.FILE.THUMBNAIL}/${fileId}`
}

/**
 * 获取文件文本内容
 * @param fileId 文件ID
 * @returns 文本内容
 */
export const getFileTextContent = async (fileId: string): Promise<string> => {
  try {
    const url = getFileDownloadUrl(fileId)
    const token = localStorage.getItem('token')
    
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Authorization': token ? `Bearer ${token}` : ''
      }
    })
    
    if (!response.ok) {
      throw new Error('获取文件内容失败')
    }
    
    return await response.text()
  } catch (error) {
    throw new Error('获取文件内容失败')
  }
}

/**
 * 获取代码语言类型（用于代码高亮）
 */
export const getCodeLanguage = (fileName: string): string => {
  const ext = fileName.split('.').pop()?.toLowerCase() || ''
  const langMap: Record<string, string> = {
    'js': 'javascript',
    'ts': 'typescript',
    'jsx': 'javascript',
    'tsx': 'typescript',
    'vue': 'vue',
    'html': 'html',
    'css': 'css',
    'scss': 'scss',
    'less': 'less',
    'json': 'json',
    'xml': 'xml',
    'yaml': 'yaml',
    'yml': 'yaml',
    'py': 'python',
    'java': 'java',
    'cpp': 'cpp',
    'c': 'c',
    'go': 'go',
    'rs': 'rust',
    'php': 'php',
    'rb': 'ruby',
    'swift': 'swift',
    'sql': 'sql',
    'sh': 'bash',
    'bash': 'bash',
    'zsh': 'bash',
    'ps1': 'powershell'
  }
  return langMap[ext] || ext
}


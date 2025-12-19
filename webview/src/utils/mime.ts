/**
 * MIME 类型工具函数
 */

/**
 * 文件类型分类（用于 Square 等页面）
 */
export type FileTypeCategory = 'image' | 'video' | 'audio' | 'doc' | 'archive' | 'other'

/**
 * 根据文件类型分类获取 MIME 类型
 * @param type 文件类型分类（如 'image', 'video', 'audio' 等）
 * @returns MIME 类型字符串
 */
export const getMimeTypeFromFileType = (type: string): string => {
  const mimeMap: Record<string, string> = {
    'image': 'image/jpeg',
    'video': 'video/mp4',
    'audio': 'audio/mpeg',
    'doc': 'application/pdf',
    'archive': 'application/zip',
    'other': 'application/octet-stream'
  }
  return mimeMap[type] || 'application/octet-stream'
}

/**
 * 根据文件扩展名获取 MIME 类型
 * @param extension 文件扩展名（不含点号，如 'jpg', 'pdf'）
 * @returns MIME 类型字符串
 */
export const getMimeTypeFromExtension = (extension: string): string => {
  const ext = extension.toLowerCase().replace(/^\./, '')
  
  const mimeMap: Record<string, string> = {
    // 图片
    'jpg': 'image/jpeg',
    'jpeg': 'image/jpeg',
    'png': 'image/png',
    'gif': 'image/gif',
    'webp': 'image/webp',
    'svg': 'image/svg+xml',
    'bmp': 'image/bmp',
    'ico': 'image/x-icon',
    
    // 视频
    'mp4': 'video/mp4',
    'avi': 'video/x-msvideo',
    'mov': 'video/quicktime',
    'wmv': 'video/x-ms-wmv',
    'flv': 'video/x-flv',
    'mkv': 'video/x-matroska',
    'webm': 'video/webm',
    
    // 音频
    'mp3': 'audio/mpeg',
    'wav': 'audio/wav',
    'flac': 'audio/flac',
    'aac': 'audio/aac',
    'ogg': 'audio/ogg',
    'm4a': 'audio/mp4',
    
    // 文档
    'pdf': 'application/pdf',
    'doc': 'application/msword',
    'docx': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    'xls': 'application/vnd.ms-excel',
    'xlsx': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    'ppt': 'application/vnd.ms-powerpoint',
    'pptx': 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
    
    // 文本
    'txt': 'text/plain',
    'html': 'text/html',
    'css': 'text/css',
    'js': 'text/javascript',
    'json': 'application/json',
    'xml': 'application/xml',
    'md': 'text/markdown',
    
    // 压缩
    'zip': 'application/zip',
    'rar': 'application/x-rar-compressed',
    '7z': 'application/x-7z-compressed',
    'tar': 'application/x-tar',
    'gz': 'application/gzip'
  }
  
  return mimeMap[ext] || 'application/octet-stream'
}

/**
 * 根据文件名获取 MIME 类型
 * @param fileName 文件名（如 'example.jpg'）
 * @returns MIME 类型字符串
 */
export const getMimeTypeFromFileName = (fileName: string): string => {
  const extension = fileName.split('.').pop() || ''
  return getMimeTypeFromExtension(extension)
}

/**
 * 根据 MIME 类型获取文件类型分类
 * @param mimeType MIME 类型
 * @returns 文件类型分类
 */
export const getFileTypeFromMimeType = (mimeType: string): FileTypeCategory => {
  const mime = mimeType.toLowerCase()
  
  if (mime.startsWith('image/')) return 'image'
  if (mime.startsWith('video/')) return 'video'
  if (mime.startsWith('audio/')) return 'audio'
  if (mime === 'application/pdf' || 
      mime.includes('word') || 
      mime.includes('excel') || 
      mime.includes('powerpoint') ||
      mime.includes('document')) return 'doc'
  if (mime.includes('zip') || 
      mime.includes('rar') || 
      mime.includes('7z') || 
      mime.includes('tar') || 
      mime.includes('gzip')) return 'archive'
  
  return 'other'
}


// 文件图标映射工具 - 基于 MIME 类型
import {
  Document,
  Picture,
  VideoPlay,
  Headset,
  Reading,
  DocumentCopy,
  PictureFilled,
  Film,
  Paperclip,
  FolderOpened,
  Postcard,
  Notebook,
  EditPen,
  Grid,
  TakeawayBox,
  CollectionTag,
  MessageBox,
  Calendar,
  PhoneFilled,
  Platform,
  Monitor,
  Connection
} from '@element-plus/icons-vue'
import type { Component } from 'vue'

// 文件图标配置接口
export interface FileIconConfig {
  icon: Component
  color: string
  name: string
}

// MIME 类型到图标的映射
const mimeIconMap: Record<string, FileIconConfig> = {
  // 图片类型 - 紫色系
  'image/jpeg': { icon: Picture, color: '#9C27B0', name: 'JPEG图片' },
  'image/jpg': { icon: Picture, color: '#9C27B0', name: 'JPG图片' },
  'image/png': { icon: PictureFilled, color: '#AB47BC', name: 'PNG图片' },
  'image/gif': { icon: PictureFilled, color: '#BA68C8', name: 'GIF动图' },
  'image/webp': { icon: Picture, color: '#CE93D8', name: 'WebP图片' },
  'image/svg+xml': { icon: Grid, color: '#E1BEE7', name: 'SVG矢量图' },
  'image/bmp': { icon: Picture, color: '#8E24AA', name: 'BMP图片' },
  'image/tiff': { icon: Picture, color: '#7B1FA2', name: 'TIFF图片' },
  'image/x-icon': { icon: Postcard, color: '#6A1B9A', name: '图标文件' },

  // 视频类型 - 红色系
  'video/mp4': { icon: VideoPlay, color: '#F44336', name: 'MP4视频' },
  'video/mpeg': { icon: Film, color: '#E53935', name: 'MPEG视频' },
  'video/quicktime': { icon: VideoPlay, color: '#D32F2F', name: 'QuickTime视频' },
  'video/x-msvideo': { icon: Film, color: '#C62828', name: 'AVI视频' },
  'video/x-flv': { icon: VideoPlay, color: '#B71C1C', name: 'FLV视频' },
  'video/x-matroska': { icon: Film, color: '#FF5252', name: 'MKV视频' },
  'video/webm': { icon: VideoPlay, color: '#FF1744', name: 'WebM视频' },

  // 音频类型 - 粉色系
  'audio/mpeg': { icon: Headset, color: '#E91E63', name: 'MP3音频' },
  'audio/wav': { icon: Headset, color: '#D81B60', name: 'WAV音频' },
  'audio/x-wav': { icon: Headset, color: '#C2185B', name: 'WAV音频' },
  'audio/ogg': { icon: Headset, color: '#AD1457', name: 'OGG音频' },
  'audio/flac': { icon: Headset, color: '#880E4F', name: 'FLAC音频' },
  'audio/aac': { icon: Headset, color: '#FF4081', name: 'AAC音频' },
  'audio/mp4': { icon: Headset, color: '#F50057', name: 'M4A音频' },

  // PDF文档 - 深红色
  'application/pdf': { icon: Reading, color: '#D32F2F', name: 'PDF文档' },

  // Word文档 - 蓝色系
  'application/msword': { icon: Document, color: '#2196F3', name: 'Word文档' },
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document': { icon: Document, color: '#1976D2', name: 'Word文档' },
  'application/vnd.oasis.opendocument.text': { icon: Document, color: '#1565C0', name: 'ODT文档' },

  // Excel表格 - 绿色系
  'application/vnd.ms-excel': { icon: Grid, color: '#4CAF50', name: 'Excel表格' },
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': { icon: Grid, color: '#388E3C', name: 'Excel表格' },
  'application/vnd.oasis.opendocument.spreadsheet': { icon: Grid, color: '#2E7D32', name: 'ODS表格' },
  'text/csv': { icon: Grid, color: '#66BB6A', name: 'CSV表格' },

  // PowerPoint演示 - 橙色系
  'application/vnd.ms-powerpoint': { icon: Monitor, color: '#FF9800', name: 'PPT演示' },
  'application/vnd.openxmlformats-officedocument.presentationml.presentation': { icon: Monitor, color: '#F57C00', name: 'PPT演示' },
  'application/vnd.oasis.opendocument.presentation': { icon: Monitor, color: '#EF6C00', name: 'ODP演示' },

  // 文本文件 - 灰色系
  'text/plain': { icon: DocumentCopy, color: '#607D8B', name: '文本文件' },
  'text/html': { icon: Platform, color: '#546E7A', name: 'HTML文件' },
  'text/css': { icon: EditPen, color: '#455A64', name: 'CSS样式' },
  'text/javascript': { icon: Platform, color: '#37474F', name: 'JavaScript' },
  'application/javascript': { icon: Platform, color: '#263238', name: 'JavaScript' },
  'application/json': { icon: DocumentCopy, color: '#78909C', name: 'JSON数据' },
  'application/xml': { icon: DocumentCopy, color: '#90A4AE', name: 'XML文件' },
  'text/xml': { icon: DocumentCopy, color: '#B0BEC5', name: 'XML文件' },

  // 代码文件 - 深灰色系
  'text/x-python': { icon: Platform, color: '#3776AB', name: 'Python代码' },
  'text/x-java': { icon: Platform, color: '#007396', name: 'Java代码' },
  'text/x-c': { icon: Platform, color: '#A8B9CC', name: 'C代码' },
  'text/x-c++': { icon: Platform, color: '#00599C', name: 'C++代码' },
  'text/x-go': { icon: Platform, color: '#00ADD8', name: 'Go代码' },
  'text/x-rust': { icon: Platform, color: '#CE422B', name: 'Rust代码' },

  // 压缩文件 - 黄色系
  'application/zip': { icon: TakeawayBox, color: '#FFC107', name: 'ZIP压缩包' },
  'application/x-zip-compressed': { icon: TakeawayBox, color: '#FFB300', name: 'ZIP压缩包' },
  'application/x-rar-compressed': { icon: TakeawayBox, color: '#FFA000', name: 'RAR压缩包' },
  'application/x-7z-compressed': { icon: TakeawayBox, color: '#FF8F00', name: '7Z压缩包' },
  'application/x-tar': { icon: TakeawayBox, color: '#FF6F00', name: 'TAR归档' },
  'application/gzip': { icon: TakeawayBox, color: '#FFCA28', name: 'GZIP压缩' },
  'application/x-bzip2': { icon: TakeawayBox, color: '#FFD54F', name: 'BZ2压缩' },

  // 可执行文件 - 青色系
  'application/x-msdownload': { icon: Platform, color: '#00BCD4', name: 'EXE程序' },
  'application/x-executable': { icon: Platform, color: '#00ACC1', name: '可执行文件' },
  'application/x-mach-binary': { icon: Platform, color: '#0097A7', name: '二进制文件' },
  'application/x-apple-diskimage': { icon: Platform, color: '#00838F', name: 'DMG镜像' },

  // 数据库文件 - 深蓝色
  'application/x-sqlite3': { icon: Connection, color: '#1565C0', name: 'SQLite数据库' },
  'application/sql': { icon: Connection, color: '#0D47A1', name: 'SQL文件' },

  // 字体文件 - 棕色系
  'font/ttf': { icon: EditPen, color: '#795548', name: 'TTF字体' },
  'font/otf': { icon: EditPen, color: '#6D4C41', name: 'OTF字体' },
  'font/woff': { icon: EditPen, color: '#5D4037', name: 'WOFF字体' },
  'font/woff2': { icon: EditPen, color: '#4E342E', name: 'WOFF2字体' },

  // 电子书 - 深紫色
  'application/epub+zip': { icon: Reading, color: '#673AB7', name: 'EPUB电子书' },
  'application/x-mobipocket-ebook': { icon: Reading, color: '#5E35B1', name: 'MOBI电子书' },

  // Markdown - 靛蓝色
  'text/markdown': { icon: Notebook, color: '#3F51B5', name: 'Markdown文档' },

  // 邮件 - 青绿色
  'message/rfc822': { icon: MessageBox, color: '#009688', name: '邮件文件' },

  // 日历 - 深绿色
  'text/calendar': { icon: Calendar, color: '#388E3C', name: '日历文件' },

  // 联系人 - 深青色
  'text/vcard': { icon: PhoneFilled, color: '#00796B', name: '联系人卡片' },

  // 种子文件 - 深绿色
  'application/x-bittorrent': { icon: CollectionTag, color: '#2E7D32', name: 'BT种子' },
}

// 默认图标（未匹配到的文件类型）
const defaultIcon: FileIconConfig = {
  icon: Paperclip,
  color: '#9E9E9E',
  name: '未知文件'
}

// 文件夹图标
export const folderIcon: FileIconConfig = {
  icon: FolderOpened,
  color: '#FFA726',
  name: '文件夹'
}

/**
 * 根据 MIME 类型获取文件图标配置
 * @param mimeType MIME 类型（从后端 mimetype.DetectFile 获取）
 * @returns 文件图标配置
 */
export const getFileIcon = (mimeType: string): FileIconConfig => {
  // 精确匹配
  if (mimeIconMap[mimeType]) {
    return mimeIconMap[mimeType]
  }

  // 模糊匹配（主类型）
  const mainType = mimeType.split('/')[0]
  switch (mainType) {
    case 'image':
      return { icon: Picture, color: '#9C27B0', name: '图片' }
    case 'video':
      return { icon: VideoPlay, color: '#F44336', name: '视频' }
    case 'audio':
      return { icon: Headset, color: '#E91E63', name: '音频' }
    case 'text':
      return { icon: DocumentCopy, color: '#607D8B', name: '文本' }
    default:
      return defaultIcon
  }
}

/**
 * 根据 MIME 类型获取文件分类
 * @param mimeType MIME 类型
 * @returns 文件分类（用于过滤和分组）
 */
export const getFileCategory = (mimeType: string): string => {
  const mainType = mimeType.split('/')[0]
  
  if (mainType === 'image') return 'image'
  if (mainType === 'video') return 'video'
  if (mainType === 'audio') return 'audio'
  if (mimeType.includes('pdf')) return 'document'
  if (mimeType.includes('word') || mimeType.includes('document')) return 'document'
  if (mimeType.includes('excel') || mimeType.includes('spreadsheet') || mimeType.includes('csv')) return 'spreadsheet'
  if (mimeType.includes('powerpoint') || mimeType.includes('presentation')) return 'presentation'
  if (mimeType.includes('zip') || mimeType.includes('rar') || mimeType.includes('7z') || mimeType.includes('tar') || mimeType.includes('gzip')) return 'archive'
  if (mainType === 'text') return 'text'
  
  return 'other'
}

/**
 * 获取所有支持的 MIME 类型列表
 */
export const getSupportedMimeTypes = (): string[] => {
  return Object.keys(mimeIconMap)
}

/**
 * 检查是否为图片类型
 */
export const isImageType = (mimeType: string): boolean => {
  return mimeType.startsWith('image/')
}

/**
 * 检查是否为视频类型
 */
export const isVideoType = (mimeType: string): boolean => {
  return mimeType.startsWith('video/')
}

/**
 * 检查是否为音频类型
 */
export const isAudioType = (mimeType: string): boolean => {
  return mimeType.startsWith('audio/')
}

/**
 * 检查是否为文档类型
 */
export const isDocumentType = (mimeType: string): boolean => {
  return mimeType.includes('pdf') || 
         mimeType.includes('word') || 
         mimeType.includes('excel') || 
         mimeType.includes('powerpoint') ||
         mimeType.includes('document') ||
         mimeType.includes('spreadsheet') ||
         mimeType.includes('presentation')
}

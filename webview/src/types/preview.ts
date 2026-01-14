/**
 * 文件预览相关类型定义
 */

import type { FileItem } from './index'

/**
 * 文件预览类型
 */
export type PreviewType =
  | 'image' // 图片
  | 'video' // 视频
  | 'audio' // 音频
  | 'pdf' // PDF
  | 'text' // 文本
  | 'code' // 代码
  | 'unsupported' // 不支持预览

/**
 * 预览选项
 */
export interface PreviewOptions {
  /** 是否自动播放（视频/音频） */
  autoplay?: boolean
  /** 是否循环播放（视频/音频） */
  loop?: boolean
  /** 是否显示控制器（视频/音频） */
  controls?: boolean
  /** 图片缩放比例 */
  zoom?: number
  /** 图片旋转角度 */
  rotate?: number
}

/**
 * 预览状态
 */
export interface PreviewState {
  /** 当前预览的文件 */
  currentFile: FileItem | null
  /** 预览类型 */
  previewType: PreviewType
  /** 是否显示预览 */
  visible: boolean
  /** 预览选项 */
  options: PreviewOptions
  /** 加载状态 */
  loading: boolean
  /** 错误信息 */
  error?: string
}

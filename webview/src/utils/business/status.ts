/**
 * 状态相关工具函数
 */

/**
 * 获取任务状态类型（用于 Element Plus Tag 组件）
 * @param state 状态码
 * @returns Element Plus Tag 类型
 */
export const getTaskStatusType = (state: number): 'info' | 'primary' | 'warning' | 'success' | 'danger' => {
  const typeMap: Record<number, 'info' | 'primary' | 'warning' | 'success' | 'danger'> = {
    0: 'info', // 初始化
    1: 'primary', // 进行中（下载中/上传中）
    2: 'warning', // 已暂停
    3: 'success', // 已完成
    4: 'danger' // 失败
  }
  return typeMap[state] || 'info'
}

/**
 * 获取上传任务状态类型
 * @param status 状态字符串
 * @returns Element Plus Tag 类型
 */
export const getUploadStatusType = (status: string): 'info' | 'primary' | 'warning' | 'success' | 'danger' => {
  const typeMap: Record<string, 'info' | 'primary' | 'warning' | 'success' | 'danger'> = {
    prechecking: 'info', // 预检中
    pending: 'info',
    uploading: 'primary',
    paused: 'warning',
    completed: 'success',
    failed: 'danger',
    cancelled: 'info'
  }
  return typeMap[status] || 'info'
}

/**
 * 获取上传任务状态文本
 * @param status 状态字符串
 * @param t 国际化翻译函数（可选，如果不提供则返回状态 key 本身）
 * @returns 状态文本
 */
export const getUploadStatusText = (status: string, t?: (key: string) => string): string => {
  const keyMap: Record<string, string> = {
    prechecking: 'tasks.prechecking',
    pending: 'tasks.pending',
    uploading: 'tasks.running',
    paused: 'tasks.paused',
    completed: 'tasks.completed',
    failed: 'tasks.failed',
    cancelled: 'tasks.cancelSuccess'
  }
  const i18nKey = keyMap[status]
  if (i18nKey && t) {
    return t(i18nKey)
  }
  return i18nKey || status
}

/**
 * 获取下载任务状态类型
 * @param state 状态码
 * @returns Element Plus Tag 类型
 */
export const getDownloadStatusType = (state: number): 'info' | 'primary' | 'warning' | 'success' | 'danger' => {
  return getTaskStatusType(state)
}

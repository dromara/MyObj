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
    0: 'info',      // 初始化
    1: 'primary',   // 进行中（下载中/上传中）
    2: 'warning',   // 已暂停
    3: 'success',   // 已完成
    4: 'danger'     // 失败
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
    'pending': 'info',
    'uploading': 'primary',
    'paused': 'warning',
    'completed': 'success',
    'failed': 'danger',
    'cancelled': 'info'
  }
  return typeMap[status] || 'info'
}

/**
 * 获取上传任务状态文本
 * @param status 状态字符串
 * @returns 状态文本
 */
// 获取上传阶段文本
export const getUploadStageText = (stage?: string): string => {
  if (!stage) return '准备中'
  const stageMap: Record<string, string> = {
    'reading': '读取文件',
    'calculating': '计算MD5',
    'uploading': '上传中',
    'completed': '已完成',
    'failed': '失败'
  }
  return stageMap[stage] || '准备中'
}

export const getUploadStatusText = (status: string): string => {
  const textMap: Record<string, string> = {
    'pending': '等待中',
    'uploading': '上传中',
    'paused': '已暂停',
    'completed': '已完成',
    'failed': '失败',
    'cancelled': '已取消'
  }
  return textMap[status] || '未知'
}

/**
 * 获取下载任务状态类型
 * @param state 状态码
 * @returns Element Plus Tag 类型
 */
export const getDownloadStatusType = (state: number): 'info' | 'primary' | 'warning' | 'success' | 'danger' => {
  return getTaskStatusType(state)
}


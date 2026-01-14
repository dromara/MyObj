import i18n from '@/i18n'

/**
 * 格式化工具函数
 */

/**
 * 格式化文件大小
 * @param bytes 字节数
 * @returns 格式化后的文件大小字符串
 */
export const formatSize = (bytes: number): string => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

/**
 * 格式化日期时间
 * @param dateStr 日期字符串
 * @param options 格式化选项
 * @returns 格式化后的日期字符串
 */
export const formatDate = (
  dateStr: string,
  options?: {
    showTime?: boolean
    showSeconds?: boolean
  }
): string => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  const formatOptions: Intl.DateTimeFormatOptions = {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  }

  if (options?.showTime) {
    formatOptions.hour = '2-digit'
    formatOptions.minute = '2-digit'
    if (options?.showSeconds) {
      formatOptions.second = '2-digit'
    }
  }

  return date.toLocaleString('zh-CN', formatOptions)
}

/**
 * 格式化速度
 * @param bytesPerSecond 每秒字节数
 * @returns 格式化后的速度字符串
 */
export const formatSpeed = (bytesPerSecond: number): string => {
  if (!bytesPerSecond || bytesPerSecond === 0) return '0 B/s'
  const k = 1024
  const sizes = ['B/s', 'KB/s', 'MB/s', 'GB/s']
  const i = Math.floor(Math.log(bytesPerSecond) / Math.log(k))
  return parseFloat((bytesPerSecond / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

/**
 * 截断URL显示
 * @param url URL字符串
 * @param maxLength 最大长度，默认50
 * @returns 截断后的URL字符串
 */
export const truncateUrl = (url: string, maxLength: number = 50): string => {
  if (url.length <= maxLength) return url
  return url.substring(0, maxLength) + '...'
}

/**
 * 格式化文件大小（简化版，用于显示）
 * @param bytes 字节数
 * @param useGB 是否使用GB单位（大于1GB时）
 * @returns 格式化后的文件大小字符串（MB或GB）
 */
export const formatFileSizeForDisplay = (bytes: number, useGB: boolean = true): string => {
  if (!bytes || bytes === 0) return '0 B'
  const mb = bytes / (1024 * 1024)
  if (useGB && bytes >= 1024 * 1024 * 1024) {
    const gb = bytes / (1024 * 1024 * 1024)
    return `${gb.toFixed(2)} GB`
  }
  return `${mb.toFixed(2)} MB`
}

/**
 * 将字节转换为 GB
 * @param bytes 字节数
 * @returns GB 数值（保留2位小数）
 */
export const bytesToGB = (bytes: number): number => {
  if (!bytes || bytes === 0) return 0
  return Math.round((bytes / (1024 * 1024 * 1024)) * 100) / 100
}

/**
 * 将 GB 转换为字节
 * @param gb GB 数值
 * @returns 字节数
 */
export const GBToBytes = (gb: number): number => {
  if (!gb || gb === 0) return 0
  return Math.round(gb * 1024 * 1024 * 1024)
}

/**
 * 格式化耗时（毫秒）
 * @param milliseconds 毫秒数
 * @returns 格式化后的耗时字符串（如：1分30秒、30秒、2小时15分）
 */
export const formatDuration = (milliseconds: number): string => {
  if (!milliseconds || milliseconds < 0) {
    return i18n.global.t('format.duration.zero') || '0秒'
  }
  
  const seconds = Math.floor(milliseconds / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)
  
  // 获取国际化文本
  const t = i18n.global.t
  
  if (days > 0) {
    const remainingHours = hours % 24
    const remainingMinutes = minutes % 60
    if (remainingHours > 0) {
      return t('format.duration.daysHoursMinutes', {
        days,
        hours: remainingHours,
        minutes: remainingMinutes
      })
    }
    return t('format.duration.daysMinutes', {
      days,
      minutes: remainingMinutes
    })
  }
  
  if (hours > 0) {
    const remainingMinutes = minutes % 60
    const remainingSeconds = seconds % 60
    if (remainingMinutes > 0) {
      return t('format.duration.hoursMinutes', {
        hours,
        minutes: remainingMinutes
      })
    }
    return t('format.duration.hoursSeconds', {
      hours,
      seconds: remainingSeconds
    })
  }
  
  if (minutes > 0) {
    const remainingSeconds = seconds % 60
    if (remainingSeconds > 0) {
      return t('format.duration.minutesSeconds', {
        minutes,
        seconds: remainingSeconds
      })
    }
    return t('format.duration.minutes', { minutes })
  }
  
  return t('format.duration.seconds', { seconds })
}

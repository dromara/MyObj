/**
 * 剪贴板工具函数
 */

import logger from '@/plugins/logger'

/**
 * 复制文本到剪贴板
 * @param text 要复制的文本
 * @returns Promise<boolean> 是否复制成功
 */
export const copyToClipboard = async (text: string): Promise<boolean> => {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch (err) {
    // 降级方案：使用传统方法
    try {
      const textArea = document.createElement('textarea')
      textArea.value = text
      textArea.style.position = 'fixed'
      textArea.style.left = '-999999px'
      textArea.style.top = '-999999px'
      document.body.appendChild(textArea)
      textArea.focus()
      textArea.select()
      const successful = document.execCommand('copy')
      document.body.removeChild(textArea)
      return successful
    } catch (error) {
      logger.error('复制失败:', error)
      return false
    }
  }
}

/**
 * 从剪贴板读取文本内容
 * @returns Promise<string | null> 剪贴板文本内容，失败返回 null
 */
export const readClipboardText = async (): Promise<string | null> => {
  try {
    // 优先使用 Clipboard API（需要 HTTPS 或 localhost）
    // 注意：此 API 需要在用户激活的上下文中调用（如点击事件），不能在定时器中调用
    if (navigator.clipboard && navigator.clipboard.readText) {
      const text = await navigator.clipboard.readText()
      return text
    }
    return null
  } catch (err: any) {
    // 权限被拒绝或其他错误
    // 注意：即使浏览器设置中允许了剪贴板权限，在定时器中调用仍然会失败
    // 这是浏览器的安全限制，需要在用户交互的上下文中调用
    if (err.name === 'NotAllowedError' || err.name === 'SecurityError') {
      // 静默处理，不记录警告（因为这是预期的行为）
      // logger.warn('剪贴板访问被拒绝，需要用户授权')
    } else {
      logger.error('读取剪贴板失败:', err)
    }
    return null
  }
}

/**
 * 从剪贴板读取文件（用于种子文件）
 * @returns Promise<File | null> 剪贴板文件，失败返回 null
 */
export const readClipboardFile = async (): Promise<File | null> => {
  try {
    // 使用 Clipboard API 读取文件
    // 注意：此 API 需要在用户激活的上下文中调用（如点击事件），不能在定时器中调用
    if (navigator.clipboard && navigator.clipboard.read) {
      const clipboardItems = await navigator.clipboard.read()
      for (const item of clipboardItems) {
        // 查找 .torrent 文件
        if (item.types.includes('application/x-bittorrent')) {
          const blob = await item.getType('application/x-bittorrent')
          return new File([blob], 'clipboard.torrent', { type: 'application/x-bittorrent' })
        }
        // 查找通用文件类型
        for (const type of item.types) {
          if (type.startsWith('application/') || type.startsWith('text/')) {
            const blob = await item.getType(type)
            const fileName = type.includes('torrent') ? 'clipboard.torrent' : 'clipboard.file'
            return new File([blob], fileName, { type })
          }
        }
      }
    }
    return null
  } catch (err: any) {
    // 权限被拒绝或其他错误
    // 注意：即使浏览器设置中允许了剪贴板权限，在定时器中调用仍然会失败
    // 这是浏览器的安全限制，需要在用户交互的上下文中调用
    if (err.name === 'NotAllowedError' || err.name === 'SecurityError') {
      // 静默处理，不记录警告（因为这是预期的行为）
      // logger.warn('剪贴板文件访问被拒绝，需要用户授权')
    } else {
      logger.error('读取剪贴板文件失败:', err)
    }
    return null
  }
}

/**
 * 将文件转换为 Base64 编码
 * @param file 文件对象
 * @returns Promise<string> Base64 编码的字符串
 */
export const fileToBase64 = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = reader.result as string
      // 移除 data URL 前缀（data:application/x-bittorrent;base64,）
      const base64 = result.split(',')[1] || result
      resolve(base64)
    }
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}
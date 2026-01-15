/**
 * 安全性工具函数
 * 提供 XSS 防护、输入验证等功能
 */

/**
 * HTML 转义
 * 防止 XSS 攻击
 */
export function escapeHtml(html: string): string {
  const div = document.createElement('div')
  div.textContent = html
  return div.innerHTML
}

/**
 * HTML 反转义
 */
export function unescapeHtml(escapedHtml: string): string {
  const div = document.createElement('div')
  div.innerHTML = escapedHtml
  return div.textContent || ''
}

/**
 * 清理用户输入
 * 移除潜在的恶意字符
 */
export function sanitizeInput(input: string): string {
  return input
    .replace(/[<>]/g, '') // 移除尖括号
    .replace(/javascript:/gi, '') // 移除 javascript: 协议
    .replace(/on\w+=/gi, '') // 移除事件处理器
    .trim()
}

/**
 * 验证文件名安全性
 * 防止路径遍历攻击
 */
export function isValidFileName(fileName: string): boolean {
  // 禁止的字符和模式
  const forbiddenChars = /[<>:"|?*\x00-\x1f]/
  const forbiddenNames = /^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])(\.|$)/i
  const pathTraversal = /\.\./

  if (forbiddenChars.test(fileName)) return false
  if (forbiddenNames.test(fileName)) return false
  if (pathTraversal.test(fileName)) return false
  if (fileName.length === 0 || fileName.length > 255) return false

  return true
}

/**
 * 验证 URL 安全性
 */
export function isValidUrl(url: string): boolean {
  try {
    const urlObj = new URL(url)
    // 只允许 http 和 https 协议
    return urlObj.protocol === 'http:' || urlObj.protocol === 'https:'
  } catch {
    return false
  }
}

/**
 * 生成 CSRF Token
 */
export function generateCSRFToken(): string {
  const array = new Uint8Array(32)
  crypto.getRandomValues(array)
  return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('')
}

/**
 * 验证 CSRF Token
 */
export function validateCSRFToken(token: string, storedToken: string): boolean {
  return token === storedToken && token.length > 0
}

/**
 * 内容安全策略（CSP）相关
 * 生成安全的随机字符串
 */
export function generateNonce(): string {
  const array = new Uint8Array(16)
  crypto.getRandomValues(array)
  return btoa(String.fromCharCode(...array))
}

/**
 * 安全的 JSON 解析
 * 防止原型污染
 */
export function safeJsonParse<T = any>(json: string, defaultValue: T): T {
  try {
    const parsed = JSON.parse(json)
    // 检查是否包含 __proto__ 或 constructor
    if (typeof parsed === 'object' && parsed !== null) {
      if ('__proto__' in parsed || 'constructor' in parsed) {
        return defaultValue
      }
    }
    return parsed
  } catch {
    return defaultValue
  }
}

/**
 * 深度克隆对象（防止原型污染）
 */
export function safeClone<T>(obj: T): T {
  if (obj === null || typeof obj !== 'object') {
    return obj
  }

  if (obj instanceof Date) {
    return new Date(obj.getTime()) as T
  }

  if (obj instanceof Array) {
    return obj.map(item => safeClone(item)) as T
  }

  if (typeof obj === 'object') {
    const cloned = {} as T
    for (const key in obj) {
      if (key !== '__proto__' && key !== 'constructor') {
        cloned[key] = safeClone(obj[key])
      }
    }
    return cloned
  }

  return obj
}

/**
 * 验证文件类型
 */
export function isValidFileType(fileName: string, allowedTypes: string[]): boolean {
  const extension = fileName.split('.').pop()?.toLowerCase()
  if (!extension) return false
  return allowedTypes.includes(extension)
}

/**
 * 验证文件大小
 */
export function isValidFileSize(size: number, maxSize: number): boolean {
  return size > 0 && size <= maxSize
}

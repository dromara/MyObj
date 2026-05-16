/**
 * 验证工具函数
 */

/**
 * 验证邮箱格式
 */
export function isValidEmail(email: string): boolean {
  const regex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return regex.test(email)
}

/**
 * 验证手机号格式（中国）
 */
export function isValidPhone(phone: string): boolean {
  const regex = /^1[3-9]\d{9}$/
  return regex.test(phone)
}

/**
 * 验证URL格式（基础验证，仅检查格式）
 * 注意：如需安全验证（限制协议），请使用 security.ts 中的 isValidUrl
 */
export function isValidUrlFormat(url: string): boolean {
  try {
    new URL(url)
    return true
  } catch {
    return false
  }
}

/**
 * 验证IP地址格式
 */
export function isValidIP(ip: string): boolean {
  const regex = /^(\d{1,3}\.){3}\d{1,3}$/
  if (!regex.test(ip)) return false

  const parts = ip.split('.')
  return parts.every(part => {
    const num = parseInt(part, 10)
    return num >= 0 && num <= 255
  })
}

/**
 * 验证身份证号格式（中国）
 */
export function isValidIDCard(idCard: string): boolean {
  const regex = /^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$/
  return regex.test(idCard)
}

/**
 * 验证密码强度
 * @param password 密码
 * @param minLength 最小长度
 * @param requireUppercase 是否需要大写字母
 * @param requireLowercase 是否需要小写字母
 * @param requireNumber 是否需要数字
 * @param requireSpecial 是否需要特殊字符
 */
export function validatePasswordStrength(
  password: string,
  minLength = 8,
  requireUppercase = true,
  requireLowercase = true,
  requireNumber = true,
  requireSpecial = false
): {
  valid: boolean
  errors: string[]
} {
  const errors: string[] = []

  if (password.length < minLength) {
    errors.push(`密码长度至少为 ${minLength} 位`)
  }

  if (requireUppercase && !/[A-Z]/.test(password)) {
    errors.push('密码必须包含大写字母')
  }

  if (requireLowercase && !/[a-z]/.test(password)) {
    errors.push('密码必须包含小写字母')
  }

  if (requireNumber && !/\d/.test(password)) {
    errors.push('密码必须包含数字')
  }

  if (requireSpecial && !/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
    errors.push('密码必须包含特殊字符')
  }

  return {
    valid: errors.length === 0,
    errors
  }
}

/**
 * 验证文件扩展名
 */
export function isValidFileExtension(filename: string, allowedExtensions: string[]): boolean {
  const extension = filename.split('.').pop()?.toLowerCase()
  if (!extension) return false
  return allowedExtensions.some(ext => ext.toLowerCase() === extension)
}

/**
 * 验证文件大小（基础验证，允许 size 为 0）
 * 注意：如需安全验证（不允许 size 为 0），请使用 security.ts 中的 isValidFileSize
 */
export function isValidFileSizeBasic(size: number, maxSize: number): boolean {
  return size <= maxSize
}

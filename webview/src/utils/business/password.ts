/**
 * 密码相关工具函数
 */

/**
 * 使用 Web Crypto API 生成安全的随机整数
 * @param max 上限（不含）
 * @returns 0 到 max-1 之间的随机整数
 */
function secureRandom(max: number): number {
  const array = new Uint32Array(1)
  crypto.getRandomValues(array)
  return array[0] % max
}

/**
 * 生成随机密码
 * @param length 密码长度，默认6
 * @param includeUpperCase 是否包含大写字母，默认false
 * @param includeSpecialChars 是否包含特殊字符，默认false
 * @returns 随机密码字符串
 */
export const generateRandomPassword = (
  length: number = 6,
  includeUpperCase: boolean = false,
  includeSpecialChars: boolean = false
): string => {
  let chars = 'abcdefghijklmnopqrstuvwxyz0123456789'

  if (includeUpperCase) {
    chars += 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
  }

  if (includeSpecialChars) {
    chars += '!@#$%^&*'
  }

  let password = ''
  for (let i = 0; i < length; i++) {
    password += chars.charAt(secureRandom(chars.length))
  }

  return password
}

/**
 * 字符串工具函数
 */

/**
 * 截断字符串
 * @param str 字符串
 * @param length 最大长度
 * @param suffix 后缀，默认 '...'
 * @returns 截断后的字符串
 */
export function truncate(str: string, length: number, suffix = '...'): string {
  if (!str || str.length <= length) return str
  return str.slice(0, length) + suffix
}

/**
 * 首字母大写
 * @param str 字符串
 * @returns 首字母大写的字符串
 */
export function capitalize(str: string): string {
  if (!str) return ''
  return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase()
}

/**
 * 驼峰命名转短横线命名
 * @param str 字符串
 * @returns 短横线命名的字符串
 */
export function camelToKebab(str: string): string {
  return str.replace(/([a-z0-9])([A-Z])/g, '$1-$2').toLowerCase()
}

/**
 * 短横线命名转驼峰命名
 * @param str 字符串
 * @returns 驼峰命名的字符串
 */
export function kebabToCamel(str: string): string {
  return str.replace(/-([a-z])/g, (_, letter) => letter.toUpperCase())
}

/**
 * 去除字符串两端的空白字符
 * @param str 字符串
 * @returns 去除空白后的字符串
 */
export function trim(str: string): string {
  return str.trim()
}

/**
 * 去除所有空白字符
 * @param str 字符串
 * @returns 去除空白后的字符串
 */
export function removeWhitespace(str: string): string {
  return str.replace(/\s+/g, '')
}

/**
 * 生成随机字符串
 * @param length 长度
 * @param chars 字符集
 * @returns 随机字符串
 */
export function randomString(
  length: number,
  chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
): string {
  let result = ''
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return result
}

/**
 * 判断字符串是否为空
 * @param str 字符串
 * @returns 是否为空
 */
export function isEmpty(str: string | null | undefined): boolean {
  return !str || str.trim().length === 0
}

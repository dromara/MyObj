/**
 * 数字工具函数
 */

/**
 * 格式化数字（添加千分位分隔符）
 * @param num 数字
 * @param decimals 小数位数
 * @returns 格式化后的字符串
 */
export function formatNumber(num: number | string, decimals = 0): string {
  const n = typeof num === 'string' ? parseFloat(num) : num
  if (isNaN(n)) return '0'

  const fixed = n.toFixed(decimals)
  const parts = fixed.split('.')
  parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  return parts.join('.')
}

/**
 * 格式化百分比
 * @param num 数字（0-1之间）
 * @param decimals 小数位数
 * @returns 百分比字符串
 */
export function formatPercent(num: number, decimals = 2): string {
  if (isNaN(num)) return '0%'
  return `${(num * 100).toFixed(decimals)}%`
}

/**
 * 限制数字范围
 * @param num 数字
 * @param min 最小值
 * @param max 最大值
 * @returns 限制后的数字
 */
export function clamp(num: number, min: number, max: number): number {
  return Math.min(Math.max(num, min), max)
}

/**
 * 生成随机数
 * @param min 最小值
 * @param max 最大值
 * @returns 随机数
 */
export function random(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

/**
 * 判断是否为数字
 * @param value 值
 * @returns 是否为数字
 */
export function isNumber(value: any): value is number {
  return typeof value === 'number' && !isNaN(value)
}

/**
 * 安全转换为数字
 * @param value 值
 * @param defaultValue 默认值
 * @returns 数字
 */
export function toNumber(value: any, defaultValue = 0): number {
  const num = typeof value === 'string' ? parseFloat(value) : Number(value)
  return isNaN(num) ? defaultValue : num
}

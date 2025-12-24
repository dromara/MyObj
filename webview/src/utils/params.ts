/**
 * 参数处理工具函数
 * 过滤掉 undefined、null 和空字符串值
 */

/**
 * 过滤无效参数（undefined、null、空字符串）
 * @param params 原始参数对象
 * @returns 过滤后的参数对象
 */
export function filterParams<T extends Record<string, any>>(params: T): Partial<T> {
  const filtered: Partial<T> = {}
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== null && value !== '') {
      filtered[key as keyof T] = value
    }
  }
  return filtered
}


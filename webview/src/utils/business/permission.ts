/**
 * 权限国际化工具函数
 * 根据权限特征码（characteristic）获取国际化后的名称和描述
 *
 * 参考 ruoyi-plus-soybean 的字典实现方式
 */

/**
 * 获取权限的国际化名称
 * @param characteristic 权限特征码
 * @param fallback 如果找不到国际化文本，返回的默认值（通常是数据库中的原始名称）
 * @param t 国际化翻译函数（可选，如果不提供则返回 fallback）
 * @returns 国际化后的权限名称
 */
export function getPermissionName(characteristic: string, fallback?: string, t?: (key: string) => string): string {
  if (!characteristic) return fallback || ''
  if (!t) return fallback || characteristic
  try {
    const key = `admin.permissions.dict.${characteristic}.name` as any
    const translated = t(key)

    if (translated === key || !translated) {
      return fallback || characteristic
    }

    return translated
  } catch (error) {
    return fallback || characteristic
  }
}

/**
 * 获取权限的国际化描述
 * @param characteristic 权限特征码
 * @param fallback 如果找不到国际化文本，返回的默认值（通常是数据库中的原始描述）
 * @param t 国际化翻译函数（可选，如果不提供则返回 fallback）
 * @returns 国际化后的权限描述
 */
export function getPermissionDescription(characteristic: string, fallback?: string, t?: (key: string) => string): string {
  if (!characteristic) return fallback || ''
  if (!t) return fallback || characteristic
  try {
    const key = `admin.permissions.dict.${characteristic}.description` as any
    const translated = t(key)

    if (translated === key || !translated) {
      return fallback || characteristic
    }

    return translated
  } catch (error) {
    return fallback || characteristic
  }
}

/**
 * 获取权限的国际化信息（名称和描述）
 * @param characteristic 权限特征码
 * @param fallbackName 如果找不到国际化文本，返回的默认名称
 * @param fallbackDescription 如果找不到国际化文本，返回的默认描述
 * @param t 国际化翻译函数（可选）
 * @returns 包含名称和描述的对象
 */
export function getPermissionInfo(
  characteristic: string,
  fallbackName?: string,
  fallbackDescription?: string,
  t?: (key: string) => string
): { name: string; description: string } {
  return {
    name: getPermissionName(characteristic, fallbackName, t),
    description: getPermissionDescription(characteristic, fallbackDescription, t)
  }
}

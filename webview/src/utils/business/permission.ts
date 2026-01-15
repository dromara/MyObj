/**
 * 权限国际化工具函数
 * 根据权限特征码（characteristic）获取国际化后的名称和描述
 *
 * 参考 ruoyi-plus-soybean 的字典实现方式
 */

import { useI18n } from '@/composables'

/**
 * 获取权限的国际化名称
 * @param characteristic 权限特征码
 * @param fallback 如果找不到国际化文本，返回的默认值（通常是数据库中的原始名称）
 * @returns 国际化后的权限名称
 */
export function getPermissionName(characteristic: string, fallback?: string): string {
  try {
    const { t } = useI18n()
    const key = `admin.permissions.dict.${characteristic}.name` as any
    const translated = t(key)

    // 如果翻译结果就是 key 本身（说明没有找到对应的翻译），返回 fallback 或原始值
    if (translated === key || !translated) {
      return fallback || characteristic
    }

    return translated
  } catch (error) {
    // 如果 useI18n 不可用（例如在非组件上下文中），返回 fallback
    return fallback || characteristic
  }
}

/**
 * 获取权限的国际化描述
 * @param characteristic 权限特征码
 * @param fallback 如果找不到国际化文本，返回的默认值（通常是数据库中的原始描述）
 * @returns 国际化后的权限描述
 */
export function getPermissionDescription(characteristic: string, fallback?: string): string {
  try {
    const { t } = useI18n()
    const key = `admin.permissions.dict.${characteristic}.description` as any
    const translated = t(key)

    // 如果翻译结果就是 key 本身（说明没有找到对应的翻译），返回 fallback 或原始值
    if (translated === key || !translated) {
      return fallback || characteristic
    }

    return translated
  } catch (error) {
    // 如果 useI18n 不可用（例如在非组件上下文中），返回 fallback
    return fallback || characteristic
  }
}

/**
 * 获取权限的国际化信息（名称和描述）
 * @param characteristic 权限特征码
 * @param fallbackName 如果找不到国际化文本，返回的默认名称
 * @param fallbackDescription 如果找不到国际化文本，返回的默认描述
 * @returns 包含名称和描述的对象
 */
export function getPermissionInfo(
  characteristic: string,
  fallbackName?: string,
  fallbackDescription?: string
): { name: string; description: string } {
  return {
    name: getPermissionName(characteristic, fallbackName),
    description: getPermissionDescription(characteristic, fallbackDescription)
  }
}

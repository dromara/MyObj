import { inject, ref, type Ref } from 'vue'
import { i18n, type LanguageEnum } from '@myobj/shared'

/**
 * 轻量级 useI18n — 不依赖 store
 * 通过 inject 获取 locale，回退到 i18n 全局实例
 */
export function useI18n() {
  const locale = inject<Ref<LanguageEnum>>('app-locale', ref(i18n.global.locale.value as LanguageEnum))

  const t = (key: string, params?: Record<string, unknown>): string => {
    if (params != null) {
      return i18n.global.t(key, params) as string
    }
    return i18n.global.t(key) as string
  }

  return { locale, t }
}

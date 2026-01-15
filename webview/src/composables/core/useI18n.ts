import { computed } from 'vue'
import { useI18n as useI18nVue } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { LanguageEnum } from '@/enums/LanguageEnum'

/**
 * 国际化 Composable
 * 提供翻译和语言切换功能
 */
export function useI18n() {
  const { t } = useI18nVue()
  const appStore = useAppStore()

  // 切换语言
  const changeLocale = (newLocale: LanguageEnum) => {
    appStore.changeLocale(newLocale)
    // Element Plus 语言会通过 ElConfigProvider 自动更新，无需重新加载页面
  }

  return {
    locale: computed(() => appStore.locale),
    t,
    setLocale: changeLocale,
    getLocale: () => appStore.locale
  }
}

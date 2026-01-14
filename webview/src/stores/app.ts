import { defineStore } from 'pinia'
import { ref, watch, computed } from 'vue'
import { useTitle } from '@vueuse/core'
import { LanguageEnum } from '@/enums/LanguageEnum'
import { StoreId } from '@/enums/StoreId'
import { $t, setLocale } from '@/i18n'
import router from '@/router'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import enUS from 'element-plus/dist/locale/en.mjs'

export const useAppStore = defineStore(StoreId.App, () => {
  // 语言设置
  const getInitialLocale = (): LanguageEnum => {
    const saved = localStorage.getItem('locale')
    if (saved && (saved === LanguageEnum.zh_CN || saved === LanguageEnum.en_US)) {
      return saved as LanguageEnum
    }
    return LanguageEnum.zh_CN
  }

  const locale = ref<LanguageEnum>(getInitialLocale())

  // Element Plus 语言包
  const elementPlusLocale = computed(() => {
    return locale.value === LanguageEnum.en_US ? enUS : zhCn
  })

  // 语言选项
  const localeOptions = [
    {
      label: '中文',
      key: LanguageEnum.zh_CN
    },
    {
      label: 'English',
      key: LanguageEnum.en_US
    }
  ]

  /**
   * 切换语言
   * @param lang 语言类型
   */
  function changeLocale(lang: LanguageEnum) {
    locale.value = lang
    setLocale(lang)
    localStorage.setItem('locale', lang)

    // 更新文档标题
    updateDocumentTitle()
  }

  /**
   * 根据当前路由更新文档标题
   */
  function updateDocumentTitle() {
    const route = router.currentRoute.value
    const routeTitle = route.meta.title as string
    const i18nKey = route.meta.i18nKey as string

    if (i18nKey) {
      const documentTitle = $t(i18nKey)
      useTitle(documentTitle)
    } else if (routeTitle) {
      useTitle(routeTitle)
    }
  }

  // 监听语言变化
  watch(locale, () => {
    updateDocumentTitle()
  })

  return {
    locale,
    localeOptions,
    elementPlusLocale,
    changeLocale,
    updateDocumentTitle
  }
})

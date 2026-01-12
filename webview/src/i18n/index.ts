import type { App } from 'vue'
import { createI18n } from 'vue-i18n'
import { LanguageEnum } from '@/enums/LanguageEnum'
import zhCN from './locales/zh-CN'
import enUS from './locales/en-US'

/**
 * 获取当前语言
 * @returns zh-CN | en-US
 */
export const getLanguage = (): LanguageEnum => {
  const saved = localStorage.getItem('locale')
  if (saved && (saved === LanguageEnum.zh_CN || saved === LanguageEnum.en_US)) {
    return saved as LanguageEnum
  }
  
  // 根据浏览器语言自动选择
  const browserLang = navigator.language || (navigator as any).userLanguage
  if (browserLang.startsWith('en')) {
    return LanguageEnum.en_US
  }
  return LanguageEnum.zh_CN
}

const i18n = createI18n({
  globalInjection: true,  // 全局注入 $t 函数
  allowComposition: true, // 允许 Composition API
  legacy: false,          // 不使用 legacy 模式
  locale: getLanguage(),
  fallbackLocale: LanguageEnum.zh_CN,
  messages: {
    [LanguageEnum.zh_CN]: zhCN,  // 'zh-CN'
    [LanguageEnum.en_US]: enUS,  // 'en-US'
    // 同时支持简化的语言代码（vue-i18n 内部可能会使用）
    'zh': zhCN,
    'en': enUS
  },
  warnHtmlMessage: false, // 禁用 HTML 消息警告
  pluralRules: {
    // 为所有语言禁用复数规则，避免 | 字符被误识别
    [LanguageEnum.zh_CN]: () => 'other',  // 'zh-CN'
    [LanguageEnum.en_US]: () => 'other',  // 'en-US'
    'zh': () => 'other',
    'en': () => 'other'
  }
})

/**
 * Setup plugin i18n
 * @param app Vue App instance
 */
export function setupI18n(app: App) {
  app.use(i18n)
}

// 导出全局翻译函数
export const $t = i18n.global.t

/**
 * 切换语言
 * @param locale 语言类型
 */
export function setLocale(locale: LanguageEnum) {
  i18n.global.locale.value = locale
  localStorage.setItem('locale', locale)
}

/**
 * 获取当前语言
 */
export function getLocale(): LanguageEnum {
  return i18n.global.locale.value as LanguageEnum
}

export default i18n

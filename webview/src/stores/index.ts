import type { App } from 'vue'
import { createPinia } from 'pinia'
import { resetSetupStore } from './plugins'

/**
 * Store 统一导出
 */
export { useUserStore } from './user'
export { useAuthStore } from './auth'
export { useAppStore } from './app'
export { useLayoutStore } from './layout'
export { useTagsViewStore } from './tagsView'

// 导出类型
export type { LayoutMode } from './layout'

/**
 * 设置 Vue Store (Pinia)
 * 初始化 Pinia 并注册插件
 *
 * @param app Vue 应用实例
 */
export function setupStore(app: App) {
  const store = createPinia()

  // 注册插件
  store.use(resetSetupStore)

  // 安装到应用
  app.use(store)
}

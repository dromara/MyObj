import type { ComponentInternalInstance as ComponentInstance } from 'vue'
import modal from '@/plugins/modal'
import cache from '@/plugins/cache'
import logger from '@/plugins/logger'

export {}

declare global {
  /** vue Instance */
  declare type ComponentInternalInstance = ComponentInstance
}

declare module 'vue' {
  interface ComponentCustomProperties {
    // 模态框对象
    $modal: typeof modal
    // 缓存对象
    $cache: typeof cache
    // 日志对象
    $log: typeof logger
    // 注意：$router 和 $route 由 Vue Router 自动注入，无需手动声明
    // 注意：Store 不需要全局注册，直接使用 useXxxStore() 即可（Pinia 官方推荐）
  }
}

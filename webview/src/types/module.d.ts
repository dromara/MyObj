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
  }
}


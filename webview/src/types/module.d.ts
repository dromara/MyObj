import type { ComponentInternalInstance as ComponentInstance } from 'vue'
import modal from '@/plugins/modal'
import cache from '@/plugins/cache'
import logger from '@/plugins/logger'
import { useUserStore } from '@/stores/user'
import { useAuthStore } from '@/stores/auth'

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
    // Store 对象（返回 store 实例的函数）
    $userStore: () => ReturnType<typeof useUserStore>
    $authStore: () => ReturnType<typeof useAuthStore>
    // 注意：$router 和 $route 由 Vue Router 自动注入，无需手动声明
  }
}


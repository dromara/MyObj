import { App, Plugin } from 'vue'
import { Router } from 'vue-router'
import modal from './modal'
import cache from './cache'
import logger from './logger'
import ElementIcons from './svgicon'
import { useUserStore } from '@/stores/user'
import { useAuthStore } from '@/stores/auth'

// 存储 router 实例，以便在插件中使用
let routerInstance: Router | null = null

export function setRouter(router: Router) {
  routerInstance = router
}

const installPlugin: Plugin = {
  install(app: App) {
    // 模态框对象
    app.config.globalProperties.$modal = modal

    // 缓存对象
    app.config.globalProperties.$cache = cache

    // 日志对象
    app.config.globalProperties.$log = logger

    // Router 对象
    if (routerInstance) {
      app.config.globalProperties.$router = routerInstance
    }

    // Store 对象（提供便捷访问）
    app.config.globalProperties.$userStore = () => useUserStore()
    app.config.globalProperties.$authStore = () => useAuthStore()

    // Element Plus 图标
    app.use(ElementIcons)
  }
}

export default installPlugin


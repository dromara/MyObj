import { App, Plugin } from 'vue'
import { Router } from 'vue-router'
import modal from './modal'
import cache from './cache'
import logger from './logger'

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

    // 注意：Store 不需要全局注册
    // Pinia 官方推荐直接使用 useXxxStore()，这是 Composition API 的标准用法
    // 参考：https://pinia.vuejs.org/zh/getting-started.html
  }
}

export default installPlugin

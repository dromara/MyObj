import { App } from 'vue'
import modal from './modal'
import cache from './cache'
import logger from './logger'
import ElementIcons from './svgicon'

export default function installPlugin(app: App) {
  // 模态框对象
  app.config.globalProperties.$modal = modal

  // 缓存对象
  app.config.globalProperties.$cache = cache

  // 日志对象
  app.config.globalProperties.$log = logger

  // Element Plus 图标
  app.use(ElementIcons)
}


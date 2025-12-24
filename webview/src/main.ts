import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import App from './App.vue'
import router from './router'
import '@/assets/styles/index.css'

// 注册插件
import plugins from './plugins/index'

// 修改 el-dialog 默认点击遮照为不关闭
import { ElDialog } from 'element-plus'
ElDialog.props.closeOnClickModal.default = false

// 配置日志（从环境变量加载配置）
// 可以在不同环境的 .env 文件中设置：
// .env.development - 开发环境配置
// .env.production - 生产环境配置
// 
// 配置项：
// VITE_LOG_LEVEL=debug          # 日志级别: debug, info, warn, error, none
// VITE_LOG_ENABLE=true           # 是否启用日志: true/false
// VITE_LOG_ENABLE_TIMESTAMP=true # 是否显示时间戳: true/false
// VITE_LOG_ENABLE_CALLER=true    # 是否显示调用者信息: true/false
import { loadLoggerConfigFromEnv } from '@/utils/logger-config'
loadLoggerConfigFromEnv()

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

// 设置 router 实例，以便在插件中使用
import { setRouter } from './plugins/index'
setRouter(router)

app.use(ElementPlus, {
  locale: zhCn,
})
app.use(plugins)
app.mount('#app')

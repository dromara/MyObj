import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
// 导入 Element Plus 深色模式 CSS（必须在默认样式之后导入）
import 'element-plus/theme-chalk/dark/css-vars.css'
import App from './App.vue'
import router from './router'
import '@/assets/styles/index.css'
import { setupI18n } from '@/i18n'

// 注册插件
import plugins from './plugins/index'

// Element Plus 图标（需要在 ElementPlus 之前注册）
import ElementIcons from './plugins/svgicon'

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
import { loadLoggerConfigFromEnv } from '@/utils/config/logger-config'
loadLoggerConfigFromEnv()

const app = createApp(App)

// 设置 Store (Pinia)
import { setupStore } from './stores'
setupStore(app)

app.use(router)

// 设置 router 实例，以便在插件中使用
import { setRouter } from './plugins/index'
setRouter(router)

// 设置国际化
setupI18n(app)

// Element Plus 全局配置（语言包现在通过 ElConfigProvider 动态切换）
app.use(ElementPlus)

// 注册 Element Plus 图标
app.use(ElementIcons)

app.use(plugins)
app.mount('#app')

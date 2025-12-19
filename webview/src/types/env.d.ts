declare module '*.vue' {
  import { DefineComponent } from 'vue'
  const Component: DefineComponent<{}, {}, any>
  export default Component
}

// 环境变量
interface ImportMetaEnv {
  // Vite 内置环境变量
  readonly MODE: string
  
  // 应用配置
  readonly VITE_APP_TITLE: string
  readonly VITE_APP_VERSION: string
  readonly VITE_APP_PORT?: string // 开发服务器端口
  readonly VITE_APP_BASE_PATH?: string // 部署路径（如 /admin/）
  readonly VITE_APP_BASE_URL?: string // API 基础 URL（后端服务器地址）
  readonly VITE_APP_BASE_API?: string // API 代理路径（如 /dev-api 或 /prod-api）
  
  // 构建配置
  readonly VITE_BUILD_COMPRESS?: string // 压缩配置: gzip, brotli, 或 gzip,brotli
  
  // 日志配置（可选）
  readonly VITE_LOG_LEVEL?: string // 日志级别: debug, info, warn, error, none
  readonly VITE_LOG_ENABLE?: string // 是否启用日志: true/false
  readonly VITE_LOG_ENABLE_TIMESTAMP?: string // 是否显示时间戳: true/false
  readonly VITE_LOG_ENABLE_CALLER?: string // 是否显示调用者信息: true/false
  
  // API 配置（可选，根据实际需要添加）
  // readonly VITE_APP_BASE_API?: string
  // readonly VITE_APP_BASE_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}


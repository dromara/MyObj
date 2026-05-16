/**
 * 环境变量管理工具
 */

/**
 * 获取环境变量
 * @param key 环境变量键
 * @param defaultValue 默认值
 * @returns 环境变量值
 */
export function getEnv(key: string, defaultValue?: string): string {
  const value = import.meta.env[key]
  if (value === undefined || value === null) {
    if (defaultValue !== undefined) {
      return defaultValue
    }
    console.warn(`环境变量 ${key} 未设置`)
    return ''
  }
  return String(value)
}

/**
 * 获取布尔类型环境变量
 */
export function getEnvBoolean(key: string, defaultValue = false): boolean {
  const value = getEnv(key, String(defaultValue))
  return value === 'true' || value === '1'
}

/**
 * 获取数字类型环境变量
 */
export function getEnvNumber(key: string, defaultValue = 0): number {
  const value = getEnv(key, String(defaultValue))
  const num = Number(value)
  return isNaN(num) ? defaultValue : num
}

/**
 * 判断是否为开发环境
 */
export function isDev(): boolean {
  return import.meta.env.DEV || import.meta.env.MODE === 'development'
}

/**
 * 判断是否为生产环境
 */
export function isProd(): boolean {
  return import.meta.env.PROD || import.meta.env.MODE === 'production'
}

/**
 * 获取应用基础路径
 */
export function getBasePath(): string {
  return getEnv('VITE_APP_BASE_PATH', '/')
}

/**
 * 获取 API 基础路径
 */
export function getApiBasePath(): string {
  return getEnv('VITE_APP_BASE_API', '/api')
}

/**
 * 获取 API 基础 URL
 */
export function getApiBaseUrl(): string {
  return getEnv('VITE_APP_BASE_URL', 'http://localhost:8080')
}

/**
 * 获取应用端口
 */
export function getAppPort(): number {
  return getEnvNumber('VITE_APP_PORT', 5173)
}

/**
 * 环境配置对象
 */
export const env = {
  // 环境信息
  isDev: isDev(),
  isProd: isProd(),
  mode: import.meta.env.MODE,

  // 应用配置
  basePath: getBasePath(),
  apiBasePath: getApiBasePath(),
  apiBaseUrl: getApiBaseUrl(),
  appPort: getAppPort(),

  // 功能开关
  enableMock: getEnvBoolean('VITE_APP_ENABLE_MOCK', false),
  enableDebug: getEnvBoolean('VITE_APP_ENABLE_DEBUG', false),

  // 其他配置
  appTitle: getEnv('VITE_APP_TITLE', 'MyObj'),
  appVersion: getEnv('VITE_APP_VERSION', '1.0.0')
}

export default env

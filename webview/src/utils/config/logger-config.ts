/**
 * 日志配置工具
 * 从环境变量加载日志配置
 */

import logger, { LogLevel } from '@/plugins/logger'

/**
 * 日志级别映射
 */
const levelMap: Record<string, LogLevel> = {
  debug: LogLevel.DEBUG,
  info: LogLevel.INFO,
  warn: LogLevel.WARN,
  error: LogLevel.ERROR,
  none: LogLevel.NONE
}

/**
 * 从环境变量加载日志配置
 *
 * 支持的环境变量：
 * - VITE_LOG_LEVEL: 日志级别 (debug, info, warn, error, none)
 * - VITE_LOG_ENABLE: 是否启用日志 (true/false)
 * - VITE_LOG_ENABLE_TIMESTAMP: 是否显示时间戳 (true/false)
 * - VITE_LOG_ENABLE_CALLER: 是否显示调用者信息 (true/false)
 *
 * 如果未设置环境变量，则使用默认配置：
 * - 开发环境: level=debug, enable=true, enableCaller=true
 * - 生产环境: level=error, enable=true, enableCaller=false
 *
 * 可以在不同环境的 .env 文件中设置不同的值：
 * - .env.development: 开发环境配置
 * - .env.production: 生产环境配置
 */
export const loadLoggerConfigFromEnv = () => {
  const isDev = import.meta.env.DEV

  // 从环境变量读取日志级别，未设置则使用默认值
  const logLevel = import.meta.env.VITE_LOG_LEVEL
  if (logLevel) {
    const level = levelMap[logLevel.toLowerCase()]
    if (level !== undefined) {
      logger.setLevel(level)
    }
  } else {
    // 默认配置：开发环境 debug，生产环境 error
    logger.setLevel(isDev ? LogLevel.DEBUG : LogLevel.ERROR)
  }

  // 从环境变量读取是否启用日志
  if (import.meta.env.VITE_LOG_ENABLE !== undefined) {
    logger.setEnable(import.meta.env.VITE_LOG_ENABLE === 'true')
  } else {
    // 默认：启用日志
    logger.setEnable(true)
  }

  // 从环境变量读取是否显示时间戳
  if (import.meta.env.VITE_LOG_ENABLE_TIMESTAMP !== undefined) {
    logger.setEnableTimestamp(import.meta.env.VITE_LOG_ENABLE_TIMESTAMP === 'true')
  } else {
    // 默认：显示时间戳
    logger.setEnableTimestamp(true)
  }

  // 从环境变量读取是否显示调用者信息
  if (import.meta.env.VITE_LOG_ENABLE_CALLER !== undefined) {
    logger.setEnableCaller(import.meta.env.VITE_LOG_ENABLE_CALLER === 'true')
  } else {
    // 默认：开发环境显示，生产环境不显示
    logger.setEnableCaller(isDev)
  }
}

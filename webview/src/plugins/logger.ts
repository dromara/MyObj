/**
 * 日志工具函数
 * 智能区分开发环境和生产环境
 */

/**
 * 判断是否为开发环境
 */
const isDev = () => {
  return import.meta.env.DEV || import.meta.env.MODE === 'development'
}

/**
 * 日志级别
 */
export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
  NONE = 4
}

/**
 * 日志配置
 */
interface LoggerConfig {
  level: LogLevel
  enable: boolean
  enableTimestamp: boolean
  enableCaller: boolean
}

const defaultConfig: LoggerConfig = {
  level: isDev() ? LogLevel.DEBUG : LogLevel.ERROR,
  enable: true, // 默认启用日志
  enableTimestamp: true,
  enableCaller: false
}

/**
 * 获取调用者信息
 */
const getCallerInfo = (): string => {
  try {
    const stack = new Error().stack
    if (!stack) return ''
    
    const stackLines = stack.split('\n')
    // 跳过 Error 和 getCallerInfo 本身，获取真正的调用者
    if (stackLines.length > 3) {
      const callerLine = stackLines[3]
      const match = callerLine.match(/at\s+(.+?)\s+\((.+?):(\d+):\d+\)/)
      if (match) {
        const [, funcName, filePath, line] = match
        const fileName = filePath.split('/').pop() || filePath
        return `${funcName}@${fileName}:${line}`
      }
    }
  } catch (error) {
    // 忽略错误
  }
  return ''
}

/**
 * 格式化日志消息
 */
const formatMessage = (level: string, message: string, config: LoggerConfig): string => {
  const parts: string[] = []
  
  if (config.enableTimestamp) {
    const now = new Date()
    const hours = String(now.getHours()).padStart(2, '0')
    const minutes = String(now.getMinutes()).padStart(2, '0')
    const seconds = String(now.getSeconds()).padStart(2, '0')
    const milliseconds = String(now.getMilliseconds()).padStart(3, '0')
    const timestamp = `${hours}:${minutes}:${seconds}.${milliseconds}`
    parts.push(`[${timestamp}]`)
  }
  
  parts.push(`[${level}]`)
  
  if (config.enableCaller) {
    const caller = getCallerInfo()
    if (caller) {
      parts.push(`[${caller}]`)
    }
  }
  
  parts.push(message)
  
  return parts.join(' ')
}

/**
 * 日志工具对象
 */
const logger = {
  /**
   * 配置
   */
  config: { ...defaultConfig },
  
  /**
   * 设置日志级别
   */
  setLevel(level: LogLevel) {
    this.config.level = level
  },
  
  /**
   * 设置是否启用日志
   */
  setEnable(enable: boolean) {
    this.config.enable = enable
  },
  
  /**
   * 设置是否显示时间戳
   */
  setEnableTimestamp(enable: boolean) {
    this.config.enableTimestamp = enable
  },
  
  /**
   * 设置是否显示调用者信息
   */
  setEnableCaller(enable: boolean) {
    this.config.enableCaller = enable
  },
  
  /**
   * 检查是否应该输出日志
   */
  shouldLog(level: LogLevel): boolean {
    // 如果未启用日志，则不输出
    if (!this.config.enable) {
      return false
    }
    
    // 检查日志级别
    return level >= this.config.level
  },
  
  /**
   * Debug 日志（开发环境）
   */
  debug(...args: any[]) {
    if (!this.shouldLog(LogLevel.DEBUG)) return
    
    const message = formatMessage('DEBUG', args.map(arg => 
      typeof arg === 'object' ? JSON.stringify(arg, null, 2) : String(arg)
    ).join(' '), this.config)
    
    console.debug(message)
  },
  
  /**
   * Info 日志
   */
  info(...args: any[]) {
    if (!this.shouldLog(LogLevel.INFO)) return
    
    const message = formatMessage('INFO', args.map(arg => 
      typeof arg === 'object' ? JSON.stringify(arg, null, 2) : String(arg)
    ).join(' '), this.config)
    
    console.info(message)
  },
  
  /**
   * Warn 日志
   */
  warn(...args: any[]) {
    if (!this.shouldLog(LogLevel.WARN)) return
    
    const message = formatMessage('WARN', args.map(arg => 
      typeof arg === 'object' ? JSON.stringify(arg, null, 2) : String(arg)
    ).join(' '), this.config)
    
    console.warn(message)
  },
  
  /**
   * Error 日志（生产环境也会输出）
   */
  error(...args: any[]) {
    if (!this.shouldLog(LogLevel.ERROR)) return
    
    const message = formatMessage('ERROR', args.map(arg => {
      if (arg instanceof Error) {
        return `${arg.message}\n${arg.stack}`
      }
      return typeof arg === 'object' ? JSON.stringify(arg, null, 2) : String(arg)
    }).join(' '), this.config)
    
    console.error(message)
  },
  
  /**
   * 分组日志（开发环境）
   */
  group(label: string) {
    if (isDev()) {
      console.group(label)
    }
  },
  
  /**
   * 结束分组
   */
  groupEnd() {
    if (isDev()) {
      console.groupEnd()
    }
  },
  
  /**
   * 表格日志（开发环境）
   */
  table(data: any) {
    if (isDev()) {
      console.table(data)
    }
  },
  
  /**
   * 时间日志（开发环境）
   */
  time(label: string) {
    if (isDev()) {
      console.time(label)
    }
  },
  
  /**
   * 结束时间日志
   */
  timeEnd(label: string) {
    if (isDev()) {
      console.timeEnd(label)
    }
  }
}

export default logger


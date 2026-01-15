/**
 * 错误处理 Composable
 * 提供统一的错误处理机制
 */

import { onErrorCaptured } from 'vue'

export interface ErrorInfo {
  message: string
  stack?: string
  component?: string
  timestamp: number
}

export function useErrorHandler() {
  const errors = ref<ErrorInfo[]>([])
  const hasError = ref(false)

  /**
   * 处理错误
   */
  const handleError = (error: Error, errorInfo?: any) => {
    const errorInfoObj: ErrorInfo = {
      message: error.message || 'Unknown error',
      stack: error.stack,
      component: errorInfo?.componentName || 'Unknown',
      timestamp: Date.now()
    }

    errors.value.push(errorInfoObj)
    hasError.value = true

    // 记录到控制台
    console.error('Error caught:', errorInfoObj)

    // 可以在这里添加错误上报逻辑
    // reportError(errorInfoObj)
  }

  /**
   * 处理 Promise 拒绝
   */
  const handleUnhandledRejection = (event: PromiseRejectionEvent) => {
    const error = event.reason instanceof Error ? event.reason : new Error(String(event.reason))

    handleError(error, { component: 'Promise Rejection' })
  }

  /**
   * 清除错误
   */
  const clearErrors = () => {
    errors.value = []
    hasError.value = false
  }

  /**
   * 获取最新错误
   */
  const getLatestError = computed(() => {
    return errors.value.length > 0 ? errors.value[errors.value.length - 1] : null
  })

  // 监听 Vue 组件错误
  onErrorCaptured((err, instance, info) => {
    handleError(err, {
      component: instance?.$options.name || 'Unknown Component',
      info
    })
    return false // 阻止错误继续传播
  })

  // 监听未处理的 Promise 拒绝
  if (typeof window !== 'undefined') {
    window.addEventListener('unhandledrejection', handleUnhandledRejection)
  }

  return {
    errors,
    hasError,
    handleError,
    clearErrors,
    getLatestError
  }
}

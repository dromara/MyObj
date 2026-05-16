/**
 * 性能监控 Composable
 * 提供性能指标收集和分析
 */

export interface PerformanceMetrics {
  // 页面加载性能
  pageLoadTime: number
  domContentLoaded: number
  firstPaint: number
  firstContentfulPaint: number

  // 资源加载性能
  resourceLoadTimes: Record<string, number>

  // 内存使用
  memoryUsage?: {
    usedJSHeapSize: number
    totalJSHeapSize: number
    jsHeapSizeLimit: number
  }

  // 网络性能
  networkInfo?: {
    effectiveType: string
    downlink: number
    rtt: number
  }
}

export function usePerformance() {
  const metrics = ref<PerformanceMetrics | null>(null)
  const isMonitoring = ref(false)

  /**
   * 获取页面加载性能指标
   */
  const getPageLoadMetrics = (): Partial<PerformanceMetrics> => {
    if (typeof window === 'undefined' || !window.performance) {
      return {}
    }

    const perfData = window.performance.timing

    return {
      pageLoadTime: perfData.loadEventEnd - perfData.navigationStart,
      domContentLoaded: perfData.domContentLoadedEventEnd - perfData.navigationStart,
      firstPaint: 0, // 需要通过 PerformanceObserver 获取
      firstContentfulPaint: 0 // 需要通过 PerformanceObserver 获取
    }
  }

  /**
   * 获取资源加载时间
   */
  const getResourceLoadTimes = (): Record<string, number> => {
    if (typeof window === 'undefined' || !window.performance) {
      return {}
    }

    const resources = window.performance.getEntriesByType('resource') as PerformanceResourceTiming[]
    const loadTimes: Record<string, number> = {}

    resources.forEach(resource => {
      const duration = resource.responseEnd - resource.requestStart
      loadTimes[resource.name] = duration
    })

    return loadTimes
  }

  /**
   * 获取内存使用情况
   */
  const getMemoryUsage = (): PerformanceMetrics['memoryUsage'] => {
    if (typeof window === 'undefined' || !(window.performance as any).memory) {
      return undefined
    }

    const memory = (window.performance as any).memory
    return {
      usedJSHeapSize: memory.usedJSHeapSize,
      totalJSHeapSize: memory.totalJSHeapSize,
      jsHeapSizeLimit: memory.jsHeapSizeLimit
    }
  }

  /**
   * 获取网络信息
   */
  const getNetworkInfo = (): PerformanceMetrics['networkInfo'] => {
    if (typeof window === 'undefined' || !(navigator as any).connection) {
      return undefined
    }

    const connection = (navigator as any).connection
    return {
      effectiveType: connection.effectiveType || 'unknown',
      downlink: connection.downlink || 0,
      rtt: connection.rtt || 0
    }
  }

  /**
   * 收集所有性能指标
   */
  const collectMetrics = (): PerformanceMetrics => {
    const pageMetrics = getPageLoadMetrics()
    const resourceTimes = getResourceLoadTimes()
    const memory = getMemoryUsage()
    const network = getNetworkInfo()

    return {
      pageLoadTime: pageMetrics.pageLoadTime || 0,
      domContentLoaded: pageMetrics.domContentLoaded || 0,
      firstPaint: pageMetrics.firstPaint || 0,
      firstContentfulPaint: pageMetrics.firstContentfulPaint || 0,
      resourceLoadTimes: resourceTimes,
      memoryUsage: memory,
      networkInfo: network
    }
  }

  /**
   * 开始监控
   */
  const startMonitoring = () => {
    if (isMonitoring.value) return

    isMonitoring.value = true

    // 监听 Paint Timing
    if (typeof window !== 'undefined' && 'PerformanceObserver' in window) {
      try {
        const observer = new PerformanceObserver(list => {
          for (const entry of list.getEntries()) {
            if (entry.entryType === 'paint') {
              const paintEntry = entry as PerformancePaintTiming
              if (paintEntry.name === 'first-paint') {
                metrics.value = {
                  ...metrics.value,
                  firstPaint: paintEntry.startTime
                } as PerformanceMetrics
              } else if (paintEntry.name === 'first-contentful-paint') {
                metrics.value = {
                  ...metrics.value,
                  firstContentfulPaint: paintEntry.startTime
                } as PerformanceMetrics
              }
            }
          }
        })

        observer.observe({ entryTypes: ['paint'] })
      } catch (e) {
        console.warn('PerformanceObserver not supported:', e)
      }
    }

    // 页面加载完成后收集指标
    if (document.readyState === 'complete') {
      metrics.value = collectMetrics()
    } else {
      window.addEventListener('load', () => {
        metrics.value = collectMetrics()
      })
    }
  }

  /**
   * 停止监控
   */
  const stopMonitoring = () => {
    isMonitoring.value = false
  }

  /**
   * 格式化性能指标为可读字符串
   */
  const formatMetrics = (m: PerformanceMetrics): string => {
    const lines: string[] = []
    lines.push(`页面加载时间: ${m.pageLoadTime.toFixed(2)}ms`)
    lines.push(`DOM 内容加载: ${m.domContentLoaded.toFixed(2)}ms`)
    if (m.firstPaint > 0) {
      lines.push(`首次绘制: ${m.firstPaint.toFixed(2)}ms`)
    }
    if (m.firstContentfulPaint > 0) {
      lines.push(`首次内容绘制: ${m.firstContentfulPaint.toFixed(2)}ms`)
    }
    if (m.memoryUsage) {
      lines.push(`内存使用: ${(m.memoryUsage.usedJSHeapSize / 1024 / 1024).toFixed(2)}MB`)
    }
    if (m.networkInfo) {
      lines.push(`网络类型: ${m.networkInfo.effectiveType}`)
      lines.push(`下载速度: ${m.networkInfo.downlink}Mbps`)
    }
    return lines.join('\n')
  }

  return {
    metrics,
    isMonitoring,
    collectMetrics,
    startMonitoring,
    stopMonitoring,
    formatMetrics,
    getMemoryUsage,
    getNetworkInfo
  }
}

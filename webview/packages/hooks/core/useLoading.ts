import { ref, computed } from 'vue'

/**
 * 全局加载状态管理
 * 提供统一的加载状态管理功能
 */
export function useLoading(initialValue = false) {
  const loading = ref(initialValue)
  const loadingText = ref<string>('')
  const loadingCount = ref(0)

  /**
   * 设置加载状态
   */
  const setLoading = (value: boolean, text?: string) => {
    loading.value = value
    if (text !== undefined) {
      loadingText.value = text
    }
    if (value) {
      loadingCount.value++
    } else {
      loadingCount.value = Math.max(0, loadingCount.value - 1)
    }
  }

  /**
   * 开始加载
   */
  const startLoading = (text?: string) => {
    setLoading(true, text)
  }

  /**
   * 停止加载
   */
  const stopLoading = () => {
    setLoading(false)
  }

  /**
   * 切换加载状态
   */
  const toggleLoading = (text?: string) => {
    setLoading(!loading.value, text)
  }

  /**
   * 执行异步操作时自动管理加载状态
   */
  const withLoading = async <T>(asyncFn: () => Promise<T>, text?: string): Promise<T> => {
    try {
      startLoading(text)
      return await asyncFn()
    } finally {
      stopLoading()
    }
  }

  /**
   * 是否正在加载（计算属性，考虑加载计数）
   */
  const isLoading = computed(() => loading.value && loadingCount.value > 0)

  return {
    loading: isLoading,
    loadingText,
    loadingCount,
    setLoading,
    startLoading,
    stopLoading,
    toggleLoading,
    withLoading
  }
}

/**
 * 全局加载状态实例（单例模式）
 */
let globalLoadingInstance: ReturnType<typeof useLoading> | null = null

/**
 * 获取全局加载状态实例
 */
export function useGlobalLoading() {
  if (!globalLoadingInstance) {
    globalLoadingInstance = useLoading(false)
  }
  return globalLoadingInstance
}

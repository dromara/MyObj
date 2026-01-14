/**
 * 懒加载 Composable
 * 使用 Intersection Observer API 实现元素懒加载
 */

export interface UseLazyLoadOptions {
  /** 根元素，默认为 null（视口） */
  root?: Element | null
  /** 根元素的边距，用于提前加载 */
  rootMargin?: string
  /** 触发加载的阈值，0-1 之间 */
  threshold?: number | number[]
  /** 是否立即加载（用于测试） */
  immediate?: boolean
}

/**
 * 懒加载 Composable
 * @param options 配置选项
 * @returns 响应式引用和加载状态
 */
export function useLazyLoad(options: UseLazyLoadOptions = {}) {
  const { root = null, rootMargin = '50px', threshold = 0.1, immediate = false } = options

  const target = ref<Element | null>(null)
  const isVisible = ref(immediate)
  const hasLoaded = ref(immediate)

  let observer: IntersectionObserver | null = null

  const observe = () => {
    if (!target.value || immediate) {
      isVisible.value = true
      hasLoaded.value = true
      return
    }

    // 检查浏览器是否支持 Intersection Observer
    if (typeof IntersectionObserver === 'undefined') {
      // 不支持时直接加载
      isVisible.value = true
      hasLoaded.value = true
      return
    }

    observer = new IntersectionObserver(
      entries => {
        entries.forEach(entry => {
          if (entry.isIntersecting) {
            isVisible.value = true
            hasLoaded.value = true
            // 加载后停止观察
            if (observer && target.value) {
              observer.unobserve(target.value)
            }
          }
        })
      },
      {
        root,
        rootMargin,
        threshold
      }
    )

    observer.observe(target.value)
  }

  const unobserve = () => {
    if (observer && target.value) {
      observer.unobserve(target.value)
      observer.disconnect()
      observer = null
    }
  }

  watch(
    () => target.value,
    newTarget => {
      if (newTarget) {
        observe()
      } else {
        unobserve()
      }
    },
    { immediate: true }
  )

  onUnmounted(() => {
    unobserve()
  })

  return {
    target,
    isVisible: readonly(isVisible),
    hasLoaded: readonly(hasLoaded)
  }
}

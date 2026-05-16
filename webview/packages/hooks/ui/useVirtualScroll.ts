/**
 * 虚拟滚动 Composable
 * 用于处理大量数据的列表渲染优化
 */

export interface VirtualScrollOptions {
  /** 每项高度（固定高度） */
  itemHeight: number
  /** 容器高度 */
  containerHeight: number
  /** 数据总数 */
  total: number
  /** 缓冲区大小（上下各保留的可见项数） */
  buffer?: number
  /** 是否启用动态高度 */
  dynamicHeight?: boolean
}

export interface VirtualScrollResult {
  /** 可见项的开始索引 */
  startIndex: ComputedRef<number>
  /** 可见项的结束索引 */
  endIndex: ComputedRef<number>
  /** 可见项列表 */
  visibleItems: ComputedRef<number[]>
  /** 总高度（用于占位） */
  totalHeight: ComputedRef<number>
  /** 偏移量（用于定位） */
  offsetY: ComputedRef<number>
  /** 更新滚动位置 */
  updateScroll: (scrollTop: number) => void
  /** 滚动到指定索引 */
  scrollToIndex: (index: number) => void
  /** 容器引用 */
  containerRef: Ref<HTMLElement | null>
}

/**
 * 虚拟滚动 Composable
 * @param options 配置选项
 */
export function useVirtualScroll(options: VirtualScrollOptions): VirtualScrollResult {
  const { itemHeight, containerHeight, total, buffer = 3 } = options

  const scrollTop = ref(0)
  const containerRef = ref<HTMLElement | null>(null)

  // 计算可见范围
  const visibleRange = computed(() => {
    const start = Math.floor(scrollTop.value / itemHeight)
    const end = Math.min(start + Math.ceil(containerHeight / itemHeight) + buffer * 2, total)

    return {
      start: Math.max(0, start - buffer),
      end
    }
  })

  // 可见项索引列表
  const visibleItems = computed(() => {
    const { start, end } = visibleRange.value
    return Array.from({ length: end - start }, (_, i) => start + i)
  })

  // 总高度
  const totalHeight = computed(() => {
    return total * itemHeight
  })

  // 偏移量
  const offsetY = computed(() => {
    return visibleRange.value.start * itemHeight
  })

  // 更新滚动位置
  const updateScroll = (newScrollTop: number) => {
    scrollTop.value = Math.max(0, newScrollTop)
  }

  // 滚动到指定索引
  const scrollToIndex = (index: number) => {
    if (index < 0 || index >= total) return

    const targetScrollTop = index * itemHeight
    if (containerRef.value) {
      containerRef.value.scrollTop = targetScrollTop
    } else {
      scrollTop.value = targetScrollTop
    }
  }

  // 处理滚动事件
  const handleScroll = (event: Event) => {
    const target = event.target as HTMLElement
    updateScroll(target.scrollTop)
  }

  // 绑定滚动事件
  onMounted(() => {
    if (containerRef.value) {
      containerRef.value.addEventListener('scroll', handleScroll, { passive: true })
    }
  })

  onUnmounted(() => {
    if (containerRef.value) {
      containerRef.value.removeEventListener('scroll', handleScroll)
    }
  })

  return {
    startIndex: computed(() => visibleRange.value.start),
    endIndex: computed(() => visibleRange.value.end),
    visibleItems,
    totalHeight,
    offsetY,
    updateScroll,
    scrollToIndex,
    containerRef
  } as VirtualScrollResult
}

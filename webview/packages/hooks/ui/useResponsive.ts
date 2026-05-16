/**
 * 响应式断点配置
 * 参考 Bootstrap 和主流设计系统
 */
export const BREAKPOINTS = {
  xs: 480, // Extra small screen / phone
  sm: 576, // Small screen / tablet
  md: 768, // Medium screen / desktop
  lg: 992, // Large screen / wide desktop
  xl: 1200, // Extra large screen / full hd
  xxl: 1600 // Extra extra large screen / large desktop
} as const

/**
 * 兼容旧版本的断点映射
 * @deprecated 使用 BREAKPOINTS.xs, BREAKPOINTS.sm 等替代
 */
export const LEGACY_BREAKPOINTS = {
  mobile: BREAKPOINTS.lg, // 移动端/平板端分界点 (992px)
  tablet: BREAKPOINTS.md, // 平板端/手机端分界点 (768px)
  phone: BREAKPOINTS.xs // 小屏手机分界点 (480px)
} as const

/**
 * 响应式检测 Composable
 * 提供窗口宽度、设备类型检测等功能
 */
export function useResponsive() {
  const windowWidth = ref(typeof window !== 'undefined' ? window.innerWidth : 1920)
  const windowHeight = ref(typeof window !== 'undefined' ? window.innerHeight : 1080)

  // 设备类型判断（基于新断点系统）
  const isXs = computed(() => windowWidth.value < BREAKPOINTS.xs)
  const isSm = computed(() => windowWidth.value >= BREAKPOINTS.xs && windowWidth.value < BREAKPOINTS.sm)
  const isMd = computed(() => windowWidth.value >= BREAKPOINTS.sm && windowWidth.value < BREAKPOINTS.md)
  const isLg = computed(() => windowWidth.value >= BREAKPOINTS.md && windowWidth.value < BREAKPOINTS.lg)
  const isXl = computed(() => windowWidth.value >= BREAKPOINTS.lg && windowWidth.value < BREAKPOINTS.xl)
  const isXxl = computed(() => windowWidth.value >= BREAKPOINTS.xl)

  // 兼容旧版本的设备类型判断
  const isMobile = computed(() => windowWidth.value < BREAKPOINTS.lg) // < 992px
  const isTablet = computed(() => windowWidth.value >= BREAKPOINTS.md && windowWidth.value < BREAKPOINTS.lg) // 768px - 991px
  const isPhone = computed(() => windowWidth.value < BREAKPOINTS.xs) // < 480px
  const isDesktop = computed(() => windowWidth.value >= BREAKPOINTS.lg) // >= 992px

  // 处理窗口大小变化
  const handleResize = () => {
    if (typeof window !== 'undefined') {
      windowWidth.value = window.innerWidth
      windowHeight.value = window.innerHeight
    }
  }

  // 监听窗口大小变化
  onMounted(() => {
    if (typeof window !== 'undefined') {
      window.addEventListener('resize', handleResize)
      // 初始化时也更新一次，确保 SSR 兼容
      handleResize()
    }
  })

  onBeforeUnmount(() => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('resize', handleResize)
    }
  })

  return {
    windowWidth,
    windowHeight,
    // 新断点系统
    isXs,
    isSm,
    isMd,
    isLg,
    isXl,
    isXxl,
    // 兼容旧版本
    isMobile,
    isTablet,
    isPhone,
    isDesktop,
    BREAKPOINTS,
    LEGACY_BREAKPOINTS
  }
}

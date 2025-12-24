/**
 * 响应式断点配置
 */
export const BREAKPOINTS = {
  mobile: 1024,  // 移动端/平板端分界点
  tablet: 768,   // 平板端/手机端分界点
  phone: 480     // 小屏手机分界点
} as const

/**
 * 响应式检测 Composable
 * 提供窗口宽度、设备类型检测等功能
 */
export function useResponsive() {
  const windowWidth = ref(typeof window !== 'undefined' ? window.innerWidth : 1920)
  const windowHeight = ref(typeof window !== 'undefined' ? window.innerHeight : 1080)

  // 设备类型判断
  const isMobile = computed(() => windowWidth.value <= BREAKPOINTS.mobile)
  const isTablet = computed(() => windowWidth.value <= BREAKPOINTS.tablet && windowWidth.value > BREAKPOINTS.phone)
  const isPhone = computed(() => windowWidth.value <= BREAKPOINTS.phone)
  const isDesktop = computed(() => windowWidth.value > BREAKPOINTS.mobile)

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
    isMobile,
    isTablet,
    isPhone,
    isDesktop,
    BREAKPOINTS
  }
}


import type { ComponentInternalInstance } from 'vue'
import { applyColorPalettes } from '@/utils/color'
import { toggleCssDarkMode } from '@/utils/theme'

type Theme = 'light' | 'dark' | 'auto'

interface CustomColors {
  primary?: string
  secondary?: string
  success?: string
  warning?: string
  danger?: string
  info?: string
}

interface ThemePreset {
  name: string
  desc: string
  theme: Theme
  grayscale: boolean
  colourWeakness: boolean
  colors: CustomColors
}

const theme = ref<Theme>('light')
const isDark = ref(false)
const customColors = ref<CustomColors>({})
const grayscale = ref(false)
const colourWeakness = ref(false)

/**
 * 主题管理 Composable
 */
export function useTheme() {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  // 从 localStorage 加载主题设置
  const loadTheme = () => {
    const saved = localStorage.getItem('theme') as Theme | null
    if (saved && ['light', 'dark', 'auto'].includes(saved)) {
      theme.value = saved
    } else {
      // 默认跟随系统
      theme.value = 'auto'
    }
    applyTheme()
  }

  // 应用主题
  const applyTheme = () => {
    if (theme.value === 'auto') {
      // 跟随系统主题
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      isDark.value = prefersDark
    } else {
      isDark.value = theme.value === 'dark'
    }
    
    // 使用 html.dark 类切换深色模式
    toggleCssDarkMode(isDark.value)
    
    // 应用自定义颜色
    applyCustomColors()
  }

  // 应用自定义颜色（使用颜色调色板）
  const applyCustomColors = () => {
    // 使用新的颜色调色板系统
    const colorsToApply: Record<string, string> = {}
    
    if (customColors.value.primary) colorsToApply.primary = customColors.value.primary
    if (customColors.value.secondary) colorsToApply.secondary = customColors.value.secondary
    if (customColors.value.success) colorsToApply.success = customColors.value.success
    if (customColors.value.warning) colorsToApply.warning = customColors.value.warning
    if (customColors.value.danger) colorsToApply.danger = customColors.value.danger
    if (customColors.value.info) colorsToApply.info = customColors.value.info
    
    if (Object.keys(colorsToApply).length > 0) {
      applyColorPalettes(colorsToApply)
    }
  }
  
  // 切换辅助颜色模式（灰度、色弱）
  const toggleAuxiliaryColorModes = () => {
    const htmlElement = document.documentElement
    const filters: string[] = []
    
    if (grayscale.value) {
      filters.push('grayscale(100%)')
    }
    
    if (colourWeakness.value) {
      filters.push('invert(80%)')
    }
    
    htmlElement.style.filter = filters.join(' ')
  }


  // 默认颜色值
  const defaultColors: CustomColors = {
    primary: '#2563eb',
    success: '#10b981',
    warning: '#f59e0b',
    danger: '#ef4444'
  }

  // 加载自定义颜色
  const loadCustomColors = () => {
    try {
      const saved = localStorage.getItem('customColors')
      if (saved) {
        const parsed = JSON.parse(saved)
        // 合并默认颜色，确保所有颜色都有值
        customColors.value = { ...defaultColors, ...parsed }
      } else {
        // 如果没有保存的颜色，使用默认值
        customColors.value = { ...defaultColors }
      }
    } catch (error) {
      proxy?.$log.error('加载自定义颜色失败:', error)
      // 出错时使用默认值
      customColors.value = { ...defaultColors }
    }
  }
  
  // 加载灰度模式
  const loadGrayscale = () => {
    const saved = localStorage.getItem('grayscale')
    if (saved === 'true') {
      grayscale.value = true
    }
  }
  
  // 加载色弱模式
  const loadColourWeakness = () => {
    const saved = localStorage.getItem('colourWeakness')
    if (saved === 'true') {
      colourWeakness.value = true
    }
  }
  
  // 设置灰度模式
  const setGrayscale = (value: boolean) => {
    grayscale.value = value
    localStorage.setItem('grayscale', String(value))
    toggleAuxiliaryColorModes()
  }
  
  // 设置色弱模式
  const setColourWeakness = (value: boolean) => {
    colourWeakness.value = value
    localStorage.setItem('colourWeakness', String(value))
    toggleAuxiliaryColorModes()
  }

  // 设置自定义颜色
  const setCustomColors = (colors: CustomColors) => {
    customColors.value = { ...customColors.value, ...colors }
    localStorage.setItem('customColors', JSON.stringify(customColors.value))
    applyCustomColors()
  }

  // 重置自定义颜色
  const resetCustomColors = () => {
    customColors.value = {}
    localStorage.removeItem('customColors')
    // 移除自定义 CSS 变量
    const root = document.documentElement
    root.style.removeProperty('--primary-color')
    root.style.removeProperty('--primary-hover')
    root.style.removeProperty('--secondary-color')
    root.style.removeProperty('--success-color')
    root.style.removeProperty('--warning-color')
    root.style.removeProperty('--danger-color')
    root.style.removeProperty('--info-color')
  }

  // 切换主题（在 light 和 dark 之间切换，跳过 auto）
  const toggleTheme = () => {
    // 如果当前是 auto，根据系统主题决定切换到 light 还是 dark
    if (theme.value === 'auto') {
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      theme.value = prefersDark ? 'light' : 'dark'
    } else if (theme.value === 'light') {
      theme.value = 'dark'
    } else {
      theme.value = 'light'
    }
    localStorage.setItem('theme', theme.value)
    applyTheme()
  }

  // 设置主题
  const setTheme = (newTheme: Theme) => {
    theme.value = newTheme
    localStorage.setItem('theme', newTheme)
    applyTheme()
  }

  // 监听系统主题变化
  let mediaQuery: MediaQueryList | null = null
  const handleSystemThemeChange = () => {
    if (theme.value === 'auto') {
      applyTheme()
    }
  }

  // 应用主题预设
  const applyPreset = (preset: ThemePreset) => {
    setTheme(preset.theme)
    setGrayscale(preset.grayscale)
    setColourWeakness(preset.colourWeakness)
    setCustomColors(preset.colors)
  }
  
  // 初始化
  onMounted(() => {
    loadTheme()
    loadCustomColors()
    loadGrayscale()
    loadColourWeakness()
    toggleAuxiliaryColorModes()
    if (typeof window !== 'undefined') {
      mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      mediaQuery.addEventListener('change', handleSystemThemeChange)
    }
  })

  onBeforeUnmount(() => {
    if (mediaQuery) {
      mediaQuery.removeEventListener('change', handleSystemThemeChange)
    }
  })

  return {
    theme: readonly(theme),
    isDark: readonly(isDark),
    customColors: readonly(customColors),
    grayscale: readonly(grayscale),
    colourWeakness: readonly(colourWeakness),
    toggleTheme,
    setTheme,
    applyTheme,
    setCustomColors,
    resetCustomColors,
    setGrayscale,
    setColourWeakness,
    applyPreset
  }
}

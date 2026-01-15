import { defineStore } from 'pinia'
import { ref } from 'vue'
import { StoreId } from '@/enums/StoreId'

export type LayoutMode =
  | 'vertical'
  | 'horizontal'
  | 'vertical-mix'
  | 'vertical-hybrid-header-first'
  | 'top-hybrid-sidebar-first'
  | 'top-hybrid-header-first'

export const useLayoutStore = defineStore(StoreId.Layout, () => {
  // 布局模式
  const layoutMode = ref<LayoutMode>('vertical')

  // 侧边栏宽度
  const sidebarWidth = ref(240)

  // 侧边栏是否折叠
  const sidebarCollapsed = ref(false)

  // 标签页是否显示（默认关闭）
  const tagsViewVisible = ref(false)

  /**
   * 设置布局模式
   */
  function setLayoutMode(mode: LayoutMode) {
    layoutMode.value = mode
    localStorage.setItem('layoutMode', mode)
  }

  /**
   * 设置侧边栏宽度
   */
  function setSidebarWidth(width: number) {
    sidebarWidth.value = width
    localStorage.setItem('sidebarWidth', String(width))
  }

  /**
   * 切换侧边栏折叠状态
   */
  function toggleSidebarCollapsed() {
    sidebarCollapsed.value = !sidebarCollapsed.value
    localStorage.setItem('sidebarCollapsed', String(sidebarCollapsed.value))
  }

  /**
   * 设置侧边栏折叠状态
   */
  function setSidebarCollapsed(collapsed: boolean) {
    sidebarCollapsed.value = collapsed
    localStorage.setItem('sidebarCollapsed', String(collapsed))
  }

  /**
   * 设置标签页显示状态
   */
  function setTagsViewVisible(visible: boolean) {
    tagsViewVisible.value = visible
    localStorage.setItem('tagsViewVisible', String(visible))
  }

  /**
   * 初始化布局设置
   */
  function initLayout() {
    const savedMode = localStorage.getItem('layoutMode') as LayoutMode | null
    const validModes: LayoutMode[] = [
      'vertical',
      'horizontal',
      'vertical-mix',
      'vertical-hybrid-header-first',
      'top-hybrid-sidebar-first',
      'top-hybrid-header-first'
    ]
    if (savedMode && validModes.includes(savedMode)) {
      layoutMode.value = savedMode
    }

    const savedWidth = localStorage.getItem('sidebarWidth')
    if (savedWidth) {
      sidebarWidth.value = Number(savedWidth)
    }

    const savedCollapsed = localStorage.getItem('sidebarCollapsed')
    if (savedCollapsed) {
      sidebarCollapsed.value = savedCollapsed === 'true'
    }

    const savedTagsViewVisible = localStorage.getItem('tagsViewVisible')
    if (savedTagsViewVisible) {
      tagsViewVisible.value = savedTagsViewVisible === 'true'
    }
  }

  /**
   * 获取布局配置
   */
  function getLayoutConfig() {
    return {
      layoutMode: layoutMode.value,
      sidebarWidth: sidebarWidth.value,
      sidebarCollapsed: sidebarCollapsed.value,
      tagsViewVisible: tagsViewVisible.value
    }
  }

  /**
   * 设置布局配置
   */
  function setLayoutConfig(config: {
    layoutMode?: LayoutMode
    sidebarWidth?: number
    sidebarCollapsed?: boolean
    tagsViewVisible?: boolean
  }) {
    if (config.layoutMode !== undefined) {
      setLayoutMode(config.layoutMode)
    }
    if (config.sidebarWidth !== undefined) {
      setSidebarWidth(config.sidebarWidth)
    }
    if (config.sidebarCollapsed !== undefined) {
      setSidebarCollapsed(config.sidebarCollapsed)
    }
    if (config.tagsViewVisible !== undefined) {
      setTagsViewVisible(config.tagsViewVisible)
    }
  }

  /**
   * 导出布局配置为 JSON
   */
  function exportLayoutConfig(): string {
    return JSON.stringify(getLayoutConfig(), null, 2)
  }

  /**
   * 导入布局配置
   */
  function importLayoutConfig(json: string): boolean {
    try {
      const config = JSON.parse(json)
      const validModes: LayoutMode[] = [
        'vertical',
        'horizontal',
        'vertical-mix',
        'vertical-hybrid-header-first',
        'top-hybrid-sidebar-first',
        'top-hybrid-header-first'
      ]
      if (config.layoutMode && validModes.includes(config.layoutMode)) {
        setLayoutConfig(config)
        return true
      }
      return false
    } catch (error) {
      console.error('导入布局配置失败:', error)
      return false
    }
  }

  /**
   * 重置布局配置为默认值
   */
  function resetLayoutConfig() {
    layoutMode.value = 'vertical'
    sidebarWidth.value = 240
    sidebarCollapsed.value = false
    tagsViewVisible.value = false
    localStorage.removeItem('layoutMode')
    localStorage.removeItem('sidebarWidth')
    localStorage.removeItem('sidebarCollapsed')
    localStorage.removeItem('tagsViewVisible')
  }

  return {
    layoutMode,
    sidebarWidth,
    sidebarCollapsed,
    tagsViewVisible,
    setLayoutMode,
    setSidebarWidth,
    toggleSidebarCollapsed,
    setSidebarCollapsed,
    setTagsViewVisible,
    initLayout,
    getLayoutConfig,
    setLayoutConfig,
    exportLayoutConfig,
    importLayoutConfig,
    resetLayoutConfig
  }
})

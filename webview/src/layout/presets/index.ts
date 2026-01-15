/**
 * 布局预设
 */
import type { LayoutMode } from '@/stores/layout'

export interface LayoutPreset {
  name: string
  mode: LayoutMode
  desc: string
  icon?: string
  recommended?: boolean
}

export const layoutPresets: LayoutPreset[] = [
  {
    name: 'vertical',
    mode: 'vertical',
    desc: '传统的左侧边栏布局，适合大多数场景',
    recommended: true
  },
  {
    name: 'horizontal',
    mode: 'horizontal',
    desc: '顶部导航栏布局，适合宽屏显示'
  },
  {
    name: 'vertical-mix',
    mode: 'vertical-mix',
    desc: '混合布局，结合垂直和水平布局的优点'
  },
  {
    name: 'vertical-hybrid-header-first',
    mode: 'vertical-hybrid-header-first',
    desc: '垂直混合布局，头部区域突出显示'
  },
  {
    name: 'top-hybrid-sidebar-first',
    mode: 'top-hybrid-sidebar-first',
    desc: '顶部布局，侧边栏在内容区域左侧'
  },
  {
    name: 'top-hybrid-header-first',
    mode: 'top-hybrid-header-first',
    desc: '顶部布局，头部区域突出显示'
  }
]

export default layoutPresets

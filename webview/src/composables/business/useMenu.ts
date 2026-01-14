import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAdmin, useI18n } from '@/composables'

/**
 * 菜单项接口
 */
export interface MenuItem {
  path?: string
  label: string
  icon: string
  isDivider?: boolean
  hidden?: boolean
}

/**
 * 菜单分组接口
 */
export interface MenuGroup {
  title?: string
  items: MenuItem[]
}

/**
 * 菜单 Composable
 * 提供菜单数据和处理逻辑
 */
export function useMenu() {
  const route = useRoute()
  const { isAdmin } = useAdmin()
  const { t } = useI18n()

  // 定义菜单分组
  const menuGroups = computed<MenuGroup[]>(() => {
    const groups: MenuGroup[] = [
      {
        title: t('menu.groups.main'),
        items: [
          { path: '/files', label: t('menu.files'), icon: 'Folder' },
          { path: '/shares', label: t('menu.shares'), icon: 'Share' },
          { path: '/offline', label: t('menu.offline'), icon: 'Download' },
          { path: '/tasks', label: t('menu.tasks'), icon: 'List' },
          { path: '/trash', label: t('menu.trash'), icon: 'Delete' }
        ]
      },
      {
        title: t('menu.groups.public'),
        items: [{ path: '/square', label: t('menu.square'), icon: 'Grid' }]
      }
    ]

    if (isAdmin.value) {
      groups.push({
        title: t('menu.groups.admin'),
        items: [{ path: '/admin', label: t('menu.admin'), icon: 'Setting' }]
      })
    }

    return groups
  })

  // 获取所有菜单项（扁平化，用于水平菜单）
  const menuItems = computed<MenuItem[]>(() => {
    return menuGroups.value.flatMap(group => group.items).filter(item => !item.hidden && item.path)
  })

  // 当前路由
  const currentRoute = computed(() => route.path)

  return {
    menuGroups,
    menuItems,
    currentRoute
  }
}

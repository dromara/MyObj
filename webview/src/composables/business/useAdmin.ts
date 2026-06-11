import { computed } from 'vue'
import { useUserStore } from '@/stores/user'

/** 管理员组 ID 常量 - 与后端保持一致 */
export const ADMIN_GROUP_ID = 1

/**
 * 管理员权限检查 Composable
 */
export function useAdmin() {
  const userStore = useUserStore()

  /**
   * 检查当前用户是否为管理员
   * 通过常量 ADMIN_GROUP_ID 判断，避免魔法数字
   */
  const isAdmin = computed(() => {
    if (!userStore.userInfo) {
      return false
    }
    return userStore.userInfo.group_id === ADMIN_GROUP_ID
  })

  /**
   * 检查是否有管理权限
   */
  const hasAdminAccess = computed(() => {
    return isAdmin.value
  })

  return {
    isAdmin,
    hasAdminAccess
  }
}

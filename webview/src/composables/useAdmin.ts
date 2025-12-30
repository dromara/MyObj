import { computed } from 'vue'
import { useUserStore } from '@/stores/user'

/**
 * 管理员权限检查 Composable
 */
export function useAdmin() {
  const userStore = useUserStore()
  
  /**
   * 检查当前用户是否为管理员
   * 通常 group_id = 1 为管理员组
   */
  const isAdmin = computed(() => {
    if (!userStore.userInfo) {
      return false
    }
    // group_id = 1 通常为管理员组
    return userStore.userInfo.group_id === 1
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


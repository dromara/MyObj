import { defineStore } from 'pinia'
import { useUserStore } from './user'
import { cache, logger } from '@myobj/shared'
import { StoreId } from '@myobj/shared'
import { uploadTaskManager } from '@/utils/file/uploadTaskManager'

/**
 * 认证 Store
 */
export const useAuthStore = defineStore(StoreId.Auth, () => {
  const userStore = useUserStore()

  // 状态
  const token = ref<string | null>(null)

  // Getters
  const isAuthenticated = computed(() => {
    return token.value !== null && userStore.isLoggedIn
  })

  /**
   * 从 localStorage 加载 token
   */
  const loadTokenFromCache = () => {
    try {
      const cached = cache.local.get('token')
      if (cached) {
        token.value = cached
      }
    } catch (error) {
      logger.error('加载 token 失败:', error)
    }
  }

  /**
   * 设置 token
   *
   * 安全说明：当前使用 localStorage 存储 token，存在以下风险：
   * - XSS 攻击可通过 document.cookie 或 localStorage 读取 token
   * - 建议生产环境使用 httpOnly cookie 存储 token，以防止 JavaScript 直接访问
   * - 如需保留 localStorage 方案，请确保 CSP 和 XSS 防护到位
   */
  const setToken = (newToken: string) => {
    token.value = newToken
    try {
      cache.local.set('token', newToken)
    } catch (error) {
      logger.error('保存 token 到缓存失败:', error)
    }
  }

  /**
   * 清除 token
   */
  const clearToken = () => {
    token.value = null
    try {
      cache.local.remove('token')
    } catch (error) {
      logger.error('清除 token 失败:', error)
    }
  }

  /**
   * 登录
   */
  const login = (newToken: string, userInfo: any) => {
    setToken(newToken)
    userStore.setUserInfo(userInfo)
    // 初始化上传任务管理器（切换到新用户的任务）
    uploadTaskManager.init()
    
    // 启动剪贴板监听（如果已启用）
    // 注意：这里不直接调用，而是通过 composable 的 watch 自动启动
    // 因为 composable 需要在组件中初始化
  }

  /**
   * 登出
   */
  const logout = () => {
    // 清空当前用户的上传任务
    uploadTaskManager.clearCurrentUserTasks()
    clearToken()
    userStore.clearUserInfo()
    // 剪贴板监听会通过 watch 自动停止
  }

  // 初始化：从缓存加载
  loadTokenFromCache()

  return {
    // 状态
    token,
    // Getters
    isAuthenticated,
    // Actions
    setToken,
    clearToken,
    login,
    logout,
    loadTokenFromCache
  }
})

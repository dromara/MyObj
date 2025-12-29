import { defineStore } from 'pinia'
import { useUserStore } from './user'
import cache from '@/plugins/cache'
import logger from '@/plugins/logger'

/**
 * 认证 Store
 */
export const useAuthStore = defineStore('auth', () => {
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
  }

  /**
   * 登出
   */
  const logout = () => {
    clearToken()
    userStore.clearUserInfo()
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


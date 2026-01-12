import { defineStore } from 'pinia'
import type { UserInfo } from '@/types'
import { getUserInfo } from '@/api/user'
import cache from '@/plugins/cache'
import logger from '@/plugins/logger'

interface StorageInfo {
  used: number
  total: number
  percentage: number
  isUnlimited: boolean
}

/**
 * 用户信息 Store
 */
export const useUserStore = defineStore('user', () => {
  // 状态
  const userInfo = ref<UserInfo | null>(null)
  const storageInfo = ref<StorageInfo>({
    used: 0,
    total: 0,
    percentage: 0,
    isUnlimited: false
  })

  // Getters
  const isLoggedIn = computed(() => userInfo.value !== null)
  const username = computed(() => userInfo.value?.user_name || '')
  const nickname = computed(() => userInfo.value?.name || '')
  const email = computed(() => userInfo.value?.email || '')
  const phone = computed(() => userInfo.value?.phone || '')

  /**
   * 从 localStorage 加载用户信息
   */
  const loadUserInfoFromCache = () => {
    try {
      const cached = cache.local.getJSON('userInfo')
      if (cached) {
        userInfo.value = cached
        updateStorageInfo(userInfo.value)
      }
    } catch (error) {
      logger.error('加载用户信息失败:', error)
    }
  }

  /**
   * 更新存储空间信息
   */
  const updateStorageInfo = (info: UserInfo | null) => {
    if (!info) {
      storageInfo.value = {
        used: 0,
        total: 0,
        percentage: 0,
        isUnlimited: false
      }
      return
    }

    // 基于 UserInfo 接口映射: space (总容量), free_space (剩余空间)
    if (info.space !== undefined) {
      const total = Number(info.space)
      const free = Number(info.free_space || 0)
      let used = 0
      
      // 如果有总容量和剩余空间，计算已用空间
      if (info.free_space !== undefined) {
        used = total - free
      } else if ((info as any).used !== undefined) {
        used = Number((info as any).used)
      }

      // 将 0 或 -1 视为无限容量
      storageInfo.value.isUnlimited = total === 0 || total === -1
      storageInfo.value.total = total
      storageInfo.value.used = used > 0 ? used : 0
      
      // 重新计算百分比
      if (!storageInfo.value.isUnlimited && storageInfo.value.total > 0) {
        storageInfo.value.percentage = Math.ceil(
          (storageInfo.value.used / storageInfo.value.total) * 100
        )
      } else {
        storageInfo.value.percentage = 0
      }
    } else {
      const capacity = (info as any).capacity || (info as any).storage_limit
      if (capacity !== undefined) {
        const capNum = Number(capacity)
        storageInfo.value.isUnlimited = capNum === 0 || capNum === -1
        storageInfo.value.total = capNum
        storageInfo.value.used = Number((info as any).used || (info as any).used_storage || 0)
        
        if (!storageInfo.value.isUnlimited && storageInfo.value.total > 0) {
          storageInfo.value.percentage = Math.ceil(
            (storageInfo.value.used / storageInfo.value.total) * 100
          )
        } else {
          storageInfo.value.percentage = 0
        }
      }
    }
  }

  /**
   * 设置用户信息
   */
  const setUserInfo = (info: UserInfo) => {
    userInfo.value = info
    updateStorageInfo(info)
    // 同步到 localStorage
    try {
      cache.local.setJSON('userInfo', info)
    } catch (error) {
      logger.error('保存用户信息到缓存失败:', error)
    }
  }

  /**
   * 更新用户信息（部分更新）
   */
  const updateUserInfo = (updates: Partial<UserInfo>) => {
    if (userInfo.value) {
      userInfo.value = { ...userInfo.value, ...updates }
      updateStorageInfo(userInfo.value)
      // 同步到 localStorage
      try {
        cache.local.setJSON('userInfo', userInfo.value)
      } catch (error) {
        logger.error('更新用户信息到缓存失败:', error)
      }
    }
  }

  /**
   * 从服务器获取用户信息
   * 如果获取失败，保留现有的用户信息，避免菜单消失等问题
   */
  const fetchUserInfo = async () => {
    // 保存当前用户信息作为后备
    const currentUserInfo = userInfo.value
    
    try {
      const res = await getUserInfo()
      if (res.code === 200 && res.data) {
        setUserInfo(res.data)
        return res.data
      } else {
        // API 返回失败，保留现有用户信息
        logger.warn('获取用户信息失败，保留现有用户信息')
        return currentUserInfo
      }
    } catch (error) {
      // API 调用失败，保留现有用户信息，避免菜单消失
      logger.error('获取用户信息失败:', error)
      return currentUserInfo
    }
  }

  /**
   * 清除用户信息
   */
  const clearUserInfo = () => {
    userInfo.value = null
    storageInfo.value = {
      used: 0,
      total: 0,
      percentage: 0,
      isUnlimited: false
    }
    cache.local.remove('userInfo')
  }

  // 初始化：从缓存加载
  loadUserInfoFromCache()

  return {
    // 状态
    userInfo,
    storageInfo,
    // Getters
    isLoggedIn,
    username,
    nickname,
    email,
    phone,
    // Actions
    setUserInfo,
    updateUserInfo,
    fetchUserInfo,
    clearUserInfo,
    loadUserInfoFromCache,
    updateStorageInfo
  }
})


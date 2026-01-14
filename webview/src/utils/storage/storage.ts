/**
 * 本地存储工具函数
 * 提供类型安全的 localStorage 和 sessionStorage 操作
 */

type StorageType = 'local' | 'session'

/**
 * 获取存储对象
 */
function getStorage(type: StorageType): Storage {
  return type === 'local' ? localStorage : sessionStorage
}

/**
 * 设置存储项
 */
export function setStorageItem<T>(key: string, value: T, type: StorageType = 'local'): boolean {
  try {
    const storage = getStorage(type)
    const serialized = JSON.stringify(value)
    storage.setItem(key, serialized)
    return true
  } catch (error) {
    console.error(`Failed to set storage item "${key}":`, error)
    return false
  }
}

/**
 * 获取存储项
 */
export function getStorageItem<T>(key: string, defaultValue?: T, type: StorageType = 'local'): T | null {
  try {
    const storage = getStorage(type)
    const item = storage.getItem(key)
    if (item === null) {
      return defaultValue ?? null
    }
    return JSON.parse(item) as T
  } catch (error) {
    console.error(`Failed to get storage item "${key}":`, error)
    return defaultValue ?? null
  }
}

/**
 * 删除存储项
 */
export function removeStorageItem(key: string, type: StorageType = 'local'): boolean {
  try {
    const storage = getStorage(type)
    storage.removeItem(key)
    return true
  } catch (error) {
    console.error(`Failed to remove storage item "${key}":`, error)
    return false
  }
}

/**
 * 清空存储
 */
export function clearStorage(type: StorageType = 'local'): boolean {
  try {
    const storage = getStorage(type)
    storage.clear()
    return true
  } catch (error) {
    console.error(`Failed to clear storage:`, error)
    return false
  }
}

/**
 * 获取所有存储键
 */
export function getStorageKeys(type: StorageType = 'local'): string[] {
  try {
    const storage = getStorage(type)
    return Object.keys(storage)
  } catch (error) {
    console.error(`Failed to get storage keys:`, error)
    return []
  }
}

/**
 * 检查存储项是否存在
 */
export function hasStorageItem(key: string, type: StorageType = 'local'): boolean {
  try {
    const storage = getStorage(type)
    return storage.getItem(key) !== null
  } catch (error) {
    return false
  }
}

/**
 * 获取存储大小（字节）
 */
export function getStorageSize(type: StorageType = 'local'): number {
  try {
    const storage = getStorage(type)
    let total = 0
    for (const key in storage) {
      if (storage.hasOwnProperty(key)) {
        total += storage[key].length + key.length
      }
    }
    return total
  } catch (error) {
    console.error(`Failed to get storage size:`, error)
    return 0
  }
}

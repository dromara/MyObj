import type { PiniaPluginContext } from 'pinia'
import { StoreId } from '@/enums/StoreId'

/**
 * 深拷贝对象（简单实现，用于 store 状态重置）
 */
function deepClone<T>(obj: T): T {
  if (obj === null || typeof obj !== 'object') {
    return obj
  }

  if (obj instanceof Date) {
    return new Date(obj.getTime()) as T
  }

  if (obj instanceof Array) {
    return obj.map(item => deepClone(item)) as T
  }

  if (typeof obj === 'object') {
    const cloned = {} as T
    for (const key in obj) {
      if (Object.prototype.hasOwnProperty.call(obj, key)) {
        cloned[key] = deepClone(obj[key])
      }
    }
    return cloned
  }

  return obj
}

/**
 * Pinia 插件：重置使用 setup 语法定义的 store 状态
 *
 * 这个插件为使用 setup 语法定义的 store 提供 $reset 方法
 * 它会保存 store 的初始状态，并在调用 $reset 时恢复到初始状态
 *
 * @param context Pinia 插件上下文
 */
export function resetSetupStore(context: PiniaPluginContext) {
  const setupStoreIds = Object.values(StoreId) as string[]

  // 只处理我们定义的 store
  if (setupStoreIds.includes(context.store.$id)) {
    const { $state } = context.store

    // 深拷贝初始状态
    const defaultStore = deepClone($state)

    // 重写 $reset 方法
    context.store.$reset = () => {
      context.store.$patch(defaultStore)
    }
  }
}

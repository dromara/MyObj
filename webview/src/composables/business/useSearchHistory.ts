/**
 * 搜索历史管理 Composable
 * 提供搜索历史的管理功能
 */
import type { ComponentInternalInstance } from 'vue'
import cache from '@/plugins/cache'

const MAX_HISTORY = 10
const HISTORY_KEY = 'searchHistory'

export function useSearchHistory() {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const searchHistory = ref<string[]>([])

  // 加载搜索历史
  const loadHistory = () => {
    try {
      const history = cache.local.getJSON(HISTORY_KEY)
      if (Array.isArray(history)) {
        searchHistory.value = history as string[]
      }
    } catch (error) {
      proxy?.$log.error('加载搜索历史失败:', error)
    }
  }

  // 添加搜索历史
  const addHistory = (keyword: string) => {
    if (!keyword.trim()) return

    // 移除重复项
    const filtered = searchHistory.value.filter(item => item !== keyword.trim())
    // 添加到开头
    filtered.unshift(keyword.trim())
    // 限制数量
    searchHistory.value = filtered.slice(0, MAX_HISTORY)

    // 保存到 localStorage
    try {
      cache.local.setJSON(HISTORY_KEY, searchHistory.value)
    } catch (error) {
      proxy?.$log.error('保存搜索历史失败:', error)
    }
  }

  // 清除搜索历史
  const clearHistory = () => {
    searchHistory.value = []
    cache.local.remove(HISTORY_KEY)
  }

  // 删除单条历史
  const removeHistory = (keyword: string) => {
    searchHistory.value = searchHistory.value.filter(item => item !== keyword)
    try {
      cache.local.setJSON(HISTORY_KEY, searchHistory.value)
    } catch (error) {
      proxy?.$log.error('删除搜索历史失败:', error)
    }
  }

  // 获取匹配的搜索建议
  const getSuggestions = (keyword: string, maxResults: number = 5) => {
    if (!keyword.trim()) {
      return searchHistory.value.slice(0, maxResults)
    }

    const lowerKeyword = keyword.toLowerCase()
    return searchHistory.value.filter(item => item.toLowerCase().includes(lowerKeyword)).slice(0, maxResults)
  }

  // 初始化
  loadHistory()

  return {
    searchHistory: readonly(searchHistory),
    addHistory,
    clearHistory,
    removeHistory,
    getSuggestions
  }
}

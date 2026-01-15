import type { ComponentInternalInstance } from 'vue'
import type { FileSearchParams, SearchResponse } from '@/api/file'
import cache from '@/plugins/cache'

const MAX_HISTORY = 10
const HISTORY_KEY = 'searchHistory'

/**
 * 通用搜索 composable（包含搜索历史管理）
 * @param searchApi 搜索 API 函数
 * @param transformResult 结果转换函数
 * @param onClear 清空搜索时的回调
 * @param enableHistory 是否启用搜索历史，默认 true
 */
export function useSearch<T>(
  searchApi: (params: FileSearchParams) => Promise<SearchResponse>,
  transformResult: (files: any[]) => T[],
  onClear?: () => void,
  enableHistory = true
) {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance

  const searchKeyword = ref('')
  const isSearching = ref(false)
  const searchResults = ref<T[]>([])
  const total = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(20)
  const searchHistory = ref<string[]>([])

  // 加载搜索历史
  const loadHistory = () => {
    if (!enableHistory) return
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
    if (!enableHistory || !keyword.trim()) return

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
    if (!enableHistory) return []
    if (!keyword.trim()) {
      return searchHistory.value.slice(0, maxResults)
    }

    const lowerKeyword = keyword.toLowerCase()
    return searchHistory.value.filter(item => item.toLowerCase().includes(lowerKeyword)).slice(0, maxResults)
  }

  // 执行搜索
  const performSearch = async (keyword: string, pageNum: number = 1, pageSizeNum: number = 20) => {
    if (!keyword.trim()) {
      // 如果关键词为空，清空搜索结果
      clearSearchResults()
      if (onClear) {
        onClear()
      }
      return
    }

    // 添加到搜索历史
    if (enableHistory) {
      addHistory(keyword)
    }

    isSearching.value = true
    try {
      const params: FileSearchParams = {
        keyword: keyword.trim(),
        page: pageNum,
        pageSize: pageSizeNum
      }

      const res = await searchApi(params)

      if (res.code === 200 && res.data) {
        searchResults.value = transformResult(res.data.files)
        total.value = res.data.total
        currentPage.value = pageNum
        pageSize.value = pageSizeNum
      } else {
        proxy?.$modal.msgError(res.message || '搜索失败')
        clearSearchResults()
      }
    } catch (error) {
      proxy?.$modal.msgError('搜索文件失败')
      proxy?.$log.error(error)
      clearSearchResults()
    } finally {
      isSearching.value = false
    }
  }

  // 清空搜索结果
  const clearSearchResults = () => {
    searchResults.value = []
    total.value = 0
    currentPage.value = 1
    pageSize.value = 20
  }

  // 清空搜索
  const clearSearch = () => {
    searchKeyword.value = ''
    clearSearchResults()
    if (onClear) {
      onClear()
    }
  }

  // 是否有搜索关键词
  const hasSearchKeyword = computed(() => {
    return searchKeyword.value.trim().length > 0
  })

  // 初始化
  if (enableHistory) {
    loadHistory()
  }

  return {
    searchKeyword,
    isSearching,
    searchResults,
    total,
    currentPage,
    pageSize,
    searchHistory: readonly(searchHistory),
    performSearch,
    clearSearch,
    clearSearchResults,
    hasSearchKeyword,
    // 搜索历史相关方法
    addHistory,
    clearHistory,
    removeHistory,
    getSuggestions
  }
}

import type { FileSearchParams, SearchResponse } from '@/api/file'

/**
 * 通用搜索 composable
 * @param searchApi 搜索 API 函数
 * @param transformResult 结果转换函数
 * @param onClear 清空搜索时的回调
 */
export function useSearch<T>(
  searchApi: (params: FileSearchParams) => Promise<SearchResponse>,
  transformResult: (files: any[]) => T[],
  onClear?: () => void
) {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance

  const searchKeyword = ref('')
  const isSearching = ref(false)
  const searchResults = ref<T[]>([])
  const total = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(20)

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

  return {
    searchKeyword,
    isSearching,
    searchResults,
    total,
    currentPage,
    pageSize,
    performSearch,
    clearSearch,
    clearSearchResults,
    hasSearchKeyword
  }
}


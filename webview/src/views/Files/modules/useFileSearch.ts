import { searchUserFiles } from '@/api/file'
import { useSearch } from '@/composables/useSearch'
import type { FileListResponse, FileItem } from '@/types'

/**
 * 文件搜索 composable（用户文件）
 */
export function useFileSearch() {
  // 结果转换函数：将后端返回的文件转换为 FileItem 格式
  const transformResult = (files: any[]): FileItem[] => {
    return files.map((file: any) => ({
      file_id: file.uf_id || file.id || '', // 优先使用 uf_id
      file_name: file.file_name || file.name || '', // 优先使用 file_name（用户文件名）
      file_size: file.size || 0,
      mime_type: file.mime || '',
      created_at: file.created_at || file.createdAt || '',
      is_enc: file.is_enc || false,
      has_thumbnail: (file.thumbnail_img && file.thumbnail_img !== '') || false,
      public: file.public || file.isPublic || false
    }))
  }

  const search = useSearch<FileItem>(searchUserFiles, transformResult)

  // 将搜索结果包装为 FileListResponse 格式（兼容现有代码）
  const searchResults = computed<FileListResponse>(() => ({
    breadcrumbs: [],
    current_path: '',
    folders: [],
    files: search.searchResults.value,
    total: search.total.value,
    page: search.currentPage.value,
    page_size: search.pageSize.value
  }))

  return {
    searchKeyword: search.searchKeyword,
    isSearching: search.isSearching,
    searchResults,
    performSearch: search.performSearch,
    clearSearch: search.clearSearch,
    hasSearchKeyword: search.hasSearchKeyword
  }
}


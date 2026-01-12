import { getFileList, getThumbnail } from '@/api/file'
import { useI18n } from '@/composables/useI18n'
import type { FileListResponse } from '@/types'

export function useFileList() {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const router = useRouter()
  const route = useRoute()
  const { t } = useI18n()

  const fileListData = ref<FileListResponse>({
    breadcrumbs: [],
    current_path: '0',
    folders: [],
    files: [],
    total: 0,
    page: 1,
    page_size: 20
  })

  const currentPage = ref(1)
  const pageSize = ref(20)
  const currentPath = ref<string>('')
  const thumbnailCache = ref<Map<string, string>>(new Map())
  const loading = ref(false)

  const breadcrumbs = computed(() => fileListData.value.breadcrumbs)

  const formatBreadcrumbName = (name: string): string => {
    if (!name) return ''
    let formatted = name.replace(/^\/+/, '')
    if (formatted === 'home' || formatted === '') {
      return t('files.home')
    }
    return formatted
  }

  const loadFileList = async () => {
    loading.value = true
    try {
      const res = await getFileList({
        virtualPath: currentPath.value,
        page: currentPage.value,
        pageSize: pageSize.value
      })
      
      if (res.code === 200) {
        fileListData.value = res.data
        
        if (res.data.current_path) {
          currentPath.value = res.data.current_path
        }
        
        // 使用 Promise.all 并发加载缩略图
        const thumbnailPromises = res.data.files
          .filter((file: any) => file.has_thumbnail && !thumbnailCache.value.has(file.file_id))
          .map(async (file: any) => {
            try {
              const blobUrl = await getThumbnail(file.file_id)
              if (blobUrl) {
                thumbnailCache.value.set(file.file_id, blobUrl)
              }
            } catch (error) {
              // 缩略图加载失败不影响主流程
              proxy?.$log.warn(t('files.thumbnailLoadFailed') + `: ${file.file_id}`, error)
            }
          })
        
        // 不等待缩略图加载完成，后台加载
        Promise.all(thumbnailPromises).catch(() => {
          // 静默处理错误
        })
      } else {
        proxy?.$modal.msgError(res.message || t('files.loadFailed'))
      }
    } catch (error) {
      proxy?.$modal.msgError(t('files.loadFileListFailed'))
      proxy?.$log.error(error)
    } finally {
      loading.value = false
    }
  }

  const navigateToPath = (path: string) => {
    // 先更新路由，watch 会自动处理 currentPath 和 loadFileList
    router.push({
      path: route.path,
      query: {
        virtualPath: path
      }
    })
    // 注意：不需要手动调用 loadFileList，watch 会自动处理
  }

  const getThumbnailUrl = (fileId: string) => {
    return thumbnailCache.value.get(fileId) || ''
  }

  const handlePageChange = (page: number) => {
    currentPage.value = page
    loadFileList()
  }

  const handleSizeChange = (size: number) => {
    pageSize.value = size
    currentPage.value = 1
    loadFileList()
  }

  // 监听路由变化，支持浏览器前进/后退
  watch(
    () => route.query.virtualPath,
    (newPath) => {
      const pathValue = newPath && typeof newPath === 'string' ? newPath : ''
      // 只有当路径真正改变时才更新
      if (currentPath.value !== pathValue) {
        currentPath.value = pathValue
        currentPage.value = 1
        loadFileList()
      }
    },
    { immediate: true } // 立即执行一次，处理初始加载
  )

  return {
    fileListData,
    currentPage,
    pageSize,
    currentPath,
    breadcrumbs,
    formatBreadcrumbName,
    loadFileList,
    navigateToPath,
    getThumbnailUrl,
    handlePageChange,
    handleSizeChange,
    loading
  }
}


import {
  getDownloadTaskList,
  cancelDownload as cancelDownloadApi,
  deleteDownload as deleteDownloadApi,
  pauseDownload,
  resumeDownload
} from '@/api/download'
import { useI18n } from '@/composables'
import type { OfflineDownloadTask } from '@/api/download'

export function useDownloadTasks() {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  const downloadLoading = ref(false)
  const downloadTasks = ref<OfflineDownloadTask[]>([])
  const currentPage = ref(1)
  const pageSize = ref(20)
  const total = ref(0)
  let isFirstLoad = true
  let lastRefreshTime = 0

  // 加载下载任务列表
  const loadDownloadTasks = async (showLoading?: boolean, page?: number, limit?: number) => {
    // 智能刷新：只在首次加载、手动刷新或距离上次刷新超过5秒时显示loading
    const shouldShowLoading =
      showLoading !== false && (isFirstLoad || !lastRefreshTime || Date.now() - lastRefreshTime > 5000)

    if (shouldShowLoading) {
      downloadLoading.value = true
    }

    // 使用传入的分页参数或当前状态
    const pageNum = page !== undefined ? page : currentPage.value
    const pageSizeNum = limit !== undefined ? limit : pageSize.value

    try {
      // 只查询网盘文件下载任务（type=7）
      const res = await getDownloadTaskList({ page: pageNum, pageSize: pageSizeNum, type: 7 })
      if (res.code === 200 && res.data) {
        downloadTasks.value = res.data.tasks || []
        total.value = res.data.total || 0
        currentPage.value = res.data.page || pageNum
        pageSize.value = res.data.page_size || pageSizeNum
      }
    } catch (error: any) {
      // 智能刷新时静默处理错误，避免频繁弹窗
      if (shouldShowLoading) {
        proxy?.$log.error('加载下载任务失败:', error)
        proxy?.$modal.msgError(t('tasks.loadDownloadTasksFailed'))
      } else {
        proxy?.$log.warn('刷新下载任务失败:', error)
      }
    } finally {
      if (shouldShowLoading) {
        downloadLoading.value = false
      }
      isFirstLoad = false
      lastRefreshTime = Date.now()
    }
  }

  // 取消下载
  const cancelDownload = async (taskId: string) => {
    try {
      await proxy?.$modal.confirm(t('tasks.confirmCancelDownload'))
      const res = await cancelDownloadApi(taskId)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('tasks.cancelSuccess'))
        loadDownloadTasks(true, currentPage.value, pageSize.value)
      } else {
        proxy?.$modal.msgError(res.message || t('tasks.cancelFailed'))
      }
    } catch (error: any) {
      if (error !== 'cancel') {
        proxy?.$modal.msgError(t('tasks.cancelFailed'))
      }
    }
  }

  // 删除下载任务
  const deleteDownloadTask = async (taskId: string) => {
    try {
      await proxy?.$modal.confirm(t('tasks.confirmDeleteDownload'))
      const res = await deleteDownloadApi(taskId)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('tasks.deleteSuccess'))
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (currentPage.value > 1 && downloadTasks.value.length === 1) {
          loadDownloadTasks(true, currentPage.value - 1, pageSize.value)
        } else {
          loadDownloadTasks(true, currentPage.value, pageSize.value)
        }
      } else {
        proxy?.$modal.msgError(res.message || t('tasks.deleteFailed'))
      }
    } catch (error: any) {
      if (error !== 'cancel') {
        proxy?.$modal.msgError(t('tasks.deleteFailed'))
      }
    }
  }

  // 暂停下载
  const pauseDownloadTask = async (taskId: string) => {
    try {
      const res = await pauseDownload(taskId)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('tasks.pauseSuccess'))
        loadDownloadTasks(false, currentPage.value, pageSize.value)
      } else {
        proxy?.$modal.msgError(res.message || t('tasks.pauseFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(t('tasks.pauseFailed'))
    }
  }

  // 恢复下载
  const resumeDownloadTask = async (taskId: string) => {
    try {
      const res = await resumeDownload(taskId)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('tasks.resumeSuccess'))
        loadDownloadTasks(false, currentPage.value, pageSize.value)
      } else {
        proxy?.$modal.msgError(res.message || t('tasks.resumeFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(t('tasks.resumeFailed'))
    }
  }

  return {
    downloadTasks,
    downloadLoading,
    currentPage,
    pageSize,
    total,
    loadDownloadTasks,
    cancelDownload,
    deleteDownloadTask,
    pauseDownloadTask,
    resumeDownloadTask
  }
}

import { 
  getDownloadTaskList, 
  cancelDownload as cancelDownloadApi, 
  deleteDownload as deleteDownloadApi,
  pauseDownload,
  resumeDownload
} from '@/api/download'
import type { OfflineDownloadTask } from '@/api/download'

export function useDownloadTasks() {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance

  const downloadLoading = ref(false)
  const downloadTasks = ref<OfflineDownloadTask[]>([])
  let isFirstLoad = true
  let lastRefreshTime = 0

  // 加载下载任务列表
  const loadDownloadTasks = async (showLoading?: boolean) => {
    // 智能刷新：只在首次加载、手动刷新或距离上次刷新超过5秒时显示loading
    const shouldShowLoading = showLoading !== false && (
      isFirstLoad || 
      !lastRefreshTime || 
      Date.now() - lastRefreshTime > 5000
    )
    
    if (shouldShowLoading) {
      downloadLoading.value = true
    }
    
    try {
      // 只查询网盘文件下载任务（type=7）
      const res = await getDownloadTaskList({ page: 1, pageSize: 100, type: 7 })
      if (res.code === 200 && res.data) {
        downloadTasks.value = res.data.tasks || []
      }
    } catch (error: any) {
      // 智能刷新时静默处理错误，避免频繁弹窗
      if (shouldShowLoading) {
        proxy?.$log.error('加载下载任务失败:', error)
        proxy?.$modal.msgError('加载下载任务失败')
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
      await proxy?.$modal.confirm('确认取消该下载任务?')
      const res = await cancelDownloadApi(taskId)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess('已取消')
        loadDownloadTasks()
      } else {
        proxy?.$modal.msgError(res.message || '取消失败')
      }
    } catch (error: any) {
      if (error !== 'cancel') {
        proxy?.$modal.msgError('取消失败')
      }
    }
  }

  // 删除下载任务
  const deleteDownloadTask = async (taskId: string) => {
    try {
      await proxy?.$modal.confirm('确认删除该任务记录?')
      const res = await deleteDownloadApi(taskId)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess('已删除')
        loadDownloadTasks()
      } else {
        proxy?.$modal.msgError(res.message || '删除失败')
      }
    } catch (error: any) {
      if (error !== 'cancel') {
        proxy?.$modal.msgError('删除失败')
      }
    }
  }

  // 暂停下载
  const pauseDownloadTask = async (taskId: string) => {
    try {
      const res = await pauseDownload(taskId)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess('已暂停')
        loadDownloadTasks()
      } else {
        proxy?.$modal.msgError(res.message || '暂停失败')
      }
    } catch (error: any) {
      proxy?.$modal.msgError('暂停失败')
    }
  }

  // 恢复下载
  const resumeDownloadTask = async (taskId: string) => {
    try {
      const res = await resumeDownload(taskId)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess('已恢复')
        loadDownloadTasks()
      } else {
        proxy?.$modal.msgError(res.message || '恢复失败')
      }
    } catch (error: any) {
      proxy?.$modal.msgError('恢复失败')
    }
  }

  return {
    downloadTasks,
    downloadLoading,
    loadDownloadTasks,
    cancelDownload,
    deleteDownloadTask,
    pauseDownloadTask,
    resumeDownloadTask
  }
}


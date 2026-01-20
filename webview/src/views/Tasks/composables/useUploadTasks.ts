import { uploadTaskManager } from '@/utils/file/uploadTaskManager'
import { findBackendTask, syncBackendTasksToFrontend } from '@/utils/file/uploadTaskSync'
import { deleteUploadTask, getUploadProgress, listExpiredUploads, getUploadTaskList } from '@/api/file'
import { formatFileSizeForDisplay } from '@/utils'
import { isUploadTaskActive, openFileDialog, uploadSingleFile } from '@/utils/file/upload'
import { useI18n } from '@/composables'

export function useUploadTasks() {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  const uploadLoading = ref(false)
  const cleanLoading = ref(false)
  const uploadTasks = ref<any[]>([])
  const allUploadTasks = ref<any[]>([]) // 存储所有任务（用于分页）

  // 分页状态
  const currentPage = ref(1)
  const pageSize = ref(20)
  const total = computed(() => allUploadTasks.value.length)

  // 防抖相关
  let syncTimer: number | null = null
  let lastSyncTime = 0
  const SYNC_DEBOUNCE_TIME = 2000 // 2秒内最多同步一次

  // 更新分页数据
  const updatePaginatedTasks = () => {
    const start = (currentPage.value - 1) * pageSize.value
    const end = start + pageSize.value
    uploadTasks.value = allUploadTasks.value.slice(start, end)
  }

  // 加载上传任务列表（带防抖）
  const loadUploadTasks = async (showLoading = true, forceSync = false) => {
    if (showLoading) {
      uploadLoading.value = true
    }
    try {
      // 先加载本地任务并更新UI
      const localTasks = uploadTaskManager.getAllTasks()
      allUploadTasks.value = localTasks
      updatePaginatedTasks()

      // 防抖：如果不是强制同步，且距离上次同步时间小于防抖时间，则跳过后端同步
      const now = Date.now()
      if (!forceSync && now - lastSyncTime < SYNC_DEBOUNCE_TIME) {
        // 取消之前的定时器
        if (syncTimer) {
          clearTimeout(syncTimer)
        }
        // 设置新的定时器，延迟执行同步
        syncTimer = window.setTimeout(async () => {
          await syncTasksFromBackend()
          lastSyncTime = Date.now()
        }, SYNC_DEBOUNCE_TIME)
        return
      }

      // 执行后端同步
      lastSyncTime = now
      await syncTasksFromBackend()
    } catch (error: any) {
      proxy?.$log.error('加载上传任务失败:', error)
    } finally {
      if (showLoading) {
        uploadLoading.value = false
      }
    }
  }

  // 从后端同步任务（包括所有状态的任务）
  const syncTasksFromBackend = async () => {
    try {
      // 获取所有任务列表（包括已完成和未完成的任务）
      // 使用较大的分页大小，获取所有任务
      const pageSize = 100
      let currentPage = 1
      let hasMore = true
      const allBackendTasks: any[] = []

      while (hasMore) {
        try {
          const response = await getUploadTaskList({
            page: currentPage,
            pageSize: pageSize
          })

          if (response.code === 200 && response.data) {
            const { tasks, total } = response.data
            if (tasks && Array.isArray(tasks)) {
              allBackendTasks.push(...tasks)
              
              // 检查是否还有更多数据
              if (allBackendTasks.length >= total || tasks.length < pageSize) {
                hasMore = false
              } else {
                currentPage++
              }
            } else {
              hasMore = false
            }
          } else {
            hasMore = false
          }
        } catch (error: any) {
          proxy?.$log.warn('获取任务列表失败:', error)
          hasMore = false
        }
      }

      // 将后端的所有任务同步到前端（包括已完成和未完成的任务）
      if (allBackendTasks.length > 0) {
        // 使用现有的同步函数，但需要适配新的数据结构
        const backendTasks = allBackendTasks.map(task => ({
          id: task.id,
          file_name: task.file_name,
          file_size: task.file_size,
          chunk_size: task.chunk_size,
          total_chunks: task.total_chunks,
          uploaded_chunks: task.uploaded_chunks,
          progress: task.progress || 0,
          status: task.status,
          error_message: task.error_message,
          path_id: task.path_id,
          create_time: task.create_time,
          update_time: task.update_time,
          expire_time: task.expire_time
        }))

        syncBackendTasksToFrontend(backendTasks)
      }

      // 更新前端任务列表
      const allTasks = uploadTaskManager.getAllTasks()
      allUploadTasks.value = allTasks
      updatePaginatedTasks()
    } catch (error: any) {
      proxy?.$log.error('同步任务失败:', error)
    }
  }

  // 处理分页变化
  const handlePagination = ({ page, limit }: { page: number; limit: number }) => {
    currentPage.value = page
    pageSize.value = limit
    updatePaginatedTasks()
  }

  // 暂停上传
  const pauseUpload = (taskId: string) => {
    uploadTaskManager.pauseTask(taskId)
    uploadTaskManager.cancelAllUploads(taskId)
    proxy?.$modal.msgSuccess(t('tasks.paused'))
  }

  // 恢复上传
  const resumeUpload = async (taskId: string) => {
    const task = uploadTaskManager.getTask(taskId)
    if (!task) {
      proxy?.$modal.msgError(t('tasks.taskNotExists'))
      return
    }

    if (task.status !== 'paused') {
      proxy?.$modal.msgError(t('tasks.taskStatusIncorrect'))
      return
    }

    if (!task.pathId) {
      proxy?.$modal.msgError(t('tasks.taskInfoIncomplete'))
      return
    }

    try {
      if (!task.precheckId) {
        proxy?.$modal.msgError(t('tasks.taskPrecheckMissing'))
        return
      }

      let progressData: any = null
      let uploaded = 0
      let total = 0
      let progressPercent = 0

      try {
        const backendTask = await findBackendTask(task.precheckId)

        if (backendTask) {
          uploaded = backendTask.uploaded_chunks || 0
          total = backendTask.total_chunks || 0
          progressPercent = backendTask.progress || 0
          progressData = {
            uploaded,
            total,
            progress: progressPercent,
            md5: []
          }
        }
      } catch (error: any) {
        // 继续尝试使用 getUploadProgress
      }

      if (!progressData || !progressData.md5) {
        const progressResponse = await getUploadProgress(task.precheckId)

        if (progressResponse.code === 200 && progressResponse.data) {
          progressData = progressResponse.data as any
          uploaded = progressData.uploaded || 0
          total = progressData.total || 0
          progressPercent = progressData.progress || (total > 0 ? (uploaded / total) * 100 : 0)
        } else {
          proxy?.$modal.msgError(t('tasks.cannotQueryProgress'))
          return
        }
      }

      if (!progressData) {
        proxy?.$modal.msgError(t('tasks.cannotGetProgress'))
        return
      }

      // 更新任务的已上传分片MD5列表和进度信息
      if (progressData.md5 && Array.isArray(progressData.md5) && progressData.md5.length > 0) {
        uploadTaskManager.updateTask(taskId, { uploadedChunkMd5s: progressData.md5 })
      }

      uploadTaskManager.updateTask(taskId, {
        progress: Math.floor(progressPercent),
        uploaded_size: Math.floor((uploaded / total) * task.file_size)
      })

      // 检查任务是否在活动状态（文件对象还在内存中或 uploadSingleFile 仍在运行）
      const isTaskActive = isUploadTaskActive(taskId)

      if (isTaskActive) {
        // 任务在活动状态，直接恢复任务状态
        // ConcurrentUploader 会自动检测到状态变化并继续上传
        uploadTaskManager.resumeTask(taskId)
        proxy?.$modal.msgSuccess(t('tasks.resumed'))
        return
      }

      // 如果任务不在活动状态，需要重新选择文件继续上传

      const sizeDisplay = formatFileSizeForDisplay(task.file_size)
      ElMessage.info({
        message: t('tasks.selectFileToResume', { fileName: task.file_name, size: sizeDisplay }),
        duration: 3000
      })

      const files = await openFileDialog(false)

      if (files.length === 0) {
        return
      }

      const selectedFile = files[0]

      if (selectedFile.name !== task.file_name || selectedFile.size !== task.file_size) {
        const expectedSize = formatFileSizeForDisplay(task.file_size)
        proxy?.$modal.msgError(t('tasks.fileMismatch', { fileName: task.file_name, size: expectedSize }))
        return
      }

      await uploadSingleFile({
        file: selectedFile,
        pathId: task.pathId!,
        taskId: taskId,
        onProgress: () => {},
        onSuccess: fileName => {
          proxy?.$modal.msgSuccess(t('tasks.uploadSuccess', { fileName }))
          loadUploadTasks(false)
        },
        onError: (error, fileName) => {
          proxy?.$modal.msgError(t('tasks.uploadFailed', { fileName, error: error.message }))
          loadUploadTasks(false)
        }
      })

      proxy?.$modal.msgSuccess(t('tasks.resumed'))
    } catch (error: any) {
      proxy?.$modal.msgError(t('tasks.resumeFailed', { error: error.message }))
      proxy?.$log.error('恢复上传失败:', error)
    }
  }

  // 取消上传
  const cancelUpload = async (taskId: string) => {
    try {
      await proxy?.$modal.confirm(t('tasks.confirmCancelUpload'))
      uploadTaskManager.cancelTask(taskId)
      uploadTaskManager.cancelAllUploads(taskId)
      proxy?.$modal.msgSuccess(t('tasks.cancelSuccess'))
      await loadUploadTasks(false)
    } catch (error) {
      // 用户取消操作
    }
  }

  // 删除上传任务
  const deleteUpload = async (taskId: string) => {
    try {
      const task = uploadTaskManager.getTask(taskId)
      if (!task) {
        proxy?.$modal.msgError(t('tasks.taskNotExists') || '任务不存在')
        return
      }

      if (task.status === 'uploading') {
        await proxy?.$modal.confirm(t('tasks.confirmDeleteUploading'))
        uploadTaskManager.cancelTask(taskId)
        uploadTaskManager.cancelAllUploads(taskId)
      } else {
        await proxy?.$modal.confirm(t('tasks.confirmDeleteUpload'))
      }

      const precheckId = task.precheckId || taskId
      try {
        await deleteUploadTask(precheckId)
      } catch (error: any) {
        proxy?.$log.warn('调用后端删除接口失败:', error)
      }

      uploadTaskManager.deleteTask(taskId)
      proxy?.$modal.msgSuccess(t('tasks.deleteSuccess'))
      // 更新所有任务列表
      allUploadTasks.value = uploadTaskManager.getAllTasks()
      // 如果当前页没有数据了，且不是第一页，则跳转到上一页
      if (uploadTasks.value.length === 1 && currentPage.value > 1) {
        currentPage.value--
      }
      updatePaginatedTasks()
    } catch (error) {
      // 用户取消操作
    }
  }

  // 获取过期任务数量
  const expiredTaskCount = ref(0)
  const getExpiredTaskCount = async () => {
    try {
      const res = await listExpiredUploads()
      if (res.code === 200 && res.data) {
        expiredTaskCount.value = res.data.length
      }
    } catch (error: any) {
      proxy?.$log.warn('获取过期任务数量失败:', error)
      expiredTaskCount.value = 0
    }
  }

  // 清理过期任务（保留，但改为显示弹窗）
  const cleanExpiredUploads = async () => {
    // 这个方法现在不再使用，改为显示过期任务弹窗
    // 保留是为了兼容性
  }

  // 一键清空所有上传任务
  const clearAllLoading = ref(false)
  const clearAllUploadTasks = async () => {
    try {
      const allTasks = uploadTaskManager.getAllTasks()
      if (allTasks.length === 0) {
        proxy?.$modal.msgWarning(t('tasks.noTasksToClear'))
        return
      }

      // 统计不同状态的任务数量
      const uploadingTasks = allTasks.filter(t => t.status === 'uploading' || t.status === 'prechecking' || t.status === 'pending')
      const otherTasks = allTasks.filter(t => !['uploading', 'prechecking', 'pending'].includes(t.status))

      let confirmMessage = ''
      if (uploadingTasks.length > 0 && otherTasks.length > 0) {
        confirmMessage = t('tasks.confirmClearAllWithUploading', {
          total: allTasks.length,
          uploading: uploadingTasks.length,
          other: otherTasks.length
        }) || `确认清空所有上传任务？\n共有 ${allTasks.length} 个任务，其中 ${uploadingTasks.length} 个正在上传/预检中，${otherTasks.length} 个已完成/失败/已取消。\n正在上传的任务将被取消。`
      } else if (uploadingTasks.length > 0) {
        confirmMessage = t('tasks.confirmClearAllUploading', {
          count: uploadingTasks.length
        }) || `确认清空所有上传任务？\n共有 ${uploadingTasks.length} 个正在上传/预检中的任务，清空将取消这些任务。`
      } else {
        confirmMessage = t('tasks.confirmClearAll', {
          count: otherTasks.length
        }) || `确认清空所有上传任务？\n共有 ${otherTasks.length} 个已完成/失败/已取消的任务将被清空。`
      }

      await proxy?.$modal.confirm(confirmMessage)

      // 先取消所有正在上传的任务
      uploadingTasks.forEach(task => {
        uploadTaskManager.cancelTask(task.id)
        uploadTaskManager.cancelAllUploads(task.id)
      })

      // 尝试删除后端任务（批量删除，忽略错误）
      const deletePromises = allTasks.map(async task => {
        if (task.precheckId) {
          try {
            await deleteUploadTask(task.precheckId)
          } catch (error: any) {
            proxy?.$log.warn(`删除后端任务失败: ${task.precheckId}`, error)
          }
        }
      })
      await Promise.allSettled(deletePromises)

      // 清空所有本地任务
      uploadTaskManager.clearAllTasks()
      proxy?.$modal.msgSuccess(t('tasks.clearAllSuccess', { count: allTasks.length }))

      // 重置分页并重新加载
      currentPage.value = 1
      allUploadTasks.value = []
      uploadTasks.value = []
    } catch (error) {
      // 用户取消操作
    } finally {
      clearAllLoading.value = false
    }
  }

  return {
    uploadTasks,
    uploadLoading,
    cleanLoading,
    clearAllLoading,
    expiredTaskCount,
    currentPage,
    pageSize,
    total,
    loadUploadTasks,
    getExpiredTaskCount,
    pauseUpload,
    resumeUpload,
    cancelUpload,
    deleteUpload,
    cleanExpiredUploads,
    clearAllUploadTasks,
    handlePagination
  }
}

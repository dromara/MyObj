import { uploadTaskManager } from '@/utils/uploadTaskManager'
import { loadAndSyncBackendTasks, findBackendTask } from '@/utils/uploadTaskSync'
import { deleteUploadTask, getUploadProgress, listExpiredUploads } from '@/api/file'
import { formatFileSizeForDisplay } from '@/utils'
import { isUploadTaskActive, openFileDialog, uploadSingleFile } from '@/utils/upload'
import { useI18n } from '@/composables/useI18n'

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

  // 更新分页数据
  const updatePaginatedTasks = () => {
    const start = (currentPage.value - 1) * pageSize.value
    const end = start + pageSize.value
    uploadTasks.value = allUploadTasks.value.slice(start, end)
  }

  // 加载上传任务列表
  const loadUploadTasks = async (showLoading = true) => {
    if (showLoading) {
      uploadLoading.value = true
    }
    try {
      const localTasks = uploadTaskManager.getAllTasks()
      allUploadTasks.value = localTasks
      updatePaginatedTasks()
      
      const syncResult = await loadAndSyncBackendTasks()
      
      if (syncResult.success) {
        const allTasks = uploadTaskManager.getAllTasks()
        allUploadTasks.value = allTasks
        updatePaginatedTasks()
      } else if (syncResult.error) {
        proxy?.$log.warn('任务同步失败:', syncResult.error)
      }
    } catch (error: any) {
      proxy?.$log.error('加载上传任务失败:', error)
    } finally {
      if (showLoading) {
        uploadLoading.value = false
      }
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
      proxy?.$modal.msgError(t('tasks.taskNotExists') || '任务不存在')
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
          progressPercent = progressData.progress || (total > 0 ? (uploaded / total * 100) : 0)
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
        proxy?.$modal.msgError(
          t('tasks.fileMismatch', { fileName: task.file_name, size: expectedSize })
        )
        return
      }

      await uploadSingleFile({
        file: selectedFile,
        pathId: task.pathId!,
        taskId: taskId,
        onProgress: () => {},
        onSuccess: (fileName) => {
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

  return {
    uploadTasks,
    uploadLoading,
    cleanLoading,
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
    handlePagination
  }
}


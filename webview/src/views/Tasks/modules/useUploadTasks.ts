import { uploadTaskManager } from '@/utils/uploadTaskManager'
import { loadAndSyncBackendTasks, findBackendTask } from '@/utils/uploadTaskSync'
import { deleteUploadTask, getUploadProgress } from '@/api/file'
import { formatFileSizeForDisplay } from '@/utils'
import { isUploadTaskActive, openFileDialog, uploadSingleFile } from '@/utils/upload'

export function useUploadTasks() {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance

  const uploadLoading = ref(false)
  const cleanLoading = ref(false)
  const uploadTasks = ref<any[]>([])

  // 加载上传任务列表
  const loadUploadTasks = async () => {
    uploadLoading.value = true
    try {
      const localTasks = uploadTaskManager.getAllTasks()
      uploadTasks.value = localTasks
      
      const syncResult = await loadAndSyncBackendTasks()
      
      if (syncResult.success) {
        const allTasks = uploadTaskManager.getAllTasks()
        uploadTasks.value = allTasks
      } else if (syncResult.error) {
        proxy?.$log.warn('任务同步失败:', syncResult.error)
      }
    } catch (error: any) {
      proxy?.$log.error('加载上传任务失败:', error)
    } finally {
      uploadLoading.value = false
    }
  }

  // 暂停上传
  const pauseUpload = (taskId: string) => {
    uploadTaskManager.pauseTask(taskId)
    uploadTaskManager.cancelAllUploads(taskId)
    proxy?.$modal.msgSuccess('已暂停')
  }

  // 恢复上传
  const resumeUpload = async (taskId: string) => {
    const task = uploadTaskManager.getTask(taskId)
    if (!task) {
      proxy?.$modal.msgError('任务不存在')
      return
    }

    if (task.status !== 'paused') {
      proxy?.$modal.msgError('任务状态不正确，无法恢复')
      return
    }

    if (!task.pathId) {
      proxy?.$modal.msgError('任务信息不完整（缺少路径信息），无法恢复上传')
      return
    }

    try {
      if (!task.precheckId) {
        proxy?.$modal.msgError('任务信息不完整（缺少预检ID），无法恢复上传。')
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
          proxy?.$modal.msgError('无法查询上传进度，预检信息可能已过期。')
          return
        }
      }
      
      if (!progressData) {
        proxy?.$modal.msgError('无法获取进度数据，无法恢复上传。')
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
        proxy?.$modal.msgSuccess('已继续上传')
        return
      }
      
      // 如果任务不在活动状态，需要重新选择文件继续上传
      
      const sizeDisplay = formatFileSizeForDisplay(task.file_size)
      ElMessage.info({
        message: `请选择文件 "${task.file_name}" (${sizeDisplay}) 继续上传`,
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
          `选择的文件与原始文件不匹配。\n` +
          `请选择文件名为 "${task.file_name}"、大小为 ${expectedSize} 的文件。`
        )
        return
      }

      await uploadSingleFile({
        file: selectedFile,
        pathId: task.pathId!,
        taskId: taskId,
        onProgress: () => {},
        onSuccess: (fileName) => {
          proxy?.$modal.msgSuccess(`文件 ${fileName} 上传成功`)
          loadUploadTasks()
        },
        onError: (error, fileName) => {
          proxy?.$modal.msgError(`文件 ${fileName} 上传失败: ${error.message}`)
          loadUploadTasks()
        }
      })

      proxy?.$modal.msgSuccess('已继续上传')
    } catch (error: any) {
      proxy?.$modal.msgError(`恢复上传失败: ${error.message}`)
      proxy?.$log.error('恢复上传失败:', error)
    }
  }

  // 取消上传
  const cancelUpload = async (taskId: string) => {
    try {
      await proxy?.$modal.confirm('确认取消该上传任务?')
      uploadTaskManager.cancelTask(taskId)
      uploadTaskManager.cancelAllUploads(taskId)
      proxy?.$modal.msgSuccess('已取消')
      loadUploadTasks()
    } catch (error) {
      // 用户取消操作
    }
  }

  // 删除上传任务
  const deleteUpload = async (taskId: string) => {
    try {
      const task = uploadTaskManager.getTask(taskId)
      if (!task) {
        proxy?.$modal.msgError('任务不存在')
        return
      }
      
      if (task.status === 'uploading') {
        await proxy?.$modal.confirm('任务正在上传中，删除将取消上传。确认删除?')
        uploadTaskManager.cancelTask(taskId)
        uploadTaskManager.cancelAllUploads(taskId)
      } else {
        await proxy?.$modal.confirm('确认删除该任务记录?')
      }
      
      const precheckId = task.precheckId || taskId
      try {
        await deleteUploadTask(precheckId)
      } catch (error: any) {
        proxy?.$log.warn('调用后端删除接口失败:', error)
      }
      
      uploadTaskManager.deleteTask(taskId)
      proxy?.$modal.msgSuccess('已删除')
      uploadTasks.value = uploadTaskManager.getAllTasks()
    } catch (error) {
      // 用户取消操作
    }
  }

  // 获取过期任务数量
  const expiredTaskCount = ref(0)
  const getExpiredTaskCount = async () => {
    try {
      const { listExpiredUploads } = await import('@/api/file')
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
    loadUploadTasks,
    getExpiredTaskCount,
    pauseUpload,
    resumeUpload,
    cancelUpload,
    deleteUpload,
    cleanExpiredUploads
  }
}


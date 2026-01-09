import { uploadTaskManager } from '@/utils/uploadTaskManager'
import { getUploadProgress, listExpiredUploads, getUploadTaskList, deleteUploadTask, type UploadTaskItem } from '@/api/file'
import { formatFileSizeForDisplay } from '@/utils'
import { isUploadTaskActive, openFileDialog, uploadSingleFile } from '@/utils/upload'

// 合并本地临时任务和后端任务列表
const mergeTasks = (backendTasks: any[], localTasks: any[]): any[] => {
    // 创建后端任务的 precheckId 集合，用于去重
    const backendPrecheckIds = new Set(backendTasks.map(t => t.precheckId || t.id))
    
    // 过滤出本地任务中不在后端的任务（临时任务或正在进行的任务）
    const localOnlyTasks = localTasks.filter(t => {
      const precheckId = t.precheckId || t.id
      // 只保留未完成的任务（pending, uploading, paused）
      return !backendPrecheckIds.has(precheckId) && 
             (t.status === 'pending' || t.status === 'uploading' || t.status === 'paused')
    })
    
    // 合并：本地临时任务在前，后端任务在后，按创建时间倒序
    const merged = [...localOnlyTasks, ...backendTasks]
    merged.sort((a, b) => {
      const timeA = new Date(a.created_at || a.create_time || 0).getTime()
      const timeB = new Date(b.created_at || b.create_time || 0).getTime()
      return timeB - timeA
    })
    
  return merged
}

export function useUploadTasks() {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance

  const uploadLoading = ref(false)
  const cleanLoading = ref(false)
  const uploadTasks = ref<any[]>([])
  const totalTasks = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(20)

  // 从后端API加载上传任务列表（分页），并合并本地临时任务
  const loadUploadTasks = async (page: number = 1, size: number = 20) => {
    uploadLoading.value = true
    try {
      // 获取本地临时任务
      const localTasks = uploadTaskManager.getAllTasks()
      
      // 获取后端任务列表
      const res = await getUploadTaskList({ page, pageSize: size })
      if (res.code === 200 && res.data) {
        // 将后端任务转换为前端任务格式
        const convertedTasks = res.data.tasks.map((task: UploadTaskItem) => {
          // 计算已上传大小
          const uploadedSize = task.total_chunks > 0 
            ? Math.floor((task.uploaded_chunks / task.total_chunks) * task.file_size)
            : 0
          
          // 映射状态
          let status: 'pending' | 'uploading' | 'paused' | 'completed' | 'failed' | 'cancelled' = 'pending'
          if (task.status === 'completed') {
            status = 'completed'
          } else if (task.status === 'failed' || task.status === 'aborted') {
            status = 'failed'
          } else if (task.status === 'uploading') {
            status = 'uploading'
          } else if (task.status === 'pending') {
            status = 'pending'
          }
          
          return {
            id: task.id,
            file_name: task.file_name,
            file_size: task.file_size,
            uploaded_size: uploadedSize,
            progress: Math.floor(task.progress),
            status: status,
            stage: status === 'uploading' ? 'uploading' : status === 'completed' ? 'completed' : status === 'failed' ? 'failed' : 'calculating',
            speed: '0 KB/s',
            created_at: task.create_time,
            error: task.error_message,
            pathId: task.path_id,
            precheckId: task.id,
            chunkSignature: task.chunk_signature,
            filesMd5: [],
            uploadedChunkMd5s: []
          }
        })
        
        // 合并本地临时任务和后端任务
        uploadTasks.value = mergeTasks(convertedTasks, localTasks)
        totalTasks.value = res.data.total
        currentPage.value = res.data.page
        pageSize.value = res.data.page_size
      } else {
        // 如果后端请求失败，至少显示本地任务
        uploadTasks.value = localTasks.filter(t => 
          t.status === 'pending' || t.status === 'uploading' || t.status === 'paused'
        )
      }
    } catch (error: any) {
      proxy?.$log.error('加载上传任务失败:', error)
      // 即使后端失败，也显示本地任务
      const localTasks = uploadTaskManager.getAllTasks()
      uploadTasks.value = localTasks.filter(t => 
        t.status === 'pending' || t.status === 'uploading' || t.status === 'paused'
      )
      // 不显示错误提示，避免干扰用户体验
    } finally {
      uploadLoading.value = false
    }
  }

  // 暂停上传（仅对正在上传的任务有效）
  const pauseUpload = (taskId: string) => {
    // 检查任务是否在活动状态（正在上传）
    const isTaskActive = isUploadTaskActive(taskId)
    if (isTaskActive) {
      uploadTaskManager.pauseTask(taskId)
      uploadTaskManager.cancelAllUploads(taskId)
      proxy?.$modal.msgSuccess('已暂停')
      // 刷新列表以更新状态
      loadUploadTasks(currentPage.value, pageSize.value)
    } else {
      proxy?.$modal.msgError('该任务未在上传中，无法暂停')
    }
  }

  // 恢复上传
  const resumeUpload = async (taskId: string) => {
    // 从列表中找到任务
    const task = uploadTasks.value.find(t => t.id === taskId || t.precheckId === taskId)
    if (!task) {
      proxy?.$modal.msgError('任务不存在')
      return
    }

    if (task.status !== 'paused' && task.status !== 'pending') {
      proxy?.$modal.msgError('任务状态不正确，无法恢复')
      return
    }

    if (!task.pathId) {
      proxy?.$modal.msgError('任务信息不完整（缺少路径信息），无法恢复上传')
      return
    }

    try {
      const precheckId = task.precheckId || taskId
      
      // 获取上传进度
      const progressResponse = await getUploadProgress(precheckId)
      
      if (progressResponse.code !== 200 || !progressResponse.data) {
        proxy?.$modal.msgError('无法查询上传进度，预检信息可能已过期。')
        return
      }
      
      // 检查任务是否在活动状态（文件对象还在内存中或 uploadSingleFile 仍在运行）
      const isTaskActive = isUploadTaskActive(taskId)
      
      if (isTaskActive) {
        // 任务在活动状态，直接恢复任务状态
        uploadTaskManager.resumeTask(taskId)
        proxy?.$modal.msgSuccess('已继续上传')
        // 刷新列表
        loadUploadTasks(currentPage.value, pageSize.value)
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
        pathId: task.pathId,
        taskId: precheckId,
        onProgress: () => {},
        onSuccess: (fileName) => {
          proxy?.$modal.msgSuccess(`文件 ${fileName} 上传成功`)
          loadUploadTasks(currentPage.value, pageSize.value)
        },
        onError: (error, fileName) => {
          proxy?.$modal.msgError(`文件 ${fileName} 上传失败: ${error.message}`)
          loadUploadTasks(currentPage.value, pageSize.value)
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
      
      // 如果任务正在上传，先取消上传
      const isTaskActive = isUploadTaskActive(taskId)
      if (isTaskActive) {
        uploadTaskManager.cancelTask(taskId)
        uploadTaskManager.cancelAllUploads(taskId)
      }
      
      proxy?.$modal.msgSuccess('已取消')
      // 刷新列表
      loadUploadTasks(currentPage.value, pageSize.value)
    } catch (error) {
      // 用户取消操作
    }
  }

  // 删除上传任务
  const deleteUpload = async (taskId: string) => {
    try {
      // 从列表中找到任务
      const task = uploadTasks.value.find(t => t.id === taskId || t.precheckId === taskId)
      if (!task) {
        proxy?.$modal.msgError('任务不存在')
        return
      }
      
      if (task.status === 'uploading') {
        await proxy?.$modal.confirm('任务正在上传中，删除将取消上传。确认删除?')
        // 如果任务正在上传，先取消上传
        const isTaskActive = isUploadTaskActive(taskId)
        if (isTaskActive) {
          uploadTaskManager.cancelTask(taskId)
          uploadTaskManager.cancelAllUploads(taskId)
        }
      } else {
        await proxy?.$modal.confirm('确认删除该任务记录?')
      }
      
      const precheckId = task.precheckId || taskId
      try {
        await deleteUploadTask(precheckId)
        proxy?.$modal.msgSuccess('已删除')
        // 刷新列表
        loadUploadTasks(currentPage.value, pageSize.value)
      } catch (error: any) {
        proxy?.$log.error('删除任务失败:', error)
        proxy?.$modal.msgError('删除任务失败')
      }
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

  // 订阅本地任务更新（用于实时显示临时任务）
  let unsubscribe: (() => void) | null = null

  // 初始化订阅
  const initTaskSubscription = () => {
    if (unsubscribe) {
      unsubscribe()
    }
    unsubscribe = uploadTaskManager.subscribe((localTasks) => {
      // 当本地任务更新时，重新合并任务列表
      if (uploadTasks.value.length > 0 || localTasks.length > 0) {
        // 获取当前的后端任务（从现有列表中提取）
        const backendTasks = uploadTasks.value.filter(t => t.precheckId && 
          !localTasks.some(lt => (lt.precheckId || lt.id) === (t.precheckId || t.id)))
        // 重新合并
        uploadTasks.value = mergeTasks(backendTasks, localTasks)
      }
    })
  }

  return {
    uploadTasks,
    uploadLoading,
    cleanLoading,
    expiredTaskCount,
    totalTasks,
    currentPage,
    pageSize,
    loadUploadTasks,
    getExpiredTaskCount,
    pauseUpload,
    resumeUpload,
    cancelUpload,
    deleteUpload,
    cleanExpiredUploads,
    initTaskSubscription
  }
}


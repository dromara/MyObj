<template>
  <div class="tasks-page">
    <el-tabs v-model="activeTab" class="task-tabs">
      <el-tab-pane label="上传任务" name="upload">
        <el-card shadow="never">
          <div style="margin-bottom: 16px; display: flex; justify-content: space-between; align-items: center;">
            <span style="font-size: 14px; color: #606266;">共 {{ uploadTasks.length }} 个任务</span>
            <el-button 
              type="warning" 
              size="small" 
              icon="Delete" 
              @click="cleanExpiredUploads"
              :loading="cleanLoading"
            >
              清理过期任务
            </el-button>
          </div>
          <el-table :data="uploadTasks" v-loading="uploadLoading">
            <el-table-column label="文件名" min-width="300">
              <template #default="{ row }">
                <div class="file-name-cell">
                  <el-icon :size="24" color="#409EFF"><Document /></el-icon>
                  <span>{{ row.file_name }}</span>
                </div>
              </template>
            </el-table-column>
            
            <el-table-column label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="getUploadStatusType(row.status)">{{ getUploadStatusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
            
            <el-table-column label="进度" width="250">
              <template #default="{ row }">
                <div class="progress-cell">
                  <el-progress 
                    :percentage="row.progress" 
                    :status="row.status === 'completed' ? 'success' : row.status === 'failed' ? 'exception' : undefined"
                  />
                  <span class="progress-info">{{ formatSize(row.uploaded_size) }} / {{ formatSize(row.file_size) }} · {{ row.speed || '0 KB/s' }}</span>
                </div>
              </template>
            </el-table-column>
            
            <el-table-column label="创建时间" width="180">
              <template #default="{ row }">
                {{ formatDate(row.created_at) }}
              </template>
            </el-table-column>
            
            <el-table-column label="操作" width="240" fixed="right">
              <template #default="{ row }">
                <!-- 上传中或等待中：显示暂停和取消 -->
                <el-button 
                  v-if="row.status === 'uploading'"
                  link 
                  icon="VideoPause" 
                  type="warning"
                  @click="pauseUpload(row.id)"
                >
                  暂停
                </el-button>
                <el-button 
                  v-if="row.status === 'paused'"
                  link 
                  icon="VideoPlay" 
                  type="primary"
                  @click="resumeUpload(row.id)"
                >
                  继续
                </el-button>
                <el-button 
                  v-if="row.status === 'uploading' || row.status === 'pending' || row.status === 'paused'"
                  link 
                  icon="Close" 
                  type="danger"
                  @click="cancelUpload(row.id)"
                >
                  取消
                </el-button>
                <!-- 所有状态都可以删除 -->
                <el-button 
                  link 
                  icon="Delete" 
                  type="danger"
                  @click="deleteUpload(row.id)"
                  :disabled="row.status === 'uploading'"
                >
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          
          <el-empty v-if="uploadTasks.length === 0 && !uploadLoading" description="暂无上传任务" />
        </el-card>
      </el-tab-pane>
      
      <el-tab-pane label="下载任务" name="download">
        <el-card shadow="never">
          <el-table :data="downloadTasks" v-loading="downloadLoading">
            <el-table-column label="文件名" min-width="300">
              <template #default="{ row }">
                <div class="file-name-cell">
                  <el-icon :size="24" color="#67C23A"><Document /></el-icon>
                  <span>{{ row.file_name || row.url }}</span>
                </div>
              </template>
            </el-table-column>
            
            <el-table-column label="类型" width="120">
              <template #default="{ row }">
                <el-tag :type="getDownloadTypeColor(row.type)" effect="plain">{{ row.type_text }}</el-tag>
              </template>
            </el-table-column>
            
            <el-table-column label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="getDownloadStatusType(row.state)">{{ row.state_text }}</el-tag>
              </template>
            </el-table-column>
            
            <el-table-column label="进度" width="250">
              <template #default="{ row }">
                <div class="progress-cell">
                  <el-progress 
                    :percentage="row.progress" 
                    :status="row.state === 3 ? 'success' : row.state === 4 ? 'exception' : undefined"
                  />
                  <span class="progress-info">{{ formatSize(row.downloaded_size) }} / {{ formatSize(row.file_size) }} · {{ formatSpeed(row.speed) }}</span>
                </div>
              </template>
            </el-table-column>
            
            <el-table-column label="创建时间" width="180">
              <template #default="{ row }">
                {{ formatDate(row.create_time) }}
              </template>
            </el-table-column>
            
            <el-table-column label="操作" width="240" fixed="right">
              <template #default="{ row }">
                <!-- 其他类型的下载任务，显示暂停/继续按钮 -->
                <el-button 
                  v-if="row.type !== 7 && row.state === 1"
                  link 
                  icon="VideoPause" 
                  type="warning"
                  @click="pauseDownloadTask(row.id)"
                >
                  暂停
                </el-button>
                <el-button 
                  v-if="row.type !== 7 && row.state === 2"
                  link 
                  icon="VideoPlay" 
                  type="primary"
                  @click="resumeDownloadTask(row.id)"
                >
                  继续
                </el-button>
                <!-- 未完成的任务显示取消按钮 -->
                <el-button 
                  v-if="row.state === 0 || row.state === 1 || row.state === 2"
                  link 
                  icon="Close" 
                  type="danger"
                  @click="cancelDownload(row.id)"
                >
                  取消
                </el-button>
                <!-- 已完成或失败的任务显示删除按钮 -->
                <el-button 
                  v-if="row.state === 3 || row.state === 4"
                  link 
                  icon="Delete" 
                  type="danger"
                  @click="deleteDownloadTask(row.id)"
                >
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          
          <el-empty v-if="downloadTasks.length === 0 && !downloadLoading" description="暂无下载任务" />
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { 
  getDownloadTaskList, 
  cancelDownload as cancelDownloadApi, 
  deleteDownload as deleteDownloadApi,
  pauseDownload,
  resumeDownload
} from '@/api/download'
import type { OfflineDownloadTask } from '@/api/download'
import { cleanExpiredUploads as cleanExpiredUploadsApi, deleteUploadTask, getUploadProgress } from '@/api/file'
import { formatSize, formatDate, formatSpeed, getUploadStatusType, getUploadStatusText, formatFileSizeForDisplay } from '@/utils'
import { isUploadTaskActive, openFileDialog, uploadSingleFile } from '@/utils/upload'
import { ElMessage } from 'element-plus'

const { proxy } = getCurrentInstance() as ComponentInternalInstance

const route = useRoute()

const activeTab = ref<string>((route.query.tab as string) || 'upload')
const uploadLoading = ref(false)
const downloadLoading = ref(false)
const cleanLoading = ref(false)
const uploadTasks = ref<any[]>([])
const downloadTasks = ref<OfflineDownloadTask[]>([])
let refreshTimer: number | null = null

// 监听路由查询参数变化
watch(() => route.query.tab, (newTab) => {
  if (newTab && (newTab === 'upload' || newTab === 'download')) {
    activeTab.value = newTab as string
  }
})

import { uploadTaskManager } from '@/utils/uploadTaskManager'
import { loadAndSyncBackendTasks, findBackendTask } from '@/utils/uploadTaskSync'

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

// 加载下载任务列表
const loadDownloadTasks = async () => {
  downloadLoading.value = true
  try {
    const res = await getDownloadTaskList({ page: 1, pageSize: 100 })
    if (res.code === 200 && res.data) {
      downloadTasks.value = res.data.tasks || []
    }
  } catch (error: any) {
    proxy?.$log.error('加载下载任务失败:', error)
    proxy?.$modal.msgError('加载下载任务失败')
  } finally {
    downloadLoading.value = false
  }
}

const pauseUpload = (taskId: string) => {
  uploadTaskManager.pauseTask(taskId)
  uploadTaskManager.cancelAllUploads(taskId)
  proxy?.$modal.msgSuccess('已暂停')
}

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

const cleanExpiredUploads = async () => {
  try {
    await proxy?.$modal.confirm('确认清理所有过期的上传任务？过期任务将无法恢复。')
    cleanLoading.value = true
    const res = await cleanExpiredUploadsApi()
    if (res.code === 200) {
      const count = res.data?.cleaned_count || 0
      if (count > 0) {
        proxy?.$modal.msgSuccess(`已清理 ${count} 个过期任务`)
        loadUploadTasks()
      } else {
        proxy?.$modal.msgSuccess('没有需要清理的过期任务')
      }
    } else {
      proxy?.$modal.msgError(res.message || '清理失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      proxy?.$modal.msgError('清理失败')
      proxy?.$log.error('清理过期任务失败:', error)
    }
  } finally {
    cleanLoading.value = false
  }
}

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

const getDownloadTypeColor = (type: number) => {
  const colorMap: Record<number, any> = {
    0: 'success',
    1: 'warning',
    2: 'warning',
    3: 'info',
    4: 'danger',
    5: 'danger',
    6: 'primary'
  }
  return colorMap[type] || 'info'
}

const getDownloadStatusType = (state: number) => {
  const typeMap: Record<number, any> = {
    0: 'info',
    1: 'primary',
    2: 'warning',
    3: 'success',
    4: 'danger'
  }
  return typeMap[state] || 'info'
}

onMounted(() => {
  loadUploadTasks()
  loadDownloadTasks()
  
  const unsubscribe = uploadTaskManager.subscribe((tasks) => {
    uploadTasks.value = tasks
  })
  
  let syncTimer: number | null = null
  const startAutoSync = () => {
    if (syncTimer) {
      clearInterval(syncTimer)
    }
    syncTimer = window.setInterval(() => {
      if (activeTab.value === 'upload') {
        loadUploadTasks()
      }
    }, 30000)
  }

  startAutoSync()
  
  refreshTimer = window.setInterval(() => {
    loadDownloadTasks()
  }, 3000)
  
  onBeforeUnmount(() => {
    unsubscribe()
    if (syncTimer) {
      clearInterval(syncTimer)
    }
  })
})

onBeforeUnmount(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.tasks-page {
  height: 100%;
}

.task-tabs {
  height: 100%;
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.progress-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.progress-info {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>

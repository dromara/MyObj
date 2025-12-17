<template>
  <div class="tasks-page">
    <el-tabs v-model="activeTab" class="task-tabs">
      <el-tab-pane label="上传任务" name="upload">
        <el-card shadow="never">
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
            
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button 
                  v-if="row.status === 'uploading' || row.status === 'pending'"
                  link 
                  :icon="Close" 
                  type="danger"
                  @click="cancelUpload(row.id)"
                >
                  取消
                </el-button>
                <el-button 
                  v-if="row.status === 'completed' || row.status === 'failed'"
                  link 
                  :icon="Delete" 
                  type="danger"
                  @click="deleteUpload(row.id)"
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
                  :icon="VideoPause" 
                  type="warning"
                  @click="pauseDownloadTask(row.id)"
                >
                  暂停
                </el-button>
                <el-button 
                  v-if="row.type !== 7 && row.state === 2"
                  link 
                  :icon="VideoPlay" 
                  type="primary"
                  @click="resumeDownloadTask(row.id)"
                >
                  继续
                </el-button>
                <!-- 未完成的任务显示取消按钮 -->
                <el-button 
                  v-if="row.state === 0 || row.state === 1 || row.state === 2"
                  link 
                  :icon="Close" 
                  type="danger"
                  @click="cancelDownload(row.id)"
                >
                  取消
                </el-button>
                <!-- 已完成或失败的任务显示删除按钮 -->
                <el-button 
                  v-if="row.state === 3 || row.state === 4"
                  link 
                  :icon="Delete" 
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
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, Close, Delete, VideoPause, VideoPlay, Download } from '@element-plus/icons-vue'
import { get } from '@/utils/request'
import { API_ENDPOINTS } from '@/config/api'
import { 
  getDownloadTaskList, 
  cancelDownload as cancelDownloadApi, 
  deleteDownload as deleteDownloadApi,
  pauseDownload,
  resumeDownload,
  getLocalFileDownloadUrl
} from '@/api/download'
import type { OfflineDownloadTask } from '@/api/download'

const activeTab = ref('upload')
const uploadLoading = ref(false)
const downloadLoading = ref(false)
const uploadTasks = ref<any[]>([])
const downloadTasks = ref<OfflineDownloadTask[]>([])
let refreshTimer: number | null = null

// 加载上传任务列表
const loadUploadTasks = async () => {
  uploadLoading.value = true
  try {
    const res = await get(API_ENDPOINTS.TASK.UPLOAD_LIST)
    if (res.code === 200 && res.data) {
      uploadTasks.value = res.data.tasks || []
    }
  } catch (error: any) {
    console.error('加载上传任务失败:', error)
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
    console.error('加载下载任务失败:', error)
    ElMessage.error('加载下载任务失败')
  } finally {
    downloadLoading.value = false
  }
}

// 取消上传任务
const cancelUpload = async (taskId: string) => {
  try {
    await ElMessageBox.confirm('确认取消该上传任务?', '提示', { type: 'warning' })
    // TODO: 调用取消上传API
    ElMessage.success('已取消')
    loadUploadTasks()
  } catch (error) {
    // 用户取消操作
  }
}

// 删除上传任务
const deleteUpload = async (taskId: string) => {
  try {
    await ElMessageBox.confirm('确认删除该任务记录?', '提示', { type: 'warning' })
    // TODO: 调用删除任务API
    ElMessage.success('已删除')
    loadUploadTasks()
  } catch (error) {
    // 用户取消操作
  }
}

// 取消下载任务
const cancelDownload = async (taskId: string) => {
  try {
    await ElMessageBox.confirm('确认取消该下载任务?', '提示', { type: 'warning' })
    const res = await cancelDownloadApi(taskId)
    if (res.code === 200) {
      ElMessage.success('已取消')
      loadDownloadTasks()
    } else {
      ElMessage.error(res.msg || '取消失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('取消失败')
    }
  }
}

// 删除下载任务
const deleteDownloadTask = async (taskId: string) => {
  try {
    await ElMessageBox.confirm('确认删除该任务记录?', '提示', { type: 'warning' })
    const res = await deleteDownloadApi(taskId)
    if (res.code === 200) {
      ElMessage.success('已删除')
      loadDownloadTasks()
    } else {
      ElMessage.error(res.msg || '删除失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 下载网盘文件
const downloadLocalFile = async (taskId: string) => {
  try {
    const token = localStorage.getItem('token')
    const downloadUrl = getLocalFileDownloadUrl(taskId)
    
    ElMessage.info('下载中，请稍候...')
    
    const response = await fetch(downloadUrl, {
      method: 'GET',
      headers: {
        'Authorization': token ? `Bearer ${token}` : ''
      }
    })
    
    if (!response.ok) {
      throw new Error('下载失败')
    }
    
    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    // 从任务信息中获取文件名
    const task = downloadTasks.value.find(t => t.id === taskId)
    link.download = task?.file_name || 'download'
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    
    ElMessage.success('下载完成')
  } catch (error: any) {
    console.error('下载文件失败:', error)
    ElMessage.error('下载失败: ' + (error.message || '未知错误'))
  }
}

// 获取上传任务状态类型
const getUploadStatusType = (status: string) => {
  const typeMap: Record<string, any> = {
    pending: 'info',
    uploading: 'primary',
    completed: 'success',
    failed: 'danger'
  }
  return typeMap[status] || 'info'
}

// 获取上传任务状态文本
const getUploadStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    pending: '等待中',
    uploading: '上传中',
    completed: '已完成',
    failed: '失败'
  }
  return textMap[status] || '未知'
}

// 暂停下载任务
const pauseDownloadTask = async (taskId: string) => {
  try {
    const res = await pauseDownload(taskId)
    if (res.code === 200) {
      ElMessage.success('已暂停')
      loadDownloadTasks()
    } else {
      ElMessage.error(res.msg || '暂停失败')
    }
  } catch (error: any) {
    ElMessage.error('暂停失败')
  }
}

// 恢复下载任务
const resumeDownloadTask = async (taskId: string) => {
  try {
    const res = await resumeDownload(taskId)
    if (res.code === 200) {
      ElMessage.success('已恢复')
      loadDownloadTasks()
    } else {
      ElMessage.error(res.msg || '恢复失败')
    }
  } catch (error: any) {
    ElMessage.error('恢复失败')
  }
}

// 获取下载任务类型颜色
const getDownloadTypeColor = (type: number) => {
  const colorMap: Record<number, any> = {
    0: 'success',   // HTTP - 绿色
    1: 'warning',   // FTP - 橙色
    2: 'warning',   // SFTP - 橙色
    3: 'info',      // S3 - 灰色
    4: 'danger',    // 种子 - 红色
    5: 'danger',    // 磁力链接 - 红色
    6: 'primary'    // 本地文件 - 蓝色
  }
  return colorMap[type] || 'info'
}

// 获取下载任务状态类型
const getDownloadStatusType = (state: number) => {
  const typeMap: Record<number, any> = {
    0: 'info',      // 初始化
    1: 'primary',   // 下载中
    2: 'warning',   // 暂停
    3: 'success',   // 完成
    4: 'danger'     // 失败
  }
  return typeMap[state] || 'info'
}

// 格式化速度
const formatSpeed = (speed: number): string => {
  if (!speed || speed === 0) return '0 B/s'
  const k = 1024
  const sizes = ['B/s', 'KB/s', 'MB/s', 'GB/s']
  const i = Math.floor(Math.log(speed) / Math.log(k))
  return (speed / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

// 格式化文件大小
const formatSize = (bytes: number): string => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

// 格式化日期
const formatDate = (dateStr: string): string => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

// 页面加载时获取任务列表
onMounted(() => {
  loadUploadTasks()
  loadDownloadTasks()
  
  // 每 3 秒自动刷新
  refreshTimer = window.setInterval(() => {
    loadUploadTasks()
    loadDownloadTasks()
  }, 3000)
})

// 页面销毁时清除定时器
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

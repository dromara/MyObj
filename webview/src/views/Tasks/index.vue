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
                <el-tag :type="getStatusType(row.status)">{{ getStatusText(row.status) }}</el-tag>
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
                  <span>{{ row.file_name }}</span>
                </div>
              </template>
            </el-table-column>
            
            <el-table-column label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">{{ getStatusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
            
            <el-table-column label="进度" width="250">
              <template #default="{ row }">
                <div class="progress-cell">
                  <el-progress 
                    :percentage="row.progress" 
                    :status="row.status === 'completed' ? 'success' : row.status === 'failed' ? 'exception' : undefined"
                  />
                  <span class="progress-info">{{ formatSize(row.downloaded_size) }} / {{ formatSize(row.file_size) }} · {{ row.speed || '0 KB/s' }}</span>
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
                  v-if="row.status === 'downloading' || row.status === 'pending'"
                  link 
                  :icon="Close" 
                  type="danger"
                  @click="cancelDownload(row.id)"
                >
                  取消
                </el-button>
                <el-button 
                  v-if="row.status === 'completed' || row.status === 'failed'"
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
import { Document, Close, Delete } from '@element-plus/icons-vue'
import { get } from '@/utils/request'
import { API_ENDPOINTS } from '@/config/api'

const activeTab = ref('upload')
const uploadLoading = ref(false)
const downloadLoading = ref(false)
const uploadTasks = ref<any[]>([])
const downloadTasks = ref<any[]>([])
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
    const res = await get(API_ENDPOINTS.TASK.DOWNLOAD_LIST)
    if (res.code === 200 && res.data) {
      downloadTasks.value = res.data.tasks || []
    }
  } catch (error: any) {
    console.error('加载下载任务失败:', error)
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
    // TODO: 调用取消下载API
    ElMessage.success('已取消')
    loadDownloadTasks()
  } catch (error) {
    // 用户取消操作
  }
}

// 删除下载任务
const deleteDownloadTask = async (taskId: string) => {
  try {
    await ElMessageBox.confirm('确认删除该任务记录?', '提示', { type: 'warning' })
    // TODO: 调用删除任务API
    ElMessage.success('已删除')
    loadDownloadTasks()
  } catch (error) {
    // 用户取消操作
  }
}

// 获取状态类型
const getStatusType = (status: string) => {
  const typeMap: Record<string, any> = {
    pending: 'info',
    uploading: 'primary',
    downloading: 'primary',
    completed: 'success',
    failed: 'danger'
  }
  return typeMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    pending: '等待中',
    uploading: '上传中',
    downloading: '下载中',
    completed: '已完成',
    failed: '失败'
  }
  return textMap[status] || '未知'
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

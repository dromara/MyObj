<template>
  <div class="tasks-page">
    <el-tabs v-model="activeTab" class="task-tabs">
      <el-tab-pane label="上传任务" name="upload">
        <UploadTaskTable
          :tasks="uploadTasks"
          :loading="uploadLoading"
          :clean-loading="cleanLoading"
          :expired-count="expiredTaskCount"
          @pause="pauseUpload"
          @resume="resumeUpload"
          @cancel="cancelUpload"
          @delete="deleteUpload"
          @view-expired="showExpiredDialog = true"
        />
        <ExpiredTasksDialog
          v-model="showExpiredDialog"
          @refresh="handleExpiredRefresh"
        />
      </el-tab-pane>
      
      <el-tab-pane label="下载任务" name="download">
        <DownloadTaskTable
          :tasks="downloadTasks"
          :loading="downloadLoading"
          @pause="pauseDownloadTask"
          @resume="resumeDownloadTask"
          @cancel="cancelDownload"
          @delete="deleteDownloadTask"
        />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import UploadTaskTable from './modules/UploadTaskTable.vue'
import DownloadTaskTable from './modules/DownloadTaskTable.vue'
import ExpiredTasksDialog from './modules/ExpiredTasksDialog.vue'
import { useUploadTasks } from './modules/useUploadTasks'
import { useDownloadTasks } from './modules/useDownloadTasks'

const route = useRoute()

const activeTab = ref<string>((route.query.tab as string) || 'upload')
let refreshTimer: number | null = null
let syncTimer: number | null = null

// 监听路由查询参数变化
watch(() => route.query.tab, (newTab) => {
  if (newTab && (newTab === 'upload' || newTab === 'download')) {
    activeTab.value = newTab as string
  }
})

// 使用 composables
const {
  uploadTasks,
  uploadLoading,
  cleanLoading,
  expiredTaskCount,
  currentPage,
  pageSize,
  loadUploadTasks,
  getExpiredTaskCount,
  pauseUpload,
  resumeUpload,
  cancelUpload,
  deleteUpload,
  initTaskSubscription
} = useUploadTasks()

const showExpiredDialog = ref(false)

// 处理过期任务刷新
const handleExpiredRefresh = () => {
  loadUploadTasks(currentPage.value, pageSize.value)
  getExpiredTaskCount()
}

const {
  downloadTasks,
  downloadLoading,
  loadDownloadTasks,
  cancelDownload,
  deleteDownloadTask,
  pauseDownloadTask,
  resumeDownloadTask
} = useDownloadTasks()

onMounted(() => {
  loadUploadTasks()
  loadDownloadTasks()
  getExpiredTaskCount() // 加载过期任务数量
  
  // 初始化本地任务订阅（用于实时显示临时任务）
  initTaskSubscription()
  
  // 启动自动刷新（30秒）
  const startAutoSync = () => {
    if (syncTimer) {
      clearInterval(syncTimer)
    }
    syncTimer = window.setInterval(() => {
      if (activeTab.value === 'upload') {
        loadUploadTasks(currentPage.value, pageSize.value)
        getExpiredTaskCount() // 定期更新过期任务数量
      }
    }, 30000)
  }

  startAutoSync()
  
  // 启动下载任务自动刷新（3秒，智能刷新不显示loading）
  refreshTimer = window.setInterval(() => {
    // 自动刷新时不显示loading
    loadDownloadTasks(false)
  }, 3000)
})

onBeforeUnmount(() => {
  if (syncTimer) {
    clearInterval(syncTimer)
  }
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

/* 移动端响应式 - 组件特定样式 */
@media (max-width: 1024px) {
  /* 任务标签页特定样式 */
  .task-tabs :deep(.el-tabs__header) {
    margin-bottom: 12px;
  }
  
  .task-tabs :deep(.el-tabs__item) {
    padding: 0 12px;
    font-size: 14px;
  }
}
</style>

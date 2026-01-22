<template>
  <div class="tasks-page">
    <el-tabs v-model="activeTab" class="task-tabs">
      <el-tab-pane :label="t('tasks.upload')" name="upload">
        <UploadTaskTable
          :tasks="uploadTasks"
          :loading="uploadLoading"
          :clean-loading="cleanLoading"
          :expired-count="expiredTaskCount"
          :current-page="uploadCurrentPage"
          :page-size="uploadPageSize"
          :total="uploadTotal"
          @pause="pauseUpload"
          @resume="resumeUpload"
          @cancel="cancelUpload"
          @delete="deleteUpload"
          @view-expired="showExpiredDialog = true"
          @clear-all="clearAllUploadTasks"
          @pagination="handleUploadPagination"
        />
        <ExpiredTasksDialog v-model="showExpiredDialog" @refresh="handleExpiredRefresh" />
      </el-tab-pane>

      <el-tab-pane :label="t('tasks.download')" name="download">
        <DownloadTaskTable
          :tasks="downloadTasks"
          :loading="downloadLoading"
          :current-page="downloadCurrentPage"
          :page-size="downloadPageSize"
          :total="downloadTotal"
          @pause="pauseDownloadTask"
          @resume="resumeDownloadTask"
          @cancel="cancelDownload"
          @delete="deleteDownloadTask"
          @pagination="handleDownloadPagination"
        />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
  import { uploadTaskManager } from '@/utils/file/uploadTaskManager'
  import { useI18n } from '@/composables'
  import UploadTaskTable from './components/UploadTaskTable.vue'
  import DownloadTaskTable from './components/DownloadTaskTable.vue'
  import ExpiredTasksDialog from './components/ExpiredTasksDialog.vue'
  import { useUploadTasks } from './composables/useUploadTasks'
  import { useDownloadTasks } from './composables/useDownloadTasks'

  const { t } = useI18n()

  const route = useRoute()

  const activeTab = ref<string>((route.query.tab as string) || 'upload')
  let refreshTimer: number | null = null
  let syncTimer: number | null = null

  // 监听路由查询参数变化
  watch(
    () => route.query.tab,
    newTab => {
      if (newTab && (newTab === 'upload' || newTab === 'download')) {
        activeTab.value = newTab as string
      }
    }
  )

  // 使用 composables
  const {
    uploadTasks,
    uploadLoading,
    cleanLoading,
    expiredTaskCount,
    currentPage: uploadCurrentPage,
    pageSize: uploadPageSize,
    total: uploadTotal,
    loadUploadTasks,
    getExpiredTaskCount,
    pauseUpload,
    resumeUpload,
    cancelUpload,
    deleteUpload,
    clearAllUploadTasks,
    handlePagination: handleUploadPagination
  } = useUploadTasks()

  const showExpiredDialog = ref(false)

  // 处理过期任务刷新
  const handleExpiredRefresh = () => {
    loadUploadTasks()
    getExpiredTaskCount()
  }

  const {
    downloadTasks,
    downloadLoading,
    currentPage: downloadCurrentPage,
    pageSize: downloadPageSize,
    total: downloadTotal,
    loadDownloadTasks,
    cancelDownload,
    deleteDownloadTask,
    pauseDownloadTask,
    resumeDownloadTask
  } = useDownloadTasks()

  // 处理下载任务分页
  const handleDownloadPagination = ({ page, limit }: { page: number; limit: number }) => {
    loadDownloadTasks(true, page, limit)
  }

  // 订阅上传任务更新
  let unsubscribe: (() => void) | null = null

  onMounted(() => {
    // 初始化上传任务管理器（确保用户隔离）
    uploadTaskManager.init()
    loadUploadTasks(true, true) // 初始加载，强制同步后端
    loadDownloadTasks(true, 1, 20) // 初始加载，第一页，每页20条
    getExpiredTaskCount() // 加载过期任务数量

    // 订阅上传任务更新（保持当前分页）
    // 注意：不强制同步，使用防抖机制
    unsubscribe = uploadTaskManager.subscribe(() => {
      // 重新加载以更新分页数据，保持当前页
      // forceSync = false，使用防抖机制，避免频繁调用后端接口
      loadUploadTasks(false, false)
    })

    // 启动自动同步（30秒）
    const startAutoSync = () => {
      if (syncTimer) {
        clearInterval(syncTimer)
      }
      syncTimer = window.setInterval(() => {
        if (activeTab.value === 'upload') {
          loadUploadTasks(false, true) // 定时任务强制同步
          getExpiredTaskCount() // 定期更新过期任务数量
        }
      }, 30000)
    }

    startAutoSync()

    // 启动下载任务自动刷新（3秒，智能刷新不显示loading）
    refreshTimer = window.setInterval(() => {
      // 自动刷新时不显示loading，保持当前分页
      if (activeTab.value === 'download') {
        loadDownloadTasks(false, downloadCurrentPage.value, downloadPageSize.value)
      }
    }, 3000)
  })

  onBeforeUnmount(() => {
    if (unsubscribe) {
      unsubscribe()
    }
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
    display: flex;
    flex-direction: column;
  }

  .task-tabs :deep(.el-tabs__content) {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .task-tabs :deep(.el-tab-pane) {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
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

  /* 深色模式样式 */
  html.dark .task-tabs :deep(.el-tabs__header) {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .task-tabs :deep(.el-tabs__item) {
    color: var(--el-text-color-primary);
    border-color: var(--el-border-color);
  }

  html.dark .task-tabs :deep(.el-tabs__item.is-active) {
    color: var(--primary-color);
    border-bottom-color: var(--primary-color);
  }

  html.dark .task-tabs :deep(.el-tabs__nav-wrap::after) {
    background-color: var(--el-border-color);
  }
</style>

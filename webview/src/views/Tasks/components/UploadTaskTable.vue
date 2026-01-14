<template>
  <el-card shadow="never" class="task-card">
    <div class="card-header">
      <span class="task-count">{{ t('tasks.taskCount', { count: tasks.length }) }}</span>
      <div class="header-actions">
        <el-button
          v-if="tasks.length > 0"
          type="danger"
          size="small"
          icon="Delete"
          @click="$emit('clear-all')"
          :loading="clearAllLoading"
          class="clear-all-btn"
        >
          {{ t('tasks.clearAll') }}
        </el-button>
        <el-button type="warning" size="small" icon="View" @click="$emit('view-expired')" class="expired-btn">
          {{ t('tasks.viewExpired')
          }}<el-badge v-if="(expiredCount || 0) > 0" :value="expiredCount" class="expired-badge" />
        </el-button>
      </div>
    </div>

    <!-- PC端：表格布局 -->
    <div class="table-wrapper" v-if="props.tasks.length > 0">
      <el-table :data="tasks" v-loading="loading" class="task-table desktop-table">
        <el-table-column :label="t('tasks.fileName')" min-width="200" class-name="mobile-name-column">
          <template #default="{ row }">
            <div class="file-name-cell">
              <el-icon :size="24" class="upload-task-icon"><Document /></el-icon>
              <file-name-tooltip :file-name="row.file_name" view-mode="table" />
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('tasks.status')" min-width="100" class-name="mobile-hide">
          <template #default="{ row }">
            <div class="status-cell">
              <el-tag :type="getUploadStatusType(row.status)">{{ getUploadStatusText(row.status) }}</el-tag>
              <!-- 秒传标识 -->
              <el-tag v-if="row.isInstantUpload && row.status === 'completed'" type="success" size="small" class="instant-tag">
                {{ t('tasks.instantUpload') }}
              </el-tag>
              <!-- 预检中时显示当前步骤 -->
              <span v-if="row.status === 'prechecking' && row.currentStep" class="current-step">
                {{ row.currentStep }}
              </span>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('tasks.progress')" min-width="200" class-name="mobile-progress-column">
          <template #default="{ row }">
            <div class="progress-cell">
              <el-progress
                :percentage="Math.max(0, Math.min(100, row.status === 'prechecking' ? (row.precheckProgress || 0) : (row.progress || 0)))"
                :status="row.status === 'completed' ? 'success' : row.status === 'failed' ? 'exception' : undefined"
                :color="row.status === 'prechecking' ? '#409EFF' : undefined"
              />
              <span class="progress-info">
                <template v-if="row.status === 'prechecking'">
                  {{ t('tasks.prechecking') }} - {{ row.precheckProgress || 0 }}%
                </template>
                <template v-else>
                  {{ formatSize(row.uploaded_size) }} / {{ formatSize(row.file_size) }} · {{ row.speed || '0 KB/s' }}
                </template>
              </span>
              <!-- 已完成任务显示总耗时和平均速度 -->
              <div v-if="row.status === 'completed'" class="completed-info">
                <span v-if="row.totalDuration" class="duration-info">
                  {{ t('tasks.totalDuration') }}: {{ formatDuration(row.totalDuration) }}
                </span>
                <span v-if="row.averageSpeed" class="avg-speed-info">
                  {{ t('tasks.averageSpeed') }}: {{ formatSpeed(row.averageSpeed) }}
                </span>
                <el-tag v-if="row.isInstantUpload" type="success" size="small" class="instant-upload-tag">
                  {{ t('tasks.instantUpload') }}
                </el-tag>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('tasks.createTime')" min-width="150" class-name="mobile-hide">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('tasks.operation')" width="180" fixed="right" class-name="mobile-actions-column">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'uploading'"
              link
              icon="VideoPause"
              type="warning"
              @click="$emit('pause', row.id)"
            >
              {{ t('tasks.pause') }}
            </el-button>
            <el-button
              v-if="row.status === 'paused'"
              link
              icon="VideoPlay"
              type="primary"
              @click="$emit('resume', row.id)"
            >
              {{ t('tasks.resume') }}
            </el-button>
            <el-button
              v-if="row.status === 'uploading' || row.status === 'pending' || row.status === 'paused' || row.status === 'prechecking'"
              link
              icon="Close"
              type="danger"
              @click="$emit('cancel', row.id)"
            >
              {{ t('tasks.cancel') }}
            </el-button>
            <el-button
              link
              icon="Delete"
              type="danger"
              @click="$emit('delete', row.id)"
              :disabled="row.status === 'uploading'"
            >
              {{ t('tasks.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 移动端：卡片布局 -->
    <div class="mobile-task-list" v-loading="loading">
      <div v-for="row in tasks" :key="row.id" class="mobile-task-item">
        <div class="task-item-header">
          <div class="task-item-info">
            <el-icon :size="20" class="task-icon upload-task-icon"><Document /></el-icon>
            <div class="task-name-wrapper">
              <file-name-tooltip :file-name="row.file_name" view-mode="list" custom-class="task-name" />
              <div class="task-meta">
                <div class="task-status-row">
                  <el-tag :type="getUploadStatusType(row.status)" size="small" effect="plain">
                    {{ getUploadStatusText(row.status) }}
                  </el-tag>
                  <el-tag v-if="row.isInstantUpload && row.status === 'completed'" type="success" size="small" class="instant-tag-mobile">
                    {{ t('tasks.instantUpload') }}
                  </el-tag>
                </div>
                <span class="task-size">{{ formatSize(row.uploaded_size) }} / {{ formatSize(row.file_size) }}</span>
              </div>
            </div>
          </div>
          <div class="task-actions">
            <el-button
              v-if="row.status === 'uploading'"
              link
              type="warning"
              @click.stop="$emit('pause', row.id)"
              class="action-btn"
            >
              <el-icon><VideoPause /></el-icon>
            </el-button>
            <el-button
              v-if="row.status === 'paused'"
              link
              type="primary"
              @click.stop="$emit('resume', row.id)"
              class="action-btn"
            >
              <el-icon><VideoPlay /></el-icon>
            </el-button>
            <el-button
              v-if="row.status === 'uploading' || row.status === 'pending' || row.status === 'paused'"
              link
              type="danger"
              @click.stop="$emit('cancel', row.id)"
              class="action-btn"
            >
              <el-icon><Close /></el-icon>
            </el-button>
            <el-button
              link
              type="danger"
              @click.stop="$emit('delete', row.id)"
              :disabled="row.status === 'uploading'"
              class="action-btn"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>
        <div class="task-progress-wrapper">
          <el-progress
            :percentage="Math.max(0, Math.min(100, row.status === 'prechecking' ? (row.precheckProgress || 0) : (row.progress || 0)))"
            :status="row.status === 'completed' ? 'success' : row.status === 'failed' ? 'exception' : undefined"
            :stroke-width="4"
            :color="row.status === 'prechecking' ? '#409EFF' : undefined"
            class="task-progress"
          />
          <div class="task-speed" v-if="row.status === 'uploading'">
            {{ row.speed || '0 KB/s' }}
          </div>
          <div v-if="row.status === 'prechecking' && row.currentStep" class="task-step">
            {{ row.currentStep }}
          </div>
          <!-- 已完成任务显示总耗时和平均速度 -->
          <div v-if="row.status === 'completed'" class="completed-info-mobile">
            <span v-if="row.totalDuration" class="duration-info">
              {{ t('tasks.totalDuration') }}: {{ formatDuration(row.totalDuration) }}
            </span>
            <span v-if="row.averageSpeed" class="avg-speed-info">
              {{ t('tasks.averageSpeed') }}: {{ formatSpeed(row.averageSpeed) }}
            </span>
            <el-tag v-if="row.isInstantUpload" type="success" size="small" class="instant-upload-tag">
              {{ t('tasks.instantUpload') }}
            </el-tag>
          </div>
        </div>
      </div>
    </div>

    <EmptyState
      v-if="tasks.length === 0 && !loading"
      type="task"
      :show-actions="false"
      compact
    />

    <!-- 分页 -->
    <pagination
      v-if="(total || 0) > 0"
      v-model:page="currentPage"
      v-model:limit="pageSize"
      :total="total || 0"
      :page-sizes="[20, 50, 100]"
      @pagination="handlePagination"
      class="pagination"
    />
  </el-card>
</template>

<script setup lang="ts">
  import { formatSize, formatDate, formatSpeed, formatDuration, getUploadStatusType, getUploadStatusText } from '@/utils'
  import { useI18n } from '@/composables'
  import EmptyState from '@/components/EmptyState/index.vue'

  const { t } = useI18n()

  const props = defineProps<{
    tasks: any[]
    loading: boolean
    cleanLoading: boolean
    expiredCount?: number
    currentPage?: number
    pageSize?: number
    total?: number
  }>()

  const emit = defineEmits<{
    pause: [taskId: string]
    resume: [taskId: string]
    cancel: [taskId: string]
    delete: [taskId: string]
    'view-expired': []
    'clear-all': []
    pagination: [{ page: number; limit: number }]
  }>()

  const clearAllLoading = computed(() => props.cleanLoading)

  const currentPage = ref(props.currentPage || 1)
  const pageSize = ref(props.pageSize || 20)

  watch(
    () => props.currentPage,
    val => {
      if (val !== undefined) {
        currentPage.value = val
      }
    }
  )

  watch(
    () => props.pageSize,
    val => {
      if (val !== undefined) {
        pageSize.value = val
      }
    }
  )

  const handlePagination = ({ page, limit }: { page: number; limit: number }) => {
    currentPage.value = page
    pageSize.value = limit
    emit('pagination', { page, limit })
  }
</script>

<style scoped>
  .task-card {
    padding: 0;
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .task-card :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
    padding: 0;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  .upload-task-icon {
    color: var(--el-color-primary);
  }

  .task-count {
    font-size: 14px;
    color: var(--el-text-color-regular);
  }

  .expired-badge {
    margin-left: 4px;
  }

  .header-actions {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .clear-all-btn {
    margin-right: 8px;
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

  .completed-info {
    margin-top: 8px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 12px;
    color: var(--el-text-color-regular);
  }

  .duration-info,
  .avg-speed-info {
    color: var(--el-text-color-secondary);
  }

  .instant-upload-tag {
    margin-top: 4px;
    align-self: flex-start;
  }

  .instant-tag {
    margin-left: 8px;
  }

  .status-cell {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .table-wrapper {
    flex: 1;
    min-height: 0;
    overflow-x: auto;
    overflow-y: hidden;
    padding: 8px;
  }

  /* PC端表格样式 */
  .desktop-table {
    display: table;
  }

  .task-table :deep(.el-table) {
    background: transparent;
    width: 100%;
    table-layout: auto;
  }

  .task-table :deep(.el-table__header-wrapper) {
    overflow-x: hidden;
  }

  .task-table :deep(.el-table__body-wrapper) {
    overflow-x: hidden;
  }

  /* 当数据为空时，表格不显示 */
  .task-table :deep(.el-table__empty-block) {
    display: none;
  }

  .task-table :deep(.el-table) {
    background: transparent !important;
    --el-table-tr-bg-color: transparent;
    --el-table-header-bg-color: transparent;
  }

  .task-table :deep(.el-table th.el-table__cell) {
    background: transparent !important;
    color: var(--el-text-color-primary);
    font-weight: 600;
    font-size: 13px;
    border-bottom-color: var(--el-border-color-lighter);
  }

  .task-table :deep(.el-table td.el-table__cell) {
    background: transparent !important;
    color: var(--el-text-color-primary);
    border-bottom-color: var(--el-border-color-lighter);
  }

  .task-table :deep(.el-table tr) {
    background: transparent !important;
    transition: all 0.2s;
  }

  .task-table :deep(.el-table--enable-row-hover .el-table__body tr:hover > td.el-table__cell) {
    background: var(--el-fill-color-lighter) !important;
  }

  .task-table :deep(.mobile-hide) {
    display: table-cell;
  }

  .task-table :deep(.mobile-name-column) {
    min-width: 200px;
  }

  .task-table :deep(.mobile-progress-column) {
    min-width: 200px;
  }

  .task-table :deep(.mobile-actions-column) {
    width: auto;
    min-width: 120px;
  }

  /* 移动端卡片列表 */
  .mobile-task-list {
    display: none;
  }

  .mobile-task-item {
    padding: 12px 16px;
    border-bottom: 1px solid var(--el-border-color-lighter);
    background: var(--el-bg-color, var(--card-bg));
    transition: background-color 0.2s;
  }

  .mobile-task-item:active {
    background-color: var(--el-fill-color-light);
  }

  .task-item-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 8px;
  }

  .task-item-info {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    flex: 1;
    min-width: 0;
  }

  .task-icon {
    flex-shrink: 0;
    margin-top: 2px;
  }

  .task-name-wrapper {
    flex: 1;
    min-width: 0;
  }

  .task-status-row {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
  }

  .instant-tag-mobile {
    margin-left: 4px;
  }

  .completed-info-mobile {
    margin-top: 8px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 12px;
    color: var(--el-text-color-regular);
  }

  .task-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--el-text-color-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin-bottom: 4px;
  }

  .task-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .task-size {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .task-actions {
    display: flex;
    align-items: center;
    gap: 4px;
    flex-shrink: 0;
    margin-left: 8px;
  }

  .action-btn {
    padding: 4px;
    min-width: auto;
  }

  .action-btn :deep(.el-icon) {
    font-size: 18px;
  }

  .task-progress-wrapper {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 8px;
  }

  .task-progress {
    flex: 1;
  }

  .task-speed {
    font-size: 12px;
    color: var(--el-color-primary);
    font-weight: 500;
    white-space: nowrap;
    min-width: 60px;
    text-align: right;
  }

  /* 移动端响应式 */
  @media (max-width: 1024px) {
    .desktop-table {
      display: none !important;
    }

    .mobile-task-list {
      display: block;
    }

    .card-header {
      padding: 12px;
      flex-wrap: wrap;
      gap: 8px;
    }

    .expired-btn {
      font-size: 12px;
      padding: 6px 12px;
    }
  }

  @media (max-width: 480px) {
    .mobile-task-item {
      padding: 10px 12px;
    }

    .task-name {
      font-size: 13px;
    }

    .task-meta {
      font-size: 11px;
    }

    .task-speed {
      font-size: 11px;
      min-width: 50px;
    }
  }

  .table-wrapper {
    flex: 1;
    min-height: 0;
    overflow-x: auto;
    overflow-y: hidden;
    padding: 8px;
  }

  .pagination {
    margin-top: 16px;
    padding: 16px;
    border-top: 1px solid var(--el-border-color-lighter);
    flex-shrink: 0;
    display: flex;
    justify-content: center;
  }

  @media (max-width: 1024px) {
    .pagination {
      padding: 12px;
    }
  }

  /* 深色模式样式 */
  html.dark .task-card {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .task-card :deep(.el-card__body) {
    background: var(--card-bg);
  }

  html.dark .card-header {
    border-bottom-color: var(--el-border-color);
  }

  html.dark .task-count {
    color: var(--el-text-color-primary);
  }

  html.dark .pagination {
    border-top-color: var(--el-border-color);
  }

  html.dark .mobile-task-item {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .mobile-task-item:active {
    background-color: rgba(59, 130, 246, 0.1);
  }
</style>

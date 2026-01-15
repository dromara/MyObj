<template>
  <el-card shadow="never" class="task-card">
    <!-- PC端：表格布局 -->
    <div class="table-wrapper" v-if="props.tasks.length > 0">
      <el-table :data="tasks" v-loading="loading" class="task-table desktop-table">
        <el-table-column :label="t('tasks.fileName')" min-width="200" class-name="mobile-name-column">
          <template #default="{ row }">
            <div class="file-name-cell">
              <el-icon :size="24" class="download-task-icon"><Document /></el-icon>
              <file-name-tooltip :file-name="row.file_name || row.url" view-mode="table" />
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('tasks.type')" min-width="90" class-name="mobile-hide">
          <template #default="{ row }">
            <el-tag :type="getDownloadTypeColor(row.type)" effect="plain">{{ row.type_text }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('tasks.status')" min-width="100" class-name="mobile-hide">
          <template #default="{ row }">
            <el-tag :type="getDownloadStatusType(row.state)">{{ row.state_text }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('tasks.progress')" min-width="200" class-name="mobile-progress-column">
          <template #default="{ row }">
            <div class="progress-cell">
              <el-progress
                :percentage="row.progress"
                :status="row.state === 3 ? 'success' : row.state === 4 ? 'exception' : undefined"
              />
              <span class="progress-info"
                >{{ formatSize(row.downloaded_size) }} / {{ formatSize(row.file_size) }} ·
                {{ formatSpeed(row.speed) }}</span
              >
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('tasks.createTime')" min-width="150" class-name="mobile-hide">
          <template #default="{ row }">
            {{ formatDate(row.create_time) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('tasks.operation')" width="180" fixed="right" class-name="mobile-actions-column">
          <template #default="{ row }">
            <el-button
              v-if="row.type !== 7 && row.state === 1"
              link
              icon="VideoPause"
              type="warning"
              @click="$emit('pause', row.id)"
            >
              {{ t('tasks.pause') }}
            </el-button>
            <el-button
              v-if="row.type !== 7 && row.state === 2"
              link
              icon="VideoPlay"
              type="primary"
              @click="$emit('resume', row.id)"
            >
              {{ t('tasks.resume') }}
            </el-button>
            <el-button
              v-if="row.state === 0 || row.state === 1 || row.state === 2"
              link
              icon="Close"
              type="danger"
              @click="$emit('cancel', row.id)"
            >
              {{ t('tasks.cancel') }}
            </el-button>
            <el-button
              v-if="row.state === 3 || row.state === 4"
              link
              icon="Delete"
              type="danger"
              @click="$emit('delete', row.id)"
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
            <el-icon :size="20" class="task-icon download-task-icon"><Document /></el-icon>
            <div class="task-name-wrapper">
              <file-name-tooltip :file-name="row.file_name || row.url" view-mode="list" custom-class="task-name" />
              <div class="task-meta">
                <el-tag :type="getDownloadStatusType(row.state)" size="small" effect="plain">
                  {{ row.state_text }}
                </el-tag>
                <el-tag v-if="row.type !== 7" :type="getDownloadTypeColor(row.type)" size="small" effect="plain">
                  {{ row.type_text }}
                </el-tag>
                <span class="task-size">{{ formatSize(row.downloaded_size) }} / {{ formatSize(row.file_size) }}</span>
              </div>
            </div>
          </div>
          <div class="task-actions">
            <el-button
              v-if="row.type !== 7 && row.state === 1"
              link
              type="warning"
              @click.stop="$emit('pause', row.id)"
              class="action-btn"
            >
              <el-icon><VideoPause /></el-icon>
            </el-button>
            <el-button
              v-if="row.type !== 7 && row.state === 2"
              link
              type="primary"
              @click.stop="$emit('resume', row.id)"
              class="action-btn"
            >
              <el-icon><VideoPlay /></el-icon>
            </el-button>
            <el-button
              v-if="row.state === 0 || row.state === 1 || row.state === 2"
              link
              type="danger"
              @click.stop="$emit('cancel', row.id)"
              class="action-btn"
            >
              <el-icon><Close /></el-icon>
            </el-button>
            <el-button
              v-if="row.state === 3 || row.state === 4"
              link
              type="danger"
              @click.stop="$emit('delete', row.id)"
              class="action-btn"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>
        <div class="task-progress-wrapper">
          <el-progress
            :percentage="row.progress"
            :status="row.state === 3 ? 'success' : row.state === 4 ? 'exception' : undefined"
            :stroke-width="4"
            class="task-progress"
          />
          <div class="task-speed" v-if="row.state === 1">
            {{ formatSpeed(row.speed) }}
          </div>
        </div>
      </div>
    </div>

    <el-empty v-if="tasks.length === 0 && !loading" :description="t('tasks.noDownloadTasks')" />

    <!-- 分页 -->
    <pagination
      v-if="(props.total || 0) > 0"
      v-model:page="currentPage"
      v-model:limit="pageSize"
      :total="props.total || 0"
      :page-sizes="[20, 50, 100]"
      @pagination="handlePagination"
      class="pagination"
    />
  </el-card>
</template>

<script setup lang="ts">
  import { formatSize, formatDate, formatSpeed } from '@/utils'
  import { useI18n } from '@/composables'
  import type { OfflineDownloadTask } from '@/api/download'

  const { t } = useI18n()

  const props = defineProps<{
    tasks: OfflineDownloadTask[]
    loading: boolean
    currentPage?: number
    pageSize?: number
    total?: number
  }>()

  const emit = defineEmits<{
    pause: [taskId: string]
    resume: [taskId: string]
    cancel: [taskId: string]
    delete: [taskId: string]
    pagination: [data: { page: number; limit: number }]
  }>()

  const currentPage = computed({
    get: () => props.currentPage || 1,
    set: (val: number) => emit('pagination', { page: val, limit: props.pageSize || 20 })
  })

  const pageSize = computed({
    get: () => props.pageSize || 20,
    set: (val: number) => emit('pagination', { page: props.currentPage || 1, limit: val })
  })

  const handlePagination = ({ page, limit }: { page: number; limit: number }) => {
    emit('pagination', { page, limit })
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

  .table-wrapper {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: 8px;
  }

  .task-table {
    width: 100%;
  }

  .task-table :deep(.el-table__body-wrapper) {
    max-height: calc(100vh - 300px);
    overflow-y: auto;
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

  /* 当数据为空时，隐藏表格的空状态显示（因为我们已经用 el-empty 显示了） */
  .task-table :deep(.el-table__empty-block) {
    display: none;
  }

  /* 确保表格在数据为空时不显示滚动条 */
  .task-table:has(.el-table__empty-block) {
    overflow: hidden;
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

  .download-task-icon {
    color: var(--el-color-success);
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
    overflow: auto;
    padding: 8px;
  }

  .task-table {
    width: 100%;
  }

  .task-table :deep(.el-table__body-wrapper) {
    max-height: calc(100vh - 300px);
    overflow-y: auto;
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

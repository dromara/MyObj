<template>
  <el-card shadow="never" class="task-card">
    <div class="card-header">
      <span class="task-count">共 {{ tasks.length }} 个任务</span>
      <el-button 
        type="warning" 
        size="small" 
        icon="View" 
        @click="$emit('view-expired')"
        class="expired-btn"
      >
        查看过期任务<el-badge v-if="(expiredCount || 0) > 0" :value="expiredCount" class="expired-badge" />
      </el-button>
    </div>
    
    <!-- PC端：表格布局 -->
    <el-table :data="tasks" v-loading="loading" class="task-table desktop-table">
      <el-table-column label="文件名" min-width="300" class-name="mobile-name-column">
        <template #default="{ row }">
          <div class="file-name-cell">
            <el-icon :size="24" color="#409EFF"><Document /></el-icon>
            <span>{{ row.file_name }}</span>
          </div>
        </template>
      </el-table-column>
      
      <el-table-column label="状态" width="120" class-name="mobile-hide">
        <template #default="{ row }">
          <el-tag :type="getUploadStatusType(row.status)">{{ getUploadStatusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      
      <el-table-column label="进度" width="250" class-name="mobile-progress-column">
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
      
      <el-table-column label="创建时间" width="180" class-name="mobile-hide">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      
      <el-table-column label="操作" width="240" fixed="right" class-name="mobile-actions-column">
        <template #default="{ row }">
          <el-button 
            v-if="row.status === 'uploading'"
            link 
            icon="VideoPause" 
            type="warning"
            @click="$emit('pause', row.id)"
          >
            暂停
          </el-button>
          <el-button 
            v-if="row.status === 'paused'"
            link 
            icon="VideoPlay" 
            type="primary"
            @click="$emit('resume', row.id)"
          >
            继续
          </el-button>
          <el-button 
            v-if="row.status === 'uploading' || row.status === 'pending' || row.status === 'paused'"
            link 
            icon="Close" 
            type="danger"
            @click="$emit('cancel', row.id)"
          >
            取消
          </el-button>
          <el-button 
            link 
            icon="Delete" 
            type="danger"
            @click="$emit('delete', row.id)"
            :disabled="row.status === 'uploading'"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 移动端：卡片布局 -->
    <div class="mobile-task-list" v-loading="loading">
      <div 
        v-for="row in tasks" 
        :key="row.id" 
        class="mobile-task-item"
      >
        <div class="task-item-header">
          <div class="task-item-info">
            <el-icon :size="20" color="#409EFF" class="task-icon"><Document /></el-icon>
            <div class="task-name-wrapper">
              <div class="task-name">{{ row.file_name }}</div>
              <div class="task-meta">
                <el-tag :type="getUploadStatusType(row.status)" size="small" effect="plain">
                  {{ getUploadStatusText(row.status) }}
                </el-tag>
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
            :percentage="row.progress" 
            :status="row.status === 'completed' ? 'success' : row.status === 'failed' ? 'exception' : undefined"
            :stroke-width="4"
            class="task-progress"
          />
          <div class="task-speed" v-if="row.status === 'uploading'">
            {{ row.speed || '0 KB/s' }}
          </div>
        </div>
      </div>
    </div>
    
    <el-empty v-if="tasks.length === 0 && !loading" description="暂无上传任务" />
  </el-card>
</template>

<script setup lang="ts">
import { formatSize, formatDate, getUploadStatusType, getUploadStatusText } from '@/utils'

defineProps<{
  tasks: any[]
  loading: boolean
  cleanLoading: boolean
  expiredCount?: number
}>()

defineEmits<{
  pause: [taskId: string]
  resume: [taskId: string]
  cancel: [taskId: string]
  delete: [taskId: string]
  'view-expired': []
}>()
</script>

<style scoped>
.task-card {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.task-count {
  font-size: 14px;
  color: #606266;
}

.expired-badge {
  margin-left: 4px;
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

/* PC端表格样式 */
.desktop-table {
  display: table;
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
  background: #fff;
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
</style>


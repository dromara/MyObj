<template>
  <el-dialog
    v-model="visible"
    title="过期任务"
    width="80%"
    :close-on-click-modal="false"
    @close="handleClose"
    class="expired-tasks-dialog"
  >
    <div class="dialog-header">
      <span>共 {{ expiredTasks.length }} 个过期任务</span>
      <div class="header-actions">
        <el-button 
          type="primary" 
          size="small" 
          :disabled="selectedTasks.length === 0"
          @click="handleBatchRenew"
          :loading="batchRenewLoading"
        >
          批量延期（{{ selectedTasks.length }}）
        </el-button>
        <el-button 
          type="danger" 
          size="small" 
          :disabled="selectedTasks.length === 0"
          @click="handleBatchDelete"
          :loading="batchDeleteLoading"
        >
          批量删除（{{ selectedTasks.length }}）
        </el-button>
      </div>
    </div>

    <!-- PC端：表格布局 -->
    <el-table 
      :data="expiredTasks" 
      v-loading="loading"
      @selection-change="handleSelectionChange"
      class="expired-tasks-table desktop-table"
    >
      <el-table-column type="selection" width="55" />
      
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
      
      <el-table-column label="进度" width="200">
        <template #default="{ row }">
          <div class="progress-cell">
            <el-progress 
              :percentage="Math.round(row.progress)" 
              :status="row.status === 'completed' ? 'success' : row.status === 'failed' ? 'exception' : undefined"
            />
            <span class="progress-info">{{ formatSize(row.uploaded_chunks * row.chunk_size) }} / {{ formatSize(row.file_size) }}</span>
          </div>
        </template>
      </el-table-column>
      
      <el-table-column label="过期时间" width="180">
        <template #default="{ row }">
          <span style="color: var(--el-color-danger)">{{ formatDate(row.expire_time) }}</span>
        </template>
      </el-table-column>
      
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button 
            link 
            icon="Refresh" 
            type="primary"
            @click="handleRenew(row.id)"
            :loading="renewingTasks.has(row.id)"
          >
            延期
          </el-button>
          <el-button 
            link 
            icon="Delete" 
            type="danger"
            @click="handleDelete(row.id)"
            :loading="deletingTasks.has(row.id)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 移动端：卡片布局 -->
    <div class="mobile-task-list" v-loading="loading">
      <div 
        v-for="row in expiredTasks" 
        :key="row.id" 
        class="mobile-task-item"
      >
        <div class="task-item-header">
          <el-checkbox 
            :model-value="selectedTasks.some(t => t.id === row.id)"
            @change="handleItemSelect(row, $event)"
            class="task-checkbox"
          />
          <div class="task-item-info">
            <el-icon :size="20" color="#409EFF" class="task-icon"><Document /></el-icon>
            <div class="task-name-wrapper">
              <div class="task-name">{{ row.file_name }}</div>
              <div class="task-meta">
                <el-tag :type="getUploadStatusType(row.status)" size="small" effect="plain">
                  {{ getUploadStatusText(row.status) }}
                </el-tag>
                <span class="task-size">{{ formatSize(row.uploaded_chunks * row.chunk_size) }} / {{ formatSize(row.file_size) }}</span>
                <span class="task-expire-time" style="color: var(--el-color-danger)">
                  {{ formatDate(row.expire_time) }}
                </span>
              </div>
            </div>
          </div>
          <div class="task-actions">
            <el-button 
              link 
              type="primary"
              @click.stop="handleRenew(row.id)"
              :loading="renewingTasks.has(row.id)"
              class="action-btn"
            >
              <el-icon><Refresh /></el-icon>
            </el-button>
            <el-button 
              link 
              type="danger"
              @click.stop="handleDelete(row.id)"
              :loading="deletingTasks.has(row.id)"
              class="action-btn"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>
        <div class="task-progress-wrapper">
          <el-progress 
            :percentage="Math.round(row.progress)" 
            :status="row.status === 'completed' ? 'success' : row.status === 'failed' ? 'exception' : undefined"
            :stroke-width="4"
            class="task-progress"
          />
        </div>
      </div>
    </div>
    
    <el-empty v-if="expiredTasks.length === 0 && !loading" description="暂无过期任务" />
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { formatSize, formatDate, getUploadStatusType, getUploadStatusText } from '@/utils'
import { listExpiredUploads, renewExpiredTask, deleteUploadTask } from '@/api/file'
import type { UncompletedUploadTask } from '@/api/file'

const { proxy } = getCurrentInstance() as ComponentInternalInstance

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'refresh': []
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const loading = ref(false)
const expiredTasks = ref<UncompletedUploadTask[]>([])
const selectedTasks = ref<UncompletedUploadTask[]>([])
const renewingTasks = ref(new Set<string>())
const deletingTasks = ref(new Set<string>())
const batchRenewLoading = ref(false)
const batchDeleteLoading = ref(false)

// 加载过期任务列表
const loadExpiredTasks = async () => {
  loading.value = true
  try {
    const res = await listExpiredUploads()
    if (res.code === 200 && res.data) {
      expiredTasks.value = res.data
    }
  } catch (error: any) {
    proxy?.$modal.msgError(error.message || '加载过期任务失败')
  } finally {
    loading.value = false
  }
}

// 监听弹窗显示，加载数据
watch(() => props.modelValue, (val) => {
  if (val) {
    loadExpiredTasks()
    selectedTasks.value = []
  }
})

// 选择变化（PC端表格）
const handleSelectionChange = (selection: UncompletedUploadTask[]) => {
  selectedTasks.value = selection
}

// 移动端单个选择
const handleItemSelect = (task: UncompletedUploadTask, checked: boolean) => {
  if (checked) {
    if (!selectedTasks.value.find(t => t.id === task.id)) {
      selectedTasks.value.push(task)
    }
  } else {
    selectedTasks.value = selectedTasks.value.filter(t => t.id !== task.id)
  }
}

// 单个延期
const handleRenew = async (taskId: string) => {
  renewingTasks.value.add(taskId)
  try {
    const res = await renewExpiredTask(taskId, 7)
    if (res.code === 200) {
      proxy?.$modal.msgSuccess('延期成功')
      await loadExpiredTasks()
      emit('refresh')
    } else {
      proxy?.$modal.msgError(res.message || '延期失败')
    }
  } catch (error: any) {
    proxy?.$modal.msgError(error.message || '延期失败')
  } finally {
    renewingTasks.value.delete(taskId)
  }
}

// 批量延期
const handleBatchRenew = async () => {
  if (selectedTasks.value.length === 0) return
  
  try {
    await proxy?.$modal.confirm(`确认延期 ${selectedTasks.value.length} 个过期任务？`)
    batchRenewLoading.value = true
    
    const promises = selectedTasks.value.map(task => renewExpiredTask(task.id, 7))
    const results = await Promise.allSettled(promises)
    
    const successCount = results.filter(r => r.status === 'fulfilled' && r.value.code === 200).length
    const failCount = results.length - successCount
    
    if (successCount > 0) {
      proxy?.$modal.msgSuccess(`成功延期 ${successCount} 个任务${failCount > 0 ? `，${failCount} 个失败` : ''}`)
      await loadExpiredTasks()
      emit('refresh')
    } else {
      proxy?.$modal.msgError('延期失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      proxy?.$modal.msgError('延期失败')
    }
  } finally {
    batchRenewLoading.value = false
  }
}

// 单个删除
const handleDelete = async (taskId: string) => {
  try {
    await proxy?.$modal.confirm('确认删除该过期任务？')
    deletingTasks.value.add(taskId)
    
    const res = await deleteUploadTask({ task_id: taskId })
    if (res.code === 200) {
      proxy?.$modal.msgSuccess('删除成功')
      await loadExpiredTasks()
      emit('refresh')
    } else {
      proxy?.$modal.msgError(res.message || '删除失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      proxy?.$modal.msgError('删除失败')
    }
  } finally {
    deletingTasks.value.delete(taskId)
  }
}

// 批量删除
const handleBatchDelete = async () => {
  if (selectedTasks.value.length === 0) return
  
  try {
    await proxy?.$modal.confirm(`确认删除 ${selectedTasks.value.length} 个过期任务？删除后无法恢复。`)
    batchDeleteLoading.value = true
    
    const promises = selectedTasks.value.map(task => deleteUploadTask({ task_id: task.id }))
    const results = await Promise.allSettled(promises)
    
    const successCount = results.filter(r => r.status === 'fulfilled' && r.value.code === 200).length
    const failCount = results.length - successCount
    
    if (successCount > 0) {
      proxy?.$modal.msgSuccess(`成功删除 ${successCount} 个任务${failCount > 0 ? `，${failCount} 个失败` : ''}`)
      await loadExpiredTasks()
      emit('refresh')
    } else {
      proxy?.$modal.msgError('删除失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      proxy?.$modal.msgError('删除失败')
    }
  } finally {
    batchDeleteLoading.value = false
  }
}

// 关闭弹窗
const handleClose = () => {
  visible.value = false
  selectedTasks.value = []
}
</script>


<style scoped>
.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.header-actions {
  display: flex;
  gap: 8px;
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

.expired-tasks-table {
  max-height: 500px;
  overflow-y: auto;
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
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 8px;
}

.task-checkbox {
  flex-shrink: 0;
  margin-top: 2px;
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

.task-expire-time {
  font-size: 12px;
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
  margin-left: 32px; /* 对齐复选框 */
}

.task-progress {
  flex: 1;
}

/* 移动端响应式 */
@media (max-width: 1024px) {
  .desktop-table {
    display: none !important;
  }
  
  .mobile-task-list {
    display: block;
    max-height: 60vh;
    overflow-y: auto;
  }
  
  .expired-tasks-dialog :deep(.el-dialog) {
    width: 95% !important;
    margin: 5vh auto;
    max-height: 90vh;
  }
  
  .expired-tasks-dialog :deep(.el-dialog__body) {
    padding: 16px;
    max-height: calc(90vh - 120px);
    overflow-y: auto;
  }
  
  .dialog-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
    padding: 12px 0;
    margin-bottom: 12px;
  }
  
  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }
  
  .header-actions .el-button {
    font-size: 12px;
    padding: 6px 12px;
  }
}

@media (max-width: 480px) {
  .expired-tasks-dialog :deep(.el-dialog) {
    width: 100% !important;
    margin: 0 !important;
    height: 100vh !important;
    max-height: 100vh !important;
    border-radius: 0 !important;
  }
  
  .expired-tasks-dialog :deep(.el-dialog__header) {
    padding: 12px 16px;
    flex-shrink: 0;
  }
  
  .expired-tasks-dialog :deep(.el-dialog__body) {
    padding: 12px;
    max-height: calc(100vh - 140px);
    overflow-y: auto;
  }
  
  .expired-tasks-dialog :deep(.el-dialog__footer) {
    padding: 12px 16px;
    flex-shrink: 0;
  }
  
  .mobile-task-item {
    padding: 10px 12px;
  }
  
  .task-name {
    font-size: 13px;
  }
  
  .task-meta {
    font-size: 11px;
  }
  
  .task-progress-wrapper {
    margin-left: 28px;
  }
}
</style>


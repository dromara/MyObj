<template>
  <el-dialog
    v-model="visible"
    :title="t('extract.title')"
    width="550px"
    :close-on-click-modal="false"
    :close-on-press-escape="true"
    class="extract-dialog"
    @close="handleClose"
    @keydown.enter.prevent
  >
    <!-- File info -->
    <div class="file-info-card">
      <el-icon :size="36" class="archive-icon"><FolderOpened /></el-icon>
      <div class="file-info-content">
        <div class="file-name">{{ fileInfo.file_name || t('common.noData') }}</div>
        <div class="file-size" v-if="fileInfo.file_size">
          {{ formatFileSize(fileInfo.file_size) }}
        </div>
      </div>
    </div>

    <!-- Select target path -->
    <el-form :model="extractForm" label-width="100px" class="extract-form">
      <el-form-item :label="t('extract.targetPath')">
        <div class="path-select-area">
          <div class="current-path" v-if="selectedPathLabel">
            <el-icon><Folder /></el-icon>
            <span>{{ selectedPathLabel }}</span>
          </div>
          <el-radio-group v-model="extractForm.target_path_id" class="path-options">
            <el-radio-button value="home">{{ t('extract.rootDir') }}</el-radio-button>
            <el-radio-button :value="currentPathID" v-if="currentPathLabel && currentPathID">
              {{ t('extract.currentDir') }} ({{ currentPathLabel }})
            </el-radio-button>
          </el-radio-group>
        </div>
      </el-form-item>

      <el-form-item v-if="fileInfo.is_enc" :label="t('extract.password')">
        <el-input
          v-model="extractForm.file_password"
          type="password"
          :placeholder="t('extract.passwordPlaceholder')"
          :show-password="true"
        />
        <div class="form-tip">{{ t('extract.passwordTip') }}</div>
      </el-form-item>
    </el-form>

    <!-- Conflict detection -->
    <div v-if="conflictInfo" class="conflict-section">
      <el-alert
        :title="t('extract.conflictTitle')"
        type="warning"
        :closable="false"
        :show-icon="true"
      />
      <div class="conflict-desc">{{ t('extract.conflictDesc', { count: conflictInfo.conflict_files.length }) }}</div>
      <div class="conflict-files">
        <el-tag
          v-for="file in conflictInfo.conflict_files.slice(0, 5)"
          :key="file"
          type="warning"
          size="small"
          class="conflict-tag"
        >{{ file }}</el-tag>
        <el-tag v-if="conflictInfo.conflict_files.length > 5" type="info" size="small" class="conflict-tag">
          +{{ conflictInfo.conflict_files.length - 5 }}
        </el-tag>
      </div>
      <div class="conflict-actions">
        <el-tooltip :content="t('extract.overwriteTip')" placement="top">
          <el-button type="warning" @click="resolveConflict('overwrite')">
            {{ t('extract.overwrite') }}
          </el-button>
        </el-tooltip>
        <el-tooltip :content="t('extract.keepBothTip')" placement="top">
          <el-button type="primary" @click="resolveConflict('keep_both')">
            {{ t('extract.keepBoth') }}
          </el-button>
        </el-tooltip>
        <el-tooltip :content="t('extract.skipConflictTip')" placement="top">
          <el-button @click="resolveConflict('cancel')">
            {{ t('extract.skipConflict') }}
          </el-button>
        </el-tooltip>
      </div>
    </div>

    <!-- Progress -->
    <div v-if="taskStatus && taskStatus !== 'idle'" class="extract-progress">
      <el-alert
        v-if="taskStatus === 'completed'"
        :title="t('extract.completed', { completed: completedFiles, total: totalFiles })"
        type="success"
        :closable="false"
        :show-icon="true"
      />
      <el-alert
        v-else-if="taskStatus === 'failed'"
        :title="errorMsg || t('extract.failed')"
        type="error"
        :show-icon="true"
      />
      <el-alert
        v-else-if="taskStatus.includes('skipped')"
        :title="taskStatus"
        type="success"
        :show-icon="true"
      />
      <el-alert
        v-else-if="taskStatus.includes('partial')"
        :title="taskStatus"
        type="warning"
        :show-icon="true"
      />
      <div v-else class="progress-area">
        <el-progress :percentage="progress" :status="progress === 100 ? 'success' : undefined" />
        <div class="progress-detail">{{ statusLabel }}</div>
        <div v-if="currentFile" class="current-file">{{ t('extract.extracting') }}: {{ currentFile }}</div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose" :disabled="isRunning">{{
          taskStatus === 'completed' || taskStatus?.includes('partial') || taskStatus?.includes('skipped') ? t('common.close') : t('common.cancel')
        }}</el-button>
        <el-button
          type="primary"
          @click="handleExtract"
          :loading="isRunning || checkingConflict"
          :disabled="!!conflictInfo"
          :tabindex="-1"
          :autofocus="false"
        >
          {{ t('extract.startExtract') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, nextTick } from 'vue'
import { useI18n } from '@/composables/core/useI18n'
import { createExtract, getExtractProgress, checkExtractConflict } from '@/api/extract'
import type { ExtractCheckResponse } from '@/api/extract'
import type { FileItem } from '@/types'

const { t } = useI18n()
const { proxy } = getCurrentInstance() as any

const emit = defineEmits<{
  (e: 'completed'): void
}>()

const visible = ref(false)
const canExtract = ref(false)
const fileInfo = reactive<{ file_id: string; file_name: string; file_size: number; is_enc: boolean }>({
  file_id: '',
  file_name: '',
  file_size: 0,
  is_enc: false
})

const extractForm = reactive({
  target_path_id: 'home',
  file_password: ''
})

const taskID = ref('')
const taskStatus = ref('')
const progress = ref(0)
const currentFile = ref('')
const completedFiles = ref(0)
const totalFiles = ref(0)
const errorMsg = ref('')
const checkingConflict = ref(false)
const conflictInfo = ref<ExtractCheckResponse | null>(null)
let pollTimer: ReturnType<typeof setInterval> | null = null

const currentPathID = ref('')
const currentPathLabel = computed(() => {
  const pro = proxy as any
  return pro?.$route?.query?.pathLabel || ''
})

const isRunning = computed(() => {
  return !!(
    taskStatus.value &&
    taskStatus.value !== 'idle' &&
    taskStatus.value !== 'completed' &&
    !taskStatus.value.includes('partial') &&
    !taskStatus.value.includes('skipped') &&
    taskStatus.value !== 'failed'
  )
})

const selectedPathLabel = computed(() => {
  return extractForm.target_path_id === 'home'
    ? t('extract.rootDir')
    : t('extract.currentDir') + (currentPathLabel.value ? ` (${currentPathLabel.value})` : '')
})

const statusLabel = computed(() => {
  switch (taskStatus.value) {
    case 'preparing': return t('extract.statusPreparing')
    case 'downloading': return t('extract.statusDownloading')
    case 'extracting': return t('extract.statusExtracting')
    case 'uploading': return t('extract.statusUploading')
    default: return taskStatus.value
  }
})

const formatFileSize = (size: number): string => {
  if (size < 1024) return size + ' B'
  if (size < 1024 * 1024) return (size / 1024).toFixed(1) + ' KB'
  if (size < 1024 * 1024 * 1024) return (size / (1024 * 1024)).toFixed(1) + ' MB'
  return (size / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
}

const open = (file: FileItem, pathID?: string) => {
  canExtract.value = false
  fileInfo.file_id = file.file_id
  fileInfo.file_name = file.file_name
  fileInfo.file_size = file.file_size
  fileInfo.is_enc = (file as any).is_enc || false
  currentPathID.value = pathID || ''
  extractForm.target_path_id = 'home'
  extractForm.file_password = ''
  taskStatus.value = ''
  progress.value = 0
  currentFile.value = ''
  completedFiles.value = 0
  totalFiles.value = 0
  errorMsg.value = ''
  conflictInfo.value = null
  checkingConflict.value = false
  visible.value = true
  nextTick(() => {
    setTimeout(() => {
      canExtract.value = true
    }, 300)
  })
}

const handleExtract = async () => {
  if (!canExtract.value) return

  // 先检测冲突
  checkingConflict.value = true
  try {
    const checkRes = await checkExtractConflict({
      file_id: fileInfo.file_id,
      target_path_id: extractForm.target_path_id,
      file_password: extractForm.file_password || undefined
    })
    checkingConflict.value = false

    if (checkRes.code === 200 && checkRes.data) {
      if (checkRes.data.has_conflict) {
        // 有冲突，显示冲突信息让用户选择
        conflictInfo.value = checkRes.data
        return
      }
    }
    // 无冲突，直接创建任务
    conflictInfo.value = null
    await doCreateExtract()
  } catch (error: any) {
    checkingConflict.value = false
    proxy?.$modal.msgError(error.message || t('extract.createFailed'))
  }
}

const resolveConflict = async (resolution: string) => {
  conflictInfo.value = null
  await doCreateExtract(resolution)
}

const doCreateExtract = async (conflictResolution?: string) => {
  try {
    proxy?.$modal.loading(t('extract.creating'))
    const res = await createExtract({
      file_id: fileInfo.file_id,
      target_path_id: extractForm.target_path_id,
      file_password: extractForm.file_password || undefined,
      conflict_resolution: conflictResolution || undefined
    })
    proxy?.$modal.closeLoading()

    if (res.code === 200 && res.data) {
      taskID.value = res.data.task_id
      taskStatus.value = res.data.status
      totalFiles.value = res.data.total_files
      startPolling()
    } else {
      proxy?.$modal.msgError(res.message || t('extract.createFailed'))
    }
  } catch (error: any) {
    proxy?.$modal.closeLoading()
    proxy?.$modal.msgError(error.message || t('extract.createFailed'))
  }
}

const startPolling = () => {
  if (pollTimer) clearInterval(pollTimer)
  pollTimer = setInterval(async () => {
    if (!taskID.value) return
    try {
      const res = await getExtractProgress(taskID.value)
      if (res.code === 200 && res.data) {
        const data = res.data
        taskStatus.value = data.status
        progress.value = data.progress
        currentFile.value = data.current_file
        completedFiles.value = data.completed
        totalFiles.value = data.total_files
        errorMsg.value = data.error_msg

        if (data.status === 'completed' || data.status.includes('partial') || data.status === 'failed') {
          if (pollTimer) {
            clearInterval(pollTimer)
            pollTimer = null
          }
          // 解压完成（成功/部分/失败）时通知父组件刷新文件列表
          if (data.status === 'completed' || data.status.includes('partial')) {
            emit('completed')
          }
        }
      }
    } catch {
      // ignore polling errors
    }
  }, 3000)
}

const handleClose = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  visible.value = false
}

watch(visible, (val) => {
  if (!val) {
    handleClose()
  }
})

defineExpose({ open })
</script>

<style scoped>
.extract-dialog :deep(.el-dialog__body) {
  padding-top: 16px;
}

.file-info-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  margin-bottom: 20px;
  background: var(--el-fill-color-light);
  border-radius: 8px;
}

.archive-icon {
  color: var(--el-color-primary);
}

.file-info-content {
  flex: 1;
  min-width: 0;
}

.file-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}

.extract-form {
  margin-bottom: 20px;
}

.path-select-area {
  width: 100%;
}

.current-path {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 10px;
  font-size: 13px;
  color: var(--el-color-primary);
}

.path-options {
  width: 100%;
}

.form-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}

.extract-progress {
  margin-top: 8px;
}

.progress-area {
  padding: 16px;
  background: var(--el-fill-color-light);
  border-radius: 8px;
}

.progress-detail {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-top: 8px;
}

.current-file {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.conflict-section {
  margin-top: 16px;
  padding: 12px;
  background: var(--el-color-warning-light-9);
  border-radius: 8px;
}

.conflict-desc {
  font-size: 13px;
  color: var(--el-text-color-regular);
  margin-top: 8px;
}

.conflict-files {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.conflict-tag {
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.conflict-actions {
  margin-top: 12px;
  display: flex;
  gap: 8px;
}
</style>

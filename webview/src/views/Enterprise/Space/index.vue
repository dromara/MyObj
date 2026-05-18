<template>
  <div class="enterprise-space">
    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" icon="Plus" size="small" @click="showMkdirDialog = true">
          {{ t('enterprise.space.mkdir') }}
        </el-button>
        <el-upload
          :show-file-list="false"
          :before-upload="handleUpload"
          :disabled="uploading"
          multiple
        >
          <el-button icon="Upload" size="small" :loading="uploading">{{ t('enterprise.space.upload') }}</el-button>
        </el-upload>
        <el-button
          icon="Delete"
          size="small"
          :disabled="selectedFiles.length === 0"
          @click="handleBatchDelete"
        >
          {{ t('common.delete') }}
        </el-button>
      </div>
      <div class="toolbar-right">
        <el-button icon="Refresh" size="small" @click="loadFiles">{{ t('common.refresh') }}</el-button>
      </div>
    </div>

    <!-- 上传进度 -->
    <div v-if="uploading" class="upload-progress-bar">
      <el-progress :percentage="uploadProgress" :stroke-width="10" :text-inside="true" />
    </div>

    <!-- 面包屑导航 -->
    <div class="breadcrumb-bar">
      <el-breadcrumb separator="/">
        <el-breadcrumb-item @click="navigateTo(0)">
          <el-icon><HomeFilled /></el-icon>
        </el-breadcrumb-item>
        <el-breadcrumb-item
          v-for="item in breadcrumbs"
          :key="item.id"
          @click="navigateTo(item.id)"
        >
          {{ item.name }}
        </el-breadcrumb-item>
      </el-breadcrumb>
    </div>

    <!-- 空间使用情况 -->
    <div v-if="spaceUsage" class="usage-bar">
      <div class="usage-info">
        <span>{{ t('enterprise.space.used') }}: {{ formatSize(spaceUsage.used_space) }}</span>
        <span>/</span>
        <span>{{ t('enterprise.space.total') }}: {{ spaceUsage.total_space > 0 ? formatSize(spaceUsage.total_space) : '∞' }}</span>
        <el-divider direction="vertical" />
        <span>{{ t('enterprise.space.fileCount') }}: {{ spaceUsage.file_count }}</span>
      </div>
      <el-progress
        v-if="spaceUsage.total_space > 0"
        :percentage="Math.min(100, Math.round((spaceUsage.used_space / spaceUsage.total_space) * 100))"
        :stroke-width="8"
        style="flex: 1; max-width: 300px"
      />
    </div>

    <!-- 文件列表 -->
    <el-table
      :data="fileList"
      v-loading="loading"
      class="file-table"
      :empty-text="t('enterprise.space.emptyFolder')"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="50" />
      <el-table-column :label="t('enterprise.info.name')" min-width="250">
        <template #default="{ row }">
          <div class="file-name-cell" @click="row._isDir && navigateTo(row.id, row.name)">
            <el-icon v-if="row._isDir" class="file-icon folder-icon"><Folder /></el-icon>
            <el-icon v-else class="file-icon"><Document /></el-icon>
            <span :class="{ 'is-folder': row._isDir }">{{ row._isDir ? row.name : row.file_name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="t('enterprise.info.storage')" width="120">
        <template #default="{ row }">
          {{ row._isDir ? '-' : formatSize(row.size) }}
        </template>
      </el-table-column>
      <el-table-column prop="created_at" :label="t('enterprise.info.createdAt')" width="180" />
      <el-table-column :label="t('common.operation')" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleRename(row)">
            {{ t('common.rename') || '重命名' }}
          </el-button>
          <el-button v-if="!row._isDir" link type="primary" size="small" @click="handleDownload(row)">
            {{ t('enterprise.space.download') }}
          </el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">
            {{ t('common.delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="pagination.page"
      v-model:page-size="pagination.pageSize"
      :total="pagination.total"
      :page-sizes="[20, 50, 100]"
      layout="total, sizes, prev, pager, next"
      @size-change="loadFiles"
      @current-change="loadFiles"
      class="pagination"
    />

    <!-- 新建文件夹对话框 -->
    <el-dialog v-model="showMkdirDialog" :title="t('enterprise.space.mkdir')" width="400px">
      <el-form :model="mkdirForm" :rules="mkdirRules" ref="mkdirFormRef" label-width="80px">
        <el-form-item :label="t('enterprise.info.name')" prop="name">
          <el-input v-model="mkdirForm.name" :placeholder="t('enterprise.info.name')" @keyup.enter="handleMkdir" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showMkdirDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="mkdirLoading" @click="handleMkdir">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 重命名对话框 -->
    <el-dialog v-model="showRenameDialog" :title="t('common.rename') || '重命名'" width="400px">
      <el-form :model="renameForm" :rules="renameRules" ref="renameFormRef" label-width="80px">
        <el-form-item :label="t('enterprise.info.name')" prop="name">
          <el-input v-model="renameForm.name" :placeholder="t('enterprise.info.name')" @keyup.enter="handleRenameSubmit" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRenameDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="renameLoading" @click="handleRenameSubmit">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import type { ComponentInternalInstance } from 'vue'
  import { enterpriseApi } from '@myobj/api'
  import { upload } from '@myobj/http'
  import { API_ENDPOINTS } from '@myobj/shared'
  import type { SharedDirEntry, SharedFileEntry, SpaceUsage } from '@myobj/shared'
  import { formatSize } from '@/utils'
  import { useI18n } from '@/composables'

  const props = defineProps<{
    enterpriseId: string
  }>()

  const {
    getSharedFileList, createSharedDir, deleteSharedFile,
    downloadSharedFile, sharedUploadPrecheck, getSpaceUsage,
    deleteSharedDir, renameSharedFile, renameSharedDir
  } = enterpriseApi

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  const loading = ref(false)
  const fileList = ref<any[]>([])
  const selectedFiles = ref<any[]>([])
  const currentPathId = ref(0)
  const breadcrumbs = ref<{ id: number; name: string }[]>([])
  const spaceUsage = ref<SpaceUsage | null>(null)

  const pagination = reactive({ page: 1, pageSize: 50, total: 0 })

  const showMkdirDialog = ref(false)
  const mkdirLoading = ref(false)
  const mkdirFormRef = ref()
  const mkdirForm = reactive({ name: '' })
  const mkdirRules = {
    name: [{ required: true, message: t('enterprise.info.name'), trigger: 'blur' }]
  }
  const uploading = ref(false)
  const uploadProgress = ref(0)

  const showRenameDialog = ref(false)
  const renameLoading = ref(false)
  const renameFormRef = ref()
  const renameForm = reactive({ name: '' })
  const renameRules = {
    name: [{ required: true, message: t('enterprise.info.name'), trigger: 'blur' }]
  }
  const renamingItem = ref<any>(null)

  const loadFiles = async () => {
    loading.value = true
    try {
      const res = await getSharedFileList({
        enterprise_id: props.enterpriseId,
        path_id: currentPathId.value || undefined,
        page: pagination.page,
        pageSize: pagination.pageSize
      })
      if (res.code === 200 && res.data) {
        const dirs = (res.data.dirs || []).map(d => ({ ...d, _isDir: true }))
        const files = (res.data.files || []).map(f => ({ ...f, _isDir: false }))
        fileList.value = [...dirs, ...files]
        pagination.total = res.data.total || 0
      }
    } catch (error: any) {
      proxy?.$log?.error(error)
    } finally {
      loading.value = false
    }
  }

  const loadSpaceUsage = async () => {
    try {
      const res = await getSpaceUsage(props.enterpriseId)
      if (res.code === 200 && res.data) {
        spaceUsage.value = res.data
      }
    } catch {}
  }

  const navigateTo = (pathId: number, name?: string) => {
    currentPathId.value = pathId
    if (pathId === 0) {
      breadcrumbs.value = []
    } else {
      const idx = breadcrumbs.value.findIndex(b => b.id === pathId)
      if (idx >= 0) {
        breadcrumbs.value = breadcrumbs.value.slice(0, idx + 1)
      } else if (name) {
        breadcrumbs.value.push({ id: pathId, name })
      }
    }
    pagination.page = 1
    loadFiles()
  }

  const handleSelectionChange = (selection: any[]) => {
    selectedFiles.value = selection
  }

  const handleMkdir = async () => {
    if (!mkdirFormRef.value) return
    await mkdirFormRef.value.validate(async (valid: boolean) => {
      if (!valid) return
      mkdirLoading.value = true
      try {
        const res = await createSharedDir({
          enterprise_id: props.enterpriseId,
          name: mkdirForm.name,
          parent_id: currentPathId.value || undefined
        })
        if (res.code === 200) {
          proxy?.$modal.msgSuccess(t('enterprise.space.mkdirSuccess'))
          showMkdirDialog.value = false
          mkdirForm.name = ''
          loadFiles()
        } else {
          proxy?.$modal.msgError(res.message || t('common.operationFailed'))
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || t('common.operationFailed'))
      } finally {
        mkdirLoading.value = false
      }
    })
  }

  const handleUpload = async (file: File) => {
    try {
      uploading.value = true
      uploadProgress.value = 0

      // Precheck
      const precheckRes = await sharedUploadPrecheck({
        enterprise_id: props.enterpriseId,
        file_name: file.name,
        file_size: file.size,
        path_id: currentPathId.value || undefined
      })
      if (precheckRes.code !== 200) {
        proxy?.$modal.msgError(precheckRes.message || t('common.operationFailed'))
        return false
      }

      // Upload using project HTTP utility
      const formData = new FormData()
      formData.append('enterprise_id', props.enterpriseId)
      formData.append('precheck_id', precheckRes.data?.precheck_id || '')
      if (currentPathId.value) formData.append('path_id', String(currentPathId.value))

      const result = await upload(
        API_ENDPOINTS.ENTERPRISE.SPACE.UPLOAD,
        file,
        formData,
        (percent) => { uploadProgress.value = Math.round(percent) }
      )

      if (result.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.space.uploadSuccess'))
        loadFiles()
        loadSpaceUsage()
      } else {
        proxy?.$modal.msgError(result.message || t('common.operationFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('common.operationFailed'))
    } finally {
      uploading.value = false
      uploadProgress.value = 0
    }
    return false
  }

  const handleDownload = async (file: SharedFileEntry) => {
    try {
      await downloadSharedFile(file.id)
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('common.operationFailed'))
    }
  }

  const handleDelete = async (item: any) => {
    try {
      await proxy?.$modal.confirm(t('enterprise.space.deleteConfirm'))
      const res = item._isDir
        ? await deleteSharedDir(item.id)
        : await deleteSharedFile({ id: item.id })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.space.deleteSuccess'))
        loadFiles()
        loadSpaceUsage()
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch {}
  }

  const handleBatchDelete = async () => {
    if (selectedFiles.value.length === 0) return
    try {
      await proxy?.$modal.confirm(t('enterprise.space.deleteConfirm'))
      const results = await Promise.allSettled(
        selectedFiles.value.map((file: any) =>
          file._isDir
            ? deleteSharedDir(file.id).then(res => ({ res, file }))
            : deleteSharedFile({ id: file.id }).then(res => ({ res, file }))
        )
      )
      const failed = results.filter(r => r.status === 'rejected' || (r.status === 'fulfilled' && r.value.res.code !== 200)).length
      if (failed > 0) {
        proxy?.$modal.msgWarning(`${t('common.deleteSuccess')} (${selectedFiles.value.length - failed}/${selectedFiles.value.length})`)
      } else {
        proxy?.$modal.msgSuccess(t('enterprise.space.deleteSuccess'))
      }
      loadFiles()
      loadSpaceUsage()
    } catch {}
  }

  const handleRename = (item: any) => {
    renamingItem.value = item
    renameForm.name = item._isDir ? item.name : item.file_name
    showRenameDialog.value = true
  }

  const handleRenameSubmit = async () => {
    if (!renameFormRef.value) return
    await renameFormRef.value.validate(async (valid: boolean) => {
      if (!valid) return
      renameLoading.value = true
      try {
        const item = renamingItem.value
        const res = item._isDir
          ? await renameSharedDir(item.id, renameForm.name)
          : await renameSharedFile(item.id, renameForm.name)
        if (res.code === 200) {
          proxy?.$modal.msgSuccess(t('common.renameSuccess') || '重命名成功')
          showRenameDialog.value = false
          loadFiles()
        } else {
          proxy?.$modal.msgError(res.message || t('common.operationFailed'))
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || t('common.operationFailed'))
      } finally {
        renameLoading.value = false
      }
    })
  }

  watch(() => props.enterpriseId, (id) => {
    if (id) {
      currentPathId.value = 0
      breadcrumbs.value = []
      loadFiles()
      loadSpaceUsage()
    }
  }, { immediate: true })
</script>

<style scoped>
  .enterprise-space {
    display: flex;
    flex-direction: column;
    gap: 12px;
    height: 100%;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
  }

  .toolbar-left {
    display: flex;
    gap: 8px;
    align-items: center;
    flex-wrap: wrap;
  }

  .toolbar-right {
    display: flex;
    gap: 8px;
  }

  .upload-progress-bar {
    padding: 0 4px;
  }

  .breadcrumb-bar {
    padding: 8px 0;
  }

  .breadcrumb-bar .el-breadcrumb-item {
    cursor: pointer;
  }

  .usage-bar {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 8px 12px;
    background: var(--el-fill-color-lighter);
    border-radius: 8px;
    font-size: 13px;
  }

  .usage-info {
    display: flex;
    align-items: center;
    gap: 4px;
    white-space: nowrap;
    color: var(--el-text-color-regular);
  }

  .file-table {
    flex: 1;
    overflow: auto;
  }

  .file-name-cell {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: default;
  }

  .file-name-cell .is-folder {
    cursor: pointer;
    color: var(--primary-color);
  }

  .file-icon {
    font-size: 18px;
    color: var(--el-text-color-secondary);
  }

  .folder-icon {
    color: var(--el-color-warning);
  }

  .pagination {
    margin-top: 8px;
    justify-content: flex-end;
  }

  @media (max-width: 768px) {
    .toolbar {
      flex-direction: column;
      align-items: stretch;
    }

    .toolbar-left,
    .toolbar-right {
      width: 100%;
    }

    .usage-bar {
      flex-direction: column;
      align-items: flex-start;
    }

    .pagination :deep(.el-pagination__sizes),
    .pagination :deep(.el-pagination__jump) {
      display: none;
    }
  }

  html.dark .usage-bar {
    background: var(--el-fill-color-dark);
  }
</style>

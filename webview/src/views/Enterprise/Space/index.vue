<template>
  <div class="enterprise-space" @dragenter.prevent="onDragEnter" @dragover.prevent @dragleave="onDragLeave" @drop.prevent="onDrop">
    <!-- 拖拽上传覆盖层 -->
    <div v-if="isDraggingOver" class="drag-overlay">
      <el-icon :size="48"><Upload /></el-icon>
      <p>释放文件以上传</p>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar-container">
      <div class="toolbar-left">
        <el-button class="action-btn" icon="Plus" @click="showMkdirDialog = true">
          {{ t('enterprise.space.mkdir') }}
        </el-button>
        <el-upload
          :show-file-list="false"
          :before-upload="handleUpload"
          :disabled="uploading"
          multiple
        >
          <el-button class="action-btn-secondary" icon="Upload" :loading="uploading">{{ t('enterprise.space.upload') }}</el-button>
        </el-upload>
        <el-button
          type="danger"
          plain
          icon="Delete"
          :disabled="selectedFiles.length === 0"
          @click="handleBatchDelete"
        >
          {{ t('common.delete') }}
        </el-button>
        <el-button
          icon="Download"
          :disabled="selectedFiles.length === 0 || selectedFiles.some(f => f._isDir)"
          @click="handleBatchDownload"
        >
          {{ t('common.download') }}
        </el-button>
        <el-button
          icon="Rank"
          :disabled="selectedFiles.length !== 1 || selectedFiles[0]._isDir"
          @click="handleMoveClick(selectedFiles[0])"
        >
          {{ t('common.move') || '移动' }}
        </el-button>
      </div>
      <div class="toolbar-right">
        <el-input
          v-model="searchKeyword"
          :placeholder="t('common.search') || '搜索文件...'"
          clearable
          style="width: 200px"
          @keyup.enter="handleSearch"
          @clear="handleSearchClear"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button-group class="view-toggle">
          <el-button :type="viewMode === 'list' ? 'primary' : ''" icon="List" @click="viewMode = 'list'" />
          <el-button :type="viewMode === 'grid' ? 'primary' : ''" icon="Grid" @click="viewMode = 'grid'" />
        </el-button-group>
        <el-button class="action-btn-secondary" icon="Refresh" @click="loadFiles">{{ t('common.refresh') }}</el-button>
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
      <span v-if="isSearching" class="search-badge">
        <el-icon><Search /></el-icon>
        搜索: "{{ searchKeyword }}"
        <el-icon class="clear-search" @click="handleSearchClear"><Close /></el-icon>
      </span>
    </div>

    <!-- 空间使用情况 -->
    <div v-if="spaceUsage && !isSearching" class="usage-bar">
      <el-icon class="usage-icon"><Coin /></el-icon>
      <div class="usage-info">
        <span>{{ t('enterprise.space.used') }}: {{ formatSize(spaceUsage.used_space) }}</span>
        <span class="usage-sep">/</span>
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

    <!-- 列表视图 -->
    <el-table
      v-if="viewMode === 'list'"
      :data="fileList"
      v-loading="loading"
      class="file-table styled-table"
      :empty-text="isSearching ? '未找到匹配文件' : t('enterprise.space.emptyFolder')"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="50" />
      <el-table-column :label="t('trash.name')" min-width="250" sortable="custom" prop="name" :sort-orders="['ascending', 'descending']" @sort-change="handleSortChange">
        <template #default="{ row }">
          <div class="file-name-cell" @click="row._isDir && navigateTo(row.id, row.name)">
            <el-icon v-if="row._isDir" class="file-icon folder-icon"><Folder /></el-icon>
            <el-icon v-else class="file-icon"><Document /></el-icon>
            <span
              :class="{ 'is-folder': row._isDir, 'is-previewable': !row._isDir && isPreviewableFile(row) }"
              @click="!row._isDir && handlePreview(row)"
            >{{ row._isDir ? row.name : row.file_name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="t('enterprise.info.storage')" width="100" sortable="custom" prop="size" @sort-change="handleSortChange">
        <template #default="{ row }">
          {{ row._isDir ? '-' : formatSize(row.size) }}
        </template>
      </el-table-column>
      <el-table-column label="上传人" width="100">
        <template #default="{ row }">
          {{ row.uploader_name || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="修改人" width="100">
        <template #default="{ row }">
          {{ row.updated_by_name || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="created_at" :label="t('enterprise.info.createdAt')" width="170" sortable="custom" @sort-change="handleSortChange" />
      <el-table-column prop="updated_at" label="修改时间" width="170" sortable="custom" @sort-change="handleSortChange" />
      <el-table-column :label="t('common.operation')" width="180" fixed="right">
        <template #default="{ row }">
          <div class="operation-cell">
            <el-button v-if="!row._isDir && isPreviewableFile(row)" link type="primary" size="small" @click="handlePreview(row)">
              预览
            </el-button>
            <el-button v-if="!row._isDir" link type="primary" size="small" @click="handleDownload(row)">
              {{ t('enterprise.space.download') }}
            </el-button>
            <el-dropdown trigger="click" @command="(cmd: string) => handleContextMenu(cmd, row)">
              <el-button link type="primary" size="small">
                更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item v-if="!row._isDir && isPreviewableFile(row)" command="preview" icon="View">预览</el-dropdown-item>
                  <el-dropdown-item command="rename" icon="Edit">重命名</el-dropdown-item>
                  <el-dropdown-item v-if="!row._isDir" command="move" icon="Rank">移动</el-dropdown-item>
                  <el-dropdown-item v-if="!row._isDir" command="share" icon="Share">分享</el-dropdown-item>
                  <el-dropdown-item v-if="!row._isDir && isExtractable(row)" command="extract" icon="FolderOpened">解压</el-dropdown-item>
                  <el-dropdown-item command="delete" icon="Delete" divided>删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <!-- 网格视图 -->
    <div v-else class="file-grid" v-loading="loading">
      <div v-if="fileList.length === 0" class="grid-empty">
        <el-empty :description="isSearching ? '未找到匹配文件' : t('enterprise.space.emptyFolder')" />
      </div>
      <div
        v-for="item in fileList"
        :key="item._isDir ? 'dir-' + item.id : 'file-' + item.id"
        class="grid-item"
        :class="{ 'is-selected': selectedFiles.some(s => s.id === item.id) }"
        @click="handleGridClick(item)"
        @dblclick="item._isDir && navigateTo(item.id, item.name)"
      >
        <div class="grid-item-icon">
          <el-icon v-if="item._isDir" :size="40" class="folder-icon"><Folder /></el-icon>
          <el-icon v-else :size="40"><Document /></el-icon>
        </div>
        <div class="grid-item-name" :title="item._isDir ? item.name : item.file_name">
          {{ item._isDir ? item.name : item.file_name }}
        </div>
        <div class="grid-item-size">
          {{ item._isDir ? '' : formatSize(item.size) }}
        </div>
        <div class="grid-item-actions">
          <el-dropdown trigger="click" @command="(cmd: string) => handleContextMenu(cmd, item)">
            <el-button icon="MoreFilled" circle size="small" />
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-if="!item._isDir && isPreviewableFile(item)" command="preview" icon="View">预览</el-dropdown-item>
                <el-dropdown-item v-if="!item._isDir" command="download" icon="Download">下载</el-dropdown-item>
                <el-dropdown-item command="rename" icon="Edit">重命名</el-dropdown-item>
                <el-dropdown-item v-if="!item._isDir" command="move" icon="Rank">移动</el-dropdown-item>
                <el-dropdown-item v-if="!item._isDir" command="share" icon="Share">分享</el-dropdown-item>
                <el-dropdown-item v-if="!item._isDir && isExtractable(item)" command="extract" icon="FolderOpened">解压</el-dropdown-item>
                <el-dropdown-item command="delete" icon="Delete" divided>删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </div>

    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @size-change="loadData"
        @current-change="loadData"
      />
    </div>

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

    <!-- 移动文件对话框 -->
    <el-dialog v-model="showMoveDialog" title="移动到" width="450px">
      <div class="move-dialog-content">
        <div class="move-current">
          <span class="move-label">文件：</span>
          <span>{{ movingItem?.file_name || movingItem?.name }}</span>
        </div>
        <div class="move-tree-container">
          <div class="move-tree-node root-node" :class="{ 'is-selected': moveTargetPathId === 0 }" @click="moveTargetPathId = 0">
            <el-icon><HomeFilled /></el-icon>
            <span>根目录</span>
          </div>
          <el-tree
            :data="pathTree"
            :props="{ label: 'name', children: 'children' }"
            node-key="id"
            default-expand-all
            highlight-current
            @current-change="(data: any) => { moveTargetPathId = data.id }"
          />
        </div>
      </div>
      <template #footer>
        <el-button @click="showMoveDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="moveLoading" @click="handleMoveSubmit">移动到此处</el-button>
      </template>
    </el-dialog>

    <!-- 文件预览 -->
    <Preview v-model="showPreview" :file="previewFile" />

    <!-- 解压对话框 -->
    <el-dialog v-model="showExtractDialog" title="解压文件" width="450px">
      <div v-if="extractConflictInfo" class="extract-conflict">
        <el-alert title="检测到同名文件" type="warning" :closable="false" show-icon />
        <p style="margin: 8px 0; font-size: 13px;">目标目录存在 {{ extractConflictInfo.conflict_files?.length || 0 }} 个同名文件</p>
        <el-radio-group v-model="extractStrategy">
          <el-radio value="skip">跳过同名文件</el-radio>
          <el-radio value="overwrite">覆盖同名文件</el-radio>
          <el-radio value="keep_both">保留两者（重命名）</el-radio>
        </el-radio-group>
      </div>
      <div v-else-if="extractTaskStatus === 'extracting'">
        <el-progress :percentage="extractProgress" />
        <p style="margin-top: 8px; font-size: 13px; color: var(--el-text-color-secondary);">
          正在解压... {{ extractCurrentFile }}
        </p>
      </div>
      <div v-else-if="extractTaskStatus === 'done'">
        <el-result icon="success" title="解压完成" />
      </div>
      <div v-else-if="extractTaskStatus === 'failed'">
        <el-result icon="error" title="解压失败" :sub-title="extractErrorMsg" />
      </div>
      <template #footer>
        <el-button @click="showExtractDialog = false">{{ extractTaskStatus === 'done' || extractTaskStatus === 'failed' ? '关闭' : '取消' }}</el-button>
        <el-button v-if="extractConflictInfo" type="primary" :loading="extractLoading" @click="handleExtractStart">开始解压</el-button>
      </template>
    </el-dialog>

    <!-- 分享对话框 -->
    <ShareDialog v-model="showShareDialog" :file-info="shareFileInfo" :custom-share-fn="handleShareCreate" />
  </div>
</template>

<script setup lang="ts">
  import type { ComponentInternalInstance } from 'vue'
  import { enterpriseApi } from '@myobj/api'
  import { upload, download } from '@myobj/http'
  import { API_ENDPOINTS } from '@myobj/shared'
  import type { SharedFileEntry, SpaceUsage, FileItem } from '@myobj/shared'
  import { formatSize } from '@/utils'
  import { useI18n } from '@/composables'
  import { isPreviewable } from '@/utils/ui/preview'
  import Preview from '@/components/Preview/index.vue'
  import ShareDialog from '@/components/ShareDialog/index.vue'

  const enterpriseId = inject<Ref<string>>('enterpriseId', ref(''))

  const {
    getSharedFileList, createSharedDir, deleteSharedFile,
    downloadSharedFile, sharedUploadPrecheck, getSpaceUsage,
    deleteSharedDir, renameSharedFile, renameSharedDir,
    searchSharedFiles, getSharedPathTree, moveSharedFile,
    createPackage, getPackageProgress, downloadPackage,
    extractCheck, extractStart, getExtractProgress,
    previewSharedFile, getSharedFileThumbnail, createShare
  } = enterpriseApi

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  const loading = ref(false)
  const fileList = ref<any[]>([])
  const selectedFiles = ref<any[]>([])
  const currentPathId = ref(0)
  const breadcrumbs = ref<{ id: number; name: string }[]>([])
  const spaceUsage = ref<SpaceUsage | null>(null)
  const viewMode = ref<'list' | 'grid'>('list')

  const pagination = reactive({ page: 1, pageSize: 50, total: 0 })

  // 排序
  const sortBy = ref('created_at')
  const sortOrder = ref('DESC')

  const handleSortChange = ({ prop, order }: { prop: string; order: string }) => {
    if (order) {
      sortBy.value = prop
      sortOrder.value = order === 'ascending' ? 'ASC' : 'DESC'
    } else {
      sortBy.value = 'created_at'
      sortOrder.value = 'DESC'
    }
    loadData()
  }

  // 拖拽上传
  const isDraggingOver = ref(false)
  let dragEnterCount = 0

  const onDragEnter = () => {
    dragEnterCount++
    isDraggingOver.value = true
  }
  const onDragLeave = () => {
    dragEnterCount--
    if (dragEnterCount <= 0) {
      dragEnterCount = 0
      isDraggingOver.value = false
    }
  }
  const onDrop = async (event: DragEvent) => {
    isDraggingOver.value = false
    dragEnterCount = 0
    const files = event.dataTransfer?.files
    if (files && files.length > 0) {
      for (let i = 0; i < files.length; i++) {
        await handleUpload(files[i])
      }
    }
  }

  // 搜索
  const searchKeyword = ref('')
  const isSearching = ref(false)

  // 新建文件夹
  const showMkdirDialog = ref(false)
  const mkdirLoading = ref(false)
  const mkdirFormRef = ref()
  const mkdirForm = reactive({ name: '' })
  const mkdirRules = {
    name: [{ required: true, message: t('enterprise.info.name'), trigger: 'blur' }]
  }

  // 上传
  const uploading = ref(false)
  const uploadProgress = ref(0)

  // 重命名
  const showRenameDialog = ref(false)
  const renameLoading = ref(false)
  const renameFormRef = ref()
  const renameForm = reactive({ name: '' })
  const renameRules = {
    name: [{ required: true, message: t('enterprise.info.name'), trigger: 'blur' }]
  }
  const renamingItem = ref<any>(null)

  // 轮询定时器（组件卸载时清理）
  let packagePollTimer: ReturnType<typeof setInterval> | null = null
  let extractPollTimer: ReturnType<typeof setInterval> | null = null

  onUnmounted(() => {
    if (packagePollTimer) { clearInterval(packagePollTimer); packagePollTimer = null }
    if (extractPollTimer) { clearInterval(extractPollTimer); extractPollTimer = null }
  })

  // 移动
  const showMoveDialog = ref(false)
  const moveLoading = ref(false)
  const movingItem = ref<any>(null)
  const moveTargetPathId = ref(0)
  const pathTree = ref<any[]>([])

  // 预览
  const showPreview = ref(false)
  const previewFile = ref<FileItem | null>(null)

  // 解压
  const showExtractDialog = ref(false)
  const extractLoading = ref(false)
  const extractConflictInfo = ref<any>(null)
  const extractStrategy = ref('skip')
  const extractTaskStatus = ref('')
  const extractProgress = ref(0)
  const extractCurrentFile = ref('')
  const extractErrorMsg = ref('')
  const extractingFileId = ref('')

  // 分享
  const showShareDialog = ref(false)
  const shareFileInfo = reactive({ file_id: '', file_name: '', file_size: 0 })

  // 可预览文件判断
  const isPreviewableFile = (file: any): boolean => {
    if (!file || file._isDir) return false
    return isPreviewable({
      file_id: file.file_id || file.id,
      file_name: file.file_name || file.name,
      file_size: file.size || 0,
      mime_type: file.mime || '',
      is_enc: file.is_enc || false,
      has_thumbnail: !!file.thumbnail,
      public: false,
      created_at: file.created_at || ''
    } as FileItem)
  }

  // 可解压文件判断（当前仅支持 ZIP）
  const isExtractable = (file: any): boolean => {
    if (!file || file._isDir) return false
    const name = (file.file_name || '').toLowerCase()
    return name.endsWith('.zip')
  }

  // 加载文件列表
  const loadFiles = async () => {
    loading.value = true
    isSearching.value = false
    try {
      const res = await getSharedFileList({
        enterprise_id: enterpriseId.value,
        path_id: currentPathId.value || undefined,
        page: pagination.page,
        pageSize: pagination.pageSize,
        sort_by: sortBy.value,
        sort_order: sortOrder.value
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

  // 统一数据加载（根据搜索状态决定调用哪个函数，分页时保留当前页码）
  const loadData = async () => {
    if (isSearching.value && searchKeyword.value.trim()) {
      await doSearch(false)
    } else {
      await loadFiles()
    }
  }

  // 搜索（Enter 触发，重置到第1页）
  const handleSearch = async () => {
    if (!searchKeyword.value.trim()) {
      handleSearchClear()
      return
    }
    pagination.page = 1
    await doSearch(true)
  }

  // 执行搜索（resetPage=true 时重置页码）
  const doSearch = async (resetPage: boolean) => {
    if (!searchKeyword.value.trim()) {
      handleSearchClear()
      return
    }
    if (resetPage) {
      pagination.page = 1
    }
    loading.value = true
    isSearching.value = true
    try {
      const res = await searchSharedFiles({
        enterprise_id: enterpriseId.value,
        keyword: searchKeyword.value.trim(),
        page: pagination.page,
        pageSize: pagination.pageSize
      })
      if (res.code === 200 && res.data) {
        const files = (res.data.files || []).map(f => ({ ...f, _isDir: false }))
        fileList.value = files
        pagination.total = res.data.total || 0
      }
    } catch (error: any) {
      proxy?.$log?.error(error)
    } finally {
      loading.value = false
    }
  }

  const handleSearchClear = () => {
    searchKeyword.value = ''
    isSearching.value = false
    pagination.page = 1
    loadFiles()
  }

  const loadSpaceUsage = async () => {
    try {
      const res = await getSpaceUsage(enterpriseId.value)
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

  // 网格视图点击选中
  const handleGridClick = (item: any) => {
    const idx = selectedFiles.value.findIndex(s => s.id === item.id)
    if (idx >= 0) {
      selectedFiles.value = selectedFiles.value.filter(s => s.id !== item.id)
    } else {
      selectedFiles.value = [...selectedFiles.value, item]
    }
  }

  const handleMkdir = async () => {
    if (!mkdirFormRef.value) return
    await mkdirFormRef.value.validate(async (valid: boolean) => {
      if (!valid) return
      mkdirLoading.value = true
      try {
        const res = await createSharedDir({
          enterprise_id: enterpriseId.value,
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
        enterprise_id: enterpriseId.value,
        file_name: file.name,
        file_size: file.size,
        path_id: currentPathId.value || undefined
      })
      if (!precheckRes || precheckRes.code !== 200) {
        proxy?.$modal.msgError(precheckRes?.message || t('common.operationFailed'))
        return false
      }

      const precheckId = precheckRes.data?.precheck_id
      if (!precheckId) {
        proxy?.$modal.msgError('预检失败：未获取到precheck_id')
        return false
      }

      const formData = new FormData()
      formData.append('enterprise_id', enterpriseId.value)
      formData.append('precheck_id', precheckId)
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
      await downloadSharedFile(file.id, enterpriseId.value)
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('common.operationFailed'))
    }
  }

  // 批量下载（打包 ZIP）
  const handleBatchDownload = async () => {
    if (selectedFiles.value.length === 0) return
    const fileIds = selectedFiles.value.filter(f => !f._isDir).map(f => f.id)
    if (fileIds.length === 0) return

    try {
      const res = await createPackage({ enterprise_id: enterpriseId.value, file_ids: fileIds })
      if (res.code !== 200 || !res.data) {
        proxy?.$modal.msgError(res.message || '创建打包任务失败')
        return
      }

      const packageId = res.data.package_id
      proxy?.$modal.msgSuccess('正在打包，请稍候...')

      // 轮询进度
      if (packagePollTimer) clearInterval(packagePollTimer)
      packagePollTimer = setInterval(async () => {
        try {
          const progressRes = await getPackageProgress(packageId)
          if (progressRes.code === 200 && progressRes.data) {
            const status = progressRes.data.status
            if (status === 'ready') {
              if (packagePollTimer) { clearInterval(packagePollTimer); packagePollTimer = null }
              // 使用 download helper 下载（带认证）
              const url = downloadPackage(packageId)
              await download(url, res.data.package_name || 'enterprise_files.zip')
              proxy?.$modal.msgSuccess('打包完成，开始下载')
            } else if (status === 'failed') {
              if (packagePollTimer) { clearInterval(packagePollTimer); packagePollTimer = null }
              proxy?.$modal.msgError('打包失败')
            }
          }
        } catch {
          if (packagePollTimer) { clearInterval(packagePollTimer); packagePollTimer = null }
        }
      }, 1000)

      // 超时清理
      setTimeout(() => { if (packagePollTimer) { clearInterval(packagePollTimer); packagePollTimer = null } }, 300000)
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || '操作失败')
    }
  }

  // 移动文件
  const handleMoveClick = async (file: any) => {
    if (!file || file._isDir) return
    movingItem.value = file
    moveTargetPathId.value = 0
    showMoveDialog.value = true

    // 加载目录树
    try {
      const res = await getSharedPathTree(enterpriseId.value)
      if (res.code === 200 && res.data) {
        pathTree.value = res.data
      }
    } catch {}
  }

  const handleMoveSubmit = async () => {
    if (!movingItem.value) return
    moveLoading.value = true
    try {
      const res = await moveSharedFile({
        enterprise_id: enterpriseId.value,
        file_id: movingItem.value.id,
        target_path_id: moveTargetPathId.value
      })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess('移动成功')
        showMoveDialog.value = false
        loadFiles()
      } else {
        proxy?.$modal.msgError(res.message || '移动失败')
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || '移动失败')
    } finally {
      moveLoading.value = false
    }
  }

  // 预览
  const handlePreview = (file: any) => {
    if (!file || file._isDir) return
    const fileId = file.file_id || file.id
    previewFile.value = {
      file_id: fileId,
      file_name: file.file_name || file.name,
      file_size: file.size || 0,
      mime_type: file.mime || '',
      is_enc: file.is_enc || false,
      has_thumbnail: !!file.thumbnail,
      public: false,
      created_at: file.created_at || '',
      preview_url: previewSharedFile(fileId),
      thumbnail_url: file.thumbnail ? getSharedFileThumbnail(fileId) : undefined
    } as FileItem
    showPreview.value = true
  }

  // 解压
  const handleExtract = async (file: any) => {
    if (!file || file._isDir) return
    extractingFileId.value = file.id
    extractConflictInfo.value = null
    extractTaskStatus.value = ''
    extractProgress.value = 0
    extractCurrentFile.value = ''
    extractErrorMsg.value = ''
    extractStrategy.value = 'skip'
    showExtractDialog.value = true

    try {
      const checkRes = await extractCheck({
        enterprise_id: enterpriseId.value,
        file_id: file.id,
        target_path_id: currentPathId.value || 0
      })
      if (checkRes.code === 200 && checkRes.data) {
        if (checkRes.data.has_conflict) {
          extractConflictInfo.value = checkRes.data
        } else {
          // 无冲突，直接解压
          await handleExtractStart()
        }
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || '检测失败')
      showExtractDialog.value = false
    }
  }

  const handleExtractStart = async () => {
    extractLoading.value = true
    try {
      const res = await extractStart({
        enterprise_id: enterpriseId.value,
        file_id: extractingFileId.value,
        target_path_id: currentPathId.value || 0,
        conflict_strategy: extractConflictInfo.value ? extractStrategy.value : undefined
      })
      if (res.code === 200 && res.data) {
        extractConflictInfo.value = null
        extractTaskStatus.value = 'extracting'
        // 轮询进度
        const taskId = res.data.task_id
        if (extractPollTimer) clearInterval(extractPollTimer)
        extractPollTimer = setInterval(async () => {
          try {
            const progressRes = await getExtractProgress(taskId)
            if (progressRes.code === 200 && progressRes.data) {
              extractProgress.value = progressRes.data.progress || 0
              extractCurrentFile.value = progressRes.data.current || ''
              extractTaskStatus.value = progressRes.data.status || ''
              if (progressRes.data.status === 'done' || progressRes.data.status === 'failed') {
                if (extractPollTimer) { clearInterval(extractPollTimer); extractPollTimer = null }
                if (progressRes.data.status === 'done') {
                  loadFiles()
                  loadSpaceUsage()
                } else {
                  extractErrorMsg.value = progressRes.data.current || '解压失败'
                }
              }
            }
          } catch {
            if (extractPollTimer) { clearInterval(extractPollTimer); extractPollTimer = null }
          }
        }, 1000)
        setTimeout(() => { if (extractPollTimer) { clearInterval(extractPollTimer); extractPollTimer = null } }, 600000)
      } else {
        proxy?.$modal.msgError(res.message || '解压失败')
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || '解压失败')
    } finally {
      extractLoading.value = false
    }
  }

  // 分享
  const handleShare = (file: any) => {
    if (!file || file._isDir) return
    shareFileInfo.file_id = file.id
    shareFileInfo.file_name = file.file_name || file.name
    shareFileInfo.file_size = file.size || 0
    showShareDialog.value = true
  }

  const handleShareCreate = (data: any) => {
    return createShare({ enterprise_id: enterpriseId.value, ...data })
  }

  // 右键菜单命令处理
  const handleContextMenu = (command: string, row: any) => {
    switch (command) {
      case 'preview':
        handlePreview(row)
        break
      case 'download':
        handleDownload(row)
        break
      case 'rename':
        handleRename(row)
        break
      case 'move':
        handleMoveClick(row)
        break
      case 'share':
        handleShare(row)
        break
      case 'extract':
        handleExtract(row)
        break
      case 'delete':
        handleDelete(row)
        break
    }
  }

  const handleDelete = async (item: any) => {
    try {
      await proxy?.$modal.confirm(t('enterprise.space.deleteConfirm'))
      const res = item._isDir
        ? await deleteSharedDir(item.id, enterpriseId.value)
        : await deleteSharedFile({ id: item.id }, enterpriseId.value)
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
            ? deleteSharedDir(file.id, enterpriseId.value).then(res => ({ res, file }))
            : deleteSharedFile({ id: file.id }, enterpriseId.value).then(res => ({ res, file }))
        )
      )
      const failed = results.filter(r => r.status === 'rejected' || (r.status === 'fulfilled' && r.value.res.code !== 200)).length
      if (failed > 0) {
        proxy?.$modal.msgWarning(`删除完成，${selectedFiles.value.length - failed}/${selectedFiles.value.length} 项成功`)
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
          ? await renameSharedDir(item.id, renameForm.name, enterpriseId.value)
          : await renameSharedFile(item.id, renameForm.name, enterpriseId.value)
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

  watch(enterpriseId, (id) => {
    if (id) {
      currentPathId.value = 0
      breadcrumbs.value = []
      isSearching.value = false
      searchKeyword.value = ''
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
    position: relative;
  }

  .drag-overlay {
    position: absolute;
    inset: 0;
    z-index: 999;
    background: rgba(var(--primary-color-rgb, 64, 158, 255), 0.12);
    border: 3px dashed var(--primary-color);
    border-radius: 12px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    pointer-events: none;
    backdrop-filter: blur(4px);
  }

  .drag-overlay .el-icon {
    color: var(--primary-color);
  }

  .drag-overlay p {
    font-size: 18px;
    font-weight: 600;
    color: var(--primary-color);
    margin: 0;
  }

  .toolbar-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
    padding: 12px 16px;
    border-radius: 12px;
    background: var(--bg-color-glass);
    backdrop-filter: blur(12px);
    border: 1px solid var(--glass-border);
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
    align-items: center;
  }

  .view-toggle {
    margin-left: 4px;
  }

  .upload-progress-bar {
    padding: 0 4px;
  }

  .breadcrumb-bar {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 8px 16px;
    border-radius: 8px;
    background: var(--bg-color-glass);
    backdrop-filter: blur(12px);
    border: 1px solid var(--glass-border);
  }

  .breadcrumb-bar :deep(.el-breadcrumb-item) {
    cursor: pointer;
  }

  .breadcrumb-bar :deep(.el-breadcrumb-item:hover .el-breadcrumb__inner) {
    color: var(--primary-color);
  }

  .search-badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 10px;
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
    border-radius: 12px;
    font-size: 12px;
    white-space: nowrap;
  }

  .clear-search {
    cursor: pointer;
    margin-left: 4px;
  }

  .clear-search:hover {
    color: var(--el-color-danger);
  }

  .usage-bar {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 16px;
    background: linear-gradient(135deg, rgba(37, 99, 235, 0.05), rgba(79, 70, 229, 0.05));
    border: 1px solid rgba(37, 99, 235, 0.1);
    border-radius: 10px;
    font-size: 13px;
  }

  .usage-icon {
    color: var(--primary-color);
    font-size: 20px;
  }

  .usage-info {
    display: flex;
    align-items: center;
    gap: 4px;
    white-space: nowrap;
    color: var(--el-text-color-regular);
  }

  .usage-sep {
    color: var(--el-text-color-placeholder);
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

  .operation-cell {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .file-name-cell .is-folder {
    cursor: pointer;
    color: var(--primary-color);
    font-weight: 600;
    transition: text-decoration 0.2s;
  }

  .file-name-cell .is-folder:hover {
    text-decoration: underline;
  }

  .file-name-cell .is-previewable {
    cursor: pointer;
    color: var(--el-color-primary);
  }

  .file-name-cell .is-previewable:hover {
    text-decoration: underline;
  }

  .file-icon {
    font-size: 18px;
    color: var(--el-text-color-secondary);
  }

  .folder-icon {
    color: var(--el-color-warning);
  }

  /* 网格视图 */
  .file-grid {
    flex: 1;
    overflow: auto;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 12px;
    padding: 4px;
  }

  .grid-empty {
    grid-column: 1 / -1;
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 200px;
  }

  .grid-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 16px 8px 12px;
    border-radius: 10px;
    cursor: pointer;
    transition: all 0.2s;
    position: relative;
    border: 2px solid transparent;
  }

  .grid-item:hover {
    background: var(--el-fill-color-light);
  }

  .grid-item.is-selected {
    background: var(--el-color-primary-light-9);
    border-color: var(--el-color-primary-light-5);
  }

  .grid-item-icon {
    margin-bottom: 8px;
    color: var(--el-text-color-secondary);
  }

  .grid-item-icon .folder-icon {
    color: var(--el-color-warning);
  }

  .grid-item-name {
    font-size: 13px;
    text-align: center;
    word-break: break-all;
    line-height: 1.4;
    max-height: 2.8em;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }

  .grid-item-size {
    font-size: 11px;
    color: var(--el-text-color-placeholder);
    margin-top: 4px;
  }

  .grid-item-actions {
    position: absolute;
    top: 4px;
    right: 4px;
    opacity: 0;
    transition: opacity 0.2s;
  }

  .grid-item:hover .grid-item-actions {
    opacity: 1;
  }

  /* 移动对话框 */
  .move-dialog-content {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .move-current {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background: var(--el-fill-color-light);
    border-radius: 6px;
    font-size: 13px;
  }

  .move-label {
    color: var(--el-text-color-secondary);
    white-space: nowrap;
  }

  .move-tree-container {
    max-height: 300px;
    overflow: auto;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    padding: 8px;
  }

  .move-tree-node {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 8px;
    cursor: pointer;
    border-radius: 4px;
    font-size: 13px;
  }

  .move-tree-node:hover,
  .move-tree-node.is-selected {
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
  }

  .root-node {
    margin-bottom: 8px;
    font-weight: 600;
  }

  .pagination-wrapper {
    margin-top: 8px;
    padding-top: 12px;
    border-top: 1px solid var(--el-border-color-lighter);
    display: flex;
    justify-content: flex-end;
  }

  @media (max-width: 768px) {
    .toolbar-container {
      flex-direction: column;
      align-items: stretch;
    }

    .toolbar-left,
    .toolbar-right {
      width: 100%;
      flex-wrap: wrap;
    }

    .usage-bar {
      flex-direction: column;
      align-items: flex-start;
    }

    .file-grid {
      grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
    }

    .pagination-wrapper :deep(.el-pagination__sizes),
    .pagination-wrapper :deep(.el-pagination__jump) {
      display: none;
    }
  }

  html.dark .toolbar-container {
    background: rgba(15, 23, 42, 0.6);
    border-color: var(--el-border-color);
  }

  html.dark .usage-bar {
    background: linear-gradient(135deg, rgba(37, 99, 235, 0.1), rgba(79, 70, 229, 0.1));
    border-color: rgba(37, 99, 235, 0.2);
  }

  html.dark .breadcrumb-bar {
    background: rgba(15, 23, 42, 0.6);
    border-color: var(--el-border-color);
  }

  .extract-conflict {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .extract-conflict .el-radio-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
</style>

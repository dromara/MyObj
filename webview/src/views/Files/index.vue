<template>
  <div class="files-page">
    <!-- Breadcrumb with Glass effect -->
    <Breadcrumb
      :breadcrumbs="breadcrumbs"
      :format-breadcrumb-name="formatBreadcrumbName"
      :current-path="currentPath"
      :refreshing="fileListLoading"
      @navigate="navigateToPath"
      @refresh="loadFileList"
      @go-back="handleGoBack"
    />

    <!-- Toolbar with Glass effect -->
    <div class="toolbar-container glass-panel">
      <div class="toolbar">
        <div class="toolbar-left">
          <!-- 显示所有按钮 -->
          <div class="toolbar-actions">
            <el-tooltip :content="t('files.upload')" placement="bottom">
              <el-button type="primary" icon="Upload" @click="handleUpload" class="action-btn">{{
                t('files.upload')
              }}</el-button>
            </el-tooltip>
            <el-button icon="FolderAdd" @click="handleNewFolder" class="action-btn-secondary">{{
              t('files.newFolder')
            }}</el-button>
            <el-button
              icon="FolderOpened"
              @click="handleMoveFile"
              :disabled="selectedFileIds.length === 0"
              class="action-btn-secondary"
              >{{ t('files.move') }}</el-button
            >
            <div class="divider-vertical"></div>
            <div class="view-switch glass-toggle">
              <el-button icon="Grid" :class="{ 'is-active': viewMode === 'grid' }" @click="viewMode = 'grid'" text />
              <el-button icon="List" :class="{ 'is-active': viewMode === 'list' }" @click="viewMode = 'list'" text />
            </div>
          </div>
        </div>

        <div class="toolbar-right" :class="{ 'is-visible': selectedCount > 0 }">
          <span class="selection-info desktop-only">{{ t('files.selected', { count: selectedCount }) }}</span>
          <el-button icon="Download" @click="handleToolbarDownload" plain circle />
          <el-button icon="Share" @click="handleToolbarShare" plain circle />
          <el-button icon="Delete" type="danger" @click="handleToolbarDelete" plain circle />
        </div>
      </div>
    </div>

    <!-- 文件列表内容区域 -->
    <div class="file-content-area">
      <!-- 骨架屏加载 -->
      <Skeleton v-if="fileListLoading || isLoading" :count="12" :view-mode="viewMode" />

      <!-- 网格视图 -->
      <FileGrid
        v-else-if="viewMode === 'grid'"
        :folders="displayData.folders"
        :files="displayData.files"
        :is-selected-folder="isSelectedFolder"
        :is-selected-file="isSelectedFile"
        :get-thumbnail-url="getThumbnailUrl"
        @toggle-folder="toggleSelectFolder"
        @toggle-file="toggleSelectFile"
        @enter-folder="enterFolder"
        @preview-file="handleFilePreview"
        @folder-action="handleFolderAction"
        @file-action="handleFileAction"
      />

      <!-- 列表视图 -->
      <FileList
        v-else
        :file-list-data="displayData"
        :get-thumbnail-url="getThumbnailUrl"
        :is-selected-folder="isSelectedFolder"
        :is-selected-file="isSelectedFile"
        @selection-change="handleSelectionChange"
        @toggle-folder="toggleSelectFolder"
        @toggle-file="toggleSelectFile"
        @row-dblclick="handleRowDblClick"
        @download-file="handleDownloadFile"
        @rename-file="handleRenameFile"
        @share-file="handleShareFile"
        @set-file-public="(file, isPublic) => handleSetFilePublic(file, isPublic)"
        @delete-file="handleDeleteFile"
        @rename-dir="handleRenameDir"
        @delete-dir="handleDeleteDir"
      />

      <!-- 空状态 -->
      <el-empty
        v-if="
          !fileListLoading &&
          !isLoading &&
          displayData.folders.length === 0 &&
          displayData.files.length === 0 &&
          !isSearching
        "
        :description="hasSearchKeyword ? t('files.noSearchResults') : t('files.emptyFolder')"
      />
    </div>

    <!-- 分页 -->
    <div v-if="displayPagination.total > 0" class="pagination-wrapper">
      <pagination
        :page="displayPagination.page"
        :limit="displayPagination.pageSize"
        :total="displayPagination.total"
        :page-sizes="[20, 50, 100]"
        float="center"
        @pagination="handlePagination"
        class="pagination"
      />
    </div>

    <!-- 新建文件夹对话框 -->
    <el-dialog v-model="showNewFolderDialog" :title="t('files.newFolder')" width="500px" @close="handleDialogClose">
      <el-form ref="folderFormRef" :model="folderForm" :rules="folderRules" label-width="100px">
        <el-form-item :label="t('files.folderName')" prop="dir_path">
          <el-input
            v-model="folderForm.dir_path"
            :placeholder="t('files.folderNamePlaceholder')"
            clearable
            maxlength="50"
            show-word-limit
            @keyup.enter="handleCreateFolder"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showNewFolderDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreateFolder">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 上传加密配置弹窗 -->
    <UploadEncryptDialog v-model="showUploadEncryptDialog" @confirm="handleUploadEncryptConfirm" />

    <!-- 移动文件对话框 -->
    <el-dialog v-model="showMoveDialog" :title="t('files.move')" width="500px">
      <el-form label-width="100px">
        <el-form-item :label="t('files.selectedFiles')">
          <el-tag v-for="fileId in selectedFileIds" :key="fileId" class="file-tag">
            {{ getFileNameForMove(fileId) }}
          </el-tag>
        </el-form-item>
        <el-form-item :label="t('files.targetFolder')">
          <el-tree-select
            v-model="targetFolderId"
            :data="folderTreeData"
            :render-after-expand="false"
            :placeholder="t('files.targetFolderPlaceholder')"
            :default-expanded-keys="[currentPath]"
            :loading="loadingTree"
            style="width: 100%"
            check-strictly
            node-key="value"
            :props="{ label: 'label', children: 'children' }"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showMoveDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="moving" @click="handleConfirmMove">{{ t('files.confirmMove') }}</el-button>
      </template>
    </el-dialog>

    <!-- 分享文件组件 -->
    <share-dialog
      v-model="showShareDialog"
      :file-info="{
        file_id: shareForm.file_id,
        file_name: shareForm.file_name,
        file_size: getFileSize(shareForm.file_id)
      }"
      @success="handleShareSuccess"
    />

    <!-- 下载密码对话框 -->
    <el-dialog v-model="showDownloadPasswordDialog" :title="t('files.downloadPassword')" width="450px">
      <el-form label-width="100px">
        <el-form-item :label="t('files.fileName')">
          <el-text>{{ downloadPasswordForm.file_name }}</el-text>
        </el-form-item>
        <el-form-item :label="t('files.filePassword')">
          <el-input
            v-model="downloadPasswordForm.file_password"
            type="password"
            :placeholder="t('files.filePasswordPlaceholder')"
            show-password
            @keyup.enter="confirmDownloadPassword"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showDownloadPasswordDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="downloadingFile" @click="confirmDownloadPassword">{{
          t('common.confirm')
        }}</el-button>
      </template>
    </el-dialog>

    <!-- 文件重命名对话框 -->
    <el-dialog
      v-model="showRenameFileDialog"
      :title="t('files.rename')"
      width="500px"
      @close="handleRenameFileDialogClose"
    >
      <el-form ref="renameFileFormRef" :model="renameFileForm" :rules="renameFileRules" label-width="100px">
        <el-form-item :label="t('files.oldFileName')">
          <el-text>{{ renameFileForm.old_file_name }}</el-text>
        </el-form-item>
        <el-form-item :label="t('files.newFileName')" prop="new_file_name">
          <el-input
            v-model="renameFileForm.new_file_name"
            :placeholder="t('files.fileNamePlaceholder')"
            clearable
            maxlength="255"
            show-word-limit
            @keyup.enter="handleConfirmRenameFile"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showRenameFileDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="renamingFile" @click="handleConfirmRenameFile">{{
          t('common.confirm')
        }}</el-button>
      </template>
    </el-dialog>

    <!-- 目录重命名对话框 -->
    <el-dialog
      v-model="showRenameDirDialog"
      :title="t('files.renameDir')"
      width="500px"
      @close="handleRenameDirDialogClose"
    >
      <el-form ref="renameDirFormRef" :model="renameDirForm" :rules="renameDirRules" label-width="100px">
        <el-form-item :label="t('files.oldDirName')">
          <el-text>{{ renameDirForm.old_dir_name }}</el-text>
        </el-form-item>
        <el-form-item :label="t('files.newDirName')" prop="new_dir_name">
          <el-input
            v-model="renameDirForm.new_dir_name"
            :placeholder="t('files.newDirNamePlaceholder')"
            clearable
            maxlength="50"
            show-word-limit
            @keyup.enter="handleConfirmRenameDir"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showRenameDirDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="renamingDir" @click="handleConfirmRenameDir">{{
          t('common.confirm')
        }}</el-button>
      </template>
    </el-dialog>

    <!-- 文件预览组件 -->
    <preview v-model="previewVisible" :file="previewFile" />
  </div>
</template>

<script setup lang="ts">
  import { handleFileUpload } from '@/utils/file/upload'
  import { useI18n } from '@/composables'
  import FileGrid from './components/FileGrid.vue'
  import FileList from './components/FileList.vue'
  import Breadcrumb from './components/Breadcrumb.vue'
  import type { FileItem, FolderItem } from '@/types'

  const { t } = useI18n()

  // 导入 composables
  import { useFileList } from './composables/useFileList'
  import { useFileSelection } from './composables/useFileSelection'
  import { useFileOperations } from './composables/useFileOperations'
  import { useFolderOperations } from './composables/useFolderOperations'
  import { useRename } from './composables/useRename'
  import { useMoveFile } from './composables/useMoveFile'
  import { useFileSearch } from './composables/useFileSearch'

  import { useUserStore } from '@/stores'

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()

  const viewMode = ref<'grid' | 'list'>('grid')

  // 使用 composables
  const {
    fileListData,
    currentPage,
    pageSize,
    currentPath,
    breadcrumbs,
    formatBreadcrumbName,
    loadFileList,
    navigateToPath,
    getThumbnailUrl,
    loading: fileListLoading
  } = useFileList()

  const {
    selectedFolderIds,
    selectedFileIds,
    selectedCount,
    isSelectedFolder,
    toggleSelectFolder,
    isSelectedFile,
    toggleSelectFile,
    handleSelectionChange
  } = useFileSelection()

  const {
    previewVisible,
    previewFile,
    showShareDialog,
    shareForm,
    showDownloadPasswordDialog,
    downloadPasswordForm,
    downloadingFile,
    getFileSize,
    handleShareSuccess,
    handleFilePreview,
    handleShareFile,
    handleDownloadFile,
    confirmDownloadPassword,
    handleDeleteFile,
    handleToolbarDownload,
    handleToolbarShare,
    handleToolbarDelete,
    handleSetFilePublic,
    handleFileAction: handleFileActionFromOps
  } = useFileOperations(fileListData, selectedFileIds, selectedFolderIds, loadFileList)

  const {
    showNewFolderDialog,
    creating,
    folderFormRef,
    folderForm,
    folderRules,
    handleNewFolder,
    handleDialogClose,
    handleCreateFolder
  } = useFolderOperations(currentPath, loadFileList)

  const {
    showRenameFileDialog,
    renamingFile,
    renameFileFormRef,
    renameFileForm,
    renameFileRules,
    showRenameDirDialog,
    renamingDir,
    renameDirFormRef,
    renameDirForm,
    renameDirRules,
    handleRenameFile,
    handleConfirmRenameFile,
    handleRenameFileDialogClose,
    handleRenameDir,
    handleConfirmRenameDir,
    handleRenameDirDialogClose,
    handleFileAction: handleFileActionFromRename,
    handleFolderAction
  } = useRename(selectedFileIds, selectedFolderIds, loadFileList)

  const {
    showMoveDialog,
    moving,
    targetFolderId,
    folderTreeData,
    loadingTree,
    getFileName,
    handleMoveFile,
    handleConfirmMove
  } = useMoveFile(currentPath, selectedFileIds, loadFileList)

  // 使用搜索 composable
  const { searchKeyword, isSearching, searchResults, performSearch, clearSearch, hasSearchKeyword } = useFileSearch()

  // 当前显示的数据（搜索模式或正常模式）
  const displayData = computed(() => {
    if (hasSearchKeyword.value) {
      return searchResults.value
    }
    return fileListData.value
  })

  // 当前显示的分页信息
  const displayPagination = computed(() => {
    if (hasSearchKeyword.value) {
      return {
        page: searchResults.value.page,
        pageSize: searchResults.value.page_size,
        total: searchResults.value.total
      }
    }
    return {
      page: currentPage.value,
      pageSize: pageSize.value,
      total: fileListData.value.total
    }
  })

  // 搜索时显示加载状态
  const isLoading = computed(() => {
    return isSearching.value
  })

  // 合并文件操作处理
  const handleFileAction = (command: string, file: FileItem): void => {
    if (command === 'preview') {
      handleFilePreview(file)
    } else if (command === 'rename') {
      handleFileActionFromRename(command, file)
    } else {
      handleFileActionFromOps(command, file)
    }
  }

  // 进入文件夹
  const enterFolder = (folder: FolderItem) => {
    if (folder.path) {
      navigateToPath(folder.path)
    }
  }

  // 处理列表视图的删除目录事件
  const handleDeleteDir = (folder: FolderItem) => {
    handleFolderAction('delete', folder)
  }

  // 列表视图行双击处理
  const handleRowDblClick = (row: FileItem | (FolderItem & { isFolder: boolean })) => {
    if ('isFolder' in row && row.isFolder) {
      navigateToPath((row as FolderItem).path)
    } else {
      handleFilePreview(row as FileItem)
    }
  }

  // 获取文件名称（用于移动文件对话框）
  const getFileNameForMove = (fileId: string): string => {
    return getFileName(fileId, fileListData)
  }

  // 上传加密配置弹窗显示状态
  const showUploadEncryptDialog = ref(false)

  // 上传文件
  const handleUpload = async () => {
    // 先显示加密配置弹窗
    showUploadEncryptDialog.value = true
  }

  // 处理上传加密配置确认
  const handleUploadEncryptConfirm = async (encryptConfig: { is_enc: boolean; file_password: string }) => {
    await handleFileUpload(
      currentPath.value,
      { chunkSize: 5 * 1024 * 1024 },
      (progress, fileName) => {
        proxy?.$log.debug(`文件 ${fileName} 上传进度: ${progress}%`)
      },
      async fileName => {
        proxy?.$modal.msgSuccess(t('files.uploadSuccess', { fileName }))
        await loadFileList()
        // 上传成功后刷新用户信息，更新存储空间显示
        await userStore?.fetchUserInfo()
      },
      (error, fileName) => {
        proxy?.$log.error(`文件 ${fileName} 上传失败:`, error)
        proxy?.$modal.msgError(t('files.uploadFailed', { fileName, error: error.message }))
      },
      true,
      () => {
        router.push({
          path: '/tasks',
          query: { tab: 'upload' }
        })
      },
      encryptConfig
    )
  }

  // 处理分页事件（统一处理）
  const handlePagination = ({ page, limit }: { page: number; limit: number }) => {
    if (hasSearchKeyword.value) {
      // 搜索模式下的分页
      performSearch(searchKeyword.value, page, limit)
    } else {
      // 正常模式下的分页
      currentPage.value = page
      pageSize.value = limit
      loadFileList()
    }
  }

  // 处理返回上一级
  const handleGoBack = () => {
    if (breadcrumbs.value.length > 1) {
      const previousPath = breadcrumbs.value[breadcrumbs.value.length - 2].path
      navigateToPath(previousPath)
    }
  }

  // 监听 Header 的搜索事件
  onMounted(() => {
    // 监听全局搜索事件
    const handleGlobalSearch = (event: Event) => {
      const customEvent = event as CustomEvent<{ keyword: string }>
      const keyword = customEvent.detail.keyword.trim()

      if (keyword) {
        // 只有当关键词变化时才执行搜索，避免重复请求
        if (searchKeyword.value !== keyword) {
          searchKeyword.value = keyword
          performSearch(keyword, 1, pageSize.value)
        }
      } else {
        // 清空搜索，并重新加载文件列表
        if (hasSearchKeyword.value) {
          clearSearch()
          // 重新加载当前目录的文件列表
          loadFileList()
        }
      }
    }

    window.addEventListener('files-search', handleGlobalSearch)

    // 检查路由参数中是否有搜索关键词
    if (route.query.search && typeof route.query.search === 'string') {
      searchKeyword.value = route.query.search
      performSearch(route.query.search, 1, pageSize.value)
    }

    // 如果路由中没有 virtualPath，确保加载根目录
    if (!route.query.virtualPath && !hasSearchKeyword.value) {
      loadFileList()
    }

    // 清理事件监听
    onBeforeUnmount(() => {
      window.removeEventListener('files-search', handleGlobalSearch)
    })
  })

  // 监听路由变化，清空搜索（当切换目录时）
  watch(
    () => route.query.virtualPath,
    () => {
      // 如果切换了目录，清空搜索
      if (hasSearchKeyword.value) {
        clearSearch()
      }
    }
  )
</script>

<style scoped>
  .files-page {
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 20px;
    overflow: hidden;
    padding: 4px;
  }

  .file-content-area {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .toolbar-container {
    padding: 16px;
    border-radius: 16px;
    flex-shrink: 0;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    min-height: 40px; /* 确保工具栏有最小高度，避免高度变化 */
  }

  .toolbar-right {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0; /* 允许收缩 */
    opacity: 0;
    visibility: hidden;
    transition:
      opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1),
      visibility 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    pointer-events: none; /* 隐藏时禁用交互 */
  }

  .toolbar-right.is-visible {
    opacity: 1;
    visibility: visible;
    pointer-events: auto; /* 显示时启用交互 */
  }

  .selection-info {
    margin-right: 16px;
    font-size: 14px;
    color: var(--text-secondary);
    font-weight: 500;
  }

  .action-btn {
    height: 40px;
    padding: 0 24px;
    border-radius: 10px;
    font-weight: 600;
    box-shadow: 0 4px 12px rgba(99, 102, 241, 0.25);
  }

  .action-btn-secondary {
    height: 40px;
    border-radius: 10px;
    border: 1px solid transparent;
    background: var(--card-bg);
    color: var(--text-regular);
  }

  .action-btn-secondary:hover {
    border-color: var(--primary-color);
    color: var(--primary-color);
    background: var(--card-bg);
  }

  .divider-vertical {
    width: 1px;
    height: 24px;
    background: var(--border-light);
    margin: 0 16px;
  }

  .glass-toggle {
    background: rgba(0, 0, 0, 0.03);
    padding: 4px;
    border-radius: 8px;
    display: flex;
    gap: 2px;
  }

  .glass-toggle .el-button {
    border-radius: 6px;
    padding: 8px;
    height: 32px;
    width: 32px;
    margin: 0;
    color: var(--text-secondary);
  }

  .glass-toggle .el-button.is-active {
    background: var(--card-bg);
    color: var(--primary-color);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  }

  html.dark .glass-toggle {
    background: rgba(255, 255, 255, 0.05);
  }

  html.dark .glass-toggle .el-button.is-active {
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  }

  .pagination-wrapper {
    flex-shrink: 0;
    padding-top: 16px;
    border-top: 1px solid var(--el-border-color-lighter);
  }

  .pagination {
    justify-content: center;
  }

  .file-tag {
    margin-right: 8px;
    margin-bottom: 8px;
  }

  /* 工具栏操作按钮 */
  .toolbar-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .desktop-only {
    display: inline;
  }

  /* 移动端响应式 - 组件特定样式 */
  @media (max-width: 1024px) {
    .files-page {
      gap: 12px;
    }

    .toolbar-container {
      margin-bottom: 12px;
    }

    .toolbar {
      flex-direction: column;
      gap: 12px;
    }

    .toolbar-left {
      width: 100%;
    }

    .toolbar-actions {
      width: 100%;
      gap: 6px;
    }

    .toolbar-actions .action-btn,
    .toolbar-actions .action-btn-secondary {
      flex: 1;
      min-width: 0;
      font-size: 13px;
      padding: 0 12px;
    }

    .toolbar-actions .action-btn {
      flex: 1.2;
    }

    .divider-vertical {
      display: none;
    }

    .view-switch {
      margin-left: auto;
    }

    .toolbar-right {
      width: 100%;
      justify-content: flex-end;
      margin-top: 0;
    }

    .selection-info {
      margin-right: 8px;
      font-size: 12px;
    }
  }

  @media (max-width: 480px) {
    .toolbar-actions {
      gap: 4px;
    }

    .toolbar-actions .action-btn,
    .toolbar-actions .action-btn-secondary {
      font-size: 12px;
      padding: 0 8px;
      height: 36px;
    }

    .toolbar-actions .action-btn-secondary {
      font-size: 11px;
    }
  }
</style>

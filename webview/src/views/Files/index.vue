<template>
  <div class="files-page">
    <!-- Breadcrumb with Glass effect -->
    <Breadcrumb 
      :breadcrumbs="breadcrumbs"
      :format-breadcrumb-name="formatBreadcrumbName"
      @navigate="navigateToPath"
    />

    <!-- Toolbar with Glass effect -->
    <div class="toolbar-container glass-panel">
      <div class="toolbar">
        <div class="toolbar-left">
          <!-- 显示所有按钮 -->
          <div class="toolbar-actions">
            <el-tooltip content="上传文件" placement="bottom">
              <el-button type="primary" icon="Upload" @click="handleUpload" class="action-btn">上传</el-button>
            </el-tooltip>
            <el-button icon="FolderAdd" @click="handleNewFolder" class="action-btn-secondary">新建文件夹</el-button>
            <el-button icon="FolderOpened" @click="handleMoveFile" :disabled="selectedFileIds.length === 0" class="action-btn-secondary">移动文件</el-button>
            <div class="divider-vertical"></div>
            <div class="view-switch glass-toggle">
              <el-button icon="Grid" :class="{ 'is-active': viewMode === 'grid' }" @click="viewMode = 'grid'" text />
              <el-button icon="List" :class="{ 'is-active': viewMode === 'list' }" @click="viewMode = 'list'" text />
            </div>
          </div>
        </div>
        
        <div class="toolbar-right" v-if="selectedCount > 0">
          <span class="selection-info desktop-only">已选 {{ selectedCount }} 项</span>
          <el-button icon="Download" @click="handleToolbarDownload" plain circle />
          <el-button icon="Share" @click="handleToolbarShare" plain circle />
          <el-button icon="Delete" type="danger" @click="handleToolbarDelete" plain circle />
        </div>
      </div>
    </div>
    
    <!-- 文件列表内容区域 -->
    <div class="file-content-area" v-loading="isLoading">
      <!-- 网格视图 -->
      <FileGrid
        v-if="viewMode === 'grid'"
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
        v-if="displayData.folders.length === 0 && displayData.files.length === 0 && !isSearching" 
        :description="hasSearchKeyword ? '未找到匹配的文件' : '暂无文件'" 
      />
    </div>
    
    <!-- 分页 -->
    <pagination
      v-if="displayPagination.total > 0"
      :page="displayPagination.page"
      :limit="displayPagination.pageSize"
      :total="displayPagination.total"
      :page-sizes="[20, 50, 100]"
      float="center"
      @pagination="handlePagination"
      class="pagination"
    />
    
    <!-- 新建文件夹对话框 -->
    <el-dialog 
      v-model="showNewFolderDialog" 
      title="新建文件夹" 
      width="500px"
      @close="handleDialogClose"
    >
      <el-form 
        ref="folderFormRef" 
        :model="folderForm" 
        :rules="folderRules"
        label-width="100px"
      >
        <el-form-item label="文件夹名称" prop="dir_path">
          <el-input 
            v-model="folderForm.dir_path" 
            placeholder="请输入文件夹名称"
            clearable
            maxlength="50"
            show-word-limit
            @keyup.enter="handleCreateFolder"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showNewFolderDialog = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreateFolder">确定</el-button>
      </template>
    </el-dialog>
    
    <!-- 移动文件对话框 -->
    <el-dialog 
      v-model="showMoveDialog" 
      title="移动文件" 
      width="500px"
    >
      <el-form label-width="100px">
        <el-form-item label="选中文件">
          <el-tag v-for="fileId in selectedFileIds" :key="fileId" class="file-tag">
            {{ getFileNameForMove(fileId) }}
          </el-tag>
        </el-form-item>
        <el-form-item label="目标目录">
          <el-tree-select
            v-model="targetFolderId"
            :data="folderTreeData"
            :render-after-expand="false"
            placeholder="请选择目标目录"
            :default-expanded-keys="[currentPath]"
            :loading="loadingTree"
            style="width: 100%"
            check-strictly
            :props="{ label: 'label', value: 'value', children: 'children' }"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showMoveDialog = false">取消</el-button>
        <el-button type="primary" :loading="moving" @click="handleConfirmMove">确定移动</el-button>
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
    <el-dialog 
      v-model="showDownloadPasswordDialog" 
      title="输入文件密码" 
      width="450px"
    >
      <el-form label-width="100px">
        <el-form-item label="文件名称">
          <el-text>{{ downloadPasswordForm.file_name }}</el-text>
        </el-form-item>
        <el-form-item label="文件密码">
          <el-input 
            v-model="downloadPasswordForm.file_password" 
            type="password"
            placeholder="请输入文件加密密码"
            show-password
            @keyup.enter="confirmDownloadPassword"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showDownloadPasswordDialog = false">取消</el-button>
        <el-button type="primary" :loading="downloadingFile" @click="confirmDownloadPassword">确定</el-button>
      </template>
    </el-dialog>

    <!-- 文件重命名对话框 -->
    <el-dialog 
      v-model="showRenameFileDialog" 
      title="重命名文件" 
      width="500px"
      @close="handleRenameFileDialogClose"
    >
      <el-form 
        ref="renameFileFormRef" 
        :model="renameFileForm" 
        :rules="renameFileRules"
        label-width="100px"
      >
        <el-form-item label="原文件名">
          <el-text>{{ renameFileForm.old_file_name }}</el-text>
        </el-form-item>
        <el-form-item label="新文件名" prop="new_file_name">
          <el-input 
            v-model="renameFileForm.new_file_name" 
            placeholder="请输入新文件名"
            clearable
            maxlength="255"
            show-word-limit
            @keyup.enter="handleConfirmRenameFile"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showRenameFileDialog = false">取消</el-button>
        <el-button type="primary" :loading="renamingFile" @click="handleConfirmRenameFile">确定</el-button>
      </template>
    </el-dialog>

    <!-- 目录重命名对话框 -->
    <el-dialog 
      v-model="showRenameDirDialog" 
      title="重命名目录" 
      width="500px"
      @close="handleRenameDirDialogClose"
    >
      <el-form 
        ref="renameDirFormRef" 
        :model="renameDirForm" 
        :rules="renameDirRules"
        label-width="100px"
      >
        <el-form-item label="原目录名">
          <el-text>{{ renameDirForm.old_dir_name }}</el-text>
        </el-form-item>
        <el-form-item label="新目录名" prop="new_dir_name">
          <el-input 
            v-model="renameDirForm.new_dir_name" 
            placeholder="请输入新目录名"
            clearable
            maxlength="50"
            show-word-limit
            @keyup.enter="handleConfirmRenameDir"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showRenameDirDialog = false">取消</el-button>
        <el-button type="primary" :loading="renamingDir" @click="handleConfirmRenameDir">确定</el-button>
      </template>
    </el-dialog>

    <!-- 文件预览组件 -->
    <preview v-model="previewVisible" :file="previewFile" />
  </div>
</template>

<script setup lang="ts">
import { handleFileUpload } from '@/utils/upload'
import FileGrid from './modules/FileGrid.vue'
import FileList from './modules/FileList.vue'
import Breadcrumb from './modules/Breadcrumb.vue'
import type { FileItem, FolderItem } from '@/types'

// 导入 composables
import { useFileList } from './modules/useFileList'
import { useFileSelection } from './modules/useFileSelection'
import { useFileOperations } from './modules/useFileOperations'
import { useFolderOperations } from './modules/useFolderOperations'
import { useRename } from './modules/useRename'
import { useMoveFile } from './modules/useMoveFile'
import { useFileSearch } from './modules/useFileSearch'

const { proxy } = getCurrentInstance() as ComponentInternalInstance
const route = useRoute()

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
  getThumbnailUrl
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
const {
  searchKeyword,
  isSearching,
  searchResults,
  performSearch,
  clearSearch,
  hasSearchKeyword
} = useFileSearch()

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

// 上传文件
const router = useRouter()
const handleUpload = async () => {
  await handleFileUpload(
    currentPath.value,
    { chunkSize: 5 * 1024 * 1024 },
    (progress, fileName) => {
      proxy?.$log.debug(`文件 ${fileName} 上传进度: ${progress}%`)
    },
    (fileName) => {
      proxy?.$modal.msgSuccess(`文件 ${fileName} 上传成功`)
      loadFileList()
    },
    (error, fileName) => {
      proxy?.$log.error(`文件 ${fileName} 上传失败:`, error)
      proxy?.$modal.msgError(`文件 ${fileName} 上传失败: ${error.message}`)
    },
    true,
    () => {
      router.push({
        path: '/tasks',
        query: { tab: 'upload' }
      })
    }
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
  gap: 16px;
  overflow: hidden;
  /* 移除 position 和 z-index，避免影响侧边栏 */
}

.file-content-area {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: auto; /* 允许横向滚动，以便在小屏幕上查看完整表格 */
}


.toolbar-container {
  padding: 16px;
  border-radius: 16px;
  margin-bottom: 20px;
  /* 移除 position 和 z-index，避免影响侧边栏 */
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
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
  background: white;
  color: var(--text-regular);
}

.action-btn-secondary:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
  background: white;
}

.divider-vertical {
  width: 1px;
  height: 24px;
  background: var(--border-light);
  margin: 0 16px;
}

.glass-toggle {
  background: rgba(0,0,0,0.03);
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
  background: white;
  color: var(--primary-color);
  box-shadow: 0 2px 4px rgba(0,0,0,0.05);
}



.pagination {
  margin-top: 16px;
  justify-content: center;
  flex-shrink: 0;
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

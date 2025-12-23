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
          <!-- 移动端：使用下拉菜单 -->
          <el-dropdown
            class="mobile-toolbar-menu"
            trigger="click"
            @command="handleToolbarCommand"
          >
            <el-button type="primary" icon="More" circle />
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="upload" icon="Upload">上传文件</el-dropdown-item>
                <el-dropdown-item command="newFolder" icon="FolderAdd">新建文件夹</el-dropdown-item>
                <el-dropdown-item command="moveFile" :disabled="selectedFileIds.length === 0" icon="FolderOpened">移动文件</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          
          <!-- 桌面端：显示所有按钮 -->
          <div class="desktop-toolbar">
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
    <div class="file-content-area">
      <!-- 网格视图 -->
      <FileGrid
        v-if="viewMode === 'grid'"
        :folders="fileListData.folders"
        :files="fileListData.files"
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
        :file-list-data="fileListData"
        :get-thumbnail-url="getThumbnailUrl"
        @selection-change="handleSelectionChange"
        @row-dblclick="handleRowDblClick"
        @download-file="handleDownloadFile"
        @rename-file="handleRenameFile"
        @share-file="handleShareFile"
        @delete-file="handleDeleteFile"
        @rename-dir="handleRenameDir"
      />
      
      <!-- 空状态 -->
      <el-empty v-if="fileListData.folders.length === 0 && fileListData.files.length === 0" description="暂无文件" />
    </div>
    
    <!-- 分页 -->
    <el-pagination
      v-if="fileListData.total > 0"
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :page-sizes="[20, 50, 100]"
      :total="fileListData.total"
      layout="total, sizes, prev, pager, next, jumper"
      @size-change="handleSizeChange"
      @current-change="handlePageChange"
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
    <Share
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
    <Preview v-model="previewVisible" :file="previewFile" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, getCurrentInstance, ComponentInternalInstance } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { handleFileUpload } from '@/utils/upload'
import Preview from '@/components/Preview/index.vue'
import Share from '@/components/Share/index.vue'
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
  getThumbnailUrl,
  handlePageChange,
  handleSizeChange
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

// 合并文件操作处理
const handleFileAction = (command: string, file: FileItem): void => {
  if (command === 'rename') {
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

// 移动端工具栏菜单命令处理
const handleToolbarCommand = (command: string) => {
  switch (command) {
    case 'upload':
      handleUpload()
      break
    case 'newFolder':
      handleNewFolder()
      break
    case 'moveFile':
      handleMoveFile()
      break
  }
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

// 初始化
// 注意：路由监听已在 useFileList 中处理，这里不需要手动设置
// 如果路由中有 virtualPath，watch 会自动处理
onMounted(() => {
  // 如果路由中没有 virtualPath，确保加载根目录
  if (!route.query.virtualPath) {
    loadFileList()
  }
  // 如果有 virtualPath，watch 会自动触发 loadFileList
})
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

/* 移动端工具栏 */
.mobile-toolbar-menu {
  display: none;
}

.desktop-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
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

  .toolbar-left {
    flex: 1;
    min-width: 0;
  }

  .toolbar-right {
    flex: 1 1 100%;
    justify-content: flex-end;
    margin-top: 8px;
  }

  .selection-info {
    margin-right: 8px;
    font-size: 12px;
  }
}
</style>

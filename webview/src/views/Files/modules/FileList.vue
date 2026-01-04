<template>
  <div class="file-list-wrapper">
    <!-- PC端：表格布局 -->
    <el-table
      v-if="!isMobile"
      :data="tableData"
      @selection-change="$emit('selection-change', $event)"
      :row-class-name="getRowClassName"
      class="file-table"
    >
    <el-table-column type="selection" width="55" class-name="mobile-hide" />
    <el-table-column label="名称" min-width="300" class-name="mobile-name-column">
      <template #default="{ row }">
        <div 
          class="file-name-cell" 
          :class="{ 'folder-cell': row.isFolder, 'file-cell': !row.isFolder }"
          @click="handleRowClick(row)"
          @dblclick="handleRowDblClick(row)"
        >
          <!-- 文件夹图标 -->
          <el-icon v-if="row.isFolder" :size="32" color="#409EFF" class="folder-icon">
            <Folder />
          </el-icon>
          <!-- 文件图标 -->
          <div v-else class="list-file-icon">
            <file-icon
              :mime-type="row.mime_type"
              :file-name="row.file_name"
              :thumbnail-url="getThumbnailUrl(row.file_id)"
              :show-thumbnail="row.has_thumbnail"
              :icon-size="24"
              :show-badge="false"
              :is-encrypted="row.is_enc"
            />
          </div>
          <div class="file-name-wrapper">
            <file-name-tooltip
              v-if="!row.isFolder"
              :file-name="row.file_name"
              view-mode="table"
            />
            <span v-else class="file-name-text folder-name">{{ row.name }}</span>
            <div v-if="!row.isFolder" class="file-tags-inline">
              <el-tag v-if="row.is_enc" size="small" type="warning" class="enc-tag-inline">
                <el-icon :size="12"><Lock /></el-icon>
                加密
              </el-tag>
              <el-tag v-if="row.public" size="small" type="success" class="public-tag-inline" effect="plain">
                <el-icon :size="12"><Share /></el-icon>
                公开
              </el-tag>
            </div>
          </div>
        </div>
      </template>
    </el-table-column>
    <el-table-column label="大小" width="120" class-name="mobile-hide">
      <template #default="{ row }">
        {{ row.isFolder ? '-' : formatSize(row.file_size) }}
      </template>
    </el-table-column>
    <el-table-column label="创建时间" width="180" class-name="mobile-hide">
      <template #default="{ row }">
        {{ formatDate(row.isFolder ? row.created_time : row.created_at) }}
      </template>
    </el-table-column>
    <el-table-column label="操作" width="120" fixed="right" class-name="mobile-actions-column" align="center">
      <template #default="{ row }">
        <div class="action-buttons">
          <!-- 移动端：显示预览按钮（如果可预览） -->
          <el-button 
            v-if="isMobile && !row.isFolder && isPreviewableFile(row)"
            icon="View" 
            circle 
            text 
            size="small"
            class="preview-btn"
            @click.stop="$emit('row-dblclick', row)"
          />
          <el-dropdown trigger="click" @command="(cmd: string) => handleAction(cmd, row)">
            <el-button icon="More" circle text size="small" />
            <template #dropdown>
              <el-dropdown-menu>
                <template v-if="!row.isFolder">
                  <el-dropdown-item 
                    v-if="isPreviewableFile(row)"
                    command="preview" 
                    icon="View"
                  >
                    预览
                  </el-dropdown-item>
                  <el-dropdown-item command="download" icon="Download">下载</el-dropdown-item>
                  <el-dropdown-item command="rename" icon="Edit">重命名</el-dropdown-item>
                  <el-dropdown-item command="share" icon="Share">分享</el-dropdown-item>
                  <el-dropdown-item 
                    v-if="!row.is_enc"
                    :command="row.public ? 'setPrivate' : 'setPublic'" 
                    :icon="row.public ? 'Lock' : 'Unlock'"
                  >
                    {{ row.public ? '取消公开' : '设为公开' }}
                  </el-dropdown-item>
                  <el-dropdown-item command="delete" icon="Delete" divided>删除</el-dropdown-item>
                </template>
                <template v-else>
                  <el-dropdown-item command="rename" icon="Edit">重命名</el-dropdown-item>
                  <el-dropdown-item command="delete" icon="Delete" divided>删除</el-dropdown-item>
                </template>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </template>
    </el-table-column>
    </el-table>

    <!-- 移动端：卡片布局 -->
    <div v-else class="mobile-file-list">
      <div
        v-for="row in tableData"
        :key="row.isFolder ? `folder-${row.id}` : `file-${row.file_id}`"
        class="mobile-file-item"
        :class="{ 'folder-item': row.isFolder, 'file-item': !row.isFolder, 'selected': isSelected(row) }"
        @click="handleMobileItemClick(row)"
      >
        <div class="mobile-item-content">
          <div class="mobile-item-icon">
            <!-- 文件夹图标 -->
            <el-icon v-if="row.isFolder" :size="40" color="#409EFF">
              <Folder />
            </el-icon>
            <!-- 文件图标 -->
            <file-icon
              v-else
              :mime-type="row.mime_type"
              :file-name="row.file_name"
              :thumbnail-url="getThumbnailUrl(row.file_id)"
              :show-thumbnail="row.has_thumbnail"
              :icon-size="40"
              :show-badge="false"
              :is-encrypted="row.is_enc"
            />
          </div>
          <div class="mobile-item-info">
            <div class="mobile-item-name-row">
              <file-name-tooltip
                v-if="!row.isFolder"
                :file-name="row.file_name"
                view-mode="list"
                custom-class="mobile-item-name"
              />
              <span v-else class="mobile-item-name folder-name">{{ row.name }}</span>
              <div v-if="!row.isFolder" class="mobile-file-tags">
                <el-tag v-if="row.is_enc" size="small" type="warning" class="mobile-enc-tag">
                  <el-icon :size="10"><Lock /></el-icon>
                  加密
                </el-tag>
                <el-tag v-if="row.public" size="small" type="success" class="mobile-public-tag">
                  <el-icon :size="10"><Share /></el-icon>
                  公开
                </el-tag>
              </div>
            </div>
            <div class="mobile-item-meta">
              <span v-if="!row.isFolder" class="mobile-item-size">{{ formatSize(row.file_size) }}</span>
              <span class="mobile-item-time">{{ formatDate(row.isFolder ? row.created_time : row.created_at) }}</span>
            </div>
          </div>
          <div class="mobile-item-actions" @click.stop>
            <!-- 预览按钮（仅文件且可预览） -->
            <el-button
              v-if="!row.isFolder && isPreviewableFile(row)"
              icon="View"
              circle
              text
              size="small"
              class="mobile-action-btn"
              @click.stop="$emit('row-dblclick', row)"
            />
            <el-dropdown trigger="click" @command="(cmd: string) => handleAction(cmd, row)">
              <el-button icon="More" circle text size="small" class="mobile-action-btn" />
              <template #dropdown>
                <el-dropdown-menu>
                  <template v-if="!row.isFolder">
                    <el-dropdown-item
                      v-if="isPreviewableFile(row)"
                      command="preview"
                      icon="View"
                    >
                      预览
                    </el-dropdown-item>
                    <el-dropdown-item command="download" icon="Download">下载</el-dropdown-item>
                    <el-dropdown-item command="rename" icon="Edit">重命名</el-dropdown-item>
                    <el-dropdown-item command="share" icon="Share">分享</el-dropdown-item>
                    <el-dropdown-item 
                      v-if="!row.is_enc"
                      :command="row.public ? 'setPrivate' : 'setPublic'" 
                      :icon="row.public ? 'Lock' : 'Unlock'"
                    >
                      {{ row.public ? '取消公开' : '设为公开' }}
                    </el-dropdown-item>
                    <el-dropdown-item command="delete" icon="Delete" divided>删除</el-dropdown-item>
                  </template>
                  <template v-else>
                    <el-dropdown-item command="rename" icon="Edit">重命名</el-dropdown-item>
                    <el-dropdown-item command="delete" icon="Delete" divided>删除</el-dropdown-item>
                  </template>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatSize, formatDate } from '@/utils'
import { useResponsive } from '@/composables/useResponsive'
import { isPreviewable } from '@/utils/preview'
import type { FileItem, FileListResponse, FolderItem } from '@/types'

const props = defineProps<{
  fileListData: FileListResponse
  getThumbnailUrl: (fileId: string) => string
  isSelectedFolder?: (id: number) => boolean
  isSelectedFile?: (id: string) => boolean
}>()

const tableData = computed(() => {
  return [
    ...props.fileListData.folders.map((f: any) => ({ ...f, isFolder: true })),
    ...props.fileListData.files.map((f: any) => ({ ...f, isFolder: false }))
  ]
})

const emit = defineEmits<{
  'selection-change': [selection: Array<{ isFolder?: boolean; id?: number; file_id?: string }>]
  'row-dblclick': [row: FileItem | FolderItem & { isFolder: boolean }]
  'toggle-folder': [id: number]
  'toggle-file': [id: string]
  'download-file': [file: FileItem]
  'rename-file': [file: FileItem]
  'share-file': [file: FileItem]
  'set-file-public': [file: FileItem, isPublic: boolean]
  'delete-file': [file: FileItem]
  'rename-dir': [folder: FolderItem]
  'delete-dir': [folder: FolderItem]
}>()

// 使用响应式检测 composable
const { isMobile } = useResponsive()

// 检查是否选中
const isSelected = (row: FileItem | (FolderItem & { isFolder: boolean })): boolean => {
  if ('isFolder' in row && row.isFolder) {
    return props.isSelectedFolder ? props.isSelectedFolder(row.id) : false
  } else {
    const file = row as FileItem
    return props.isSelectedFile ? props.isSelectedFile(file.file_id) : false
  }
}

// 判断文件是否可预览
const isPreviewableFile = (file: FileItem | (FolderItem & { isFolder: boolean })): boolean => {
  if ('isFolder' in file && file.isFolder) {
    return false
  }
  return isPreviewable(file as FileItem)
}

// 获取行的类名（用于区分文件夹和文件）
const getRowClassName = ({ row }: { row: any }) => {
  return row.isFolder ? 'folder-row' : 'file-row'
}

// 处理行点击（文件夹单机进入，文件用于选择）
const handleRowClick = (row: FileItem | (FolderItem & { isFolder: boolean })) => {
  if ('isFolder' in row && row.isFolder) {
    // 文件夹：单击进入
    emit('row-dblclick', row)
  } else {
    // 文件：用于选择（由表格的 selection-change 处理）
    // 这里不做任何操作，保持原有的选择行为
  }
}

// 处理移动端卡片点击
const handleMobileItemClick = (row: FileItem | (FolderItem & { isFolder: boolean })) => {
  if ('isFolder' in row && row.isFolder) {
    // 文件夹：单击进入
    emit('row-dblclick', row)
  } else {
    // 文件：切换选择状态
    const file = row as FileItem
    emit('toggle-file', file.file_id)
  }
}

// 处理行双击
const handleRowDblClick = (row: FileItem | (FolderItem & { isFolder: boolean })) => {
  emit('row-dblclick', row)
}

const handleAction = (command: string, row: FileItem | (FolderItem & { isFolder: boolean })) => {
  if ('isFolder' in row && row.isFolder) {
    if (command === 'rename') {
      emit('rename-dir', row as FolderItem)
    } else if (command === 'delete') {
      emit('delete-dir', row as FolderItem)
    }
  } else {
    switch (command) {
      case 'preview':
        emit('row-dblclick', row)
        break
      case 'download':
        emit('download-file', row as FileItem)
        break
      case 'rename':
        emit('rename-file', row as FileItem)
        break
      case 'share':
        emit('share-file', row as FileItem)
        break
      case 'setPublic':
        emit('set-file-public', row as FileItem, true)
        break
      case 'setPrivate':
        emit('set-file-public', row as FileItem, false)
        break
      case 'delete':
        emit('delete-file', row as FileItem)
        break
    }
  }
}
</script>

<style scoped>
.file-list-wrapper {
  width: 100%;
  overflow-x: auto;
  overflow-y: visible;
}

.file-table {
  width: 100%;
  min-width: 775px; /* 确保表格最小宽度：55(选择) + 300(名称) + 120(大小) + 180(时间) + 120(操作) */
}

/* 表格行 hover 效果 */
.file-table :deep(.el-table__body tr:hover) {
  background-color: var(--el-table-row-hover-bg-color) !important;
}

.file-table :deep(.el-table__body tr.folder-row:hover) {
  background-color: rgba(64, 158, 255, 0.06) !important;
}

.file-table :deep(.el-table__body tr.folder-row:hover .file-name-cell) {
  background: transparent !important;
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  margin: -8px -12px;
  border-radius: 6px;
  transition: all 0.2s ease;
}

/* 文件夹样式 */
.file-name-cell.folder-cell {
  cursor: pointer;
}

.file-name-cell.folder-cell:hover {
  background: rgba(64, 158, 255, 0.1) !important;
}

.file-name-cell.folder-cell:active {
  background: rgba(64, 158, 255, 0.15) !important;
}

.folder-icon {
  flex-shrink: 0;
  transition: transform 0.2s ease;
}

.file-name-cell.folder-cell:hover .folder-icon {
  transform: scale(1.1);
}

/* 文件样式 */
.file-name-cell.file-cell {
  cursor: default;
}

.file-name-wrapper {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  overflow: hidden;
}

.file-name-text {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--el-text-color-primary, #303133) !important;
  font-size: 14px;
  line-height: 1.5;
}

.file-name-text.folder-name {
  color: var(--el-color-primary) !important;
  font-weight: 500;
}

.file-tags-inline {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
  white-space: nowrap;
}

.enc-tag-inline {
  border: none;
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
  font-size: 11px;
  padding: 2px 8px;
  height: 20px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border-radius: 4px;
  font-weight: 500;
}

.public-tag-inline {
  border: 1px solid rgba(16, 185, 129, 0.3);
  background: rgba(16, 185, 129, 0.08);
  color: #10b981;
  font-size: 11px;
  padding: 2px 6px;
  height: 20px;
  display: inline-flex;
  align-items: center;
  gap: 3px;
  border-radius: 4px;
  font-weight: 500;
  transition: all 0.2s;
  white-space: nowrap;
  line-height: 1;
}

.public-tag-inline :deep(.el-tag__content) {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  white-space: nowrap;
  line-height: 1;
}

.public-tag-inline :deep(.el-icon) {
  flex-shrink: 0;
}

.public-tag-inline:hover {
  background: rgba(16, 185, 129, 0.12);
  border-color: rgba(16, 185, 129, 0.4);
}

.public-tag-inline .el-icon {
  color: #10b981;
}

.list-file-icon {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
}

.file-table :deep(.mobile-name-column) {
  min-width: auto;
  width: 100%;
}

.file-table :deep(.mobile-name-column .cell) {
  padding: 0 !important;
}

.file-table :deep(.mobile-actions-column) {
  width: auto;
  min-width: 80px;
}

/* 表格移动端隐藏列 - 只在移动端生效 */
@media (max-width: 1024px) {
  .file-table :deep(.mobile-hide) {
    display: none;
  }
  
  .file-table :deep(.mobile-name-column) {
    min-width: auto;
    width: 100%;
  }
}

.file-table :deep(.mobile-actions-column .cell) {
  padding: 8px 4px !important;
  text-align: center;
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 4px;
  justify-content: center;
}

.preview-btn {
  color: var(--el-color-primary);
}

@media (max-width: 1024px) {
  .file-table :deep(.mobile-actions-column) {
    width: auto;
    min-width: 80px;
  }
  
  .file-table :deep(.mobile-actions-column .el-button) {
    padding: 4px 8px;
    font-size: 12px;
  }
}

@media (max-width: 480px) {
  .file-table :deep(.mobile-actions-column) {
    width: auto;
    min-width: 60px;
  }
  
  .file-table :deep(.mobile-actions-column .el-button) {
    padding: 2px 4px;
    font-size: 11px;
  }
}

/* 移动端卡片布局 */
.mobile-file-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px 0;
}

.mobile-file-item {
  background: white;
  border-radius: 12px;
  padding: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  transition: all 0.2s ease;
  border: 2px solid transparent;
}

.mobile-file-item:active {
  transform: scale(0.98);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}

.mobile-file-item.folder-item {
  cursor: pointer;
}

.mobile-file-item.folder-item:active {
  background: rgba(64, 158, 255, 0.05);
  border-color: rgba(64, 158, 255, 0.2);
}

.mobile-file-item.selected {
  border-color: var(--el-color-primary);
  background: rgba(64, 158, 255, 0.04);
}

.mobile-item-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.mobile-item-icon {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.mobile-item-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.mobile-item-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.mobile-item-name {
  flex: 1;
  min-width: 0;
  font-size: 15px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mobile-item-name.folder-name {
  color: var(--el-color-primary);
  font-weight: 600;
}

.mobile-file-tags {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-left: 6px;
  flex-shrink: 0;
}

.mobile-enc-tag {
  border: none;
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
  font-size: 10px;
  padding: 2px 6px;
  height: 18px;
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.mobile-public-tag {
  border: none;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  font-size: 10px;
  padding: 2px 6px;
  height: 18px;
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.mobile-item-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.mobile-item-size {
  font-weight: 500;
}

.mobile-item-time {
  color: var(--el-text-color-placeholder);
}

.mobile-item-actions {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 4px;
}

.mobile-action-btn {
  color: var(--el-text-color-regular);
}

.mobile-action-btn:hover {
  color: var(--el-color-primary);
}

@media (max-width: 480px) {
  .mobile-file-item {
    padding: 10px;
    border-radius: 10px;
  }

  .mobile-item-icon {
    width: 36px;
    height: 36px;
  }

  .mobile-item-name {
    font-size: 14px;
  }

  .mobile-item-meta {
    font-size: 11px;
    gap: 8px;
  }
}
</style>


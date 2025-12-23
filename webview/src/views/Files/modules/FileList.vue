<template>
  <div class="file-list-wrapper">
    <el-table
      :data="tableData"
      @selection-change="$emit('selection-change', $event)"
      class="file-table"
    >
    <el-table-column type="selection" width="55" class-name="mobile-hide" />
    <el-table-column label="名称" min-width="300" class-name="mobile-name-column">
      <template #default="{ row }">
        <div class="file-name-cell" @dblclick="$emit('row-dblclick', row)">
          <!-- 文件夹图标 -->
          <el-icon v-if="row.isFolder" :size="32" color="#409EFF">
            <Folder />
          </el-icon>
          <!-- 文件图标 -->
          <div v-else class="list-file-icon">
            <FileIcon
              :mime-type="row.mime_type"
              :file-name="row.file_name"
              :thumbnail-url="getThumbnailUrl(row.file_id)"
              :show-thumbnail="row.has_thumbnail"
              :icon-size="24"
              :show-badge="false"
              :is-encrypted="row.is_enc"
            />
          </div>
          <span class="file-name-text">{{ row.isFolder ? row.name : row.file_name }}</span>
          <el-tag v-if="!row.isFolder && row.is_enc" size="small" type="warning" class="enc-tag-inline">
            <el-icon :size="12"><Lock /></el-icon>
            加密
          </el-tag>
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
        <el-dropdown trigger="click" @command="(cmd: string) => handleAction(cmd, row)">
          <el-button icon="More" circle text size="small" />
          <template #dropdown>
            <el-dropdown-menu>
              <template v-if="!row.isFolder">
                <el-dropdown-item command="download" icon="Download">下载</el-dropdown-item>
                <el-dropdown-item command="rename" icon="Edit">重命名</el-dropdown-item>
                <el-dropdown-item command="share" icon="Share">分享</el-dropdown-item>
                <el-dropdown-item command="delete" icon="Delete" divided>删除</el-dropdown-item>
              </template>
              <template v-else>
                <el-dropdown-item command="rename" icon="Edit">重命名</el-dropdown-item>
              </template>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>
    </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatSize, formatDate } from '@/utils'
import FileIcon from '@/components/FileIcon/index.vue'
import type { FileItem, FileListResponse, FolderItem } from '@/types'

const props = defineProps<{
  fileListData: FileListResponse
  getThumbnailUrl: (fileId: string) => string
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
  'download-file': [file: FileItem]
  'rename-file': [file: FileItem]
  'share-file': [file: FileItem]
  'delete-file': [file: FileItem]
  'rename-dir': [folder: FolderItem]
}>()

const handleAction = (command: string, row: FileItem | (FolderItem & { isFolder: boolean })) => {
  if ('isFolder' in row && row.isFolder) {
    if (command === 'rename') {
      emit('rename-dir', row as FolderItem)
    }
  } else {
    switch (command) {
      case 'download':
        emit('download-file', row as FileItem)
        break
      case 'rename':
        emit('rename-file', row as FileItem)
        break
      case 'share':
        emit('share-file', row as FileItem)
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

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
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
  display: inline-block;
}

.enc-tag-inline {
  border: none;
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
  font-size: 11px;
  padding: 2px 8px;
  height: 20px;
  margin-left: 8px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
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
</style>


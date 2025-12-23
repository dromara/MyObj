<template>
  <div class="file-grid">
    <!-- 文件夹 -->
    <div
      v-for="folder in folders"
      :key="'folder-' + folder.id"
      class="file-card scale-up"
      :class="{ selected: isSelectedFolder(folder.id) }"
      @click="$emit('toggle-folder', folder.id)"
      @dblclick="$emit('enter-folder', folder)"
    >
      <div class="file-card-actions" @click.stop>
        <el-dropdown trigger="click" @command="(cmd: string) => $emit('folder-action', cmd, folder)">
          <el-button icon="More" circle text />
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="rename" icon="Edit">重命名</el-dropdown-item>
              <el-dropdown-item command="delete" icon="Delete" divided>删除</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
      <div class="file-icon">
        <el-icon :size="64" color="#409EFF">
          <Folder />
        </el-icon>
      </div>
      <div class="file-name">{{ folder.name }}</div>
      <div class="file-info">{{ formatDate(folder.created_time) }}</div>
    </div>

    <!-- 文件 -->
    <div
      v-for="file in files"
      :key="'file-' + file.file_id"
      class="file-card scale-up"
      :class="{ selected: isSelectedFile(file.file_id) }"
      @click="$emit('toggle-file', file.file_id)"
      @dblclick="$emit('preview-file', file)"
    >
      <div class="file-card-actions" @click.stop>
        <el-dropdown trigger="click" @command="(cmd: string) => $emit('file-action', cmd, file)">
          <el-button icon="More" circle text />
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="download" icon="Download">下载</el-dropdown-item>
              <el-dropdown-item command="rename" icon="Edit">重命名</el-dropdown-item>
              <el-dropdown-item command="share" icon="Share">分享</el-dropdown-item>
              <el-dropdown-item command="delete" icon="Delete" divided>删除</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
      <div class="file-icon">
        <FileIcon
          :mime-type="file.mime_type"
          :file-name="file.file_name"
          :thumbnail-url="getThumbnailUrl(file.file_id)"
          :show-thumbnail="file.has_thumbnail"
          :icon-size="56"
          :is-encrypted="file.is_enc"
        />
      </div>
      <div class="file-name">{{ file.file_name }}</div>
      <div class="file-info">
        {{ formatSize(file.file_size) }} · {{ formatDate(file.created_at) }}
        <el-tag v-if="file.is_enc" size="small" type="warning" class="enc-tag">
          <el-icon><Lock /></el-icon>
        </el-tag>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatSize, formatDate } from '@/utils'
import FileIcon from '@/components/FileIcon/index.vue'
import type { FileItem, FolderItem } from '@/types'

defineProps<{
  folders: FolderItem[]
  files: FileItem[]
  isSelectedFolder: (id: number) => boolean
  isSelectedFile: (id: string) => boolean
  getThumbnailUrl: (fileId: string) => string
}>()

defineEmits<{
  'toggle-folder': [id: number]
  'toggle-file': [id: string]
  'enter-folder': [folder: FolderItem]
  'preview-file': [file: FileItem]
  'folder-action': [command: string, folder: FolderItem]
  'file-action': [command: string, file: FileItem]
}>()
</script>

<style scoped>
.file-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 20px;
  padding: 4px;
}

.file-card {
  background: white;
  border-radius: 16px;
  padding: 12px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  border: 2px solid transparent;
  box-shadow: 0 2px 6px rgba(0,0,0,0.02);
  position: relative;
  overflow: hidden;
}

.file-card-actions {
  position: absolute;
  top: 8px;
  right: 8px;
  opacity: 0;
  transition: opacity 0.2s;
  z-index: 2; /* 降低z-index，确保不会覆盖其他元素 */
  pointer-events: auto; /* 确保可以点击 */
}

.file-card:hover .file-card-actions {
  opacity: 1;
}

.file-card-actions .el-button {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(8px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.file-card:hover {
  transform: translateY(-4px) scale(1.02);
  box-shadow: 0 12px 24px -8px rgba(0,0,0,0.08);
  z-index: 1; /* 降低z-index，避免覆盖工具栏 */
}

.file-card.selected {
  border-color: var(--primary-color);
  background: rgba(37, 99, 235, 0.04);
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.1);
}

.file-icon {
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.3s;
}

.file-card:hover .file-icon {
  transform: scale(1.1);
}

.file-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  text-align: center;
  margin-top: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-info {
  font-size: 11px;
  color: var(--text-placeholder);
  text-align: center;
  margin-top: 4px;
}

.enc-tag {
  border: none;
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
  font-size: 11px;
  padding: 2px 6px;
  height: 18px;
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

@media (max-width: 1024px) {
  .file-grid {
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
    gap: 12px;
    padding: 4px;
  }
  
  .file-card {
    padding: 8px;
  }
  
  .file-icon {
    height: 60px;
  }
  
  .file-name {
    font-size: 12px;
    margin-top: 6px;
  }
  
  .file-card-actions {
    opacity: 1;
  }
}

@media (max-width: 480px) {
  .file-grid {
    grid-template-columns: repeat(auto-fill, minmax(80px, 1fr));
    gap: 8px;
  }
  
  .file-card {
    padding: 6px;
  }
  
  .file-icon {
    height: 50px;
  }
  
  .file-name {
    font-size: 11px;
  }
  
  .file-card-actions {
    opacity: 1;
  }
  
  .file-card-actions .el-button {
    width: 28px;
    height: 28px;
    padding: 0;
  }
}
</style>


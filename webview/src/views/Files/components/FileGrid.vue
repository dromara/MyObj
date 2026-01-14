<template>
  <div class="file-grid">
    <!-- 文件夹 -->
    <div
      v-for="folder in folders"
      :key="'folder-' + folder.id"
      class="file-card folder-card scale-up"
      :class="{ selected: isSelectedFolder(folder.id) }"
      @click="handleFolderClick(folder)"
      @dblclick="$emit('enter-folder', folder)"
    >
      <div class="file-card-actions" @click.stop>
        <el-dropdown trigger="click" @command="(cmd: string) => $emit('folder-action', cmd, folder)">
          <el-button icon="More" circle text />
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="rename" icon="Edit">{{ t('files.rename') }}</el-dropdown-item>
              <el-dropdown-item command="delete" icon="Delete" divided>{{ t('files.delete') }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
      <div class="file-icon">
        <el-icon :size="64" class="folder-icon">
          <Folder />
        </el-icon>
      </div>
      <div class="file-name folder-name">{{ folder.name }}</div>
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
        <!-- 移动端：显示预览按钮（如果可预览） -->
        <el-button
          v-if="isMobile && isPreviewableFile(file)"
          icon="View"
          circle
          text
          class="preview-btn"
          @click.stop="$emit('preview-file', file)"
        />
        <el-dropdown trigger="click" @command="(cmd: string) => handleFileAction(cmd, file)">
          <el-button icon="More" circle text />
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-if="isPreviewableFile(file)" command="preview" icon="View">
                {{ t('files.preview') }}
              </el-dropdown-item>
              <el-dropdown-item command="download" icon="Download">{{ t('files.download') }}</el-dropdown-item>
              <el-dropdown-item command="rename" icon="Edit">{{ t('files.rename') }}</el-dropdown-item>
              <el-dropdown-item command="share" icon="Share">{{ t('files.share') }}</el-dropdown-item>
              <el-dropdown-item
                v-if="!file.is_enc"
                :command="file.public ? 'setPrivate' : 'setPublic'"
                :icon="file.public ? 'Lock' : 'Unlock'"
              >
                {{ file.public ? t('files.cancelPublic') : t('files.setPublic') }}
              </el-dropdown-item>
              <el-dropdown-item command="delete" icon="Delete" divided>{{ t('files.delete') }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
      <div class="file-icon">
        <file-icon
          :mime-type="file.mime_type"
          :file-name="file.file_name"
          :thumbnail-url="getThumbnailUrl(file.file_id)"
          :show-thumbnail="file.has_thumbnail"
          :icon-size="56"
          :is-encrypted="file.is_enc"
        />
      </div>
      <file-name-tooltip :file-name="file.file_name" view-mode="grid" tag="div" custom-class="file-name" />
      <div class="file-info">
        <div class="file-info-text">{{ formatSize(file.file_size) }} · {{ formatDate(file.created_at) }}</div>
        <div class="file-tags">
          <el-tag v-if="file.is_enc" size="small" type="warning" class="enc-tag">
            <el-icon><Lock /></el-icon>
          </el-tag>
          <el-tag v-if="file.public" size="small" type="success" class="public-tag" effect="plain">
            <el-icon><Share /></el-icon>
            {{ t('share.public') }}
          </el-tag>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { formatSize, formatDate } from '@/utils'
  import { useResponsive } from '@/composables/ui/useResponsive'
  import { isPreviewable } from '@/utils/ui/preview'
  import type { FileItem, FolderItem } from '@/types'
  import { useI18n } from '@/composables/core/useI18n'

  const { t } = useI18n()

  defineProps<{
    folders: FolderItem[]
    files: FileItem[]
    isSelectedFolder: (id: number) => boolean
    isSelectedFile: (id: string) => boolean
    getThumbnailUrl: (fileId: string) => string
  }>()

  const emit = defineEmits<{
    'toggle-folder': [id: number]
    'toggle-file': [id: string]
    'enter-folder': [folder: FolderItem]
    'preview-file': [file: FileItem]
    'folder-action': [command: string, folder: FolderItem]
    'file-action': [command: string, file: FileItem]
  }>()

  // 使用响应式检测 composable
  const { isMobile } = useResponsive()

  // 判断文件是否可预览
  const isPreviewableFile = (file: FileItem): boolean => {
    return isPreviewable(file)
  }

  // 处理文件夹点击（单击进入，双击也进入）
  const handleFolderClick = (folder: FolderItem) => {
    emit('enter-folder', folder)
  }

  // 处理文件操作
  const handleFileAction = (command: string, file: FileItem) => {
    if (command === 'preview') {
      emit('preview-file', file)
    } else {
      emit('file-action', command, file)
    }
  }
</script>

<style scoped>
  .file-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 20px;
    padding: 4px;
  }

  .file-card {
    background: var(--card-bg);
    border-radius: 16px;
    padding: 12px;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border: 2px solid transparent;
    /* 增强阴影层次 */
    box-shadow:
      0 1px 3px rgba(0, 0, 0, 0.08),
      0 4px 12px rgba(0, 0, 0, 0.04);
    position: relative;
    overflow: hidden;
    /* 添加渐变边框效果 */
    background-image:
      linear-gradient(var(--card-bg), var(--card-bg)),
      linear-gradient(135deg, rgba(37, 99, 235, 0.05), rgba(79, 70, 229, 0.05));
    background-origin: border-box;
    background-clip: padding-box, border-box;
  }

  html.dark .file-card {
    box-shadow:
      0 1px 3px rgba(0, 0, 0, 0.3),
      0 4px 12px rgba(0, 0, 0, 0.2);
    background-image:
      linear-gradient(var(--card-bg), var(--card-bg)),
      linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(99, 102, 241, 0.1));
  }

  .file-card-actions {
    position: absolute;
    top: 8px;
    right: 8px;
    display: flex;
    align-items: center;
    gap: 4px;
    opacity: 0;
    transition: opacity 0.2s;
    z-index: 2;
    pointer-events: auto;
  }

  .file-card:hover .file-card-actions {
    opacity: 1;
  }

  .preview-btn {
    background: rgba(255, 255, 255, 0.9);
    backdrop-filter: blur(8px);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }

  .file-card-actions .el-button {
    background: rgba(255, 255, 255, 0.9);
    backdrop-filter: blur(8px);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }

  html.dark .preview-btn {
    background: rgba(30, 41, 59, 0.9);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  }

  html.dark .file-card-actions .el-button {
    background: rgba(30, 41, 59, 0.9);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  }

  .file-card:hover {
    transform: translateY(-6px) scale(1.03);
    background: var(--card-hover-bg);
    /* 增强 hover 阴影效果 */
    box-shadow:
      0 8px 24px rgba(37, 99, 235, 0.12),
      0 4px 12px rgba(0, 0, 0, 0.08),
      0 2px 4px rgba(0, 0, 0, 0.04);
    border-color: rgba(37, 99, 235, 0.3);
    z-index: 1; /* 降低z-index，避免覆盖工具栏 */
    /* hover 时的背景渐变 */
    background-image:
      linear-gradient(var(--card-hover-bg), var(--card-hover-bg)),
      linear-gradient(135deg, rgba(37, 99, 235, 0.15), rgba(79, 70, 229, 0.15));
  }

  html.dark .file-card:hover {
    box-shadow:
      0 8px 24px rgba(59, 130, 246, 0.2),
      0 4px 12px rgba(0, 0, 0, 0.3),
      0 2px 4px rgba(0, 0, 0, 0.2);
    border-color: rgba(59, 130, 246, 0.4);
    background-image:
      linear-gradient(var(--card-hover-bg), var(--card-hover-bg)),
      linear-gradient(135deg, rgba(59, 130, 246, 0.2), rgba(99, 102, 241, 0.2));
  }

  /* 文件夹卡片特殊样式 */
  .folder-card:hover {
    background: rgba(64, 158, 255, 0.08);
    border-color: rgba(64, 158, 255, 0.3);
  }

  html.dark .folder-card:hover {
    background: rgba(59, 130, 246, 0.15);
    border-color: rgba(59, 130, 246, 0.4);
  }

  .folder-card:hover .file-icon {
    transform: scale(1.15);
  }

  .file-card.selected {
    border-color: var(--primary-color);
    border-width: 2px;
    background: linear-gradient(rgba(37, 99, 235, 0.06), rgba(37, 99, 235, 0.03));
    /* 选中状态的阴影和光晕效果 */
    box-shadow:
      0 0 0 4px rgba(37, 99, 235, 0.12),
      0 8px 24px rgba(37, 99, 235, 0.15),
      0 4px 12px rgba(0, 0, 0, 0.08);
    /* 选中动画 */
    animation: selectedPulse 0.3s ease-out;
  }

  html.dark .file-card.selected {
    background: linear-gradient(rgba(59, 130, 246, 0.12), rgba(59, 130, 246, 0.06));
    box-shadow:
      0 0 0 4px rgba(59, 130, 246, 0.2),
      0 8px 24px rgba(59, 130, 246, 0.25),
      0 4px 12px rgba(0, 0, 0, 0.3);
  }

  /* 选中状态的光晕动画 */
  .file-card.selected::after {
    content: '';
    position: absolute;
    inset: -4px;
    border-radius: 16px;
    background: radial-gradient(circle at center, rgba(37, 99, 235, 0.2) 0%, transparent 70%);
    pointer-events: none;
    z-index: -1;
    animation: glow 2s ease-in-out infinite;
  }

  html.dark .file-card.selected::after {
    background: radial-gradient(circle at center, rgba(59, 130, 246, 0.3) 0%, transparent 70%);
  }

  @keyframes selectedPulse {
    0% {
      transform: scale(1);
      box-shadow:
        0 0 0 2px rgba(37, 99, 235, 0.1),
        0 4px 12px rgba(37, 99, 235, 0.1);
    }
    50% {
      transform: scale(1.02);
      box-shadow:
        0 0 0 4px rgba(37, 99, 235, 0.15),
        0 8px 24px rgba(37, 99, 235, 0.2);
    }
    100% {
      transform: scale(1);
      box-shadow:
        0 0 0 4px rgba(37, 99, 235, 0.12),
        0 8px 24px rgba(37, 99, 235, 0.15);
    }
  }

  html.dark .file-card.selected {
    animation: selectedPulseDark 0.3s ease-out;
  }

  @keyframes selectedPulseDark {
    0% {
      transform: scale(1);
      box-shadow:
        0 0 0 2px rgba(59, 130, 246, 0.2),
        0 4px 12px rgba(59, 130, 246, 0.15);
    }
    50% {
      transform: scale(1.02);
      box-shadow:
        0 0 0 4px rgba(59, 130, 246, 0.25),
        0 8px 24px rgba(59, 130, 246, 0.3);
    }
    100% {
      transform: scale(1);
      box-shadow:
        0 0 0 4px rgba(59, 130, 246, 0.2),
        0 8px 24px rgba(59, 130, 246, 0.25);
    }
  }

  @keyframes glow {
    0%,
    100% {
      opacity: 0.5;
    }
    50% {
      opacity: 1;
    }
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

  .file-name.folder-name {
    color: var(--el-color-primary) !important;
    font-weight: 600;
  }

  .file-info {
    font-size: 11px;
    color: var(--text-placeholder);
    text-align: center;
    margin-top: 4px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
  }

  .file-info-text {
    line-height: 1.4;
  }

  .file-tags {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    justify-content: center;
  }

  .folder-icon {
    color: var(--el-color-primary);
  }

  .enc-tag {
    border: none;
    background: linear-gradient(135deg, var(--warning-color) 0%, var(--warning-color) 100%);
    color: white;
    font-size: 11px;
    padding: 2px 6px;
    height: 18px;
    display: inline-flex;
    align-items: center;
    gap: 2px;
  }

  .public-tag {
    border: 1px solid var(--success-color);
    background: rgba(16, 185, 129, 0.08);
    color: var(--success-color);
    font-size: 11px;
    padding: 2px 6px;
    height: 18px;
    display: inline-flex;
    align-items: center;
    gap: 3px;
    border-radius: 4px;
    font-weight: 500;
    transition: all 0.2s;
    white-space: nowrap;
    line-height: 1;
  }

  .public-tag:hover {
    background: rgba(16, 185, 129, 0.12);
    border-color: rgba(16, 185, 129, 0.4);
  }

  .public-tag .el-icon {
    color: var(--success-color);
    flex-shrink: 0;
  }

  html.dark .public-tag {
    background: rgba(16, 185, 129, 0.15);
    border-color: rgba(16, 185, 129, 0.4);
  }

  .public-tag :deep(.el-tag__content) {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    white-space: nowrap;
    line-height: 1;
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

    .preview-btn {
      opacity: 1 !important;
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

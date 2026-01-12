<template>
  <div class="file-icon-wrapper" :class="{ 'has-thumbnail': showThumbnail && !isEncrypted }">
    <!-- 缩略图（加密文件不显示缩略图） -->
    <img 
      v-if="showThumbnail && thumbnailUrl && !isEncrypted" 
      :src="thumbnailUrl" 
      class="thumbnail-image"
      @error="handleImageError"
    />
    
    <!-- 文件类型图标 -->
    <div 
      v-else 
      class="file-icon-card" 
      :style="{ 
        background: gradientBg,
        borderImage: `linear-gradient(135deg, ${iconConfig.color}40, ${iconConfig.color}20) 1`
      }"
    >
      <div class="icon-glow"></div>
      <el-icon :size="iconSize" :color="iconConfig.color" class="file-icon-element">
        <component :is="iconConfig.icon" />
      </el-icon>
      
      <!-- 文件类型标签 -->
      <div class="file-type-badge" v-if="showBadge">
        {{ fileExtension }}
      </div>
    </div>
    
    <!-- 加密标识 -->
    <div v-if="isEncrypted" class="encryption-badge" :title="t('files.encrypted')">
      <el-icon :size="14">
        <Lock />
      </el-icon>
    </div>
  </div>
</template>

<script setup lang="ts">
import { getFileIcon } from '@/utils/fileIcon'
import type { FileIconConfig } from '@/utils/fileIcon'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

interface Props {
  mimeType: string
  fileName?: string
  thumbnailUrl?: string
  showThumbnail?: boolean
  iconSize?: number
  showBadge?: boolean
  isEncrypted?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  iconSize: 48,
  showThumbnail: false,
  showBadge: true,
  isEncrypted: false
})

const imageError = ref(false)

// 获取图标配置
const iconConfig = computed<FileIconConfig>(() => {
  return getFileIcon(props.mimeType)
})

// 获取文件扩展名
const fileExtension = computed(() => {
  if (!props.fileName) return ''
  const parts = props.fileName.split('.')
  if (parts.length > 1) {
    const ext = parts[parts.length - 1].toUpperCase()
    return ext.length > 4 ? ext.substring(0, 4) : ext
  }
  return ''
})

// 增强的渐变背景
const gradientBg = computed(() => {
  const color = iconConfig.value.color
  // 更丰富的渐变效果，从浅到深
  return `linear-gradient(135deg, ${color}20 0%, ${color}10 50%, ${color}30 100%)`
})

// 边框渐变
const borderGradient = computed(() => {
  const color = iconConfig.value.color
  return `linear-gradient(135deg, ${color}40, ${color}20)`
})

const handleImageError = () => {
  imageError.value = true
}
</script>

<style scoped>
.file-icon-wrapper {
  position: relative;
  display: inline-block;
  width: 100%;
  height: 100%;
}

.file-icon-card {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  border: 2px solid transparent;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  /* 添加微妙的阴影 */
  box-shadow: 
    inset 0 1px 2px rgba(255, 255, 255, 0.5),
    0 1px 3px rgba(0, 0, 0, 0.1);
}

/* 图标光晕效果 */
.icon-glow {
  position: absolute;
  inset: -20%;
  background: radial-gradient(
    circle at center,
    currentColor 0%,
    transparent 70%
  );
  opacity: 0;
  transition: opacity 0.3s;
  pointer-events: none;
  z-index: 0;
}

.file-icon-card:hover .icon-glow {
  opacity: 0.15;
}

.file-icon-element {
  position: relative;
  z-index: 1;
  transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.file-icon-card:hover .file-icon-element {
  transform: scale(1.1);
}

.file-icon-card:hover {
  /* Parent card handles hover interaction */
  transform: none;
  box-shadow: none;
}

.file-type-badge {
  position: absolute;
  bottom: 4px;
  right: 4px;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: bold;
  line-height: 1;
  backdrop-filter: blur(4px);
}

html.dark .file-type-badge {
  background: rgba(255, 255, 255, 0.2);
  color: var(--el-text-color-primary);
}

.thumbnail-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 12px;
  border: 2px solid var(--el-border-color-lighter);
  transition: all 0.3s ease;
}

.thumbnail-image:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  border-color: var(--el-color-primary);
}

html.dark .thumbnail-image:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
}

.has-thumbnail {
  background: transparent;
}

.encryption-badge {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 24px;
  height: 24px;
  background: linear-gradient(135deg, var(--warning-color) 0%, var(--warning-color) 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--el-text-color-primary);
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.4);
  z-index: 10;
  animation: pulse 2s ease-in-out infinite;
}

html.dark .encryption-badge {
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.6);
}

@keyframes pulse {
  0%, 100% {
    box-shadow: 0 2px 8px rgba(245, 158, 11, 0.4);
  }
  50% {
    box-shadow: 0 2px 12px rgba(245, 158, 11, 0.6);
  }
}

/* 移动端响应式 */
@media (max-width: 1024px) {
  .file-type-badge {
    font-size: 9px;
    padding: 1px 4px;
    bottom: 2px;
    right: 2px;
  }
  
  .encryption-badge {
    width: 20px;
    height: 20px;
    top: 2px;
    right: 2px;
  }
  
  .encryption-badge .el-icon {
    font-size: 12px;
  }
}

@media (max-width: 480px) {
  .file-type-badge {
    font-size: 8px;
    padding: 1px 3px;
  }
  
  .encryption-badge {
    width: 18px;
    height: 18px;
  }
  
  .encryption-badge .el-icon {
    font-size: 10px;
  }
}
</style>

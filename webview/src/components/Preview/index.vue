<template>
  <el-dialog
    v-model="visible"
    :title="currentFile?.file_name || '文件预览'"
    width="90%"
    :close-on-click-modal="false"
    :close-on-press-escape="true"
    @close="handleClose"
    class="file-preview-dialog"
  >
    <!-- 加载状态 -->
    <div v-if="loading" class="preview-loading">
      <el-icon class="is-loading" :size="48"><Loading /></el-icon>
      <p>加载中...</p>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="preview-error">
      <el-icon :size="48" color="#f56c6c"><WarningFilled /></el-icon>
      <p>{{ error }}</p>
      <el-button type="primary" @click="handleRetry">重试</el-button>
    </div>

    <!-- 图片预览 -->
    <div v-else-if="previewType === 'image'" class="preview-image-container">
      <div class="image-wrapper">
        <img
          :src="imageUrl"
          :style="imageStyle"
          class="preview-image"
          :alt="currentFile?.file_name"
          @load="handleImageLoad"
          @error="handleImageError"
        />
      </div>
      <!-- 图片工具栏 -->
      <div class="preview-toolbar">
        <el-button-group>
          <el-button :icon="ZoomIn" @click="zoomImage(0.1)">放大</el-button>
          <el-button :icon="ZoomOut" @click="zoomImage(-0.1)">缩小</el-button>
          <el-button :icon="RefreshRight" @click="rotateImage(90)">旋转</el-button>
          <el-button :icon="Refresh" @click="resetImageZoom">重置</el-button>
          <el-button :icon="Download" @click="handleDownload">下载</el-button>
        </el-button-group>
      </div>
    </div>

    <!-- 视频预览 -->
    <div v-else-if="previewType === 'video'" class="preview-video-container">
      <video
        :src="videoUrl"
        :autoplay="options.autoplay"
        :loop="options.loop"
        :controls="options.controls"
        class="preview-video"
        @loadstart="handleVideoLoad"
        @error="handleVideoError"
      >
        您的浏览器不支持视频播放
      </video>
      <div class="preview-toolbar">
        <el-button :icon="Download" @click="handleDownload">下载</el-button>
      </div>
    </div>

    <!-- 音频预览 -->
    <div v-else-if="previewType === 'audio'" class="preview-audio-container">
      <div class="audio-wrapper">
        <el-icon :size="64" color="var(--primary-color)"><Headset /></el-icon>
        <p class="audio-filename">{{ currentFile?.file_name }}</p>
        <audio
          :src="audioUrl"
          :autoplay="options.autoplay"
          :loop="options.loop"
          :controls="options.controls"
          class="preview-audio"
          @loadstart="handleAudioLoad"
          @error="handleAudioError"
        >
          您的浏览器不支持音频播放
        </audio>
      </div>
      <div class="preview-toolbar">
        <el-button :icon="Download" @click="handleDownload">下载</el-button>
      </div>
    </div>

    <!-- PDF 预览 -->
    <div v-else-if="previewType === 'pdf'" class="preview-pdf-container">
      <el-alert
        title="PDF 预览提示"
        description="如果 PDF 无法正常显示，请点击下载按钮下载后查看"
        type="info"
        :closable="false"
        class="mb-4"
      />
      <iframe
        :src="pdfUrl"
        class="preview-pdf"
        @load="handlePdfLoad"
        @error="handlePdfError"
      ></iframe>
      <div class="preview-toolbar">
        <el-button :icon="Download" @click="handleDownload">下载</el-button>
      </div>
    </div>

    <!-- 文本/代码预览 -->
    <div v-else-if="previewType === 'text' || previewType === 'code'" class="preview-text-container">
      <div class="preview-text-header">
        <span class="text-type-label">
          {{ previewType === 'code' ? '代码预览' : '文本预览' }}
        </span>
        <el-button :icon="Download" size="small" @click="handleDownload">下载</el-button>
      </div>
      <pre
        :class="['preview-text-content', previewType === 'code' ? `language-${codeLanguage}` : '']"
      ><code>{{ textContent }}</code></pre>
    </div>

    <!-- 不支持预览 -->
    <div v-else class="preview-unsupported">
      <el-icon :size="64" color="#909399"><Document /></el-icon>
      <p class="unsupported-title">不支持预览此文件类型</p>
      <p class="unsupported-desc">
        文件类型: {{ currentFile?.mime_type || '未知' }}
      </p>
      <el-button type="primary" :icon="Download" @click="handleDownload">下载文件</el-button>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import type { FileItem } from '@/types'
import type { PreviewType, PreviewOptions } from '@/types/preview'
import { detectFileType, getFilePreviewUrl, getFileDownloadUrl, getFileTextContent, getCodeLanguage } from '@/utils/preview'
import { download } from '@/utils/request'
import { API_BASE_URL, API_ENDPOINTS } from '@/config/api'
import { ZoomIn, ZoomOut, RefreshRight, Refresh, Download, Loading, WarningFilled, Headset, Document } from '@element-plus/icons-vue'

interface Props {
  modelValue: boolean
  file: FileItem | null
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  file: null
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const { proxy } = getCurrentInstance() as ComponentInternalInstance

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const currentFile = computed(() => props.file)
const previewType = computed<PreviewType>(() => {
  if (!currentFile.value) return 'unsupported'
  return detectFileType(currentFile.value)
})

const loading = ref(false)
const error = ref<string | undefined>(undefined)
const options = ref<PreviewOptions>({
  autoplay: false,
  loop: false,
  controls: true,
  zoom: 1,
  rotate: 0
})

const imageUrl = ref('')
const videoUrl = ref('')
const audioUrl = ref('')
const pdfUrl = ref('')
const textContent = ref('')
const codeLanguage = ref('')

// 图片样式
const imageStyle = computed(() => {
  const zoom = options.value.zoom || 1
  const rotate = options.value.rotate || 0
  return {
    transform: `scale(${zoom}) rotate(${rotate}deg)`,
    transformOrigin: 'center center',
    transition: 'transform 0.3s ease'
  }
})

// 加载文件内容
const loadFileContent = async () => {
  if (!currentFile.value) return

  loading.value = true
  error.value = undefined

  try {
    const file = currentFile.value
    const fileId = file.file_id

    switch (previewType.value) {
      case 'image':
        // 优先使用缩略图，如果没有则使用预览URL
        if (file.has_thumbnail) {
          imageUrl.value = `${API_BASE_URL}${API_ENDPOINTS.FILE.THUMBNAIL}/${fileId}`
        } else {
          imageUrl.value = getFilePreviewUrl(fileId)
        }
        break
      case 'video':
        videoUrl.value = getFilePreviewUrl(fileId)
        break
      case 'audio':
        audioUrl.value = getFilePreviewUrl(fileId)
        break
      case 'pdf':
        pdfUrl.value = getFilePreviewUrl(fileId)
        break
      case 'text':
      case 'code':
        textContent.value = await getFileTextContent(fileId)
        if (previewType.value === 'code') {
          codeLanguage.value = getCodeLanguage(file.file_name)
        }
        break
    }

    loading.value = false
  } catch (err: any) {
    loading.value = false
    error.value = err?.message || '加载文件失败'
    proxy?.$log.error('加载文件内容失败', err)
  }
}

// 缩放图片
const zoomImage = (delta: number) => {
  if (previewType.value !== 'image') return
  const currentZoom = options.value.zoom || 1
  const newZoom = Math.max(0.1, Math.min(5, currentZoom + delta))
  options.value.zoom = newZoom
}

// 重置图片
const resetImageZoom = () => {
  if (previewType.value !== 'image') return
  options.value.zoom = 1
  options.value.rotate = 0
}

// 旋转图片
const rotateImage = (angle: number) => {
  if (previewType.value !== 'image') return
  const currentRotate = options.value.rotate || 0
  options.value.rotate = (currentRotate + angle) % 360
}

// 下载文件
const handleDownload = () => {
  if (!currentFile.value) return
  const fileId = currentFile.value.file_id
  const fileName = currentFile.value.file_name
  const url = getFileDownloadUrl(fileId)
  
  download(url, fileName)
}

// 重试
const handleRetry = () => {
  loadFileContent()
}

// 关闭预览
const handleClose = () => {
  visible.value = false
  // 清理资源
  imageUrl.value = ''
  videoUrl.value = ''
  audioUrl.value = ''
  pdfUrl.value = ''
  textContent.value = ''
  error.value = undefined
  // 重置选项
  options.value = {
    autoplay: false,
    loop: false,
    controls: true,
    zoom: 1,
    rotate: 0
  }
}

// 图片加载完成
const handleImageLoad = () => {
  loading.value = false
}

// 图片加载错误
const handleImageError = () => {
  loading.value = false
  error.value = '图片加载失败'
}

// 视频加载
const handleVideoLoad = () => {
  loading.value = false
}

// 视频加载错误
const handleVideoError = () => {
  loading.value = false
  error.value = '视频加载失败'
}

// 音频加载
const handleAudioLoad = () => {
  loading.value = false
}

// 音频加载错误
const handleAudioError = () => {
  loading.value = false
  error.value = '音频加载失败'
}

// PDF 加载
const handlePdfLoad = () => {
  loading.value = false
}

// PDF 加载错误
const handlePdfError = () => {
  loading.value = false
  error.value = 'PDF 加载失败'
}

// 监听文件变化
watch(() => currentFile.value, async (newFile) => {
  if (newFile && visible.value) {
    await loadFileContent()
  }
}, { immediate: true })

// 监听可见性变化
watch(visible, async (newVisible) => {
  if (newVisible && currentFile.value) {
    await loadFileContent()
  } else {
    handleClose()
  }
})
</script>

<style scoped>
.file-preview-dialog :deep(.el-dialog__body) {
  padding: 20px;
  max-height: 80vh;
  overflow-y: auto;
}

.preview-loading,
.preview-error,
.preview-unsupported {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  gap: 16px;
}

.preview-loading p,
.preview-error p {
  margin: 0;
  color: var(--text-secondary);
}

.preview-image-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.image-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  background: var(--bg-color);
  border-radius: 8px;
  overflow: auto;
  padding: 20px;
}

.preview-image {
  max-width: 100%;
  max-height: 70vh;
  object-fit: contain;
}

.preview-video-container,
.preview-audio-container,
.preview-pdf-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.preview-video {
  width: 100%;
  max-height: 70vh;
  background: #000;
  border-radius: 8px;
}

.audio-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 40px;
}

.audio-filename {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.preview-audio {
  width: 100%;
  max-width: 500px;
}

.preview-pdf {
  width: 100%;
  height: 70vh;
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.preview-text-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.preview-text-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.text-type-label {
  font-size: 14px;
  color: var(--text-secondary);
  font-weight: 500;
}

.preview-text-content {
  background: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  max-height: 60vh;
  overflow: auto;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', 'source-code-pro', monospace;
  font-size: 14px;
  line-height: 1.6;
  margin: 0;
}

.preview-text-content code {
  color: var(--text-primary);
  background: transparent;
  padding: 0;
}

.preview-toolbar {
  display: flex;
  justify-content: center;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
}

.unsupported-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.unsupported-desc {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0;
}
</style>


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
          <el-button icon="ZoomIn" @click="zoomImage(0.1)">放大</el-button>
          <el-button icon="ZoomOut" @click="zoomImage(-0.1)">缩小</el-button>
          <el-button icon="RefreshRight" @click="rotateImage(90)">旋转</el-button>
          <el-button icon="Refresh" @click="resetImageZoom">重置</el-button>
          <el-button icon="Printer" @click="handlePrint">打印</el-button>
          <el-button icon="Download" @click="handleDownload">下载</el-button>
        </el-button-group>
      </div>
    </div>

    <!-- 视频预览 -->
    <div v-else-if="previewType === 'video'" class="preview-video-container">
      <plyr-player
        v-if="videoUrl"
        :src="videoUrl"
        :autoplay="options.autoplay"
        :loop="options.loop"
        class="preview-video-plyr"
        @ready="handleVideoReady"
        @error="handleVideoError"
      />
      <div class="preview-toolbar">
        <el-button icon="Printer" @click="handlePrint">打印</el-button>
        <el-button icon="Download" @click="handleDownload">下载</el-button>
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
        <el-button icon="Download" @click="handleDownload">下载</el-button>
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
        <el-button icon="Printer" @click="handlePrint">打印</el-button>
        <el-button icon="Download" @click="handleDownload">下载</el-button>
      </div>
    </div>

    <!-- 文本/代码预览 -->
    <div v-else-if="previewType === 'text' || previewType === 'code'" class="preview-text-container">
      <div class="preview-text-header">
        <span class="text-type-label">
          {{ previewType === 'code' ? '代码预览' : '文本预览' }}
        </span>
        <el-button-group>
          <el-button icon="Printer" size="small" @click="handlePrint">打印</el-button>
          <el-button icon="Download" size="small" @click="handleDownload">下载</el-button>
        </el-button-group>
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
      <div class="unsupported-actions">
        <el-button type="primary" icon="Printer" @click="handlePrint">打印</el-button>
        <el-button icon="Download" @click="handleDownload">下载文件</el-button>
      </div>
    </div>
    
    <!-- 下载密码对话框 -->
    <el-dialog
      v-model="showDownloadPasswordDialog"
      title="输入文件密码"
      width="400px"
      :close-on-click-modal="false"
    >
      <div class="download-password-form">
        <el-text>{{ downloadPasswordForm.file_name }}</el-text>
        <el-form-item label="文件密码" style="margin-top: 16px;">
          <el-input
            v-model="downloadPasswordForm.file_password"
            type="password"
            placeholder="请输入文件密码"
            show-password
            @keyup.enter="confirmDownloadPassword"
          />
        </el-form-item>
      </div>
      <template #footer>
        <el-button @click="showDownloadPasswordDialog = false">取消</el-button>
        <el-button type="primary" :loading="downloadingFile" @click="confirmDownloadPassword">确定</el-button>
      </template>
    </el-dialog>
  </el-dialog>
</template>

<script setup lang="ts">
import type { FileItem } from '@/types'
import type { PreviewType, PreviewOptions } from '@/types/preview'
import { detectFileType, getFilePreviewUrl, getFileTextContent, getCodeLanguage } from '@/utils/preview'
import { useFileDownload } from '@/composables/useFileDownload'
import { API_BASE_URL, API_ENDPOINTS } from '@/config/api'
import { createVideoPlayPrecheck, getVideoStreamUrl } from '@/api/video'
import { printImage, printPDF, printText, printOfficeDocument, isPrintableType, isOfficeDocument } from '@/utils/print'

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

// 使用统一的文件下载 composable（Preview 组件不需要跳转到任务中心）
const {
  showDownloadPasswordDialog,
  downloadPasswordForm,
  downloadingFile,
  handleDownload: handleFileDownload,
  confirmDownloadPassword
} = useFileDownload()

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
        // 优先使用缩略图，如果没有则使用预览URL（blob URL）
        if (file.has_thumbnail) {
          // 缩略图也需要通过fetch获取（带认证），然后创建blob URL
          try {
            const token = proxy?.$cache.local.get('token')
            const thumbnailUrl = `${API_BASE_URL}${API_ENDPOINTS.FILE.THUMBNAIL}/${fileId}`
            const response = await fetch(thumbnailUrl, {
              headers: {
                'Authorization': token ? `Bearer ${token}` : ''
              }
            })
            if (response.ok) {
              const blob = await response.blob()
              imageUrl.value = window.URL.createObjectURL(blob)
            } else {
              // 缩略图获取失败，使用预览URL
              imageUrl.value = await getFilePreviewUrl(fileId)
            }
          } catch (err) {
            // 缩略图获取失败，使用预览URL
            imageUrl.value = await getFilePreviewUrl(fileId)
          }
        } else {
          imageUrl.value = await getFilePreviewUrl(fileId)
        }
        break
      case 'video':
        // 视频使用 Plyr 播放器（支持 Range 请求）
        try {
          const res = await createVideoPlayPrecheck(fileId)
          if (res.code === 200 && res.data) {
            // 获取 JWT token 并添加到 URL 参数中
            const jwtToken = proxy?.$cache.local.get('token')
            // 构建视频流 URL（包含 playToken 和 JWT token）
            videoUrl.value = getVideoStreamUrl(res.data.play_token, jwtToken || undefined)
          } else {
            throw new Error(res.message || '获取视频播放 Token 失败')
          }
        } catch (err) {
          error.value = err instanceof Error ? err.message : '加载视频失败'
        }
        break
      case 'audio':
        audioUrl.value = await getFilePreviewUrl(fileId)
        break
      case 'pdf':
        pdfUrl.value = await getFilePreviewUrl(fileId)
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
const handleDownload = async () => {
  if (!currentFile.value) return
  await handleFileDownload(currentFile.value)
}

// 打印文件
const handlePrint = async () => {
  if (!currentFile.value) return
  
  try {
    const file = currentFile.value
    const mimeType = file.mime_type || ''
    
    // 检查是否为Office文档（Excel、Word、PowerPoint等）
    if (isOfficeDocument(mimeType)) {
      // Office文档：获取文件URL并尝试打印
      try {
        const fileUrl = await getFilePreviewUrl(file.file_id)
        await printOfficeDocument(fileUrl, file.file_name, {
          title: file.file_name
        })
      } catch (error: any) {
        proxy?.$log.error('打印Office文档失败', error)
        // 如果打印失败，提示用户下载后打印
        proxy?.$modal.msgWarning('无法直接打印此文档类型，请先下载文件，然后用相应的Office软件（如Excel、Word）打开并打印。')
      }
      return
    }
    
    // 检查是否支持打印
    if (!isPrintableType(mimeType)) {
      proxy?.$modal.msgWarning('该文件类型不支持打印')
      return
    }
    
    switch (previewType.value) {
      case 'image':
        // 打印时始终使用原图，不使用缩略图
        try {
          const originalImageUrl = await getFilePreviewUrl(file.file_id)
          await printImage(originalImageUrl, {
            title: file.file_name
          })
        } catch (error: any) {
          proxy?.$log.error('获取原图失败', error)
          proxy?.$modal.msgError('获取原图失败，无法打印')
        }
        break
        
      case 'pdf':
        if (pdfUrl.value) {
          await printPDF(pdfUrl.value, {
            title: file.file_name
          })
        } else {
          proxy?.$modal.msgWarning('PDF未加载完成，请稍候再试')
        }
        break
        
      case 'text':
      case 'code':
        if (textContent.value) {
          await printText(textContent.value, file.file_name, {
            title: file.file_name
          })
        } else {
          proxy?.$modal.msgWarning('文本内容未加载完成，请稍候再试')
        }
        break
        
      default:
        // 对于其他不支持预览的文件，尝试直接打印文件URL
        try {
          const fileUrl = await getFilePreviewUrl(file.file_id)
          await printOfficeDocument(fileUrl, file.file_name, {
            title: file.file_name
          })
        } catch (error: any) {
          proxy?.$log.error('打印失败', error)
          proxy?.$modal.msgWarning('无法直接打印此文件类型，请先下载文件，然后用相应的软件打开并打印。')
        }
    }
  } catch (error: any) {
    proxy?.$log.error('打印失败', error)
    proxy?.$modal.msgError(error.message || '打印失败')
  }
}

// 重试
const handleRetry = () => {
  loadFileContent()
}

// 清理blob URL
const cleanupBlobUrls = () => {
  if (imageUrl.value && imageUrl.value.startsWith('blob:')) {
    window.URL.revokeObjectURL(imageUrl.value)
  }
  if (videoUrl.value && videoUrl.value.startsWith('blob:')) {
    window.URL.revokeObjectURL(videoUrl.value)
  }
  if (audioUrl.value && audioUrl.value.startsWith('blob:')) {
    window.URL.revokeObjectURL(audioUrl.value)
  }
  if (pdfUrl.value && pdfUrl.value.startsWith('blob:')) {
    window.URL.revokeObjectURL(pdfUrl.value)
  }
}

// 关闭预览
const handleClose = () => {
  cleanupBlobUrls()
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

// 视频加载错误（已由 Plyr 内部处理，不需要单独处理）

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

// 视频播放器就绪
const handleVideoReady = () => {
  loading.value = false
}

// 视频播放错误
const handleVideoError = (errorMessage: string) => {
  loading.value = false
  error.value = errorMessage
  proxy?.$log.error('视频播放错误', errorMessage)
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

// 组件卸载时清理
onUnmounted(() => {
  cleanupBlobUrls()
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

.preview-video-plyr {
  width: 100%;
  height: 70vh;
  max-height: 70vh;
  min-height: 400px;
  border-radius: 8px;
  overflow: hidden;
  background: #000;
}

.preview-video-plyr :deep(.plyr) {
  height: 100%;
}

.preview-video-plyr :deep(.plyr__video-wrapper) {
  height: 100%;
}

.preview-video-plyr :deep(video) {
  width: 100%;
  height: 100%;
}

/* 确保预览窗口内容可见 */
.file-preview-dialog :deep(.el-dialog__body) {
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.preview-video-container {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
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

.unsupported-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  margin-top: 16px;
}

/* 移动端响应式 */
@media (max-width: 1024px) {
  .file-preview-dialog :deep(.el-dialog) {
    width: 95% !important;
    margin: 5vh auto;
  }
  
  .file-preview-dialog :deep(.el-dialog__body) {
    padding: 12px;
    max-height: 85vh;
  }
  
  .preview-loading,
  .preview-error,
  .preview-unsupported {
    min-height: 300px;
    gap: 12px;
  }
  
  .image-wrapper {
    min-height: 300px;
    padding: 12px;
  }
  
  .preview-image {
    max-height: 60vh;
  }
  
  .preview-video {
    max-height: 60vh;
  }
  
  .preview-pdf {
    height: 60vh;
  }
  
  .preview-text-content {
    max-height: 50vh;
    padding: 12px;
    font-size: 12px;
  }
  
  .preview-toolbar {
    padding-top: 12px;
  }
  
  .preview-toolbar :deep(.el-button-group) {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  
  .preview-toolbar :deep(.el-button) {
    flex: 1;
    min-width: 0;
    padding: 8px 12px;
  }
  
  .preview-toolbar :deep(.el-button span) {
    display: none;
  }
  
  .audio-wrapper {
    padding: 24px 16px;
    gap: 12px;
  }
  
  .audio-filename {
    font-size: 14px;
  }
  
  .preview-audio {
    max-width: 100%;
  }
}

@media (max-width: 480px) {
  .file-preview-dialog :deep(.el-dialog) {
    width: 100% !important;
    margin: 0;
    height: 100vh;
    border-radius: 0;
  }
  
  .file-preview-dialog :deep(.el-dialog__header) {
    padding: 12px;
  }
  
  .file-preview-dialog :deep(.el-dialog__body) {
    padding: 8px;
    max-height: calc(100vh - 120px);
  }
  
  .preview-image {
    max-height: 50vh;
  }
  
  .preview-video {
    max-height: 50vh;
  }
  
  .preview-pdf {
    height: 50vh;
  }
  
  .preview-text-content {
    max-height: 45vh;
    font-size: 11px;
  }
}
</style>


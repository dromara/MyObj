<template>
  <el-dialog
    v-model="visible"
    :title="currentFile?.file_name || t('preview.title')"
    width="90%"
    :close-on-click-modal="false"
    :close-on-press-escape="true"
    @close="handleClose"
    :class="['file-preview-dialog', `preview-${previewType}-active`]"
  >
    <!-- 加载状态 -->
    <div v-if="loading" class="preview-loading">
      <el-icon class="is-loading" :size="48"><Loading /></el-icon>
      <p>{{ t('preview.loading') }}</p>
    </div>

    <!-- 加密文件提示 -->
    <div v-else-if="currentFile?.is_enc" class="preview-encrypted">
      <el-icon :size="64" class="encrypted-icon"><Lock /></el-icon>
      <p class="encrypted-title">{{ t('preview.encrypted.title') }}</p>
      <p class="encrypted-desc">{{ t('preview.encrypted.desc') }}</p>
      <div class="encrypted-actions">
        <el-button type="primary" icon="Download" @click="handleDownload">{{ t('preview.encrypted.download') }}</el-button>
      </div>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="preview-error">
      <el-icon :size="48" class="error-icon"><WarningFilled /></el-icon>
      <p>{{ error }}</p>
      <el-button type="primary" @click="handleRetry">{{ t('preview.error.retry') }}</el-button>
    </div>

    <!-- 图片预览 -->
    <div v-else-if="previewType === 'image'" class="preview-image-container">
      <div class="image-wrapper" @wheel="handleImageWheel" @dblclick="resetImageZoom">
        <img
          :src="imageUrl"
          :style="imageStyle"
          class="preview-image"
          :alt="currentFile?.file_name"
          @load="handleImageLoad"
          @error="handleImageError"
          @mousedown="handleImageMouseDown"
        />
        <!-- 图片导航提示 -->
        <div v-if="imageZoom > 1" class="image-nav-hint">
          <el-icon><InfoFilled /></el-icon>
          <span>{{ t('preview.image.hint') }}</span>
        </div>
      </div>
      <!-- 图片工具栏 -->
      <div class="preview-toolbar">
        <div class="toolbar-left">
          <el-button-group>
            <el-button icon="ZoomIn" @click="zoomImage(0.1)">{{ t('preview.image.zoomIn') }}</el-button>
            <el-button icon="ZoomOut" @click="zoomImage(-0.1)">{{ t('preview.image.zoomOut') }}</el-button>
            <el-button icon="RefreshRight" @click="rotateImage(90)">{{ t('preview.image.rotate') }}</el-button>
            <el-button icon="Refresh" @click="resetImageZoom">{{ t('preview.image.reset') }}</el-button>
          </el-button-group>
        </div>
        <div class="toolbar-right">
          <el-button-group>
            <el-button v-if="canPrint" icon="Printer" @click="handlePrint">{{ t('preview.image.print') }}</el-button>
            <el-button icon="Download" @click="handleDownload">{{ t('preview.image.download') }}</el-button>
            <el-button @click="toggleFullscreen">
              <el-icon><FullScreen /></el-icon>
              {{ t('preview.image.fullscreen') }}
            </el-button>
          </el-button-group>
        </div>
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
        <el-button v-if="canPrint" icon="Printer" @click="handlePrint">{{ t('preview.video.print') }}</el-button>
        <el-button icon="Download" @click="handleDownload">{{ t('preview.video.download') }}</el-button>
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
          {{ t('preview.audio.notSupported') }}
        </audio>
      </div>
      <div class="preview-toolbar">
        <el-button icon="Download" @click="handleDownload">{{ t('preview.audio.download') }}</el-button>
      </div>
    </div>

    <!-- PDF 预览 -->
    <div v-else-if="previewType === 'pdf'" class="preview-pdf-container">
      <el-alert
        :title="t('preview.pdf.title')"
        :description="t('preview.pdf.description')"
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
        <el-button v-if="canPrint" icon="Printer" @click="handlePrint">{{ t('preview.pdf.print') }}</el-button>
        <el-button icon="Download" @click="handleDownload">{{ t('preview.pdf.download') }}</el-button>
      </div>
    </div>

    <!-- 文本/代码预览 -->
    <div v-else-if="previewType === 'text' || previewType === 'code'" class="preview-text-container">
      <div class="preview-text-header">
        <span class="text-type-label">
          {{ previewType === 'code' ? t('preview.code.title') : t('preview.text.title') }}
        </span>
        <el-button-group>
          <el-button v-if="canPrint" icon="Printer" size="small" @click="handlePrint">{{ t('preview.text.print') }}</el-button>
          <el-button icon="Download" size="small" @click="handleDownload">{{ t('preview.text.download') }}</el-button>
        </el-button-group>
      </div>
      <pre
        :class="['preview-text-content', previewType === 'code' ? `language-${codeLanguage}` : '']"
      ><code>{{ textContent }}</code></pre>
    </div>

    <!-- 不支持预览 -->
    <div v-else class="preview-unsupported">
      <el-icon :size="64" class="unsupported-icon"><Document /></el-icon>
      <p class="unsupported-title">{{ t('preview.notSupported.title') }}</p>
      <p class="unsupported-desc">
        {{ t('preview.notSupported.mimeType') }}: {{ currentFile?.mime_type || t('preview.notSupported.unknown') }}
      </p>
      <div class="unsupported-actions">
        <el-button v-if="canPrint" type="primary" icon="Printer" @click="handlePrint">{{ t('preview.notSupported.print') }}</el-button>
        <el-button icon="Download" @click="handleDownload">{{ t('preview.notSupported.download') }}</el-button>
      </div>
    </div>
    
    <!-- 下载密码对话框 -->
    <el-dialog
      v-model="showDownloadPasswordDialog"
      :title="t('preview.downloadPassword.title')"
      width="400px"
      :close-on-click-modal="false"
    >
      <div class="download-password-form">
        <el-text>{{ downloadPasswordForm.file_name }}</el-text>
        <el-form-item :label="t('preview.downloadPassword.label')" style="margin-top: 16px;">
          <el-input
            v-model="downloadPasswordForm.file_password"
            type="password"
            :placeholder="t('preview.downloadPassword.placeholder')"
            show-password
            @keyup.enter="confirmDownloadPassword"
          />
        </el-form-item>
      </div>
      <template #footer>
        <el-button @click="showDownloadPasswordDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="downloadingFile" @click="confirmDownloadPassword">{{ t('common.confirm') }}</el-button>
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
import { Lock, InfoFilled, FullScreen } from '@element-plus/icons-vue'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

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

// 判断当前文件是否支持打印
const canPrint = computed(() => {
  if (!currentFile.value) return false
  const mimeType = currentFile.value.mime_type || ''
  // isPrintableType 已经包含了图片、PDF、文本、代码、Office文档等所有支持打印的类型
  return isPrintableType(mimeType)
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

// 加载视频内容
const loadVideoContent = async (fileId: string) => {
  try {
    const res = await createVideoPlayPrecheck(fileId)
    if (res.code === 200 && res.data) {
      // 获取 JWT token 并添加到 URL 参数中
      const jwtToken = proxy?.$cache.local.get('token')
      // 构建视频流 URL（包含 playToken 和 JWT token）
      videoUrl.value = getVideoStreamUrl(res.data.play_token, jwtToken || undefined)
    } else {
      throw new Error(res.message || t('preview.video.getTokenFailed'))
    }
  } catch (err: any) {
    const errorMessage = err?.response?.data?.message || err?.message || t('preview.video.loadFailed')
    throw new Error(errorMessage)
  }
}

// 图片位置和缩放状态
const imagePosition = ref({ x: 0, y: 0 })
const isDragging = ref(false)
const dragStart = ref({ x: 0, y: 0 })
const isFullscreen = ref(false)

// 图片样式
const imageStyle = computed(() => {
  const zoom = options.value.zoom || 1
  const rotate = options.value.rotate || 0
  const x = imagePosition.value.x
  const y = imagePosition.value.y
  return {
    transform: `translate(${x}px, ${y}px) scale(${zoom}) rotate(${rotate}deg)`,
    transformOrigin: 'center center',
    transition: isDragging.value ? 'none' : 'transform 0.3s ease',
    cursor: zoom > 1 ? 'grab' : 'default'
  }
})

// 图片缩放值（用于显示）
const imageZoom = computed(() => options.value.zoom || 1)

// 滚轮缩放
const handleImageWheel = (e: WheelEvent) => {
  if (previewType.value !== 'image') return
  e.preventDefault()
  const delta = e.deltaY > 0 ? -0.1 : 0.1
  zoomImage(delta)
}

// 图片鼠标按下（开始拖拽）
const handleImageMouseDown = (e: MouseEvent) => {
  if (previewType.value !== 'image' || imageZoom.value <= 1) return
  e.preventDefault()
  isDragging.value = true
  dragStart.value = {
    x: e.clientX - imagePosition.value.x,
    y: e.clientY - imagePosition.value.y
  }
  
  const handleMouseMove = (moveEvent: MouseEvent) => {
    if (!isDragging.value) return
    imagePosition.value = {
      x: moveEvent.clientX - dragStart.value.x,
      y: moveEvent.clientY - dragStart.value.y
    }
  }
  
  const handleMouseUp = () => {
    isDragging.value = false
    document.removeEventListener('mousemove', handleMouseMove)
    document.removeEventListener('mouseup', handleMouseUp)
  }
  
  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)
}

// 全屏切换
const toggleFullscreen = () => {
  if (!currentFile.value) return
  
  const dialog = document.querySelector('.file-preview-dialog .el-dialog') as HTMLElement
  if (!dialog) return
  
  if (!isFullscreen.value) {
    // 进入全屏
    if (dialog.requestFullscreen) {
      dialog.requestFullscreen()
    } else if ((dialog as any).webkitRequestFullscreen) {
      (dialog as any).webkitRequestFullscreen()
    } else if ((dialog as any).mozRequestFullScreen) {
      (dialog as any).mozRequestFullScreen()
    } else if ((dialog as any).msRequestFullscreen) {
      (dialog as any).msRequestFullscreen()
    }
    isFullscreen.value = true
  } else {
    // 退出全屏
    if (document.exitFullscreen) {
      document.exitFullscreen()
    } else if ((document as any).webkitExitFullscreen) {
      (document as any).webkitExitFullscreen()
    } else if ((document as any).mozCancelFullScreen) {
      (document as any).mozCancelFullScreen()
    } else if ((document as any).msExitFullscreen) {
      (document as any).msExitFullscreen()
    }
    isFullscreen.value = false
  }
}

// 监听全屏状态变化
onMounted(() => {
  const handleFullscreenChange = () => {
    isFullscreen.value = !!(
      document.fullscreenElement ||
      (document as any).webkitFullscreenElement ||
      (document as any).mozFullScreenElement ||
      (document as any).msFullscreenElement
    )
  }
  
  document.addEventListener('fullscreenchange', handleFullscreenChange)
  document.addEventListener('webkitfullscreenchange', handleFullscreenChange)
  document.addEventListener('mozfullscreenchange', handleFullscreenChange)
  document.addEventListener('MSFullscreenChange', handleFullscreenChange)
  
  onBeforeUnmount(() => {
    document.removeEventListener('fullscreenchange', handleFullscreenChange)
    document.removeEventListener('webkitfullscreenchange', handleFullscreenChange)
    document.removeEventListener('mozfullscreenchange', handleFullscreenChange)
    document.removeEventListener('MSFullscreenChange', handleFullscreenChange)
  })
})

// 加载文件内容
const loadFileContent = async () => {
  if (!currentFile.value) return

  // 如果文件已加密，不加载预览
  if (currentFile.value.is_enc) {
    loading.value = false
    return
  }

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
          await loadVideoContent(fileId)
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
  imagePosition.value = { x: 0, y: 0 }
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
  // 重置图片位置
  imagePosition.value = { x: 0, y: 0 }
  isDragging.value = false
  isFullscreen.value = false
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
  display: flex;
  flex-direction: column;
  min-height: 0;
  /* 根据内容类型动态设置高度和滚动 */
  max-height: 80vh;
  overflow-y: auto;
}

/* 视频预览时，完全禁用滚动条 - 基础覆盖 */
.file-preview-dialog.preview-video-active :deep(.el-dialog__body) {
  overflow: hidden !important;
  overflow-y: hidden !important;
  overflow-x: hidden !important;
}

/* 视频预览时，限制 dialog 整体高度 */
.file-preview-dialog.preview-video-active :deep(.el-dialog) {
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}

.file-preview-dialog.preview-video-active :deep(.el-dialog__wrapper) {
  overflow: hidden;
}

.preview-loading,
.preview-error,
.preview-unsupported,
.preview-encrypted {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  gap: 16px;
}

.preview-encrypted .encrypted-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.preview-encrypted .encrypted-desc {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 8px 0 0 0;
}

.preview-encrypted .encrypted-actions {
  margin-top: 24px;
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
  overflow: hidden;
  padding: 20px;
  position: relative;
  cursor: grab;
}

.image-wrapper:active {
  cursor: grabbing;
}

.image-nav-hint {
  position: absolute;
  top: 12px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0, 0, 0, 0.7);
  color: white;
  padding: 8px 16px;
  border-radius: 20px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 6px;
  z-index: 10;
  backdrop-filter: blur(8px);
  animation: fadeIn 0.3s ease;
}

.preview-image {
  max-width: 100%;
  max-height: 70vh;
  object-fit: contain;
}

.preview-audio-container,
.preview-pdf-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.preview-video {
  width: 100%;
  max-height: 70vh;
  background: var(--el-bg-color-page, #000);
  border-radius: 8px;
}

.encrypted-icon {
  color: var(--el-color-warning);
}

.error-icon {
  color: var(--el-color-danger);
}

.unsupported-icon {
  color: var(--el-text-color-placeholder);
}

.preview-video-plyr {
  width: 100%;
  flex: 1;
  min-height: 0;
  /* 根据浏览器视口高度自适应，而不是根据视频比例 */
  height: 100%;
  border-radius: 8px;
  overflow: hidden;
  background: var(--el-bg-color-page, #000);
  /* 最小高度确保在小屏幕上也能正常显示 */
  min-height: 400px;
}

.preview-video-plyr :deep(.plyr) {
  width: 100%;
  height: 100%;
  max-width: 100%;
  max-height: 100%;
}

.preview-video-plyr :deep(.plyr__video-wrapper) {
  width: 100%;
  height: 100%;
  position: relative;
}

.preview-video-plyr :deep(video) {
  width: 100%;
  height: 100%;
  object-fit: contain; /* 视频适应容器，保持原始比例 */
}

/* 视频预览容器：使用 flexbox 自适应高度，避免滚动条 */
.preview-video-container {
  flex: 1;
  min-height: 0;
  max-height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  gap: 0; /* 移除 gap，避免额外空间 */
  /* 确保容器不会超出父元素 */
  box-sizing: border-box;
  /* 让播放器根据可用空间自适应 */
  height: 100%;
}


/* 视频预览时，el-dialog__body 不显示滚动条 - 使用更高优先级覆盖全局样式 */
.file-preview-dialog.preview-video-active :deep(.el-dialog__body),
.file-preview-dialog.preview-video-active.el-dialog .el-dialog__body {
  overflow: hidden !important;
  overflow-y: hidden !important;
  overflow-x: hidden !important;
  /* 使用固定高度：视口高度 - header高度(约60px) */
  height: calc(80vh - 60px) !important;
  max-height: calc(80vh - 60px) !important;
  padding: 20px;
  display: flex !important;
  flex-direction: column;
  box-sizing: border-box;
  /* 确保内容不会溢出 */
  position: relative;
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
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  border-top: 1px solid var(--border-color);
  flex-shrink: 0; /* 工具栏不收缩，保持固定高度 */
  margin-top: auto; /* 在 flex 容器中自动推到底部 */
  min-height: 54px; /* 确保工具栏有最小高度：按钮高度(32px) + padding(16px) + border(1px) */
  box-sizing: border-box;
  gap: 16px;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
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
  
  /* 移动端视频预览时，el-dialog__body 不显示滚动条 */
  .file-preview-dialog.preview-video-active :deep(.el-dialog__body) {
    overflow: hidden !important;
    overflow-y: hidden !important;
    overflow-x: hidden !important;
    height: calc(85vh - 60px) !important;
    max-height: calc(85vh - 60px) !important;
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
  
  /* 小屏幕视频预览时，el-dialog__body 不显示滚动条 */
  .file-preview-dialog.preview-video-active :deep(.el-dialog__body) {
    overflow: hidden !important;
    overflow-y: hidden !important;
    overflow-x: hidden !important;
    height: calc(100vh - 120px) !important;
    max-height: calc(100vh - 120px) !important;
    padding: 8px;
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

<!-- 非 scoped 样式，用于覆盖全局样式 -->
<style>
/* 视频预览弹窗：覆盖全局的 el-dialog__body 样式 */
/* 使用多种选择器确保能够匹配到，包括可能的 DOM 结构变化 */
.file-preview-dialog.preview-video-active .el-dialog__body,
.el-dialog.file-preview-dialog.preview-video-active .el-dialog__body,
.file-preview-dialog.preview-video-active.el-dialog .el-dialog__body,
.el-overlay-dialog .file-preview-dialog.preview-video-active .el-dialog__body,
.el-overlay-dialog .el-dialog.file-preview-dialog.preview-video-active .el-dialog__body {
  overflow: hidden !important;
  overflow-y: hidden !important;
  overflow-x: hidden !important;
  height: calc(80vh - 60px) !important;
  max-height: calc(80vh - 60px) !important;
  display: flex !important;
  flex-direction: column !important;
  box-sizing: border-box !important;
}

/* 移动端视频预览 */
@media (max-width: 1024px) {
  .file-preview-dialog.preview-video-active .el-dialog__body,
  .el-dialog.file-preview-dialog.preview-video-active .el-dialog__body {
    height: calc(85vh - 60px) !important;
    max-height: calc(85vh - 60px) !important;
  }
}

/* 小屏幕视频预览 */
@media (max-width: 480px) {
  .file-preview-dialog.preview-video-active .el-dialog__body,
  .el-dialog.file-preview-dialog.preview-video-active .el-dialog__body {
    height: calc(100vh - 120px) !important;
    max-height: calc(100vh - 120px) !important;
  }
}
</style>


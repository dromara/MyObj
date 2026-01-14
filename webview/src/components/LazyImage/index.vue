<template>
  <div ref="containerRef" class="lazy-image-container" :style="containerStyle">
    <!-- 占位符 -->
    <div v-if="!hasLoaded" class="lazy-image-placeholder" :class="{ 'is-loading': isLoading }">
      <el-icon v-if="!isLoading" class="placeholder-icon"><Picture /></el-icon>
      <div v-else class="loading-spinner"></div>
    </div>

    <!-- 实际图片 -->
    <img
      v-if="hasLoaded"
      :src="src"
      :alt="alt"
      :class="imageClass"
      :style="imageStyle"
      @load="handleLoad"
      @error="handleError"
    />

    <!-- 错误状态 -->
    <div v-if="hasError" class="lazy-image-error">
      <el-icon class="error-icon"><PictureFilled /></el-icon>
      <span v-if="showErrorText" class="error-text">{{ errorText }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useLazyLoad } from '@/composables'

  interface Props {
    /** 图片地址 */
    src: string
    /** 图片描述 */
    alt?: string
    /** 占位符背景色 */
    placeholderColor?: string
    /** 图片宽度 */
    width?: string | number
    /** 图片高度 */
    height?: string | number
    /** 图片对象适配方式 */
    objectFit?: 'fill' | 'contain' | 'cover' | 'none' | 'scale-down'
    /** 是否显示错误文本 */
    showErrorText?: boolean
    /** 错误文本 */
    errorText?: string
    /** 懒加载配置 */
    lazyOptions?: {
      rootMargin?: string
      threshold?: number | number[]
    }
  }

  const props = withDefaults(defineProps<Props>(), {
    alt: '',
    placeholderColor: '#f0f0f0',
    objectFit: 'cover',
    showErrorText: false,
    errorText: '图片加载失败',
    lazyOptions: () => ({
      rootMargin: '50px',
      threshold: 0.1
    })
  })

  const containerRef = ref<HTMLElement>()
  const isLoading = ref(false)
  const hasError = ref(false)

  const { target, hasLoaded } = useLazyLoad({
    rootMargin: props.lazyOptions.rootMargin,
    threshold: props.lazyOptions.threshold
  })

  // 设置观察目标
  watch(
    () => containerRef.value,
    el => {
      if (el) {
        target.value = el
      }
    },
    { immediate: true }
  )

  const containerStyle = computed(() => {
    const style: Record<string, string> = {}
    if (props.width) {
      style.width = typeof props.width === 'number' ? `${props.width}px` : props.width
    }
    if (props.height) {
      style.height = typeof props.height === 'number' ? `${props.height}px` : props.height
    }
    return style
  })

  const imageClass = computed(() => {
    return ['lazy-image', `object-fit-${props.objectFit}`]
  })

  const imageStyle = computed(() => {
    return {
      objectFit: props.objectFit
    }
  })

  const handleLoad = () => {
    isLoading.value = false
    hasError.value = false
  }

  const handleError = () => {
    isLoading.value = false
    hasError.value = true
  }

  // 当图片开始加载时显示加载状态
  watch(
    () => hasLoaded.value,
    loaded => {
      if (loaded) {
        isLoading.value = true
      }
    }
  )
</script>

<style scoped>
  .lazy-image-container {
    position: relative;
    overflow: hidden;
    background: var(--el-fill-color-light);
    border-radius: 4px;
  }

  .lazy-image-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--el-fill-color-light);
    transition: opacity 0.3s ease;
  }

  html.dark .lazy-image-placeholder {
    background: var(--el-fill-color);
  }

  .lazy-image-placeholder.is-loading {
    opacity: 0.6;
  }

  .placeholder-icon {
    font-size: 32px;
    color: var(--el-text-color-placeholder);
  }

  .loading-spinner {
    width: 24px;
    height: 24px;
    border: 2px solid var(--el-border-color);
    border-top-color: var(--el-color-primary);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .lazy-image {
    width: 100%;
    height: 100%;
    display: block;
    transition: opacity 0.3s ease;
  }

  .lazy-image.object-fit-fill {
    object-fit: fill;
  }

  .lazy-image.object-fit-contain {
    object-fit: contain;
  }

  .lazy-image.object-fit-cover {
    object-fit: cover;
  }

  .lazy-image.object-fit-none {
    object-fit: none;
  }

  .lazy-image.object-fit-scale-down {
    object-fit: scale-down;
  }

  .lazy-image-error {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    background: var(--el-fill-color-light);
    color: var(--el-text-color-placeholder);
  }

  html.dark .lazy-image-error {
    background: var(--el-fill-color);
  }

  .error-icon {
    font-size: 32px;
  }

  .error-text {
    font-size: 12px;
    text-align: center;
  }
</style>

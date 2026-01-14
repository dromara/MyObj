<template>
  <div v-if="hasError" class="error-boundary">
    <div class="error-content">
      <el-icon class="error-icon"><WarningFilled /></el-icon>
      <h2 class="error-title">{{ t('error.title') }}</h2>
      <p class="error-message">{{ errorMessage }}</p>
      <div class="error-actions">
        <el-button type="primary" @click="handleRetry">
          {{ t('error.retry') }}
        </el-button>
        <el-button @click="handleReload">
          {{ t('error.reload') }}
        </el-button>
        <el-button v-if="showDetails" @click="showErrorDetails = !showErrorDetails">
          {{ showErrorDetails ? t('error.hideDetails') : t('error.showDetails') }}
        </el-button>
      </div>
      <div v-if="showErrorDetails && errorDetails" class="error-details">
        <pre>{{ errorDetails }}</pre>
      </div>
    </div>
  </div>
  <slot v-else />
</template>

<script setup lang="ts">
  import { useI18n } from '@/composables'

  interface Props {
    fallback?: string
    onError?: (error: Error, errorInfo: any) => void
  }

  const props = withDefaults(defineProps<Props>(), {
    fallback: undefined,
    onError: undefined
  })

  const { t } = useI18n()
  const hasError = ref(false)
  const errorMessage = ref('')
  const errorDetails = ref('')
  const showErrorDetails = ref(false)

  const showDetails = computed(() => !!errorDetails.value)

  // 错误处理
  const handleError = (error: Error, errorInfo?: any) => {
    hasError.value = true
    errorMessage.value = error.message || props.fallback || t('error.unknown')
    errorDetails.value = error.stack || JSON.stringify(errorInfo, null, 2)

    // 调用外部错误处理函数
    if (props.onError) {
      props.onError(error, errorInfo)
    }

    // 记录错误日志
    console.error('ErrorBoundary caught an error:', error, errorInfo)
  }

  // 重试
  const handleRetry = () => {
    hasError.value = false
    errorMessage.value = ''
    errorDetails.value = ''
    showErrorDetails.value = false
  }

  // 重新加载页面
  const handleReload = () => {
    window.location.reload()
  }

  // 暴露错误处理方法供外部调用
  defineExpose({
    handleError
  })
</script>

<style scoped>
  .error-boundary {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 400px;
    padding: 40px 20px;
    background: var(--bg-color);
  }

  .error-content {
    text-align: center;
    max-width: 600px;
    padding: 40px;
    background: var(--card-bg);
    border-radius: 12px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  html.dark .error-content {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  .error-icon {
    font-size: 64px;
    color: var(--danger-color);
    margin-bottom: 20px;
  }

  .error-title {
    font-size: 24px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 16px 0;
  }

  .error-message {
    font-size: 16px;
    color: var(--text-regular);
    margin: 0 0 24px 0;
    line-height: 1.6;
  }

  .error-actions {
    display: flex;
    gap: 12px;
    justify-content: center;
    flex-wrap: wrap;
  }

  .error-details {
    margin-top: 24px;
    padding: 16px;
    background: var(--el-fill-color-light);
    border-radius: 8px;
    text-align: left;
    max-height: 300px;
    overflow-y: auto;
  }

  .error-details pre {
    margin: 0;
    font-size: 12px;
    color: var(--text-secondary);
    white-space: pre-wrap;
    word-break: break-all;
  }

  html.dark .error-details {
    background: var(--el-fill-color);
  }

  @media (max-width: 768px) {
    .error-content {
      padding: 24px;
    }

    .error-icon {
      font-size: 48px;
    }

    .error-title {
      font-size: 20px;
    }

    .error-message {
      font-size: 14px;
    }

    .error-actions {
      flex-direction: column;
    }

    .error-actions .el-button {
      width: 100%;
    }
  }
</style>

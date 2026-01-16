<template>
  <el-dialog
    v-model="visible"
    :title="t('onboarding.welcome.title')"
    width="600px"
    :close-on-click-modal="false"
    :show-close="false"
    class="onboarding-welcome-dialog"
  >
    <div class="welcome-content">
      <div class="welcome-header">
        <el-icon class="welcome-icon" :size="64">
          <StarFilled />
        </el-icon>
        <h3>{{ t('onboarding.welcome.subtitle') }}</h3>
        <p class="welcome-description">{{ t('onboarding.welcome.description') }}</p>
      </div>

      <div class="features-list">
        <div v-for="(feature, index) in features" :key="index" class="feature-item">
          <el-icon class="feature-icon" :size="24">
            <component :is="feature.icon" />
          </el-icon>
          <div class="feature-content">
            <h4>{{ feature.title }}</h4>
            <p>{{ feature.description }}</p>
          </div>
        </div>
      </div>

      <div class="welcome-hint">
        <el-icon><InfoFilled /></el-icon>
        <span>{{ t('onboarding.welcome.hint') }}</span>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleSkip">{{ t('onboarding.welcome.skip') }}</el-button>
        <el-button type="primary" @click="handleStart">{{ t('onboarding.welcome.start') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
  import { StarFilled, Upload, FolderAdd, Search, Share, Setting, InfoFilled } from '@element-plus/icons-vue'
  import { useOnboarding, useI18n } from '@/composables'

  const { showWelcomeDialog, completeWelcome, startFeaturesTour } = useOnboarding()
  const { t } = useI18n()

  const visible = computed({
    get: () => showWelcomeDialog.value,
    set: (val) => {
      if (!val) {
        completeWelcome(true)
      }
    }
  })

  const features = computed(() => [
    {
      icon: Upload,
      title: t('onboarding.welcome.features.upload'),
      description: t('onboarding.welcome.features.uploadDesc')
    },
    {
      icon: FolderAdd,
      title: t('onboarding.welcome.features.manage'),
      description: t('onboarding.welcome.features.manageDesc')
    },
    {
      icon: Share,
      title: t('onboarding.welcome.features.share'),
      description: t('onboarding.welcome.features.shareDesc')
    },
    {
      icon: Search,
      title: t('onboarding.welcome.features.search'),
      description: t('onboarding.welcome.features.searchDesc')
    },
    {
      icon: Setting,
      title: t('onboarding.welcome.features.customize'),
      description: t('onboarding.welcome.features.customizeDesc')
    }
  ])

  const handleSkip = () => {
    completeWelcome(true)
  }

  const handleStart = () => {
    completeWelcome(false)
    // 延迟启动功能引导，确保欢迎对话框已关闭
    setTimeout(() => {
      startFeaturesTour()
    }, 300)
  }
</script>

<style scoped>
  .onboarding-welcome-dialog :deep(.el-dialog__body) {
    padding: 24px;
  }

  .welcome-content {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .welcome-header {
    text-align: center;
    padding: 16px 0;
  }

  .welcome-icon {
    color: var(--el-color-primary);
    margin-bottom: 16px;
  }

  .welcome-header h3 {
    margin: 0 0 12px 0;
    font-size: 24px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .welcome-description {
    margin: 0;
    font-size: 14px;
    color: var(--el-text-color-regular);
    line-height: 1.6;
  }

  .features-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .feature-item {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    padding: 16px;
    background: var(--el-bg-color-page);
    border-radius: 8px;
    border: 1px solid var(--el-border-color-lighter);
    transition: all 0.2s;
  }

  .feature-item:hover {
    background: var(--el-fill-color-light);
    border-color: var(--el-border-color);
  }

  .feature-icon {
    color: var(--el-color-primary);
    flex-shrink: 0;
    margin-top: 2px;
  }

  .feature-content {
    flex: 1;
  }

  .feature-content h4 {
    margin: 0 0 6px 0;
    font-size: 16px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .feature-content p {
    margin: 0;
    font-size: 13px;
    color: var(--el-text-color-regular);
    line-height: 1.5;
  }

  .welcome-hint {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px;
    background: var(--el-color-info-light-9);
    border-radius: 6px;
    color: var(--el-color-info);
    font-size: 13px;
  }

  html.dark .welcome-hint {
    background: rgba(64, 158, 255, 0.1);
    color: var(--el-color-info-light-3);
  }

  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }
</style>

<style>
  /* Driver.js 主题样式 */
  .driverjs-theme {
    --driver-color-primary: var(--el-color-primary);
    --driver-color-secondary: var(--el-color-primary-light-3);
    --driver-color-text: var(--el-text-color-primary);
    --driver-color-backdrop: rgba(0, 0, 0, 0.75);
  }

  .driver-popover {
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

  .driver-popover-title {
    font-size: 18px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .driver-popover-description {
    font-size: 14px;
    color: var(--el-text-color-regular);
    line-height: 1.6;
  }

  html.dark .driver-popover {
    background: var(--el-bg-color);
    border-color: var(--el-border-color);
  }

  html.dark .driver-popover-title {
    color: var(--el-text-color-primary);
  }

  html.dark .driver-popover-description {
    color: var(--el-text-color-regular);
  }
</style>

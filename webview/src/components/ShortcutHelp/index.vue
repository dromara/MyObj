<template>
  <el-dialog v-model="visible" :title="t('shortcuts.title')" width="500px" :close-on-click-modal="true" :close-on-press-escape="true">
    <div class="shortcut-hint">
      <el-icon><InfoFilled /></el-icon>
      <span>{{ t('shortcuts.hint') }}</span>
    </div>
    <div class="shortcut-list">
      <div v-for="(shortcut, index) in shortcuts" :key="index" class="shortcut-item">
        <div class="shortcut-keys">
          <kbd v-if="shortcut.ctrl || shortcut.meta">{{ isMac ? 'Cmd' : 'Ctrl' }}</kbd>
          <kbd v-if="shortcut.shift">Shift</kbd>
          <kbd v-if="shortcut.alt">Alt</kbd>
          <kbd>{{ shortcut.key.toUpperCase() }}</kbd>
        </div>
        <div class="shortcut-description">
          {{ shortcut.description || t('shortcuts.noDescription') }}
        </div>
      </div>
    </div>

    <!-- 分隔线（仅登录后显示） -->
    <el-divider v-if="isLoggedIn" />

    <!-- 新手引导区域（仅登录后显示） -->
    <div v-if="isLoggedIn" class="onboarding-section">
      <div class="onboarding-hint">
        <el-icon><Guide /></el-icon>
        <span>{{ t('settings.onboarding.hint') }}</span>
      </div>
      <el-button
        type="primary"
        @click="handleRestartOnboarding"
        class="onboarding-button"
      >
        <el-icon><RefreshRight /></el-icon>
        {{ t('settings.onboarding.reset') }}
      </el-button>
    </div>

    <template #footer>
      <el-button @click="visible = false">{{ t('shortcuts.close') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
  import type { ComponentInternalInstance } from 'vue'
  import { ElMessageBox } from 'element-plus'
  import { useKeyboardShortcuts, useI18n, useOnboarding } from '@/composables'
  import { useAuthStore } from '@/stores/auth'

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { shortcuts, showHelp, toggleHelp } = useKeyboardShortcuts()
  const { resetOnboarding } = useOnboarding()
  const { t } = useI18n()
  const authStore = useAuthStore()
  
  // 检查是否已登录
  const isLoggedIn = computed(() => !!authStore.token)

  // 检测是否为 Mac 系统
  const isMac = computed(() => {
    return /Mac|iPhone|iPod|iPad/i.test(navigator.userAgent)
  })

  const visible = computed({
    get: () => showHelp.value,
    set: val => {
      if (!val && showHelp.value) {
        toggleHelp()
      }
    }
  })

  // 处理重新开始新手引导
  const handleRestartOnboarding = async () => {
    try {
      await ElMessageBox.confirm(
        t('settings.onboarding.resetConfirm'),
        t('settings.onboarding.title'),
        {
          confirmButtonText: t('common.confirm'),
          cancelButtonText: t('common.cancel'),
          type: 'info'
        }
      )
      
      // 重置新手引导
      resetOnboarding()
      
      // 关闭快捷键帮助对话框
      visible.value = false
      
      // 提示用户
      proxy?.$modal.msgSuccess(t('settings.onboarding.resetSuccess'))
    } catch {
      // 用户取消，不做任何操作
    }
  }
</script>

<style scoped>
  .shortcut-hint {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px;
    margin-bottom: 16px;
    background: var(--el-color-info-light-9);
    border-radius: 6px;
    color: var(--el-color-info);
    font-size: 13px;
  }

  html.dark .shortcut-hint {
    background: rgba(64, 158, 255, 0.1);
    color: var(--el-color-info-light-3);
  }

  .shortcut-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .shortcut-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px;
    background: var(--bg-color-glass, rgba(255, 255, 255, 0.5));
    border-radius: 8px;
    border: 1px solid var(--border-light, #f3f4f6);
  }

  .shortcut-keys {
    display: flex;
    gap: 4px;
    align-items: center;
  }

  .shortcut-keys kbd {
    padding: 4px 8px;
    background: var(--card-bg, white);
    border: 1px solid var(--border-color, #e5e7eb);
    border-radius: 4px;
    font-size: 12px;
    font-weight: 600;
    color: var(--text-primary, #111827);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
    min-width: 32px;
    text-align: center;
  }

  .shortcut-description {
    color: var(--text-regular, #374151);
    font-size: 14px;
  }

  html.dark .shortcut-keys kbd {
    background: rgba(30, 41, 59, 0.6);
    border-color: rgba(255, 255, 255, 0.1);
    color: var(--text-primary);
  }

  /* 分隔线样式 */
  :deep(.el-divider) {
    margin: 20px 0;
    border-color: var(--border-light, #f3f4f6);
  }

  html.dark :deep(.el-divider) {
    border-color: var(--el-border-color);
  }

  /* 新手引导区域 */
  .onboarding-section {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px;
    background: var(--bg-color-glass, rgba(255, 255, 255, 0.5));
    border-radius: 8px;
    border: 1px solid var(--border-light, #f3f4f6);
  }

  html.dark .onboarding-section {
    background: rgba(30, 41, 59, 0.3);
    border-color: var(--el-border-color);
  }

  .onboarding-hint {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--text-regular, #374151);
    font-size: 14px;
  }

  html.dark .onboarding-hint {
    color: var(--el-text-color-regular);
  }

  .onboarding-hint .el-icon {
    color: var(--el-color-primary);
    font-size: 16px;
  }

  .onboarding-button {
    width: 100%;
    justify-content: center;
  }

  /* 确保按钮文字是白色 - 蓝色按钮配白色文字 */
  :deep(.onboarding-button.el-button--primary) {
    color: #ffffff !important;
  }

  :deep(.onboarding-button.el-button--primary .el-icon) {
    color: #ffffff !important;
  }

  :deep(.onboarding-button.el-button--primary span) {
    color: #ffffff !important;
  }
</style>

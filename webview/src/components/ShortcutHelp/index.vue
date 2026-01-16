<template>
  <el-dialog v-model="visible" :title="t('shortcuts.title')" width="500px" :close-on-click-modal="true">
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
    <template #footer>
      <el-button @click="visible = false">{{ t('shortcuts.close') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
  import { InfoFilled } from '@element-plus/icons-vue'
  import { useKeyboardShortcuts, useI18n } from '@/composables'

  const { shortcuts, showHelp, toggleHelp } = useKeyboardShortcuts()
  const { t } = useI18n()

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
</style>

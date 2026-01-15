<template>
  <div class="layout-mode-grid">
    <div
      v-for="(config, mode) in layoutConfigs"
      :key="mode"
      class="layout-mode-item"
      :class="{ 'is-active': currentMode === mode }"
      @click="handleModeChange(mode as LayoutMode)"
    >
      <div class="layout-mode-content">
        <el-tooltip :content="getModeDescription(mode)" placement="bottom">
          <div class="layout-preview" :class="config.previewClass">
            <template v-if="mode === 'vertical-hybrid-header-first'">
              <div class="preview-sider preview-sider-small"></div>
              <div class="preview-sider preview-sider-large"></div>
              <div class="preview-header preview-header-primary"></div>
              <div class="preview-main"></div>
            </template>
            <template v-else-if="mode === 'top-hybrid-sidebar-first'">
              <div class="preview-header preview-header-secondary"></div>
              <div class="preview-sider preview-sider-horizontal"></div>
              <div class="preview-main"></div>
            </template>
            <template v-else-if="mode === 'top-hybrid-header-first'">
              <div class="preview-header preview-header-primary"></div>
              <div class="preview-sider preview-sider-horizontal"></div>
              <div class="preview-main"></div>
            </template>
            <template v-else>
              <div class="preview-sider" v-if="config.showSider"></div>
              <div class="preview-header" v-if="config.showHeader"></div>
              <div class="preview-main"></div>
            </template>
          </div>
        </el-tooltip>
        <div class="layout-mode-name-wrapper">
          <p class="layout-mode-name">{{ getModeName(mode) }}</p>
          <el-tag v-if="mode === 'vertical'" size="small" type="success" class="default-badge">
            {{ t('layout.mode.default') }}
          </el-tag>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from '@/composables'
  import type { LayoutMode } from '@/stores/layout'

  const { t } = useI18n()

  interface Props {
    modelValue: LayoutMode
    disabled?: boolean
  }

  const props = withDefaults(defineProps<Props>(), {
    disabled: false
  })

  const emit = defineEmits<{
    'update:modelValue': [mode: LayoutMode]
  }>()

  const currentMode = computed(() => props.modelValue)

  interface LayoutConfig {
    previewClass: string
    showSider: boolean
    showHeader: boolean
  }

  const layoutConfigs: Record<LayoutMode, LayoutConfig> = {
    vertical: {
      previewClass: 'preview-vertical',
      showSider: true,
      showHeader: true
    },
    horizontal: {
      previewClass: 'preview-horizontal',
      showSider: false,
      showHeader: true
    },
    'vertical-mix': {
      previewClass: 'preview-vertical-mix',
      showSider: true,
      showHeader: true
    },
    'vertical-hybrid-header-first': {
      previewClass: 'preview-vertical-hybrid-header-first',
      showSider: true,
      showHeader: true
    },
    'top-hybrid-sidebar-first': {
      previewClass: 'preview-top-hybrid-sidebar-first',
      showSider: true,
      showHeader: true
    },
    'top-hybrid-header-first': {
      previewClass: 'preview-top-hybrid-header-first',
      showSider: true,
      showHeader: true
    }
  }

  function handleModeChange(mode: LayoutMode) {
    if (props.disabled) return
    emit('update:modelValue', mode)
  }

  function getModeName(mode: string): string {
    return t(`layout.mode.${mode}`)
  }

  function getModeDescription(mode: string): string {
    return t(`layout.mode.${mode}Desc`)
  }
</script>

<style scoped>
  .layout-mode-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: 16px;
    padding: 8px 0;
  }

  /* 超大屏幕：3列布局，更宽松 */
  @media (min-width: 1600px) {
    .layout-mode-grid {
      grid-template-columns: repeat(3, 1fr);
      gap: 20px;
    }
  }

  /* 大屏幕：3列布局 */
  @media (min-width: 1200px) and (max-width: 1599px) {
    .layout-mode-grid {
      grid-template-columns: repeat(3, 1fr);
      gap: 18px;
    }
  }

  /* 中等屏幕：2-3列自适应 */
  @media (min-width: 768px) and (max-width: 1199px) {
    .layout-mode-grid {
      grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
      gap: 16px;
    }
  }

  .layout-mode-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    padding: 8px;
    border-radius: 8px;
    transition: all 0.2s ease;
  }

  .layout-mode-item:hover {
    background: var(--el-fill-color-light);
  }

  .layout-mode-item.is-active {
    background: var(--el-fill-color-light);
  }

  .layout-mode-item.is-active .layout-preview {
    box-shadow: 0 0 0 2px var(--primary-color);
  }

  .layout-preview {
    width: 96px;
    height: 64px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 4px;
    background: var(--card-bg);
    transition: all 0.2s ease;
  }

  html.dark .layout-preview {
    border-color: var(--border-color);
  }

  .preview-vertical {
    display: flex;
    gap: 4px;
  }

  .preview-vertical .preview-sider {
    width: 20px;
    height: 100%;
    background: var(--primary-color);
    border-radius: 2px;
    opacity: 0.8;
  }

  .preview-vertical .preview-header {
    width: 100%;
    height: 12px;
    background: var(--primary-color);
    border-radius: 2px;
    opacity: 0.6;
  }

  .preview-vertical .preview-main {
    flex: 1;
    background: var(--el-fill-color-light);
    border-radius: 2px;
  }

  .preview-horizontal {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .preview-horizontal .preview-header {
    width: 100%;
    height: 12px;
    background: var(--primary-color);
    border-radius: 2px;
    opacity: 0.8;
  }

  .preview-horizontal .preview-main {
    flex: 1;
    background: var(--el-fill-color-light);
    border-radius: 2px;
  }

  .preview-vertical-mix {
    display: flex;
    gap: 4px;
  }

  .preview-vertical-mix .preview-sider {
    width: 8px;
    height: 100%;
    background: var(--primary-color);
    border-radius: 2px;
    opacity: 0.8;
  }

  .preview-vertical-mix .preview-header {
    width: 100%;
    height: 12px;
    background: var(--primary-color);
    border-radius: 2px;
    opacity: 0.6;
  }

  .preview-vertical-mix .preview-main {
    flex: 1;
    background: var(--el-fill-color-light);
    border-radius: 2px;
  }

  .preview-vertical-hybrid-header-first {
    display: flex;
    gap: 4px;
  }

  .preview-vertical-hybrid-header-first .preview-sider-small {
    width: 8px;
    height: 100%;
    background: var(--primary-color);
    border-radius: 2px;
    opacity: 0.8;
  }

  .preview-vertical-hybrid-header-first .preview-sider-large {
    width: 16px;
    height: 100%;
    background: var(--primary-color);
    border-radius: 2px;
    opacity: 0.6;
  }

  .preview-vertical-hybrid-header-first .preview-header-primary {
    width: 100%;
    height: 12px;
    background: var(--primary-color);
    border-radius: 2px;
    opacity: 0.8;
  }

  .preview-vertical-hybrid-header-first .preview-main {
    flex: 1;
    background: var(--el-fill-color-light);
    border-radius: 2px;
  }

  .preview-top-hybrid-sidebar-first {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .preview-top-hybrid-sidebar-first .preview-header-secondary {
    width: 100%;
    height: 12px;
    background: var(--primary-color);
    border-radius: 2px;
    opacity: 0.6;
  }

  .preview-top-hybrid-sidebar-first .preview-sider-horizontal {
    width: 20px;
    height: 100%;
    background: var(--primary-color);
    border-radius: 2px;
    opacity: 0.8;
  }

  .preview-top-hybrid-sidebar-first .preview-main {
    flex: 1;
    background: var(--el-fill-color-light);
    border-radius: 2px;
  }

  .preview-top-hybrid-header-first {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .preview-top-hybrid-header-first .preview-header-primary {
    width: 100%;
    height: 12px;
    background: var(--primary-color);
    border-radius: 2px;
    opacity: 0.8;
  }

  .preview-top-hybrid-header-first .preview-sider-horizontal {
    width: 20px;
    height: 100%;
    background: var(--el-fill-color-light);
    border-radius: 2px;
    opacity: 0.5;
  }

  .preview-top-hybrid-header-first .preview-main {
    flex: 1;
    background: var(--el-fill-color-light);
    border-radius: 2px;
  }

  .layout-mode-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    width: 100%;
  }

  .layout-mode-name-wrapper {
    display: flex;
    align-items: center;
    gap: 6px;
    justify-content: center;
    flex-wrap: wrap;
  }

  .layout-mode-name {
    margin: 0;
    font-size: 12px;
    color: var(--text-regular);
    text-align: center;
    font-weight: 500;
  }

  .layout-mode-item.is-active .layout-mode-name {
    color: var(--primary-color);
    font-weight: 600;
  }

  .default-badge {
    font-size: 10px;
    padding: 2px 6px;
    line-height: 1.2;
  }

  @media (max-width: 768px) {
    .layout-mode-grid {
      grid-template-columns: repeat(2, 1fr);
      gap: 12px;
    }

    .layout-preview {
      width: 80px;
      height: 56px;
    }
  }
</style>

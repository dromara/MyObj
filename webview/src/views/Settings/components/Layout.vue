<template>
  <div class="layout-settings">
    <el-form label-width="120px">
      <!-- 布局模式 -->
      <el-form-item :label="t('layout.mode.title')">
        <LayoutModeCard v-model="currentLayoutMode" :disabled="isMobile" @update:modelValue="handleLayoutModeChange" />
        <div v-if="isMobile" class="layout-tip">
          <el-text type="info" size="small">{{ t('layout.mode.mobileTip') }}</el-text>
        </div>
      </el-form-item>

      <!-- 侧边栏设置 -->
      <el-form-item :label="t('layout.sidebar.title')">
        <div class="sidebar-settings">
          <div class="setting-item">
            <label>{{ t('layout.sidebar.width') }}</label>
            <el-input-number
              v-model="currentSidebarWidth"
              :min="200"
              :max="400"
              :step="10"
              @change="handleSidebarWidthChange"
            />
            <span class="unit">px</span>
          </div>
          <div class="setting-item">
            <el-switch
              v-model="currentSidebarCollapsed"
              :active-text="t('layout.sidebar.collapsed')"
              @change="handleSidebarCollapsedChange"
            />
          </div>
        </div>
      </el-form-item>

      <!-- 标签页设置 -->
      <el-form-item :label="t('layout.tagsView.title')">
        <div class="setting-item">
          <el-switch
            v-model="currentTagsViewVisible"
            :active-text="t('layout.tagsView.visible')"
            @change="handleTagsViewVisibleChange"
          />
        </div>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
  import { useI18n, useResponsive } from '@/composables'
  import { useLayoutStore } from '@/stores'
  import LayoutModeCard from '@/components/LayoutModeCard/index.vue'

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const layoutStore = useLayoutStore()
  const { t } = useI18n()
  const { isMobile } = useResponsive()

  const currentLayoutMode = ref(layoutStore.layoutMode)
  const currentSidebarWidth = ref(layoutStore.sidebarWidth)
  const currentSidebarCollapsed = ref(layoutStore.sidebarCollapsed)
  const currentTagsViewVisible = ref(layoutStore.tagsViewVisible)

  // 初始化布局设置
  onMounted(() => {
    layoutStore.initLayout()
    currentLayoutMode.value = layoutStore.layoutMode
    currentSidebarWidth.value = layoutStore.sidebarWidth
    currentSidebarCollapsed.value = layoutStore.sidebarCollapsed
    currentTagsViewVisible.value = layoutStore.tagsViewVisible
  })

  // 监听布局模式变化
  watch(
    () => layoutStore.layoutMode,
    newMode => {
      currentLayoutMode.value = newMode
    }
  )

  // 监听侧边栏宽度变化
  watch(
    () => layoutStore.sidebarWidth,
    newWidth => {
      currentSidebarWidth.value = newWidth
    }
  )

  // 监听侧边栏折叠状态变化
  watch(
    () => layoutStore.sidebarCollapsed,
    newCollapsed => {
      currentSidebarCollapsed.value = newCollapsed
    }
  )

  // 监听标签页显示状态变化
  watch(
    () => layoutStore.tagsViewVisible,
    newVisible => {
      currentTagsViewVisible.value = newVisible
    }
  )

  const handleLayoutModeChange = (mode: typeof layoutStore.layoutMode) => {
    layoutStore.setLayoutMode(mode)
    proxy?.$modal.msgSuccess(t('layout.mode.changed', { mode: t(`layout.mode.${mode}`) }))
  }

  const handleSidebarWidthChange = (width: number | undefined) => {
    if (width !== undefined && width !== null) {
      layoutStore.setSidebarWidth(width)
      proxy?.$modal.msgSuccess(t('layout.sidebar.widthChanged', { width }))
    }
  }

  const handleSidebarCollapsedChange = (val: string | number | boolean) => {
    const collapsed = val === true || val === 'true'
    layoutStore.setSidebarCollapsed(collapsed)
    proxy?.$modal.msgSuccess(collapsed ? t('layout.sidebar.collapsedEnabled') : t('layout.sidebar.collapsedDisabled'))
  }

  const handleTagsViewVisibleChange = (val: string | number | boolean) => {
    const visible = val === true || val === 'true'
    layoutStore.setTagsViewVisible(visible)
    proxy?.$modal.msgSuccess(visible ? t('layout.tagsView.visibleEnabled') : t('layout.tagsView.visibleDisabled'))
  }
</script>

<style scoped>
  .layout-settings {
    padding: 8px 0;
    width: 100%;
    max-width: 1000px;
  }

  .layout-tip {
    margin-top: 12px;
    padding: 8px 12px;
    background: var(--el-fill-color-light);
    border-radius: 6px;
  }

  html.dark .layout-tip {
    background: var(--el-fill-color-light);
  }

  .sidebar-settings {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .setting-item {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .setting-item label {
    min-width: 80px;
    font-size: 14px;
    color: var(--text-regular);
    font-weight: 500;
  }

  .setting-item .unit {
    font-size: 14px;
    color: var(--text-secondary);
  }
</style>

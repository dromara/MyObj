<template>
  <div class="layout-container">
    <!-- 背景图案 -->
    <BackgroundPattern :pattern="backgroundPattern" />

    <!-- 垂直布局 -->
    <template v-if="layoutMode === 'vertical'">
      <el-container direction="vertical" class="layout-content">
        <Header />
        <el-container class="main-container layout-vertical">
          <Sidebar />
          <div class="content-wrapper">
            <TagsView v-if="layoutStore.tagsViewVisible" />
            <AppMain />
          </div>
        </el-container>
      </el-container>
    </template>

    <!-- 水平布局 -->
    <template v-else-if="layoutMode === 'horizontal'">
      <el-container direction="vertical" class="layout-content">
        <Header />
        <el-container class="main-container layout-horizontal">
          <!-- 水平布局时不显示侧边栏，菜单在 Header 中 -->
          <div class="content-wrapper">
            <TagsView v-if="layoutStore.tagsViewVisible" />
            <AppMain />
          </div>
        </el-container>
      </el-container>
    </template>

    <!-- 垂直混合布局 -->
    <template v-else-if="layoutMode === 'vertical-mix'">
      <el-container direction="vertical" class="layout-content">
        <Header />
        <el-container class="main-container layout-vertical-mix">
          <Sidebar mode="icon" :show-storage-card="false" class="mix-sider-small" />
          <Sidebar mode="full" class="mix-sider-large" />
          <div class="content-wrapper">
            <TagsView v-if="layoutStore.tagsViewVisible" />
            <AppMain />
          </div>
        </el-container>
      </el-container>
    </template>

    <!-- 垂直混合布局 - 头部优先 -->
    <template v-else-if="layoutMode === 'vertical-hybrid-header-first'">
      <el-container direction="vertical" class="layout-content">
        <Header class="hybrid-header-primary" />
        <el-container class="main-container layout-vertical-hybrid-header-first">
          <Sidebar mode="icon" :show-storage-card="false" class="mix-sider-small" />
          <Sidebar mode="full" class="mix-sider-large" />
          <div class="content-wrapper">
            <TagsView v-if="layoutStore.tagsViewVisible" />
            <AppMain />
          </div>
        </el-container>
      </el-container>
    </template>

    <!-- 顶部混合布局 - 侧边栏优先 -->
    <template v-else-if="layoutMode === 'top-hybrid-sidebar-first'">
      <el-container direction="vertical" class="layout-content">
        <Header class="hybrid-header-secondary" />
        <el-container class="main-container layout-top-hybrid-sidebar-first">
          <Sidebar class="hybrid-sider-horizontal" />
          <div class="content-wrapper">
            <TagsView v-if="layoutStore.tagsViewVisible" />
            <AppMain />
          </div>
        </el-container>
      </el-container>
    </template>

    <!-- 顶部混合布局 - 头部优先 -->
    <template v-else-if="layoutMode === 'top-hybrid-header-first'">
      <el-container direction="vertical" class="layout-content">
        <Header class="hybrid-header-primary" />
        <el-container class="main-container layout-top-hybrid-header-first">
          <!-- 顶部混合-头部优先：菜单在 Header 中，侧边栏不显示（因为没有二级菜单） -->
          <div class="content-wrapper">
            <TagsView v-if="layoutStore.tagsViewVisible" />
            <AppMain />
          </div>
        </el-container>
      </el-container>
    </template>
  </div>
</template>

<script setup lang="ts">
  import { Header, Sidebar, AppMain, TagsView } from './components'
  import { useLayoutStore } from '@/stores'

  const layoutStore = useLayoutStore()

  // 初始化布局配置
  onMounted(() => {
    layoutStore.initLayout()
  })

  // 计算布局模式
  const layoutMode = computed(() => layoutStore.layoutMode)

  // 监听布局模式变化，动态调整布局
  watch(
    () => layoutStore.layoutMode,
    () => {
      // 布局模式变化时，可能需要重新计算高度等
      nextTick(() => {
        // 可以在这里添加布局切换的动画或逻辑
      })
    }
  )

  // 背景图案设置（可以从设置页面配置）
  const backgroundPattern = ref<'none' | 'grid' | 'dots' | 'gradient' | 'waves' | 'particles'>('none')

  // 监听背景图案变化事件
  const handleBackgroundPatternChange = (event: Event) => {
    const customEvent = event as CustomEvent<{ pattern: string }>
    if (customEvent.detail?.pattern) {
      const pattern = customEvent.detail.pattern
      if (['none', 'grid', 'dots', 'gradient', 'waves', 'particles'].includes(pattern)) {
        backgroundPattern.value = pattern as any
      }
    }
  }

  // 从 localStorage 加载背景图案设置
  const loadBackgroundPattern = () => {
    const saved = localStorage.getItem('backgroundPattern')
    if (saved && ['none', 'grid', 'dots', 'gradient', 'waves', 'particles'].includes(saved)) {
      backgroundPattern.value = saved as any
    }
  }

  // 组件挂载时加载设置并监听事件
  onMounted(() => {
    loadBackgroundPattern()

    // 添加事件监听器（使用 capture 确保能捕获到事件）
    window.addEventListener('background-pattern-changed', handleBackgroundPatternChange, true)
  })

  // 清理事件监听器
  onBeforeUnmount(() => {
    window.removeEventListener('background-pattern-changed', handleBackgroundPatternChange, true)
  })
</script>

<style scoped>
  .layout-container {
    width: 100%;
    height: 100vh;
    background: var(--bg-color);
    overflow: hidden;
    position: relative;
  }

  .layout-content {
    position: relative;
    z-index: 1;
  }

  .layout-container :deep(.el-container) {
    height: 100%;
  }

  .main-container {
    height: calc(100vh - 64px);
    overflow: hidden;
    transition: all 0.3s ease;
  }

  /* 内容包装器 */
  .content-wrapper {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-width: 0;
  }

  /* 当 TagsView 显示时，调整 AppMain 高度 */
  .content-wrapper:has(.tags-view-container) .layout-main {
    height: calc(100% - 34px);
  }

  /* 混合布局头部高度调整 */
  .hybrid-header-primary + .tags-view-container ~ .main-container,
  .layout-content:has(.hybrid-header-primary) .main-container {
    height: calc(100vh - 80px);
  }

  .layout-content:has(.hybrid-header-primary) .content-wrapper:has(.tags-view-container) .layout-main {
    height: calc(100% - 34px);
  }

  .hybrid-header-secondary + .tags-view-container ~ .main-container,
  .layout-content:has(.hybrid-header-secondary) .main-container {
    height: calc(100vh - 48px);
  }

  .layout-content:has(.hybrid-header-secondary) .content-wrapper:has(.tags-view-container) .layout-main {
    height: calc(100% - 34px);
  }

  /* 布局模式样式 */
  .main-container.layout-vertical {
    display: flex;
    flex-direction: row;
  }

  .main-container.layout-horizontal {
    display: flex;
    flex-direction: column;
  }

  /* 垂直混合布局 */
  .main-container.layout-vertical-mix {
    display: flex;
    flex-direction: row;
    align-items: stretch;
  }

  .main-container.layout-vertical-mix .mix-sider-small {
    width: 64px !important;
    min-width: 64px !important;
    max-width: 64px !important;
    flex-shrink: 0;
    flex-grow: 0;
  }

  .main-container.layout-vertical-mix .mix-sider-large {
    width: 200px !important;
    min-width: 200px !important;
    max-width: 200px !important;
    flex-shrink: 0;
    flex-grow: 0;
  }

  .main-container.layout-vertical-mix .mix-sider-small :deep(.layout-aside) {
    width: 64px !important;
  }

  .main-container.layout-vertical-mix .mix-sider-large :deep(.layout-aside) {
    width: 200px !important;
  }

  /* 垂直混合布局 - 头部优先 */
  .main-container.layout-vertical-hybrid-header-first {
    display: flex;
    flex-direction: row;
    align-items: stretch;
  }

  .main-container.layout-vertical-hybrid-header-first .mix-sider-small {
    width: 64px !important;
    min-width: 64px !important;
    max-width: 64px !important;
    flex-shrink: 0;
    flex-grow: 0;
  }

  .main-container.layout-vertical-hybrid-header-first .mix-sider-large {
    width: 200px !important;
    min-width: 200px !important;
    max-width: 200px !important;
    flex-shrink: 0;
    flex-grow: 0;
  }

  .main-container.layout-vertical-hybrid-header-first .mix-sider-small :deep(.layout-aside) {
    width: 64px !important;
  }

  .main-container.layout-vertical-hybrid-header-first .mix-sider-large :deep(.layout-aside) {
    width: 200px !important;
  }

  .hybrid-header-primary {
    height: 80px !important;
  }

  /* 顶部混合布局 - 侧边栏优先 */
  .main-container.layout-top-hybrid-sidebar-first {
    display: flex;
    flex-direction: row;
    align-items: stretch;
  }

  .main-container.layout-top-hybrid-sidebar-first .hybrid-sider-horizontal {
    width: 240px !important;
    min-width: 240px !important;
    max-width: 240px !important;
    flex-shrink: 0;
    flex-grow: 0;
  }

  .main-container.layout-top-hybrid-sidebar-first .hybrid-sider-horizontal :deep(.layout-aside) {
    width: 240px !important;
  }

  .hybrid-header-secondary {
    height: 48px !important;
  }

  /* 顶部混合布局 - 头部优先 */
  .main-container.layout-top-hybrid-header-first {
    display: flex;
    flex-direction: row;
    align-items: stretch;
  }

  .main-container.layout-top-hybrid-header-first .hybrid-sider-horizontal {
    width: 240px !important;
    min-width: 240px !important;
    max-width: 240px !important;
    flex-shrink: 0;
    flex-grow: 0;
  }

  .main-container.layout-top-hybrid-header-first .hybrid-sider-horizontal :deep(.layout-aside) {
    width: 240px !important;
  }
</style>

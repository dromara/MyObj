<template>
  <!-- 移动端抽屉式侧边栏 -->
  <el-drawer
    v-if="isMobile"
    v-model="sidebarVisible"
    :with-header="false"
    size="280px"
    direction="ltr"
    :modal="true"
    :show-close="false"
    :close-on-click-modal="true"
    :modal-class="'sidebar-drawer-modal'"
    class="sidebar-drawer"
    @close="handleDrawerClose"
  >
    <div class="drawer-content">
      <el-menu :default-active="currentRoute" router @select="handleMenuSelect" class="premium-menu">
      <template v-for="group in menuGroups" :key="group.title || 'group'">
        <div v-if="group.title" class="menu-group-title">
          {{ group.title }}
        </div>
        <template v-for="item in group.items" :key="item.path || 'item'">
          <el-menu-item v-if="!item.hidden && item.path" :index="item.path">
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </el-menu-item>
        </template>
      </template>
      </el-menu>

      <div class="storage-card-wrapper">
        <StorageCard />
      </div>
    </div>
  </el-drawer>

  <!-- 桌面端固定侧边栏（水平布局时隐藏，顶部混合-头部优先时也隐藏，因为菜单在 Header 中） -->
  <el-aside
    v-if="!isMobile && layoutStore.layoutMode !== 'horizontal' && layoutStore.layoutMode !== 'top-hybrid-header-first'"
    :width="isCollapsed || props.mode === 'icon' ? '64px' : `${sidebarWidth}px`"
    class="layout-aside"
    :class="[`sider-mode-${props.mode}`, { 'is-collapsed': isCollapsed }]"
  >
    <el-menu
      :default-active="currentRoute"
      router
      @select="handleMenuSelect"
      class="premium-menu"
      :collapse="isCollapsed || props.mode === 'icon'"
      :collapse-transition="false"
    >
      <template v-for="group in menuGroups" :key="group.title || 'group'">
        <div v-if="group.title && props.mode !== 'icon' && !isCollapsed" class="menu-group-title">
          {{ group.title }}
        </div>
        <template v-for="item in group.items" :key="item.path || 'item'">
          <el-tooltip v-if="props.mode === 'icon' || isCollapsed" :content="item.label" placement="right">
            <el-menu-item v-if="!item.hidden && item.path" :index="item.path">
              <el-icon><component :is="item.icon" /></el-icon>
            </el-menu-item>
          </el-tooltip>
          <el-menu-item v-else-if="!item.hidden && item.path" :index="item.path">
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </el-menu-item>
        </template>
      </template>
    </el-menu>

    <div v-if="props.showStorageCard" class="storage-card-container">
      <StorageCard />
    </div>
  </el-aside>
</template>

<script setup lang="ts">
  import StorageCard from '../StorageCard/index.vue'
  import { useResponsive } from '@/composables'
  import { useMenu } from '@/composables'
  import { useLayoutStore } from '@/stores'

  interface Props {
    /** 侧边栏模式：'full' 完整菜单, 'icon' 仅图标, 'child' 子菜单 */
    mode?: 'full' | 'icon' | 'child'
    /** 是否显示存储卡片 */
    showStorageCard?: boolean
  }

  const props = withDefaults(defineProps<Props>(), {
    mode: 'full',
    showStorageCard: true
  })

  const { isMobile } = useResponsive()
  const layoutStore = useLayoutStore()
  const { menuGroups, currentRoute } = useMenu()

  // 初始化布局设置
  onMounted(() => {
    layoutStore.initLayout()
  })

  const sidebarWidth = computed(() => layoutStore.sidebarWidth)
  const isCollapsed = computed(() => layoutStore.sidebarCollapsed)

  const sidebarVisible = ref(false)

  const handleMenuSelect = () => {
    // Router handles navigation automatically
    // 移动端点击菜单后关闭侧边栏（使用 nextTick 确保路由跳转完成后再关闭）
    if (isMobile.value) {
      nextTick(() => {
        closeSidebar()
      })
    }
  }

  const closeSidebar = () => {
    sidebarVisible.value = false
  }

  // 处理抽屉关闭事件
  const handleDrawerClose = () => {
    // el-drawer 关闭时会触发此事件，确保状态同步
    sidebarVisible.value = false
  }

  // 监听侧边栏切换事件
  const handleToggleSidebar = () => {
    sidebarVisible.value = !sidebarVisible.value
  }

  // 点击外部关闭侧边栏（参考 plus-ui）
  const handleClickOutside = (event: MouseEvent) => {
    if (isMobile.value && sidebarVisible.value) {
      const target = event.target as HTMLElement
      // 检查点击是否在侧边栏外部
      const drawer = document.querySelector('.sidebar-drawer')
      if (drawer && !drawer.contains(target)) {
        closeSidebar()
      }
    }
  }

  // 监听窗口大小变化，自动调整侧边栏状态
  watch(
    isMobile,
    newVal => {
      if (newVal) {
        // 切换到移动端时，如果侧边栏打开则关闭
        if (sidebarVisible.value) {
          closeSidebar()
        }
      } else {
        // 切换到桌面端时，侧边栏始终显示（桌面端使用固定侧边栏，不依赖 drawer）
        sidebarVisible.value = false
      }
    },
    { immediate: true }
  )

  onMounted(() => {
    // 移动端默认隐藏侧边栏
    if (isMobile.value) {
      sidebarVisible.value = false
    }

    window.addEventListener('toggle-sidebar', handleToggleSidebar)
    // 添加点击外部关闭功能
    document.addEventListener('click', handleClickOutside, true)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('toggle-sidebar', handleToggleSidebar)
    document.removeEventListener('click', handleClickOutside, true)
  })
</script>

<style scoped>
  .layout-aside {
    background: var(--card-bg);
    box-shadow: 4px 0 24px rgba(0, 0, 0, 0.02);
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  html.dark .sidebar-container {
    box-shadow: 4px 0 24px rgba(0, 0, 0, 0.3);
    display: flex;
    flex-direction: column;
    padding: 16px 0;
    z-index: 10;
    height: 100%;
    overflow-y: auto;
    flex-shrink: 0;
  }

  .sidebar-drawer :deep(.el-drawer__body) {
    padding: 0;
    display: flex;
    flex-direction: column;
    height: 100%;
    position: relative;
  }

  .drawer-content {
    display: flex;
    flex-direction: column;
    height: 100%;
    flex: 1;
    min-height: 100%;
  }

  .storage-card-wrapper {
    flex-shrink: 0;
    margin-top: auto;
  }

  .storage-card-container {
    flex-shrink: 0;
    padding: 12px;
    margin-top: auto;
  }

  .premium-menu {
    border: none;
    flex: 1;
    padding: 5px 12px;
    background: transparent;
    overflow-y: auto;
    min-height: 0;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  /* 菜单项文字和图标过渡动画 */
  .premium-menu :deep(.el-menu-item span) {
    transition: opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1), transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    display: inline-block;
  }

  /* 折叠时隐藏文字 */
  .layout-aside.is-collapsed .premium-menu :deep(.el-menu-item span),
  .premium-menu.el-menu--collapse :deep(.el-menu-item span) {
    opacity: 0;
    transform: translateX(-10px);
    width: 0;
    overflow: hidden;
  }

  /* 展开时显示文字 */
  .premium-menu:not(.el-menu--collapse) :deep(.el-menu-item span) {
    opacity: 1;
    transform: translateX(0);
  }

  /* 菜单分组标题过渡 */
  .menu-group-title {
    transition: opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1), transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .layout-aside.is-collapsed .menu-group-title,
  .premium-menu.el-menu--collapse ~ .menu-group-title {
    opacity: 0;
    transform: translateX(-10px);
    height: 0;
    padding: 0;
    margin: 0;
    overflow: hidden;
  }

  .premium-menu :deep(.el-menu-item) {
    height: 48px;
    margin-bottom: 4px;
    border-radius: 10px;
    color: var(--text-regular);
    font-weight: 500;
    border: none;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1), padding 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    position: relative;
    overflow: hidden;
  }

  .premium-menu :deep(.el-menu-item::before) {
    content: '';
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 3px;
    background: var(--primary-color);
    transform: scaleY(0);
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border-radius: 0 3px 3px 0;
  }

  .premium-menu :deep(.el-menu-item:hover) {
    background: var(--el-fill-color-light);
    color: var(--primary-color);
    transform: translateX(2px);
  }

  .premium-menu :deep(.el-menu-item:hover::before) {
    transform: scaleY(1);
  }

  html.dark .premium-menu :deep(.el-menu-item:hover) {
    background: rgba(99, 102, 241, 0.15);
    box-shadow: 0 2px 8px rgba(59, 130, 246, 0.2);
  }

  .premium-menu :deep(.el-menu-item.is-active) {
    background: linear-gradient(135deg, var(--primary-color) 0%, var(--secondary-color) 100%);
    color: white;
    box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
    transform: translateX(0);
  }

  .premium-menu :deep(.el-menu-item.is-active::before) {
    transform: scaleY(1);
    background: rgba(255, 255, 255, 0.3);
    width: 4px;
  }

  html.dark .premium-menu :deep(.el-menu-item.is-active) {
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
  }

  .premium-menu :deep(.el-icon) {
    font-size: 18px;
    margin-right: 12px;
    transition: margin-right 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    flex-shrink: 0;
  }

  /* 折叠时图标居中，移除右边距 */
  .layout-aside.is-collapsed .premium-menu :deep(.el-icon),
  .premium-menu.el-menu--collapse :deep(.el-icon) {
    margin-right: 0;
  }

  .menu-group-title {
    padding: 12px 16px 8px;
    font-size: 11px;
    font-weight: 600;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    opacity: 0.7;
  }

  html.dark .menu-group-title {
    color: var(--text-secondary);
    opacity: 0.6;
  }


  /* 移动端响应式 */
  @media (max-width: 991px) {
    .layout-aside {
      display: none;
    }
  }

  /* 移动端抽屉遮罩层样式（参考 plus-ui） */
  :deep(.sidebar-drawer-modal) {
    background: rgba(0, 0, 0, 0.3);
    backdrop-filter: blur(2px);
    transition: opacity 0.3s ease;
  }

  html.dark :deep(.sidebar-drawer-modal) {
    background: rgba(0, 0, 0, 0.5);
  }

  /* 图标模式样式 */
  .sider-mode-icon {
    width: 64px !important;
  }

  .sider-mode-icon .storage-card-container {
    display: none;
  }

  /* 折叠状态样式 */
  .layout-aside.is-collapsed {
    width: 64px !important;
  }

  .layout-aside.is-collapsed .storage-card-container {
    display: none;
    transition: opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .layout-aside.is-collapsed .premium-menu :deep(.el-menu-item) {
    padding: 0 20px !important;
    justify-content: center;
  }

  /* 菜单项内容过渡 */
  .premium-menu :deep(.el-menu-item) {
    display: flex;
    align-items: center;
  }

  /* 展开时菜单项内容 */
  .premium-menu:not(.el-menu--collapse) :deep(.el-menu-item) {
    padding: 0 16px;
  }

  .sider-mode-icon .premium-menu :deep(.el-menu-item) {
    padding: 0 20px !important;
    justify-content: center;
  }

  .sider-mode-icon .premium-menu :deep(.el-icon) {
    margin-right: 0;
  }
</style>

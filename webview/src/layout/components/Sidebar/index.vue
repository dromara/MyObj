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
    class="sidebar-drawer"
    @close="handleDrawerClose"
  >
    <div class="drawer-content">
      <el-menu
        :default-active="currentRoute"
        router
        @select="handleMenuSelect"
        class="premium-menu"
      >
        <el-menu-item index="/files">
          <el-icon><Folder /></el-icon>
          <span>{{ t('menu.files') }}</span>
        </el-menu-item>
        <el-menu-item index="/shares">
          <el-icon><Share /></el-icon>
          <span>{{ t('menu.shares') }}</span>
        </el-menu-item>
        <el-menu-item index="/offline">
          <el-icon><Download /></el-icon>
          <span>{{ t('menu.offline') }}</span>
        </el-menu-item>
        <el-menu-item index="/tasks">
          <el-icon><List /></el-icon>
          <span>{{ t('menu.tasks') }}</span>
        </el-menu-item>
        <el-menu-item index="/trash">
          <el-icon><Delete /></el-icon>
          <span>{{ t('menu.trash') }}</span>
        </el-menu-item>
        <div class="menu-divider"></div>
        <el-menu-item index="/square">
          <el-icon><Grid /></el-icon>
          <span>{{ t('menu.square') }}</span>
        </el-menu-item>
        <!-- 协作功能暂时隐藏 -->
        <!-- <el-menu-item index="/collaboration">
          <el-icon><UserFilled /></el-icon>
          <span>{{ t('menu.collaboration') }}</span>
        </el-menu-item> -->
        <el-menu-item v-if="isAdmin" index="/admin">
          <el-icon><Setting /></el-icon>
          <span>{{ t('menu.admin') }}</span>
        </el-menu-item>
      </el-menu>
      
      <div class="storage-card-wrapper">
        <StorageCard />
      </div>
    </div>
  </el-drawer>
  
  <!-- 桌面端固定侧边栏 -->
  <el-aside
    v-if="!isMobile"
    width="240px"
    class="layout-aside"
  >
    <el-menu
      :default-active="currentRoute"
      router
      @select="handleMenuSelect"
      class="premium-menu"
    >
      <el-menu-item index="/files">
        <el-icon><Folder /></el-icon>
        <span>{{ t('menu.files') }}</span>
      </el-menu-item>
      <el-menu-item index="/shares">
        <el-icon><Share /></el-icon>
        <span>{{ t('menu.shares') }}</span>
      </el-menu-item>
      <el-menu-item index="/offline">
        <el-icon><Download /></el-icon>
        <span>{{ t('menu.offline') }}</span>
      </el-menu-item>
      <el-menu-item index="/tasks">
        <el-icon><List /></el-icon>
        <span>{{ t('menu.tasks') }}</span>
      </el-menu-item>
      <el-menu-item index="/trash">
        <el-icon><Delete /></el-icon>
        <span>{{ t('menu.trash') }}</span>
      </el-menu-item>
      <div class="menu-divider"></div>
      <el-menu-item index="/square">
        <el-icon><Grid /></el-icon>
        <span>{{ t('menu.square') }}</span>
      </el-menu-item>
      <!-- 协作功能暂时隐藏 -->
      <!-- <el-menu-item index="/collaboration">
        <el-icon><UserFilled /></el-icon>
        <span>{{ t('menu.collaboration') }}</span>
      </el-menu-item> -->
      <el-menu-item v-if="isAdmin" index="/admin">
        <el-icon><Setting /></el-icon>
        <span>{{ t('menu.admin') }}</span>
      </el-menu-item>
    </el-menu>
    
    <div class="storage-card-container">
      <StorageCard />
    </div>
  </el-aside>
</template>

<script setup lang="ts">
import StorageCard from '../StorageCard/index.vue'
import { useAdmin } from '@/composables/useAdmin'
import { useI18n } from '@/composables/useI18n'

const route = useRoute()
const { isAdmin } = useAdmin()
const { t } = useI18n()

const currentRoute = computed(() => route.path)
const sidebarVisible = ref(false)
const isMobile = ref(false)

// 检测是否为移动端
const checkMobile = () => {
  isMobile.value = window.innerWidth <= 1024
  
  if (isMobile.value) {
    // 移动端默认隐藏侧边栏
    sidebarVisible.value = false
  } else {
    // 桌面端默认显示侧边栏
    sidebarVisible.value = true
  }
}

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

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  window.addEventListener('toggle-sidebar', handleToggleSidebar)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', checkMobile)
  window.removeEventListener('toggle-sidebar', handleToggleSidebar)
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
}

.premium-menu :deep(.el-menu-item) {
  height: 48px;
  margin-bottom: 4px;
  border-radius: 10px;
  color: var(--text-regular);
  font-weight: 500;
  border: none;
}

.premium-menu :deep(.el-menu-item:hover) {
  background: var(--el-fill-color-light);
  color: var(--primary-color);
}

html.dark .premium-menu :deep(.el-menu-item:hover) {
  background: rgba(99, 102, 241, 0.15);
}

.premium-menu :deep(.el-menu-item.is-active) {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--secondary-color) 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
}

html.dark .menu-item:hover {
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

html.dark .premium-menu :deep(.el-menu-item.is-active) {
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.premium-menu :deep(.el-icon) {
  font-size: 18px;
  margin-right: 12px;
}

.menu-divider {
  height: 1px;
  background: var(--border-light);
  margin: 12px 16px;
}

/* 移动端响应式 */
@media (max-width: 1024px) {
  .layout-aside {
    display: none;
  }
}
</style>


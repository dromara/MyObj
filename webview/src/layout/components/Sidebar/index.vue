<template>
  <!-- 移动端遮罩层 -->
  <div
    v-if="isMobile && sidebarVisible"
    class="sidebar-overlay"
    @click="closeSidebar"
  ></div>
  
  <!-- 移动端抽屉式侧边栏 -->
  <el-drawer
    v-if="isMobile"
    v-model="sidebarVisible"
    :with-header="false"
    size="280px"
    direction="ltr"
    :modal="true"
    :show-close="false"
    class="sidebar-drawer"
  >
    <el-menu
      :default-active="currentRoute"
      router
      @select="handleMenuSelect"
      class="premium-menu"
    >
      <el-menu-item index="/files" @click="handleMenuClick">
        <el-icon><Folder /></el-icon>
        <span>我的文件</span>
      </el-menu-item>
      <el-menu-item index="/shares" @click="handleMenuClick">
        <el-icon><Share /></el-icon>
        <span>我的分享</span>
      </el-menu-item>
      <el-menu-item index="/offline" @click="handleMenuClick">
        <el-icon><Download /></el-icon>
        <span>离线下载</span>
      </el-menu-item>
      <el-menu-item index="/tasks" @click="handleMenuClick">
        <el-icon><List /></el-icon>
        <span>传输列表</span>
      </el-menu-item>
      <el-menu-item index="/trash" @click="handleMenuClick">
        <el-icon><Delete /></el-icon>
        <span>回收站</span>
      </el-menu-item>
      <div class="menu-divider"></div>
      <el-menu-item index="/square" @click="handleMenuClick">
        <el-icon><Grid /></el-icon>
        <span>文件广场</span>
      </el-menu-item>
    </el-menu>
    
    <StorageCard />
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
        <span>我的文件</span>
      </el-menu-item>
      <el-menu-item index="/shares">
        <el-icon><Share /></el-icon>
        <span>我的分享</span>
      </el-menu-item>
      <el-menu-item index="/offline">
        <el-icon><Download /></el-icon>
        <span>离线下载</span>
      </el-menu-item>
      <el-menu-item index="/tasks">
        <el-icon><List /></el-icon>
        <span>传输列表</span>
      </el-menu-item>
      <el-menu-item index="/trash">
        <el-icon><Delete /></el-icon>
        <span>回收站</span>
      </el-menu-item>
      <div class="menu-divider"></div>
      <el-menu-item index="/square">
        <el-icon><Grid /></el-icon>
        <span>文件广场</span>
      </el-menu-item>
    </el-menu>
    
    <StorageCard />
  </el-aside>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { Folder, Share, Download, List, Delete, Grid } from '@element-plus/icons-vue'
import StorageCard from '../StorageCard/index.vue'

const route = useRoute()

const currentRoute = computed(() => route.path)
const sidebarVisible = ref(false)
const isMobile = ref(false)

// 检测是否为移动端
const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
  // 桌面端默认显示侧边栏
  if (!isMobile.value) {
    sidebarVisible.value = true
  }
}

const handleMenuSelect = () => {
  // Router handles navigation automatically
  // 移动端点击菜单后关闭侧边栏
  if (isMobile.value) {
    closeSidebar()
  }
}

const handleMenuClick = () => {
  // 移动端点击菜单项后关闭侧边栏
  if (isMobile.value) {
    closeSidebar()
  }
}

const closeSidebar = () => {
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
  background: white;
  box-shadow: 4px 0 24px rgba(0, 0, 0, 0.02);
  display: flex;
  flex-direction: column;
  padding: 16px 0;
  z-index: 10;
  height: 100%;
  overflow-y: auto;
  flex-shrink: 0;
}

.sidebar-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 999;
  backdrop-filter: blur(2px);
}

.sidebar-drawer :deep(.el-drawer__body) {
  padding: 0;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.premium-menu {
  border: none;
  flex: 1;
  padding: 0 12px;
  background: transparent;
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
  background: rgba(99, 102, 241, 0.08);
  color: var(--primary-color);
}

.premium-menu :deep(.el-menu-item.is-active) {
  background: linear-gradient(135deg, #2563eb 0%, #4f46e5 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
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
@media (max-width: 768px) {
  .layout-aside {
    display: none;
  }
}
</style>


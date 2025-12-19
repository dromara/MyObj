<template>
  <el-aside width="240px" class="layout-aside">
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
import StorageCard from '../StorageCard/index.vue'

const route = useRoute()

const currentRoute = computed(() => route.path)

const handleMenuSelect = () => {
  // Router handles navigation automatically
}
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
</style>


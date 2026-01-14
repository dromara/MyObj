<template>
  <div class="horizontal-menu">
    <el-menu
      :default-active="currentRoute"
      mode="horizontal"
      router
      @select="handleMenuSelect"
      class="horizontal-menu-list"
    >
      <template v-for="item in menuItems" :key="item.path">
        <el-menu-item v-if="!item.hidden && item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </el-menu-item>
      </template>
    </el-menu>
  </div>
</template>

<script setup lang="ts">
  import { useMenu } from '@/composables'

  const { menuItems, currentRoute } = useMenu()

  const handleMenuSelect = () => {
    // Router handles navigation automatically
  }
</script>

<style scoped>
  .horizontal-menu {
    width: 100%;
    display: flex;
    align-items: center;
    height: 100%;
    overflow: hidden;
  }

  .horizontal-menu-list {
    width: 100%;
    height: 100%;
    border-bottom: none;
    background: transparent;
    display: flex;
    align-items: center;
  }

  .horizontal-menu-list :deep(.el-menu) {
    display: flex;
    align-items: center;
    height: 100%;
    border-bottom: none;
    background: transparent;
  }

  .horizontal-menu-list :deep(.el-menu-item) {
    height: 100%;
    line-height: 64px;
    padding: 0 16px;
    margin: 0 2px;
    border-radius: 4px;
    transition: all 0.2s ease;
    white-space: nowrap;
    display: inline-flex;
    align-items: center;
  }

  .horizontal-menu-list :deep(.el-menu-item:hover) {
    background: var(--el-fill-color-light);
    color: var(--primary-color);
  }

  .horizontal-menu-list :deep(.el-menu-item.is-active) {
    background: var(--primary-color) !important;
    color: white !important;
    border-bottom: 2px solid var(--primary-color);
  }

  .horizontal-menu-list :deep(.el-menu-item.is-active .el-icon) {
    color: white !important;
  }

  .horizontal-menu-list :deep(.el-menu-item.is-active span) {
    color: white !important;
  }

  .horizontal-menu-list :deep(.el-icon) {
    margin-right: 6px;
    font-size: 16px;
    flex-shrink: 0;
    color: inherit;
  }

  .horizontal-menu-list :deep(.el-menu-item span) {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: inherit;
  }

  html.dark .horizontal-menu-list :deep(.el-menu-item:hover) {
    background: rgba(255, 255, 255, 0.1);
  }

  html.dark .horizontal-menu-list :deep(.el-menu-item.is-active) {
    background: var(--primary-color) !important;
    color: white !important;
  }

  html.dark .horizontal-menu-list :deep(.el-menu-item.is-active .el-icon) {
    color: white !important;
  }

  html.dark .horizontal-menu-list :deep(.el-menu-item.is-active span) {
    color: white !important;
  }

  /* 响应式：小屏幕时隐藏图标文字，只显示图标 */
  @media (max-width: 1200px) {
    .horizontal-menu-list :deep(.el-menu-item span) {
      display: none;
    }
    .horizontal-menu-list :deep(.el-menu-item) {
      padding: 0 12px;
    }
    .horizontal-menu-list :deep(.el-icon) {
      margin-right: 0;
    }
  }

  /* 超小屏幕时进一步缩小 */
  @media (max-width: 768px) {
    .horizontal-menu-list :deep(.el-menu-item) {
      padding: 0 8px;
      margin: 0 1px;
    }
  }
</style>

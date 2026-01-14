<template>
  <div class="admin-page">
    <el-card shadow="never" class="admin-header-card">
      <div class="admin-header">
        <div class="header-left">
          <el-icon :size="28" class="admin-icon"><Setting /></el-icon>
          <h2>{{ t('route.admin') }}</h2>
        </div>
      </div>
    </el-card>

    <el-card shadow="never" class="admin-content-card">
      <el-tabs v-model="activeTab" class="admin-tabs">
        <el-tab-pane :label="t('route.adminUsers')" name="users">
          <AdminUsers />
        </el-tab-pane>
        <el-tab-pane :label="t('route.adminGroups')" name="groups">
          <AdminGroups />
        </el-tab-pane>
        <el-tab-pane :label="t('route.adminPermissions')" name="permissions">
          <AdminPermissions />
        </el-tab-pane>
        <el-tab-pane :label="t('route.adminDisks')" name="disks">
          <AdminDisks />
        </el-tab-pane>
        <el-tab-pane :label="t('route.adminSystem')" name="system">
          <AdminSystem />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
  import AdminUsers from './Users/index.vue'
  import AdminGroups from './Groups/index.vue'
  import AdminPermissions from './Permissions/index.vue'
  import AdminDisks from './Disks/index.vue'
  import AdminSystem from './System/index.vue'
  import { useI18n } from '@/composables'

  const route = useRoute()
  const router = useRouter()
  const { t } = useI18n()

  const activeTab = ref('users')

  // 根据路由设置活动标签
  watch(
    () => route.name,
    name => {
      if (name === 'AdminUsers') activeTab.value = 'users'
      else if (name === 'AdminGroups') activeTab.value = 'groups'
      else if (name === 'AdminPermissions') activeTab.value = 'permissions'
      else if (name === 'AdminDisks') activeTab.value = 'disks'
      else if (name === 'AdminSystem') activeTab.value = 'system'
    },
    { immediate: true }
  )

  // 标签切换时更新路由
  watch(activeTab, tab => {
    const routeMap: Record<string, string> = {
      users: '/admin/users',
      groups: '/admin/groups',
      permissions: '/admin/permissions',
      disks: '/admin/disks',
      system: '/admin/system'
    }
    if (routeMap[tab] && route.path !== routeMap[tab]) {
      router.push(routeMap[tab])
    }
  })
</script>

<style scoped>
  .admin-page {
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 4px;
  }

  .admin-header-card {
    flex-shrink: 0;
  }

  .admin-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .admin-icon {
    color: var(--primary-color);
    filter: drop-shadow(0 2px 4px rgba(99, 102, 241, 0.3));
  }

  html.dark .admin-icon {
    filter: drop-shadow(0 2px 4px rgba(99, 102, 241, 0.5));
  }

  .admin-header h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 700;
    color: var(--text-primary);
  }

  .admin-content-card {
    flex: 1;
    overflow: hidden;
  }

  .admin-tabs {
    height: 100%;
  }

  .admin-tabs :deep(.el-tabs__content) {
    height: calc(100% - 55px);
    overflow: auto;
  }

  .admin-tabs :deep(.el-tab-pane) {
    height: 100%;
  }

  /* 深色模式样式 */
  html.dark .admin-header-card {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .admin-content-card {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .admin-tabs :deep(.el-tabs__header) {
    background: transparent;
    border-bottom-color: var(--el-border-color);
  }

  html.dark .admin-tabs :deep(.el-tabs__item) {
    color: var(--el-text-color-regular);
  }

  html.dark .admin-tabs :deep(.el-tabs__item.is-active) {
    color: var(--primary-color);
  }

  html.dark .admin-tabs :deep(.el-tabs__item:hover) {
    color: var(--primary-color);
  }
</style>

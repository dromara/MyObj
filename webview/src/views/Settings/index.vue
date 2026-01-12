<template>
  <div class="settings-page">
    <el-card shadow="never" class="settings-card">
      <template #header>
        <div class="card-header">
          <h2>{{ t('settings.title') }}</h2>
        </div>
      </template>
      
      <el-tabs v-model="activeTab" class="settings-tabs">
        <el-tab-pane :label="t('settings.userInfo.title')" name="userInfo">
          <UserInfo />
        </el-tab-pane>
        
        <el-tab-pane :label="t('settings.password.title')" name="password">
          <Password />
        </el-tab-pane>
        
        <el-tab-pane :label="t('settings.appearance')" name="appearance">
          <Appearance />
        </el-tab-pane>
        
        <el-tab-pane :label="t('settings.apiKey.title')" name="apiKey">
          <ApiKey />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import UserInfo from './modules/UserInfo.vue'
import Password from './modules/Password.vue'
import Appearance from './modules/Appearance.vue'
import ApiKey from './modules/ApiKey.vue'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()
const activeTab = ref('userInfo')
</script>

<style scoped>
.settings-page {
  padding: 24px;
  min-height: calc(100vh - 64px);
}

.settings-card {
  max-width: 900px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.settings-tabs {
  margin-top: 8px;
}

.settings-tabs :deep(.el-tab-pane) {
  padding: 24px 0;
}

@media (max-width: 1024px) {
  .settings-page {
    padding: 16px;
  }
  
  .settings-tabs :deep(.el-tab-pane) {
    padding: 16px 0;
  }
}

/* 深色模式样式 */
html.dark .settings-card {
  background: var(--card-bg);
  border-color: var(--el-border-color);
}

html.dark .settings-card :deep(.el-card__header) {
  background: var(--card-bg);
  border-bottom-color: var(--el-border-color);
}

html.dark .settings-card :deep(.el-card__body) {
  background: var(--card-bg);
}

html.dark .card-header h2 {
  color: var(--el-text-color-primary);
}

html.dark .settings-tabs :deep(.el-tabs__header) {
  background: var(--card-bg);
  border-color: var(--el-border-color);
}

html.dark .settings-tabs :deep(.el-tabs__item) {
  color: var(--el-text-color-primary);
  border-color: var(--el-border-color);
}

html.dark .settings-tabs :deep(.el-tabs__item.is-active) {
  color: var(--primary-color);
  border-bottom-color: var(--primary-color);
}

html.dark .settings-tabs :deep(.el-tabs__nav-wrap::after) {
  background-color: var(--el-border-color);
}
</style>


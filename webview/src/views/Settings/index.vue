<template>
  <div class="settings-page">
    <el-card shadow="never" class="settings-card">
      <template #header>
        <div class="card-header">
          <h2>系统设置</h2>
        </div>
      </template>
      
      <el-tabs v-model="activeTab" class="settings-tabs">
        <el-tab-pane label="用户信息" name="userInfo">
          <UserInfo />
        </el-tab-pane>
        
        <el-tab-pane label="修改密码" name="password">
          <Password />
        </el-tab-pane>
        
        <el-tab-pane label="API Key" name="apiKey">
          <ApiKey />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import UserInfo from './modules/UserInfo.vue'
import Password from './modules/Password.vue'
import ApiKey from './modules/ApiKey.vue'

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
</style>


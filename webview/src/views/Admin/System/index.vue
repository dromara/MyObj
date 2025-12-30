<template>
  <div class="admin-system">
    <el-card shadow="never" class="config-card">
      <template #header>
        <div class="card-header">
          <el-icon><Setting /></el-icon>
          <span>系统配置</span>
        </div>
      </template>
      
      <el-form :model="configData" label-width="150px" v-loading="loading">
        <el-form-item label="允许用户注册">
          <el-switch 
            v-model="configData.allow_register"
            style="--el-switch-off-color: #f56c6c"
          />
          <div class="form-tip">关闭后，新用户将无法注册账号</div>
        </el-form-item>
        
        <el-form-item label="启用 WebDAV">
          <el-switch 
            v-model="configData.webdav_enabled"
            style="--el-switch-off-color: #f56c6c"
          />
          <div class="form-tip">启用后，用户可以通过 WebDAV 协议访问文件</div>
        </el-form-item>
        
        <el-form-item class="button-form-item">
          <el-button type="primary" :loading="saving" @click="handleSave">保存配置</el-button>
          <el-button @click="loadConfig">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="info-card">
      <template #header>
        <div class="card-header">
          <el-icon><InfoFilled /></el-icon>
          <span>系统信息</span>
        </div>
      </template>
      
      <el-descriptions :column="2" border v-loading="loading">
        <el-descriptions-item label="系统版本">
          {{ systemInfo.version || '未知' }}
        </el-descriptions-item>
        <el-descriptions-item label="总用户数">
          {{ systemInfo.total_users || 0 }}
        </el-descriptions-item>
        <el-descriptions-item label="总文件数">
          {{ systemInfo.total_files || 0 }}
        </el-descriptions-item>
        <el-descriptions-item label="运行时间">
          {{ systemInfo.uptime || '未知' }}
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import type { ComponentInternalInstance } from 'vue'
import { getSystemConfig, updateSystemConfig, type SystemConfig } from '@/api/admin'

const { proxy } = getCurrentInstance() as ComponentInternalInstance

const loading = ref(false)
const saving = ref(false)
const configData = reactive<SystemConfig>({
  allow_register: true,
  webdav_enabled: true,
  version: '',
  total_users: 0,
  total_files: 0
})

const systemInfo = reactive<{
  version: string
  total_users: number
  total_files: number
  uptime: string
}>({
  version: '',
  total_users: 0,
  total_files: 0,
  uptime: ''
})

// 加载配置
const loadConfig = async () => {
  loading.value = true
  try {
    const res = await getSystemConfig()
    if (res.code === 200 && res.data) {
      Object.assign(configData, res.data)
      Object.assign(systemInfo, res.data)
    } else {
      proxy?.$modal.msg('系统配置功能开发中')
    }
  } catch (error: any) {
    if (error.response?.status === 404 || error.message?.includes('404')) {
      proxy?.$modal.msg('系统配置功能开发中')
    } else {
      proxy?.$modal.msgError('加载配置失败')
    }
    proxy?.$log?.error(error)
  } finally {
    loading.value = false
  }
}

// 保存配置
const handleSave = async () => {
  saving.value = true
  try {
    const res = await updateSystemConfig({
      allow_register: configData.allow_register,
      webdav_enabled: configData.webdav_enabled
    })
    if (res.code === 200) {
      proxy?.$modal.msgSuccess('配置保存成功')
      loadConfig()
    } else {
      proxy?.$modal.msgError(res.message || '保存失败')
    }
  } catch (error: any) {
    if (error.response?.status === 404 || error.message?.includes('404')) {
      proxy?.$modal.msg('系统配置功能开发中')
    } else {
      proxy?.$modal.msgError(error.message || '保存失败')
    }
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadConfig()
})
</script>

<style scoped>
.admin-system {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.config-card,
.info-card {
  flex-shrink: 0;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.form-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 8px;
  margin-left: 0;
}

/* Switch 样式优化 */
.config-card :deep(.el-form-item) {
  margin-bottom: 24px;
}

.config-card :deep(.el-form-item__content) {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.config-card :deep(.button-form-item .el-form-item__content) {
  flex-direction: row;
  gap: 12px;
}

.config-card :deep(.el-switch) {
  margin-bottom: 4px;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .config-card,
  .info-card {
    margin-bottom: 16px;
  }

  .config-card :deep(.el-form),
  .info-card :deep(.el-descriptions) {
    font-size: 14px;
  }

  .config-card :deep(.el-form-item__label) {
    width: 120px !important;
    font-size: 13px;
  }

  .info-card :deep(.el-descriptions) {
    font-size: 12px;
  }

  .info-card :deep(.el-descriptions__label) {
    width: 100px !important;
    font-size: 12px;
  }
}

@media (max-width: 480px) {
  .config-card :deep(.el-form-item__label) {
    width: 100px !important;
    font-size: 12px;
  }

  .info-card :deep(.el-descriptions) {
    font-size: 11px;
  }

  .info-card :deep(.el-descriptions__label) {
    width: 80px !important;
    font-size: 11px;
  }

  .info-card :deep(.el-descriptions__content) {
    font-size: 11px;
  }
}
</style>


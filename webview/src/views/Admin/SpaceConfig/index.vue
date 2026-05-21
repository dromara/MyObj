<template>
  <div class="space-config">
    <el-card shadow="never" class="config-card">
      <template #header>
        <div class="card-header">
          <el-icon><Coin /></el-icon>
          <span>{{ t('admin.spaceConfig.title') }}</span>
        </div>
      </template>

      <el-form :model="configData" label-width="180px" v-loading="loading">
        <el-form-item :label="t('admin.spaceConfig.defaultEnterpriseSpace')">
          <div class="space-input-group">
            <el-switch
              v-model="enterpriseUnlimited"
              :active-text="t('admin.spaceConfig.unlimited')"
              inactive-text=""
              @change="handleEnterpriseUnlimitedChange"
            />
            <el-input-number
              v-if="!enterpriseUnlimited"
              v-model="enterpriseSpaceGB"
              :min="0"
              :max="999999"
              :precision="2"
              :step="1"
              controls-position="right"
            />
            <span v-if="!enterpriseUnlimited" class="unit-label">GB</span>
          </div>
          <div class="form-tip">{{ t('admin.spaceConfig.defaultEnterpriseSpaceTip') }}</div>
        </el-form-item>

        <el-form-item class="button-form-item">
          <el-button type="primary" :loading="saving" @click="handleSave">{{ t('admin.system.saveConfig') }}</el-button>
          <el-button @click="loadConfig">{{ t('admin.system.reset') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
  import type { ComponentInternalInstance } from 'vue'
  import { adminApi } from '@myobj/api'
  import type { SpaceConfig } from '@myobj/shared'
  const { getSpaceConfig, updateSpaceConfig } = adminApi
  import { useI18n } from '@/composables'

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  const loading = ref(false)
  const saving = ref(false)

  // 企业空间：0 = 无限
  const enterpriseUnlimited = ref(true)
  const enterpriseSpaceGB = ref(0)

  const configData = reactive<SpaceConfig>({
    default_enterprise_space: 0
  })

  const BYTES_PER_GB = 1024 * 1024 * 1024

  // 加载配置
  const loadConfig = async () => {
    loading.value = true
    try {
      const res = await getSpaceConfig()
      if (res.code === 200 && res.data) {
        configData.default_enterprise_space = res.data.default_enterprise_space

        // 转换为GB显示
        if (res.data.default_enterprise_space === 0) {
          enterpriseUnlimited.value = true
          enterpriseSpaceGB.value = 0
        } else {
          enterpriseUnlimited.value = false
          enterpriseSpaceGB.value = Number((res.data.default_enterprise_space / BYTES_PER_GB).toFixed(2))
        }
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('admin.system.loadConfigFailed'))
    } finally {
      loading.value = false
    }
  }

  // 无限空间切换
  const handleEnterpriseUnlimitedChange = (val: boolean) => {
    if (val) {
      enterpriseSpaceGB.value = 0
    }
  }

  // 保存配置
  const handleSave = async () => {
    saving.value = true
    try {
      const enterpriseSpaceBytes = enterpriseUnlimited.value ? 0 : Math.round(enterpriseSpaceGB.value * BYTES_PER_GB)

      const res = await updateSpaceConfig({
        default_enterprise_space: enterpriseSpaceBytes
      })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('admin.system.configSaveSuccess'))
        loadConfig()
      } else {
        proxy?.$modal.msgError(res.message || t('admin.system.saveFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('admin.system.saveFailed'))
    } finally {
      saving.value = false
    }
  }

  onMounted(() => {
    loadConfig()
  })
</script>

<style scoped>
  .space-config {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .config-card {
    flex-shrink: 0;
  }

  .card-header {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 600;
  }

  .space-input-group {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .unit-label {
    font-size: 14px;
    color: var(--el-text-color-regular);
  }

  .form-tip {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    margin-top: 8px;
    margin-left: 0;
  }

  .config-card ::deep(.el-form-item) {
    margin-bottom: 24px;
  }

  .config-card ::deep(.el-form-item__content) {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
  }

  .config-card ::deep(.button-form-item .el-form-item__content) {
    flex-direction: row;
    gap: 12px;
  }

  /* 移动端适配 */
  @media (max-width: 768px) {
    .config-card ::deep(.el-form-item__label) {
      width: 140px !important;
      font-size: 13px;
    }
  }

  @media (max-width: 480px) {
    .config-card ::deep(.el-form-item__label) {
      width: 110px !important;
      font-size: 12px;
    }
  }

  /* 深色模式样式 */
  html.dark .config-card {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .config-card ::deep(.el-card__header) {
    background: var(--card-bg);
    border-bottom-color: var(--el-border-color);
  }

  html.dark .card-header {
    color: var(--el-text-color-primary);
  }

  html.dark .form-tip {
    color: var(--el-text-color-secondary);
  }

  html.dark .config-card ::deep(.el-form-item__label) {
    color: var(--el-text-color-primary);
  }
</style>

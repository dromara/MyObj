<template>
  <div class="admin-system">
    <el-card shadow="never" class="config-card">
      <template #header>
        <div class="card-header">
          <el-icon><Setting /></el-icon>
          <span>{{ t('admin.system.title') }}</span>
        </div>
      </template>

      <el-form :model="configData" label-width="150px" v-loading="loading">
        <el-form-item :label="t('admin.system.allowRegister')">
          <el-switch v-model="configData.allow_register" style="--el-switch-off-color: var(--el-color-danger)" />
          <div class="form-tip">{{ t('admin.system.allowRegisterTip') }}</div>
        </el-form-item>

        <el-form-item :label="t('admin.system.enableWebDAV')">
          <el-switch v-model="configData.webdav_enabled" style="--el-switch-off-color: var(--el-color-danger)" />
          <div class="form-tip">{{ t('admin.system.enableWebDAVTip') }}</div>
        </el-form-item>

        <el-form-item class="button-form-item">
          <el-button type="primary" :loading="saving" @click="handleSave">{{ t('admin.system.saveConfig') }}</el-button>
          <el-button @click="loadConfig">{{ t('admin.system.reset') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="info-card">
      <template #header>
        <div class="card-header">
          <el-icon><InfoFilled /></el-icon>
          <span>{{ t('admin.system.systemInfo') }}</span>
        </div>
      </template>

      <el-descriptions :column="2" border v-loading="loading">
        <el-descriptions-item :label="t('admin.system.systemVersion')">
          {{ systemInfo.version || t('admin.system.unknown') }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('admin.system.totalUsers')">
          {{ systemInfo.total_users || 0 }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('admin.system.totalFiles')">
          {{ systemInfo.total_files || 0 }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('admin.system.uptime')">
          {{ systemInfo.uptime || t('admin.system.unknown') }}
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card shadow="never" class="update-card">
      <template #header>
        <div class="card-header">
          <el-icon><Upload /></el-icon>
          <span>{{ t('admin.system.update') || '系统更新' }}</span>
        </div>
      </template>

      <div class="update-content">
        <div class="update-info">
          <span>{{ t('admin.system.currentVersion') || '当前版本' }}: <el-tag size="small">{{ systemInfo.version || 'unknown' }}</el-tag></span>
        </div>
        <el-button type="primary" :loading="checkingUpdate" @click="handleCheckUpdate">
          {{ t('admin.system.checkUpdate') || '检测更新' }}
        </el-button>
      </div>

      <div v-if="updateInfo" class="update-result">
        <el-alert
          v-if="!updateInfo.has_update"
          :title="t('admin.system.isLatest') || '已是最新版本'"
          type="success"
          show-icon
          :closable="false"
        />
        <div v-else class="update-available">
          <el-alert
            :title="`${t('admin.system.newVersionAvailable') || '发现新版本'}: v${updateInfo.latest_version}`"
            type="warning"
            show-icon
            :closable="false"
          >
            <template #default>
              <div class="update-detail">
                <div v-if="updateInfo.file_size" class="update-size">
                  {{ t('admin.system.fileSize') || '文件大小' }}: {{ formatSize(updateInfo.file_size) }}
                </div>
                <div v-if="updateInfo.release_note" class="update-note">
                  <strong>{{ t('admin.system.releaseNote') || '更新日志' }}:</strong>
                  <pre class="release-note-content">{{ updateInfo.release_note }}</pre>
                </div>
              </div>
            </template>
          </el-alert>
          <el-button
            type="danger"
            :loading="upgrading"
            @click="handleUpgrade"
            style="margin-top: 12px"
          >
            {{ t('admin.system.startUpgrade') || '立即升级' }}
          </el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
  import type { ComponentInternalInstance } from 'vue'
  import { adminApi } from '@myobj/api'
  import type { SystemConfig, UpdateInfo } from '@myobj/shared'
  const { getSystemConfig, updateSystemConfig, checkUpdate, performUpgrade } = adminApi
  import { useI18n } from '@/composables'

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  const loading = ref(false)
  const saving = ref(false)
  const checkingUpdate = ref(false)
  const upgrading = ref(false)
  const updateInfo = ref<UpdateInfo | null>(null)

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

  // 格式化文件大小
  const formatSize = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  // 加载配置
  const loadConfig = async () => {
    loading.value = true
    try {
      const res = await getSystemConfig()
      if (res.code === 200 && res.data) {
        Object.assign(configData, res.data)
        Object.assign(systemInfo, res.data)
      } else {
        proxy?.$modal.msg(t('admin.system.featureDeveloping'))
      }
    } catch (error: any) {
      if (error.response?.status === 404 || error.message?.includes('404')) {
        proxy?.$modal.msg(t('admin.system.featureDeveloping'))
      } else {
        proxy?.$modal.msgError(t('admin.system.loadConfigFailed'))
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
        proxy?.$modal.msgSuccess(t('admin.system.configSaveSuccess'))
        loadConfig()
      } else {
        proxy?.$modal.msgError(res.message || t('admin.system.saveFailed'))
      }
    } catch (error: any) {
      if (error.response?.status === 404 || error.message?.includes('404')) {
        proxy?.$modal.msg(t('admin.system.featureDeveloping'))
      } else {
        proxy?.$modal.msgError(error.message || t('admin.system.saveFailed'))
      }
    } finally {
      saving.value = false
    }
  }

  // 检查更新
  const handleCheckUpdate = async () => {
    checkingUpdate.value = true
    updateInfo.value = null
    try {
      const res = await checkUpdate()
      if (res.code === 200 && res.data) {
        updateInfo.value = res.data
      } else {
        proxy?.$modal.msgError(res.message || '检查更新失败')
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || '检查更新失败')
      proxy?.$log?.error(error)
    } finally {
      checkingUpdate.value = false
    }
  }

  // 执行升级
  const handleUpgrade = async () => {
    if (!updateInfo.value?.download_url) return

    try {
      await proxy?.$modal.confirm(
        `确定要升级到 v${updateInfo.value.latest_version} 吗？升级过程中服务将短暂中断。`,
        '确认升级',
        { confirmButtonText: '立即升级', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }

    upgrading.value = true
    try {
      const res = await performUpgrade(updateInfo.value.download_url)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess('升级任务已启动，服务即将重启，请稍候...')
      } else {
        proxy?.$modal.msgError(res.message || '升级失败')
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || '升级失败')
    } finally {
      upgrading.value = false
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
  .info-card,
  .update-card {
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

  .update-content {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }

  .update-info {
    font-size: 14px;
    color: var(--el-text-color-regular);
  }

  .update-result {
    margin-top: 16px;
  }

  .update-detail {
    margin-top: 8px;
  }

  .update-size {
    font-size: 13px;
    color: var(--el-text-color-secondary);
    margin-bottom: 8px;
  }

  .release-note-content {
    font-family: monospace;
    font-size: 12px;
    white-space: pre-wrap;
    word-break: break-all;
    max-height: 200px;
    overflow-y: auto;
    background: var(--el-fill-color-light);
    padding: 8px;
    border-radius: 4px;
    margin: 4px 0 0 0;
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
    .info-card,
    .update-card {
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

    .update-content {
      flex-direction: column;
      align-items: flex-start;
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

  /* 深色模式样式 */
  html.dark .admin-system {
    background: transparent;
  }

  html.dark .config-card,
  html.dark .info-card,
  html.dark .update-card {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .config-card :deep(.el-card__header),
  html.dark .info-card :deep(.el-card__header),
  html.dark .update-card :deep(.el-card__header) {
    background: var(--card-bg);
    border-bottom-color: var(--el-border-color);
  }

  html.dark .card-header {
    color: var(--el-text-color-primary);
  }

  html.dark .form-tip {
    color: var(--el-text-color-secondary);
  }

  html.dark .config-card :deep(.el-form-item__label) {
    color: var(--el-text-color-primary);
  }

  html.dark .info-card :deep(.el-descriptions) {
    background: transparent;
  }

  html.dark .info-card :deep(.el-descriptions__label) {
    color: var(--el-text-color-regular);
    background: var(--el-bg-color-page);
  }

  html.dark .info-card :deep(.el-descriptions__content) {
    color: var(--el-text-color-primary);
    background: var(--el-bg-color);
  }

  html.dark .info-card :deep(.el-descriptions__table) {
    border-color: var(--el-border-color);
  }

  html.dark .info-card :deep(.el-descriptions__table th),
  html.dark .info-card :deep(.el-descriptions__table td) {
    border-color: var(--el-border-color);
  }
</style>

<template>
  <div class="api-key-manager">
    <!-- 生成 API Key 区域 -->
    <el-card shadow="never" class="generate-card">
      <template #header>
        <div class="card-header">
          <span>{{ t('settings.apiKey.generateNew') }}</span>
        </div>
      </template>

      <el-form ref="generateFormRef" :model="generateForm" label-width="120px" label-position="left">
        <el-form-item :label="t('settings.apiKey.expiresDays')">
          <el-input-number
            v-model="generateForm.expiresDays"
            :min="0"
            :max="365"
            :placeholder="t('settings.apiKey.expiresDaysPlaceholder')"
            style="width: 200px"
          />
          <div class="form-tip">{{ t('settings.apiKey.expiresDaysTip') }}</div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="generating" @click="handleGenerate">
            {{ t('settings.apiKey.generate') }}
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 生成结果 -->
      <el-alert
        v-if="newApiKey"
        :title="t('settings.apiKey.generatedTitle')"
        type="success"
        :closable="false"
        show-icon
        class="api-key-alert"
      >
        <template #default>
          <div class="api-key-result">
            <div class="key-item">
              <span class="key-label">{{ t('settings.apiKey.apiKeyLabel') }}</span>
              <el-input :value="newApiKey.key" readonly class="key-input">
                <template #append>
                  <el-button @click="copyApiKey(newApiKey.key)" icon="CopyDocument">
                    {{ t('settings.apiKey.copy') }}
                  </el-button>
                </template>
              </el-input>
            </div>
            <div class="key-item">
              <span class="key-label">{{ t('settings.apiKey.publicKeyLabel') }}</span>
              <div class="public-key-wrapper">
                <el-input :value="newApiKey.public_key" type="textarea" :rows="4" readonly class="key-input" />
                <el-button @click="copyApiKey(newApiKey.public_key)" icon="CopyDocument" style="margin-top: 8px">
                  {{ t('settings.apiKey.copyPublicKey') }}
                </el-button>
              </div>
            </div>
            <div class="key-tip">
              <el-icon><Warning /></el-icon>
              <span>{{ t('settings.apiKey.warning') }}</span>
            </div>
          </div>
        </template>
      </el-alert>
    </el-card>

    <!-- API Key 列表 -->
    <el-card shadow="never" class="list-card">
      <template #header>
        <div class="card-header">
          <span>{{ t('settings.apiKey.myApiKeys') }}</span>
          <el-button type="primary" size="small" icon="Refresh" :loading="loading" @click="loadApiKeys">
            {{ t('settings.apiKey.refresh') }}
          </el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="apiKeyList"
        :empty-text="t('settings.apiKey.noApiKeys')"
        class="api-key-table"
      >
        <el-table-column :label="t('settings.apiKey.apiKeyLabel')" min-width="200">
          <template #default="{ row }">
            <code class="api-key-code">{{ row.key }}</code>
          </template>
        </el-table-column>

        <el-table-column :label="t('settings.apiKey.createTime')" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('settings.apiKey.expireTime')" width="180">
          <template #default="{ row }">
            <span v-if="row.expires_at && row.expires_at !== null">{{ formatDate(row.expires_at) }}</span>
            <el-tag v-else type="success" size="small">{{ t('settings.apiKey.neverExpires') }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('settings.apiKey.status')" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.is_expired" type="danger" size="small">{{ t('settings.apiKey.expired') }}</el-tag>
            <el-tag v-else type="success" size="small">{{ t('settings.apiKey.valid') }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.operation')" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              type="danger"
              size="small"
              icon="Delete"
              :loading="deletingIds.includes(row.id)"
              @click="handleDelete(row.id)"
            >
              {{ t('settings.apiKey.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
  import { generateApiKey, listApiKeys, deleteApiKey } from '@/api/user'
  import { copyToClipboard } from '@/utils'
  import { formatDate } from '@/utils'
  import type { ApiKeyInfo, GenerateApiKeyResponse } from '@/api/user'
  import { useI18n } from '@/composables'

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  const loading = ref(false)
  const generating = ref(false)
  const deletingIds = ref<number[]>([])

  const generateForm = reactive({
    expiresDays: 0
  })

  const apiKeyList = ref<ApiKeyInfo[]>([])
  const newApiKey = ref<GenerateApiKeyResponse | null>(null)

  // 加载 API Key 列表
  const loadApiKeys = async () => {
    loading.value = true
    try {
      const result = await listApiKeys()
      if (result.code === 200) {
        apiKeyList.value = result.data || []
      } else {
        proxy?.$modal.msgError(result.message || t('settings.apiKey.loadFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('settings.apiKey.loadFailed'))
    } finally {
      loading.value = false
    }
  }

  // 生成 API Key
  const handleGenerate = async () => {
    generating.value = true
    try {
      // 如果 expiresDays 为 0，传递 0 表示永不过期；否则传递实际值
      const result = await generateApiKey({
        expires_days: generateForm.expiresDays
      })

      if (result.code === 200) {
        newApiKey.value = result.data as GenerateApiKeyResponse
        proxy?.$modal.msgSuccess(t('settings.apiKey.generateSuccess'))
        // 重新加载列表
        await loadApiKeys()
        // 重置表单
        generateForm.expiresDays = 0
      } else {
        proxy?.$modal.msgError(result.message || t('settings.apiKey.generateFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('settings.apiKey.generateFailed'))
    } finally {
      generating.value = false
    }
  }

  // 删除 API Key
  const handleDelete = async (id: number) => {
    try {
      await ElMessageBox.confirm(t('settings.apiKey.deleteConfirm'), t('settings.apiKey.deleteTitle'), {
        confirmButtonText: t('settings.apiKey.deleteConfirmButton'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      })

      deletingIds.value.push(id)
      try {
        const result = await deleteApiKey({ api_key_id: id })
        if (result.code === 200) {
          proxy?.$modal.msgSuccess(t('settings.apiKey.deleteSuccess'))
          await loadApiKeys()
        } else {
          proxy?.$modal.msgError(result.message || t('settings.apiKey.deleteFailed'))
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || t('settings.apiKey.deleteFailed'))
      } finally {
        deletingIds.value = deletingIds.value.filter(deleteId => deleteId !== id)
      }
    } catch {
      // 用户取消
    }
  }

  // 复制 API Key
  const copyApiKey = async (text: string) => {
    const success = await copyToClipboard(text)
    if (success) {
      proxy?.$modal.msgSuccess(t('settings.apiKey.copySuccess'))
    } else {
      proxy?.$modal.msgError(t('settings.apiKey.copyFailed'))
    }
  }

  onMounted(() => {
    loadApiKeys()
  })
</script>

<style scoped>
  .api-key-manager {
    width: 100%;
    max-width: 1200px;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .form-tip {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    margin-top: 4px;
  }

  .api-key-alert {
    margin-top: 16px;
  }

  .api-key-result {
    margin-top: 12px;
  }

  .key-item {
    margin-bottom: 16px;
  }

  .key-item:last-child {
    margin-bottom: 0;
  }

  .key-label {
    display: block;
    font-weight: 600;
    margin-bottom: 8px;
    color: var(--el-text-color-primary);
  }

  .key-input {
    width: 100%;
  }

  .key-tip {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 16px;
    padding: 12px;
    background: var(--el-warning-color-light-9);
    border-radius: 4px;
    color: var(--el-warning-color);
    font-size: 14px;
  }

  .api-key-table {
    width: 100%;
  }

  .api-key-code {
    font-family: 'Courier New', monospace;
    font-size: 13px;
    color: var(--el-text-color-primary);
    background: var(--el-fill-color-light);
    padding: 2px 6px;
    border-radius: 3px;
  }

  @media (max-width: 768px) {
    .api-key-manager {
      max-width: 100%;
    }

    .api-key-table {
      font-size: 12px;
    }

    .api-key-code {
      font-size: 11px;
    }
  }

  /* 深色模式样式 */
  html.dark .api-key-manager {
    color: var(--el-text-color-primary);
  }

  html.dark .generate-card {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .generate-card :deep(.el-card__header) {
    background: var(--card-bg);
    border-bottom-color: var(--el-border-color);
  }

  html.dark .generate-card :deep(.el-card__body) {
    background: var(--card-bg);
  }

  html.dark .list-card {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .list-card :deep(.el-card__header) {
    background: var(--card-bg);
    border-bottom-color: var(--el-border-color);
  }

  html.dark .list-card :deep(.el-card__body) {
    background: var(--card-bg);
  }

  html.dark .card-header {
    color: var(--el-text-color-primary);
  }

  html.dark .form-tip {
    color: var(--el-text-color-secondary);
  }

  html.dark .key-label {
    color: var(--el-text-color-primary);
  }

  html.dark .key-input :deep(.el-input__wrapper) {
    background-color: var(--el-bg-color);
    border-color: var(--el-border-color);
  }

  html.dark .key-input :deep(.el-input__inner) {
    color: var(--el-text-color-primary);
  }

  html.dark .key-input :deep(.el-textarea__inner) {
    background-color: var(--el-bg-color);
    border-color: var(--el-border-color);
    color: var(--el-text-color-primary);
  }

  html.dark .key-tip {
    background: var(--el-warning-color-light-9);
    color: var(--el-warning-color);
  }

  html.dark .api-key-table {
    background: var(--card-bg);
  }

  html.dark .api-key-table :deep(.el-table__header-wrapper) {
    background: var(--el-bg-color-page);
  }

  html.dark .api-key-table :deep(.el-table__header th) {
    background: var(--el-bg-color-page);
    color: var(--el-text-color-primary);
    border-color: var(--el-border-color);
  }

  html.dark .api-key-table :deep(.el-table__body tr) {
    background: var(--card-bg);
  }

  html.dark .api-key-table :deep(.el-table__body tr:hover > td) {
    background: var(--el-fill-color-light);
  }

  html.dark .api-key-code {
    background: var(--el-fill-color-light);
    color: var(--el-text-color-primary);
  }
</style>

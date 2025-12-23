<template>
  <div class="api-key-manager">
    <!-- 生成 API Key 区域 -->
    <el-card shadow="never" class="generate-card">
      <template #header>
        <div class="card-header">
          <span>生成新的 API Key</span>
        </div>
      </template>
      
      <el-form
        ref="generateFormRef"
        :model="generateForm"
        label-width="120px"
        label-position="left"
      >
        <el-form-item label="过期天数">
          <el-input-number
            v-model="generateForm.expiresDays"
            :min="0"
            :max="365"
            placeholder="0 表示永不过期"
            style="width: 200px"
          />
          <div class="form-tip">设置为 0 表示永不过期，最大 365 天</div>
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" :loading="generating" @click="handleGenerate">
            生成 API Key
          </el-button>
        </el-form-item>
      </el-form>
      
      <!-- 生成结果 -->
      <el-alert
        v-if="newApiKey"
        :title="`API Key 已生成（请妥善保管，仅显示一次）`"
        type="success"
        :closable="false"
        show-icon
        class="api-key-alert"
      >
        <template #default>
          <div class="api-key-result">
            <div class="key-item">
              <span class="key-label">API Key:</span>
              <el-input
                :value="newApiKey.key"
                readonly
                class="key-input"
              >
                <template #append>
                  <el-button @click="copyApiKey(newApiKey.key)" :icon="CopyDocument">
                    复制
                  </el-button>
                </template>
              </el-input>
            </div>
            <div class="key-item">
              <span class="key-label">公钥:</span>
              <div class="public-key-wrapper">
                <el-input
                  :value="newApiKey.public_key"
                  type="textarea"
                  :rows="4"
                  readonly
                  class="key-input"
                />
                <el-button 
                  @click="copyApiKey(newApiKey.public_key)" 
                  :icon="CopyDocument"
                  style="margin-top: 8px"
                >
                  复制公钥
                </el-button>
              </div>
            </div>
            <div class="key-tip">
              <el-icon><Warning /></el-icon>
              <span>请妥善保管 API Key 和公钥，API Key 生成后无法再次查看</span>
            </div>
          </div>
        </template>
      </el-alert>
    </el-card>
    
    <!-- API Key 列表 -->
    <el-card shadow="never" class="list-card" style="margin-top: 24px">
      <template #header>
        <div class="card-header">
          <span>我的 API Key</span>
          <el-button
            type="primary"
            size="small"
            :icon="Refresh"
            :loading="loading"
            @click="loadApiKeys"
          >
            刷新
          </el-button>
        </div>
      </template>
      
      <el-table
        v-loading="loading"
        :data="apiKeyList"
        empty-text="暂无 API Key"
        class="api-key-table"
      >
        <el-table-column label="API Key" min-width="200">
          <template #default="{ row }">
            <code class="api-key-code">{{ row.key }}</code>
          </template>
        </el-table-column>
        
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="过期时间" width="180">
          <template #default="{ row }">
            <span v-if="row.expires_at && row.expires_at !== null">{{ formatDate(row.expires_at) }}</span>
            <el-tag v-else type="success" size="small">永不过期</el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.is_expired" type="danger" size="small">已过期</el-tag>
            <el-tag v-else type="success" size="small">有效</el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              type="danger"
              size="small"
              :icon="Delete"
              :loading="deletingIds.includes(row.id)"
              @click="handleDelete(row.id)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, getCurrentInstance, ComponentInternalInstance } from 'vue'
import { generateApiKey, listApiKeys, deleteApiKey } from '@/api/user'
import { copyToClipboard } from '@/utils'
import { formatDate } from '@/utils'
import { CopyDocument, Refresh, Delete, Warning } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'
import type { ApiKeyInfo, GenerateApiKeyResponse } from '@/api/user'

const { proxy } = getCurrentInstance() as ComponentInternalInstance

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
      proxy?.$modal.msgError(result.message || '加载失败')
    }
  } catch (error: any) {
    proxy?.$modal.msgError(error.message || '加载失败')
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
      proxy?.$modal.msgSuccess('API Key 生成成功')
      // 重新加载列表
      await loadApiKeys()
      // 重置表单
      generateForm.expiresDays = 0
    } else {
      proxy?.$modal.msgError(result.message || '生成失败')
    }
  } catch (error: any) {
    proxy?.$modal.msgError(error.message || '生成失败')
  } finally {
    generating.value = false
  }
}

// 删除 API Key
const handleDelete = async (id: number) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除此 API Key 吗？删除后将无法恢复，使用此 Key 的应用将无法继续访问。',
      '删除 API Key',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    deletingIds.value.push(id)
    try {
      const result = await deleteApiKey({ api_key_id: id })
      if (result.code === 200) {
        proxy?.$modal.msgSuccess('删除成功')
        await loadApiKeys()
      } else {
        proxy?.$modal.msgError(result.message || '删除失败')
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || '删除失败')
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
    proxy?.$modal.msgSuccess('已复制到剪贴板')
  } else {
    proxy?.$modal.msgError('复制失败')
  }
}

onMounted(() => {
  loadApiKeys()
})
</script>

<style scoped>
.api-key-manager {
  max-width: 1000px;
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
</style>


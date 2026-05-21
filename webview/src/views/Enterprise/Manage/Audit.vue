<template>
  <div class="enterprise-audit page-card">
    <div class="section-header">
      <h3><el-icon><Document /></el-icon> {{ t('enterprise.audit.title') }}</h3>
      <div class="section-actions">
        <el-input
          v-model="keyword"
          :placeholder="t('enterprise.audit.searchByKeyword')"
          clearable
          style="width: 280px"
          @clear="loadAuditLogs"
          @keyup.enter="loadAuditLogs"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button class="action-btn-secondary" icon="Refresh" @click="loadAuditLogs">{{ t('common.refresh') }}</el-button>
        <el-button type="success" plain icon="Download" @click="handleExport">{{ t('enterprise.audit.export') }}</el-button>
      </div>
    </div>

    <el-table :data="auditList" v-loading="loading" class="data-table styled-table" :empty-text="t('common.noData')">
      <el-table-column prop="user_name" :label="t('enterprise.audit.operator')" width="120" />
      <el-table-column prop="action" :label="t('enterprise.audit.action')" width="150">
        <template #default="{ row }">
          <el-tag :type="getActionTagType(row.action) as any" size="small">{{ row.action }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="target_name" :label="t('enterprise.audit.target')" min-width="150" />
      <el-table-column prop="detail" :label="t('enterprise.audit.detail')" min-width="200" show-overflow-tooltip />
      <el-table-column prop="ip" :label="t('enterprise.audit.ip')" width="130" />
      <el-table-column prop="created_at" :label="t('enterprise.audit.time')" width="180" />
    </el-table>

    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @size-change="loadAuditLogs"
        @current-change="loadAuditLogs"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
  import type { ComponentInternalInstance } from 'vue'
  import { enterpriseApi } from '@myobj/api'
  import type { EnterpriseAuditLog } from '@myobj/shared'
  import { useI18n } from '@/composables'

  const enterpriseId = inject<Ref<string>>('enterpriseId', ref(''))

  const { getEnterpriseAuditLogs, exportEnterpriseAuditLogs } = enterpriseApi
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  const loading = ref(false)
  const auditList = ref<EnterpriseAuditLog[]>([])
  const keyword = ref('')
  const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

  const loadAuditLogs = async () => {
    loading.value = true
    try {
      const res = await getEnterpriseAuditLogs({
        enterprise_id: enterpriseId.value,
        page: pagination.page,
        pageSize: pagination.pageSize,
        keyword: keyword.value || undefined
      })
      if (res.code === 200 && res.data) {
        auditList.value = res.data.list || []
        pagination.total = res.data.total || 0
      }
    } catch (error: any) {
      proxy?.$log?.error(error)
    } finally {
      loading.value = false
    }
  }

  const getActionTagType = (action: string): string => {
    const lower = action.toLowerCase()
    if (lower.includes('create') || lower.includes('add') || lower.includes('invite')) return 'success'
    if (lower.includes('delete') || lower.includes('remove')) return 'danger'
    if (lower.includes('update') || lower.includes('edit') || lower.includes('rename')) return 'warning'
    if (lower.includes('upload')) return 'primary'
    if (lower.includes('download')) return 'info'
    return ''
  }

  const handleExport = async () => {
    try {
      await exportEnterpriseAuditLogs({
        enterprise_id: enterpriseId.value,
        keyword: keyword.value || undefined
      })
      proxy?.$modal.msgSuccess(t('enterprise.audit.exportSuccess'))
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('common.operationFailed'))
    }
  }

  watch(enterpriseId, (id) => {
    auditList.value = []
    if (id) {
      loadAuditLogs()
    }
  }, { immediate: true })
</script>

<style scoped>
  .enterprise-audit {
    display: flex;
    flex-direction: column;
    gap: 16px;
    height: 100%;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
  }

  .section-header h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 700;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .section-header h3 .el-icon {
    color: var(--primary-color);
  }

  .section-actions {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .data-table {
    width: 100%;
  }

  .pagination-wrapper {
    margin-top: 8px;
    padding-top: 12px;
    border-top: 1px solid var(--el-border-color-lighter);
    display: flex;
    justify-content: flex-end;
  }

  @media (max-width: 768px) {
    .section-header {
      flex-direction: column;
      align-items: flex-start;
    }

    .section-actions {
      width: 100%;
      flex-wrap: wrap;
    }

    .pagination-wrapper ::deep(.el-pagination__sizes),
    .pagination-wrapper ::deep(.el-pagination__jump) {
      display: none;
    }
  }
</style>

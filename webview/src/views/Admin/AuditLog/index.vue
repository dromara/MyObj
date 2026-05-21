<template>
  <div class="admin-audit">
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button icon="Refresh" @click="loadAuditLogs">{{ t('admin.audit.refresh') }}</el-button>
        <el-button icon="Download" @click="handleExport">{{ t('admin.audit.export') }}</el-button>
      </div>
      <div class="toolbar-right">
        <el-select
          v-model="filters.action"
          :placeholder="t('admin.audit.allActions')"
          clearable
          style="width: 140px"
          @change="handleFilter"
        >
          <el-option
            v-for="item in actionOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-date-picker
          v-model="dateRange"
          type="datetimerange"
          range-separator="-"
          :start-placeholder="t('admin.audit.startTime')"
          :end-placeholder="t('admin.audit.endTime')"
          value-format="YYYY-MM-DD HH:mm:ss"
          style="width: 360px"
          @change="handleFilter"
        />
        <el-input
          v-model="filters.keyword"
          :placeholder="t('admin.audit.searchPlaceholder')"
          clearable
          style="width: 260px"
          @clear="handleFilter"
          @keyup.enter="handleFilter"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
    </div>

    <el-table :data="logList" v-loading="loading" class="admin-table" :empty-text="t('admin.audit.noLogs')">
      <el-table-column prop="created_at" :label="t('admin.audit.time')" width="180" />
      <el-table-column prop="user_name" :label="t('admin.audit.user')" width="120" />
      <el-table-column :label="t('admin.audit.action')" width="120">
        <template #default="{ row }">
          <el-tag :type="getActionTagType(row.action)" size="small">
            {{ getActionLabel(row.action) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('admin.audit.targetType')" width="100">
        <template #default="{ row }">
          {{ getTargetLabel(row.target_type) }}
        </template>
      </el-table-column>
      <el-table-column prop="target_name" :label="t('admin.audit.targetName')" min-width="160" show-overflow-tooltip />
      <el-table-column prop="target_path" :label="t('admin.audit.targetPath')" min-width="200" show-overflow-tooltip />
      <el-table-column prop="detail" :label="t('admin.audit.detail')" min-width="200" show-overflow-tooltip />
      <el-table-column prop="ip" :label="t('admin.audit.ip')" width="140" />
    </el-table>

    <el-pagination
      v-model:current-page="pagination.page"
      v-model:page-size="pagination.pageSize"
      :total="pagination.total"
      :page-sizes="[20, 50, 100]"
      layout="total, sizes, prev, pager, next, jumper"
      @size-change="loadAuditLogs"
      @current-change="loadAuditLogs"
      class="pagination"
    />
  </div>
</template>

<script setup lang="ts">
  import type { ComponentInternalInstance } from 'vue'
  import { adminApi } from '@myobj/api'
  import type { AuditLogEntry } from '@myobj/shared'
  import { useI18n } from '@/composables'

  const { getAuditLogList, exportAuditLog } = adminApi
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  const loading = ref(false)
  const logList = ref<AuditLogEntry[]>([])
  const dateRange = ref<[string, string] | null>(null)

  const filters = reactive({
    keyword: '',
    action: ''
  })

  const pagination = reactive({
    page: 1,
    pageSize: 20,
    total: 0
  })

  const actionOptions = computed(() => [
    { value: 'upload', label: t('admin.audit.actions.upload') },
    { value: 'download', label: t('admin.audit.actions.download') },
    { value: 'rename', label: t('admin.audit.actions.rename') },
    { value: 'move', label: t('admin.audit.actions.move') },
    { value: 'delete', label: t('admin.audit.actions.delete') },
    { value: 'open', label: t('admin.audit.actions.open') },
    { value: 'mkdir', label: t('admin.audit.actions.mkdir') },
    { value: 'set_public', label: t('admin.audit.actions.set_public') },
    { value: 'extract', label: t('admin.audit.actions.extract') },
    { value: 'package', label: t('admin.audit.actions.package') },
    { value: 'share', label: t('admin.audit.actions.share') },
    { value: 'restore', label: t('admin.audit.actions.restore') },
    { value: 'permanent_delete', label: t('admin.audit.actions.permanent_delete') },
    { value: 'empty_recycle', label: t('admin.audit.actions.empty_recycle') }
  ])

  const getActionLabel = (action: string) => {
    const map: Record<string, string> = {
      upload: t('admin.audit.actions.upload'),
      download: t('admin.audit.actions.download'),
      rename: t('admin.audit.actions.rename'),
      move: t('admin.audit.actions.move'),
      delete: t('admin.audit.actions.delete'),
      open: t('admin.audit.actions.open'),
      mkdir: t('admin.audit.actions.mkdir'),
      set_public: t('admin.audit.actions.set_public'),
      extract: t('admin.audit.actions.extract'),
      package: t('admin.audit.actions.package'),
      share: t('admin.audit.actions.share'),
      restore: t('admin.audit.actions.restore'),
      permanent_delete: t('admin.audit.actions.permanent_delete'),
      empty_recycle: t('admin.audit.actions.empty_recycle')
    }
    return map[action] || action
  }

  const getActionTagType = (action: string) => {
    const map: Record<string, string> = {
      upload: 'success',
      download: '',
      rename: 'warning',
      move: 'warning',
      delete: 'danger',
      mkdir: 'success',
      restore: 'success',
      permanent_delete: 'danger',
      empty_recycle: 'danger'
    }
    return (map[action] || '') as any
  }

  const getTargetLabel = (target: string) => {
    const map: Record<string, string> = {
      file: t('admin.audit.targets.file'),
      dir: t('admin.audit.targets.dir')
    }
    return map[target] || target
  }

  const loadAuditLogs = async () => {
    loading.value = true
    try {
      const params: any = {
        page: pagination.page,
        pageSize: pagination.pageSize
      }
      if (filters.keyword) params.keyword = filters.keyword
      if (filters.action) params.action = filters.action
      if (dateRange.value) {
        params.start_time = dateRange.value[0]
        params.end_time = dateRange.value[1]
      }

      const res = await getAuditLogList(params)
      if (res.code === 200 && res.data) {
        logList.value = res.data.list || []
        pagination.total = res.data.total || 0
      } else {
        logList.value = []
        pagination.total = 0
      }
    } catch (error: any) {
      proxy?.$modal.msgError(t('admin.audit.loadFailed'))
      proxy?.$log?.error(error)
    } finally {
      loading.value = false
    }
  }

  const handleFilter = () => {
    pagination.page = 1
    loadAuditLogs()
  }

  const handleExport = async () => {
    try {
      const params: any = {}
      if (filters.keyword) params.keyword = filters.keyword
      if (filters.action) params.action = filters.action
      if (dateRange.value) {
        params.start_time = dateRange.value[0]
        params.end_time = dateRange.value[1]
      }
      await exportAuditLog(params)
      proxy?.$modal.msgSuccess(t('admin.audit.exportSuccess'))
    } catch (error: any) {
      proxy?.$modal.msgError(t('admin.audit.exportFailed'))
      proxy?.$log?.error(error)
    }
  }

  onMounted(() => {
    loadAuditLogs()
  })
</script>

<style scoped>
  .admin-audit {
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 16px;
  }

  .toolbar-left {
    display: flex;
    gap: 12px;
  }

  .toolbar-right {
    display: flex;
    gap: 12px;
    align-items: center;
    flex-wrap: wrap;
  }

  .admin-table {
    flex: 1;
    overflow: auto;
  }

  .pagination {
    margin-top: 16px;
    justify-content: flex-end;
  }

  /* 移动端适配 */
  @media (max-width: 768px) {
    .toolbar {
      flex-direction: column;
      align-items: stretch;
      gap: 12px;
    }

    .toolbar-left {
      flex-wrap: wrap;
    }

    .toolbar-right {
      width: 100%;
    }

    .toolbar-right .el-input,
    .toolbar-right .el-select,
    .toolbar-right .el-date-picker {
      width: 100% !important;
    }

    .admin-table {
      font-size: 12px;
    }

    .admin-table ::deep(.el-table__cell) {
      padding: 8px 4px;
    }

    .pagination {
      justify-content: center;
    }

    .pagination ::deep(.el-pagination__sizes),
    .pagination ::deep(.el-pagination__jump) {
      display: none;
    }
  }

  @media (max-width: 480px) {
    .toolbar-left .el-button {
      flex: 1;
      min-width: 0;
    }

    .admin-table ::deep(.el-table__cell) {
      padding: 6px 2px;
      font-size: 11px;
    }
  }

  /* 深色模式样式 */
  html.dark .admin-audit {
    background: transparent;
  }

  html.dark .pagination {
    border-top-color: var(--el-border-color);
  }
</style>

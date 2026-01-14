<template>
  <div class="table-header-operation">
    <div class="operation-left">
      <slot name="prefix" />
      <el-button v-if="showAdd" type="primary" icon="Plus" @click="$emit('add')">
        {{ t('common.add') }}
      </el-button>
      <el-button
        v-if="showBatchDelete && checkedCount > 0"
        type="danger"
        icon="Delete"
        @click="$emit('batch-delete')"
        :loading="batchDeleting"
      >
        {{ t('common.batchDelete') }}
      </el-button>
      <el-button v-if="showExport" icon="Download" @click="$emit('export')">
        {{ t('common.export') }}
      </el-button>
      <slot />
    </div>
    <div class="operation-right">
      <el-button v-if="showRefresh" icon="Refresh" circle @click="$emit('refresh')" :loading="refreshing" />
      <el-button v-if="showColumnSetting" icon="Setting" circle @click="$emit('column-setting')" />
      <slot name="after" />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from '@/composables'

  const { t } = useI18n()

  interface Props {
    /** 显示添加按钮 */
    showAdd?: boolean
    /** 显示批量删除按钮 */
    showBatchDelete?: boolean
    /** 显示导出按钮 */
    showExport?: boolean
    /** 显示刷新按钮 */
    showRefresh?: boolean
    /** 显示列设置按钮 */
    showColumnSetting?: boolean
    /** 选中数量 */
    checkedCount?: number
    /** 批量删除加载状态 */
    batchDeleting?: boolean
    /** 刷新加载状态 */
    refreshing?: boolean
  }

  withDefaults(defineProps<Props>(), {
    showAdd: false,
    showBatchDelete: false,
    showExport: false,
    showRefresh: true,
    showColumnSetting: true,
    checkedCount: 0,
    batchDeleting: false,
    refreshing: false
  })

  defineEmits<{
    add: []
    'batch-delete': []
    export: []
    refresh: []
    'column-setting': []
  }>()
</script>

<style scoped>
  .table-header-operation {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    gap: 12px;
  }

  .operation-left {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .operation-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  /* 优化按钮交互效果 */
  .operation-left .el-button,
  .operation-right .el-button {
    transition: all 0.2s ease;
  }

  .operation-left .el-button:hover:not(.is-disabled):not(.is-loading),
  .operation-right .el-button:hover:not(.is-disabled):not(.is-loading) {
    transform: translateY(-1px);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }

  .operation-right .el-button.is-circle:hover:not(.is-disabled):not(.is-loading) {
    background: var(--el-fill-color-light);
    color: var(--primary-color);
  }

  html.dark .operation-left .el-button:hover:not(.is-disabled):not(.is-loading),
  html.dark .operation-right .el-button:hover:not(.is-disabled):not(.is-loading) {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  }

  @media (max-width: 768px) {
    .table-header-operation {
      flex-direction: column;
      align-items: stretch;
    }

    .operation-left,
    .operation-right {
      width: 100%;
      justify-content: space-between;
    }
  }
</style>

<template>
  <div v-if="checkedCount > 0" class="table-row-check-alert">
    <div class="alert-content">
      <span class="checked-text">
        {{ t('common.selectedCount', { count: checkedCount }) }}
      </span>
      <el-button link type="primary" size="small" @click="$emit('clear')">
        {{ t('common.clearSelection') }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from '@/composables'

  const { t } = useI18n()

  interface Props {
    checkedCount: number
  }

  defineProps<Props>()

  defineEmits<{
    clear: []
  }>()
</script>

<style scoped>
  .table-row-check-alert {
    margin-bottom: 16px;
    padding: 12px 16px;
    background: linear-gradient(135deg, rgba(37, 99, 235, 0.08) 0%, rgba(79, 70, 229, 0.08) 100%);
    border: 1px solid rgba(37, 99, 235, 0.2);
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(37, 99, 235, 0.1);
    animation: slideDown 0.3s ease-out;
    transition: all 0.2s ease;
  }

  .table-row-check-alert:hover {
    box-shadow: 0 4px 12px rgba(37, 99, 235, 0.15);
    transform: translateY(-1px);
  }

  html.dark .table-row-check-alert {
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.15) 0%, rgba(99, 102, 241, 0.15) 100%);
    border-color: rgba(59, 130, 246, 0.3);
    box-shadow: 0 2px 8px rgba(59, 130, 246, 0.2);
  }

  html.dark .table-row-check-alert:hover {
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
  }

  @keyframes slideDown {
    from {
      opacity: 0;
      transform: translateY(-10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .alert-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .checked-text {
    font-size: 14px;
    color: var(--el-color-primary);
    font-weight: 500;
  }

  @media (max-width: 768px) {
    .table-row-check-alert {
      padding: 10px 12px;
    }

    .checked-text {
      font-size: 13px;
    }
  }
</style>

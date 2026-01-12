<template>
  <div v-if="checkedCount > 0" class="table-row-check-alert">
    <div class="alert-content">
      <span class="checked-text">
        {{ t('common.selectedCount', { count: checkedCount }) }}
      </span>
      <el-button
        link
        type="primary"
        size="small"
        @click="$emit('clear')"
      >
        {{ t('common.clearSelection') }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'

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
  background: var(--el-color-primary-light-9);
  border: 1px solid var(--el-color-primary-light-7);
  border-radius: 8px;
}

html.dark .table-row-check-alert {
  background: rgba(64, 158, 255, 0.1);
  border-color: rgba(64, 158, 255, 0.2);
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

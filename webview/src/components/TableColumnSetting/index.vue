<template>
  <el-drawer v-model="visible" :title="t('common.columnSetting')" size="400px" direction="rtl">
    <div class="column-setting-content">
      <div class="setting-header">
        <el-button link type="primary" @click="handleReset">
          {{ t('common.reset') }}
        </el-button>
      </div>

      <el-divider />

      <div class="column-list">
        <div v-for="check in columnChecks" :key="check.key" class="column-item">
          <el-checkbox v-model="check.checked" :disabled="!check.visible">
            {{ check.title }}
          </el-checkbox>
        </div>
      </div>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
  import type { TableCheck } from '@/composables'
  import { useI18n } from '@/composables'

  const { t } = useI18n()

  interface Props {
    modelValue: boolean
    columnChecks: TableCheck[]
  }

  const props = defineProps<Props>()

  const emit = defineEmits<{
    'update:modelValue': [value: boolean]
    change: [checks: TableCheck[]]
  }>()

  const visible = computed({
    get: () => props.modelValue,
    set: value => emit('update:modelValue', value)
  })

  const localChecks = ref<TableCheck[]>([])

  watch(
    () => props.columnChecks,
    newChecks => {
      localChecks.value = newChecks.map(check => ({ ...check }))
    },
    { immediate: true, deep: true }
  )

  watch(
    localChecks,
    newChecks => {
      emit('change', newChecks)
    },
    { deep: true }
  )

  const handleReset = () => {
    localChecks.value = localChecks.value.map(check => ({
      ...check,
      checked: check.visible !== false
    }))
  }
</script>

<style scoped>
  .column-setting-content {
    padding: 16px;
  }

  .setting-header {
    display: flex;
    justify-content: flex-end;
    margin-bottom: 8px;
  }

  .column-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .column-item {
    display: flex;
    align-items: center;
    padding: 10px 12px;
    border-radius: 6px;
    transition: all 0.2s ease;
    cursor: pointer;
  }

  .column-item:hover {
    background-color: var(--el-fill-color-light);
    transform: translateX(4px);
  }

  html.dark .column-item:hover {
    background-color: var(--el-fill-color-light);
  }

  .column-item :deep(.el-checkbox) {
    width: 100%;
  }

  .column-item :deep(.el-checkbox__label) {
    flex: 1;
  }
</style>

<template>
  <div class="breadcrumb-container glass-panel-sm">
    <el-breadcrumb separator="/">
      <el-breadcrumb-item 
        v-for="item in breadcrumbs" 
        :key="item.id"
        @click="$emit('navigate', item.path)"
        class="breadcrumb-item"
      >
        {{ formatBreadcrumbName(item.name) }}
      </el-breadcrumb-item>
    </el-breadcrumb>
  </div>
</template>

<script setup lang="ts">
import type { Breadcrumb } from '@/types'

defineProps<{
  breadcrumbs: Breadcrumb[]
  formatBreadcrumbName: (name: string) => string
}>()

defineEmits<{
  navigate: [path: string]
}>()
</script>

<style scoped>
.breadcrumb-container {
  margin-bottom: 20px;
  padding: 12px 20px;
  border-radius: 12px;
  transition: all 0.3s;
}

.breadcrumb-item {
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}
</style>


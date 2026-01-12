<template>
  <div :class="{ hidden: hidden }" class="pagination-container" :data-float="float">
    <el-pagination
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :background="background"
      :layout="layout"
      :page-sizes="pageSizes"
      :pager-count="pagerCount"
      :total="total"
      @size-change="handleSizeChange"
      @current-change="handleCurrentChange"
    />
  </div>
</template>

<script setup lang="ts">
import { useResponsive } from '@/composables/useResponsive'
import { scrollTo } from '@/utils/scroll-to'

interface Props {
  total: number
  page?: number
  limit?: number
  pageSizes?: number[]
  pagerCount?: number
  layout?: string
  background?: boolean
  autoScroll?: boolean
  hidden?: boolean
  float?: 'left' | 'right' | 'center' | 'flex-start' | 'flex-end'
}

const props = withDefaults(defineProps<Props>(), {
  page: 1,
  limit: 20,
  pageSizes: () => [20, 50, 100],
  layout: 'total, sizes, prev, pager, next, jumper',
  background: true,
  autoScroll: true,
  hidden: false,
  float: 'center'
})

const emit = defineEmits<{
  'update:page': [page: number]
  'update:limit': [limit: number]
  'pagination': [data: { page: number; limit: number }]
}>()

// 使用响应式检测 composable
const { isMobile } = useResponsive()

// 动态 pagerCount：移动端5，PC端7
const pagerCount = computed(() => {
  if (props.pagerCount !== undefined) {
    return props.pagerCount
  }
  return isMobile.value ? 5 : 7
})

const currentPage = computed({
  get() {
    return props.page
  },
  set(val: number) {
    emit('update:page', val)
  }
})

const pageSize = computed({
  get() {
    return props.limit
  },
  set(val: number) {
    emit('update:limit', val)
  }
})

function handleSizeChange(val: number) {
  // 如果当前页超出范围，重置到第一页
  if (currentPage.value * val > props.total) {
    currentPage.value = 1
  }
  emit('pagination', { page: currentPage.value, limit: val })
  if (props.autoScroll) {
    scrollTo(0, 300)
  }
}

function handleCurrentChange(val: number) {
  emit('pagination', { page: val, limit: pageSize.value })
  if (props.autoScroll) {
    scrollTo(0, 300)
  }
}
</script>

<style scoped>
.pagination-container {
  display: flex;
  width: 100%;
}

.pagination-container[data-float="left"],
.pagination-container[data-float="flex-start"] {
  justify-content: flex-start;
}

.pagination-container[data-float="right"],
.pagination-container[data-float="flex-end"] {
  justify-content: flex-end;
}

.pagination-container[data-float="center"] {
  justify-content: center;
}

.pagination-container.hidden {
  display: none;
}

/* 移动端优化 */
@media (max-width: 1024px) {
  .pagination-container {
    :deep(.el-pagination) {
      justify-content: center;
      flex-wrap: wrap;
    }

    /* 移动端隐藏部分元素，简化显示 */
    :deep(.el-pagination__total) {
      display: none;
    }

    :deep(.el-pagination__sizes) {
      margin-right: 0;
    }

    :deep(.el-pagination__jump) {
      display: none;
    }
  }
}

@media (max-width: 768px) {
  .pagination-container {
    :deep(.el-pagination) {
      padding: 8px 0;
    }

    /* 进一步简化移动端显示 */
    :deep(.el-pagination__sizes) {
      display: none;
    }
  }
}

/* 深色模式样式 */
html.dark .pagination-container :deep(.el-pagination) {
  color: var(--el-text-color-primary);
}

html.dark .pagination-container :deep(.el-pagination__total) {
  color: var(--el-text-color-primary);
}

html.dark .pagination-container :deep(.el-pagination button) {
  background-color: var(--el-bg-color);
  color: var(--el-text-color-primary);
  border-color: var(--el-border-color);
}

html.dark .pagination-container :deep(.el-pagination button:hover) {
  color: var(--primary-color);
}

html.dark .pagination-container :deep(.el-pagination button.is-active) {
  background-color: var(--primary-color);
  color: var(--el-text-color-primary);
  border-color: var(--primary-color);
}

html.dark .pagination-container :deep(.el-pagination .el-pager li) {
  background-color: var(--el-bg-color);
  color: var(--el-text-color-primary);
  border-color: var(--el-border-color);
}

html.dark .pagination-container :deep(.el-pagination .el-pager li:hover) {
  color: var(--primary-color);
}

html.dark .pagination-container :deep(.el-pagination .el-pager li.is-active) {
  background-color: var(--primary-color);
  color: var(--el-text-color-primary);
  border-color: var(--primary-color);
}

html.dark .pagination-container :deep(.el-pagination .el-select) {
  background-color: var(--el-bg-color);
}

html.dark .pagination-container :deep(.el-pagination .el-select .el-input__wrapper) {
  background-color: var(--el-bg-color);
  border-color: var(--el-border-color);
}

html.dark .pagination-container :deep(.el-pagination .el-select .el-input__inner) {
  background-color: var(--el-bg-color);
  color: var(--el-text-color-primary);
  border-color: var(--el-border-color);
}

html.dark .pagination-container :deep(.el-pagination .el-select:hover .el-input__wrapper) {
  border-color: var(--primary-color);
}

html.dark .pagination-container :deep(.el-pagination .el-select.is-focus .el-input__wrapper) {
  border-color: var(--primary-color);
}

html.dark .pagination-container :deep(.el-pagination .el-input__inner) {
  background-color: var(--el-bg-color);
  color: var(--el-text-color-primary);
  border-color: var(--el-border-color);
}

/* 选择框下拉菜单深色模式样式 */
html.dark .pagination-container :deep(.el-select-dropdown) {
  background-color: var(--el-bg-color);
  border-color: var(--el-border-color);
}

html.dark .pagination-container :deep(.el-select-dropdown__item) {
  background-color: var(--el-bg-color);
  color: var(--el-text-color-primary);
}

html.dark .pagination-container :deep(.el-select-dropdown__item:hover) {
  background-color: var(--el-fill-color-light);
  color: var(--primary-color);
}

html.dark .pagination-container :deep(.el-select-dropdown__item.selected) {
  background-color: var(--el-fill-color-light);
  color: var(--primary-color);
  font-weight: 600;
}

html.dark .pagination-container :deep(.el-select-dropdown__item.selected::after) {
  color: var(--primary-color);
}
</style>


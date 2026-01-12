# 表格管理 Composables 使用指南

本文档介绍如何使用表格管理的 composables 和组件，实现统一的表格功能。

## 核心 Composables

### 1. useTable - 基础表格管理

提供数据、加载状态、分页等基础功能。

```typescript
import { useTable, type TableColumn } from '@/composables/useTable'

const columns = ref<TableColumn[]>([
  { key: 'id', label: 'ID', width: 100 },
  { key: 'name', label: '名称', minWidth: 200 },
  { key: 'status', label: '状态', width: 120 }
])

const { data, loading, total, currentPage, pageSize, visibleColumns, refresh } = useTable({
  getData: async () => {
    // 加载数据的逻辑
    const res = await fetchData({ page: currentPage.value, pageSize: pageSize.value })
    data.value = res.list
    total.value = res.total
  },
  columns,
  pagination: true
})
```

### 2. useTableOperate - 表格操作

提供增删改查、批量操作等功能。

```typescript
import { useTableOperate } from '@/composables/useTableOperate'

const {
  checkedRowKeys,
  checkedRows,
  handleAdd,
  handleEdit,
  handleDelete,
  handleBatchDelete,
  clearSelection
} = useTableOperate({
  data,
  getId: (item) => item.id,
  onAdd: () => {
    // 添加逻辑
  },
  onEdit: (id) => {
    // 编辑逻辑
  },
  onDelete: async (id) => {
    // 删除逻辑
    await deleteItem(id)
    await refresh()
  },
  onBatchDelete: async (ids) => {
    // 批量删除逻辑
    await batchDelete(ids)
    await refresh()
  }
})
```

### 3. useTableColumn - 列管理

提供列的显示/隐藏、排序、持久化等功能。

```typescript
import { useTableColumn } from '@/composables/useTableColumn'

const { applyColumnSettings, updateColumnChecks } = useTableColumn()

// 应用保存的列设置
applyColumnSettings(columns, 'shares-table')

// 监听列设置变化并保存
updateColumnChecks(columnChecks, 'shares-table')
```

## 组件

### 1. TableHeaderOperation - 表格头部操作栏

```vue
<TableHeaderOperation
  :show-add="true"
  :show-batch-delete="true"
  :show-export="true"
  :checked-count="checkedRowKeys.length"
  @add="handleAdd"
  @batch-delete="handleBatchDelete"
  @export="handleExport"
  @refresh="refresh"
  @column-setting="showColumnSetting = true"
/>
```

### 2. TableColumnSetting - 列设置弹窗

```vue
<TableColumnSetting
  v-model="showColumnSetting"
  :column-checks="columnChecks"
  @change="handleColumnChange"
/>
```

### 3. TableRowCheckAlert - 选中行提示

```vue
<TableRowCheckAlert
  :checked-count="checkedRowKeys.length"
  @clear="clearSelection"
/>
```

## 完整示例

```vue
<template>
  <div class="page">
    <!-- 表格头部操作栏 -->
    <TableHeaderOperation
      :show-add="true"
      :show-batch-delete="true"
      :checked-count="checkedRowKeys.length"
      @add="handleAdd"
      @batch-delete="handleBatchDelete"
      @refresh="refresh"
      @column-setting="showColumnSetting = true"
    />

    <!-- 选中行提示 -->
    <TableRowCheckAlert
      :checked-count="checkedRowKeys.length"
      @clear="clearSelection"
    />

    <!-- 表格 -->
    <el-table
      :data="data"
      v-loading="loading"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="55" />
      <el-table-column
        v-for="col in visibleColumns"
        :key="col.key"
        :prop="col.key"
        :label="col.label"
        :width="col.width"
        :min-width="col.minWidth"
      />
    </el-table>

    <!-- 分页 -->
    <el-pagination
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :total="total"
      layout="total, sizes, prev, pager, next, jumper"
    />

    <!-- 列设置 -->
    <TableColumnSetting
      v-model="showColumnSetting"
      :column-checks="columnChecks"
      @change="handleColumnChange"
    />
  </div>
</template>

<script setup lang="ts">
import { useTable, useTableOperate, useTableColumn } from '@/composables'
// 注意：TableHeaderOperation, TableColumnSetting, TableRowCheckAlert 等组件已全局注册，无需显式导入

const columns = ref([...])
const showColumnSetting = ref(false)

const { data, loading, total, currentPage, pageSize, columnChecks, visibleColumns, refresh } = useTable({
  getData: loadData,
  columns,
  pagination: true
})

const {
  checkedRowKeys,
  handleAdd,
  handleDelete,
  handleBatchDelete,
  clearSelection,
  handleSelectionChange
} = useTableOperate({
  data,
  getId: (item) => item.id,
  onDelete: async (id) => {
    await deleteItem(id)
    await refresh()
  },
  onBatchDelete: async (ids) => {
    await batchDelete(ids)
    await refresh()
  }
})

const { updateColumnChecks } = useTableColumn()
updateColumnChecks(columnChecks, 'table-key')

const handleColumnChange = (checks) => {
  columnChecks.value = checks
}
</script>
```

## 特性

- ✅ 组合式 API + 自定义 Hook：逻辑复用
- ✅ 响应式设计：移动端适配
- ✅ 列管理：显示/隐藏、排序、持久化
- ✅ 统一操作模式：增删改查标准化
- ✅ 类型安全：完整的 TypeScript 支持
- ✅ 国际化：使用 `t()` 统一翻译

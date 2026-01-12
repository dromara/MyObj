import type { Ref } from 'vue'

export interface UseTableOperateOptions<T> {
  /** 数据列表 */
  data: Ref<T[]>
  /** 获取 ID 的函数 */
  getId: (item: T) => string | number
  /** 添加操作 */
  onAdd?: () => void
  /** 编辑操作 */
  onEdit?: (id: string | number) => void
  /** 删除操作 */
  onDelete?: (id: string | number) => Promise<void>
  /** 批量删除操作 */
  onBatchDelete?: (ids: (string | number)[]) => Promise<void>
}

/**
 * 表格操作 Composable
 * 提供增删改查、批量操作等功能
 */
export function useTableOperate<T = any>(options: UseTableOperateOptions<T>) {
  const { data, getId, onAdd, onEdit, onDelete, onBatchDelete } = options

  // 选中行
  const checkedRowKeys = ref<(string | number)[]>([])
  const checkedRows = computed(() => {
    return data.value.filter(item => checkedRowKeys.value.includes(getId(item)))
  })

  // 抽屉/对话框状态
  const drawerVisible = ref(false)
  const operateType = ref<'add' | 'edit'>('add')
  const editingData = ref<T | null>(null)

  /**
   * 添加
   */
  const handleAdd = () => {
    operateType.value = 'add'
    editingData.value = null
    drawerVisible.value = true
    if (onAdd) {
      onAdd()
    }
  }

  /**
   * 编辑
   */
  const handleEdit = (id: string | number) => {
    const item = data.value.find(d => getId(d) === id)
    if (item) {
      operateType.value = 'edit'
      editingData.value = item
      drawerVisible.value = true
      if (onEdit) {
        onEdit(id)
      }
    }
  }

  /**
   * 删除
   */
  const handleDelete = async (id: string | number) => {
    if (onDelete) {
      await onDelete(id)
      // 从选中列表中移除
      checkedRowKeys.value = checkedRowKeys.value.filter(key => key !== id)
    }
  }

  /**
   * 批量删除
   */
  const handleBatchDelete = async () => {
    if (onBatchDelete && checkedRowKeys.value.length > 0) {
      await onBatchDelete(checkedRowKeys.value)
      checkedRowKeys.value = []
    }
  }

  /**
   * 清空选中
   */
  const clearSelection = () => {
    checkedRowKeys.value = []
  }

  /**
   * 选中变化
   */
  const handleSelectionChange = (keys: (string | number)[]) => {
    checkedRowKeys.value = keys
  }

  /**
   * 删除后回调
   */
  const onDeleted = () => {
    clearSelection()
  }

  /**
   * 批量删除后回调
   */
  const onBatchDeleted = () => {
    clearSelection()
  }

  return {
    // 状态
    checkedRowKeys,
    checkedRows,
    drawerVisible,
    operateType,
    editingData,
    // 方法
    handleAdd,
    handleEdit,
    handleDelete,
    handleBatchDelete,
    clearSelection,
    handleSelectionChange,
    onDeleted,
    onBatchDeleted
  }
}

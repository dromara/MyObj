import type { Ref } from 'vue'
import cache from '@/plugins/cache'

// ==================== 类型定义 ====================

export interface TableColumn {
  key: string
  label: string
  width?: number | string
  minWidth?: number | string
  align?: 'left' | 'center' | 'right'
  fixed?: boolean | 'left' | 'right'
  visible?: boolean
  order?: number
  className?: string
  [key: string]: any
}

export interface TableCheck {
  key: string
  title: string
  checked: boolean
  visible: boolean
}

export interface UseTableOptions {
  /** 获取数据的函数 */
  getData: () => Promise<void>
  /** 列配置 */
  columns: Ref<TableColumn[]>
  /** 是否启用分页 */
  pagination?: boolean
  /** 初始页码 */
  initialPage?: number
  /** 初始每页数量 */
  initialPageSize?: number
  /** 列可见性检查函数 */
  getColumnVisible?: (column: TableColumn) => boolean
}

/**
 * 基础表格管理 Composable
 * 提供数据、加载状态、分页等基础功能
 */
export function useTable<T = any>(options: UseTableOptions) {
  const { getData, columns, initialPage = 1, initialPageSize = 10, getColumnVisible } = options

  // 数据状态
  const data = ref<T[]>([]) as Ref<T[]>
  const loading = ref(false)
  const total = ref(0)

  // 分页状态
  const currentPage = ref(initialPage)
  const pageSize = ref(initialPageSize)

  // 列管理
  const columnChecks = ref<TableCheck[]>([])

  /**
   * 初始化列检查项
   */
  const initColumnChecks = () => {
    columnChecks.value = columns.value
      .filter(col => col.key)
      .map(col => ({
        key: col.key,
        title: col.label,
        checked: col.visible !== false,
        visible: getColumnVisible ? getColumnVisible(col) : true
      }))
  }

  /**
   * 获取可见的列
   */
  const visibleColumns = computed(() => {
    const checksMap = new Map(columnChecks.value.map(check => [check.key, check.checked]))
    return columns.value.filter(col => {
      if (!col.key) return true
      return checksMap.get(col.key) !== false
    })
  })

  /**
   * 重新加载列配置
   */
  const reloadColumns = () => {
    initColumnChecks()
  }

  /**
   * 加载数据
   */
  const loadData = async () => {
    loading.value = true
    try {
      await getData()
    } finally {
      loading.value = false
    }
  }

  /**
   * 刷新数据
   */
  const refresh = () => {
    return loadData()
  }

  // 初始化
  onMounted(() => {
    initColumnChecks()
  })

  return {
    // 数据
    data,
    loading,
    total,
    // 分页
    currentPage,
    pageSize,
    // 列管理
    columns,
    columnChecks,
    visibleColumns,
    // 方法
    loadData,
    refresh,
    reloadColumns
  }
}

// ==================== 列管理功能 ====================

const COLUMN_SETTINGS_KEY = 'table-column-settings'

/**
 * 列管理 Composable
 * 提供列的显示/隐藏、排序、持久化等功能
 */
export function useTableColumn() {
  /**
   * 保存列设置到 localStorage
   */
  const saveColumnSettings = (tableKey: string, checks: TableCheck[]) => {
    try {
      const settings = cache.local.getJSON(COLUMN_SETTINGS_KEY) || {}
      settings[tableKey] = checks.map(check => ({
        key: check.key,
        checked: check.checked,
        visible: check.visible
      }))
      cache.local.setJSON(COLUMN_SETTINGS_KEY, settings)
    } catch (error) {
      console.error('保存列设置失败:', error)
    }
  }

  /**
   * 从 localStorage 加载列设置
   */
  const loadColumnSettings = (tableKey: string): Partial<Record<string, boolean>> | null => {
    try {
      const settings = cache.local.getJSON(COLUMN_SETTINGS_KEY)
      if (settings && settings[tableKey]) {
        const result: Record<string, boolean> = {}
        settings[tableKey].forEach((item: { key: string; checked: boolean }) => {
          result[item.key] = item.checked
        })
        return result
      }
    } catch (error) {
      console.error('加载列设置失败:', error)
    }
    return null
  }

  /**
   * 应用列设置到列配置
   */
  const applyColumnSettings = (columns: Ref<TableColumn[]>, tableKey: string) => {
    const settings = loadColumnSettings(tableKey)
    if (settings) {
      columns.value = columns.value.map(col => {
        if (col.key && settings[col.key] !== undefined) {
          return { ...col, visible: settings[col.key] }
        }
        return col
      })
    }
  }

  /**
   * 更新列检查项
   */
  const updateColumnChecks = (checks: Ref<TableCheck[]>, tableKey: string) => {
    watch(
      checks,
      newChecks => {
        saveColumnSettings(tableKey, newChecks)
      },
      { deep: true }
    )
  }

  /**
   * 重置列设置
   */
  const resetColumnSettings = (tableKey: string) => {
    try {
      const settings = cache.local.getJSON(COLUMN_SETTINGS_KEY) || {}
      delete settings[tableKey]
      cache.local.setJSON(COLUMN_SETTINGS_KEY, settings)
    } catch (error) {
      console.error('重置列设置失败:', error)
    }
  }

  return {
    saveColumnSettings,
    loadColumnSettings,
    applyColumnSettings,
    updateColumnChecks,
    resetColumnSettings
  }
}

// ==================== 过滤功能 ====================

export interface TableFilterConfig {
  [key: string]: any
}

export interface UseTableFilterOptions<T> {
  /** 数据列表 */
  data: Ref<T[]>
  /** 默认过滤配置 */
  defaultFilters?: TableFilterConfig
  /** 自定义过滤函数 */
  customFilter?: (item: T, filters: TableFilterConfig) => boolean
}

/**
 * 表格过滤 Composable
 * 提供过滤功能
 */
export function useTableFilter<T = any>(options: UseTableFilterOptions<T>) {
  const { data, defaultFilters = {}, customFilter } = options

  // 过滤状态
  const filters = ref<TableFilterConfig>({ ...defaultFilters })

  /**
   * 设置过滤条件
   */
  const setFilter = (key: string, value: any) => {
    filters.value[key] = value
  }

  /**
   * 清除过滤条件
   */
  const clearFilter = (key?: string) => {
    if (key) {
      delete filters.value[key]
    } else {
      filters.value = {}
    }
  }

  /**
   * 重置过滤条件
   */
  const resetFilters = () => {
    filters.value = { ...defaultFilters }
  }

  /**
   * 过滤数据
   */
  const filteredData = computed(() => {
    if (Object.keys(filters.value).length === 0) {
      return data.value
    }

    if (customFilter) {
      return data.value.filter(item => customFilter(item, filters.value))
    }

    // 默认过滤逻辑：字符串包含匹配
    return data.value.filter(item => {
      return Object.entries(filters.value).every(([key, value]) => {
        if (value === null || value === undefined || value === '') {
          return true
        }

        const itemValue = (item as any)[key]
        if (itemValue === null || itemValue === undefined) {
          return false
        }

        // 字符串匹配
        if (typeof value === 'string' && typeof itemValue === 'string') {
          return itemValue.toLowerCase().includes(value.toLowerCase())
        }

        // 精确匹配
        return itemValue === value
      })
    })
  })

  return {
    filters,
    filteredData,
    setFilter,
    clearFilter,
    resetFilters
  }
}

// ==================== 排序功能 ====================

export interface TableSortConfig {
  prop: string
  order: 'ascending' | 'descending' | null
}

export interface UseTableSortOptions<T> {
  /** 数据列表 */
  data: Ref<T[]>
  /** 默认排序配置 */
  defaultSort?: TableSortConfig
}

/**
 * 表格排序 Composable
 * 提供排序功能
 */
export function useTableSort<T = any>(options: UseTableSortOptions<T>) {
  const { data, defaultSort } = options

  // 排序状态
  const sortConfig = ref<TableSortConfig | null>(defaultSort || null)

  /**
   * 处理排序变化
   */
  const handleSortChange = (config: { prop: string; order: string }) => {
    if (config.order) {
      sortConfig.value = {
        prop: config.prop,
        order: config.order as 'ascending' | 'descending'
      }
    } else {
      sortConfig.value = null
    }
  }

  /**
   * 排序数据
   */
  const sortedData = computed(() => {
    if (!sortConfig.value || !sortConfig.value.order) {
      return data.value
    }

    const { prop, order } = sortConfig.value
    const sorted = [...data.value]

    sorted.sort((a, b) => {
      const aVal = (a as any)[prop]
      const bVal = (b as any)[prop]

      if (aVal === bVal) return 0

      let result = 0
      if (aVal < bVal) {
        result = -1
      } else if (aVal > bVal) {
        result = 1
      }

      return order === 'ascending' ? result : -result
    })

    return sorted
  })

  /**
   * 清除排序
   */
  const clearSort = () => {
    sortConfig.value = null
  }

  return {
    sortConfig,
    sortedData,
    handleSortChange,
    clearSort
  }
}

// ==================== 操作功能 ====================

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

import type { Ref } from 'vue'

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
  const {
    getData,
    columns,
    initialPage = 1,
    initialPageSize = 10,
    getColumnVisible
  } = options

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

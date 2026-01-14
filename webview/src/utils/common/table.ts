/**
 * 表格工具函数
 */

/**
 * 格式化表格数据（用于导出等场景）
 */
export function formatTableData<T extends Record<string, any>>(
  data: T[],
  columns: Array<{ key: string; label: string }>
): Array<Record<string, any>> {
  return data.map(item => {
    const formatted: Record<string, any> = {}
    columns.forEach(col => {
      formatted[col.label] = item[col.key] ?? ''
    })
    return formatted
  })
}

/**
 * 导出表格数据为 CSV
 */
export function exportTableToCSV<T extends Record<string, any>>(
  data: T[],
  columns: Array<{ key: string; label: string }>,
  filename = 'export.csv'
): void {
  const formatted = formatTableData(data, columns)
  const headers = columns.map(col => col.label).join(',')
  const rows = formatted.map(row =>
    Object.values(row)
      .map(val => {
        // 处理包含逗号、引号或换行符的值
        if (typeof val === 'string' && (val.includes(',') || val.includes('"') || val.includes('\n'))) {
          return `"${val.replace(/"/g, '""')}"`
        }
        return val ?? ''
      })
      .join(',')
  )

  const csv = [headers, ...rows].join('\n')
  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  const url = URL.createObjectURL(blob)

  link.setAttribute('href', url)
  link.setAttribute('download', filename)
  link.style.visibility = 'hidden'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

/**
 * 导出表格数据为 Excel（使用 CSV 格式，可被 Excel 打开）
 */
export function exportTableToExcel<T extends Record<string, any>>(
  data: T[],
  columns: Array<{ key: string; label: string }>,
  filename = 'export.xlsx'
): void {
  exportTableToCSV(data, columns, filename.replace('.xlsx', '.csv'))
}

/**
 * 表格数据排序
 */
export function sortTableData<T extends Record<string, any>>(
  data: T[],
  sortKey: string,
  order: 'asc' | 'desc' = 'asc'
): T[] {
  return [...data].sort((a, b) => {
    const aVal = a[sortKey]
    const bVal = b[sortKey]

    if (aVal === bVal) return 0

    const comparison = aVal > bVal ? 1 : -1
    return order === 'asc' ? comparison : -comparison
  })
}

/**
 * 表格数据过滤
 */
export function filterTableData<T extends Record<string, any>>(data: T[], filters: Record<string, any>): T[] {
  return data.filter(item => {
    return Object.entries(filters).every(([key, value]) => {
      if (value === null || value === undefined || value === '') return true

      const itemValue = item[key]
      if (typeof value === 'string') {
        return String(itemValue).toLowerCase().includes(value.toLowerCase())
      }

      return itemValue === value
    })
  })
}

/**
 * 表格数据搜索
 */
export function searchTableData<T extends Record<string, any>>(data: T[], keyword: string, searchKeys: string[]): T[] {
  if (!keyword) return data

  const lowerKeyword = keyword.toLowerCase()
  return data.filter(item => {
    return searchKeys.some(key => {
      const value = item[key]
      return value && String(value).toLowerCase().includes(lowerKeyword)
    })
  })
}

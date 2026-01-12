import type { Ref } from 'vue'
import type { TableCheck, TableColumn } from './useTable'
import cache from '@/plugins/cache'

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
  const applyColumnSettings = (
    columns: Ref<TableColumn[]>,
    tableKey: string
  ) => {
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
  const updateColumnChecks = (
    checks: Ref<TableCheck[]>,
    tableKey: string
  ) => {
    watch(
      checks,
      (newChecks) => {
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

/**
 * 文件名相关工具函数
 */

/**
 * 判断文件名是否可能被截断（用于决定是否显示 tooltip）
 * @param fileName 文件名
 * @param maxLength 最大显示长度（可选，默认根据视图模式判断）
 * @param viewMode 视图模式：'grid' | 'list' | 'table'（可选）
 * @returns 是否可能被截断
 */
export function isFileNameTruncated(
  fileName: string | null | undefined,
  maxLength?: number,
  viewMode?: 'grid' | 'list' | 'table'
): boolean {
  if (!fileName) return false

  // 如果提供了 maxLength，直接使用
  if (maxLength !== undefined) {
    return fileName.length > maxLength
  }

  // 根据视图模式判断
  if (viewMode === 'grid') {
    // 网格视图：2行大约可以显示30-40个字符（取决于卡片宽度）
    return fileName.length > 35
  } else if (viewMode === 'list') {
    // 列表视图（移动端）：单行大约可以显示20-30个字符
    return fileName.length > 25
  } else {
    // 表格视图：单行大约可以显示20-30个字符
    return fileName.length > 25
  }
}

/**
 * 获取文件名的显示样式类
 * @param viewMode 视图模式
 * @returns 样式类名
 */
export function getFileNameClass(viewMode?: 'grid' | 'list' | 'table'): string {
  const baseClass = 'file-name-text'
  if (viewMode === 'grid') {
    return `${baseClass} ${baseClass}--grid`
  } else if (viewMode === 'list') {
    return `${baseClass} ${baseClass}--list`
  } else {
    return `${baseClass} ${baseClass}--table`
  }
}

/**
 * 可访问性增强 Composable
 * 提供 ARIA 属性管理和键盘导航支持
 */

export interface AccessibilityOptions {
  /** 元素角色 */
  role?: string
  /** 元素标签 */
  label?: string
  /** 是否可聚焦 */
  focusable?: boolean
  /** Tab 索引 */
  tabIndex?: number
  /** 是否隐藏（对屏幕阅读器） */
  hidden?: boolean
  /** 是否必需 */
  required?: boolean
  /** 是否禁用 */
  disabled?: boolean
  /** 是否选中 */
  checked?: boolean
  /** 是否展开 */
  expanded?: boolean
  /** 当前值 */
  current?: boolean | 'page' | 'step' | 'location' | 'date' | 'time'
  /** 描述 */
  describedBy?: string
  /** 错误信息 ID */
  errorMessage?: string
}

/**
 * 可访问性增强 Composable
 * @param options 可访问性配置
 */
export function useAccessibility(options: AccessibilityOptions = {}) {
  const {
    role,
    label,
    focusable = true,
    tabIndex,
    hidden = false,
    required = false,
    disabled = false,
    checked,
    expanded,
    current,
    describedBy,
    errorMessage
  } = options

  /**
   * 生成 ARIA 属性对象
   */
  const getAriaAttrs = (): Record<string, string | boolean | number | undefined> => {
    const attrs: Record<string, string | boolean | number | undefined> = {}

    if (role) attrs.role = role
    if (label) attrs['aria-label'] = label
    if (hidden) attrs['aria-hidden'] = hidden
    if (required) attrs['aria-required'] = required
    if (disabled) attrs['aria-disabled'] = disabled
    if (checked !== undefined) attrs['aria-checked'] = checked
    if (expanded !== undefined) attrs['aria-expanded'] = expanded
    if (current !== undefined) attrs['aria-current'] = current
    if (describedBy) attrs['aria-describedby'] = describedBy
    if (errorMessage) attrs['aria-errormessage'] = errorMessage

    // Tab 索引
    if (focusable && tabIndex !== undefined) {
      attrs.tabindex = tabIndex
    } else if (!focusable) {
      attrs.tabindex = -1
    }

    return attrs
  }

  /**
   * 处理键盘事件
   */
  const handleKeydown = (
    event: KeyboardEvent,
    handlers: {
      onEnter?: () => void
      onEscape?: () => void
      onArrowUp?: () => void
      onArrowDown?: () => void
      onArrowLeft?: () => void
      onArrowRight?: () => void
      onTab?: () => void
      onSpace?: () => void
      custom?: (key: string, event: KeyboardEvent) => void
    }
  ) => {
    const { key } = event

    switch (key) {
      case 'Enter':
        handlers.onEnter?.()
        break
      case 'Escape':
        handlers.onEscape?.()
        break
      case 'ArrowUp':
        handlers.onArrowUp?.()
        break
      case 'ArrowDown':
        handlers.onArrowDown?.()
        break
      case 'ArrowLeft':
        handlers.onArrowLeft?.()
        break
      case 'ArrowRight':
        handlers.onArrowRight?.()
        break
      case 'Tab':
        handlers.onTab?.()
        break
      case ' ':
        event.preventDefault()
        handlers.onSpace?.()
        break
      default:
        handlers.custom?.(key, event)
    }
  }

  /**
   * 聚焦元素
   */
  const focusElement = (element: HTMLElement | null) => {
    if (element && typeof element.focus === 'function') {
      element.focus()
    }
  }

  /**
   * 管理焦点陷阱（用于模态框等）
   */
  const createFocusTrap = (container: HTMLElement) => {
    const focusableElements = container.querySelectorAll<HTMLElement>(
      'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])'
    )

    const firstElement = focusableElements[0]
    const lastElement = focusableElements[focusableElements.length - 1]

    const handleTabKey = (event: KeyboardEvent) => {
      if (event.key !== 'Tab') return

      if (event.shiftKey) {
        // Shift + Tab
        if (document.activeElement === firstElement) {
          event.preventDefault()
          lastElement?.focus()
        }
      } else {
        // Tab
        if (document.activeElement === lastElement) {
          event.preventDefault()
          firstElement?.focus()
        }
      }
    }

    container.addEventListener('keydown', handleTabKey)

    return () => {
      container.removeEventListener('keydown', handleTabKey)
    }
  }

  return {
    getAriaAttrs,
    handleKeydown,
    focusElement,
    createFocusTrap
  }
}

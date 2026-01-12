import { useRouter } from 'vue-router'

interface ShortcutConfig {
  key: string
  ctrl?: boolean
  shift?: boolean
  alt?: boolean
  meta?: boolean
  handler: () => void
  description?: string
}

/**
 * 键盘快捷键管理 Composable
 */
export function useKeyboardShortcuts() {
  const router = useRouter()
  const shortcuts = ref<ShortcutConfig[]>([])
  const showHelp = ref(false)

  // 注册快捷键
  const registerShortcut = (config: ShortcutConfig) => {
    shortcuts.value.push(config)
  }

  // 处理键盘事件
  const handleKeyDown = (event: KeyboardEvent) => {
    // 忽略在输入框中的快捷键
    const target = event.target as HTMLElement
    if (
      target.tagName === 'INPUT' ||
      target.tagName === 'TEXTAREA' ||
      target.isContentEditable
    ) {
      // 允许全局快捷键（如 Ctrl+K 搜索）
      if (event.key === 'k' && (event.ctrlKey || event.metaKey)) {
        event.preventDefault()
        // 触发搜索
        const searchInput = document.querySelector('.search-input input') as HTMLInputElement
        if (searchInput) {
          searchInput.focus()
        }
      }
      return
    }

    // 检查是否匹配任何快捷键
    for (const shortcut of shortcuts.value) {
      if (
        event.key.toLowerCase() === shortcut.key.toLowerCase() &&
        !!event.ctrlKey === !!shortcut.ctrl &&
        !!event.shiftKey === !!shortcut.shift &&
        !!event.altKey === !!shortcut.alt &&
        !!event.metaKey === !!shortcut.meta
      ) {
        event.preventDefault()
        shortcut.handler()
        break
      }
    }

    // 显示快捷键帮助（按 ? 键）
    if (event.key === '?' && !event.ctrlKey && !event.metaKey) {
      event.preventDefault()
      showHelp.value = !showHelp.value
    }
  }

  // 注册默认快捷键
  const registerDefaultShortcuts = () => {
    registerShortcut({
      key: 'k',
      ctrl: true,
      handler: () => {
        const searchInput = document.querySelector('.search-input input') as HTMLInputElement
        if (searchInput) {
          searchInput.focus()
        }
      },
      description: '聚焦搜索框'
    })

    registerShortcut({
      key: 'n',
      ctrl: true,
      handler: () => {
        // 新建文件夹（如果在文件页面）
        const newFolderBtn = document.querySelector('[aria-label="新建文件夹"]') as HTMLElement
        if (newFolderBtn) {
          newFolderBtn.click()
        }
      },
      description: '新建文件夹'
    })

    registerShortcut({
      key: 'u',
      ctrl: true,
      handler: () => {
        // 上传文件
        const uploadBtn = document.querySelector('[aria-label="上传文件"]') as HTMLElement
        if (uploadBtn) {
          uploadBtn.click()
        }
      },
      description: '上传文件'
    })

    registerShortcut({
      key: 'f',
      ctrl: true,
      handler: () => {
        router.push('/files')
      },
      description: '跳转到文件页面'
    })

    registerShortcut({
      key: 's',
      ctrl: true,
      handler: () => {
        router.push('/shares')
      },
      description: '跳转到分享页面'
    })

    registerShortcut({
      key: 'Escape',
      handler: () => {
        // 关闭对话框或取消选择
        const dialogs = document.querySelectorAll('.el-overlay')
        if (dialogs.length > 0) {
          const closeBtn = document.querySelector('.el-dialog__close') as HTMLElement
          if (closeBtn) {
            closeBtn.click()
          }
        }
      },
      description: '关闭对话框'
    })
  }

  // 初始化
  onMounted(() => {
    registerDefaultShortcuts()
    window.addEventListener('keydown', handleKeyDown)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleKeyDown)
  })

  // 切换帮助显示
  const toggleHelp = () => {
    showHelp.value = !showHelp.value
  }

  return {
    shortcuts: readonly(shortcuts),
    showHelp,
    registerShortcut,
    toggleHelp
  }
}

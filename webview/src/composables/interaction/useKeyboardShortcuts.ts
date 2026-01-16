import { useRouter } from 'vue-router'
import type { ComponentInternalInstance } from 'vue'
import { useI18n } from '@/composables'

interface ShortcutConfig {
  key: string
  ctrl?: boolean
  shift?: boolean
  alt?: boolean
  meta?: boolean
  handler: () => void
  description?: string
}

// 将状态提升到模块级别，确保所有调用共享同一个状态
const shortcuts = ref<ShortcutConfig[]>([])
const showHelp = ref(false)
let isInitialized = false // 初始化标志，确保只初始化一次
let keydownHandler: ((event: KeyboardEvent) => void) | null = null

/**
 * 键盘快捷键管理 Composable
 */
export function useKeyboardShortcuts() {
  const router = useRouter()
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t, locale } = useI18n()

  // 注册快捷键
  const registerShortcut = (config: ShortcutConfig) => {
    shortcuts.value.push(config)
  }

  // 处理键盘事件
  const handleKeyDown = (event: KeyboardEvent) => {
    // 先检查是否匹配任何快捷键（在检查输入框之前）
    // 这样可以确保即使焦点在输入框上，也能阻止浏览器默认行为
    // 安全检查：event.key 可能为 undefined（某些特殊按键或浏览器兼容性问题）
    if (!event.key) {
      return
    }
    const key = event.key.toLowerCase()
    const isCtrlOrMeta = event.ctrlKey || event.metaKey
    
    // 检查是否匹配快捷键（需要阻止浏览器默认行为的组合键）
    // 避免与浏览器默认行为冲突的快捷键
    const shouldPreventDefault = isCtrlOrMeta && ['u', 'w', 't', 'r', 'p', 'e'].includes(key)
    
    // 忽略在输入框中的快捷键（除了全局快捷键）
    const target = event.target as HTMLElement
    const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable
    
    if (isInput) {
      // 允许全局快捷键（如 Ctrl+K 搜索）
      if (key === 'k' && isCtrlOrMeta) {
        event.preventDefault()
        event.stopPropagation()
        const searchInput = document.querySelector('.search-input input') as HTMLInputElement
        if (searchInput) {
          searchInput.focus()
        }
        return
      }
      // 对于其他快捷键，如果在输入框中，不处理（但如果是需要阻止默认行为的，仍然阻止）
      if (shouldPreventDefault) {
        event.preventDefault()
        event.stopPropagation()
      }
      return
    }

    // 检查是否匹配任何快捷键
    let matched = false
    for (const shortcut of shortcuts.value) {
      // 对于 ctrl 或 meta，在 Windows/Linux 上使用 Ctrl，在 Mac 上使用 Cmd (Meta)
      // 如果 shortcut.ctrl 为 true，则允许使用 Ctrl 或 Meta（跨平台兼容）
      const ctrlOrMetaPressed = event.ctrlKey || event.metaKey
      const ctrlOrMetaRequired = shortcut.ctrl || shortcut.meta
      const ctrlMatch = ctrlOrMetaRequired ? ctrlOrMetaPressed : !ctrlOrMetaPressed
      
      // 安全检查：shortcut.key 必须存在
      if (!shortcut.key) {
        continue
      }
      
      if (
        key === shortcut.key.toLowerCase() &&
        ctrlMatch &&
        !!event.shiftKey === !!shortcut.shift &&
        !!event.altKey === !!shortcut.alt
      ) {
        // 先阻止浏览器默认行为，防止快捷键冲突
        event.preventDefault()
        event.stopPropagation()
        // 然后执行处理函数
        try {
          shortcut.handler()
        } catch (error) {
          proxy?.$log.warn('快捷键处理函数执行失败:', error)
        }
        matched = true
        break
      }
    }

    // 如果匹配到快捷键，确保阻止浏览器默认行为
    if (matched) {
      return
    }

    // 显示快捷键帮助（按 ? 键）
    if (event.key === '?' && !event.ctrlKey && !event.metaKey) {
      event.preventDefault()
      showHelp.value = !showHelp.value
    }
  }

  // 注册默认快捷键
  const registerDefaultShortcuts = () => {
    // 先清空数组，避免重复注册
    shortcuts.value = []
    // 触发上传的辅助函数
    const triggerUpload = () => {
      let uploadBtn: HTMLElement | null = null

      // 方式1: 查找 .toolbar-actions 中的第一个按钮（上传按钮是第一个）
      const toolbarActions = document.querySelector('.toolbar-actions')
      if (toolbarActions) {
        // 查找第一个 .action-btn（可能在 tooltip 内）
        uploadBtn = toolbarActions.querySelector('.action-btn') as HTMLElement
        // 如果没找到，查找第一个 type="primary" 的按钮
        if (!uploadBtn) {
          uploadBtn = toolbarActions.querySelector('.el-button.type-primary') as HTMLElement
        }
        // 如果还是没找到，查找第一个按钮（无论类型）
        if (!uploadBtn) {
          const buttons = Array.from(toolbarActions.querySelectorAll('.el-button'))
          if (buttons.length > 0) {
            uploadBtn = buttons[0] as HTMLElement
          }
        }
      }

      // 方式2: 查找工具栏中的第一个主要按钮
      if (!uploadBtn) {
        const toolbar = document.querySelector('.toolbar, .toolbar-container')
        if (toolbar) {
          // 查找第一个 type="primary" 的按钮
          uploadBtn = toolbar.querySelector('.el-button.type-primary') as HTMLElement
          // 如果没找到，查找第一个 .action-btn
          if (!uploadBtn) {
            uploadBtn = toolbar.querySelector('.action-btn') as HTMLElement
          }
        }
      }

      // 方式3: 通过按钮文本内容查找（支持中英文）
      if (!uploadBtn) {
        const buttons = Array.from(document.querySelectorAll('.el-button'))
        uploadBtn = buttons.find((btn) => {
          const text = btn.textContent?.trim()
          return (
            text === '上传文件' ||
            text === 'Upload File' ||
            text?.includes('上传') ||
            text?.toLowerCase().includes('upload')
          )
        }) as HTMLElement
      }

      // 方式4: 查找包含 Upload 图标的按钮（Element Plus 图标）
      if (!uploadBtn) {
        // Element Plus 的图标可能通过 i 标签或 svg 实现
        const buttons = Array.from(document.querySelectorAll('.el-button'))
        uploadBtn = buttons.find((btn) => {
          // 检查按钮内是否有 Upload 相关的图标
          const icon = btn.querySelector('i, svg')
          if (icon) {
            const iconClass = icon.className || ''
            return iconClass.includes('upload') || iconClass.includes('Upload')
          }
          return false
        }) as HTMLElement
      }

      if (uploadBtn) {
        // 确保按钮可见且可点击
        if (uploadBtn.offsetParent !== null && !uploadBtn.hasAttribute('disabled')) {
          uploadBtn.click()
        } else {
          proxy?.$log.warn('上传按钮不可用（可能被禁用或隐藏）')
        }
      } else {
        proxy?.$log.warn('未找到上传按钮，快捷键 Ctrl+U 无法执行。')
      }
    }

    // 触发新建文件夹的辅助函数
    const triggerNewFolder = () => {
      let newFolderBtn: HTMLElement | null = null

      // 方式1: 查找 .toolbar-actions 中的第二个按钮（新建文件夹是第二个）
      const toolbarActions = document.querySelector('.toolbar-actions')
      if (toolbarActions) {
        const buttons = Array.from(toolbarActions.querySelectorAll('.el-button'))
        // 跳过第一个按钮（上传），取第二个按钮（新建文件夹）
        if (buttons.length > 1) {
          newFolderBtn = buttons[1] as HTMLElement
        }
      }

      // 方式2: 通过 class 查找（action-btn-secondary 是新建文件夹按钮的类名）
      if (!newFolderBtn) {
        newFolderBtn = document.querySelector('.action-btn-secondary') as HTMLElement
      }

      // 方式3: 通过按钮文本查找（支持中英文）
      if (!newFolderBtn) {
        const buttons = Array.from(document.querySelectorAll('.el-button'))
        newFolderBtn = buttons.find((btn) => {
          const text = btn.textContent?.trim()
          return (
            text === '新建文件夹' ||
            text === 'New Folder' ||
            text?.includes('新建') ||
            text?.toLowerCase().includes('new folder')
          )
        }) as HTMLElement
      }

      // 方式4: 查找包含 FolderAdd 图标的按钮
      if (!newFolderBtn) {
        const buttons = Array.from(document.querySelectorAll('.el-button'))
        newFolderBtn = buttons.find((btn) => {
          const icon = btn.querySelector('i, svg')
          if (icon) {
            const iconClass = icon.className || ''
            return iconClass.includes('folder-add') || iconClass.includes('FolderAdd')
          }
          return false
        }) as HTMLElement
      }

      if (newFolderBtn) {
        if (newFolderBtn.offsetParent !== null && !newFolderBtn.hasAttribute('disabled')) {
          newFolderBtn.click()
        } else {
          proxy?.$log.warn('新建文件夹按钮不可用（可能被禁用或隐藏）')
        }
      } else {
        proxy?.$log.warn('未找到新建文件夹按钮，快捷键 Ctrl+N 无法执行。')
      }
    }

    registerShortcut({
      key: 'e',
      ctrl: true,
      handler: () => {
        // 检查是否在文件页面
        const isFilesPage = router.currentRoute.value.name === 'Files' || router.currentRoute.value.path === '/files'
        
        // 如果不在文件页面，先跳转到文件页面
        if (!isFilesPage) {
          router.push('/files').then(() => {
            // 使用重试机制等待页面加载完成
            let retryCount = 0
            const maxRetries = 20 // 最多重试 20 次（2秒）
            const retryInterval = 100 // 每次间隔 100ms
            
            const tryTriggerNewFolder = () => {
              const toolbarActions = document.querySelector('.toolbar-actions')
              if (toolbarActions) {
                // 页面已加载，尝试触发新建文件夹
                triggerNewFolder()
              } else if (retryCount < maxRetries) {
                // 页面还未加载完成，继续等待
                retryCount++
                setTimeout(tryTriggerNewFolder, retryInterval)
              } else {
                // 超时，输出警告
                proxy?.$log.warn('页面加载超时，无法执行新建文件夹操作')
              }
            }
            
            // 开始重试
            setTimeout(tryTriggerNewFolder, retryInterval)
          })
          return
        }

        // 在文件页面，直接触发新建文件夹
        triggerNewFolder()
      },
      description: t('shortcuts.newFolder')
    })

    registerShortcut({
      key: 'k',
      ctrl: true,
      handler: () => {
        const searchInput = document.querySelector('.search-input input') as HTMLInputElement
        if (searchInput) {
          searchInput.focus()
        }
      },
      description: t('shortcuts.focusSearch')
    })

    registerShortcut({
      key: 'u',
      ctrl: true,
      handler: () => {
        // 检查是否在文件页面
        const isFilesPage = router.currentRoute.value.name === 'Files' || router.currentRoute.value.path === '/files'
        
        // 如果不在文件页面，先跳转到文件页面
        if (!isFilesPage) {
          router.push('/files').then(() => {
            // 使用重试机制等待页面加载完成
            let retryCount = 0
            const maxRetries = 20 // 最多重试 20 次（2秒）
            const retryInterval = 100 // 每次间隔 100ms
            
            const tryTriggerUpload = () => {
              const toolbarActions = document.querySelector('.toolbar-actions')
              if (toolbarActions) {
                // 页面已加载，尝试触发上传
                triggerUpload()
              } else if (retryCount < maxRetries) {
                // 页面还未加载完成，继续等待
                retryCount++
                setTimeout(tryTriggerUpload, retryInterval)
              } else {
                // 超时，输出警告
                proxy?.$log.warn('页面加载超时，无法执行上传操作')
              }
            }
            
            // 开始重试
            setTimeout(tryTriggerUpload, retryInterval)
          })
          return
        }

        // 在文件页面，直接触发上传
        triggerUpload()
      },
      description: t('shortcuts.uploadFile')
    })

    registerShortcut({
      key: 'f',
      ctrl: true,
      handler: () => {
        router.push('/files')
      },
      description: t('shortcuts.goToFiles')
    })

    registerShortcut({
      key: 's',
      ctrl: true,
      handler: () => {
        router.push('/shares')
      },
      description: t('shortcuts.goToShares')
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
      description: t('shortcuts.closeDialog')
    })
  }

  // 监听语言变化，重新注册快捷键以更新描述（需要在 onMounted 之前定义）
  watch(locale, () => {
    shortcuts.value = []
    registerDefaultShortcuts()
  })

  // 初始化（只执行一次）
  onMounted(() => {
    // 注册快捷键（每次都会执行，但 registerDefaultShortcuts 内部会先清空数组）
    registerDefaultShortcuts()
    
    // 事件监听器只注册一次
    if (!isInitialized) {
      keydownHandler = handleKeyDown
      // 使用 capture 模式确保在事件捕获阶段就能处理，优先于浏览器默认行为
      window.addEventListener('keydown', keydownHandler, true)
      isInitialized = true
    }
  })

  onBeforeUnmount(() => {
    // 只有最后一个组件卸载时才移除事件监听器
    // 注意：这里无法准确判断是否是最后一个，所以暂时不移除
    // 因为事件监听器是全局的，应该一直存在
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

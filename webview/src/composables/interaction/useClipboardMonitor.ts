/**
 * 剪贴板监听 Composable
 * 用于自动识别用户复制的下载链接并提示创建下载任务
 */

import type { ComponentInternalInstance } from 'vue'
import { onMounted, onBeforeUnmount, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from '@/composables'
import { readClipboardText, readClipboardFile, fileToBase64 } from '@/utils/ui/clipboard'
import { detectLinkType, formatLinkDisplayName, type DetectedLink } from '@/utils/clipboard/linkDetector'
import logger from '@/plugins/logger'
import cache from '@/plugins/cache'

// 用户设置键名
const SETTINGS_KEY = 'clipboardMonitorEnabled'
const PROCESSED_LINKS_KEY = 'processedClipboardLinks'

// 防抖时间（毫秒）
const DEBOUNCE_TIME = 3000 // 3秒内不重复提示相同链接

// 检查间隔（毫秒）
const CHECK_INTERVAL = 2000 // 每2秒检查一次剪贴板

interface ProcessedLink {
  content: string
  timestamp: number
}

/**
 * 剪贴板监听 Composable
 */
export function useClipboardMonitor() {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()
  const router = useRouter()
  const authStore = useAuthStore()

  // 状态
  const isEnabled = ref(false)
  const isMonitoring = ref(false)
  let checkInterval: number | null = null
  let lastClipboardContent = ''
  let lastCheckTime = 0

  /**
   * 加载用户设置
   */
  const loadSettings = () => {
    try {
      const saved = cache.local.get(SETTINGS_KEY)
      isEnabled.value = saved === 'true' || saved === '1'
    } catch (error) {
      logger.error('加载剪贴板监听设置失败:', error)
      isEnabled.value = false
    }
  }

  /**
   * 保存用户设置
   */
  const saveSettings = (enabled: boolean) => {
    try {
      cache.local.set(SETTINGS_KEY, enabled)
      isEnabled.value = enabled
      if (enabled && authStore.isAuthenticated) {
        startMonitoring()
      } else {
        stopMonitoring()
      }
    } catch (error) {
      logger.error('保存剪贴板监听设置失败:', error)
    }
  }

  /**
   * 获取已处理的链接列表
   */
  const getProcessedLinks = (): ProcessedLink[] => {
    try {
      const saved = cache.local.getJSON(PROCESSED_LINKS_KEY)
      if (Array.isArray(saved)) {
        // 清理过期的记录（超过防抖时间的记录）
        const now = Date.now()
        return saved.filter((link: ProcessedLink) => now - link.timestamp < DEBOUNCE_TIME)
      }
      return []
    } catch (error) {
      logger.error('获取已处理链接列表失败:', error)
      return []
    }
  }

  /**
   * 保存已处理的链接
   */
  const saveProcessedLink = (content: string) => {
    try {
      const processedLinks = getProcessedLinks()
      processedLinks.push({
        content,
        timestamp: Date.now()
      })
      cache.local.setJSON(PROCESSED_LINKS_KEY, processedLinks)
    } catch (error) {
      logger.error('保存已处理链接失败:', error)
    }
  }

  /**
   * 检查链接是否已处理过
   */
  const isLinkProcessed = (content: string): boolean => {
    const processedLinks = getProcessedLinks()
    return processedLinks.some(link => link.content === content)
  }

  /**
   * 处理检测到的链接
   */
  const handleDetectedLink = async (link: DetectedLink) => {
    // 检查是否已处理过
    if (isLinkProcessed(link.content)) {
      return
    }

    // 标记为已处理
    saveProcessedLink(link.content)

    // 显示确认对话框
    const displayName = formatLinkDisplayName(link)
    const linkTypeKey =
      link.type === 'http'
        ? 'settings.clipboardMonitor.linkType.http'
        : link.type === 'magnet'
          ? 'settings.clipboardMonitor.linkType.magnet'
          : 'settings.clipboardMonitor.linkType.torrent'
    const linkTypeText = t(linkTypeKey)

    try {
      proxy?.$log.debug('显示确认对话框', { linkType: link.type, displayName, linkTypeText })
      
      await ElMessageBox.confirm(
        t('settings.clipboardMonitor.confirmMessage', {
          linkType: linkTypeText,
          displayName
        }),
        t('settings.clipboardMonitor.confirmTitle'),
        {
          confirmButtonText: t('common.confirm'),
          cancelButtonText: t('common.cancel'),
          type: 'info'
        }
      )

      // 用户确认，创建下载任务
      proxy?.$log.debug('用户确认创建下载任务', { linkType: link.type, displayName })
      
      // 不在这里显示加载提示，让 createDownloadTask 根据实际步骤显示不同的提示
      try {
        await createDownloadTask(link)
      } catch (taskError: any) {
        // 错误已在 createDownloadTask 中处理，这里只记录日志
        proxy?.$log.error('创建下载任务异常:', taskError)
        throw taskError // 重新抛出，让外层 catch 处理
      }
    } catch (error: any) {
      // 用户取消，不做任何操作
      if (error === 'cancel' || error?.action === 'cancel' || error?.message === 'cancel') {
        proxy?.$log.debug('用户取消创建下载任务')
      } else {
        // 其他错误，记录日志
        proxy?.$log.error('确认对话框或创建任务错误:', error)
        // 如果错误不是用户取消，显示错误消息
        if (error?.message && !error.message.includes('cancel')) {
          ElMessage.error(error.message || '操作失败')
        }
      }
    }
  }

  /**
   * 创建下载任务
   */
  const createDownloadTask = async (link: DetectedLink) => {
    try {
      proxy?.$log.debug('开始创建下载任务', { linkType: link.type, content: link.content.substring(0, 50) })
      
      if (link.type === 'http') {
        // HTTP/HTTPS 下载 - 跳转到离线下载页面，让用户确认
        proxy?.$log.debug('跳转到离线下载页面，填充 HTTP 链接', { url: link.url || link.content })
        
        // 将链接内容存储到 sessionStorage，供离线下载页面使用
        sessionStorage.setItem('clipboardTorrentContent', link.url || link.content)
        sessionStorage.setItem('clipboardTorrentType', 'http')
        
        // 跳转到离线下载页面，并传递参数表示需要打开弹窗
        router
          .push({
            path: '/offline',
            query: { openDialog: 'true', inputType: 'text' }
          })
          .catch(err => {
            proxy?.$log.error('跳转到离线下载页面失败:', err)
          })
      } else if (link.type === 'magnet' || link.type === 'torrent') {
        // 磁力链接或种子文件下载
        // 不在这里解析，直接跳转到离线下载页面，让页面自动填充并解析
        const content = link.type === 'magnet' ? link.magnet || link.content : link.torrentBase64 || link.content
        proxy?.$log.debug('跳转到离线下载页面，自动填充并解析', { type: link.type })
        
        // 将种子内容存储到 sessionStorage，供离线下载页面使用
        sessionStorage.setItem('clipboardTorrentContent', content)
        sessionStorage.setItem('clipboardTorrentType', link.type)
        
        // 跳转到离线下载页面，并传递参数表示需要打开弹窗
        // 根据类型设置 inputType：magnet 使用 'text'，torrent 使用 'file'
        const inputTypeParam = link.type === 'magnet' ? 'text' : 'file'
        router
          .push({
            path: '/offline',
            query: { openDialog: 'true', inputType: inputTypeParam, autoParse: 'true' }
          })
          .catch(err => {
            proxy?.$log.error('跳转到离线下载页面失败:', err)
          })
      }
    } catch (error: any) {
      logger.error('创建下载任务失败:', error)
      const errorMessage = error?.message || error?.response?.data?.message || '未知错误'
      proxy?.$log.error('创建下载任务失败详情', { 
        error, 
        errorMessage,
        errorStack: error?.stack,
        errorResponse: error?.response
      })
      ElMessage.error(t('settings.clipboardMonitor.createFailed', { error: errorMessage }))
      // 重新抛出错误，让外层 catch 处理
      throw error
    }
  }

  /**
   * 处理粘贴事件（用户主动粘贴时触发）
   */
  const handlePaste = async (event: ClipboardEvent) => {
    // 检查是否启用且已登录
    if (!isEnabled.value || !authStore.isAuthenticated) {
      return
    }

    // 防抖：避免频繁处理
    const now = Date.now()
    if (now - lastCheckTime < DEBOUNCE_TIME) {
      return
    }
    lastCheckTime = now

    try {
      // 从粘贴事件中获取剪贴板数据
      const clipboardData = event.clipboardData
      if (!clipboardData) {
        return
      }

      // 先检查文本内容
      const text = clipboardData.getData('text')
      if (text && text !== lastClipboardContent) {
        lastClipboardContent = text

        // 检测链接类型
        const detectedLink = detectLinkType(text)
        if (detectedLink) {
          await handleDetectedLink(detectedLink)
          return
        }
      }

      // 检查文件（种子文件）
      const items = clipboardData.items
      if (items) {
        for (let i = 0; i < items.length; i++) {
          const item = items[i]
          if (item.kind === 'file' && item.type === 'application/x-bittorrent') {
            const file = item.getAsFile()
            if (file && file.name.endsWith('.torrent')) {
              try {
                const base64 = await fileToBase64(file)
                const detectedLink = detectLinkType(base64)
                if (detectedLink && detectedLink.content !== lastClipboardContent) {
                  lastClipboardContent = detectedLink.content
                  await handleDetectedLink(detectedLink)
                }
              } catch (error) {
                logger.error('处理种子文件失败:', error)
              }
            }
          }
        }
      }
    } catch (error) {
      logger.warn('处理粘贴事件失败:', error)
    }
  }

  /**
   * 检查剪贴板内容（尝试使用 Clipboard API，需要用户交互上下文）
   * 注意：此方法在定时器中调用时可能会失败，因为 Clipboard API 需要用户激活
   */
  const checkClipboard = async () => {
    // 检查是否启用且已登录
    if (!isEnabled.value || !authStore.isAuthenticated) {
      return
    }

    // 防抖：避免频繁检查
    const now = Date.now()
    if (now - lastCheckTime < CHECK_INTERVAL / 2) {
      return
    }
    lastCheckTime = now

    try {
      // 尝试使用 Clipboard API（可能失败，因为需要用户激活）
      // 先尝试读取文本
      const text = await readClipboardText()
      if (text && text !== lastClipboardContent) {
        lastClipboardContent = text

        // 检测链接类型
        const detectedLink = detectLinkType(text)
        if (detectedLink) {
          await handleDetectedLink(detectedLink)
          return
        }
      }

      // 如果文本检测失败，尝试读取文件（种子文件）
      const file = await readClipboardFile()
      if (file && file.name.endsWith('.torrent')) {
        try {
          const base64 = await fileToBase64(file)
          const detectedLink = detectLinkType(base64)
          if (detectedLink && detectedLink.content !== lastClipboardContent) {
            lastClipboardContent = detectedLink.content
            await handleDetectedLink(detectedLink)
          }
        } catch (error) {
          logger.error('处理种子文件失败:', error)
        }
      }
    } catch (error) {
      // 权限错误等，静默处理（不记录为错误，因为这是预期的行为）
      // Clipboard API 在定时器中调用时会失败，这是正常的安全限制
    }
  }

  /**
   * 开始监听
   */
  const startMonitoring = () => {
    if (isMonitoring.value || !isEnabled.value || !authStore.isAuthenticated) {
      return
    }

    isMonitoring.value = true
    lastClipboardContent = ''
    lastCheckTime = 0

    // 监听粘贴事件（主要方式，用户粘贴时触发）
    document.addEventListener('paste', handlePaste, true)

    // 尝试定时检查（可能失败，因为 Clipboard API 需要用户激活）
    // 这个定时器主要用于处理用户在其他地方复制后切换回页面的情况
    // 但可能因为安全限制而无法工作
    checkInterval = window.setInterval(() => {
      checkClipboard()
    }, CHECK_INTERVAL)

    proxy?.$log.debug('剪贴板监听已启动')
  }

  /**
   * 停止监听
   */
  const stopMonitoring = () => {
    if (!isMonitoring.value) {
      return
    }

    isMonitoring.value = false

    // 移除粘贴事件监听
    document.removeEventListener('paste', handlePaste, true)

    if (checkInterval !== null) {
      clearInterval(checkInterval)
      checkInterval = null
    }

    proxy?.$log.debug('剪贴板监听已停止')
  }

  /**
   * 监听登录状态变化
   */
  watch(
    () => authStore.isAuthenticated,
    (isAuthenticated) => {
      if (isAuthenticated && isEnabled.value) {
        startMonitoring()
      } else {
        stopMonitoring()
      }
    }
  )

  /**
   * 监听启用状态变化
   */
  watch(
    () => isEnabled.value,
    (enabled) => {
      if (enabled && authStore.isAuthenticated) {
        startMonitoring()
      } else {
        stopMonitoring()
      }
    }
  )

  // 初始化
  onMounted(() => {
    loadSettings()
    if (isEnabled.value && authStore.isAuthenticated) {
      startMonitoring()
    }
  })

  onBeforeUnmount(() => {
    stopMonitoring()
  })

  return {
    isEnabled: readonly(isEnabled),
    isMonitoring: readonly(isMonitoring),
    enable: () => saveSettings(true),
    disable: () => saveSettings(false),
    toggle: () => saveSettings(!isEnabled.value),
    start: startMonitoring,
    stop: stopMonitoring
  }
}

import { driver } from 'driver.js'
import 'driver.js/dist/driver.css'
import type { ComponentInternalInstance } from 'vue'
import { useI18n } from '@/composables'

export interface OnboardingStep {
  element: string
  popover: {
    title: string
    description: string
    position?: 'left' | 'right' | 'top' | 'bottom'
  }
}

const STORAGE_KEY_WELCOME = 'onboarding_welcome_completed'
const STORAGE_KEY_FEATURES = 'onboarding_features_completed'

/**
 * 新手引导管理 Composable
 */
export function useOnboarding() {
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()
  const driverInstance = ref<ReturnType<typeof driver> | null>(null)
  const showWelcomeDialog = ref(false)
  
  // 防止重复调用的标志
  const isCheckingOnboarding = ref(false)

  // 使用 ref 而不是 computed，以便可以手动刷新
  const isWelcomeCompleted = ref(false)
  const isFeaturesCompleted = ref(false)

  // 检查引导状态
  const checkOnboardingStatus = () => {
    if (typeof window === 'undefined') {
      isWelcomeCompleted.value = true
      isFeaturesCompleted.value = true
      return
    }
    isWelcomeCompleted.value = localStorage.getItem(STORAGE_KEY_WELCOME) === 'true'
    isFeaturesCompleted.value = localStorage.getItem(STORAGE_KEY_FEATURES) === 'true'
  }

  // 初始化 Driver 实例
  const initDriver = (forceRecreate = false) => {
    // 如果强制重新创建，先销毁旧实例
    if (forceRecreate && driverInstance.value) {
      try {
        if (driverInstance.value.isActive()) {
          driverInstance.value.destroy()
        }
      } catch (e) {
        proxy?.$log.warn('强制重新创建时销毁旧实例出错:', e)
      }
      driverInstance.value = null
    }
    
    // 如果实例已存在且未强制重新创建，直接返回
    if (driverInstance.value && !forceRecreate) {
      // 但如果实例不在活动状态，也清理掉
      try {
        if (!driverInstance.value.isActive()) {
          driverInstance.value = null
        } else {
          return driverInstance.value
        }
      } catch (e) {
        // 如果检查状态出错，也清理掉
        driverInstance.value = null
      }
    }

    // 清理可能残留的 driver.js DOM 元素
    try {
      const driverOverlay = document.querySelector('.driver-overlay')
      if (driverOverlay) {
        driverOverlay.remove()
      }
      const driverHighlightedElement = document.querySelector('.driver-highlighted-element')
      if (driverHighlightedElement) {
        driverHighlightedElement.remove()
      }
      const driverPopover = document.querySelector('.driver-popover')
      if (driverPopover) {
        driverPopover.remove()
      }
    } catch (e) {
      proxy?.$log.warn('清理残留 DOM 元素时出错:', e)
    }

    // 根据 driver.js 源码，直接在配置中设置按钮文本
    driverInstance.value = driver({
      showProgress: true,
      allowClose: true,
      showButtons: ['next', 'previous', 'close'],
      // 直接配置按钮文本（driver.js 支持这些配置项）
      nextBtnText: t('common.next') || '下一步',
      prevBtnText: t('common.prev') || '上一步',
      doneBtnText: t('common.finish') || '完成',
      onHighlightStarted: (element) => {
        // 高亮开始时，确保目标元素可见
        if (element) {
          element.scrollIntoView({ behavior: 'smooth', block: 'center' })
        }
      },
      onDestroyed: () => {
        // 引导销毁时，标记为已完成
        localStorage.setItem(STORAGE_KEY_FEATURES, 'true')
        checkOnboardingStatus()
      },
      onCloseClick: () => {
        // 点击关闭按钮时，标记为已完成并销毁引导
        localStorage.setItem(STORAGE_KEY_FEATURES, 'true')
        checkOnboardingStatus()
        // 确保引导被销毁
        if (driverInstance.value) {
          driverInstance.value.destroy()
        }
        // 重置检查标志，防止自动重启
        isCheckingOnboarding.value = false
      }
    })

    return driverInstance.value
  }

  // 显示欢迎对话框
  const showWelcome = () => {
    // 先更新状态
    checkOnboardingStatus()
    if (isWelcomeCompleted.value) {
      return
    }
    showWelcomeDialog.value = true
  }

  // 完成欢迎引导
  const completeWelcome = (skip = false) => {
    showWelcomeDialog.value = false
    if (!skip) {
      localStorage.setItem(STORAGE_KEY_WELCOME, 'true')
      // 更新状态
      checkOnboardingStatus()
      // 欢迎弹窗关闭后，继续检查功能引导
      setTimeout(() => {
        isCheckingOnboarding.value = false
        checkAndStartOnboarding()
      }, 300)
    } else {
      isCheckingOnboarding.value = false
    }
  }

  // 开始功能引导
  const startFeaturesTour = () => {
    // 如果已经有活动的引导，先销毁它
    if (driverInstance.value) {
      try {
        if (driverInstance.value.isActive()) {
          driverInstance.value.destroy()
        }
      } catch (e) {
        proxy?.$log.warn('销毁已有引导实例时出错:', e)
      }
      driverInstance.value = null
    }
    
    // 清理可能残留的 DOM 元素
    try {
      const driverOverlay = document.querySelector('.driver-overlay')
      if (driverOverlay) {
        driverOverlay.remove()
      }
      const driverHighlightedElement = document.querySelector('.driver-highlighted-element')
      if (driverHighlightedElement) {
        driverHighlightedElement.remove()
      }
      const driverPopover = document.querySelector('.driver-popover')
      if (driverPopover) {
        driverPopover.remove()
      }
    } catch (e) {
      proxy?.$log.warn('清理残留 DOM 元素时出错:', e)
    }
    
    // 初始化新的引导实例
    initDriver()

    // 定义步骤配置（支持多个选择器作为备选）
    const stepConfigs: Array<{
      selectors: string[]
      popover: {
        title: string
        description: string
        position: 'bottom' | 'right'
      }
    }> = [
      {
        // 上传按钮：在 Files 页面的 toolbar-actions 中，在 tooltip 内
        // 尝试多个选择器以确保能找到
        selectors: [
          '.files-page .toolbar-actions .action-btn',
          '.files-page .toolbar-actions .el-button--primary',
          '.files-page .toolbar-actions button.el-button--primary',
          '.toolbar-actions .action-btn',
          '.toolbar-actions .el-button--primary',
          '.toolbar-actions button.el-button--primary'
        ],
        popover: {
          title: t('onboarding.features.upload.title'),
          description: t('onboarding.features.upload.description'),
          position: 'bottom'
        }
      },
      {
        // 新建文件夹按钮：在 Files 页面的 toolbar-actions 中，第一个 action-btn-secondary
        // 注意：有多个 action-btn-secondary 按钮（新建文件夹和移动文件），需要取第一个
        // 使用 :nth-of-type 或直接通过 querySelectorAll 取第一个
        selectors: [
          '.files-page .toolbar-actions .action-btn-secondary',
          '.files-page .toolbar-actions button.action-btn-secondary',
          '.files-page .toolbar-actions .el-button.action-btn-secondary',
          '.toolbar-actions .action-btn-secondary',
          '.toolbar-actions button.action-btn-secondary',
          '.toolbar-actions .el-button.action-btn-secondary'
        ],
        popover: {
          title: t('onboarding.features.newFolder.title'),
          description: t('onboarding.features.newFolder.description'),
          position: 'bottom'
        }
      },
      {
        // 搜索输入框：在 Header 组件中（全局）
        selectors: [
          '.search-input input',
          '.header-center .search-input input',
          '.layout-header .search-input input'
        ],
        popover: {
          title: t('onboarding.features.search.title'),
          description: t('onboarding.features.search.description'),
          position: 'bottom'
        }
      },
      {
        // 离线下载：左侧导航菜单中的离线下载菜单项
        // Element Plus 的 el-menu-item 使用 router 时，会渲染成包含 router-link 的结构
        // 通过文本内容查找菜单项（支持多语言）
        selectors: [
          // 优先使用国际化文本进行精确匹配（支持中文和英文）
          `.premium-menu .el-menu-item:contains("${t('menu.offline')}")`,
          `.layout-aside .el-menu-item:contains("${t('menu.offline')}")`,
          // 通过路由路径查找（最可靠的方式）- Element Plus 会将 index 作为属性
          '.premium-menu .el-menu-item[data-index="/offline"]',
          '.layout-aside .el-menu-item[data-index="/offline"]',
          '.el-menu-item[data-index="/offline"]',
          // 通过位置查找（第三个菜单项，在"我的文件"和"我的分享"之后）
          // 注意：需要排除分组标题，所以可能需要调整索引
          '.premium-menu .el-menu-item:nth-of-type(3)',
          '.layout-aside .el-menu-item:nth-child(3)',
          // 通过路由链接查找
          '.el-menu-item:has(a[href="/offline"])',
          '.el-menu-item:has(router-link[to="/offline"])'
        ],
        popover: {
          title: t('onboarding.features.offlineDownload.title'),
          description: t('onboarding.features.offlineDownload.description'),
          position: 'right'
        }
      },
      {
        // 个性化设置：Header 中的用户头像
        selectors: [
          '.user-profile',
          '.layout-header .user-profile',
          '.header-right .user-profile',
          '.user-profile .user-avatar-img'
        ],
        popover: {
          title: t('onboarding.features.customize.title'),
          description: t('onboarding.features.customize.description'),
          position: 'bottom'
        }
      }
    ]

    // 查找元素的辅助函数（尝试多个选择器）
    const findElement = (selectors: string[], isFirstOfMultiple = false): { element: Element; selector: string } | null => {
      for (const selector of selectors) {
        // 如果选择器包含 :contains()，使用文本内容查找
        if (selector.includes(':contains(')) {
          const textMatch = selector.match(/:contains\("([^"]+)"\)/)
          if (textMatch) {
            const searchText = textMatch[1]
            const baseSelector = selector.replace(/:contains\("([^"]+)"\).*/, '')
            const elements = document.querySelectorAll(baseSelector)
            
            // 获取搜索文本的所有可能翻译（中文和英文）
            const searchTexts = [
              searchText,
              searchText === '离线下载' ? 'Offline Download' : searchText === 'Offline Download' ? '离线下载' : searchText
            ]
            
            for (const el of Array.from(elements)) {
              // 获取元素的文本内容（去除空白字符）
              const textContent = el.textContent?.trim() || ''
              const textContentLower = textContent.toLowerCase()
              
              // 检查是否匹配任一搜索文本
              for (const search of searchTexts) {
                const searchLower = search.toLowerCase()
                
                // 精确匹配或完全包含搜索文本
                if (textContentLower === searchLower || textContentLower.includes(searchLower)) {
                  // 额外检查：确保不是部分匹配（比如"我的文件"不应该匹配"离线下载"）
                  // 如果搜索文本是"离线下载"或"Offline Download"，确保元素文本不包含"我的文件"等无关文本
                  if (searchLower.includes('离线') || searchLower.includes('offline')) {
                    // 如果是离线下载相关的搜索，排除包含"我的文件"、"我的分享"等的元素
                    if (textContentLower.includes('我的文件') || 
                        textContentLower.includes('my files') ||
                        textContentLower.includes('我的分享') ||
                        textContentLower.includes('my shares') ||
                        textContentLower.includes('分享') ||
                        textContentLower.includes('shares')) {
                      continue
                    }
                    
                    // 额外验证：检查元素是否包含 router-link 指向 /offline
                    const routerLink = el.querySelector('a[href="/offline"], router-link[to="/offline"]')
                    if (!routerLink) {
                      // 如果没有找到指向 /offline 的链接，可能不是正确的菜单项，继续查找
                      continue
                    }
                  }
                  return { element: el, selector: baseSelector }
                }
              }
            }
          }
          continue
        }
        
        // 特殊处理：查找离线下载菜单项时，通过检查路由链接来验证
        if (selector.includes('el-menu-item') && (selector.includes('offline') || selector.includes('nth-of-type(3)') || selector.includes('nth-child(3)'))) {
          const elements = document.querySelectorAll(selector)
          for (const el of Array.from(elements)) {
            // 检查是否包含指向 /offline 的路由链接
            const routerLink = el.querySelector('a[href="/offline"], router-link[to="/offline"]')
            if (routerLink) {
              return { element: el, selector }
            }
            // 或者检查文本内容是否包含"离线下载"或"Offline Download"
            const textContent = el.textContent?.trim() || ''
            const textContentLower = textContent.toLowerCase()
            if ((textContentLower.includes('离线下载') || textContentLower.includes('offline download')) &&
                !textContentLower.includes('我的文件') &&
                !textContentLower.includes('my files') &&
                !textContentLower.includes('我的分享') &&
                !textContentLower.includes('my shares')) {
              return { element: el, selector }
            }
          }
          continue
        }
        
        // 使用 querySelectorAll 然后取第一个，这样即使有多个匹配也能找到
        const elements = document.querySelectorAll(selector)
        if (elements.length > 0) {
          // 如果允许多个匹配，取第一个；否则只取唯一元素
          if (isFirstOfMultiple || elements.length === 1) {
            return { element: elements[0], selector }
          }
          // 如果有多个元素但不允许，继续尝试下一个选择器
        }
      }
      return null
    }

    // 检查是否在 Files 页面，如果不在，先跳转（前两步需要 Files 页面）
    const isFilesPage = window.location.pathname === '/files' || 
                       document.querySelector('.files-page') !== null
    if (!isFilesPage) {
      proxy?.$router.push('/files').then(() => {
        // 等待页面渲染完成后再启动引导
        setTimeout(() => {
          // 重置检查标志，允许重新启动引导
          isCheckingOnboarding.value = false
          startFeaturesTour()
        }, 800)
      }).catch((error) => {
        proxy?.$log.error('跳转到文件页面失败:', error)
        // 重置检查标志
        isCheckingOnboarding.value = false
      })
      return
    }
    
    // 收集有效的引导步骤（只收集当前页面能找到的元素，找不到的步骤会被跳过）
    const validSteps: OnboardingStep[] = []
    
    const collectValidSteps = () => {
      for (let i = 0; i < stepConfigs.length; i++) {
        const config = stepConfigs[i]
        // 新建文件夹按钮（第二个步骤，索引为1）可能有多个匹配，需要取第一个
        const isFirstOfMultiple = i === 1
        const result = findElement(config.selectors, isFirstOfMultiple)
        if (result) {
          validSteps.push({
            element: result.selector,
            popover: config.popover
          })
        } else {
          proxy?.$log.warn(`引导步骤 ${i + 1} 目标元素不存在: ${config.selectors.join(', ')}`)
        }
      }
      
      if (validSteps.length === 0) {
        proxy?.$log.warn('没有可用的引导步骤，可能页面尚未加载完成')
        return
      }
      
      // 转换为 Driver.js 格式
      const driverSteps = validSteps.map((step) => {
        return {
          element: step.element,
          popover: {
            title: step.popover.title,
            description: step.popover.description,
            side: (step.popover.position || 'bottom') as 'left' | 'right' | 'top' | 'bottom',
            align: 'start' as const
          }
        }
      })
      
      driverInstance.value?.setSteps(driverSteps)
      driverInstance.value?.drive()
      
      // 监听引导完成（通过检查是否还有活动的高亮）
      // 注意：onDestroyed 回调已经会处理完成标记，这里只是备用检查
      const checkComplete = setInterval(() => {
        if (!driverInstance.value?.isActive()) {
          clearInterval(checkComplete)
          // 如果 onDestroyed 没有触发，这里作为备用标记
          if (localStorage.getItem(STORAGE_KEY_FEATURES) !== 'true') {
            localStorage.setItem(STORAGE_KEY_FEATURES, 'true')
            checkOnboardingStatus()
          }
        }
      }, 500)
      
      // 10秒后清除检查（防止内存泄漏）
      setTimeout(() => {
        clearInterval(checkComplete)
      }, 10000)
    }
    
    // 调用 collectValidSteps 开始收集步骤
    collectValidSteps()
  }

  // 重置引导状态
  const resetOnboarding = () => {
    localStorage.removeItem(STORAGE_KEY_WELCOME)
    localStorage.removeItem(STORAGE_KEY_FEATURES)
    showWelcomeDialog.value = false
    if (driverInstance.value) {
      driverInstance.value.destroy()
      driverInstance.value = null
    }
    // 更新状态并立即检查
    checkOnboardingStatus()
    checkAndStartOnboarding()
  }

  // 检查并启动引导
  const checkAndStartOnboarding = () => {
    // 如果正在检查中，直接返回，防止重复调用
    if (isCheckingOnboarding.value) {
      return
    }
    
    // 先更新状态
    checkOnboardingStatus()
    
    // 设置检查标志
    isCheckingOnboarding.value = true
    
    // 延迟执行，确保页面已加载
    setTimeout(() => {
      // 如果已经完成，不再启动
      if (isWelcomeCompleted.value && isFeaturesCompleted.value) {
        isCheckingOnboarding.value = false
        return
      }
      
      // 优先显示欢迎弹窗
      if (!isWelcomeCompleted.value) {
        showWelcome()
        // 注意：不重置 isCheckingOnboarding，因为欢迎弹窗关闭后会继续检查功能引导
        return // 显示欢迎弹窗后，不继续执行功能引导
      }
      
      // 如果欢迎已完成但功能引导未完成，检查是否在文件页面
      if (!isFeaturesCompleted.value) {
        // 检查是否已经有活动的引导（如果引导正在进行中，不要重新启动）
        if (driverInstance.value?.isActive()) {
          isCheckingOnboarding.value = false
          return
        }
        
        const isFilesPage = window.location.pathname === '/files' || 
                           document.querySelector('.files-page') !== null
        if (isFilesPage) {
          // 再延迟一点，确保页面元素已渲染
          setTimeout(() => {
            startFeaturesTour()
            isCheckingOnboarding.value = false
          }, 500)
        } else {
          // 不在文件页面，先跳转到文件页面，然后再启动引导
          proxy?.$router.push('/files').then(() => {
            // 等待页面渲染完成后再启动引导
            setTimeout(() => {
              // 再次检查是否已经有活动的引导
              if (driverInstance.value?.isActive()) {
                isCheckingOnboarding.value = false
                return
              }
              startFeaturesTour()
              isCheckingOnboarding.value = false
            }, 800) // 给页面渲染更多时间
          }).catch((error) => {
            proxy?.$log.error('跳转到文件页面失败:', error)
            isCheckingOnboarding.value = false
          })
        }
      } else {
        isCheckingOnboarding.value = false
      }
    }, 500)
  }

  // 初始化
  onMounted(() => {
    initDriver()
    // 初始化时检查状态
    checkOnboardingStatus()
    // 立即检查是否需要启动引导（首次加载时）
    checkAndStartOnboarding()
  })

  // 清理
  onBeforeUnmount(() => {
    if (driverInstance.value) {
      driverInstance.value.destroy()
    }
  })

  return {
    showWelcomeDialog,
    isWelcomeCompleted: readonly(isWelcomeCompleted),
    isFeaturesCompleted: readonly(isFeaturesCompleted),
    showWelcome,
    completeWelcome,
    startFeaturesTour,
    resetOnboarding,
    checkAndStartOnboarding,
    checkOnboardingStatus
  }
}

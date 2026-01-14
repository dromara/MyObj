<template>
  <el-header class="layout-header glass-panel">
    <div class="header-left">
      <!-- 移动端汉堡菜单按钮（非水平布局时显示） -->
      <el-button v-if="!showHorizontalMenu" class="mobile-menu-btn" icon="Menu" circle text @click="toggleSidebar" />
      <!-- 桌面端 Logo -->
      <div class="logo-wrapper desktop-logo">
        <el-image :src="logoImage" alt="MyObj Logo" class="logo-image" fit="contain" :lazy="false" />
        <span class="logo-text">{{ t('header.title') }}</span>
      </div>
    </div>

    <div class="header-center">
      <!-- 水平布局时显示菜单 -->
      <HorizontalMenu v-if="showHorizontalMenu" />
      <!-- 非水平布局时显示搜索框 -->
      <div v-else class="search-wrapper" ref="searchWrapperRef">
        <el-input
          v-model="searchKeyword"
          :placeholder="t('files.searchPlaceholder')"
          prefix-icon="Search"
          clearable
          @input="handleSearchInput"
          @keyup.enter="handleSearch"
          @keyup.down="handleArrowDown"
          @keyup.up="handleArrowUp"
          @focus="showSuggestions = true"
          @blur="handleSearchBlur"
          @clear="handleSearchClear"
          class="search-input glass-input"
        />
        <!-- 搜索建议 -->
        <SearchSuggestions
          v-if="showSuggestions"
          :suggestions="searchSuggestions"
          :visible="showSuggestions && searchSuggestions.length > 0"
          @select="handleSuggestionSelect"
          @clear="handleClearHistory"
          @delete="handleDeleteHistory"
        />
      </div>
    </div>

    <div class="header-right">
      <!-- 全屏切换按钮（桌面端） -->
      <el-tooltip
        :content="isFullscreen ? t('header.exitFullscreen') : t('header.fullscreen')"
        placement="bottom"
        class="desktop-only"
      >
        <el-button
          class="header-action-btn"
          :icon="isFullscreen ? 'CopyDocument' : 'FullScreen'"
          circle
          text
          @click="toggleFullscreen"
        />
      </el-tooltip>
      <!-- 主题切换按钮 -->
      <el-tooltip :content="isDark ? t('header.switchToLight') : t('header.switchToDark')" placement="bottom">
        <el-button
          class="header-action-btn theme-toggle-btn"
          :icon="isDark ? 'Sunny' : 'Moon'"
          circle
          text
          @click="toggleTheme"
        />
      </el-tooltip>
      <!-- 移动端：搜索按钮 -->
      <el-button class="mobile-search-btn" icon="Search" circle text @click="showSearchDialog = true" />
      <!-- 移动端：Logo -->
      <div class="mobile-logo">
        <el-image :src="logoImage" alt="MyObj Logo" class="logo-image-mobile" fit="contain" :lazy="false" />
      </div>
      <!-- 用户头像（桌面端和移动端共用） -->
      <el-dropdown @command="handleCommand" trigger="click">
        <div class="user-profile glass-hover">
          <el-avatar :size="32" :style="{ background: avatarColor }" class="user-avatar-img">
            {{ avatarText }}
          </el-avatar>
          <span class="username desktop-only">{{ userStore.nickname || userStore.username }}</span>
          <el-icon class="el-icon--right desktop-only"><CaretBottom /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu class="premium-dropdown">
            <el-dropdown-item command="settings">
              <el-icon><Setting /></el-icon>
              {{ t('menu.settings') }}
            </el-dropdown-item>
            <el-dropdown-item divided command="logout">
              <el-icon><SwitchButton /></el-icon>
              {{ t('header.logout') }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <!-- 移动端搜索对话框 -->
    <el-dialog
      v-model="showSearchDialog"
      width="90%"
      :close-on-click-modal="true"
      :close-on-press-escape="true"
      :show-close="false"
      :append-to-body="true"
      :center="false"
      class="search-dialog"
      @closed="handleSearchDialogClosed"
    >
      <template #header>
        <div class="search-dialog-header">
          <el-icon class="search-icon"><Search /></el-icon>
          <span class="search-title">{{ t('header.search') }}</span>
        </div>
      </template>

      <div class="search-dialog-body">
        <el-input
          ref="searchDialogInputRef"
          v-model="searchKeyword"
          :placeholder="t('files.searchPlaceholder')"
          prefix-icon="Search"
          clearable
          @input="handleSearchInput"
          @keyup.enter="handleSearchAndClose"
          @clear="handleSearchClear"
          class="search-dialog-input"
          size="large"
        />
      </div>

      <template #footer>
        <div class="search-dialog-footer">
          <el-button class="cancel-btn" @click="showSearchDialog = false">{{ t('common.cancel') }}</el-button>
          <el-button class="search-btn" type="primary" @click="handleSearchAndClose">
            <el-icon><Search /></el-icon>
            {{ t('common.search') }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </el-header>
</template>

<script setup lang="ts">
  import { useTheme, useSearchHistory, useI18n } from '@/composables'
  import { useFullscreen } from '@vueuse/core'
  import logoImage from '@/assets/images/LOGO.png'

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  import { useUserStore, useAuthStore, useLayoutStore } from '@/stores'
  import HorizontalMenu from '../HorizontalMenu/index.vue'

  const router = useRouter()
  const route = useRoute()
  const userStore = useUserStore()
  const authStore = useAuthStore()
  const layoutStore = useLayoutStore()

  // 是否显示水平菜单
  // - 水平布局：显示水平菜单
  // - 顶部混合-头部优先：显示水平菜单
  // - 垂直混合-头部优先：显示水平菜单（一级菜单在 Header）
  // - 顶部混合-侧边栏优先：不显示（菜单在侧边栏中，Header 显示搜索框）
  const showHorizontalMenu = computed(() => {
    const mode = layoutStore.layoutMode
    return mode === 'horizontal' || mode === 'top-hybrid-header-first' || mode === 'vertical-hybrid-header-first'
  })
  const { isDark, toggleTheme } = useTheme()
  const { searchHistory, addHistory, clearHistory, removeHistory } = useSearchHistory()
  const { isFullscreen, toggle: toggleFullscreen } = useFullscreen()

  const searchKeyword = ref('')
  let searchTimer: ReturnType<typeof setTimeout> | null = null
  const showSearchDialog = ref(false)
  const searchDialogInputRef = ref()
  const showSuggestions = ref(false)
  const selectedSuggestionIndex = ref(-1)

  const avatarText = computed(() => {
    const nickname = userStore.nickname || userStore.username || ''
    return nickname ? nickname.charAt(0).toUpperCase() : 'U'
  })

  const avatarColor = computed(() => {
    // 使用 CSS 变量，但需要转换为实际颜色值
    // 由于 CSS 变量在 style 中可以直接使用，这里返回变量名
    const colors = [
      'var(--primary-color)',
      'var(--secondary-color)',
      'var(--danger-color)',
      'var(--success-color)',
      'var(--warning-color)'
    ]
    const name = userStore.nickname || userStore.username || 'User'
    let hash = 0
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash)
    }
    return colors[Math.abs(hash) % colors.length]
  })

  // 搜索建议（响应式追踪 searchHistory）
  const searchSuggestions = computed(() => {
    // 直接访问 searchHistory.value 以确保响应式更新
    const history = searchHistory.value
    const keyword = searchKeyword.value.trim()
    const maxResults = 5

    if (!keyword) {
      return history.slice(0, maxResults)
    }

    const lowerKeyword = keyword.toLowerCase()
    return history.filter(item => item.toLowerCase().includes(lowerKeyword)).slice(0, maxResults)
  })

  // 触发搜索（带防抖）
  const triggerSearch = (keyword: string) => {
    const trimmedKeyword = keyword.trim()

    // 添加到搜索历史
    if (trimmedKeyword) {
      addHistory(trimmedKeyword)
    }

    const currentPath = route.path

    // 根据当前页面决定搜索行为
    if (currentPath === '/files') {
      // 在 Files 页面，触发搜索事件
      const event = new CustomEvent('files-search', { detail: { keyword: trimmedKeyword } })
      window.dispatchEvent(event)
    } else if (currentPath === '/square') {
      // 在 Square 页面，触发搜索事件
      const event = new CustomEvent('square-search', { detail: { keyword: trimmedKeyword } })
      window.dispatchEvent(event)
    } else {
      // 不在 Files 或 Square 页面
      if (trimmedKeyword) {
        // 有关键词，跳转到 Files 页面并搜索
        router.push({
          path: '/files',
          query: { search: trimmedKeyword }
        })
      }
      // 没有关键词，不执行任何操作
    }

    showSuggestions.value = false
  }

  // 处理搜索建议选择
  const handleSuggestionSelect = (keyword: string) => {
    searchKeyword.value = keyword
    triggerSearch(keyword)
  }

  // 处理清除历史
  const handleClearHistory = () => {
    clearHistory()
  }

  // 处理删除历史
  const handleDeleteHistory = (keyword: string) => {
    removeHistory(keyword)
  }

  // 处理搜索框失焦
  const handleSearchBlur = () => {
    // 延迟隐藏，允许点击建议项
    setTimeout(() => {
      showSuggestions.value = false
    }, 200)
  }

  // 处理方向键导航
  const handleArrowDown = (e: KeyboardEvent) => {
    if (searchSuggestions.value.length === 0) return
    e.preventDefault()
    selectedSuggestionIndex.value = Math.min(selectedSuggestionIndex.value + 1, searchSuggestions.value.length - 1)
  }

  const handleArrowUp = (e: KeyboardEvent) => {
    if (searchSuggestions.value.length === 0) return
    e.preventDefault()
    selectedSuggestionIndex.value = Math.max(selectedSuggestionIndex.value - 1, -1)
  }

  // 处理输入事件（带防抖，500ms）
  const handleSearchInput = () => {
    // 清除之前的定时器
    if (searchTimer) {
      clearTimeout(searchTimer)
    }

    // 设置新的定时器，500ms 后执行搜索
    searchTimer = setTimeout(() => {
      triggerSearch(searchKeyword.value)
    }, 500)
  }

  // 处理清空事件
  const handleSearchClear = () => {
    // 清除定时器
    if (searchTimer) {
      clearTimeout(searchTimer)
      searchTimer = null
    }
    // 立即触发清空搜索
    triggerSearch('')
  }

  // 处理回车事件（立即搜索，不防抖）
  const handleSearch = () => {
    // 清除定时器
    if (searchTimer) {
      clearTimeout(searchTimer)
      searchTimer = null
    }
    // 立即触发搜索
    triggerSearch(searchKeyword.value)
  }

  // 处理搜索并关闭对话框
  const handleSearchAndClose = () => {
    handleSearch()
    showSearchDialog.value = false
  }

  // 处理搜索对话框关闭事件
  const handleSearchDialogClosed = () => {
    // 对话框关闭后，如果需要可以清空搜索关键词
    // 这里不清空，保留搜索关键词以便用户继续搜索
  }

  const handleCommand = (command: string) => {
    if (command === 'logout') {
      authStore.logout()
      router.push('/login')
      proxy?.$modal.msgSuccess(t('header.logoutSuccess'))
    } else if (command === 'settings') {
      router.push('/settings')
    }
  }

  const toggleSidebar = () => {
    // 触发侧边栏显示/隐藏事件
    const event = new CustomEvent('toggle-sidebar')
    window.dispatchEvent(event)
  }

  // 监听路由变化，切换页面时清空搜索框
  watch(
    () => route.path,
    (newPath, oldPath) => {
      // 当从其他页面切换到 Square 或 Files 页面时，清空搜索框
      if (oldPath && (newPath === '/square' || newPath === '/files')) {
        // 清除定时器
        if (searchTimer) {
          clearTimeout(searchTimer)
          searchTimer = null
        }
        // 清空搜索框
        searchKeyword.value = ''
        // 触发清空搜索事件
        if (newPath === '/files') {
          const event = new CustomEvent('files-search', { detail: { keyword: '' } })
          window.dispatchEvent(event)
        } else if (newPath === '/square') {
          const event = new CustomEvent('square-search', { detail: { keyword: '' } })
          window.dispatchEvent(event)
        }
      }
    }
  )

  // 监听搜索对话框显示，自动聚焦输入框
  watch(showSearchDialog, newVal => {
    if (newVal) {
      nextTick(() => {
        searchDialogInputRef.value?.focus()
      })
    }
  })

  // 组件卸载时清理定时器
  onBeforeUnmount(() => {
    if (searchTimer) {
      clearTimeout(searchTimer)
      searchTimer = null
    }
  })
</script>

<style scoped>
  .layout-header {
    height: 64px !important;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 24px;
    z-index: 100;
    position: relative;
    border-bottom: 1px solid var(--glass-border);
    flex-shrink: 0;
  }

  .header-left {
    min-width: 240px;
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .mobile-menu-btn {
    display: none !important;
  }

  @media (max-width: 1024px) {
    .mobile-menu-btn {
      display: inline-flex !important;
    }

    .header-left {
      min-width: auto;
      gap: 8px;
    }

    .logo-text {
      font-size: 18px;
    }

    /* 水平布局时，header-center 需要更多空间 */
    .header-center:has(.horizontal-menu) {
      max-width: none;
      flex: 1;
      margin: 0 8px;
      justify-content: flex-start;
    }
  }

  .logo-wrapper {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .logo-image {
    width: 65px;
    height: 65px;
    display: block;
    transition: transform 0.2s ease;
  }

  .logo-image :deep(.el-image__inner) {
    width: 65px;
    height: 65px;
    object-fit: contain;
    filter: drop-shadow(0 2px 4px rgba(99, 102, 241, 0.3));
    transition: transform 0.2s ease;
  }

  .logo-image:hover :deep(.el-image__inner) {
    transform: scale(1.05);
  }

  .logo-image-mobile {
    width: 36px;
    height: 36px;
    display: block;
  }

  .logo-image-mobile :deep(.el-image__inner) {
    width: 36px;
    height: 36px;
    object-fit: contain;
    filter: drop-shadow(0 2px 4px rgba(99, 102, 241, 0.3));
  }

  .logo-text {
    font-size: 20px;
    font-weight: 700;
    background: linear-gradient(135deg, var(--primary-color) 0%, var(--secondary-color) 100%);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
    letter-spacing: -0.5px;
  }

  .header-center {
    flex: 1;
    max-width: 500px;
    margin: 0 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 0;
    overflow: hidden;
  }

  /* 水平布局时，header-center 需要更多空间 */
  .header-center:has(.horizontal-menu) {
    max-width: none;
    flex: 1;
    margin: 0 16px;
    justify-content: flex-start;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  /* 统一头部操作按钮样式 */
  .header-action-btn {
    width: 36px;
    height: 36px;
    padding: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    transition: all 0.2s ease;
    color: var(--text-regular);
  }

  .header-action-btn:hover {
    background: var(--el-fill-color-light);
    color: var(--primary-color);
    transform: translateY(-1px);
  }

  html.dark .header-action-btn:hover {
    background: var(--el-fill-color-light);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
  }

  /* 主题切换按钮保持原有样式，但继承 header-action-btn 的基础样式 */

  .mobile-search-btn {
    display: none !important;
  }

  .mobile-logo {
    display: none !important;
  }

  .search-wrapper {
    position: relative;
    width: 100%;
    max-width: 600px;
  }

  .search-input :deep(.el-input__wrapper) {
    background: var(--bg-color-glass, rgba(255, 255, 255, 0.5));
    backdrop-filter: blur(8px);
    border-radius: 12px;
    padding-left: 16px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.2);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .search-input :deep(.el-input__wrapper):hover {
    background: var(--card-bg, rgba(255, 255, 255, 0.7));
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
    border-color: rgba(37, 99, 235, 0.3);
    transform: translateY(-1px);
  }

  .search-input :deep(.el-input__wrapper.is-focus) {
    background: var(--card-bg, var(--el-bg-color));
    box-shadow:
      0 6px 20px rgba(37, 99, 235, 0.15),
      0 0 0 3px rgba(37, 99, 235, 0.1);
    border-color: var(--primary-color);
    transform: translateY(-1px);
  }

  html.dark .search-input :deep(.el-input__wrapper) {
    background: rgba(30, 41, 59, 0.6);
    border-color: rgba(255, 255, 255, 0.1);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
  }

  html.dark .search-input :deep(.el-input__wrapper):hover {
    background: rgba(30, 41, 59, 0.8);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    border-color: rgba(59, 130, 246, 0.3);
  }

  html.dark .search-input :deep(.el-input__wrapper.is-focus) {
    background: var(--el-bg-color);
    box-shadow:
      0 6px 20px rgba(59, 130, 246, 0.25),
      0 0 0 3px rgba(59, 130, 246, 0.15);
    border-color: var(--primary-color);
  }

  .user-profile {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 12px;
    border-radius: 30px;
    cursor: pointer;
    background: transparent;
    transition: all 0.2s;
    border: 1px solid transparent;
  }

  .user-profile:hover {
    background: var(--el-fill-color-light, rgba(255, 255, 255, 0.6));
    border-color: var(--border-color);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  }

  html.dark .user-profile:hover {
    background: var(--el-fill-color-light);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
  }

  .username {
    font-weight: 600;
    font-size: 14px;
    color: var(--text-primary);
  }

  .desktop-only {
    display: inline;
  }

  .desktop-user {
    display: block;
  }

  /* 移动端响应式 */
  @media (max-width: 991px) {
    .layout-header {
      padding: 0 12px;
      height: 56px !important;
    }

    .header-left {
      flex-shrink: 0;
      gap: 8px;
      min-width: auto;
    }

    .logo-text {
      display: none;
    }

    .logo-image {
      width: 40px;
      height: 40px;
    }

    .logo-image :deep(.el-image__inner) {
      width: 40px;
      height: 40px;
    }

    .header-center {
      flex: 1;
      max-width: none;
      margin: 0 8px;
      min-width: 0;
      overflow: hidden;
    }

    .header-right {
      flex-shrink: 0;
      min-width: auto;
      display: flex;
      align-items: center;
      gap: 4px;
    }

    .header-action-btn {
      width: 32px;
      height: 32px;
    }

    .user-profile {
      padding: 4px;
      gap: 0;
    }

    .user-avatar-img {
      width: 28px !important;
      height: 28px !important;
    }

    .desktop-only {
      display: none;
    }
  }

  @media screen and (max-width: 480px) {
    .layout-header {
      padding: 0 8px !important;
      display: flex !important;
      align-items: center !important;
      justify-content: space-between !important;
      overflow: visible !important;
      flex-wrap: nowrap !important;
    }

    .header-center {
      display: none !important;
      width: 0 !important;
      margin: 0 !important;
      padding: 0 !important;
      flex: 0 0 0 !important;
    }

    .desktop-logo {
      display: none !important;
    }

    .header-left {
      flex: 0 0 auto !important;
      flex-shrink: 0 !important;
      flex-grow: 0 !important;
      gap: 8px !important;
      min-width: auto !important;
      max-width: none !important;
      width: auto !important;
      overflow: visible !important;
    }

    .header-right {
      display: flex !important;
      visibility: visible !important;
      opacity: 1 !important;
      align-items: center !important;
      justify-content: flex-end !important;
      gap: 6px !important;
      flex: 0 0 auto !important;
      flex-shrink: 0 !important;
      flex-grow: 0 !important;
      min-width: 120px !important;
      width: auto !important;
      max-width: none !important;
      margin-left: auto !important;
      padding: 0 !important;
      overflow: visible !important;
      position: relative !important;
      z-index: 10 !important;
    }

    .header-right > * {
      flex-shrink: 0 !important;
      flex-grow: 0 !important;
      flex: 0 0 auto !important;
    }

    .header-right .el-button,
    .header-right .mobile-search-btn {
      flex: 0 0 auto !important;
      flex-grow: 0 !important;
      flex-shrink: 0 !important;
      width: auto !important;
      min-width: auto !important;
      max-width: none !important;
    }

    .mobile-search-btn {
      display: inline-flex !important;
      visibility: visible !important;
      opacity: 1 !important;
      width: 36px !important;
      height: 36px !important;
      min-width: 36px !important;
      min-height: 36px !important;
      max-width: 36px !important;
      max-height: 36px !important;
      padding: 0 !important;
      margin: 0 !important;
      flex-shrink: 0 !important;
      position: relative !important;
      z-index: 1 !important;
    }

    .mobile-search-btn :deep(button),
    .mobile-search-btn :deep(.el-button) {
      display: inline-flex !important;
      visibility: visible !important;
      opacity: 1 !important;
      width: 36px !important;
      height: 36px !important;
      min-width: 36px !important;
      min-height: 36px !important;
    }

    .mobile-search-btn :deep(.el-icon) {
      display: inline-block !important;
      visibility: visible !important;
      opacity: 1 !important;
      font-size: 18px !important;
      width: 18px !important;
      height: 18px !important;
    }

    .mobile-logo {
      display: flex !important;
      visibility: visible !important;
      opacity: 1 !important;
      align-items: center !important;
      justify-content: center !important;
      width: 36px !important;
      height: 36px !important;
      min-width: 36px !important;
      min-height: 36px !important;
      max-width: 36px !important;
      max-height: 36px !important;
      flex-shrink: 0 !important;
      margin: 0 !important;
      padding: 0 !important;
      position: relative !important;
      z-index: 1 !important;
    }

    .mobile-logo .logo-image-mobile {
      display: block !important;
      visibility: visible !important;
      opacity: 1 !important;
      width: 36px !important;
      height: 36px !important;
    }

    .mobile-logo .logo-image-mobile :deep(.el-image__inner) {
      width: 36px !important;
      height: 36px !important;
    }

    .user-profile {
      display: flex !important;
      visibility: visible !important;
      opacity: 1 !important;
      padding: 0 !important;
      gap: 0 !important;
      flex-shrink: 0 !important;
      align-items: center !important;
      min-width: auto !important;
      position: relative !important;
      z-index: 1 !important;
    }

    .user-avatar-img {
      display: block !important;
      visibility: visible !important;
      opacity: 1 !important;
      flex-shrink: 0 !important;
      width: 32px !important;
      height: 32px !important;
      min-width: 32px !important;
      min-height: 32px !important;
      max-width: 32px !important;
      max-height: 32px !important;
    }
  }

  /* 搜索对话框样式 */
  .search-dialog :deep(.el-dialog) {
    border-radius: 24px;
    overflow: hidden;
    background: linear-gradient(135deg, rgba(255, 255, 255, 0.98) 0%, rgba(255, 255, 255, 1) 100%);
    backdrop-filter: blur(30px);
    box-shadow:
      0 24px 80px rgba(0, 0, 0, 0.12),
      0 8px 24px rgba(0, 0, 0, 0.08),
      0 0 0 1px rgba(255, 255, 255, 0.8);
    margin-top: 15vh !important;
    margin-bottom: auto !important;
    top: 0 !important;
    transform: translateY(0) !important;
    position: fixed !important;
  }

  .search-dialog :deep(.el-dialog__wrapper) {
    display: flex !important;
    align-items: flex-start !important;
    justify-content: center !important;
    padding-top: 0 !important;
  }

  .search-dialog :deep(.el-dialog__header) {
    padding: 28px 28px 20px;
    border-bottom: none;
    background: transparent;
  }

  .search-dialog :deep(.el-dialog__body) {
    padding: 0 28px 24px;
  }

  .search-dialog :deep(.el-dialog__footer) {
    padding: 20px 28px 28px;
    border-top: 1px solid rgba(0, 0, 0, 0.05);
    background: linear-gradient(to bottom, rgba(255, 255, 255, 0.6), rgba(255, 255, 255, 0.8));
  }

  .search-dialog-header {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 0;
  }

  .search-icon {
    font-size: 26px;
    color: var(--primary-color);
    filter: drop-shadow(0 3px 6px rgba(99, 102, 241, 0.4));
    animation: pulse 2s ease-in-out infinite;
  }

  @keyframes pulse {
    0%,
    100% {
      transform: scale(1);
      opacity: 1;
    }
    50% {
      transform: scale(1.05);
      opacity: 0.9;
    }
  }

  .search-title {
    font-size: 22px;
    font-weight: 700;
    background: linear-gradient(135deg, var(--primary-color) 0%, var(--secondary-color) 100%);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
    letter-spacing: -0.5px;
    text-shadow: 0 2px 4px rgba(99, 102, 241, 0.1);
  }

  .search-dialog-body {
    padding: 12px 0;
  }

  .search-dialog-input {
    width: 100%;
  }

  .search-dialog-input :deep(.el-input__wrapper) {
    background: linear-gradient(135deg, rgba(255, 255, 255, 0.95) 0%, rgba(255, 255, 255, 1) 100%);
    backdrop-filter: blur(12px);
    border-radius: 18px;
    padding: 14px 24px;
    box-shadow:
      0 6px 20px rgba(0, 0, 0, 0.06),
      inset 0 1px 2px rgba(255, 255, 255, 0.9),
      inset 0 -1px 2px rgba(0, 0, 0, 0.02);
    border: 2px solid rgba(99, 102, 241, 0.12);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    height: 60px;
  }

  .search-dialog-input :deep(.el-input__inner) {
    font-size: 17px;
    color: var(--text-primary);
    font-weight: 500;
  }

  .search-dialog-input :deep(.el-input__wrapper):hover {
    background: linear-gradient(135deg, rgba(255, 255, 255, 1) 0%, rgba(255, 255, 255, 1) 100%);
    border-color: rgba(99, 102, 241, 0.25);
    box-shadow:
      0 8px 24px rgba(99, 102, 241, 0.12),
      inset 0 1px 2px rgba(255, 255, 255, 0.9),
      inset 0 -1px 2px rgba(0, 0, 0, 0.02);
    transform: translateY(-1px);
  }

  .search-dialog-input :deep(.el-input__wrapper.is-focus) {
    background: var(--card-bg, var(--el-bg-color));
    border-color: var(--primary-color);
    box-shadow:
      0 12px 32px rgba(99, 102, 241, 0.18),
      0 0 0 5px rgba(99, 102, 241, 0.08),
      inset 0 1px 2px rgba(255, 255, 255, 0.9),
      inset 0 -1px 2px rgba(0, 0, 0, 0.02);
    transform: translateY(-2px);
  }

  html.dark .search-dialog-input :deep(.el-input__wrapper.is-focus) {
    box-shadow:
      0 12px 32px rgba(99, 102, 241, 0.3),
      0 0 0 5px rgba(99, 102, 241, 0.15),
      inset 0 1px 2px rgba(255, 255, 255, 0.1),
      inset 0 -1px 2px rgba(0, 0, 0, 0.1);
  }

  .search-dialog-input :deep(.el-input__prefix) {
    color: var(--primary-color);
    font-size: 22px;
    margin-right: 12px;
  }

  .search-dialog-input :deep(.el-input__suffix) {
    color: var(--text-secondary);
  }

  .search-dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  .cancel-btn {
    padding: 13px 28px;
    border-radius: 14px;
    font-weight: 600;
    font-size: 15px;
    border: 2px solid rgba(0, 0, 0, 0.08);
    background: linear-gradient(135deg, rgba(255, 255, 255, 0.9) 0%, rgba(255, 255, 255, 1) 100%);
    color: var(--text-primary);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  }

  .cancel-btn:hover {
    background: linear-gradient(135deg, rgba(0, 0, 0, 0.03) 0%, rgba(0, 0, 0, 0.05) 100%);
    border-color: rgba(0, 0, 0, 0.12);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
    transform: translateY(-1px);
  }

  .search-btn {
    padding: 13px 32px;
    border-radius: 14px;
    font-weight: 600;
    font-size: 15px;
    background: linear-gradient(135deg, var(--primary-color) 0%, var(--secondary-color) 100%);
    border: none;
    box-shadow:
      0 6px 20px rgba(99, 102, 241, 0.35),
      0 2px 8px rgba(99, 102, 241, 0.2);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    display: flex;
    align-items: center;
    gap: 8px;
    color: white;
  }

  .search-btn:hover {
    box-shadow:
      0 8px 28px rgba(99, 102, 241, 0.45),
      0 4px 12px rgba(99, 102, 241, 0.25);
    transform: translateY(-2px);
    background: linear-gradient(135deg, var(--primary-color) 0%, var(--secondary-color) 100%);
  }

  .search-btn:active {
    transform: translateY(0);
    box-shadow:
      0 4px 16px rgba(99, 102, 241, 0.35),
      0 2px 6px rgba(99, 102, 241, 0.2);
  }

  .search-btn :deep(.el-icon) {
    font-size: 18px;
  }

  /* 深色模式：下拉菜单 */
  html.dark .premium-dropdown {
    background-color: var(--el-bg-color);
    border-color: var(--el-border-color);
  }

  html.dark .premium-dropdown :deep(.el-dropdown-menu__item) {
    color: var(--el-text-color-primary);
    background-color: transparent;
  }

  html.dark .premium-dropdown :deep(.el-dropdown-menu__item:hover) {
    background-color: var(--el-fill-color-light);
    color: var(--primary-color);
  }

  html.dark .premium-dropdown :deep(.el-dropdown-menu__item.is-divided) {
    border-top-color: var(--el-border-color);
  }
</style>

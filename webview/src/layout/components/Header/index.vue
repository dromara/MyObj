<template>
  <el-header class="layout-header glass-panel">
    <div class="header-left">
      <!-- 移动端汉堡菜单按钮 -->
      <el-button
        class="mobile-menu-btn"
        icon="Menu"
        circle
        text
        @click="toggleSidebar"
      />
      <div class="logo-wrapper">
        <el-icon :size="32" class="logo-icon"><Folder /></el-icon>
        <span class="logo-text">MyObj 云盘</span>
      </div>
    </div>
    
    <div class="header-center">
      <div class="search-wrapper">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索文件、资料..."
          prefix-icon="Search"
          clearable
          @keyup.enter="handleSearch"
          class="search-input glass-input"
        />
      </div>
    </div>
    
    <div class="header-right">
      <!-- 移动端只显示头像，隐藏用户名 -->
      <el-dropdown @command="handleCommand" trigger="click">
        <div class="user-profile glass-hover">
          <el-avatar :size="32" :style="{ background: avatarColor }" class="user-avatar-img">
            {{ avatarText }}
          </el-avatar>
          <span class="username desktop-only">{{ userStore.nickname }}</span>
          <el-icon class="el-icon--right desktop-only"><CaretBottom /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu class="premium-dropdown">
            <el-dropdown-item command="settings">
              <el-icon><Setting /></el-icon>
              系统设置
            </el-dropdown-item>
            <el-dropdown-item divided command="logout">
              <el-icon><SwitchButton /></el-icon>
              退出登录
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </el-header>
</template>

<script setup lang="ts">
import { useUserStore } from '@/stores/user'
import { useAuthStore } from '@/stores/auth'

const { proxy } = getCurrentInstance() as ComponentInternalInstance

const router = useRouter()
const userStore = useUserStore()
const authStore = useAuthStore()

const searchKeyword = ref('')

const avatarText = computed(() => {
  return userStore.nickname ? userStore.nickname.charAt(0).toUpperCase() : 'U'
})

const avatarColor = computed(() => {
  const colors = ['#6366f1', '#8b5cf6', '#ec4899', '#10b981', '#f59e0b']
  const name = userStore.nickname
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return colors[Math.abs(hash) % colors.length]
})

const handleSearch = () => {
  proxy?.$log.debug('Search:', searchKeyword.value)
}

const handleCommand = (command: string) => {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
    proxy?.$modal.msgSuccess('已退出登录')
  } else if (command === 'settings') {
    router.push('/settings')
  }
}

const toggleSidebar = () => {
  // 触发侧边栏显示/隐藏事件
  const event = new CustomEvent('toggle-sidebar')
  window.dispatchEvent(event)
}
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
}

.logo-wrapper {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-icon {
  color: var(--primary-color);
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
}

.search-input :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.5);
  backdrop-filter: blur(4px);
  border-radius: 12px;
  padding-left: 16px;
  box-shadow: none;
  border: 1px solid transparent;
  transition: all 0.3s ease;
}

.search-input :deep(.el-input__wrapper):hover,
.search-input :deep(.el-input__wrapper.is-focus) {
  background: white;
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
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
  background: rgba(255, 255, 255, 0.6);
  border-color: var(--border-color);
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}

.username {
  font-weight: 600;
  font-size: 14px;
  color: var(--text-primary);
}

.desktop-only {
  display: inline;
}

/* 移动端响应式 */
@media (max-width: 1024px) {
  .layout-header {
    padding: 0 12px;
  }
  
  .header-center {
    flex: 1;
    max-width: none;
    margin: 0 12px;
  }
  
  .header-right {
    flex-shrink: 0;
  }
  
  .user-profile {
    padding: 4px;
    gap: 0;
  }
  
  .desktop-only {
    display: none;
  }
}

@media (max-width: 480px) {
  .header-center {
    display: none;
  }
}
</style>


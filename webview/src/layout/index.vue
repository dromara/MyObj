<template>
  <div class="layout-container">
    <el-container>
      <!-- Premium Glass Header -->
      <el-header class="layout-header glass-panel">
        <div class="header-left">
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
              :prefix-icon="Search"
              clearable
              @keyup.enter="handleSearch"
              class="search-input glass-input"
            />
          </div>
        </div>
        
        <div class="header-right">
          <el-dropdown @command="handleCommand" trigger="click">
            <div class="user-profile glass-hover">
              <el-avatar :size="32" :style="{ background: getAvatarColor(userInfo.name) }" class="user-avatar-img">
                {{ getAvatarText(userInfo.name) }}
              </el-avatar>
              <span class="username">{{ userInfo.name }}</span>
              <el-icon class="el-icon--right"><CaretBottom /></el-icon>
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
      
      <el-container class="main-container">
        <!-- Clean Sidebar -->
        <el-aside width="240px" class="layout-aside">
          <el-menu
            :default-active="currentRoute"
            router
            @select="handleMenuSelect"
            class="premium-menu"
          >
            <el-menu-item index="/files">
              <el-icon><Folder /></el-icon>
              <span>我的文件</span>
            </el-menu-item>
            <el-menu-item index="/shares">
              <el-icon><Share /></el-icon>
              <span>我的分享</span>
            </el-menu-item>
            <el-menu-item index="/offline">
              <el-icon><Download /></el-icon>
              <span>离线下载</span>
            </el-menu-item>
            <el-menu-item index="/tasks">
              <el-icon><List /></el-icon>
              <span>传输列表</span>
            </el-menu-item>
            <el-menu-item index="/trash">
              <el-icon><Delete /></el-icon>
              <span>回收站</span>
            </el-menu-item>
            <div class="menu-divider"></div>
            <el-menu-item index="/square">
              <el-icon><Grid /></el-icon>
              <span>文件广场</span>
            </el-menu-item>
          </el-menu>
          
          <!-- Storage Info -->
          <div class="storage-card glass-panel-sm">
            <div class="storage-header">
              <span class="storage-title">存储空间</span>
              <span class="storage-text" v-if="storageInfo.isUnlimited">无限容量</span>
              <span class="storage-text" v-else>{{ storageInfo.percentage }}%</span>
            </div>
            <el-progress 
              :percentage="storageInfo.isUnlimited ? 100 : storageInfo.percentage" 
              :color="storageInfo.isUnlimited ? '#8b5cf6' : customColors"
              :show-text="false"
              :stroke-width="8" 
              class="storage-progress"
            />
            <div class="storage-detail">
              {{ formatStorageSize(storageInfo.used) }} / {{ storageInfo.isUnlimited ? '∞' : formatStorageSize(storageInfo.total) }}
            </div>
            <el-button type="primary" link class="upgrade-btn">升级扩容</el-button>
          </div>
        </el-aside>
        
        <!-- Main Content -->
        <el-main class="layout-main">
          <router-view v-slot="{ Component }">
            <transition name="fade-scale" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Folder,
  FolderAdd,
  Search,
  Upload,
  Download,
  User,
  Setting,
  SwitchButton,
  Share,
  Delete,
  Grid,
  List,
  CaretBottom,
  Star
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import md5 from 'js-md5'

const router = useRouter()
const route = useRoute()

const searchKeyword = ref('')
const userInfo = ref({
  name: '',
  username: '',
  email: ''
})

const storageInfo = ref({
  used: 0,
  total: 0,
  percentage: 0,
  isUnlimited: false
})

const customColors = [
  { color: '#10b981', percentage: 60 },
  { color: '#f59e0b', percentage: 80 },
  { color: '#ef4444', percentage: 100 },
]

// Mock Interfaces matching original file logic
interface UserInfo {
  name: string;
  username: string;
  email: string;
}

// Keeping original logic structure
const initUserInfo = async () => {
  try {
    const res = await request.get('/user/info')
    if (res.code === 200) {
      userInfo.value = res.data
      updateStorageInfo(res.data)
    }
  } catch (error) {
    console.error('获取用户信息失败', error)
  }
}

const updateStorageInfo = (userInfo: any) => {
  if (userInfo.capacity) {
     storageInfo.value.total = userInfo.capacity
     storageInfo.value.used = userInfo.used || 0
     storageInfo.value.isUnlimited = userInfo.capacity === -1
     
     if (!storageInfo.value.isUnlimited && storageInfo.value.total > 0) {
       storageInfo.value.percentage = Math.ceil((storageInfo.value.used / storageInfo.value.total) * 100)
     } else {
       storageInfo.value.percentage = 0
     }
  } else if (userInfo.storage_limit !== undefined) {
    // Backwards compatibility with previous code structure if needed
    if (userInfo.storage_limit === -1) {
      storageInfo.value.isUnlimited = true
      storageInfo.value.total = 0
    } else {
      storageInfo.value.isUnlimited = false
      storageInfo.value.total = userInfo.storage_limit
    }
    storageInfo.value.used = userInfo.used_storage || 0
    if (!storageInfo.value.isUnlimited && storageInfo.value.total > 0) {
      storageInfo.value.percentage = Math.floor((storageInfo.value.used / storageInfo.value.total) * 100)
    } else {
      storageInfo.value.percentage = 0
    }
  }
}

const formatStorageSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const currentRoute = computed(() => route.path)

const handleSearch = () => {
  // Global search or emit event
  console.log('Search:', searchKeyword.value)
  // Implement actual search routing if needed, or event bus
}

const handleCommand = (command: string) => {
  if (command === 'logout') {
    localStorage.removeItem('token')
    router.push('/login')
    ElMessage.success('已退出登录')
  } else if (command === 'settings') {
    ElMessage.info('设置功能开发中')
  }
}

const handleMenuSelect = (index: string) => {
  // Router handles navigation automatically
}

const getAvatarText = (name: string) => {
  return name ? name.charAt(0).toUpperCase() : 'U'
}

const getAvatarColor = (name: string) => {
  const colors = ['#6366f1', '#8b5cf6', '#ec4899', '#10b981', '#f59e0b']
  // Simple hash to pick color
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return colors[Math.abs(hash) % colors.length]
}

onMounted(() => {
  initUserInfo()
})
</script>

<style scoped>
.layout-container {
  width: 100%;
  height: 100vh;
  background: var(--bg-color);
  overflow: hidden;
}

.layout-header {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  z-index: 100;
  position: relative;
  /* Glassmorphism handled by .glass-panel global class */
  border-bottom: 1px solid var(--glass-border);
}

.header-left {
  min-width: 240px;
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

.main-container {
  height: calc(100vh - 64px);
}

.layout-aside {
  background: white; /* Clean white sidebar */
  /* No border right, using shadow */
  box-shadow: 4px 0 24px rgba(0, 0, 0, 0.02);
  display: flex;
  flex-direction: column;
  padding: 16px 0;
  z-index: 10;
}

.premium-menu {
  border: none;
  flex: 1;
  padding: 0 12px;
  background: transparent;
}

.premium-menu :deep(.el-menu-item) {
  height: 48px;
  margin-bottom: 4px;
  border-radius: 10px;
  color: var(--text-regular);
  font-weight: 500;
  border: none;
}

.premium-menu :deep(.el-menu-item:hover) {
  background: rgba(99, 102, 241, 0.08); /* Primary color light */
  color: var(--primary-color);
}

.premium-menu :deep(.el-menu-item.is-active) {
  background: linear-gradient(135deg, #2563eb 0%, #4f46e5 100%); /* Blue to Indigo match */
  color: white;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
}

.premium-menu :deep(.el-icon) {
  font-size: 18px;
  margin-right: 12px;
}

.menu-divider {
  height: 1px;
  background: var(--border-light);
  margin: 12px 16px;
}

.storage-card {
  margin: 16px;
  padding: 20px;
  border-radius: 16px;
  background: var(--bg-color-overlay, #f8fafc); /* Use var or fallback */
  border: 1px solid var(--border-light, #e2e8f0);
}

.storage-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 13px;
  color: var(--text-primary, #333);
  font-weight: 600;
}

.storage-progress {
  margin-bottom: 8px;
}

.storage-detail {
  font-size: 12px;
  color: var(--text-secondary, #666);
  text-align: right;
  margin-bottom: 16px;
  font-family: var(--font-family-mono, monospace); /* Optional: distinct font for numbers */
}

.upgrade-btn {
  width: 100%;
  border-radius: 8px;
  height: 36px;
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(99, 102, 241, 0.1);
  color: var(--primary-color);
  transition: all 0.2s;
}

.upgrade-btn:hover {
  background: rgba(99, 102, 241, 0.2);
  transform: translateY(-1px);
}

.layout-main {
  padding: 24px;
  background: var(--bg-color);
  overflow-y: auto;
}

/* Transitions */
.fade-scale-enter-active,
.fade-scale-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-scale-enter-from {
  opacity: 0;
  transform: scale(0.98);
}

.fade-scale-leave-to {
  opacity: 0;
  transform: scale(1.02);
}
</style>

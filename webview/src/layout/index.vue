<template>
  <div class="layout-container">
    <el-container>
      <!-- 顶部导航 -->
      <el-header class="layout-header">
        <div class="header-left">
          <el-icon :size="28" color="#409EFF"><Folder /></el-icon>
          <span class="logo-text">MyObj 网盘</span>
        </div>
        
        <div class="header-center">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索文件..."
            :prefix-icon="Search"
            clearable
            @keyup.enter="handleSearch"
            class="search-input"
          />
        </div>
        
        <div class="header-right">
          <el-tooltip content="上传文件" placement="bottom">
            <el-button type="primary" :icon="Upload" @click="handleUpload">上传</el-button>
          </el-tooltip>
          
          <el-dropdown @command="handleCommand">
            <div class="user-avatar">
              <el-avatar :size="36" :style="{ background: getAvatarColor(userInfo.name) }">
                {{ getAvatarText(userInfo.name) }}
              </el-avatar>
              <span class="username">{{ userInfo.name }}</span>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="settings">
                  <el-icon><Setting /></el-icon>
                  设置
                </el-dropdown-item>
                <el-dropdown-item command="logout" divided>
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      
      <el-container class="main-container">
        <!-- 侧边栏 -->
        <el-aside width="220px" class="layout-aside">
          <el-menu
            :default-active="currentRoute"
            router
            @select="handleMenuSelect"
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
              <span>任务中心</span>
            </el-menu-item>
            <el-menu-item index="/trash">
              <el-icon><Delete /></el-icon>
              <span>回收站</span>
            </el-menu-item>
            <el-menu-item index="/square">
              <el-icon><Grid /></el-icon>
              <span>文件广场</span>
            </el-menu-item>
          </el-menu>
          
          <!-- 存储空间信息 -->
          <div class="storage-info">
            <div class="storage-header">
              <span class="storage-title">存储空间</span>
              <span class="storage-text" v-if="storageInfo.isUnlimited">无限制</span>
              <span class="storage-text" v-else>{{ formatStorageSize(storageInfo.used) }} / {{ formatStorageSize(storageInfo.total) }}</span>
            </div>
            <el-progress 
              v-if="!storageInfo.isUnlimited"
              :percentage="storageInfo.percentage" 
              :stroke-width="6" 
            />
            <div v-if="storageInfo.isUnlimited" class="unlimited-badge">
              <el-icon><Star /></el-icon>
              <span>无限容量</span>
            </div>
          </div>
        </el-aside>
        
        <!-- 主内容区 -->
        <el-main class="layout-main">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
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
  Magnet,
  Grid,
  List,
  Star
} from '@element-plus/icons-vue'

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

// 初始化用户信息
const initUserInfo = () => {
  const storedUserInfo = localStorage.getItem('userInfo')
  if (storedUserInfo) {
    try {
      const parsed = JSON.parse(storedUserInfo)
      userInfo.value = {
        name: parsed.name || parsed.user_name || '用户',
        username: parsed.user_name || '',
        email: parsed.email || ''
      }
      
      // 更新存储信息
      updateStorageInfo(parsed)
    } catch (error) {
      console.error('解析用户信息失败:', error)
      userInfo.value.name = localStorage.getItem('username') || '用户'
    }
  } else {
    userInfo.value.name = localStorage.getItem('username') || '用户'
  }
}

// 更新存储空间信息
const updateStorageInfo = (userInfo: any) => {
  const totalSpace = userInfo.space || 0
  const freeSpace = userInfo.free_space || 0
  
  // 如果总空间小于等于0，视为无限制
  if (totalSpace <= 0) {
    storageInfo.value = {
      used: 0,
      total: 0,
      percentage: 0,
      isUnlimited: true
    }
  } else {
    const usedSpace = totalSpace - freeSpace
    const percentage = totalSpace > 0 ? Math.min((usedSpace / totalSpace) * 100, 100) : 0
    
    storageInfo.value = {
      used: usedSpace,
      total: totalSpace,
      percentage: Math.round(percentage * 10) / 10,
      isUnlimited: false
    }
  }
}

// 格式化存储大小
const formatStorageSize = (bytes: number): string => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

const currentRoute = computed(() => route.path)

// 判断是否在广场页面
const isSquarePage = computed(() => route.path === '/square')

const handleSearch = () => {
  if (searchKeyword.value) {
    if (isSquarePage.value) {
      // 在广场页面，触发广场搜索
      router.push({
        path: '/square',
        query: { keyword: searchKeyword.value }
      })
    } else {
      // 在其他页面，搜索当前用户文件
      ElMessage.success(`搜索: ${searchKeyword.value}`)
      // TODO: 调用搜索用户文件API
    }
  }
}

const handleUpload = () => {
  ElMessage.info('上传功能')
}

const handleCommand = (command: string) => {
  if (command === 'logout') {
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    localStorage.removeItem('userInfo')
    router.push('/login')
  } else if (command === 'settings') {
    ElMessage.info('设置功能')
  }
}

const handleMenuSelect = (index: string) => {
  console.log('选中菜单:', index)
}

// 获取头像文字（取名字的第一个字符）
const getAvatarText = (name: string) => {
  return name ? name.charAt(0).toUpperCase() : 'U'
}

// 根据名字生成颜色
const getAvatarColor = (name: string) => {
  const colors = [
    '#409EFF', '#67C23A', '#E6A23C', '#F56C6C', '#909399',
    '#0078d4', '#00b7c3', '#8764b8', '#498205', '#ff8c00'
  ]
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return colors[Math.abs(hash) % colors.length]
}

// 页面加载时初始化
import { onMounted } from 'vue'

onMounted(() => {
  initUserInfo()
})
</script>

<style scoped>
.layout-container {
  width: 100%;
  height: 100vh;
}

.el-container {
  height: 100%;
}

.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: white;
  border-bottom: 1px solid var(--el-border-color);
  padding: 0 24px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.header-center {
  flex: 1;
  max-width: 600px;
  margin: 0 40px;
}

.search-input {
  width: 100%;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-avatar {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: 8px;
  transition: all 0.3s;
}

.user-avatar:hover {
  background: var(--el-fill-color-light);
}

.username {
  font-size: 14px;
  font-weight: 500;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.main-container {
  height: calc(100vh - 60px);
}

.layout-aside {
  background: white;
  border-right: 1px solid var(--el-border-color);
  display: flex;
  flex-direction: column;
}

.el-menu {
  border: none;
  flex: 1;
  padding: 8px 0;
}

.el-menu-item {
  margin: 4px 12px;
  border-radius: 8px;
  height: 42px;
  line-height: 42px;
}

.el-menu-item:hover {
  background: var(--el-fill-color-light);
}

.el-menu-item.is-active {
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-weight: 500;
}

.storage-info {
  padding: 16px;
  margin: 16px 12px;
  background: var(--el-fill-color-light);
  border-radius: 8px;
}

.storage-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
}

.storage-title {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.storage-text {
  font-size: 14px;
  font-weight: 500;
}

.upgrade-btn {
  width: 100%;
  margin-top: 12px;
}

.unlimited-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-top: 12px;
  padding: 8px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
}

.layout-main {
  background: #f5f7fa;
  padding: 24px;
}
</style>

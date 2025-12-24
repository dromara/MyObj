<script setup lang="ts">
import { useUserStore } from '@/stores/user'
import { useAuthStore } from '@/stores/auth'

const { proxy } = getCurrentInstance() as ComponentInternalInstance
const router = useRouter()
const userStore = useUserStore()
const authStore = useAuthStore()

const emit = defineEmits<{
  search: [keyword: string]
  upload: []
  'new-folder': []
  'offline-download': []
  'torrent-download': []
}>()

const searchInput = ref('')

const handleSearch = () => {
  emit('search', searchInput.value)
}

const handleUpload = () => {
  emit('upload')
}

const handleNewFolder = () => {
  emit('new-folder')
}

const handleOfflineDownload = () => {
  emit('offline-download')
}

const handleTorrentDownload = () => {
  emit('torrent-download')
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
</script>

<template>
  <div class="navbar">
    <div class="navbar-left">
      <div class="logo">
        <el-icon :size="32" color="var(--primary-color)">
          <Folder />
        </el-icon>
        <span class="logo-text">云盘系统</span>
      </div>
    </div>
    
      <div class="navbar-center">
      <el-input
        v-model="searchInput"
        placeholder="搜索文件..."
        clearable
        @keyup.enter="handleSearch"
        @clear="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
    </div>
    
    <div class="navbar-right">
      <el-button
        icon="FolderAdd"
        circle
        @click="handleNewFolder"
        title="新建文件夹"
      />
      
      <el-button
        icon="Download"
        circle
        @click="handleOfflineDownload"
        title="离线下载"
      />
      
      <el-button
        icon="Magnet"
        circle
        @click="handleTorrentDownload"
        title="种子下载"
      />
      
      <el-button
        type="primary"
        icon="Upload"
        @click="handleUpload"
      >
        上传
      </el-button>
      
      <el-dropdown @command="handleCommand">
        <div class="user-info">
          <el-avatar :size="32" icon="User" />
          <span class="username">{{ userStore.username || '用户' }}</span>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item :command="'settings'" icon="Setting">
              设置
            </el-dropdown-item>
            <el-dropdown-item :command="'logout'" icon="SwitchButton" divided>
              退出登录
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<style scoped>
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  background: white;
  border-bottom: 1px solid var(--border-color);
  padding: 0 24px;
}

.navbar-left {
  display: flex;
  align-items: center;
}

.logo {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.navbar-center {
  flex: 1;
  display: flex;
  justify-content: center;
  max-width: 600px;
  margin: 0 40px;
}

.navbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 12px;
  border-radius: 20px;
  transition: background 0.3s;
}

.user-info:hover {
  background: var(--bg-color);
}

.username {
  font-size: 14px;
  color: var(--text-primary);
}

/* 移动端响应式 */
@media (max-width: 1024px) {
  .navbar {
    padding: 0 12px;
    height: 56px;
  }
  
  .logo-text {
    font-size: 16px;
  }
  
  .navbar-center {
    display: none;
  }
  
  .navbar-right {
    gap: 8px;
  }
  
  .navbar-right .el-button {
    padding: 8px;
  }
  
  .navbar-right .el-button span {
    display: none;
  }
  
  .user-info {
    padding: 4px 8px;
  }
  
  .username {
    display: none;
  }
}

@media (max-width: 480px) {
  .navbar {
    padding: 0 8px;
  }
  
  .logo {
    gap: 4px;
  }
  
  .logo .el-icon {
    font-size: 24px;
  }
  
  .logo-text {
    font-size: 14px;
  }
  
  .navbar-right {
    gap: 4px;
  }
  
  .navbar-right .el-button {
    padding: 6px;
  }
}
</style>

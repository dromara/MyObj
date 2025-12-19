<script setup lang="ts">
const { proxy } = getCurrentInstance() as ComponentInternalInstance

const emit = defineEmits<{
  search: [keyword: string]
  upload: []
  'new-folder': []
  'offline-download': []
  'torrent-download': []
}>()

const searchInput = ref('')
const userInfo = ref({
  username: proxy?.$cache.local.get('username') || '用户',
  avatar: ''
})

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
    proxy?.$cache.local.remove('token')
    proxy?.$cache.local.remove('username')
    window.location.href = '/login'
  } else if (command === 'settings') {
    proxy?.$modal.msg('设置功能')
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
          <span class="username">{{ userInfo.username }}</span>
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
</style>

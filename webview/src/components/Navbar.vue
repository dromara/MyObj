<script setup>
import { ref } from 'vue'

const emit = defineEmits(['search', 'upload', 'new-folder', 'offline-download', 'torrent-download'])
const searchInput = ref('')
const userInfo = ref({
  username: localStorage.getItem('username') || '用户',
  avatar: ''
})
const showUserMenu = ref(false)

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

const handleUserClick = () => {
  showUserMenu.value = !showUserMenu.value
}

const handleLogout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('username')
  window.location.href = '/login'
}

const handleSettings = () => {
  alert('设置功能')
}
</script>

<template>
  <div class="navbar">
    <div class="navbar-left">
      <div class="logo">
        <svg viewBox="0 0 24 24" width="32" height="32">
          <path fill="currentColor" d="M10 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/>
        </svg>
        <span class="logo-text">云盘系统</span>
      </div>
    </div>
    
    <div class="navbar-center">
      <div class="search-box">
        <input 
          v-model="searchInput" 
          type="text" 
          placeholder="搜索文件..."
          @keyup.enter="handleSearch"
        />
        <button @click="handleSearch">
          <svg viewBox="0 0 24 24" width="20" height="20">
            <path fill="currentColor" d="M15.5 14h-.79l-.28-.27C15.41 12.59 16 11.11 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/>
          </svg>
        </button>
      </div>
    </div>
    
    <div class="navbar-right">
      <button class="icon-btn" @click="handleNewFolder" title="新建文件夹">
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path fill="currentColor" d="M20 6h-8l-2-2H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm-1 8h-3v3h-2v-3h-3v-2h3V9h2v3h3v2z"/>
        </svg>
      </button>
      
      <button class="icon-btn" @click="handleOfflineDownload" title="离线下载">
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path fill="currentColor" d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/>
        </svg>
      </button>
      
      <button class="icon-btn" @click="handleTorrentDownload" title="种子下载">
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path fill="currentColor" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
        </svg>
      </button>
      
      <button class="upload-btn" @click="handleUpload">
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path fill="currentColor" d="M9 16h6v-6h4l-7-7-7 7h4zm-4 2h14v2H5z"/>
        </svg>
        上传
      </button>
      
      <div class="user-info" @click="handleUserClick">
        <div class="avatar">
          <svg viewBox="0 0 24 24" width="32" height="32">
            <path fill="currentColor" d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
          </svg>
        </div>
        <span class="username">{{ userInfo.username }}</span>
        
        <div v-if="showUserMenu" class="user-menu">
          <div class="menu-item" @click.stop="handleSettings">
            <svg viewBox="0 0 24 24" width="18" height="18">
              <path fill="currentColor" d="M19.14 12.94c.04-.3.06-.61.06-.94 0-.32-.02-.64-.07-.94l2.03-1.58c.18-.14.23-.41.12-.61l-1.92-3.32c-.12-.22-.37-.29-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94L14.4 2.81c-.04-.24-.24-.41-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L2.74 8.87c-.12.21-.08.47.12.61l2.03 1.58c-.05.3-.09.63-.09.94s.02.64.07.94l-2.03 1.58c-.18.14-.23.41-.12.61l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.47-.12-.61l-2.01-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z"/>
            </svg>
            <span>设置</span>
          </div>
          <div class="menu-item" @click.stop="handleLogout">
            <svg viewBox="0 0 24 24" width="18" height="18">
              <path fill="currentColor" d="M17 7l-1.41 1.41L18.17 11H8v2h10.17l-2.58 2.58L17 17l5-5zM4 5h8V3H4c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h8v-2H4V5z"/>
            </svg>
            <span>退出登录</span>
          </div>
        </div>
      </div>
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
  color: var(--primary-color);
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

.search-box {
  display: flex;
  align-items: center;
  width: 100%;
  max-width: 500px;
  background: var(--bg-color);
  border-radius: 20px;
  padding: 0 16px;
  gap: 8px;
}

.search-box input {
  flex: 1;
  border: none;
  background: transparent;
  outline: none;
  padding: 10px 0;
  font-size: 14px;
  color: var(--text-primary);
}

.search-box button {
  background: none;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  padding: 4px;
}

.search-box button:hover {
  color: var(--primary-color);
}

.navbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: var(--bg-color);
  border: none;
  border-radius: 6px;
  cursor: pointer;
  color: var(--text-secondary);
  transition: all 0.3s;
}

.icon-btn:hover {
  background: var(--primary-color);
  color: white;
  transform: translateY(-1px);
}

.upload-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 20px;
  background: var(--primary-color);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.upload-btn:hover {
  background: #66b1ff;
  transform: translateY(-1px);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 12px;
  border-radius: 20px;
  transition: background 0.3s;
  position: relative;
}

.user-info:hover {
  background: var(--bg-color);
}

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--bg-color);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
}

.username {
  font-size: 14px;
  color: var(--text-primary);
}

.user-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 8px;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 150px;
  z-index: 1000;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  transition: background 0.3s;
  color: var(--text-primary);
}

.menu-item:hover {
  background: var(--bg-color);
}

.menu-item:first-child {
  border-radius: 8px 8px 0 0;
}

.menu-item:last-child {
  border-radius: 0 0 8px 8px;
}

.menu-item span {
  font-size: 14px;
}
</style>

<script setup>
import { ref } from 'vue'

const emit = defineEmits(['menu-click'])

const menuItems = ref([
  { id: 1, name: '我的文件', icon: 'folder', active: true, type: 'files' },
  { id: 4, name: '我的分享', icon: 'share', active: false, type: 'shares' },
  { id: 5, name: '离线下载', icon: 'download', active: false, type: 'offline' },
  { id: 6, name: '回收站', icon: 'trash', active: false, type: 'trash' }
])

const storageInfo = ref({
  used: 25.6,
  total: 100,
  percentage: 25.6
})

const handleMenuClick = (item) => {
  menuItems.value.forEach(menu => menu.active = false)
  item.active = true
  emit('menu-click', item.name)
}

const getIcon = (iconName) => {
  const icons = {
    folder: 'M10 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z',
    share: 'M18 16.08c-.76 0-1.44.3-1.96.77L8.91 12.7c.05-.23.09-.46.09-.7s-.04-.47-.09-.7l7.05-4.11c.54.5 1.25.81 2.04.81 1.66 0 3-1.34 3-3s-1.34-3-3-3-3 1.34-3 3c0 .24.04.47.09.7L8.04 9.81C7.5 9.31 6.79 9 6 9c-1.66 0-3 1.34-3 3s1.34 3 3 3c.79 0 1.5-.31 2.04-.81l7.12 4.16c-.05.21-.08.43-.08.65 0 1.61 1.31 2.92 2.92 2.92 1.61 0 2.92-1.31 2.92-2.92s-1.31-2.92-2.92-2.92z',
    download: 'M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z',
    trash: 'M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z'
  }
  return icons[iconName] || icons.folder
}
</script>

<template>
  <div class="sidebar">
    <div class="menu-list">
      <div 
        v-for="item in menuItems" 
        :key="item.id"
        class="menu-item"
        :class="{ active: item.active }"
        @click="handleMenuClick(item)"
      >
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path fill="currentColor" :d="getIcon(item.icon)" />
        </svg>
        <span>{{ item.name }}</span>
      </div>
    </div>
    
    <div class="storage-info">
      <div class="storage-header">
        <span class="storage-title">存储空间</span>
        <span class="storage-text">{{ storageInfo.used }}GB / {{ storageInfo.total }}GB</span>
      </div>
      <div class="storage-bar">
        <div class="storage-bar-fill" :style="{ width: storageInfo.percentage + '%' }"></div>
      </div>
      <button class="upgrade-btn">升级容量</button>
    </div>
  </div>
</template>

<style scoped>
.sidebar {
  width: 220px;
  background: white;
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  padding: 16px 0;
}

.menu-list {
  flex: 1;
  padding: 0 12px;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  margin-bottom: 4px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
  color: var(--text-regular);
}

.menu-item:hover {
  background: var(--bg-color);
  color: var(--primary-color);
}

.menu-item.active {
  background: #ecf5ff;
  color: var(--primary-color);
}

.menu-item span {
  font-size: 14px;
  font-weight: 500;
}

.storage-info {
  padding: 16px;
  margin: 16px 12px 0;
  background: var(--bg-color);
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
  color: var(--text-secondary);
}

.storage-text {
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 500;
}

.storage-bar {
  width: 100%;
  height: 6px;
  background: #e4e7ed;
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 12px;
}

.storage-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--primary-color), #66b1ff);
  border-radius: 3px;
  transition: width 0.3s;
}

.upgrade-btn {
  width: 100%;
  padding: 8px;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.3s;
}

.upgrade-btn:hover {
  background: var(--primary-color);
  color: white;
  border-color: var(--primary-color);
}
</style>

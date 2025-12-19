<script setup lang="ts">

interface MenuItem {
  id: number
  name: string
  icon: string
  active: boolean
  type: string
}

const emit = defineEmits<{
  'menu-click': [name: string]
}>()

const menuItems = ref<MenuItem[]>([
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

const handleMenuClick = (item: MenuItem) => {
  menuItems.value.forEach((menu: MenuItem) => menu.active = false)
  item.active = true
  emit('menu-click', item.name)
}

const handleMenuSelect = (index: string) => {
  const item = menuItems.value.find((m: MenuItem) => m.id.toString() === index)
  if (item) {
    handleMenuClick(item)
  }
}

// 返回图标组件名称（已全局注册，可直接使用字符串）
const getIcon = (iconName: string) => {
  const iconMap: Record<string, string> = {
    folder: 'Folder',
    share: 'Share',
    download: 'Download',
    trash: 'Delete'
  }
  return iconMap[iconName] || 'Folder'
}
</script>

<template>
  <div class="sidebar">
    <el-menu
      :default-active="menuItems.find((item: MenuItem) => item.active)?.id.toString()"
      class="menu-list"
      @select="handleMenuSelect"
    >
      <el-menu-item
        v-for="item in menuItems"
        :key="item.id"
        :index="item.id.toString()"
      >
        <el-icon><component :is="getIcon(item.icon)" /></el-icon>
        <span>{{ item.name }}</span>
      </el-menu-item>
    </el-menu>
    
    <div class="storage-info">
      <div class="storage-header">
        <span class="storage-title">存储空间</span>
        <span class="storage-text">{{ storageInfo.used }}GB / {{ storageInfo.total }}GB</span>
      </div>
      <el-progress
        :percentage="storageInfo.percentage"
        :stroke-width="6"
        :show-text="false"
        class="storage-bar"
      />
      <el-button
        type="primary"
        size="small"
        class="upgrade-btn"
        style="width: 100%"
      >
        升级容量
      </el-button>
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
  border-right: none;
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
  margin-bottom: 12px;
}
</style>

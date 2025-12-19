<template>
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
    <el-button v-if="!storageInfo.isUnlimited" type="primary" link class="upgrade-btn">升级扩容</el-button>
  </div>
</template>

<script setup lang="ts">
const { proxy } = getCurrentInstance() as ComponentInternalInstance

interface StorageInfo {
  used: number
  total: number
  percentage: number
  isUnlimited: boolean
}

const storageInfo = ref<StorageInfo>({
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

const formatStorageSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const updateStorageInfo = (info: any) => {
  // 基于 UserInfo 接口映射: space (总容量), free_space (剩余空间)
  if (info.space !== undefined) {
    const total = Number(info.space)
    const free = Number(info.free_space || 0)
    let used = 0
    
    // 如果有总容量和剩余空间，计算已用空间
    if (info.free_space !== undefined) {
      used = total - free
    } else if (info.used !== undefined) {
      used = Number(info.used)
    }

    // 将 0 或 -1 视为无限容量
    storageInfo.value.isUnlimited = total === 0 || total === -1
    storageInfo.value.total = total
    storageInfo.value.used = used > 0 ? used : 0
    
    // 重新计算百分比
    if (!storageInfo.value.isUnlimited && storageInfo.value.total > 0) {
      storageInfo.value.percentage = Math.ceil((storageInfo.value.used / storageInfo.value.total) * 100)
    } else {
      storageInfo.value.percentage = 0
    }
  } else {
    const capacity = info.capacity || info.storage_limit
    if (capacity !== undefined) {
       const capNum = Number(capacity)
       storageInfo.value.isUnlimited = capNum === 0 || capNum === -1
       storageInfo.value.total = capNum
       storageInfo.value.used = Number(info.used || info.used_storage || 0)
       
       if (!storageInfo.value.isUnlimited && storageInfo.value.total > 0) {
         storageInfo.value.percentage = Math.ceil((storageInfo.value.used / storageInfo.value.total) * 100)
       } else {
         storageInfo.value.percentage = 0
       }
    }
  }
}

const initStorageInfo = () => {
  try {
    const user = proxy?.$cache.local.getJSON('userInfo')
    if (user) {
      updateStorageInfo(user)
    }
  } catch (error) {
    proxy?.$log.error('获取存储信息失败', error)
  }
}

onMounted(() => {
  initStorageInfo()
})
</script>

<style scoped>
.storage-card {
  margin: 16px;
  padding: 20px;
  border-radius: 16px;
  background: var(--bg-color-overlay, #f8fafc);
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
  margin-bottom: 0px; 
  font-family: var(--font-family-mono, monospace);
}

.upgrade-btn {
  width: 100%;
  margin-top: 16px;
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
</style>


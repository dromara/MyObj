<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  currentPath: {
    type: String,
    default: '我的文件'
  },
  currentType: {
    type: String,
    default: 'files'
  },
  searchKeyword: {
    type: String,
    default: ''
  }
})

const viewMode = ref('grid') // grid or list
const selectedFiles = ref([])

const files = ref([
  { id: 1, name: '工作文档', type: 'folder', size: '-', modified: '2024-11-10', icon: 'folder' },
  { id: 2, name: '个人照片', type: 'folder', size: '-', modified: '2024-11-08', icon: 'folder' },
  { id: 3, name: '项目资料.pdf', type: 'pdf', size: '2.5 MB', modified: '2024-11-12', icon: 'pdf' },
  { id: 4, name: '设计稿.png', type: 'image', size: '1.8 MB', modified: '2024-11-11', icon: 'image' },
  { id: 5, name: '会议记录.docx', type: 'doc', size: '500 KB', modified: '2024-11-09', icon: 'doc' },
  { id: 6, name: '数据分析.xlsx', type: 'excel', size: '3.2 MB', modified: '2024-11-07', icon: 'excel' },
  { id: 7, name: '演示文稿.pptx', type: 'ppt', size: '4.5 MB', modified: '2024-11-06', icon: 'ppt' },
  { id: 8, name: '视频教程.mp4', type: 'video', size: '125 MB', modified: '2024-11-05', icon: 'video' }
])

const filteredFiles = computed(() => {
  if (!props.searchKeyword) return files.value
  return files.value.filter(file => 
    file.name.toLowerCase().includes(props.searchKeyword.toLowerCase())
  )
})

const selectFile = (file) => {
  const index = selectedFiles.value.findIndex(f => f.id === file.id)
  if (index > -1) {
    selectedFiles.value.splice(index, 1)
  } else {
    selectedFiles.value.push(file)
  }
}

const isSelected = (file) => {
  return selectedFiles.value.some(f => f.id === file.id)
}

const getFileIcon = (iconType) => {
  const icons = {
    folder: { path: 'M10 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z', color: '#409eff' },
    pdf: { path: 'M14 2H6c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V8l-6-6zm4 18H6V4h7v5h5v11z', color: '#f56c6c' },
    image: { path: 'M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z', color: '#67c23a' },
    doc: { path: 'M14 2H6c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V8l-6-6zM6 20V4h7v5h5v11H6z', color: '#409eff' },
    excel: { path: 'M14 2H6c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V8l-6-6zM6 20V4h7v5h5v11H6z', color: '#67c23a' },
    ppt: { path: 'M14 2H6c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V8l-6-6zM6 20V4h7v5h5v11H6z', color: '#e6a23c' },
    video: { path: 'M18 4l2 4h-3l-2-4h-2l2 4h-3l-2-4H8l2 4H7L5 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V4h-4z', color: '#909399' }
  }
  return icons[iconType] || icons.folder
}

const handleFileClick = (file) => {
  if (file.type === 'folder') {
    alert(`打开文件夹: ${file.name}`)
  } else {
    alert(`预览文件: ${file.name}`)
  }
}

const handleDownload = () => {
  if (selectedFiles.value.length === 0) {
    alert('请选择要下载的文件')
  } else {
    alert(`下载 ${selectedFiles.value.length} 个文件`)
  }
}

const handleDelete = () => {
  if (selectedFiles.value.length === 0) {
    alert('请选择要删除的文件')
  } else {
    if (confirm(`确定要删除 ${selectedFiles.value.length} 个文件吗?`)) {
      alert('已删除')
      selectedFiles.value = []
    }
  }
}

const handleShare = () => {
  if (selectedFiles.value.length === 0) {
    alert('请选择要分享的文件')
  } else {
    alert(`分享 ${selectedFiles.value.length} 个文件`)
  }
}
</script>

<template>
  <div class="file-list-container">
    <div class="toolbar">
      <div class="breadcrumb">
        <span class="breadcrumb-item">{{ currentPath }}</span>
      </div>
      
      <div class="toolbar-actions">
        <div class="view-mode">
          <button 
            :class="{ active: viewMode === 'grid' }"
            @click="viewMode = 'grid'"
            title="网格视图"
          >
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path fill="currentColor" d="M3 3h8v8H3zm10 0h8v8h-8zM3 13h8v8H3zm10 0h8v8h-8z"/>
            </svg>
          </button>
          <button 
            :class="{ active: viewMode === 'list' }"
            @click="viewMode = 'list'"
            title="列表视图"
          >
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path fill="currentColor" d="M3 13h2v-2H3v2zm0 4h2v-2H3v2zm0-8h2V7H3v2zm4 4h14v-2H7v2zm0 4h14v-2H7v2zM7 7v2h14V7H7z"/>
            </svg>
          </button>
        </div>
        
        <div class="action-buttons" v-if="selectedFiles.length > 0">
          <button @click="handleDownload" class="action-btn">
            <svg viewBox="0 0 24 24" width="18" height="18">
              <path fill="currentColor" d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/>
            </svg>
            下载
          </button>
          <button @click="handleShare" class="action-btn">
            <svg viewBox="0 0 24 24" width="18" height="18">
              <path fill="currentColor" d="M18 16.08c-.76 0-1.44.3-1.96.77L8.91 12.7c.05-.23.09-.46.09-.7s-.04-.47-.09-.7l7.05-4.11c.54.5 1.25.81 2.04.81 1.66 0 3-1.34 3-3s-1.34-3-3-3-3 1.34-3 3c0 .24.04.47.09.7L8.04 9.81C7.5 9.31 6.79 9 6 9c-1.66 0-3 1.34-3 3s1.34 3 3 3c.79 0 1.5-.31 2.04-.81l7.12 4.16c-.05.21-.08.43-.08.65 0 1.61 1.31 2.92 2.92 2.92 1.61 0 2.92-1.31 2.92-2.92s-1.31-2.92-2.92-2.92z"/>
            </svg>
            分享
          </button>
          <button @click="handleDelete" class="action-btn danger">
            <svg viewBox="0 0 24 24" width="18" height="18">
              <path fill="currentColor" d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
            </svg>
            删除
          </button>
        </div>
      </div>
    </div>
    
    <div class="file-grid" v-if="viewMode === 'grid'">
      <div 
        v-for="file in filteredFiles" 
        :key="file.id"
        class="file-card"
        :class="{ selected: isSelected(file) }"
        @click="selectFile(file)"
        @dblclick="handleFileClick(file)"
      >
        <div class="file-icon">
          <svg viewBox="0 0 24 24" width="48" height="48">
            <path :fill="getFileIcon(file.icon).color" :d="getFileIcon(file.icon).path" />
          </svg>
        </div>
        <div class="file-name" :title="file.name">{{ file.name }}</div>
        <div class="file-info">{{ file.size }} · {{ file.modified }}</div>
      </div>
    </div>
    
    <div class="file-table" v-else>
      <table>
        <thead>
          <tr>
            <th width="40"></th>
            <th>文件名</th>
            <th width="120">大小</th>
            <th width="150">修改时间</th>
          </tr>
        </thead>
        <tbody>
          <tr 
            v-for="file in filteredFiles" 
            :key="file.id"
            :class="{ selected: isSelected(file) }"
            @click="selectFile(file)"
            @dblclick="handleFileClick(file)"
          >
            <td>
              <input 
                type="checkbox" 
                :checked="isSelected(file)"
                @click.stop="selectFile(file)"
              />
            </td>
            <td class="file-name-cell">
              <svg viewBox="0 0 24 24" width="24" height="24">
                <path :fill="getFileIcon(file.icon).color" :d="getFileIcon(file.icon).path" />
              </svg>
              <span>{{ file.name }}</span>
            </td>
            <td>{{ file.size }}</td>
            <td>{{ file.modified }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    
    <div v-if="filteredFiles.length === 0" class="empty-state">
      <svg viewBox="0 0 24 24" width="64" height="64">
        <path fill="#dcdfe6" d="M20 6h-8l-2-2H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm0 12H4V6h5.17l2 2H20v10zm-8-4h2v2h-2zm0-6h2v4h-2z"/>
      </svg>
      <p>{{ searchKeyword ? '未找到匹配的文件' : '文件夹为空' }}</p>
    </div>
  </div>
</template>

<style scoped>
.file-list-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: white;
  overflow: hidden;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color);
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
}

.breadcrumb-item {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.view-mode {
  display: flex;
  gap: 4px;
  background: var(--bg-color);
  border-radius: 6px;
  padding: 4px;
}

.view-mode button {
  padding: 6px;
  background: transparent;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s;
}

.view-mode button:hover {
  color: var(--primary-color);
}

.view-mode button.active {
  background: white;
  color: var(--primary-color);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s;
}

.action-btn:hover {
  background: var(--bg-color);
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.action-btn.danger:hover {
  border-color: var(--danger-color);
  color: var(--danger-color);
}

.file-grid {
  flex: 1;
  padding: 24px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 16px;
  overflow-y: auto;
  align-content: start;
}

.file-card {
  padding: 16px;
  border: 2px solid transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.file-card:hover {
  background: var(--bg-color);
}

.file-card.selected {
  background: #ecf5ff;
  border-color: var(--primary-color);
}

.file-icon {
  display: flex;
  align-items: center;
  justify-content: center;
}

.file-name {
  font-size: 13px;
  color: var(--text-primary);
  text-align: center;
  word-break: break-all;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.file-info {
  font-size: 12px;
  color: var(--text-secondary);
  text-align: center;
}

.file-table {
  flex: 1;
  overflow-y: auto;
  padding: 0 24px;
}

.file-table table {
  width: 100%;
  border-collapse: collapse;
}

.file-table thead th {
  text-align: left;
  padding: 12px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-color);
  position: sticky;
  top: 0;
  z-index: 1;
}

.file-table tbody tr {
  cursor: pointer;
  transition: background 0.3s;
}

.file-table tbody tr:hover {
  background: var(--bg-color);
}

.file-table tbody tr.selected {
  background: #ecf5ff;
}

.file-table tbody td {
  padding: 12px;
  font-size: 14px;
  color: var(--text-primary);
  border-bottom: 1px solid var(--border-color);
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  color: var(--text-secondary);
}

.empty-state p {
  font-size: 14px;
}
</style>

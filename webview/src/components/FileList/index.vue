<script setup lang="ts">
const { proxy } = getCurrentInstance() as ComponentInternalInstance

interface Props {
  currentPath?: string
  currentType?: string
  searchKeyword?: string
}

const props = withDefaults(defineProps<Props>(), {
  currentPath: '我的文件',
  currentType: 'files',
  searchKeyword: ''
})

const viewMode = ref<'grid' | 'list'>('grid')
const selectedFiles = ref<FileItem[]>([])

interface FileItem {
  id: number
  name: string
  type: string
  size: string
  modified: string
  icon: string
}

const files = ref<FileItem[]>([
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
  return files.value.filter((file: FileItem) => 
    file.name.toLowerCase().includes(props.searchKeyword.toLowerCase())
  )
})

const selectFile = (file: FileItem) => {
  const index = selectedFiles.value.findIndex((f: FileItem) => f.id === file.id)
  if (index > -1) {
    selectedFiles.value.splice(index, 1)
  } else {
    selectedFiles.value.push(file)
  }
}

const isSelected = (file: FileItem) => {
  return selectedFiles.value.some((f: FileItem) => f.id === file.id)
}

const getFileIcon = (iconType: string) => {
  const icons: Record<string, { path: string; color: string }> = {
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

const handleFileClick = (file: FileItem) => {
  if (file.type === 'folder') {
    proxy?.$modal.msg(`打开文件夹: ${file.name}`)
  } else {
    proxy?.$modal.msg(`预览文件: ${file.name}`)
  }
}

const handleDownload = async () => {
  if (selectedFiles.value.length === 0) {
    proxy?.$modal.msgWarning('请选择要下载的文件')
    return
  }
  proxy?.$modal.msg(`下载 ${selectedFiles.value.length} 个文件`)
}

const handleDelete = async () => {
  if (selectedFiles.value.length === 0) {
    proxy?.$modal.msgWarning('请选择要删除的文件')
    return
  }
  try {
    await proxy?.$modal.confirm(`确定要删除 ${selectedFiles.value.length} 个文件吗?`)
    proxy?.$modal.msgSuccess('已删除')
    selectedFiles.value = []
  } catch (error) {
    // 用户取消操作
  }
}

const handleShare = () => {
  if (selectedFiles.value.length === 0) {
    proxy?.$modal.msgWarning('请选择要分享的文件')
    return
  }
  proxy?.$modal.msg(`分享 ${selectedFiles.value.length} 个文件`)
}
</script>

<template>
  <div class="file-list-container">
    <div class="toolbar">
      <div class="breadcrumb">
        <span class="breadcrumb-item">{{ currentPath }}</span>
      </div>
      
      <div class="toolbar-actions">
        <el-button-group>
          <el-button
            :type="viewMode === 'grid' ? 'primary' : 'default'"
            icon="Grid"
            @click="viewMode = 'grid'"
            title="网格视图"
          />
          <el-button
            :type="viewMode === 'list' ? 'primary' : 'default'"
            icon="List"
            @click="viewMode = 'list'"
            title="列表视图"
          />
        </el-button-group>
        
        <div class="action-buttons" v-if="selectedFiles.length > 0">
          <el-button
            icon="Download"
            @click="handleDownload"
          >
            下载
          </el-button>
          <el-button
            icon="Share"
            @click="handleShare"
          >
            分享
          </el-button>
          <el-button
            type="danger"
            icon="Delete"
            @click="handleDelete"
          >
            删除
          </el-button>
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

.action-buttons {
  display: flex;
  gap: 8px;
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
  line-clamp: 2;
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

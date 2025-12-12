<template>
  <div class="square-container">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <div class="breadcrumb">
        <el-icon :size="24" color="#409EFF"><Grid /></el-icon>
        <span class="breadcrumb-item">文件广场</span>
        <span class="breadcrumb-desc">探索用户分享的公开文件</span>
      </div>
      
      <div class="toolbar-actions">
        <!-- 搜索框 -->
        <el-input
          v-model="searchKeyword"
          placeholder="搜索广场文件..."
          :prefix-icon="Search"
          clearable
          @input="handleSearch"
          style="width: 300px"
        />
        
        <!-- 视图切换 -->
        <div class="view-mode">
          <el-button-group>
            <el-button :type="viewMode === 'grid' ? 'primary' : ''" :icon="Grid" @click="viewMode = 'grid'" />
            <el-button :type="viewMode === 'list' ? 'primary' : ''" :icon="List" @click="viewMode = 'list'" />
          </el-button-group>
        </div>
      </div>
    </div>
    
    <!-- 筛选栏 -->
    <div class="filter-bar">
      <el-radio-group v-model="fileTypeFilter" @change="handleFilterChange">
        <el-radio-button label="all">全部文件</el-radio-button>
        <el-radio-button label="image">图片</el-radio-button>
        <el-radio-button label="video">视频</el-radio-button>
        <el-radio-button label="doc">文档</el-radio-button>
        <el-radio-button label="audio">音频</el-radio-button>
        <el-radio-button label="archive">压缩包</el-radio-button>
        <el-radio-button label="other">其他</el-radio-button>
      </el-radio-group>
      
      <el-select v-model="sortBy" placeholder="排序方式" style="width: 150px" @change="handleSortChange">
        <el-option label="最新上传" value="newest" />
        <el-option label="最多下载" value="downloads" />
        <el-option label="文件大小" value="size" />
        <el-option label="文件名称" value="name" />
      </el-select>
    </div>
    
    <!-- 文件网格视图 -->
    <div v-if="viewMode === 'grid'" class="file-grid" v-loading="loading">
      <el-card
        v-for="file in filteredFiles"
        :key="file.id"
        shadow="hover"
        class="file-card"
        @click="handleFileClick(file)"
      >
        <div class="file-icon">
          <el-icon :size="64" :color="getFileColor(file.type)">
            <component :is="getFileIcon(file.type)" />
          </el-icon>
        </div>
        <div class="file-name" :title="file.name">{{ file.name }}</div>
        <div class="file-meta">
          <div class="file-info">
            <span>{{ formatFileSize(file.size) }}</span>
            <span>·</span>
            <span>{{ file.ownerName }}</span>
          </div>
          <div class="file-stats">
            <el-icon><View /></el-icon>
            <span>{{ file.viewCount }}</span>
            <el-icon><Download /></el-icon>
            <span>{{ file.downloadCount }}</span>
          </div>
        </div>
        <div class="file-actions">
          <el-button type="primary" size="small" :icon="Download" @click.stop="handleDownload(file)">
            下载
          </el-button>
        </div>
      </el-card>
      
      <!-- 空状态 -->
      <div v-if="filteredFiles.length === 0 && !loading" class="empty-state">
        <el-empty description="暂无公开文件" />
      </div>
    </div>
    
    <!-- 文件列表视图 -->
    <el-table
      v-else
      :data="filteredFiles"
      v-loading="loading"
      @row-click="handleFileClick"
      style="width: 100%"
    >
      <el-table-column label="文件名" min-width="300">
        <template #default="{ row }">
          <div class="file-name-cell">
            <el-icon :size="24" :color="getFileColor(row.type)">
              <component :is="getFileIcon(row.type)" />
            </el-icon>
            <span>{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="大小" width="120">
        <template #default="{ row }">
          {{ formatFileSize(row.size) }}
        </template>
      </el-table-column>
      <el-table-column label="上传者" width="150" prop="ownerName" />
      <el-table-column label="浏览" width="100">
        <template #default="{ row }">
          <span>{{ row.viewCount }}</span>
        </template>
      </el-table-column>
      <el-table-column label="下载" width="100">
        <template #default="{ row }">
          <span>{{ row.downloadCount }}</span>
        </template>
      </el-table-column>
      <el-table-column label="上传时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" size="small" :icon="Download" @click.stop="handleDownload(row)">
            下载
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 分页 -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="handlePageChange"
        @size-change="handleSizeChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Grid,
  List,
  Search,
  Download,
  View,
  Document,
  Picture,
  VideoCamera,
  Headset,
  Files,
  FolderOpened
} from '@element-plus/icons-vue'

const route = useRoute()

// 响应式数据
const viewMode = ref<'grid' | 'list'>('grid')
const searchKeyword = ref('')
const fileTypeFilter = ref('all')
const sortBy = ref('newest')
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 模拟数据 - 实际应从后端API获取
const publicFiles = ref([
  {
    id: '1',
    name: 'Vue3开发指南.pdf',
    type: 'doc',
    size: 2560000,
    ownerName: '张三',
    viewCount: 128,
    downloadCount: 56,
    createdAt: '2024-11-10 15:30:00'
  },
  {
    id: '2',
    name: '风景照片.jpg',
    type: 'image',
    size: 1800000,
    ownerName: '李四',
    viewCount: 256,
    downloadCount: 89,
    createdAt: '2024-11-11 10:20:00'
  },
  {
    id: '3',
    name: '教学视频.mp4',
    type: 'video',
    size: 125000000,
    ownerName: '王五',
    viewCount: 512,
    downloadCount: 234,
    createdAt: '2024-11-12 09:15:00'
  },
  {
    id: '4',
    name: '项目源码.zip',
    type: 'archive',
    size: 4500000,
    ownerName: '赵六',
    viewCount: 345,
    downloadCount: 167,
    createdAt: '2024-11-09 14:45:00'
  },
  {
    id: '5',
    name: '音乐专辑.mp3',
    type: 'audio',
    size: 3200000,
    ownerName: '孙七',
    viewCount: 678,
    downloadCount: 432,
    createdAt: '2024-11-08 16:00:00'
  }
])

// 筛选后的文件列表
const filteredFiles = computed(() => {
  let files = publicFiles.value

  // 根据文件类型筛选
  if (fileTypeFilter.value !== 'all') {
    files = files.filter(file => file.type === fileTypeFilter.value)
  }

  // 根据搜索关键词筛选
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    files = files.filter(file => 
      file.name.toLowerCase().includes(keyword) ||
      file.ownerName.toLowerCase().includes(keyword)
    )
  }

  // 排序
  files = [...files].sort((a, b) => {
    switch (sortBy.value) {
      case 'newest':
        return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      case 'downloads':
        return b.downloadCount - a.downloadCount
      case 'size':
        return b.size - a.size
      case 'name':
        return a.name.localeCompare(b.name, 'zh-CN')
      default:
        return 0
    }
  })

  total.value = files.length
  return files
})

// 获取文件图标
const getFileIcon = (type: string) => {
  const iconMap: Record<string, any> = {
    doc: Document,
    image: Picture,
    video: VideoCamera,
    audio: Headset,
    archive: Files,
    other: FolderOpened
  }
  return iconMap[type] || Document
}

// 获取文件颜色
const getFileColor = (type: string) => {
  const colorMap: Record<string, string> = {
    doc: '#409EFF',
    image: '#67C23A',
    video: '#E6A23C',
    audio: '#F56C6C',
    archive: '#909399',
    other: '#909399'
  }
  return colorMap[type] || '#909399'
}

// 格式化文件大小
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

// 格式化时间
const formatTime = (time: string): string => {
  return time
}

// 搜索处理
const handleSearch = () => {
  currentPage.value = 1
}

// 筛选处理
const handleFilterChange = () => {
  currentPage.value = 1
}

// 排序处理
const handleSortChange = () => {
  currentPage.value = 1
}

// 点击文件
const handleFileClick = (file: any) => {
  ElMessage.info(`预览文件: ${file.name}`)
}

// 下载文件
const handleDownload = (file: any) => {
  ElMessage.success(`开始下载: ${file.name}`)
  // TODO: 调用下载API
}

// 分页处理
const handlePageChange = (page: number) => {
  currentPage.value = page
  // TODO: 加载对应页的数据
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  // TODO: 重新加载数据
}

// 加载公开文件列表
const loadPublicFiles = async () => {
  loading.value = true
  try {
    // TODO: 调用后端API获取公开文件列表
    // const response = await searchPublicFiles({
    //   keyword: searchKeyword.value,
    //   type: fileTypeFilter.value,
    //   sortBy: sortBy.value,
    //   page: currentPage.value,
    //   pageSize: pageSize.value
    // })
    // publicFiles.value = response.data.files
    // total.value = response.data.total
    
    // 模拟加载延迟
    await new Promise(resolve => setTimeout(resolve, 500))
  } catch (error) {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadPublicFiles()
  // 监听 URL 参数变化
  const keyword = route.query.keyword as string
  if (keyword) {
    searchKeyword.value = keyword
  }
})

// 监听路由参数变化
watch(() => route.query.keyword, (newKeyword) => {
  if (newKeyword) {
    searchKeyword.value = newKeyword as string
    loadPublicFiles()
  }
})
</script>

<style scoped>
.square-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: white;
  border-radius: 8px;
  overflow: hidden;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid var(--el-border-color);
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 12px;
}

.breadcrumb-item {
  font-size: 18px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.breadcrumb-desc {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  margin-left: 8px;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  background: var(--el-fill-color-light);
  border-bottom: 1px solid var(--el-border-color);
}

.file-grid {
  flex: 1;
  padding: 24px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 20px;
  overflow-y: auto;
  align-content: start;
}

.file-card {
  cursor: pointer;
  transition: all 0.3s;
}

.file-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.file-card :deep(.el-card__body) {
  padding: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.file-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 12px;
}

.file-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  text-align: center;
  word-break: break-all;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.file-meta {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.file-info {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.file-stats {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.file-stats .el-icon {
  font-size: 14px;
}

.file-stats span {
  margin-left: 4px;
}

.file-actions {
  width: 100%;
  margin-top: 4px;
}

.file-actions .el-button {
  width: 100%;
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.empty-state {
  grid-column: 1 / -1;
  display: flex;
  justify-content: center;
  padding: 60px 0;
}

.pagination {
  padding: 20px;
  border-top: 1px solid var(--el-border-color);
  display: flex;
  justify-content: center;
}
</style>

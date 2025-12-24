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
        <!-- 视图切换按钮 -->
        <el-button-group>
          <el-button :type="viewMode === 'grid' ? 'primary' : ''" icon="Grid" @click="viewMode = 'grid'" />
          <el-button :type="viewMode === 'list' ? 'primary' : ''" icon="List" @click="viewMode = 'list'" />
        </el-button-group>
      </div>
    </div>
    
    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="filter-type-group">
        <el-radio-group v-model="fileTypeFilter" @change="handleFilterChange" class="type-radio-group">
          <el-radio-button label="all">全部</el-radio-button>
          <el-radio-button label="image">图片</el-radio-button>
          <el-radio-button label="video">视频</el-radio-button>
          <el-radio-button label="doc">文档</el-radio-button>
          <el-radio-button label="audio">音频</el-radio-button>
          <el-radio-button label="archive">压缩包</el-radio-button>
          <el-radio-button label="other">其他</el-radio-button>
        </el-radio-group>
      </div>
      
      <div class="filter-sort-group">
        <el-select v-model="sortBy" placeholder="排序方式" class="sort-select" @change="handleSortChange">
          <el-option label="最新上传" value="time" />
          <el-option label="文件大小" value="size" />
          <el-option label="文件名称" value="name" />
        </el-select>
      </div>
    </div>
    
    <!-- 文件网格视图 -->
    <div v-if="viewMode === 'grid'" class="file-grid" v-loading="loading">
      <el-card
        v-for="file in filteredFiles"
        :key="file.uf_id"
        shadow="hover"
        class="file-card"
        @click="handleFileClick(file)"
        @dblclick="handleFileClick(file)"
      >
        <div class="file-icon">
          <el-icon :size="64" :color="getFileIconColor(file.mime_type)">
            <component :is="getFileIconName(file.mime_type)" />
          </el-icon>
        </div>
        <file-name-tooltip :file-name="file.file_name" view-mode="grid" tag="div" custom-class="file-name" />
        <div class="file-meta">
          <div class="file-info">
            <span>{{ formatFileSize(file.file_size) }}</span>
            <span>·</span>
            <span>{{ file.owner_name }}</span>
          </div>
          <div class="file-stats">
            <span>{{ formatTime(file.created_at) }}</span>
          </div>
        </div>
        <div class="file-actions">
          <el-button type="primary" size="small" icon="Download" @click.stop="handleDownload(file)">
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
    <!-- PC端：表格布局 -->
    <el-table
      v-else-if="!isMobile"
      :data="filteredFiles"
      v-loading="loading"
      @row-click="handleFileClick"
      style="width: 100%"
    >
      <el-table-column label="文件名" min-width="300">
        <template #default="{ row }">
          <div class="file-name-cell">
            <el-icon :size="24" :color="getFileIconColor(row.mime_type)">
              <component :is="getFileIconName(row.mime_type)" />
            </el-icon>
            <file-name-tooltip :file-name="row.file_name" view-mode="table" />
          </div>
        </template>
      </el-table-column>
      <el-table-column label="大小" width="120">
        <template #default="{ row }">
          {{ formatFileSize(row.file_size) }}
        </template>
      </el-table-column>
      <el-table-column label="上传者" width="150" prop="owner_name" />
      <el-table-column label="上传时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" size="small" icon="Download" @click.stop="handleDownload(row)">
            下载
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 移动端：卡片列表布局 -->
    <div v-else class="mobile-file-list" v-loading="loading">
      <div
        v-for="file in filteredFiles"
        :key="file.uf_id"
        class="mobile-file-item"
        @click="handleFileClick(file)"
      >
        <div class="mobile-item-content">
          <div class="mobile-item-icon">
            <el-icon :size="40" :color="getFileIconColor(file.mime_type)">
              <component :is="getFileIconName(file.mime_type)" />
            </el-icon>
          </div>
          <div class="mobile-item-info">
            <div class="mobile-item-name-row">
              <file-name-tooltip :file-name="file.file_name" view-mode="list" custom-class="mobile-item-name" />
            </div>
            <div class="mobile-item-meta">
              <span class="mobile-item-size">{{ formatFileSize(file.file_size) }}</span>
              <span class="mobile-item-owner">{{ file.owner_name }}</span>
              <span class="mobile-item-time">{{ formatTime(file.created_at) }}</span>
            </div>
          </div>
          <div class="mobile-item-actions" @click.stop>
            <el-button
              type="primary"
              size="small"
              icon="Download"
              class="mobile-download-btn"
              @click.stop="handleDownload(file)"
            >
              下载
            </el-button>
          </div>
        </div>
      </div>
      
      <!-- 空状态 -->
      <div v-if="filteredFiles.length === 0 && !loading" class="mobile-empty-state">
        <el-empty description="暂无公开文件" />
      </div>
    </div>
    
    <!-- 分页 -->
    <div class="pagination">
      <pagination
        v-model:page="currentPage"
        v-model:limit="pageSize"
        :total="total"
        :page-sizes="[20, 50, 100]"
        @pagination="handlePagination"
      />
    </div>

    <!-- 文件预览组件 -->
    <preview v-model="previewVisible" :file="previewFile" />
  </div>
</template>

<script setup lang="ts">
import { formatSize, formatDate } from '@/utils'
import { useResponsive } from '@/composables/useResponsive'
import { getPublicFileList, searchPublicFiles, type PublicFileItem, type PublicFileListParams } from '@/api/file'
import { useSearch } from '@/composables/useSearch'
import { getFileIcon } from '@/utils/fileIcon'

const { proxy } = getCurrentInstance() as ComponentInternalInstance
const route = useRoute()

// 使用响应式检测 composable
const { isMobile } = useResponsive()

// 响应式数据
const viewMode = ref<'grid' | 'list'>(isMobile.value ? 'list' : 'grid')
const fileTypeFilter = ref('all')
const sortBy = ref('time')
const loading = ref(false)
const isSearchMode = ref(false) // 是否处于搜索模式

// 公开文件列表
const publicFiles = ref<PublicFileItem[]>([])

// 使用通用搜索 composable
const transformResult = (files: any[]): PublicFileItem[] => {
  return files.map((file: any) => ({
    uf_id: file.uf_id || file.id || '',
    file_name: file.file_name || file.name || '',
    file_size: file.size || 0,
    mime_type: file.mime || file.mime_type || '',
    created_at: file.created_at || file.createdAt || '',
    owner_name: file.owner_name || 'Unknown',
    has_thumbnail: (file.thumbnail_img && file.thumbnail_img !== '') || false
  }))
}

const search = useSearch<PublicFileItem>(
  searchPublicFiles,
  transformResult,
  () => {
    // 清空搜索时的回调：切换到正常模式
    isSearchMode.value = false
    currentPage.value = 1
    loadPublicFiles()
  }
)

// 兼容现有代码的变量
const searchKeyword = search.searchKeyword
const currentPage = search.currentPage
const pageSize = search.pageSize
const total = search.total

// 筛选后的文件列表（搜索模式或正常模式）
const filteredFiles = computed(() => {
  if (isSearchMode.value) {
    return search.searchResults.value
  }
  return publicFiles.value || []
})

// 获取文件图标名称
const getFileIconName = (mimeType: string) => {
  return getFileIcon(mimeType).icon
}

// 获取文件图标颜色
const getFileIconColor = (mimeType: string) => {
  return getFileIcon(mimeType).color
}

const formatFileSize = formatSize

const formatTime = (time: string): string => {
  return formatDate(time, { showTime: true })
}


// 搜索处理（使用后端搜索 API）
const performSearch = async (keyword: string, pageNum: number = 1, pageSizeNum: number = 20) => {
  if (!keyword.trim()) {
    // 如果关键词为空，切换到正常模式
    isSearchMode.value = false
    currentPage.value = 1
    loadPublicFiles()
    return
  }

  isSearchMode.value = true
  await search.performSearch(keyword, pageNum, pageSizeNum)
}

// 筛选处理
const handleFilterChange = () => {
  currentPage.value = 1
  loadPublicFiles()
}

// 排序处理
const handleSortChange = () => {
  currentPage.value = 1
  loadPublicFiles()
}

// 文件预览
const previewVisible = ref(false)
const previewFile = ref<any>(null)

// 点击文件
const handleFileClick = (file: PublicFileItem) => {
  // 将 Square 的文件格式转换为 Preview 组件需要的格式
  previewFile.value = {
    file_id: file.uf_id,
    file_name: file.file_name,
    file_size: file.file_size,
    mime_type: file.mime_type,
    is_enc: false,
    has_thumbnail: file.has_thumbnail,
    created_at: file.created_at
  }
  previewVisible.value = true
}


// 下载文件
const handleDownload = (file: PublicFileItem) => {
  proxy?.$modal.msgSuccess(`开始下载: ${file.file_name}`)
  // TODO: 调用下载API
}

// 分页处理
const handlePagination = ({ page, limit }: { page: number; limit: number }) => {
  if (isSearchMode.value && searchKeyword.value.trim()) {
    // 搜索模式下的分页
    performSearch(searchKeyword.value, page, limit)
  } else {
    // 正常模式下的分页
    currentPage.value = page
    pageSize.value = limit
    loadPublicFiles()
  }
}

// 加载公开文件列表
const loadPublicFiles = async () => {
  // 如果正在搜索，不显示加载状态（避免冲突）
  if (!isSearchMode.value) {
    loading.value = true
  }
  try {
    // 构建请求参数，只有当类型不是 'all' 时才传递 type
    const params: PublicFileListParams = {
      sortBy: sortBy.value,
      page: currentPage.value,
      pageSize: pageSize.value
    }
    // 只有当类型不是 'all' 时才添加 type 参数
    if (fileTypeFilter.value !== 'all') {
      params.type = fileTypeFilter.value
    }
    
    const response = await getPublicFileList(params)
    
    if (response.code === 200 && response.data) {
      // 确保 files 是数组，如果为 null 或 undefined 则使用空数组
      publicFiles.value = response.data.files || []
      total.value = response.data.total || 0
    } else {
      proxy?.$modal.msgError(response.message || '加载失败')
      // 加载失败时也确保是空数组
      publicFiles.value = []
    }
  } catch (error) {
    proxy?.$log.error('加载公开文件列表失败:', error)
    proxy?.$modal.msgError('加载失败')
  } finally {
    if (!isSearchMode.value) {
      loading.value = false
    }
  }
}

onMounted(() => {
  // 监听全局搜索事件
  const handleGlobalSearch = (event: Event) => {
    const customEvent = event as CustomEvent<{ keyword: string }>
    const keyword = customEvent.detail.keyword.trim()
    
    if (keyword) {
      // 只有当关键词变化时才执行搜索，避免重复请求
      if (searchKeyword.value !== keyword) {
        searchKeyword.value = keyword
        performSearch(keyword, 1, pageSize.value)
      }
    } else {
      // 清空搜索
      search.clearSearch()
    }
  }

  window.addEventListener('square-search', handleGlobalSearch)

  // 检查路由参数中是否有搜索关键词
  const keyword = route.query.keyword as string
  if (keyword) {
    searchKeyword.value = keyword
    performSearch(keyword, 1, pageSize.value)
  } else {
    loadPublicFiles()
  }

  // 清理事件监听
  onBeforeUnmount(() => {
    window.removeEventListener('square-search', handleGlobalSearch)
  })
})

// 监听路由参数变化
watch(() => route.query.keyword, (newKeyword) => {
  if (newKeyword) {
    searchKeyword.value = newKeyword as string
    performSearch(newKeyword as string, 1, pageSize.value)
  } else if (isSearchMode.value) {
    // 如果路由参数被清空，且当前在搜索模式，则切换到正常模式
    isSearchMode.value = false
    searchKeyword.value = ''
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
  gap: 16px;
}

.filter-type-group {
  flex: 1;
  min-width: 0;
}

.type-radio-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.filter-sort-group {
  flex-shrink: 0;
}

.sort-select {
  width: 150px;
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

.file-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.file-card :deep(.el-card__body) {
  padding: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-height: 0;
}

.file-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 12px;
  flex-shrink: 0;
}

.file-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  text-align: center;
  width: 100%;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  word-break: break-word;
  line-height: 1.4;
  min-height: 2.8em; /* 固定高度：2行 * 1.4行高 */
  max-height: 2.8em;
}

.file-meta {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex-shrink: 0;
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
  margin-top: auto;
  flex-shrink: 0;
  padding-top: 8px;
}

.file-actions .el-button {
  width: 100%;
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.file-name-text {
  display: inline-block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 250px;
  vertical-align: middle;
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


/* 移动端响应式 - 组件特定样式 */
@media (max-width: 1024px) {
  .square-container {
    border-radius: 0;
  }
  
  .toolbar {
    padding: 12px 16px;
    gap: 12px;
  }
  
  .breadcrumb {
    flex: 1 1 100%;
    order: 1;
  }
  
  .breadcrumb-item {
    font-size: 16px;
  }
  
  .breadcrumb-desc {
    display: none;
  }
  
  .toolbar-actions {
    flex: 1 1 100%;
    order: 2;
    width: 100%;
  }
  
  .filter-bar {
    padding: 12px 16px;
    flex-direction: column;
    gap: 12px;
  }

  .filter-type-group {
    width: 100%;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .type-radio-group {
    display: flex;
    gap: 6px;
    min-width: max-content;
  }
  
  .filter-bar :deep(.el-radio-button) {
    flex-shrink: 0;
  }
  
  .filter-bar :deep(.el-radio-button__inner) {
    padding: 8px 12px;
    font-size: 12px;
    white-space: nowrap;
  }
  
  .filter-sort-group {
    width: 100%;
  }

  .sort-select {
    width: 100%;
  }
  
  .file-grid {
    padding: 12px;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 12px;
  }
  
  .file-card :deep(.el-card__body) {
    padding: 12px;
    gap: 8px;
  }
  
  .file-icon {
    padding: 8px;
  }
  
  .file-name {
    font-size: 12px;
    min-height: 2.52em; /* 2行 * 1.26行高（12px * 1.05） */
    max-height: 2.52em;
  }
  
  .file-info {
    font-size: 11px;
  }
  
  .file-stats {
    font-size: 11px;
  }
  
  .pagination {
    padding: 12px;
  }
}

@media (max-width: 480px) {
  .file-grid {
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
    gap: 8px;
    padding: 8px;
  }
  
  .file-card :deep(.el-card__body) {
    padding: 8px;
    gap: 6px;
  }
  
  .file-icon {
    padding: 6px;
  }
  
  .file-name {
    font-size: 11px;
    min-height: 2.31em; /* 2行 * 1.155行高（11px * 1.05） */
    max-height: 2.31em;
  }
  
  .filter-bar :deep(.el-radio-button__inner) {
    padding: 6px 10px;
    font-size: 11px;
  }

  .toolbar {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .breadcrumb {
    width: 100%;
  }

  .toolbar-actions {
    width: 100%;
  }
}

/* 移动端卡片列表布局 */
.mobile-file-list {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  overflow-y: auto;
}

.mobile-file-item {
  background: white;
  border-radius: 12px;
  padding: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  transition: all 0.2s ease;
  border: 1px solid var(--el-border-color-lighter);
  cursor: pointer;
}

.mobile-file-item:active {
  transform: scale(0.98);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  background: var(--el-fill-color-light);
}

.mobile-item-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.mobile-item-icon {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.mobile-item-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.mobile-item-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.mobile-item-name {
  flex: 1;
  min-width: 0;
  font-size: 15px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mobile-item-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  flex-wrap: wrap;
}

.mobile-item-size {
  font-weight: 500;
  color: var(--el-text-color-regular);
}

.mobile-item-owner {
  color: var(--el-text-color-secondary);
}

.mobile-item-time {
  color: var(--el-text-color-placeholder);
}

.mobile-item-actions {
  flex-shrink: 0;
}

.mobile-download-btn {
  padding: 8px 16px;
}

.mobile-empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 60px 0;
  min-height: 200px;
}

@media (max-width: 480px) {
  .mobile-file-list {
    padding: 8px;
    gap: 6px;
  }

  .mobile-file-item {
    padding: 10px;
    border-radius: 10px;
  }

  .mobile-item-icon {
    width: 36px;
    height: 36px;
  }

  .mobile-item-name {
    font-size: 14px;
  }

  .mobile-item-meta {
    font-size: 11px;
    gap: 8px;
  }

  .mobile-download-btn {
    padding: 6px 12px;
    font-size: 12px;
  }
}
</style>

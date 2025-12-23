<template>
  <div class="trash-page">
    <!-- 工具栏 -->
    <el-card shadow="never" class="toolbar-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <!-- 移动端：使用下拉菜单 -->
          <el-dropdown
            class="mobile-toolbar-menu"
            trigger="click"
            @command="handleToolbarCommand"
          >
            <el-button type="primary" icon="More" circle />
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="restore" :disabled="selectedIds.length === 0" icon="RefreshRight">还原</el-dropdown-item>
                <el-dropdown-item command="delete" :disabled="selectedIds.length === 0" icon="Delete">永久删除</el-dropdown-item>
                <el-dropdown-item divided command="empty" icon="Delete">清空回收站</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          
          <!-- 桌面端：显示所有按钮 -->
          <div class="desktop-toolbar">
            <el-button 
              icon="RefreshRight" 
              :disabled="selectedIds.length === 0"
              @click="handleRestore"
            >
              还原
            </el-button>
            <el-button 
              icon="Delete" 
              type="danger"
              :disabled="selectedIds.length === 0"
              @click="handleDeletePermanently"
            >
              永久删除
            </el-button>
            <el-divider direction="vertical" />
            <el-button 
              icon="Delete" 
              type="danger"
              @click="handleEmptyTrash"
            >
              清空回收站
            </el-button>
          </div>
        </div>
        
        <div class="toolbar-right" v-if="selectedIds.length > 0">
          <el-tag type="info">已选择 {{ selectedIds.length }} 项</el-tag>
        </div>
      </div>
    </el-card>
    
    <!-- 文件列表 -->
    <el-table
      v-loading="loading"
      :data="fileList"
      @selection-change="handleSelectionChange"
      class="trash-table"
    >
      <el-table-column type="selection" width="55" class-name="mobile-hide" />
      <el-table-column label="名称" min-width="300" class-name="mobile-name-column">
        <template #default="{ row }">
          <div class="file-name-cell" @dblclick="handleFilePreview(row)">
            <div class="list-file-icon">
              <FileIcon
                :mime-type="row.mime_type"
                :file-name="row.file_name"
                :thumbnail-url="getThumbnailUrl(row.file_id)"
                :show-thumbnail="row.has_thumbnail"
                :icon-size="24"
                :show-badge="false"
                :is-encrypted="row.is_enc"
              />
            </div>
            <span>{{ row.file_name }}</span>
            <el-tag v-if="row.is_enc" size="small" type="warning" class="enc-tag-inline">
              <el-icon :size="12"><Lock /></el-icon>
              加密
            </el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="大小" width="120" class-name="mobile-hide">
        <template #default="{ row }">
          {{ formatSize(row.file_size) }}
        </template>
      </el-table-column>
      <el-table-column label="删除时间" width="180" class-name="mobile-hide">
        <template #default="{ row }">
          {{ formatDate(row.deleted_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right" class-name="mobile-actions-column">
        <template #default="{ row }">
          <el-button link icon="RefreshRight" @click.stop="handleRestoreFile(row)">还原</el-button>
          <el-button link icon="Delete" type="danger" @click.stop="handleDeleteFilepermanently(row)">永久删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 空状态 -->
    <el-empty v-if="!loading && fileList.length === 0" description="回收站为空" />
    
    <!-- 分页 -->
    <el-pagination
      v-if="total > 0"
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :page-sizes="[20, 50, 100]"
      :total="total"
      layout="total, sizes, prev, pager, next, jumper"
      @size-change="handleSizeChange"
      @current-change="handlePageChange"
      class="pagination"
    />

    <!-- 文件预览组件 -->
    <Preview v-model="previewVisible" :file="previewFile" />
  </div>
</template>

<script setup lang="ts">
import { 
  getRecycledList, 
  restoreFile, 
  deleteFilePermanently, 
  emptyRecycled,
  type RecycledItem 
} from '@/api/recycled'
import { getThumbnailUrl } from '@/api/file'
import { formatSize, formatDate } from '@/utils'
import Preview from '@/components/Preview/index.vue'
import type { FileItem } from '@/types'

const { proxy } = getCurrentInstance() as ComponentInternalInstance

// 数据
const loading = ref(false)
const fileList = ref<RecycledItem[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const selectedIds = ref<string[]>([])

// 加载回收站列表
const loadRecycledList = async () => {
  loading.value = true
  try {
    const res = await getRecycledList({
      page: currentPage.value,
      pageSize: pageSize.value
    })
    
    if (res.code === 200 && res.data) {
      fileList.value = res.data.items || []
      total.value = res.data.total || 0
    } else {
      proxy?.$modal.msgError(res.message || '获取回收站列表失败')
    }
  } catch (error: any) {
    proxy?.$modal.msgError(error.message || '获取回收站列表失败')
  } finally {
    loading.value = false
  }
}

// 选择变化
const handleSelectionChange = (selection: RecycledItem[]) => {
  selectedIds.value = selection.map(item => item.recycled_id)
}

// 文件预览
const previewVisible = ref(false)
const previewFile = ref<FileItem | null>(null)

const handleFilePreview = (item: RecycledItem) => {
  // 将 RecycledItem 转换为 FileItem 格式
  previewFile.value = {
    file_id: item.file_id,
    file_name: item.file_name,
    file_size: item.file_size,
    mime_type: item.mime_type,
    is_enc: item.is_enc,
    has_thumbnail: item.has_thumbnail,
    created_at: item.deleted_at
  }
  previewVisible.value = true
}

// 还原文件（批量）
const handleRestore = async () => {
  if (selectedIds.value.length === 0) {
    proxy?.$modal.msgWarning('请先选择要还原的文件')
    return
  }
  
  try {
    await proxy?.$modal.confirm(`确定要还原 ${selectedIds.value.length} 个文件吗？`)
    let successCount = 0
    let failedCount = 0
    
    for (const recycledId of selectedIds.value) {
      try {
        const res = await restoreFile(recycledId)
        if (res.code === 200) {
          successCount++
        } else {
          failedCount++
        }
      } catch {
        failedCount++
      }
    }
    
    if (successCount > 0) {
      proxy?.$modal.msgSuccess(`成功还原 ${successCount} 个文件`)
    }
    if (failedCount > 0) {
      proxy?.$modal.msgWarning(`${failedCount} 个文件还原失败`)
    }
    
    selectedIds.value = []
    await loadRecycledList()
  } catch (error: any) {
    if (error !== 'cancel') {
      // 用户取消操作
    }
  }
}

// 移动端工具栏菜单命令处理
const handleToolbarCommand = (command: string) => {
  switch (command) {
    case 'restore':
      handleRestore()
      break
    case 'delete':
      handleDeletePermanently()
      break
    case 'empty':
      handleEmptyTrash()
      break
  }
}

// 还原单个文件
const handleRestoreFile = async (item: RecycledItem) => {
  try {
    await proxy?.$modal.confirm(`确定要还原 "${item.file_name}" 吗？`)
    const res = await restoreFile(item.recycled_id)
    if (res.code === 200) {
      proxy?.$modal.msgSuccess('还原成功')
      await loadRecycledList()
    } else {
      proxy?.$modal.msgError(res.message || '还原失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      proxy?.$modal.msgError(error.message || '还原失败')
    }
  }
}

// 永久删除（批量）
const handleDeletePermanently = async () => {
  if (selectedIds.value.length === 0) {
    proxy?.$modal.msgWarning('请先选择要删除的文件')
    return
  }
  
  try {
    await proxy?.$modal.confirm(`确定要永久删除 ${selectedIds.value.length} 个文件吗？此操作不可恢复！`)
    let successCount = 0
    let failedCount = 0
    
    for (const recycledId of selectedIds.value) {
      try {
        const res = await deleteFilePermanently(recycledId)
        if (res.code === 200) {
          successCount++
        } else {
          failedCount++
        }
      } catch {
        failedCount++
      }
    }
    
    if (successCount > 0) {
      proxy?.$modal.msgSuccess(`成功删除 ${successCount} 个文件`)
    }
    if (failedCount > 0) {
      proxy?.$modal.msgWarning(`${failedCount} 个文件删除失败`)
    }
    
    selectedIds.value = []
    await loadRecycledList()
  } catch (error: any) {
    if (error !== 'cancel') {
      // 用户取消操作
    }
  }
}

// 永久删除单个文件
const handleDeleteFilepermanently = async (item: RecycledItem) => {
  try {
    await proxy?.$modal.confirm(`确定要永久删除 "${item.file_name}" 吗？此操作不可恢复！`)
    const res = await deleteFilePermanently(item.recycled_id)
    if (res.code === 200) {
      proxy?.$modal.msgSuccess('删除成功')
      await loadRecycledList()
    } else {
      proxy?.$modal.msgError(res.message || '删除失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      proxy?.$modal.msgError(error.message || '删除失败')
    }
  }
}

// 清空回收站
const handleEmptyTrash = async () => {
  if (total.value === 0) {
    proxy?.$modal.msg('回收站已经是空的')
    return
  }
  
  try {
    await proxy?.$modal.confirm(`确定要清空回收站吗？将永久删除所有 ${total.value} 个文件，此操作不可恢复！`)
    loading.value = true
    try {
      const res = await emptyRecycled()
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(res.message || '清空成功')
        await loadRecycledList()
      } else {
        proxy?.$modal.msgError(res.message || '清空失败')
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || '清空失败')
    } finally {
      loading.value = false
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      // 用户取消操作
    }
  }
}

// 分页
const handlePageChange = (page: number) => {
  currentPage.value = page
  loadRecycledList()
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  loadRecycledList()
}

// 初始化
onMounted(() => {
  loadRecycledList()
})
</script>

<style scoped>
.trash-page {
  padding: 20px;
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.toolbar-card {
  margin-bottom: 20px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toolbar-left {
  display: flex;
  gap: 10px;
  align-items: center;
}

.toolbar-right {
  display: flex;
  gap: 10px;
  align-items: center;
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.list-file-icon {
  flex-shrink: 0;
}

.file-name-cell span {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.enc-tag-inline {
  border: none;
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
  font-size: 11px;
  padding: 2px 8px;
  height: 20px;
  margin-left: 8px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

/* 移动端工具栏 */
.mobile-toolbar-menu {
  display: none;
}

.desktop-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 表格移动端优化 */
.trash-table :deep(.mobile-hide) {
  display: table-cell;
}

.trash-table :deep(.mobile-name-column) {
  min-width: 200px;
}

.trash-table :deep(.mobile-actions-column) {
  width: auto;
  min-width: 120px;
}

/* 移动端响应式 */
@media (max-width: 768px) {
  .trash-page {
    padding: 12px;
  }
  
  .toolbar-card {
    margin-bottom: 12px;
    padding: 12px;
  }
  
  .toolbar {
    flex-wrap: wrap;
    gap: 8px;
  }
  
  .toolbar-left {
    flex: 1;
    min-width: 0;
  }
  
  .mobile-toolbar-menu {
    display: inline-flex;
  }
  
  .desktop-toolbar {
    display: none;
  }
  
  .toolbar-right {
    flex: 1 1 100%;
    justify-content: flex-end;
    margin-top: 8px;
  }
  
  .file-name-cell {
    gap: 8px;
  }
  
  .file-name-cell span {
    font-size: 13px;
    max-width: 200px;
  }
  
  /* 表格移动端隐藏列 */
  .trash-table :deep(.mobile-hide) {
    display: none;
  }
  
  .trash-table :deep(.mobile-name-column) {
    min-width: auto;
    width: 100%;
  }
  
  .trash-table :deep(.mobile-actions-column) {
    width: auto;
    min-width: 80px;
  }
  
  /* 操作按钮在移动端使用图标按钮 */
  .trash-table :deep(.mobile-actions-column .el-button) {
    padding: 4px 8px;
    font-size: 12px;
  }
  
  .trash-table :deep(.mobile-actions-column .el-button span) {
    display: none;
  }
  
  .trash-table :deep(.mobile-actions-column .el-button .el-icon) {
    margin: 0;
  }
}

@media (max-width: 480px) {
  .file-name-cell span {
    max-width: 150px;
  }
  
  .trash-table :deep(.mobile-actions-column) {
    width: auto;
    min-width: 60px;
  }
  
  .trash-table :deep(.mobile-actions-column .el-button) {
    padding: 4px;
  }
}
</style>

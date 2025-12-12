<template>
  <div class="trash-page">
    <!-- 工具栏 -->
    <el-card shadow="never" class="toolbar-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button 
            :icon="RefreshRight" 
            :disabled="selectedIds.length === 0"
            @click="handleRestore"
          >
            还原
          </el-button>
          <el-button 
            :icon="Delete" 
            type="danger"
            :disabled="selectedIds.length === 0"
            @click="handleDeletePermanently"
          >
            永久删除
          </el-button>
          <el-divider direction="vertical" />
          <el-button 
            :icon="Delete" 
            type="danger"
            @click="handleEmptyTrash"
          >
            清空回收站
          </el-button>
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
    >
      <el-table-column type="selection" width="55" />
      <el-table-column label="名称" min-width="300">
        <template #default="{ row }">
          <div class="file-name-cell">
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
      <el-table-column label="大小" width="120">
        <template #default="{ row }">
          {{ formatSize(row.file_size) }}
        </template>
      </el-table-column>
      <el-table-column label="删除时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.deleted_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link :icon="RefreshRight" @click.stop="handleRestoreFile(row)">还原</el-button>
          <el-button link :icon="Delete" type="danger" @click.stop="handleDeleteFilepermanently(row)">永久删除</el-button>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, RefreshRight, Lock } from '@element-plus/icons-vue'
import FileIcon from '@/components/FileIcon.vue'
import { 
  getRecycledList, 
  restoreFile, 
  deleteFilePermanently, 
  emptyRecycled,
  type RecycledItem 
} from '@/api/recycled'
import { getThumbnailUrl } from '@/api/file'

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
      ElMessage.error(res.message || '获取回收站列表失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '获取回收站列表失败')
  } finally {
    loading.value = false
  }
}

// 格式化文件大小
const formatSize = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

// 格式化日期
const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 选择变化
const handleSelectionChange = (selection: RecycledItem[]) => {
  selectedIds.value = selection.map(item => item.recycled_id)
}

// 还原文件（批量）
const handleRestore = async () => {
  if (selectedIds.value.length === 0) {
    ElMessage.warning('请先选择要还原的文件')
    return
  }
  
  ElMessageBox.confirm(
    `确定要还原 ${selectedIds.value.length} 个文件吗？`,
    '提示',
    {
      type: 'info',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    }
  ).then(async () => {
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
      ElMessage.success(`成功还原 ${successCount} 个文件`)
    }
    if (failedCount > 0) {
      ElMessage.warning(`${failedCount} 个文件还原失败`)
    }
    
    selectedIds.value = []
    await loadRecycledList()
  }).catch(() => {
    // 用户取消
  })
}

// 还原单个文件
const handleRestoreFile = async (item: RecycledItem) => {
  ElMessageBox.confirm(
    `确定要还原 "${item.file_name}" 吗？`,
    '提示',
    {
      type: 'info',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    }
  ).then(async () => {
    try {
      const res = await restoreFile(item.recycled_id)
      if (res.code === 200) {
        ElMessage.success('还原成功')
        await loadRecycledList()
      } else {
        ElMessage.error(res.message || '还原失败')
      }
    } catch (error: any) {
      ElMessage.error(error.message || '还原失败')
    }
  }).catch(() => {
    // 用户取消
  })
}

// 永久删除（批量）
const handleDeletePermanently = async () => {
  if (selectedIds.value.length === 0) {
    ElMessage.warning('请先选择要删除的文件')
    return
  }
  
  ElMessageBox.confirm(
    `确定要永久删除 ${selectedIds.value.length} 个文件吗？此操作不可恢复！`,
    '警告',
    {
      type: 'error',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
    }
  ).then(async () => {
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
      ElMessage.success(`成功删除 ${successCount} 个文件`)
    }
    if (failedCount > 0) {
      ElMessage.warning(`${failedCount} 个文件删除失败`)
    }
    
    selectedIds.value = []
    await loadRecycledList()
  }).catch(() => {
    // 用户取消
  })
}

// 永久删除单个文件
const handleDeleteFilepermanently = async (item: RecycledItem) => {
  ElMessageBox.confirm(
    `确定要永久删除 "${item.file_name}" 吗？此操作不可恢复！`,
    '警告',
    {
      type: 'error',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
    }
  ).then(async () => {
    try {
      const res = await deleteFilePermanently(item.recycled_id)
      if (res.code === 200) {
        ElMessage.success('删除成功')
        await loadRecycledList()
      } else {
        ElMessage.error(res.message || '删除失败')
      }
    } catch (error: any) {
      ElMessage.error(error.message || '删除失败')
    }
  }).catch(() => {
    // 用户取消
  })
}

// 清空回收站
const handleEmptyTrash = async () => {
  if (total.value === 0) {
    ElMessage.info('回收站已经是空的')
    return
  }
  
  ElMessageBox.confirm(
    `确定要清空回收站吗？将永久删除所有 ${total.value} 个文件，此操作不可恢复！`,
    '警告',
    {
      type: 'error',
      confirmButtonText: '确定清空',
      cancelButtonText: '取消',
    }
  ).then(async () => {
    loading.value = true
    try {
      const res = await emptyRecycled()
      if (res.code === 200) {
        ElMessage.success(res.message || '清空成功')
        await loadRecycledList()
      } else {
        ElMessage.error(res.message || '清空失败')
      }
    } catch (error: any) {
      ElMessage.error(error.message || '清空失败')
    } finally {
      loading.value = false
    }
  }).catch(() => {
    // 用户取消
  })
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
</style>

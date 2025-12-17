<template>
  <div class="shares-page">
    <div class="header-card glass-panel">
      <div class="header">
        <div class="title-section">
          <el-icon :size="24" class="title-icon"><Share /></el-icon>
          <h2>我的分享</h2>
        </div>
        <el-button type="primary" :icon="Refresh" @click="loadShareList">刷新</el-button>
      </div>
    </div>

    <div class="table-card glass-panel">
      <el-table :data="shareList" v-loading="loading" style="width: 100%" empty-text="暂无分享记录">
        <el-table-column label="文件名" min-width="200">
          <template #default="{ row }">
            <div class="file-name-cell">
              <el-icon :size="18" color="#409EFF"><Document /></el-icon>
              <span class="file-name">{{ row.file_name }}</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="分享链接" min-width="350">
          <template #default="{ row }">
            <div class="link-cell">
              <el-input 
                :model-value="getShareUrl(row.token)" 
                readonly
                size="small"
              >
                <template #append>
                  <el-button :icon="CopyDocument" @click="copyShareLink(row)">复制</el-button>
                </template>
              </el-input>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="访问密码" width="120" align="center">
          <template #default>
            <el-tag type="warning" effect="plain" size="small">
              <el-icon><Lock /></el-icon>
              已设置
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="下载次数" prop="download_count" width="100" align="center">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.download_count }}</el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="过期时间" width="180" align="center">
          <template #default="{ row }">
            <div class="time-cell">
              <el-icon :size="14"><Clock /></el-icon>
              <span>{{ formatDate(row.expires_at) }}</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="创建时间" width="180" align="center">
          <template #default="{ row }">
            <div class="time-cell">
              <el-icon :size="14"><Calendar /></el-icon>
              <span>{{ formatDate(row.created_at) }}</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-button-group>
              <el-button link type="primary" :icon="Edit" @click="handleUpdatePassword(row)">修改密码</el-button>
              <el-button link type="danger" :icon="Delete" @click="handleDelete(row)">删除</el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="shareList.length === 0 && !loading" description="暂无分享记录" />
    </div>
    
    <!-- 修改密码对话框 -->
    <el-dialog 
      v-model="showPasswordDialog" 
      title="修改分享密码" 
      width="450px"
      :close-on-click-modal="false"
    >
      <el-form label-width="80px">
        <el-form-item label="文件名">
          <el-input v-model="currentShare.file_name" disabled />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input 
            v-model="newPassword" 
            placeholder="请输入新的访问密码"
            maxlength="20"
            show-word-limit
            clearable
          >
            <template #append>
              <el-button @click="generateRandomPassword">随机生成</el-button>
            </template>
          </el-input>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showPasswordDialog = false">取消</el-button>
        <el-button type="primary" :loading="updating" @click="handleConfirmUpdatePassword">确定修改</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Refresh, 
  CopyDocument, 
  Delete,
  Share,
  Document,
  Lock,
  Clock,
  Calendar,
  Edit
} from '@element-plus/icons-vue'
import { getShareList, deleteShare, updateSharePassword } from '@/api/share'
import type { ShareInfo } from '@/types'

const loading = ref(false)
const shareList = ref<ShareInfo[]>([])
const showPasswordDialog = ref(false)
const updating = ref(false)
const newPassword = ref('')
const currentShare = reactive<Partial<ShareInfo>>({})

// 加载分享列表
const loadShareList = async () => {
  loading.value = true
  try {
    const res = await getShareList()
    if (res.code === 200) {
      shareList.value = res.data || []
    } else {
      ElMessage.error(res.message || '加载失败')
    }
  } catch (error) {
    ElMessage.error('加载分享列表失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

// 获取分享链接（不包含密码）
const getShareUrl = (token: string) => {
  return `${window.location.origin}/api/share/download?token=${token}`
}

// 复制分享链接
const copyShareLink = async (share: ShareInfo) => {
  const shareUrl = getShareUrl(share.token)
  try {
    await navigator.clipboard.writeText(shareUrl)
    ElMessage.success('已复制到剪贴板')
  } catch (err) {
    ElMessage.error('复制失败')
  }
}

// 删除分享
const handleDelete = (share: ShareInfo) => {
  ElMessageBox.confirm('确定要删除该分享吗？', '提示', {
    type: 'warning'
  }).then(async () => {
    try {
      const res = await deleteShare(share.id)
      if (res.code === 200) {
        ElMessage.success('删除成功')
        loadShareList()
      } else {
        ElMessage.error(res.message || '删除失败')
      }
    } catch (error: any) {
      ElMessage.error(error.message || '删除失败')
    }
  })
}

// 打开修改密码对话框
const handleUpdatePassword = (share: ShareInfo) => {
  Object.assign(currentShare, share)
  newPassword.value = ''
  showPasswordDialog.value = true
}

// 生成随机密码
const generateRandomPassword = () => {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789'
  let password = ''
  for (let i = 0; i < 6; i++) {
    password += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  newPassword.value = password
}

// 确认修改密码
const handleConfirmUpdatePassword = async () => {
  if (!newPassword.value) {
    ElMessage.warning('请输入新密码')
    return
  }
  
  updating.value = true
  try {
    const res = await updateSharePassword(currentShare.id!, newPassword.value)
    if (res.code === 200) {
      ElMessage.success('修改密码成功')
      showPasswordDialog.value = false
      loadShareList()
    } else {
      ElMessage.error(res.message || '修改密码失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '修改密码失败')
  } finally {
    updating.value = false
  }
}

// 格式化日期
const formatDate = (dateStr: string): string => {
  if (!dateStr || dateStr === '') {
    return '-'
  }
  
  // 处理 "2006-01-02 15:04:05" 格式
  // 将空格替换为 T 以符合 ISO 8601 格式
  const isoDateStr = dateStr.replace(' ', 'T')
  const date = new Date(isoDateStr)
  
  if (isNaN(date.getTime())) {
    return dateStr // 如果解析失败，返回原始字符串
  }
  
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
}

onMounted(() => {
  loadShareList()
})
</script>

<style scoped>
.shares-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 4px;
}

.header-card {
  padding: 16px 24px;
  border-radius: 16px;
  display: flex;
  align-items: center;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.title-section {
  display: flex;
  align-items: center;
  gap: 12px;
}

.title-icon {
  color: var(--primary-color);
  filter: drop-shadow(0 2px 4px rgba(99, 102, 241, 0.3));
}

.title-section h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  background: linear-gradient(135deg, var(--text-primary) 0%, var(--text-secondary) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.table-card {
  flex: 1;
  border-radius: 16px;
  padding: 8px; /* Inner padding for glass effect */
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

:deep(.el-table) {
  background: transparent !important;
  --el-table-tr-bg-color: transparent;
  --el-table-header-bg-color: transparent;
}

:deep(.el-table th.el-table__cell) {
  background: transparent !important;
  color: var(--text-secondary);
  font-weight: 600;
  font-size: 13px;
}

:deep(.el-table tr) {
  background: transparent !important;
  transition: all 0.2s;
}

:deep(.el-table--enable-row-hover .el-table__body tr:hover > td.el-table__cell) {
  background: rgba(99, 102, 241, 0.05) !important;
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.file-name {
  font-weight: 500;
  color: var(--text-primary);
}

:deep(.el-button-group) {
  display: flex;
  gap: 4px;
}

:deep(.el-tag) {
  border-radius: 6px;
}
</style>

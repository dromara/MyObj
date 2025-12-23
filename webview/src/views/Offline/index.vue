<template>
  <div class="offline-page">
    <!-- 标题栏 -->
    <el-card shadow="never" class="header-card">
      <div class="page-header">
        <div class="header-left">
          <h2>离线下载</h2>
          <el-tag type="info">{{ taskList.length }} 个任务</el-tag>
        </div>
        <div class="header-right">
          <el-button type="primary" icon="Link" @click="showUrlDialog = true">新建 URL 下载</el-button>
          <el-button icon="Refresh" @click="refreshTaskList" circle />
        </div>
      </div>
    </el-card>

    <!-- 任务列表 -->
    <el-card shadow="never" class="task-list-card">
      <el-table :data="taskList" v-loading="loading" style="width: 100%">
        <el-table-column label="文件名" min-width="300">
          <template #default="{ row }">
            <div class="file-name-cell">
              <el-icon :size="24" color="#409EFF"><Document /></el-icon>
              <div class="file-info">
                <div class="file-name">{{ row.file_name || '未知文件' }}</div>
                <div class="file-url" v-if="row.url">{{ truncateUrl(row.url) }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.state)">{{ row.state_text }}</el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="进度" width="200">
          <template #default="{ row }">
            <div class="progress-cell">
              <el-progress 
                :percentage="row.progress" 
                :status="row.state === 3 ? 'success' : row.state === 4 ? 'exception' : undefined"
              />
              <span class="progress-text">{{ formatSize(row.downloaded_size) }} / {{ formatSize(row.file_size) }}</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="速度" width="120">
          <template #default="{ row }">
            <span v-if="row.state === 1">{{ formatSpeed(row.speed) }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.create_time) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button 
                v-if="row.state === 1"
                link 
                icon="VideoPause" 
                @click="pauseTask(row.id)"
              >
                暂停
              </el-button>
              <el-button 
                v-if="row.state === 2"
                link 
                icon="VideoPlay" 
                type="success"
                @click="resumeTask(row.id)"
              >
                继续
              </el-button>
              <el-button 
                v-if="row.state === 0 || row.state === 1"
                link 
                icon="Close" 
                type="warning"
                @click="cancelTask(row.id)"
              >
                取消
              </el-button>
              <el-button 
                v-if="row.state === 3 || row.state === 4"
                link 
                icon="Delete" 
                type="danger"
                @click="deleteTask(row.id)"
              >
                删除
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
      
      <el-empty v-if="taskList.length === 0 && !loading" description="暂无下载任务" />
    </el-card>

    <!-- URL 下载对话框 -->
    <el-dialog v-model="showUrlDialog" title="新建 URL 下载" width="600px" @open="buildFolderTree">
      <el-form :model="urlForm" :rules="urlRules" ref="urlFormRef" label-width="100px">
        <el-form-item label="下载链接" prop="url">
          <el-input 
            v-model="urlForm.url" 
            placeholder="请输入 HTTP/HTTPS 下载链接"
            type="textarea"
            :rows="3"
          />
        </el-form-item>
        <el-form-item label="保存位置">
          <el-tree-select
            v-model="urlForm.virtual_path"
            :data="folderTreeData"
            :render-after-expand="false"
            placeholder="请选择保存目录（默认：/离线下载/）"
            :loading="loadingTree"
            style="width: 100%"
            check-strictly
            :props="{ label: 'label', value: 'value', children: 'children' }"
            :default-expand-all="true"
            node-key="value"
          />
        </el-form-item>
        <el-form-item label="加密存储">
          <el-switch v-model="urlForm.enable_encryption" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUrlDialog = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreateUrlDownload">创建任务</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import {
  getDownloadTaskList,
  createOfflineDownload,
  pauseDownload,
  resumeDownload,
  cancelDownload,
  deleteDownload,
  type OfflineDownloadTask
} from '@/api/download'
import { getVirtualPathTree } from '@/api/file'
import { formatSize, formatDate, formatSpeed, truncateUrl, getTaskStatusType } from '@/utils'

const { proxy } = getCurrentInstance() as ComponentInternalInstance

const loading = ref(false)
const creating = ref(false)
const taskList = ref<OfflineDownloadTask[]>([])
const showUrlDialog = ref(false)
let refreshTimer: number | null = null
const loadingTree = ref(false)
const folderTreeData = ref<any[]>([])

const urlFormRef = ref<FormInstance>()

const urlForm = reactive({
  url: '',
  virtual_path: '',
  enable_encryption: false
})

const urlRules: FormRules = {
  url: [
    { required: true, message: '请输入下载链接', trigger: 'blur' },
    { pattern: /^https?:\/\//, message: '请输入正确的 HTTP/HTTPS 链接', trigger: 'blur' }
  ]
}

// 加载任务列表
const loadTaskList = async () => {
  loading.value = true
  try {
    // 查询所有类型的离线下载任务（type < 7），不包含网盘文件下载（type=7）
    // 由于后端不支持 type < 7 的查询，这里先查询所有任务，然后在前端过滤
    // 或者可以分别查询 type=0,1,2,3,4,5,6，但这样需要多次请求
    // 为了简化，暂时保持前端过滤，但可以优化为后端支持范围查询
    const res = await getDownloadTaskList({
      page: 1,
      pageSize: 100,
      state: -1 // 查询所有状态
    })
    if (res.code === 200 && res.data) {
      // 过滤掉网盘下载任务（type=7），只显示离线下载（type=0-6）
      taskList.value = (res.data.tasks || []).filter((task: any) => task.type !== 7)
    }
  } catch (error: any) {
    proxy?.$modal.msgError(error.message || '加载任务列表失败')
  } finally {
    loading.value = false
  }
}

// 刷新任务列表
const refreshTaskList = () => {
  loadTaskList()
}

// 构建文件夹树结构
const buildFolderTree = async () => {
  loadingTree.value = true
  try {
    const res = await getVirtualPathTree()
    
    if (res.code !== 200 || !res.data) {
      proxy?.$modal.msgError('获取目录树失败')
      return
    }
    
    // 后端返回的是 VirtualPath 数组
    const virtualPaths = res.data as Array<{
      id: number
      path: string
      parent_level: string
      is_dir: boolean
    }>
    
    // 构建树形结构
    const pathMap = new Map<string, any>()
    const rootNodes: any[] = []
    
    // 第一步：创建所有节点
    virtualPaths.forEach(vp => {
      const nodeId = String(vp.id)
      // 获取路径最后一段作为显示名称
      const pathParts = vp.path.split('/').filter(p => p !== '')
      const displayName = pathParts.length > 0 ? pathParts[pathParts.length - 1] : vp.path || '根目录'
      
      pathMap.set(nodeId, {
        value: nodeId,
        label: displayName,
        children: [],
        _raw: vp
      })
    })
    
    // 第二步：构建父子关系
    virtualPaths.forEach(vp => {
      const nodeId = String(vp.id)
      const node = pathMap.get(nodeId)
      
      if (!node) return
      
      // 如果有父级路径，添加到父节点的 children
      if (vp.parent_level && vp.parent_level !== '' && vp.parent_level !== '0') {
        const parentNode = pathMap.get(vp.parent_level)
        if (parentNode) {
          parentNode.children.push(node)
        } else {
          // 父节点不存在，作为根节点
          rootNodes.push(node)
        }
      } else {
        // 没有父级，是根节点
        rootNodes.push(node)
      }
    })
    
    // 清理空 children 数组
    const cleanEmptyChildren = (nodes: any[]) => {
      nodes.forEach(node => {
        if (node.children && node.children.length === 0) {
          delete node.children
        } else if (node.children) {
          cleanEmptyChildren(node.children)
        }
      })
    }
    cleanEmptyChildren(rootNodes)
    
    folderTreeData.value = rootNodes
  } catch (error: any) {
    proxy?.$modal.msgError(error.message || '获取目录树失败')
  } finally {
    loadingTree.value = false
  }
}

// 创建 URL 下载任务
const handleCreateUrlDownload = async () => {
  if (!urlFormRef.value) return
  
  await urlFormRef.value.validate(async (valid: boolean) => {
    if (valid) {
      creating.value = true
      try {
        const res = await createOfflineDownload({
          url: urlForm.url,
          virtual_path: urlForm.virtual_path || undefined,
          enable_encryption: urlForm.enable_encryption
        })
        
        if (res.code === 200) {
          proxy?.$modal.msgSuccess('任务创建成功')
          showUrlDialog.value = false
          urlForm.url = ''
          urlForm.virtual_path = ''
          urlForm.enable_encryption = false
          loadTaskList()
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || '创建任务失败')
      } finally {
        creating.value = false
      }
    }
  })
}

// 暂停任务
const pauseTask = async (taskId: string) => {
  try {
    await pauseDownload(taskId)
    proxy?.$modal.msgSuccess('已暂停')
    loadTaskList()
  } catch (error: any) {
    proxy?.$modal.msgError(error.message || '暂停失败')
  }
}

// 恢复任务
const resumeTask = async (taskId: string) => {
  try {
    await resumeDownload(taskId)
    proxy?.$modal.msgSuccess('已恢复')
    loadTaskList()
  } catch (error: any) {
    proxy?.$modal.msgError(error.message || '恢复失败')
  }
}

// 取消任务
const cancelTask = async (taskId: string) => {
  try {
    await proxy?.$modal.confirm('确认取消该任务？')
    
    await cancelDownload(taskId)
    proxy?.$modal.msgSuccess('已取消')
    loadTaskList()
  } catch (error: any) {
    if (error !== 'cancel') {
      proxy?.$modal.msgError(error.message || '取消失败')
    }
  }
}

// 删除任务
const deleteTask = async (taskId: string) => {
  try {
    await proxy?.$modal.confirm('确认删除该任务？')
    
    await deleteDownload(taskId)
    proxy?.$modal.msgSuccess('已删除')
    loadTaskList()
  } catch (error: any) {
    if (error !== 'cancel') {
      proxy?.$modal.msgError(error.message || '删除失败')
    }
  }
}

// 使用 getTaskStatusType 作为 getStatusType 的别名
const getStatusType = getTaskStatusType

// 页面加载时获取任务列表
onMounted(() => {
  loadTaskList()
  
  // 每 3 秒自动刷新任务状态
  refreshTimer = window.setInterval(() => {
    loadTaskList()
  }, 3000)
})

// 页面销毁时清除定时器
onBeforeUnmount(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.offline-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.header-card {
  flex-shrink: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-left h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.header-right {
  display: flex;
  gap: 12px;
}

.task-list-card {
  flex: 1;
  overflow: hidden;
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.file-info {
  flex: 1;
  overflow: hidden;
}

.file-name {
  font-size: 14px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-url {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-top: 2px;
}

.progress-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.progress-text {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>

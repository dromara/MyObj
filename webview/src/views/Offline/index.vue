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
          <el-button type="primary" icon="Upload" @click="showTorrentDialog = true">新建种子下载</el-button>
          <el-button icon="Refresh" @click="refreshTaskList">刷新</el-button>
        </div>
      </div>
    </el-card>

    <!-- 任务列表 -->
    <el-card shadow="never" class="task-list-card">
      <!-- PC端：表格布局 -->
      <el-table 
        :data="taskList" 
        v-loading="loading" 
        class="offline-table desktop-table"
        empty-text="暂无下载任务"
      >
        <el-table-column label="文件名" min-width="300" class-name="mobile-name-column">
          <template #default="{ row }">
            <div class="file-name-cell">
              <el-icon :size="24" color="#409EFF"><Document /></el-icon>
              <div class="file-info">
                <file-name-tooltip :file-name="row.file_name || '未知文件'" view-mode="table" custom-class="file-name" />
                <div class="file-url mobile-hide" v-if="row.url">{{ truncateUrl(row.url) }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="状态" width="120" class-name="mobile-hide">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.state)">{{ row.state_text }}</el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="进度" width="200" class-name="mobile-progress-column">
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
        
        <el-table-column label="速度" width="120" class-name="mobile-hide">
          <template #default="{ row }">
            <span v-if="row.state === 1">{{ formatSpeed(row.speed) }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        
        <el-table-column label="创建时间" width="180" class-name="mobile-hide">
          <template #default="{ row }">
            {{ formatDate(row.create_time) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="200" fixed="right" class-name="mobile-actions-column">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button 
                v-if="row.state === 1"
                link 
                icon="VideoPause" 
                type="warning"
                @click="pauseTask(row.id)"
                size="small"
              >
                暂停
              </el-button>
              <el-button 
                v-if="row.state === 2"
                link 
                icon="VideoPlay" 
                type="primary"
                @click="resumeTask(row.id)"
                size="small"
              >
                继续
              </el-button>
              <el-button 
                v-if="row.state === 0 || row.state === 1 || row.state === 2"
                link 
                icon="Close" 
                type="danger"
                @click="cancelTask(row.id)"
                size="small"
              >
                取消
              </el-button>
              <el-button 
                v-if="row.state === 3 || row.state === 4"
                link 
                icon="Delete" 
                type="danger"
                @click="deleteTask(row.id)"
                size="small"
              >
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 移动端：卡片布局 -->
      <div class="mobile-task-list" v-loading="loading">
        <div 
          v-for="row in taskList" 
          :key="row.id" 
          class="mobile-task-item"
        >
          <div class="task-item-header">
            <div class="task-item-info">
              <el-icon :size="24" color="#409EFF" class="task-icon"><Document /></el-icon>
              <div class="task-name-wrapper">
                <file-name-tooltip :file-name="row.file_name || row.url || '未知文件'" view-mode="list" custom-class="task-name" />
                <div class="task-meta">
                  <el-tag :type="getStatusType(row.state)" size="small" effect="plain">
                    {{ row.state_text }}
                  </el-tag>
                  <span class="task-size">{{ formatSize(row.downloaded_size) }} / {{ formatSize(row.file_size) }}</span>
                  <span v-if="row.state === 1" class="task-speed">{{ formatSpeed(row.speed) }}</span>
                </div>
                <div v-if="row.url" class="task-url">{{ truncateUrl(row.url, 40) }}</div>
              </div>
            </div>
            <div class="task-actions">
              <el-button 
                v-if="row.state === 1"
                link 
                type="warning"
                @click.stop="pauseTask(row.id)"
                class="action-btn"
              >
                <el-icon><VideoPause /></el-icon>
              </el-button>
              <el-button 
                v-if="row.state === 2"
                link 
                type="primary"
                @click.stop="resumeTask(row.id)"
                class="action-btn"
              >
                <el-icon><VideoPlay /></el-icon>
              </el-button>
              <el-button 
                v-if="row.state === 0 || row.state === 1 || row.state === 2"
                link 
                type="danger"
                @click.stop="cancelTask(row.id)"
                class="action-btn"
              >
                <el-icon><Close /></el-icon>
              </el-button>
              <el-button 
                v-if="row.state === 3 || row.state === 4"
                link 
                type="danger"
                @click.stop="deleteTask(row.id)"
                class="action-btn"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </div>
          <div class="task-progress-wrapper">
            <el-progress 
              :percentage="row.progress" 
              :status="row.state === 3 ? 'success' : row.state === 4 ? 'exception' : undefined"
              :stroke-width="6"
              text-inside
              class="task-progress"
            />
          </div>
        </div>
      </div>
      
      <el-empty v-if="taskList.length === 0 && !loading" description="暂无下载任务" />
    </el-card>

    <!-- 种子下载对话框 -->
    <el-dialog 
      v-model="showTorrentDialog" 
      title="新建种子下载" 
      :width="isMobile ? '95%' : '800px'"
      @open="handleTorrentDialogOpen"
      @close="handleTorrentDialogClose"
      class="torrent-download-dialog"
    >
      <el-tabs v-model="torrentInputType" class="torrent-tabs">
        <el-tab-pane label="上传种子文件" name="file">
          <el-upload
            ref="torrentUploadRef"
            :auto-upload="false"
            :on-change="handleTorrentFileChange"
            :limit="1"
            accept=".torrent"
            drag
            class="torrent-upload"
          >
            <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
            <div class="el-upload__text">
              将种子文件拖到此处，或<em>点击上传</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">
                支持 .torrent 文件，最大 10MB
              </div>
            </template>
          </el-upload>
          <div v-if="torrentFileName" class="torrent-file-info">
            <el-icon><Document /></el-icon>
            <span>{{ torrentFileName }}</span>
            <el-button link type="danger" @click="clearTorrentFile">清除</el-button>
          </div>
        </el-tab-pane>
        <el-tab-pane label="输入磁力链接" name="magnet">
          <el-form-item label="磁力链接">
            <el-input 
              v-model="torrentForm.magnetLink" 
              placeholder="请输入磁力链接（magnet:?xt=urn:btih:...）"
              type="textarea"
              :rows="3"
            />
          </el-form-item>
        </el-tab-pane>
      </el-tabs>

      <!-- 解析按钮 -->
      <div v-if="!torrentParseResult" class="parse-section">
        <el-button 
          type="primary" 
          :loading="parsing" 
          :disabled="!canParse"
          @click="handleParseTorrent"
          style="width: 100%"
        >
          解析种子
        </el-button>
      </div>

      <!-- 解析结果：文件列表 -->
      <div v-if="torrentParseResult" class="parse-result-section">
        <div class="torrent-info">
          <h4>{{ torrentParseResult.name }}</h4>
          <div class="torrent-meta">
            <el-tag type="info">共 {{ torrentParseResult.files.length }} 个文件</el-tag>
            <el-tag type="info">{{ formatSize(torrentParseResult.total_size) }}</el-tag>
          </div>
        </div>
        <el-divider />
        <div class="file-selection-section">
          <div class="selection-header">
            <el-checkbox 
              v-model="selectAllFiles" 
              :indeterminate="isIndeterminate"
              @change="handleSelectAll"
            >
              全选
            </el-checkbox>
            <span class="selected-count">已选择 {{ selectedFileIndexes.length }} 个文件</span>
          </div>
          <el-scrollbar height="300px" class="file-list-scrollbar">
            <el-table 
              ref="torrentFileTableRef"
              :data="torrentParseResult.files" 
              @selection-change="handleFileSelectionChange"
              :row-key="(row: any) => row.index"
            >
              <el-table-column type="selection" width="55" :reserve-selection="true" />
              <el-table-column label="文件名" min-width="200">
                <template #default="{ row }">
                  <file-name-tooltip :file-name="row.name" view-mode="table" custom-class="torrent-file-name" />
                </template>
              </el-table-column>
              <el-table-column label="大小" width="120">
                <template #default="{ row }">
                  {{ formatSize(row.size) }}
                </template>
              </el-table-column>
              <el-table-column label="路径" min-width="150" class-name="mobile-hide">
                <template #default="{ row }">
                  <span class="file-path">{{ row.path }}</span>
                </template>
              </el-table-column>
            </el-table>
          </el-scrollbar>
        </div>
      </div>

      <!-- 下载配置表单 -->
      <el-form 
        v-if="torrentParseResult" 
        :model="torrentForm" 
        :rules="torrentRules" 
        ref="torrentFormRef" 
        label-width="100px"
        style="margin-top: 20px"
      >
        <el-form-item label="保存位置">
          <el-tree-select
            v-model="torrentForm.virtual_path"
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
          <el-switch v-model="torrentForm.enable_encryption" />
        </el-form-item>
        <el-form-item 
          v-if="torrentForm.enable_encryption" 
          label="加密密码" 
          prop="file_password"
        >
          <el-input 
            v-model="torrentForm.file_password" 
            type="password"
            placeholder="请输入加密密码"
            show-password
            maxlength="32"
          />
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px;">
            下载文件时需要使用此密码解密
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showTorrentDialog = false">取消</el-button>
        <el-button 
          v-if="!torrentParseResult" 
          type="primary" 
          :loading="parsing" 
          :disabled="!canParse"
          @click="handleParseTorrent"
        >
          解析种子
        </el-button>
        <el-button 
          v-else
          type="primary" 
          :loading="creatingTorrent" 
          :disabled="selectedFileIndexes.length === 0"
          @click="handleStartTorrentDownload"
        >
          开始下载（{{ selectedFileIndexes.length }} 个文件）
        </el-button>
      </template>
    </el-dialog>

    <!-- URL 下载对话框 -->
    <el-dialog 
      v-model="showUrlDialog" 
      title="新建 URL 下载" 
      :width="isMobile ? '95%' : '600px'"
      @open="buildFolderTree"
      class="url-download-dialog"
    >
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
        <el-form-item 
          v-if="urlForm.enable_encryption" 
          label="加密密码" 
          prop="file_password"
        >
          <el-input 
            v-model="urlForm.file_password" 
            type="password"
            placeholder="请输入加密密码"
            show-password
            maxlength="32"
          />
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px;">
            下载文件时需要使用此密码解密
          </div>
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
  parseTorrent,
  startTorrentDownload,
  type OfflineDownloadTask,
  type ParseTorrentResponse,
  type TorrentFileInfo
} from '@/api/download'
import { getVirtualPathTree } from '@/api/file'
import { formatSize, formatDate, formatSpeed, truncateUrl, getTaskStatusType } from '@/utils'
import { useResponsive } from '@/composables/useResponsive'

const { proxy } = getCurrentInstance() as ComponentInternalInstance

// 使用响应式检测 composable
const { isMobile } = useResponsive()

const loading = ref(false)
const creating = ref(false)
const taskList = ref<OfflineDownloadTask[]>([])
const showUrlDialog = ref(false)
const showTorrentDialog = ref(false)
let refreshTimer: number | null = null // 支持 setTimeout 和 setInterval
const loadingTree = ref(false)
const folderTreeData = ref<any[]>([])

const urlFormRef = ref<FormInstance>()
const torrentFormRef = ref<FormInstance>()
const torrentUploadRef = ref()
const torrentFileTableRef = ref()

const urlForm = reactive({
  url: '',
  virtual_path: '',
  enable_encryption: false,
  file_password: ''
})

// 种子下载相关状态
const torrentInputType = ref<'file' | 'magnet'>('file')
const torrentFileName = ref('')
const torrentFileContent = ref('') // Base64 编码的种子文件内容
const parsing = ref(false)
const creatingTorrent = ref(false)
const torrentParseResult = ref<ParseTorrentResponse | null>(null)
const selectedFileIndexes = ref<number[]>([])
const selectAllFiles = ref(false)
const isIndeterminate = ref(false)

const torrentForm = reactive({
  magnetLink: '',
  virtual_path: '',
  enable_encryption: false,
  file_password: ''
})

const urlRules: FormRules = {
  url: [
    { required: true, message: '请输入下载链接', trigger: 'blur' },
    { pattern: /^https?:\/\//, message: '请输入正确的 HTTP/HTTPS 链接', trigger: 'blur' }
  ],
  file_password: [
    { 
      validator: (_rule: any, value: any, callback: any) => {
        if (urlForm.enable_encryption && !value) {
          callback(new Error('加密存储时必须设置密码'))
        } else if (value && value.length < 6) {
          callback(new Error('密码长度至少为6位'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

const torrentRules: FormRules = {
  file_password: [
    { 
      validator: (_rule: any, value: any, callback: any) => {
        if (torrentForm.enable_encryption && !value) {
          callback(new Error('加密存储时必须设置密码'))
        } else if (value && value.length < 6) {
          callback(new Error('密码长度至少为6位'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 计算是否可以解析种子
const canParse = computed(() => {
  if (torrentInputType.value === 'file') {
    return !!torrentFileContent.value
  } else {
    return !!torrentForm.magnetLink && torrentForm.magnetLink.trim().startsWith('magnet:')
  }
})

// 加载任务列表
const loadTaskList = async () => {
  // 智能刷新时不显示 loading，避免频繁闪烁
  // 只在手动刷新或首次加载时显示 loading
  const isManualRefresh = !refreshTimer
  if (isManualRefresh) {
    loading.value = true
  }
  
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
      const newTasks = (res.data.tasks || []).filter((task: any) => task.type !== 7)
      
      // 确保数据更新（即使值相同，也要触发响应式更新）
      // 通过创建新数组来触发 Vue 的响应式更新
      taskList.value = newTasks.map((task: any) => ({ ...task }))
      
      // 调试日志：检查数据更新（仅在开发环境）
      if (import.meta.env.DEV) {
        const downloadingTasks = newTasks.filter((t: any) => t.state === 1)
        if (downloadingTasks.length > 0) {
          downloadingTasks.forEach((task: any) => {
            proxy?.$log?.debug('任务数据更新', {
              id: task.id,
              progress: task.progress,
              speed: task.speed,
              downloaded_size: task.downloaded_size,
              update_time: task.update_time
            })
          })
        }
      }
    }
  } catch (error: any) {
    // 智能刷新时静默处理错误，避免频繁弹窗
    if (isManualRefresh) {
      proxy?.$modal.msgError(error.message || '加载任务列表失败')
    } else {
      proxy?.$log.warn('刷新任务列表失败:', error)
    }
  } finally {
    if (isManualRefresh) {
      loading.value = false
    }
  }
}

// 刷新任务列表
const refreshTaskList = () => {
  loadTaskList().then(() => {
    // 刷新后重新启动智能刷新
    startSmartRefresh()
  })
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
          enable_encryption: urlForm.enable_encryption,
          file_password: urlForm.enable_encryption ? urlForm.file_password : undefined
        })
        
        if (res.code === 200) {
          proxy?.$modal.msgSuccess('任务创建成功')
          showUrlDialog.value = false
          urlForm.url = ''
          urlForm.virtual_path = ''
          urlForm.enable_encryption = false
          urlForm.file_password = ''
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

// 处理种子文件选择
const handleTorrentFileChange = (file: any) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    const result = e.target?.result as string
    // 移除 data URL 前缀（如 "data:application/x-bittorrent;base64,"）
    const base64Content = result.includes(',') ? result.split(',')[1] : result
    torrentFileContent.value = base64Content
    torrentFileName.value = file.name
  }
  reader.onerror = () => {
    proxy?.$modal.msgError('读取种子文件失败')
  }
  reader.readAsDataURL(file.raw)
}

// 清除种子文件
const clearTorrentFile = () => {
  torrentFileContent.value = ''
  torrentFileName.value = ''
  if (torrentUploadRef.value) {
    torrentUploadRef.value.clearFiles()
  }
}

// 处理种子对话框打开
const handleTorrentDialogOpen = () => {
  buildFolderTree()
}

// 处理种子对话框关闭
const handleTorrentDialogClose = () => {
  // 重置所有状态
  torrentInputType.value = 'file'
  torrentFileName.value = ''
  torrentFileContent.value = ''
  torrentForm.magnetLink = ''
  torrentForm.virtual_path = ''
  torrentForm.enable_encryption = false
  torrentForm.file_password = ''
  torrentParseResult.value = null
  selectedFileIndexes.value = []
  selectAllFiles.value = false
  isIndeterminate.value = false
  if (torrentUploadRef.value) {
    torrentUploadRef.value.clearFiles()
  }
}

// 解析种子
const handleParseTorrent = async () => {
  if (!canParse.value) {
    proxy?.$modal.msgWarning('请先上传种子文件或输入磁力链接')
    return
  }

  parsing.value = true
  try {
    const content = torrentInputType.value === 'file' 
      ? torrentFileContent.value 
      : torrentForm.magnetLink.trim()

    const res = await parseTorrent({ content })
    
    if (res.code === 200 && res.data) {
      torrentParseResult.value = res.data
      // 等待 DOM 更新后设置默认全选
      await nextTick()
      // 默认全选所有文件
      if (torrentFileTableRef.value && res.data.files.length > 0) {
        res.data.files.forEach((file: TorrentFileInfo) => {
          torrentFileTableRef.value.toggleRowSelection(file, true)
        })
      }
      selectedFileIndexes.value = res.data.files.map((f: TorrentFileInfo) => f.index)
      selectAllFiles.value = true
      isIndeterminate.value = false
      proxy?.$modal.msgSuccess('解析成功')
    } else {
      proxy?.$modal.msgError(res.message || '解析失败')
    }
  } catch (error: any) {
    proxy?.$modal.msgError(error.message || '解析失败')
  } finally {
    parsing.value = false
  }
}

// 处理文件选择变化
const handleFileSelectionChange = (selection: TorrentFileInfo[]) => {
  selectedFileIndexes.value = selection.map((f: TorrentFileInfo) => f.index)
  const total = torrentParseResult.value?.files.length || 0
  const selected = selectedFileIndexes.value.length
  selectAllFiles.value = selected === total && total > 0
  isIndeterminate.value = selected > 0 && selected < total
}

// 处理全选
const handleSelectAll = (val: boolean | string | number) => {
  if (!torrentParseResult.value || !torrentFileTableRef.value) return
  
  const checked = Boolean(val)
  
  if (checked) {
    // 全选所有行
    torrentParseResult.value.files.forEach((file: TorrentFileInfo) => {
      torrentFileTableRef.value.toggleRowSelection(file, true)
    })
    selectedFileIndexes.value = torrentParseResult.value.files.map(f => f.index)
  } else {
    // 取消全选
    torrentParseResult.value.files.forEach((file: TorrentFileInfo) => {
      torrentFileTableRef.value.toggleRowSelection(file, false)
    })
    selectedFileIndexes.value = []
  }
  isIndeterminate.value = false
}

// 开始种子下载
const handleStartTorrentDownload = async () => {
  if (!torrentFormRef.value || !torrentParseResult.value) return
  
  if (selectedFileIndexes.value.length === 0) {
    proxy?.$modal.msgWarning('请至少选择一个文件')
    return
  }

  await torrentFormRef.value.validate(async (valid: boolean) => {
    if (valid) {
      creatingTorrent.value = true
      try {
        const content = torrentInputType.value === 'file' 
          ? torrentFileContent.value 
          : torrentForm.magnetLink.trim()

        const res = await startTorrentDownload({
          content,
          file_indexes: selectedFileIndexes.value,
          virtual_path: torrentForm.virtual_path || undefined,
          enable_encryption: torrentForm.enable_encryption,
          file_password: torrentForm.enable_encryption ? torrentForm.file_password : undefined
        })
        
        if (res.code === 200 && res.data) {
          proxy?.$modal.msgSuccess(`任务创建成功，共创建 ${res.data.task_count} 个下载任务`)
          showTorrentDialog.value = false
          loadTaskList()
        } else {
          proxy?.$modal.msgError(res.message || '创建任务失败')
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || '创建任务失败')
      } finally {
        creatingTorrent.value = false
      }
    }
  })
}

// 智能刷新：根据任务状态使用不同的刷新频率
const startSmartRefresh = () => {
  if (refreshTimer) {
    clearTimeout(refreshTimer)
    clearInterval(refreshTimer)
  }
  
  const refresh = async () => {
    await loadTaskList()
    
    // 检查是否有正在下载的任务
    const hasActiveTasks = taskList.value.some((task: any) => task.state === 1) // state=1 表示下载中
    
    if (hasActiveTasks) {
      // 有正在下载的任务，1秒后再次刷新（快速更新）
      refreshTimer = window.setTimeout(refresh, 1000)
    } else {
      // 没有正在下载的任务，3秒后再次刷新（节省资源）
      refreshTimer = window.setTimeout(refresh, 3000)
    }
  }
  
  // 初始延迟1秒后开始刷新
  refreshTimer = window.setTimeout(refresh, 1000)
}

// 页面加载时获取任务列表
onMounted(() => {
  loadTaskList()
  
  // 启动智能刷新
  startSmartRefresh()
})

// 页面销毁时清除定时器
onBeforeUnmount(() => {
  if (refreshTimer) {
    // 支持 setTimeout 和 setInterval
    if (typeof refreshTimer === 'number') {
      clearTimeout(refreshTimer)
      clearInterval(refreshTimer)
    }
    refreshTimer = null
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

.action-buttons {
  display: flex;
  gap: 8px;
  justify-content: center;
}

/* PC端表格样式 */
.desktop-table {
  display: table;
}

/* 表格移动端隐藏列 */
.offline-table :deep(.mobile-hide) {
  display: table-cell;
}

.offline-table :deep(.mobile-name-column) {
  min-width: 200px;
}

.offline-table :deep(.mobile-progress-column) {
  min-width: 180px;
}

.offline-table :deep(.mobile-actions-column) {
  width: auto;
  min-width: 120px;
}

/* 移动端卡片列表 */
.mobile-task-list {
  display: none;
}

.mobile-task-item {
  padding: 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color-overlay);
  transition: background-color 0.2s;
  border-radius: 8px;
  margin-bottom: 12px;
}

.mobile-task-item:last-child {
  border-bottom: none;
  margin-bottom: 0;
}

.mobile-task-item:active {
  background-color: var(--el-fill-color-light);
}

.task-item-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.task-item-info {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.task-icon {
  flex-shrink: 0;
  margin-top: 2px;
}

.task-name-wrapper {
  flex: 1;
  min-width: 0;
}

.task-name {
  font-size: 15px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 6px;
}

.task-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 4px;
}

.task-size {
  white-space: nowrap;
}

.task-speed {
  color: var(--el-color-primary);
  font-weight: 500;
  white-space: nowrap;
}

.task-url {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-top: 4px;
}

.task-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
  margin-left: 8px;
}

.action-btn {
  padding: 4px;
  min-width: auto;
}

.action-btn :deep(.el-icon) {
  font-size: 18px;
}

.task-progress-wrapper {
  width: 100%;
}

.task-progress {
  width: 100%;
}

/* 移动端响应式 */
@media (max-width: 1024px) {
  .desktop-table {
    display: none !important;
  }
  
  .mobile-task-list {
    display: block;
  }
  
  .header-card {
    padding: 12px 16px;
  }
  
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .header-left {
    width: 100%;
    justify-content: space-between;
  }
  
  .header-left h2 {
    font-size: 18px;
  }
  
  .header-right {
    width: 100%;
    justify-content: space-between;
    gap: 8px;
  }
  
  .header-right .el-button:first-child {
    flex: 1;
  }
  
  .header-right .el-button:last-child {
    flex-shrink: 0;
    min-width: auto;
  }
  
  .file-info {
    min-width: 0;
  }
  
  .file-url {
    font-size: 11px;
  }
  
  .url-download-dialog :deep(.el-dialog) {
    width: 95% !important;
    margin: 0 auto;
  }
  
  .url-download-dialog :deep(.el-form-item__label) {
    font-size: 14px;
  }
}

@media (max-width: 480px) {
  .mobile-task-item {
    padding: 12px;
  }
  
  .task-name {
    font-size: 14px;
  }
  
  .task-meta {
    font-size: 11px;
  }
  
  .task-url {
    font-size: 10px;
  }
  
  .url-download-dialog :deep(.el-dialog) {
    width: 100% !important;
    margin: 0;
    border-radius: 0;
  }
  
  .url-download-dialog :deep(.el-form-item__label) {
    font-size: 13px;
  }
}

/* 种子下载对话框样式 */
.torrent-download-dialog :deep(.el-dialog) {
  border-radius: 8px;
}

.torrent-tabs {
  margin-bottom: 20px;
}

.torrent-upload {
  width: 100%;
}

.torrent-upload :deep(.el-upload) {
  width: 100%;
}

.torrent-upload :deep(.el-upload-dragger) {
  width: 100%;
  padding: 40px 20px;
}

.torrent-file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  padding: 12px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  font-size: 14px;
}

.torrent-file-info .el-icon {
  color: var(--el-color-primary);
}

.parse-section {
  margin-top: 20px;
}

.parse-result-section {
  margin-top: 20px;
}

.torrent-info {
  margin-bottom: 16px;
}

.torrent-info h4 {
  margin: 0 0 8px 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.torrent-meta {
  display: flex;
  gap: 8px;
  align-items: center;
}

.file-selection-section {
  margin-top: 16px;
}

.selection-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding: 8px 0;
}

.selected-count {
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.file-list-scrollbar {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
}

.torrent-file-name {
  font-size: 14px;
}

.file-path {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 移动端响应式 */
@media (max-width: 1024px) {
  .torrent-download-dialog :deep(.el-dialog) {
    width: 95% !important;
    margin: 0 auto;
  }

  .torrent-download-dialog :deep(.el-form-item__label) {
    font-size: 14px;
  }

  .file-list-scrollbar {
    height: 200px !important;
  }
}

@media (max-width: 480px) {
  .torrent-download-dialog :deep(.el-dialog) {
    width: 100% !important;
    margin: 0;
    border-radius: 0;
  }

  .torrent-download-dialog :deep(.el-form-item__label) {
    font-size: 13px;
  }

  .torrent-info h4 {
    font-size: 14px;
  }

  .file-list-scrollbar {
    height: 150px !important;
  }
}
</style>

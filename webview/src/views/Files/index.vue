<template>
  <div class="files-page">
    <!-- Breadcrumb with Glass effect -->
    <div class="breadcrumb-container glass-panel-sm">
      <el-breadcrumb separator="/">
        <el-breadcrumb-item 
          v-for="item in breadcrumbs" 
          :key="item.id"
          @click="navigateToPath(item.path)"
          class="breadcrumb-item"
        >
          {{ item.name }}
        </el-breadcrumb-item>
      </el-breadcrumb>
    </div>

    <!-- Toolbar with Glass effect -->
    <div class="toolbar-container glass-panel">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-tooltip content="上传文件" placement="bottom">
            <el-button type="primary" :icon="Upload" @click="handleUpload" class="action-btn">上传</el-button>
          </el-tooltip>
          <el-button :icon="FolderAdd" @click="handleNewFolder" class="action-btn-secondary">新建文件夹</el-button>
          <el-button :icon="FolderOpened" @click="handleMoveFile" :disabled="selectedFileIds.length === 0" class="action-btn-secondary">移动文件</el-button>
          <div class="divider-vertical"></div>
          <div class="view-switch glass-toggle">
            <el-button :icon="Grid" :class="{ 'is-active': viewMode === 'grid' }" @click="viewMode = 'grid'" text />
            <el-button :icon="List" :class="{ 'is-active': viewMode === 'list' }" @click="viewMode = 'list'" text />
          </div>
        </div>
        
        <div class="toolbar-right" v-if="selectedCount > 0">
          <span class="selection-info">已选 {{ selectedCount }} 项</span>
          <el-button :icon="Download" @click="handleToolbarDownload" plain circle />
          <el-button :icon="Share" @click="handleToolbarShare" plain circle />
          <el-button :icon="Delete" type="danger" @click="handleToolbarDelete" plain circle />
        </div>
      </div>
    </div>
    
    <!-- 网格视图 -->
    <div v-if="viewMode === 'grid'" class="file-grid">
      <!-- 文件夹 -->
      <div
        v-for="folder in fileListData.folders"
        :key="'folder-' + folder.id"
        class="file-card scale-up"
        :class="{ selected: isSelectedFolder(folder.id) }"
        @click="toggleSelectFolder(folder.id)"
        @dblclick="enterFolder(folder)"
      >
        <div class="file-icon">
          <el-icon :size="64" color="#409EFF">
            <Folder />
          </el-icon>
        </div>
        <div class="file-name">{{ folder.name }}</div>
        <div class="file-info">{{ formatDate(folder.created_time) }}</div>
      </div>

      <!-- 文件 -->
      <div
        v-for="file in fileListData.files"
        :key="'file-' + file.file_id"
        class="file-card scale-up"
        :class="{ selected: isSelectedFile(file.file_id) }"
        @click="toggleSelectFile(file.file_id)"
        @contextmenu.prevent="handleContextMenu($event, file)"
      >
        <div class="file-icon">
          <FileIcon
            :mime-type="file.mime_type"
            :file-name="file.file_name"
            :thumbnail-url="getThumbnailUrl(file.file_id)"
            :show-thumbnail="file.has_thumbnail"
            :icon-size="56"
            :is-encrypted="file.is_enc"
          />
        </div>
        <div class="file-name">{{ file.file_name }}</div>
        <div class="file-info">
          {{ formatSize(file.file_size) }} · {{ formatDate(file.created_at) }}
          <el-tag v-if="file.is_enc" size="small" type="warning" class="enc-tag">
            <el-icon><Lock /></el-icon>
          </el-tag>
        </div>
      </div>
    </div>
    
    <!-- 列表视图 -->
    <el-table
      v-else
      :data="[...fileListData.folders.map(f => ({ ...f, isFolder: true })), ...fileListData.files.map(f => ({ ...f, isFolder: false }))]"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="55" />
      <el-table-column label="名称" min-width="300">
        <template #default="{ row }">
          <div class="file-name-cell" @dblclick="row.isFolder ? navigateToPath(row.path) : handleFilePreview(row)">
            <!-- 文件夹图标 -->
            <el-icon v-if="row.isFolder" :size="32" color="#409EFF">
              <Folder />
            </el-icon>
            <!-- 文件图标 -->
            <div v-else class="list-file-icon">
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
            <span>{{ row.isFolder ? row.name : row.file_name }}</span>
            <el-tag v-if="!row.isFolder && row.is_enc" size="small" type="warning" class="enc-tag-inline">
              <el-icon :size="12"><Lock /></el-icon>
              加密
            </el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="大小" width="120">
        <template #default="{ row }">
          {{ row.isFolder ? '-' : formatSize(row.file_size) }}
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.isFolder ? row.created_time : row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <template v-if="!row.isFolder">
            <el-button link :icon="Download" @click.stop="handleDownloadFile(row)">下载</el-button>
            <el-button link :icon="Share" @click.stop="handleShareFile(row)">分享</el-button>
            <el-button link :icon="Delete" type="danger" @click.stop="handleDeleteFile(row)">删除</el-button>
          </template>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 空状态 -->
    <el-empty v-if="fileListData.folders.length === 0 && fileListData.files.length === 0" description="暂无文件" />
    
    <!-- 分页 -->
    <el-pagination
      v-if="fileListData.total > 0"
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :page-sizes="[20, 50, 100]"
      :total="fileListData.total"
      layout="total, sizes, prev, pager, next, jumper"
      @size-change="handleSizeChange"
      @current-change="handlePageChange"
      class="pagination"
    />
    
    <!-- 新建文件夹对话框 -->
    <el-dialog 
      v-model="showNewFolderDialog" 
      title="新建文件夹" 
      width="500px"
      @close="handleDialogClose"
    >
      <el-form 
        ref="folderFormRef" 
        :model="folderForm" 
        :rules="folderRules"
        label-width="100px"
      >
        <el-form-item label="文件夹名称" prop="dir_path">
          <el-input 
            v-model="folderForm.dir_path" 
            placeholder="请输入文件夹名称"
            clearable
            maxlength="50"
            show-word-limit
            @keyup.enter="handleCreateFolder"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showNewFolderDialog = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreateFolder">确定</el-button>
      </template>
    </el-dialog>
    
    <!-- 移动文件对话框 -->
    <el-dialog 
      v-model="showMoveDialog" 
      title="移动文件" 
      width="500px"
    >
      <el-form label-width="100px">
        <el-form-item label="选中文件">
          <el-tag v-for="fileId in selectedFileIds" :key="fileId" class="file-tag">
            {{ getFileName(fileId) }}
          </el-tag>
        </el-form-item>
        <el-form-item label="目标目录">
          <el-tree-select
            v-model="targetFolderId"
            :data="folderTreeData"
            :render-after-expand="false"
            placeholder="请选择目标目录"
            :default-expanded-keys="[currentPath]"
            :loading="loadingTree"
            style="width: 100%"
            check-strictly
            :props="{ label: 'label', value: 'value', children: 'children' }"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showMoveDialog = false">取消</el-button>
        <el-button type="primary" :loading="moving" @click="handleConfirmMove">确定移动</el-button>
      </template>
    </el-dialog>
    
    <!-- 分享文件对话框 -->
    <el-dialog 
      v-model="showShareDialog" 
      title="分享文件" 
      width="500px"
    >
      <el-form label-width="100px">
        <el-form-item label="文件名称">
          <el-input v-model="shareForm.file_name" disabled />
        </el-form-item>
        <el-form-item label="有效期">
          <el-radio-group v-model="shareForm.expire_days">
            <el-radio-button 
              v-for="option in expireOptions" 
              :key="option.value" 
              :label="option.value"
            >
              {{ option.label }}
            </el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="访问密码">
          <el-input 
            v-model="shareForm.password" 
            placeholder="请输入访问密码"
            maxlength="20"
            show-word-limit
          >
            <template #append>
              <el-button @click="shareForm.password = generateRandomPassword()">随机生成</el-button>
            </template>
          </el-input>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showShareDialog = false">取消</el-button>
        <el-button type="primary" :loading="sharing" @click="handleConfirmShare">确定分享</el-button>
      </template>
    </el-dialog>
    
    <!-- 下载密码对话框 -->
    <el-dialog 
      v-model="showDownloadPasswordDialog" 
      title="输入文件密码" 
      width="450px"
    >
      <el-form label-width="100px">
        <el-form-item label="文件名称">
          <el-text>{{ downloadPasswordForm.file_name }}</el-text>
        </el-form-item>
        <el-form-item label="文件密码">
          <el-input 
            v-model="downloadPasswordForm.file_password" 
            type="password"
            placeholder="请输入文件加密密码"
            show-password
            @keyup.enter="confirmDownloadPassword"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showDownloadPasswordDialog = false">取消</el-button>
        <el-button type="primary" :loading="downloadingFile" @click="confirmDownloadPassword">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, reactive } from 'vue'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import {
  Grid,
  List,
  Download,
  Share,
  Delete,
  Folder,
  FolderAdd,
  FolderOpened,
  Lock,
  Upload
} from '@element-plus/icons-vue'
import { getFileList, getThumbnail, moveFile, getVirtualPathTree, uploadFile, uploadPrecheck } from '@/api/file'
import { createShare } from '@/api/share'
import { createFolder } from '@/api/folder'
import { createLocalFileDownload, getDownloadTaskList, getLocalFileDownloadUrl } from '@/api/download'
import FileIcon from '@/components/FileIcon.vue'
import type { FileListResponse, FolderItem, FileItem } from '@/types'
import { useRoute, useRouter } from 'vue-router'
import { calculateFileMD5 } from '@/utils/crypto'

const viewMode = ref<'grid' | 'list'>('grid')
const selectedFolderIds = ref<number[]>([])
const selectedFileIds = ref<string[]>([])
const currentPage = ref(1)
const pageSize = ref(20)
const currentPath = ref<string>('') // 虚拟路径ID，初始为空字符串，从后端获取
const showNewFolderDialog = ref(false)
const creating = ref(false)
const folderFormRef = ref<FormInstance>()

// 移动文件对话框
const showMoveDialog = ref(false)
const moving = ref(false)
const targetFolderId = ref<string>('')
const folderTreeData = ref<any[]>([])
const loadingTree = ref(false)

// 分享对话框
const showShareDialog = ref(false)
const sharing = ref(false)
const shareForm = reactive({
  file_id: '',
  file_name: '',
  expire_days: 7,
  password: ''
})

// 下载密码对话框
const showDownloadPasswordDialog = ref(false)
const downloadPasswordForm = reactive({
  file_id: '',
  file_name: '',
  file_password: ''
})
const downloadingFile = ref(false)

// 预设有效期选项
const expireOptions = [
  { label: '1天', value: 1 },
  { label: '3天', value: 3 },
  { label: '7天', value: 7 },
  { label: '30天', value: 30 }
]

// 文件夹表单
const folderForm = reactive({
  dir_path: ''
})

// 表单验证规则
const folderRules: FormRules = {
  dir_path: [
    { required: true, message: '请输入文件夹名称', trigger: 'blur' },
    { min: 1, max: 50, message: '文件夹名称长度在1-50个字符', trigger: 'blur' },
    { 
      pattern: /^[^\\/:*?"<>|]+$/, 
      message: '文件夹名称不能包含特殊字符 \\ / : * ? " < > |', 
      trigger: 'blur' 
    }
  ]
}

// 文件列表数据
const fileListData = ref<FileListResponse>({
  breadcrumbs: [],
  current_path: '0',
  folders: [],
  files: [],
  total: 0,
  page: 1,
  page_size: 20
})

// 缩略图URL缓存（避免重复请求）
const thumbnailCache = ref<Map<string, string>>(new Map())

// 面包屑数据（后端已经返回面包屑，直接使用）
const breadcrumbs = computed(() => fileListData.value.breadcrumbs)

// 所有选中项数量
const selectedCount = computed(() => selectedFolderIds.value.length + selectedFileIds.value.length)

// 加载文件列表
const loadFileList = async () => {
  try {
    const res = await getFileList({
      virtualPath: currentPath.value,
      page: currentPage.value,
      pageSize: pageSize.value
    })
    
    if (res.code === 200) {
      fileListData.value = res.data
      
      // 更新当前路径ID（从后端返回的真实路径ID）
      if (res.data.current_path) {
        currentPath.value = res.data.current_path
      }
      
      // 异步加载所有文件的缩略图
      res.data.files.forEach(async (file: any) => {
        if (file.has_thumbnail && !thumbnailCache.value.has(file.file_id)) {
          const blobUrl = await getThumbnail(file.file_id)
          if (blobUrl) {
            thumbnailCache.value.set(file.file_id, blobUrl)
          }
        }
      })
    } else {
      ElMessage.error(res.message || '加载失败')
    }
  } catch (error) {
    ElMessage.error('加载文件列表失败')
    console.error(error)
  }
}

let router = useRouter();
let route = useRoute();
// 导航到指定路径
const navigateToPath = (path: string) => {
  currentPath.value = path
  currentPage.value = 1
  selectedFolderIds.value = []
  selectedFileIds.value = []
 router.push({
  path: route.path,
  query: {
    virtualPath: path
  }
 })
  loadFileList()
}

// 获取缩略图URL
const getThumbnailUrl = (fileId: string) => {
  return thumbnailCache.value.get(fileId) || ''
}

// 图片加载失败处理
const handleImageError = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.style.display = 'none'
}

// 格式化文件大小
const formatSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

// 格式化日期
const formatDate = (dateStr: string): string => {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

// 文件夹选择
const isSelectedFolder = (id: number) => selectedFolderIds.value.includes(id)
const toggleSelectFolder = (id: number) => {
  const index = selectedFolderIds.value.indexOf(id)
  if (index > -1) {
    selectedFolderIds.value.splice(index, 1)
  } else {
    selectedFolderIds.value.push(id)
  }
}

// 文件选择
const isSelectedFile = (id: string) => selectedFileIds.value.includes(id)
const toggleSelectFile = (id: string) => {
  const index = selectedFileIds.value.indexOf(id)
  if (index > -1) {
    selectedFileIds.value.splice(index, 1)
  } else {
    selectedFileIds.value.push(id)
  }
}

// 列表视图选择变化
const handleSelectionChange = (selection: any[]) => {
  selectedFolderIds.value = selection.filter(s => s.isFolder).map(s => s.id)
  selectedFileIds.value = selection.filter(s => !s.isFolder).map(s => s.file_id)
}

// 文件预览
const handleFilePreview = (file: FileItem) => {
  ElMessage.info(`预览文件: ${file.file_name}`)
}

// 进入文件夹
const enterFolder = (folder: any) => {
  if (folder.path) {
    navigateToPath(folder.path)
  }
}

// 新建文件夹
const handleNewFolder = () => {
  showNewFolderDialog.value = true
  folderForm.dir_path = ''
}

const handleUpload = async () => {
  // 创建文件选择输入框
  const input = document.createElement('input')
  input.type = 'file'
  input.multiple = true // 允许选择多个文件
  
  // 监听文件选择事件
  input.onchange = async (e) => {
    const target = e.target as HTMLInputElement
    const files = target.files
    
    if (!files || files.length === 0) {
      return
    }
    
    // 显示上传提示
    ElMessage.info(`开始上传 ${files.length} 个文件`)
    
    // 遍历处理每个文件
    for (let i = 0; i < files.length; i++) {
      const file = files[i]
      
      try {
        // 1. 计算文件MD5
        const fileMD5 = await calculateFileMD5(file)
        
        // 2. 准备上传预检参数
        const precheckParams = {
          chunk_signature: fileMD5, // 使用文件MD5作为分片签名
          file_name: file.name,
          file_size: file.size,
          files_md5: [fileMD5,fileMD5,fileMD5],
          path_id: currentPath.value
        }
        
        // 3. 调用上传预检接口
        const precheckResponse = await uploadPrecheck(precheckParams)
        console.log('上传预检接口响应:', precheckResponse)
        //秒传成功代码
        if (precheckResponse.code === 200) {
          ElMessage.success(`文件 ${file.name} 秒传成功`)
          loadFileList()
          continue
        }
        
        // 4. 处理接口返回结果
        if (precheckResponse.code === 201) {
          ElMessage.success(`文件 ${file.name} 预检成功`)
          console.log('上传预检结果:', precheckResponse.data)
          
          // 5. 调用上传API（不分片上传）
          const uploadParams = {
            precheck_id: precheckResponse.data,
            file: file,
            chunk_index: 0, // 小文件只有一个分片，索引为0
            total_chunks: 1, // 小文件总分片数为1
            chunk_md5: fileMD5, // 使用文件MD5作为分片MD5
            is_enc: false, // 默认不加密，可根据用户设置调整
          }
          try {
            const uploadResponse = await uploadFile(uploadParams)
            console.log('上传接口响应:', uploadResponse)
            
            // 6. 处理上传结果
            if (uploadResponse.code === 200) {
              ElMessage.success(`文件 ${file.name} 上传成功`)
              loadFileList()
            } else {
              ElMessage.error(`文件 ${file.name} 上传失败: ${uploadResponse.message}`)
            }
          } catch (error: any) {
            console.error(`上传文件 ${file.name} 时出错:`, error)
            ElMessage.error(`上传文件 ${file.name} 时出错: ${error.message}`)
          } finally {
            loadFileList()
          }
        } else {
          ElMessage.error(`文件 ${file.name} 预检失败: ${precheckResponse.message}`)
        }
      } catch (error: any) {
        console.error(`处理文件 ${file.name} 时出错:`, error)
        ElMessage.error(`处理文件 ${file.name} 时出错: ${error.message}`)
      }
    }
  }
  
  // 触发文件选择对话框
  input.click()
}


// 关闭对话框
const handleDialogClose = () => {
  folderFormRef.value?.resetFields()
}

// 创建文件夹
const handleCreateFolder = async () => {
  if (!folderFormRef.value) return
  
  await folderFormRef.value.validate(async (valid) => {
    if (valid) {
      creating.value = true
      try {
        // 调用创建文件夹API
        const res = await createFolder({
          parent_level: currentPath.value,
          dir_path: folderForm.dir_path
        })
        
        if (res.code === 200) {
          ElMessage.success('文件夹创建成功')
          showNewFolderDialog.value = false
          folderForm.dir_path = ''
          // 刷新文件列表
          loadFileList()
        } else {
          ElMessage.error(res.message || '创建文件夹失败')
        }
      } catch (error: any) {
        ElMessage.error(error.message || '创建文件夹失败')
      } finally {
        creating.value = false
      }
    }
  })
}

// 工具栏批量操作
const handleToolbarDownload = async () => {
  if (selectedFileIds.value.length === 0) {
    ElMessage.warning('请先选择要下载的文件')
    return
  }
  
  // 如果只选择了一个文件，调用单文件下载逻辑
  if (selectedFileIds.value.length === 1) {
    const fileId = selectedFileIds.value[0]
    const file = fileListData.value.files.find(f => f.file_id === fileId)
    if (file) {
      await handleDownloadFile(file)
    }
  } else {
    // 多个文件下载
    ElMessage.info('批量下载功能开发中')
  }
}

const handleToolbarShare = () => {
  if (selectedFileIds.value.length === 0) {
    ElMessage.warning('请先选择要分享的文件')
    return
  }
  if (selectedFileIds.value.length > 1) {
    ElMessage.warning('一次只能分享一个文件')
    return
  }
  
  const fileId = selectedFileIds.value[0]
  const file = fileListData.value.files.find(f => f.file_id === fileId)
  if (!file) {
    ElMessage.error('文件不存在')
    return
  }
  
  // 打开分享对话框
  handleShareFile(file)
}

const handleToolbarDelete = async () => {
  const totalCount = selectedFileIds.value.length + selectedFolderIds.value.length
  
  if (totalCount === 0) {
    ElMessage.warning('请先选择要删除的文件')
    return
  }
  
  ElMessageBox.confirm(
    `确定要删除 ${totalCount} 个文件吗？删除后将移动到回收站。`,
    '提示',
    {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    }
  ).then(async () => {
    try {
      // 删除文件
      if (selectedFileIds.value.length > 0) {
        const { deleteFiles } = await import('@/api/file')
        const result = await deleteFiles({ file_ids: selectedFileIds.value })
        if (result.code === 200) {
          ElMessage.success(result.message || '删除成功')
        } else {
          ElMessage.error(result.message || '删除失败')
        }
      }
      
      // TODO: 处理文件夹删除
      if (selectedFolderIds.value.length > 0) {
        ElMessage.warning('文件夹删除功能待开发')
      }
      
      // 清空选中并刷新列表
      selectedFileIds.value = []
      selectedFolderIds.value = []
      await loadFileList()
    } catch (error: any) {
      ElMessage.error(error.message || '删除失败')
    }
  }).catch(() => {
    // 用户取消
  })
}

// 单个文件操作
const handleDownloadFile = async (file: FileItem) => {
  // 检查是否加密文件
  if (file.is_enc) {
    // 加密文件，弹窗输入密码
    downloadPasswordForm.file_id = file.file_id
    downloadPasswordForm.file_name = file.file_name
    downloadPasswordForm.file_password = ''
    showDownloadPasswordDialog.value = true
  } else {
    // 非加密文件，直接下载
    await executeDownload(file.file_id, '')
  }
}

// 执行下载
const executeDownload = async (fileId: string, password: string) => {
  try {
    downloadingFile.value = true
    const res = await createLocalFileDownload({
      file_id: fileId,
      file_password: password
    })
    
    if (res.code === 200) {
      const taskId = res.data?.task_id
      if (!taskId) {
        ElMessage.error('任务创建失败')
        return
      }
      
      ElMessage.success('准备下载中，请稍候...')
      showDownloadPasswordDialog.value = false
      
      // 轮询任务状态，等待准备完成
      let retryCount = 0
      const maxRetries = 30 // 最多轮询30次，每次1秒，共30秒
      
      const checkTaskStatus = async () => {
        try {
          const taskRes = await getDownloadTaskList({ page: 1, pageSize: 100, state: -1 })
          if (taskRes.code === 200 && taskRes.data) {
            const task = taskRes.data.tasks?.find((t: any) => t.id === taskId)
            
            if (task) {
              if (task.state === 3) {
                // 准备完成，触发下载
                const downloadUrl = getLocalFileDownloadUrl(taskId)
                const link = document.createElement('a')
                link.href = downloadUrl
                link.download = task.file_name || 'download'
                link.style.display = 'none'
                document.body.appendChild(link)
                link.click()
                document.body.removeChild(link)
                
                ElMessage.success('下载已开始')
                downloadingFile.value = false
                return
              } else if (task.state === 4) {
                // 失败
                ElMessage.error(task.error_msg || '下载准备失败')
                downloadingFile.value = false
                return
              }
            }
            
            // 继续轮询
            retryCount++
            if (retryCount < maxRetries) {
              setTimeout(checkTaskStatus, 1000)
            } else {
              ElMessage.warning('准备超时，请到任务中心查看')
              downloadingFile.value = false
            }
          }
        } catch (error: any) {
          console.error('查询任务状态失败:', error)
          retryCount++
          if (retryCount < maxRetries) {
            setTimeout(checkTaskStatus, 1000)
          } else {
            downloadingFile.value = false
          }
        }
      }
      
      // 开始轮询
      setTimeout(checkTaskStatus, 1000)
    } else {
      ElMessage.error(res.msg || '创建下载任务失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '创建下载任务失败')
  } finally {
    // 注意：轮询过程中不关闭 loading，由轮询逼辑关闭
    if (!downloadingFile.value) {
      downloadingFile.value = false
    }
  }
}

// 确认密码并下载
const confirmDownloadPassword = async () => {
  if (!downloadPasswordForm.file_password) {
    ElMessage.warning('请输入文件密码')
    return
  }
  await executeDownload(downloadPasswordForm.file_id, downloadPasswordForm.file_password)
}

const handleShareFile = (file: FileItem) => {
  shareForm.file_id = file.file_id
  shareForm.file_name = file.file_name
  shareForm.expire_days = 7
  shareForm.password = generateRandomPassword()
  showShareDialog.value = true
}

const handleDeleteFile = async (file: FileItem) => {
  ElMessageBox.confirm(
    `确定要删除 "${file.file_name}" 吗？删除后将移动到回收站。`,
    '提示',
    {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    }
  ).then(async () => {
    try {
      const { deleteFiles } = await import('@/api/file')
      const result = await deleteFiles({ file_ids: [file.file_id] })
      if (result.code === 200) {
        ElMessage.success('删除成功')
        // 清空选中状态
        selectedFileIds.value = []
        selectedFolderIds.value = []
        await loadFileList()
      } else {
        ElMessage.error(result.message || '删除失败')
      }
    } catch (error: any) {
      ElMessage.error(error.message || '删除失败')
    }
  }).catch(() => {
    // 用户取消
  })
}

// 分页处理
const handlePageChange = (page: number) => {
  currentPage.value = page
  loadFileList()
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  loadFileList()
}

// 生成随机密码
const generateRandomPassword = () => {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789'
  let password = ''
  for (let i = 0; i < 6; i++) {
    password += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return password
}

// 复制到剪贴板
const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制到剪贴板')
  } catch (err) {
    ElMessage.error('复制失败')
  }
}

// 确认分享
const handleConfirmShare = async () => {
  if (!shareForm.password) {
    ElMessage.warning('请设置访问密码')
    return
  }
  
  sharing.value = true
  try {
    // 计算过期时间
    const expireDate = new Date()
    expireDate.setDate(expireDate.getDate() + shareForm.expire_days)
    const expireStr = expireDate.toISOString().slice(0, 19).replace('T', ' ')
    
    const res = await createShare({
      file_id: shareForm.file_id,
      expire: expireStr,
      password: shareForm.password
    })
    
    if (res.code === 200) {
      // 后端返回的 token，构建分享链接（不包含密码，用户在页面输入）
      const token = res.data.split('/').pop()
      const shareUrl = `${window.location.origin}/api/share/download?token=${token}`
      
      ElMessageBox.alert(
        `<div style="line-height: 2;">
          <p><strong>分享链接：</strong></p>
          <p style="word-break: break-all; background: #f5f7fa; padding: 8px; border-radius: 4px;">${shareUrl}</p>
          <p style="color: #e6a23c; margin-top: 8px;"><strong>访问密码：</strong>${shareForm.password}</p>
          <p style="color: #909399; font-size: 13px;">提示：访问链接时需要输入此密码</p>
          <p><strong>有效期：</strong>${shareForm.expire_days}天</p>
        </div>`,
        '分享成功',
        {
          confirmButtonText: '复制链接',
          dangerouslyUseHTMLString: true,
          callback: () => {
            copyToClipboard(shareUrl)
          }
        }
      )
      
      showShareDialog.value = false
      selectedFileIds.value = []
    } else {
      ElMessage.error(res.message || '分享失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '分享失败')
  } finally {
    sharing.value = false
  }
}

// 构建文件夹树结构（从后端获取完整的路径数据）
const buildFolderTree = async () => {
  loadingTree.value = true
  try {
    const res = await getVirtualPathTree()
    
    if (res.code !== 200 || !res.data) {
      ElMessage.error('获取目录树失败')
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
      pathMap.set(nodeId, {
        value: nodeId,
        label: vp.path.replace(/^\//, '') || '根目录', // 去除前缀斜杠
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
    ElMessage.error(error.message || '获取目录树失败')
  } finally {
    loadingTree.value = false
  }
}

// 获取文件名称
const getFileName = (fileId: string): string => {
  const file = fileListData.value.files.find(f => f.file_id === fileId)
  return file ? file.file_name : ''
}

// 打开移动文件对话框
const handleMoveFile = async () => {
  if (selectedFileIds.value.length === 0) {
    ElMessage.warning('请先选择要移动的文件')
    return
  }
  
  showMoveDialog.value = true
  targetFolderId.value = ''
  
  // 构建文件夹树（异步加载）
  await buildFolderTree()
}

// 确认移动
const handleConfirmMove = async () => {
  if (!targetFolderId.value) {
    ElMessage.warning('请选择目标目录')
    return
  }
  
  if (targetFolderId.value === currentPath.value) {
    ElMessage.warning('目标目录与当前目录相同')
    return
  }
  
  moving.value = true
  try {
    // 逐个移动文件
    for (const fileId of selectedFileIds.value) {
      const res = await moveFile({
        file_id: fileId,
        source_path: currentPath.value,
        target_path: targetFolderId.value
      })
      
      if (res.code !== 200) {
        ElMessage.error(`移动文件失败: ${res.message}`)
        return
      }
    }
    
    ElMessage.success(`成功移动 ${selectedFileIds.value.length} 个文件`)
    showMoveDialog.value = false
    selectedFileIds.value = []
    
    // 刷新文件列表
    loadFileList()
  } catch (error: any) {
    ElMessage.error(error.message || '移动文件失败')
  } finally {
    moving.value = false
  }
}

// 初始化
onMounted(() => {
  loadFileList()
})
</script>

<style scoped>
.files-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.breadcrumb-container {
  margin-bottom: 20px;
  padding: 12px 20px;
  border-radius: 12px;
  transition: all 0.3s;
}

.breadcrumb-item {
  font-size: 14px;
  font-weight: 500;
}

.toolbar-container {
  padding: 16px;
  border-radius: 16px;
  margin-bottom: 20px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.selection-info {
  margin-right: 16px;
  font-size: 14px;
  color: var(--text-secondary);
  font-weight: 500;
}

.action-btn {
  height: 40px;
  padding: 0 24px;
  border-radius: 10px;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.25);
}

.action-btn-secondary {
  height: 40px;
  border-radius: 10px;
  border: 1px solid transparent;
  background: white;
  color: var(--text-regular);
}

.action-btn-secondary:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
  background: white;
}

.divider-vertical {
  width: 1px;
  height: 24px;
  background: var(--border-light);
  margin: 0 16px;
}

.glass-toggle {
  background: rgba(0,0,0,0.03);
  padding: 4px;
  border-radius: 8px;
  display: flex;
  gap: 2px;
}

.glass-toggle .el-button {
  border-radius: 6px;
  padding: 8px;
  height: 32px;
  width: 32px;
  margin: 0;
  color: var(--text-secondary);
}

.glass-toggle .el-button.is-active {
  background: white;
  color: var(--primary-color);
  box-shadow: 0 2px 4px rgba(0,0,0,0.05);
}

.file-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 20px;
  padding: 4px;
}

.file-card {
  background: white;
  border-radius: 16px;
  padding: 12px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  border: 2px solid transparent;
  box-shadow: 0 2px 6px rgba(0,0,0,0.02);
  position: relative;
  overflow: hidden;
}

.file-card:hover {
  transform: translateY(-4px) scale(1.02);
  box-shadow: 0 12px 24px -8px rgba(0,0,0,0.08);
  z-index: 10;
}

.file-card.selected {
  border-color: var(--primary-color);
  background: rgba(37, 99, 235, 0.04);
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.1);
}

.file-icon {
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.3s;
}

.file-card:hover .file-icon {
  transform: scale(1.1);
}

.file-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  text-align: center;
  margin-top: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-info {
  font-size: 11px;
  color: var(--text-placeholder);
  text-align: center;
  margin-top: 4px;
}


.enc-tag {
  border: none;
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
  font-size: 11px;
  padding: 2px 6px;
  height: 18px;
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
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
}

.list-file-icon {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
}

.breadcrumb-card {
  flex-shrink: 0;
  margin-bottom: 16px;
}

.breadcrumb-item {
  cursor: pointer;
}

.breadcrumb-item:hover {
  color: var(--el-color-primary);
}

.pagination {
  margin-top: 16px;
  justify-content: center;
}

.file-tag {
  margin-right: 8px;
  margin-bottom: 8px;
}
</style>

<template>
  <div class="offline-page">
    <!-- 标题栏 -->
    <el-card shadow="never" class="header-card">
      <div class="page-header">
        <div class="header-left">
          <h2>{{ t('offline.title') }}</h2>
          <el-tag type="info">{{ t('offline.taskCount', { count: total }) }}</el-tag>
        </div>
        <div class="header-right">
          <el-button type="primary" icon="Plus" @click="showDownloadDialog = true">{{
            t('offline.newDownload')
          }}</el-button>
          <el-button icon="Refresh" @click="refreshTaskList">{{ t('common.refresh') }}</el-button>
        </div>
      </div>
    </el-card>

    <!-- 任务列表 -->
    <el-card shadow="never" class="task-list-card">
      <!-- PC端：表格布局 -->
      <el-table :data="taskList" v-loading="loading" class="offline-table desktop-table">
        <el-table-column :label="t('tasks.fileName')" min-width="300" class-name="mobile-name-column">
          <template #default="{ row }">
            <div class="file-name-cell">
              <el-icon :size="24" class="offline-icon"><Document /></el-icon>
              <div class="file-info">
                <file-name-tooltip
                  :file-name="row.file_name || t('offline.unknownFile')"
                  view-mode="table"
                  custom-class="file-name"
                />
                <div class="file-url mobile-hide" v-if="row.url">{{ truncateUrl(row.url) }}</div>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('tasks.status')" width="120" class-name="mobile-hide">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.state)">{{ row.state_text }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('tasks.progress')" width="200" class-name="mobile-progress-column">
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

        <el-table-column :label="t('offline.speed')" width="120" class-name="mobile-hide">
          <template #default="{ row }">
            <span v-if="row.state === 1">{{ formatSpeed(row.speed) }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('tasks.createTime')" width="180" class-name="mobile-hide">
          <template #default="{ row }">
            {{ formatDate(row.create_time) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('offline.errorInfo')" min-width="200" class-name="mobile-hide">
          <template #default="{ row }">
            <el-tooltip v-if="row.error_msg" :content="row.error_msg" placement="top">
              <span class="error-msg-text">{{ row.error_msg }}</span>
            </el-tooltip>
            <span v-else class="no-error-text">-</span>
          </template>
        </el-table-column>

        <el-table-column :label="t('tasks.operation')" width="200" fixed="right" class-name="mobile-actions-column">
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
                {{ t('tasks.pause') }}
              </el-button>
              <el-button
                v-if="row.state === 2"
                link
                icon="VideoPlay"
                type="primary"
                @click="resumeTask(row.id)"
                size="small"
              >
                {{ t('tasks.resume') }}
              </el-button>
              <el-button
                v-if="row.state === 0 || row.state === 1 || row.state === 2"
                link
                icon="Close"
                type="danger"
                @click="cancelTask(row.id)"
                size="small"
              >
                {{ t('tasks.cancel') }}
              </el-button>
              <el-button
                v-if="row.state === 3 || row.state === 4"
                link
                icon="Delete"
                type="danger"
                @click="deleteTask(row.id)"
                size="small"
              >
                {{ t('tasks.delete') }}
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 移动端：卡片布局 -->
      <div class="mobile-task-list" v-loading="loading">
        <div v-for="row in taskList" :key="row.id" class="mobile-task-item">
          <div class="task-item-header">
            <div class="task-item-info">
              <el-icon :size="24" class="task-icon offline-icon"><Document /></el-icon>
              <div class="task-name-wrapper">
                <file-name-tooltip
                  :file-name="row.file_name || row.url || t('offline.unknownFile')"
                  view-mode="list"
                  custom-class="task-name"
                />
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
              <el-button v-if="row.state === 1" link type="warning" @click.stop="pauseTask(row.id)" class="action-btn">
                <el-icon><VideoPause /></el-icon>
              </el-button>
              <el-button v-if="row.state === 2" link type="primary" @click.stop="resumeTask(row.id)" class="action-btn">
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

      <el-empty v-if="taskList.length === 0 && !loading" :description="t('offline.noDownloads')" />
    </el-card>

    <!-- 分页 -->
    <div v-if="total > 0" class="pagination-wrapper">
      <pagination
        :page="currentPage"
        :limit="pageSize"
        :total="total"
        :page-sizes="[20, 50, 100]"
        float="center"
        @pagination="handlePagination"
        class="pagination"
      />
    </div>

    <!-- 统一下载对话框 -->
    <el-dialog
      v-model="showDownloadDialog"
      :title="t('offline.newDownload')"
      :width="isMobile ? '95%' : '800px'"
      @open="handleDownloadDialogOpen"
      @close="handleDownloadDialogClose"
      :destroy-on-close="true"
      class="download-dialog"
    >
      <template v-if="showDownloadDialog">
        <!-- 输入区域：支持文本输入和文件上传 -->
        <div class="input-section">
          <el-tabs v-model="inputType" class="input-tabs">
            <el-tab-pane :label="t('offline.inputLink')" name="text">
              <el-form-item :label="t('offline.downloadLink')">
                <el-input
                  v-model="downloadForm.inputText"
                  :placeholder="t('offline.downloadLinkPlaceholder')"
                  type="textarea"
                  :rows="3"
                  @input="handleInputTextChange"
                />
                <div class="input-tip">
                  <el-icon><InfoFilled /></el-icon>
                  <span>{{ t('offline.downloadTip') }}</span>
                </div>
              </el-form-item>
            </el-tab-pane>
            <el-tab-pane :label="t('offline.uploadTorrent')" name="file">
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
                  {{ t('offline.dragTorrentHere') }}
                </div>
                <template #tip>
                  <div class="el-upload__tip">
                    {{ t('offline.torrentFileTip') }}
                  </div>
                </template>
              </el-upload>
              <div v-if="torrentFileName" class="torrent-file-info">
                <el-icon><Document /></el-icon>
                <span>{{ torrentFileName }}</span>
                <el-button link type="danger" @click="clearTorrentFile">{{ t('offline.clear') }}</el-button>
              </div>
            </el-tab-pane>
          </el-tabs>

          <!-- 输入类型提示 -->
          <div v-if="detectedInputType" class="detected-type-tip">
            <el-icon :class="detectedInputType === 'url' ? 'input-icon-success' : 'input-icon-primary'" :size="16">
              <component :is="detectedInputType === 'url' ? 'Check' : 'InfoFilled'" />
            </el-icon>
            <span v-if="detectedInputType === 'url'">{{ t('offline.detectedAsUrl') }}</span>
            <span v-else-if="detectedInputType === 'magnet'">{{ t('offline.detectedAsMagnet') }}</span>
            <span v-else-if="detectedInputType === 'torrent'">{{ t('offline.detectedAsTorrent') }}</span>
          </div>
        </div>

        <!-- URL 下载模式：直接显示表单 -->
        <el-form
          v-if="detectedInputType === 'url' && !torrentParseResult"
          :model="downloadForm"
          :rules="downloadRules"
          ref="downloadFormRef"
          label-width="100px"
          style="margin-top: 20px"
        >
          <el-form-item :label="t('offline.saveLocation')">
            <el-tree-select
              v-model="downloadForm.virtual_path"
              :data="folderTreeData"
              :render-after-expand="false"
              :placeholder="t('offline.selectSaveDirectory')"
              :loading="loadingTree"
              style="width: 100%"
              check-strictly
              :props="{ label: 'label', children: 'children' }"
              :default-expand-all="true"
              node-key="value"
            />
          </el-form-item>
          <el-form-item :label="t('offline.encryptStorage')">
            <el-switch v-model="downloadForm.enable_encryption" />
          </el-form-item>
          <el-form-item
            v-if="downloadForm.enable_encryption"
            :label="t('offline.encryptPassword')"
            prop="file_password"
          >
            <el-input
              v-model="downloadForm.file_password"
              type="password"
              :placeholder="t('offline.encryptPasswordPlaceholder')"
              show-password
              maxlength="32"
            />
            <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px">
              {{ t('offline.encryptPasswordTip') }}
            </div>
          </el-form-item>
        </el-form>

        <!-- 种子/磁力链接模式：解析按钮 -->
        <div
          v-if="(detectedInputType === 'magnet' || detectedInputType === 'torrent') && !torrentParseResult"
          class="parse-section"
        >
          <el-button
            type="primary"
            :loading="parsing"
            :disabled="!canParse"
            @click="handleParseTorrent"
            style="width: 100%"
          >
            {{ t('offline.parseTorrent') }}
          </el-button>
        </div>

        <!-- 解析结果：文件列表 -->
        <div v-if="torrentParseResult" class="parse-result-section">
          <div class="torrent-info">
            <h4>{{ torrentParseResult.name }}</h4>
            <div class="torrent-meta">
              <el-tag type="info">{{ t('offline.fileCount', { count: torrentParseResult.files.length }) }}</el-tag>
              <el-tag type="info">{{ formatSize(torrentParseResult.total_size) }}</el-tag>
            </div>
          </div>
          <el-divider />
          <div class="file-selection-section">
            <div class="selection-header">
              <div class="selection-left">
                <el-checkbox v-model="selectAllFiles" :indeterminate="isIndeterminate" @change="handleSelectAll">
                  {{ t('offline.allSelect') }}
                </el-checkbox>
                <span class="selected-count">{{
                  t('offline.selectedFiles', { count: selectedFileIndexes.length })
                }}</span>
              </div>
              <div class="file-type-filters">
                <el-button-group>
                  <el-button
                    :type="selectedFileType === 'all' ? 'primary' : 'default'"
                    size="small"
                    @click="handleFileTypeFilter('all')"
                  >
                    {{ t('offline.fileTypeAll') }}
                  </el-button>
                  <el-button
                    :type="selectedFileType === 'video' ? 'primary' : 'default'"
                    size="small"
                    @click="handleFileTypeFilter('video')"
                  >
                    {{ t('offline.fileTypeVideo') }}
                  </el-button>
                  <el-button
                    :type="selectedFileType === 'audio' ? 'primary' : 'default'"
                    size="small"
                    @click="handleFileTypeFilter('audio')"
                  >
                    {{ t('offline.fileTypeAudio') }}
                  </el-button>
                  <el-button
                    :type="selectedFileType === 'image' ? 'primary' : 'default'"
                    size="small"
                    @click="handleFileTypeFilter('image')"
                  >
                    {{ t('offline.fileTypeImage') }}
                  </el-button>
                  <el-button
                    :type="selectedFileType === 'doc' ? 'primary' : 'default'"
                    size="small"
                    @click="handleFileTypeFilter('doc')"
                  >
                    {{ t('offline.fileTypeDoc') }}
                  </el-button>
                  <el-button
                    :type="selectedFileType === 'archive' ? 'primary' : 'default'"
                    size="small"
                    @click="handleFileTypeFilter('archive')"
                  >
                    {{ t('offline.fileTypeArchive') }}
                  </el-button>
                </el-button-group>
              </div>
            </div>
            <el-scrollbar height="300px" class="file-list-scrollbar">
              <el-table
                ref="torrentFileTableRef"
                :data="filteredTorrentFiles"
                @selection-change="handleFileSelectionChange"
                :row-key="(row: any) => row.index"
              >
                <el-table-column type="selection" width="55" :reserve-selection="true" />
                <el-table-column :label="t('tasks.fileName')" min-width="250">
                  <template #default="{ row }">
                    <file-name-tooltip :file-name="row.name" view-mode="table" custom-class="torrent-file-name" />
                  </template>
                </el-table-column>
                <el-table-column :label="t('offline.fileType')" width="100">
                  <template #default="{ row }">
                    <el-tag :type="getFileTypeTagType(getFileTypeFromName(row.name))" size="small">
                      {{ getFileTypeText(getFileTypeFromName(row.name)) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column :label="t('tasks.fileSize')" width="110">
                  <template #default="{ row }">
                    {{ formatSize(row.size) }}
                  </template>
                </el-table-column>
              </el-table>
            </el-scrollbar>
          </div>
        </div>

        <!-- 种子下载配置表单（解析后显示） -->
        <el-form
          v-if="torrentParseResult"
          :model="downloadForm"
          :rules="downloadRules"
          ref="downloadFormRef"
          label-width="100px"
          style="margin-top: 20px"
        >
          <el-form-item :label="t('offline.saveLocation')">
            <el-tree-select
              v-model="downloadForm.virtual_path"
              :data="folderTreeData"
              :render-after-expand="false"
              :placeholder="t('offline.selectSaveDirectory')"
              :loading="loadingTree"
              style="width: 100%"
              check-strictly
              :props="{ label: 'label', children: 'children' }"
              :default-expand-all="true"
              node-key="value"
            />
          </el-form-item>
          <el-form-item :label="t('offline.encryptStorage')">
            <el-switch v-model="downloadForm.enable_encryption" />
          </el-form-item>
          <el-form-item
            v-if="downloadForm.enable_encryption"
            :label="t('offline.encryptPassword')"
            prop="file_password"
          >
            <el-input
              v-model="downloadForm.file_password"
              type="password"
              :placeholder="t('offline.encryptPasswordPlaceholder')"
              show-password
              maxlength="32"
            />
            <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px">
              {{ t('offline.encryptPasswordTip') }}
            </div>
          </el-form-item>
        </el-form>
      </template>

      <template #footer>
        <el-button @click="showDownloadDialog = false">{{ t('common.cancel') }}</el-button>
        <!-- URL 下载模式 -->
        <el-button
          v-if="detectedInputType === 'url' && !torrentParseResult"
          type="primary"
          :loading="creating"
          @click="handleCreateUrlDownload"
        >
          {{ t('offline.createTask') }}
        </el-button>
        <!-- 种子/磁力链接模式：解析按钮 -->
        <el-button
          v-else-if="(detectedInputType === 'magnet' || detectedInputType === 'torrent') && !torrentParseResult"
          type="primary"
          :loading="parsing"
          :disabled="!canParse"
          @click="handleParseTorrent"
        >
          {{ t('offline.parseTorrent') }}
        </el-button>
        <!-- 种子/磁力链接模式：开始下载按钮 -->
        <el-button
          v-else-if="torrentParseResult"
          type="primary"
          :loading="creatingTorrent"
          :disabled="selectedFileIndexes.length === 0"
          @click="handleStartTorrentDownload"
        >
          {{ t('offline.startDownload', { count: selectedFileIndexes.length }) }}
        </el-button>
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
  import { useResponsive, useI18n } from '@/composables'
  import { getFileTypeFromMimeType, getMimeTypeFromFileName, type FileTypeCategory } from '@/utils/file/mime'

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  // 使用响应式检测 composable
  const { isMobile } = useResponsive()

  const loading = ref(false)
  const creating = ref(false)
  const taskList = ref<OfflineDownloadTask[]>([])
  const showDownloadDialog = ref(false) // 统一的下载对话框
  let refreshTimer: number | null = null // 支持 setTimeout 和 setInterval
  const loadingTree = ref(false)
  const folderTreeData = ref<any[]>([])

  // 分页状态
  const currentPage = ref(1)
  const pageSize = ref(20)
  const total = ref(0)

  const downloadFormRef = ref<FormInstance>()
  const torrentUploadRef = ref()
  const torrentFileTableRef = ref()

  // 输入类型：文本输入或文件上传
  const inputType = ref<'text' | 'file'>('text')

  // 统一的下载表单
  const downloadForm = reactive({
    inputText: '', // 文本输入（URL 或磁力链接）
    virtual_path: '',
    enable_encryption: false,
    file_password: ''
  })

  // 种子下载相关状态
  const torrentFileName = ref('')
  const torrentFileContent = ref('') // Base64 编码的种子文件内容
  const parsing = ref(false)
  const creatingTorrent = ref(false)
  const torrentParseResult = ref<ParseTorrentResponse | null>(null)
  const selectedFileIndexes = ref<number[]>([])
  const selectAllFiles = ref(false)
  const isIndeterminate = ref(false)
  const selectedFileType = ref<'all' | 'video' | 'audio' | 'image' | 'doc' | 'archive' | 'other'>('all')

  // 检测到的输入类型
  const detectedInputType = ref<'url' | 'magnet' | 'torrent' | null>(null)

  // 统一的表单验证规则
  const downloadRules: FormRules = {
    inputText: [
      {
        validator: (_rule: any, value: any, callback: any) => {
          if (inputType.value === 'text' && !value?.trim()) {
            callback(new Error(t('offline.enterDownloadLink')))
          } else if (
            inputType.value === 'text' &&
            detectedInputType.value === 'url' &&
            !/^https?:\/\//.test(value?.trim())
          ) {
            callback(new Error(t('files.formatError')))
          } else {
            callback()
          }
        },
        trigger: 'blur'
      }
    ],
    file_password: [
      {
        validator: (_rule: any, value: any, callback: any) => {
          if (downloadForm.enable_encryption && !value) {
            callback(new Error(t('offline.encryptPasswordRequired')))
          } else if (value && value.length < 6) {
            callback(new Error(t('offline.passwordMinLength')))
          } else {
            callback()
          }
        },
        trigger: 'blur'
      }
    ]
  }

  // 输入类型识别函数
  const detectInputType = (input: string | File | null): 'url' | 'magnet' | 'torrent' | null => {
    if (!input) return null

    if (input instanceof File) {
      return input.name.toLowerCase().endsWith('.torrent') ? 'torrent' : null
    }

    const text = input.trim()
    if (text.startsWith('magnet:')) {
      return 'magnet'
    }
    if (text.startsWith('http://') || text.startsWith('https://')) {
      return 'url'
    }

    return null
  }

  // 监听输入变化，自动识别类型
  const handleInputTextChange = () => {
    if (inputType.value === 'text' && downloadForm.inputText) {
      detectedInputType.value = detectInputType(downloadForm.inputText)
    } else if (inputType.value === 'file') {
      detectedInputType.value = torrentFileContent.value ? 'torrent' : null
    } else {
      detectedInputType.value = null
    }
  }

  // 计算是否可以解析种子
  const canParse = computed(() => {
    if (inputType.value === 'file') {
      return !!torrentFileContent.value
    } else {
      return !!downloadForm.inputText && (detectedInputType.value === 'magnet' || detectedInputType.value === 'url')
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
      // 使用范围查询，查询 type < 7 的任务（即 type 0-6，离线下载任务）
      // 这样可以直接在后端过滤掉 type=7 的任务，总数更准确
      const res = await getDownloadTaskList({
        page: currentPage.value,
        pageSize: pageSize.value,
        state: -1, // 查询所有状态
        typeMax: 7 // 查询 type < 7 的任务（即 type 0-6，离线下载任务）
      })
      if (res.code === 200 && res.data) {
        // 更新任务列表（后端已经过滤了 type=7 的任务）
        taskList.value = (res.data.tasks || []).map((task: any) => ({ ...task }))

        // 更新总数（后端返回的是过滤后的准确总数）
        total.value = res.data.total || 0

        // 调试日志：检查数据更新（仅在开发环境）
        if (import.meta.env.DEV) {
          const downloadingTasks = taskList.value.filter((t: any) => t.state === 1)
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
        proxy?.$modal.msgError(error.message || t('offline.loadTaskListFailed'))
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
    // 重置到第一页
    currentPage.value = 1
    loadTaskList().then(() => {
      // 刷新后重新启动智能刷新
      startSmartRefresh()
    })
  }

  // 处理分页事件
  const handlePagination = ({ page, limit }: { page: number; limit: number }) => {
    currentPage.value = page
    pageSize.value = limit
    // 使用后端分页，直接加载对应页的数据
    loadTaskList()
  }

  // 构建文件夹树结构
  const buildFolderTree = async () => {
    loadingTree.value = true
    try {
      const res = await getVirtualPathTree()

      if (res.code !== 200 || !res.data) {
        proxy?.$modal.msgError(t('offline.getFolderTreeFailed'))
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
        const displayName = pathParts.length > 0 ? pathParts[pathParts.length - 1] : vp.path || t('offline.rootDir')

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
      proxy?.$modal.msgError(error.message || t('offline.getFolderTreeFailed'))
    } finally {
      loadingTree.value = false
    }
  }

  // 创建 URL 下载任务
  const handleCreateUrlDownload = async () => {
    if (!downloadFormRef.value) return

    await downloadFormRef.value.validate(async (valid: boolean) => {
      if (valid) {
        if (detectedInputType.value !== 'url') {
          proxy?.$modal.msgWarning(t('offline.enterValidUrl'))
          return
        }

        creating.value = true
        try {
          const res = await createOfflineDownload({
            url: downloadForm.inputText.trim(),
            virtual_path: downloadForm.virtual_path || undefined,
            enable_encryption: downloadForm.enable_encryption,
            file_password: downloadForm.enable_encryption ? downloadForm.file_password : undefined
          })

          if (res.code === 200) {
            proxy?.$modal.msgSuccess(t('offline.taskCreatedSuccess'))
            showDownloadDialog.value = false
            // 创建任务后重置到第一页并刷新
            currentPage.value = 1
            loadTaskList()
          }
        } catch (error: any) {
          proxy?.$modal.msgError(error.message || t('offline.taskCreatedFailed'))
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
      proxy?.$modal.msgSuccess(t('tasks.pauseSuccess'))
      loadTaskList()
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('tasks.pauseFailed'))
    }
  }

  // 恢复任务
  const resumeTask = async (taskId: string) => {
    try {
      await resumeDownload(taskId)
      proxy?.$modal.msgSuccess(t('tasks.resumeSuccess'))
      loadTaskList()
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('tasks.resumeFailed'))
    }
  }

  // 取消任务
  const cancelTask = async (taskId: string) => {
    try {
      await proxy?.$modal.confirm(t('offline.confirmCancelTask'))

      await cancelDownload(taskId)
      proxy?.$modal.msgSuccess(t('tasks.cancelSuccess'))
      loadTaskList()
    } catch (error: any) {
      if (error !== 'cancel') {
        proxy?.$modal.msgError(error.message || t('tasks.cancelFailed'))
      }
    }
  }

  // 删除任务
  const deleteTask = async (taskId: string) => {
    try {
      await proxy?.$modal.confirm(t('offline.confirmDeleteTask'))

      await deleteDownload(taskId)
      proxy?.$modal.msgSuccess(t('tasks.deleteSuccess'))
      loadTaskList()
    } catch (error: any) {
      if (error !== 'cancel') {
        proxy?.$modal.msgError(error.message || t('tasks.deleteFailed'))
      }
    }
  }

  // 使用 getTaskStatusType 作为 getStatusType 的别名
  const getStatusType = getTaskStatusType

  // 处理种子文件选择
  const handleTorrentFileChange = (file: any) => {
    const reader = new FileReader()
    reader.onload = e => {
      const result = e.target?.result as string
      // 移除 data URL 前缀（如 "data:application/x-bittorrent;base64,"）
      const base64Content = result.includes(',') ? result.split(',')[1] : result
      torrentFileContent.value = base64Content
      torrentFileName.value = file.name
      // 自动识别类型
      detectedInputType.value = detectInputType(file.raw)
    }
    reader.onerror = () => {
      proxy?.$modal.msgError(t('offline.readTorrentFailed'))
    }
    reader.readAsDataURL(file.raw)
  }

  // 清除种子文件
  const clearTorrentFile = () => {
    torrentFileContent.value = ''
    torrentFileName.value = ''
    detectedInputType.value = null
    if (torrentUploadRef.value) {
      torrentUploadRef.value.clearFiles()
    }
  }

  // 处理下载对话框打开
  const handleDownloadDialogOpen = () => {
    buildFolderTree()
    // 重置状态
    inputType.value = 'text'
    detectedInputType.value = null
  }

  // 处理下载对话框关闭
  const handleDownloadDialogClose = () => {
    // 重置所有状态
    inputType.value = 'text'
    downloadForm.inputText = ''
    downloadForm.virtual_path = ''
    downloadForm.enable_encryption = false
    downloadForm.file_password = ''
    torrentFileName.value = ''
    torrentFileContent.value = ''
    torrentParseResult.value = null
    selectedFileIndexes.value = []
    selectAllFiles.value = false
    isIndeterminate.value = false
    selectedFileType.value = 'all'
    detectedInputType.value = null
    if (torrentUploadRef.value) {
      torrentUploadRef.value.clearFiles()
    }
  }

  // 解析种子
  const handleParseTorrent = async () => {
    if (!canParse.value) {
      proxy?.$modal.msgWarning(t('offline.uploadTorrentFirst'))
      return
    }

    if (detectedInputType.value !== 'magnet' && detectedInputType.value !== 'torrent') {
      proxy?.$modal.msgWarning(t('offline.enterMagnetOrTorrent'))
      return
    }

    parsing.value = true
    try {
      const content = inputType.value === 'file' ? torrentFileContent.value : downloadForm.inputText.trim()

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
        proxy?.$modal.msgSuccess(t('offline.parseSuccess'))
      } else {
        proxy?.$modal.msgError(res.message || t('offline.parseFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('offline.parseFailed'))
    } finally {
      parsing.value = false
    }
  }

  // 根据文件名获取文件类型
  const getFileTypeFromName = (fileName: string): FileTypeCategory => {
    const mimeType = getMimeTypeFromFileName(fileName)
    return getFileTypeFromMimeType(mimeType)
  }

  // 获取文件类型标签的类型（用于 el-tag）
  const getFileTypeTagType = (fileType: FileTypeCategory): 'success' | 'warning' | 'info' | 'primary' | 'danger' | undefined => {
    const typeMap: Record<FileTypeCategory, 'success' | 'warning' | 'info' | 'primary' | 'danger' | undefined> = {
      image: 'success',
      video: 'warning',
      audio: 'info',
      doc: 'primary',
      archive: 'danger',
      other: undefined
    }
    return typeMap[fileType]
  }

  // 获取文件类型的显示文本
  const getFileTypeText = (fileType: FileTypeCategory): string => {
    const typeKey = fileType === 'other' ? 'fileTypeOther' : `fileType${fileType.charAt(0).toUpperCase() + fileType.slice(1)}`
    return t(`offline.${typeKey}`)
  }

  // 筛选后的文件列表
  const filteredTorrentFiles = computed(() => {
    if (!torrentParseResult.value) return []
    if (selectedFileType.value === 'all') {
      return torrentParseResult.value.files
    }
    return torrentParseResult.value.files.filter((file: TorrentFileInfo) => {
      const fileType = getFileTypeFromName(file.name)
      return fileType === selectedFileType.value
    })
  })

  // 处理文件类型筛选
  const handleFileTypeFilter = (type: 'all' | 'video' | 'audio' | 'image' | 'doc' | 'archive' | 'other') => {
    selectedFileType.value = type
    // 筛选后，更新全选状态
    nextTick(() => {
      if (torrentFileTableRef.value && torrentParseResult.value) {
        const filtered = filteredTorrentFiles.value
        const selectedInFiltered = filtered.filter((f: TorrentFileInfo) =>
          selectedFileIndexes.value.includes(f.index)
        )
        const total = filtered.length
        const selected = selectedInFiltered.length
        selectAllFiles.value = selected === total && total > 0
        isIndeterminate.value = selected > 0 && selected < total
      }
    })
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
    if (!downloadFormRef.value || !torrentParseResult.value) return

    if (selectedFileIndexes.value.length === 0) {
      proxy?.$modal.msgWarning(t('offline.selectFileFirst'))
      return
    }

    await downloadFormRef.value.validate(async (valid: boolean) => {
      if (valid) {
        creatingTorrent.value = true
        try {
          const content = inputType.value === 'file' ? torrentFileContent.value : downloadForm.inputText.trim()

          const res = await startTorrentDownload({
            content,
            file_indexes: selectedFileIndexes.value,
            virtual_path: downloadForm.virtual_path || undefined,
            enable_encryption: downloadForm.enable_encryption,
            file_password: downloadForm.enable_encryption ? downloadForm.file_password : undefined
          })

          if (res.code === 200 && res.data) {
            proxy?.$modal.msgSuccess(t('offline.taskCreatedWithCount', { count: res.data.task_count }))
            showDownloadDialog.value = false
            // 创建任务后重置到第一页并刷新
            currentPage.value = 1
            loadTaskList()
          } else {
            proxy?.$modal.msgError(res.message || t('offline.taskCreatedFailed'))
          }
        } catch (error: any) {
          proxy?.$modal.msgError(error.message || t('offline.taskCreatedFailed'))
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
      // 智能刷新时重新获取数据以更新任务状态和进度
      // 使用 loadTaskList，但不会显示 loading（因为 isManualRefresh = false）
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

  .header-right .el-button {
    transition: all 0.2s;
  }

  html.dark .header-right .el-button {
    background-color: var(--el-bg-color);
    border-color: var(--el-border-color);
    color: var(--el-text-color-primary);
  }

  html.dark .header-right .el-button:hover {
    background-color: var(--el-fill-color-light);
    border-color: var(--primary-color);
    color: var(--primary-color);
  }

  html.dark .header-right .el-button--primary {
    background-color: var(--primary-color);
    border-color: var(--primary-color);
    color: var(--el-text-color-primary);
  }

  html.dark .header-right .el-button--primary:hover {
    background-color: var(--primary-hover);
    border-color: var(--primary-hover);
  }

  .task-list-card {
    flex: 1;
    overflow: hidden;
  }

  .pagination-wrapper {
    flex-shrink: 0;
    padding-top: 16px;
    border-top: 1px solid var(--el-border-color-lighter);
  }

  .pagination {
    justify-content: center;
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

  .error-msg-text {
    color: var(--el-color-danger);
    font-size: 13px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: block;
    cursor: pointer;
  }

  .no-error-text {
    color: var(--el-text-color-placeholder);
    font-size: 13px;
  }

  /* PC端表格样式 */
  .desktop-table {
    display: table;
  }

  /* 隐藏表格自带的空状态显示，使用手动的 el-empty */
  .offline-table :deep(.el-table__empty-block) {
    display: none;
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

    .download-dialog :deep(.el-dialog) {
      width: 95% !important;
      margin: 0 auto;
    }

    .download-dialog :deep(.el-form-item__label) {
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

    .download-dialog :deep(.el-dialog) {
      width: 100% !important;
      margin: 0;
      border-radius: 0;
    }

    .download-dialog :deep(.el-form-item__label) {
      font-size: 13px;
    }
  }

  /* 统一下载对话框样式 */
  .download-dialog :deep(.el-dialog) {
    border-radius: 8px;
  }


  .input-section {
    margin-bottom: 20px;
  }

  .input-tabs {
    margin-bottom: 16px;
  }

  .input-tip {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-top: 8px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .input-tip .el-icon {
    font-size: 14px;
    color: var(--el-color-info);
  }

  .detected-type-tip {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    margin-top: 12px;
    background: var(--el-fill-color-light);
    border-radius: 4px;
    font-size: 13px;
    color: var(--el-text-color-primary);
  }

  .detected-type-tip .el-icon {
    font-size: 16px;
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
    flex-wrap: wrap;
    gap: 12px;
  }

  .selection-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .file-type-filters {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .file-type-filters .el-button-group {
    display: flex;
    flex-wrap: wrap;
  }

  @media (max-width: 768px) {
    .selection-header {
      flex-direction: column;
      align-items: flex-start;
    }

    .file-type-filters {
      width: 100%;
      justify-content: flex-start;
    }

    .file-type-filters .el-button-group {
      width: 100%;
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 4px;
    }

    .file-type-filters .el-button-group .el-button {
      width: 100%;
    }
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
    .download-dialog :deep(.el-dialog) {
      width: 95% !important;
      margin: 0 auto;
    }

    .download-dialog :deep(.el-form-item__label) {
      font-size: 14px;
    }

    .file-list-scrollbar {
      height: 200px !important;
    }
  }

  @media (max-width: 480px) {
    .download-dialog :deep(.el-dialog) {
      width: 100% !important;
      margin: 0;
      border-radius: 0;
    }

    .download-dialog :deep(.el-form-item__label) {
      font-size: 13px;
    }

    .torrent-info h4 {
      font-size: 14px;
    }

    .file-list-scrollbar {
      height: 150px !important;
    }
  }

  /* 深色模式样式 */
  html.dark .offline-page {
    background: var(--card-bg);
  }

  html.dark .header-card {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .header-card :deep(.el-card__body) {
    background: var(--card-bg);
  }

  html.dark .task-list-card {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .task-list-card :deep(.el-card__body) {
    background: var(--card-bg);
  }

  html.dark .offline-table {
    background: var(--card-bg);
  }

  html.dark .mobile-task-item {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .download-dialog :deep(.el-dialog) {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .download-dialog :deep(.el-dialog__header) {
    background: var(--card-bg);
    border-bottom-color: var(--el-border-color);
  }

  html.dark .download-dialog :deep(.el-dialog__title) {
    color: var(--el-text-color-primary);
  }

  html.dark .download-dialog :deep(.el-dialog__body) {
    background: var(--card-bg);
    color: var(--el-text-color-primary);
  }

  html.dark .download-dialog :deep(.el-form-item__label) {
    color: var(--el-text-color-primary);
  }

  html.dark .download-dialog :deep(.el-input__wrapper) {
    background-color: var(--el-bg-color);
    border-color: var(--el-border-color);
  }

  html.dark .download-dialog :deep(.el-input__inner) {
    color: var(--el-text-color-primary);
  }
</style>

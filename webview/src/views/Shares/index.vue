<template>
  <div class="shares-page">
    <!-- 头部卡片 -->
    <div class="header-card glass-panel">
      <div class="header">
        <div class="header-left">
          <div class="title-section">
            <el-icon :size="24" class="title-icon"><Share /></el-icon>
            <h2>{{ t('share.myShares') }}</h2>
            <span class="share-count">{{ t('share.shareCount', { count: shareList.length }) }}</span>
          </div>
          <div v-if="selectedShares.length > 0" class="batch-selection-info">
            <span class="selected-count">{{ t('share.selectedCount', { count: selectedShares.length }) }}</span>
            <el-button type="danger" icon="Delete" size="small" @click="handleBatchDelete" :loading="batchDeleting">
              {{ t('share.batchDelete') }}
            </el-button>
            <el-button link size="small" @click="clearSelection">
              {{ t('share.cancelSelect') }}
            </el-button>
          </div>
        </div>
        <el-button type="primary" icon="Refresh" @click="loadShareList" :loading="loading">{{
          t('common.refresh')
        }}</el-button>
      </div>
    </div>

    <!-- PC端：表格布局 -->
    <div class="table-card glass-panel">
      <el-table
        ref="tableRef"
        :data="shareList"
        v-loading="loading"
        class="shares-table desktop-table"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column :label="t('tasks.fileName')" min-width="250" class-name="mobile-name-column">
          <template #default="{ row }">
            <div class="file-name-cell">
              <el-icon :size="24" class="share-icon"><Document /></el-icon>
              <file-name-tooltip :file-name="row.file_name" view-mode="table" custom-class="file-name" />
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('share.shareLink')" min-width="400" class-name="mobile-link-column">
          <template #default="{ row }">
            <div class="link-cell">
              <el-input :model-value="getShareUrl(row.token)" readonly size="small" class="share-link-input">
                <template #append>
                  <el-button icon="CopyDocument" @click="copyShareLink(row)" :loading="copyingId === row.id">
                    {{ t('common.copy') }}
                  </el-button>
                </template>
              </el-input>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('share.sharePassword')" width="100" align="center" class-name="mobile-hide">
          <template #default="{ row }">
            <el-tooltip :content="row.password_hash ? t('share.hasPassword') : t('share.noPassword')" placement="top">
              <div
                class="status-badge"
                :class="{ 'has-password': row.password_hash, 'no-password': !row.password_hash }"
              >
                <el-icon :size="16"><Lock /></el-icon>
              </div>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column :label="t('share.downloadCount')" width="100" align="center" class-name="mobile-hide">
          <template #default="{ row }">
            <el-tooltip :content="t('share.downloadedTimes', { count: row.download_count || 0 })" placement="top">
              <div class="download-badge">
                <el-icon :size="14"><Download /></el-icon>
                <span class="download-count-text">{{ row.download_count || 0 }}</span>
              </div>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column :label="t('share.expireDate')" width="180" align="center" class-name="mobile-hide">
          <template #default="{ row }">
            <div class="time-cell">
              <el-icon :size="14"><Clock /></el-icon>
              <span :class="{ 'expired-text': isExpired(row.expires_at) }">
                {{ formatDate(row.expires_at) }}
              </span>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('share.createTime')" width="180" align="center" class-name="mobile-hide">
          <template #default="{ row }">
            <div class="time-cell">
              <el-icon :size="14"><Calendar /></el-icon>
              <span>{{ formatDate(row.created_at) }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column
          :label="t('tasks.operation')"
          width="200"
          fixed="right"
          align="center"
          class-name="mobile-actions-column"
        >
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button link type="primary" icon="Edit" @click="handleUpdatePassword(row)" size="small">
                {{ t('share.modifyPassword') }}
              </el-button>
              <el-button link type="danger" icon="Delete" @click="handleDelete(row)" size="small">
                {{ t('common.delete') }}
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 移动端：卡片布局 -->
      <div class="mobile-share-list" v-loading="loading">
        <div
          v-for="row in shareList"
          :key="row.id"
          class="mobile-share-item"
          :class="{ selected: isShareSelected(row.id) }"
        >
          <div class="share-item-header">
            <el-checkbox
              :model-value="isShareSelected(row.id)"
              @change="toggleShareSelection(row)"
              class="mobile-checkbox"
            />
            <div class="share-item-info">
              <el-icon :size="24" class="share-icon"><Document /></el-icon>
              <div class="share-name-wrapper">
                <file-name-tooltip :file-name="row.file_name" view-mode="list" custom-class="share-name" />
                <div class="share-meta">
                  <div
                    class="mobile-status-badge"
                    :class="{ 'has-password': row.password_hash, 'no-password': !row.password_hash }"
                  >
                    <el-icon :size="14"><Lock /></el-icon>
                    <span class="status-text">{{ row.password_hash ? t('share.password') : t('share.public') }}</span>
                  </div>
                  <div class="mobile-download-badge">
                    <el-icon :size="12"><Download /></el-icon>
                    <span class="download-text">{{ row.download_count || 0 }}</span>
                  </div>
                </div>
              </div>
            </div>
            <div class="share-actions">
              <el-button link type="primary" @click.stop="handleUpdatePassword(row)" class="action-btn">
                <el-icon><Edit /></el-icon>
              </el-button>
              <el-button link type="danger" @click.stop="handleDelete(row)" class="action-btn">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </div>

          <div class="share-link-wrapper">
            <el-input :model-value="getShareUrl(row.token)" readonly size="small" class="mobile-share-link-input">
              <template #append>
                <el-button icon="CopyDocument" @click="copyShareLink(row)" :loading="copyingId === row.id" size="small">
                  {{ t('common.copy') }}
                </el-button>
              </template>
            </el-input>
          </div>

          <div class="share-time-info">
            <div class="time-item">
              <el-icon :size="12"><Clock /></el-icon>
              <span :class="{ 'expired-text': isExpired(row.expires_at) }">
                {{ t('share.expire') }}：{{ formatDate(row.expires_at) }}
              </span>
            </div>
            <div class="time-item">
              <el-icon :size="12"><Calendar /></el-icon>
              <span>{{ t('share.create') }}：{{ formatDate(row.created_at) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 空状态显示 -->
      <el-empty v-if="shareList.length === 0 && !loading" :description="t('share.noShareRecords')" />
    </div>

    <!-- 修改密码对话框 -->
    <el-dialog
      v-model="showPasswordDialog"
      :title="t('share.updatePassword')"
      :width="isMobile ? '95%' : '450px'"
      :close-on-click-modal="false"
      class="password-dialog"
    >
      <el-form label-width="80px">
        <el-form-item :label="t('tasks.fileName')">
          <el-input v-model="currentShare.file_name" disabled />
        </el-form-item>
        <el-form-item :label="t('share.newPassword')">
          <el-input
            v-model="newPassword"
            :placeholder="t('share.updatePasswordPlaceholder')"
            maxlength="20"
            show-word-limit
            clearable
          >
            <template #append>
              <el-button @click="handleGenerateRandomPassword" size="small">{{ t('common.generate') }}</el-button>
            </template>
          </el-input>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showPasswordDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="updating" @click="handleConfirmUpdatePassword">{{
          t('share.confirmUpdate')
        }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import { useResponsive, useI18n } from '@/composables'
  import { getShareList, deleteShare, updateSharePassword } from '@/api/share'
  import type { ShareInfo } from '@/types'
  import { formatDate, getShareUrl, generateRandomPassword, copyToClipboard } from '@/utils'

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  // 使用响应式检测 composable
  const { isMobile } = useResponsive()

  const loading = ref(false)
  const shareList = ref<ShareInfo[]>([])
  const showPasswordDialog = ref(false)
  const updating = ref(false)
  const newPassword = ref('')
  const currentShare = reactive<Partial<ShareInfo>>({})
  const copyingId = ref<number | null>(null)
  const selectedShares = ref<ShareInfo[]>([])
  const batchDeleting = ref(false)
  const tableRef = ref()

  onMounted(() => {
    loadShareList()
  })

  // 检查是否过期
  const isExpired = (expiresAt: string): boolean => {
    return new Date(expiresAt) < new Date()
  }

  // 加载分享列表
  const loadShareList = async () => {
    loading.value = true
    try {
      const res = await getShareList()
      if (res.code === 200) {
        shareList.value = res.data || []
      } else {
        proxy?.$modal.msgError(res.message || t('common.loadFailed'))
      }
    } catch (error) {
      proxy?.$modal.msgError(t('share.loadShareListFailed'))
      proxy?.$log.error(error)
    } finally {
      loading.value = false
    }
  }

  // 复制分享链接
  const copyShareLink = async (share: ShareInfo) => {
    copyingId.value = share.id
    const shareUrl = getShareUrl(share.token)
    const success = await copyToClipboard(shareUrl)
    if (success) {
      proxy?.$modal.msgSuccess(t('common.copied'))
    } else {
      proxy?.$modal.msgError(t('common.copyFailed'))
    }
    setTimeout(() => {
      copyingId.value = null
    }, 500)
  }

  // 删除分享
  const handleDelete = async (share: ShareInfo) => {
    try {
      await proxy?.$modal.confirm(t('share.confirmDeleteShare'))
      const res = await deleteShare(share.id)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('common.deleteSuccess'))
        // 从选中列表中移除
        const index = selectedShares.value.findIndex(s => s.id === share.id)
        if (index > -1) {
          selectedShares.value.splice(index, 1)
        }
        loadShareList()
      } else {
        proxy?.$modal.msgError(res.message || t('common.deleteFailed'))
      }
    } catch (error: any) {
      if (error !== 'cancel') {
        proxy?.$modal.msgError(error.message || t('common.deleteFailed'))
      }
    }
  }

  // 打开修改密码对话框
  const handleUpdatePassword = (share: ShareInfo) => {
    Object.assign(currentShare, share)
    newPassword.value = ''
    showPasswordDialog.value = true
  }

  // 生成随机密码
  const handleGenerateRandomPassword = () => {
    newPassword.value = generateRandomPassword(6)
  }

  // 确认修改密码
  const handleConfirmUpdatePassword = async () => {
    updating.value = true
    try {
      const res = await updateSharePassword(currentShare.id!, newPassword.value || '')
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(
          newPassword.value ? t('share.updatePasswordSuccess') : t('share.cancelPasswordSuccess')
        )
        showPasswordDialog.value = false
        loadShareList()
      } else {
        proxy?.$modal.msgError(res.message || t('share.updatePasswordFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('share.updatePasswordFailed'))
    } finally {
      updating.value = false
    }
  }

  // 表格选择变化
  const handleSelectionChange = (selection: ShareInfo[]) => {
    selectedShares.value = selection
  }

  // 检查分享是否被选中（移动端）
  const isShareSelected = (shareId: number): boolean => {
    return selectedShares.value.some(share => share.id === shareId)
  }

  // 切换分享选择状态（移动端）
  const toggleShareSelection = (share: ShareInfo) => {
    const index = selectedShares.value.findIndex(s => s.id === share.id)
    if (index > -1) {
      selectedShares.value.splice(index, 1)
    } else {
      selectedShares.value.push(share)
    }
  }

  // 清空选择
  const clearSelection = () => {
    selectedShares.value = []
    // 清空表格多选框
    tableRef.value?.clearSelection()
  }

  // 批量删除
  const handleBatchDelete = async () => {
    if (selectedShares.value.length === 0) {
      proxy?.$modal.msgWarning(t('files.selectDeleteFilesFirst'))
      return
    }

    try {
      await proxy?.$modal.confirm(t('share.confirmBatchDeleteShare', { count: selectedShares.value.length }))
      batchDeleting.value = true

      // 提示开发中
      proxy?.$modal.msg(t('share.batchDeletePending'))

      // 清空选择（包括表格多选框）
      selectedShares.value = []
      tableRef.value?.clearSelection()
    } catch (error: any) {
      if (error !== 'cancel') {
        proxy?.$modal.msgError(error.message || t('files.operationFailed'))
      }
    } finally {
      batchDeleting.value = false
    }
  }
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
    align-items: flex-start;
    width: 100%;
    gap: 12px;
  }

  .header-left {
    display: flex;
    flex-direction: column;
    gap: 12px;
    flex: 1;
    min-width: 0;
  }

  .title-section {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
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
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .share-count {
    font-size: 14px;
    color: var(--text-secondary);
    margin-left: 8px;
  }

  .batch-selection-info {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background: var(--el-fill-color-light);
    border-radius: 8px;
    flex-wrap: wrap;
  }

  html.dark .batch-selection-info {
    background: rgba(99, 102, 241, 0.15);
  }

  .selected-count {
    font-size: 14px;
    color: var(--primary-color);
    font-weight: 500;
  }

  .table-card {
    flex: 1;
    border-radius: 16px;
    padding: 8px;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  /* PC端表格样式 */
  .desktop-table {
    display: table;
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
    background: var(--el-fill-color-lighter) !important;
  }

  /* 隐藏表格自带的空状态显示，使用手动的 el-empty */
  :deep(.el-table__empty-block) {
    display: none;
  }

  .file-name-cell {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .file-name {
    font-weight: 500;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .link-cell {
    width: 100%;
  }

  .share-link-input {
    width: 100%;
  }

  .share-link-input :deep(.el-input__inner) {
    font-size: 13px;
    font-family: 'Courier New', monospace;
  }

  .action-buttons {
    display: flex;
    gap: 8px;
    justify-content: center;
  }

  :deep(.el-tag) {
    border-radius: 6px;
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  .time-cell {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    font-size: 13px;
    color: var(--text-secondary);
  }

  .expired-text {
    color: var(--el-color-danger) !important;
    font-weight: 500;
  }

  /* PC端状态徽章样式 */
  .status-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    cursor: pointer;
    transition: all 0.2s;
  }

  .status-badge.has-password {
    background: var(--el-warning-color-light-9);
    color: var(--el-color-warning);
  }

  .status-badge.has-password:hover {
    background: var(--el-warning-color-light-8);
    transform: scale(1.1);
  }

  html.dark .status-badge.has-password {
    background: rgba(230, 162, 60, 0.2);
  }

  html.dark .status-badge.has-password:hover {
    background: rgba(230, 162, 60, 0.3);
  }

  .status-badge.no-password {
    background: var(--el-fill-color-light);
    color: var(--el-color-info);
  }

  .status-badge.no-password:hover {
    background: var(--el-fill-color);
    transform: scale(1.1);
  }

  html.dark .status-badge.no-password {
    background: rgba(144, 147, 153, 0.15);
  }

  html.dark .status-badge.no-password:hover {
    background: rgba(144, 147, 153, 0.25);
  }

  .download-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    padding: 4px 8px;
    border-radius: 12px;
    background: var(--el-fill-color-light);
    color: var(--el-color-info);
    cursor: pointer;
    transition: all 0.2s;
    font-size: 13px;
  }

  .download-badge:hover {
    background: var(--el-fill-color);
    transform: translateY(-1px);
  }

  html.dark .download-badge {
    background: rgba(144, 147, 153, 0.15);
  }

  html.dark .download-badge:hover {
    background: rgba(144, 147, 153, 0.25);
  }

  .download-count-text {
    font-weight: 500;
    font-size: 12px;
  }

  /* 移动端卡片列表 */
  .mobile-share-list {
    display: none;
  }

  .mobile-share-item {
    padding: 16px;
    border-bottom: 1px solid var(--el-border-color-lighter);
    background: var(--el-bg-color-overlay);
    transition: all 0.2s;
    border-radius: 8px;
    margin-bottom: 12px;
    border: 2px solid transparent;
  }

  .mobile-share-item.selected {
    background: var(--el-fill-color-light);
    border-color: var(--primary-color);
  }

  html.dark .mobile-share-item.selected {
    background: rgba(99, 102, 241, 0.15);
  }

  .mobile-checkbox {
    flex-shrink: 0;
    margin-right: 12px;
  }

  .mobile-share-item:last-child {
    border-bottom: none;
    margin-bottom: 0;
  }

  .mobile-share-item:active {
    background-color: var(--el-fill-color-light);
  }

  .share-item-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 12px;
  }

  .share-item-info {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    flex: 1;
    min-width: 0;
  }

  .share-icon {
    flex-shrink: 0;
    margin-top: 2px;
  }

  .share-name-wrapper {
    flex: 1;
    min-width: 0;
  }

  .share-name {
    font-size: 15px;
    font-weight: 500;
    color: var(--el-text-color-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin-bottom: 6px;
  }

  .share-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  /* 移动端状态徽章样式 */
  .mobile-status-badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 3px 8px;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 500;
    white-space: nowrap;
  }

  .mobile-status-badge.has-password {
    background: var(--el-warning-color-light-9);
    color: var(--el-color-warning);
  }

  html.dark .mobile-status-badge.has-password {
    background: rgba(230, 162, 60, 0.2);
  }

  .mobile-status-badge.no-password {
    background: var(--el-fill-color-light);
    color: var(--el-color-info);
  }

  html.dark .mobile-status-badge.no-password {
    background: rgba(144, 147, 153, 0.15);
  }

  .status-text {
    font-size: 11px;
  }

  .mobile-download-badge {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    padding: 3px 8px;
    border-radius: 12px;
    background: var(--el-fill-color-light);
    color: var(--el-text-color-secondary);
    font-size: 11px;
    white-space: nowrap;
  }

  html.dark .mobile-download-badge {
    background: rgba(144, 147, 153, 0.15);
  }

  .download-text {
    font-weight: 500;
    font-size: 11px;
  }

  .share-actions {
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

  .share-link-wrapper {
    margin-bottom: 12px;
  }

  .mobile-share-link-input {
    width: 100%;
  }

  .mobile-share-link-input :deep(.el-input__inner) {
    font-size: 12px;
    font-family: 'Courier New', monospace;
  }

  .share-time-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .time-item {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  /* 移动端响应式 */
  @media (max-width: 1024px) {
    .desktop-table {
      display: none !important;
    }

    .mobile-share-list {
      display: block;
    }

    .header-card {
      padding: 12px 16px;
    }

    .header {
      flex-direction: column;
      align-items: stretch;
    }

    .header-left {
      width: 100%;
    }

    .title-section {
      gap: 8px;
    }

    .title-section h2 {
      font-size: 18px;
    }

    .share-count {
      font-size: 12px;
      margin-left: 4px;
    }

    .batch-selection-info {
      gap: 6px;
      padding: 6px 10px;
    }

    .selected-count {
      font-size: 13px;
    }

    .header .el-button {
      width: 100%;
      margin-top: 8px;
    }

    .password-dialog :deep(.el-dialog) {
      width: 95% !important;
      margin: 0 auto;
    }

    .password-dialog :deep(.el-form-item__label) {
      font-size: 14px;
    }
  }

  @media (max-width: 480px) {
    .header-card {
      padding: 10px 12px;
    }

    .title-section {
      gap: 6px;
    }

    .title-section h2 {
      font-size: 16px;
    }

    .share-count {
      font-size: 11px;
    }

    .batch-selection-info {
      gap: 4px;
      padding: 6px 8px;
    }

    .selected-count {
      font-size: 12px;
    }

    .batch-selection-info .el-button {
      font-size: 12px;
      padding: 4px 8px;
    }

    .mobile-share-item {
      padding: 12px;
    }

    .share-name {
      font-size: 14px;
    }

    .share-meta {
      font-size: 11px;
    }

    .share-time-info {
      font-size: 11px;
    }

    .mobile-share-link-input :deep(.el-input__inner) {
      font-size: 11px;
    }

    .password-dialog :deep(.el-dialog) {
      width: 100% !important;
      margin: 0;
      border-radius: 0;
    }

    .password-dialog :deep(.el-form-item__label) {
      font-size: 13px;
    }
  }

  /* 表格移动端隐藏列 */
  .shares-table :deep(.mobile-hide) {
    display: table-cell;
  }

  .shares-table :deep(.mobile-name-column) {
    min-width: 200px;
  }

  .shares-table :deep(.mobile-link-column) {
    min-width: 300px;
  }

  .shares-table :deep(.mobile-actions-column) {
    width: auto;
    min-width: 120px;
  }
</style>

<template>
  <div class="trash-page">
    <!-- 工具栏 -->
    <el-card shadow="never" class="toolbar-card">
      <div class="toolbar">
        <div class="toolbar-content">
          <!-- 选择提示 -->
          <div class="toolbar-selection" v-if="selectedIds.length > 0">
            <el-tag type="info" size="small">{{ t('common.selected', { count: selectedIds.length }) }}</el-tag>
          </div>

          <!-- 操作按钮组 -->
          <div class="toolbar-actions">
            <!-- 第一行：需要选择的操作 -->
            <div class="action-row action-row-primary">
              <el-button
                icon="RefreshRight"
                :disabled="selectedIds.length === 0"
                @click="handleRestore"
                size="small"
                class="action-btn"
              >
                {{ t('trash.restore') }}
              </el-button>
              <el-button
                icon="Delete"
                type="danger"
                :disabled="selectedIds.length === 0"
                @click="handleDeletePermanently"
                size="small"
                class="action-btn"
              >
                {{ t('trash.permanentDelete') }}
              </el-button>
            </div>

            <!-- 第二行：独立操作 -->
            <div class="action-row action-row-secondary">
              <el-button
                icon="Delete"
                type="danger"
                @click="handleEmptyTrash"
                size="small"
                class="action-btn action-btn-full"
              >
                {{ t('trash.empty') }}
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <!-- PC端：表格布局 -->
    <el-table
      v-loading="loading"
      :data="fileList"
      @selection-change="handleSelectionChange"
      class="trash-table desktop-table"
    >
      <el-table-column type="selection" width="55" class-name="mobile-hide" />
      <el-table-column :label="t('trash.name')" min-width="300" class-name="mobile-name-column">
        <template #default="{ row }">
          <div class="file-name-cell">
            <div class="list-file-icon">
              <file-icon
                :mime-type="row.mime_type"
                :file-name="row.file_name"
                :thumbnail-url="getThumbnailUrl(row.file_id)"
                :show-thumbnail="row.has_thumbnail"
                :icon-size="24"
                :show-badge="false"
                :is-encrypted="row.is_enc"
              />
            </div>
            <div class="file-name-content">
              <file-name-tooltip :file-name="row.file_name" view-mode="table" />
              <div class="file-name-tags">
                <el-tag v-if="row.is_enc" size="small" type="warning" class="enc-tag-inline">
                  <el-icon :size="12"><Lock /></el-icon>
                  {{ t('trash.encrypted') }}
                </el-tag>
                <el-tooltip
                  v-if="isExpired(row.deleted_at) || isExpiringSoon(row.deleted_at)"
                  :content="
                    t('trash.willPermanentDelete', { date: formatDate(getExpireTime(row.deleted_at).toISOString()) })
                  "
                  placement="top"
                >
                  <el-tag
                    :type="isExpired(row.deleted_at) ? 'danger' : 'warning'"
                    size="small"
                    effect="plain"
                    class="expire-tag-inline"
                  >
                    <el-icon :size="10"><Warning /></el-icon>
                    {{ getExpireStatusText(row.deleted_at) }}
                  </el-tag>
                </el-tooltip>
              </div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="t('trash.size')" width="120" class-name="mobile-hide">
        <template #default="{ row }">
          {{ formatSize(row.file_size) }}
        </template>
      </el-table-column>
      <el-table-column :label="t('trash.deletedTime')" width="200" class-name="mobile-hide">
        <template #default="{ row }">
          <div class="time-cell">
            <el-icon :size="14"><Clock /></el-icon>
            <span>{{ formatDate(row.deleted_at) }}</span>
          </div>
          <div
            v-if="isExpired(row.deleted_at) || isExpiringSoon(row.deleted_at)"
            class="expire-info-cell"
            :class="{
              expired: isExpired(row.deleted_at),
              'expiring-soon': isExpiringSoon(row.deleted_at) && !isExpired(row.deleted_at)
            }"
          >
            <el-icon :size="12"><Warning /></el-icon>
            <span class="expire-text">{{ getExpireStatusText(row.deleted_at) }}</span>
            <el-tooltip
              :content="
                t('trash.willPermanentDelete', { date: formatDate(getExpireTime(row.deleted_at).toISOString()) })
              "
              placement="top"
            >
              <el-icon :size="12" class="info-icon"><InfoFilled /></el-icon>
            </el-tooltip>
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="t('trash.operation')" width="200" fixed="right" class-name="mobile-actions-column">
        <template #default="{ row }">
          <div class="action-buttons">
            <el-button link icon="RefreshRight" type="primary" @click.stop="handleRestoreFile(row)" size="small">{{
              t('trash.restore')
            }}</el-button>
            <el-button link icon="Delete" type="danger" @click.stop="handleDeleteFilepermanently(row)" size="small">{{
              t('trash.permanentDelete')
            }}</el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <!-- 移动端：卡片布局 -->
    <div class="mobile-trash-list" v-loading="loading">
      <div
        v-for="row in fileList"
        :key="row.recycled_id"
        class="mobile-trash-item"
        :class="{ selected: selectedIds.includes(row.recycled_id) }"
        @click="toggleSelectItem(row)"
      >
        <div class="trash-item-header">
          <div class="trash-item-info">
            <el-checkbox
              :model-value="selectedIds.includes(row.recycled_id)"
              @change="() => toggleSelectItem(row)"
              @click.stop
              class="trash-checkbox"
            />
            <div class="list-file-icon">
              <file-icon
                :mime-type="row.mime_type"
                :file-name="row.file_name"
                :thumbnail-url="getThumbnailUrl(row.file_id)"
                :show-thumbnail="row.has_thumbnail"
                :icon-size="24"
                :show-badge="false"
                :is-encrypted="row.is_enc"
              />
            </div>
            <div class="trash-name-wrapper">
              <file-name-tooltip :file-name="row.file_name" view-mode="list" custom-class="trash-name" />
              <div class="trash-meta">
                <span class="trash-size">{{ formatSize(row.file_size) }}</span>
                <el-tag v-if="row.is_enc" size="small" type="warning" effect="plain" class="enc-tag">
                  <el-icon :size="10"><Lock /></el-icon>
                  {{ t('trash.encrypted') }}
                </el-tag>
                <span class="trash-time">
                  <el-icon :size="12"><Clock /></el-icon>
                  {{ formatDate(row.deleted_at) }}
                </span>
              </div>
              <!-- 过期提示 -->
              <div
                v-if="isExpired(row.deleted_at) || isExpiringSoon(row.deleted_at)"
                class="trash-expire-warning"
                :class="{
                  expired: isExpired(row.deleted_at),
                  'expiring-soon': isExpiringSoon(row.deleted_at) && !isExpired(row.deleted_at)
                }"
              >
                <el-icon :size="12"><Warning /></el-icon>
                <span class="expire-text">{{ getExpireStatusText(row.deleted_at) }}</span>
              </div>
            </div>
          </div>
          <div class="trash-actions">
            <el-button link type="primary" @click.stop="handleRestoreFile(row)" class="action-btn">
              <el-icon><RefreshRight /></el-icon>
            </el-button>
            <el-button link type="danger" @click.stop="handleDeleteFilepermanently(row)" class="action-btn">
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <EmptyState
      v-if="!loading && fileList.length === 0"
      type="trash"
      :show-actions="false"
      compact
    />

    <!-- 分页 -->
    <pagination
      v-if="total > 0"
      v-model:page="currentPage"
      v-model:limit="pageSize"
      :total="total"
      :page-sizes="[20, 50, 100]"
      @pagination="handlePagination"
      class="pagination"
    />
  </div>
</template>

<script setup lang="ts">
  import { getRecycledList, restoreFile, deleteFilePermanently, emptyRecycled, type RecycledItem } from '@/api/recycled'
  import { getThumbnailUrl } from '@/api/file'
  import { formatSize, formatDate } from '@/utils'
  import { useI18n } from '@/composables'
  import { useUserStore } from '@/stores'
  import EmptyState from '@/components/EmptyState/index.vue'

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const userStore = useUserStore()
  const { t } = useI18n()

  // 数据
  const loading = ref(false)
  const fileList = ref<RecycledItem[]>([])
  const total = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(20)
  const selectedIds = ref<string[]>([])

  // 回收站保留天数（后端配置为30天）
  const RECYCLED_RETENTION_DAYS = 30

  // 计算过期时间
  const getExpireTime = (deletedAt: string): Date => {
    const deleted = new Date(deletedAt)
    return new Date(deleted.getTime() + RECYCLED_RETENTION_DAYS * 24 * 60 * 60 * 1000)
  }

  // 检查是否已过期
  const isExpired = (deletedAt: string): boolean => {
    const expireTime = getExpireTime(deletedAt)
    return new Date() > expireTime
  }

  // 检查是否即将过期（3天内）
  const isExpiringSoon = (deletedAt: string): boolean => {
    if (isExpired(deletedAt)) return false
    const expireTime = getExpireTime(deletedAt)
    const now = new Date()
    const daysLeft = Math.ceil((expireTime.getTime() - now.getTime()) / (24 * 60 * 60 * 1000))
    return daysLeft <= 3
  }

  // 获取剩余天数
  const getRemainingDays = (deletedAt: string): number => {
    const expireTime = getExpireTime(deletedAt)
    const now = new Date()
    const daysLeft = Math.ceil((expireTime.getTime() - now.getTime()) / (24 * 60 * 60 * 1000))
    return Math.max(0, daysLeft)
  }

  // 获取过期状态文本
  const getExpireStatusText = (deletedAt: string): string => {
    if (isExpired(deletedAt)) {
      return t('trash.expired')
    }
    const daysLeft = getRemainingDays(deletedAt)
    if (daysLeft === 0) {
      return t('trash.expireToday')
    } else if (daysLeft === 1) {
      return t('trash.expireTomorrow')
    } else {
      return t('trash.expireDays', { days: daysLeft })
    }
  }

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
        proxy?.$modal.msgError(res.message || t('trash.getListFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('trash.getListFailed'))
    } finally {
      loading.value = false
    }
  }

  // 选择变化
  const handleSelectionChange = (selection: RecycledItem[]) => {
    selectedIds.value = selection.map(item => item.recycled_id)
  }

  // 移动端切换选择
  const toggleSelectItem = (item: RecycledItem) => {
    const index = selectedIds.value.indexOf(item.recycled_id)
    if (index > -1) {
      selectedIds.value.splice(index, 1)
    } else {
      selectedIds.value.push(item.recycled_id)
    }
  }

  // 还原文件（批量）
  const handleRestore = async () => {
    if (selectedIds.value.length === 0) {
      proxy?.$modal.msgWarning(t('trash.pleaseSelectRestore'))
      return
    }

    try {
      await proxy?.$modal.confirm(t('trash.confirmRestore', { count: selectedIds.value.length }))
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
        proxy?.$modal.msgSuccess(t('trash.successRestore', { count: successCount }))
      }
      if (failedCount > 0) {
        proxy?.$modal.msgWarning(t('trash.restoreFilesFailed', { count: failedCount }))
      }

      selectedIds.value = []
      await loadRecycledList()
    } catch (error: any) {
      if (error !== 'cancel') {
        // 用户取消操作
      }
    }
  }

  // 还原单个文件
  const handleRestoreFile = async (item: RecycledItem) => {
    try {
      await proxy?.$modal.confirm(t('trash.confirmRestoreFile', { fileName: item.file_name }))
      const res = await restoreFile(item.recycled_id)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('trash.restoreSuccess'))
        await loadRecycledList()
      } else {
        proxy?.$modal.msgError(res.message || t('trash.restoreFailed'))
      }
    } catch (error: any) {
      if (error !== 'cancel') {
        proxy?.$modal.msgError(error.message || t('trash.restoreFailed'))
      }
    }
  }

  // 永久删除（批量）
  const handleDeletePermanently = async () => {
    if (selectedIds.value.length === 0) {
      proxy?.$modal.msgWarning(t('trash.pleaseSelectDelete'))
      return
    }

    try {
      await proxy?.$modal.confirm(t('trash.confirmPermanentDelete', { count: selectedIds.value.length }))
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
        proxy?.$modal.msgSuccess(t('trash.successDelete', { count: successCount }))
        // 永久删除成功后刷新用户信息，更新存储空间显示
        await userStore.fetchUserInfo()
      }
      if (failedCount > 0) {
        proxy?.$modal.msgWarning(t('trash.deleteFilesFailed', { count: failedCount }))
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
      await proxy?.$modal.confirm(t('trash.confirmPermanentDeleteFile', { fileName: item.file_name }))
      const res = await deleteFilePermanently(item.recycled_id)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('trash.deleteSuccess'))
        await loadRecycledList()
        // 永久删除成功后刷新用户信息，更新存储空间显示
        await userStore.fetchUserInfo()
      } else {
        proxy?.$modal.msgError(res.message || t('trash.deleteFailed'))
      }
    } catch (error: any) {
      if (error !== 'cancel') {
        proxy?.$modal.msgError(error.message || t('trash.deleteFailed'))
      }
    }
  }

  // 清空回收站
  const handleEmptyTrash = async () => {
    if (total.value === 0) {
      proxy?.$modal.msg(t('trash.alreadyEmpty'))
      return
    }

    try {
      await proxy?.$modal.confirm(t('trash.emptyConfirmAll', { count: total.value }))
      loading.value = true
      try {
        const res = await emptyRecycled()
        if (res.code === 200) {
          proxy?.$modal.msgSuccess(res.message || t('trash.emptySuccess'))
          await loadRecycledList()
          // 清空回收站成功后刷新用户信息，更新存储空间显示
          await userStore.fetchUserInfo()
        } else {
          proxy?.$modal.msgError(res.message || t('trash.emptyFailed'))
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || t('trash.emptyFailed'))
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
  const handlePagination = ({ page, limit }: { page: number; limit: number }) => {
    currentPage.value = page
    pageSize.value = limit
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
    background: var(--bg-color);
    min-height: calc(100vh - 60px);
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .toolbar-card {
    flex-shrink: 0;
  }

  .toolbar {
    width: 100%;
  }

  .toolbar-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 16px;
  }

  .toolbar-selection {
    display: flex;
    align-items: center;
    flex-shrink: 0;
  }

  .toolbar-actions {
    display: flex;
    align-items: center;
    gap: 10px;
    flex: 1;
    justify-content: flex-end;
  }

  .action-row {
    display: flex;
    gap: 10px;
    align-items: center;
  }

  .action-row-primary {
    display: flex;
  }

  .action-row-secondary {
    display: flex;
  }

  .action-btn {
    min-width: auto;
  }

  .action-btn-full {
    min-width: 100px;
  }

  .file-name-cell {
    display: flex;
    align-items: flex-start;
    gap: 12px;
  }

  .list-file-icon {
    flex-shrink: 0;
    margin-top: 2px;
  }

  .file-name-content {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .file-name-content > span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    line-height: 1.5;
  }

  .file-name-tags {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
  }

  .expire-tag-inline {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    font-size: 11px;
  }

  .enc-tag-inline {
    border: none;
    background: linear-gradient(135deg, var(--warning-color) 0%, var(--warning-color) 100%);
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

  .action-buttons {
    display: flex;
    gap: 8px;
    justify-content: center;
  }

  .time-cell {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: var(--el-text-color-regular);
    margin-bottom: 4px;
  }

  .expire-info-cell {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 3px 8px;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 500;
    margin-top: 4px;
  }

  .expire-info-cell.expired {
    background: rgba(245, 101, 101, 0.1);
    color: var(--el-color-danger);
    border: 1px solid rgba(245, 101, 101, 0.3);
  }

  .expire-info-cell.expiring-soon {
    background: rgba(230, 162, 60, 0.1);
    color: var(--el-color-warning);
    border: 1px solid rgba(230, 162, 60, 0.3);
  }

  .expire-info-cell .expire-text {
    font-size: 11px;
    font-weight: 500;
  }

  .expire-info-cell .info-icon {
    margin-left: 2px;
    cursor: help;
    opacity: 0.7;
    transition: opacity 0.2s;
  }

  .expire-info-cell .info-icon:hover {
    opacity: 1;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: center;
    flex-shrink: 0;
  }

  /* PC端表格样式 */
  .desktop-table {
    display: table;
  }

  /* 隐藏表格自带的空状态显示，使用手动的 el-empty */
  .trash-table :deep(.el-table__empty-block) {
    display: none;
  }

  .trash-table :deep(.el-table) {
    background: transparent !important;
    --el-table-tr-bg-color: transparent;
    --el-table-header-bg-color: transparent;
  }

  .trash-table :deep(.el-table th.el-table__cell) {
    background: transparent !important;
    color: var(--el-text-color-primary);
    font-weight: 600;
    font-size: 13px;
    border-bottom-color: var(--el-border-color-lighter);
  }

  .trash-table :deep(.el-table td.el-table__cell) {
    background: transparent !important;
    color: var(--el-text-color-primary);
    border-bottom-color: var(--el-border-color-lighter);
  }

  .trash-table :deep(.el-table tr) {
    background: transparent !important;
    transition: all 0.2s;
  }

  .trash-table :deep(.el-table--enable-row-hover .el-table__body tr:hover > td.el-table__cell) {
    background: var(--el-fill-color-lighter) !important;
  }

  /* 表格移动端隐藏列 */
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

  /* 移动端卡片列表 */
  .mobile-trash-list {
    display: none;
  }

  .mobile-trash-item {
    padding: 16px;
    border-bottom: 1px solid var(--el-border-color-lighter);
    background: var(--el-bg-color-overlay);
    transition: background-color 0.2s;
    border-radius: 8px;
    margin-bottom: 12px;
  }

  .mobile-trash-item:last-child {
    border-bottom: none;
    margin-bottom: 0;
  }

  .mobile-trash-item:active {
    background-color: var(--el-fill-color-light);
  }

  .mobile-trash-item.selected {
    border-color: var(--el-color-primary);
    background-color: var(--el-color-primary-light-9);
  }

  .trash-item-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
  }

  .trash-item-info {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    flex: 1;
    min-width: 0;
  }

  .trash-checkbox {
    flex-shrink: 0;
    margin-top: 2px;
  }

  .trash-name-wrapper {
    flex: 1;
    min-width: 0;
  }

  .trash-name {
    font-size: 15px;
    font-weight: 500;
    color: var(--el-text-color-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin-bottom: 6px;
  }

  .trash-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .trash-size {
    white-space: nowrap;
  }

  .enc-tag {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    padding: 2px 6px;
  }

  .trash-time {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    white-space: nowrap;
  }

  /* 过期警告提示 */
  .trash-expire-warning {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 500;
    margin-top: 6px;
    white-space: nowrap;
  }

  .trash-expire-warning.expired {
    background: rgba(245, 101, 101, 0.1);
    color: var(--el-color-danger);
    border: 1px solid rgba(245, 101, 101, 0.3);
  }

  .trash-expire-warning.expiring-soon {
    background: rgba(230, 162, 60, 0.1);
    color: var(--el-color-warning);
    border: 1px solid rgba(230, 162, 60, 0.3);
  }

  .trash-expire-warning .el-icon {
    flex-shrink: 0;
  }

  .expire-text {
    font-size: 11px;
    font-weight: 500;
  }

  .trash-actions {
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

  /* 移动端响应式 */
  @media (max-width: 1024px) {
    .desktop-table {
      display: none !important;
    }

    .mobile-trash-list {
      display: block;
    }

    .trash-page {
      padding: 12px;
      gap: 12px;
    }

    .toolbar-card {
      padding: 12px 16px;
    }

    .toolbar-content {
      flex-direction: column;
      align-items: stretch;
      gap: 10px;
    }

    .toolbar-selection {
      width: 100%;
      justify-content: flex-start;
    }

    .toolbar-actions {
      flex-direction: column;
      width: 100%;
      gap: 8px;
      align-items: stretch;
    }

    .action-row {
      width: 100%;
      gap: 8px;
    }

    .action-row-primary {
      display: flex;
    }

    .action-row-secondary {
      display: flex;
    }

    .action-btn {
      flex: 1;
      font-size: 13px;
      padding: 8px 12px;
    }

    .action-btn-full {
      flex: 1;
      width: 100%;
    }

    .pagination {
      margin-top: 12px;
    }
  }

  @media (max-width: 480px) {
    .mobile-trash-item {
      padding: 12px;
    }

    .trash-name {
      font-size: 14px;
    }

    .trash-meta {
      font-size: 11px;
    }
  }

  /* 深色模式样式 */
  html.dark .toolbar-card {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .toolbar-card :deep(.el-card__body) {
    background: var(--card-bg);
  }

  html.dark .mobile-trash-item {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .mobile-trash-item.selected {
    background-color: rgba(59, 130, 246, 0.15);
    border-color: var(--primary-color);
  }

  html.dark .mobile-trash-item:active {
    background-color: rgba(59, 130, 246, 0.1);
  }

  html.dark .pagination {
    border-top-color: var(--el-border-color);
  }
</style>

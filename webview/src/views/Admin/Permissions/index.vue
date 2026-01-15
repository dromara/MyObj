<template>
  <div class="admin-permissions">
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" icon="Plus" @click="handleCreate">{{
          t('admin.permissions.newPermission')
        }}</el-button>
        <el-button type="danger" icon="Delete" :disabled="selectedRows.length === 0" @click="handleBatchDelete">
          {{ t('admin.permissions.batchDeleteWithCount', { count: selectedRows.length }) }}
        </el-button>
        <el-button icon="Refresh" @click="loadPowerList">{{ t('common.refresh') }}</el-button>
      </div>
    </div>

    <el-table
      :data="powerList"
      v-loading="loading"
      class="admin-table"
      :empty-text="t('admin.permissions.noPermissions')"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="55" />
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" :label="t('admin.permissions.permissionName')" min-width="150">
        <template #default="{ row }">
          {{ getPermissionName(row.characteristic, row.name) }}
        </template>
      </el-table-column>
      <el-table-column prop="description" :label="t('admin.permissions.description')" min-width="200">
        <template #default="{ row }">
          {{ getPermissionDescription(row.characteristic, row.description) }}
        </template>
      </el-table-column>
      <el-table-column prop="characteristic" :label="t('admin.permissions.characteristic')" min-width="200">
        <template #default="{ row }">
          <code style="font-size: 12px; color: var(--el-color-primary)">{{ row.characteristic }}</code>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" :label="t('admin.users.createTime')" width="180" />
      <el-table-column :label="t('admin.users.operation')" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="handleEdit(row)">
            {{ t('admin.users.edit') }}
          </el-button>
          <el-button link type="danger" @click="handleDelete(row)">
            {{ t('admin.users.delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页组件 -->
    <Pagination
      v-model:page="pagination.page"
      v-model:limit="pagination.pageSize"
      :total="pagination.total"
      @pagination="handlePagination"
    />

    <!-- 创建/编辑权限对话框 -->
    <el-dialog v-model="showDialog" :title="dialogTitle" :width="dialogWidth" @close="handleDialogClose">
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item :label="t('admin.permissions.permissionName')" prop="name">
          <el-input v-model="formData.name" :placeholder="t('admin.permissions.permissionNamePlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('admin.permissions.description')" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            :placeholder="t('admin.permissions.descriptionPlaceholder')"
          />
        </el-form-item>
        <el-form-item :label="t('admin.permissions.characteristic')" prop="characteristic">
          <el-input v-model="formData.characteristic" :placeholder="t('admin.permissions.characteristicPlaceholder')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSubmit">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import type { ComponentInternalInstance } from 'vue'
  import {
    getAdminPowerList,
    createAdminPower,
    updateAdminPower,
    deleteAdminPower,
    batchDeleteAdminPower,
    type AdminPower,
    type CreatePowerRequest,
    type UpdatePowerRequest
  } from '@/api/admin'
  import type { FormRules, FormInstance } from 'element-plus'
  import { useResponsive, useI18n } from '@/composables'
  import { getPermissionName, getPermissionDescription } from '@/utils/business/permission'

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { isMobile, isPhone } = useResponsive()
  const { t } = useI18n()

  const loading = ref(false)
  const powerList = ref<AdminPower[]>([])
  const selectedRows = ref<AdminPower[]>([])
  const showDialog = ref(false)
  const dialogTitle = ref('')
  const formRef = ref<FormInstance>()
  const isEdit = ref(false)
  const formData = ref<CreatePowerRequest & { id?: number }>({
    name: '',
    description: '',
    characteristic: ''
  })

  // 分页数据
  const pagination = reactive({
    page: 1,
    pageSize: 20,
    total: 0
  })

  // 对话框宽度响应式
  const dialogWidth = computed(() => {
    if (isPhone.value) return '90%'
    if (isMobile.value) return '80%'
    return '600px'
  })

  const formRules: FormRules = {
    name: [{ required: true, message: t('admin.permissions.permissionNameRequired'), trigger: 'blur' }],
    description: [{ required: true, message: t('admin.permissions.descriptionRequired'), trigger: 'blur' }],
    characteristic: [{ required: true, message: t('admin.permissions.characteristicRequired'), trigger: 'blur' }]
  }

  // 加载权限列表
  const loadPowerList = async () => {
    loading.value = true
    try {
      const res = await getAdminPowerList({
        page: pagination.page,
        pageSize: pagination.pageSize
      })
      if (res.code === 200 && res.data) {
        powerList.value = res.data.powers || []
        pagination.total = res.data.total || 0
      } else {
        proxy?.$modal.msgError(t('admin.permissions.loadListFailed'))
        powerList.value = []
        pagination.total = 0
      }
    } catch (error: any) {
      proxy?.$modal.msgError(t('admin.permissions.loadListFailed'))
      proxy?.$log?.error(error)
      powerList.value = []
      pagination.total = 0
    } finally {
      loading.value = false
    }
  }

  // 分页变化处理
  const handlePagination = ({ page, limit }: { page: number; limit: number }) => {
    pagination.page = page
    pagination.pageSize = limit
    loadPowerList()
  }

  // 新建权限
  const handleCreate = () => {
    isEdit.value = false
    dialogTitle.value = t('admin.permissions.newPermission')
    formData.value = {
      name: '',
      description: '',
      characteristic: ''
    }
    showDialog.value = true
    nextTick(() => {
      formRef.value?.clearValidate()
    })
  }

  // 编辑权限
  const handleEdit = (row: AdminPower) => {
    isEdit.value = true
    dialogTitle.value = t('admin.permissions.editPermission')
    formData.value = {
      id: row.id,
      name: row.name,
      description: row.description,
      characteristic: row.characteristic
    }
    showDialog.value = true
    nextTick(() => {
      formRef.value?.clearValidate()
    })
  }

  // 选择变化
  const handleSelectionChange = (selection: AdminPower[]) => {
    selectedRows.value = selection
  }

  // 删除权限
  const handleDelete = (row: AdminPower) => {
    proxy?.$modal
      .confirm(t('admin.permissions.confirmDelete', { name: row.name }))
      .then(async () => {
        try {
          const res = await deleteAdminPower(row.id)
          if (res.code === 200) {
            proxy?.$modal.msgSuccess(t('admin.users.deleteSuccess'))
            loadPowerList()
          } else {
            proxy?.$modal.msgError(res.message || t('admin.users.deleteFailed'))
          }
        } catch (error: any) {
          proxy?.$modal.msgError(error.response?.data?.message || t('admin.users.deleteFailed'))
          proxy?.$log?.error(error)
        }
      })
      .catch(() => {})
  }

  // 批量删除权限
  const handleBatchDelete = () => {
    if (selectedRows.value.length === 0) {
      proxy?.$modal.msgWarning(t('admin.permissions.pleaseSelectDelete'))
      return
    }

    const names = selectedRows.value.map(row => row.name).join('、')
    proxy?.$modal
      .confirm(t('admin.permissions.confirmBatchDelete', { count: selectedRows.value.length }) + `\n${names}`)
      .then(async () => {
        try {
          const ids = selectedRows.value.map(row => row.id)
          const res = await batchDeleteAdminPower({ ids })
          if (res.code === 200) {
            proxy?.$modal.msgSuccess(res.message || t('admin.permissions.batchDeleteSuccess'))
            selectedRows.value = []
            loadPowerList()
          } else {
            proxy?.$modal.msgError(res.message || t('admin.permissions.batchDeleteFailed'))
          }
        } catch (error: any) {
          proxy?.$modal.msgError(error.response?.data?.message || t('admin.permissions.batchDeleteFailed'))
          proxy?.$log?.error(error)
        }
      })
      .catch(() => {})
  }

  // 提交表单
  const handleSubmit = async () => {
    if (!formRef.value) return

    await formRef.value.validate(async valid => {
      if (!valid) return

      try {
        if (isEdit.value) {
          const res = await updateAdminPower(formData.value as UpdatePowerRequest)
          if (res.code === 200) {
            proxy?.$modal.msgSuccess(t('admin.users.updateSuccess'))
            showDialog.value = false
            loadPowerList()
          } else {
            proxy?.$modal.msgError(res.message || t('admin.users.updateFailed'))
          }
        } else {
          const res = await createAdminPower(formData.value as CreatePowerRequest)
          if (res.code === 200) {
            proxy?.$modal.msgSuccess(t('admin.users.createSuccess'))
            showDialog.value = false
            loadPowerList()
          } else {
            proxy?.$modal.msgError(res.message || t('admin.users.createFailed'))
          }
        }
      } catch (error: any) {
        proxy?.$modal.msgError(
          error.response?.data?.message ||
            (isEdit.value ? t('admin.users.updateFailed') : t('admin.users.createFailed'))
        )
        proxy?.$log?.error(error)
      }
    })
  }

  // 对话框关闭
  const handleDialogClose = () => {
    formRef.value?.resetFields()
    formRef.value?.clearValidate()
  }

  onMounted(() => {
    loadPowerList()
  })
</script>

<style scoped>
  .admin-permissions {
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 16px;
    min-height: 0;
    overflow: hidden;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    flex-shrink: 0;
  }

  .toolbar-left {
    display: flex;
    gap: 12px;
  }

  .admin-table {
    flex: 1;
    min-height: 0;
    overflow: auto;
  }

  /* 分页组件样式 */
  .admin-permissions :deep(.pagination-container) {
    flex-shrink: 0;
    margin-top: 16px;
  }

  /* 平板端适配 (768px - 1024px) */
  @media (max-width: 1024px) and (min-width: 769px) {
    .toolbar {
      gap: 10px;
    }

    .toolbar-left {
      flex-wrap: wrap;
    }

    .admin-table {
      font-size: 13px;
    }

    .admin-table :deep(.el-table__cell) {
      padding: 10px 6px;
    }

    /* 对话框适配 */
    .admin-permissions :deep(.el-dialog) {
      margin: 5vh auto;
    }
  }

  /* 移动端/平板端适配 (≤768px) */
  @media (max-width: 768px) {
    .admin-permissions {
      gap: 12px;
    }

    .toolbar {
      flex-wrap: wrap;
      gap: 8px;
    }

    .toolbar-left {
      flex: 1;
      min-width: 0;
      width: 100%;
    }

    .toolbar-left .el-button {
      flex: 1;
      min-width: 0;
      font-size: 13px;
    }

    .admin-table {
      font-size: 12px;
    }

    .admin-table :deep(.el-table__cell) {
      padding: 8px 4px;
    }

    /* 表格列优化 */
    .admin-table :deep(.el-table-column--selection) {
      width: 45px !important;
    }

    /* 分页组件优化 */
    .admin-permissions :deep(.pagination-container) {
      margin-top: 12px;
    }

    /* 对话框适配 */
    .admin-permissions :deep(.el-dialog) {
      margin: 5vh auto;
      max-height: 90vh;
    }

    .admin-permissions :deep(.el-dialog__body) {
      max-height: calc(90vh - 120px);
      overflow-y: auto;
    }
  }

  /* 手机端适配 (≤480px) */
  @media (max-width: 480px) {
    .admin-permissions {
      gap: 10px;
    }

    .toolbar {
      gap: 6px;
    }

    .toolbar-left .el-button {
      font-size: 12px;
      padding: 8px 12px;
    }

    /* 按钮文字优化 */
    .toolbar-left .el-button span {
      font-size: 12px;
    }

    .admin-table {
      font-size: 11px;
    }

    .admin-table :deep(.el-table__cell) {
      padding: 6px 2px;
    }

    .admin-table :deep(.el-table-column--selection) {
      width: 40px !important;
    }

    /* 隐藏部分表格列在极小屏幕上 */
    .admin-table :deep(.el-table__header-wrapper),
    .admin-table :deep(.el-table__body-wrapper) {
      overflow-x: auto;
    }

    /* 对话框进一步优化 */
    .admin-permissions :deep(.el-dialog) {
      width: 90% !important;
      margin: 3vh auto;
    }

    .admin-permissions :deep(.el-dialog__header) {
      padding: 15px;
    }

    .admin-permissions :deep(.el-dialog__body) {
      padding: 15px;
      max-height: calc(90vh - 100px);
    }

    .admin-permissions :deep(.el-form-item__label) {
      font-size: 13px;
      width: 80px !important;
    }

    .admin-permissions :deep(.el-input),
    .admin-permissions :deep(.el-textarea) {
      font-size: 14px;
    }
  }

  /* 深色模式样式 */
  html.dark .admin-permissions {
    background: transparent;
  }

  html.dark .pagination-container {
    border-top-color: var(--el-border-color);
  }

  html.dark :deep(.el-dialog) {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark :deep(.el-dialog__header) {
    background: var(--card-bg);
    border-bottom-color: var(--el-border-color);
  }

  html.dark :deep(.el-dialog__title) {
    color: var(--el-text-color-primary);
  }

  html.dark :deep(.el-dialog__body) {
    background: var(--card-bg);
    color: var(--el-text-color-primary);
  }

  html.dark :deep(.el-form-item__label) {
    color: var(--el-text-color-primary);
  }
</style>

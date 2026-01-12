<template>
  <div class="admin-users">
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" icon="Plus" @click="handleCreate">{{ t('admin.users.newUser') }}</el-button>
        <el-button icon="Refresh" @click="loadUserList">{{ t('common.refresh') }}</el-button>
      </div>
      <div class="toolbar-right">
        <el-input
          v-model="searchKeyword"
          :placeholder="t('admin.users.searchPlaceholder')"
          clearable
          style="width: 300px"
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
    </div>

    <el-table
      :data="userList"
      v-loading="loading"
      class="admin-table"
      :empty-text="t('admin.users.noUsers')"
    >
      <el-table-column prop="user_name" :label="t('admin.users.username')" min-width="120" />
      <el-table-column prop="name" :label="t('admin.users.nickname')" min-width="120" />
      <el-table-column prop="email" :label="t('admin.users.email')" min-width="180" />
      <el-table-column prop="phone" :label="t('admin.users.phone')" min-width="120" />
      <el-table-column prop="group_name" :label="t('admin.users.userGroup')" width="120">
        <template #default="{ row }">
          <el-tag>{{ row.group_name || t('admin.users.userGroup') + row.group_id }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('admin.users.storageSpace')" width="150">
        <template #default="{ row }">
          {{ formatStorage(row.space) }}
        </template>
      </el-table-column>
      <el-table-column prop="state" :label="t('admin.users.status')" width="100" align="center">
        <template #default="{ row }">
          <el-switch 
            v-model="row.state" 
            :active-value="0" 
            :inactive-value="1" 
            :disabled="row.group_id === 1"
            @change="handleStatusChange(row)"
          />
        </template>
      </el-table-column>
      <el-table-column prop="created_at" :label="t('admin.users.createTime')" width="180" />
      <el-table-column :label="t('admin.users.operation')" width="150" fixed="right">
        <template #default="{ row }">
          <el-button 
            v-if="row.group_id !== 1" 
            link 
            type="primary" 
            @click="handleEdit(row)"
          >
            {{ t('admin.users.edit') }}
          </el-button>
          <el-button 
            v-if="row.group_id !== 1" 
            link 
            type="danger" 
            @click="handleDelete(row)"
          >
            {{ t('admin.users.delete') }}
          </el-button>
          <span v-if="row.group_id === 1" style="color: var(--el-text-color-secondary); font-size: 12px;">
            {{ t('admin.users.adminCannotOperate') }}
          </span>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="pagination.page"
      v-model:page-size="pagination.pageSize"
      :total="pagination.total"
      :page-sizes="[10, 20, 50, 100]"
      layout="total, sizes, prev, pager, next, jumper"
      @size-change="loadUserList"
      @current-change="loadUserList"
      class="pagination"
    />

    <!-- 创建/编辑用户对话框 -->
    <el-dialog
      v-model="showDialog"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item :label="t('admin.users.username')" prop="user_name">
          <el-input v-model="formData.user_name" :disabled="isEdit" />
        </el-form-item>
        <el-form-item :label="t('admin.users.password')" prop="password" v-if="!isEdit">
          <el-input v-model="formData.password" type="password" show-password />
        </el-form-item>
        <el-form-item :label="t('admin.users.nickname')">
          <el-input v-model="formData.name" />
        </el-form-item>
        <el-form-item :label="t('admin.users.email')">
          <el-input v-model="formData.email" />
        </el-form-item>
        <el-form-item :label="t('admin.users.phone')">
          <el-input v-model="formData.phone" />
        </el-form-item>
        <el-form-item :label="t('admin.users.userGroup')" prop="group_id">
          <el-select v-model="formData.group_id" style="width: 100%">
            <el-option
              v-for="group in groupList"
              :key="group.id"
              :label="group.name"
              :value="group.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('admin.users.storageSpaceGB')" prop="space">
          <el-input-number 
            v-model="formData.space" 
            :min="0" 
            :max="maxSpaceInGB" 
            style="width: 100%" 
          />
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px;">
            {{ t('admin.users.unlimitedSpace') }}
            <span v-if="maxSpaceInGB > 0 && maxSpaceInGB < 999999" style="color: var(--el-color-warning);">
              {{ t('admin.users.groupLimit', { limit: maxSpaceInGB }) }}
            </span>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import type { ComponentInternalInstance } from 'vue'
import { 
  getAdminUserList, 
  createAdminUser, 
  updateAdminUser, 
  deleteAdminUser, 
  toggleUserState,
  getAdminGroupList,
  type AdminUser,
  type CreateUserRequest,
  type UpdateUserRequest,
  type AdminGroup
} from '@/api/admin'
import { formatSize, bytesToGB, GBToBytes } from '@/utils'
import type { FormRules } from 'element-plus'
import { useI18n } from '@/composables/useI18n'

const { proxy } = getCurrentInstance() as ComponentInternalInstance
const { t } = useI18n()

const loading = ref(false)
const submitting = ref(false)
const userList = ref<AdminUser[]>([])
const groupList = ref<AdminGroup[]>([])
const searchKeyword = ref('')
const showDialog = ref(false)
const isEdit = ref(false)
const formRef = ref()
const formData = reactive<CreateUserRequest & { id?: string }>({
  user_name: '',
  password: '',
  name: '',
  email: '',
  phone: '',
  group_id: 1,
  space: 0
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const dialogTitle = computed(() => isEdit.value ? t('admin.users.editUser') : t('admin.users.newUser'))

// 根据选中的组，计算存储空间的最大值（GB）
const maxSpaceInGB = computed(() => {
  if (!formData.group_id || formData.group_id === 0) {
    return 999999 // 未选择组时，不限制
  }
  const selectedGroup = groupList.value.find(g => g.id === formData.group_id)
  if (!selectedGroup) {
    return 999999
  }
  // 如果组有存储空间限制（space > 0），则用户不能超过组限制
  if (selectedGroup.space > 0) {
    return bytesToGB(selectedGroup.space)
  }
  // 组无限制（space = 0），用户也不限制
  return 999999
})

const formRules: FormRules = {
  user_name: [{ required: true, message: t('admin.users.usernameRequired'), trigger: 'blur' }],
  password: [{ required: true, message: t('admin.users.passwordRequired'), trigger: 'blur' }],
  name: [{ required: true, message: t('admin.users.nicknameRequired'), trigger: 'blur' }],
  email: [
    { required: true, message: t('admin.users.emailRequired'), trigger: 'blur' },
    { type: 'email', message: t('admin.users.emailFormat'), trigger: 'blur' }
  ],
  phone: [{ required: true, message: t('admin.users.phoneRequired'), trigger: 'blur' }],
  group_id: [{ required: true, message: t('admin.users.groupRequired'), trigger: 'change' }],
  space: [{ required: true, message: t('admin.users.spaceRequired'), trigger: 'blur' }]
}

// 格式化存储空间
const formatStorage = (bytes: number) => {
  if (bytes === 0 || bytes === -1) return t('admin.users.unlimited')
  return formatSize(bytes)
}

// 加载用户列表
const loadUserList = async () => {
  loading.value = true
  try {
    const res = await getAdminUserList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: searchKeyword.value || undefined
    })
    if (res.code === 200 && res.data) {
      userList.value = res.data.users || []
      pagination.total = res.data.total || 0
    } else {
      // 后端接口未实现，提示开发中
      proxy?.$modal.msg(t('admin.users.featureDeveloping'))
      userList.value = []
      pagination.total = 0
    }
  } catch (error: any) {
    proxy?.$modal.msgError(t('admin.users.loadListFailed'))
    proxy?.$log?.error(error)
  } finally {
    loading.value = false
  }
}

// 加载组列表
const loadGroupList = async () => {
  try {
    const res = await getAdminGroupList()
    if (res.code === 200 && res.data) {
      groupList.value = res.data.groups || []
    } else {
      // 后端接口未实现，使用默认组
      groupList.value = [
        { id: 1, name: '管理员', group_default: 0, space: 0, created_at: '' },
        { id: 2, name: '普通用户', group_default: 1, space: 0, created_at: '' }
      ]
    }
  } catch (error: any) {
    proxy?.$log?.error(error)
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  loadUserList()
}

// 创建用户
const handleCreate = () => {
  isEdit.value = false
  Object.assign(formData, {
    user_name: '',
    password: '',
    name: '',
    email: '',
    phone: '',
    group_id: 1,
    space: 0
  })
  showDialog.value = true
}

// 编辑用户
const handleEdit = (user: AdminUser) => {
  // 禁止操作管理员组
  if (user.group_id === 1) {
    proxy?.$modal.msgWarning(t('admin.users.cannotEditAdmin'))
    return
  }
  
  isEdit.value = true
  // 将字节转换为 GB（后端存储的是字节，前端表单输入的是 GB）
  Object.assign(formData, {
    id: user.id,
    user_name: user.user_name,
    name: user.name,
    email: user.email,
    phone: user.phone,
    group_id: user.group_id,
    space: bytesToGB(user.space)
  })
  showDialog.value = true
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (valid) {
      submitting.value = true
      try {
        // 将 GB 转换为字节（前端输入的是 GB，后端存储的是字节）
        const submitData = {
          ...formData,
          space: GBToBytes(formData.space)
        }
        
        if (isEdit.value) {
          const res = await updateAdminUser(submitData as UpdateUserRequest)
          if (res.code === 200) {
            proxy?.$modal.msgSuccess(t('admin.users.updateSuccess'))
            showDialog.value = false
            loadUserList()
          } else {
            proxy?.$modal.msgError(res.message || t('admin.users.updateFailed'))
          }
        } else {
          const res = await createAdminUser(submitData as CreateUserRequest)
          if (res.code === 200) {
            proxy?.$modal.msgSuccess(t('admin.users.createSuccess'))
            showDialog.value = false
            loadUserList()
          } else {
            proxy?.$modal.msgError(res.message || t('admin.users.createFailed'))
          }
        }
      } catch (error: any) {
        if (error.response?.status === 404 || error.message?.includes('404')) {
          proxy?.$modal.msg(t('admin.users.featureDeveloping'))
        } else {
          proxy?.$modal.msgError(error.message || t('common.operationFailed'))
        }
      } finally {
        submitting.value = false
      }
    }
  })
}

// 删除用户
const handleDelete = async (user: AdminUser) => {
  // 禁止操作管理员组
  if (user.group_id === 1) {
    proxy?.$modal.msgWarning(t('admin.users.cannotDeleteAdmin'))
    return
  }
  
  try {
    await proxy?.$modal.confirm(t('admin.users.confirmDelete', { name: user.user_name }))
    try {
      const res = await deleteAdminUser(user.id)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('admin.users.deleteSuccess'))
        loadUserList()
      } else {
        proxy?.$modal.msgError(res.message || t('admin.users.deleteFailed'))
      }
    } catch (error: any) {
      if (error.response?.status === 404 || error.message?.includes('404')) {
        proxy?.$modal.msg(t('admin.users.featureDeveloping'))
      } else {
        proxy?.$modal.msgError(error.message || t('admin.users.deleteFailed'))
      }
    }
  } catch (error: any) {
    // 用户取消
  }
}

// 用户状态修改（参考 plus-ui 实现）
const handleStatusChange = async (row: AdminUser) => {
  // 禁止操作管理员组
  if (row.group_id === 1) {
    proxy?.$modal.msgWarning(t('admin.users.cannotOperateAdmin'))
    // 回滚状态
    row.state = row.state === 0 ? 1 : 0
    return
  }
  
  const action = row.state === 0 ? t('admin.users.enabled') : t('admin.users.disabled')
  const actionKey = row.state === 0 ? 'enable' : 'disable'
  try {
    await proxy?.$modal.confirm(t('admin.users.confirmStatusChange', { action, name: row.user_name }))
    try {
      const res = await toggleUserState(row.id, row.state)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t(`admin.users.${actionKey}Success`))
      } else {
        proxy?.$modal.msgError(res.message || t(`admin.users.${actionKey}Failed`))
        // 操作失败，回滚状态
        row.state = row.state === 0 ? 1 : 0
      }
    } catch (error: any) {
      // 操作失败，回滚状态
      row.state = row.state === 0 ? 1 : 0
      if (error.response?.status === 404 || error.message?.includes('404')) {
        proxy?.$modal.msg(t('admin.users.featureDeveloping'))
      } else {
        proxy?.$modal.msgError(error.message || t(`admin.users.${actionKey}Failed`))
      }
    }
  } catch (error: any) {
    // 用户取消，回滚状态
    row.state = row.state === 0 ? 1 : 0
  }
}

// 关闭对话框
const handleDialogClose = () => {
  formRef.value?.resetFields()
}

onMounted(() => {
  loadGroupList()
  loadUserList()
})
</script>

<style scoped>
.admin-users {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.toolbar-left {
  display: flex;
  gap: 12px;
}

.toolbar-right {
  display: flex;
  gap: 12px;
}

.admin-table {
  flex: 1;
  overflow: auto;
}

.pagination {
  margin-top: 16px;
  justify-content: flex-end;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .toolbar {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .toolbar-left {
    flex-wrap: wrap;
  }

  .toolbar-right {
    width: 100%;
  }

  .toolbar-right .el-input {
    width: 100% !important;
  }

  .admin-table {
    font-size: 12px;
  }

  .admin-table :deep(.el-table__cell) {
    padding: 8px 4px;
  }

  .pagination {
    justify-content: center;
  }

  .pagination :deep(.el-pagination__sizes),
  .pagination :deep(.el-pagination__jump) {
    display: none;
  }
}

@media (max-width: 480px) {
  .toolbar-left .el-button {
    flex: 1;
    min-width: 0;
  }

  .admin-table :deep(.el-table__cell) {
    padding: 6px 2px;
    font-size: 11px;
  }

  .admin-table :deep(.el-table-column--selection) {
    width: 40px !important;
  }
}

/* 深色模式样式 */
html.dark .admin-users {
  background: transparent;
}

html.dark .pagination {
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

html.dark :deep(.el-input__wrapper) {
  background-color: var(--el-bg-color);
  border-color: var(--el-border-color);
}

html.dark :deep(.el-input__inner) {
  color: var(--el-text-color-primary);
}

html.dark :deep(.el-select .el-input__wrapper) {
  background-color: var(--el-bg-color);
  border-color: var(--el-border-color);
}

html.dark :deep(.el-input-number .el-input__wrapper) {
  background-color: var(--el-bg-color);
  border-color: var(--el-border-color);
}
</style>


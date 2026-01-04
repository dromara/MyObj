<template>
  <div class="admin-users">
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" icon="Plus" @click="handleCreate">新建用户</el-button>
        <el-button icon="Refresh" @click="loadUserList">刷新</el-button>
      </div>
      <div class="toolbar-right">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索用户名、邮箱..."
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
      empty-text="暂无用户"
    >
      <el-table-column prop="user_name" label="用户名" min-width="120" />
      <el-table-column prop="name" label="昵称" min-width="120" />
      <el-table-column prop="email" label="邮箱" min-width="180" />
      <el-table-column prop="phone" label="手机号" min-width="120" />
      <el-table-column prop="group_name" label="用户组" width="120">
        <template #default="{ row }">
          <el-tag>{{ row.group_name || `组${row.group_id}` }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="存储空间" width="150">
        <template #default="{ row }">
          {{ formatStorage(row.space) }}
        </template>
      </el-table-column>
      <el-table-column prop="state" label="状态" width="100" align="center">
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
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button 
            v-if="row.group_id !== 1" 
            link 
            type="primary" 
            @click="handleEdit(row)"
          >
            编辑
          </el-button>
          <el-button 
            v-if="row.group_id !== 1" 
            link 
            type="danger" 
            @click="handleDelete(row)"
          >
            删除
          </el-button>
          <span v-if="row.group_id === 1" style="color: var(--el-text-color-secondary); font-size: 12px;">
            管理员不可操作
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
        <el-form-item label="用户名" prop="user_name">
          <el-input v-model="formData.user_name" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!isEdit">
          <el-input v-model="formData.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="formData.name" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="formData.email" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="formData.phone" />
        </el-form-item>
        <el-form-item label="用户组" prop="group_id">
          <el-select v-model="formData.group_id" style="width: 100%">
            <el-option
              v-for="group in groupList"
              :key="group.id"
              :label="group.name"
              :value="group.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="存储空间(GB)" prop="space">
          <el-input-number 
            v-model="formData.space" 
            :min="0" 
            :max="maxSpaceInGB" 
            style="width: 100%" 
          />
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px;">
            0 表示无限空间
            <span v-if="maxSpaceInGB > 0 && maxSpaceInGB < 999999" style="color: var(--el-color-warning);">
              （组限制：{{ maxSpaceInGB }} GB）
            </span>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
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

const { proxy } = getCurrentInstance() as ComponentInternalInstance

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

const dialogTitle = computed(() => isEdit.value ? '编辑用户' : '新建用户')

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
  user_name: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
  group_id: [{ required: true, message: '请选择用户组', trigger: 'change' }],
  space: [{ required: true, message: '请输入存储空间', trigger: 'blur' }]
}

// 格式化存储空间
const formatStorage = (bytes: number) => {
  if (bytes === 0 || bytes === -1) return '无限'
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
      proxy?.$modal.msg('用户管理功能开发中')
      userList.value = []
      pagination.total = 0
    }
  } catch (error: any) {
    proxy?.$modal.msgError('加载用户列表失败')
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
    proxy?.$modal.msgWarning('不能编辑管理员组用户')
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
            proxy?.$modal.msgSuccess('更新成功')
            showDialog.value = false
            loadUserList()
          } else {
            proxy?.$modal.msgError(res.message || '更新失败')
          }
        } else {
          const res = await createAdminUser(submitData as CreateUserRequest)
          if (res.code === 200) {
            proxy?.$modal.msgSuccess('创建成功')
            showDialog.value = false
            loadUserList()
          } else {
            proxy?.$modal.msgError(res.message || '创建失败')
          }
        }
      } catch (error: any) {
        if (error.response?.status === 404 || error.message?.includes('404')) {
          proxy?.$modal.msg('用户管理功能开发中')
        } else {
          proxy?.$modal.msgError(error.message || '操作失败')
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
    proxy?.$modal.msgWarning('不能删除管理员组用户')
    return
  }
  
  try {
    await proxy?.$modal.confirm(`确定要删除用户 "${user.user_name}" 吗？`)
    try {
      const res = await deleteAdminUser(user.id)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess('删除成功')
        loadUserList()
      } else {
        proxy?.$modal.msgError(res.message || '删除失败')
      }
    } catch (error: any) {
      if (error.response?.status === 404 || error.message?.includes('404')) {
        proxy?.$modal.msg('用户管理功能开发中')
      } else {
        proxy?.$modal.msgError(error.message || '删除失败')
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
    proxy?.$modal.msgWarning('不能操作管理员组用户')
    // 回滚状态
    row.state = row.state === 0 ? 1 : 0
    return
  }
  
  const text = row.state === 0 ? '启用' : '禁用'
  try {
    await proxy?.$modal.confirm(`确认要"${text}""${row.user_name}"用户吗?`)
    try {
      const res = await toggleUserState(row.id, row.state)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(`${text}成功`)
      } else {
        proxy?.$modal.msgError(res.message || `${text}失败`)
        // 操作失败，回滚状态
        row.state = row.state === 0 ? 1 : 0
      }
    } catch (error: any) {
      // 操作失败，回滚状态
      row.state = row.state === 0 ? 1 : 0
      if (error.response?.status === 404 || error.message?.includes('404')) {
        proxy?.$modal.msg('用户管理功能开发中')
      } else {
        proxy?.$modal.msgError(error.message || `${text}失败`)
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
</style>


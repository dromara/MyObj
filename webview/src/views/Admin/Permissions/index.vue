<template>
  <div class="admin-permissions">
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" icon="Plus" @click="handleCreate">新建权限</el-button>
        <el-button 
          type="danger" 
          icon="Delete" 
          :disabled="selectedRows.length === 0"
          @click="handleBatchDelete"
        >
          批量删除 ({{ selectedRows.length }})
        </el-button>
        <el-button icon="Refresh" @click="loadPowerList">刷新</el-button>
      </div>
    </div>

    <el-table
      :data="powerList"
      v-loading="loading"
      class="admin-table"
      empty-text="暂无权限"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="55" />
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="权限名称" min-width="150" />
      <el-table-column prop="description" label="描述" min-width="200" />
      <el-table-column prop="characteristic" label="特征码" min-width="200">
        <template #default="{ row }">
          <code style="font-size: 12px; color: var(--el-color-primary);">{{ row.characteristic }}</code>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button 
            link 
            type="primary" 
            @click="handleEdit(row)"
          >
            编辑
          </el-button>
          <el-button 
            link 
            type="danger" 
            @click="handleDelete(row)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 创建/编辑权限对话框 -->
    <el-dialog
      v-model="showDialog"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="权限名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入权限名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input 
            v-model="formData.description" 
            type="textarea" 
            :rows="3"
            placeholder="请输入权限描述" 
          />
        </el-form-item>
        <el-form-item label="特征码" prop="characteristic">
          <el-input 
            v-model="formData.characteristic" 
            placeholder="请输入权限特征码" 
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
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

const { proxy } = getCurrentInstance() as ComponentInternalInstance

const loading = ref(false)
const powerList = ref<AdminPower[]>([])
const selectedRows = ref<AdminPower[]>([])
const showDialog = ref(false)
const dialogTitle = ref('新建权限')
const formRef = ref<FormInstance>()
const isEdit = ref(false)
const formData = ref<CreatePowerRequest & { id?: number }>({
  name: '',
  description: '',
  characteristic: ''
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入权限名称', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入权限描述', trigger: 'blur' }
  ],
  characteristic: [
    { required: true, message: '请输入权限特征码', trigger: 'blur' }
  ]
}

// 加载权限列表
const loadPowerList = async () => {
  loading.value = true
  try {
    const res = await getAdminPowerList()
    if (res.code === 200 && res.data) {
      powerList.value = res.data.powers || []
    } else {
      proxy?.$modal.msgError('加载权限列表失败')
      powerList.value = []
    }
  } catch (error: any) {
    proxy?.$modal.msgError('加载权限列表失败')
    proxy?.$log?.error(error)
  } finally {
    loading.value = false
  }
}

// 新建权限
const handleCreate = () => {
  isEdit.value = false
  dialogTitle.value = '新建权限'
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
  dialogTitle.value = '编辑权限'
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
  proxy?.$modal.confirm(`确定要删除权限 "${row.name}" 吗？`).then(async () => {
    try {
      const res = await deleteAdminPower(row.id)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess('删除成功')
        loadPowerList()
      } else {
        proxy?.$modal.msgError(res.message || '删除失败')
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.response?.data?.message || '删除失败')
      proxy?.$log?.error(error)
    }
  }).catch(() => {})
}

// 批量删除权限
const handleBatchDelete = () => {
  if (selectedRows.value.length === 0) {
    proxy?.$modal.msgWarning('请选择要删除的权限')
    return
  }

  const names = selectedRows.value.map(row => row.name).join('、')
  proxy?.$modal.confirm(`确定要删除选中的 ${selectedRows.value.length} 个权限吗？\n${names}`).then(async () => {
    try {
      const ids = selectedRows.value.map(row => row.id)
      const res = await batchDeleteAdminPower({ ids })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(res.message || '批量删除成功')
        selectedRows.value = []
        loadPowerList()
      } else {
        proxy?.$modal.msgError(res.message || '批量删除失败')
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.response?.data?.message || '批量删除失败')
      proxy?.$log?.error(error)
    }
  }).catch(() => {})
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return

    try {
      if (isEdit.value) {
        const res = await updateAdminPower(formData.value as UpdatePowerRequest)
        if (res.code === 200) {
          proxy?.$modal.msgSuccess('更新成功')
          showDialog.value = false
          loadPowerList()
        } else {
          proxy?.$modal.msgError(res.message || '更新失败')
        }
      } else {
        const res = await createAdminPower(formData.value as CreatePowerRequest)
        if (res.code === 200) {
          proxy?.$modal.msgSuccess('创建成功')
          showDialog.value = false
          loadPowerList()
        } else {
          proxy?.$modal.msgError(res.message || '创建失败')
        }
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.response?.data?.message || (isEdit.value ? '更新失败' : '创建失败'))
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
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.toolbar-left {
  display: flex;
  gap: 12px;
}

.admin-table {
  flex: 1;
  overflow: auto;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .toolbar {
    flex-wrap: wrap;
    gap: 8px;
  }

  .toolbar-left {
    flex: 1;
    min-width: 0;
  }

  .toolbar-left .el-button {
    flex: 1;
    min-width: 0;
  }

  .admin-table {
    font-size: 12px;
  }

  .admin-table :deep(.el-table__cell) {
    padding: 8px 4px;
  }
}

@media (max-width: 480px) {
  .admin-table :deep(.el-table__cell) {
    padding: 6px 2px;
    font-size: 11px;
  }
}
</style>

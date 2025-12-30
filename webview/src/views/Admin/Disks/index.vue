<template>
  <div class="admin-disks">
    <div class="toolbar">
      <el-button type="primary" icon="Plus" @click="handleCreate">添加磁盘</el-button>
      <el-button icon="Refresh" @click="loadDiskList">刷新</el-button>
    </div>

    <el-table
      :data="diskList"
      v-loading="loading"
      class="admin-table"
      empty-text="暂无磁盘"
    >
      <el-table-column prop="id" label="磁盘ID" min-width="200" />
      <el-table-column prop="disk_path" label="磁盘路径" min-width="250" />
      <el-table-column prop="data_path" label="数据路径" min-width="250" />
      <el-table-column label="大小" width="150">
        <template #default="{ row }">
          {{ formatStorage(row.size) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 创建/编辑磁盘对话框 -->
    <el-dialog
      v-model="showDialog"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="100px">
        <!-- 输入方式选择 -->
        <el-form-item label="输入方式" v-if="!isEdit">
          <el-radio-group v-model="inputMode" @change="handleInputModeChange">
            <el-radio-button value="manual">手动输入</el-radio-button>
            <el-radio-button value="scan">扫描选择</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <!-- 扫描磁盘选择 -->
        <template v-if="inputMode === 'scan' && !isEdit">
          <el-form-item label="选择磁盘" prop="selected_disk">
            <el-select
              v-model="formData.selected_disk"
              placeholder="请选择磁盘"
              style="width: 100%"
              @change="handleDiskSelect"
              :loading="scanLoading"
            >
              <el-option
                v-for="disk in scannedDisks"
                :key="disk.mount"
                :label="`${disk.mount} (${formatBytes(disk.total)})`"
                :value="disk.mount"
              >
                <div style="display: flex; justify-content: space-between; align-items: center;">
                  <span>{{ disk.mount }}</span>
                  <span style="color: var(--el-text-color-secondary); font-size: 12px;">
                    {{ formatBytes(disk.total) }} / 可用: {{ formatBytes(disk.avail) }}
                  </span>
                </div>
              </el-option>
            </el-select>
            <el-button
              type="primary"
              text
              style="margin-top: 8px;"
              @click="handleScanDisks"
              :loading="scanLoading"
            >
              <el-icon><Refresh /></el-icon>
              {{ scanLoading ? '扫描中...' : '重新扫描' }}
            </el-button>
          </el-form-item>
        </template>

        <!-- 手动输入或编辑时的表单 -->
        <el-form-item label="磁盘路径" prop="disk_path">
          <el-input
            v-model="formData.disk_path"
            placeholder="例如: D:\storage 或 /mnt/storage"
            :disabled="inputMode === 'scan' && !isEdit"
          />
        </el-form-item>
        <el-form-item label="数据路径" prop="data_path">
          <el-input
            v-model="formData.data_path"
            placeholder="例如: D:\storage\data 或 /mnt/storage/data"
          />
        </el-form-item>
        <el-form-item label="大小(GB)" prop="size">
          <el-input-number
            v-model="formData.size"
            :min="0"
            :max="999999"
            style="width: 100%"
            :disabled="inputMode === 'scan' && !isEdit && formData.selected_disk !== ''"
          />
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
  getAdminDiskList,
  createAdminDisk,
  updateAdminDisk,
  deleteAdminDisk,
  scanDisks,
  type AdminDisk,
  type ScannedDiskInfo
} from '@/api/admin'
import { formatSize, bytesToGB } from '@/utils'

const { proxy } = getCurrentInstance() as ComponentInternalInstance

const loading = ref(false)
const submitting = ref(false)
const scanLoading = ref(false)
const diskList = ref<AdminDisk[]>([])
const scannedDisks = ref<ScannedDiskInfo[]>([])
const showDialog = ref(false)
const isEdit = ref(false)
const inputMode = ref<'manual' | 'scan'>('manual')
const formRef = ref()
const formData = reactive({
  id: '',
  disk_path: '',
  data_path: '',
  size: 0,
  selected_disk: ''
})

const dialogTitle = computed(() => isEdit.value ? '编辑磁盘' : '添加磁盘')

const formRules = {
  disk_path: [{ required: true, message: '请输入磁盘路径', trigger: 'blur' }],
  data_path: [{ required: true, message: '请输入数据路径', trigger: 'blur' }],
  size: [{ required: true, message: '请输入磁盘大小', trigger: 'blur' }],
  selected_disk: [
    {
      validator: (_rule: any, value: any, callback: any) => {
        if (inputMode.value === 'scan' && !value) {
          callback(new Error('请选择磁盘'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ]
}

// 格式化存储空间（后端返回的是字节，需要转换为GB显示）
const formatStorage = (bytes: number) => {
  if (bytes === 0) return '未设置'
  // return formatSize(bytes)
  return bytes + 'GB'
}

// 格式化字节（用于扫描磁盘显示）
const formatBytes = (bytes: number) => {
  return formatSize(bytes)
}

// 加载磁盘列表
const loadDiskList = async () => {
  loading.value = true
  try {
    const res = await getAdminDiskList()
    if (res.code === 200 && res.data) {
      diskList.value = res.data.disks || []
    } else {
      proxy?.$modal.msg('磁盘管理功能开发中')
      diskList.value = []
    }
  } catch (error: any) {
    if (error.response?.status === 404 || error.message?.includes('404')) {
      proxy?.$modal.msg('磁盘管理功能开发中')
    } else {
      proxy?.$modal.msgError('加载磁盘列表失败')
    }
    proxy?.$log?.error(error)
  } finally {
    loading.value = false
  }
}

// 创建磁盘
const handleCreate = () => {
  isEdit.value = false
  inputMode.value = 'manual'
  Object.assign(formData, {
    id: '',
    disk_path: '',
    data_path: '',
    size: 0,
    selected_disk: ''
  })
  scannedDisks.value = []
  showDialog.value = true
}

// 扫描磁盘
const handleScanDisks = async () => {
  scanLoading.value = true
  try {
    const res = await scanDisks()
    if (res.code === 200 && res.data) {
      scannedDisks.value = res.data
      if (scannedDisks.value.length === 0) {
        proxy?.$modal.msgWarning('未扫描到可用磁盘')
      }
    } else {
      proxy?.$modal.msgError(res.message || '扫描磁盘失败')
    }
  } catch (error: any) {
    proxy?.$modal.msgError('扫描磁盘失败')
    proxy?.$log?.error(error)
  } finally {
    scanLoading.value = false
  }
}

// 输入方式改变
const handleInputModeChange = (mode: string | number | boolean | undefined) => {
  const inputType = mode as 'manual' | 'scan'
  if (inputType === 'scan' && scannedDisks.value.length === 0) {
    handleScanDisks()
  }
  // 清空表单数据
  formData.disk_path = ''
  formData.data_path = ''
  formData.size = 0
  formData.selected_disk = ''
}

// 磁盘选择改变
const handleDiskSelect = (mount: string) => {
  const selectedDisk = scannedDisks.value.find(d => d.mount === mount)
  if (selectedDisk) {
    formData.disk_path = selectedDisk.mount
    // 自动生成数据路径
    const dataPath = selectedDisk.mount.endsWith('/') || selectedDisk.mount.endsWith('\\')
      ? `${selectedDisk.mount}data`
      : `${selectedDisk.mount}/data`
    formData.data_path = dataPath
    // 将字节转换为GB（向下取整）
    formData.size = Math.floor(selectedDisk.total / (1024 * 1024 * 1024))
  }
}

// 编辑磁盘
const handleEdit = (disk: AdminDisk) => {
  isEdit.value = true
  // 后端返回的 size 是字节，需要转换为 GB 用于表单输入
  Object.assign(formData, {
    id: disk.id,
    disk_path: disk.disk_path,
    data_path: disk.data_path,
    size: bytesToGB(disk.size) // 将字节转换为 GB
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
        // 前端输入的是 GB，后端期望的也是 GB（后端会转换为字节）
        const submitData = {
          ...formData,
          size: formData.size // 保持 GB，后端会转换
        }
        if (isEdit.value) {
          const res = await updateAdminDisk(submitData)
          if (res.code === 200) {
            proxy?.$modal.msgSuccess('更新成功')
            showDialog.value = false
            loadDiskList()
          } else {
            proxy?.$modal.msgError(res.message || '更新失败')
          }
        } else {
          const res = await createAdminDisk(submitData)
          if (res.code === 200) {
            proxy?.$modal.msgSuccess('创建成功')
            showDialog.value = false
            loadDiskList()
          } else {
            proxy?.$modal.msgError(res.message || '创建失败')
          }
        }
      } catch (error: any) {
        if (error.response?.status === 404 || error.message?.includes('404')) {
          proxy?.$modal.msg('磁盘管理功能开发中')
        } else {
          proxy?.$modal.msgError(error.message || '操作失败')
        }
      } finally {
        submitting.value = false
      }
    }
  })
}

// 删除磁盘
const handleDelete = async (disk: AdminDisk) => {
  try {
    await proxy?.$modal.confirm(`确定要删除磁盘 "${disk.disk_path}" 吗？`)
    try {
      const res = await deleteAdminDisk(disk.id)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess('删除成功')
        loadDiskList()
      } else {
        proxy?.$modal.msgError(res.message || '删除失败')
      }
    } catch (error: any) {
      if (error.response?.status === 404 || error.message?.includes('404')) {
        proxy?.$modal.msg('磁盘管理功能开发中')
      } else {
        proxy?.$modal.msgError(error.message || '删除失败')
      }
    }
  } catch (error: any) {
    // 用户取消
  }
}

// 关闭对话框
const handleDialogClose = () => {
  formRef.value?.resetFields()
  scannedDisks.value = []
  inputMode.value = 'manual'
}

onMounted(() => {
  loadDiskList()
})
</script>

<style scoped>
.admin-disks {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.toolbar {
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

  .toolbar .el-button {
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


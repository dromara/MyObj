<template>
  <div class="admin-groups">
    <div class="toolbar">
      <el-button type="primary" icon="Plus" @click="handleCreate">新建组</el-button>
      <el-button icon="Refresh" @click="loadGroupList">刷新</el-button>
    </div>

    <el-table
      :data="groupList"
      v-loading="loading"
      class="admin-table"
      empty-text="暂无组"
    >
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="组名称" min-width="150" />
      <el-table-column label="默认组" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="row.group_default === 1 ? 'success' : 'info'">
            {{ row.group_default === 1 ? '是' : '否' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="存储空间" width="150">
        <template #default="{ row }">
          {{ formatStorage(row.space) }}
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <template v-if="row.id !== 1">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="primary" @click="handleAssignPower(row)">分配权限</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
          <span v-else style="color: var(--el-text-color-secondary); font-size: 12px;">
            管理员不可操作
          </span>
        </template>
      </el-table-column>
    </el-table>

    <!-- 创建/编辑组对话框 -->
    <el-dialog
      v-model="showDialog"
      :title="dialogTitle"
      width="500px"
      @close="handleDialogClose"
    >
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="组名称" prop="name">
          <el-input v-model="formData.name" />
        </el-form-item>
        <el-form-item label="存储空间(GB)" prop="space">
          <el-input-number v-model="formData.space" :min="0" :max="999999" style="width: 100%" />
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px;">
            0 表示无限空间
          </div>
        </el-form-item>
        <el-form-item label="默认组">
          <el-switch v-model="formData.group_default" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 分配权限抽屉 -->
    <el-drawer
      v-model="showPowerDrawer"
      :size="drawerSize"
      :with-header="true"
      direction="rtl"
    >
      <template #header>
        <div class="drawer-header">
          <div class="drawer-title">
            <el-icon><Key /></el-icon>
            <span>为 "{{ currentGroup?.name }}" 分配权限</span>
          </div>
        </div>
      </template>

      <div class="power-drawer-content">
        <!-- 搜索和全选 -->
        <div class="power-toolbar">
          <el-input
            v-model="powerSearchKeyword"
            placeholder="搜索权限名称或特征码..."
            clearable
            style="flex: 1; margin-right: 12px;"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-button 
            :icon="isAllSelected ? 'Select' : 'Close'"
            @click="handleToggleSelectAll"
          >
            {{ isAllSelected ? '取消全选' : '全选' }}
          </el-button>
        </div>

        <!-- 权限分类列表 -->
        <div class="power-categories">
          <div
            v-for="(category, categoryKey) in categorizedPowers"
            :key="categoryKey"
            class="power-category"
          >
            <div class="category-header" @click="toggleCategory(categoryKey)">
              <el-icon class="category-icon" :class="{ 'is-expanded': expandedCategories[categoryKey] }">
                <ArrowRight />
              </el-icon>
              <span class="category-name">{{ category.name }}</span>
              <el-tag size="small" type="info" class="category-count">
                {{ category.powers.length }}
              </el-tag>
              <el-checkbox
                :model-value="isCategorySelected(categoryKey)"
                :indeterminate="isCategoryIndeterminate(categoryKey)"
                @change="(val) => handleCategorySelect(categoryKey, val as boolean | string | number)"
                @click.stop
                class="category-checkbox"
              />
            </div>
            <el-collapse-transition>
              <div v-show="expandedCategories[categoryKey]" class="category-powers">
                <el-checkbox-group v-model="selectedPowerIds">
                  <div
                    v-for="power in category.powers"
                    :key="power.id"
                    class="power-item"
                  >
                    <el-checkbox :label="power.id">
                      <div class="power-content">
                        <div class="power-name">{{ power.name }}</div>
                        <div class="power-description">{{ power.description || '暂无描述' }}</div>
                        <div class="power-characteristic">
                          <el-icon><Key /></el-icon>
                          <code>{{ power.characteristic }}</code>
                        </div>
                      </div>
                    </el-checkbox>
                  </div>
                </el-checkbox-group>
              </div>
            </el-collapse-transition>
          </div>
        </div>

        <!-- 已选权限统计 -->
        <div class="power-summary">
          <el-text type="info">
            已选择 <strong>{{ selectedPowerIds.length }}</strong> / {{ totalPowersCount }} 个权限
          </el-text>
        </div>
      </div>
      <template #footer>
        <el-button @click="showPowerDrawer = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleAssignPowerSubmit">确定</el-button>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import type { ComponentInternalInstance } from 'vue'
import {
  getAdminGroupList,
  createAdminGroup,
  updateAdminGroup,
  deleteAdminGroup,
  getAdminPowerList,
  getGroupPowers,
  assignPowerToGroup,
  type AdminGroup,
  type AdminPower
} from '@/api/admin'
import { bytesToGB, GBToBytes } from '@/utils'

const { proxy } = getCurrentInstance() as ComponentInternalInstance

const loading = ref(false)
const submitting = ref(false)
const groupList = ref<AdminGroup[]>([])
const powerList = ref<AdminPower[]>([])
const showDialog = ref(false)
const showPowerDrawer = ref(false)
const isEdit = ref(false)
const currentGroup = ref<AdminGroup | null>(null)
const selectedPowerIds = ref<number[]>([])
const powerSearchKeyword = ref('')
const expandedCategories = ref<Record<string, boolean>>({})
const formRef = ref()
const formData = reactive({
  id: 0,
  name: '',
  space: 0,
  group_default: 0
})

// 权限分类映射（根据特征码前缀）
const categoryMap: Record<string, string> = {
  'file:': '文件管理',
  'dir:': '目录管理',
  'user:': '用户管理',
  'share:': '分享管理',
  'download:': '下载管理',
  'offline:': '离线下载',
  'admin:': '系统管理',
  'recycled:': '回收站'
}

// 计算抽屉大小（响应式）
const drawerSize = computed(() => {
  if (window.innerWidth <= 768) {
    return '90%'
  }
  if (window.innerWidth <= 1024) {
    return '70%'
  }
  return '600px'
})

// 过滤后的权限列表
const filteredPowerList = computed(() => {
  if (!powerSearchKeyword.value) {
    return powerList.value
  }
  const keyword = powerSearchKeyword.value.toLowerCase()
  return powerList.value.filter(power => 
    power.name.toLowerCase().includes(keyword) ||
    power.characteristic.toLowerCase().includes(keyword) ||
    (power.description && power.description.toLowerCase().includes(keyword))
  )
})

// 归类后的权限
const categorizedPowers = computed(() => {
  const categories: Record<string, { name: string; powers: AdminPower[] }> = {}
  
  filteredPowerList.value.forEach(power => {
    // 提取特征码前缀（如 "file:upload" -> "file:"）
    const prefix = power.characteristic.split(':')[0] + ':'
    const categoryName = categoryMap[prefix] || '其他权限'
    
    if (!categories[categoryName]) {
      categories[categoryName] = {
        name: categoryName,
        powers: []
      }
    }
    categories[categoryName].powers.push(power)
  })
  
  // 按分类名称排序
  const sortedCategories: Record<string, { name: string; powers: AdminPower[] }> = {}
  Object.keys(categories).sort().forEach(key => {
    sortedCategories[key] = categories[key]
    // 每个分类内的权限按ID排序
    sortedCategories[key].powers.sort((a, b) => a.id - b.id)
  })
  
  return sortedCategories
})

// 总权限数
const totalPowersCount = computed(() => filteredPowerList.value.length)

// 是否全选
const isAllSelected = computed(() => {
  if (totalPowersCount.value === 0) return false
  return selectedPowerIds.value.length === totalPowersCount.value
})

// 切换全选
const handleToggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedPowerIds.value = []
  } else {
    selectedPowerIds.value = filteredPowerList.value.map(p => p.id)
  }
}

// 切换分类展开/收起
const toggleCategory = (categoryKey: string) => {
  expandedCategories.value[categoryKey] = !expandedCategories.value[categoryKey]
}

// 检查分类是否全选
const isCategorySelected = (categoryKey: string) => {
  const category = categorizedPowers.value[categoryKey]
  if (!category || category.powers.length === 0) return false
  return category.powers.every(power => selectedPowerIds.value.includes(power.id))
}

// 检查分类是否半选（部分选中）
const isCategoryIndeterminate = (categoryKey: string) => {
  const category = categorizedPowers.value[categoryKey]
  if (!category || category.powers.length === 0) return false
  const selectedCount = category.powers.filter(power => selectedPowerIds.value.includes(power.id)).length
  return selectedCount > 0 && selectedCount < category.powers.length
}

// 分类全选/取消全选
const handleCategorySelect = (categoryKey: string, checked: boolean | string | number) => {
  const isChecked = checked === true || checked === 1 || checked === '1'
  const category = categorizedPowers.value[categoryKey]
  if (!category) return
  
  if (isChecked) {
    // 全选该分类
    category.powers.forEach(power => {
      if (!selectedPowerIds.value.includes(power.id)) {
        selectedPowerIds.value.push(power.id)
      }
    })
  } else {
    // 取消全选该分类
    category.powers.forEach(power => {
      const index = selectedPowerIds.value.indexOf(power.id)
      if (index > -1) {
        selectedPowerIds.value.splice(index, 1)
      }
    })
  }
}

const dialogTitle = computed(() => isEdit.value ? '编辑组' : '新建组')

const formRules = {
  name: [{ required: true, message: '请输入组名称', trigger: 'blur' }],
  space: [{ required: true, message: '请输入存储空间', trigger: 'blur' }]
}

// 格式化存储空间
const formatStorage = (bytes: number) => {
  if (bytes === 0 || bytes === -1) return '无限'
  return bytesToGB(bytes) + ' GB'
}

// 加载组列表
const loadGroupList = async () => {
  loading.value = true
  try {
    const res = await getAdminGroupList()
    if (res.code === 200 && res.data) {
      groupList.value = res.data.groups || []
    } else {
      proxy?.$modal.msg('组管理功能开发中')
      groupList.value = []
    }
  } catch (error: any) {
    if (error.response?.status === 404 || error.message?.includes('404')) {
      proxy?.$modal.msg('组管理功能开发中')
    } else {
      proxy?.$modal.msgError('加载组列表失败')
    }
    proxy?.$log?.error(error)
  } finally {
    loading.value = false
  }
}

// 加载权限列表
const loadPowerList = async () => {
  try {
    const res = await getAdminPowerList()
    if (res.code === 200 && res.data) {
      powerList.value = res.data.powers || []
    } else {
      proxy?.$modal.msg('权限管理功能开发中')
      powerList.value = []
    }
  } catch (error: any) {
    proxy?.$log?.error(error)
  }
}

// 创建组
const handleCreate = () => {
  isEdit.value = false
  Object.assign(formData, {
    id: 0,
    name: '',
    space: 0,
    group_default: 0
  })
  showDialog.value = true
}

// 编辑组
const handleEdit = (group: AdminGroup) => {
  // 禁止操作管理员组
  if (group.id === 1) {
    proxy?.$modal.msgWarning('不能编辑管理员组')
    return
  }
  isEdit.value = true
  // 将字节转换为 GB（后端存储的是字节，前端表单输入的是 GB）
  Object.assign(formData, {
    id: group.id,
    name: group.name,
    space: bytesToGB(group.space),
    group_default: group.group_default
  })
  showDialog.value = true
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (valid) {
      // 禁止编辑管理员组
      if (isEdit.value && formData.id === 1) {
        proxy?.$modal.msgWarning('不能编辑管理员组')
        return
      }
      submitting.value = true
      try {
        // 将 GB 转换为字节（前端输入的是 GB，后端存储的是字节）
        const submitData = {
          ...formData,
          space: GBToBytes(formData.space)
        }
        
        if (isEdit.value) {
          const res = await updateAdminGroup(submitData)
          if (res.code === 200) {
            proxy?.$modal.msgSuccess('更新成功')
            showDialog.value = false
            loadGroupList()
          } else {
            proxy?.$modal.msgError(res.message || '更新失败')
          }
        } else {
          const res = await createAdminGroup(submitData)
          if (res.code === 200) {
            proxy?.$modal.msgSuccess('创建成功')
            showDialog.value = false
            loadGroupList()
          } else {
            proxy?.$modal.msgError(res.message || '创建失败')
          }
        }
      } catch (error: any) {
        if (error.response?.status === 404 || error.message?.includes('404')) {
          proxy?.$modal.msg('组管理功能开发中')
        } else {
          proxy?.$modal.msgError(error.message || '操作失败')
        }
      } finally {
        submitting.value = false
      }
    }
  })
}

// 删除组
const handleDelete = async (group: AdminGroup) => {
  // 禁止操作管理员组
  if (group.id === 1) {
    proxy?.$modal.msgWarning('不能删除管理员组')
    return
  }
  try {
    await proxy?.$modal.confirm(`确定要删除组 "${group.name}" 吗？`)
    try {
      const res = await deleteAdminGroup(group.id)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess('删除成功')
        loadGroupList()
      } else {
        proxy?.$modal.msgError(res.message || '删除失败')
      }
    } catch (error: any) {
      if (error.response?.status === 404 || error.message?.includes('404')) {
        proxy?.$modal.msg('组管理功能开发中')
      } else {
        proxy?.$modal.msgError(error.message || '删除失败')
      }
    }
  } catch (error: any) {
    // 用户取消
  }
}

// 分配权限
const handleAssignPower = async (group: AdminGroup) => {
  // 禁止操作管理员组
  if (group.id === 1) {
    proxy?.$modal.msgWarning('不能修改管理员组权限')
    return
  }
  currentGroup.value = group
  powerSearchKeyword.value = ''
  
  // 初始化所有分类为展开状态
  Object.keys(categorizedPowers.value).forEach(key => {
    expandedCategories.value[key] = true
  })
  
  try {
    const res = await getGroupPowers(group.id)
    if (res.code === 200 && res.data) {
      selectedPowerIds.value = res.data.power_ids || []
    } else {
      selectedPowerIds.value = []
    }
  } catch (error: any) {
    selectedPowerIds.value = []
  }
  showPowerDrawer.value = true
}

// 提交权限分配
const handleAssignPowerSubmit = async () => {
  if (!currentGroup.value) return
  submitting.value = true
  try {
    const res = await assignPowerToGroup({
      group_id: currentGroup.value.id,
      power_ids: selectedPowerIds.value
    })
    if (res.code === 200) {
      proxy?.$modal.msgSuccess('权限分配成功')
      showPowerDrawer.value = false
    } else {
      proxy?.$modal.msgError(res.message || '权限分配失败')
    }
  } catch (error: any) {
    if (error.response?.status === 404 || error.message?.includes('404')) {
      proxy?.$modal.msg('权限管理功能开发中')
    } else {
      proxy?.$modal.msgError(error.message || '权限分配失败')
    }
  } finally {
    submitting.value = false
  }
}

// 关闭对话框
const handleDialogClose = () => {
  formRef.value?.resetFields()
}

onMounted(() => {
  loadPowerList()
  loadGroupList()
})
</script>

<style scoped>
.admin-groups {
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

  .admin-table :deep(.el-table-column--selection) {
    width: 40px !important;
  }
}

/* 权限分配抽屉样式 */
.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.drawer-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.power-drawer-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 16px;
}

.power-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.power-categories {
  flex: 1;
  overflow-y: auto;
  padding-right: 8px;
}

.power-category {
  margin-bottom: 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  overflow: hidden;
  background: var(--el-bg-color-page);
}

.category-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  background: var(--el-bg-color);
  transition: background 0.2s;
  user-select: none;
}

.category-header:hover {
  background: var(--el-fill-color-light);
}

.category-icon {
  margin-right: 8px;
  transition: transform 0.3s;
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.category-icon.is-expanded {
  transform: rotate(90deg);
}

.category-name {
  flex: 1;
  font-weight: 500;
  font-size: 14px;
  color: var(--el-text-color-primary);
}

.category-count {
  margin-right: 12px;
}

.category-checkbox {
  margin-left: auto;
}

.category-powers {
  padding: 8px 16px 16px;
  background: var(--el-bg-color-page);
}

.power-item {
  padding: 12px;
  margin-bottom: 8px;
  border-radius: 4px;
  background: var(--el-bg-color);
  transition: background 0.2s;
}

.power-item:hover {
  background: var(--el-fill-color-light);
}

.power-item :deep(.el-checkbox) {
  width: 100%;
  align-items: flex-start;
}

.power-item :deep(.el-checkbox__label) {
  width: 100%;
  padding-left: 8px;
}

.power-content {
  width: 100%;
}

.power-name {
  font-weight: 500;
  font-size: 14px;
  color: var(--el-text-color-primary);
  margin-bottom: 4px;
}

.power-description {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 6px;
  line-height: 1.4;
}

.power-characteristic {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--el-text-color-placeholder);
}

.power-characteristic code {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  background: var(--el-fill-color-light);
  padding: 2px 6px;
  border-radius: 3px;
  color: var(--el-color-primary);
}

.power-summary {
  padding: 12px;
  text-align: center;
  border-top: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
  border-radius: 4px;
}

.drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .power-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .power-toolbar .el-input {
    margin-right: 0 !important;
    margin-bottom: 8px;
  }

  .power-toolbar .el-button {
    width: 100%;
  }

  .category-header {
    padding: 10px 12px;
  }

  .category-powers {
    padding: 8px 12px 12px;
  }

  .power-item {
    padding: 10px;
  }

  .power-name {
    font-size: 13px;
  }

  .power-description {
    font-size: 11px;
  }

  .power-characteristic {
    font-size: 10px;
  }
}
</style>


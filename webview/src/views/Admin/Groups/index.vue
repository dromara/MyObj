<template>
  <div class="admin-groups">
    <div class="toolbar">
      <el-button type="primary" icon="Plus" @click="handleCreate">{{ t('admin.groups.newGroup') }}</el-button>
      <el-button icon="Refresh" @click="loadGroupList">{{ t('common.refresh') }}</el-button>
    </div>

    <el-table :data="groupList" v-loading="loading" class="admin-table" :empty-text="t('admin.groups.noGroups')">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" :label="t('admin.groups.groupName')" min-width="150" />
      <el-table-column :label="t('admin.groups.defaultGroup')" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="row.group_default === 1 ? 'success' : 'info'">
            {{ row.group_default === 1 ? t('admin.groups.yes') : t('admin.groups.no') }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('admin.users.storageSpace')" width="150">
        <template #default="{ row }">
          {{ formatStorage(row.space) }}
        </template>
      </el-table-column>
      <el-table-column prop="created_at" :label="t('admin.users.createTime')" width="180" />
      <el-table-column :label="t('admin.users.operation')" width="200" fixed="right">
        <template #default="{ row }">
          <template v-if="row.id !== 1">
            <el-button link type="primary" @click="handleEdit(row)">{{ t('admin.users.edit') }}</el-button>
            <el-button link type="primary" @click="handleAssignPower(row)">{{
              t('admin.groups.assignPower')
            }}</el-button>
            <el-button link type="danger" @click="handleDelete(row)">{{ t('admin.users.delete') }}</el-button>
          </template>
          <span v-else style="color: var(--el-text-color-secondary); font-size: 12px">
            {{ t('admin.users.adminCannotOperate') }}
          </span>
        </template>
      </el-table-column>
    </el-table>

    <!-- 创建/编辑组对话框 -->
    <el-dialog v-model="showDialog" :title="dialogTitle" width="500px" @close="handleDialogClose">
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item :label="t('admin.groups.groupName')" prop="name">
          <el-input v-model="formData.name" />
        </el-form-item>
        <el-form-item :label="t('admin.users.storageSpaceGB')" prop="space">
          <el-input-number v-model="formData.space" :min="0" :max="999999" style="width: 100%" />
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px">
            {{ t('admin.users.unlimitedSpace') }}
          </div>
        </el-form-item>
        <el-form-item :label="t('admin.groups.defaultGroup')">
          <el-switch v-model="formData.group_default" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 分配权限抽屉 -->
    <el-drawer v-model="showPowerDrawer" :size="drawerSize" :with-header="true" direction="rtl">
      <template #header>
        <div class="drawer-header">
          <div class="drawer-title">
            <el-icon><Key /></el-icon>
            <span>{{ t('admin.groups.assignPowerFor', { name: currentGroup?.name || '' }) }}</span>
          </div>
        </div>
      </template>

      <div class="power-drawer-content">
        <!-- 搜索和全选 -->
        <div class="power-toolbar">
          <el-input
            v-model="powerSearchKeyword"
            :placeholder="t('admin.groups.searchPower')"
            clearable
            style="flex: 1; margin-right: 12px"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-button :icon="isAllSelected ? 'Select' : 'Close'" @click="handleToggleSelectAll">
            {{ isAllSelected ? t('admin.groups.cancelSelectAll') : t('admin.groups.selectAll') }}
          </el-button>
        </div>

        <!-- 权限分类列表 -->
        <div class="power-categories">
          <div v-for="(category, categoryKey) in categorizedPowers" :key="categoryKey" class="power-category">
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
                @change="val => handleCategorySelect(categoryKey, val as boolean | string | number)"
                @click.stop
                class="category-checkbox"
              />
            </div>
            <el-collapse-transition>
              <div v-show="expandedCategories[categoryKey]" class="category-powers">
                <el-checkbox-group v-model="selectedPowerIds">
                  <div v-for="power in category.powers" :key="power.id" class="power-item">
                    <el-checkbox :label="power.id">
                      <div class="power-content">
                        <div class="power-name">{{ getPermissionName(power.characteristic, power.name) }}</div>
                        <div class="power-description">
                          {{
                            getPermissionDescription(power.characteristic, power.description) ||
                            t('admin.groups.noDescription')
                          }}
                        </div>
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
            {{ t('admin.groups.selectedPowers', { count: selectedPowerIds.length, total: totalPowersCount }) }}
          </el-text>
        </div>
      </div>
      <template #footer>
        <el-button @click="showPowerDrawer = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="handleAssignPowerSubmit">{{
          t('common.confirm')
        }}</el-button>
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
  import { useI18n } from '@/composables'
  import { getPermissionName, getPermissionDescription } from '@/utils/business/permission'

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

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
    'file:': t('route.files'),
    'dir:': t('route.files'),
    'user:': t('route.adminUsers'),
    'share:': t('route.shares'),
    'download:': t('route.tasks'),
    'offline:': t('route.offline'),
    'admin:': t('route.admin'),
    'recycled:': t('route.trash')
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
    return powerList.value.filter(power => {
      // 搜索原始名称和描述
      const matchesOriginal =
        power.name.toLowerCase().includes(keyword) ||
        power.characteristic.toLowerCase().includes(keyword) ||
        (power.description && power.description.toLowerCase().includes(keyword))

      // 搜索国际化后的名称和描述
      const i18nName = getPermissionName(power.characteristic, power.name).toLowerCase()
      const i18nDescription = getPermissionDescription(power.characteristic, power.description || '').toLowerCase()
      const matchesI18n = i18nName.includes(keyword) || i18nDescription.includes(keyword)

      return matchesOriginal || matchesI18n
    })
  })

  // 归类后的权限
  const categorizedPowers = computed(() => {
    const categories: Record<string, { name: string; powers: AdminPower[] }> = {}

    filteredPowerList.value.forEach(power => {
      // 提取特征码前缀（如 "file:upload" -> "file:"）
      const prefix = power.characteristic.split(':')[0] + ':'
      const categoryName = categoryMap[prefix] || t('admin.groups.otherPowers')

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
    Object.keys(categories)
      .sort()
      .forEach(key => {
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

  const dialogTitle = computed(() => (isEdit.value ? t('admin.groups.editGroup') : t('admin.groups.newGroup')))

  const formRules = {
    name: [{ required: true, message: t('admin.groups.groupNameRequired'), trigger: 'blur' }],
    space: [{ required: true, message: t('admin.users.spaceRequired'), trigger: 'blur' }]
  }

  // 格式化存储空间
  const formatStorage = (bytes: number) => {
    if (bytes === 0 || bytes === -1) return t('admin.groups.unlimited')
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
        proxy?.$modal.msg(t('admin.groups.featureDeveloping'))
        groupList.value = []
      }
    } catch (error: any) {
      if (error.response?.status === 404 || error.message?.includes('404')) {
        proxy?.$modal.msg(t('admin.groups.featureDeveloping'))
      } else {
        proxy?.$modal.msgError(t('admin.groups.loadListFailed'))
      }
      proxy?.$log?.error(error)
    } finally {
      loading.value = false
    }
  }

  // 加载权限列表
  const loadPowerList = async () => {
    try {
      // 获取所有权限（传递足够大的 pageSize 以获取全部权限）
      const res = await getAdminPowerList({ page: 1, pageSize: 1000 })
      if (res.code === 200 && res.data) {
        powerList.value = res.data.powers || []
      } else {
        proxy?.$modal.msg(t('admin.groups.featureDeveloping'))
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
      proxy?.$modal.msgWarning(t('admin.groups.cannotEditAdmin'))
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
          proxy?.$modal.msgWarning(t('admin.groups.cannotEditAdmin'))
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
              proxy?.$modal.msgSuccess(t('admin.users.updateSuccess'))
              showDialog.value = false
              loadGroupList()
            } else {
              proxy?.$modal.msgError(res.message || t('admin.users.updateFailed'))
            }
          } else {
            const res = await createAdminGroup(submitData)
            if (res.code === 200) {
              proxy?.$modal.msgSuccess(t('admin.users.createSuccess'))
              showDialog.value = false
              loadGroupList()
            } else {
              proxy?.$modal.msgError(res.message || t('admin.users.createFailed'))
            }
          }
        } catch (error: any) {
          if (error.response?.status === 404 || error.message?.includes('404')) {
            proxy?.$modal.msg(t('admin.groups.featureDeveloping'))
          } else {
            proxy?.$modal.msgError(error.message || t('common.operationFailed'))
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
      proxy?.$modal.msgWarning(t('admin.groups.cannotDeleteAdmin'))
      return
    }
    try {
      await proxy?.$modal.confirm(t('admin.groups.confirmDelete', { name: group.name }))
      try {
        const res = await deleteAdminGroup(group.id)
        if (res.code === 200) {
          proxy?.$modal.msgSuccess(t('admin.users.deleteSuccess'))
          loadGroupList()
        } else {
          proxy?.$modal.msgError(res.message || t('admin.users.deleteFailed'))
        }
      } catch (error: any) {
        if (error.response?.status === 404 || error.message?.includes('404')) {
          proxy?.$modal.msg(t('admin.groups.featureDeveloping'))
        } else {
          proxy?.$modal.msgError(error.message || t('admin.users.deleteFailed'))
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
      proxy?.$modal.msgWarning(t('admin.groups.cannotModifyAdminPower'))
      return
    }
    currentGroup.value = group
    powerSearchKeyword.value = ''

    // 初始化所有分类为展开状态
    Object.keys(categorizedPowers.value).forEach(key => {
      expandedCategories.value[key] = true
    })

    // 刷新权限列表（确保获取最新的所有权限）
    await loadPowerList()

    // 获取该组已拥有的权限ID
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
        proxy?.$modal.msgSuccess(t('admin.groups.assignPowerSuccess'))
        showPowerDrawer.value = false
      } else {
        proxy?.$modal.msgError(res.message || t('admin.groups.assignPowerFailed'))
      }
    } catch (error: any) {
      if (error.response?.status === 404 || error.message?.includes('404')) {
        proxy?.$modal.msg(t('admin.groups.featureDeveloping'))
      } else {
        proxy?.$modal.msgError(error.message || t('admin.groups.assignPowerFailed'))
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

  /* 深色模式样式 */
  html.dark .admin-groups {
    background: transparent;
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

  html.dark :deep(.el-drawer) {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark :deep(.el-drawer__header) {
    background: var(--card-bg);
    border-bottom-color: var(--el-border-color);
  }

  html.dark :deep(.el-drawer__title) {
    color: var(--el-text-color-primary);
  }

  html.dark :deep(.el-drawer__body) {
    background: var(--card-bg);
    color: var(--el-text-color-primary);
  }

  html.dark .power-toolbar {
    border-bottom-color: var(--el-border-color);
  }

  html.dark .power-category {
    border-color: var(--el-border-color);
    background: var(--el-bg-color-page);
  }

  html.dark .category-header {
    background: var(--el-bg-color);
    color: var(--el-text-color-primary);
  }

  html.dark .category-header:hover {
    background: var(--el-fill-color-light);
  }

  html.dark .category-icon {
    color: var(--el-text-color-secondary);
  }

  html.dark .category-name {
    color: var(--el-text-color-primary);
  }

  html.dark .category-powers {
    background: var(--el-bg-color-page);
  }

  html.dark .power-item {
    background: var(--el-bg-color);
  }

  html.dark .power-item:hover {
    background: var(--el-fill-color-light);
  }

  html.dark .power-name {
    color: var(--el-text-color-primary);
  }

  html.dark .power-description {
    color: var(--el-text-color-secondary);
  }

  html.dark .power-characteristic {
    color: var(--el-text-color-placeholder);
  }

  html.dark .power-characteristic code {
    background: var(--el-fill-color-light);
    color: var(--el-color-primary);
  }

  html.dark .power-summary {
    border-top-color: var(--el-border-color);
    background: var(--el-bg-color);
  }

  html.dark .drawer-title {
    color: var(--el-text-color-primary);
  }
</style>

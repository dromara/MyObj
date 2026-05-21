<template>
  <div class="enterprise-roles page-card">
    <div class="section-header">
      <h3><el-icon><Key /></el-icon> {{ t('enterprise.role.title') }}</h3>
      <el-button class="action-btn" icon="Plus" @click="showRoleDialog = true; resetRoleForm()">
        {{ t('enterprise.role.create') }}
      </el-button>
    </div>

    <el-table :data="roleList" v-loading="loading" class="data-table styled-table" :empty-text="t('common.noData')">
      <el-table-column prop="name" :label="t('enterprise.role.name')" min-width="150" />
      <el-table-column prop="is_admin" :label="t('enterprise.role.isAdmin')" width="100" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.is_admin" effect="dark" round size="small">Admin</el-tag>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="is_default" :label="t('enterprise.role.isDefault')" width="100" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.is_default" effect="plain" round type="success" size="small">Default</el-tag>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('common.operation')" width="200" fixed="right">
        <template #default="{ row }">
          <template v-if="row.is_admin">
            <el-tooltip content="管理员角色拥有全部权限，不可修改" placement="top">
              <el-button link type="info" size="small" disabled>
                {{ t('common.edit') }}
              </el-button>
            </el-tooltip>
            <el-tooltip content="不能删除管理员角色" placement="top">
              <el-button link type="info" size="small" disabled>
                {{ t('common.delete') }}
              </el-button>
            </el-tooltip>
          </template>
          <template v-else>
            <el-button link type="primary" size="small" @click="handleEditRole(row)">
              {{ t('common.edit') }}
            </el-button>
            <el-button link type="danger" size="small" @click="handleDeleteRole(row)">
              {{ t('common.delete') }}
            </el-button>
          </template>
        </template>
      </el-table-column>
    </el-table>

    <!-- Role Dialog -->
    <el-dialog v-model="showRoleDialog" :title="editingRole ? t('enterprise.role.update') : t('enterprise.role.create')" width="550px">
      <el-form :model="roleForm" :rules="roleRules" ref="roleFormRef" label-width="80px">
        <el-form-item :label="t('enterprise.role.name')" prop="name">
          <el-input v-model="roleForm.name" :placeholder="t('enterprise.role.name')" />
        </el-form-item>
        <el-form-item :label="t('enterprise.role.powers')">
          <el-collapse v-model="expandedGroups" class="power-collapse">
            <el-collapse-item
              v-for="group in groupedPowers"
              :key="group.key"
              :name="group.key"
              :title="group.label"
            >
              <template #title>
                <div class="power-group-header">
                  <span>{{ group.label }}</span>
                  <el-tag size="small" type="info">{{ getGroupSelectedCount(group) }}/{{ group.powers.length }}</el-tag>
                </div>
              </template>
              <el-checkbox-group v-model="roleForm.power_ids">
                <div class="power-checkbox-list">
                  <el-checkbox
                    v-for="power in group.powers"
                    :key="power.id"
                    :label="power.id"
                  >
                    {{ power.description || power.name }}
                  </el-checkbox>
                </div>
              </el-checkbox-group>
            </el-collapse-item>
          </el-collapse>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRoleDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="savingRole" @click="handleSaveRole">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import type { ComponentInternalInstance } from 'vue'
  import { enterpriseApi } from '@myobj/api'
  import type { EnterpriseRole, EnterprisePower } from '@myobj/shared'
  import { useI18n } from '@/composables'

  const enterpriseId = inject<Ref<string>>('enterpriseId', ref(''))

  const { getRoleList, createRole, updateRole, deleteRole, getAllPowers } = enterpriseApi
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  const loading = ref(false)
  const roleList = ref<EnterpriseRole[]>([])
  const allPowers = ref<EnterprisePower[]>([])
  const expandedGroups = ref<string[]>([])

  const showRoleDialog = ref(false)
  const savingRole = ref(false)
  const editingRole = ref<EnterpriseRole | null>(null)
  const roleFormRef = ref()
  const roleForm = reactive({ name: '', power_ids: [] as number[] })
  const roleRules = {
    name: [{ required: true, message: t('enterprise.role.name'), trigger: 'blur' }]
  }

  const groupLabels: Record<string, string> = {
    'enterprise:member': '成员管理',
    'enterprise:role': '角色管理',
    'enterprise:space': '共享空间',
    'enterprise:audit': '审计日志',
    'enterprise:manage': '企业管理'
  }

  const groupedPowers = computed(() => {
    const groups: Record<string, EnterprisePower[]> = {}
    for (const power of allPowers.value) {
      // enterprise:member:invite -> enterprise:member
      const char = power.characteristic || ''
      const lastColon = char.lastIndexOf(':')
      const groupKey = lastColon > 0 ? char.substring(0, lastColon) : char
      if (!groups[groupKey]) groups[groupKey] = []
      groups[groupKey].push(power)
    }
    return Object.entries(groups)
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([key, powers]) => ({
        key,
        label: groupLabels[key] || key,
        powers: powers.sort((a, b) => a.id - b.id)
      }))
  })

  const getGroupSelectedCount = (group: { powers: EnterprisePower[] }) => {
    return group.powers.filter(p => roleForm.power_ids.includes(p.id)).length
  }

  const loadRoles = async () => {
    loading.value = true
    try {
      const res = await getRoleList(enterpriseId.value)
      if (res.code === 200 && res.data) {
        roleList.value = Array.isArray(res.data) ? res.data : (res.data.list || [])
      }
    } catch (error: any) {
      proxy?.$log?.error(error)
    } finally {
      loading.value = false
    }
  }

  const loadPowers = async () => {
    try {
      const res = await getAllPowers()
      if (res.code === 200 && res.data) {
        allPowers.value = res.data
      }
    } catch {}
  }

  const resetRoleForm = () => {
    editingRole.value = null
    roleForm.name = ''
    roleForm.power_ids = []
  }

  const handleEditRole = (role: EnterpriseRole) => {
    editingRole.value = role
    roleForm.name = role.name
    roleForm.power_ids = [...(role.power_ids || [])]
    showRoleDialog.value = true
  }

  const handleSaveRole = async () => {
    if (!roleFormRef.value) return
    await roleFormRef.value.validate(async (valid: boolean) => {
      if (!valid) return
      savingRole.value = true
      try {
        if (editingRole.value) {
          const res = await updateRole({ enterprise_id: enterpriseId.value, role_id: editingRole.value.id, name: roleForm.name, power_ids: roleForm.power_ids })
          if (res.code === 200) {
            proxy?.$modal.msgSuccess(t('enterprise.role.updateSuccess'))
            showRoleDialog.value = false
            loadRoles()
          } else {
            proxy?.$modal.msgError(res.message || t('common.operationFailed'))
          }
        } else {
          const res = await createRole({ enterprise_id: enterpriseId.value, name: roleForm.name, power_ids: roleForm.power_ids })
          if (res.code === 200) {
            proxy?.$modal.msgSuccess(t('enterprise.role.createSuccess'))
            showRoleDialog.value = false
            loadRoles()
          } else {
            proxy?.$modal.msgError(res.message || t('common.operationFailed'))
          }
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || t('common.operationFailed'))
      } finally {
        savingRole.value = false
      }
    })
  }

  const handleDeleteRole = async (role: EnterpriseRole) => {
    try {
      await proxy?.$modal.confirm(t('enterprise.role.deleteConfirm'))
      const res = await deleteRole({ enterprise_id: enterpriseId.value, role_id: role.id })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.role.deleteSuccess'))
        loadRoles()
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch {}
  }

  watch(enterpriseId, (id) => {
    roleList.value = []
    allPowers.value = []
    if (id) {
      loadRoles()
      loadPowers()
    }
  }, { immediate: true })
</script>

<style scoped>
  .enterprise-roles {
    display: flex;
    flex-direction: column;
    gap: 16px;
    height: 100%;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .section-header h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 700;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .section-header h3 .el-icon {
    color: var(--primary-color);
  }

  .data-table {
    width: 100%;
  }

  .power-collapse {
    width: 100%;
    border: none;
  }

  .power-collapse ::deep(.el-collapse-item__header) {
    font-size: 14px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    background: var(--el-fill-color-lighter);
    padding: 0 12px;
    border-radius: 6px;
    margin-bottom: 4px;
  }

  .power-collapse ::deep(.el-collapse-item__content) {
    padding: 8px 0;
  }

  .power-group-header {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
  }

  .power-checkbox-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 0 12px;
  }
</style>

<template>
  <div class="enterprise-manage">
    <!-- 成员管理 -->
    <div class="section">
      <div class="section-header">
        <h3>{{ t('enterprise.member.title') }}</h3>
        <div class="section-actions">
          <el-button v-if="isAdmin" type="primary" icon="Plus" size="small" @click="showInviteDialog = true">
            {{ t('enterprise.member.invite') }}
          </el-button>
          <el-button v-if="!isAdmin" type="danger" icon="Close" size="small" @click="handleLeave">
            {{ t('enterprise.member.leave') || '退出企业' }}
          </el-button>
          <el-button icon="Refresh" size="small" @click="loadMembers">{{ t('common.refresh') }}</el-button>
        </div>
      </div>

      <el-table :data="memberList" v-loading="memberLoading" class="data-table" :empty-text="t('common.noData')">
        <el-table-column prop="user_name" :label="t('enterprise.member.userName')" min-width="120" />
        <el-table-column prop="role_name" :label="t('enterprise.member.roleName')" width="120">
          <template #default="{ row }">
            <el-tag :type="row.is_admin ? 'danger' : 'info'" size="small">
              {{ row.role_name || '-' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="joined_at" :label="t('enterprise.member.joinedAt')" width="180" />
        <el-table-column prop="status" :label="t('enterprise.member.status')" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 0 ? 'success' : 'danger'" size="small">
              {{ row.status === 0 ? t('enterprise.member.statusActive') : t('enterprise.member.statusDisabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column v-if="isAdmin" :label="t('common.operation')" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleChangeRole(row)">
              {{ t('enterprise.member.changeRole') }}
            </el-button>
            <el-button link type="danger" size="small" @click="handleRemoveMember(row)">
              {{ t('enterprise.member.remove') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="memberPagination.page"
        v-model:page-size="memberPagination.pageSize"
        :total="memberPagination.total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="loadMembers"
        @current-change="loadMembers"
        class="pagination"
      />
    </div>

    <!-- 角色管理 -->
    <div v-if="isAdmin" class="section">
      <div class="section-header">
        <h3>{{ t('enterprise.role.title') }}</h3>
        <el-button type="primary" icon="Plus" size="small" @click="showRoleDialog = true; resetRoleForm()">
          {{ t('enterprise.role.create') }}
        </el-button>
      </div>

      <el-table :data="roleList" v-loading="roleLoading" class="data-table" :empty-text="t('common.noData')">
        <el-table-column prop="name" :label="t('enterprise.role.name')" min-width="150" />
        <el-table-column prop="is_admin" :label="t('enterprise.role.isAdmin')" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_admin ? 'danger' : 'info'" size="small">
              {{ row.is_admin ? '✓' : '-' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_default" :label="t('enterprise.role.isDefault')" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_default" type="success" size="small">✓</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.operation')" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleEditRole(row)">
              {{ t('common.edit') }}
            </el-button>
            <el-button link type="danger" size="small" @click="handleDeleteRole(row)">
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 审计日志 -->
    <div v-if="isAdmin" class="section">
      <div class="section-header">
        <h3>{{ t('enterprise.audit.title') }}</h3>
        <div class="section-actions">
          <el-input
            v-model="auditKeyword"
            :placeholder="t('enterprise.audit.searchByKeyword')"
            clearable
            size="small"
            style="width: 200px"
            @clear="loadAuditLogs"
            @keyup.enter="loadAuditLogs"
          >
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-button icon="Refresh" size="small" @click="loadAuditLogs">{{ t('common.refresh') }}</el-button>
          <el-button icon="Download" size="small" @click="handleExportAudit">{{ t('enterprise.audit.export') }}</el-button>
        </div>
      </div>

      <el-table :data="auditList" v-loading="auditLoading" class="data-table" :empty-text="t('common.noData')">
        <el-table-column prop="user_name" :label="t('enterprise.audit.operator')" width="120" />
        <el-table-column prop="action" :label="t('enterprise.audit.action')" width="150">
          <template #default="{ row }">
            <el-tag size="small">{{ row.action }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="target_name" :label="t('enterprise.audit.target')" min-width="150" />
        <el-table-column prop="detail" :label="t('enterprise.audit.detail')" min-width="200" show-overflow-tooltip />
        <el-table-column prop="ip" :label="t('enterprise.audit.ip')" width="130" />
        <el-table-column prop="created_at" :label="t('enterprise.audit.time')" width="180" />
      </el-table>

      <el-pagination
        v-model:current-page="auditPagination.page"
        v-model:page-size="auditPagination.pageSize"
        :total="auditPagination.total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @size-change="loadAuditLogs"
        @current-change="loadAuditLogs"
        class="pagination"
      />
    </div>

    <!-- 邀请成员对话框 -->
    <el-dialog v-model="showInviteDialog" :title="t('enterprise.member.invite')" width="500px">
      <el-form :model="inviteForm" :rules="inviteRules" ref="inviteFormRef" label-width="80px">
        <el-form-item :label="t('enterprise.member.userName')" prop="user_name">
          <el-input v-model="inviteForm.user_name" :placeholder="t('enterprise.member.userName')" @keyup.enter="handleInvite" />
        </el-form-item>
      </el-form>

      <el-divider />

      <div class="invite-code-section">
        <h4>{{ t('enterprise.info.inviteCode') }}</h4>
        <div class="invite-code-display">
          <el-input v-model="inviteCode" readonly>
            <template #append>
              <el-button @click="copyInviteCode">{{ t('common.copy') }}</el-button>
            </template>
          </el-input>
        </div>
        <el-button text size="small" @click="handleRefreshCode">{{ t('enterprise.member.refreshCode') }}</el-button>
      </div>

      <template #footer>
        <el-button @click="showInviteDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="inviting" @click="handleInvite">{{ t('enterprise.member.invite') }}</el-button>
      </template>
    </el-dialog>

    <!-- 变更角色对话框 -->
    <el-dialog v-model="showRoleChangeDialog" :title="t('enterprise.member.changeRole')" width="400px">
      <el-form label-width="80px">
        <el-form-item :label="t('enterprise.role.name')">
          <el-select v-model="selectedRoleId" style="width: 100%">
            <el-option
              v-for="role in roleList"
              :key="role.id"
              :label="role.name"
              :value="role.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRoleChangeDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="changingRole" @click="confirmChangeRole">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 角色编辑对话框 -->
    <el-dialog v-model="showRoleDialog" :title="editingRole ? t('enterprise.role.update') : t('enterprise.role.create')" width="500px">
      <el-form :model="roleForm" :rules="roleRules" ref="roleFormRef" label-width="80px">
        <el-form-item :label="t('enterprise.role.name')" prop="name">
          <el-input v-model="roleForm.name" :placeholder="t('enterprise.role.name')" />
        </el-form-item>
        <el-form-item :label="t('enterprise.role.powers')">
          <el-checkbox-group v-model="roleForm.power_ids">
            <el-checkbox
              v-for="power in allPowers"
              :key="power.id"
              :label="power.id"
            >
              {{ power.description || power.name }}
            </el-checkbox>
          </el-checkbox-group>
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
  import type { EnterpriseMember, EnterpriseRole, EnterpriseAuditLog, EnterprisePower } from '@myobj/shared'
  import { useI18n } from '@/composables'

  const props = defineProps<{
    enterpriseId: string
    isAdmin: boolean
  }>()

  const emit = defineEmits<{
    refresh: []
  }>()

  const {
    getMemberList, inviteMember, removeMember, updateMemberRole, leaveEnterprise,
    getInviteCode, refreshInviteCode,
    getRoleList, createRole, updateRole, deleteRole,
    getAllPowers,
    getEnterpriseAuditLogs, exportEnterpriseAuditLogs
  } = enterpriseApi

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  // Members
  const memberLoading = ref(false)
  const memberList = ref<EnterpriseMember[]>([])
  const memberPagination = reactive({ page: 1, pageSize: 20, total: 0 })

  // Roles
  const roleLoading = ref(false)
  const roleList = ref<EnterpriseRole[]>([])
  const allPowers = ref<EnterprisePower[]>([])

  // Audit
  const auditLoading = ref(false)
  const auditList = ref<EnterpriseAuditLog[]>([])
  const auditKeyword = ref('')
  const auditPagination = reactive({ page: 1, pageSize: 20, total: 0 })

  // Dialogs
  const showInviteDialog = ref(false)
  const inviting = ref(false)
  const inviteForm = reactive({ user_name: '' })
  const inviteFormRef = ref()
  const inviteRules = {
    user_name: [{ required: true, message: t('enterprise.member.userName'), trigger: 'blur' }]
  }
  const inviteCode = ref('')

  const showRoleChangeDialog = ref(false)
  const changingRole = ref(false)
  const selectedRoleId = ref('')
  const changingMemberId = ref('')

  const showRoleDialog = ref(false)
  const savingRole = ref(false)
  const editingRole = ref<EnterpriseRole | null>(null)
  const roleFormRef = ref()
  const roleForm = reactive({ name: '', power_ids: [] as number[] })
  const roleRules = {
    name: [{ required: true, message: t('enterprise.role.name'), trigger: 'blur' }]
  }

  // Load members
  const loadMembers = async () => {
    memberLoading.value = true
    try {
      const res = await getMemberList({
        enterprise_id: props.enterpriseId,
        page: memberPagination.page,
        pageSize: memberPagination.pageSize
      })
      if (res.code === 200 && res.data) {
        memberList.value = res.data.list || []
        memberPagination.total = res.data.total || 0
      }
    } catch (error: any) {
      proxy?.$log?.error(error)
    } finally {
      memberLoading.value = false
    }
  }

  // Load roles
  const loadRoles = async () => {
    roleLoading.value = true
    try {
      const res = await getRoleList(props.enterpriseId)
      if (res.code === 200 && res.data) {
        roleList.value = Array.isArray(res.data) ? res.data : (res.data.list || [])
      }
    } catch (error: any) {
      proxy?.$log?.error(error)
    } finally {
      roleLoading.value = false
    }
  }

  // Load powers
  const loadPowers = async () => {
    try {
      const res = await getAllPowers()
      if (res.code === 200 && res.data) {
        allPowers.value = res.data
      }
    } catch (error: any) {
      proxy?.$log?.error(error)
    }
  }

  // Load audit logs
  const loadAuditLogs = async () => {
    auditLoading.value = true
    try {
      const res = await getEnterpriseAuditLogs({
        enterprise_id: props.enterpriseId,
        page: auditPagination.page,
        pageSize: auditPagination.pageSize,
        keyword: auditKeyword.value || undefined
      })
      if (res.code === 200 && res.data) {
        auditList.value = res.data.list || []
        auditPagination.total = res.data.total || 0
      }
    } catch (error: any) {
      proxy?.$log?.error(error)
    } finally {
      auditLoading.value = false
    }
  }

  // Invite
  const loadInviteCode = async () => {
    try {
      const res = await getInviteCode(props.enterpriseId)
      if (res.code === 200 && res.data) {
        inviteCode.value = res.data.invite_code
      }
    } catch (error: any) {
      proxy?.$log?.error(error)
    }
  }

  const handleInvite = async () => {
    if (!inviteFormRef.value) return
    await inviteFormRef.value.validate(async (valid: boolean) => {
      if (!valid) return
      inviting.value = true
      try {
        const res = await inviteMember({ enterprise_id: props.enterpriseId, user_name: inviteForm.user_name })
        if (res.code === 200) {
          proxy?.$modal.msgSuccess(t('enterprise.member.inviteSuccess'))
          showInviteDialog.value = false
          inviteForm.user_name = ''
          loadMembers()
        } else {
          proxy?.$modal.msgError(res.message || t('common.operationFailed'))
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || t('common.operationFailed'))
      } finally {
        inviting.value = false
      }
    })
  }

  const copyInviteCode = () => {
    navigator.clipboard.writeText(inviteCode.value)
    proxy?.$modal.msgSuccess(t('common.copied'))
  }

  const handleRefreshCode = async () => {
    try {
      await proxy?.$modal.confirm(t('enterprise.member.refreshCodeConfirm'))
      const res = await refreshInviteCode(props.enterpriseId)
      if (res.code === 200 && res.data) {
        inviteCode.value = res.data.invite_code
        proxy?.$modal.msgSuccess(t('enterprise.member.refreshCodeSuccess'))
      }
    } catch {}
  }

  // Member operations
  const handleChangeRole = (member: EnterpriseMember) => {
    changingMemberId.value = member.id
    selectedRoleId.value = member.role_id
    showRoleChangeDialog.value = true
  }

  const confirmChangeRole = async () => {
    changingRole.value = true
    try {
      const res = await updateMemberRole({
        enterprise_id: props.enterpriseId,
        member_id: changingMemberId.value,
        role_id: selectedRoleId.value
      })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.member.changeRoleSuccess'))
        showRoleChangeDialog.value = false
        loadMembers()
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('common.operationFailed'))
    } finally {
      changingRole.value = false
    }
  }

  const handleRemoveMember = async (member: EnterpriseMember) => {
    try {
      await proxy?.$modal.confirm(t('enterprise.member.removeConfirm'))
      const res = await removeMember({ enterprise_id: props.enterpriseId, member_id: member.id })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.member.removeSuccess'))
        loadMembers()
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch {}
  }

  // Role operations
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
          const res = await updateRole({ enterprise_id: props.enterpriseId, role_id: editingRole.value.id, name: roleForm.name, power_ids: roleForm.power_ids })
          if (res.code === 200) {
            proxy?.$modal.msgSuccess(t('enterprise.role.updateSuccess'))
            showRoleDialog.value = false
            loadRoles()
          } else {
            proxy?.$modal.msgError(res.message || t('common.operationFailed'))
          }
        } else {
          const res = await createRole({ enterprise_id: props.enterpriseId, name: roleForm.name, power_ids: roleForm.power_ids })
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
      const res = await deleteRole({ enterprise_id: props.enterpriseId, role_id: role.id })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.role.deleteSuccess'))
        loadRoles()
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch {}
  }

  // Audit export
  const handleExportAudit = async () => {
    try {
      await exportEnterpriseAuditLogs({
        enterprise_id: props.enterpriseId,
        keyword: auditKeyword.value || undefined
      })
      proxy?.$modal.msgSuccess(t('enterprise.audit.exportSuccess'))
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('common.operationFailed'))
    }
  }

  // Leave enterprise
  const handleLeave = async () => {
    try {
      await proxy?.$modal.confirm(t('enterprise.member.leaveConfirm') || '确定要退出该企业吗？退出后将无法访问企业资源。')
      const res = await leaveEnterprise({ enterprise_id: props.enterpriseId })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.member.leaveSuccess') || '退出成功')
        emit('refresh')
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch {}
  }

  watch(() => props.enterpriseId, (id) => {
    roleList.value = []
    allPowers.value = []
    memberList.value = []
    auditList.value = []
    if (id) {
      loadMembers()
      if (props.isAdmin) {
        loadRoles()
        loadPowers()
        loadAuditLogs()
        loadInviteCode()
      }
    }
  }, { immediate: true })
</script>

<style scoped>
  .enterprise-manage {
    display: flex;
    flex-direction: column;
    gap: 24px;
    height: 100%;
    overflow: auto;
  }

  .section {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
  }

  .section-header h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
  }

  .section-actions {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .data-table {
    width: 100%;
  }

  .pagination {
    margin-top: 8px;
    justify-content: flex-end;
  }

  .invite-code-section {
    text-align: center;
  }

  .invite-code-section h4 {
    margin: 0 0 12px 0;
    font-size: 14px;
    color: var(--el-text-color-regular);
  }

  .invite-code-display {
    max-width: 400px;
    margin: 0 auto;
  }

  @media (max-width: 768px) {
    .section-header {
      flex-direction: column;
      align-items: flex-start;
    }

    .section-actions {
      width: 100%;
      flex-wrap: wrap;
    }

    .pagination ::deep(.el-pagination__sizes),
    .pagination ::deep(.el-pagination__jump) {
      display: none;
    }
  }

  html.dark .section-header h3 {
    color: var(--el-text-color-primary);
  }
</style>

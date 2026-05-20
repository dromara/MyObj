<template>
  <div class="enterprise-members page-card">
    <div class="section-header">
      <h3><el-icon><User /></el-icon> {{ t('enterprise.member.title') }}</h3>
      <div class="section-actions">
        <el-button v-if="isAdmin" class="action-btn" icon="Plus" @click="showInviteDialog = true">
          {{ t('enterprise.member.invite') }}
        </el-button>
        <el-button v-if="!isAdmin" type="danger" plain icon="Close" @click="handleLeave">
          {{ t('enterprise.member.leave') }}
        </el-button>
        <el-button class="action-btn-secondary" icon="Refresh" @click="loadMembers">{{ t('common.refresh') }}</el-button>
      </div>
    </div>

    <el-table :data="memberList" v-loading="loading" class="data-table styled-table" :empty-text="t('common.noData')">
      <el-table-column prop="user_name" :label="t('enterprise.member.userName')" min-width="120" />
      <el-table-column prop="role_name" :label="t('enterprise.member.roleName')" width="120">
        <template #default="{ row }">
          <el-tag v-if="row.is_admin" class="role-tag-admin" size="small">{{ row.role_name || '-' }}</el-tag>
          <el-tag v-else type="info" size="small">{{ row.role_name || '-' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="joined_at" :label="t('enterprise.member.joinedAt')" width="180" />
      <el-table-column prop="status" :label="t('enterprise.member.status')" width="100" align="center">
        <template #default="{ row }">
          <span :class="['status-indicator', row.status === 0 ? 'status-active' : 'status-disabled']">
            {{ row.status === 0 ? t('enterprise.member.statusActive') : t('enterprise.member.statusDisabled') }}
          </span>
        </template>
      </el-table-column>
      <el-table-column v-if="isAdmin" :label="t('common.operation')" width="200" fixed="right">
        <template #default="{ row }">
          <template v-if="row.is_admin">
            <el-tooltip :content="t('enterprise.member.cannotModifyAdmin') || '无法操作管理员'" placement="top">
              <el-button link type="info" size="small" disabled>
                {{ t('enterprise.member.changeRole') }}
              </el-button>
            </el-tooltip>
            <el-tooltip :content="t('enterprise.member.cannotRemoveAdmin') || '无法移除管理员'" placement="top">
              <el-button link type="info" size="small" disabled>
                {{ t('enterprise.member.remove') }}
              </el-button>
            </el-tooltip>
          </template>
          <template v-else>
            <el-button link type="primary" size="small" @click="handleChangeRole(row)">
              {{ t('enterprise.member.changeRole') }}
            </el-button>
            <el-button link type="danger" size="small" @click="handleRemoveMember(row)">
              {{ t('enterprise.member.remove') }}
            </el-button>
          </template>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-wrapper">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="loadMembers"
        @current-change="loadMembers"
      />
    </div>

    <!-- Invite Dialog -->
    <el-dialog v-model="showInviteDialog" :title="t('enterprise.member.invite')" width="500px">
      <el-form :model="inviteForm" :rules="inviteRules" ref="inviteFormRef" label-width="80px">
        <el-form-item :label="t('enterprise.member.userName')" prop="user_name">
          <el-input v-model="inviteForm.user_name" :placeholder="t('enterprise.member.userName')" @keyup.enter="handleInvite" />
        </el-form-item>
      </el-form>

      <el-divider />

      <div class="invite-code-section glass-panel">
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

    <!-- Change Role Dialog -->
    <el-dialog v-model="showRoleChangeDialog" :title="t('enterprise.member.changeRole')" width="400px">
      <el-form label-width="80px">
        <el-form-item :label="t('enterprise.role.name')">
          <el-select v-model="selectedRoleId" style="width: 100%" filterable>
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
  </div>
</template>

<script setup lang="ts">
  import type { ComponentInternalInstance } from 'vue'
  import { enterpriseApi } from '@myobj/api'
  import type { EnterpriseMember, EnterpriseRole } from '@myobj/shared'
  import { useI18n } from '@/composables'

  const enterpriseId = inject<Ref<string>>('enterpriseId', ref(''))
  const isAdmin = inject<Ref<boolean>>('isAdmin', ref(false))
  const loadEnterprises = inject<() => Promise<void>>('loadEnterprises', async () => {})

  const {
    getMemberList, inviteMember, removeMember, updateMemberRole, leaveEnterprise,
    getInviteCode, refreshInviteCode, getRoleList
  } = enterpriseApi

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  const loading = ref(false)
  const memberList = ref<EnterpriseMember[]>([])
  const roleList = ref<EnterpriseRole[]>([])
  const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

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

  const loadMembers = async () => {
    loading.value = true
    try {
      const res = await getMemberList({
        enterprise_id: enterpriseId.value,
        page: pagination.page,
        pageSize: pagination.pageSize
      })
      if (res.code === 200 && res.data) {
        memberList.value = res.data.list || []
        pagination.total = res.data.total || 0
      }
    } catch (error: any) {
      proxy?.$log?.error(error)
    } finally {
      loading.value = false
    }
  }

  const loadRoles = async () => {
    try {
      const res = await getRoleList(enterpriseId.value)
      if (res.code === 200 && res.data) {
        roleList.value = Array.isArray(res.data) ? res.data : (res.data.list || [])
      }
    } catch {}
  }

  const loadInviteCode = async () => {
    try {
      const res = await getInviteCode(enterpriseId.value)
      if (res.code === 200 && res.data) {
        inviteCode.value = res.data.invite_code
      }
    } catch {}
  }

  const handleInvite = async () => {
    if (!inviteFormRef.value) return
    await inviteFormRef.value.validate(async (valid: boolean) => {
      if (!valid) return
      inviting.value = true
      try {
        const res = await inviteMember({ enterprise_id: enterpriseId.value, user_name: inviteForm.user_name })
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
      const res = await refreshInviteCode(enterpriseId.value)
      if (res.code === 200 && res.data) {
        inviteCode.value = res.data.invite_code
        proxy?.$modal.msgSuccess(t('enterprise.member.refreshCodeSuccess'))
      }
    } catch {}
  }

  const handleChangeRole = async (member: EnterpriseMember) => {
    changingMemberId.value = member.id
    selectedRoleId.value = member.role_id
    if (roleList.value.length === 0) {
      await loadRoles()
    }
    showRoleChangeDialog.value = true
  }

  const confirmChangeRole = async () => {
    changingRole.value = true
    try {
      const res = await updateMemberRole({
        enterprise_id: enterpriseId.value,
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
      const res = await removeMember({ enterprise_id: enterpriseId.value, member_id: member.id })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.member.removeSuccess'))
        loadMembers()
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch {}
  }

  const handleLeave = async () => {
    try {
      await proxy?.$modal.confirm(t('enterprise.member.leaveConfirm'))
      const res = await leaveEnterprise({ enterprise_id: enterpriseId.value })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.member.leaveSuccess'))
        loadEnterprises()
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch {}
  }

  watch(enterpriseId, (id) => {
    memberList.value = []
    roleList.value = []
    if (id) {
      loadMembers()
      loadRoles()
      if (isAdmin.value) {
        loadInviteCode()
      }
    }
  }, { immediate: true })
</script>

<style scoped>
  .enterprise-members {
    display: flex;
    flex-direction: column;
    gap: 16px;
    height: 100%;
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
    font-weight: 700;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .section-header h3 .el-icon {
    color: var(--primary-color);
  }

  .section-actions {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .data-table {
    width: 100%;
  }

  .role-tag-admin {
    background: linear-gradient(135deg, rgba(37, 99, 235, 0.1), rgba(79, 70, 229, 0.1));
    color: var(--primary-color);
    border: none;
    font-weight: 600;
  }

  .status-indicator {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
  }

  .status-indicator::before {
    content: '';
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }

  .status-active {
    color: var(--el-color-success);
  }

  .status-active::before {
    background: var(--el-color-success);
  }

  .status-disabled {
    color: var(--el-color-danger);
  }

  .status-disabled::before {
    background: var(--el-color-danger);
  }

  .pagination-wrapper {
    margin-top: 8px;
    padding-top: 12px;
    border-top: 1px solid var(--el-border-color-lighter);
    display: flex;
    justify-content: flex-end;
  }

  .invite-code-section {
    text-align: center;
    padding: 16px;
    border-radius: 10px;
    background: var(--bg-color-glass);
    backdrop-filter: blur(12px);
    border: 1px solid var(--glass-border);
  }

  .invite-code-section h4 {
    margin: 0 0 12px 0;
    font-size: 14px;
    font-weight: 600;
    color: var(--primary-color);
  }

  .invite-code-display {
    max-width: 400px;
    margin: 0 auto 8px;
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

    .pagination-wrapper :deep(.el-pagination__sizes),
    .pagination-wrapper :deep(.el-pagination__jump) {
      display: none;
    }
  }

  html.dark .role-tag-admin {
    background: linear-gradient(135deg, rgba(37, 99, 235, 0.2), rgba(79, 70, 229, 0.2));
  }

  html.dark .invite-code-section {
    background: rgba(15, 23, 42, 0.6);
    border-color: var(--el-border-color);
  }
</style>

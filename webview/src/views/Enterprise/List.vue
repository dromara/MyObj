<template>
  <div class="enterprise-list-page">
    <div class="list-header">
      <h2 class="gradient-text">{{ t('enterprise.title') }}</h2>
      <div class="header-actions">
        <el-badge :value="pendingInvites.length" :hidden="pendingInvites.length === 0" :max="99">
          <el-button class="action-btn-secondary" icon="Bell" @click="showPendingDialog = true">
            {{ t('enterprise.member.pendingInvites') }}
          </el-button>
        </el-badge>
        <el-button class="action-btn-secondary" icon="Link" @click="showJoinDialog = true">
          {{ t('enterprise.member.joinByCode') }}
        </el-button>
        <el-button class="action-btn" icon="Plus" @click="showCreateDialog = true">
          {{ t('enterprise.create') }}
        </el-button>
      </div>
    </div>

    <div v-loading="loading" class="enterprise-grid">
      <el-empty v-if="!loading && enterpriseList.length === 0" :description="t('enterprise.noEnterprise')">
        <el-button class="action-btn" @click="showCreateDialog = true">{{ t('enterprise.create') }}</el-button>
      </el-empty>

      <div
        v-for="(item, index) in enterpriseList"
        :key="item.id"
        class="enterprise-card"
        :style="{ '--index': index }"
        @click="enterEnterprise(item.id)"
      >
        <div class="card-header">
          <div class="card-icon-wrapper">
            <el-icon :size="28" class="card-icon"><OfficeBuilding /></el-icon>
          </div>
          <div class="card-status">
            <el-tag v-if="item.state === 1" type="danger" size="small">
              {{ t('enterprise.info.stateDisabled') }}
            </el-tag>
            <el-tag v-if="item.is_admin === 1 || item.creator_id === userId" class="role-tag-admin" size="small">
              {{ t('enterprise.role.isAdmin') }}
            </el-tag>
          </div>
        </div>
        <div class="card-body">
          <h3 class="card-name">{{ item.name }}</h3>
          <p class="card-desc">{{ item.description || t('enterprise.info.description') }}</p>
          <div class="card-meta">
            <span class="meta-item"><el-icon class="meta-icon"><User /></el-icon> {{ item.member_count || 0 }} {{ t('enterprise.info.memberCount') }}</span>
            <span class="meta-item"><el-icon class="meta-icon"><Coin /></el-icon> {{ formatSpace(item) }}</span>
          </div>
        </div>
        <div class="card-footer">
          <span class="card-time">{{ item.created_at }}</span>
        </div>
      </div>
    </div>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreateDialog" :title="t('enterprise.create')" width="500px">
      <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="100px">
        <el-form-item :label="t('enterprise.info.name')" prop="name">
          <el-input v-model="createForm.name" :placeholder="t('enterprise.info.name')" />
        </el-form-item>
        <el-form-item :label="t('enterprise.info.description')">
          <el-input v-model="createForm.description" type="textarea" :rows="3" :placeholder="t('enterprise.info.description')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- Join Dialog -->
    <el-dialog v-model="showJoinDialog" :title="t('enterprise.member.joinByCode')" width="450px">
      <el-form :model="joinForm" ref="joinFormRef" label-width="80px">
        <el-form-item :label="t('enterprise.info.inviteCode')" :rules="[{ required: true, message: t('enterprise.member.inviteCodePlaceholder') }]" prop="invite_code">
          <el-input v-model="joinForm.invite_code" :placeholder="t('enterprise.member.inviteCodePlaceholder')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showJoinDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="joining" @click="handleJoin">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- Pending Invites Dialog -->
    <el-dialog v-model="showPendingDialog" :title="t('enterprise.member.pendingInvites')" width="550px">
      <div v-if="pendingInvites.length === 0" style="text-align: center; padding: 20px;">
        <el-empty :description="t('enterprise.member.noPendingInvites')" />
      </div>
      <el-table v-else :data="pendingInvites" style="width: 100%">
        <el-table-column :label="t('enterprise.info.name')" prop="enterprise_name" />
        <el-table-column :label="t('enterprise.member.inviter') || '邀请人'" prop="inviter_name" />
        <el-table-column :label="t('enterprise.member.inviteTime') || '邀请时间'" prop="created_at" width="170" />
        <el-table-column :label="t('common.operation')" width="120" align="center">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleAcceptInvite(row.id)">
              {{ t('enterprise.member.accept') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import type { ComponentInternalInstance } from 'vue'
  import type { FormRules } from 'element-plus'
  import { enterpriseApi } from '@myobj/api'
  import type { Enterprise } from '@myobj/shared'
  import { useI18n } from '@/composables'
  import { useUserStore } from '@/stores'

  const { getMyEnterprises, createEnterprise, joinEnterprise, acceptInvite, getPendingInvites } = enterpriseApi
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()
  const userStore = useUserStore()
  const router = useRouter()

  const userId = computed(() => userStore.userInfo?.id)
  const loading = ref(false)
  const enterpriseList = ref<Enterprise[]>([])
  const showCreateDialog = ref(false)
  const creating = ref(false)
  const createFormRef = ref()
  const showJoinDialog = ref(false)
  const joining = ref(false)
  const joinFormRef = ref()
  const showPendingDialog = ref(false)
  const pendingInvites = ref<any[]>([])

  const createForm = reactive({ name: '', description: '' })
  const joinForm = reactive({ invite_code: '' })

  const createRules: FormRules = {
    name: [
      { required: true, message: t('enterprise.info.name'), trigger: 'blur' },
      { min: 2, max: 50, message: t('enterprise.info.nameLength') || '名称长度为2-50个字符', trigger: 'blur' }
    ]
  }

  const formatSpace = (item: Enterprise) => {
    if (item.space_unlimited) {
      if (item.global_max_space && item.global_max_space > 0) {
        const units = ['B', 'KB', 'MB', 'GB', 'TB']
        let idx = 0
        let size = item.global_max_space
        while (size >= 1024 && idx < units.length - 1) {
          size /= 1024
          idx++
        }
        return `${size.toFixed(1)} ${units[idx]}`
      }
      return '∞'
    }
    if (!item.space || item.space === 0) return '0 B'
    const units = ['B', 'KB', 'MB', 'GB', 'TB']
    let idx = 0
    let size = item.space
    while (size >= 1024 && idx < units.length - 1) {
      size /= 1024
      idx++
    }
    return `${size.toFixed(1)} ${units[idx]}`
  }

  const enterEnterprise = (id: string) => {
    router.push(`/enterprise/${id}/space`)
  }

  const loadEnterprises = async () => {
    loading.value = true
    try {
      const res = await getMyEnterprises()
      if (res.code === 200 && res.data) {
        enterpriseList.value = Array.isArray(res.data) ? res.data : (res.data.list || [])
      }
    } catch (error: any) {
      console.error('[List.vue] getMyEnterprises error:', error)
      proxy?.$log?.error(error)
    } finally {
      loading.value = false
    }
  }

  const handleCreate = async () => {
    if (!createFormRef.value) return
    await createFormRef.value.validate(async (valid: boolean) => {
      if (!valid) return
      creating.value = true
      try {
        const res = await createEnterprise(createForm)
        if (res.code === 200) {
          proxy?.$modal.msgSuccess(t('enterprise.createSuccess'))
          showCreateDialog.value = false
          createForm.name = ''
          createForm.description = ''
          await loadEnterprises()
          if (res.data?.enterprise_id) {
            router.push(`/enterprise/${res.data.enterprise_id}/members`)
          }
        } else {
          proxy?.$modal.msgError(res.message || t('common.createFailed'))
        }
      } catch (error: any) {
        proxy?.$modal.msgError(error.message || t('common.createFailed'))
      } finally {
        creating.value = false
      }
    })
  }

  const handleJoin = async () => {
    if (!joinForm.invite_code.trim()) {
      proxy?.$modal.msgError(t('enterprise.member.inviteCodePlaceholder'))
      return
    }
    joining.value = true
    try {
      const res = await joinEnterprise({ invite_code: joinForm.invite_code.trim() })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.member.joinSuccess'))
        showJoinDialog.value = false
        joinForm.invite_code = ''
        await loadEnterprises()
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('common.operationFailed'))
    } finally {
      joining.value = false
    }
  }

  const loadPendingInvites = async () => {
    try {
      const res = await getPendingInvites()
      if (res.code === 200) {
        pendingInvites.value = Array.isArray(res.data) ? res.data : []
      }
    } catch {
      pendingInvites.value = []
    }
  }

  const handleAcceptInvite = async (inviteId: string) => {
    try {
      const res = await acceptInvite(inviteId)
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.member.acceptSuccess'))
        await loadPendingInvites()
        await loadEnterprises()
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('common.operationFailed'))
    }
  }

  onMounted(() => {
    loadEnterprises()
    loadPendingInvites()
  })
</script>

<style scoped>
  @keyframes fadeInUp {
    from {
      opacity: 0;
      transform: translateY(16px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .enterprise-list-page {
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 20px;
    padding: 4px;
  }

  .list-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
  }

  .list-header h2 {
    margin: 0;
    font-size: 22px;
    font-weight: 800;
    letter-spacing: -0.5px;
  }

  .header-actions {
    display: flex;
    gap: 10px;
    align-items: center;
    flex-wrap: wrap;
  }

  .enterprise-grid {
    flex: 1;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 16px;
    align-content: start;
    overflow: auto;
  }

  .enterprise-card {
    border: 1px solid var(--el-border-color-light);
    border-top: 3px solid transparent;
    border-image: linear-gradient(90deg, var(--primary-color), var(--secondary-color)) 1;
    border-radius: 12px;
    padding: 16px;
    cursor: pointer;
    transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
    display: flex;
    flex-direction: column;
    gap: 12px;
    background: var(--el-bg-color);
    animation: fadeInUp 0.4s cubic-bezier(0.4, 0, 0.2, 1) forwards;
    animation-delay: calc(var(--index, 0) * 0.08s);
    opacity: 0;
  }

  .enterprise-card:hover {
    border-color: var(--el-color-primary);
    box-shadow: 0 8px 24px rgba(37, 99, 235, 0.12);
    transform: translateY(-3px);
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
  }

  .card-icon-wrapper {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    background: linear-gradient(135deg, rgba(37, 99, 235, 0.08), rgba(79, 70, 229, 0.08));
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .card-icon {
    color: var(--primary-color);
    filter: drop-shadow(0 2px 4px rgba(37, 99, 235, 0.3));
  }

  .card-status {
    display: flex;
    gap: 6px;
  }

  .role-tag-admin {
    background: linear-gradient(135deg, rgba(37, 99, 235, 0.1), rgba(79, 70, 229, 0.1));
    color: var(--primary-color);
    border: none;
    font-weight: 600;
  }

  .card-body {
    flex: 1;
  }

  .card-name {
    margin: 0 0 6px 0;
    font-size: 16px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .card-desc {
    margin: 0 0 10px 0;
    font-size: 13px;
    color: var(--el-text-color-secondary);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    line-height: 1.4;
  }

  .card-meta {
    display: flex;
    gap: 16px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .meta-item {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .meta-icon {
    color: var(--primary-color);
  }

  .card-footer {
    border-top: 1px solid var(--el-border-color-lighter);
    padding-top: 10px;
  }

  .card-time {
    font-size: 12px;
    color: var(--el-text-color-placeholder);
  }

  @media (max-width: 768px) {
    .list-header {
      flex-direction: column;
      align-items: flex-start;
    }

    .header-actions {
      width: 100%;
    }

    .enterprise-grid {
      grid-template-columns: 1fr;
    }
  }

  html.dark .enterprise-card {
    background: var(--el-bg-color-overlay);
    border-color: var(--el-border-color);
  }

  html.dark .enterprise-card:hover {
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
  }

  html.dark .card-icon-wrapper {
    background: linear-gradient(135deg, rgba(37, 99, 235, 0.15), rgba(79, 70, 229, 0.15));
  }

  html.dark .role-tag-admin {
    background: linear-gradient(135deg, rgba(37, 99, 235, 0.2), rgba(79, 70, 229, 0.2));
  }
</style>

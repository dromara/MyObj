<template>
  <div class="enterprise-settings">
    <!-- 企业信息 -->
    <el-card shadow="never" class="settings-card">
      <template #header>
        <span class="card-header-title"><el-icon><OfficeBuilding /></el-icon> {{ t('enterprise.info.title') }}</span>
      </template>
      <el-form :model="infoForm" label-width="120px" style="max-width: 600px">
        <el-form-item :label="t('enterprise.info.name')">
          <el-input v-model="infoForm.name" />
        </el-form-item>
        <el-form-item :label="t('enterprise.info.description')">
          <el-input v-model="infoForm.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item>
          <el-button class="action-btn" :loading="savingInfo" @click="handleSaveInfo">{{ t('common.save') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 存储配额 -->
    <el-card shadow="never" class="settings-card">
      <template #header>
        <span class="card-header-title"><el-icon><Coin /></el-icon> {{ t('enterprise.setQuota') }}</span>
      </template>
      <el-form :model="quotaForm" label-width="120px" style="max-width: 600px">
        <el-form-item :label="t('enterprise.info.storage')">
          <div style="display: flex; align-items: center; gap: 12px; flex-wrap: wrap">
            <el-switch
              v-model="quotaForm.spaceUnlimited"
              :active-text="t('admin.spaceConfig.unlimited')"
              :disabled="globalMaxGB > 0"
            />
            <template v-if="!quotaForm.spaceUnlimited">
              <el-input-number v-model="quotaForm.spaceGB" :min="0" :max="globalMaxGB > 0 ? globalMaxGB : 999999" style="width: 180px" />
              <span style="color: var(--el-text-color-secondary)">GB</span>
            </template>
          </div>
          <div style="display: flex; align-items: center; gap: 20px; font-size: 12px; margin-top: 4px; flex-wrap: wrap">
            <span style="color: var(--el-text-color-secondary)">
              {{ t('enterprise.space.used') }}: {{ spaceUsage ? formatSize(spaceUsage.used_space) : '-' }}
            </span>
            <span v-if="globalMaxGB > 0" style="color: var(--el-color-warning)">
              {{ '系统设置的企业空间上限：' + globalMaxGB + ' GB' }}
            </span>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button class="action-btn" :loading="savingQuota" @click="handleSaveQuota">{{ t('common.save') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 危险操作 -->
    <el-card shadow="never" class="settings-card danger-card">
      <template #header>
        <span style="color: var(--el-color-danger)">{{ t('common.warning') }}</span>
      </template>

      <div class="danger-section">
        <div v-if="isCreator" class="danger-item">
          <div class="danger-info">
            <h4>{{ t('enterprise.transfer') }}</h4>
            <p>{{ t('enterprise.transferConfirm') }}</p>
          </div>
          <el-tooltip
            :content="t('enterprise.transferNoTarget') || '没有可转让的成员'"
            :disabled="transferableMembers.length > 0"
            placement="top"
          >
            <el-button type="warning" :disabled="transferableMembers.length === 0" @click="showTransferDialog = true">
              {{ t('enterprise.transfer') }}
            </el-button>
          </el-tooltip>
        </div>

        <el-divider v-if="isCreator" />

        <div class="danger-item">
          <div class="danger-info">
            <h4>{{ enterpriseInfo?.state === 0 ? t('enterprise.disable') : t('enterprise.enable') }}</h4>
            <p>{{ enterpriseInfo?.state === 0 ? t('enterprise.disableDesc') : t('enterprise.enableDesc') }}</p>
          </div>
          <el-button
            :type="enterpriseInfo?.state === 0 ? 'warning' : 'success'"
            @click="handleToggleState"
          >
            {{ enterpriseInfo?.state === 0 ? t('enterprise.disable') : t('enterprise.enable') }}
          </el-button>
        </div>

        <el-divider />

        <div v-if="isCreator" class="danger-item">
          <div class="danger-info">
            <h4 style="color: var(--el-color-danger)">{{ t('enterprise.dissolve') }}</h4>
            <p>{{ t('enterprise.dissolveConfirm') }}</p>
          </div>
          <el-button type="danger" @click="handleDissolve">{{ t('enterprise.dissolve') }}</el-button>
        </div>
      </div>
    </el-card>

    <!-- 转让所有权对话框 -->
    <el-dialog v-model="showTransferDialog" :title="t('enterprise.transfer')" width="500px">
      <el-form :model="transferForm" label-width="100px">
        <el-form-item :label="t('enterprise.transferTarget') || '转让给'">
          <el-select v-model="transferForm.member_id" :placeholder="t('enterprise.transferSelectPlaceholder') || '请选择转让目标'" style="width: 100%" filterable>
            <el-option
              v-for="member in transferableMembers"
              :key="member.id"
              :label="member.user_name + (member.role_name ? ` (${member.role_name})` : '')"
              :value="member.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showTransferDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="warning" :loading="transferring" :disabled="!transferForm.member_id" @click="handleTransfer">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import type { ComponentInternalInstance } from 'vue'
  import { enterpriseApi } from '@myobj/api'
  import type { Enterprise, SpaceUsage, EnterpriseMember } from '@myobj/shared'
  import { formatSize, GBToBytes, bytesToGB } from '@/utils'
  import { useI18n } from '@/composables'
  import { useUserStore } from '@/stores'

  const enterpriseId = inject<Ref<string>>('enterpriseId', ref(''))
  const loadEnterprises = inject<() => Promise<void>>('loadEnterprises', async () => {})

  const {
    getEnterpriseInfo, updateEnterprise, transferOwnership,
    dissolveEnterprise, toggleEnterpriseState, setEnterpriseQuota,
    getSpaceUsage, getMemberList
  } = enterpriseApi

  const userStore = useUserStore()

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()
  const router = useRouter()

  const enterpriseInfo = ref<Enterprise | null>(null)
  const spaceUsage = ref<SpaceUsage | null>(null)
  const globalMaxGB = ref(0)

  const infoForm = reactive({ name: '', description: '' })
  const quotaForm = reactive({ spaceGB: 0, spaceUnlimited: false })
  const savingInfo = ref(false)
  const savingQuota = ref(false)

  const showTransferDialog = ref(false)
  const transferring = ref(false)
  const transferForm = reactive({ member_id: '' })
  const memberList = ref<EnterpriseMember[]>([])

  const transferableMembers = computed(() => {
    const currentUserId = userStore.userInfo?.id
    return memberList.value.filter(m => m.user_id !== currentUserId && m.status === 0)
  })

  const isCreator = computed(() => enterpriseInfo.value?.creator_id === userStore.userInfo?.id)

  const loadMembers = async () => {
    try {
      const res = await getMemberList({
        enterprise_id: enterpriseId.value,
        page: 1,
        pageSize: 999
      })
      if (res.code === 200 && res.data) {
        memberList.value = res.data.list || []
      }
    } catch {}
  }

  const loadInfo = async () => {
    try {
      const res = await getEnterpriseInfo(enterpriseId.value)
      if (res.code === 200 && res.data) {
        enterpriseInfo.value = res.data
        infoForm.name = res.data.name
        infoForm.description = res.data.description || ''
        quotaForm.spaceUnlimited = res.data.space_unlimited || false
        globalMaxGB.value = res.data.global_max_space ? bytesToGB(res.data.global_max_space) : 0
      }
    } catch (error: any) {
      proxy?.$log?.error(error)
    }
  }

  const loadSpaceUsage = async () => {
    try {
      const res = await getSpaceUsage(enterpriseId.value)
      if (res.code === 200 && res.data) {
        spaceUsage.value = res.data
        quotaForm.spaceGB = res.data.total_space > 0 ? bytesToGB(res.data.total_space) : 0
      }
    } catch (error: any) {
      proxy?.$log?.error(error)
    }
  }

  const handleSaveInfo = async () => {
    if (!infoForm.name.trim()) return
    savingInfo.value = true
    try {
      const res = await updateEnterprise({
        enterprise_id: enterpriseId.value,
        name: infoForm.name,
        description: infoForm.description
      })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('common.saveSuccess'))
        loadEnterprises()
      } else {
        proxy?.$modal.msgError(res.message || t('common.saveFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('common.saveFailed'))
    } finally {
      savingInfo.value = false
    }
  }

  const handleSaveQuota = async () => {
    savingQuota.value = true
    try {
      const res = await setEnterpriseQuota({
        enterprise_id: enterpriseId.value,
        space: quotaForm.spaceUnlimited ? 0 : GBToBytes(quotaForm.spaceGB),
        space_unlimited: quotaForm.spaceUnlimited
      })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.setQuotaSuccess'))
        loadSpaceUsage()
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('common.operationFailed'))
    } finally {
      savingQuota.value = false
    }
  }

  const handleToggleState = async () => {
    const newState = enterpriseInfo.value?.state === 0 ? 1 : 0
    const desc = newState === 0 ? t('enterprise.enableDesc') : t('enterprise.disableDesc')
    try {
      await proxy?.$modal.confirm(`${desc}`)
      const res = await toggleEnterpriseState({
        enterprise_id: enterpriseId.value,
        state: newState
      })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(newState === 0 ? t('enterprise.enableSuccess') : t('enterprise.disableSuccess'))
        loadInfo()
        loadEnterprises()
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch (error: any) {
      if (error !== 'cancel') {
        proxy?.$modal.msgError(error.message || t('common.operationFailed'))
      }
    }
  }

  const handleTransfer = async () => {
    if (!transferForm.member_id) return
    const targetMember = memberList.value.find(m => m.id === transferForm.member_id)
    if (!targetMember) return
    try {
      await proxy?.$modal.confirm(t('enterprise.transferConfirm'))
      transferring.value = true
      const res = await transferOwnership({
        enterprise_id: enterpriseId.value,
        new_owner_id: targetMember.user_id,
        new_owner_name: targetMember.user_name
      })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.transferSuccess'))
        showTransferDialog.value = false
        transferForm.member_id = ''
        loadInfo()
        loadMembers()
        loadEnterprises()
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch (error: any) {
      if (error !== 'cancel') {
        proxy?.$modal.msgError(error.message || t('common.operationFailed'))
      }
    } finally {
      transferring.value = false
    }
  }

  const handleDissolve = async () => {
    try {
      await proxy?.$modal.confirm(t('enterprise.dissolveConfirm'))
      const res = await dissolveEnterprise({ enterprise_id: enterpriseId.value })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.dissolveSuccess'))
        loadEnterprises()
        router.push('/enterprise/list')
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch (error: any) {
      if (error !== 'cancel') {
        proxy?.$modal.msgError(error.message || t('common.operationFailed'))
      }
    }
  }

  watch(enterpriseId, (id) => {
    if (id) {
      loadInfo()
      loadSpaceUsage()
      loadMembers()
    }
  }, { immediate: true })
</script>

<style scoped>
  .enterprise-settings {
    display: flex;
    flex-direction: column;
    gap: 16px;
    height: 100%;
    overflow: auto;
  }

  .settings-card {
    border-radius: 12px;
    transition: box-shadow 0.3s;
  }

  .settings-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  }

  .card-header-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 600;
  }

  .card-header-title .el-icon {
    color: var(--primary-color);
  }

  .danger-card {
    border-color: var(--el-color-danger-light-5);
    border-top: 3px solid;
    border-image: linear-gradient(90deg, var(--el-color-danger), #f97316) 1;
  }

  .danger-section {
    display: flex;
    flex-direction: column;
    gap: 0;
  }

  .danger-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 0;
  }

  .danger-info h4 {
    margin: 0 0 4px 0;
    font-size: 14px;
    font-weight: 600;
  }

  .danger-info p {
    margin: 0;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  @media (max-width: 768px) {
    .danger-item {
      flex-direction: column;
      align-items: flex-start;
      gap: 8px;
    }
  }

  html.dark .settings-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  html.dark .danger-card {
    border-color: var(--el-color-danger-dark-2);
  }
</style>

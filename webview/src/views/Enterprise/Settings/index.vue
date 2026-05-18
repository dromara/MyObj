<template>
  <div class="enterprise-settings">
    <!-- 企业信息 -->
    <el-card shadow="never">
      <template #header>
        <span>{{ t('enterprise.info.title') }}</span>
      </template>
      <el-form :model="infoForm" label-width="120px" style="max-width: 600px">
        <el-form-item :label="t('enterprise.info.name')">
          <el-input v-model="infoForm.name" />
        </el-form-item>
        <el-form-item :label="t('enterprise.info.description')">
          <el-input v-model="infoForm.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="savingInfo" @click="handleSaveInfo">{{ t('common.save') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 存储配额 -->
    <el-card shadow="never">
      <template #header>
        <span>{{ t('enterprise.setQuota') }}</span>
      </template>
      <el-form :model="quotaForm" label-width="120px" style="max-width: 600px">
        <el-form-item :label="t('enterprise.info.storage')">
          <el-input-number v-model="quotaForm.spaceGB" :min="0" :max="999999" style="width: 200px" />
          <span style="margin-left: 8px; color: var(--el-text-color-secondary)">GB</span>
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px">
            {{ t('enterprise.space.used') }}: {{ spaceUsage ? formatSize(spaceUsage.used_space) : '-' }}
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="savingQuota" @click="handleSaveQuota">{{ t('common.save') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 危险操作 -->
    <el-card shadow="never" class="danger-card">
      <template #header>
        <span style="color: var(--el-color-danger)">{{ t('common.warning') }}</span>
      </template>

      <div class="danger-section">
        <div class="danger-item">
          <div class="danger-info">
            <h4>{{ t('enterprise.transfer') }}</h4>
            <p>{{ t('enterprise.transferConfirm') }}</p>
          </div>
          <el-button type="warning" @click="showTransferDialog = true">{{ t('enterprise.transfer') }}</el-button>
        </div>

        <el-divider />

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

        <div class="danger-item">
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
        <el-form-item :label="t('enterprise.member.userName')">
          <el-input v-model="transferForm.user_name" :placeholder="t('enterprise.member.userName')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showTransferDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="warning" :loading="transferring" @click="handleTransfer">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import type { ComponentInternalInstance } from 'vue'
  import { enterpriseApi } from '@myobj/api'
  import type { Enterprise, SpaceUsage } from '@myobj/shared'
  import { formatSize, GBToBytes, bytesToGB } from '@/utils'
  import { useI18n } from '@/composables'

  const props = defineProps<{
    enterpriseId: string
  }>()

  const emit = defineEmits<{
    refresh: []
    dissolved: []
  }>()

  const {
    getEnterpriseInfo, updateEnterprise, transferOwnership,
    dissolveEnterprise, toggleEnterpriseState, setEnterpriseQuota,
    getSpaceUsage
  } = enterpriseApi

  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()

  const enterpriseInfo = ref<Enterprise | null>(null)
  const spaceUsage = ref<SpaceUsage | null>(null)

  const infoForm = reactive({ name: '', description: '' })
  const quotaForm = reactive({ spaceGB: 0 })
  const savingInfo = ref(false)
  const savingQuota = ref(false)

  const showTransferDialog = ref(false)
  const transferring = ref(false)
  const transferForm = reactive({ user_name: '' })

  const loadInfo = async () => {
    try {
      const res = await getEnterpriseInfo(props.enterpriseId)
      if (res.code === 200 && res.data) {
        enterpriseInfo.value = res.data
        infoForm.name = res.data.name
        infoForm.description = res.data.description || ''
      }
    } catch (error: any) {
      proxy?.$log?.error(error)
    }
  }

  const loadSpaceUsage = async () => {
    try {
      const res = await getSpaceUsage(props.enterpriseId)
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
        enterprise_id: props.enterpriseId,
        name: infoForm.name,
        description: infoForm.description
      })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('common.saveSuccess'))
        emit('refresh')
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
        enterprise_id: props.enterpriseId,
        space: GBToBytes(quotaForm.spaceGB)
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
        enterprise_id: props.enterpriseId,
        state: newState
      })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(newState === 0 ? t('enterprise.enableSuccess') : t('enterprise.disableSuccess'))
        loadInfo()
        emit('refresh')
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
    if (!transferForm.user_name.trim()) return
    try {
      await proxy?.$modal.confirm(t('enterprise.transferConfirm'))
      transferring.value = true
      const res = await transferOwnership({
        enterprise_id: props.enterpriseId,
        new_owner_name: transferForm.user_name
      })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.transferSuccess'))
        showTransferDialog.value = false
        transferForm.user_name = ''
        emit('refresh')
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
      const res = await dissolveEnterprise({ enterprise_id: props.enterpriseId })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.dissolveSuccess'))
        emit('dissolved')
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch {}
  }

  watch(() => props.enterpriseId, (id) => {
    if (id) {
      loadInfo()
      loadSpaceUsage()
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

  .danger-card {
    border-color: var(--el-color-danger-light-5);
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

  html.dark .danger-card {
    border-color: var(--el-color-danger-dark-2);
  }
</style>

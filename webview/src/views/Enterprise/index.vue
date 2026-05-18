<template>
  <div class="enterprise-page">
    <el-card shadow="never" class="enterprise-header-card">
      <div class="enterprise-header">
        <div class="header-left">
          <el-icon :size="28" class="enterprise-icon"><OfficeBuilding /></el-icon>
          <h2>{{ t('enterprise.title') }}</h2>
          <el-dropdown v-if="enterpriseList.length > 0" @command="handleSwitchEnterprise" trigger="click">
            <el-tag class="enterprise-tag" effect="plain" size="large">
              {{ currentEnterpriseName }}
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-tag>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item
                  v-for="item in enterpriseList"
                  :key="item.id"
                  :command="item.id"
                  :disabled="item.id === currentEnterpriseId"
                >
                  {{ item.name }}
                  <el-icon v-if="item.id === currentEnterpriseId" class="el-icon--right"><Check /></el-icon>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
        <div class="header-right">
          <el-badge :value="pendingInvites.length" :hidden="pendingInvites.length === 0" :max="99">
            <el-button icon="Bell" @click="showPendingDialog = true">
              {{ t('enterprise.member.pendingInvites') || '待处理邀请' }}
            </el-button>
          </el-badge>
          <el-button icon="Link" @click="showJoinDialog = true">
            {{ t('enterprise.member.joinByCode') || '邀请码加入' }}
          </el-button>
          <el-button type="primary" icon="Plus" @click="showCreateDialog = true">
            {{ t('enterprise.create') }}
          </el-button>
          <el-button icon="Refresh" @click="loadEnterprises">{{ t('common.refresh') }}</el-button>
        </div>
      </div>
    </el-card>

    <el-card v-if="!currentEnterpriseId && !loading" v-loading="loading" shadow="never" class="enterprise-empty-card">
      <el-empty :description="t('enterprise.noEnterprise')">
        <el-button type="primary" @click="showCreateDialog = true">{{ t('enterprise.create') }}</el-button>
      </el-empty>
    </el-card>

    <template v-else>
      <el-card shadow="never" class="enterprise-content-card">
        <el-tabs v-model="activeTab" class="enterprise-tabs">
          <el-tab-pane :label="t('route.enterpriseManage')" name="manage">
            <EnterpriseManage
              v-if="currentEnterpriseId"
              :enterprise-id="currentEnterpriseId"
              :is-admin="isAdminOfCurrent"
              @refresh="loadEnterprises"
            />
          </el-tab-pane>
          <el-tab-pane :label="t('route.enterpriseSpace')" name="space">
            <EnterpriseSpace
              v-if="currentEnterpriseId"
              :enterprise-id="currentEnterpriseId"
            />
          </el-tab-pane>
          <el-tab-pane v-if="isAdminOfCurrent" :label="t('route.enterpriseSettings')" name="settings">
            <EnterpriseSettings
              :enterprise-id="currentEnterpriseId"
              @refresh="loadEnterprises"
              @dissolved="handleDissolved"
            />
          </el-tab-pane>
        </el-tabs>
      </el-card>
    </template>

    <!-- 创建企业对话框 -->
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

    <!-- 邀请码加入对话框 -->
    <el-dialog v-model="showJoinDialog" :title="t('enterprise.member.joinByCode') || '邀请码加入企业'" width="450px">
      <el-form :model="joinForm" ref="joinFormRef" label-width="80px">
        <el-form-item :label="t('enterprise.member.inviteCode') || '邀请码'" :rules="[{ required: true, message: '请输入邀请码' }]" prop="invite_code">
          <el-input v-model="joinForm.invite_code" :placeholder="t('enterprise.member.inviteCodePlaceholder') || '请输入企业邀请码'" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showJoinDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="joining" @click="handleJoin">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 待处理邀请对话框 -->
    <el-dialog v-model="showPendingDialog" :title="t('enterprise.member.pendingInvites') || '待处理邀请'" width="550px">
      <div v-if="pendingInvites.length === 0" style="text-align: center; padding: 20px;">
        <el-empty :description="t('enterprise.member.noPendingInvites') || '暂无待处理邀请'" />
      </div>
      <el-table v-else :data="pendingInvites" style="width: 100%">
        <el-table-column :label="t('enterprise.info.name') || '企业名称'" prop="enterprise_name" />
        <el-table-column :label="t('enterprise.member.inviter') || '邀请人'" prop="inviter_name" />
        <el-table-column :label="t('enterprise.member.inviteTime') || '邀请时间'" prop="created_at" width="170" />
        <el-table-column :label="t('common.action') || '操作'" width="120" align="center">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleAcceptInvite(row.id)">
              {{ t('enterprise.member.accept') || '接受' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import type { ComponentInternalInstance } from 'vue'
  import { enterpriseApi } from '@myobj/api'
  import type { Enterprise } from '@myobj/shared'
  import { useI18n } from '@/composables'
  import { useUserStore } from '@/stores'
  import EnterpriseManage from './Manage/index.vue'
  import EnterpriseSpace from './Space/index.vue'
  import EnterpriseSettings from './Settings/index.vue'
  import type { FormRules } from 'element-plus'

  const { getMyEnterprises, switchEnterprise, createEnterprise, joinEnterprise, acceptInvite, getPendingInvites } = enterpriseApi
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()
  const userStore = useUserStore()

  const route = useRoute()
  const router = useRouter()

  const loading = ref(false)
  const enterpriseList = ref<Enterprise[]>([])
  const currentEnterpriseId = ref<string>('')
  const currentEnterprise = ref<Enterprise | null>(null)
  const activeTab = ref('manage')
  const showCreateDialog = ref(false)
  const creating = ref(false)
  const createFormRef = ref()
  const showJoinDialog = ref(false)
  const joining = ref(false)
  const joinFormRef = ref()
  const showPendingDialog = ref(false)
  const pendingInvites = ref<any[]>([])

  const createForm = reactive({
    name: '',
    description: ''
  })

  const joinForm = reactive({
    invite_code: ''
  })

  const createRules: FormRules = {
    name: [
      { required: true, message: t('enterprise.info.name'), trigger: 'blur' },
      { min: 2, max: 50, message: t('enterprise.info.nameLength') || '名称长度为2-50个字符', trigger: 'blur' }
    ]
  }

  const currentEnterpriseName = computed(() => currentEnterprise.value?.name || t('enterprise.noEnterprise'))
  const isAdminOfCurrent = computed(() => currentEnterprise.value?.role === 'admin' || currentEnterprise.value?.creator_id === userStore.userInfo?.id)

  // 根据路由设置活动标签
  watch(
    () => route.name,
    name => {
      if (name === 'EnterpriseManage') activeTab.value = 'manage'
      else if (name === 'EnterpriseSpace') activeTab.value = 'space'
      else if (name === 'EnterpriseSettings') activeTab.value = 'settings'
    },
    { immediate: true }
  )

  // 标签切换时更新路由
  watch(activeTab, tab => {
    const routeMap: Record<string, string> = {
      manage: '/enterprise/manage',
      space: '/enterprise/space',
      settings: '/enterprise/settings'
    }
    if (routeMap[tab] && route.path !== routeMap[tab]) {
      router.push(routeMap[tab])
    }
  })

  const loadEnterprises = async () => {
    loading.value = true
    try {
      const res = await getMyEnterprises()
      if (res.code === 200 && res.data) {
        enterpriseList.value = Array.isArray(res.data) ? res.data : (res.data.list || [])
        // 自动选择第一个企业
        if (enterpriseList.value.length > 0 && !currentEnterpriseId.value) {
          const cached = localStorage.getItem('currentEnterpriseId')
          const found = cached ? enterpriseList.value.find(e => e.id === cached) : null
          selectEnterprise(found ? found.id : enterpriseList.value[0].id)
        }
      }
    } catch (error: any) {
      proxy?.$log?.error(error)
    } finally {
      loading.value = false
    }
  }

  const selectEnterprise = (id: string) => {
    currentEnterpriseId.value = id
    currentEnterprise.value = enterpriseList.value.find(e => e.id === id) || null
    localStorage.setItem('currentEnterpriseId', id)
  }

  const handleSwitchEnterprise = async (id: string) => {
    try {
      const res = await switchEnterprise({ enterprise_id: id })
      if (res.code === 200) {
        selectEnterprise(id)
        proxy?.$modal.msgSuccess(t('enterprise.switch'))
      } else {
        proxy?.$modal.msgError(res.message || t('common.operationFailed'))
      }
    } catch (error: any) {
      proxy?.$modal.msgError(error.message || t('common.operationFailed'))
    }
  }

  const handleCreate = async () => {
    if (!createFormRef.value) return
    await createFormRef.value.validate(async (valid: boolean) => {
      if (valid) {
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
              selectEnterprise(res.data.enterprise_id)
            }
          } else {
            proxy?.$modal.msgError(res.message || t('common.createFailed'))
          }
        } catch (error: any) {
          proxy?.$modal.msgError(error.message || t('common.createFailed'))
        } finally {
          creating.value = false
        }
      }
    })
  }

  const handleDissolved = () => {
    currentEnterpriseId.value = ''
    currentEnterprise.value = null
    localStorage.removeItem('currentEnterpriseId')
    loadEnterprises()
  }

  const handleJoin = async () => {
    if (!joinForm.invite_code.trim()) {
      proxy?.$modal.msgError(t('enterprise.member.inviteCodePlaceholder') || '请输入邀请码')
      return
    }
    joining.value = true
    try {
      const res = await joinEnterprise({ invite_code: joinForm.invite_code.trim() })
      if (res.code === 200) {
        proxy?.$modal.msgSuccess(t('enterprise.member.joinSuccess') || '加入成功')
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
        proxy?.$modal.msgSuccess(t('enterprise.member.acceptSuccess') || '已接受邀请')
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
  .enterprise-page {
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 4px;
  }

  .enterprise-header-card {
    flex-shrink: 0;
  }

  .enterprise-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .header-right {
    display: flex;
    gap: 12px;
  }

  .enterprise-icon {
    color: var(--primary-color);
    filter: drop-shadow(0 2px 4px rgba(99, 102, 241, 0.3));
  }

  html.dark .enterprise-icon {
    filter: drop-shadow(0 2px 4px rgba(99, 102, 241, 0.5));
  }

  .enterprise-header h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 700;
    color: var(--text-primary);
  }

  .enterprise-tag {
    cursor: pointer;
    font-size: 14px;
  }

  .enterprise-empty-card {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .enterprise-content-card {
    flex: 1;
    overflow: hidden;
  }

  .enterprise-tabs {
    height: 100%;
  }

  .enterprise-tabs :deep(.el-tabs__content) {
    height: calc(100% - 55px);
    overflow: auto;
  }

  .enterprise-tabs :deep(.el-tab-pane) {
    height: 100%;
  }

  html.dark .enterprise-header-card {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .enterprise-content-card {
    background: var(--card-bg);
    border-color: var(--el-border-color);
  }

  html.dark .enterprise-tabs :deep(.el-tabs__header) {
    background: transparent;
    border-bottom-color: var(--el-border-color);
  }

  html.dark .enterprise-tabs :deep(.el-tabs__item) {
    color: var(--el-text-color-regular);
  }

  html.dark .enterprise-tabs :deep(.el-tabs__item.is-active) {
    color: var(--primary-color);
  }

  html.dark .enterprise-tabs :deep(.el-tabs__item:hover) {
    color: var(--primary-color);
  }

  @media (max-width: 768px) {
    .enterprise-header {
      flex-direction: column;
      gap: 12px;
    }

    .header-left {
      flex-wrap: wrap;
    }

    .header-right {
      width: 100%;
      justify-content: flex-end;
    }
  }
</style>

<template>
  <div class="enterprise-layout">
    <aside class="enterprise-sidebar">
      <div class="sidebar-header">
        <el-icon :size="24" class="enterprise-icon"><OfficeBuilding /></el-icon>
        <span class="sidebar-title">{{ t('enterprise.title') }}</span>
      </div>

      <div class="sidebar-enterprise-selector">
        <el-dropdown v-if="enterpriseList.length > 0" @command="handleSelectEnterprise" trigger="click" placement="bottom-start">
          <div class="enterprise-selector-trigger">
            <div class="selector-info">
              <div class="selector-name">{{ currentEnterprise?.name || t('enterprise.noEnterprise') }}</div>
              <div class="selector-meta" v-if="currentEnterprise">
                {{ currentEnterprise.member_count || 0 }} {{ t('enterprise.info.memberCount') }}
              </div>
            </div>
            <el-icon><ArrowDown /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item
                v-for="item in enterpriseList"
                :key="item.id"
                :command="item.id"
                :disabled="item.id === currentEnterpriseId"
              >
                <div class="dropdown-enterprise-item">
                  <span>{{ item.name }}</span>
                  <el-icon v-if="item.id === currentEnterpriseId" class="el-icon--right"><Check /></el-icon>
                </div>
              </el-dropdown-item>
              <el-dropdown-item divided command="__list__">
                <el-icon><List /></el-icon>
                {{ t('enterprise.list.allEnterprises') || '所有企业' }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button v-else text type="primary" @click="router.push('/enterprise/list')">
          {{ t('enterprise.create') }}
        </el-button>
      </div>

      <el-menu
        v-if="currentEnterpriseId"
        :default-active="activeMenu"
        class="sidebar-menu"
        router
      >
        <el-menu-item v-if="hasPower('enterprise:member:view')" :index="`/enterprise/${currentEnterpriseId}/members`">
          <el-icon><User /></el-icon>
          <template #title>{{ t('enterprise.member.title') }}</template>
        </el-menu-item>
        <el-menu-item v-if="hasPower('enterprise:role:manage')" :index="`/enterprise/${currentEnterpriseId}/roles`">
          <el-icon><Key /></el-icon>
          <template #title>{{ t('enterprise.role.title') }}</template>
        </el-menu-item>
        <el-menu-item :index="`/enterprise/${currentEnterpriseId}/space`">
          <el-icon><FolderOpened /></el-icon>
          <template #title>{{ t('enterprise.space.title') }}</template>
        </el-menu-item>
        <el-menu-item v-if="hasPower('enterprise:audit:view')" :index="`/enterprise/${currentEnterpriseId}/audit`">
          <el-icon><Document /></el-icon>
          <template #title>{{ t('enterprise.audit.title') }}</template>
        </el-menu-item>
        <el-menu-item v-if="hasPower('enterprise:manage')" :index="`/enterprise/${currentEnterpriseId}/settings`">
          <el-icon><Setting /></el-icon>
          <template #title>{{ t('route.enterpriseSettings') }}</template>
        </el-menu-item>
      </el-menu>

      <div class="sidebar-footer">
        <el-button text @click="router.push('/enterprise/list')" style="width: 100%">
          <el-icon><List /></el-icon>
          <span>{{ t('enterprise.list.allEnterprises') || '所有企业' }}</span>
        </el-button>
      </div>
    </aside>

    <main class="enterprise-main">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
  import type { ComponentInternalInstance } from 'vue'
  import { enterpriseApi } from '@myobj/api'
  import type { Enterprise } from '@myobj/shared'
  import { useI18n } from '@/composables'
  import { useUserStore } from '@/stores'

  const { getMyEnterprises } = enterpriseApi
  const { proxy } = getCurrentInstance() as ComponentInternalInstance
  const { t } = useI18n()
  const userStore = useUserStore()
  const route = useRoute()
  const router = useRouter()

  const loading = ref(false)
  const enterpriseList = ref<Enterprise[]>([])
  const currentEnterpriseId = ref<string>('')
  const currentEnterprise = ref<Enterprise | null>(null)

  const isAdmin = computed(() =>
    currentEnterprise.value?.is_admin === 1 || currentEnterprise.value?.creator_id === userStore.userInfo?.id
  )

  const hasPower = (power: string) => {
    return isAdmin.value || (currentEnterprise.value?.powers?.includes(power) ?? false)
  }

  const activeMenu = computed(() => route.path)

  const loadEnterprises = async () => {
    loading.value = true
    try {
      const res = await getMyEnterprises()
      if (res.code === 200 && res.data) {
        enterpriseList.value = Array.isArray(res.data) ? res.data : (res.data.list || [])
        const id = route.params.id as string
        if (id) {
          currentEnterpriseId.value = id
          currentEnterprise.value = enterpriseList.value.find(e => e.id === id) || null
        }
      }
    } catch (error: any) {
      console.error('[Layout.vue] getMyEnterprises error:', error)
      proxy?.$log?.error(error)
    } finally {
      loading.value = false
    }
  }

  const handleSelectEnterprise = (command: string) => {
    if (command === '__list__') {
      router.push('/enterprise/list')
      return
    }
    currentEnterpriseId.value = command
    currentEnterprise.value = enterpriseList.value.find(e => e.id === command) || null
    const currentRouteName = route.name as string
    const subPath = currentRouteName?.startsWith('Enterprise') && currentRouteName !== 'EnterpriseList'
      ? route.path.split('/').slice(3).join('/')
      : 'space'
    router.push(`/enterprise/${command}/${subPath}`)
  }

  watch(() => route.params.id, (id) => {
    if (id && id !== currentEnterpriseId.value) {
      currentEnterpriseId.value = id as string
      currentEnterprise.value = enterpriseList.value.find(e => e.id === id) || null
    }
  })

  provide('enterpriseId', currentEnterpriseId)
  provide('isAdmin', isAdmin)
  provide('loadEnterprises', loadEnterprises)

  onMounted(() => {
    loadEnterprises()
  })
</script>

<style scoped>
  .enterprise-layout {
    display: flex;
    height: 100%;
    overflow: hidden;
  }

  .enterprise-sidebar {
    width: 240px;
    min-width: 240px;
    display: flex;
    flex-direction: column;
    border-right: 1px solid var(--el-border-color-light);
    background: linear-gradient(180deg, var(--el-bg-color) 0%, var(--el-bg-color-overlay) 100%);
  }

  .sidebar-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 16px 16px 12px;
    flex-shrink: 0;
  }

  .enterprise-icon {
    color: var(--primary-color);
    filter: drop-shadow(0 2px 4px rgba(37, 99, 235, 0.3));
  }

  .sidebar-title {
    font-size: 16px;
    font-weight: 800;
    letter-spacing: -0.5px;
    color: var(--el-text-color-primary);
  }

  .sidebar-enterprise-selector {
    padding: 0 12px 12px;
    flex-shrink: 0;
  }

  .enterprise-selector-trigger {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 12px;
    border-radius: 8px;
    cursor: pointer;
    background: var(--el-fill-color-light);
    border-left: 3px solid var(--primary-color);
    transition: all 0.2s;
  }

  .enterprise-selector-trigger:hover {
    background: var(--el-fill-color);
    border-left-width: 4px;
  }

  .selector-info {
    flex: 1;
    min-width: 0;
  }

  .selector-name {
    font-size: 14px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .selector-meta {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    margin-top: 2px;
  }

  .dropdown-enterprise-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
  }

  .sidebar-menu {
    flex: 1;
    border-right: none;
    overflow-y: auto;
  }

  .sidebar-menu ::deep(.el-menu-item) {
    height: 44px;
    line-height: 44px;
    margin: 2px 8px;
    border-radius: 8px;
    transition: all 0.2s;
  }

  .sidebar-menu ::deep(.el-menu-item.is-active) {
    background: var(--el-color-primary-light-9);
    border-left: 3px solid var(--primary-color);
    padding-left: 13px;
    font-weight: 600;
  }

  .sidebar-footer {
    padding: 8px;
    border-top: 1px solid var(--el-border-color-lighter);
    flex-shrink: 0;
  }

  .sidebar-footer ::deep(.el-button) {
    border-radius: 8px;
    transition: all 0.2s;
  }

  .sidebar-footer ::deep(.el-button:hover) {
    background: var(--el-fill-color-light);
  }

  .enterprise-main {
    flex: 1;
    overflow: auto;
    padding: 16px;
    min-width: 0;
  }

  html.dark .enterprise-sidebar {
    background: linear-gradient(180deg, var(--el-bg-color-overlay) 0%, rgba(15, 23, 42, 0.9) 100%);
    border-right-color: var(--el-border-color);
  }

  html.dark .enterprise-selector-trigger {
    background: var(--el-fill-color-dark);
  }

  html.dark .enterprise-selector-trigger:hover {
    background: var(--el-fill-color-darker);
  }

  html.dark .sidebar-menu ::deep(.el-menu-item.is-active) {
    background: rgba(var(--el-color-primary-rgb), 0.15);
  }
</style>

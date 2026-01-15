<template>
  <div class="collaboration-page">
    <el-card shadow="never" class="page-header-card">
      <div class="page-header">
        <div class="header-left">
          <el-icon :size="28" class="page-icon"><UserFilled /></el-icon>
          <h2>{{ t('collaboration.title') }}</h2>
        </div>
        <div class="header-right">
          <el-button type="primary" icon="Plus" @click="showInviteDialog = true">
            {{ t('collaboration.invite') }}
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card shadow="never" class="content-card">
      <el-tabs v-model="activeTab" class="collaboration-tabs">
        <el-tab-pane :label="t('collaboration.workspace')" name="workspace">
          <div class="workspace-container">
            <div class="workspace-list">
              <el-card
                v-for="workspace in workspaces"
                :key="workspace.id"
                class="workspace-card"
                shadow="hover"
                @click="selectWorkspace(workspace)"
              >
                <div class="workspace-header">
                  <el-icon :size="24" color="var(--primary-color)"><FolderOpened /></el-icon>
                  <h3>{{ workspace.name }}</h3>
                </div>
                <div class="workspace-info">
                  <div class="info-item">
                    <el-icon><User /></el-icon>
                    <span>{{ workspace.memberCount }} {{ t('collaboration.members') }}</span>
                  </div>
                  <div class="info-item">
                    <el-icon><Document /></el-icon>
                    <span>{{ workspace.fileCount }} {{ t('files.title') }}</span>
                  </div>
                </div>
                <div class="workspace-footer">
                  <el-tag :type="workspace.permission === 'admin' ? 'danger' : 'primary'">
                    {{ t(`collaboration.${workspace.permission}`) }}
                  </el-tag>
                  <span class="update-time">{{ formatTime(workspace.updateTime) }}</span>
                </div>
              </el-card>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane :label="t('collaboration.team')" name="team">
          <div class="team-container">
            <el-table :data="teamMembers" style="width: 100%">
              <el-table-column prop="name" :label="t('collaboration.members')" width="200">
                <template #default="{ row }">
                  <div class="member-cell">
                    <el-avatar :size="32" :style="{ backgroundColor: row.avatarColor }">
                      {{ row.name.charAt(0) }}
                    </el-avatar>
                    <span>{{ row.name }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="email" :label="t('collaboration.email')" width="250" />
              <el-table-column prop="permission" :label="t('collaboration.permissions')" width="150">
                <template #default="{ row }">
                  <el-tag :type="row.permission === 'admin' ? 'danger' : 'primary'">
                    {{ t(`collaboration.${row.permission}`) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="joinTime" :label="t('collaboration.joinTime')" width="180" />
              <el-table-column :label="t('common.operation')" width="150">
                <template #default="{ row }">
                  <el-button
                    v-if="currentUserPermission === 'admin'"
                    text
                    type="danger"
                    @click="handleRemoveMember(row)"
                  >
                    {{ t('common.delete') }}
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 邀请成员对话框 -->
    <el-dialog v-model="showInviteDialog" :title="t('collaboration.invite')" width="500px">
      <el-form :model="inviteForm" label-width="100px">
        <el-form-item :label="t('collaboration.email')">
          <el-input v-model="inviteForm.email" :placeholder="t('collaboration.emailPlaceholder')" clearable />
        </el-form-item>
        <el-form-item :label="t('collaboration.permissions')">
          <el-select v-model="inviteForm.permission" style="width: 100%">
            <el-option :label="t('collaboration.view')" value="view" />
            <el-option :label="t('collaboration.edit')" value="edit" />
            <el-option :label="t('collaboration.admin')" value="admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showInviteDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleInvite">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from '@/composables'

  const { t } = useI18n()
  const { proxy } = getCurrentInstance() as ComponentInternalInstance

  const activeTab = ref('workspace')
  const showInviteDialog = ref(false)
  const currentUserPermission = ref('admin') // 当前用户权限

  const inviteForm = reactive({
    email: '',
    permission: 'view'
  })

  const workspaces = ref([
    {
      id: 1,
      name: '项目文档',
      memberCount: 5,
      fileCount: 23,
      permission: 'admin',
      updateTime: new Date()
    },
    {
      id: 2,
      name: '设计资源',
      memberCount: 3,
      fileCount: 12,
      permission: 'edit',
      updateTime: new Date(Date.now() - 86400000)
    }
  ])

  const teamMembers = ref([
    {
      id: 1,
      name: '张三',
      email: 'zhangsan@example.com',
      permission: 'admin',
      avatarColor: '#6366f1',
      joinTime: '2024-01-01'
    },
    {
      id: 2,
      name: '李四',
      email: 'lisi@example.com',
      permission: 'edit',
      avatarColor: '#8b5cf6',
      joinTime: '2024-01-15'
    }
  ])

  const selectWorkspace = (workspace: any) => {
    proxy?.$modal.msg(t('collaboration.selectWorkspace', { name: workspace.name }))
  }

  const handleInvite = () => {
    if (!inviteForm.email) {
      proxy?.$modal.msgWarning(t('collaboration.emailPlaceholder'))
      return
    }
    proxy?.$modal.msgSuccess(t('collaboration.inviteSuccess'))
    showInviteDialog.value = false
    inviteForm.email = ''
    inviteForm.permission = 'view'
  }

  const handleRemoveMember = (member: any) => {
    proxy?.$modal.confirm(t('collaboration.removeMemberConfirm', { name: member.name })).then(() => {
      proxy?.$modal.msgSuccess(t('collaboration.removeMemberSuccess'))
    })
  }

  const formatTime = (date: Date) => {
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    const days = Math.floor(diff / 86400000)

    if (days === 0) return t('collaboration.today')
    if (days === 1) return t('collaboration.yesterday')
    if (days < 7) return t('collaboration.daysAgo', { days })
    return date.toLocaleDateString()
  }
</script>

<style scoped>
  .collaboration-page {
    height: 100%;
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 4px;
  }

  .page-header-card {
    flex-shrink: 0;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .page-icon {
    color: var(--primary-color);
  }

  .page-header h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 700;
    color: var(--text-primary);
  }

  .content-card {
    flex: 1;
    overflow: hidden;
  }

  .collaboration-tabs {
    height: 100%;
  }

  .collaboration-tabs :deep(.el-tabs__content) {
    height: calc(100% - 55px);
    overflow: auto;
  }

  .workspace-container {
    padding: 16px;
  }

  .workspace-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 16px;
  }

  .workspace-card {
    cursor: pointer;
    transition: all 0.3s;
  }

  .workspace-card:hover {
    transform: translateY(-4px);
  }

  .workspace-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;
  }

  .workspace-header h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
  }

  .workspace-info {
    display: flex;
    gap: 16px;
    margin-bottom: 12px;
    color: var(--text-secondary);
    font-size: 14px;
  }

  .info-item {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .workspace-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-top: 12px;
    border-top: 1px solid var(--border-light);
  }

  .update-time {
    font-size: 12px;
    color: var(--text-secondary);
  }

  .team-container {
    padding: 16px;
  }

  .member-cell {
    display: flex;
    align-items: center;
    gap: 8px;
  }
</style>

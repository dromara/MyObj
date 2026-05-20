import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
// 路由后置守卫：更新文档标题和 SEO
import { useAppStore } from '@/stores/app'
import { useSEO } from '@/composables'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/share/:token',
    name: 'ShareDownload',
    component: () => import('@/views/ShareDownload/index.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/layout/index.vue'),
    redirect: '/files',
    meta: { requiresAuth: true },
    children: [
      {
        path: '/files',
        name: 'Files',
        component: () => import('@/views/Files/index.vue'),
        meta: { title: '我的文件', i18nKey: 'route.files' }
      },
      {
        path: '/shares',
        name: 'Shares',
        component: () => import('@/views/Shares/index.vue'),
        meta: { title: '我的分享', i18nKey: 'route.shares' }
      },
      {
        path: '/offline',
        name: 'Offline',
        component: () => import('@/views/Offline/index.vue'),
        meta: { title: '离线下载', i18nKey: 'route.offline' }
      },
      {
        path: '/tasks',
        name: 'Tasks',
        component: () => import('@/views/Tasks/index.vue'),
        meta: { title: '任务中心', i18nKey: 'route.tasks' }
      },
      {
        path: '/trash',
        name: 'Trash',
        component: () => import('@/views/Trash/index.vue'),
        meta: { title: '回收站', i18nKey: 'route.trash' }
      },
      {
        path: '/square',
        name: 'Square',
        component: () => import('@/views/Square/index.vue'),
        meta: { title: '文件广场', i18nKey: 'route.square' }
      },
      {
        path: '/settings',
        name: 'Settings',
        component: () => import('@/views/Settings/index.vue'),
        meta: { title: '系统设置', i18nKey: 'route.settings' }
      },
      // 协作功能暂时隐藏
      // {
      //   path: '/collaboration',
      //   name: 'Collaboration',
      //   component: () => import('@/views/Collaboration/index.vue'),
      //   meta: { title: '协作', i18nKey: 'route.collaboration' }
      // },
      {
        path: '/enterprise',
        name: 'Enterprise',
        component: () => import('@/views/Enterprise/Layout.vue'),
        redirect: '/enterprise/list',
        meta: { title: '企业空间', i18nKey: 'route.enterprise' },
        children: [
          {
            path: 'list',
            name: 'EnterpriseList',
            component: () => import('@/views/Enterprise/List.vue'),
            meta: { title: '企业列表', i18nKey: 'route.enterpriseList' }
          },
          {
            path: ':id/members',
            name: 'EnterpriseMembers',
            component: () => import('@/views/Enterprise/Manage/Members.vue'),
            meta: { title: '成员管理', i18nKey: 'route.enterpriseMembers' }
          },
          {
            path: ':id/roles',
            name: 'EnterpriseRoles',
            component: () => import('@/views/Enterprise/Manage/Roles.vue'),
            meta: { title: '角色管理', i18nKey: 'route.enterpriseRoles' }
          },
          {
            path: ':id/space',
            name: 'EnterpriseSpace',
            component: () => import('@/views/Enterprise/Space/index.vue'),
            meta: { title: '共享空间', i18nKey: 'route.enterpriseSpace' }
          },
          {
            path: ':id/audit',
            name: 'EnterpriseAudit',
            component: () => import('@/views/Enterprise/Manage/Audit.vue'),
            meta: { title: '审计日志', i18nKey: 'route.enterpriseAudit' }
          },
          {
            path: ':id/settings',
            name: 'EnterpriseSettings',
            component: () => import('@/views/Enterprise/Settings/index.vue'),
            meta: { title: '企业设置', i18nKey: 'route.enterpriseSettings' }
          }
        ]
      },
      {
        path: '/admin',
        name: 'Admin',
        component: () => import('@/views/Admin/index.vue'),
        meta: { title: '系统管理', i18nKey: 'route.admin', requiresAdmin: true },
        redirect: '/admin/users',
        children: [
          {
            path: 'users',
            name: 'AdminUsers',
            component: () => import('@/views/Admin/Users/index.vue'),
            meta: { title: '用户管理', i18nKey: 'route.adminUsers' }
          },
          {
            path: 'groups',
            name: 'AdminGroups',
            component: () => import('@/views/Admin/Groups/index.vue'),
            meta: { title: '组管理', i18nKey: 'route.adminGroups' }
          },
          {
            path: 'permissions',
            name: 'AdminPermissions',
            component: () => import('@/views/Admin/Permissions/index.vue'),
            meta: { title: '权限管理', i18nKey: 'route.adminPermissions' }
          },
          {
            path: 'disks',
            name: 'AdminDisks',
            component: () => import('@/views/Admin/Disks/index.vue'),
            meta: { title: '磁盘管理', i18nKey: 'route.adminDisks' }
          },
          {
            path: 'system',
            name: 'AdminSystem',
            component: () => import('@/views/Admin/System/index.vue'),
            meta: { title: '系统配置', i18nKey: 'route.adminSystem' }
          },
          {
            path: 'space-config',
            name: 'AdminSpaceConfig',
            component: () => import('@/views/Admin/SpaceConfig/index.vue'),
            meta: { title: '空间配置', i18nKey: 'route.adminSpaceConfig' }
          },
          {
            path: 'audit',
            name: 'AdminAuditLog',
            component: () => import('@/views/Admin/AuditLog/index.vue'),
            meta: { title: '审计日志', i18nKey: 'route.adminAuditLog' }
          }
        ]
      }
    ]
  },
  {
    path: '/redirect',
    component: () => import('@/layout/index.vue'),
    meta: { hidden: true },
    children: [
      {
        path: '/redirect/:path(.*)',
        component: () => import('@/views/Redirect/index.vue'),
        meta: { title: '重定向', i18nKey: 'route.redirect', hidden: true }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/files'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach(async (to) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.token) {
    return '/login'
  }
  if (to.path === '/login' && authStore.token) {
    return '/files'
  }
  if (to.meta.requiresAdmin) {
    const { useAdmin } = await import('@/composables/business/useAdmin')
    const { isAdmin } = useAdmin()
    if (!isAdmin.value) {
      return '/files'
    }
  }
  return true
})

router.afterEach(to => {
  const appStore = useAppStore()
  appStore.updateDocumentTitle()

  // 更新 SEO 信息
  const seo = useSEO({
    title: (to.meta.title as string) || 'MyObj 网盘系统',
    description: (to.meta.description as string) || 'MyObj 网盘系统 - 安全、高效的文件存储和管理平台'
  })
  seo.applySEO()
})

export default router

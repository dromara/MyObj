import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

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
        meta: { title: '我的文件' }
      },
      {
        path: '/shares',
        name: 'Shares',
        component: () => import('@/views/Shares/index.vue'),
        meta: { title: '我的分享' }
      },
      {
        path: '/offline',
        name: 'Offline',
        component: () => import('@/views/Offline/index.vue'),
        meta: { title: '离线下载' }
      },
      {
        path: '/tasks',
        name: 'Tasks',
        component: () => import('@/views/Tasks/index.vue'),
        meta: { title: '任务中心' }
      },
      {
        path: '/trash',
        name: 'Trash',
        component: () => import('@/views/Trash/index.vue'),
        meta: { title: '回收站' }
      },
      {
        path: '/square',
        name: 'Square',
        component: () => import('@/views/Square/index.vue'),
        meta: { title: '文件广场' }
      },
      {
        path: '/settings',
        name: 'Settings',
        component: () => import('@/views/Settings/index.vue'),
        meta: { title: '系统设置' }
      },
      {
        path: '/admin',
        name: 'Admin',
        component: () => import('@/views/Admin/index.vue'),
        meta: { title: '系统管理', requiresAdmin: true },
        redirect: '/admin/users',
        children: [
          {
            path: 'users',
            name: 'AdminUsers',
            component: () => import('@/views/Admin/Users/index.vue'),
            meta: { title: '用户管理' }
          },
          {
            path: 'groups',
            name: 'AdminGroups',
            component: () => import('@/views/Admin/Groups/index.vue'),
            meta: { title: '组管理' }
          },
          {
            path: 'permissions',
            name: 'AdminPermissions',
            component: () => import('@/views/Admin/Permissions/index.vue'),
            meta: { title: '权限管理' }
          },
          {
            path: 'disks',
            name: 'AdminDisks',
            component: () => import('@/views/Admin/Disks/index.vue'),
            meta: { title: '磁盘管理' }
          },
          {
            path: 'system',
            name: 'AdminSystem',
            component: () => import('@/views/Admin/System/index.vue'),
            meta: { title: '系统配置' }
          }
        ]
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
router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()
  
  if (to.meta.requiresAuth && !authStore.token) {
    next('/login')
  } else if (to.path === '/login' && authStore.token) {
    next('/files')
  } else if (to.meta.requiresAdmin) {
    // 检查管理员权限
    const { useAdmin } = await import('@/composables/useAdmin')
    const { isAdmin } = useAdmin()
    if (!isAdmin.value) {
      next('/files')
    } else {
      next()
    }
  } else {
    next()
  }
})

export default router

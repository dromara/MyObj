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
  } else {
    next()
  }
})

export default router

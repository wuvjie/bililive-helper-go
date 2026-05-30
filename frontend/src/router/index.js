import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue'),
    meta: { title: '仪表盘', icon: 'chart-bar' }
  },
  {
    path: '/streamers',
    name: 'Streamers',
    component: () => import('../views/Streamers.vue'),
    meta: { title: '主播管理', icon: 'users' }
  },
  {
    path: '/tasks',
    name: 'Tasks',
    component: () => import('../views/Tasks.vue'),
    meta: { title: '任务中心', icon: 'refresh' }
  },
  {
    path: '/logs',
    name: 'Logs',
    component: () => import('../views/Logs.vue'),
    meta: { title: '操作日志', icon: 'document-text' }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('../views/Settings.vue'),
    meta: { title: '系统设置', icon: 'cog' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router

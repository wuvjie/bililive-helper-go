import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  { path: '/', component: () => import('../views/Dashboard.vue'), name: 'Dashboard' },
  { path: '/streamers', component: () => import('../views/Streamers.vue'), name: 'Streamers' },
  { path: '/tasks', component: () => import('../views/Tasks.vue'), name: 'Tasks' },
  { path: '/history', component: () => import('../views/Logs.vue'), name: 'History' },
  { path: '/settings', component: () => import('../views/Settings.vue'), name: 'Settings' }
]

export const router = createRouter({
  history: createWebHashHistory(),
  routes
})

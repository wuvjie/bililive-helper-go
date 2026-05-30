<script setup>
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '../stores/app'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()

const navItems = [
  {
    path: '/dashboard',
    name: '仪表盘',
    icon: 'chart-bar',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/></svg>`
  },
  {
    path: '/streamers',
    name: '主播管理',
    icon: 'users',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 4.354a4 4 0 110 7.292 15.242 15.242 0 01-4.574 2.077M19 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75M9 7a4 4 0 11-8 0 4 4 0 018 0zM5 21v-2a4 4 0 014-4h6a4 4 0 014 4v2"/></svg>`
  },
  {
    path: '/tasks',
    name: '任务中心',
    icon: 'refresh',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>`
  },
  {
    path: '/logs',
    name: '操作日志',
    icon: 'document-text',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>`
  },
  {
    path: '/settings',
    name: '系统设置',
    icon: 'cog',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/><path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/></svg>`
  }
]

const sidebarWidth = computed(() => appStore.sidebarExpanded ? 'w-[220px]' : 'w-[64px]')

function isActive(path) {
  return route.path === path || route.path.startsWith(path + '/')
}

function navigateTo(path) {
  router.push(path)
}
</script>

<template>
  <aside
    :class="[
      'h-screen bg-[#1e2126] flex flex-col transition-all duration-300 ease-in-out fixed left-0 top-0 z-40',
      sidebarWidth,
      appStore.isMobile ? 'hidden' : ''
    ]"
  >
    <!-- Logo 区域 -->
    <div class="h-14 flex items-center px-4 border-b border-white/10">
      <div class="w-8 h-8 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center flex-shrink-0">
        <svg class="w-5 h-5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/>
        </svg>
      </div>
      <Transition name="fade">
        <span v-if="appStore.sidebarExpanded" class="ml-3 text-white font-semibold text-sm whitespace-nowrap">
          Bililive Helper
        </span>
      </Transition>
    </div>

    <!-- 导航菜单 -->
    <nav class="flex-1 py-4 overflow-y-auto">
      <div class="space-y-1 px-2">
        <button
          v-for="item in navItems"
          :key="item.path"
          @click="navigateTo(item.path)"
          :class="[
            'w-full flex items-center rounded-lg transition-all duration-200 group',
            appStore.sidebarExpanded ? 'px-3 py-2.5' : 'px-0 py-2.5 justify-center',
            isActive(item.path)
              ? 'bg-white/15 text-white'
              : 'text-white/60 hover:bg-white/10 hover:text-white/90'
          ]"
        >
          <div
            v-html="item.svg"
            :class="[
              'flex-shrink-0 w-5 h-5',
              isActive(item.path) ? 'text-white' : 'text-white/60 group-hover:text-white/90'
            ]"
          ></div>
          <Transition name="fade">
            <span
              v-if="appStore.sidebarExpanded"
              :class="[
                'ml-3 text-sm font-medium whitespace-nowrap',
                isActive(item.path) ? 'text-white' : 'text-white/60 group-hover:text-white/90'
              ]"
            >
              {{ item.name }}
            </span>
          </Transition>

          <!-- 选中指示器 -->
          <div
            v-if="isActive(item.path) && !appStore.sidebarExpanded"
            class="absolute left-0 w-1 h-6 bg-blue-500 rounded-r"
          ></div>
        </button>
      </div>
    </nav>

    <!-- 底部展开/收起按钮 -->
    <div class="p-2 border-t border-white/10">
      <button
        @click="appStore.toggleSidebar()"
        class="w-full flex items-center justify-center py-2 rounded-lg text-white/50 hover:bg-white/10 hover:text-white/80 transition-all"
      >
        <svg
          :class="[
            'w-5 h-5 transition-transform duration-300',
            appStore.sidebarExpanded ? 'rotate-180' : ''
          ]"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <path d="M11 19l-7-7 7-7m8 14l-7-7 7-7"/>
        </svg>
      </button>
    </div>
  </aside>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

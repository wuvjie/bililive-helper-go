<script setup>
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '../stores/app'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()

const navItems = [
  { path: '/dashboard', name: '仪表盘', svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/></svg>' },
  { path: '/streamers', name: '主播管理', svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 4.354a4 4 0 110 7.292 15.242 0 01-4.574 2.077M19 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75M9 7a4 4 0 11-8 0 4 4 0 018 0z"/></svg>' },
  { path: '/tasks', name: '任务中心', svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>' },
  { path: '/logs', name: '操作日志', svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>' },
  { path: '/settings', name: '系统设置', svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37"/><path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/></svg>' },
]

const sidebarWidth = computed(() => appStore.sidebarExpanded ? 'w-[200px]' : 'w-[64px]')

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
      'h-screen bg-white border-r border-[#dee0e3] flex flex-col transition-all duration-200 ease-in-out fixed left-0 top-0 z-40',
      sidebarWidth,
      appStore.isMobile ? 'hidden' : ''
    ]"
  >
    <!-- Logo -->
    <div class="h-14 flex items-center px-4 border-b border-[#f0f1f5]">
      <div class="w-8 h-8 bg-[#3370ff] rounded-lg flex items-center justify-center flex-shrink-0">
        <svg class="w-5 h-5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/>
        </svg>
      </div>
      <transition name="fade">
        <span v-if="appStore.sidebarExpanded" class="ml-3 text-[15px] font-semibold text-[#1f2329] whitespace-nowrap">
          Bililive Helper
        </span>
      </transition>
    </div>

    <!-- Nav -->
    <nav class="flex-1 py-3 overflow-y-auto">
      <div class="space-y-0.5 px-2">
        <button
          v-for="item in navItems"
          :key="item.path"
          @click="navigateTo(item.path)"
          :class="[
            'w-full flex items-center rounded-lg transition-all duration-150 relative',
            appStore.sidebarExpanded ? 'px-3 py-2.5 gap-3' : 'px-0 py-2.5 justify-center',
            isActive(item.path)
              ? 'bg-[#e8f3ff] text-[#3370ff]'
              : 'text-[#4e5969] hover:bg-[#f2f3f5] hover:text-[#1f2329]'
          ]"
        >
          <div v-html="item.svg" class="w-[18px] h-[18px] flex-shrink-0"></div>
          <transition name="fade">
            <span v-if="appStore.sidebarExpanded" class="text-[14px] whitespace-nowrap">{{ item.name }}</span>
          </transition>
          <div v-if="isActive(item.path)" class="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-[18px] bg-[#3370ff] rounded-r"></div>
        </button>
      </div>
    </nav>

    <!-- Collapse -->
    <div class="p-2 border-t border-[#f0f1f5]">
      <button @click="appStore.toggleSidebar()" class="w-full flex items-center py-2 px-3 rounded-lg text-[#86909c] hover:bg-[#f2f3f5] hover:text-[#4e5969] transition-all">
        <svg :class="['w-[18px] h-[18px] transition-transform duration-200 flex-shrink-0', appStore.sidebarExpanded ? 'rotate-180' : '']" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M11 19l-7-7 7-7m8 14l-7-7 7-7"/>
        </svg>
        <transition name="fade">
          <span v-if="appStore.sidebarExpanded" class="ml-2 text-[13px] whitespace-nowrap">收起导航</span>
        </transition>
      </button>
    </div>
  </aside>
</template>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.15s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>

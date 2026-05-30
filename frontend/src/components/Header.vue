<script setup>
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '../stores/app'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()

const sidebarWidth = computed(() => {
  if (appStore.isMobile) return '0px'
  return appStore.sidebarExpanded ? '200px' : '64px'
})

const tabs = [
  { path: '/dashboard', name: '仪表盘' },
  { path: '/streamers', name: '主播管理' },
  { path: '/tasks', name: '任务中心' },
]

function isActive(path) {
  return route.path === path || route.path.startsWith(path + '/')
}
</script>

<template>
  <header
    class="h-14 bg-white border-b border-[#dee0e3] flex items-center justify-between px-4 fixed top-0 right-0 z-50 transition-all duration-200"
    :style="{ left: sidebarWidth }"
  >
    <!-- 左：标签页 -->
    <div class="flex items-center gap-6">
      <nav class="hidden md:flex items-center gap-1">
        <button
          v-for="tab in tabs"
          :key="tab.path"
          @click="router.push(tab.path)"
          :class="[
            'px-3 py-1.5 text-[14px] rounded-md transition-colors',
            isActive(tab.path)
              ? 'text-[#3370ff] bg-[#e8f3ff] font-medium'
              : 'text-[#4e5969] hover:text-[#1f2329] hover:bg-[#f2f3f5]'
          ]"
        >
          {{ tab.name }}
        </button>
      </nav>
    </div>

    <!-- 右：搜索 + 图标 + 用户 -->
    <div class="flex items-center gap-2">
      <div class="relative hidden lg:block">
        <svg class="absolute left-2.5 top-1/2 -translate-y-1/2 w-[14px] h-[14px] text-[#86909c]" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/></svg>
        <input type="text" placeholder="搜索功能、配置..." class="w-64 pl-8 pr-3 py-[6px] bg-[#f2f3f5] border border-[#e5e6eb] rounded-lg text-[13px] text-[#1f2329] placeholder:text-[#86909c] focus:outline-none focus:border-[#3370ff] focus:bg-white transition-all">
      </div>
      <button class="w-8 h-8 rounded-lg flex items-center justify-center text-[#4e5969] hover:bg-[#f2f3f5] transition-colors">
        <svg class="w-[18px] h-[18px]" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
      </button>
      <button class="w-8 h-8 rounded-lg flex items-center justify-center text-[#4e5969] hover:bg-[#f2f3f5] transition-colors relative">
        <svg class="w-[18px] h-[18px]" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"/></svg>
        <span v-if="appStore.notifications > 0" class="absolute -top-0.5 -right-0.5 w-4 h-4 bg-[#f54a45] text-white text-[10px] font-bold rounded-full flex items-center justify-center leading-none">
          {{ appStore.notifications > 99 ? '99+' : appStore.notifications }}
        </span>
      </button>
      <button class="w-8 h-8 rounded-lg flex items-center justify-center text-[#4e5969] hover:bg-[#f2f3f5] transition-colors">
        <svg class="w-[18px] h-[18px]" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37"/><path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
      </button>
      <div class="w-px h-5 bg-[#e5e6eb] mx-1"></div>
      <div class="flex items-center gap-2 cursor-pointer hover:bg-[#f2f3f5] rounded-lg px-2 py-1 transition-colors">
        <div class="w-7 h-7 bg-[#3370ff] rounded-full flex items-center justify-center">
          <span class="text-white text-[12px] font-semibold">管</span>
        </div>
        <span class="text-[13px] text-[#1f2329] hidden xl:block">管理员</span>
        <svg class="w-3 h-3 text-[#86909c]" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M19 9l-7 7-7-7"/></svg>
      </div>
    </div>
  </header>
</template>

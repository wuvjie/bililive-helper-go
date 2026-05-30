<script setup>
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '../stores/app'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()

const props = defineProps({
  currentUser: {
    type: Object,
    default: () => ({
      name: '管理员',
      avatar: null
    })
  }
})

const initials = computed(() => {
  return props.currentUser.name.charAt(0).toUpperCase()
})

// 侧边栏宽度
const sidebarWidth = computed(() => {
  if (appStore.isMobile) return '0px'
  return appStore.sidebarExpanded ? '200px' : '64px'
})

// 顶部标签页导航
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
    class="h-14 bg-white border-b border-gray-200 flex items-center justify-between px-4 fixed top-0 right-0 z-30 transition-all duration-300"
    :style="{ left: sidebarWidth }"
  >
    <!-- 左侧：Logo + 标签页 -->
    <div class="flex items-center gap-6">
      <!-- Logo -->
      <div class="flex items-center gap-2">
        <div class="w-8 h-8 bg-[#3370ff] rounded-lg flex items-center justify-center">
          <svg class="w-5 h-5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/>
          </svg>
        </div>
        <span class="text-base font-semibold text-gray-900">Bililive Helper</span>
      </div>

      <!-- 标签页导航 -->
      <nav class="hidden md:flex items-center gap-1">
        <button
          v-for="tab in tabs"
          :key="tab.path"
          @click="router.push(tab.path)"
          :class="[
            'px-4 py-2 text-sm font-medium rounded-lg transition-colors',
            isActive(tab.path)
              ? 'bg-[#e8f3ff] text-[#3370ff]'
              : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
          ]"
        >
          {{ tab.name }}
        </button>
      </nav>
    </div>

    <!-- 右侧：搜索 + 图标 + 用户 -->
    <div class="flex items-center gap-3">
      <!-- 搜索框 -->
      <div class="relative hidden lg:block">
        <svg class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
        </svg>
        <input
          type="text"
          placeholder="搜索功能、配置..."
          class="w-72 pl-10 pr-4 py-2 bg-gray-100 border-0 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:bg-white transition-all"
        >
      </div>

      <!-- 图标按钮 -->
      <button class="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
      </button>
      <button class="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors relative">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"/>
        </svg>
        <span
          v-if="appStore.notifications > 0"
          class="absolute top-1 right-1 w-4 h-4 bg-red-500 text-white text-xs font-bold rounded-full flex items-center justify-center"
        >
          {{ appStore.notifications > 99 ? '99+' : appStore.notifications }}
        </span>
      </button>
      <button class="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
        </svg>
      </button>

      <!-- 分隔线 -->
      <div class="w-px h-6 bg-gray-200"></div>

      <!-- 用户头像 -->
      <div class="flex items-center gap-2 cursor-pointer hover:bg-gray-100 rounded-lg px-2 py-1 transition-colors">
        <div class="w-8 h-8 bg-[#3370ff] rounded-full flex items-center justify-center">
          <span class="text-white text-sm font-semibold">{{ initials }}</span>
        </div>
        <div class="hidden xl:block">
          <p class="text-sm font-medium text-gray-700">{{ currentUser.name }}</p>
        </div>
        <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
        </svg>
      </div>
    </div>
  </header>
</template>

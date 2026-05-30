<script setup>
import { computed } from 'vue'
import { useAppStore } from '../stores/app'

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
  return appStore.sidebarExpanded ? '220px' : '64px'
})
</script>

<template>
  <header
    class="h-14 bg-white border-b border-gray-200 flex items-center justify-between px-6 fixed top-0 right-0 z-30 shadow-sm transition-all duration-300"
    :style="{ left: sidebarWidth }"
  >

    <!-- 左侧：面包屑 / 页面标题 -->
    <div class="flex items-center">
      <slot name="left">
        <h1 class="text-base font-semibold text-gray-800">Bililive Helper</h1>
      </slot>
    </div>

    <!-- 右侧：状态 + 搜索 + 通知 + 用户 -->
    <div class="flex items-center gap-4">

      <!-- 搜索框 -->
      <div class="relative hidden md:block">
        <svg class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
        </svg>
        <input
          type="text"
          placeholder="搜索... (⌘K)"
          class="w-64 pl-10 pr-4 py-2 bg-gray-100 border-0 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:bg-white transition-all"
        >
      </div>

      <!-- 运行状态指示器 -->
      <div
        :class="[
          'flex items-center gap-2 px-3 py-1.5 rounded-full text-sm font-medium',
          appStore.isRunning
            ? 'bg-green-50 text-green-700'
            : 'bg-gray-100 text-gray-500'
        ]"
      >
        <span
          :class="[
            'w-2 h-2 rounded-full',
            appStore.isRunning ? 'bg-green-500 animate-pulse' : 'bg-gray-400'
          ]"
        ></span>
        {{ appStore.isRunning ? '运行中' : '空闲' }}
      </div>

      <!-- 通知铃铛 -->
      <button class="relative p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"/>
        </svg>
        <span
          v-if="appStore.notifications > 0"
          class="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-xs font-bold rounded-full flex items-center justify-center"
        >
          {{ appStore.notifications > 99 ? '99+' : appStore.notifications }}
        </span>
      </button>

      <!-- 分隔线 -->
      <div class="w-px h-6 bg-gray-200"></div>

      <!-- 用户头像 -->
      <div class="flex items-center gap-3 cursor-pointer hover:bg-gray-100 rounded-lg px-2 py-1 transition-colors">
        <div class="w-8 h-8 bg-gradient-to-br from-blue-500 to-blue-600 rounded-full flex items-center justify-center">
          <span class="text-white text-sm font-semibold">{{ initials }}</span>
        </div>
        <div class="hidden lg:block">
          <p class="text-sm font-medium text-gray-700">{{ currentUser.name }}</p>
        </div>
        <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
        </svg>
      </div>
    </div>
  </header>
</template>

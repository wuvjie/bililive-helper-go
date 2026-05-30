<script setup>
import { computed, onMounted } from 'vue'
import { useAppStore } from './stores/app'
import { useStreamerData } from './composables/useStreamerData'
import { useConfig } from './composables/useConfig'
import Sidebar from './components/Sidebar.vue'
import Header from './components/Header.vue'

const appStore = useAppStore()

// 初始化数据
const {
  streamers,
  diskUsage,
  totalGB,
  detail,
  fetchStatus,
  fetchDetail
} = useStreamerData()

const {
  config,
  schedule,
  fetchConfig,
  fetchSchedule
} = useConfig()

// 侧边栏宽度（动态，移动端无边距）
const sidebarWidth = computed(() => {
  if (appStore.isMobile) return '0px'
  return appStore.sidebarExpanded ? '200px' : '64px'
})

// 移动端底部导航项
const mobileNavItems = [
  { path: '/dashboard', name: '首页', svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/></svg>' },
  { path: '/streamers', name: '主播', svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 4.354a4 4 0 110 7.292 15.242 0 01-4.574 2.077M19 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/></svg>' },
  { path: '/tasks', name: '任务', svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>' },
  { path: '/settings', name: '设置', svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/><path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/></svg>' }
]

// 初始化
onMounted(async () => {
  await fetchConfig()
  await fetchStatus()
  await fetchSchedule()
  await fetchDetail()

  // 自动刷新
  setInterval(async () => {
    if (!appStore.isRunning) {
      await fetchDetail()
    }
  }, 60000)
})
</script>

<template>
  <div class="min-h-screen bg-[#f5f6f8]">

    <!-- 侧边栏 -->
    <Sidebar />

    <!-- 主内容区 -->
    <div class="transition-all duration-300" :style="{ marginLeft: sidebarWidth }">

      <!-- 顶部导航栏 -->
      <Header />

      <!-- 页面内容 -->
      <main class="min-h-screen bg-[#f5f6f8]" :style="{ marginTop: '56px', paddingLeft: appStore.isMobile ? '0' : '0', paddingBottom: appStore.isMobile ? '64px' : '0' }">
        <router-view
          v-slot="{ Component }"
          :streamers="streamers"
          :diskUsage="diskUsage"
          :totalGB="totalGB"
          :detail="detail"
          :config="config"
          :schedule="schedule"
        >
          <transition name="page" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>

    <!-- 移动端底部导航 -->
    <nav v-if="appStore.isMobile" class="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 z-40">
      <div class="flex items-center justify-around py-2">
        <router-link
          v-for="item in mobileNavItems"
          :key="item.path"
          :to="item.path"
          class="flex flex-col items-center py-1 px-3 text-gray-500 hover:text-blue-600"
          active-class="text-blue-600"
        >
          <div v-html="item.svg" class="w-6 h-6 mb-1"></div>
          <span class="text-xs">{{ item.name }}</span>
        </router-link>
      </div>
    </nav>
  </div>
</template>

<style>
/* 页面过渡动画 */
.page-enter-active,
.page-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.page-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.page-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* 全局滚动条样式 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}
</style>

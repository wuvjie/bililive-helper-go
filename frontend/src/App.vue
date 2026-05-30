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

// 侧边栏宽度（动态）
const sidebarWidth = computed(() => appStore.sidebarExpanded ? 'ml-[220px]' : 'ml-[64px]')

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
    <div :class="['transition-all duration-300', sidebarWidth]">

      <!-- 顶部导航栏 -->
      <Header />

      <!-- 页面内容 -->
      <main class="pt-14 min-h-screen">
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

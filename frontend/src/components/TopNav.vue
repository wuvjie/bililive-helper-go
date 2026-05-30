<script setup>
defineProps({
  running: {
    type: Boolean,
    default: false
  }
})
</script>

<template>
  <header class="sticky top-0 z-40 bg-white/80 backdrop-blur-lg border-b border-gray-200">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex items-center justify-between h-16">

        <!-- 左侧：Logo + 标题 -->
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl flex items-center justify-center shadow-lg">
            <svg class="w-6 h-6 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/>
            </svg>
          </div>
          <h1 class="text-xl font-bold text-gray-900 tracking-tight">Bililive Helper</h1>
        </div>

        <!-- 右侧：状态 + 退出 -->
        <div class="flex items-center gap-4">
          <!-- 运行状态指示器 -->
          <Transition name="fade">
            <div v-if="running" class="flex items-center gap-2 px-3 py-1.5 bg-amber-50 text-amber-700 rounded-full text-sm font-medium">
              <span class="w-2 h-2 bg-amber-500 rounded-full animate-pulse"></span>
              运行中
            </div>
          </Transition>

          <!-- 退出登录 -->
          <a href="/logout" class="text-sm font-medium text-gray-500 hover:text-gray-700 transition-colors">
            退出登录
          </a>
        </div>
      </div>
    </div>
  </header>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: scale(0.9);
}
</style>

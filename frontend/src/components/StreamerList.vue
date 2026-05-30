<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  streamers: {
    type: Array,
    default: () => []
  },
  totalGB: {
    type: Number,
    default: 1
  },
  running: {
    type: Boolean,
    default: false
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['run', 'openManualMerge', 'refresh'])

const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = 10

const filtered = computed(() => {
  const kw = searchQuery.value.toLowerCase().trim()
  let list = props.streamers || []

  if (kw) {
    list = list.filter(s => s.name.toLowerCase().includes(kw))
  }

  return list
})

const totalPages = computed(() => Math.max(1, Math.ceil(filtered.value.length / pageSize)))

const paginated = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filtered.value.slice(start, start + pageSize)
})

watch(totalPages, (newVal) => {
  if (currentPage.value > newVal) {
    currentPage.value = Math.max(1, newVal)
  }
})

function prevPage() {
  currentPage.value = Math.max(1, currentPage.value - 1)
}

function nextPage() {
  currentPage.value = Math.min(totalPages.value, currentPage.value + 1)
}

function getUsagePercent(sizeGB) {
  return Math.min(100, (sizeGB / props.totalGB) * 100)
}

function getUsageColor(sizeGB) {
  const percent = (sizeGB / props.totalGB) * 100
  if (percent > 15) return 'bg-red-500'
  if (percent > 10) return 'bg-amber-500'
  return 'bg-blue-500'
}
</script>

<template>
  <div class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
    <!-- 顶部工具栏 -->
    <div class="p-4 border-b border-gray-100">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-900">
          主播列表
          <span class="ml-2 text-sm font-normal text-gray-500">({{ filtered.length }})</span>
        </h2>
        <button
          @click="emit('refresh')"
          :disabled="loading"
          class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 disabled:opacity-50 transition-colors"
        >
          刷新
        </button>
      </div>

      <!-- 搜索框 -->
      <div class="relative">
        <svg class="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
        </svg>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索主播..."
          class="w-full pl-10 pr-4 py-3 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
        >
      </div>
    </div>

    <!-- 主播卡片列表 -->
    <div class="p-4 space-y-3 max-h-[600px] overflow-y-auto">
      <div v-if="paginated.length === 0" class="flex flex-col items-center justify-center py-12 text-gray-400">
        <svg class="w-12 h-12 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"/>
        </svg>
        <p class="text-sm">未找到匹配的主播</p>
      </div>

      <div
        v-for="s in paginated"
        :key="s.name"
        class="p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
      >
        <div class="flex items-start justify-between">
          <!-- 左侧信息 -->
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 mb-2">
              <span :class="[
                'w-2 h-2 rounded-full flex-shrink-0',
                s.is_running ? 'bg-green-500 animate-pulse' : 'bg-gray-300'
              ]"></span>
              <h3 class="font-medium text-gray-900 truncate">{{ s.name }}</h3>
            </div>

            <div class="flex items-center gap-4 text-sm text-gray-500 mb-3">
              <span class="flex items-center gap-1">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/>
                </svg>
                {{ s.size_gb.toFixed(1) }} GB
              </span>
              <span class="flex items-center gap-1">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                </svg>
                {{ s.files }} 文件
              </span>
            </div>

            <!-- 进度条 -->
            <div class="w-full bg-gray-200 rounded-full h-2">
              <div
                :class="['h-2 rounded-full transition-all duration-500', getUsageColor(s.size_gb)]"
                :style="{ width: getUsagePercent(s.size_gb) + '%' }"
              ></div>
            </div>
          </div>

          <!-- 右侧操作按钮 -->
          <div class="flex flex-col gap-2 ml-4">
            <button
              @click="emit('run', 'merge', s.name)"
              :disabled="running"
              class="px-4 py-2 text-xs font-medium text-white bg-blue-500 rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              合并
            </button>
            <button
              @click="emit('run', 'clean', s.name)"
              :disabled="running"
              class="px-4 py-2 text-xs font-medium text-red-600 bg-red-50 rounded-lg hover:bg-red-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              清理
            </button>
            <button
              @click="emit('openManualMerge', s.name)"
              :disabled="running"
              class="px-4 py-2 text-xs font-medium text-gray-600 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              手动
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="totalPages > 1" class="px-4 py-3 border-t border-gray-100 flex items-center justify-center gap-4">
      <button
        @click="prevPage"
        :disabled="currentPage <= 1"
        class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        上一页
      </button>
      <span class="text-sm text-gray-500">
        {{ currentPage }} / {{ totalPages }}
      </span>
      <button
        @click="nextPage"
        :disabled="currentPage >= totalPages"
        class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        下一页
      </button>
    </div>
  </div>
</template>

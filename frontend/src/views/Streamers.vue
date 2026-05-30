<script setup>
import { ref, computed } from 'vue'
import { useAppStore } from '../stores/app'

const props = defineProps({
  streamers: { type: Array, default: () => [] },
  totalGB: { type: Number, default: 1 },
  running: { type: Boolean, default: false }
})

const emit = defineEmits(['run', 'openManualMerge'])

const appStore = useAppStore()

const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = 12
const statusFilter = ref('all')

const filtered = computed(() => {
  let list = props.streamers || []

  // 搜索过滤
  const kw = searchQuery.value.toLowerCase().trim()
  if (kw) {
    list = list.filter(s => s.name.toLowerCase().includes(kw))
  }

  // 状态过滤
  if (statusFilter.value === 'running') {
    list = list.filter(s => s.is_running)
  } else if (statusFilter.value === 'stopped') {
    list = list.filter(s => !s.is_running)
  }

  return list
})

const totalPages = computed(() => Math.max(1, Math.ceil(filtered.value.length / pageSize)))

const paginated = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filtered.value.slice(start, start + pageSize)
})

function getUsageColor(sizeGB) {
  const percent = (sizeGB / props.totalGB) * 100
  if (percent > 15) return 'bg-red-500'
  if (percent > 10) return 'bg-amber-500'
  return 'bg-blue-500'
}

function getUsagePercent(sizeGB) {
  return Math.min(100, (sizeGB / props.totalGB) * 100)
}
</script>

<template>
  <div class="p-6">
    <!-- 面包屑 -->
    <div class="mb-6">
      <nav class="flex items-center text-sm text-gray-500">
        <span class="text-gray-400">首页</span>
        <svg class="w-4 h-4 mx-2 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
        </svg>
        <span class="text-gray-700 font-medium">主播管理</span>
      </nav>
      <h1 class="mt-2 text-2xl font-bold text-gray-900">主播管理</h1>
    </div>

    <!-- 操作栏 -->
    <div class="bg-white rounded-xl p-4 border border-gray-100 mb-6">
      <div class="flex flex-col md:flex-row md:items-center gap-4">
        <!-- 搜索框 -->
        <div class="relative flex-1">
          <svg class="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
          </svg>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索主播..."
            class="w-full pl-10 pr-4 py-2.5 bg-gray-50 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
        </div>

        <!-- 状态筛选 -->
        <div class="flex items-center gap-2">
          <button
            @click="statusFilter = 'all'"
            :class="[
              'px-4 py-2 text-sm font-medium rounded-lg transition-colors',
              statusFilter === 'all'
                ? 'bg-blue-500 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            ]"
          >
            全部 ({{ streamers.length }})
          </button>
          <button
            @click="statusFilter = 'running'"
            :class="[
              'px-4 py-2 text-sm font-medium rounded-lg transition-colors',
              statusFilter === 'running'
                ? 'bg-green-500 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            ]"
          >
            监控中 ({{ streamers.filter(s => s.is_running).length }})
          </button>
          <button
            @click="statusFilter = 'stopped'"
            :class="[
              'px-4 py-2 text-sm font-medium rounded-lg transition-colors',
              statusFilter === 'stopped'
                ? 'bg-gray-500 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            ]"
          >
            已停止 ({{ streamers.filter(s => !s.is_running).length }})
          </button>
        </div>
      </div>
    </div>

    <!-- 主播卡片网格 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
      <div
        v-for="s in paginated"
        :key="s.name"
        class="bg-white rounded-xl p-5 border border-gray-100 hover:shadow-md transition-all group"
      >
        <!-- 头部：名称 + 状态 -->
        <div class="flex items-center justify-between mb-4">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 bg-gradient-to-br from-blue-500 to-blue-600 rounded-full flex items-center justify-center">
              <span class="text-white font-semibold">{{ s.name.charAt(0) }}</span>
            </div>
            <div>
              <h3 class="font-semibold text-gray-900 truncate">{{ s.name }}</h3>
              <p class="text-xs text-gray-500">{{ s.files }} 个文件</p>
            </div>
          </div>
          <span
            :class="[
              'px-2 py-1 text-xs font-medium rounded-full',
              s.is_running
                ? 'bg-green-100 text-green-700'
                : 'bg-gray-100 text-gray-500'
            ]"
          >
            {{ s.is_running ? '监控中' : '已停止' }}
          </span>
        </div>

        <!-- 磁盘占用 -->
        <div class="mb-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-sm text-gray-500">磁盘占用</span>
            <span class="text-sm font-semibold text-gray-900">{{ s.size_gb.toFixed(1) }} GB</span>
          </div>
          <div class="w-full bg-gray-100 rounded-full h-2">
            <div
              :class="['h-2 rounded-full transition-all', getUsageColor(s.size_gb)]"
              :style="{ width: getUsagePercent(s.size_gb) + '%' }"
            ></div>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="flex gap-2">
          <button
            @click="emit('run', 'merge', s.name)"
            :disabled="running"
            class="flex-1 px-3 py-2 text-sm font-medium text-blue-600 bg-blue-50 rounded-lg hover:bg-blue-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            合并
          </button>
          <button
            @click="emit('run', 'clean', s.name)"
            :disabled="running"
            class="flex-1 px-3 py-2 text-sm font-medium text-red-600 bg-red-50 rounded-lg hover:bg-red-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            清理
          </button>
          <button
            @click="emit('openManualMerge', s.name)"
            :disabled="running"
            class="px-3 py-2 text-sm font-medium text-gray-600 bg-gray-100 rounded-lg hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            手动
          </button>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="totalPages > 1" class="mt-6 flex items-center justify-center gap-4">
      <button
        @click="currentPage = Math.max(1, currentPage - 1)"
        :disabled="currentPage <= 1"
        class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        上一页
      </button>
      <span class="text-sm text-gray-500">
        第 {{ currentPage }} / {{ totalPages }} 页
      </span>
      <button
        @click="currentPage = Math.min(totalPages, currentPage + 1)"
        :disabled="currentPage >= totalPages"
        class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        下一页
      </button>
    </div>
  </div>
</template>

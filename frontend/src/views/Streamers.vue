<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  streamers: { type: Array, default: () => [] },
  totalGB: { type: Number, default: 1 },
  running: { type: Boolean, default: false }
})

const emit = defineEmits(['run', 'openManualMerge'])

const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = 12
const statusFilter = ref('all')

const filtered = computed(() => {
  let list = props.streamers || []
  const kw = searchQuery.value.toLowerCase().trim()
  if (kw) list = list.filter(s => s.name.toLowerCase().includes(kw))
  if (statusFilter.value === 'running') list = list.filter(s => s.is_running)
  else if (statusFilter.value === 'stopped') list = list.filter(s => !s.is_running)
  return list
})

const totalPages = computed(() => Math.max(1, Math.ceil(filtered.value.length / pageSize)))
const paginated = computed(() => filtered.value.slice((currentPage.value - 1) * pageSize, currentPage.value * pageSize))

function getUsageColor(sizeGB) {
  const p = (sizeGB / props.totalGB) * 100
  if (p > 15) return 'bg-red-500'
  if (p > 10) return 'bg-amber-500'
  return 'bg-blue-500'
}
</script>

<template>
  <div class="p-8">
    <!-- 标题行 -->
    <div class="flex items-center justify-between mb-5">
      <h1 class="text-xl font-semibold text-gray-900">主播管理</h1>
      <div class="flex items-center gap-2">
        <button v-for="f in [{k:'all',l:'全部',c:''},{k:'running',l:'监控中',c:'text-green-600'},{k:'stopped',l:'已停止',c:'text-gray-500'}]" :key="f.k"
          @click="statusFilter=f.k; currentPage=1"
          :class="['px-3 py-1.5 text-sm rounded-lg transition-colors', statusFilter===f.k ? 'bg-blue-500 text-white' : 'text-gray-500 hover:bg-gray-100']">
          {{ f.l }} ({{ f.k==='all' ? streamers.length : f.k==='running' ? streamers.filter(s=>s.is_running).length : streamers.filter(s=>!s.is_running).length }})
        </button>
      </div>
    </div>

    <!-- 搜索 -->
    <div class="mb-5">
      <div class="relative max-w-md">
        <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/></svg>
        <input v-model="searchQuery" placeholder="搜索主播..." class="w-full pl-9 pr-4 py-2 bg-white border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent">
      </div>
    </div>

    <!-- 卡片网格 -->
    <div class="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
      <div v-for="s in paginated" :key="s.name" class="bg-white rounded-xl border border-gray-100 overflow-hidden hover:shadow-md transition-shadow">
        <div class="p-4">
          <div class="flex items-center gap-3 mb-3">
            <div class="w-10 h-10 rounded-full bg-blue-500 text-white flex items-center justify-center flex-shrink-0 text-sm font-semibold">{{ s.name.charAt(0) }}</div>
            <div class="min-w-0 flex-1">
              <div class="text-sm font-medium text-gray-900 truncate">{{ s.name }}</div>
              <div class="text-xs text-gray-400 mt-0.5">{{ s.files }} 个文件</div>
            </div>
            <span :class="['text-xs px-2 py-0.5 rounded-full', s.is_running ? 'bg-green-50 text-green-600' : 'bg-gray-100 text-gray-400']">
              {{ s.is_running ? '监控中' : '已停止' }}
            </span>
          </div>
          <div class="flex items-center gap-2 text-xs text-gray-400 mb-3">
            <span>磁盘占用</span>
            <div class="flex-1 bg-gray-100 rounded-full h-1">
              <div :class="['h-1 rounded-full', getUsageColor(s.size_gb)]" :style="{width: Math.min(100, s.size_gb/totalGB*100)+'%'}"></div>
            </div>
            <span class="font-medium text-gray-600">{{ s.size_gb.toFixed(1) }} GB</span>
          </div>
          <div class="flex gap-2">
            <button @click="emit('run','merge',s.name)" :disabled="running" class="flex-1 py-1.5 text-xs font-medium text-blue-600 bg-blue-50 rounded-lg hover:bg-blue-100 disabled:opacity-40 transition-colors">合并</button>
            <button @click="emit('run','clean',s.name)" :disabled="running" class="flex-1 py-1.5 text-xs font-medium text-red-600 bg-red-50 rounded-lg hover:bg-red-100 disabled:opacity-40 transition-colors">清理</button>
            <button @click="emit('openManualMerge',s.name)" :disabled="running" class="flex-1 py-1.5 text-xs font-medium text-gray-600 bg-gray-50 rounded-lg hover:bg-gray-100 disabled:opacity-40 transition-colors">手动</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="totalPages > 1" class="mt-5 flex items-center justify-center gap-3">
      <button @click="currentPage=Math.max(1,currentPage-1)" :disabled="currentPage<=1" class="px-3 py-1.5 text-sm text-gray-600 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-40">上一页</button>
      <span class="text-sm text-gray-400">{{ currentPage }} / {{ totalPages }}</span>
      <button @click="currentPage=Math.min(totalPages,currentPage+1)" :disabled="currentPage>=totalPages" class="px-3 py-1.5 text-sm text-gray-600 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-40">下一页</button>
    </div>
  </div>
</template>

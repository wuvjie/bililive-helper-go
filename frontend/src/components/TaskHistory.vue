<script setup>
import { ref, watch } from 'vue'
import { useApi } from '../composables/useApi'

const { get } = useApi()

const items = ref([])
const total = ref(0)
const pages = ref(0)
const currentPage = ref(1)
const taskFilter = ref('')
const loading = ref(false)

const perPage = 10

async function fetchHistory() {
  loading.value = true

  try {
    const params = new URLSearchParams({
      page: currentPage.value,
      per_page: perPage
    })

    if (taskFilter.value) {
      params.set('task', taskFilter.value)
    }

    const data = await get(`/api/history?${params.toString()}`)

    items.value = data.items || []
    total.value = data.total || 0
    pages.value = data.pages || 0
  } catch (err) {
    console.error('Failed to fetch history:', err)
  } finally {
    loading.value = false
  }
}

function formatHistoryTime(ts) {
  if (!ts) return '--'

  const date = new Date(ts.replace(' ', 'T'))
  if (isNaN(date.getTime())) return ts.substring(5, 16)

  const now = new Date()
  const h = date.getHours().toString().padStart(2, '0')
  const m = date.getMinutes().toString().padStart(2, '0')
  const t = `${h}:${m}`

  if (date.toDateString() === now.toDateString()) {
    return t
  }

  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)

  if (date.toDateString() === yesterday.toDateString()) {
    return `昨天 ${t}`
  }

  return `${(date.getMonth() + 1).toString().padStart(2, '0')}-${date.getDate().toString().padStart(2, '0')} ${t}`
}

function prevPage() {
  if (currentPage.value > 1) {
    currentPage.value--
    fetchHistory()
  }
}

function nextPage() {
  if (currentPage.value < pages.value) {
    currentPage.value++
    fetchHistory()
  }
}

watch(taskFilter, () => {
  currentPage.value = 1
  fetchHistory()
})

// Initial fetch
fetchHistory()
</script>

<template>
  <div class="h-full flex flex-col">
    <!-- 顶部工具栏 -->
    <div class="p-4 border-b border-gray-100">
      <div class="flex items-center justify-between">
        <select
          v-model="taskFilter"
          class="px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        >
          <option value="">全部任务</option>
          <option value="merge">仅合并</option>
          <option value="clean">仅清理</option>
        </select>
      </div>
    </div>

    <!-- 历史列表 -->
    <div class="flex-1 overflow-y-auto p-4 space-y-3">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
      </div>

      <div v-else-if="items.length === 0" class="flex flex-col items-center justify-center py-12 text-gray-400">
        <svg class="w-12 h-12 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
        <p class="text-sm">暂无操作记录</p>
      </div>

      <div
        v-for="r in items"
        :key="r.id"
        class="flex items-center p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
      >
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2 mb-1">
            <span class="text-sm font-medium text-gray-900">
              {{ r.task === 'merge' ? '合并' : '清理' }}任务
            </span>
          </div>
          <p class="text-sm text-gray-500 truncate">
            {{ r.detail || r.streamer }}
          </p>
          <p class="text-xs text-gray-400 mt-1">
            {{ formatHistoryTime(r.time) }}
          </p>
        </div>
        <div class="ml-4">
          <span
            :class="[
              'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
              r.status === 'success'
                ? 'bg-green-100 text-green-800'
                : 'bg-red-100 text-red-800'
            ]"
          >
            {{ r.status === 'success' ? '成功' : '失败' }}
          </span>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="pages > 1" class="px-4 py-3 border-t border-gray-100 flex items-center justify-center gap-4">
      <button
        @click="prevPage"
        :disabled="currentPage <= 1"
        class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        上一页
      </button>
      <span class="text-sm text-gray-500">
        {{ currentPage }} / {{ pages }}
      </span>
      <button
        @click="nextPage"
        :disabled="currentPage >= pages"
        class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        下一页
      </button>
    </div>
  </div>
</template>

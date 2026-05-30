<script setup>
import { ref, watch } from 'vue'
import { useApi } from '../composables/useApi'

const { get } = useApi()

const activeTab = ref('history')

// 任务历史
const historyItems = ref([])
const historyPage = ref(1)
const historyPages = ref(0)
const taskFilter = ref('')
const historyLoading = ref(false)

// 系统日志
const logType = ref('merge')
const logFiles = ref([])
const selectedLogFile = ref('')
const logContent = ref('')
const logContentHtml = ref('')
const autoScroll = ref(true)

async function fetchHistory() {
  historyLoading.value = true
  try {
    const params = new URLSearchParams({ page: historyPage.value, per_page: 10 })
    if (taskFilter.value) params.set('task', taskFilter.value)
    const data = await get(`/api/history?${params.toString()}`)
    historyItems.value = data.items || []
    historyPages.value = data.pages || 0
  } catch (err) {
    console.error('Failed to fetch history:', err)
  } finally {
    historyLoading.value = false
  }
}

async function loadLogFiles() {
  try {
    const data = await get(`/api/logs/list/${logType.value}`)
    logFiles.value = data || []
    if (data && data.length > 0) {
      selectedLogFile.value = data[0].filename
      await showLog()
    }
  } catch (err) {
    console.error('Failed to load log files:', err)
  }
}

async function showLog() {
  logContentHtml.value = "<span class='text-gray-400'>加载中...</span>"
  try {
    const response = await fetch(`/api/logs/content/${logType.value}?file=${selectedLogFile.value}`)
    const text = await response.text()
    logContent.value = text || '(空)'
    logContentHtml.value = highlightLog(logContent.value)
    if (autoScroll.value) {
      setTimeout(() => {
        const el = document.getElementById('logContent')
        if (el) el.scrollTop = el.scrollHeight
      }, 50)
    }
  } catch (err) {
    logContentHtml.value = "<span class='text-red-400'>加载失败</span>"
  }
}

function escapeHtml(text) {
  return (text || '').toString().replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

function highlightLog(text) {
  return escapeHtml(text)
    .replace(/ERROR/gi, '<span class="text-red-400 font-bold">ERROR</span>')
    .replace(/WARN/gi, '<span class="text-yellow-400 font-bold">WARN</span>')
    .replace(/✅|SUCCESS/gi, '<span class="text-green-400">$&</span>')
    .replace(/❌/g, '<span class="text-red-400">❌</span>')
}

function formatHistoryTime(ts) {
  if (!ts) return '--'
  const date = new Date(ts.replace(' ', 'T'))
  if (isNaN(date.getTime())) return ts.substring(5, 16)
  const now = new Date()
  const h = date.getHours().toString().padStart(2, '0')
  const m = date.getMinutes().toString().padStart(2, '0')
  if (date.toDateString() === now.toDateString()) return `${h}:${m}`
  return `${(date.getMonth() + 1).toString().padStart(2, '0')}-${date.getDate().toString().padStart(2, '0')} ${h}:${m}`
}

watch(activeTab, (tab) => {
  if (tab === 'history') fetchHistory()
  else loadLogFiles()
})

watch(logType, () => loadLogFiles())

// 初始加载
fetchHistory()
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
        <span class="text-gray-700 font-medium">任务中心</span>
      </nav>
      <h1 class="mt-2 text-2xl font-bold text-gray-900">任务中心</h1>
    </div>

    <!-- Tab 切换 -->
    <div class="bg-white rounded-xl border border-gray-100 overflow-hidden">
      <div class="flex border-b border-gray-200">
        <button
          @click="activeTab = 'history'"
          :class="[
            'px-6 py-4 text-sm font-medium transition-colors',
            activeTab === 'history'
              ? 'text-blue-600 border-b-2 border-blue-600 bg-blue-50'
              : 'text-gray-500 hover:text-gray-700 hover:bg-gray-50'
          ]"
        >
          任务历史
        </button>
        <button
          @click="activeTab = 'logs'"
          :class="[
            'px-6 py-4 text-sm font-medium transition-colors',
            activeTab === 'logs'
              ? 'text-blue-600 border-b-2 border-blue-600 bg-blue-50'
              : 'text-gray-500 hover:text-gray-700 hover:bg-gray-50'
          ]"
        >
          系统日志
        </button>
      </div>

      <!-- 任务历史 -->
      <div v-show="activeTab === 'history'" class="p-6">
        <!-- 筛选 -->
        <div class="flex items-center gap-4 mb-6">
          <select
            v-model="taskFilter"
            @change="historyPage = 1; fetchHistory()"
            class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">全部任务</option>
            <option value="merge">仅合并</option>
            <option value="clean">仅清理</option>
          </select>
          <button
            @click="fetchHistory"
            class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
          >
            刷新
          </button>
        </div>

        <!-- 列表 -->
        <div class="space-y-3">
          <div v-if="historyLoading" class="flex justify-center py-8">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
          </div>

          <div v-else-if="historyItems.length === 0" class="text-center py-8 text-gray-400">
            暂无操作记录
          </div>

          <div
            v-for="r in historyItems"
            :key="r.id"
            class="flex items-center p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
          >
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-1">
                <span class="font-medium text-gray-900">{{ r.task === 'merge' ? '合并' : '清理' }}任务</span>
              </div>
              <p class="text-sm text-gray-500">{{ r.detail || r.streamer }}</p>
              <p class="text-xs text-gray-400 mt-1">{{ formatHistoryTime(r.time) }}</p>
            </div>
            <span
              :class="[
                'px-3 py-1 text-sm font-medium rounded-full',
                r.status === 'success' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
              ]"
            >
              {{ r.status === 'success' ? '成功' : '失败' }}
            </span>
          </div>
        </div>

        <!-- 分页 -->
        <div v-if="historyPages > 1" class="mt-6 flex items-center justify-center gap-4">
          <button
            @click="historyPage--; fetchHistory()"
            :disabled="historyPage <= 1"
            class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50"
          >
            上一页
          </button>
          <span class="text-sm text-gray-500">{{ historyPage }} / {{ historyPages }}</span>
          <button
            @click="historyPage++; fetchHistory()"
            :disabled="historyPage >= historyPages"
            class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50"
          >
            下一页
          </button>
        </div>
      </div>

      <!-- 系统日志 -->
      <div v-show="activeTab === 'logs'" class="bg-gray-900">
        <!-- 工具栏 -->
        <div class="flex items-center gap-4 p-4 border-b border-gray-700">
          <select
            v-model="logType"
            class="px-3 py-2 bg-gray-800 text-white border border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="merge">合并</option>
            <option value="clean">清理</option>
          </select>

          <select
            v-model="selectedLogFile"
            @change="showLog"
            class="flex-1 px-3 py-2 bg-gray-800 text-white border border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option v-for="f in logFiles" :key="f.filename" :value="f.filename">{{ f.date }}</option>
          </select>

          <label class="flex items-center gap-2 text-sm text-gray-300 cursor-pointer">
            <input type="checkbox" v-model="autoScroll" class="rounded bg-gray-800 border-gray-700">
            自动滚动
          </label>

          <button
            @click="showLog"
            class="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm hover:bg-gray-600 transition-colors"
          >
            刷新
          </button>
        </div>

        <!-- 日志内容 -->
        <div id="logContent" class="p-4 h-[500px] overflow-auto font-mono text-sm text-gray-300 leading-relaxed">
          <div v-html="logContentHtml || logContent || '选择日志文件'"></div>
        </div>
      </div>
    </div>
  </div>
</template>

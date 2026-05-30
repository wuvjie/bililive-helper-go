<script setup>
import { ref, watch } from 'vue'
import { useApi } from '../composables/useApi'

const { get } = useApi()
const activeTab = ref('history')
const historyItems = ref([])
const historyPage = ref(1)
const historyPages = ref(0)
const taskFilter = ref('')
const loading = ref(false)
const logType = ref('merge')
const logFiles = ref([])
const selectedLogFile = ref('')
const logContent = ref('')
const logContentHtml = ref('')
const autoScroll = ref(true)

async function fetchHistory() {
  loading.value = true
  try {
    const params = new URLSearchParams({ page: historyPage.value, per_page: 10 })
    if (taskFilter.value) params.set('task', taskFilter.value)
    const data = await get('/api/history?' + params.toString())
    historyItems.value = data.items || []
    historyPages.value = data.pages || 0
  } catch (e) { console.error(e) } finally { loading.value = false }
}

async function loadLogFiles() {
  try {
    const data = await get('/api/logs/list/' + logType.value)
    logFiles.value = data || []
    if (data?.length) { selectedLogFile.value = data[0].filename; await showLog() }
  } catch (e) { console.error(e) }
}

async function showLog() {
  logContentHtml.value = '<span style="color:#9ca3af">加载中...</span>'
  try {
    const r = await fetch('/api/logs/content/' + logType.value + '?file=' + selectedLogFile.value)
    const t = await r.text()
    logContent.value = t || '(空)'
    logContentHtml.value = highlight(logContent.value)
    if (autoScroll.value) setTimeout(() => { const el = document.getElementById('logBox'); if (el) el.scrollTop = el.scrollHeight }, 50)
  } catch { logContentHtml.value = '<span style="color:#ef4444">加载失败</span>' }
}

function highlight(t) {
  return (t||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')
    .replace(/ERROR/gi,'<span style="color:#f87171;font-weight:600">ERROR</span>')
    .replace(/WARN/gi,'<span style="color:#fbbf24;font-weight:600">WARN</span>')
    .replace(/✅|SUCCESS/gi,'<span style="color:#34d399">$&</span>')
    .replace(/❌/g,'<span style="color:#f87171">❌</span>')
}

function fmtTime(ts) {
  if (!ts) return '--'
  const d = new Date(ts.replace(' ','T'))
  if (isNaN(d)) return ts.substring(5,16)
  const now = new Date()
  const t = d.getHours().toString().padStart(2,'0')+':'+d.getMinutes().toString().padStart(2,'0')
  if (d.toDateString()===now.toDateString()) return t
  return (d.getMonth()+1).toString().padStart(2,'0')+'-'+d.getDate().toString().padStart(2,'0')+' '+t
}

watch(activeTab, t => { if (t==='history') fetchHistory(); else loadLogFiles() })
watch(logType, () => loadLogFiles())
fetchHistory()
</script>

<template>
  <div class="p-8">
    <h1 class="text-xl font-semibold text-gray-900 mb-5">任务中心</h1>

    <div class="bg-white rounded-xl border border-gray-100 overflow-hidden">
      <div class="flex border-b border-gray-100">
        <button @click="activeTab='history'" :class="['px-5 py-3 text-sm font-medium transition-colors', activeTab==='history'?'text-blue-600 border-b-2 border-blue-600':'text-gray-500 hover:text-gray-700']">任务历史</button>
        <button @click="activeTab='logs'" :class="['px-5 py-3 text-sm font-medium transition-colors', activeTab==='logs'?'text-blue-600 border-b-2 border-blue-600':'text-gray-500 hover:text-gray-700']">系统日志</button>
      </div>

      <div v-show="activeTab==='history'" class="p-5">
        <div class="flex items-center gap-3 mb-4">
          <select v-model="taskFilter" @change="historyPage=1;fetchHistory()" class="px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
            <option value="">全部任务</option><option value="merge">仅合并</option><option value="clean">仅清理</option>
          </select>
          <button @click="fetchHistory" class="px-3 py-2 text-sm text-gray-600 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">刷新</button>
        </div>
        <div class="divide-y divide-gray-100">
          <div v-if="loading" class="py-8 text-center text-sm text-gray-400">加载中...</div>
          <div v-else-if="!historyItems.length" class="py-8 text-center text-sm text-gray-400">暂无记录</div>
          <div v-for="r in historyItems" :key="r.id" class="flex items-center justify-between py-3">
            <div>
              <div class="text-sm font-medium text-gray-800">{{ r.task==='merge'?'合并':'清理' }}任务</div>
              <div class="text-xs text-gray-400 mt-0.5">{{ r.detail || r.streamer }} · {{ fmtTime(r.time) }}</div>
            </div>
            <span :class="['text-xs font-medium px-2 py-0.5 rounded-full', r.status==='success'?'bg-green-50 text-green-600':'bg-red-50 text-red-600']">{{ r.status==='success'?'成功':'失败' }}</span>
          </div>
        </div>
        <div v-if="historyPages>1" class="mt-4 flex items-center justify-center gap-3">
          <button @click="historyPage--;fetchHistory()" :disabled="historyPage<=1" class="px-3 py-1.5 text-sm text-gray-600 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-40">上一页</button>
          <span class="text-sm text-gray-400">{{ historyPage }} / {{ historyPages }}</span>
          <button @click="historyPage++;fetchHistory()" :disabled="historyPage>=historyPages" class="px-3 py-1.5 text-sm text-gray-600 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-40">下一页</button>
        </div>
      </div>

      <div v-show="activeTab==='logs'">
        <div class="flex items-center gap-3 p-4 border-b border-gray-100">
          <select v-model="logType" class="px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"><option value="merge">合并</option><option value="clean">清理</option></select>
          <select v-model="selectedLogFile" @change="showLog" class="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
            <option v-for="f in logFiles" :key="f.filename" :value="f.filename">{{ f.date }}</option>
          </select>
          <label class="flex items-center gap-1.5 text-sm text-gray-500 cursor-pointer whitespace-nowrap">
            <input type="checkbox" v-model="autoScroll" class="rounded border-gray-300"> 自动滚动
          </label>
          <button @click="showLog" class="px-3 py-2 text-sm text-gray-600 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors whitespace-nowrap">刷新</button>
        </div>
        <div id="logBox" class="bg-gray-900 p-4 h-[500px] overflow-auto font-mono text-xs leading-relaxed">
          <div class="text-gray-300 whitespace-pre-wrap" v-html="logContentHtml||logContent||'选择日志文件'"></div>
        </div>
      </div>
    </div>
  </div>
</template>

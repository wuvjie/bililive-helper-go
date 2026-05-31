<script setup>
import { ref, watch, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const api = useApi()
const historyList = ref([])
const loading = ref(false)
const page = ref(1)
const perPage = ref(20)
const total = ref(0)
const pages = ref(0)
const filterTask = ref('')

const showTaskLog = ref(false)
const taskLogs = ref([])
const taskLogContent = ref('')
const selectedTask = ref('')
const loadingTaskLog = ref(false)

async function fetchLogs() {
  loading.value = true
  try {
    let url = `/history?page=${page.value}&per_page=${perPage.value}`
    if (filterTask.value) url += `&task=${filterTask.value}`
    const res = await api.get(url)
    historyList.value = res.items || []
    total.value = res.total || 0
    pages.value = res.pages || 0
  } catch (e) {
    console.error('获取日志失败', e)
  } finally {
    loading.value = false
  }
}

function goToPage(p) {
  if (p < 1 || p > pages.value) return
  page.value = p
  fetchLogs()
}

async function doExport() {
  try {
    const res = await api.get('/history/export')
    const data = Array.isArray(res) ? res : []
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `bililive-history-${new Date().toISOString().split('T')[0]}.json`
    a.click()
    URL.revokeObjectURL(url)
  } catch (e) {
    alert('导出失败: ' + e.message)
  }
}

async function loadTaskLogs(task) {
  selectedTask.value = task
  taskLogContent.value = ''
  loadingTaskLog.value = true
  showTaskLog.value = true
  try {
    const res = await api.get(`/logs/list/${task}`)
    taskLogs.value = Array.isArray(res) ? res : []
    if (taskLogs.value.length > 0) {
      await loadLogContent(task, taskLogs.value[0].filename)
    }
  } catch (e) {
    taskLogs.value = []
  } finally {
    loadingTaskLog.value = false
  }
}

async function loadLogContent(task, filename) {
  try {
    const res = await fetch(`/api/logs/content/${task}?file=${encodeURIComponent(filename)}`, {
      headers: { 'Authorization': `Bearer ${localStorage.getItem('token') || ''}` }
    })
    taskLogContent.value = await res.text()
  } catch (e) {
    taskLogContent.value = '加载失败: ' + e.message
  }
}

function getTaskClass(task) {
  if (!task) return 'bg-gray-100 text-gray-600'
  const t = task.toLowerCase()
  if (t.includes('merge') || t.includes('合并')) return 'bg-[#e8f3ff] text-[#3370ff]'
  if (t.includes('clean') || t.includes('清理')) return 'bg-[#fff0f0] text-[#f54a45]'
  if (t.includes('config')) return 'bg-[#f0f1f5] text-[#646a73]'
  return 'bg-[#f8f9fa] text-[#646a73]'
}

function formatSize(gb) {
  if (!gb && gb !== 0) return '-'
  if (gb < 1) return (gb * 1024).toFixed(1) + ' MB'
  return gb.toFixed(2) + ' GB'
}

onMounted(fetchLogs)
</script>

<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-5">
      <h1 class="text-[17px] font-semibold text-[#1f2329]">操作日志</h1>
      <div class="flex gap-2">
        <select v-model="filterTask" @change="page = 1; fetchLogs()" class="px-3 py-2 border border-[#dee0e3] rounded-lg text-[13px] bg-white outline-none">
          <option value="">全部类型</option>
          <option value="merge">合并</option>
          <option value="clean">清理</option>
          <option value="config">配置</option>
        </select>
        <button @click="fetchLogs" class="px-3 py-2 border border-[#dee0e3] hover:bg-gray-50 text-[13px] rounded-lg bg-white">
          {{ loading ? '...' : '刷新' }}
        </button>
        <button @click="doExport" class="px-3 py-2 border border-[#dee0e3] hover:bg-gray-50 text-[13px] rounded-lg bg-white">
          导出
        </button>
      </div>
    </div>

    <div class="bg-white border border-[#dee0e3] rounded-xl overflow-hidden shadow-sm hover:shadow-md transition-shadow">
      <table class="w-full text-left text-[14px]">
        <thead class="bg-[#f8f9fa] text-[#646a73] border-b border-[#dee0e3]">
          <tr>
            <th class="px-5 py-4 font-medium w-[15%]">时间</th>
            <th class="px-5 py-4 font-medium w-[10%]">类型</th>
            <th class="px-5 py-4 font-medium w-[15%]">主播</th>
            <th class="px-5 py-4 font-medium w-[38%]">详情</th>
            <th class="px-5 py-4 font-medium w-[12%]">状态</th>
            <th class="px-5 py-4 font-medium w-[10%]">日志</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#dee0e3]">
          <tr v-for="(item, i) in historyList" :key="item.id || i" class="hover:bg-[#fcfcfc] transition-colors">
            <td class="px-5 py-3.5 text-[#646a73] text-[13px]">{{ item.time || '-' }}</td>
            <td class="px-5 py-3.5">
              <span :class="['px-2 py-1 rounded text-xs font-medium', getTaskClass(item.task)]">
                {{ item.task || '未知' }}
              </span>
            </td>
            <td class="px-5 py-3.5 text-[#1f2329] text-[13px]">{{ item.streamer || '-' }}</td>
            <td class="px-5 py-3.5 text-[#1f2329] text-[13px]">
              <div v-if="item.detail">{{ item.detail }}</div>
              <div v-else class="flex gap-3 text-[#646a73]">
                <span v-if="item.files_count">文件: {{ item.files_count }}</span>
                <span v-if="item.merged_bytes">合并: {{ formatSize(item.merged_bytes / (1024*1024*1024)) }}</span>
                <span v-if="item.freed_bytes">释放: {{ formatSize(item.freed_bytes / (1024*1024*1024)) }}</span>
                <span v-if="item.duration">耗时: {{ item.duration }}</span>
              </div>
            </td>
            <td class="px-5 py-3.5">
              <span v-if="item.status === 'success'" class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-[#e8f8f0] text-[#00b578] text-xs font-medium rounded-md">
                <span class="w-1.5 h-1.5 rounded-full bg-[#00b578]"></span>成功
              </span>
              <span v-else-if="item.status === 'error'" class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-[#fff0f0] text-[#f54a45] text-xs font-medium rounded-md">
                <span class="w-1.5 h-1.5 rounded-full bg-[#f54a45]"></span>失败
              </span>
              <span v-else class="text-[#646a73] text-[13px]">{{ item.status || '-' }}</span>
            </td>
            <td class="px-5 py-3.5">
              <button v-if="item.task === 'merge' || item.task === 'clean'"
                @click="loadTaskLogs(item.task)"
                class="text-[#3370ff] hover:underline text-[12px]">
                查看日志
              </button>
            </td>
          </tr>
          <tr v-if="historyList.length === 0 && !loading">
            <td colspan="6" class="px-6 py-12 text-center text-[#8f959e] text-sm">暂无操作记录</td>
          </tr>
        </tbody>
      </table>

      <div v-if="pages > 1" class="flex items-center justify-between px-5 py-3 border-t border-[#dee0e3] bg-[#f8f9fa] text-[13px]">
        <span class="text-[#8f959e]">共 {{ total }} 条记录</span>
        <div class="flex items-center gap-1">
          <button @click="goToPage(page - 1)" :disabled="page <= 1"
            class="px-3 py-1.5 rounded border border-[#dee0e3] bg-white disabled:opacity-40 hover:bg-gray-50">上一页</button>
          <span class="px-3 py-1.5 text-[#1f2329] font-medium">{{ page }} / {{ pages }}</span>
          <button @click="goToPage(page + 1)" :disabled="page >= pages"
            class="px-3 py-1.5 rounded border border-[#dee0e3] bg-white disabled:opacity-40 hover:bg-gray-50">下一页</button>
        </div>
      </div>
    </div>

    <div v-if="showTaskLog" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50" @click.self="showTaskLog = false">
      <div class="bg-white rounded-xl w-[800px] max-h-[80vh] flex flex-col shadow-xl">
        <div class="flex items-center justify-between px-6 py-4 border-b border-[#dee0e3]">
          <h3 class="text-[15px] font-semibold text-[#1f2329]">任务日志: {{ selectedTask }}</h3>
          <button @click="showTaskLog = false" class="text-[#8f959e] hover:text-[#1f2329] text-xl">&times;</button>
        </div>
        <div v-if="taskLogs.length > 0" class="flex gap-2 px-6 py-3 border-b border-[#f0f1f5] overflow-x-auto">
          <button v-for="tl in taskLogs" :key="tl.filename"
            @click="loadLogContent(selectedTask, tl.filename)"
            class="px-3 py-1.5 rounded text-[12px] border border-[#dee0e3] bg-white hover:bg-[#f8f9fa] whitespace-nowrap">
            {{ tl.date }}
          </button>
        </div>
        <div v-if="loadingTaskLog" class="p-6 text-center text-[#8f959e]">加载中...</div>
        <pre v-else class="flex-1 overflow-auto p-6 text-[13px] font-mono text-[#1f2329] bg-[#fafbfc] leading-relaxed whitespace-pre-wrap">{{ taskLogContent || '暂无日志内容' }}</pre>
      </div>
    </div>
  </div>
</template>

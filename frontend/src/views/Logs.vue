<script setup>
import { ref, watch } from 'vue'
import { useApi } from '../composables/useApi'

const { get } = useApi()
const logType = ref('merge')
const logFiles = ref([])
const selectedFile = ref('')
const content = ref('')
const contentHtml = ref('')
const autoScroll = ref(true)

async function loadLogFiles() {
  try {
    const data = await get('/api/logs/list/' + logType.value)
    logFiles.value = data || []
    if (data?.length) { selectedFile.value = data[0].filename; await showLog() }
  } catch (e) { console.error(e) }
}

async function showLog() {
  contentHtml.value = '<span style="color:#9ca3af">加载中...</span>'
  try {
    const r = await fetch('/api/logs/content/' + logType.value + '?file=' + selectedFile.value)
    const t = await r.text()
    content.value = t || '(空)'
    contentHtml.value = highlight(content.value)
    if (autoScroll.value) setTimeout(() => { const el = document.getElementById('logView'); if (el) el.scrollTop = el.scrollHeight }, 50)
  } catch { contentHtml.value = '<span style="color:#ef4444">加载失败</span>' }
}

function highlight(t) {
  return (t||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')
    .replace(/ERROR/gi,'<span style="color:#f87171;font-weight:600">ERROR</span>')
    .replace(/WARN/gi,'<span style="color:#fbbf24;font-weight:600">WARN</span>')
    .replace(/✅|SUCCESS/gi,'<span style="color:#34d399">$&</span>')
    .replace(/❌/g,'<span style="color:#f87171">❌</span>')
}

watch(logType, () => loadLogFiles())
loadLogFiles()
</script>

<template>
  <div class="p-8">
    <h1 class="text-xl font-semibold text-gray-900 mb-5">操作日志</h1>

    <div class="bg-white rounded-xl border border-gray-100 overflow-hidden">
      <div class="flex items-center gap-3 p-4 border-b border-gray-100">
        <select v-model="logType" class="px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"><option value="merge">合并日志</option><option value="clean">清理日志</option></select>
        <select v-model="selectedFile" @change="showLog" class="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
          <option v-for="f in logFiles" :key="f.filename" :value="f.filename">{{ f.date }}</option>
        </select>
        <label class="flex items-center gap-1.5 text-sm text-gray-500 cursor-pointer whitespace-nowrap">
          <input type="checkbox" v-model="autoScroll" class="rounded border-gray-300"> 自动滚动
        </label>
        <button @click="showLog" class="px-3 py-2 text-sm text-gray-600 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors whitespace-nowrap">刷新</button>
      </div>
      <div id="logView" class="bg-gray-900 p-4 h-[600px] overflow-auto font-mono text-xs leading-relaxed">
        <div class="text-gray-300 whitespace-pre-wrap" v-html="contentHtml||content||'选择日志文件'"></div>
      </div>
    </div>
  </div>
</template>

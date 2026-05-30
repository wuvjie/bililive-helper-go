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
    const data = await get(`/api/logs/list/${logType.value}`)
    logFiles.value = data || []

    if (data && data.length > 0) {
      selectedFile.value = data[0].filename
      await showLog()
    } else {
      content.value = '无日志数据'
      contentHtml.value = highlightLog('无日志数据')
    }
  } catch (err) {
    console.error('Failed to load log files:', err)
  }
}

async function showLog() {
  contentHtml.value = "<span class='text-gray-400'>请求中...</span>"

  try {
    const response = await fetch(`/api/logs/content/${logType.value}?file=${selectedFile.value}`)
    const text = await response.text()
    content.value = text || '(空)'
    contentHtml.value = highlightLog(content.value)

    if (autoScroll.value) {
      setTimeout(() => {
        const histLog = document.getElementById('histLog')
        if (histLog) histLog.scrollTop = histLog.scrollHeight
      }, 50)
    }
  } catch (err) {
    contentHtml.value = "<span class='text-red-400'>读取失败</span>"
  }
}

function escapeHtml(text) {
  return (text || '').toString()
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
}

function highlightLog(text) {
  return escapeHtml(text)
    .replace(/ERROR/gi, '<span class="text-red-400 font-bold">ERROR</span>')
    .replace(/WARN/gi, '<span class="text-yellow-400 font-bold">WARN</span>')
    .replace(/✅|SUCCESS/gi, '<span class="text-green-400">$&</span>')
    .replace(/ℹ|信息/g, '<span class="text-blue-400">$&</span>')
    .replace(/❌/g, '<span class="text-red-400">❌</span>')
}

watch(logType, () => {
  loadLogFiles()
})

// Initial load
loadLogFiles()
</script>

<template>
  <div class="h-full flex flex-col bg-gray-950">
    <!-- 顶部工具栏 -->
    <div class="flex items-center gap-3 p-4 bg-gray-900 border-b border-gray-800">
      <select
        v-model="logType"
        class="px-3 py-2 bg-gray-800 text-white border border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
      >
        <option value="merge">合并</option>
        <option value="clean">清理</option>
      </select>

      <select
        v-model="selectedFile"
        @change="showLog"
        class="flex-1 px-3 py-2 bg-gray-800 text-white border border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
      >
        <option v-for="f in logFiles" :key="f.filename" :value="f.filename">
          {{ f.date }}
        </option>
      </select>

      <label class="flex items-center gap-2 text-sm text-gray-300 cursor-pointer">
        <input
          type="checkbox"
          v-model="autoScroll"
          class="w-4 h-4 rounded bg-gray-800 border-gray-700 text-blue-500 focus:ring-blue-500 focus:ring-offset-gray-900"
        >
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
    <div
      id="histLog"
      class="flex-1 overflow-auto p-4 font-mono text-sm leading-relaxed"
    >
      <div class="text-gray-300" v-html="contentHtml || content || '选择日志文件'"></div>
    </div>
  </div>
</template>

<style scoped>
/* 终端滚动条样式 */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: #1f2937;
}

::-webkit-scrollbar-thumb {
  background: #4b5563;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #6b7280;
}
</style>

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
    }
  } catch (err) {
    console.error('Failed to load log files:', err)
  }
}

async function showLog() {
  contentHtml.value = "<span class='text-gray-400'>加载中...</span>"
  try {
    const response = await fetch(`/api/logs/content/${logType.value}?file=${selectedFile.value}`)
    const text = await response.text()
    content.value = text || '(空)'
    contentHtml.value = highlightLog(content.value)
    if (autoScroll.value) {
      setTimeout(() => {
        const el = document.getElementById('logContent')
        if (el) el.scrollTop = el.scrollHeight
      }, 50)
    }
  } catch (err) {
    contentHtml.value = "<span class='text-red-400'>加载失败</span>"
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

watch(logType, () => loadLogFiles())
loadLogFiles()
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
        <span class="text-gray-700 font-medium">操作日志</span>
      </nav>
      <h1 class="mt-2 text-2xl font-bold text-gray-900">操作日志</h1>
    </div>

    <!-- 日志查看器 -->
    <div class="bg-white rounded-xl border border-gray-100 overflow-hidden">
      <!-- 工具栏 -->
      <div class="flex items-center gap-4 p-4 border-b border-gray-200">
        <select
          v-model="logType"
          class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="merge">合并日志</option>
          <option value="clean">清理日志</option>
        </select>

        <select
          v-model="selectedFile"
          @change="showLog"
          class="flex-1 px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option v-for="f in logFiles" :key="f.filename" :value="f.filename">{{ f.date }}</option>
        </select>

        <label class="flex items-center gap-2 text-sm text-gray-600 cursor-pointer">
          <input type="checkbox" v-model="autoScroll" class="rounded border-gray-300">
          自动滚动
        </label>

        <button
          @click="showLog"
          class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
        >
          刷新
        </button>
      </div>

      <!-- 日志内容 -->
      <div id="logContent" class="bg-gray-900 p-4 h-[600px] overflow-auto font-mono text-sm text-gray-300 leading-relaxed">
        <div v-html="contentHtml || content || '选择日志文件'"></div>
      </div>
    </div>
  </div>
</template>

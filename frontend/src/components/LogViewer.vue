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
  contentHtml.value = "<span style='color:var(--muted)'>请求中...</span>"

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
    contentHtml.value = "<span style='color:var(--err)'>读取失败</span>"
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
    .replace(/ERROR/gi, '<span class="te">ERROR</span>')
    .replace(/WARN/gi, '<span class="tw">WARN</span>')
    .replace(/✅|SUCCESS/gi, '<span class="tok">$&</span>')
    .replace(/ℹ|信息/g, '<span class="ti">$&</span>')
    .replace(/❌/g, '<span class="te">❌</span>')
}

watch(logType, () => {
  loadLogFiles()
})

// Initial load
loadLogFiles()
</script>

<template>
  <div class="term-wrapper">
    <div class="term" id="histLog" v-html="contentHtml || content || '选择日志文件'"></div>
  </div>
</template>

<style scoped>
.term-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 16px;
  min-width: 0;
}

.term {
  font-family: var(--font-mono);
  background: transparent;
  color: inherit;
  padding: 0;
  border: none;
  overflow-y: auto;
  overflow-x: hidden;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  word-break: break-all;
  font-size: 13px;
  line-height: 1.6;
  margin: 0;
  box-shadow: none;
  flex: 1;
  min-width: 0;
  width: 100%;
}

:deep(.te) { color: #ff7b72; }
:deep(.tw) { color: #d2a8ff; }
:deep(.tok) { color: #7ee787; }
:deep(.ti) { color: #79c0ff; }
</style>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useStreamerData } from '../composables/useStreamerData'
import { useConfig } from '../composables/useConfig'
import { useApi } from '../composables/useApi'
import { useSSE } from '../composables/useSSE'

const { streamers } = useStreamerData()
const { schedule, fetchSchedule } = useConfig()
const api = useApi()
const sse = useSSE()

const selectedStreamer = ref('')
const streamerFiles = ref([])
const selectedFiles = ref([])
const loadingFiles = ref(false)

const nonMergedFiles = computed(() => streamerFiles.value.filter(f => !f.is_merged))

async function loadFiles() {
  if (!selectedStreamer.value) {
    streamerFiles.value = []
    selectedFiles.value = []
    return
  }
  loadingFiles.value = true
  try {
    const res = await api.get(`/streamers/${encodeURIComponent(selectedStreamer.value)}/files`)
    streamerFiles.value = Array.isArray(res) ? res : []
    selectedFiles.value = []
  } catch (e) {
    console.error('获取文件列表失败', e)
    streamerFiles.value = []
  } finally {
    loadingFiles.value = false
  }
}

function toggleFile(name) {
  const idx = selectedFiles.value.indexOf(name)
  if (idx >= 0) {
    selectedFiles.value.splice(idx, 1)
  } else {
    selectedFiles.value.push(name)
  }
}

function selectAllNonMerged() {
  selectedFiles.value = nonMergedFiles.value.map(f => f.name)
}

function formatBytes(bytes) {
  if (!bytes) return '0 B'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
}

function fmtTime(ts) {
  if (!ts || ts <= 0) return '未排期'
  return new Date(ts * 1000).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

async function startMerge() {
  if (!selectedStreamer.value || selectedFiles.value.length < 2) return
  sse.clear()
  sse.addLine(`开始手动合并: ${selectedStreamer.value} (${selectedFiles.value.length} 个文件)`)
  const ok = await sse.startSSE('/api/merge/manual', { streamer: selectedStreamer.value, files: selectedFiles.value })
  if (ok) {
    sse.addLine('任务完成', 'success')
    loadFiles()
    fetchSchedule()
  }
}

async function startClean() {
  sse.clear()
  sse.addLine('开始清理任务...')
  const ok = await sse.startSSE('/api/clean', { streamer: selectedStreamer.value || '' })
  if (ok) {
    sse.addLine('清理任务完成', 'success')
    fetchSchedule()
  }
}
</script>

<template>
  <div class="p-6">
    <h1 class="text-[17px] font-semibold text-[#1f2329] mb-5">任务中心</h1>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-5">

      <div class="space-y-5">
        <div class="bg-white rounded-xl p-5 border border-[#dee0e3] shadow-sm">
          <h3 class="text-[15px] font-semibold text-[#1f2329] mb-3">定时任务状态</h3>
          <div class="space-y-3">
            <div class="flex items-center justify-between p-3 bg-[#f8f9fa] rounded-lg">
              <div>
                <div class="text-[14px] font-medium text-[#1f2329]">自动合并</div>
                <div class="text-[12px] text-[#8f959e]">间隔: {{ schedule?.merge?.interval || '-' }} 分钟 · 下次: {{ schedule?.merge?.enabled ? fmtTime(schedule.merge.next_run) : '已暂停' }}</div>
              </div>
              <span :class="['px-2 py-1 rounded text-xs font-medium', schedule?.merge?.enabled ? 'bg-[#e8f8f0] text-[#00b578]' : 'bg-gray-100 text-gray-500']">
                {{ schedule?.merge?.enabled ? '已启用' : '已暂停' }}
              </span>
            </div>
            <div class="flex items-center justify-between p-3 bg-[#f8f9fa] rounded-lg">
              <div>
                <div class="text-[14px] font-medium text-[#1f2329]">自动清理</div>
                <div class="text-[12px] text-[#8f959e]">间隔: {{ schedule?.clean?.interval || '-' }} 分钟 · 下次: {{ schedule?.clean?.enabled ? fmtTime(schedule.clean.next_run) : '已暂停' }}</div>
              </div>
              <span :class="['px-2 py-1 rounded text-xs font-medium', schedule?.clean?.enabled ? 'bg-[#e8f8f0] text-[#00b578]' : 'bg-gray-100 text-gray-500']">
                {{ schedule?.clean?.enabled ? '已启用' : '已暂停' }}
              </span>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-xl p-5 border border-[#dee0e3] shadow-sm">
          <h3 class="text-[15px] font-semibold text-[#1f2329] mb-1">手动操作</h3>
          <p class="text-[13px] text-[#8f959e] mb-4">选择主播和文件后手动触发合并，至少选择 2 个文件。</p>

          <div class="space-y-4">
            <div>
              <label class="block text-[13px] font-medium text-[#1f2329] mb-1.5">选择主播</label>
              <select v-model="selectedStreamer" @change="loadFiles" class="w-full px-3 py-2 border border-[#dee0e3] rounded-lg text-[14px] outline-none focus:border-[#3370ff] focus:ring-1 focus:ring-[#3370ff] bg-white">
                <option value="" disabled>请选择主播</option>
                <option v-for="s in streamers" :key="s.name" :value="s.name">{{ s.name }} ({{ s.files }} 个文件, {{ s.size_gb?.toFixed(1) || 0 }} GB)</option>
              </select>
            </div>

            <div v-if="selectedStreamer">
              <div class="flex items-center justify-between mb-1.5">
                <label class="text-[13px] font-medium text-[#1f2329]">
                  选择文件
                  <span class="text-[#8f959e] font-normal">(已选 {{ selectedFiles.length }}/{{ nonMergedFiles.length }})</span>
                </label>
                <button @click="selectAllNonMerged" class="text-[12px] text-[#3370ff] hover:underline">全选未合并</button>
              </div>
              <div v-if="loadingFiles" class="text-center py-4 text-[#8f959e] text-sm">加载文件列表...</div>
              <div v-else-if="streamerFiles.length === 0" class="text-center py-4 text-[#8f959e] text-sm">无录播文件</div>
              <div v-else class="max-h-[200px] overflow-y-auto border border-[#dee0e3] rounded-lg divide-y divide-[#f0f1f5]">
                <label v-for="f in streamerFiles" :key="f.name"
                  class="flex items-center gap-3 px-3 py-2 hover:bg-[#f8f9fa] cursor-pointer text-[13px]"
                  :class="{ 'opacity-50': f.is_merged }">
                  <input type="checkbox" :checked="selectedFiles.includes(f.name)" @change="toggleFile(f.name)" :disabled="f.is_merged" class="rounded" />
                  <span class="flex-1 truncate text-[#1f2329]">{{ f.name }}</span>
                  <span class="text-[#8f959e] whitespace-nowrap">{{ f.size_str }}</span>
                  <span v-if="f.is_merged" class="text-[12px] text-[#00b578] bg-[#e8f8f0] px-1.5 py-0.5 rounded">已合并</span>
                </label>
              </div>
            </div>

            <div class="flex gap-3">
              <button @click="startMerge" :disabled="sse.isRunning.value || selectedFiles.length < 2"
                class="flex-1 py-2.5 bg-[#3370ff] hover:bg-[#5384ff] disabled:bg-[#a9c4ff] text-white text-[14px] font-medium rounded-lg transition-colors">
                {{ sse.isRunning.value ? '执行中...' : `合并 (${selectedFiles.length} 个文件)` }}
              </button>
              <button @click="startClean" :disabled="sse.isRunning.value"
                class="py-2.5 px-4 border border-[#dee0e3] hover:bg-gray-50 disabled:opacity-50 text-[#f54a45] text-[14px] font-medium rounded-lg transition-colors bg-white">
                清理
              </button>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-[#1e1e1e] rounded-xl p-4 border border-[#333] flex flex-col h-[520px]">
        <div class="flex items-center justify-between mb-3 pb-3 border-b border-[#333]">
          <div class="flex items-center gap-3">
            <div class="flex gap-1.5">
              <div class="w-3 h-3 rounded-full bg-[#ff5f56]"></div>
              <div class="w-3 h-3 rounded-full bg-[#ffbd2e]"></div>
              <div class="w-3 h-3 rounded-full bg-[#27c93f]"></div>
            </div>
            <span class="text-[#858585] text-xs font-mono">task_output.log</span>
          </div>
          <span v-if="sse.isRunning.value" class="text-[#27c93f] text-xs animate-pulse">● 运行中</span>
        </div>
        <div class="flex-1 overflow-y-auto font-mono text-[13px] space-y-1">
          <div v-if="sse.lines.value.length === 0" class="text-[#5c6370]">等待任务触发...</div>
          <div v-for="(line, i) in sse.lines.value" :key="i"
            :class="line.type === 'error' ? 'text-[#ff6b6b]' : line.type === 'success' ? 'text-[#27c93f]' : 'text-[#ccc]'">
            <span class="text-[#5c6370]">[{{ line.time }}]</span> {{ line.text }}
          </div>
          <div v-if="sse.isRunning.value" class="text-[#ccc] animate-pulse">▋</div>
        </div>
      </div>

    </div>
  </div>
</template>

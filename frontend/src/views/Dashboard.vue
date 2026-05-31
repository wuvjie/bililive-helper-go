<script setup>
import { ref, onMounted } from 'vue'
import { useStreamerData } from '../composables/useStreamerData'
import { useConfig } from '../composables/useConfig'
import { useApi } from '../composables/useApi'

const { diskUsage, totalGB, detail } = useStreamerData()
const { schedule } = useConfig()
const api = useApi()

const recentLogs = ref([])

onMounted(async () => {
  try {
    const res = await api.get('/history?limit=5')
    recentLogs.value = Array.isArray(res) ? res.slice(0, 5) : (res.data || []).slice(0, 5)
  } catch (e) {
    console.error('获取动态失败', e)
  }
})

function fmtTime(ts) {
  if (!ts || ts <= 0) return '未排期'
  return new Date(ts * 1000).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function fmtUptime(seconds) {
  if (!seconds) return '--'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  return `${h} 小时 ${m} 分钟`
}
</script>

<template>
  <div class="p-6">
    <h1 class="text-[17px] font-semibold text-[#1f2329] mb-4">概览视图</h1>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-5">
      <div class="bg-white rounded-xl p-5 border border-[#dee0e3] flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-[#e8f3ff] text-[#3370ff] flex items-center justify-center">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
        </div>
        <div>
          <div class="text-[13px] text-[#646a73] mb-1">系统连续运行</div>
          <div class="text-xl font-bold text-[#1f2329]">{{ fmtUptime(detail.uptime) }}</div>
        </div>
      </div>

      <div class="bg-white rounded-xl p-5 border border-[#dee0e3] flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-[#e8f8f0] text-[#00b578] flex items-center justify-center">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581"/></svg>
        </div>
        <div>
          <div class="text-[13px] text-[#646a73] mb-1">下次合并时间</div>
          <div class="text-xl font-bold text-[#1f2329]">{{ schedule?.merge?.enabled ? fmtTime(schedule.merge.next_run) : '已暂停' }}</div>
        </div>
      </div>

      <div class="bg-white rounded-xl p-5 border border-[#dee0e3] flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-[#fff0f0] text-[#f54a45] flex items-center justify-center">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
        </div>
        <div>
          <div class="text-[13px] text-[#646a73] mb-1">下次清理时间</div>
          <div class="text-xl font-bold text-[#1f2329]">{{ schedule?.clean?.enabled ? fmtTime(schedule.clean.next_run) : '已暂停' }}</div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-5">

      <div class="bg-white rounded-xl p-5 border border-[#dee0e3]">
        <h3 class="text-[15px] font-semibold text-[#1f2329] mb-4">系统与存储状态</h3>
        <div class="space-y-5">
          <div>
            <div class="flex items-center justify-between mb-2">
              <span class="text-[14px] text-[#1f2329]">目标挂载盘容量 (NAS)</span>
              <span :class="['text-[14px] font-semibold', diskUsage > 90 ? 'text-[#f54a45]' : 'text-[#1f2329]']">
                {{ diskUsage.toFixed(1) }}%
              </span>
            </div>
            <div class="w-full bg-[#f0f1f5] rounded-full h-2">
              <div :class="['h-2 rounded-full transition-all', diskUsage > 90 ? 'bg-[#f54a45]' : (diskUsage > 80 ? 'bg-[#f88339]' : 'bg-[#00b578]')]" :style="{width: diskUsage + '%'}"></div>
            </div>
            <p class="text-[12px] text-[#8f959e] mt-2">已用 / 总计: {{ totalGB.toFixed(1) }} GB</p>
          </div>
        </div>
      </div>

      <div class="bg-white rounded-xl p-5 border border-[#dee0e3]">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-[15px] font-semibold text-[#1f2329]">系统动态</h3>
          <router-link to="/history" class="text-[13px] text-[#3370ff] hover:underline">查看全部</router-link>
        </div>

        <div v-if="recentLogs.length === 0" class="text-center py-6 text-[#8f959e] text-sm">暂无近期动态</div>
        <ul v-else class="space-y-4">
          <li v-for="(log, i) in recentLogs" :key="i" class="flex gap-3 text-[13px]">
            <div class="text-[#8f959e] whitespace-nowrap">{{ log.time || log.timestamp }}</div>
            <div class="text-[#1f2329] line-clamp-2">{{ log.message || log.detail }}</div>
          </li>
        </ul>
      </div>

    </div>
  </div>
</template>

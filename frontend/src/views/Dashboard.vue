<script setup>
import { ref, computed, onMounted } from 'vue'
import { useStreamerData } from '../composables/useStreamerData'
import { useConfig } from '../composables/useConfig'
import { useStats } from '../composables/useStats'
import { useApi } from '../composables/useApi'

const { diskUsage, totalGB, detail } = useStreamerData()
const { schedule } = useConfig()
const { stats } = useStats()
const api = useApi()

const recentLogs = ref([])
const health = ref(null)

onMounted(async () => {
  try {
    const res = await api.get('/history?per_page=5')
    recentLogs.value = Array.isArray(res) ? res.slice(0, 5) : (res.items || []).slice(0, 5)
  } catch (e) {
    console.error('获取动态失败', e)
  }

  try {
    health.value = await api.get('/setup/check')
  } catch (e) {
    console.error('获取健康状态失败', e)
  }
})

function fmtTime(ts) {
  if (!ts || ts <= 0) return '未排期'
  return new Date(ts * 1000).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function formatBytes(bytes) {
  if (!bytes) return '0'
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
}

const maxDailyCount = computed(() => {
  if (!stats.value.daily?.length) return 1
  return Math.max(1, ...stats.value.daily.map(d => d.merge_count + d.clean_count))
})

function dayLabel(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr + 'T00:00:00')
  return d.toLocaleDateString('zh-CN', { weekday: 'short' })
}
</script>

<template>
  <div class="p-6" style="padding: 32px 40px;">
    <h2 class="section-title">概览视图</h2>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-5">
      <div class="bg-white rounded-xl p-5 border border-[#dee0e3] shadow-sm hover:shadow-md transition-shadow flex items-center gap-4">
        <div class="w-10 h-10 rounded-lg bg-[#e8f3ff] text-[#3370ff] flex items-center justify-center text-xl flex-shrink-0">✨</div>
        <div>
          <div class="text-[13px] text-[#646a73] mb-1">今日合并</div>
          <div class="text-2xl font-bold text-[#1f2329]">{{ stats.today.merge_count }} <span class="text-[13px] font-normal text-[#8f959e]">次</span></div>
          <div class="text-[12px] text-[#8f959e]">{{ formatBytes(stats.today.merge_bytes) }}</div>
        </div>
      </div>

      <div class="bg-white rounded-xl p-5 border border-[#dee0e3] shadow-sm hover:shadow-md transition-shadow flex items-center gap-4">
        <div class="w-10 h-10 rounded-lg bg-[#fff2e8] text-[#f88339] flex items-center justify-center text-xl flex-shrink-0">🧹</div>
        <div>
          <div class="text-[13px] text-[#646a73] mb-1">今日清理</div>
          <div class="text-2xl font-bold text-[#1f2329]">{{ stats.today.clean_count }} <span class="text-[13px] font-normal text-[#8f959e]">次</span></div>
          <div class="text-[12px] text-[#8f959e]">{{ formatBytes(stats.today.clean_bytes) }}</div>
        </div>
      </div>

      <div class="bg-white rounded-xl p-5 border border-[#dee0e3] shadow-sm hover:shadow-md transition-shadow flex items-center gap-4">
        <div class="w-10 h-10 rounded-lg bg-[#e8f8f0] text-[#00b578] flex items-center justify-center text-xl flex-shrink-0">📊</div>
        <div>
          <div class="text-[13px] text-[#646a73] mb-1">本月合并</div>
          <div class="text-2xl font-bold text-[#1f2329]">{{ stats.month.merge_count }} <span class="text-[13px] font-normal text-[#8f959e]">次</span></div>
          <div class="text-[12px] text-[#8f959e]">{{ formatBytes(stats.month.merge_bytes) }}</div>
        </div>
      </div>

      <div class="bg-white rounded-xl p-5 border border-[#dee0e3] shadow-sm hover:shadow-md transition-shadow flex items-center gap-4">
        <div class="w-10 h-10 rounded-lg bg-[#fff0f0] text-[#f54a45] flex items-center justify-center text-xl flex-shrink-0">📦</div>
        <div>
          <div class="text-[13px] text-[#646a73] mb-1">本月清理</div>
          <div class="text-2xl font-bold text-[#1f2329]">{{ stats.month.clean_count }} <span class="text-[13px] font-normal text-[#8f959e]">次</span></div>
          <div class="text-[12px] text-[#8f959e]">{{ formatBytes(stats.month.clean_bytes) }}</div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-5">
      <div class="bg-white rounded-xl p-5 border border-[#dee0e3] shadow-sm hover:shadow-md transition-shadow flex items-center gap-4">
        <div class="w-10 h-10 rounded-lg flex items-center justify-center text-lg" :class="health?.ffmpeg_ok ? 'bg-[#e8f8f0]' : 'bg-[#fff0f0]'">
          {{ health?.ffmpeg_ok ? '✅' : '⚠️' }}
        </div>
        <div>
          <div class="text-[13px] text-[#646a73] mb-1">系统状态</div>
          <div class="text-xl font-bold" :class="health?.ffmpeg_ok && health?.target_dir_exists ? 'text-[#00b578]' : 'text-[#f54a45]'">
            {{ health ? (health.ffmpeg_ok && health.target_dir_exists ? '正常' : '异常') : '...' }}
          </div>
        </div>
      </div>

      <div class="bg-white rounded-xl p-5 border border-[#dee0e3] shadow-sm hover:shadow-md transition-shadow flex items-center gap-4">
        <div class="w-10 h-10 rounded-lg bg-[#e8f8f0] flex items-center justify-center text-lg">🔄</div>
        <div>
          <div class="text-[13px] text-[#646a73] mb-1">下次合并</div>
          <div class="text-xl font-bold text-[#1f2329]">{{ schedule?.merge?.enabled ? fmtTime(schedule.merge.next_run) : '已暂停' }}</div>
        </div>
      </div>

      <div class="bg-white rounded-xl p-5 border border-[#dee0e3] shadow-sm hover:shadow-md transition-shadow flex items-center gap-4">
        <div class="w-10 h-10 rounded-lg bg-[#fff0f0] flex items-center justify-center text-lg">🗑️</div>
        <div>
          <div class="text-[13px] text-[#646a73] mb-1">下次清理</div>
          <div class="text-xl font-bold text-[#1f2329]">{{ schedule?.clean?.enabled ? fmtTime(schedule.clean.next_run) : '已暂停' }}</div>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-5">

      <div class="bg-white rounded-xl p-5 border border-[#dee0e3] shadow-sm hover:shadow-md transition-shadow">
        <h3 class="text-[15px] font-semibold text-[#1f2329] mb-4">存储状态</h3>
        <div>
          <div class="flex items-center justify-between mb-2">
            <span class="text-[14px] text-[#1f2329]">磁盘使用率</span>
            <span :class="['text-[14px] font-semibold', diskUsage > 90 ? 'text-[#f54a45]' : 'text-[#1f2329]']">
              {{ diskUsage.toFixed(1) }}%
            </span>
          </div>
          <div class="w-full bg-[#f0f1f5] rounded-full h-2.5">
            <div :class="['h-2.5 rounded-full transition-all', diskUsage > 90 ? 'bg-[#f54a45]' : (diskUsage > 80 ? 'bg-[#f88339]' : 'bg-[#00b578]')]" :style="{width: diskUsage + '%'}"></div>
          </div>
          <p class="text-[12px] text-[#8f959e] mt-2">总计: {{ totalGB.toFixed(1) }} GB</p>
        </div>
        <div v-if="detail?.pending" class="mt-4 pt-3 border-t border-[#f0f1f5]">
          <div class="flex items-center justify-between text-[14px]">
            <span class="text-[#646a73]">待合并</span>
            <span class="font-semibold text-[#1f2329]">{{ detail.pending.original_files || 0 }} 个 · {{ (detail.pending.original_size_gb || 0).toFixed(2) }} GB</span>
          </div>
        </div>
      </div>

      <div class="bg-white rounded-xl p-5 border border-[#dee0e3] shadow-sm hover:shadow-md transition-shadow">
        <h3 class="text-[15px] font-semibold text-[#1f2329] mb-4">近 7 天趋势</h3>
        <div v-if="stats.daily?.length" class="flex items-end gap-2 h-[140px]">
          <div v-for="day in stats.daily" :key="day.date" class="flex-1 flex flex-col items-center gap-1 h-full justify-end">
            <div class="w-full flex flex-col items-center gap-0.5">
              <div class="w-full max-w-[32px] bg-[#3370ff] rounded-t" :style="{ height: ((day.merge_count / maxDailyCount) * 100) + 'px', minHeight: day.merge_count ? '4px' : '0' }"></div>
              <div class="w-full max-w-[32px] bg-[#f54a45] rounded-b" :style="{ height: ((day.clean_count / maxDailyCount) * 60) + 'px', minHeight: day.clean_count ? '4px' : '0' }"></div>
            </div>
            <span class="text-[10px] text-[#8f959e] mt-1">{{ dayLabel(day.date) }}</span>
          </div>
        </div>
        <div class="flex gap-4 mt-3 text-[12px]">
          <span class="flex items-center gap-1.5"><span class="w-2.5 h-2.5 bg-[#3370ff] rounded-sm"></span> 合并</span>
          <span class="flex items-center gap-1.5"><span class="w-2.5 h-2.5 bg-[#f54a45] rounded-sm"></span> 清理</span>
        </div>
      </div>

      <div class="bg-white rounded-xl p-5 border border-[#dee0e3] shadow-sm hover:shadow-md transition-shadow">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-[15px] font-semibold text-[#1f2329]">最近操作</h3>
          <router-link to="/history" class="text-[13px] text-[#3370ff] hover:underline">查看全部</router-link>
        </div>

        <div v-if="recentLogs.length === 0" class="text-center py-6 text-[#8f959e] text-sm">暂无近期操作</div>
        <ul v-else class="space-y-3">
          <li v-for="(log, i) in recentLogs" :key="i" class="flex items-start gap-3 text-[13px]">
            <span :class="['inline-block w-2 h-2 rounded-full mt-1.5 flex-shrink-0', log.status === 'success' ? 'bg-[#00b578]' : log.status === 'error' ? 'bg-[#f54a45]' : 'bg-[#3370ff]']"></span>
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <span class="text-[#646a73] whitespace-nowrap">{{ log.time }}</span>
                <span class="text-[#1f2329] font-medium">{{ log.task }}</span>
                <span v-if="log.streamer" class="text-[#8f959e]">{{ log.streamer }}</span>
              </div>
              <div v-if="log.detail" class="text-[#646a73] text-[12px] mt-0.5 truncate">{{ log.detail }}</div>
            </div>
          </li>
        </ul>
      </div>

    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useTaskRunner } from '../composables/useTaskRunner'

const router = useRouter()
const { run } = useTaskRunner()

const props = defineProps({
  streamers: { type: Array, default: () => [] },
  diskUsage: { type: Number, default: 0 },
  totalGB: { type: Number, default: 1 },
  detail: { type: Object, default: () => ({}) },
  schedule: { type: Object, default: () => ({}) }
})

const stats = computed(() => [
  { label: '主播总数', value: props.streamers.length, iconColor: 'bg-[#e8f3ff] text-[#3370ff]' },
  { label: '监控中', value: props.streamers.filter(s => s.is_running).length, iconColor: 'bg-[#e8f8f0] text-[#00b578]' },
  { label: '磁盘占用', value: props.diskUsage.toFixed(1) + '%', iconColor: 'bg-[#e8f3ff] text-[#3370ff]' },
  { label: '总容量', value: props.totalGB.toFixed(1) + ' GB', iconColor: 'bg-[#f0e8ff] text-[#8b5cf6]' },
])

function fmtTime(ts) {
  if (!ts || ts <= 0) return '--:--'
  const d = new Date(ts * 1000)
  return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function handleMerge() { if (confirm('确认执行全局合并？')) run('merge', '') }
function handleClean() { if (confirm('确认执行全局清理？')) run('clean', '') }
</script>

<template>
  <div class="p-6">
    <h1 class="text-[17px] font-semibold text-[#1f2329] mb-4">系统概览</h1>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-4">
      <div v-for="(s, i) in stats" :key="i" class="bg-white rounded-xl p-4 border border-[#dee0e3]">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[13px] text-[#646a73]">{{ s.label }}</span>
          <div :class="['w-8 h-8 rounded-lg flex items-center justify-center', s.iconColor]">
            <svg v-if="i===0" class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/></svg>
            <svg v-if="i===1" class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/></svg>
            <svg v-if="i===2" class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/></svg>
            <svg v-if="i===3" class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z"/></svg>
          </div>
        </div>
        <div class="text-[22px] font-bold text-[#1f2329] leading-tight">{{ s.value }}</div>
      </div>
    </div>

    <!-- 快捷操作 + 调度状态 -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-3">
      <div class="lg:col-span-2 bg-white rounded-xl p-4 border border-[#dee0e3]">
        <h3 class="text-[13px] text-[#646a73] mb-3">快捷操作</h3>
        <div class="grid grid-cols-3 gap-3">
          <button @click="handleMerge" class="flex flex-col items-center gap-2 py-4 rounded-xl bg-[#f0f5ff] hover:bg-[#e0ecff] transition-colors">
            <div class="w-10 h-10 rounded-xl bg-[#3370ff] flex items-center justify-center">
              <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581"/></svg>
            </div>
            <span class="text-[13px] text-[#1f2329]">全局合并</span>
          </button>
          <button @click="handleClean" class="flex flex-col items-center gap-2 py-4 rounded-xl bg-[#fff0f0] hover:bg-[#ffe0e0] transition-colors">
            <div class="w-10 h-10 rounded-xl bg-[#f54a45] flex items-center justify-center">
              <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
            </div>
            <span class="text-[13px] text-[#1f2329]">全局清理</span>
          </button>
          <button @click="router.push('/settings')" class="flex flex-col items-center gap-2 py-4 rounded-xl bg-[#f5f6f8] hover:bg-[#eaebef] transition-colors">
            <div class="w-10 h-10 rounded-xl bg-[#86909c] flex items-center justify-center">
              <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37"/><path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
            </div>
            <span class="text-[13px] text-[#1f2329]">系统设置</span>
          </button>
        </div>
      </div>

      <div class="bg-white rounded-xl p-4 border border-[#dee0e3]">
        <h3 class="text-[13px] text-[#646a73] mb-3">调度状态</h3>
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="w-2 h-2 rounded-full bg-[#00b578]"></span>
              <span class="text-[13px] text-[#1f2329]">自动合并</span>
            </div>
            <span class="text-[13px] text-[#86909c]">{{ schedule?.merge?.enabled ? fmtTime(schedule?.merge?.next_run) : '已暂停' }}</span>
          </div>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="w-2 h-2 rounded-full bg-[#00b578]"></span>
              <span class="text-[13px] text-[#1f2329]">自动清理</span>
            </div>
            <span class="text-[13px] text-[#86909c]">{{ schedule?.clean?.enabled ? fmtTime(schedule?.clean?.next_run) : '已暂停' }}</span>
          </div>
          <div class="pt-2 border-t border-[#f0f1f5]">
            <div class="flex items-center justify-between mb-1">
              <span class="text-[13px] text-[#1f2329]">磁盘占用</span>
              <span class="text-[13px] font-semibold text-[#1f2329]">{{ diskUsage.toFixed(1) }}%</span>
            </div>
            <div class="w-full bg-[#f0f1f5] rounded-full h-1.5">
              <div class="h-1.5 rounded-full bg-[#00b578] transition-all" :style="{width: diskUsage + '%'}"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

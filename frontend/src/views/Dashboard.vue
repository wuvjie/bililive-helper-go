<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '../stores/app'
import { useTaskRunner } from '../composables/useTaskRunner'

const router = useRouter()
const appStore = useAppStore()
const { run } = useTaskRunner()

const props = defineProps({
  streamers: { type: Array, default: () => [] },
  diskUsage: { type: Number, default: 0 },
  totalGB: { type: Number, default: 1 },
  detail: { type: Object, default: () => ({}) },
  config: { type: Object, default: () => ({}) },
  schedule: { type: Object, default: () => ({}) }
})

const stats = computed(() => [
  { label: '主播总数', value: props.streamers.length, color: 'text-blue-500 bg-blue-50' },
  { label: '监控中', value: props.streamers.filter(s => s.is_running).length, color: 'text-green-500 bg-green-50' },
  { label: '磁盘占用', value: props.diskUsage.toFixed(1) + '%', color: props.diskUsage > 80 ? 'text-red-500 bg-red-50' : props.diskUsage > 60 ? 'text-amber-500 bg-amber-50' : 'text-blue-500 bg-blue-50' },
  { label: '总容量', value: props.totalGB.toFixed(1) + ' GB', color: 'text-purple-500 bg-purple-50' },
])

function formatTime(ts) {
  if (!ts || ts <= 0) return '--:--'
  const d = new Date(ts * 1000)
  return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function handleMerge() {
  if (confirm('确认执行全局合并？')) run('merge', '')
}
function handleClean() {
  if (confirm('确认执行全局清理？')) run('clean', '')
}
</script>

<template>
  <div class="p-8">
    <h1 class="text-xl font-semibold text-gray-900 mb-6">系统概览</h1>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      <div v-for="(s, i) in stats" :key="i" class="bg-white rounded-xl p-5 border border-gray-100">
        <div class="flex items-center justify-between mb-3">
          <span class="text-sm text-gray-500">{{ s.label }}</span>
          <div :class="['w-9 h-9 rounded-lg flex items-center justify-center', s.color]">
            <svg v-if="i===0" class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/></svg>
            <svg v-if="i===1" class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>
            <svg v-if="i===2" class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/></svg>
            <svg v-if="i===3" class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z"/><path d="M20.489 9H15V3.512A9.025 9.025 0 0120.489 9z"/></svg>
          </div>
        </div>
        <div class="text-2xl font-bold text-gray-900">{{ s.value }}</div>
      </div>
    </div>

    <!-- 快捷操作 + 调度状态 -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <div class="lg:col-span-2 bg-white rounded-xl p-5 border border-gray-100">
        <h3 class="text-sm font-medium text-gray-500 mb-4">快捷操作</h3>
        <div class="grid grid-cols-3 gap-3">
          <button @click="handleMerge" class="flex flex-col items-center gap-2 p-4 rounded-xl bg-blue-50 hover:bg-blue-100 transition-colors">
            <div class="w-10 h-10 rounded-xl bg-blue-500 text-white flex items-center justify-center">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581"/></svg>
            </div>
            <span class="text-sm text-gray-700">全局合并</span>
          </button>
          <button @click="handleClean" class="flex flex-col items-center gap-2 p-4 rounded-xl bg-red-50 hover:bg-red-100 transition-colors">
            <div class="w-10 h-10 rounded-xl bg-red-500 text-white flex items-center justify-center">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
            </div>
            <span class="text-sm text-gray-700">全局清理</span>
          </button>
          <button @click="router.push('/settings')" class="flex flex-col items-center gap-2 p-4 rounded-xl bg-gray-50 hover:bg-gray-100 transition-colors">
            <div class="w-10 h-10 rounded-xl bg-gray-200 text-gray-600 flex items-center justify-center">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37"/><path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
            </div>
            <span class="text-sm text-gray-700">系统设置</span>
          </button>
        </div>
      </div>

      <div class="bg-white rounded-xl p-5 border border-gray-100">
        <h3 class="text-sm font-medium text-gray-500 mb-4">调度状态</h3>
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="w-2 h-2 rounded-full bg-green-500"></span>
              <span class="text-sm text-gray-700">自动合并</span>
            </div>
            <span class="text-sm text-gray-400">{{ schedule?.merge?.enabled ? formatTime(schedule?.merge?.next_run) : '已暂停' }}</span>
          </div>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="w-2 h-2 rounded-full bg-green-500"></span>
              <span class="text-sm text-gray-700">自动清理</span>
            </div>
            <span class="text-sm text-gray-400">{{ schedule?.clean?.enabled ? formatTime(schedule?.clean?.next_run) : '已暂停' }}</span>
          </div>
          <div class="pt-2">
            <div class="flex items-center justify-between mb-1.5">
              <span class="text-sm text-gray-700">磁盘占用</span>
              <span class="text-sm font-semibold text-gray-900">{{ diskUsage.toFixed(1) }}%</span>
            </div>
            <div class="w-full bg-gray-100 rounded-full h-1.5">
              <div class="h-1.5 rounded-full bg-green-500 transition-all" :style="{width: diskUsage + '%'}"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

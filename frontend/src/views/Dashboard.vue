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

// 统计数据
const stats = computed(() => [
  {
    label: '主播总数',
    value: props.streamers.length,
    icon: 'users',
    color: 'blue'
  },
  {
    label: '监控中',
    value: props.streamers.filter(s => s.is_running).length,
    icon: 'play',
    color: 'green'
  },
  {
    label: '磁盘占用',
    value: props.diskUsage.toFixed(1) + '%',
    icon: 'database',
    color: props.diskUsage > 80 ? 'red' : props.diskUsage > 60 ? 'amber' : 'blue'
  },
  {
    label: '总容量',
    value: props.totalGB.toFixed(1) + ' GB',
    icon: 'chart-pie',
    color: 'purple'
  }
])

const colorClasses = {
  blue: 'bg-blue-50 text-blue-600',
  green: 'bg-green-50 text-green-600',
  red: 'bg-red-50 text-red-600',
  amber: 'bg-amber-50 text-amber-600',
  purple: 'bg-purple-50 text-purple-600'
}

function formatTime(ts) {
  if (!ts || ts <= 0) return '--:--'
  const date = new Date(ts * 1000)
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function handleMerge() {
  if (confirm('确认执行全局合并？')) {
    run('merge', '')
  }
}

function handleClean() {
  if (confirm('确认执行全局清理？')) {
    run('clean', '')
  }
}

function goToSettings() {
  router.push('/settings')
}
</script>

<template>
  <div class="p-8">
    <!-- 面包屑 -->
    <div class="mb-8">
      <nav class="flex items-center text-sm text-gray-500 mb-2">
        <span class="text-gray-400">首页</span>
        <svg class="w-4 h-4 mx-2 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
        </svg>
        <span class="text-gray-700 font-medium">仪表盘</span>
      </nav>
      <h1 class="text-3xl font-bold text-gray-900">系统概览</h1>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      <div
        v-for="stat in stats"
        :key="stat.label"
        class="bg-white rounded-xl p-6 border border-gray-100 hover:shadow-lg transition-all"
      >
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 mb-1">{{ stat.label }}</p>
            <p class="text-2xl font-bold text-gray-900">{{ stat.value }}</p>
          </div>
          <div :class="['w-12 h-12 rounded-xl flex items-center justify-center', colorClasses[stat.color]]">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path v-if="stat.icon === 'users'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 7.292 15.242 0 01-4.574 2.077M19 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/>
              <path v-if="stat.icon === 'play'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
              <path v-if="stat.icon === 'database'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/>
              <path v-if="stat.icon === 'chart-pie'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z"/>
            </svg>
          </div>
        </div>
      </div>
    </div>

    <!-- 快捷操作 + 调度状态 -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">

      <!-- 快捷操作 -->
      <div class="lg:col-span-2 bg-white rounded-xl p-6 border border-gray-100">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">快捷操作</h2>
        <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
          <button
            @click="handleMerge"
            class="flex flex-col items-center p-4 bg-gray-50 rounded-xl hover:bg-blue-50 hover:text-blue-600 transition-colors group"
          >
            <div class="w-12 h-12 bg-blue-100 text-blue-600 rounded-xl flex items-center justify-center mb-3 group-hover:bg-blue-200 transition-colors">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
              </svg>
            </div>
            <span class="text-sm font-medium text-gray-700 group-hover:text-blue-600">全局合并</span>
          </button>

          <button
            @click="handleClean"
            class="flex flex-col items-center p-4 bg-gray-50 rounded-xl hover:bg-red-50 hover:text-red-600 transition-colors group"
          >
            <div class="w-12 h-12 bg-red-100 text-red-600 rounded-xl flex items-center justify-center mb-3 group-hover:bg-red-200 transition-colors">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
              </svg>
            </div>
            <span class="text-sm font-medium text-gray-700 group-hover:text-red-600">全局清理</span>
          </button>

          <button
            @click="goToSettings"
            class="flex flex-col items-center p-4 bg-gray-50 rounded-xl hover:bg-purple-50 hover:text-purple-600 transition-colors group"
          >
            <div class="w-12 h-12 bg-purple-100 text-purple-600 rounded-xl flex items-center justify-center mb-3 group-hover:bg-purple-200 transition-colors">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
              </svg>
            </div>
            <span class="text-sm font-medium text-gray-700 group-hover:text-purple-600">系统设置</span>
          </button>
        </div>
      </div>

      <!-- 调度状态 -->
      <div class="bg-white rounded-xl p-6 border border-gray-100">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">调度状态</h2>
        <div class="space-y-4">
          <!-- 自动合并 -->
          <div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
            <div class="flex items-center gap-3">
              <div :class="['w-2 h-2 rounded-full', schedule?.merge?.enabled ? 'bg-green-500' : 'bg-gray-300']"></div>
              <span class="text-sm font-medium text-gray-700">自动合并</span>
            </div>
            <span class="text-sm text-gray-500">
              {{ schedule?.merge?.enabled ? formatTime(schedule?.merge?.next_run) : '已暂停' }}
            </span>
          </div>

          <!-- 自动清理 -->
          <div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
            <div class="flex items-center gap-3">
              <div :class="['w-2 h-2 rounded-full', schedule?.clean?.enabled ? 'bg-green-500' : 'bg-gray-300']"></div>
              <span class="text-sm font-medium text-gray-700">自动清理</span>
            </div>
            <span class="text-sm text-gray-500">
              {{ schedule?.clean?.enabled ? formatTime(schedule?.clean?.next_run) : '已暂停' }}
            </span>
          </div>

          <!-- 磁盘状态 -->
          <div class="p-3 bg-gray-50 rounded-lg">
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm font-medium text-gray-700">磁盘占用</span>
              <span :class="['text-sm font-semibold', diskUsage > 80 ? 'text-red-600' : diskUsage > 60 ? 'text-amber-600' : 'text-green-600']">
                {{ diskUsage.toFixed(1) }}%
              </span>
            </div>
            <div class="w-full bg-gray-200 rounded-full h-2">
              <div
                :class="['h-2 rounded-full transition-all', diskUsage > 80 ? 'bg-red-500' : diskUsage > 60 ? 'bg-amber-500' : 'bg-green-500']"
                :style="{ width: diskUsage + '%' }"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

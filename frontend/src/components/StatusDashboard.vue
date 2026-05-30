<script setup>
defineProps({
  diskUsage: {
    type: Number,
    default: 0
  },
  totalGB: {
    type: Number,
    default: 1
  },
  schedule: {
    type: Object,
    default: null
  },
  running: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['run'])

function formatSimpleTime(ts) {
  if (!ts || ts <= 0) return '--:--'

  const date = new Date(ts * 1000)
  const now = new Date()
  const diff = (date - now) / 1000 / 60

  const h = date.getHours().toString().padStart(2, '0')
  const m = date.getMinutes().toString().padStart(2, '0')
  const t = `${h}:${m}`

  if (Math.abs(diff) < 1) return '即将'

  if (diff < 0) {
    const ago = Math.abs(diff)
    if (ago < 60) return `${Math.round(ago)}m 前`
    return `${Math.round(ago / 60)}h 前`
  }

  return `${t} (${diff < 60 ? Math.round(diff) + 'm' : Math.round(diff / 60) + 'h'}后)`
}

function getDiskColor(usage) {
  if (usage > 85) return 'text-red-600'
  if (usage > 65) return 'text-amber-600'
  return 'text-gray-900'
}

function getDiskBarColor(usage) {
  if (usage > 85) return 'bg-red-500'
  if (usage > 65) return 'bg-amber-500'
  return 'bg-gray-900'
}
</script>

<template>
  <section class="bg-white rounded-xl shadow-sm border border-gray-100 p-6 mb-6">
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 items-center">

      <!-- 磁盘占用仪表盘 -->
      <div class="flex flex-col items-center">
        <div class="relative w-32 h-32">
          <!-- 环形进度条背景 -->
          <svg class="w-32 h-32 transform -rotate-90">
            <circle cx="64" cy="64" r="56" stroke="#e5e7eb" stroke-width="8" fill="none" />
            <circle
              cx="64" cy="64" r="56"
              :class="getDiskBarColor(diskUsage)"
              stroke-width="8"
              fill="none"
              stroke-linecap="round"
              :stroke-dasharray="351.86"
              :stroke-dashoffset="351.86 * (1 - diskUsage / 100)"
              class="transition-all duration-1000 ease-out"
            />
          </svg>
          <!-- 中心数值 -->
          <div class="absolute inset-0 flex flex-col items-center justify-center">
            <span :class="['text-2xl font-bold', getDiskColor(diskUsage)]">
              {{ diskUsage.toFixed(1) }}%
            </span>
            <span class="text-xs text-gray-500 mt-1">
              {{ (totalGB * (diskUsage / 100)).toFixed(1) }}G / {{ totalGB.toFixed(1) }}G
            </span>
          </div>
        </div>
        <span class="text-sm font-medium text-gray-600 mt-3">磁盘占用</span>
      </div>

      <!-- 自动合并状态 -->
      <div class="flex items-center gap-4 p-4 bg-gray-50 rounded-lg">
        <div :class="[
          'w-10 h-10 rounded-full flex items-center justify-center',
          schedule?.merge?.enabled ? 'bg-green-100' : 'bg-gray-100'
        ]">
          <span :class="[
            'w-3 h-3 rounded-full',
            schedule?.merge?.enabled ? 'bg-green-500 animate-pulse' : 'bg-gray-300'
          ]"></span>
        </div>
        <div>
          <p class="text-sm font-medium text-gray-900">自动合并</p>
          <p class="text-sm text-gray-500">
            {{ schedule?.merge?.enabled ? formatSimpleTime(schedule?.merge?.next_run) : '已暂停' }}
          </p>
          <p v-if="schedule?.merge?.enabled" class="text-xs text-gray-400 mt-0.5">
            间隔 {{ schedule?.merge?.interval }} 分钟
          </p>
        </div>
      </div>

      <!-- 自动清理状态 -->
      <div class="flex items-center gap-4 p-4 bg-gray-50 rounded-lg">
        <div :class="[
          'w-10 h-10 rounded-full flex items-center justify-center',
          schedule?.clean?.enabled ? 'bg-blue-100' : 'bg-gray-100'
        ]">
          <span :class="[
            'w-3 h-3 rounded-full',
            schedule?.clean?.enabled ? 'bg-blue-500 animate-pulse' : 'bg-gray-300'
          ]"></span>
        </div>
        <div>
          <p class="text-sm font-medium text-gray-900">自动清理</p>
          <p class="text-sm text-gray-500">
            {{ schedule?.clean?.enabled ? formatSimpleTime(schedule?.clean?.next_run) : '已暂停' }}
          </p>
          <p v-if="schedule?.clean?.enabled" class="text-xs text-gray-400 mt-0.5">
            间隔 {{ schedule?.clean?.interval }} 分钟
          </p>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="flex flex-col gap-3">
        <button
          @click="emit('run', 'merge', '')"
          :disabled="running"
          class="flex items-center justify-center gap-2 px-4 py-3 bg-gradient-to-r from-blue-500 to-blue-600 text-white rounded-lg font-medium text-sm hover:from-blue-600 hover:to-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-sm hover:shadow-md"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
          </svg>
          全局合并
        </button>
        <button
          @click="emit('run', 'clean', '')"
          :disabled="running"
          class="flex items-center justify-center gap-2 px-4 py-3 bg-white border border-gray-200 text-gray-700 rounded-lg font-medium text-sm hover:bg-gray-50 hover:border-gray-300 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
          </svg>
          全局清理
        </button>
      </div>
    </div>
  </section>
</template>

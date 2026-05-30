<script setup>
defineProps({
  schedule: {
    type: Object,
    required: true
  },
  config: {
    type: Object,
    required: true
  }
})

function clampInput(event, max) {
  const value = parseFloat(event.target.value)
  if (!isNaN(value) && value > max) {
    event.target.value = max
    event.target.dispatchEvent(new Event('input'))
  }
}
</script>

<template>
  <div class="space-y-4">
    <h3 class="text-sm font-medium text-gray-700">任务调度</h3>

    <!-- 自动合并任务 -->
    <div class="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
      <div class="flex items-center gap-4">
        <button
          @click="schedule.merge_enabled = !schedule.merge_enabled"
          :class="[
            'relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
            schedule.merge_enabled ? 'bg-blue-600' : 'bg-gray-300'
          ]"
        >
          <span
            :class="[
              'inline-block h-4 w-4 transform rounded-full bg-white transition-transform',
              schedule.merge_enabled ? 'translate-x-6' : 'translate-x-1'
            ]"
          ></span>
        </button>
        <div>
          <p class="font-medium text-gray-900">自动合并任务</p>
          <p class="text-sm text-gray-500">后台定期执行录像碎片拼接</p>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <input
          v-model.number="schedule.merge_interval"
          :disabled="!schedule.merge_enabled"
          type="number"
          min="10" max="1440"
          class="w-20 px-3 py-2 border border-gray-200 rounded-lg text-sm text-right focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50 disabled:bg-gray-100"
        >
        <span class="text-sm text-gray-500">分钟</span>
      </div>
    </div>

    <!-- 自动清理任务 -->
    <div class="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
      <div class="flex items-center gap-4">
        <button
          @click="schedule.clean_enabled = !schedule.clean_enabled"
          :class="[
            'relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
            schedule.clean_enabled ? 'bg-blue-600' : 'bg-gray-300'
          ]"
        >
          <span
            :class="[
              'inline-block h-4 w-4 transform rounded-full bg-white transition-transform',
              schedule.clean_enabled ? 'translate-x-6' : 'translate-x-1'
            ]"
          ></span>
        </button>
        <div>
          <p class="font-medium text-gray-900">自动清理任务</p>
          <p class="text-sm text-gray-500">触及空间警戒阈值时自动释放容量</p>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <input
          v-model.number="schedule.clean_interval"
          :disabled="!schedule.clean_enabled"
          type="number"
          min="10" max="1440"
          class="w-20 px-3 py-2 border border-gray-200 rounded-lg text-sm text-right focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50 disabled:bg-gray-100"
        >
        <span class="text-sm text-gray-500">分钟</span>
      </div>
    </div>

    <!-- 系统静默时段 -->
    <div class="p-4 bg-gray-50 rounded-lg">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <p class="font-medium text-gray-900">系统静默时段</p>
          <p class="text-sm text-gray-500">此时段内自动任务强制休眠</p>
        </div>
        <div class="flex items-center gap-3">
          <div class="flex items-center gap-1 bg-white px-3 py-2 border border-gray-200 rounded-lg">
            <input
              v-model.number="config.BACKUP_START_HOUR"
              type="number"
              min="0" max="23"
              class="w-12 text-center border-0 p-0 focus:outline-none focus:ring-0 text-sm"
            >
            <span class="text-gray-400">:</span>
            <input
              v-model.number="config.BACKUP_START_MINUTE"
              type="number"
              min="0" max="59"
              class="w-12 text-center border-0 p-0 focus:outline-none focus:ring-0 text-sm"
            >
          </div>
          <span class="text-gray-500">至</span>
          <div class="flex items-center gap-1 bg-white px-3 py-2 border border-gray-200 rounded-lg">
            <input
              v-model.number="config.BACKUP_END_HOUR"
              type="number"
              min="0" max="23"
              class="w-12 text-center border-0 p-0 focus:outline-none focus:ring-0 text-sm"
            >
            <span class="text-gray-400">:</span>
            <input
              v-model.number="config.BACKUP_END_MINUTE"
              type="number"
              min="0" max="59"
              class="w-12 text-center border-0 p-0 focus:outline-none focus:ring-0 text-sm"
            >
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useToast } from '../composables/useToast'
import TaskScheduler from './TaskScheduler.vue'

const props = defineProps({
  config: {
    type: Object,
    required: true
  },
  schedule: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['save', 'recommend'])

const { toast } = useToast()

const showHelp = ref(false)

const whitelistStr = computed({
  get() {
    const wl = props.config.WHITELIST_KEYWORDS || []
    return Array.isArray(wl) ? wl.join(',') : wl
  },
  set(val) {
    props.config.WHITELIST_KEYWORDS = val
      .split(/[,，]/)
      .map(s => s.trim())
      .filter(s => s)
  }
})

const safeMinutes = computed({
  get() {
    return props.config.SAFE_AGE_MINUTES != null ? props.config.SAFE_AGE_MINUTES : 120
  },
  set(v) {
    v = parseFloat(v)
    if (isNaN(v)) v = 120
    v = Math.max(1, Math.min(720, Math.round(v)))
    props.config.SAFE_AGE_MINUTES = v
    props.config.SAFE_MODE = 'hours'
    props.config.SAFE_DAYS = Math.max(1, Math.round(v / 60 / 24))
  }
})

// Watch and validate numeric fields
watch(() => props.config.BACKUP_START_HOUR, (v) => {
  if (typeof v !== 'number' || isNaN(v)) props.config.BACKUP_START_HOUR = 0
  else if (v < 0) props.config.BACKUP_START_HOUR = 0
  else if (v > 23) props.config.BACKUP_START_HOUR = 23
})

watch(() => props.config.BACKUP_END_HOUR, (v) => {
  if (typeof v !== 'number' || isNaN(v)) props.config.BACKUP_END_HOUR = 0
  else if (v < 0) props.config.BACKUP_END_HOUR = 0
  else if (v > 23) props.config.BACKUP_END_HOUR = 23
})

watch(() => props.config.BACKUP_START_MINUTE, (v) => {
  if (typeof v !== 'number' || isNaN(v)) props.config.BACKUP_START_MINUTE = 0
  else if (v < 0) props.config.BACKUP_START_MINUTE = 0
  else if (v > 59) props.config.BACKUP_START_MINUTE = 59
})

watch(() => props.config.BACKUP_END_MINUTE, (v) => {
  if (typeof v !== 'number' || isNaN(v)) props.config.BACKUP_END_MINUTE = 0
  else if (v < 0) props.config.BACKUP_END_MINUTE = 0
  else if (v > 59) props.config.BACKUP_END_MINUTE = 59
})

watch(() => props.config.TRIGGER_THRESHOLD, (v) => {
  if (typeof v === 'number' && !isNaN(v)) {
    props.config.TRIGGER_THRESHOLD = Math.max(0, Math.min(100, v))
  }
})

watch(() => props.config.TARGET_THRESHOLD, (v) => {
  if (typeof v === 'number' && !isNaN(v)) {
    props.config.TARGET_THRESHOLD = Math.max(0, Math.min(100, v))
  }
})

function validate() {
  if (props.config.TRIGGER_THRESHOLD < 0 || props.config.TRIGGER_THRESHOLD > 100) {
    toast('阈值必须在 0-100 之间', 'err')
    return false
  }
  if (props.config.TARGET_THRESHOLD < 0 || props.config.TARGET_THRESHOLD > 100) {
    toast('阈值必须在 0-100 之间', 'err')
    return false
  }
  return true
}

function handleSave() {
  if (validate()) {
    emit('save')
  }
}
</script>

<template>
  <div class="h-full overflow-y-auto">
    <!-- 帮助说明 -->
    <div class="p-4 border-b border-gray-100">
      <button
        @click="showHelp = !showHelp"
        class="flex items-center gap-2 text-sm text-gray-600 hover:text-gray-900 transition-colors"
      >
        <svg :class="['w-4 h-4 transition-transform', showHelp ? 'rotate-90' : '']" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
        </svg>
        {{ showHelp ? '收起说明' : '帮助说明' }}
      </button>

      <Transition name="slide">
        <div v-if="showHelp" class="mt-4 p-4 bg-blue-50 rounded-lg text-sm text-gray-700 space-y-2">
          <p><strong>空间警戒阈值：</strong>自动清理的触发红线。建议设为 80% - 90%。</p>
          <p><strong>安全回落水位：</strong>自动清理的目标终点。建议与警戒阈值保持 10% - 20% 差值。</p>
          <p><strong>保底留存件数：</strong>无论空间多紧缺，每人强制保留的最新录像数。</p>
          <p><strong>新文件保护期：</strong>刚结束录制的视频在此时间内绝对免删，防止误清。</p>
        </div>
      </Transition>
    </div>

    <!-- 表单内容 -->
    <div class="p-5 space-y-6">

      <!-- 存储目录 -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-2">录像存储目录</label>
        <input
          v-model="config.TARGET_DIR"
          type="text"
          class="w-full px-4 py-3 border border-gray-200 rounded-lg text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        >
      </div>

      <!-- 免删名单 -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-2">免删保护名单</label>
        <input
          v-model="whitelistStr"
          type="text"
          placeholder="留存,高能,勿删"
          class="w-full px-4 py-3 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        >
        <p class="text-xs text-gray-500 mt-1.5">含此词缀录像绝对免删（逗号分隔）</p>
      </div>

      <!-- 阈值设置 -->
      <div class="grid grid-cols-2 gap-5">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">空间警戒阈值</label>
          <div class="flex items-center gap-2">
            <input
              v-model.number="config.TRIGGER_THRESHOLD"
              type="number"
              min="0" max="100"
              class="w-24 px-4 py-3 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
            <span class="text-gray-500">%</span>
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">安全回落水位</label>
          <div class="flex items-center gap-2">
            <input
              v-model.number="config.TARGET_THRESHOLD"
              type="number"
              min="0" max="100"
              class="w-24 px-4 py-3 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
            <span class="text-gray-500">%</span>
          </div>
        </div>
      </div>

      <!-- 其他设置 -->
      <div class="grid grid-cols-2 gap-5">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">保底留存件数</label>
          <div class="flex items-center gap-2">
            <input
              v-model.number="config.MIN_KEEP_PER_STREAMER"
              type="number"
              min="1" max="50"
              class="w-24 px-4 py-3 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
            <span class="text-gray-500">个</span>
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">单次清理限额</label>
          <div class="flex items-center gap-2">
            <input
              v-model.number="config.MAX_DELETE_PER_RUN"
              type="number"
              min="1" max="100"
              class="w-24 px-4 py-3 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
            <span class="text-gray-500">个</span>
          </div>
        </div>
      </div>

      <!-- 时间设置 -->
      <div class="grid grid-cols-3 gap-5">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">新文件保护期</label>
          <div class="flex items-center gap-2">
            <input
              v-model.number="safeMinutes"
              type="number"
              min="1" max="720"
              class="w-24 px-4 py-3 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
            <span class="text-gray-500">分钟</span>
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">断流分割判定</label>
          <div class="flex items-center gap-2">
            <input
              v-model.number="config.GAP_MINUTES"
              type="number"
              min="1" max="1440"
              class="w-24 px-4 py-3 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
            <span class="text-gray-500">分钟</span>
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">合并延迟缓冲</label>
          <div class="flex items-center gap-2">
            <input
              v-model.number="config.MERGE_AGE_MINUTES"
              type="number"
              min="0" max="1440"
              class="w-24 px-4 py-3 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
            <span class="text-gray-500">分钟</span>
          </div>
        </div>
      </div>

      <!-- 任务调度 -->
      <TaskScheduler :schedule="schedule" :config="config" />

      <!-- 操作按钮 -->
      <div class="flex items-center gap-3 pt-4 border-t border-gray-100">
        <button
          @click="emit('recommend')"
          class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
        >
          智能推荐
        </button>
        <button
          @click="handleSave"
          class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors"
        >
          保存配置
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.slide-enter-from,
.slide-leave-to {
  max-height: 0;
  opacity: 0;
  margin-top: 0;
  padding-top: 0;
  padding-bottom: 0;
}

.slide-enter-to,
.slide-leave-from {
  max-height: 200px;
  opacity: 1;
}
</style>

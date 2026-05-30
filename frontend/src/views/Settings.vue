<script setup>
import { ref, computed, watch } from 'vue'
import { useToast } from '../composables/useToast'

const props = defineProps({
  config: { type: Object, required: true },
  schedule: { type: Object, required: true }
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
    props.config.WHITELIST_KEYWORDS = val.split(/[,，]/).map(s => s.trim()).filter(s => s)
  }
})

const safeMinutes = computed({
  get() {
    return props.config.SAFE_AGE_MINUTES != null ? props.config.SAFE_AGE_MINUTES : 120
  },
  set(v) {
    v = parseFloat(v) || 120
    v = Math.max(1, Math.min(720, Math.round(v)))
    props.config.SAFE_AGE_MINUTES = v
    props.config.SAFE_MODE = 'hours'
    props.config.SAFE_DAYS = Math.max(1, Math.round(v / 60 / 24))
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
  if (validate()) emit('save')
}
</script>

<template>
  <div class="p-6">
    <!-- 面包屑 -->
    <div class="mb-6">
      <nav class="flex items-center text-sm text-gray-500">
        <span class="text-gray-400">首页</span>
        <svg class="w-4 h-4 mx-2 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
        </svg>
        <span class="text-gray-700 font-medium">系统设置</span>
      </nav>
      <h1 class="mt-2 text-2xl font-bold text-gray-900">系统设置</h1>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- 左侧：设置表单 -->
      <div class="lg:col-span-2 space-y-6">
        <!-- 基础设置 -->
        <div class="bg-white rounded-xl p-6 border border-gray-100">
          <h2 class="text-lg font-semibold text-gray-900 mb-4">基础设置</h2>

          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">录像存储目录</label>
              <input
                v-model="config.TARGET_DIR"
                type="text"
                class="w-full px-4 py-3 border border-gray-200 rounded-lg text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">免删保护名单</label>
              <input
                v-model="whitelistStr"
                type="text"
                placeholder="留存,高能,勿删"
                class="w-full px-4 py-3 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
              <p class="text-xs text-gray-500 mt-1">含此词缀录像绝对免删（逗号分隔）</p>
            </div>
          </div>
        </div>

        <!-- 阈值设置 -->
        <div class="bg-white rounded-xl p-6 border border-gray-100">
          <h2 class="text-lg font-semibold text-gray-900 mb-4">阈值设置</h2>

          <div class="grid grid-cols-2 gap-6">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">空间警戒阈值</label>
              <div class="flex items-center gap-2">
                <input v-model.number="config.TRIGGER_THRESHOLD" type="number" min="0" max="100" class="w-24 px-4 py-3 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
                <span class="text-gray-500">%</span>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">安全回落水位</label>
              <div class="flex items-center gap-2">
                <input v-model.number="config.TARGET_THRESHOLD" type="number" min="0" max="100" class="w-24 px-4 py-3 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
                <span class="text-gray-500">%</span>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">保底留存件数</label>
              <div class="flex items-center gap-2">
                <input v-model.number="config.MIN_KEEP_PER_STREAMER" type="number" min="1" max="50" class="w-24 px-4 py-3 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
                <span class="text-gray-500">个</span>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">单次清理限额</label>
              <div class="flex items-center gap-2">
                <input v-model.number="config.MAX_DELETE_PER_RUN" type="number" min="1" max="100" class="w-24 px-4 py-3 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
                <span class="text-gray-500">个</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 时间设置 -->
        <div class="bg-white rounded-xl p-6 border border-gray-100">
          <h2 class="text-lg font-semibold text-gray-900 mb-4">时间设置</h2>

          <div class="grid grid-cols-3 gap-6">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">新文件保护期</label>
              <div class="flex items-center gap-2">
                <input v-model.number="safeMinutes" type="number" min="1" max="720" class="w-24 px-4 py-3 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
                <span class="text-gray-500">分钟</span>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">断流分割判定</label>
              <div class="flex items-center gap-2">
                <input v-model.number="config.GAP_MINUTES" type="number" min="1" max="1440" class="w-24 px-4 py-3 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
                <span class="text-gray-500">分钟</span>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">合并延迟缓冲</label>
              <div class="flex items-center gap-2">
                <input v-model.number="config.MERGE_AGE_MINUTES" type="number" min="0" max="1440" class="w-24 px-4 py-3 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
                <span class="text-gray-500">分钟</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="flex gap-4">
          <button @click="emit('recommend')" class="px-6 py-3 text-sm font-medium text-gray-700 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors">
            智能推荐
          </button>
          <button @click="handleSave" class="px-6 py-3 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors">
            保存配置
          </button>
        </div>
      </div>

      <!-- 右侧：任务调度 -->
      <div class="bg-white rounded-xl p-6 border border-gray-100 h-fit">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">任务调度</h2>

        <div class="space-y-4">
          <!-- 自动合并 -->
          <div class="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
            <div class="flex items-center gap-3">
              <button
                @click="schedule.merge_enabled = !schedule.merge_enabled"
                :class="['relative w-11 h-6 rounded-full transition-colors', schedule.merge_enabled ? 'bg-blue-600' : 'bg-gray-300']"
              >
                <span :class="['absolute top-1 w-4 h-4 bg-white rounded-full transition-transform', schedule.merge_enabled ? 'left-6' : 'left-1']"></span>
              </button>
              <div>
                <p class="font-medium text-gray-900">自动合并</p>
                <p class="text-xs text-gray-500">后台定期执行</p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <input v-model.number="schedule.merge_interval" :disabled="!schedule.merge_enabled" type="number" min="10" max="1440" class="w-20 px-3 py-2 border border-gray-200 rounded-lg text-sm text-right focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50">
              <span class="text-sm text-gray-500">分钟</span>
            </div>
          </div>

          <!-- 自动清理 -->
          <div class="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
            <div class="flex items-center gap-3">
              <button
                @click="schedule.clean_enabled = !schedule.clean_enabled"
                :class="['relative w-11 h-6 rounded-full transition-colors', schedule.clean_enabled ? 'bg-blue-600' : 'bg-gray-300']"
              >
                <span :class="['absolute top-1 w-4 h-4 bg-white rounded-full transition-transform', schedule.clean_enabled ? 'left-6' : 'left-1']"></span>
              </button>
              <div>
                <p class="font-medium text-gray-900">自动清理</p>
                <p class="text-xs text-gray-500">触及阈值时执行</p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <input v-model.number="schedule.clean_interval" :disabled="!schedule.clean_enabled" type="number" min="10" max="1440" class="w-20 px-3 py-2 border border-gray-200 rounded-lg text-sm text-right focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50">
              <span class="text-sm text-gray-500">分钟</span>
            </div>
          </div>

          <!-- 静默时段 -->
          <div class="p-4 bg-gray-50 rounded-lg">
            <p class="font-medium text-gray-900 mb-3">系统静默时段</p>
            <div class="flex items-center gap-3">
              <input v-model.number="config.BACKUP_START_HOUR" type="number" min="0" max="23" class="w-16 px-3 py-2 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
              <span class="text-gray-400">:</span>
              <input v-model.number="config.BACKUP_START_MINUTE" type="number" min="0" max="59" class="w-16 px-3 py-2 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
              <span class="text-gray-500">至</span>
              <input v-model.number="config.BACKUP_END_HOUR" type="number" min="0" max="23" class="w-16 px-3 py-2 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
              <span class="text-gray-400">:</span>
              <input v-model.number="config.BACKUP_END_MINUTE" type="number" min="0" max="59" class="w-16 px-3 py-2 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
            </div>
            <p class="text-xs text-gray-500 mt-2">此时段内自动任务强制休眠</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

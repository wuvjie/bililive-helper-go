<script setup>
import { ref, computed } from 'vue'
import { useToast } from '../composables/useToast'

const props = defineProps({
  config: { type: Object, required: true },
  schedule: { type: Object, required: true }
})

const emit = defineEmits(['save', 'recommend'])
const { toast } = useToast()

const whitelistStr = computed({
  get: () => (props.config.WHITELIST_KEYWORDS || []).join(','),
  set: v => { props.config.WHITELIST_KEYWORDS = v.split(/[,，]/).map(s => s.trim()).filter(Boolean) }
})

const safeMinutes = computed({
  get: () => props.config.SAFE_AGE_MINUTES ?? 120,
  set: v => {
    v = Math.max(1, Math.min(720, Math.round(parseFloat(v) || 120)))
    props.config.SAFE_AGE_MINUTES = v
    props.config.SAFE_MODE = 'hours'
    props.config.SAFE_DAYS = Math.max(1, Math.round(v / 60 / 24))
  }
})

function handleSave() {
  if (props.config.TRIGGER_THRESHOLD < 0 || props.config.TRIGGER_THRESHOLD > 100) { toast('阈值必须在 0-100', 'err'); return }
  emit('save')
}
</script>

<template>
  <div class="p-8">
    <h1 class="text-xl font-semibold text-gray-900 mb-5">系统设置</h1>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-5">
      <!-- 左侧设置 -->
      <div class="lg:col-span-2 space-y-4">
        <!-- 基础设置 -->
        <div class="bg-white rounded-xl p-5 border border-gray-100">
          <h3 class="text-sm font-medium text-gray-900 mb-4">基础设置</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm text-gray-600 mb-1.5">录像存储目录</label>
              <input v-model="config.TARGET_DIR" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent">
            </div>
            <div>
              <label class="block text-sm text-gray-600 mb-1.5">免删保护名单</label>
              <input v-model="whitelistStr" placeholder="留存,高能,勿删" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent">
              <p class="text-xs text-gray-400 mt-1">含此词缀的录像绝对免删，逗号分隔</p>
            </div>
          </div>
        </div>

        <!-- 阈值设置 -->
        <div class="bg-white rounded-xl p-5 border border-gray-100">
          <h3 class="text-sm font-medium text-gray-900 mb-4">阈值设置</h3>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm text-gray-600 mb-1.5">空间警戒阈值</label>
              <div class="flex items-center gap-1.5">
                <input v-model.number="config.TRIGGER_THRESHOLD" type="number" min="0" max="100" class="w-20 px-3 py-2 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
                <span class="text-sm text-gray-400">%</span>
              </div>
            </div>
            <div>
              <label class="block text-sm text-gray-600 mb-1.5">安全回落水位</label>
              <div class="flex items-center gap-1.5">
                <input v-model.number="config.TARGET_THRESHOLD" type="number" min="0" max="100" class="w-20 px-3 py-2 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
                <span class="text-sm text-gray-400">%</span>
              </div>
            </div>
            <div>
              <label class="block text-sm text-gray-600 mb-1.5">保底留存件数</label>
              <div class="flex items-center gap-1.5">
                <input v-model.number="config.MIN_KEEP_PER_STREAMER" type="number" min="1" max="50" class="w-20 px-3 py-2 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
                <span class="text-sm text-gray-400">个</span>
              </div>
            </div>
            <div>
              <label class="block text-sm text-gray-600 mb-1.5">单次清理限额</label>
              <div class="flex items-center gap-1.5">
                <input v-model.number="config.MAX_DELETE_PER_RUN" type="number" min="1" max="100" class="w-20 px-3 py-2 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
                <span class="text-sm text-gray-400">个</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 时间设置 -->
        <div class="bg-white rounded-xl p-5 border border-gray-100">
          <h3 class="text-sm font-medium text-gray-900 mb-4">时间设置</h3>
          <div class="grid grid-cols-3 gap-4">
            <div>
              <label class="block text-sm text-gray-600 mb-1.5">新文件保护期</label>
              <div class="flex items-center gap-1.5">
                <input v-model.number="safeMinutes" type="number" min="1" max="720" class="w-20 px-3 py-2 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
                <span class="text-sm text-gray-400">分钟</span>
              </div>
            </div>
            <div>
              <label class="block text-sm text-gray-600 mb-1.5">断流分割判定</label>
              <div class="flex items-center gap-1.5">
                <input v-model.number="config.GAP_MINUTES" type="number" min="1" max="1440" class="w-20 px-3 py-2 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
                <span class="text-sm text-gray-400">分钟</span>
              </div>
            </div>
            <div>
              <label class="block text-sm text-gray-600 mb-1.5">合并延迟缓冲</label>
              <div class="flex items-center gap-1.5">
                <input v-model.number="config.MERGE_AGE_MINUTES" type="number" min="0" max="1440" class="w-20 px-3 py-2 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
                <span class="text-sm text-gray-400">分钟</span>
              </div>
            </div>
          </div>
        </div>

        <div class="flex gap-3">
          <button @click="emit('recommend')" class="px-4 py-2 text-sm font-medium text-gray-600 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors">智能推荐</button>
          <button @click="handleSave" class="px-5 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors">保存配置</button>
        </div>
      </div>

      <!-- 右侧调度 -->
      <div class="bg-white rounded-xl p-5 border border-gray-100 h-fit">
        <h3 class="text-sm font-medium text-gray-900 mb-4">任务调度</h3>
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <button @click="schedule.merge_enabled=!schedule.merge_enabled" :class="['relative w-10 h-[22px] rounded-full transition-colors', schedule.merge_enabled?'bg-blue-600':'bg-gray-300']">
                <span :class="['absolute top-[3px] w-4 h-4 bg-white rounded-full transition-transform shadow-sm', schedule.merge_enabled?'left-[22px]':'left-[3px]']"></span>
              </button>
              <div><div class="text-sm font-medium text-gray-900">自动合并</div><div class="text-xs text-gray-400">后台定期执行</div></div>
            </div>
            <div class="flex items-center gap-1.5">
              <input v-model.number="schedule.merge_interval" :disabled="!schedule.merge_enabled" type="number" min="10" max="1440" class="w-16 px-2 py-1.5 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-40">
              <span class="text-xs text-gray-400">分钟</span>
            </div>
          </div>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <button @click="schedule.clean_enabled=!schedule.clean_enabled" :class="['relative w-10 h-[22px] rounded-full transition-colors', schedule.clean_enabled?'bg-blue-600':'bg-gray-300']">
                <span :class="['absolute top-[3px] w-4 h-4 bg-white rounded-full transition-transform shadow-sm', schedule.clean_enabled?'left-[22px]':'left-[3px]']"></span>
              </button>
              <div><div class="text-sm font-medium text-gray-900">自动清理</div><div class="text-xs text-gray-400">触及阈值时执行</div></div>
            </div>
            <div class="flex items-center gap-1.5">
              <input v-model.number="schedule.clean_interval" :disabled="!schedule.clean_enabled" type="number" min="10" max="1440" class="w-16 px-2 py-1.5 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-40">
              <span class="text-xs text-gray-400">分钟</span>
            </div>
          </div>
          <div class="pt-2 border-t border-gray-100">
            <div class="text-sm font-medium text-gray-900 mb-2">系统静默时段</div>
            <div class="flex items-center gap-1.5">
              <input v-model.number="config.BACKUP_START_HOUR" type="number" min="0" max="23" class="w-12 px-2 py-1.5 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
              <span class="text-gray-400">:</span>
              <input v-model.number="config.BACKUP_START_MINUTE" type="number" min="0" max="59" class="w-12 px-2 py-1.5 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
              <span class="text-xs text-gray-400 mx-1">至</span>
              <input v-model.number="config.BACKUP_END_HOUR" type="number" min="0" max="23" class="w-12 px-2 py-1.5 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
              <span class="text-gray-400">:</span>
              <input v-model.number="config.BACKUP_END_MINUTE" type="number" min="0" max="59" class="w-12 px-2 py-1.5 border border-gray-200 rounded-lg text-sm text-center focus:outline-none focus:ring-2 focus:ring-blue-500">
            </div>
            <p class="text-xs text-gray-400 mt-1.5">此时段内自动任务强制休眠</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

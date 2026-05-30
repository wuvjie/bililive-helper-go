<script setup>
import { ref, computed, watch } from 'vue'
import { useApi } from '../composables/useApi'
import { useToast } from '../composables/useToast'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  streamer: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['close', 'merge'])

const { get } = useApi()
const { toast } = useToast()

const files = ref([])
const selected = ref([])
const loading = ref(false)

const allChecked = computed(() => {
  return files.value.length > 0 && selected.value.length === files.value.length
})

const selectedSize = computed(() => {
  return files.value
    .filter(f => selected.value.includes(f.name))
    .reduce((sum, f) => sum + f.size, 0)
})

async function fetchFiles() {
  if (!props.streamer) return

  loading.value = true
  files.value = []
  selected.value = []

  try {
    const data = await get(`/api/files/${encodeURIComponent(props.streamer)}`)
    files.value = data || []
  } catch (err) {
    if (err.message !== '未登录') {
      toast(err.message || '拉取失败', 'err')
      emit('close')
    }
  } finally {
    loading.value = false
  }
}

function toggleFile(name) {
  const idx = selected.value.indexOf(name)
  if (idx === -1) {
    selected.value.push(name)
  } else {
    selected.value.splice(idx, 1)
  }
}

function toggleAll() {
  if (allChecked.value) {
    selected.value = []
  } else {
    selected.value = files.value.map(f => f.name)
  }
}

function formatSize(bytes) {
  if (bytes >= 1 << 30) return (bytes / (1 << 30)).toFixed(2) + ' GB'
  if (bytes >= 1 << 20) return (bytes / (1 << 20)).toFixed(2) + ' MB'
  return (bytes / 1024).toFixed(2) + ' KB'
}

function handleMerge() {
  if (selected.value.length < 2) return
  emit('merge', [...selected.value])
}

watch(() => props.visible, (newVal) => {
  if (newVal) {
    fetchFiles()
  }
})
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center p-4" @click.self="emit('close')">
        <!-- 背景遮罩 -->
        <div class="absolute inset-0 bg-black/50 backdrop-blur-sm"></div>

        <!-- 弹窗内容 -->
        <div class="relative bg-white rounded-2xl shadow-2xl w-full max-w-3xl max-h-[85vh] flex flex-col transform transition-all">

          <!-- 头部 -->
          <div class="px-6 py-4 border-b border-gray-200">
            <div class="flex items-center justify-between">
              <h2 class="text-lg font-semibold text-gray-900">手动合并 — {{ streamer }}</h2>
              <button
                @click="emit('close')"
                class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
              </button>
            </div>
          </div>

          <!-- 文件列表 -->
          <div class="flex-1 overflow-y-auto p-6">
            <div v-if="loading" class="flex flex-col items-center justify-center py-12">
              <div class="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-500 mb-4"></div>
              <p class="text-gray-500">加载文件列表...</p>
            </div>

            <div v-else-if="files.length === 0" class="flex flex-col items-center justify-center py-12 text-gray-400">
              <svg class="w-12 h-12 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5 19a2 2 0 01-2-2V7a2 2 0 012-2h4l2 2h4a2 2 0 012 2v1M5 19h14a2 2 0 002-2v-5a2 2 0 00-2-2H9a2 2 0 00-2 2v5a2 2 0 01-2 2z"/>
              </svg>
              <p>暂无视频文件</p>
            </div>

            <div v-else class="space-y-2">
              <!-- 全选 -->
              <label class="flex items-center gap-3 p-3 bg-gray-50 rounded-lg cursor-pointer hover:bg-gray-100 transition-colors">
                <input
                  type="checkbox"
                  :checked="allChecked"
                  @change="toggleAll"
                  class="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                >
                <span class="text-sm font-medium text-gray-700">全选 ({{ files.length }} 个文件)</span>
              </label>

              <!-- 文件列表 -->
              <div class="space-y-1">
                <div
                  v-for="f in files"
                  :key="f.name"
                  @click="toggleFile(f.name)"
                  :class="[
                    'flex items-center gap-3 p-3 rounded-lg cursor-pointer transition-colors',
                    selected.includes(f.name) ? 'bg-blue-50 border border-blue-200' : 'hover:bg-gray-50 border border-transparent'
                  ]"
                >
                  <input
                    type="checkbox"
                    :checked="selected.includes(f.name)"
                    @click.stop="toggleFile(f.name)"
                    class="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  >
                  <div class="flex-1 min-w-0">
                    <p class="text-sm font-medium text-gray-900 truncate">{{ f.name }}</p>
                    <p class="text-xs text-gray-500">{{ f.size_str }}</p>
                  </div>
                  <span :class="[
                    'px-2 py-1 text-xs font-medium rounded',
                    f.is_merged ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'
                  ]">
                    {{ f.is_merged ? '合并' : '原片' }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- 底部操作栏 -->
          <div class="px-6 py-4 border-t border-gray-200 bg-gray-50 rounded-b-2xl">
            <div class="flex items-center justify-between">
              <div class="text-sm text-gray-500">
                已选 {{ selected.length }} 个文件
                <span v-if="selected.length > 0" class="text-gray-400">
                  ({{ formatSize(selectedSize) }})
                </span>
              </div>
              <div class="flex items-center gap-3">
                <button
                  @click="selected = []"
                  :disabled="selected.length === 0"
                  class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  清空
                </button>
                <button
                  @click="handleMerge"
                  :disabled="selected.length < 2"
                  class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  开始合并
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: all 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .relative,
.modal-leave-to .relative {
  transform: scale(0.95) translateY(10px);
}
</style>

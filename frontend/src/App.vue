<script setup>
import { ref, onMounted } from 'vue'
import TopNav from './components/TopNav.vue'
import StatusDashboard from './components/StatusDashboard.vue'
import StreamerList from './components/StreamerList.vue'
import TaskHistory from './components/TaskHistory.vue'
import LogViewer from './components/LogViewer.vue'
import ConfigPanel from './components/ConfigPanel.vue'
import ManualMerge from './components/ManualMerge.vue'
import { useStreamerData } from './composables/useStreamerData'
import { useConfig } from './composables/useConfig'
import { useTaskRunner } from './composables/useTaskRunner'
import { useToast } from './composables/useToast'

const {
  streamers,
  diskUsage,
  totalGB,
  detail,
  fetchStatus,
  fetchDetail
} = useStreamerData()

const {
  config,
  schedule,
  fetchConfig,
  fetchSchedule,
  saveConfig,
  saveSchedule,
  resetSchedule
} = useConfig()

const {
  running,
  progress,
  progressLabel,
  logs,
  autoScroll,
  run,
  startManualMerge
} = useTaskRunner()

const { toast } = useToast()

// Modal states
const showMergeModal = ref(false)
const selectedStreamer = ref('')
const showConfirmModal = ref(false)
const confirmTask = ref('')
const confirmTarget = ref('')
const confirmLabel = ref('')

// Tab navigation
const activeTab = ref('history')
const tabs = [
  { id: 'history', name: '操作历史' },
  { id: 'logs', name: '系统日志' },
  { id: 'config', name: '设置' },
]

function handleRun(task, streamer) {
  confirmTask.value = task
  confirmTarget.value = streamer || '全局'
  confirmLabel.value = task === 'merge' ? '合并' : '清理'
  showConfirmModal.value = true
}

function confirmRun() {
  showConfirmModal.value = false
  const onComplete = () => {
    fetchStatus()
    fetchDetail()
    fetchConfig()
  }
  run(confirmTask.value, confirmTarget.value, { onComplete })
}

function handleOpenManualMerge(streamer) {
  selectedStreamer.value = streamer
  showMergeModal.value = true
}

function handleMerge(files) {
  showMergeModal.value = false
  const onComplete = () => {
    fetchStatus()
    fetchDetail()
    fetchConfig()
  }
  startManualMerge(selectedStreamer.value, files, { onComplete })
}

async function handleSaveConfig() {
  await saveConfig()
  await saveSchedule()
}

async function handleRecommend() {
  toast('正在分析配置...', 'info')
}

// Initialize
onMounted(async () => {
  await fetchConfig()
  await fetchStatus()
  await fetchSchedule()
  await fetchDetail()

  setInterval(async () => {
    if (!running.value) {
      await fetchDetail()
    }
  }, 60000)
})
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Top navigation -->
    <TopNav :running="running" />

    <!-- Main content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">

      <!-- Dashboard -->
      <StatusDashboard
        :disk-usage="diskUsage"
        :total-g-b="totalGB"
        :schedule="detail.schedule"
        :running="running"
        @run="handleRun"
      />

      <!-- Two-column layout (desktop) -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">

        <!-- Left: Streamer list (2/3) -->
        <div class="lg:col-span-2">
          <StreamerList
            :streamers="streamers"
            :total-g-b="totalGB"
            :running="running"
            :loading="false"
            @run="handleRun"
            @open-manual-merge="handleOpenManualMerge"
            @refresh="fetchStatus"
          />
        </div>

        <!-- Right: Tab panel (1/3) -->
        <div class="bg-white rounded-lg shadow-sm overflow-hidden">
          <!-- Tab switcher -->
          <div class="flex border-b border-gray-200">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              @click="activeTab = tab.id"
              :class="[
                'flex-1 px-4 py-3 text-sm font-medium transition-colors',
                activeTab === tab.id
                  ? 'text-blue-600 border-b-2 border-blue-600 bg-blue-50'
                  : 'text-gray-500 hover:text-gray-700 hover:bg-gray-50'
              ]"
            >
              {{ tab.name }}
            </button>
          </div>

          <!-- Tab content -->
          <div class="h-[480px] overflow-hidden">
            <TaskHistory v-show="activeTab === 'history'" />
            <LogViewer v-show="activeTab === 'logs'" />
            <ConfigPanel v-show="activeTab === 'config'" :config="config" :schedule="schedule" @save="handleSaveConfig" @recommend="handleRecommend" />
          </div>
        </div>
      </div>
    </main>

    <!-- Manual merge modal -->
    <ManualMerge
      :visible="showMergeModal"
      :streamer="selectedStreamer"
      @close="showMergeModal = false"
      @merge="handleMerge"
    />

    <!-- Confirm run modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showConfirmModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm" @click.self="showConfirmModal = false">
          <div class="bg-white rounded-xl shadow-2xl w-full max-w-md mx-4 overflow-hidden transform transition-all">
            <div class="p-6">
              <div class="flex items-center justify-between mb-4">
                <h3 class="text-lg font-semibold text-gray-900">确认{{ confirmLabel }}</h3>
                <button @click="showConfirmModal = false" class="text-gray-400 hover:text-gray-600">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                </button>
              </div>
              <p class="text-gray-600 mb-2">
                即将对 <span class="font-semibold text-gray-900">{{ confirmTarget }}</span> 执行 <span class="font-semibold text-gray-900">{{ confirmLabel }}</span>
              </p>
              <p class="text-sm text-gray-500">此操作将立即开始，且无法中途撤销。</p>
            </div>
            <div class="px-6 py-4 bg-gray-50 border-t flex justify-end gap-3">
              <button @click="showConfirmModal = false" class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border rounded-lg hover:bg-gray-50">
                取消
              </button>
              <button @click="confirmRun" class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700">
                确认执行
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style>
/* Transition animations */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(10px) scale(0.98);
}
</style>

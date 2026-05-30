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

// Manual merge modal
const showMergeModal = ref(false)
const selectedStreamer = ref('')

// Confirm run modal
const showConfirmModal = ref(false)
const confirmTask = ref('')
const confirmTarget = ref('')
const confirmLabel = ref('')

// Log tab
const activeTab = ref('history')

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
  // TODO: Implement recommend
}

// Initialize
onMounted(async () => {
  await fetchConfig()
  await fetchStatus()
  await fetchSchedule()
  await fetchDetail()

  // Auto-refresh every minute
  setInterval(async () => {
    if (!running.value) {
      await fetchDetail()
    }
  }, 60000)
})
</script>

<template>
  <TopNav :running="running" />

  <StatusDashboard
    :disk-usage="diskUsage"
    :total-g-b="totalGB"
    :schedule="detail.schedule"
    :running="running"
    @run="handleRun"
  />

  <div class="main-grid">
    <StreamerList
      :streamers="streamers"
      :total-g-b="totalGB"
      :running="running"
      :loading="false"
      @run="handleRun"
      @open-manual-merge="handleOpenManualMerge"
      @refresh="fetchStatus"
    />

    <div class="card">
      <div class="card-head" style="gap: 12px;">
        <select v-model="activeTab" class="tab-select" style="cursor:pointer; width: auto; min-width: 90px;">
          <option value="history">操作历史</option>
          <option value="logs">系统日志</option>
        </select>
      </div>

      <div style="flex:1; display:flex; flex-direction:column; overflow:hidden">
        <TaskHistory v-show="activeTab === 'history'" />
        <LogViewer v-show="activeTab === 'logs'" />
      </div>
    </div>
  </div>

  <ConfigPanel
    :config="config"
    :schedule="schedule"
    @save="handleSaveConfig"
    @recommend="handleRecommend"
  />

  <ManualMerge
    :visible="showMergeModal"
    :streamer="selectedStreamer"
    @close="showMergeModal = false"
    @merge="handleMerge"
  />

  <!-- Confirm Run Modal -->
  <div v-if="showConfirmModal" class="overlay" @click.self="showConfirmModal = false">
    <div class="modal">
      <div class="card-head" style="display:flex; justify-content:space-between;">
        <h2 style="margin:0;">确认{{ confirmLabel }}</h2>
        <button class="modal-x" @click="showConfirmModal = false">&times;</button>
      </div>
      <div style="padding:32px 24px;text-align:center">
        <p style="font-size:15px;color:var(--text);font-weight:500;margin-bottom:8px">
          即将对 <strong>{{ confirmTarget }}</strong> 执行 <b>{{ confirmLabel }}</b>
        </p>
        <p style="font-size:13px;color:var(--muted)">此操作将立即开始，且无法中途撤销。</p>
      </div>
      <div class="card-foot" style="display:flex;gap:12px;justify-content:flex-end">
        <button class="btn btn-ghost auto-w" @click="showConfirmModal = false">取消</button>
        <button class="btn btn-pri auto-w" @click="confirmRun">确认执行</button>
      </div>
    </div>
  </div>
</template>

<style>
.main-grid {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(0, 1fr);
  gap: var(--gap);
  align-items: stretch;
  margin-bottom: var(--gap);
}

@media (max-width: 1024px) {
  .main-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

.card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--card-shadow);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.card-head {
  height: var(--row-h);
  padding: 0 16px;
  border-bottom: 1px solid var(--border);
  background: color-mix(in srgb, var(--bg-sub) 50%, transparent);
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.card-body {
  padding: 20px;
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.card-foot {
  padding: 12px 20px;
  background: var(--bg-sub);
  border-top: 1px solid var(--border);
}

.tab-select {
  border: none !important;
  background: transparent !important;
  font-weight: 600;
  font-size: 14px;
  padding-left: 0;
  box-shadow: none !important;
  color: var(--text) !important;
}

.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  animation: fade-in 0.2s ease-out;
}

.modal {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 40px -10px rgba(0,0,0,0.2);
  animation: pop-up 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-x {
  background: transparent;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: var(--muted);
  line-height: 1;
  padding: 4px;
  border-radius: 6px;
  transition: background 0.2s;
}

.modal-x:hover {
  background: var(--hover);
  color: var(--text);
}

@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes pop-up {
  0% { opacity: 0; transform: scale(0.96) translateY(10px); }
  100% { opacity: 1; transform: scale(1) translateY(0); }
}
</style>

<template>
  <div class="tasks-page">
    <div class="layout-grid">

      <div class="control-panel">
        <div class="feishu-card">
          <h3 class="card-title">触发手动任务</h3>
          <p class="desc">针对漏录或未自动触发的视频分段进行手动合并。</p>

          <div class="form-group">
            <label>主播标识 (Streamer ID)</label>
            <input v-model="form.streamer" type="text" class="feishu-input" placeholder="例如: douyin_game_001" />
          </div>
          <div class="form-group">
            <label>目标日期 (Date)</label>
            <input v-model="form.date" type="text" class="feishu-input" placeholder="例如: 2026-05-30" />
          </div>
          <button class="feishu-btn" style="width: 100%;" @click="triggerTask" :disabled="isSubmitting">
            <span class="icon">🚀</span> {{ isSubmitting ? '正在提交...' : '立即执行合并' }}
          </button>
        </div>

        <div class="feishu-card mt-20">
          <h3 class="card-title">当前任务队列</h3>
          <ul class="queue-list">
            <li class="queue-item running">
              <span class="status-icon">🔄</span>
              <div class="task-info">
                <div class="task-name">合并任务: dy_outdoor_xx</div>
                <div class="task-progress">
                  <div class="progress-bar" style="width: 65%;"></div>
                </div>
              </div>
            </li>
            <li class="queue-item pending">
              <span class="status-icon">⏳</span>
              <div class="task-info">
                <div class="task-name">合并任务: bili_vup_123</div>
                <div class="task-meta">等待执行中...</div>
              </div>
            </li>
          </ul>
        </div>
      </div>

      <div class="terminal-panel feishu-card">
        <div class="terminal-header">
          <div class="mac-buttons">
            <span class="mac-btn red"></span>
            <span class="mac-btn yellow"></span>
            <span class="mac-btn green"></span>
          </div>
          <div class="terminal-title">system_merger.log (Live)</div>
          <button class="clear-btn" @click="logs = []">清屏</button>
        </div>

        <div class="terminal-window" ref="terminalWindow">
          <div v-for="(log, index) in logs" :key="index" class="log-line" :class="log.level">
            <span class="log-time">[{{ log.time }}]</span>
            <span class="log-level">[{{ log.level.toUpperCase() }}]</span>
            <span class="log-msg">{{ log.message }}</span>
          </div>
          <div class="log-line"><span class="cursor">_</span></div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'

const isSubmitting = ref(false)
const form = reactive({ streamer: '', date: '' })
const terminalWindow = ref(null)

// 模拟初始日志
const logs = ref([
  { time: '14:22:01', level: 'info', message: 'BiliLive Helper Task Scheduler started.' },
  { time: '14:22:05', level: 'info', message: 'Found 1 pending task in queue.' },
  { time: '14:22:06', level: 'info', message: 'Starting ffmpeg merge for dy_outdoor_xx [2026-05-30]' },
  { time: '14:22:08', level: 'debug', message: 'ffmpeg -f concat -safe 0 -i list.txt -c copy output.mp4' },
])

const scrollToBottom = () => {
  nextTick(() => {
    if (terminalWindow.value) {
      terminalWindow.value.scrollTop = terminalWindow.value.scrollHeight
    }
  })
}

const triggerTask = () => {
  if (!form.streamer || !form.date) return alert('请填写完整信息')
  isSubmitting.value = true

  // 模拟终端写入动画
  setTimeout(() => {
    logs.value.push({ time: new Date().toLocaleTimeString('en-GB'), level: 'info', message: `Manual trigger received for: ${form.streamer}` })
    scrollToBottom()
  }, 500)

  setTimeout(() => {
    logs.value.push({ time: new Date().toLocaleTimeString('en-GB'), level: 'warning', message: `Task queued at position #3.` })
    scrollToBottom()
    isSubmitting.value = false
    form.streamer = ''
    form.date = ''
  }, 1200)
}

onMounted(() => scrollToBottom())
</script>

<style scoped>
.layout-grid { display: grid; grid-template-columns: 320px 1fr; gap: 24px; height: calc(100vh - 120px); }
.card-title { margin: 0 0 8px 0; font-size: 16px; font-weight: 600; color: var(--text-title); }
.desc { font-size: 13px; color: var(--text-placeholder); margin-bottom: 20px; line-height: 1.5; }
.mt-20 { margin-top: 20px; }

/* 表单与队列 */
.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 13px; font-weight: 500; color: var(--text-title); margin-bottom: 8px; }
.queue-list { list-style: none; padding: 0; margin: 0; }
.queue-item { display: flex; gap: 12px; padding: 12px 0; border-bottom: 1px solid var(--border-color); }
.queue-item:last-child { border-bottom: none; }
.status-icon { font-size: 16px; animation: spin 2s linear infinite; }
.queue-item.pending .status-icon { animation: none; filter: grayscale(1); }
@keyframes spin { 100% { transform: rotate(360deg); } }
.task-info { flex: 1; }
.task-name { font-size: 13px; font-weight: 500; color: var(--text-title); margin-bottom: 6px; }
.task-meta { font-size: 12px; color: var(--text-placeholder); }
.task-progress { height: 4px; background: #ebedf0; border-radius: 2px; overflow: hidden; }
.progress-bar { height: 100%; background: var(--color-primary); }

/* 终端样式 */
.terminal-panel { display: flex; flex-direction: column; padding: 0; overflow: hidden; background: #1e1e1e; border: 1px solid #333; }
.terminal-header { height: 40px; background: #2d2d2d; display: flex; align-items: center; justify-content: space-between; padding: 0 16px; border-bottom: 1px solid #000; }
.mac-buttons { display: flex; gap: 8px; }
.mac-btn { width: 12px; height: 12px; border-radius: 50%; }
.mac-btn.red { background: #ff5f56; }
.mac-btn.yellow { background: #ffbd2e; }
.mac-btn.green { background: #27c93f; }
.terminal-title { color: #858585; font-size: 13px; font-family: monospace; font-weight: 500; }
.clear-btn { background: none; border: none; color: #858585; cursor: pointer; font-size: 12px; }
.clear-btn:hover { color: #fff; }

.terminal-window { flex: 1; padding: 16px; overflow-y: auto; font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace; font-size: 13px; line-height: 1.6; color: #cccccc; }
.log-time { color: #858585; margin-right: 8px; }
.log-level { margin-right: 8px; font-weight: 600; }
.log-line.info .log-level { color: #3b8eea; }
.log-line.warning .log-level { color: #d19a66; }
.log-line.warning .log-msg { color: #d19a66; }
.log-line.error .log-level { color: #e06c75; }
.log-line.error .log-msg { color: #e06c75; }
.log-line.debug .log-level { color: #98c379; }
.log-line.debug .log-msg { color: #5c6370; }
.cursor { animation: blink 1s step-end infinite; background: #fff; color: transparent; display: inline-block; width: 8px; }
@keyframes blink { 50% { opacity: 0; } }
</style>

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

const emit = defineEmits(['save', 'saveSchedule', 'resetSchedule', 'recommend'])

const { toast } = useToast()

const activeHelp = ref('')

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
  if (typeof v !== 'number' || isNaN(v)) {
    props.config.BACKUP_START_HOUR = 0
  } else if (v < 0) {
    props.config.BACKUP_START_HOUR = 0
  } else if (v > 23) {
    props.config.BACKUP_START_HOUR = 23
  }
})

watch(() => props.config.BACKUP_END_HOUR, (v) => {
  if (typeof v !== 'number' || isNaN(v)) {
    props.config.BACKUP_END_HOUR = 0
  } else if (v < 0) {
    props.config.BACKUP_END_HOUR = 0
  } else if (v > 23) {
    props.config.BACKUP_END_HOUR = 23
  }
})

watch(() => props.config.BACKUP_START_MINUTE, (v) => {
  if (typeof v !== 'number' || isNaN(v)) {
    props.config.BACKUP_START_MINUTE = 0
  } else if (v < 0) {
    props.config.BACKUP_START_MINUTE = 0
  } else if (v > 59) {
    props.config.BACKUP_START_MINUTE = 59
  }
})

watch(() => props.config.BACKUP_END_MINUTE, (v) => {
  if (typeof v !== 'number' || isNaN(v)) {
    props.config.BACKUP_END_MINUTE = 0
  } else if (v < 0) {
    props.config.BACKUP_END_MINUTE = 0
  } else if (v > 59) {
    props.config.BACKUP_END_MINUTE = 59
  }
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
  <div class="card" style="margin-bottom:24px;">
    <div class="card-head" style="display:flex; justify-content:space-between;">
      <h2 style="margin:0;">基础策略</h2>
      <div style="display:flex;gap:12px">
        <button class="btn btn-ghost auto-w" @click="activeHelp = activeHelp === 'config' ? '' : 'config'">
          {{ activeHelp === 'config' ? '收起说明' : '帮助说明' }}
        </button>
        <button class="btn btn-ghost auto-w" @click="emit('recommend')">智能推荐</button>
        <button class="btn btn-pri auto-w" @click="handleSave">保存配置</button>
      </div>
    </div>

    <div class="help-wrap" :class="{open: activeHelp === 'config'}">
      <div class="help">
        <dl>
          <dt>空间警戒阈值</dt><dd>自动清理的触发红线。建议设为 80% - 90%。</dd>
          <dt>安全回落水位</dt><dd>自动清理的目标终点。建议与警戒阈值保持 10% - 20% 差值。</dd>
          <dt>保底留存件数</dt><dd>无论空间多紧缺，每人强制保留的最新录像数。</dd>
          <dt>新文件保护期</dt><dd>刚结束录制的视频在此时间内绝对免删，防止误清。</dd>
          <dt>断流分割判定</dt><dd>断流超过此时长，合并时将切分为新场次。</dd>
          <dt>合并延迟缓冲</dt><dd>录制结束后推迟合并的安全等待时间，防止文件损坏。</dd>
          <dt>单次清理限额</dt><dd>每轮清理最多允许删除的文件数，防止失控误删。</dd>
        </dl>
      </div>
    </div>

    <div class="card-body">
      <div class="setting-grid" style="grid-template-columns: 1fr 1fr;">
        <label class="setting-box" style="align-items:flex-start;flex-direction:column;gap:12px;cursor:text">
          <div class="setting-info" style="margin:0;width:100%">
            <div class="setting-title">录像存储目录</div>
            <div class="setting-desc">扫描与清理的基础路径</div>
          </div>
          <input type="text" v-model="config.TARGET_DIR" style="font-family:var(--font-mono);width:100%;background:var(--card);box-shadow:inset 0 1px 2px rgba(0,0,0,.02)">
        </label>

        <label class="setting-box" style="align-items:flex-start;flex-direction:column;gap:12px;cursor:text">
          <div class="setting-info" style="margin:0;width:100%">
            <div class="setting-title">免删保护名单</div>
            <div class="setting-desc">含此词缀录像绝对免删（逗号分隔）</div>
          </div>
          <input type="text" v-model="whitelistStr" style="width:100%;background:var(--card);box-shadow:inset 0 1px 2px rgba(0,0,0,.02)" placeholder="留存,高能,勿删">
        </label>
      </div>

      <div class="setting-grid" style="grid-template-columns: 1fr 1fr;">
        <label class="setting-box">
          <div class="setting-info">
            <div class="setting-title">空间警戒阈值</div>
            <div class="setting-desc">触发自动清理的占用红线</div>
          </div>
          <div class="setting-control">
            <input type="number" v-model.number="config.TRIGGER_THRESHOLD" min="0" max="100">
            <span class="setting-unit">%</span>
          </div>
        </label>

        <label class="setting-box">
          <div class="setting-info">
            <div class="setting-title">安全回落水位</div>
            <div class="setting-desc">清理后的空间安全底线</div>
          </div>
          <div class="setting-control">
            <input type="number" v-model.number="config.TARGET_THRESHOLD" min="0" max="100">
            <span class="setting-unit">%</span>
          </div>
        </label>
      </div>

      <div class="setting-grid" style="grid-template-columns: 1fr 1fr;">
        <label class="setting-box">
          <div class="setting-info">
            <div class="setting-title">保底留存件数</div>
            <div class="setting-desc">主播强制保留的最新件数</div>
          </div>
          <div class="setting-control">
            <input type="number" v-model.number="config.MIN_KEEP_PER_STREAMER" min="1" max="50">
            <span class="setting-unit">个</span>
          </div>
        </label>

        <label class="setting-box">
          <div class="setting-info">
            <div class="setting-title">单次清理限额</div>
            <div class="setting-desc">每轮清理最多删除的文件数</div>
          </div>
          <div class="setting-control">
            <input type="number" v-model.number="config.MAX_DELETE_PER_RUN" min="1" max="100">
            <span class="setting-unit">个</span>
          </div>
        </label>
      </div>

      <div class="setting-grid" style="grid-template-columns: 1fr 1fr 1fr;">
        <label class="setting-box">
          <div class="setting-info">
            <div class="setting-title">新文件保护期</div>
            <div class="setting-desc">录制完的视频绝对免删时长</div>
          </div>
          <div class="setting-control">
            <input type="number" v-model.number="safeMinutes" min="1" max="720">
            <span class="setting-unit">MIN</span>
          </div>
        </label>

        <label class="setting-box">
          <div class="setting-info">
            <div class="setting-title">断流分割判定</div>
            <div class="setting-desc">断流超时即切分为新场次</div>
          </div>
          <div class="setting-control">
            <input type="number" v-model.number="config.GAP_MINUTES" min="1" max="1440">
            <span class="setting-unit">MIN</span>
          </div>
        </label>

        <label class="setting-box">
          <div class="setting-info">
            <div class="setting-title">合并延迟缓冲</div>
            <div class="setting-desc">录像后延迟合并的等待时间</div>
          </div>
          <div class="setting-control">
            <input type="number" v-model.number="config.MERGE_AGE_MINUTES" min="0" max="1440">
            <span class="setting-unit">MIN</span>
          </div>
        </label>
      </div>

      <!-- Task Scheduler -->
      <TaskScheduler :schedule="schedule" :config="config" />
    </div>
  </div>
</template>

<style scoped>
.card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--card-shadow);
  overflow: hidden;
}

.card-head {
  height: var(--row-h);
  padding: 0 16px;
  border-bottom: 1px solid var(--border);
  background: color-mix(in srgb, var(--bg-sub) 50%, transparent);
  display: flex;
  align-items: center;
  flex-shrink: 0;
  justify-content: space-between;
}

.card-body {
  padding: 20px;
  overflow-y: auto;
}

.setting-grid {
  display: grid;
  gap: 16px;
  margin-bottom: 16px;
}

.setting-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  background: var(--bg-sub);
  border: 1px solid var(--border);
  border-radius: 8px;
  transition: all 0.25s ease;
  cursor: pointer;
}

.setting-box:hover {
  border-color: var(--muted);
}

.setting-box:focus-within {
  border-color: var(--pri);
  background: var(--card);
  box-shadow: 0 0 0 3px var(--pri-r);
}

.setting-info {
  flex: 1;
  margin-right: 16px;
}

.setting-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
}

.setting-desc {
  font-size: 12px;
  color: var(--muted);
  margin-top: 4px;
  line-height: 1.5;
}

.setting-control {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.setting-unit {
  font-size: 12px;
  font-weight: 600;
  color: var(--muted);
  width: 28px;
  text-align: left;
  display: inline-block;
}

.help-wrap {
  overflow: hidden;
  max-height: 0;
  transition: max-height 0.4s ease;
}

.help-wrap.open {
  max-height: 600px;
}

.help {
  background: var(--bg-sub);
  border-left: 3px solid var(--muted);
  border-radius: 0 8px 8px 0;
  padding: 16px 20px;
  margin: 0 20px 20px;
  font-size: 12px;
  line-height: 1.6;
}

.help dt {
  font-weight: 600;
  color: var(--text);
  display: inline-block;
  margin-right: 6px;
}

.help dt::after {
  content: ":";
  color: var(--muted);
}

.help dd {
  display: inline;
  color: var(--text2);
}

.help dd::after {
  content: "";
  display: block;
  margin-bottom: 8px;
}

@media (max-width: 768px) {
  .setting-grid {
    grid-template-columns: 1fr !important;
    gap: 10px;
  }

  .setting-box {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
    padding: 12px;
  }

  .setting-info {
    margin-right: 0;
  }

  .setting-control {
    justify-content: flex-end;
  }
}
</style>

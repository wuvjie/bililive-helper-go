<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const api = useApi()
const currentTab = ref('general')
const saving = ref(false)
const saveMsg = ref('')
const config = reactive({})
const scheduleForm = reactive({
  merge_interval: 60,
  clean_interval: 360,
  merge_enabled: true,
  clean_enabled: true,
  BACKUP_START_HOUR: 4,
  BACKUP_START_MINUTE: 0,
  BACKUP_END_HOUR: 12,
  BACKUP_END_MINUTE: 0
})
const health = ref(null)
const cleanEstimate = ref(null)
const recommend = ref(null)
const importText = ref('')
const importMsg = ref('')
const exportData = ref(null)

const WHITELIST_INPUT = ref('')

async function loadConfig() {
  try {
    const res = await api.get('/config')
    Object.assign(config, res)
    WHITELIST_INPUT.value = (res.WHITELIST_KEYWORDS || []).join(', ')
  } catch (e) {
    console.error('加载配置失败', e)
  }
}

async function loadSchedule() {
  try {
    const res = await api.get('/schedule')
    scheduleForm.merge_interval = res.merge?.interval || 60
    scheduleForm.clean_interval = res.clean?.interval || 360
    scheduleForm.merge_enabled = res.merge?.enabled ?? true
    scheduleForm.clean_enabled = res.clean?.enabled ?? true
  } catch (e) {
    console.error('加载调度失败', e)
  }
}

async function loadHealth() {
  try {
    health.value = await api.get('/setup/check')
  } catch (e) { /* ignore */ }
}

async function loadCleanEstimate() {
  try {
    cleanEstimate.value = await api.get('/clean/estimate')
  } catch (e) { /* ignore */ }
}

async function loadRecommend() {
  try {
    recommend.value = await api.get('/config/recommend')
  } catch (e) { /* ignore */ }
}

async function saveConfig() {
  saving.value = true
  saveMsg.value = ''
  try {
    const payload = { ...config }
    if (WHITELIST_INPUT.value.trim()) {
      payload.WHITELIST_KEYWORDS = WHITELIST_INPUT.value.split(',').map(s => s.trim()).filter(Boolean)
    } else {
      payload.WHITELIST_KEYWORDS = []
    }
    await api.post('/config', payload)
    saveMsg.value = '配置已保存'
    setTimeout(() => { saveMsg.value = '' }, 3000)
  } catch (e) {
    saveMsg.value = '保存失败: ' + e.message
  } finally {
    saving.value = false
  }
}

async function saveSchedule() {
  saving.value = true
  saveMsg.value = ''
  try {
    await api.post('/schedule', scheduleForm)
    saveMsg.value = '调度设置已保存'
    setTimeout(() => { saveMsg.value = '' }, 3000)
  } catch (e) {
    saveMsg.value = '保存失败: ' + e.message
  } finally {
    saving.value = false
  }
}

async function doExport() {
  try {
    exportData.value = await api.get('/config/export')
  } catch (e) {
    alert('导出失败: ' + e.message)
  }
}

function downloadExport() {
  if (!exportData.value) return
  const blob = new Blob([JSON.stringify(exportData.value, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `bililive-helper-config-${new Date().toISOString().split('T')[0]}.json`
  a.click()
  URL.revokeObjectURL(url)
}

async function doImport() {
  importMsg.value = ''
  try {
    const data = JSON.parse(importText.value)
    await api.post('/config/import', data)
    importMsg.value = '导入成功'
    loadConfig()
    loadSchedule()
  } catch (e) {
    importMsg.value = '导入失败: ' + e.message
  }
}

function applyRecommend() {
  if (!recommend.value) return
  const keys = ['TRIGGER_THRESHOLD', 'TARGET_THRESHOLD', 'MIN_KEEP_PER_STREAMER',
    'SAFE_AGE_MINUTES', 'MERGE_AGE_MINUTES', 'MAX_DELETE_PER_RUN', 'GAP_MINUTES']
  for (const k of keys) {
    if (recommend.value[k] !== undefined) config[k] = recommend.value[k]
  }
  saveMsg.value = '已应用推荐配置，请点击保存'
}

function tabClass(tab) {
  return currentTab.value === tab ? 'active' : ''
}

onMounted(() => {
  loadConfig()
  loadSchedule()
})
</script>

<template>
  <div class="settings-page">
    <div class="settings-layout">
      <div class="settings-sidebar">
        <h3 class="nav-header">系统设置</h3>
        <ul class="settings-nav">
          <li :class="tabClass('general')" @click="currentTab = 'general'">⚙️ 通用配置</li>
          <li :class="tabClass('storage')" @click="currentTab = 'storage'; loadCleanEstimate()">💾 存储管理</li>
          <li :class="tabClass('schedule')" @click="currentTab = 'schedule'">⏰ 定时调度</li>
          <li :class="tabClass('health')" @click="currentTab = 'health'; loadHealth()">🩺 系统诊断</li>
          <li :class="tabClass('recommend')" @click="currentTab = 'recommend'; loadRecommend()">🤖 AI 推荐</li>
          <li :class="tabClass('backup')" @click="currentTab = 'backup'; doExport()">📦 配置备份</li>
        </ul>
      </div>

      <div class="settings-content">

        <!-- 通用配置 -->
        <div v-show="currentTab === 'general'" class="settings-section">
          <h2 class="section-title">通用配置</h2>

          <div class="form-row">
            <label>录播存储目录</label>
            <input v-model="config.TARGET_DIR" type="text" class="feishu-input" placeholder="/path/to/recordings" />
            <div class="help-text">所有主播录播文件的根目录。</div>
          </div>

          <div class="form-row">
            <label>合并间隔判定 (分钟)</label>
            <input v-model.number="config.MERGE_AGE_MINUTES" type="number" min="5" max="180" class="feishu-input" style="width: 120px;" />
            <div class="help-text">相邻分片间隔超过此时间则视为不同场次，分开合并。默认 30 分钟。</div>
          </div>

          <div class="form-row">
            <label>分片间隔判定 (分钟)</label>
            <input v-model.number="config.GAP_MINUTES" type="number" min="1" max="120" class="feishu-input" style="width: 120px;" />
            <div class="help-text">同场次内分片的最大允许间隔。默认 60 分钟。</div>
          </div>

          <div class="form-row">
            <label>安全模式</label>
            <select v-model="config.SAFE_MODE" class="feishu-input" style="width: 200px;">
              <option value="hours">按小时保护</option>
              <option value="days">按天保护</option>
              <option value="count">按数量保护</option>
              <option value="off">关闭</option>
            </select>
            <div class="help-text">保护策略：在安全范围内的文件不会被自动清理。</div>
          </div>

          <div class="form-row" v-if="config.SAFE_MODE === 'hours'">
            <label>安全保护时长 (小时)</label>
            <input v-model.number="config.SAFE_AGE_MINUTES" type="number" min="30" class="feishu-input" style="width: 120px;" />
            <div class="help-text">此时间内的录播文件不会被清理。实际单位为分钟。</div>
          </div>

          <div class="form-row" v-if="config.SAFE_MODE === 'days'">
            <label>安全保护天数</label>
            <input v-model.number="config.SAFE_DAYS" type="number" min="1" class="feishu-input" style="width: 120px;" />
          </div>

          <div class="form-row">
            <label>单次最大删除数</label>
            <input v-model.number="config.MAX_DELETE_PER_RUN" type="number" min="1" max="100" class="feishu-input" style="width: 120px;" />
            <div class="help-text">每次清理任务最多删除的文件数。默认 10。</div>
          </div>

          <div class="form-row">
            <label>白名单关键词</label>
            <input v-model="WHITELIST_INPUT" type="text" class="feishu-input" placeholder="关键词1, 关键词2, ..." />
            <div class="help-text">包含这些关键词的主播或文件不会被清理。逗号分隔。</div>
          </div>

          <div class="form-actions">
            <button class="feishu-btn" @click="saveConfig" :disabled="saving">{{ saving ? '保存中...' : '保存配置' }}</button>
            <span v-if="saveMsg" class="save-msg" :class="saveMsg.includes('失败') ? 'error' : 'success'">{{ saveMsg }}</span>
          </div>
        </div>

        <!-- 存储管理 -->
        <div v-show="currentTab === 'storage'" class="settings-section">
          <h2 class="section-title">存储管理</h2>

          <div class="alert-box warning" v-if="cleanEstimate">
            <span>⚠️</span>
            当前有 <strong>{{ cleanEstimate.file_count }}</strong> 个可清理文件，共 <strong>{{ cleanEstimate.total_size_gb?.toFixed(2) }} GB</strong>
          </div>

          <div class="form-row">
            <label>清理触发阈值 (%)</label>
            <div class="slider-group">
              <input type="range" min="50" max="99" v-model.number="config.TRIGGER_THRESHOLD" class="feishu-slider" />
              <span class="slider-value">{{ config.TRIGGER_THRESHOLD }}%</span>
            </div>
            <div class="help-text">磁盘使用率达到此值时自动触发清理任务。</div>
          </div>

          <div class="form-row">
            <label>清理目标水位 (%)</label>
            <div class="slider-group">
              <input type="range" min="30" max="89" v-model.number="config.TARGET_THRESHOLD" class="feishu-slider" />
              <span class="slider-value">{{ config.TARGET_THRESHOLD }}%</span>
            </div>
            <div class="help-text">清理持续进行直到磁盘使用率回落到此值。</div>
          </div>

          <div class="form-row">
            <label>每主播最少保留文件数</label>
            <input v-model.number="config.MIN_KEEP_PER_STREAMER" type="number" min="0" max="50" class="feishu-input" style="width: 120px;" />
            <div class="help-text">清理时每个主播至少保留的最新文件数。默认 3。</div>
          </div>

          <div class="form-actions">
            <button class="feishu-btn" @click="saveConfig" :disabled="saving">{{ saving ? '保存中...' : '保存配置' }}</button>
            <span v-if="saveMsg" class="save-msg" :class="saveMsg.includes('失败') ? 'error' : 'success'">{{ saveMsg }}</span>
          </div>
        </div>

        <!-- 定时调度 -->
        <div v-show="currentTab === 'schedule'" class="settings-section">
          <h2 class="section-title">定时调度</h2>

          <div class="form-row">
            <label class="checkbox-label">
              <input type="checkbox" v-model="scheduleForm.merge_enabled" />
              启用自动合并
            </label>
          </div>
          <div class="form-row" v-if="scheduleForm.merge_enabled">
            <label>合并任务间隔 (分钟)</label>
            <input v-model.number="scheduleForm.merge_interval" type="number" min="10" max="1440" class="feishu-input" style="width: 120px;" />
            <div class="help-text">范围 10-1440 分钟。</div>
          </div>

          <div class="form-row" style="margin-top: 32px;">
            <label class="checkbox-label">
              <input type="checkbox" v-model="scheduleForm.clean_enabled" />
              启用自动清理
            </label>
          </div>
          <div class="form-row" v-if="scheduleForm.clean_enabled">
            <label>清理任务间隔 (分钟)</label>
            <input v-model.number="scheduleForm.clean_interval" type="number" min="10" max="1440" class="feishu-input" style="width: 120px;" />
          </div>

          <div class="form-row" style="margin-top: 32px;">
            <label>备份窗口（高负载时段暂停任务）</label>
            <div class="time-range">
              <div>
                <span class="time-label">开始</span>
                <input v-model.number="scheduleForm.BACKUP_START_HOUR" type="number" min="0" max="23" class="feishu-input time-input" /> :
                <input v-model.number="scheduleForm.BACKUP_START_MINUTE" type="number" min="0" max="59" class="feishu-input time-input" />
              </div>
              <span class="time-sep">—</span>
              <div>
                <span class="time-label">结束</span>
                <input v-model.number="scheduleForm.BACKUP_END_HOUR" type="number" min="0" max="23" class="feishu-input time-input" /> :
                <input v-model.number="scheduleForm.BACKUP_END_MINUTE" type="number" min="0" max="59" class="feishu-input time-input" />
              </div>
            </div>
            <div class="help-text">在此时段内不会触发自动合并/清理任务。</div>
          </div>

          <div class="form-actions">
            <button class="feishu-btn" @click="saveSchedule" :disabled="saving">{{ saving ? '保存中...' : '保存调度设置' }}</button>
            <span v-if="saveMsg" class="save-msg" :class="saveMsg.includes('失败') ? 'error' : 'success'">{{ saveMsg }}</span>
          </div>
        </div>

        <!-- 系统诊断 -->
        <div v-show="currentTab === 'health'" class="settings-section">
          <h2 class="section-title">系统诊断</h2>
          <div v-if="!health" class="help-text">加载中...</div>
          <div v-else class="diagnostics">
            <div class="diag-item">
              <span class="diag-label">ffmpeg</span>
              <span :class="health.ffmpeg_ok ? 'diag-ok' : 'diag-fail'">{{ health.ffmpeg_ok ? '✓ 正常' : '✗ 未找到' }}</span>
              <span class="diag-detail" v-if="health.ffmpeg_path">{{ health.ffmpeg_path }}</span>
            </div>
            <div class="diag-item">
              <span class="diag-label">ffprobe</span>
              <span :class="health.ffprobe_ok ? 'diag-ok' : 'diag-fail'">{{ health.ffprobe_ok ? '✓ 正常' : '✗ 未找到' }}</span>
              <span class="diag-detail" v-if="health.ffprobe_path">{{ health.ffprobe_path }}</span>
            </div>
            <div class="diag-item">
              <span class="diag-label">进程组创建</span>
              <span :class="health.ffmpeg_process_group_ok ? 'diag-ok' : 'diag-fail'">{{ health.ffmpeg_process_group_ok ? '✓ 正常' : '✗ 异常' }}</span>
            </div>
            <div class="diag-item">
              <span class="diag-label">目标目录</span>
              <span :class="health.target_dir_exists ? 'diag-ok' : 'diag-fail'">{{ health.target_dir_exists ? '✓ 存在' : '✗ 不存在' }}</span>
              <span class="diag-detail">{{ health.target_dir }}</span>
            </div>
            <div class="diag-item" v-if="health.target_dir_exists">
              <span class="diag-label">目录可写</span>
              <span :class="health.target_dir_writable ? 'diag-ok' : 'diag-fail'">{{ health.target_dir_writable ? '✓ 可写' : '✗ 不可写' }}</span>
            </div>
            <div class="diag-item" v-if="health.target_dir_exists">
              <span class="diag-label">主播数 / 文件数</span>
              <span class="diag-value">{{ health.streamer_count }} / {{ health.video_count }}</span>
            </div>
            <div class="diag-item" v-if="health.target_dir_exists">
              <span class="diag-label">总大小</span>
              <span class="diag-value">{{ health.total_size_gb?.toFixed(2) }} GB</span>
            </div>
            <div class="diag-item">
              <span class="diag-label">磁盘总量</span>
              <span class="diag-value">{{ health.disk_total_gb?.toFixed(1) }} GB ({{ health.disk_usage_pct?.toFixed(1) }}%)</span>
            </div>
            <div class="diag-item">
              <span class="diag-label">磁盘剩余</span>
              <span class="diag-value">{{ health.disk_free_gb?.toFixed(1) }} GB</span>
            </div>
          </div>
          <button class="feishu-btn" style="margin-top: 20px;" @click="loadHealth">重新检测</button>
        </div>

        <!-- AI 推荐 -->
        <div v-show="currentTab === 'recommend'" class="settings-section">
          <h2 class="section-title">AI 配置推荐</h2>
          <div v-if="!recommend" class="help-text">正在分析系统状态...</div>
          <div v-else>
            <div class="alert-box" :class="{
              warning: recommend.risk_level === 'high' || recommend.risk_level === 'critical',
              success: recommend.risk_level === 'low' || recommend.risk_level === 'normal'
            }">
              <span>{{ recommend.risk_level === 'critical' ? '🔴' : recommend.risk_level === 'high' ? '🟠' : recommend.risk_level === 'normal' ? '🟡' : '🟢' }}</span>
              <div>
                <strong>风险等级: {{ recommend.risk_level }}</strong>
                <div class="help-text" style="margin-top: 4px;">{{ recommend.reason }}</div>
              </div>
            </div>

            <div class="diag-grid" v-if="recommend.analysis">
              <div class="diag-card">
                <div class="diag-card-value">{{ recommend.analysis.streamer_count }}</div>
                <div class="diag-card-label">主播数</div>
              </div>
              <div class="diag-card">
                <div class="diag-card-value">{{ recommend.analysis.total_videos }}</div>
                <div class="diag-card-label">视频文件</div>
              </div>
              <div class="diag-card">
                <div class="diag-card-value">{{ recommend.analysis.daily_output_gb?.toFixed(1) }} GB</div>
                <div class="diag-card-label">日均产出</div>
              </div>
              <div class="diag-card">
                <div class="diag-card-value">{{ recommend.analysis.days_until_full?.toFixed(0) }} 天</div>
                <div class="diag-card-label">预计满盘</div>
              </div>
            </div>

            <h3 style="font-size: 15px; margin: 24px 0 12px; font-weight: 600;">推荐参数</h3>
            <table class="recommend-table">
              <thead>
                <tr><th>参数</th><th>当前值</th><th>推荐值</th></tr>
              </thead>
              <tbody>
                <tr><td>清理触发阈值</td><td>{{ config.TRIGGER_THRESHOLD }}%</td><td>{{ recommend.TRIGGER_THRESHOLD }}%</td></tr>
                <tr><td>清理目标水位</td><td>{{ config.TARGET_THRESHOLD }}%</td><td>{{ recommend.TARGET_THRESHOLD }}%</td></tr>
                <tr><td>每主播最少保留</td><td>{{ config.MIN_KEEP_PER_STREAMER }}</td><td>{{ recommend.MIN_KEEP_PER_STREAMER }}</td></tr>
                <tr><td>安全保护时长</td><td>{{ config.SAFE_AGE_MINUTES }} 分钟</td><td>{{ recommend.SAFE_AGE_MINUTES }} 分钟</td></tr>
                <tr><td>合并间隔判定</td><td>{{ config.MERGE_AGE_MINUTES }} 分钟</td><td>{{ recommend.MERGE_AGE_MINUTES }} 分钟</td></tr>
                <tr><td>单次最大删除</td><td>{{ config.MAX_DELETE_PER_RUN }}</td><td>{{ recommend.MAX_DELETE_PER_RUN }}</td></tr>
              </tbody>
            </table>

            <button class="feishu-btn" style="margin-top: 16px;" @click="applyRecommend">应用推荐配置</button>
            <button class="feishu-btn feishu-btn-outline" style="margin-left: 12px;" @click="loadRecommend">重新分析</button>
          </div>
        </div>

        <!-- 配置备份 -->
        <div v-show="currentTab === 'backup'" class="settings-section">
          <h2 class="section-title">配置备份与恢复</h2>

          <div class="backup-section">
            <h3 style="font-size: 15px; font-weight: 600; margin-bottom: 8px;">导出配置</h3>
            <p class="help-text" style="margin-bottom: 12px;">导出当前系统配置、调度设置和最近操作记录。</p>
            <div style="display: flex; gap: 12px;">
              <button class="feishu-btn" @click="doExport">获取导出数据</button>
              <button v-if="exportData" class="feishu-btn feishu-btn-outline" @click="downloadExport">下载 JSON 文件</button>
            </div>
            <textarea v-if="exportData" readonly class="feishu-input json-editor" :value="JSON.stringify(exportData, null, 2)" />
          </div>

          <div class="backup-section" style="margin-top: 32px;">
            <h3 style="font-size: 15px; font-weight: 600; margin-bottom: 8px;">导入配置</h3>
            <p class="help-text" style="margin-bottom: 12px;">粘贴之前导出的 JSON 配置数据来恢复系统设置。</p>
            <textarea v-model="importText" class="feishu-input json-editor" placeholder='粘贴 JSON 配置数据...' />
            <div style="margin-top: 12px; display: flex; align-items: center; gap: 12px;">
              <button class="feishu-btn" @click="doImport" :disabled="!importText.trim()">导入</button>
              <span v-if="importMsg" :class="importMsg.includes('失败') ? 'text-[#f54a45]' : 'text-[#00b578]'" class="text-sm">{{ importMsg }}</span>
            </div>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<style scoped>
.settings-page { overflow: hidden; }
.settings-layout { display: flex; height: calc(100vh - 140px); min-height: 500px; background: white; border: 1px solid var(--border-color); border-radius: 12px; }

.settings-sidebar { width: 220px; background: #fafbfc; border-right: 1px solid var(--border-color); padding: 24px 0; flex-shrink: 0; }
.nav-header { margin: 0 0 16px 24px; font-size: 12px; color: var(--text-placeholder); font-weight: 500; text-transform: uppercase; letter-spacing: 0.5px; }
.settings-nav { list-style: none; padding: 0; margin: 0; }
.settings-nav li { padding: 11px 24px; font-size: 14px; color: var(--text-regular); cursor: pointer; border-left: 3px solid transparent; transition: all 0.2s; }
.settings-nav li:hover { background: #f0f1f2; color: var(--text-title); }
.settings-nav li.active { background: var(--color-primary-bg); color: var(--color-primary); border-left-color: var(--color-primary); font-weight: 500; }

.settings-content { flex: 1; padding: 32px 40px; overflow-y: auto; }
.section-title { margin: 0 0 24px 0; font-size: 20px; font-weight: 600; color: var(--text-title); }

.form-row { margin-bottom: 20px; max-width: 600px; }
.form-row label { display: block; font-weight: 500; margin-bottom: 6px; color: var(--text-title); font-size: 14px; }
.help-text { font-size: 12px; color: var(--text-placeholder); margin-top: 4px; }

.checkbox-label { display: flex; align-items: center; gap: 8px; cursor: pointer; font-weight: 400 !important; }
.slider-group { display: flex; align-items: center; gap: 16px; }
.feishu-slider { flex: 1; accent-color: var(--color-primary); }
.slider-value { font-weight: 600; color: var(--color-primary); width: 40px; }

.alert-box { padding: 12px 16px; border-radius: var(--radius-md); font-size: 14px; margin-bottom: 24px; display: flex; align-items: flex-start; gap: 10px; }
.alert-box.warning { background: #fff2e8; border: 1px solid #ffd591; color: #d46b08; }
.alert-box.success { background: #e8f8f0; border: 1px solid #b7eb8f; color: #389e0d; }

.form-actions { margin-top: 28px; padding-top: 20px; border-top: 1px solid var(--border-color); display: flex; align-items: center; gap: 12px; }
.save-msg { font-size: 13px; }
.save-msg.success { color: #00b578; }
.save-msg.error { color: #f54a45; }

.time-range { display: flex; align-items: center; gap: 12px; }
.time-label { font-size: 12px; color: var(--text-placeholder); margin-right: 6px; }
.time-input { width: 60px; text-align: center; }
.time-sep { color: var(--text-placeholder); }

.diagnostics { display: flex; flex-direction: column; gap: 1px; background: var(--border-color); border: 1px solid var(--border-color); border-radius: var(--radius-md); overflow: hidden; }
.diag-item { display: flex; align-items: center; padding: 12px 16px; background: white; gap: 12px; }
.diag-label { font-weight: 500; font-size: 14px; color: var(--text-title); min-width: 140px; }
.diag-ok { color: #00b578; font-weight: 500; font-size: 13px; }
.diag-fail { color: #f54a45; font-weight: 500; font-size: 13px; }
.diag-value { font-size: 13px; color: var(--text-regular); }
.diag-detail { font-size: 12px; color: var(--text-placeholder); margin-left: auto; }

.diag-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-bottom: 16px; }
.diag-card { background: #f8f9fa; border-radius: var(--radius-md); padding: 16px; text-align: center; }
.diag-card-value { font-size: 20px; font-weight: 700; color: var(--text-title); }
.diag-card-label { font-size: 12px; color: var(--text-placeholder); margin-top: 4px; }

.recommend-table { width: 100%; border-collapse: collapse; font-size: 14px; max-width: 500px; }
.recommend-table th { text-align: left; padding: 8px 12px; background: #f8f9fa; color: var(--text-regular); font-weight: 500; border-bottom: 1px solid var(--border-color); }
.recommend-table td { padding: 8px 12px; border-bottom: 1px solid #f0f1f5; color: var(--text-title); }

.backup-section { max-width: 700px; }
.json-editor { font-family: "SFMono-Regular", Consolas, monospace; height: 250px; resize: vertical; background: #fafbfc; line-height: 1.5; font-size: 13px; margin-top: 8px; }
</style>

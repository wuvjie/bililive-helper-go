<template>
  <div class="settings">
    <el-card shadow="hover">
      <template #header><span>⚙️ 全局设置</span></template>
      <el-tabs v-model="activeTab" tab-position="left" class="settings-tabs">

        <!-- Tab 1: General Config -->
        <el-tab-pane label="基本配置" name="general">
          <el-form label-width="144px" label-position="right" v-loading="configLoading" class="settings-form">
            <el-form-item label="录制目录">
              <el-input v-model="config.TARGET_DIR" placeholder="/path/to/recordings" class="mono-input" style="max-width: 520px" />
            </el-form-item>
            <el-form-item label="合并保护期(分钟)" :class="{ 'field-highlight': highlightFields.includes('MERGE_AGE_MINUTES') }">
              <el-input-number v-model="config.MERGE_AGE_MINUTES" :min="5" :max="1440" class="compact-counter" />
            </el-form-item>
            <el-form-item label="片段间隔(分钟)" :class="{ 'field-highlight': highlightFields.includes('GAP_MINUTES') }">
              <el-input-number v-model="config.GAP_MINUTES" :min="1" :max="120" class="compact-counter" />
            </el-form-item>
            <el-form-item label="安全模式" :class="{ 'field-highlight': highlightFields.includes('SAFE_MODE') }">
              <el-select v-model="config.SAFE_MODE" class="compact-select" @change="onSafeModeChange">
                <el-option label="按小时" value="hours" />
                <el-option label="按天" value="days" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="config.SAFE_MODE === 'hours'" label="清理保护期(分钟)" :class="{ 'field-highlight': highlightFields.includes('SAFE_AGE_MINUTES') }">
              <el-input-number v-model="config.SAFE_AGE_MINUTES" :min="10" :max="1440" class="compact-counter" />
            </el-form-item>
            <el-form-item v-if="config.SAFE_MODE === 'days'" label="清理保护期(天)">
              <el-input-number v-model="config.SAFE_DAYS" :min="1" :max="365" class="compact-counter" />
            </el-form-item>
            <el-form-item label="单次最大删除数" :class="{ 'field-highlight': highlightFields.includes('MAX_DELETE_PER_RUN') }">
              <el-input-number v-model="config.MAX_DELETE_PER_RUN" :min="1" :max="1000" class="compact-counter" />
            </el-form-item>
            <el-form-item label="白名单关键词">
              <el-input v-model="config.WHITELIST_KEYWORDS" placeholder="关键词用英文逗号分隔" class="mono-input" style="max-width: 520px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="savingConfig" style="width: 128px" @click="handleSaveConfig">保存配置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- Tab 2: Storage -->
        <el-tab-pane label="存储管理" name="storage">
          <el-form label-width="144px" label-position="left" class="settings-form">
            <el-form-item label="触发清理阈值">
              <div class="slider-row">
                <el-slider v-model="config.TRIGGER_THRESHOLD" :min="50" :max="99" :format-tooltip="(v: number) => v + '%'" class="slider-flex" @change="onTriggerChange" />
                <span class="slider-value">{{ config.TRIGGER_THRESHOLD }}%</span>
              </div>
            </el-form-item>
            <el-form-item label="目标清理阈值">
              <div class="slider-row">
                <el-slider v-model="config.TARGET_THRESHOLD" :min="30" :max="89" :format-tooltip="(v: number) => v + '%'" class="slider-flex" @change="onTargetChange" />
                <span class="slider-value">{{ config.TARGET_THRESHOLD }}%</span>
              </div>
            </el-form-item>
            <el-form-item label="每主播最少保留">
              <el-input-number v-model="config.MIN_KEEP_PER_STREAMER" :min="1" :max="50" />
            </el-form-item>
            <el-form-item label="清理预估">
              <span v-if="cleanEstimate" class="clean-estimate">
                预计清理 <span class="ctx-num">{{ cleanEstimate.file_count }}</span> 个文件，共释放 <span class="ctx-num-green">{{ cleanEstimate.total_size_gb?.toFixed(2) }} GB</span>
              </span>
              <span v-else class="placeholder">加载中...</span>
            </el-form-item>
            <el-form-item>
              <div class="storage-actions">
                <el-button type="primary" :loading="savingConfig" @click="handleSaveConfig">保存配置</el-button>
                <button class="emergency-trigger" @click="emergencyDialogVisible = true">
                  <svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                  <span>紧急清理...</span>
                </button>
              </div>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- Tab 3: Schedule -->
        <el-tab-pane label="定时任务" name="schedule">
          <el-form label-width="144px" label-position="left" v-loading="scheduleLoading" class="settings-form">
            <div class="section-divider">🔄 自动合并</div>
            <el-form-item label="是否启用">
              <el-switch v-model="scheduleForm.merge_enabled" />
            </el-form-item>
            <el-form-item label="执行间隔(分钟)">
              <el-input-number v-model="scheduleForm.merge_interval" :min="10" :max="1440" />
            </el-form-item>
            <div class="section-divider">🧹 自动清理</div>
            <el-form-item label="是否启用">
              <el-switch v-model="scheduleForm.clean_enabled" />
            </el-form-item>
            <el-form-item label="执行间隔(分钟)">
              <el-input-number v-model="scheduleForm.clean_interval" :min="10" :max="1440" />
            </el-form-item>
            <div class="section-divider">⏳ 备份窗口（暂停任务）</div>
            <el-form-item label="开始时间">
              <el-time-picker v-model="backupStart" format="HH:mm" value-format="HH:mm" style="width: 160px" />
            </el-form-item>
            <el-form-item label="结束时间">
              <el-time-picker v-model="backupEnd" format="HH:mm" value-format="HH:mm" style="width: 160px" />
            </el-form-item>
            <div class="backup-hint">支持跨午夜，如 23:00 - 06:00</div>
            <el-form-item>
              <el-button type="primary" :loading="savingSchedule" style="width: 128px" @click="handleSaveSchedule">保存计划</el-button>
            </el-form-item>
            <div class="section-divider">🎮 手动触发</div>
            <el-form-item>
              <el-button :disabled="taskRunning" style="width: 148px" @click="triggerManualTask('merge')" class="ghost-trigger">
                🔄 立即执行合并
              </el-button>
              <el-button :disabled="taskRunning" style="width: 148px" @click="triggerManualTask('clean')" class="ghost-trigger">
                🧹 立即执行清理
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- Tab 4: Diagnostics -->
        <el-tab-pane label="系统诊断" name="diagnostics">
          <div class="diag-table" v-loading="diagLoading">
            <div class="diag-row">
              <span class="diag-label">FFmpeg</span>
              <span class="diag-val"><span class="status-dot-sm" :class="setupData?.ffmpeg_ok ? 'dot-ok' : 'dot-err'" />{{ setupData?.ffmpeg_ok ? '正常' : '异常' }}<span class="diag-path">{{ setupData?.ffmpeg_path }}</span></span>
            </div>
            <div class="diag-row">
              <span class="diag-label">FFprobe</span>
              <span class="diag-val"><span class="status-dot-sm" :class="setupData?.ffprobe_ok ? 'dot-ok' : 'dot-err'" />{{ setupData?.ffprobe_ok ? '正常' : '异常' }}</span>
            </div>
            <div class="diag-row">
              <span class="diag-label">目标目录存在</span>
              <span class="diag-val"><span class="status-dot-sm" :class="setupData?.target_dir_exists ? 'dot-ok' : 'dot-err'" />{{ setupData?.target_dir_exists ? '是' : '否' }}</span>
            </div>
            <div class="diag-row">
              <span class="diag-label">目录可写</span>
              <span class="diag-val"><span class="status-dot-sm" :class="setupData?.target_dir_writable ? 'dot-ok' : 'dot-err'" />{{ setupData?.target_dir_writable ? '是' : '否' }}</span>
            </div>
            <div class="diag-row">
              <span class="diag-label">主播数</span>
              <span class="mono-val">{{ setupData?.streamer_count || 0 }}</span>
            </div>
            <div class="diag-row">
              <span class="diag-label">视频数</span>
              <span class="mono-val">{{ setupData?.video_count || 0 }}</span>
            </div>
            <div class="diag-row">
              <span class="diag-label">总大小</span>
              <span class="mono-val">{{ setupData?.total_size_gb?.toFixed(2) }} GB</span>
            </div>
            <div class="diag-row">
              <span class="diag-label">磁盘总容量</span>
              <span class="mono-val">{{ setupData?.disk_total_gb?.toFixed(1) }} GB</span>
            </div>
            <div class="diag-row">
              <span class="diag-label">磁盘剩余</span>
              <span class="mono-val">{{ setupData?.disk_free_gb?.toFixed(1) }} GB</span>
            </div>
            <div class="diag-row diag-row-last">
              <span class="diag-label">磁盘使用率</span>
              <div class="diag-progress">
                <el-progress :percentage="setupData?.disk_usage_pct || 0" :stroke-width="6" :format="() => ''" style="flex:1" />
                <span class="mono-val diag-pct">{{ (setupData?.disk_usage_pct || 0).toFixed(1) }}%</span>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- Tab 5: AI Recommendation -->
        <el-tab-pane label="智能推荐" name="recommend">
          <div v-loading="recommendLoading" class="recommend-container">
            <!-- Empty state -->
            <div v-if="!recommend" class="recommend-empty">
              <span class="empty-icon">📊</span>
              <p class="empty-text">系统将根据您近 7 天的磁盘产出模型自动调校阈值</p>
              <button class="btn-primary" @click="loadRecommend">加载智能推荐值</button>
            </div>

            <!-- Loaded state -->
            <template v-else>
              <!-- Risk card -->
              <div class="risk-card" :class="'risk-' + (recommend.risk_level || 'info')">
                <span class="risk-icon">⚠️</span>
                <div class="risk-body">
                  <span class="risk-level">风险等级: {{ recommend.risk_level?.toUpperCase() }}</span>
                  <span v-if="recommend.reason" class="risk-reason">{{ recommend.reason }}</span>
                </div>
              </div>

              <!-- Summary indicators -->
              <div class="recommend-summary">
                <div class="summary-item">
                  <span class="summary-label">磁盘使用率</span>
                  <span class="mono-val">{{ recommend.current_usage?.toFixed(1) }}%</span>
                </div>
                <div class="summary-item">
                  <span class="summary-label">总容量</span>
                  <span class="mono-val">{{ recommend.total_gb?.toFixed(0) }} GB</span>
                </div>
                <div class="summary-item">
                  <span class="summary-label">剩余</span>
                  <span class="mono-val">{{ recommend.free_gb?.toFixed(1) }} GB</span>
                </div>
                <div class="summary-item">
                  <span class="summary-label">主播数</span>
                  <span class="mono-val">{{ recommend.analysis?.streamer_count || 0 }}</span>
                </div>
                <div class="summary-item">
                  <span class="summary-label">日产出</span>
                  <span class="mono-val">{{ (recommend.analysis?.daily_output_gb || 0).toFixed(1) }} GB</span>
                </div>
                <div class="summary-item">
                  <span class="summary-label">可维持</span>
                  <span class="mono-val">{{ (recommend.analysis?.days_until_full || 0).toFixed(0) }} 天</span>
                </div>
              </div>

              <!-- Comparison table -->
              <div class="recommend-table-wrap">
                <div class="recommend-table-header">
                  <span class="rt-col1">配置参数</span>
                  <span class="rt-col2">当前值</span>
                  <span class="rt-col3">推荐值</span>
                </div>
                <div v-for="(row, i) in recommendTable" :key="i" class="recommend-table-row">
                  <span class="rt-col1">{{ row.key }}</span>
                  <span class="rt-col2 mono-val">{{ row.current }}</span>
                  <span class="rt-col3"><span class="recommend-tag">{{ row.recommended }}</span></span>
                </div>
              </div>

              <div class="recommend-actions">
                <button class="btn-primary" @click="applyRecommend">一键应用推荐值</button>
              </div>
            </template>
          </div>
        </el-tab-pane>

        <!-- Tab 6: Backup -->
        <el-tab-pane label="配置备份" name="backup">
          <div class="backup-grid">
            <!-- Export card -->
            <div class="backup-card">
              <div class="backup-card-header">
                <span>📤</span><span>导出配置</span>
              </div>
              <p class="backup-desc">将当前系统的主播列表、存储阈值、定时任务等全部参数备份为 JSON 文件。</p>
              <div class="backup-spacer"></div>
              <div class="backup-card-footer">
                <button class="btn-primary btn-full" @click="handleExport">生成并导出 JSON</button>
              </div>
              <textarea v-if="exportJson" class="backup-preview" readonly :value="exportJson" />
            </div>

            <!-- Import card -->
            <div class="backup-card">
              <div class="backup-card-header">
                <span>📥</span><span>导入配置</span>
              </div>
              <p class="backup-desc">在下方粘贴合法的备份 JSON，或选择一个 JSON 文件导入，恢复将覆盖当前所有设置。</p>
              <input
                ref="importFileRef"
                type="file"
                accept=".json"
                style="display: none"
                @change="handleFileSelect"
              />
              <button class="btn-ghost-file" @click="importFileRef?.click()">📂 选择 JSON 文件</button>
              <textarea
                class="backup-textarea"
                v-model="importJson"
                placeholder="或在此粘贴配置 JSON 字符串..."
              />
              <div class="backup-card-footer">
                <button class="btn-primary btn-full" @click="handleImport">导入恢复</button>
              </div>
            </div>
          </div>
        </el-tab-pane>

      </el-tabs>
    </el-card>

    <!-- Emergency Clean Dialog -->
    <el-dialog
      v-model="emergencyDialogVisible"
      title="紧急清理"
      width="440px"
      class="emergency-dialog"
      :close-on-click-modal="!emergencySSE.isRunning.value"
      destroy-on-close
      @close="closeEmergencyDialog"
    >
      <div v-if="!emergencySSE.lines.value.length" class="emergency-form">
        <div class="emergency-title-row">
          <span class="emergency-dot"></span>
          <span class="emergency-title">紧急磁盘清理</span>
        </div>
        <p class="emergency-desc">将临时降低磁盘使用率目标阈值，强制清理到指定百分比以下。清理结束后阈值自动恢复。</p>
        <div class="emergency-input-row">
          <span class="emergency-label">目标磁盘百分比</span>
          <el-input-number v-model="emergencyTargetPct" :min="10" :max="99" :step="5" />
          <span class="emergency-unit">%</span>
        </div>
      </div>
      <pre v-else class="emergency-terminal" v-loading="emergencyLoading">{{ emergencySSE.lines.value.map(l => l.text).join('\n') }}</pre>
      <template #footer>
        <button
          v-if="!emergencySSE.lines.value.length"
          class="emergency-confirm"
          :disabled="emergencyLoading"
          @click="handleEmergencyClean"
        >
          确认紧急清理
        </button>
        <button
          v-else-if="emergencySSE.isRunning.value"
          class="emergency-confirm"
          @click="emergencySSE.abort()"
        >
          中止
        </button>
        <button
          v-else
          class="btn-primary"
          @click="closeEmergencyDialog"
        >
          关闭
        </button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onActivated } from "vue";
import { onBeforeRouteLeave } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { getConfig, saveConfig, getConfigRecommend, getConfigExport, importConfig as apiImportConfig } from "@/api/config";
import { getSchedule, saveSchedule, runTask } from "@/api/schedule";
import { setupCheck } from "@/api/setup";
import { getCleanEstimate } from "@/api/task";
import { useSSE } from "@/utils/sse";
import type { ScheduleStatus, SetupCheck, CleanEstimate, ConfigRecommend, ConfigDTO } from "@/api/types";

const activeTab = ref("general");
const savingConfig = ref(false);
const savingSchedule = ref(false);
const configLoading = ref(false);
const scheduleLoading = ref(false);
const diagLoading = ref(false);
const recommendLoading = ref(false);
const highlightFields = ref<string[]>([]);

// Using ref instead of reactive for the config object to allow full replacement from API response
const config = ref<ConfigDTO>({} as ConfigDTO);
const scheduleForm = reactive({
  merge_enabled: true,
  clean_enabled: true,
  merge_interval: 360,
  clean_interval: 720
});
const backupStart = ref("");
const backupEnd = ref("");

const setupData = ref<SetupCheck>();
const cleanEstimate = ref<CleanEstimate>();
const recommend = ref<ConfigRecommend>();
const exportJson = ref("");
const importJson = ref("");
const importFileRef = ref<HTMLInputElement | null>(null);

const taskRunning = ref(false);

// Emergency clean
const emergencyDialogVisible = ref(false);
const emergencyTargetPct = ref(70);
const emergencyLoading = ref(false);
const emergencySSE = useSSE();

// --- Fix 1: Unsaved changes guard ---
const isDirty = ref(false);
let suppressDirty = false; // suppress watcher during initial load / save reset

watch(config, () => { if (!suppressDirty) isDirty.value = true; }, { deep: true, flush: 'sync' });
watch(scheduleForm, () => { if (!suppressDirty) isDirty.value = true; }, { deep: true, flush: 'sync' });

onBeforeRouteLeave(async (_to, _from, next) => {
  if (!isDirty.value) { next(); return; }
  try {
    await ElMessageBox.confirm("您有未保存的配置变更，确定要离开吗？", "未保存的变更", {
      confirmButtonText: "离开",
      cancelButtonText: "取消",
      type: "warning"
    });
    next();
  } catch {
    next(false);
  }
});

const recommendTable = computed(() => {
  if (!recommend.value) return [];
  const r = recommend.value;
  const cfg = config.value;
  return [
    { key: "触发清理阈值 (%)", current: cfg.TRIGGER_THRESHOLD, recommended: r.TRIGGER_THRESHOLD },
    { key: "目标清理阈值 (%)", current: cfg.TARGET_THRESHOLD, recommended: r.TARGET_THRESHOLD },
    { key: "每主播最少保留", current: cfg.MIN_KEEP_PER_STREAMER, recommended: r.MIN_KEEP_PER_STREAMER },
    { key: "清理保护期 (分钟)", current: cfg.SAFE_AGE_MINUTES, recommended: r.SAFE_AGE_MINUTES },
    { key: "安全模式", current: cfg.SAFE_MODE, recommended: r.SAFE_MODE },
    { key: "合并保护期 (分钟)", current: cfg.MERGE_AGE_MINUTES, recommended: r.MERGE_AGE_MINUTES },
    { key: "单次最大删除数", current: cfg.MAX_DELETE_PER_RUN, recommended: r.MAX_DELETE_PER_RUN },
    { key: "片段间隔 (分钟)", current: cfg.GAP_MINUTES, recommended: r.GAP_MINUTES }
  ];
});

// Threshold cross-validation: trigger must always be > target
function onTriggerChange(val: number) {
  if (config.value.TARGET_THRESHOLD >= val) {
    config.value.TARGET_THRESHOLD = val - 1;
  }
}
function onTargetChange(val: number) {
  if (config.value.TRIGGER_THRESHOLD <= val) {
    config.value.TRIGGER_THRESHOLD = val + 1;
  }
}

// --- Fix 2: Auto-convert protection period when switching safe mode ---
function onSafeModeChange(mode: string) {
  if (mode === "days") {
    // Convert minutes -> days (round to nearest integer, min=1)
    const minutes = config.value.SAFE_AGE_MINUTES || 60;
    config.value.SAFE_DAYS = Math.max(1, Math.round(minutes / 1440));
  } else {
    // Convert days -> minutes
    const days = config.value.SAFE_DAYS || 1;
    config.value.SAFE_AGE_MINUTES = Math.round(days * 1440);
  }
}

async function handleSaveConfig() {
  // Safety-net: re-validate before submit
  if (config.value.TRIGGER_THRESHOLD <= config.value.TARGET_THRESHOLD) {
    ElMessage.error("触发阈值必须大于目标阈值，请调整后重试");
    return;
  }
  savingConfig.value = true;
  try {
    const payload = { ...config.value };
    if (typeof payload.WHITELIST_KEYWORDS === "string") {
      payload.WHITELIST_KEYWORDS = payload.WHITELIST_KEYWORDS.split(",").map((s: string) => s.trim()).filter(Boolean);
    }
    await saveConfig(payload);
    isDirty.value = false;
    ElMessage.success("配置已保存");
  } finally {
    savingConfig.value = false;
  }
}

async function handleSaveSchedule() {
  savingSchedule.value = true;
  try {
    const data: Record<string, any> = { ...scheduleForm };
    if (backupStart.value) {
      const [h, m] = backupStart.value.split(":");
      data.BACKUP_START_HOUR = parseInt(h);
      data.BACKUP_START_MINUTE = parseInt(m);
    }
    if (backupEnd.value) {
      const [h, m] = backupEnd.value.split(":");
      data.BACKUP_END_HOUR = parseInt(h);
      data.BACKUP_END_MINUTE = parseInt(m);
    }
    await saveSchedule(data);
    isDirty.value = false;
    ElMessage.success("计划已保存");
  } finally {
    savingSchedule.value = false;
  }
}

async function loadRecommend() {
  recommendLoading.value = true;
  try {
    recommend.value = await getConfigRecommend();
  } finally {
    recommendLoading.value = false;
  }
}

async function applyRecommend() {
  if (!recommend.value) return;
  const r = recommend.value;
  suppressDirty = true;
  config.value.TRIGGER_THRESHOLD = r.TRIGGER_THRESHOLD;
  config.value.TARGET_THRESHOLD = r.TARGET_THRESHOLD;
  config.value.MIN_KEEP_PER_STREAMER = r.MIN_KEEP_PER_STREAMER;
  config.value.SAFE_AGE_MINUTES = r.SAFE_AGE_MINUTES;
  config.value.SAFE_MODE = r.SAFE_MODE;
  config.value.MERGE_AGE_MINUTES = r.MERGE_AGE_MINUTES;
  config.value.MAX_DELETE_PER_RUN = r.MAX_DELETE_PER_RUN;
  config.value.GAP_MINUTES = r.GAP_MINUTES;
  suppressDirty = false;
  try {
    await handleSaveConfig();
  } catch {
    // handleSaveConfig already showed the error (e.g. threshold validation);
    // leave isDirty true so the user can manually adjust and save
  }
}

async function handleExport() {
  try {
    const data = await getConfigExport();
    exportJson.value = JSON.stringify(data, null, 2);
    const blob = new Blob([exportJson.value], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `config-backup-${new Date().toISOString().slice(0, 10)}.json`;
    a.click();
    URL.revokeObjectURL(url);
    ElMessage.success("导出成功");
  } catch {
    // Error handled by interceptor
  }
}

function handleFileSelect(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = (e) => {
    const text = e.target?.result;
    if (typeof text === "string") {
      importJson.value = text;
      ElMessage.success(`已加载文件: ${file.name}`);
    }
  };
  reader.onerror = () => {
    ElMessage.error("文件读取失败");
  };
  reader.readAsText(file);
  // Reset input so the same file can be selected again
  input.value = "";
}

async function handleImport() {
  if (!importJson.value.trim()) {
    ElMessage.warning("请粘贴配置内容");
    return;
  }
  try {
    const data = JSON.parse(importJson.value);
    try {
      await ElMessageBox.confirm(
        "导入将覆盖当前所有配置，是否继续？",
        "确认导入配置",
        { confirmButtonText: "导入", cancelButtonText: "取消", type: "warning" }
      );
    } catch {
      return;
    }
    await apiImportConfig(data);
    ElMessage.success("导入成功");
    suppressDirty = true;
    const c = await getConfig();
    config.value = c;
    if (Array.isArray(c.WHITELIST_KEYWORDS)) {
      (config.value as Record<string, unknown>).WHITELIST_KEYWORDS = c.WHITELIST_KEYWORDS.join(", ");
    }
    suppressDirty = false;
    isDirty.value = false;
  } catch (e: any) {
    suppressDirty = false;
    if (e instanceof SyntaxError) {
      ElMessage.error("JSON 格式错误");
    }
    // Other errors handled by HTTP interceptor
  }
}

async function triggerManualTask(task: "merge" | "clean") {
  const label = task === "merge" ? "合并" : "清理";
  try {
    await ElMessageBox.confirm(
      `即将手动触发${label}任务，是否继续？`,
      `确认执行${label}`,
      { confirmButtonText: "执行", cancelButtonText: "取消", type: "warning" }
    );
  } catch {
    return;
  }
  taskRunning.value = true;
  try {
    await runTask(task);
    ElMessage.success(`${label}任务已触发`);
  } catch {
    // Error handled by interceptor
  } finally {
    taskRunning.value = false;
  }
}

async function handleEmergencyClean() {
  emergencySSE.clear();
  emergencyLoading.value = true;
  try {
    await emergencySSE.startSSE("/api/clean/emergency", {
      target_pct: emergencyTargetPct.value,
      confirm: true
    });
    const lastLine = emergencySSE.lines.value.at(-1);
    if (lastLine && /✅/.test(lastLine.text)) {
      ElMessage.success("紧急清理完成");
    }
  } catch {
    // Error handled by SSE
  } finally {
    emergencyLoading.value = false;
  }
}

function closeEmergencyDialog() {
  emergencySSE.abort();
  emergencyDialogVisible.value = false;
}

onMounted(async () => {
  suppressDirty = true;
  const [c, s, d, ce] = await Promise.allSettled([
    getConfig(),
    getSchedule(),
    setupCheck(),
    getCleanEstimate()
  ]);
  if (c.status === "fulfilled") {
    config.value = c.value;
    if (Array.isArray(c.value.WHITELIST_KEYWORDS)) {
      (config.value as Record<string, unknown>).WHITELIST_KEYWORDS = c.value.WHITELIST_KEYWORDS.join(", ");
    }
    const cv = c.value;
    if (cv.BACKUP_START_HOUR != null) {
      backupStart.value = `${String(cv.BACKUP_START_HOUR).padStart(2, "0")}:${String(cv.BACKUP_START_MINUTE || 0).padStart(2, "0")}`;
    }
    if (cv.BACKUP_END_HOUR != null) {
      backupEnd.value = `${String(cv.BACKUP_END_HOUR).padStart(2, "0")}:${String(cv.BACKUP_END_MINUTE || 0).padStart(2, "0")}`;
    }
  }
  if (s.status === "fulfilled") {
    const sv = s.value;
    scheduleForm.merge_enabled = sv.merge_enabled;
    scheduleForm.clean_enabled = sv.clean_enabled;
    scheduleForm.merge_interval = sv.merge_interval;
    scheduleForm.clean_interval = sv.clean_interval;
  }
  if (d.status === "fulfilled") setupData.value = d.value;
  if (ce.status === "fulfilled") cleanEstimate.value = ce.value;
  suppressDirty = false;
  isDirty.value = false;
});

// Refresh data when component is re-activated by keep-alive router-view
onActivated(async () => {
  suppressDirty = true;
  const [c, s, d, ce] = await Promise.allSettled([
    getConfig(),
    getSchedule(),
    setupCheck(),
    getCleanEstimate()
  ]);
  if (c.status === "fulfilled") {
    config.value = c.value;
    if (Array.isArray(c.value.WHITELIST_KEYWORDS)) {
      (config.value as Record<string, unknown>).WHITELIST_KEYWORDS = c.value.WHITELIST_KEYWORDS.join(", ");
    }
  }
  if (s.status === "fulfilled") {
    const sv = s.value;
    scheduleForm.merge_enabled = sv.merge_enabled;
    scheduleForm.clean_enabled = sv.clean_enabled;
    scheduleForm.merge_interval = sv.merge_interval;
    scheduleForm.clean_interval = sv.clean_interval;
  }
  if (d.status === "fulfilled") setupData.value = d.value;
  if (ce.status === "fulfilled") cleanEstimate.value = ce.value;
  suppressDirty = false;
  isDirty.value = false;
});
</script>

<style scoped>
.settings-tabs {
  min-height: 500px;
}
.placeholder { color: var(--stone); }

/* Form label styling — left-aligned, fixed width, muted color */
.settings-form :deep(.el-form-item__label) {
  color: var(--steel) !important;
  font-size: 13px !important;
  font-weight: 500 !important;
  line-height: 32px !important;
}
.settings-form :deep(.el-form-item) {
  margin-bottom: 18px;
}
/* Input Number — mono digits */
.settings-form :deep(.el-input__inner) {
  font-family: var(--font-mono);
}

/* Field highlight animation for recommended values */
.field-highlight {
  animation: highlight-pulse 3s ease-out forwards;
}
@keyframes highlight-pulse {
  0% { background: #fdf6e3; border-radius: 6px; }
  100% { background: transparent; }
}

/* Slider row — constrained to 520px matrix, vertically centered */
.slider-row { max-width: 520px; width: 100%; display: flex; align-items: center; gap: 12px; transform: translateY(-1px); }
.slider-flex { flex: 1; }
.slider-value { font-family: var(--font-mono); font-size: 13px; font-weight: 500; color: var(--ink); min-width: 40px; text-align: right; }

/* Clean estimate — mono numbers, muted */
.clean-estimate { font-size: 13px; color: var(--slate); }
.clean-estimate .ctx-num {
  font-family: var(--font-mono); font-weight: 500; color: var(--charcoal);
}
.clean-estimate .ctx-num-green {
  font-family: var(--font-mono); font-weight: 500; color: #448361;
}

/* Storage actions row — save + emergency side by side */
.storage-actions { display: flex; align-items: center; gap: 12px; }
.emergency-trigger {
  background: transparent; border: none; padding: 0; cursor: pointer;
  font-size: 12px; font-weight: 500; color: #888;
  display: flex; align-items: center; gap: 4px;
  transition: color 0.15s;
}
.emergency-trigger:hover { color: #e0564c; }
.emergency-trigger svg { width: 14px; height: 14px; color: #aeaeb2; transition: color 0.15s; }
.emergency-trigger:hover svg { color: #e0564c; }

/* Section divider — replaces el-divider for cleaner Notion look */
.section-divider {
  font-size: 13px; font-weight: 600; color: var(--ink);
  padding-bottom: 8px; margin-bottom: 4px;
  border-bottom: 1px solid #f1f1ef;
  margin-top: 12px;
}
.section-divider:first-child { margin-top: 0; }

/* Diagnostics table — pure white, hairline rows */
.diag-table {
  border: 1px solid var(--hairline); border-radius: var(--r-md);
  overflow: hidden; max-width: 640px;
}
.diag-row {
  display: flex; justify-content: space-between; align-items: center;
  padding: 10px 16px;
  border-bottom: 1px solid #f1f1ef;
}
.diag-row-last { border-bottom: none; }
.diag-label { font-size: 13px; color: var(--steel); width: 110px; flex-shrink: 0; white-space: nowrap; }
.diag-val { display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--charcoal); padding-right: 1px; }
.diag-path { font-size: 11px; color: var(--stone); margin-left: 8px; font-family: var(--font-mono); }
.mono-val { font-family: var(--font-mono); font-size: 13px; color: var(--ink); }
.status-dot-sm { width: 6px; height: 6px; border-radius: 50%; display: inline-block; flex-shrink: 0; }
.dot-ok { background: #448361; }
.dot-err { background: #e03131; }
.diag-progress { display: flex; align-items: center; gap: 12px; flex: 1; max-width: 280px; }
.diag-pct { font-size: 12px; color: #448361; font-weight: 500; min-width: 44px; text-align: right; line-height: 1; }

/* Recommend tab — empty state & loaded state */
.recommend-container { display: flex; flex-direction: column; gap: 16px; }
.recommend-empty {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 12px; padding: 80px 0; text-align: center;
}
.empty-icon { font-size: 28px; margin-bottom: 4px; }
.empty-text { font-size: 13px; color: var(--steel); max-width: 320px; }

/* Risk card */
.risk-card {
  display: flex; align-items: flex-start; gap: 10px;
  padding: 14px 16px; border-radius: var(--r-md);
  font-size: 13px; line-height: 1.6;
}
.risk-high { background: #fdf1e6; border: 1px solid #f5e1ce; }
.risk-critical { background: #fde8e8; border: 1px solid #f5c6c6; }
.risk-low { background: #f0f7f4; border: 1px solid #d4e8dc; }
.risk-normal { background: #fdf1e6; border: 1px solid #f5e1ce; }
.risk-info { background: #f1f1ef; border: 1px solid var(--hairline); }
.risk-icon { font-size: 14px; flex-shrink: 0; margin-top: 1px; }
.risk-body { display: flex; flex-wrap: wrap; gap: 4px; color: var(--charcoal); }
.risk-level { font-weight: 600; color: var(--warning); margin-right: 8px; }
.risk-reason { color: var(--slate); }
.risk-body :deep(.ctx-num) { font-family: var(--font-mono); font-weight: 600; color: var(--ink); }

/* Summary indicators */
.recommend-summary {
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 1px;
  border: 1px solid var(--hairline); border-radius: var(--r-sm); overflow: hidden;
}
.summary-item {
  display: flex; justify-content: space-between; align-items: center;
  padding: 10px 14px; background: var(--canvas);
}
.summary-label { font-size: 13px; color: var(--steel); }

/* Comparison table — pure white, no gray patches */
.recommend-table-wrap {
  border: 1px solid var(--hairline); border-radius: var(--r-md); overflow: hidden;
}
.recommend-table-header {
  display: grid; grid-template-columns: 1fr 120px 120px;
  padding: 8px 16px; border-bottom: 1px solid var(--hairline);
  font-size: 12px; font-weight: 500; color: var(--steel); text-transform: uppercase; letter-spacing: 0.3px;
}
.recommend-table-row {
  display: grid; grid-template-columns: 1fr 120px 120px;
  padding: 10px 16px; border-bottom: 1px solid #f1f1ef;
  font-size: 13px;
}
.recommend-table-row:last-child { border-bottom: none; }
.rt-col1 { color: var(--charcoal); min-width: 0; }
.rt-col2 { text-align: right; font-family: var(--font-mono); color: var(--ink); }
.rt-col3 { text-align: right; }
.recommend-tag {
  display: inline-block; font-family: var(--font-mono); font-weight: 600;
  color: #2b593f; background: #f0f7f4;
  padding: 1px 8px; border-radius: var(--r-xs); font-size: 13px;
  min-width: 32px; text-align: right;
}

/* Button — shared with other tabs */
.btn-primary {
  height: 36px; padding: 0 20px;
  background: var(--primary); border: 1px solid var(--primary);
  border-radius: var(--r-md); font-size: 13px; font-weight: 500;
  color: #fff; cursor: pointer; transition: all 0.15s;
}
.btn-primary:hover { background: var(--primary-pressed); }

/* Emergency confirm button — red destructive action */
.emergency-confirm {
  height: 36px; padding: 0 20px;
  background: #e0564c; border: 1px solid #e0564c;
  border-radius: var(--r-md); font-size: 13px; font-weight: 500;
  color: #fff; cursor: pointer; transition: all 0.15s;
}
.emergency-confirm:hover { background: #c7463d; border-color: #c7463d; }
.emergency-confirm:disabled { opacity: 0.6; cursor: not-allowed; }

/* Backup tab — dual card split layout */
.backup-grid {
  display: grid; grid-template-columns: 1fr 1fr; gap: 16px;
  max-width: 720px;
}
.backup-card {
  background: #fafafa; border: 1px solid var(--hairline);
  border-radius: var(--r-md); padding: 20px;
  display: flex; flex-direction: column; gap: 12px;
  min-height: 240px;
}
.backup-card-header {
  display: flex; align-items: center; gap: 6px;
  font-size: 14px; font-weight: 600; color: var(--ink);
}
.backup-desc { font-size: 12px; color: var(--steel); line-height: 1.6; }
.backup-card-footer { margin-top: auto; }
.backup-textarea {
  width: 100%; height: 100px; padding: 12px;
  background: var(--canvas); border: 1px solid var(--hairline);
  border-radius: var(--r-md); resize: none;
  font-family: var(--font-mono); font-size: 12px; line-height: 1.6;
  color: var(--ink); transition: border-color 0.15s;
}
.backup-textarea:focus { outline: none; border-color: var(--ink); }
.backup-textarea::placeholder { color: var(--stone); }
.backup-preview {
  width: 100%; height: 100px; padding: 12px;
  background: var(--canvas); border: 1px solid var(--hairline);
  border-radius: var(--r-md); resize: none;
  font-family: var(--font-mono); font-size: 11px; line-height: 1.5;
  color: var(--slate);
}
.backup-spacer { flex: 1; min-height: 0; }
.backup-hint { font-size: 12px; color: var(--stone); margin: -8px 0 8px 144px; }
.btn-full { width: 100%; }
.btn-ghost-file {
  width: 100%; height: 36px;
  background: var(--canvas); border: 1px dashed var(--hairline);
  border-radius: var(--r-md); font-size: 13px; font-weight: 500;
  color: var(--steel); cursor: pointer; transition: all 0.15s;
}
.btn-ghost-file:hover {
  color: var(--ink); border-color: var(--ink); background: var(--highlight);
}

.recommend-actions { margin-top: 12px; }

@media (max-width: 768px) {
  .backup-grid { grid-template-columns: 1fr; }
}

/* Time Picker icon — thinner stroke, vertical align fix */
.settings-form :deep(.el-input__prefix .el-icon),
.settings-form :deep(.el-input__suffix .el-icon) {
  color: var(--steel);
  transform: translateY(-0.5px);
}
.ghost-trigger {
  color: var(--steel) !important;
  border: 1px solid var(--hairline) !important;
  background: transparent !important;
  border-radius: var(--r-sm) !important;
}
.ghost-trigger:hover {
  color: var(--ink) !important;
  border-color: var(--hairline-strong) !important;
  background: var(--highlight) !important;
}

/* Emergency dialog internal layout */
.emergency-title-row {
  display: flex; align-items: center; gap: 8px;
  margin-bottom: 8px;
}
.emergency-dot {
  width: 8px; height: 8px; border-radius: 50%;
  background: #e0564c; flex-shrink: 0;
}
.emergency-title {
  font-size: 14px; font-weight: 600; color: var(--ink);
}
.emergency-form { display: flex; flex-direction: column; gap: 12px; }
.emergency-desc { font-size: 13px; color: var(--slate); line-height: 1.6; margin: 0; }
.emergency-input-row {
  display: flex; align-items: center; gap: 12px;
}
.emergency-label { font-size: 13px; color: var(--charcoal); white-space: nowrap; }
.emergency-unit { font-size: 13px; color: var(--steel); }
.emergency-terminal {
  background: #191919;
  color: #b0b0b0;
  padding: 16px;
  border: 1px solid #2e2e2e;
  border-radius: var(--r-sm);
  max-height: 300px;
  overflow-y: auto;
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  min-height: 120px;
}

/* Basic config — compact counters & mono inputs */
.compact-counter,
.compact-select {
  width: 120px !important;
}
.mono-input :deep(.el-input__inner) {
  font-family: var(--font-mono, 'SF Mono', 'Fira Code', monospace);
}

/* Emergency clean dialog — compact geek overrides */
:deep(.emergency-dialog .el-dialog) {
  max-width: 420px;
  border-radius: 6px;
  overflow: hidden;
  box-shadow: 0 12px 32px rgba(0,0,0,0.06);
}
:deep(.emergency-dialog .el-dialog__header) {
  padding: 16px 20px 8px;
  border-bottom: none;
}
:deep(.emergency-dialog .el-dialog__body) {
  padding: 8px 20px 16px;
}
:deep(.emergency-dialog .el-dialog__footer) {
  padding: 12px 20px 16px;
  background: rgba(250, 250, 250, 0.4);
  border-top: 1px solid #f1f1ef;
}
:deep(.emergency-dialog .el-dialog__footer .el-button) {
  height: 32px;
  padding: 0 16px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}
:deep(.emergency-dialog .el-input-number--small) {
  width: 96px;
}
:deep(.emergency-dialog .el-input-number--small .el-input__wrapper) {
  border-radius: 6px;
}
</style>

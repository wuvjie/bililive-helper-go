<template>
  <div class="settings">
    <el-card shadow="hover" class="settings-card">
      <template #header>
        <div class="settings-header">
          <span class="settings-title">⚙️ 系统设置</span>
          <span class="settings-subtitle">管理录像合并路径、磁盘清理策略与调度参数</span>
        </div>
      </template>

      <el-tabs v-model="activeTab" tab-position="left" class="settings-tabs">
        <!-- Tab 1: 全局策略 (合并原基本配置与存储管理) -->
        <el-tab-pane label="全局策略" name="policy">
          <div v-loading="configLoading" class="policy-pane">
            <el-form label-width="160px" label-position="left" class="settings-form">
              <!-- Section 1: 录制与合并 -->
              <div class="form-section-card">
                <div class="section-title">
                  <span class="section-icon">📁</span>
                  <span>录制与合并规则</span>
                </div>
                <div class="section-body">

                  <el-form-item label="断流重连容差">
                    <div class="inline-unit-row">
                      <el-input-number v-model="config.GAP_MINUTES" :min="1" :max="120" class="compact-counter" />
                      <span class="unit-text">分钟</span>
                    </div>
                    <div class="field-hint">同一主播在此时间差内断流重连，将被判定为同一场直播并自动合并为整段视频。</div>
                  </el-form-item>

                  <el-form-item label="完播静置等待">
                    <div class="inline-unit-row">
                      <el-input-number v-model="config.MERGE_AGE_MINUTES" :min="5" :max="1440" class="compact-counter" />
                      <span class="unit-text">分钟</span>
                    </div>
                    <div class="field-hint">录制文件生成并静置满此时间后才允许合并，确保主播已彻底关播且文件写入完毕，防止误读活动切片。</div>
                  </el-form-item>
                </div>
              </div>

              <!-- Section 2: 存储清理策略 -->
              <div class="form-section-card">
                <div class="section-title">
                  <span class="section-icon">🧹</span>
                  <span>磁盘自动清理策略</span>
                </div>
                <div class="section-body">
                  <el-form-item label="触发清理阈值">
                    <div class="slider-row">
                      <el-slider
                        v-model="config.TRIGGER_THRESHOLD"
                        :min="50"
                        :max="99"
                        :format-tooltip="formatPercent"
                        class="slider-flex"
                        @change="onTriggerChange"
                      />
                      <span class="slider-value mono-val">{{ config.TRIGGER_THRESHOLD }}%</span>
                    </div>
                    <div class="field-hint">磁盘使用率达到此警戒水位时，定时清理任务将正式启动。</div>
                  </el-form-item>

                  <el-form-item label="目标清理阈值">
                    <div class="slider-row">
                      <el-slider
                        v-model="config.TARGET_THRESHOLD"
                        :min="30"
                        :max="89"
                        :format-tooltip="formatPercent"
                        class="slider-flex"
                        @change="onTargetChange"
                      />
                      <span class="slider-value mono-val">{{ config.TARGET_THRESHOLD }}%</span>
                    </div>
                    <div class="field-hint">清理时将从最旧录播开始删除，直至磁盘使用率降至此目标水位。</div>
                  </el-form-item>

                  <el-form-item label="新录像免删保护">
                    <div class="safe-age-group">
                      <el-select v-model="config.SAFE_MODE" class="compact-select" @change="onSafeModeChange" style="width: 92px; flex-shrink: 0;">
                        <el-option label="按小时" value="hours" />
                        <el-option label="按天" value="days" />
                      </el-select>
                      <div v-if="config.SAFE_MODE === 'hours'" class="inline-unit-row">
                        <el-input-number v-model="safeAgeHours" :min="1" :max="168" class="compact-counter" />
                        <span class="unit-text">小时</span>
                      </div>
                      <div v-if="config.SAFE_MODE === 'days'" class="inline-unit-row">
                        <el-input-number v-model="config.SAFE_DAYS" :min="1" :max="365" class="compact-counter" />
                        <span class="unit-text">天</span>
                      </div>
                    </div>
                    <div class="field-hint">在此保护期内生成的最新录像享有免删保护，即使磁盘水位达到警戒线，清理任务也绝不会删除保护期内的文件。</div>
                  </el-form-item>

                  <el-form-item label="每主播最少保留">
                    <div class="inline-unit-row">
                      <el-input-number v-model="config.MIN_KEEP_PER_STREAMER" :min="1" :max="50" class="compact-counter" />
                      <span class="unit-text">场</span>
                    </div>
                    <div class="field-hint">每个主播至少保留最近的场次数，无论磁盘空间多么紧张也不会被全删。</div>
                  </el-form-item>

                  <el-form-item label="单次最大删除数">
                    <div class="inline-unit-row">
                      <el-input-number v-model="config.MAX_DELETE_PER_RUN" :min="1" :max="1000" class="compact-counter" />
                      <span class="unit-text">个</span>
                    </div>
                    <div class="field-hint">限制单次清理任务允许删除的文件上限，避免引起磁盘长时间高负荷 I/O 抖动（默认 50 个）。</div>
                  </el-form-item>

                  <el-form-item label="白名单关键词">
                    <el-input
                      :model-value="whitelistString"
                      @update:model-value="onWhitelistUpdate"
                      placeholder="关键词用英文逗号分隔，如: 留存, 纪念, 精华"
                      class="mono-input"
                      style="max-width: 520px"
                    />
                    <div class="field-hint">包含任一关键词的视频文件名或主播目录，将永远受保护不被清理。</div>
                  </el-form-item>
                </div>
              </div>

              <!-- Form actions -->
              <div class="policy-actions">
                <el-button type="primary" :loading="savingConfig" class="btn-save-policy" @click="handleSaveConfig">
                  保存全局策略
                </el-button>
              </div>
            </el-form>
          </div>
        </el-tab-pane>

        <!-- Tab 2: 任务调度 -->
        <el-tab-pane label="定时任务" name="schedule">
          <ScheduleTab
            :form="scheduleForm"
            :backup-start="backupStart"
            :backup-end="backupEnd"
            :saving="savingSchedule"
            :loading="scheduleLoading"
            @update:form="Object.assign(scheduleForm, $event)"
            @update:backup-start="backupStart = $event"
            @update:backup-end="backupEnd = $event"
            @save="handleSaveSchedule"
          />
        </el-tab-pane>

        <!-- Tab 3: 配置备份与恢复 -->
        <el-tab-pane label="配置备份" name="backup">
          <div class="backup-grid">
            <!-- Export card -->
            <div class="backup-card">
              <div class="backup-card-header">
                <span class="backup-icon">📤</span>
                <span class="backup-title">导出配置备份</span>
              </div>
              <p class="backup-desc">将系统当前的阈值策略、定时调度参数以及各主播的个性化截取与保护规则导出为 JSON 备份文件。</p>
              <div class="backup-spacer"></div>
              <div class="backup-action-group">
                <button class="btn-action-half btn-export" @click="handleExport">📥 下载配置文件 (.json)</button>
                <button class="btn-action-half btn-copy" @click="handleCopyConfig">📋 复制配置代码</button>
              </div>
            </div>

            <!-- Import card -->
            <div class="backup-card">
              <div class="backup-card-header">
                <span class="backup-icon">📥</span>
                <span class="backup-title">导入配置恢复</span>
              </div>
              <p class="backup-desc">选择之前导出的 JSON 备份文件，导入并校验通过后将自动覆盖应用现有设置。</p>
              <input
                ref="importFileRef"
                type="file"
                accept=".json"
                style="display: none"
                @change="handleFileSelect"
              />
              <div class="import-file-box" @click="importFileRef?.click()">
                <span class="import-file-icon">📂</span>
                <span class="import-file-text">{{ selectedImportFileName || '点击选择本地 JSON 备份文件...' }}</span>
              </div>
              <div class="import-manual-toggle">
                <span class="toggle-link" @click="showManualPaste = !showManualPaste">
                  {{ showManualPaste ? '▾ 收起手动粘贴' : '▸ 或手动粘贴 JSON 代码' }}
                </span>
                <textarea
                  v-if="showManualPaste"
                  class="backup-textarea"
                  v-model="importJson"
                  placeholder="在此直接粘贴配置 JSON 字符串..."
                />
              </div>
              <div class="backup-spacer"></div>
              <div class="backup-card-footer">
                <button class="btn-primary btn-full" :disabled="!selectedImportFileName && !importJson" @click="handleImport">导入并应用配置</button>
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from "vue";
import { onBeforeRouteLeave } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { getConfig, saveConfig, getConfigExport, importConfig as apiImportConfig } from "@/api/config";
import { getSchedule, saveSchedule } from "@/api/schedule";
import type { ScheduleStatus, ConfigDTO } from "@/api/types";
import ScheduleTab from "./tabs/ScheduleTab.vue";
import type { ScheduleForm } from "./tabs/ScheduleTab.vue";

const formatPercent = (v: number | number[]) => Number(v) + "%";

const activeTab = ref("policy");
const savingConfig = ref(false);
const savingSchedule = ref(false);
const configLoading = ref(false);
const scheduleLoading = ref(false);

const config = ref<ConfigDTO>({} as ConfigDTO);

const scheduleForm = reactive<ScheduleForm>({
  merge_enabled: true,
  clean_enabled: true,
  merge_interval: 360,
  clean_interval: 720
});
const backupStart = ref("");
const backupEnd = ref("");

const exportJson = ref("");
const selectedImportFileName = ref("");
const showManualPaste = ref(false);

async function handleCopyConfig() {
  try {
    const res = await getConfigExport();
    const jsonStr = JSON.stringify(res, null, 2);
    await navigator.clipboard.writeText(jsonStr);
    ElMessage.success("已复制配置 JSON 到剪贴板");
  } catch {
    ElMessage.error("复制配置失败");
  }
}

const importJson = ref("");
const importFileRef = ref<HTMLInputElement | null>(null);

const isDirty = ref(false);
let suppressDirty = false;

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

const safeAgeHours = computed({
  get: () => {
    const mins = config.value.SAFE_AGE_MINUTES ?? 120;
    return Math.max(1, Math.round(mins / 60));
  },
  set: (val: number) => {
    config.value.SAFE_AGE_MINUTES = (val || 1) * 60;
  }
});

const whitelistString = computed(() => {
  if (!config.value.WHITELIST_KEYWORDS) return "";
  if (Array.isArray(config.value.WHITELIST_KEYWORDS)) {
    return config.value.WHITELIST_KEYWORDS.join(", ");
  }
  return String(config.value.WHITELIST_KEYWORDS);
});

function onWhitelistUpdate(val: string) {
  (config.value as Record<string, unknown>).WHITELIST_KEYWORDS = val;
}

function onSafeModeChange(val: string) {
  if (val === "hours" && !config.value.SAFE_AGE_MINUTES) config.value.SAFE_AGE_MINUTES = 120;
  if (val === "days" && !config.value.SAFE_DAYS) config.value.SAFE_DAYS = 1;
}

function onTriggerChange(val: number) {
  if (val <= config.value.TARGET_THRESHOLD) {
    config.value.TARGET_THRESHOLD = Math.max(30, val - 10);
  }
}

function onTargetChange(val: number) {
  if (val >= config.value.TRIGGER_THRESHOLD) {
    config.value.TRIGGER_THRESHOLD = Math.min(99, val + 10);
  }
}

async function loadConfigData() {
  configLoading.value = true;
  try {
    suppressDirty = true;
    const c = await getConfig();
    config.value = c;
    suppressDirty = false;
    isDirty.value = false;
  } finally {
    configLoading.value = false;
  }
}

async function loadScheduleData() {
  scheduleLoading.value = true;
  try {
    suppressDirty = true;
    const s: ScheduleStatus = await getSchedule();
    scheduleForm.merge_enabled = s.merge.enabled;
    scheduleForm.clean_enabled = s.clean.enabled;
    scheduleForm.merge_interval = s.merge.interval;
    scheduleForm.clean_interval = s.clean.interval;
    if (s.backup_start_hour !== undefined && s.backup_start_minute !== undefined) {
      backupStart.value = `${String(s.backup_start_hour).padStart(2, '0')}:${String(s.backup_start_minute).padStart(2, '0')}`;
    }
    if (s.backup_end_hour !== undefined && s.backup_end_minute !== undefined) {
      backupEnd.value = `${String(s.backup_end_hour).padStart(2, '0')}:${String(s.backup_end_minute).padStart(2, '0')}`;
    }
    suppressDirty = false;
    isDirty.value = false;
  } finally {
    scheduleLoading.value = false;
  }
}

async function handleSaveConfig() {
  if (config.value.TRIGGER_THRESHOLD <= config.value.TARGET_THRESHOLD) {
    ElMessage.error("触发阈值必须大于目标阈值");
    throw new Error("Validation failed");
  }

  savingConfig.value = true;
  try {
    const payload = { ...config.value };
    if (typeof payload.WHITELIST_KEYWORDS === "string") {
      payload.WHITELIST_KEYWORDS = (payload.WHITELIST_KEYWORDS as string)
        .split(/[,，]/)
        .map((s: string) => s.trim())
        .filter(Boolean);
    }
    await saveConfig(payload);
    ElMessage.success("全局策略已成功保存");
    suppressDirty = true;
    isDirty.value = false;
    await loadConfigData();
  } finally {
    savingConfig.value = false;
  }
}

async function handleSaveSchedule() {
  savingSchedule.value = true;
  try {
    const payload: Record<string, unknown> = {
      merge_enabled: scheduleForm.merge_enabled,
      clean_enabled: scheduleForm.clean_enabled,
      merge_interval: scheduleForm.merge_interval,
      clean_interval: scheduleForm.clean_interval,
    };
    if (backupStart.value) {
      const [h, m] = backupStart.value.split(":").map(Number);
      payload.backup_start_hour = h;
      payload.backup_start_minute = m;
    }
    if (backupEnd.value) {
      const [h, m] = backupEnd.value.split(":").map(Number);
      payload.backup_end_hour = h;
      payload.backup_end_minute = m;
    }
    await saveSchedule(payload);
    ElMessage.success("任务调度设置已保存");
    suppressDirty = true;
    isDirty.value = false;
  } finally {
    savingSchedule.value = false;
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
    ElMessage.success("配置备份已导出");
  } catch {
    // Handled by HTTP interceptor
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
      ElMessage.success(`已读取文件: ${file.name}`);
    }
  };
  reader.onerror = () => {
    ElMessage.error("文件读取失败");
  };
  reader.readAsText(file);
  input.value = "";
}

async function handleImport() {
  if (!importJson.value.trim()) {
    ElMessage.warning("请粘贴配置内容或选择文件");
    return;
  }
  try {
    const data = JSON.parse(importJson.value);
    try {
      await ElMessageBox.confirm(
        "导入将完全覆盖当前所有系统策略与主播规则，是否继续？",
        "确认导入配置",
        { confirmButtonText: "确认导入", cancelButtonText: "取消", type: "warning" }
      );
    } catch {
      return;
    }
    await apiImportConfig(data);
    ElMessage.success("配置导入恢复成功");
    await loadConfigData();
  } catch (e: any) {
    if (e instanceof SyntaxError) {
      ElMessage.error("JSON 格式错误，请检查配置内容");
    }
  }
}

onMounted(() => {
  loadConfigData();
  loadScheduleData();
});
</script>

<style scoped>
.settings {
  max-width: 1040px;
  margin: 0 auto;
}

.settings-card {
  border-radius: 12px;
  border: 1px solid var(--hairline);
  background: var(--canvas);
}

.settings-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.settings-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--charcoal);
}

.settings-subtitle {
  font-size: 12px;
  color: var(--steel);
}

.settings-tabs {
  min-height: 540px;
  padding: 12px 0;
}

.policy-pane {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 0 16px 20px 24px;
}

.form-section-card {
  background: var(--surface);
  border: 1px solid var(--hairline);
  border-radius: var(--r-md);
  padding: 18px 20px;
  margin-bottom: 20px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--charcoal);
  padding-bottom: 12px;
  border-bottom: 1px solid var(--hairline);
  margin-bottom: 16px;
}

.section-icon {
  font-size: 16px;
}

.section-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.settings-form :deep(.el-form-item__content) {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}


.field-hint {
  width: 100%;
  font-size: 12px;
  color: var(--steel);
  line-height: 1.5;
  margin-top: 6px;
}

.field-hint code {
  background: var(--surface-hover);
  padding: 1px 5px;
  border-radius: 4px;
  font-family: var(--font-mono);
  font-size: 11px;
}

.inline-unit-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  white-space: nowrap;
}

.unit-text {
  font-size: 13px;
  color: var(--slate);
  white-space: nowrap;
  flex-shrink: 0;
  display: inline-block;
}

.slider-row {
  display: flex;
  align-items: center;
  gap: 16px;
  width: 100%;
  max-width: 480px;
}

.slider-flex {
  flex: 1;
}

.slider-value {
  min-width: 44px;
  text-align: right;
  font-size: 14px;
}

.safe-age-group {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 32px;
  flex-wrap: nowrap;
}

.policy-actions {
  display: flex;
  padding-top: 8px;
}

.btn-save-policy {
  padding: 10px 32px;
  font-size: 14px;
  font-weight: 600;
  border-radius: var(--r-md);
  background: var(--primary);
  border-color: var(--primary);
}

.btn-save-policy:hover {
  background: var(--primary-pressed);
  border-color: var(--primary-pressed);
}

/* Backup tab styling */
.backup-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  padding: 8px 16px 20px 24px;
}

.backup-card {
  display: flex;
  flex-direction: column;
  background: var(--surface);
  border: 1px solid var(--hairline);
  border-radius: var(--r-md);
  padding: 20px;
}

.backup-card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.backup-icon {
  font-size: 18px;
}

.backup-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--charcoal);
}

.backup-desc {
  font-size: 13px;
  color: var(--slate);
  line-height: 1.5;
  margin-bottom: 16px;
}

.backup-spacer {
  flex: 1;
}

.btn-full {
  width: 100%;
}

.btn-ghost-file {
  background: #fff;
  border: 1px dashed var(--hairline-strong);
  color: var(--charcoal);
  border-radius: var(--r-md);
  padding: 8px 16px;
  font-size: 13px;
  cursor: pointer;
  margin-bottom: 12px;
  transition: border-color 0.15s, background 0.15s;
}

.btn-ghost-file:hover {
  border-color: var(--charcoal);
  background: var(--highlight);
}

.backup-textarea {
  width: 100%;
  height: 140px;
  padding: 10px;
  font-family: var(--font-mono);
  font-size: 12px;
  border: 1px solid var(--hairline);
  border-radius: var(--r-md);
  background: #fff;
  color: var(--ink);
  resize: vertical;
  margin-bottom: 14px;
  box-sizing: border-box;
}

.backup-textarea:focus {
  outline: none;
  border-color: var(--charcoal);
}

.backup-preview {
  width: 100%;
  height: 140px;
  padding: 10px;
  font-family: var(--font-mono);
  font-size: 12px;
  border: 1px solid var(--hairline);
  border-radius: var(--r-md);
  background: #faf9f8;
  color: var(--steel);
  resize: vertical;
  margin-top: 14px;
  box-sizing: border-box;
}



.backup-action-group {
  display: flex;
  gap: 12px;
  margin-top: 16px;
}
.btn-action-half {
  flex: 1;
  height: 38px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  border-radius: var(--r-md, 6px);
  box-sizing: border-box;
  cursor: pointer;
  transition: all 0.15s ease;
}
.btn-action-half.btn-export {
  background: var(--primary);
  color: var(--on-primary);
  border: 1px solid var(--primary);
}
.btn-action-half.btn-export:hover {
  background: var(--primary-pressed);
  border-color: var(--primary-pressed);
}
.btn-action-half.btn-copy {
  background: var(--surface, #fafafa);
  border: 1px solid var(--hairline, #dcdfe6);
  color: var(--charcoal, #333);
}
.btn-action-half.btn-copy:hover {
  background: var(--highlight, #f0f0f0);
  border-color: var(--steel, #aaa);
}
.import-file-box {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--canvas, #fff);
  border: 1px dashed var(--hairline, #dcdfe6);
  padding: 12px 16px;
  border-radius: var(--r-sm, 4px);
  cursor: pointer;
  margin: 12px 0 8px;
  transition: border-color 0.15s;
}
.import-file-box:hover {
  border-color: #409eff;
}
.import-file-icon {
  font-size: 18px;
}
.import-file-text {
  font-size: 13px;
  color: var(--slate, #606266);
}
.import-manual-toggle {
  margin: 6px 0 12px;
}
.toggle-link {
  font-size: 12px;
  color: var(--steel, #888);
  cursor: pointer;
  user-select: none;
}
.toggle-link:hover {
  color: var(--ink, #333);
}

@media (max-width: 680px) {
  .policy-pane {
    padding: 0 4px 16px 8px;
  }
  .settings-form :deep(.el-form-item) {
    flex-direction: column;
    align-items: flex-start;
  }
  .settings-form :deep(.el-form-item__label) {
    width: 100% !important;
    margin-bottom: 6px;
    height: auto;
    line-height: 1.4;
  }
  .settings-form :deep(.el-form-item__content) {
    width: 100%;
  }
  .safe-age-group {
    width: 100%;
    overflow-x: auto;
  }
  .backup-grid {
    grid-template-columns: 1fr;
    padding: 8px 8px 16px 8px;
  }
}
</style>

<template>
  <div class="settings">
    <el-card shadow="hover" class="settings-card">
      <template #header>
        <div class="settings-header">
          <div class="header-left">
            <div class="title-group-flex">
              <div class="ui-icon-box icon-box-purple">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
              </div>
              <span class="settings-title">系统设置</span>
            </div>
            <span class="settings-subtitle">管理自动化录制合并、磁盘巡检保护、计划调度与系统凭据</span>
          </div>
          <span v-if="isFormDirty()" class="unsaved-badge">
            <span class="badge-dot"></span> 未保存的配置更改
          </span>
        </div>
      </template>

      <el-tabs v-model="activeTab" :tab-position="tabPosition" class="settings-tabs">
        <!-- Tab 1: 全局策略 -->
        <el-tab-pane name="policy">
          <template #label>
            <span class="settings-tab-label">
              <svg class="tab-svg-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="4" y1="21" x2="4" y2="14"/><line x1="4" y1="10" x2="4" y2="3"/><line x1="12" y1="21" x2="12" y2="12"/><line x1="12" y1="8" x2="12" y2="3"/><line x1="20" y1="21" x2="20" y2="16"/><line x1="20" y1="12" x2="20" y2="3"/><line x1="1" y1="14" x2="7" y2="14"/><line x1="9" y1="8" x2="15" y2="8"/><line x1="17" y1="16" x2="23" y2="16"/></svg>
              <span>全局策略</span>
            </span>
          </template>
          <div v-loading="configLoading" class="tab-pane-wrap">
            <PolicyTab
              :config="config"
              :saving="savingConfig"
              :total-disk-gb="totalDiskGb"
              @update:config="config = $event"
              @reset-defaults="handleResetDefaults"
            />
          </div>
        </el-tab-pane>

        <!-- Tab 2: 任务调度 -->
        <el-tab-pane name="schedule">
          <template #label>
            <span class="settings-tab-label">
              <svg class="tab-svg-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
              <span>定时调度</span>
            </span>
          </template>
          <div v-loading="scheduleLoading" class="tab-pane-wrap">
            <ScheduleTab
              :form="scheduleForm"
              :backup-start="backupStart"
              :backup-end="backupEnd"
              :saving="savingSchedule"
              :loading="scheduleLoading"
              @update:form="Object.assign(scheduleForm, $event)"
              @update:backup-start="backupStart = $event"
              @update:backup-end="backupEnd = $event"
              />
          </div>
        </el-tab-pane>

        <!-- Tab 3: 配置备份与恢复 -->
        <el-tab-pane name="backup">
          <template #label>
            <span class="settings-tab-label">
              <svg class="tab-svg-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
              <span>配置备份</span>
            </span>
          </template>
          <div class="tab-pane-wrap">
            <BackupTab
              :selected-file-name="selectedImportFileName"
              :import-json="importJson"
              @export-file="handleExport"
              @copy-json="handleCopyConfig"
              @file-selected="handleFileSelect"
              @update:import-json="importJson = $event"
              @apply-import="handleImport"
            />
          </div>
        </el-tab-pane>

        <!-- Tab 4: 安全与密码 -->
        <el-tab-pane name="security">
          <template #label>
            <span class="settings-tab-label">
              <svg class="tab-svg-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
              <span>安全设置</span>
            </span>
          </template>
          <div class="tab-pane-wrap">
            <SecurityTab />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- Floating Sticky Save Bar (吸底浮动保存栏) -->
    <transition name="save-bar-slide">
      <div v-if="isFormDirty()" class="sticky-save-bar">
        <div class="save-bar-inner">
          <div class="save-bar-left">
            <span class="save-bar-pulse-dot"></span>
            <span class="save-bar-title">检测到您修改了配置参数</span>
            <span class="save-bar-sub">离开前请记得保存生效</span>
          </div>
          <div class="save-bar-actions">
            <button class="btn-bar-ghost" @click="handleDiscardChanges">放弃变更</button>
            <button
              class="btn-bar-primary"
              :disabled="savingConfig || savingSchedule"
              @click="handleSaveSettings"
            >
              {{ (savingConfig || savingSchedule) ? '保存中...' : '立即保存配置' }}
            </button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onUnmounted } from "vue";
import { useRoute, onBeforeRouteLeave } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { getConfig, saveConfig, getConfigExport, importConfig as apiImportConfig } from "@/api/config";
import { getSchedule, saveSchedule } from "@/api/schedule";
import { getStatusDetail } from "@/api/status";
import { copyToClipboard } from "@/utils/clipboard";
import type { ScheduleStatus, ConfigDTO } from "@/api/types";

import PolicyTab from "./tabs/PolicyTab.vue";
import ScheduleTab from "./tabs/ScheduleTab.vue";
import type { ScheduleForm } from "./tabs/ScheduleTab.vue";
import BackupTab from "./tabs/BackupTab.vue";
import SecurityTab from "./tabs/SecurityTab.vue";

const route = useRoute();
const activeTab = ref(typeof route.query.tab === "string" ? route.query.tab : "policy");
const savingConfig = ref(false);
const savingSchedule = ref(false);
const configLoading = ref(false);
const scheduleLoading = ref(false);

const config = ref<ConfigDTO>({} as ConfigDTO);
const totalDiskGb = ref<number>(0);

const scheduleForm = reactive<ScheduleForm>({
  merge_enabled: true,
  clean_enabled: true,
  merge_interval: 360,
  clean_interval: 720
});
const backupStart = ref("");
const backupEnd = ref("");

const selectedImportFileName = ref("");
const importJson = ref("");

// Responsive tab position (Desktop: left, Mobile/Vertical: top)
const isMobile = ref(typeof window !== "undefined" ? window.innerWidth < 768 : false);
function handleResize() {
  isMobile.value = window.innerWidth < 768;
}
const tabPosition = computed(() => (isMobile.value ? "top" : "left"));

// Baseline Snapshot System (数学级快照比对，杜绝误报)
const initialPolicySnapshot = ref("");
const initialScheduleSnapshot = ref("");

function normalizeWhitelist(wl: unknown): string {
  if (!wl) return "";
  if (Array.isArray(wl)) {
    return wl.map((s: string) => String(s).trim()).filter(Boolean).sort().join(",");
  }
  if (typeof wl === "string") {
    return wl.split(/[,，]/).map((s: string) => s.trim()).filter(Boolean).sort().join(",");
  }
  return String(wl);
}

function serializePolicyConfig(c: ConfigDTO): string {
  if (!c || Object.keys(c).length === 0) return "";
  return JSON.stringify({
    gap: Number(c.GAP_MINUTES || 30),
    mergeAge: Number(c.MERGE_AGE_MINUTES || 30),
    trigger: Number(c.TRIGGER_THRESHOLD || 80),
    target: Number(c.TARGET_THRESHOLD || 60),
    safeMode: c.SAFE_MODE || "hours",
    safeAgeMinutes: Number(c.SAFE_AGE_MINUTES || 120),
    safeDays: Number(c.SAFE_DAYS || 1),
    minKeep: Number(c.MIN_KEEP_PER_STREAMER || 3),
    maxDelete: Number(c.MAX_DELETE_PER_RUN || 50),
    whitelist: normalizeWhitelist(c.WHITELIST_KEYWORDS)
  });
}

function serializeScheduleConfig(form: ScheduleForm, start: string, end: string): string {
  return JSON.stringify({
    merge_enabled: Boolean(form.merge_enabled),
    clean_enabled: Boolean(form.clean_enabled),
    merge_interval: Number(form.merge_interval || 360),
    clean_interval: Number(form.clean_interval || 720),
    start: start || "",
    end: end || ""
  });
}

function isFormDirty(): boolean {
  if (initialPolicySnapshot.value && serializePolicyConfig(config.value) !== initialPolicySnapshot.value) {
    return true;
  }
  if (initialScheduleSnapshot.value && serializeScheduleConfig(scheduleForm, backupStart.value, backupEnd.value) !== initialScheduleSnapshot.value) {
    return true;
  }
  return false;
}

onBeforeRouteLeave(async (_to, _from, next) => {
  if (!isFormDirty()) { next(); return; }
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

// Loaders
async function loadConfigData() {
  configLoading.value = true;
  try {
    const c = await getConfig();
    // 兼容历史配置：若从旧环境加载出的单次删除上限仍为旧版默认的 10，自动升级为标准推荐值 50
    if (c.MAX_DELETE_PER_RUN === 10 || !c.MAX_DELETE_PER_RUN) {
      c.MAX_DELETE_PER_RUN = 50;
    }
    config.value = c;
    initialPolicySnapshot.value = serializePolicyConfig(c);
  } finally {
    configLoading.value = false;
  }
}

async function loadScheduleData() {
  scheduleLoading.value = true;
  try {
    const s: ScheduleStatus = await getSchedule();
    scheduleForm.merge_enabled = s.merge.enabled;
    scheduleForm.clean_enabled = s.clean.enabled;
    scheduleForm.merge_interval = s.merge.interval;
    scheduleForm.clean_interval = s.clean.interval;
    let start = "";
    let end = "";
    if (s.backup_start_hour !== undefined && s.backup_start_minute !== undefined) {
      start = `${String(s.backup_start_hour).padStart(2, '0')}:${String(s.backup_start_minute).padStart(2, '0')}`;
    }
    if (s.backup_end_hour !== undefined && s.backup_end_minute !== undefined) {
      end = `${String(s.backup_end_hour).padStart(2, '0')}:${String(s.backup_end_minute).padStart(2, '0')}`;
    }
    backupStart.value = start;
    backupEnd.value = end;
    initialScheduleSnapshot.value = serializeScheduleConfig(scheduleForm, start, end);
  } finally {
    scheduleLoading.value = false;
  }
}

async function loadDiskStatus() {
  try {
    const detail = await getStatusDetail();
    if (detail?.disk?.total_gb) {
      totalDiskGb.value = detail.disk.total_gb;
    }
  } catch {
    // Non-critical
  }
}

// Handlers
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
    initialPolicySnapshot.value = serializePolicyConfig(config.value);
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
    } else {
      payload.backup_start_hour = 0;
      payload.backup_start_minute = 0;
    }
    if (backupEnd.value) {
      const [h, m] = backupEnd.value.split(":").map(Number);
      payload.backup_end_hour = h;
      payload.backup_end_minute = m;
    } else {
      payload.backup_end_hour = 0;
      payload.backup_end_minute = 0;
    }
    await saveSchedule(payload);
    ElMessage.success("任务调度设置已保存");
    initialScheduleSnapshot.value = serializeScheduleConfig(scheduleForm, backupStart.value, backupEnd.value);
  } finally {
    savingSchedule.value = false;
  }
}

const OFFICIAL_RECOMMENDED_POLICY = {
  GAP_MINUTES: 30,
  MERGE_AGE_MINUTES: 30,
  TRIGGER_THRESHOLD: 80,
  TARGET_THRESHOLD: 60,
  SAFE_MODE: "hours",
  SAFE_AGE_MINUTES: 120,
  SAFE_DAYS: 1,
  MIN_KEEP_PER_STREAMER: 3,
  MAX_DELETE_PER_RUN: 50,
  WHITELIST_KEYWORDS: ["留存", "纪念", "高能", "生日", "勿删"]
};

async function handleResetDefaults() {
  const targetConfig = {
    ...config.value,
    ...OFFICIAL_RECOMMENDED_POLICY
  };
  const targetSerialized = serializePolicyConfig(targetConfig);
  const currentFormSerialized = serializePolicyConfig(config.value);
  const initialSavedSerialized = initialPolicySnapshot.value;

  // 1. 若当前表单值已完全等于官方预设：
  if (currentFormSerialized === targetSerialized) {
    if (currentFormSerialized === initialSavedSerialized) {
      ElMessage.info("当前配置已是官方推荐预设，无需更改");
    } else {
      ElMessage.info("当前表单已是官方推荐预设，请点击底部「立即保存配置」生效");
    }
    return;
  }

  // 2. 当前表单值与官方预设不同，弹出确认提示：
  try {
    await ElMessageBox.confirm(
      "是否将全局合并与磁盘清理策略重置为系统官方推荐参数？（断流 30 分钟、静置 30 分钟、水位 80%/60%、免删 2 小时、保底 3 场、单次最大 50 个）",
      "恢复推荐预设",
      { confirmButtonText: "确认恢复", cancelButtonText: "取消", type: "warning" }
    );
  } catch {
    return;
  }

  config.value = targetConfig;

  // 3. 应用预设后，比对新表单与数据库基准快照的状态闭环（三态自适应反馈律）：
  if (targetSerialized === initialSavedSerialized) {
    // 恰好还原回数据库原值（草稿被抵消，保存栏自动平滑收起）
    ElMessage.info("已还原为已保存的推荐预设");
  } else {
    // 产生实际配置变更，保存栏平滑升起，正面引导保存
    ElMessage.success("已恢复官方推荐预设，请点击底部「立即保存配置」生效");
  }
}

function handleDiscardChanges() {
  loadConfigData();
  loadScheduleData();
  ElMessage.info("已放弃并还原未保存的更改");
}

function isPolicyDirty(): boolean {
  return Boolean(initialPolicySnapshot.value && serializePolicyConfig(config.value) !== initialPolicySnapshot.value);
}

function isScheduleDirty(): boolean {
  return Boolean(initialScheduleSnapshot.value && serializeScheduleConfig(scheduleForm, backupStart.value, backupEnd.value) !== initialScheduleSnapshot.value);
}

async function handleSaveSettings() {
  const policyDirty = isPolicyDirty();
  const scheduleDirty = isScheduleDirty();

  if (!policyDirty && !scheduleDirty) {
    return;
  }

  if (policyDirty && config.value.TRIGGER_THRESHOLD <= config.value.TARGET_THRESHOLD) {
    ElMessage.error("触发阈值必须大于目标阈值");
    return;
  }

  savingConfig.value = policyDirty;
  savingSchedule.value = scheduleDirty;

  try {
    const promises: Promise<unknown>[] = [];

    if (policyDirty) {
      const payload = { ...config.value };
      if (typeof payload.WHITELIST_KEYWORDS === "string") {
        payload.WHITELIST_KEYWORDS = (payload.WHITELIST_KEYWORDS as string)
          .split(/[,，]/)
          .map((s: string) => s.trim())
          .filter(Boolean);
      }
      promises.push(saveConfig(payload));
    }

    if (scheduleDirty) {
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
      } else {
        payload.backup_start_hour = 0;
        payload.backup_start_minute = 0;
      }
      if (backupEnd.value) {
        const [h, m] = backupEnd.value.split(":").map(Number);
        payload.backup_end_hour = h;
        payload.backup_end_minute = m;
      } else {
        payload.backup_end_hour = 0;
        payload.backup_end_minute = 0;
      }
      promises.push(saveSchedule(payload));
    }

    await Promise.all(promises);

    // Guaranteed EXACTLY ONE Toast authority
    if (policyDirty && scheduleDirty) {
      ElMessage.success("全局策略与调度设置已成功保存");
    } else if (policyDirty) {
      ElMessage.success("全局策略已成功保存");
    } else {
      ElMessage.success("任务调度设置已保存");
    }

    // Refresh snapshots
    if (policyDirty) {
      initialPolicySnapshot.value = serializePolicyConfig(config.value);
      await loadConfigData();
    }
    if (scheduleDirty) {
      initialScheduleSnapshot.value = serializeScheduleConfig(scheduleForm, backupStart.value, backupEnd.value);
    }
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error || err?.message || "保存配置失败");
  } finally {
    savingConfig.value = false;
    savingSchedule.value = false;
  }
}

async function handleCopyConfig() {
  try {
    const res = await getConfigExport();
    const jsonStr = JSON.stringify(res, null, 2);
    await copyToClipboard(jsonStr, "已复制配置 JSON 到剪贴板");
  } catch {
    ElMessage.error("复制配置失败");
  }
}

async function handleExport() {
  try {
    const data = await getConfigExport();
    const exportStr = JSON.stringify(data, null, 2);
    const blob = new Blob([exportStr], { type: "application/json" });
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

function handleFileSelect(file: File) {
  selectedImportFileName.value = file.name;
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
    await loadScheduleData();
  } catch (e: any) {
    if (e instanceof SyntaxError) {
      ElMessage.error("JSON 格式错误，请检查配置内容");
    }
  }
}

onMounted(() => {
  window.addEventListener("resize", handleResize);
  loadConfigData();
  loadScheduleData();
  loadDiskStatus();
});

onUnmounted(() => {
  window.removeEventListener("resize", handleResize);
});
</script>

<style scoped>
.settings {
  max-width: 1040px;
  margin: 0 auto;
  position: relative;
  padding-bottom: 60px;
}

.settings-card {
  border-radius: 12px;
  border: 1px solid var(--hairline);
  background: var(--surface-card);
}

.settings-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.header-left {
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

.unsaved-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #fff8f0;
  border: 1px solid rgba(217, 115, 13, 0.3);
  color: #b35600;
  font-size: 12px;
  font-weight: 500;
  padding: 3px 8px;
  border-radius: 12px;
}
.badge-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #b45309;
}

.settings-tabs {
  min-height: 540px;
  padding: 12px 0;
}
.settings-tab-label {
  display: inline-flex;
  align-items: center;
  gap: 7px;
}
.tab-svg-ico {
  width: 14px;
  height: 14px;
  stroke-width: 2;
  flex-shrink: 0;
  opacity: 0.85;
}
.el-tabs__item:hover .tab-svg-ico,
.el-tabs__item.is-active .tab-svg-ico {
  opacity: 1;
}

.tab-pane-wrap {
  padding: 0 16px 20px 24px;
}

/* Floating Sticky Save Bar (吸底保存条) */
.sticky-save-bar {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 100;
  width: calc(100% - 48px);
  max-width: 640px;
}
.save-bar-inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #252526;
  border: 1px solid #3e3e42;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.25);
  padding: 10px 18px;
  border-radius: var(--r-xl, 14px);
  backdrop-filter: blur(12px);
  color: #fff;
}
.save-bar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.save-bar-pulse-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #b45309;
  box-shadow: 0 0 0 2px rgba(217, 115, 13, 0.4);
}
.save-bar-title {
  font-size: 13px;
  font-weight: 600;
  color: #ffffff;
}
.save-bar-sub {
  font-size: 12px;
  color: #888888;
}
.save-bar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}
.btn-bar-ghost {
  background: transparent;
  border: 1px solid #555555;
  color: #cccccc;
  font-size: 12px;
  padding: 5px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-bar-ghost:hover {
  background: #333333;
  color: #ffffff;
}
.btn-bar-primary {
  background: #448361;
  border: 1px solid #448361;
  color: #ffffff;
  font-size: 12px;
  font-weight: 600;
  padding: 5px 16px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-bar-primary:hover:not(:disabled) {
  background: #36694e;
}
.btn-bar-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.save-bar-slide-enter-active,
.save-bar-slide-leave-active {
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.save-bar-slide-enter-from,
.save-bar-slide-leave-to {
  opacity: 0;
  transform: translate(-50%, 20px);
}

@media (max-width: 768px) {
  .tab-pane-wrap {
    padding: 12px 4px 16px 4px;
  }
  .save-bar-sub {
    display: none;
  }
  .save-bar-inner {
    padding: 8px 12px;
  }
}

.title-group-flex {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}
.unsaved-badge .badge-dot {
  animation: livePulseWarning 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}


html.dark {
  .settings-card {
    background: var(--surface-card);
    border-color: var(--hairline);
  }
  .unsaved-badge {
    background: rgba(245, 158, 11, 0.12) !important;
    border-color: rgba(245, 158, 11, 0.25) !important;
    color: #fbbf24 !important;
  }
  .save-bar-inner {
    background: #1c1c20;
    border-color: #2e2e33;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.6);
  }
}

</style>

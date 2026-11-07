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

      <!-- Continuous Scroll Layout with Sticky Anchor Sidebar -->
      <div class="settings-layout">
        <!-- Left: Sticky Anchor Navigation -->
        <aside class="settings-sidebar">
          <nav class="settings-nav">
            <button
              v-for="item in navSections"
              :key="item.id"
              type="button"
              class="nav-anchor-btn"
              :class="{ active: activeSectionId === item.id }"
              @click="scrollToSection(item.id)"
            >
              <span class="nav-btn-icon">
                <svg v-if="item.id === 'policy'" class="tab-svg-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="4" y1="21" x2="4" y2="14"/><line x1="4" y1="10" x2="4" y2="3"/><line x1="12" y1="21" x2="12" y2="12"/><line x1="12" y1="8" x2="12" y2="3"/><line x1="20" y1="21" x2="20" y2="16"/><line x1="20" y1="12" x2="20" y2="3"/><line x1="1" y1="14" x2="7" y2="14"/><line x1="9" y1="8" x2="15" y2="8"/><line x1="17" y1="16" x2="23" y2="16"/></svg>
                <svg v-else-if="item.id === 'schedule'" class="tab-svg-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                <svg v-else-if="item.id === 'backup'" class="tab-svg-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                <svg v-else-if="item.id === 'security'" class="tab-svg-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
              </span>
              <span class="nav-btn-text">{{ item.label }}</span>
            </button>
          </nav>
        </aside>

        <!-- Right: Continuous Stream of Sections -->
        <main class="settings-content" ref="contentRef">
          <!-- 1. 全局策略 -->
          <section id="section-policy" class="settings-section" data-section="policy">
            <div v-loading="configLoading" class="section-pane-wrap">
              <PolicyTab
                :config="config"
                :saving="savingConfig"
                :total-disk-gb="totalDiskGb"
                @update:config="config = $event"
                @reset-defaults="handleResetDefaults"
              />
            </div>
          </section>

          <div class="section-divider" />

          <!-- 2. 任务调度 -->
          <section id="section-schedule" class="settings-section" data-section="schedule">
            <div v-loading="scheduleLoading" class="section-pane-wrap">
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
          </section>

          <div class="section-divider" />

          <!-- 3. 配置备份与恢复 -->
          <section id="section-backup" class="settings-section" data-section="backup">
            <div class="section-pane-wrap">
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
          </section>

          <div class="section-divider" />

          <!-- 4. 安全设置 -->
          <section id="section-security" class="settings-section" data-section="security">
            <div class="section-pane-wrap">
              <SecurityTab />
            </div>
          </section>
        </main>
      </div>
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
import { ref, reactive, computed, watch, onMounted, onUnmounted, onActivated, nextTick } from "vue";
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

const navSections = [
  { id: "policy", label: "全局策略" },
  { id: "schedule", label: "定时调度" },
  { id: "backup", label: "配置备份" },
  { id: "security", label: "安全设置" }
];

const activeSectionId = ref<string>(
  typeof route.query.tab === "string" && ["policy", "schedule", "backup", "security"].includes(route.query.tab)
    ? (route.query.tab as string)
    : "policy"
);

let isUserClicking = false;
let userClickTimer: ReturnType<typeof setTimeout> | null = null;

function scrollToSection(id: string) {
  activeSectionId.value = id;
  isUserClicking = true;
  if (userClickTimer) clearTimeout(userClickTimer);

  const el = document.getElementById(`section-${id}`);
  if (el) {
    const headerOffset = 70;
    const elementPosition = el.getBoundingClientRect().top;
    const offsetPosition = elementPosition + window.pageYOffset - headerOffset;
    window.scrollTo({
      top: offsetPosition,
      behavior: "smooth"
    });
  }

  userClickTimer = setTimeout(() => {
    isUserClicking = false;
  }, 750);
}

// ScrollSpy via IntersectionObserver
let observer: IntersectionObserver | null = null;

function setupScrollSpy() {
  if (observer) {
    observer.disconnect();
  }
  const sections = document.querySelectorAll(".settings-section");
  observer = new IntersectionObserver(
    (entries) => {
      if (isUserClicking) return;
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          const sec = entry.target.getAttribute("data-section");
          if (sec) {
            activeSectionId.value = sec;
          }
        }
      });
    },
    {
      rootMargin: "-15% 0px -70% 0px",
      threshold: 0
    }
  );
  sections.forEach((sec) => observer?.observe(sec));
}

function handleWindowScroll() {
  if (isUserClicking) return;
  // If scrolled close to bottom of page, highlight the last section
  if (window.innerHeight + window.scrollY >= document.documentElement.scrollHeight - 60) {
    activeSectionId.value = "security";
  }
}

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

function isPolicyDirty(): boolean {
  return Boolean(initialPolicySnapshot.value && serializePolicyConfig(config.value) !== initialPolicySnapshot.value);
}

function isScheduleDirty(): boolean {
  return Boolean(initialScheduleSnapshot.value && serializeScheduleConfig(scheduleForm, backupStart.value, backupEnd.value) !== initialScheduleSnapshot.value);
}

function handleDiscardChanges() {
  if (initialPolicySnapshot.value) {
    try {
      loadConfigData();
    } catch {
      // ignore
    }
  }
  if (initialScheduleSnapshot.value) {
    try {
      loadScheduleData();
    } catch {
      // ignore
    }
  }
  ElMessage.info("已放弃未保存的配置更改，恢复为上次保存状态");
}

// 离开路由守卫：若有未保存改动弹窗拦截
onBeforeRouteLeave((to, from, next) => {
  if (isFormDirty()) {
    ElMessageBox.confirm(
      "您在当前设置页面有未保存的配置变更，离开后变更将会丢失，是否确认离开？",
      "未保存配置提示",
      {
        confirmButtonText: "离开页面",
        cancelButtonText: "留在此页",
        confirmButtonClass: "el-button--danger",
        type: "warning"
      }
    )
      .then(() => next())
      .catch(() => next(false));
  } else {
    next();
  }
});

async function loadConfigData() {
  configLoading.value = true;
  try {
    const data = await getConfig();
    config.value = { ...data };
    initialPolicySnapshot.value = serializePolicyConfig(data);
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error || err?.message || "加载全局策略失败");
  } finally {
    configLoading.value = false;
  }
}

async function loadScheduleData() {
  scheduleLoading.value = true;
  try {
    const data: ScheduleStatus = await getSchedule();
    scheduleForm.merge_enabled = data.merge.enabled;
    scheduleForm.clean_enabled = data.clean.enabled;
    scheduleForm.merge_interval = data.merge.interval;
    scheduleForm.clean_interval = data.clean.interval;

    if (data.backup?.start_hour !== undefined && data.backup?.start_minute !== undefined) {
      backupStart.value = `${String(data.backup.start_hour).padStart(2, "0")}:${String(data.backup.start_minute).padStart(2, "0")}`;
    }
    if (data.backup?.end_hour !== undefined && data.backup?.end_minute !== undefined) {
      backupEnd.value = `${String(data.backup.end_hour).padStart(2, "0")}:${String(data.backup.end_minute).padStart(2, "0")}`;
    }

    initialScheduleSnapshot.value = serializeScheduleConfig(scheduleForm, backupStart.value, backupEnd.value);
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error || err?.message || "加载任务调度设置失败");
  } finally {
    scheduleLoading.value = false;
  }
}

async function loadDiskStatus() {
  try {
    const res = await getStatusDetail();
    if (res?.disk?.total_gb) {
      totalDiskGb.value = res.disk.total_gb;
    }
  } catch {
    // optional status
  }
}

const OFFICIAL_RECOMMENDED_POLICY = {
  GAP_MINUTES: 30,
  MERGE_AGE_MINUTES: 30,
  TRIGGER_THRESHOLD: 80,
  TARGET_THRESHOLD: 60,
  SAFE_MODE: "days",
  SAFE_AGE_MINUTES: 1440 * 3,
  SAFE_DAYS: 3,
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

  if (currentFormSerialized === targetSerialized) {
    if (currentFormSerialized === initialSavedSerialized) {
      ElMessage.info("当前配置已是官方推荐预设，无需更改");
    } else {
      ElMessage.info("表单已应用官方推荐预设，请点击下方保存按钮提交生效");
    }
    return;
  }

  try {
    await ElMessageBox.confirm(
      "是否将全局合并与磁盘清理策略重置为系统官方推荐参数？（断流 30 分钟、静置 30 分钟、水位 80%/60%、免删 3 天、白名单保护关键词）\n\n重置后需点击下方保存按钮生效。",
      "恢复官方推荐预设",
      {
        confirmButtonText: "确认应用预设",
        cancelButtonText: "取消",
        type: "warning"
      }
    );
  } catch {
    return;
  }

  config.value = targetConfig;
  ElMessage.success("已载入官方推荐策略预设，请检查并点击下方保存按钮生效");
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
    ElMessage.success("系统配置已成功保存生效");

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
  reader.readAsText(file);
}

async function handleImport() {
  if (!importJson.value.trim()) {
    ElMessage.warning("请先选择备份文件或粘贴配置 JSON");
    return;
  }
  try {
    const data = JSON.parse(importJson.value);
    try {
      await ElMessageBox.confirm(
        "导入备份将完全覆盖当前的合并策略、清理阈值和调度周期，是否继续？",
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

// Watch route query to jump to section
watch(
  () => route.query.tab,
  (newTab) => {
    if (typeof newTab === "string" && ["policy", "schedule", "backup", "security"].includes(newTab)) {
      scrollToSection(newTab);
    }
  }
);

onActivated(() => {
  if (typeof route.query.tab === "string" && ["policy", "schedule", "backup", "security"].includes(route.query.tab)) {
    nextTick(() => scrollToSection(route.query.tab as string));
  }
  loadConfigData();
  loadScheduleData();
  loadDiskStatus();
});

onMounted(() => {
  window.addEventListener("scroll", handleWindowScroll, { passive: true });
  loadConfigData();
  loadScheduleData();
  loadDiskStatus();
  nextTick(() => {
    setupScrollSpy();
    if (typeof route.query.tab === "string" && ["policy", "schedule", "backup", "security"].includes(route.query.tab)) {
      scrollToSection(route.query.tab);
    }
  });
});

onUnmounted(() => {
  window.removeEventListener("scroll", handleWindowScroll);
  if (observer) observer.disconnect();
  if (userClickTimer) clearTimeout(userClickTimer);
});
</script>

<style scoped>
.settings {
  max-width: 1080px;
  margin: 0 auto;
  position: relative;
  padding-bottom: 80px;
}

.settings-card {
  border-radius: var(--r-lg, 12px);
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

.title-group-flex {
  display: inline-flex;
  align-items: center;
  gap: 8px;
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

/* Two-column Continuous Scroll Layout */
.settings-layout {
  display: flex;
  gap: 24px;
  padding: 16px 0 24px 0;
  align-items: flex-start;
}

/* Sticky Left Sidebar Navigation */
.settings-sidebar {
  width: 180px;
  flex-shrink: 0;
  position: sticky;
  top: 80px;
  z-index: 10;
}

.settings-nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
  background: var(--bg-hover, #f4f4f5);
  padding: 6px;
  border-radius: var(--r-md, 8px);
  border: 1px solid var(--hairline, #e4e4e7);
}

.nav-anchor-btn {
  display: flex;
  align-items: center;
  gap: 9px;
  height: 38px;
  padding: 0 12px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--slate, #71717a);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
  text-align: left;
  width: 100%;

  .nav-btn-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .tab-svg-ico {
    width: 15px;
    height: 15px;
    stroke-width: 2;
    opacity: 0.8;
  }

  &:hover {
    color: var(--charcoal, #18181b);
    background: rgba(0, 0, 0, 0.04);
  }

  &.active {
    background: var(--surface, #ffffff);
    color: #4f46e5;
    font-weight: 600;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);

    .tab-svg-ico {
      opacity: 1;
      color: #4f46e5;
    }
  }
}

/* Right Content Stream */
.settings-content {
  max-width: 860px;
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 28px;
}

.settings-section {
  scroll-margin-top: 80px;
}

.section-pane-wrap {
  width: 100%;
}

.section-divider {
  height: 1px;
  background: var(--hairline, #e4e4e7);
  margin: 4px 0;
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
}
.save-bar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}
.save-bar-pulse-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #f59e0b;
  box-shadow: 0 0 0 0 rgba(245, 158, 11, 0.7);
  animation: pulse-amber 2s infinite;
}
.save-bar-title {
  font-size: 13px;
  font-weight: 600;
  color: #ffffff;
}
.save-bar-sub {
  font-size: 12px;
  color: #a1a1aa;
}
.save-bar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.btn-bar-ghost {
  height: 32px;
  padding: 0 14px;
  font-size: 12px;
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: #d4d4d8;
  border-radius: var(--r-sm, 6px);
  cursor: pointer;
  transition: all 0.15s ease;
}
.btn-bar-ghost:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #ffffff;
}
.btn-bar-primary {
  height: 32px;
  padding: 0 16px;
  font-size: 12px;
  font-weight: 600;
  background: #4f46e5;
  border: 1px solid #4338ca;
  color: #ffffff;
  border-radius: var(--r-sm, 6px);
  cursor: pointer;
  transition: all 0.15s ease;
}
.btn-bar-primary:hover:not(:disabled) {
  background: #4338ca;
}
.btn-bar-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Animations */
.save-bar-slide-enter-active,
.save-bar-slide-leave-active {
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.save-bar-slide-enter-from,
.save-bar-slide-leave-to {
  opacity: 0;
  transform: translate(-50%, 20px);
}

@keyframes pulse-amber {
  0% { box-shadow: 0 0 0 0 rgba(245, 158, 11, 0.7); }
  70% { box-shadow: 0 0 0 6px rgba(245, 158, 11, 0); }
  100% { box-shadow: 0 0 0 0 rgba(245, 158, 11, 0); }
}

/* Dark Mode Overrides */
html.dark {
  .settings-nav {
    background: #18181b;
    border-color: #27272a;
  }
  .nav-anchor-btn {
    color: #a1a1aa;

    &:hover {
      color: #f4f4f5;
      background: rgba(255, 255, 255, 0.05);
    }

    &.active {
      background: #27272a;
      color: #818cf8;
      box-shadow: 0 1px 4px rgba(0, 0, 0, 0.4);

      .tab-svg-ico {
        color: #818cf8;
      }
    }
  }
  .section-divider {
    background: #27272a;
  }
}

/* Responsive */
@media (max-width: 768px) {
  .settings-layout {
    flex-direction: column;
    gap: 16px;
  }
  .settings-sidebar {
    width: 100%;
    position: sticky;
    top: 60px;
  }
  .settings-nav {
    flex-direction: row;
    overflow-x: auto;
    padding: 4px;
    gap: 4px;
  }
  .nav-anchor-btn {
    white-space: nowrap;
    flex-shrink: 0;
  }
  .save-bar-sub {
    display: none;
  }
  .save-bar-inner {
    padding: 8px 12px;
  }
}
</style>

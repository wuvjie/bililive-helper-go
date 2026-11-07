<template>
  <div class="settings">
    <!-- Page Header (Top Level, outside card) -->
    <header class="settings-page-header">
      <div class="header-text-wrap">
        <h1 class="page-title">系统设置</h1>
        <p class="page-desc">管理自动化录制合并、磁盘巡检保护、计划调度与系统安全凭据</p>
      </div>
    </header>

    <!-- Main Two-Column Layout -->
    <div class="settings-layout">
      <!-- Left: Sticky Sidebar Navigation (200px) -->
      <aside class="settings-sidebar">
        <div class="sidebar-sticky-inner">
          <div class="sidebar-caption">配置目录</div>
          <nav class="settings-nav">
            <button
              v-for="item in navSections"
              :key="item.id"
              class="nav-tab-btn"
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
              <span v-if="activeSectionId === item.id" class="nav-active-pill" />
            </button>
          </nav>
        </div>
      </aside>

      <!-- Right: Continuous Flow of Sections (All same width, zero nested border madness) -->
      <main class="settings-content" ref="contentRef">
        <!-- 1. 全局策略 -->
        <section id="section-policy" class="settings-section-card" data-section="policy">
          <div class="section-card-header">
            <div class="section-header-title-row">
              <div class="ui-icon-box icon-box-emerald">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
              </div>
              <div class="section-header-texts">
                <h2 class="section-main-title">全局策略</h2>
                <p class="section-sub-title">配置录播分片合并静置容差时间、磁盘巡检保护水位以及新录像免删天数</p>
              </div>
            </div>
          </div>
          <div class="section-card-body" v-loading="configLoading">
            <PolicyTab
              :config="config"
              :saving="savingConfig"
              :total-disk-gb="totalDiskGb"
              @update:config="config = $event"
              @reset-defaults="handleResetDefaults"
            />
          </div>
        </section>

        <!-- 2. 定时调度 -->
        <section id="section-schedule" class="settings-section-card" data-section="schedule">
          <div class="section-card-header">
            <div class="section-header-title-row">
              <div class="ui-icon-box icon-box-blue">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
              </div>
              <div class="section-header-texts">
                <h2 class="section-main-title">定时调度</h2>
                <p class="section-sub-title">设定自动分片合并与磁盘清理巡检的运行周期与自动执行开关</p>
              </div>
            </div>
          </div>
          <div class="section-card-body" v-loading="scheduleLoading">
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

        <!-- 3. 配置备份与恢复 (统一整卡) -->
        <section id="section-backup" class="settings-section-card" data-section="backup">
          <div class="section-card-header">
            <div class="section-header-title-row">
              <div class="ui-icon-box icon-box-purple">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
              </div>
              <div class="section-header-texts">
                <h2 class="section-main-title">配置备份与恢复</h2>
                <p class="section-sub-title">导出当前系统全部策略规则与调度参数，或从已有备份中导入恢复</p>
              </div>
            </div>
          </div>
          <div class="section-card-body">
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

        <!-- 4. 安全设置 (统一整卡) -->
        <section id="section-security" class="settings-section-card" data-section="security">
          <div class="section-card-header">
            <div class="section-header-title-row">
              <div class="ui-icon-box icon-box-amber">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
              </div>
              <div class="section-header-texts">
                <h2 class="section-main-title">安全设置</h2>
                <p class="section-sub-title">修改 Web 管理控制台访问密码与安全凭据</p>
              </div>
            </div>
          </div>
          <div class="section-card-body">
            <SecurityTab />
          </div>
        </section>
      </main>
    </div>

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
import { ref, reactive, watch, onMounted, onUnmounted, onActivated, nextTick } from "vue";
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

// Helper: 获取最外层滚动容器 (.layout-content)
function getScrollContainer(): HTMLElement | null {
  return document.querySelector(".layout-content") as HTMLElement | null;
}

function scrollToSection(id: string) {
  activeSectionId.value = id;
  isUserClicking = true;
  if (userClickTimer) clearTimeout(userClickTimer);

  const container = getScrollContainer();
  const el = document.getElementById(`section-${id}`);
  if (container && el) {
    const containerRect = container.getBoundingClientRect();
    const elRect = el.getBoundingClientRect();
    const targetScroll = container.scrollTop + (elRect.top - containerRect.top) - 16;
    container.scrollTo({
      top: Math.max(0, targetScroll),
      behavior: "smooth"
    });
  } else if (el) {
    el.scrollIntoView({ behavior: "smooth", block: "start" });
  }

  userClickTimer = setTimeout(() => {
    isUserClicking = false;
  }, 750);
}

function handleContainerScroll() {
  if (isUserClicking) return;
  const container = getScrollContainer();
  if (!container) return;
  if (container.scrollHeight - container.scrollTop - container.clientHeight < 60) {
    activeSectionId.value = "security";
  }
}

let observer: IntersectionObserver | null = null;

function setupScrollSpy() {
  if (observer) {
    observer.disconnect();
    observer = null;
  }
  const container = getScrollContainer();
  const sections = document.querySelectorAll(".settings-section-card");
  if (!sections.length) return;

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
      root: container,
      rootMargin: "-10% 0px -60% 0px",
      threshold: [0, 0.2]
    }
  );
  sections.forEach((sec) => observer?.observe(sec));
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

// Baseline Snapshot System
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

// 离开路由守卫
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
  const container = getScrollContainer();
  if (container) {
    container.addEventListener("scroll", handleContainerScroll, { passive: true });
  }
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
  const container = getScrollContainer();
  if (container) {
    container.removeEventListener("scroll", handleContainerScroll);
  }
  if (observer) observer.disconnect();
  if (userClickTimer) clearTimeout(userClickTimer);
});
</script>

<style scoped>
.settings {
  max-width: 1100px;
  margin: 0 auto;
  position: relative;
  padding-bottom: 90px;
}

/* Page Level Header */
.settings-page-header {
  margin-bottom: 24px;
}
.page-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--charcoal);
  margin: 0 0 4px 0;
  letter-spacing: -0.3px;
}
.page-desc {
  font-size: 13px;
  color: var(--steel);
  margin: 0;
}

/* Layout */
.settings-layout {
  display: flex;
  gap: 28px;
  align-items: flex-start;
}
@media (max-width: 860px) {
  .settings-layout {
    flex-direction: column;
  }
  .settings-sidebar {
    width: 100% !important;
    position: static !important;
  }
  .settings-nav {
    display: flex !important;
    flex-direction: row !important;
    overflow-x: auto;
    padding-bottom: 6px;
  }
}

/* Sticky Sidebar (200px) */
.settings-sidebar {
  width: 200px;
  flex-shrink: 0;
  position: sticky;
  top: 16px;
}
.sidebar-sticky-inner {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.sidebar-caption {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--muted);
  letter-spacing: 0.8px;
  padding-left: 12px;
}
.settings-nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.nav-tab-btn {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  height: 40px;
  padding: 0 14px;
  border-radius: var(--r-sm, 6px);
  border: none;
  background: transparent;
  color: var(--slate, #64748b);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  text-align: left;
  transition: all 0.15s ease;
  white-space: nowrap;
}
.nav-tab-btn:hover {
  color: var(--charcoal);
  background: var(--surface-subtle, rgba(0, 0, 0, 0.04));
}
.nav-tab-btn.active {
  color: var(--color-primary, #3b82f6);
  background: var(--highlight, rgba(59, 130, 246, 0.08));
  font-weight: 600;
}
.nav-active-pill {
  position: absolute;
  left: 0;
  top: 8px;
  bottom: 8px;
  width: 3px;
  background: var(--color-primary, #3b82f6);
  border-radius: 0 2px 2px 0;
}
.tab-svg-ico {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

/* Right Content Main Stream (Uniform 100% width, no width mismatches) */
.settings-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 28px;
}

/* Unified Section Card (Every section looks identical in width, radius & borders) */
.settings-section-card {
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--hairline, #e4e4e7);
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.03);
}

.section-card-header {
  padding: 18px 22px;
  border-bottom: 1px solid var(--hairline, #e4e4e7);
  background: var(--surface-subtle, rgba(0, 0, 0, 0.015));
}
.section-header-title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.section-header-texts {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.section-main-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--charcoal);
}
.section-sub-title {
  margin: 0;
  font-size: 12px;
  color: var(--steel);
}

.section-card-body {
  padding: 22px;
}

/* Eliminate nested card borders inside PolicyTab, ScheduleTab, BackupTab, SecurityTab */
.settings-section-card :deep(.form-section-card),
.settings-section-card :deep(.backup-card) {
  border: none !important;
  background: transparent !important;
  padding: 0 !important;
  margin: 0 !important;
  box-shadow: none !important;
}
.settings-section-card :deep(.form-section-card + .form-section-card),
.settings-section-card :deep(.backup-card + .backup-card) {
  margin-top: 24px !important;
  padding-top: 24px !important;
  border-top: 1px solid var(--hairline) !important;
}

/* Floating Sticky Save Bar (吸底保存条) */
.sticky-save-bar {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 1000;
  width: calc(100% - 48px);
  max-width: 820px;
  animation: barFloat 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.save-bar-inner {
  background: rgba(24, 24, 27, 0.94);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 40px;
  padding: 10px 18px 10px 22px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.35), 0 2px 8px rgba(0, 0, 0, 0.2);
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
  animation: pulse-warn-dot 1.8s infinite;
}
@keyframes pulse-warn-dot {
  0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(245, 158, 11, 0.7); }
  70% { transform: scale(1); box-shadow: 0 0 0 6px rgba(245, 158, 11, 0); }
  100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(245, 158, 11, 0); }
}
.save-bar-title {
  color: #f4f4f5;
  font-size: 13px;
  font-weight: 600;
}
.save-bar-sub {
  color: #a1a1aa;
  font-size: 12px;
}
.save-bar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}
.btn-bar-ghost {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #d4d4d8;
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-bar-ghost:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #ffffff;
}
.btn-bar-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #3b82f6;
  border: 1px solid #60a5fa;
  color: #ffffff;
  padding: 6px 18px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.4);
  transition: all 0.15s;
}
.btn-bar-primary:hover:not(:disabled) {
  background: #2563eb;
  transform: translateY(-1px);
}
.btn-bar-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Animations */
.save-bar-slide-enter-active,
.save-bar-slide-leave-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.save-bar-slide-enter-from,
.save-bar-slide-leave-to {
  opacity: 0;
  transform: translate(-50%, 20px);
}

/* Dark mode tweaks */
html.dark .nav-tab-btn {
  color: #94a3b8;
}
html.dark .nav-tab-btn:hover {
  color: #f1f5f9;
  background: rgba(255, 255, 255, 0.05);
}
html.dark .nav-tab-btn.active {
  color: #60a5fa;
  background: rgba(59, 130, 246, 0.12);
}
html.dark .section-card-header {
  background: rgba(255, 255, 255, 0.02);
}
</style>
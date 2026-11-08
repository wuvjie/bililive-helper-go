<template>
  <div class="settings">
    <!-- Sticky Top UnderlineNav Header (GitHub Primer 规范横向导航栏) -->
    <div class="settings-top-nav-bar">
      <div class="top-nav-inner">
        <nav class="settings-nav-underline" role="tablist" aria-label="系统设置分类导航">
          <button
            v-for="item in navSections"
            :key="item.id"
            type="button"
            role="tab"
            :aria-selected="activeSectionId === item.id"
            class="nav-underline-item"
            :class="{ active: activeSectionId === item.id }"
            @click="handleTabClick(item.id)"
          >
            <span class="nav-item-icon">
              <svg v-if="item.id === 'policy'" class="tab-svg-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="4" y1="21" x2="4" y2="14"/><line x1="4" y1="10" x2="4" y2="3"/><line x1="12" y1="21" x2="12" y2="12"/><line x1="12" y1="8" x2="12" y2="3"/><line x1="20" y1="21" x2="20" y2="16"/><line x1="20" y1="12" x2="20" y2="3"/><line x1="1" y1="14" x2="7" y2="14"/><line x1="9" y1="8" x2="15" y2="8"/><line x1="17" y1="16" x2="23" y2="16"/></svg>
              <svg v-else-if="item.id === 'schedule'" class="tab-svg-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
            </span>
            <span class="nav-item-text">{{ item.label }}</span>
            <span v-if="item.id === 'policy' && isPolicyDirty()" class="nav-tab-dirty-dot" title="包含未保存的策略更改" />
            <span v-if="item.id === 'schedule' && isScheduleDirty()" class="nav-tab-dirty-dot" title="包含未保存的调度更改" />
          </button>
        </nav>

        <div class="top-nav-actions">
          <button
            v-if="activeSectionId === 'policy'"
            type="button"
            class="top-preset-action-btn"
            @click="handleResetPolicyDefaults"
            title="将全局合并与磁盘清理策略参数恢复为官方推荐预设"
          >
            <svg class="reset-action-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>
            <span>恢复策略预设</span>
          </button>
          <button
            v-else-if="activeSectionId === 'schedule'"
            type="button"
            class="top-preset-action-btn"
            @click="handleResetScheduleDefaults"
            title="将定时调度运行周期与避让窗口恢复为官方推荐预设"
          >
            <svg class="reset-action-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>
            <span>恢复调度预设</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Main Full-Width Content Stream (Native Smooth Flow) -->
    <main class="settings-content">
      <Transition name="tab-fade" mode="out-in">
        <!-- 1. 全局策略 -->
        <div
          v-if="activeSectionId === 'policy'"
          key="policy"
          id="section-policy"
          class="settings-module-wrap"
        >
          <div class="module-body" v-loading="configLoading">
            <PolicyTab
              :config="config"
              :saving="savingConfig"
              :total-disk-gb="totalDiskGb"
              @update:config="config = $event"
            />
          </div>
        </div>

        <!-- 2. 定时调度 -->
        <div
          v-else-if="activeSectionId === 'schedule'"
          key="schedule"
          id="section-schedule"
          class="settings-module-wrap"
        >
          <div class="module-body" v-loading="scheduleLoading">
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
        </div>
      </Transition>
    </main>

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
              {{ (savingConfig || savingSchedule) ? '保存中...' : '保存变更' }}
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- 恢复预设确认弹窗 (GitHub Primer 规范卡片弹窗) -->
    <el-dialog
      v-model="presetModalVisible"
      :title="presetModalType === 'policy' ? '恢复策略预设' : '恢复调度预设'"
      width="460px"
      destroy-on-close
      class="preset-confirm-dialog"
    >
      <div class="preset-dialog-body">
        <div class="preset-dialog-desc">
          {{ presetModalType === 'policy' ? '将录制合并规则与磁盘清理水位恢复为官方推荐的基准参数：' : '将自动化分片合并与磁盘巡检周期恢复为官方推荐的基准参数：' }}
        </div>

        <!-- 结构化参数展示卡片 -->
        <div class="preset-summary-card">
          <div v-if="presetModalType === 'policy'" class="preset-params-list">
            <div class="preset-param-item">
              <span class="param-lbl">断流重连容差</span>
              <span class="param-val">30 分钟</span>
            </div>
            <div class="preset-param-item">
              <span class="param-lbl">完播静置等待</span>
              <span class="param-val">30 分钟</span>
            </div>
            <div class="preset-param-item">
              <span class="param-lbl">触发清理水位</span>
              <span class="param-val">80% 触发 · 降至 60%</span>
            </div>
            <div class="preset-param-item">
              <span class="param-lbl">免删保护底线</span>
              <span class="param-val">保留最近 3 天录像</span>
            </div>
          </div>

          <div v-else class="preset-params-list">
            <div class="preset-param-item">
              <span class="param-lbl">自动分片合并</span>
              <span class="param-val">启用 · 每 360 分钟</span>
            </div>
            <div class="preset-param-item">
              <span class="param-lbl">自动磁盘清理</span>
              <span class="param-val">启用 · 每 720 分钟</span>
            </div>
            <div class="preset-param-item">
              <span class="param-lbl">避让时间窗口</span>
              <span class="param-val">全天候运行 (无避让)</span>
            </div>
          </div>
        </div>

        <div class="preset-dialog-hint">
          <svg class="hint-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/>
          </svg>
          <span>预设参数将载入当前表单，您可在检查无误后，点击底部的「保存变更」正式生效。</span>
        </div>
      </div>

      <template #footer>
        <div class="preset-dialog-footer">
          <el-button @click="presetModalVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmApplyPreset">确认</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onActivated } from "vue";
import { useRoute, useRouter, onBeforeRouteLeave } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { getConfig, saveConfig } from "@/api/config";
import { getSchedule, saveSchedule } from "@/api/schedule";
import { getStatusDetail } from "@/api/status";
import type { ScheduleStatus, ConfigDTO } from "@/api/types";

import PolicyTab from "./tabs/PolicyTab.vue";
import ScheduleTab from "./tabs/ScheduleTab.vue";
import type { ScheduleForm } from "./tabs/ScheduleTab.vue";

const route = useRoute();
const router = useRouter();

const navSections = [
  { id: "policy", label: "全局策略" },
  { id: "schedule", label: "定时调度" }
];

const activeSectionId = ref<string>(
  typeof route.query.tab === "string" && ["policy", "schedule"].includes(route.query.tab)
    ? (route.query.tab as string)
    : "policy"
);

function updateRouteQuery(tabId: string) {
  if (route.query.tab !== tabId) {
    router.replace({ query: { ...route.query, tab: tabId } }).catch(() => {});
  }
}

function switchToSection(id: string) {
  if (activeSectionId.value === id) return;
  activeSectionId.value = id;
  updateRouteQuery(id);
  const container = document.querySelector(".layout-content") as HTMLElement | null;
  if (container) {
    container.scrollTop = 0;
  }
}

function handleTabClick(id: string) {
  switchToSection(id);
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

    const startHour = data.backup_start_hour ?? data.backup?.start_hour;
    const startMinute = data.backup_start_minute ?? data.backup?.start_minute;
    if (startHour !== undefined && startMinute !== undefined) {
      backupStart.value = `${String(startHour).padStart(2, "0")}:${String(startMinute).padStart(2, "0")}`;
    }

    const endHour = data.backup_end_hour ?? data.backup?.end_hour;
    const endMinute = data.backup_end_minute ?? data.backup?.end_minute;
    if (endHour !== undefined && endMinute !== undefined) {
      backupEnd.value = `${String(endHour).padStart(2, "0")}:${String(endMinute).padStart(2, "0")}`;
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

const OFFICIAL_RECOMMENDED_SCHEDULE = {
  merge_enabled: true,
  clean_enabled: true,
  merge_interval: 360,
  clean_interval: 720
};

const presetModalVisible = ref(false);
const presetModalType = ref<"policy" | "schedule">("policy");

function handleResetPolicyDefaults() {
  const targetConfig = {
    ...config.value,
    ...OFFICIAL_RECOMMENDED_POLICY
  };
  const targetSerialized = serializePolicyConfig(targetConfig);
  const currentFormSerialized = serializePolicyConfig(config.value);
  const initialSavedSerialized = initialPolicySnapshot.value;

  if (currentFormSerialized === targetSerialized) {
    if (currentFormSerialized === initialSavedSerialized) {
      ElMessage.info("当前策略已是官方推荐预设，无需更改");
    } else {
      ElMessage.info("表单已应用官方推荐策略预设，请检查并点击下方保存变更生效");
    }
    return;
  }

  presetModalType.value = "policy";
  presetModalVisible.value = true;
}

function handleResetScheduleDefaults() {
  const isAlreadyRecommended =
    scheduleForm.merge_enabled === OFFICIAL_RECOMMENDED_SCHEDULE.merge_enabled &&
    scheduleForm.clean_enabled === OFFICIAL_RECOMMENDED_SCHEDULE.clean_enabled &&
    scheduleForm.merge_interval === OFFICIAL_RECOMMENDED_SCHEDULE.merge_interval &&
    scheduleForm.clean_interval === OFFICIAL_RECOMMENDED_SCHEDULE.clean_interval &&
    !backupStart.value &&
    !backupEnd.value;

  if (isAlreadyRecommended) {
    if (!isScheduleDirty()) {
      ElMessage.info("当前定时调度已是官方推荐预设，无需更改");
    } else {
      ElMessage.info("表单已应用官方推荐调度预设，请检查并点击下方保存变更生效");
    }
    return;
  }

  presetModalType.value = "schedule";
  presetModalVisible.value = true;
}

function confirmApplyPreset() {
  presetModalVisible.value = false;
  if (presetModalType.value === "policy") {
    config.value = {
      ...config.value,
      ...OFFICIAL_RECOMMENDED_POLICY
    };
    ElMessage.success("已载入官方推荐策略预设，请检查并点击下方保存变更生效");
  } else {
    scheduleForm.merge_enabled = true;
    scheduleForm.clean_enabled = true;
    scheduleForm.merge_interval = 360;
    scheduleForm.clean_interval = 720;
    backupStart.value = "";
    backupEnd.value = "";
    ElMessage.success("已载入官方推荐调度预设，请检查并点击下方保存变更生效");
  }
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

// Watch route query to jump to section
watch(
  () => route.query.tab,
  (newTab) => {
    if (typeof newTab === "string" && ["policy", "schedule"].includes(newTab)) {
      if (activeSectionId.value !== newTab) {
        switchToSection(newTab);
      }
    }
  }
);

onActivated(() => {
  if (typeof route.query.tab === "string" && ["policy", "schedule"].includes(route.query.tab)) {
    if (activeSectionId.value !== route.query.tab) {
      activeSectionId.value = route.query.tab;
    }
  }
  loadConfigData();
  loadScheduleData();
  loadDiskStatus();
});

onMounted(() => {
  loadConfigData();
  loadScheduleData();
  loadDiskStatus();
  if (typeof route.query.tab === "string" && ["policy", "schedule"].includes(route.query.tab)) {
    activeSectionId.value = route.query.tab;
  }
});
</script>

<style scoped>
.settings {
  max-width: 980px;
  margin: 0 auto;
  position: relative;
  padding-bottom: 90px;
}

/* Page Level Header */
.settings-page-header {
  margin-bottom: 16px;
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

/* Sticky Top UnderlineNav Header (GitHub Primer 规范横向吸顶导航栏：毛玻璃背景 + 两极分布排版) */
.settings-top-nav-bar {
  position: sticky;
  top: -24px;
  margin-top: -24px;
  padding-top: 24px;
  background: color-mix(in srgb, var(--canvas, #f6f8fa) 88%, transparent);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--hairline, #e4e4e7);
  margin-bottom: 24px;
  z-index: 20;
}

html.dark .settings-top-nav-bar {
  background: color-mix(in srgb, var(--canvas, #0d1117) 88%, transparent);
  border-bottom-color: #30363d;
}

.top-nav-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0 4px;
  box-sizing: border-box;
}

.top-nav-actions {
  display: flex;
  align-items: center;
  margin-left: auto;
}

.top-preset-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  color: var(--color-fg-muted, #57606a);
  background: transparent;
  border: 1px solid transparent;
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;
  white-space: nowrap;

  .reset-action-svg {
    width: 14px;
    height: 14px;
    transition: transform 0.25s ease;
    flex-shrink: 0;
  }

  &:hover {
    color: var(--color-accent-fg, #0969da);
    background: var(--color-canvas-subtle, #f6f8fa);
    border-color: var(--hairline, #e4e4e7);

    .reset-action-svg {
      transform: rotate(-45deg);
    }
  }
}

html.dark .top-preset-action-btn {
  color: #8b949e;

  &:hover {
    color: #58a6ff;
    background: #21262d;
    border-color: #30363d;
  }
}

.settings-nav-underline {
  display: flex;
  align-items: center;
  gap: 28px;
  height: 44px;
  user-select: none;
  overflow-x: auto;
  scrollbar-width: none;
  margin-bottom: -1px; /* Overlap border */
}

.settings-nav-underline::-webkit-scrollbar {
  display: none;
}

.nav-underline-item {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 44px;
  padding: 0 4px;
  background: transparent;
  border: none;
  font-family: var(--font-sans);
  font-size: 13.5px;
  font-weight: 500;
  color: var(--steel);
  cursor: pointer;
  white-space: nowrap;
  outline: none;
  transition: color 0.15s ease;
}

.nav-underline-item:hover {
  color: var(--charcoal);
}

.nav-underline-item.active {
  color: var(--ink);
  font-weight: 600;
}

.nav-underline-item.active::after {
  content: "";
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: var(--primary);
  border-radius: 2px 2px 0 0;
  box-shadow: 0 -1px 3px color-mix(in srgb, var(--primary) 30%, transparent);
}

.nav-underline-item.active .tab-svg-ico {
  color: var(--primary);
}

.nav-item-icon {
  display: flex;
  align-items: center;
  justify-content: center;
}

.tab-svg-ico {
  width: 15px;
  height: 15px;
  flex-shrink: 0;
  color: inherit;
  transition: color 0.15s ease;
}

.nav-item-text {
  line-height: 1;
}

.nav-tab-dirty-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #f59e0b;
  margin-left: 2px;
  flex-shrink: 0;
  box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.25);
  animation: livePulseWarning 2.4s infinite;
}

/* Settings Content Main Stream */
.settings-content {
  width: 100%;
  max-width: 1040px;
  margin: 0 auto;
  min-width: 0;
}

.settings-module-wrap {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Unified Section Card */
.settings-section-card {
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--hairline, #e4e4e7);
  border-radius: var(--r-md, 8px);
  overflow: hidden;
  box-shadow: var(--shadow-card);
  transition: box-shadow 0.2s ease, border-color 0.2s ease;
}

.settings-section-card:hover {
  box-shadow: var(--shadow-card-hover);
}

.section-card-header {
  padding: 16px 22px;
  border-bottom: 1px solid var(--hairline, #e4e4e7);
  background: transparent;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.section-header-title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.section-main-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--charcoal);
}

.section-header-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 8px;
  border-radius: var(--r-sm, 6px);
  background: transparent;
  border: none;
  color: var(--steel, #71717a);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
  line-height: 1;
}

.section-header-action-btn:hover {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.05));
  color: var(--charcoal, #27272a);
}

.reset-action-svg {
  width: 13px;
  height: 13px;
  color: var(--stone, #a1a1aa);
  transition: transform 0.25s ease, color 0.15s ease;
}

.section-header-action-btn:hover .reset-action-svg {
  transform: rotate(-45deg);
  color: var(--charcoal, #27272a);
}

html.dark .section-header-action-btn {
  border: none;
  color: #a1a1aa;
}

html.dark .section-header-action-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #f4f4f5;
}

html.dark .section-header-action-btn:hover .reset-action-svg {
  color: #f4f4f5;
}

.section-card-body {
  padding: 22px;
}

/* Clean Subtle Tab Fade Transition */
.tab-fade-enter-active,
.tab-fade-leave-active {
  transition: opacity 0.15s ease;
}

.tab-fade-enter-from,
.tab-fade-leave-to {
  opacity: 0;
}

/* Form sub-section styling inside cards */
.settings-section-card :deep(.form-section-card) {
  border: none !important;
  background: transparent !important;
  padding: 0 !important;
  margin: 0 !important;
  box-shadow: none !important;
}
.settings-section-card :deep(.form-section-card + .form-section-card) {
  margin-top: 28px !important;
  padding-top: 0 !important;
  border-top: none !important;
}

/* Floating Sticky Save Bar (吸底浮层保存栏：居中对齐主内容区域，宽度贴合设置卡片，整体风格与系统原生主题深度融合) */
.sticky-save-bar {
  position: fixed;
  bottom: 24px;
  right: 24px;
  left: 244px; /* 展开侧栏 220px + 24px 页面 padding */
  display: flex;
  justify-content: center;
  z-index: 1000;
  pointer-events: none;
}

:global(.layout-container.sidebar-collapsed) .sticky-save-bar {
  left: 88px; /* 折叠侧栏 64px + 24px 页面 padding */
}

@media (max-width: 768px) {
  .sticky-save-bar {
    left: 14px;
    right: 14px;
    bottom: 16px;
  }
}

.save-bar-inner {
  pointer-events: auto;
  width: 100%;
  max-width: 980px; /* 与设置卡片 max-width 完全贴合对齐 */
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--hairline, #e4e4e7);
  border-radius: 8px;
  padding: 10px 18px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.08), 0 2px 6px rgba(0, 0, 0, 0.04);
  gap: 16px;
  box-sizing: border-box;
}

html.dark .save-bar-inner {
  background: #161b22;
  border-color: #30363d;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.45), 0 2px 8px rgba(0, 0, 0, 0.3);
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
  background: #d97706;
  box-shadow: 0 0 0 0 rgba(217, 119, 6, 0.6);
  animation: pulse-warn-dot 1.8s infinite;
}

@keyframes pulse-warn-dot {
  0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(217, 119, 6, 0.6); }
  70% { transform: scale(1); box-shadow: 0 0 0 6px rgba(217, 119, 6, 0); }
  100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(217, 119, 6, 0); }
}

.save-bar-title {
  color: var(--charcoal, #24292f);
  font-size: 13px;
  font-weight: 600;
}

html.dark .save-bar-title {
  color: #f0f6fc;
}

.save-bar-sub {
  color: var(--stone, #71717a);
  font-size: 12px;
}

html.dark .save-bar-sub {
  color: #8b949e;
}

.save-bar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.btn-bar-ghost {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 5px 14px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 6px;
  color: var(--color-btn-text, #24292f);
  background: var(--color-btn-bg, #ffffff);
  border: 1px solid var(--color-btn-border, #d0d7de);
  box-shadow: 0 1px 0 rgba(27, 31, 36, 0.04);
  cursor: pointer;
  transition: all 0.15s ease;
  line-height: 20px;

  &:hover {
    background: var(--color-btn-hover-bg, #f3f4f6);
    border-color: #cf222e;
    color: #cf222e;
  }
}

html.dark .btn-bar-ghost {
  color: #c9d1d9;
  background: #21262d;
  border-color: #30363d;
  box-shadow: 0 1px 0 rgba(0, 0, 0, 0.1);

  &:hover {
    background: #30363d;
    border-color: #f87171;
    color: #f87171;
  }
}

.btn-bar-primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 5px 18px;
  font-size: 12px;
  font-weight: 600;
  border-radius: 6px;
  color: #ffffff;
  background: var(--color-accent-emphasis, #0969da);
  border: 1px solid rgba(27, 31, 36, 0.15);
  box-shadow: 0 1px 0 rgba(27, 31, 36, 0.1);
  cursor: pointer;
  transition: all 0.15s ease;
  line-height: 20px;

  &:hover:not(:disabled) {
    background: #0860ca;
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

html.dark .btn-bar-primary {
  background: #1f6feb;
  border-color: rgba(240, 246, 252, 0.1);

  &:hover:not(:disabled) {
    background: #388bfd;
  }
}

/* Animations */
.save-bar-slide-enter-active,
.save-bar-slide-leave-active {
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.save-bar-slide-enter-from,
.save-bar-slide-leave-to {
  opacity: 0;
  transform: translateY(16px);
}

/* Dark mode tweaks */
html.dark .nav-underline-item {
  color: var(--steel);
}
html.dark .nav-underline-item:hover {
  color: var(--charcoal);
}
html.dark .nav-underline-item.active {
  color: var(--ink);
}
html.dark .nav-underline-item.active::after {
  background: var(--primary);
}
html.dark .section-card-header {
  background: rgba(255, 255, 255, 0.02);
}

/* Preset Confirm Modal Dialog */
:deep(.preset-confirm-dialog) {
  border-radius: var(--r-md, 8px);
  overflow: hidden;

  .el-dialog__header {
    margin-right: 0;
    padding: 16px 20px 12px;
    border-bottom: 1px solid var(--hairline, #e4e4e7);
  }

  .el-dialog__title {
    font-size: 15px;
    font-weight: 600;
    color: var(--charcoal, #24292f);
  }

  .el-dialog__body {
    padding: 18px 20px 14px;
  }

  .el-dialog__footer {
    padding: 12px 20px 16px;
    border-top: 1px solid var(--hairline, #e4e4e7);
  }
}

html.dark :deep(.preset-confirm-dialog) {
  background: #161b22;
  border: 1px solid #30363d;

  .el-dialog__header {
    border-bottom-color: #30363d;
  }
  .el-dialog__title {
    color: #f0f6fc;
  }
  .el-dialog__footer {
    border-top-color: #30363d;
  }
}

.preset-dialog-desc {
  font-size: 13px;
  color: var(--color-fg-muted, #57606a);
  margin-bottom: 14px;
  line-height: 1.5;
}

html.dark .preset-dialog-desc {
  color: #8b949e;
}

.preset-summary-card {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.02));
  border: 1px solid var(--hairline, #e4e4e7);
  border-radius: 6px;
  padding: 12px 16px;
  margin-bottom: 14px;
}

html.dark .preset-summary-card {
  background: #0d1117;
  border-color: #30363d;
}

.preset-params-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.preset-param-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12.5px;
}

.preset-param-item .param-lbl {
  color: var(--color-fg-muted, #57606a);
}

html.dark .preset-param-item .param-lbl {
  color: #8b949e;
}

.preset-param-item .param-val {
  font-weight: 600;
  color: var(--charcoal, #24292f);
  font-family: var(--font-mono, monospace);
  font-size: 12px;
}

html.dark .preset-param-item .param-val {
  color: #c9d1d9;
}

.preset-dialog-hint {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 12px;
  color: var(--color-fg-muted, #57606a);
  line-height: 1.45;

  .hint-svg {
    width: 14px;
    height: 14px;
    flex-shrink: 0;
    margin-top: 1px;
    color: var(--color-accent-fg, #0969da);
  }
}

html.dark .preset-dialog-hint {
  color: #8b949e;

  .hint-svg {
    color: #58a6ff;
  }
}

.preset-dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;

  :deep(.el-button) {
    min-width: 68px;
    height: 32px;
    font-size: 13px;
  }
}
</style>
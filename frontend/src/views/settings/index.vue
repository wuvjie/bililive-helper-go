<template>
  <div class="settings-view">
    <!-- 极简单行顶栏：UnderlineNav + 右侧恢复预设按钮 -->
    <header class="settings-top-bar">
      <nav class="settings-underline-nav" role="tablist">
        <button
          v-for="item in navTabs"
          :key="item.id"
          type="button"
          role="tab"
          :aria-selected="activeTab === item.id"
          class="underline-nav-tab"
          :class="{ active: activeTab === item.id }"
          @click="switchTab(item.id)"
        >
          <span class="tab-label">{{ item.label }}</span>
          <span
            v-if="(item.id === 'policy' && isPolicyDirty) || (item.id === 'schedule' && isScheduleDirty)"
            class="tab-dirty-dot"
            title="包含未保存的修改"
          ></span>
        </button>
      </nav>

      <div class="top-bar-actions">
        <button
          type="button"
          class="btn-primer-outline"
          :title="activeTab === 'policy' ? '将录播合并与磁盘清理策略恢复为官方推荐预设' : '将自动化合并与磁盘巡检调度恢复为官方推荐预设'"
          @click="handleOpenPresetModal"
        >
          <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/>
            <path d="M3 3v5h5"/>
          </svg>
          <span>恢复推荐预设</span>
        </button>
      </div>
    </header>

    <!-- Main Content Stream -->
    <main class="settings-body" v-loading="loading">
      <Transition name="tab-fade" mode="out-in">
        <!-- 1. 全局策略 Tab -->
        <PolicyTab
          v-if="activeTab === 'policy'"
          key="policy"
          :config="config"
          :total-disk-gb="totalDiskGb"
          :saving="saving"
          @update:config="config = $event"
        />

        <!-- 2. 定时调度 Tab -->
        <ScheduleTab
          v-else-if="activeTab === 'schedule'"
          key="schedule"
          :form="scheduleForm"
          :backup-start="backupStart"
          :backup-end="backupEnd"
          :saving="saving"
          :loading="scheduleLoading"
          @update:form="Object.assign(scheduleForm, $event)"
          @update:backup-start="backupStart = $event"
          @update:backup-end="backupEnd = $event"
        />
      </Transition>
    </main>

    <!-- Floating Sticky Save Bar (吸底浮动保存栏 - 单一保存权威法则) -->
    <Transition name="save-bar-slide">
      <div v-if="isFormDirty" class="sticky-save-bar">
        <div class="save-bar-inner">
          <div class="save-bar-status">
            <span class="save-pulse-dot"></span>
            <span class="save-bar-text">检测到配置已修改，离开前请保存生效</span>
          </div>
          <div class="save-bar-actions">
            <button
              type="button"
              class="btn-save-discard"
              :disabled="saving"
              @click="handleDiscardChanges"
            >
              放弃更改
            </button>
            <button
              type="button"
              class="btn-save-commit"
              :disabled="saving"
              @click="handleSaveAll"
            >
              {{ saving ? '正在保存...' : '保存配置' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 恢复预设确认弹窗 -->
    <el-dialog
      v-model="presetModalVisible"
      :title="presetModalType === 'policy' ? '恢复策略推荐预设' : '恢复调度推荐预设'"
      width="460px"
      destroy-on-close
      class="preset-dialog"
    >
      <div class="preset-dialog-content">
        <p class="preset-intro-text">
          {{ presetModalType === 'policy'
            ? '将全局录播合并规则与磁盘清理水位恢复为官方推荐基准：'
            : '将后台自动化分片合并与磁盘巡检调度恢复为官方推荐基准：'
          }}
        </p>

        <div class="preset-values-box">
          <div v-if="presetModalType === 'policy'" class="preset-items-grid">
            <div class="preset-item"><span class="k">断流重连容差</span><span class="v">30 分钟</span></div>
            <div class="preset-item"><span class="k">完播静置等待</span><span class="v">30 分钟</span></div>
            <div class="preset-item"><span class="k">触发清理水位</span><span class="v">80% 触发 · 降至 60%</span></div>
            <div class="preset-item"><span class="k">免删保护底线</span><span class="v">保留最近 3 天录像</span></div>
            <div class="preset-item"><span class="k">每主播最少保留</span><span class="v">3 场</span></div>
            <div class="preset-item"><span class="k">单次最多清理</span><span class="v">50 个</span></div>
          </div>
          <div v-else class="preset-items-grid">
            <div class="preset-item"><span class="k">自动分片合并</span><span class="v">启用 · 每 360 分钟</span></div>
            <div class="preset-item"><span class="k">自动磁盘清理</span><span class="v">启用 · 每 720 分钟</span></div>
            <div class="preset-item"><span class="k">避让时间窗口</span><span class="v">全天候运行（无避让）</span></div>
          </div>
        </div>

        <p class="preset-notice-hint">
          预设参数载入后，可在表单中再次检查，点击吸底保存栏的「保存配置」正式写入系统。
        </p>
      </div>

      <template #footer>
        <div class="preset-dialog-actions">
          <el-button @click="presetModalVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmApplyPreset">载入预设</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onActivated } from "vue";
import { useRoute, useRouter, onBeforeRouteLeave } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { getConfig, saveConfig } from "@/api/config";
import { getSchedule, saveSchedule } from "@/api/schedule";
import { getStatusDetail } from "@/api/status";
import type { ConfigDTO, ScheduleStatus } from "@/api/types";
import PolicyTab from "./tabs/PolicyTab.vue";
import ScheduleTab, { type ScheduleForm } from "./tabs/ScheduleTab.vue";

const route = useRoute();
const router = useRouter();

const navTabs = [
  { id: "policy", label: "全局策略" },
  { id: "schedule", label: "定时调度" }
];

const activeTab = ref<string>(
  typeof route.query.tab === "string" && ["policy", "schedule"].includes(route.query.tab)
    ? (route.query.tab as string)
    : "policy"
);

function switchTab(tabId: string) {
  if (activeTab.value === tabId) return;
  activeTab.value = tabId;
  router.replace({ query: { ...route.query, tab: tabId } }).catch(() => {});
}

// 数据状态
const loading = ref(false);
const scheduleLoading = ref(false);
const saving = ref(false);

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

// 基准快照（用于高精度脏数据检测）
const initialPolicySnapshot = ref("");
const initialScheduleSnapshot = ref("");

function serializePolicy(c: ConfigDTO): string {
  if (!c || Object.keys(c).length === 0) return "";
  const wl = Array.isArray(c.WHITELIST_KEYWORDS)
    ? c.WHITELIST_KEYWORDS.slice().sort().join(",")
    : String(c.WHITELIST_KEYWORDS || "");
  return JSON.stringify({
    gap: Number(c.GAP_MINUTES ?? 30),
    mergeAge: Number(c.MERGE_AGE_MINUTES ?? 30),
    trigger: Number(c.TRIGGER_THRESHOLD ?? 80),
    target: Number(c.TARGET_THRESHOLD ?? 60),
    safeMode: c.SAFE_MODE || "days",
    safeDays: Number(c.SAFE_DAYS ?? 3),
    safeAgeMinutes: Number(c.SAFE_AGE_MINUTES ?? 120),
    minKeep: Number(c.MIN_KEEP_PER_STREAMER ?? 3),
    maxDelete: Number(c.MAX_DELETE_PER_RUN ?? 50),
    whitelist: wl
  });
}

function serializeSchedule(form: ScheduleForm, start: string, end: string): string {
  return JSON.stringify({
    merge_enabled: Boolean(form.merge_enabled),
    clean_enabled: Boolean(form.clean_enabled),
    merge_interval: Number(form.merge_interval || 360),
    clean_interval: Number(form.clean_interval || 720),
    start: start || "",
    end: end || ""
  });
}

const isPolicyDirty = computed(() => {
  return Boolean(initialPolicySnapshot.value && serializePolicy(config.value) !== initialPolicySnapshot.value);
});

const isScheduleDirty = computed(() => {
  return Boolean(
    initialScheduleSnapshot.value &&
    serializeSchedule(scheduleForm, backupStart.value, backupEnd.value) !== initialScheduleSnapshot.value
  );
});

const isFormDirty = computed(() => isPolicyDirty.value || isScheduleDirty.value);

// 数据加载
async function loadConfigData() {
  loading.value = true;
  try {
    const data = await getConfig();
    config.value = { ...data };
    initialPolicySnapshot.value = serializePolicy(data);
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error || err?.message || "获取全局策略失败");
  } finally {
    loading.value = false;
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

    const sh = data.backup_start_hour ?? data.backup?.start_hour;
    const sm = data.backup_start_minute ?? data.backup?.start_minute;
    if (sh !== undefined && sm !== undefined) {
      backupStart.value = `${String(sh).padStart(2, "0")}:${String(sm).padStart(2, "0")}`;
    } else {
      backupStart.value = "";
    }

    const eh = data.backup_end_hour ?? data.backup?.end_hour;
    const em = data.backup_end_minute ?? data.backup?.end_minute;
    if (eh !== undefined && em !== undefined) {
      backupEnd.value = `${String(eh).padStart(2, "0")}:${String(em).padStart(2, "0")}`;
    } else {
      backupEnd.value = "";
    }

    initialScheduleSnapshot.value = serializeSchedule(scheduleForm, backupStart.value, backupEnd.value);
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error || err?.message || "获取定时调度失败");
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
    // optional
  }
}

// 放弃更改
function handleDiscardChanges() {
  if (isPolicyDirty.value) {
    loadConfigData();
  }
  if (isScheduleDirty.value) {
    loadScheduleData();
  }
  ElMessage.info("已放弃未保存的配置更改");
}

// 保存配置
async function handleSaveAll() {
  if (!isFormDirty.value) return;
  saving.value = true;

  try {
    const tasks: Promise<unknown>[] = [];

    if (isPolicyDirty.value) {
      const payload = { ...config.value };
      if (typeof payload.WHITELIST_KEYWORDS === "string") {
        payload.WHITELIST_KEYWORDS = (payload.WHITELIST_KEYWORDS as string)
          .split(/[,，]/)
          .map(s => s.trim())
          .filter(Boolean);
      }
      tasks.push(saveConfig(payload));
    }

    if (isScheduleDirty.value) {
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
      tasks.push(saveSchedule(payload));
    }

    await Promise.all(tasks);
    ElMessage.success("系统配置已成功保存生效");

    if (isPolicyDirty.value) {
      initialPolicySnapshot.value = serializePolicy(config.value);
    }
    if (isScheduleDirty.value) {
      initialScheduleSnapshot.value = serializeSchedule(scheduleForm, backupStart.value, backupEnd.value);
    }
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error || err?.message || "保存系统配置失败");
  } finally {
    saving.value = false;
  }
}

// 离开路由守卫
onBeforeRouteLeave((to, from, next) => {
  if (isFormDirty.value) {
    ElMessageBox.confirm(
      "当前设置页面存在尚未保存的配置变更，离开后将丢失修改，是否确认离开？",
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

// 恢复预设弹窗
const presetModalVisible = ref(false);
const presetModalType = ref<"policy" | "schedule">("policy");

const RECOMMENDED_POLICY: Partial<ConfigDTO> = {
  GAP_MINUTES: 30,
  MERGE_AGE_MINUTES: 30,
  TRIGGER_THRESHOLD: 80,
  TARGET_THRESHOLD: 60,
  SAFE_MODE: "days",
  SAFE_DAYS: 3,
  SAFE_AGE_MINUTES: 1440 * 3,
  MIN_KEEP_PER_STREAMER: 3,
  MAX_DELETE_PER_RUN: 50,
  WHITELIST_KEYWORDS: ["留存", "纪念", "高能", "生日", "勿删"]
};

const RECOMMENDED_SCHEDULE: ScheduleForm = {
  merge_enabled: true,
  clean_enabled: true,
  merge_interval: 360,
  clean_interval: 720
};

function handleOpenPresetModal() {
  presetModalType.value = activeTab.value as "policy" | "schedule";
  presetModalVisible.value = true;
}

function confirmApplyPreset() {
  if (presetModalType.value === "policy") {
    config.value = {
      ...config.value,
      ...RECOMMENDED_POLICY
    } as ConfigDTO;
    ElMessage.success("已载入官方推荐策略预设，请检查并点击下方「保存配置」");
  } else {
    Object.assign(scheduleForm, RECOMMENDED_SCHEDULE);
    backupStart.value = "";
    backupEnd.value = "";
    ElMessage.success("已载入官方推荐调度预设，请检查并点击下方「保存配置」");
  }
  presetModalVisible.value = false;
}

onMounted(() => {
  loadConfigData();
  loadScheduleData();
  loadDiskStatus();
});

onActivated(() => {
  loadConfigData();
  loadScheduleData();
  loadDiskStatus();
});
</script>

<style scoped>
.settings-view {
  max-width: 960px;
  margin: 0 auto;
  padding: 0 16px 80px;
  position: relative;
}

/* 单行极简顶栏：UnderlineNav + 右侧操作 */
.settings-top-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 0 0;
  margin-bottom: 20px;
  border-bottom: 1px solid var(--hairline, #e5e3df);
}

.settings-underline-nav {
  display: flex;
  gap: 8px;
  margin-bottom: -1px;
}

.underline-nav-tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--steel, #787671);
  font-size: 13.5px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
  line-height: 1.5;
}

.underline-nav-tab:hover {
  color: var(--ink, #1a1a1a);
  border-bottom-color: var(--hairline-strong, #c8c4be);
}

.underline-nav-tab.active {
  color: var(--primary, #2563eb);
  font-weight: 600;
  border-bottom-color: var(--primary, #2563eb);
}

.tab-dirty-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--warning, #d97706);
}

.top-bar-actions {
  flex-shrink: 0;
  padding-bottom: 8px;
}

.btn-primer-outline {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 30px;
  padding: 0 10px;
  border-radius: var(--r-sm, 4px);
  border: 1px solid var(--hairline-strong, #c8c4be);
  background: var(--surface-card, #ffffff);
  color: var(--charcoal, #37352f);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
  box-shadow: var(--shadow-xs);
}

.btn-primer-outline:hover {
  border-color: var(--primary, #2563eb);
  color: var(--primary, #2563eb);
  background: color-mix(in srgb, var(--primary, #2563eb) 6%, transparent);
}

.btn-svg-sm {
  width: 13px;
  height: 13px;
}

/* Settings Content Transition */
.settings-body {
  min-height: 300px;
}

.tab-fade-enter-active,
.tab-fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.tab-fade-enter-from {
  opacity: 0;
  transform: translateY(4px);
}

.tab-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* Floating Sticky Save Bar */
.sticky-save-bar {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 1000;
  width: calc(100% - 32px);
  max-width: 660px;
}

.save-bar-slide-enter-active,
.save-bar-slide-leave-active {
  transition: transform 0.22s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.22s ease;
}

.save-bar-slide-enter-from,
.save-bar-slide-leave-to {
  opacity: 0;
  transform: translate(-50%, 20px);
}

.save-bar-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--hairline-strong, #c8c4be);
  border-radius: var(--r-xl, 14px);
  padding: 10px 16px 10px 20px;
  box-shadow: var(--shadow-popover, 0 8px 24px rgba(0, 0, 0, 0.12));
  backdrop-filter: var(--glass-blur);
}

.save-bar-status {
  display: flex;
  align-items: center;
  gap: 10px;
}

.save-pulse-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--warning, #d97706);
  box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.25);
  animation: livePulseWarning 2s infinite;
}

.save-bar-text {
  font-size: 13px;
  font-weight: 500;
  color: var(--charcoal, #37352f);
}

.save-bar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.btn-save-discard {
  height: 32px;
  padding: 0 12px;
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--r-sm, 4px);
  color: var(--slate, #5d5b54);
  font-size: 12.5px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-save-discard:hover {
  color: var(--error, #e03131);
  background: color-mix(in srgb, var(--error, #e03131) 8%, transparent);
}

.btn-save-commit {
  height: 32px;
  padding: 0 16px;
  background: var(--primary, #2563eb);
  border: 1px solid var(--primary, #2563eb);
  border-radius: var(--r-sm, 4px);
  color: var(--on-primary, #ffffff);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
  box-shadow: var(--shadow-xs);
}

.btn-save-commit:hover {
  background: var(--primary-hover, #1d4ed8);
  border-color: var(--primary-hover, #1d4ed8);
}

.btn-save-commit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Preset Dialog */
.preset-dialog-content {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.preset-intro-text {
  font-size: 13px;
  color: var(--charcoal, #37352f);
  margin: 0;
  line-height: 1.45;
}

.preset-values-box {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.02));
  border: 1px solid var(--hairline, #e5e3df);
  border-radius: var(--r-md, 6px);
  padding: 12px 16px;
}

.preset-items-grid {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.preset-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12.5px;
}

.preset-item .k {
  color: var(--steel, #787671);
}

.preset-item .v {
  font-weight: 600;
  color: var(--ink, #1a1a1a);
}

.preset-notice-hint {
  font-size: 12px;
  color: var(--steel, #787671);
  margin: 0;
  line-height: 1.4;
}

.preset-dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 640px) {
  .settings-top-bar {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  .sticky-save-bar {
    width: calc(100% - 20px);
  }
  .save-bar-inner {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
    padding: 12px 16px;
  }
  .save-bar-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>

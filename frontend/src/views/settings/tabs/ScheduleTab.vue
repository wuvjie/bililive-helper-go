<template>
  <div class="settings-tab-pane">
    <!-- Box 1: 自动化定时调度任务 -->
    <section class="primer-box">
      <header class="primer-box-header">
        <div class="box-header-title-row">
          <svg class="box-header-icon icon-blue" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
          <h2 class="box-title">自动化定时调度任务</h2>
        </div>
      </header>

      <div class="primer-box-body">
        <div class="schedule-task-grid">
          <!-- 1. 自动合并视频分片 -->
          <div class="schedule-task-card" :class="{ 'task-disabled': !form.merge_enabled }">
            <div class="task-card-header">
              <div class="task-card-title-group">
                <span class="task-card-title">自动合并视频分片</span>
                <span class="status-pill" :class="form.merge_enabled ? 'pill-active' : 'pill-inactive'">
                  {{ form.merge_enabled ? '已启用' : '已停用' }}
                </span>
              </div>
              <el-switch
                :model-value="form.merge_enabled"
                class="primer-switch"
                @update:model-value="updateField('merge_enabled', Boolean($event))"
              />
            </div>
            <div class="task-card-body">
              <label class="field-label-sm">任务执行周期</label>
              <div class="interval-input-row">
                <el-input-number
                  :model-value="form.merge_interval"
                  :min="10"
                  :max="10080"
                  :step="30"
                  :disabled="!form.merge_enabled"
                  size="default"
                  class="stepper-compact"
                  @update:model-value="updateField('merge_interval', Number($event))"
                />
                <span class="unit-text">分钟</span>
                <span class="interval-hint mono-val">（每 {{ (form.merge_interval / 60).toFixed(1) }} 小时）</span>
              </div>
            </div>
          </div>

          <!-- 2. 自动清理磁盘空间 -->
          <div class="schedule-task-card" :class="{ 'task-disabled': !form.clean_enabled }">
            <div class="task-card-header">
              <div class="task-card-title-group">
                <span class="task-card-title">自动清理磁盘空间</span>
                <span class="status-pill" :class="form.clean_enabled ? 'pill-active' : 'pill-inactive'">
                  {{ form.clean_enabled ? '已启用' : '已停用' }}
                </span>
              </div>
              <el-switch
                :model-value="form.clean_enabled"
                class="primer-switch"
                @update:model-value="updateField('clean_enabled', Boolean($event))"
              />
            </div>
            <div class="task-card-body">
              <label class="field-label-sm">任务执行周期</label>
              <div class="interval-input-row">
                <el-input-number
                  :model-value="form.clean_interval"
                  :min="10"
                  :max="10080"
                  :step="60"
                  :disabled="!form.clean_enabled"
                  size="default"
                  class="stepper-compact"
                  @update:model-value="updateField('clean_interval', Number($event))"
                />
                <span class="unit-text">分钟</span>
                <span class="interval-hint mono-val">（每 {{ (form.clean_interval / 60).toFixed(1) }} 小时）</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Box 2: 定时任务避让时段 -->
    <section class="primer-box">
      <header class="primer-box-header">
        <div class="box-header-title-row">
          <svg class="box-header-icon icon-purple" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
          </svg>
          <div class="box-header-title-group">
            <h2 class="box-title">定时任务避让时段</h2>
            <span class="status-pill" :class="isQuietEnabled ? 'pill-active' : 'pill-inactive'">
              {{ isQuietEnabled ? '已启用' : '已停用' }}
            </span>
          </div>
          <el-switch
            :model-value="isQuietEnabled"
            class="primer-switch"
            @update:model-value="toggleQuiet(Boolean($event))"
          />
        </div>
      </header>

      <div class="primer-box-body">
        <!-- 未启用状态：直接干脆展示全天候说明，不显示虚假的 00:00 与状态呼吸灯 -->
        <div v-if="!isQuietEnabled" class="quiet-disabled-notice">
          未开启避让时段，所有自动化任务全天候 24 小时正常调度运行。
        </div>

        <!-- 已启用状态：对称双列网格（左侧时间范围，右侧实时运行状态） -->
        <div v-else class="quiet-window-grid">
          <!-- 1. 避让时间范围选择 -->
          <div class="quiet-field-cell">
            <label class="field-label-sm">每日避让时段范围</label>
            <div class="window-picker-flow">
              <div class="time-range-group">
                <el-time-picker
                  :model-value="backupStart"
                  format="HH:mm"
                  value-format="HH:mm"
                  placeholder="开始时间"
                  size="default"
                  class="time-picker-item"
                  @update:model-value="onStartTimeChange"
                />
                <span class="range-separator">至</span>
                <el-time-picker
                  :model-value="backupEnd"
                  format="HH:mm"
                  value-format="HH:mm"
                  placeholder="结束时间"
                  size="default"
                  class="time-picker-item"
                  @update:model-value="onEndTimeChange"
                />
              </div>
            </div>
          </div>

          <!-- 2. 实时窗口运行状态指示 -->
          <div class="quiet-field-cell">
            <label class="field-label-sm">当前实时运行状态</label>
            <div class="window-status-card" :class="isQuietActive ? 'status-card-quiet' : 'status-card-active'">
              <span class="status-pulse-dot" :class="isQuietActive ? 'pulse-amber' : 'pulse-emerald'"></span>
              <span class="status-main-text">{{ windowStatusTitle }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";

export interface ScheduleForm {
  merge_enabled: boolean;
  clean_enabled: boolean;
  merge_interval: number;
  clean_interval: number;
}

const props = defineProps<{
  form: ScheduleForm;
  backupStart: string;
  backupEnd: string;
  saving?: boolean;
  loading?: boolean;
}>();

const emit = defineEmits<{
  (e: "update:form", val: ScheduleForm): void;
  (e: "update:backupStart", val: string): void;
  (e: "update:backupEnd", val: string): void;
}>();

function updateField<K extends keyof ScheduleForm>(key: K, value: ScheduleForm[K]) {
  emit("update:form", {
    ...props.form,
    [key]: value
  });
}

const isQuietEnabled = computed(() => {
  return Boolean(props.backupStart && props.backupEnd && props.backupStart !== props.backupEnd);
});

const lastValidStart = ref(props.backupStart || "02:00");
const lastValidEnd = ref(props.backupEnd || "06:00");

function toggleQuiet(enabled: boolean) {
  if (enabled) {
    const s = lastValidStart.value && lastValidEnd.value && lastValidStart.value !== lastValidEnd.value
      ? lastValidStart.value
      : "02:00";
    const e = lastValidStart.value && lastValidEnd.value && lastValidStart.value !== lastValidEnd.value
      ? lastValidEnd.value
      : "06:00";
    emit("update:backupStart", s);
    emit("update:backupEnd", e);
  } else {
    if (props.backupStart && props.backupEnd && props.backupStart !== props.backupEnd) {
      lastValidStart.value = props.backupStart;
      lastValidEnd.value = props.backupEnd;
    }
    emit("update:backupStart", "");
    emit("update:backupEnd", "");
  }
}

function onStartTimeChange(val: string | null) {
  const newStart = val || "";
  emit("update:backupStart", newStart);
  if (newStart && props.backupEnd && newStart !== props.backupEnd) {
    lastValidStart.value = newStart;
  }
}

function onEndTimeChange(val: string | null) {
  const newEnd = val || "";
  emit("update:backupEnd", newEnd);
  if (props.backupStart && newEnd && props.backupStart !== newEnd) {
    lastValidEnd.value = newEnd;
  }
}

// 避让窗口状态计算
const isQuietActive = computed(() => {
  if (!isQuietEnabled.value) return false;
  const now = new Date();
  const currentMinutes = now.getHours() * 60 + now.getMinutes();

  const [sh, sm] = props.backupStart.split(":").map(Number);
  const [eh, em] = props.backupEnd.split(":").map(Number);
  const startMinutes = sh * 60 + sm;
  const endMinutes = eh * 60 + em;

  if (startMinutes <= endMinutes) {
    return currentMinutes >= startMinutes && currentMinutes < endMinutes;
  }
  // 跨午夜
  return currentMinutes >= startMinutes || currentMinutes < endMinutes;
});

const windowStatusTitle = computed(() => {
  if (isQuietActive.value) {
    return `当前处于避让休眠时段（${props.backupStart} ~ ${props.backupEnd}），暂停后台合并与清理`;
  }
  return `当前处于正常运行窗口（每日避让时段: ${props.backupStart} ~ ${props.backupEnd}）`;
});
</script>

<style scoped>
.settings-tab-pane {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* GitHub Primer Box Model */
.primer-box {
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--hairline, #e5e3df);
  border-radius: var(--r-md, 6px);
  overflow: hidden;
  box-shadow: var(--shadow-xs);
}

.primer-box-header {
  padding: 12px 18px;
  background: var(--surface-secondary, rgba(0, 0, 0, 0.02));
  border-bottom: 1px solid var(--hairline, #e5e3df);
}

.box-header-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.box-header-title-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.box-header-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.icon-blue {
  color: var(--primary, #2563eb);
}

.icon-purple {
  color: #8b5cf6;
}

.box-title {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--ink, #1a1a1a);
  margin: 0;
  line-height: 1.3;
}

.primer-box-body {
  padding: 0;
}

/* 定时任务对称双列网格 */
.schedule-task-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  padding: 16px 18px;
}

.schedule-task-card {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.015));
  border: 1px solid var(--hairline, #e5e3df);
  border-radius: var(--r-md, 6px);
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  transition: all 0.15s ease;
}

.schedule-task-card.task-disabled {
  opacity: 0.6;
  background: var(--surface, #fbfbfa);
}

.task-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.primer-switch {
  flex-shrink: 0;
  margin-left: auto;
}

.task-card-title-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.task-card-title {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal, #37352f);
}

.status-pill {
  font-size: 11px;
  font-weight: 500;
  padding: 1px 7px;
  border-radius: var(--r-full, 9999px);
  line-height: 1.3;
}

.pill-active {
  background: rgba(16, 185, 129, 0.12);
  color: var(--success, #059669);
  border: 1px solid rgba(16, 185, 129, 0.25);
}

.pill-inactive {
  background: var(--surface, #f4f3f0);
  color: var(--stone, #a4a097);
  border: 1px solid var(--hairline, #e5e3df);
}

.task-card-body {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label-sm {
  font-size: 12px;
  color: var(--steel, #787671);
  font-weight: 500;
}

.interval-input-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.stepper-compact {
  width: 120px;
}

.unit-text {
  font-size: 13px;
  color: var(--slate, #5d5b54);
  font-weight: 500;
}

.interval-hint {
  font-size: 12px;
  color: var(--steel, #787671);
  white-space: nowrap;
  display: inline-block;
  min-width: 96px;
  font-variant-numeric: tabular-nums;
  font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace;
}

/* 避让时间窗口网格 */
.quiet-window-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  padding: 16px 18px;
  align-items: start;
}

.quiet-field-cell {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-label {
  font-size: 13.5px;
  font-weight: 500;
  color: var(--charcoal, #37352f);
  line-height: 1.35;
}

.window-picker-flow {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.time-range-group {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.time-picker-item {
  width: 115px !important;
}

.range-separator {
  font-size: 12px;
  color: var(--steel, #787671);
}

.quiet-disabled-notice {
  padding: 14px 18px;
  font-size: 13px;
  color: var(--steel, #787671);
  line-height: 1.5;
}

.window-status-card {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  border-radius: var(--r-sm, 4px);
  border: 1px solid transparent;
}

.status-card-active {
  background: rgba(16, 185, 129, 0.08);
  border-color: rgba(16, 185, 129, 0.22);
}

.status-card-quiet {
  background: rgba(245, 158, 11, 0.08);
  border-color: rgba(245, 158, 11, 0.22);
}

.status-pulse-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.pulse-emerald {
  background: var(--success, #059669);
  box-shadow: 0 0 0 2px rgba(16, 185, 129, 0.2);
}

.pulse-amber {
  background: var(--warning, #d97706);
  box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.2);
}

.status-main-text {
  font-size: 12.5px;
  font-weight: 500;
  color: var(--charcoal, #37352f);
}

@media (max-width: 768px) {
  .schedule-task-grid,
  .quiet-window-grid {
    grid-template-columns: 1fr;
    gap: 14px;
  }
}
</style>

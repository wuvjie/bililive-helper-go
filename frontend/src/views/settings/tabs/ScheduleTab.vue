<template>
  <div class="settings-tab-pane">
    <!-- Box 1: 自动化媒体流水线调度 -->
    <section class="primer-box">
      <header class="primer-box-header">
        <div class="box-header-title-row">
          <svg class="box-header-icon icon-blue" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
          <div class="box-header-text">
            <h2 class="box-title">定时流水线调度</h2>
            <p class="box-desc">配置后台自动化巡检扫描周期，按需静默执行视频分片合并与磁盘空间释放</p>
          </div>
        </div>
      </header>

      <div class="primer-box-body">
        <!-- 1. 自动分片合并 -->
        <div class="primer-row">
          <div class="row-info">
            <div class="row-label-row">
              <label class="row-label">自动分片合并</label>
              <span class="status-pill" :class="form.merge_enabled ? 'pill-active' : 'pill-inactive'">
                {{ form.merge_enabled ? '已启用' : '已停用' }}
              </span>
            </div>
            <p class="row-desc">后台定时扫描主播录像目录，将符合静置容差的已完播切片无损合并为完整视频</p>
          </div>
          <div class="row-control">
            <div class="schedule-control-flow">
              <el-switch
                :model-value="form.merge_enabled"
                class="primer-switch"
                @update:model-value="updateField('merge_enabled', Boolean($event))"
              />
              <div v-if="form.merge_enabled" class="control-unit-group">
                <el-input-number
                  :model-value="form.merge_interval"
                  :min="10"
                  :max="10080"
                  :step="30"
                  size="default"
                  class="stepper-compact"
                  @update:model-value="updateField('merge_interval', Number($event))"
                />
                <span class="unit-text">分钟</span>
                <span class="interval-hint mono-val">（每 {{ (form.merge_interval / 60).toFixed(1) }} 小时）</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 2. 自动磁盘清理 -->
        <div class="primer-row">
          <div class="row-info">
            <div class="row-label-row">
              <label class="row-label">自动磁盘清理</label>
              <span class="status-pill" :class="form.clean_enabled ? 'pill-active' : 'pill-inactive'">
                {{ form.clean_enabled ? '已启用' : '已停用' }}
              </span>
            </div>
            <p class="row-desc">后台定时检查磁盘水位深度，并在达到触发阈值时按策略自动淘汰旧录像</p>
          </div>
          <div class="row-control">
            <div class="schedule-control-flow">
              <el-switch
                :model-value="form.clean_enabled"
                class="primer-switch"
                @update:model-value="updateField('clean_enabled', Boolean($event))"
              />
              <div v-if="form.clean_enabled" class="control-unit-group">
                <el-input-number
                  :model-value="form.clean_interval"
                  :min="10"
                  :max="10080"
                  :step="60"
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

    <!-- Box 2: 避让时间窗口 (免打扰时段) -->
    <section class="primer-box">
      <header class="primer-box-header">
        <div class="box-header-title-row">
          <svg class="box-header-icon icon-purple" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
          </svg>
          <div class="box-header-text">
            <h2 class="box-title">避让时间窗口 (免打扰时段)</h2>
            <p class="box-desc">在设定的时段内自动暂停后台高 I/O 任务，避让业务高峰或夜间数据冷备</p>
          </div>
        </div>
      </header>

      <div class="primer-box-body">
        <!-- 1. 避让时间范围选择 -->
        <div class="primer-row">
          <div class="row-info">
            <label class="row-label">避让时间范围</label>
            <p class="row-desc">在此时间段内定时合并与磁盘清理将自动挂起延后；留空表示全天候运行（无避让）</p>
          </div>
          <div class="row-control">
            <div class="window-picker-flow">
              <div class="time-range-group">
                <el-time-picker
                  :model-value="backupStart"
                  format="HH:mm"
                  value-format="HH:mm"
                  placeholder="开始时间"
                  size="default"
                  class="time-picker-item"
                  @update:model-value="$emit('update:backupStart', $event || '')"
                />
                <span class="range-separator">至</span>
                <el-time-picker
                  :model-value="backupEnd"
                  format="HH:mm"
                  value-format="HH:mm"
                  placeholder="结束时间"
                  size="default"
                  class="time-picker-item"
                  @update:model-value="$emit('update:backupEnd', $event || '')"
                />
              </div>

              <button
                v-if="backupStart || backupEnd"
                type="button"
                class="btn-reset-window"
                title="清除避让时段，恢复全天候正常调度"
                @click="clearQuietWindow"
              >
                <span>恢复全天候运行</span>
              </button>
            </div>
          </div>
        </div>

        <!-- 2. 实时窗口运行状态指示 -->
        <div class="primer-row window-status-row">
          <div class="window-status-card" :class="isQuietActive ? 'status-card-quiet' : 'status-card-active'">
            <div class="status-icon-col">
              <span class="status-pulse-dot" :class="isQuietActive ? 'pulse-amber' : 'pulse-emerald'"></span>
            </div>
            <div class="status-text-col">
              <span class="status-main-text">{{ windowStatusTitle }}</span>
              <span class="status-sub-text">{{ windowStatusDesc }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

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

function clearQuietWindow() {
  emit("update:backupStart", "");
  emit("update:backupEnd", "");
}

// 避让窗口状态计算
const isQuietActive = computed(() => {
  if (!props.backupStart || !props.backupEnd) return false;
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
  if (!props.backupStart || !props.backupEnd) {
    return "全天候 24 小时正常调度运行中";
  }
  if (isQuietActive.value) {
    return "当前处于避让休眠时段（自动化任务已自动顺延挂起）";
  }
  return "当前处于正常调度窗口（后台任务按周期正规运行）";
});

const windowStatusDesc = computed(() => {
  if (!props.backupStart || !props.backupEnd) {
    return "未设定避让时间窗口，后台合并与清理任务全天候按各自设定周期自动执行";
  }
  if (isQuietActive.value) {
    return `当前时间命中避让窗口（${props.backupStart} ~ ${props.backupEnd}），所有自动化高 I/O 操作已挂起，到达 ${props.backupEnd} 后自动恢复调度。`;
  }
  return `设定避让窗口为每日 ${props.backupStart} ~ ${props.backupEnd}，在进入该时段前后台自动化任务正常调度。`;
});
</script>

<style scoped>
.settings-tab-pane {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* GitHub Primer Box Model */
.primer-box {
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--hairline, #e5e3df);
  border-radius: var(--r-md, 6px);
  overflow: hidden;
  box-shadow: var(--shadow-xs);
  transition: border-color 0.15s ease;
}

.primer-box-header {
  padding: 14px 20px;
  background: var(--surface-secondary, rgba(0, 0, 0, 0.02));
  border-bottom: 1px solid var(--hairline, #e5e3df);
}

.box-header-title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.box-header-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.icon-blue {
  color: var(--primary, #2563eb);
}

.icon-purple {
  color: #8b5cf6;
}

.box-header-text {
  flex: 1;
  min-width: 0;
}

.box-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--ink, #1a1a1a);
  margin: 0 0 2px 0;
  line-height: 1.3;
}

.box-desc {
  font-size: 12px;
  color: var(--steel, #787671);
  margin: 0;
  line-height: 1.4;
}

/* Primer Rows */
.primer-box-body {
  padding: 0;
}

.primer-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--hairline, #e5e3df);
  gap: 24px;
}

.primer-row:last-child {
  border-bottom: none;
}

.row-info {
  flex: 1;
  min-width: 0;
}

.row-label-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 3px;
}

.row-label {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal, #37352f);
  line-height: 1.35;
}

.row-desc {
  font-size: 12px;
  color: var(--steel, #787671);
  margin: 0;
  line-height: 1.45;
}

.row-control {
  flex-shrink: 0;
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

.schedule-control-flow {
  display: inline-flex;
  align-items: center;
  gap: 14px;
}

.control-unit-group {
  display: inline-flex;
  align-items: center;
  gap: 6px;
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
}

/* 避让窗口样式 */
.window-picker-flow {
  display: flex;
  align-items: center;
  gap: 12px;
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

.btn-reset-window {
  display: inline-flex;
  align-items: center;
  height: 28px;
  padding: 0 10px;
  border-radius: var(--r-sm, 4px);
  border: 1px solid var(--hairline-strong, #c8c4be);
  background: transparent;
  color: var(--slate, #5d5b54);
  font-size: 11.5px;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.btn-reset-window:hover {
  border-color: var(--primary, #2563eb);
  color: var(--primary, #2563eb);
  background: color-mix(in srgb, var(--primary, #2563eb) 6%, transparent);
}

.window-status-row {
  display: block;
  padding: 14px 20px;
}

.window-status-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 16px;
  border-radius: var(--r-md, 6px);
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

.status-icon-col {
  padding-top: 3px;
}

.status-pulse-dot {
  display: block;
  width: 8px;
  height: 8px;
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

.status-text-col {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.status-main-text {
  font-size: 13px;
  font-weight: 600;
  color: var(--charcoal, #37352f);
}

.status-sub-text {
  font-size: 12px;
  color: var(--steel, #787671);
  line-height: 1.4;
}

@media (max-width: 640px) {
  .primer-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  .schedule-control-flow {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  .window-picker-flow {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
}
</style>

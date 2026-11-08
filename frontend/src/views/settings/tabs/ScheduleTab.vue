<template>
  <div class="schedule-tab" v-loading="loading">
    <!-- Card 1: 自动化任务调度计划 (双列并行对称网格) -->
    <div class="setting-group-card">
      <div class="group-card-header">
        <div class="group-header-text">
          <h3 class="group-title">自动化任务调度计划</h3>
          <p class="group-desc">配置录像分片合并与磁盘巡检清理的主定时调度器，后台周期性轮询执行。</p>
        </div>
      </div>
      <div class="group-card-body">
        <div class="schedule-tasks-grid">
          <!-- Task 1: 自动分片合并 -->
          <div class="task-tile-card" :class="{ 'tile-disabled': !form.merge_enabled }">
            <div class="tile-card-header">
              <div class="tile-header-left">
                <svg class="tile-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"/>
                  <line x1="7" y1="2" x2="7" y2="22"/>
                  <line x1="17" y1="2" x2="17" y2="22"/>
                  <line x1="2" y1="12" x2="22" y2="12"/>
                  <line x1="2" y1="7" x2="7" y2="7"/>
                  <line x1="2" y1="17" x2="7" y2="17"/>
                  <line x1="17" y1="17" x2="22" y2="17"/>
                  <line x1="17" y1="7" x2="22" y2="7"/>
                </svg>
                <span class="tile-title">自动分片合并</span>
              </div>
              <el-switch
                :model-value="form.merge_enabled"
                @update:model-value="updateFormField('merge_enabled', !!$event)"
              />
            </div>

            <div class="tile-card-body">
              <div class="tile-field-row">
                <label class="tile-field-label">
                  <span>合并调度周期</span>
                </label>
                <div class="inline-unit-row">
                  <el-input-number
                    :model-value="form.merge_interval"
                    :min="10"
                    :max="1440"
                    :step="15"
                    class="compact-counter"
                    :disabled="!form.merge_enabled"
                    @update:model-value="updateFormField('merge_interval', Number($event))"
                  />
                  <span class="unit-text">分钟</span>
                </div>
              </div>

              <div class="tile-hint-box">
                <svg class="hint-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <line x1="12" y1="16" x2="12" y2="12"/>
                  <line x1="12" y1="8" x2="12.01" y2="8"/>
                </svg>
                <span>推荐 360 分钟（6小时），任务执行完毕后休眠进入下次循环。</span>
              </div>
            </div>
          </div>

          <!-- Task 2: 自动磁盘清理 -->
          <div class="task-tile-card" :class="{ 'tile-disabled': !form.clean_enabled }">
            <div class="tile-card-header">
              <div class="tile-header-left">
                <svg class="tile-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="3 6 5 6 21 6"/>
                  <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                </svg>
                <span class="tile-title">自动磁盘清理</span>
              </div>
              <el-switch
                :model-value="form.clean_enabled"
                @update:model-value="updateFormField('clean_enabled', !!$event)"
              />
            </div>

            <div class="tile-card-body">
              <div class="tile-field-row">
                <label class="tile-field-label">
                  <span>巡检调度周期</span>
                </label>
                <div class="inline-unit-row">
                  <el-input-number
                    :model-value="form.clean_interval"
                    :min="10"
                    :max="1440"
                    :step="30"
                    class="compact-counter"
                    :disabled="!form.clean_enabled"
                    @update:model-value="updateFormField('clean_interval', Number($event))"
                  />
                  <span class="unit-text">分钟</span>
                </div>
              </div>

              <div class="tile-hint-box">
                <svg class="hint-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <line x1="12" y1="16" x2="12" y2="12"/>
                  <line x1="12" y1="8" x2="12.01" y2="8"/>
                </svg>
                <span>推荐 720 分钟（12小时），定期比对磁盘剩余水位并触发清理。</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Card 2: 免打扰避让时段 -->
    <div class="setting-group-card">
      <div class="group-card-header">
        <div class="group-header-text">
          <h3 class="group-title">免打扰避让时段</h3>
          <p class="group-desc">在此设定时段内自动暂停合并与清理，其余时段恢复正常计划调度（支持跨午夜）。</p>
        </div>
      </div>
      <div class="group-card-body">
        <div class="setting-row">
          <div class="setting-info">
            <div class="setting-title">
              <span>避让时间范围</span>
            </div>
          </div>
          <div class="setting-control">
            <div class="time-range-row">
              <el-time-picker
                :model-value="backupStart"
                format="HH:mm"
                value-format="HH:mm"
                placeholder="开始时间"
                class="time-picker-box"
                @update:model-value="emit('update:backupStart', $event ?? '')"
              />
              <span class="time-sep">至</span>
              <el-time-picker
                :model-value="backupEnd"
                format="HH:mm"
                value-format="HH:mm"
                placeholder="结束时间"
                class="time-picker-box"
                @update:model-value="emit('update:backupEnd', $event ?? '')"
              />
            </div>
          </div>
        </div>

        <!-- 避让窗口实时状态提示条 -->
        <div v-if="backupStart && backupEnd" class="schedule-status-banner">
          <div class="status-banner-left">
            <svg class="status-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <polyline points="12 6 12 12 16 14"/>
            </svg>
            <span class="status-text">
              每日 <strong>{{ backupStart }}</strong> 至 <strong>{{ backupEnd }}</strong> 处于静默避让窗口
            </span>
          </div>
          <span class="status-badge">定时避让生效中</span>
        </div>
        <div v-else class="schedule-status-banner status-banner-neutral">
          <div class="status-banner-left">
            <svg class="status-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <line x1="12" y1="8" x2="12" y2="12"/>
              <line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
            <span class="status-text">全天候调度运行中（未设置避让时段）</span>
          </div>
          <span class="status-badge-neutral">24h 持续作业</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
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
  loading: boolean;
}>();

const emit = defineEmits<{
  "update:form": [value: ScheduleForm];
  "update:backupStart": [value: string];
  "update:backupEnd": [value: string];
  save?: [];
}>();

function updateFormField<K extends keyof ScheduleForm>(field: K, val: ScheduleForm[K]) {
  emit("update:form", {
    ...props.form,
    [field]: val,
  });
}
</script>

<style scoped>
.schedule-tab {
  display: flex;
  flex-direction: column;
}

/* Unified Group Card (GitHub Primer Section Card) */
.setting-group-card {
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--hairline, #e4e4e7);
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.setting-group-card:hover {
  border-color: var(--color-border-default, #afb8c1);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.06);
}

html.dark .setting-group-card {
  background: #161b22;
  border-color: #30363d;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

html.dark .setting-group-card:hover {
  border-color: #484f58;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
}

.group-card-header {
  padding: 16px 22px;
  border-bottom: 1px solid var(--hairline, #e4e4e7);
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

html.dark .group-card-header {
  border-bottom-color: #30363d;
}

.group-header-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.group-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--charcoal, #24292f);
  margin: 0;
  line-height: 1.3;
}

html.dark .group-title {
  color: #f0f6fc;
}

.group-desc {
  font-size: 13px;
  color: var(--color-fg-muted, #57606a);
  margin: 0;
  line-height: 1.45;
}

html.dark .group-desc {
  color: #8b949e;
}

.group-card-body {
  padding: 8px 22px 18px;
}

/* Symmetrical Dual-Task Grid */
.schedule-tasks-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 18px;
  padding: 12px 0 6px;
}

@media (max-width: 768px) {
  .schedule-tasks-grid {
    grid-template-columns: 1fr;
  }
}

.task-tile-card {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.02));
  border: 1px solid var(--hairline, #e4e4e7);
  border-radius: 8px;
  padding: 16px 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.task-tile-card:hover {
  border-color: var(--color-border-default, #afb8c1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

html.dark .task-tile-card {
  background: #0d1117;
  border-color: #30363d;
}

html.dark .task-tile-card:hover {
  border-color: #484f58;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
}

.task-tile-card.tile-disabled {
  opacity: 0.72;
}

.tile-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--hairline, #e4e4e7);
}

html.dark .tile-card-header {
  border-bottom-color: #21262d;
}

.tile-header-left {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.tile-icon {
  width: 16px;
  height: 16px;
  color: var(--color-accent-fg, #0969da);
  flex-shrink: 0;
}

html.dark .tile-icon {
  color: #58a6ff;
}

.tile-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--charcoal, #24292f);
}

html.dark .tile-title {
  color: #f0f6fc;
}

.tile-card-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tile-field-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 36px;
}

.tile-field-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  color: var(--color-fg-default, #24292f);
}

html.dark .tile-field-label {
  color: #c9d1d9;
}

.tile-hint-box {
  display: flex;
  align-items: flex-start;
  gap: 7px;
  font-size: 12px;
  line-height: 1.45;
  color: var(--color-fg-muted, #57606a);
  background: var(--surface-card, #ffffff);
  padding: 8px 10px;
  border-radius: 6px;
  border: 1px solid var(--hairline, #e4e4e7);
}

html.dark .tile-hint-box {
  background: rgba(255, 255, 255, 0.02);
  border-color: #21262d;
  color: #8b949e;
}

.hint-icon {
  width: 13px;
  height: 13px;
  margin-top: 2px;
  flex-shrink: 0;
  color: var(--color-fg-muted, #57606a);
}

html.dark .hint-icon {
  color: #8b949e;
}

/* Setting Rows */
.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 0;
  gap: 20px;
  min-height: 44px;
  box-sizing: border-box;
}

.setting-info {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
}

.setting-title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal, #24292f);
}

html.dark .setting-title {
  color: #c9d1d9;
}

.label-help-ico {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  font-size: 10px;
  font-weight: 600;
  background: var(--hairline, #e4e4e7);
  color: var(--steel, #71717a);
  cursor: help;
  transition: all 0.15s ease;
  user-select: none;
}

.label-help-ico:hover {
  background: var(--color-accent-fg, #0969da);
  color: #ffffff;
}

.setting-control {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  flex-shrink: 0;
}

.compact-counter {
  width: 120px;
  height: 32px !important;
  line-height: 32px !important;
}

:deep(.compact-counter .el-input__wrapper) {
  height: 32px !important;
  box-sizing: border-box !important;
}

:deep(.compact-counter .el-input-number__decrease),
:deep(.compact-counter .el-input-number__increase) {
  height: 30px !important;
  line-height: 30px !important;
}

.inline-unit-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.unit-text {
  font-size: 12px;
  color: var(--color-fg-muted, #57606a);
  font-weight: 500;
  white-space: nowrap;
  min-width: 24px;
}

html.dark .unit-text {
  color: #8b949e;
}

/* Time Picker Range */
.time-range-row {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.time-picker-box {
  width: 124px !important;
}

.time-sep {
  font-size: 13px;
  color: var(--color-fg-muted, #57606a);
  font-weight: 500;
}

/* Schedule Status Banner */
.schedule-status-banner {
  margin: 6px 0 4px;
  padding: 10px 14px;
  border-radius: 6px;
  background: rgba(9, 105, 218, 0.06);
  border: 1px solid rgba(9, 105, 218, 0.18);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

html.dark .schedule-status-banner {
  background: rgba(56, 139, 253, 0.1);
  border-color: rgba(56, 139, 253, 0.25);
}

.status-banner-neutral {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.02));
  border-color: var(--hairline, #e4e4e7);
}

html.dark .status-banner-neutral {
  background: rgba(255, 255, 255, 0.03);
  border-color: #30363d;
}

.status-banner-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12.5px;
  color: var(--charcoal, #24292f);
}

html.dark .status-banner-left {
  color: #c9d1d9;
}

.status-svg {
  width: 15px;
  height: 15px;
  color: var(--color-accent-fg, #0969da);
  flex-shrink: 0;
}

.status-banner-neutral .status-svg {
  color: var(--color-fg-muted, #57606a);
}

.status-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 12px;
  background: #0969da;
  color: #ffffff;
  font-weight: 500;
}

.status-badge-neutral {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 12px;
  background: rgba(0, 0, 0, 0.06);
  color: var(--color-fg-muted, #57606a);
  font-weight: 500;
}

html.dark .status-badge-neutral {
  background: rgba(255, 255, 255, 0.08);
  color: #8b949e;
}
</style>

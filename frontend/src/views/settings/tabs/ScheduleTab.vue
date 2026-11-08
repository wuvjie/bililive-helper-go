<template>
  <div class="schedule-tab" v-loading="loading">
    <!-- Card 1: 自动分片合并计划 -->
    <div class="setting-group-card">
      <div class="group-card-header">
        <div class="group-header-text">
          <h3 class="group-title">自动分片合并计划</h3>
          <p class="group-desc">主调度器周期性激活录像分片合并流水线，将零散切片拼合为完整录播。</p>
        </div>
      </div>
      <div class="group-card-body">
        <div class="setting-row">
          <div class="setting-info">
            <div class="setting-title">
              <span>开启自动合并</span>
              <el-tooltip content="主调度器周期性激活录像分片合并流水线任务" placement="top" :show-after="300">
                <span class="label-help-ico">?</span>
              </el-tooltip>
            </div>
          </div>
          <div class="setting-control">
            <el-switch
              :model-value="form.merge_enabled"
              @update:model-value="emit('update:form', { ...form, merge_enabled: !!$event })"
            />
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-info">
            <div class="setting-title">
              <span>合并调度周期</span>
              <el-tooltip content="任务执行完毕后等待指定间隔再启动下次合并（推荐 360 分钟 / 6 小时）" placement="top" :show-after="300">
                <span class="label-help-ico">?</span>
              </el-tooltip>
            </div>
          </div>
          <div class="setting-control">
            <div class="inline-unit-row">
              <el-input-number
                :model-value="form.merge_interval"
                :min="10"
                :max="1440"
                :step="15"
                class="compact-counter"
                @update:model-value="emit('update:form', { ...form, merge_interval: Number($event) })"
              />
              <span class="unit-text">分钟</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Card 2: 自动磁盘巡检与清理计划 -->
    <div class="setting-group-card">
      <div class="group-card-header">
        <div class="group-header-text">
          <h3 class="group-title">自动磁盘巡检与清理计划</h3>
          <p class="group-desc">巡检器定时扫描磁盘可用水位，若触达警戒线则自动启动清理冲程。</p>
        </div>
      </div>
      <div class="group-card-body">
        <div class="setting-row">
          <div class="setting-info">
            <div class="setting-title">
              <span>开启自动清理</span>
              <el-tooltip content="定期巡检磁盘可用率，若触达警戒线则自动启动清理" placement="top" :show-after="300">
                <span class="label-help-ico">?</span>
              </el-tooltip>
            </div>
          </div>
          <div class="setting-control">
            <el-switch
              :model-value="form.clean_enabled"
              @update:model-value="emit('update:form', { ...form, clean_enabled: !!$event })"
            />
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-info">
            <div class="setting-title">
              <span>巡检调度周期</span>
              <el-tooltip content="巡检器检查磁盘可用水位的轮询间隔（推荐 720 分钟 / 12 小时）" placement="top" :show-after="300">
                <span class="label-help-ico">?</span>
              </el-tooltip>
            </div>
          </div>
          <div class="setting-control">
            <div class="inline-unit-row">
              <el-input-number
                :model-value="form.clean_interval"
                :min="10"
                :max="1440"
                :step="30"
                class="compact-counter"
                @update:model-value="emit('update:form', { ...form, clean_interval: Number($event) })"
              />
              <span class="unit-text">分钟</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Card 3: 免打扰避让时段 -->
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
              <el-tooltip content="避让时段内主调度器进入休眠，不抢占 CPU 与磁盘 I/O 资源" placement="top" :show-after="300">
                <span class="label-help-ico">?</span>
              </el-tooltip>
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

defineProps<{
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
</script>

<style scoped>
.schedule-tab {
  display: flex;
  flex-direction: column;
}

/* Unified Group Card (GitHub Primer Section Card) */
.setting-group-card {
  background: var(--color-canvas-default, #ffffff);
  border: 1px solid var(--color-border-subtle, #d0d7de);
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.setting-group-card:hover {
  border-color: var(--color-border-default, #afb8c1);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.06);
}

html.dark .setting-group-card {
  background: var(--color-canvas-subtle, #161b22);
  border-color: #30363d;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

html.dark .setting-group-card:hover {
  border-color: #484f58;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
}

.group-card-header {
  padding: 14px 20px;
  border-bottom: 1px solid var(--color-border-subtle, #d0d7de);
  background: var(--color-canvas-subtle, #f6f8fa);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

html.dark .group-card-header {
  border-bottom-color: #30363d;
  background: rgba(255, 255, 255, 0.02);
}

.group-header-text {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.group-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-fg-default, #24292f);
  margin: 0;
}

html.dark .group-title {
  color: #f0f6fc;
}

.group-desc {
  font-size: 12px;
  color: var(--color-fg-muted, #57606a);
  margin: 0;
  line-height: 1.4;
}

html.dark .group-desc {
  color: #8b949e;
}

.group-card-body {
  padding: 4px 20px 8px;
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

.setting-row:not(:last-child) {
  border-bottom: 1px solid var(--color-border-subtle, #d0d7de);
}

html.dark .setting-row:not(:last-child) {
  border-bottom-color: #21262d;
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
  color: var(--color-fg-default, #24292f);
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
  margin: 12px 0 8px;
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
  background: var(--color-canvas-subtle, #f6f8fa);
  border-color: var(--color-border-subtle, #d0d7de);
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
  color: var(--color-fg-default, #24292f);
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

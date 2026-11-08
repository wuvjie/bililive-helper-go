<template>
  <div class="schedule-tab" v-loading="loading">
    <!-- Dual Schedule Plans (2-Column Responsive Grid) -->
    <div class="schedule-two-col-grid">
      <!-- Card 1: 自动分片合并计划 -->
      <div class="setting-group sched-panel-card">
        <div class="subhead">
          <h3 class="subhead-title">自动分片合并计划</h3>
        </div>

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

      <!-- Card 2: 自动磁盘巡检与清理计划 -->
      <div class="setting-group sched-panel-card">
        <div class="subhead">
          <h3 class="subhead-title">自动磁盘巡检与清理计划</h3>
        </div>

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

    <!-- Group 3: 免打扰避让时段 -->
    <div class="setting-group">
      <div class="subhead">
        <h3 class="subhead-title">免打扰避让时段</h3>
      </div>

      <div class="setting-row">
        <div class="setting-info">
          <div class="setting-title">
            <span>避让时间范围</span>
            <el-tooltip content="在此设定时段内自动暂停合并与清理，其余时段恢复正常计划调度（支持跨午夜）" placement="top" :show-after="300">
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
            避让保护已激活：每天 <strong>{{ backupStart }}</strong> 至 <strong>{{ backupEnd }}</strong> 自动暂停合并与清理
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

.setting-group {
  display: flex;
  flex-direction: column;
}
.setting-group + .setting-group {
  margin-top: 32px;
}

/* GitHub Primer Subhead (子分类标题) */
.subhead {
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--hairline);
}

.subhead-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.2px;
}

/* Setting Rows - 单行清爽布局与渐进式展示 */
.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
  gap: 20px;
  min-height: 40px;
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
  color: var(--charcoal);
  letter-spacing: -0.1px;
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
  background: var(--hairline);
  color: var(--slate);
  cursor: help;
  transition: all 0.15s ease;
  user-select: none;
}

.label-help-ico:hover {
  background: var(--primary);
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
  :deep(.el-input__wrapper) {
    height: 32px !important;
    box-sizing: border-box !important;
  }
  :deep(.el-input-number__decrease),
  :deep(.el-input-number__increase) {
    height: 30px !important;
    line-height: 30px !important;
  }
}

.inline-unit-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.unit-text {
  font-size: 12px;
  color: var(--steel);
  font-weight: 500;
  white-space: nowrap;
  min-width: 28px;
  text-align: left;
}

.time-range-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.time-picker-box {
  width: 120px !important;
}

.time-sep {
  font-size: 12px;
  color: var(--steel);
}

/* Schedule Status Banner */
.schedule-status-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: color-mix(in srgb, var(--primary, #2563eb) 6%, transparent);
  border: 1px solid color-mix(in srgb, var(--primary, #2563eb) 20%, transparent);
  border-radius: var(--r-sm, 6px);
  padding: 10px 14px;
  margin-top: 14px;
  gap: 10px;
}
.status-banner-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--charcoal, #27272a);
}
.status-svg {
  width: 14px;
  height: 14px;
  color: var(--primary, #2563eb);
  flex-shrink: 0;
}
.status-badge {
  font-size: 11px;
  font-weight: 600;
  color: var(--primary, #2563eb);
  background: color-mix(in srgb, var(--primary, #2563eb) 12%, transparent);
  padding: 2px 8px;
  border-radius: var(--r-xs, 4px);
  white-space: nowrap;
}
.status-banner-neutral {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.025));
  border-color: var(--hairline, #e4e4e7);
}
.status-banner-neutral .status-svg {
  color: var(--stone, #a1a1aa);
}
.status-badge-neutral {
  font-size: 11px;
  font-weight: 500;
  color: var(--steel, #71717a);
  background: var(--surface-secondary, rgba(0, 0, 0, 0.05));
  padding: 2px 8px;
  border-radius: var(--r-xs, 4px);
  white-space: nowrap;
}
html.dark .schedule-status-banner {
  background: rgba(59, 130, 246, 0.08);
  border-color: rgba(59, 130, 246, 0.25);
}
html.dark .status-banner-left {
  color: #f4f4f5;
}
html.dark .status-badge {
  background: rgba(59, 130, 246, 0.2);
  color: #93c5fd;
}
html.dark .status-banner-neutral {
  background: rgba(255, 255, 255, 0.03);
  border-color: rgba(255, 255, 255, 0.08);
}
html.dark .status-badge-neutral {
  background: rgba(255, 255, 255, 0.08);
  color: #a1a1aa;
}

.schedule-two-col-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  margin-bottom: 20px;

  @media (max-width: 768px) {
    grid-template-columns: 1fr;
  }
}

.sched-panel-card {
  background: var(--color-canvas-subtle, #f6f8fa);
  border: 1px solid var(--color-border-subtle, #d0d7de);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 0 !important;

  .subhead {
    margin-bottom: 8px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--color-border-subtle, #d0d7de);
  }
}

html.dark .sched-panel-card {
  background: rgba(255, 255, 255, 0.02);
  border-color: #30363d;

  .subhead {
    border-bottom-color: #30363d;
  }
}

</style>

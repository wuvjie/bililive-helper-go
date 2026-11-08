<template>
  <div class="schedule-tab" v-loading="loading">
    <!-- Group 1: 自动分片合并计划 -->
    <div class="setting-group">
      <div class="category-divider">
        <div class="category-badge">
          <div class="mini-icon-box icon-box-emerald">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
              <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
            </svg>
          </div>
          <span class="category-label">自动分片合并计划</span>
        </div>
        <span class="category-line" />
      </div>

      <div class="setting-row">
        <div class="setting-info">
          <div class="setting-title">开启自动合并</div>
          <div class="setting-desc">周期性自动扫描各主播目录并合并已完播的视频分片</div>
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
          <div class="setting-title">合并调度周期</div>
          <div class="setting-desc">任务执行完毕后等待指定间隔再启动下次合并（推荐 360 分钟 / 6 小时）</div>
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

    <!-- Group 2: 自动磁盘巡检与清理计划 -->
    <div class="setting-group">
      <div class="category-divider">
        <div class="category-badge">
          <div class="mini-icon-box icon-box-amber">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
            </svg>
          </div>
          <span class="category-label">自动磁盘巡检与清理计划</span>
        </div>
        <span class="category-line" />
      </div>

      <div class="setting-row">
        <div class="setting-info">
          <div class="setting-title">开启自动清理</div>
          <div class="setting-desc">定期巡检磁盘可用率，仅当达到警戒水位时触发清理任务</div>
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
          <div class="setting-title">巡检调度周期</div>
          <div class="setting-desc">巡检器检查磁盘可用水位的轮询间隔（推荐 720 分钟 / 12 小时）</div>
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

    <!-- Group 3: 免打扰避让时段 -->
    <div class="setting-group">
      <div class="category-divider">
        <div class="category-badge">
          <div class="mini-icon-box icon-box-purple">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
            </svg>
          </div>
          <span class="category-label">免打扰避让时段</span>
        </div>
        <span class="category-line" />
      </div>

      <div class="setting-row">
        <div class="setting-info">
          <div class="setting-title">避让时间范围</div>
          <div class="setting-desc">在此设定时段内自动暂停合并与清理，避开直播高峰或高占用期（支持跨午夜）</div>
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
  margin-top: 24px;
}

/* Category Header with Mini Icon Box */
.category-divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 4px 0 16px 0;
}
.category-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: var(--surface-secondary, rgba(0, 0, 0, 0.04));
  padding: 4px 10px 4px 6px;
  border-radius: var(--r-sm, 6px);
  border: 1px solid var(--hairline, #e4e4e7);
}
.mini-icon-box {
  width: 18px;
  height: 18px;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.mini-icon-box svg {
  width: 11px;
  height: 11px;
}
.icon-box-emerald {
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
}
.icon-box-amber {
  background: rgba(245, 158, 11, 0.12);
  color: #d97706;
}
.icon-box-purple {
  background: rgba(139, 92, 246, 0.12);
  color: #7c3aed;
}
.category-label {
  font-size: 11.5px;
  font-weight: 700;
  letter-spacing: 0.4px;
  color: var(--charcoal, #27272a);
  white-space: nowrap;
}
.category-line {
  flex: 1;
  height: 1px;
  background: var(--hairline, #e6e4df);
}

html.dark .category-badge {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.1);
}
html.dark .category-label {
  color: #f4f4f5;
}
html.dark .category-line {
  background: rgba(255, 255, 255, 0.08);
}
html.dark .icon-box-emerald {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
}
html.dark .icon-box-amber {
  background: rgba(245, 158, 11, 0.2);
  color: #fbbf24;
}
html.dark .icon-box-purple {
  background: rgba(139, 92, 246, 0.2);
  color: #c084fc;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 13px 0;
  border-bottom: 1px solid var(--hairline);
  gap: 20px;
}
.setting-row:last-child {
  border-bottom: none;
}

.setting-info {
  display: flex;
  flex-direction: column;
  gap: 3px;
  flex: 1;
  min-width: 0;
}

.setting-title {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.1px;
}

.setting-desc {
  font-size: 12px;
  color: var(--steel);
  line-height: 1.45;
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
  padding: 8px 14px;
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
</style>

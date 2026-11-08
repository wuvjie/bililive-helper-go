<template>
  <div class="schedule-grid">
    <!-- Card 1: 自动合并视频分片 -->
    <div class="card sched-card">
      <div class="card-title card-title-flex">
        <div class="title-left">
          <div class="ui-icon-box icon-box-emerald">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect width="18" height="18" x="3" y="3" rx="2"/>
              <path d="M7 3v18"/>
              <path d="M17 3v18"/>
              <path d="M3 7.5h4"/>
              <path d="M3 12h18"/>
              <path d="M3 16.5h4"/>
              <path d="M17 7.5h4"/>
              <path d="M17 16.5h4"/>
            </svg>
          </div>
          <span class="title-text">自动合并视频分片</span>
          <span class="badge-status" :class="schedule?.merge?.is_running ? 'badge-emerald' : schedule?.merge?.enabled ? 'badge-standing' : 'badge-off'">
            <span class="status-dot-sm" :class="schedule?.merge?.is_running ? 'dot-emerald pulse-dot' : schedule?.merge?.enabled ? 'dot-standing' : 'dot-off'" />
            {{ schedule?.merge?.is_running ? '正在合并录像...' : schedule?.merge?.enabled ? '定时调度中' : '定时已停用' }}
          </span>
        </div>
      </div>

      <div class="sched-metric-grid">
        <div class="sched-metric-cell">
          <span class="metric-lbl">上次运行</span>
          <span class="mono-val">{{ formatDisplayTime(schedule?.merge?.last_run) }}</span>
        </div>
        <div class="sched-metric-cell">
          <span class="metric-lbl">调度周期</span>
          <span class="mono-val">每 {{ schedule?.merge?.interval || 360 }} 分钟</span>
        </div>
        <div class="sched-metric-cell">
          <span class="metric-lbl">下次预计</span>
          <span class="mono-val">{{ formatDisplayTime(schedule?.merge?.next_run) }}</span>
        </div>
      </div>
    </div>

    <!-- Card 2: 自动清理磁盘空间 -->
    <div class="card sched-card">
      <div class="card-title card-title-flex">
        <div class="title-left">
          <div class="ui-icon-box icon-box-amber">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
            </svg>
          </div>
          <span class="title-text">自动清理磁盘空间</span>
          <span class="badge-status" :class="schedule?.clean?.is_running ? 'badge-amber' : schedule?.clean?.enabled ? 'badge-standing' : 'badge-off'">
            <span class="status-dot-sm" :class="schedule?.clean?.is_running ? 'dot-amber' : schedule?.clean?.enabled ? 'dot-standing' : 'dot-off'" />
            {{ schedule?.clean?.is_running ? '正在清理磁盘...' : schedule?.clean?.enabled ? '定时调度中' : '定时已停用' }}
          </span>
        </div>
      </div>

      <div class="sched-metric-grid">
        <div class="sched-metric-cell">
          <span class="metric-lbl">上次运行</span>
          <span class="mono-val">{{ formatDisplayTime(schedule?.clean?.last_run) }}</span>
        </div>
        <div class="sched-metric-cell">
          <span class="metric-lbl">调度周期</span>
          <span class="mono-val">每 {{ schedule?.clean?.interval || 720 }} 分钟</span>
        </div>
        <div class="sched-metric-cell">
          <span class="metric-lbl">下次预计</span>
          <span class="mono-val">{{ formatDisplayTime(schedule?.clean?.next_run) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatTime } from "@/utils/format";
import type { ScheduleStatus } from "@/api/types";

defineProps<{
  schedule?: ScheduleStatus;
}>();

function formatDisplayTime(ts?: number): string {
  if (!ts || ts <= 0) return "未安排 / 待定";
  return formatTime(ts);
}
</script>

<style scoped>
.schedule-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.sched-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

.card-title-flex {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 32px;
  margin-bottom: 16px;
  box-sizing: border-box;
}

.title-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.title-text {
  font-size: 15px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.2px;
}

.badge-status {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 8px;
  border-radius: var(--r-full);
  font-size: 11px;
  font-weight: 500;
}

.badge-emerald {
  background: color-mix(in srgb, #059669 10%, transparent);
  color: #059669;
}

.badge-amber {
  background: color-mix(in srgb, #d97706 10%, transparent);
  color: #b45309;
}

.badge-standing {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.04));
  color: var(--slate);
  border: 1px solid var(--hairline);
}

.badge-off {
  background: rgba(148, 163, 184, 0.1);
  color: #94a3b8;
}

.dot-standing {
  background: #10b981;
}

.dot-amber {
  background: #f59e0b;
}

.dot-emerald {
  background: #10b981;
}

.pulse-dot {
  animation: schedPulse 1.6s infinite;
}

@keyframes schedPulse {
  0% { transform: scale(0.9); opacity: 0.8; }
  50% { transform: scale(1.3); opacity: 1; }
  100% { transform: scale(0.9); opacity: 0.8; }
}

/* Hairline ruled metric grid (3 columns) */
.sched-metric-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  border-top: 1px solid var(--hairline);
  padding-top: 12px;
}

@media (max-width: 640px) {
  .sched-metric-grid {
    grid-template-columns: 1fr;
    gap: 8px;
  }
  .sched-metric-cell + .sched-metric-cell {
    border-left: none;
    border-top: 1px solid var(--hairline);
    padding-top: 8px;
    padding-left: 0;
  }
}

.sched-metric-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 0 14px;
  box-sizing: border-box;
}

.sched-metric-cell:first-child {
  padding-left: 0;
}

.sched-metric-cell:last-child {
  padding-right: 0;
}

.sched-metric-cell + .sched-metric-cell {
  border-left: 1px solid var(--hairline);
}

.metric-lbl {
  font-size: 11.5px;
  color: var(--stone);
  font-weight: 500;
}

.mono-val {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal);
  font-family: var(--font-mono);
}

@media (max-width: 860px) {
  .schedule-grid {
    grid-template-columns: 1fr;
  }
}
</style>

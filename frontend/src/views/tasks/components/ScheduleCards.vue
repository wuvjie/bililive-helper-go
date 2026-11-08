<template>
  <div class="schedule-grid">
    <!-- Card 1: 自动分片合并 -->
    <div class="card sched-card">
      <div class="card-title card-title-flex">
        <div class="title-left">
          <div class="ui-icon-box icon-box-emerald">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
              <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
            </svg>
          </div>
          <div class="title-text-group">
            <span class="title-text">自动分片合并</span>
            <span class="sched-policy-pill">每 {{ schedule?.merge?.interval || 360 }} 分钟</span>
          </div>
          <span class="badge-status" :class="schedule?.merge?.is_running ? 'badge-emerald' : schedule?.merge?.enabled ? 'badge-standing' : 'badge-off'">
            <span class="status-dot-sm" :class="schedule?.merge?.is_running ? 'dot-emerald pulse-dot' : schedule?.merge?.enabled ? 'dot-standing' : 'dot-off'" />
            {{ schedule?.merge?.is_running ? '正在合并视频...' : schedule?.merge?.enabled ? '调度生效中' : '定时已暂停' }}
          </span>
        </div>
        <button
          class="btn-task-action btn-merge-action"
          :disabled="isRunning"
          @click="$emit('run-global', 'merge')"
          title="立即启动一次全局分片合并任务"
        >
          <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="currentColor">
            <polygon points="5 3 19 12 5 21 5 3"/>
          </svg>
          <span>立即执行合并</span>
        </button>
      </div>

      <div class="sched-metric-grid">
        <div class="sched-metric-cell">
          <span class="metric-lbl">上次运行</span>
          <span class="mono-val">{{ formatDisplayTime(schedule?.merge?.last_run) }}</span>
        </div>
        <div class="sched-metric-cell">
          <span class="metric-lbl">下次预计</span>
          <span class="mono-val">{{ formatDisplayTime(schedule?.merge?.next_run) }}</span>
        </div>
      </div>
    </div>

    <!-- Card 2: 自动磁盘清理 -->
    <div class="card sched-card">
      <div class="card-title card-title-flex">
        <div class="title-left">
          <div class="ui-icon-box icon-box-amber">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
            </svg>
          </div>
          <div class="title-text-group">
            <span class="title-text">自动磁盘清理</span>
            <span class="sched-policy-pill">每 {{ schedule?.clean?.interval || 720 }} 分钟</span>
          </div>
          <span class="badge-status" :class="schedule?.clean?.is_running ? 'badge-amber' : schedule?.clean?.enabled ? 'badge-standing' : 'badge-off'">
            <span class="status-dot-sm" :class="schedule?.clean?.is_running ? 'dot-amber' : schedule?.clean?.enabled ? 'dot-standing' : 'dot-off'" />
            {{ schedule?.clean?.is_running ? '正在清理磁盘...' : schedule?.clean?.enabled ? '巡检生效中' : '保护未开启' }}
          </span>
        </div>
        <button
          class="btn-task-action btn-clean-action"
          :disabled="isRunning"
          @click="$emit('run-global', 'clean')"
          title="立即启动一次全局磁盘巡检与清理"
        >
          <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
          </svg>
          <span>立即执行清理</span>
        </button>
      </div>

      <div class="sched-metric-grid">
        <div class="sched-metric-cell">
          <span class="metric-lbl">上次运行</span>
          <span class="mono-val">{{ formatDisplayTime(schedule?.clean?.last_run) }}</span>
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
  isRunning: boolean;
}>();

defineEmits<{
  (e: "run-global", task: "merge" | "clean"): void;
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
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
}

.card-title-flex {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
  gap: 12px;
  flex-wrap: wrap;
}

.title-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.title-text-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.title-text {
  font-size: 15px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.2px;
}

.sched-policy-pill {
  font-size: 11px;
  color: var(--steel);
  background: var(--surface-secondary, rgba(0, 0, 0, 0.04));
  padding: 1px 7px;
  border-radius: var(--r-xs);
  border: 1px solid var(--hairline);
  font-weight: 500;
  white-space: nowrap;
}

html.dark .sched-policy-pill {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.1);
  color: #a1a1aa;
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

.btn-task-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: var(--r-sm);
  font-size: 12.5px;
  font-weight: 600;
  cursor: pointer;
  border-width: 1px;
  border-style: solid;
  transition: all 0.15s ease;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  line-height: 1.4;
}

.btn-merge-action {
  background: color-mix(in srgb, #059669 12%, transparent);
  color: #047857;
  border-color: color-mix(in srgb, #059669 35%, transparent);
}
.btn-merge-action:hover:not(:disabled) {
  background: #059669;
  border-color: #047857;
  color: #ffffff;
  box-shadow: 0 2px 6px rgba(5, 150, 105, 0.25);
}

.btn-clean-action {
  background: color-mix(in srgb, #d97706 12%, transparent);
  color: #b45309;
  border-color: color-mix(in srgb, #d97706 35%, transparent);
}
.btn-clean-action:hover:not(:disabled) {
  background: #d97706;
  border-color: #b45309;
  color: #ffffff;
  box-shadow: 0 2px 6px rgba(217, 119, 6, 0.25);
}

.btn-task-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-svg-sm {
  width: 12px;
  height: 12px;
}

/* Hairline ruled metric grid (2 columns) */
.sched-metric-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  border-top: 1px solid var(--hairline);
  padding-top: 12px;
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

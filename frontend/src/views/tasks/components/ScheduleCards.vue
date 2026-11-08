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
          <span class="title-text">自动分片合并</span>
        </div>
        <span class="badge-status" :class="schedule?.merge?.is_running ? 'badge-emerald' : schedule?.merge?.enabled ? 'badge-standing' : 'badge-off'">
          <span class="status-dot-sm" :class="schedule?.merge?.is_running ? 'dot-emerald pulse-dot' : schedule?.merge?.enabled ? 'dot-standing' : 'dot-off'" />
          {{ schedule?.merge?.is_running ? '正在合并视频...' : schedule?.merge?.enabled ? '定时调度已启用' : '定时调度已暂停' }}
        </span>
      </div>
      <div class="card-body sched-body">
        <div class="sched-row">
          <span class="sched-label">调度周期</span>
          <span class="mono-val">每 {{ schedule?.merge?.interval || 360 }} <span class="unit-lbl">分钟</span></span>
        </div>
        <div class="sched-row">
          <span class="sched-label">上次运行</span>
          <span class="mono-val-sm">{{ formatDisplayTime(schedule?.merge?.last_run) }}</span>
        </div>
        <div class="sched-row">
          <span class="sched-label">下次预计</span>
          <span class="mono-val-sm">{{ formatDisplayTime(schedule?.merge?.next_run) }}</span>
        </div>
        <div class="sched-action-row">
          <button
            class="btn-task-action btn-merge-action"
            :disabled="isRunning"
            @click="$emit('run-global', 'merge')"
          >
            <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>
            </svg>
            <span>立即全局合并</span>
          </button>
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
          <span class="title-text">自动磁盘清理</span>
        </div>
        <span class="badge-status" :class="schedule?.clean?.is_running ? 'badge-amber' : schedule?.clean?.enabled ? 'badge-standing' : 'badge-off'">
          <span class="status-dot-sm" :class="schedule?.clean?.is_running ? 'dot-amber' : schedule?.clean?.enabled ? 'dot-standing' : 'dot-off'" />
          {{ schedule?.clean?.is_running ? '正在清理磁盘...' : schedule?.clean?.enabled ? '磁盘巡检已开启' : '巡检保护未开启' }}
        </span>
      </div>
      <div class="card-body sched-body">
        <div class="sched-row">
          <span class="sched-label">巡检周期</span>
          <span class="mono-val">每 {{ schedule?.clean?.interval || 720 }} <span class="unit-lbl">分钟</span></span>
        </div>
        <div class="sched-row">
          <span class="sched-label">上次运行</span>
          <span class="mono-val-sm">{{ formatDisplayTime(schedule?.clean?.last_run) }}</span>
        </div>
        <div class="sched-row">
          <span class="sched-label">下次预计</span>
          <span class="mono-val-sm">{{ formatDisplayTime(schedule?.clean?.next_run) }}</span>
        </div>
        <div class="sched-action-row">
          <button
            class="btn-task-action btn-clean-action"
            :disabled="isRunning"
            @click="$emit('run-global', 'clean')"
          >
            <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
            </svg>
            <span>立即全局清理</span>
          </button>
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
  margin-bottom: 16px;
}

.title-left {
  display: flex;
  align-items: center;
  gap: 10px;
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

.sched-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
  flex: 1;
  justify-content: space-between;
}

.sched-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
  padding: 4px 0;
  border-bottom: 1px dashed var(--hairline);
}

.sched-label {
  color: var(--slate);
}

.sched-action-row {
  margin-top: 14px;
  display: flex;
  width: 100%;
}

.btn-task-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  gap: 6px;
  padding: 8px 16px;
  border-radius: var(--r-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.15s ease;
  box-sizing: border-box;
}

.btn-merge-action {
  background: color-mix(in srgb, #059669 10%, transparent);
  color: #047857;
  border-color: color-mix(in srgb, #059669 25%, transparent);
}
.btn-merge-action:hover:not(:disabled) {
  background: #059669;
  color: #ffffff;
}

.btn-clean-action {
  background: color-mix(in srgb, #d97706 10%, transparent);
  color: #b45309;
  border-color: color-mix(in srgb, #d97706 25%, transparent);
}
.btn-clean-action:hover:not(:disabled) {
  background: #d97706;
  color: #ffffff;
}

.btn-task-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-svg-sm {
  width: 14px;
  height: 14px;
}

@media (max-width: 860px) {
  .schedule-grid {
    grid-template-columns: 1fr;
  }
}
</style>

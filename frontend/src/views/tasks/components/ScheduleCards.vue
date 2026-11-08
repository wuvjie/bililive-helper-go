<template>
  <div class="schedule-grid">
    <!-- Card 1: 自动分片合并 -->
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
          <span class="title-text">自动分片合并</span>
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
          <span class="metric-lbl">调度周期</span>
          <span class="mono-val">每 {{ schedule?.merge?.interval || 360 }} 分钟</span>
        </div>
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
          <span class="title-text">自动磁盘清理</span>
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
          <span class="metric-lbl">调度周期</span>
          <span class="mono-val">每 {{ schedule?.clean?.interval || 720 }} 分钟</span>
        </div>
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

.sched-policy-meta {
  font-size: 12px;
  color: var(--color-fg-muted, #57606a);
  font-family: var(--font-mono, ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace);
  font-weight: 500;
  white-space: nowrap;
}

html.dark .sched-policy-meta {
  color: #8b949e;
}

.title-right-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sched-jump-link {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  color: var(--color-fg-muted, #57606a);
  background: transparent;
  border: 1px solid transparent;
  text-decoration: none;
  transition: all 0.15s ease;

  &:hover {
    color: var(--color-accent-fg, #0969da);
    background: var(--color-canvas-subtle, #f6f8fa);
    border-color: var(--color-border-subtle, #d0d7de);
  }

  .sched-jump-icon {
    width: 14px;
    height: 14px;
  }
}

html.dark .sched-jump-link {
  color: #8b949e;

  &:hover {
    color: #58a6ff;
    background: #21262d;
    border-color: #30363d;
  }
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
  padding: 5px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  background: var(--color-btn-bg, #ffffff);
  color: var(--color-btn-text, #24292f);
  border: 1px solid var(--color-btn-border, #d0d7de);
  box-shadow: 0 1px 0 rgba(27, 31, 36, 0.04);
  transition: all 0.15s cubic-bezier(0.16, 1, 0.3, 1);
  line-height: 20px;

  &:hover:not(:disabled) {
    background: var(--color-btn-hover-bg, #f3f4f6);
    border-color: var(--color-btn-hover-border, #d0d7de);
    color: var(--color-btn-text, #24292f);
  }

  &:active:not(:disabled) {
    background: var(--color-btn-active-bg, #ebecf0);
  }

  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
}

.btn-merge-action .btn-svg-sm {
  color: #1a7f37;
}

.btn-clean-action .btn-svg-sm {
  color: #9a6700;
}

html.dark .btn-task-action {
  background: #21262d;
  color: #c9d1d9;
  border-color: #30363d;
  box-shadow: 0 1px 0 rgba(0, 0, 0, 0.1);

  &:hover:not(:disabled) {
    background: #30363d;
    border-color: #8b949e;
    color: #f0f6fc;
  }

  .btn-merge-action .btn-svg-sm {
    color: #3fb950;
  }

  .btn-clean-action .btn-svg-sm {
    color: #d29922;
  }
}

.btn-task-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-svg-sm {
  width: 12px;
  height: 12px;
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

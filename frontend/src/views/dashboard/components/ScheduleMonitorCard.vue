<template>
  <div class="card schedule-monitor-card">
    <div class="card-title">
      <div class="card-title-left">
        <div class="ui-icon-box icon-box-purple">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
        </div>
        <span class="card-title-main">自动化调度监控</span>
        <span class="monitor-mode-tag" :class="isTaskRunning ? 'tag-running' : 'tag-in-service'">
          <span class="status-pulse-dot" :class="isTaskRunning ? 'dot-active' : 'dot-in-service'"></span>
          {{ isTaskRunning ? '任务执行中' : '自动调度生效中' }}
        </span>
      </div>
      <router-link to="/tasks" class="card-header-action" title="前往任务中心查看流式终端与执行记录">
        任务中心 ↗
      </router-link>
    </div>

    <div class="card-body">
      <div class="card-top-metric">
        <div class="runtime-top">
          <span class="runtime-label">当前任务状态</span>
          <div class="runtime-state-indicator">
            <span class="state-dot" :class="isTaskRunning ? 'dot-running' : 'dot-idle'"></span>
            <span class="state-text" :class="{ 'text-running': isTaskRunning }">
              {{ isTaskRunning ? (detail?.schedule?.merge?.is_running ? '正在合并视频...' : '正在清理磁盘...') : '待命就绪 · 暂无活动任务' }}
            </span>
          </div>
        </div>
        <div class="runtime-subline">
          <span class="runtime-subtext">
            {{ isTaskRunning ? '后台正在处理视频合并或磁盘清理任务，前往任务中心可查看实时终端输出' : '后台自动化任务已就绪，按设定周期自动执行' }}
          </span>
        </div>
      </div>

      <div class="data-grid">
        <div class="data-cell">
          <span class="data-label">合并周期</span>
          <span class="mono-val" :class="{ 'val-disabled': !detail?.schedule?.merge_enabled }">
            {{ detail?.schedule?.merge_enabled ? ('每 ' + (detail?.schedule?.merge_interval || 360) + ' 分钟') : '已暂停' }}
          </span>
        </div>
        <div class="data-cell">
          <span class="data-label">下次合并</span>
          <span class="mono-val" :class="{ 'val-disabled': !detail?.schedule?.merge_enabled }">
            {{ detail?.schedule?.merge_enabled ? formatNextTime(detail?.schedule?.merge?.next_run) : '-' }}
          </span>
        </div>
        <div class="data-cell">
          <span class="data-label">清理周期</span>
          <span class="mono-val" :class="{ 'val-disabled': !detail?.schedule?.clean_enabled }">
            {{ detail?.schedule?.clean_enabled ? ('每 ' + (detail?.schedule?.clean_interval || 720) + ' 分钟') : '未开启' }}
          </span>
        </div>
        <div class="data-cell">
          <span class="data-label">下次清理</span>
          <span class="mono-val" :class="{ 'val-disabled': !detail?.schedule?.clean_enabled }">
            {{ detail?.schedule?.clean_enabled ? formatNextTime(detail?.schedule?.clean?.next_run) : '-' }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { DetailResponse, SetupCheck } from "@/api/types";

const props = defineProps<{
  detail?: DetailResponse;
  setup?: SetupCheck;
}>();

const isTaskRunning = computed(() => {
  return Boolean(
    props.detail?.schedule?.running ||
    props.detail?.schedule?.merge?.is_running ||
    props.detail?.schedule?.clean?.is_running
  );
});

function formatNextTime(ts?: number): string {
  if (!ts || ts <= 0) return "未安排";
  const date = new Date(ts * 1000);
  const now = new Date();
  const diffMinutes = Math.round((date.getTime() - now.getTime()) / 60000);
  if (diffMinutes <= 0) return "即将执行";
  if (diffMinutes < 60) return `${diffMinutes} 分钟后`;

  const hours = String(date.getHours()).padStart(2, "0");
  const mins = String(date.getMinutes()).padStart(2, "0");
  const timeStr = `${hours}:${mins}`;

  const isToday = date.toDateString() === now.toDateString();
  if (isToday) return `今日 ${timeStr}`;

  const tomorrow = new Date(now);
  tomorrow.setDate(tomorrow.getDate() + 1);
  if (date.toDateString() === tomorrow.toDateString()) return `明日 ${timeStr}`;

  return `${date.getMonth() + 1}/${date.getDate()} ${timeStr}`;
}
</script>

<style scoped>
.schedule-monitor-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

.card-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 32px;
  margin-bottom: 16px;
  box-sizing: border-box;
}

.card-title-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-title-main {
  font-size: 15px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.2px;
  line-height: 1;
}

.monitor-mode-tag {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 8px;
  border-radius: var(--r-full);
  font-size: 11px;
  font-weight: 500;
  line-height: 1.4;
}

.status-pulse-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
}

.tag-running {
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
}
.dot-active {
  background: #10b981;
  animation: schedPulse 1.6s infinite;
}

.tag-in-service {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.04));
  color: var(--slate);
  border: 1px solid var(--hairline);
}
.dot-in-service {
  background: #10b981;
}

@keyframes schedPulse {
  0% { transform: scale(0.9); opacity: 0.8; }
  50% { transform: scale(1.3); opacity: 1; }
  100% { transform: scale(0.9); opacity: 0.8; }
}

.card-header-action {
  font-size: 12px;
  font-weight: 500;
  color: var(--primary);
  text-decoration: none;
  transition: opacity 0.15s ease;
  line-height: 1;
}
.card-header-action:hover {
  text-decoration: underline;
}

/* Deconstructed top metric - fixed height 56px, exactly matching storage card */
.card-top-metric {
  height: 56px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--hairline);
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.runtime-top {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  height: 24px;
}

.runtime-label {
  font-size: 13px;
  color: var(--charcoal);
  font-weight: 500;
  line-height: 1;
}

.runtime-state-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  line-height: 1;
}

.state-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
  flex-shrink: 0;
}

.dot-idle {
  background: #10b981;
}

.dot-running {
  background: #10b981;
  animation: schedPulse 1.6s infinite;
}

.state-text {
  font-size: 12.5px;
  font-weight: 500;
  color: var(--slate);
  line-height: 1;
}

.text-running {
  color: #059669;
  font-weight: 600;
}

.runtime-subline {
  height: 16px;
  display: flex;
  align-items: center;
}

.runtime-subtext {
  font-size: 11.5px;
  color: var(--stone);
  line-height: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Hairline ruled grid (2x2) - locked 52px row height for absolute cross-card alignment */
.data-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 52px 52px;
  position: relative;
  box-sizing: border-box;
}

.data-cell {
  padding: 6px 12px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 3px;
  box-sizing: border-box;
  overflow: hidden;
}

.data-cell:nth-child(odd) {
  padding-left: 0;
  border-right: 1px solid var(--hairline);
}

.data-cell:nth-child(even) {
  padding-right: 0;
}

.data-cell:nth-child(-n+2) {
  border-bottom: 1px solid var(--hairline);
}

.data-label {
  font-size: 12px;
  color: var(--stone);
  display: flex;
  align-items: center;
  gap: 4px;
  line-height: 1.2;
}

.val-disabled {
  color: var(--stone);
}
</style>

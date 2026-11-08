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
        <div class="runtime-top" :class="{ 'has-substrip': isTaskRunning }">
          <span class="runtime-label">当前任务状态</span>
          <div class="runtime-state-indicator">
            <span class="state-dot" :class="isTaskRunning ? 'dot-running' : 'dot-idle'"></span>
            <span class="state-text" :class="{ 'text-running': isTaskRunning }">
              {{ isTaskRunning ? (detail?.schedule?.merge?.is_running ? '正在合并视频...' : '正在清理磁盘...') : '待命就绪 · 暂无活动任务' }}
            </span>
          </div>
        </div>
        <div v-if="isTaskRunning" class="runtime-substrip">
          <span class="substrip-text">
            后台正在处理媒体流水线任务，任务中心可查看终端流式输出
          </span>
        </div>
      </div>

      <div class="data-grid">
        <div class="data-cell">
          <span class="data-label">自动合并周期</span>
          <span class="mono-val">{{ detail?.schedule?.merge_enabled ? ('每 ' + (detail?.schedule?.merge_interval || 360) + ' 分钟') : '已暂停' }}</span>
        </div>
        <div class="data-cell">
          <span class="data-label">下次合并预计</span>
          <span class="mono-val">{{ formatNextTime(detail?.schedule?.merge?.next_run) }}</span>
        </div>
        <div class="data-cell">
          <span class="data-label">自动清理周期</span>
          <span class="mono-val">{{ detail?.schedule?.clean_enabled ? ('每 ' + (detail?.schedule?.clean_interval || 720) + ' 分钟') : '未开启' }}</span>
        </div>
        <div class="data-cell">
          <span class="data-label">下次清理预计</span>
          <span class="mono-val">{{ formatNextTime(detail?.schedule?.clean?.next_run) }}</span>
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

const streamerCount = computed(() => {
  return props.setup?.streamer_count ?? props.detail?.streamers?.length ?? 0;
});

const videoCount = computed(() => {
  return props.setup?.video_count ?? props.detail?.streamers?.reduce((acc, s) => acc + (s.files || 0), 0) ?? 0;
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
  return `${hours}:${mins}`;
}
</script>

<style scoped>
.schedule-monitor-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
}

.card-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
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
}
.card-header-action:hover {
  text-decoration: underline;
}

/* Deconstructed top metric - hairline underline, no box */
.card-top-metric {
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--hairline);
}

.runtime-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  min-height: 28px;
}

.runtime-top.has-substrip {
  margin-bottom: 6px;
}

.runtime-label {
  font-size: 13px;
  color: var(--charcoal);
  font-weight: 500;
}

.runtime-state-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
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
}

.text-running {
  color: #059669;
  font-weight: 600;
}

.runtime-substrip {
  font-size: 11.5px;
  color: var(--stone);
  line-height: 1.45;
}

/* Hairline ruled grid (2x2) */
.data-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  position: relative;
}

.data-cell {
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  box-sizing: border-box;
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
  padding-top: 0;
}

.data-cell:nth-child(n+3) {
  padding-bottom: 4px;
}

.data-label {
  font-size: 12px;
  color: var(--stone);
  display: flex;
  align-items: center;
  gap: 4px;
}
</style>

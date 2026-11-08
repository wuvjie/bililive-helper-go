<template>
  <div class="card schedule-monitor-card">
    <div class="card-title">
      <div class="card-title-left">
        <div class="ui-icon-box icon-box-purple">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>
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
          <div class="runtime-badge" :class="isTaskRunning ? 'badge-running' : 'badge-idle'">
            <span>{{ isTaskRunning ? (detail?.schedule?.merge?.is_running ? '正在合并视频...' : '正在清理磁盘...') : '空闲待命中' }}</span>
          </div>
        </div>
        <div class="runtime-substrip">
          <span class="substrip-text">
            {{ isTaskRunning ? '后台正在处理媒体流水线任务，任务中心可查看终端流式输出' : '定时调度器与目录监听器持续常驻，主播关播分片将自动静置后合并' }}
          </span>
        </div>
      </div>

      <div class="data-grid">
        <div class="data-cell">
          <span class="data-label">自动合并</span>
          <span class="mono-val">{{ detail?.schedule?.merge_enabled ? ('每 ' + (detail?.schedule?.merge_interval || 360) + ' 分钟') : '已暂停' }}</span>
        </div>
        <div class="data-cell">
          <span class="data-label">自动清理</span>
          <span class="mono-val">{{ detail?.schedule?.clean_enabled ? ('每 ' + (detail?.schedule?.clean_interval || 720) + ' 分钟') : '未开启' }}</span>
        </div>
        <div class="data-cell">
          <span class="data-label">录制主播</span>
          <span class="mono-val">{{ streamerCount }} 位</span>
        </div>
        <div class="data-cell">
          <span class="data-label">
            视频总数
            <el-tooltip content="全部主播录像文件总数（含未合并分段切片与已合并成品）" placement="top">
              <span class="label-help-ico">?</span>
            </el-tooltip>
          </span>
          <span class="mono-val">{{ videoCount }} 个</span>
        </div>
      </div>

      <div class="bottom-health-bar">
        <div class="health-cell">
          <span class="health-cell-label">下次合并</span>
          <span class="health-cell-val mono-val">
            <strong>{{ formatNextTime(detail?.schedule?.merge?.next_run) }}</strong>
          </span>
        </div>
        <div class="health-cell">
          <span class="health-cell-label">下次清理</span>
          <span class="health-cell-val mono-val">
            <strong>{{ formatNextTime(detail?.schedule?.clean?.next_run) }}</strong>
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

.card-header-action {
  font-size: 12px;
  color: var(--primary);
  text-decoration: none;
  font-weight: 500;
  transition: color 0.15s ease;
}

.card-header-action:hover {
  text-decoration: underline;
}

.monitor-mode-tag {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 7px;
  border-radius: var(--r-full);
  font-size: 11px;
  font-weight: 500;
}

.tag-in-service {
  background: color-mix(in srgb, #059669 10%, transparent);
  color: #059669;
}

.tag-running {
  background: color-mix(in srgb, #d97706 12%, transparent);
  color: #b45309;
}

.status-pulse-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
}

.dot-in-service {
  background: #10b981;
}

.dot-active {
  background: #f59e0b;
  animation: pulseDot 1.5s infinite;
}

@keyframes pulseDot {
  0% { transform: scale(0.9); opacity: 0.8; }
  50% { transform: scale(1.3); opacity: 1; }
  100% { transform: scale(0.9); opacity: 0.8; }
}

.card-top-metric {
  margin-bottom: 20px;
  background: var(--surface-secondary, rgba(0, 0, 0, 0.02));
  padding: 12px 14px;
  border-radius: var(--r-sm);
  border: 1px solid var(--hairline);
}

html.dark .card-top-metric {
  background: rgba(255, 255, 255, 0.03);
}

.runtime-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.runtime-label {
  font-size: 13px;
  color: var(--slate);
  font-weight: 500;
}

.runtime-badge {
  font-size: 12px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: var(--r-full);
}

.badge-idle {
  background: color-mix(in srgb, #059669 10%, transparent);
  color: #059669;
}

.badge-running {
  background: color-mix(in srgb, #d97706 12%, transparent);
  color: #b45309;
}

.runtime-substrip {
  font-size: 11px;
  color: var(--stone);
  line-height: 1.4;
}

.data-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}

.data-cell {
  background: var(--surface);
  border: 1px solid var(--hairline);
  padding: 10px 14px;
  border-radius: var(--r-sm);
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.data-label {
  font-size: 12px;
  color: var(--stone);
  display: flex;
  align-items: center;
  gap: 4px;
}

.label-help-ico {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  font-size: 10px;
  background: var(--hairline);
  color: var(--slate);
  cursor: help;
}

.bottom-health-bar {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  padding-top: 14px;
  border-top: 1px dashed var(--hairline);
}

.health-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px 12px;
  border-radius: var(--r-sm);
  background: var(--surface-secondary, rgba(0, 0, 0, 0.025));
  border: 1px solid var(--hairline);
  box-sizing: border-box;
}

.health-cell-label {
  font-size: 11px;
  color: var(--stone);
  font-weight: 500;
}

.health-cell-val {
  font-size: 13px;
  color: var(--charcoal);
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.health-cell-val strong {
  font-size: 14px;
  font-weight: 600;
  color: var(--charcoal);
}

html.dark .health-cell {
  background: rgba(255, 255, 255, 0.03);
  border-color: rgba(255, 255, 255, 0.08);
}
</style>

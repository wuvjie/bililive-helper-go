<template>
  <div class="card trend-card">
    <div class="card-title">
      <div class="card-title-left">
        <div class="ui-icon-box icon-box-teal">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="20" x2="18" y2="10"/>
            <line x1="12" y1="20" x2="12" y2="4"/>
            <line x1="6" y1="20" x2="6" y2="14"/>
          </svg>
        </div>
        <span class="card-title-main">近 7 天产出与清理趋势</span>
      </div>
      <span v-if="hasTrendData" class="trend-sub-summary">
        累计合并 {{ totalMergeGB }} · 释放 {{ totalCleanGB }}
      </span>
    </div>

    <div class="card-body trend-body">
      <div v-if="hasTrendData" class="trend-chart">
        <div v-for="day in stats?.daily" :key="day.date" class="trend-col">
          <div class="trend-bars">
            <el-tooltip :content="getMergeTooltip(day)" placement="top">
              <div class="trend-bar bar-merge" :style="{ height: barHeight(day.merge_bytes) + 'px' }" />
            </el-tooltip>
            <el-tooltip :content="getCleanTooltip(day)" placement="top">
              <div class="trend-bar bar-clean" :style="{ height: barHeight(day.clean_bytes) + 'px' }" />
            </el-tooltip>
          </div>
          <span class="trend-date">{{ day.date.slice(5) }}</span>
        </div>
      </div>

      <div v-else class="trend-empty empty-illustration">
        <div class="empty-icon-wrapper" style="background: rgba(20, 184, 166, 0.09); color: #0d9488;">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="20" x2="18" y2="10"/>
            <line x1="12" y1="20" x2="12" y2="4"/>
            <line x1="6" y1="20" x2="6" y2="14"/>
          </svg>
        </div>
        <span class="empty-title">近 7 天暂无自动化处理数据</span>
        <span class="empty-desc">当录像分片合并或触发磁盘清理时，每日吞吐量将自动在此展现</span>
      </div>

      <div class="trend-legend">
        <span><span class="legend-dot" style="background: #787774" />合并量</span>
        <span><span class="legend-dot" style="background: var(--success)" />释放量</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { formatBytes } from "@/utils/format";
import type { StatsResponse, DayStats } from "@/api/types";

const props = defineProps<{
  stats?: StatsResponse;
}>();

const hasTrendData = computed(() => {
  if (!props.stats?.daily || props.stats.daily.length === 0) return false;
  return props.stats.daily.some(d => (d.merge_bytes > 0 || d.clean_bytes > 0));
});

const totalMergeGB = computed(() => {
  const bytes = props.stats?.daily?.reduce((acc, d) => acc + (d.merge_bytes || 0), 0) || 0;
  return formatBytes(bytes);
});

const totalCleanGB = computed(() => {
  const bytes = props.stats?.daily?.reduce((acc, d) => acc + (d.clean_bytes || 0), 0) || 0;
  return formatBytes(bytes);
});

const maxBytes = computed(() => {
  if (!props.stats?.daily) return 1;
  return Math.max(...props.stats.daily.map(d => Math.max(d.merge_bytes || 0, d.clean_bytes || 0)), 1);
});

function barHeight(b: number): number {
  if (!b || b <= 0) return 4;
  return Math.max(6, Math.min(130, Math.round((b / maxBytes.value) * 130)));
}

function getMergeTooltip(day: DayStats): string {
  if (!day.merge_bytes || day.merge_bytes <= 0) return `${day.date} 无合并任务`;
  const count = day.merge_count ? `${day.merge_count} 次` : "";
  return `合并 ${count} (${formatBytes(day.merge_bytes)})`;
}

function getCleanTooltip(day: DayStats): string {
  if (!day.clean_bytes || day.clean_bytes <= 0) return `${day.date} 无清理操作`;
  const count = day.clean_count ? `${day.clean_count} 次` : "";
  return `释放 ${count} (${formatBytes(day.clean_bytes)})`;
}
</script>

<style scoped>
.trend-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
}

.card-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.card-title-left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.card-title-main {
  font-size: 15px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.2px;
  white-space: nowrap;
}

.trend-sub-summary {
  font-size: 12px;
  color: var(--steel);
  font-family: var(--font-mono);
  white-space: nowrap;
}

.trend-body {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  flex: 1;
  min-height: 180px;
}

.trend-chart {
  display: flex;
  align-items: flex-end;
  justify-content: space-around;
  height: 170px;
  padding: 10px 0;
  border-bottom: 1px solid var(--hairline);
}

.trend-col {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.trend-bars {
  display: flex;
  align-items: flex-end;
  gap: 4px;
  height: 130px;
}

.trend-bar {
  width: 14px;
  border-radius: 3px 3px 0 0;
  cursor: pointer;
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.trend-bar:hover {
  opacity: 0.85;
  transform: scaleY(1.03);
}

.bar-merge {
  background: #787774;
}

.bar-clean {
  background: var(--success);
}

.trend-date {
  font-size: 11px;
  font-family: var(--font-mono);
  color: var(--stone);
}

.trend-empty {
  padding: 30px 0;
}

.trend-legend {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 24px;
  font-size: 12px;
  color: var(--slate);
  margin-top: 14px;
}

.trend-legend span {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 2px;
}
</style>

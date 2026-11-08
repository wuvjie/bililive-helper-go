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
      <div v-if="hasTrendData" class="trend-chart-wrapper">
        <!-- Y-Axis 3-tier Hairline Gridlines & Labels -->
        <div class="trend-y-axis">
          <span class="y-label">{{ maxLabel }}</span>
          <span class="y-label">{{ midLabel }}</span>
          <span class="y-label">0</span>
        </div>

        <div class="trend-stage">
          <!-- 3 Hairline Reference Lines -->
          <div class="gridline gridline-top"></div>
          <div class="gridline gridline-mid"></div>
          <div class="gridline gridline-zero"></div>

          <!-- 7 Columns with Column Hover Tracks -->
          <div class="trend-cols">
            <div
              v-for="day in stats?.daily"
              :key="day.date"
              class="trend-col-track"
            >
              <el-tooltip placement="top" :show-after="80">
                <template #content>
                  <div class="trend-tooltip">
                    <div class="tooltip-header">{{ day.date }}</div>
                    <div class="tooltip-item">
                      <span class="tooltip-dot dot-merge"></span>
                      <span class="tooltip-lbl">生产合并:</span>
                      <span class="tooltip-val">{{ formatBytes(day.merge_bytes) }} ({{ day.merge_count || 0 }} 次)</span>
                    </div>
                    <div class="tooltip-item">
                      <span class="tooltip-dot dot-clean"></span>
                      <span class="tooltip-lbl">释放清理:</span>
                      <span class="tooltip-val">{{ formatBytes(day.clean_bytes) }} ({{ day.clean_count || 0 }} 次)</span>
                    </div>
                    <div class="tooltip-divider"></div>
                    <div class="tooltip-net">
                      <span>净空间变化:</span>
                      <span :class="getNetDiffClass(day)">{{ getNetDiffText(day) }}</span>
                    </div>
                  </div>
                </template>

                <div class="trend-col-inner">
                  <div class="trend-bars">
                    <div class="trend-bar bar-merge" :style="{ height: barHeight(day.merge_bytes) + 'px' }" />
                    <div class="trend-bar bar-clean" :style="{ height: barHeight(day.clean_bytes) + 'px' }" />
                  </div>
                  <span class="trend-date">{{ day.date.slice(5) }}</span>
                </div>
              </el-tooltip>
            </div>
          </div>
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
        <span><span class="legend-dot dot-merge" />生产合并</span>
        <span><span class="legend-dot dot-clean" />释放清理</span>
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
  if (!props.stats?.daily || props.stats.daily.length === 0) return 1;
  const max = Math.max(...props.stats.daily.map(d => Math.max(d.merge_bytes || 0, d.clean_bytes || 0)), 0);
  return max > 0 ? max : 1;
});

const maxLabel = computed(() => {
  return formatBytes(maxBytes.value);
});

const midLabel = computed(() => {
  return formatBytes(Math.round(maxBytes.value / 2));
});

function barHeight(b: number): number {
  if (!b || b <= 0) return 3;
  return Math.max(5, Math.min(116, Math.round((b / maxBytes.value) * 116)));
}

function getNetDiffText(day: DayStats): string {
  const net = (day.merge_bytes || 0) - (day.clean_bytes || 0);
  if (net === 0) return "0 B";
  if (net > 0) return `+${formatBytes(net)}`;
  return `-${formatBytes(Math.abs(net))}`;
}

function getNetDiffClass(day: DayStats): string {
  const net = (day.merge_bytes || 0) - (day.clean_bytes || 0);
  if (net > 0) return "text-inflow";
  if (net < 0) return "text-outflow";
  return "";
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
  min-height: 190px;
}

/* Precision chart container with 3-tier hairline gridlines */
.trend-chart-wrapper {
  display: flex;
  align-items: flex-start;
  height: 160px;
  padding: 10px 0 0 0;
  gap: 10px;
}

.trend-y-axis {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  height: 120px;
  text-align: right;
  min-width: 48px;
  padding-right: 4px;
}

.y-label {
  font-size: 10px;
  font-family: var(--font-mono);
  color: var(--stone);
  line-height: 1;
}

.trend-stage {
  position: relative;
  flex: 1;
  height: 154px;
}

.gridline {
  position: absolute;
  left: 0;
  right: 0;
  height: 1px;
  background: var(--hairline);
  pointer-events: none;
}

.gridline-top {
  top: 0;
}

.gridline-mid {
  top: 60px;
}

.gridline-zero {
  top: 120px;
}

.trend-cols {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: flex-start;
  justify-content: space-around;
  height: 100%;
}

.trend-col-track {
  flex: 1;
  height: 154px;
  display: flex;
  flex-direction: column;
  align-items: center;
  border-radius: var(--r-xs);
  transition: background 0.15s ease;
  cursor: pointer;
  padding: 0 2px;
}

.trend-col-track:hover {
  background: rgba(0, 0, 0, 0.035);
}

html.dark .trend-col-track:hover {
  background: rgba(255, 255, 255, 0.045);
}

.trend-col-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  height: 100%;
}

.trend-bars {
  display: flex;
  align-items: flex-end;
  justify-content: center;
  gap: 5px;
  height: 120px;
}

.trend-bar {
  width: 12px;
  border-radius: 2px 2px 0 0;
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.trend-bar:hover {
  opacity: 0.9;
  transform: scaleY(1.03);
}

.bar-merge, .dot-merge {
  background: #3b82f6 !important;
}

.bar-clean, .dot-clean {
  background: #10b981 !important;
}

.trend-date {
  font-size: 11px;
  font-family: var(--font-mono);
  color: var(--stone);
  margin-top: 10px;
  line-height: 1;
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
  margin-top: 10px;
}

.trend-legend span {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 2px;
}

/* Tooltip internal styling */
.trend-tooltip {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  padding: 2px 0;
}

.tooltip-header {
  font-weight: 600;
  margin-bottom: 2px;
  font-family: var(--font-mono);
}

.tooltip-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.tooltip-dot {
  width: 6px;
  height: 6px;
  border-radius: 2px;
  flex-shrink: 0;
}

.tooltip-val {
  font-family: var(--font-mono);
}

.tooltip-divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.15);
  margin: 4px 0;
}

.tooltip-net {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-weight: 500;
  font-family: var(--font-mono);
}

.text-inflow {
  color: #58a6ff;
}

.text-outflow {
  color: #3fb950;
}
</style>

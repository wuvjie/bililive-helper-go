<template>
  <div class="card storage-card">
    <div class="card-title">
      <div class="card-title-left">
        <div class="ui-icon-box icon-box-blue">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 12H2M5.45 5.11L2 12v6a2 2 0 002 2h16a2 2 0 002-2v-6l-3.45-6.89A2 2 0 0016.76 4H7.24a2 2 0 00-1.79 1.11z"/>
            <line x1="6" y1="16" x2="6.01" y2="16"/>
            <line x1="10" y1="16" x2="10.01" y2="16"/>
          </svg>
        </div>
        <span class="card-title-main">存储空间监控</span>
      </div>
    </div>

    <div class="card-body">
      <div class="card-top-metric">
        <div class="disk-top">
          <span class="disk-label">磁盘使用率</span>
          <span class="disk-pct mono-val" :style="{ color: diskColor }">
            {{ detail?.disk?.usage_pct?.toFixed(1) || 0 }}%
          </span>
        </div>
        <el-progress
          :percentage="detail?.disk?.usage_pct || 0"
          :color="diskColor"
          :stroke-width="6"
          :format="() => ''"
        />
      </div>

      <div class="data-grid">
        <div class="data-cell">
          <span class="data-label">总容量</span>
          <span class="mono-val">{{ detail?.disk?.total_gb?.toFixed(1) }} GB</span>
        </div>
        <div class="data-cell">
          <span class="data-label">已用</span>
          <span class="mono-val">{{ detail?.disk?.used_gb?.toFixed(1) }} GB</span>
        </div>
        <div class="data-cell">
          <span class="data-label">剩余</span>
          <span class="mono-val">{{ detail?.disk?.free_gb?.toFixed(1) }} GB</span>
        </div>
        <div class="data-cell">
          <span class="data-label">
            录播库占用
            <el-tooltip content="当前已录制的视频文件所占用的磁盘物理空间总和" placement="top">
              <span class="label-help-ico">?</span>
            </el-tooltip>
          </span>
          <span class="mono-val">{{ totalRecordingsGB }} GB</span>
        </div>
      </div>

      <div class="bottom-health-bar">
        <div class="health-cell">
          <span class="health-cell-label">近 7 天产出</span>
          <span class="health-cell-val mono-val">
            <strong>{{ stats?.daily_output_gb?.toFixed(1) || 0 }}</strong>
            <span class="unit-lbl">GB/天</span>
          </span>
        </div>
        <div class="health-cell" :class="{ 'health-cell-warn': (stats?.days_until_full || 999) < 7 }">
          <span class="health-cell-label">余量预估</span>
          <span class="health-cell-val mono-val">
            <strong>{{ (stats?.days_until_full && stats?.days_until_full > 0) ? stats.days_until_full.toFixed(0) : '充足' }}</strong>
            <span v-if="stats?.days_until_full && stats?.days_until_full > 0" class="unit-lbl">天</span>
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { DetailResponse, StatsResponse } from "@/api/types";

const props = defineProps<{
  detail?: DetailResponse;
  stats?: StatsResponse;
}>();

const diskColor = computed(() => {
  const pct = props.detail?.disk?.usage_pct || 0;
  if (pct >= 90) return "#d9730d";
  return "#448361";
});

const totalRecordingsGB = computed(() => {
  const sumBytes = props.detail?.streamers?.reduce((acc, s) => {
    if (s.size_bytes !== undefined && s.size_bytes !== null) {
      return acc + s.size_bytes;
    }
    return acc + (s.size_gb ? s.size_gb * (1024 ** 3) : 0);
  }, 0) || 0;
  return (sumBytes / (1024 ** 3)).toFixed(1);
});
</script>

<style scoped>
.storage-card {
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

/* Deconstructed top metric - hairline underline, no box */
.card-top-metric {
  margin-bottom: 16px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--hairline);
}

.disk-top {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 8px;
}

.disk-label {
  font-size: 13px;
  color: var(--slate);
  font-weight: 500;
}

.disk-pct {
  font-size: 18px;
  font-weight: 600;
}

/* Hairline Ruled Metric Grid (十字发丝线标尺网格) */
.data-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  border-bottom: 1px solid var(--hairline);
  margin-bottom: 14px;
}

.data-cell {
  background: transparent;
  border: none;
  padding: 10px 14px;
  border-radius: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.data-cell:nth-child(odd) {
  border-right: 1px solid var(--hairline);
  padding-left: 0;
}

.data-cell:nth-child(even) {
  padding-right: 0;
}

.data-cell:nth-child(-n+2) {
  border-bottom: 1px solid var(--hairline);
  padding-top: 0;
}

.data-cell:nth-child(n+3) {
  padding-bottom: 12px;
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

/* Flattened bottom health bar - no outer boxes */
.bottom-health-bar {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  padding-top: 2px;
}

.health-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 0;
  border-radius: 0;
  background: transparent;
  border: none;
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
  font-size: 15px;
  font-weight: 600;
  color: var(--charcoal);
}

.health-cell-val .unit-lbl {
  font-size: 11px;
  color: var(--stone);
  font-weight: 400;
}

.health-cell-warn strong {
  color: #d9730d;
}
</style>

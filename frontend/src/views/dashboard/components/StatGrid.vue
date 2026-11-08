<template>
  <div class="stat-grid">
    <!-- Card 1: 今日合并 -->
    <div class="stat-card accent-blue">
      <span class="stat-label">今日合并</span>
      <span class="stat-value">{{ stats?.today?.merge_count || 0 }}</span>
      <span class="stat-sub">
        产出 {{ stats?.today?.merge_bytes ? formatBytes(stats.today.merge_bytes) : '0.00 GB' }}
      </span>
    </div>

    <!-- Card 2: 7日日均产出 -->
    <div class="stat-card accent-green">
      <span class="stat-label">7日日均产出</span>
      <div class="stat-value-row">
        <span class="stat-value">{{ stats?.daily_output_gb ? stats.daily_output_gb.toFixed(1) : '0.0' }}</span>
        <span class="stat-unit">GB/天</span>
      </div>
      <span class="stat-sub">近 7 天平均录播增量</span>
    </div>

    <!-- Card 3: 存储余量预估 -->
    <div class="stat-card" :class="isRunwayUrgent ? 'accent-amber' : 'accent-sky'">
      <span class="stat-label">存储余量预估</span>
      <div class="stat-value-row">
        <span class="stat-value">{{ runwayDisplay.val }}</span>
        <span v-if="runwayDisplay.unit" class="stat-unit">{{ runwayDisplay.unit }}</span>
      </div>
      <span class="stat-sub">{{ runwayDisplay.sub }}</span>
    </div>

    <!-- Card 4: 累计清理释放 -->
    <div class="stat-card accent-purple">
      <span class="stat-label">累计清理释放</span>
      <span class="stat-value">{{ totalFreedStats.count }}</span>
      <span class="stat-sub">
        释放 {{ freedSpaceDisplay.val }} {{ freedSpaceDisplay.unit }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { formatBytes } from "@/utils/format";
import type { StatsResponse } from "@/api/types";

const props = defineProps<{
  stats?: StatsResponse;
  yesterdayStats?: { count: number; bytes: number };
  totalFreedStats: { count: number; bytes: number };
  freedSpaceDisplay: { val: string; unit: string };
}>();

const isRunwayUrgent = computed(() => {
  const days = props.stats?.days_until_full;
  return days !== undefined && days !== null && days > 0 && days < 7;
});

const runwayDisplay = computed(() => {
  const days = props.stats?.days_until_full;
  if (days && days > 0) {
    return {
      val: days.toFixed(0),
      unit: "天",
      sub: days < 7 ? "余量告急 · 请关注磁盘容量" : "基于近期写入速率测算"
    };
  }
  return {
    val: "充足",
    unit: "",
    sub: "未检测到磁盘耗尽风险"
  };
});
</script>

<style scoped>
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-card {
  position: relative;
  padding: 16px 20px;
  border-radius: var(--r-md);
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--hairline);
  box-shadow: var(--shadow-card);
  display: flex;
  flex-direction: column;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
  overflow: hidden;
}

.stat-card::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
}

.stat-card.accent-blue::before { background: #3b82f6; }
.stat-card.accent-green::before { background: #10b981; }
.stat-card.accent-sky::before { background: #0ea5e9; }
.stat-card.accent-amber::before { background: #f59e0b; }
.stat-card.accent-purple::before { background: #8b5cf6; }

.stat-card:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-card-hover);
}

.stat-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--steel);
  margin-bottom: 6px;
}

.stat-value-row {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.stat-value {
  font-size: 30px;
  font-weight: 700;
  line-height: 1.1;
  font-family: var(--font-mono);
  color: var(--ink);
  letter-spacing: -1px;
}

.stat-unit {
  font-size: 13px;
  font-weight: 600;
  color: var(--steel);
  font-family: var(--font-mono);
}

.stat-sub {
  font-size: 12px;
  color: var(--stone);
  margin-top: 6px;
}

@media (max-width: 900px) {
  .stat-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .stat-grid {
    grid-template-columns: 1fr;
  }
}
</style>

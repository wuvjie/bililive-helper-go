<template>
  <div class="stat-grid">
    <div class="stat-card accent-blue">
      <span class="stat-label">今日合并</span>
      <span class="stat-value">{{ stats?.today?.merge_count || 0 }}</span>
      <span class="stat-sub">
        产出 {{ stats?.today?.merge_bytes ? formatBytes(stats.today.merge_bytes) : '0.00 GB' }}
      </span>
    </div>
    <div class="stat-card accent-green">
      <span class="stat-label">昨日合并</span>
      <span class="stat-value">{{ yesterdayStats.count }}</span>
      <span class="stat-sub">
        产出 {{ yesterdayStats.bytes ? formatBytes(yesterdayStats.bytes) : '0.00 GB' }}
      </span>
    </div>
    <div class="stat-card accent-sky">
      <span class="stat-label">本月合并</span>
      <span class="stat-value">{{ stats?.month?.merge_count || 0 }}</span>
      <span class="stat-sub">
        产出 {{ stats?.month?.merge_bytes ? formatBytes(stats.month.merge_bytes) : '0.00 GB' }}
      </span>
    </div>
    <div class="stat-card accent-purple">
      <span class="stat-label">累计清理</span>
      <span class="stat-value">{{ totalFreedStats.count }}</span>
      <span class="stat-sub">
        释放 {{ freedSpaceDisplay.val }} {{ freedSpaceDisplay.unit }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatBytes } from "@/utils/format";
import type { StatsResponse } from "@/api/types";

defineProps<{
  stats?: StatsResponse;
  yesterdayStats: { count: number; bytes: number };
  totalFreedStats: { count: number; bytes: number };
  freedSpaceDisplay: { val: string; unit: string };
}>();
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

.stat-value {
  font-size: 32px;
  font-weight: 700;
  line-height: 1.1;
  font-family: var(--font-mono);
  color: var(--ink);
  letter-spacing: -1px;
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

<template>
  <div class="stat-grid">
    <div class="stat-card stat-peach">
      <span class="stat-label">今日合并</span>
      <span class="stat-value">{{ stats?.today?.merge_count || 0 }}</span>
      <span class="stat-sub">
        产出 {{ stats?.today?.merge_bytes ? formatBytes(stats.today.merge_bytes) : '0.00 GB' }}
      </span>
    </div>
    <div class="stat-card stat-mint">
      <span class="stat-label">昨日合并</span>
      <span class="stat-value">{{ yesterdayStats.count }}</span>
      <span class="stat-sub">
        产出 {{ yesterdayStats.bytes ? formatBytes(yesterdayStats.bytes) : '0.00 GB' }}
      </span>
    </div>
    <div class="stat-card stat-sky">
      <span class="stat-label">本月合并</span>
      <span class="stat-value">{{ stats?.month?.merge_count || 0 }}</span>
      <span class="stat-sub">
        产出 {{ stats?.month?.merge_bytes ? formatBytes(stats.month.merge_bytes) : '0.00 GB' }}
      </span>
    </div>
    <div class="stat-card stat-lavender">
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
  padding: 16px 20px;
  border-radius: var(--r-md);
  display: flex;
  flex-direction: column;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
  border: 1px solid transparent;
}

.stat-card:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.stat-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--charcoal);
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

/* Pastel tints */
.stat-peach {
  background: #fdf5f0;
  border-color: #f7e1d2;
}
.stat-peach .stat-label { color: #8f4c28; }

.stat-mint {
  background: #f0f7f3;
  border-color: #d1ead9;
}
.stat-mint .stat-label { color: #2d6a4f; }

.stat-sky {
  background: #f0f5fa;
  border-color: #d2e3f3;
}
.stat-sky .stat-label { color: #235d8a; }

.stat-lavender {
  background: #f5f2f9;
  border-color: #dfd5ed;
}
.stat-lavender .stat-label { color: #5a3d7a; }

/* Dark mode overrides */
html.dark .stat-peach {
  background: rgba(217, 115, 13, 0.1);
  border-color: rgba(217, 115, 13, 0.2);
}
html.dark .stat-peach .stat-label { color: #fb923c; }

html.dark .stat-mint {
  background: rgba(68, 131, 97, 0.1);
  border-color: rgba(68, 131, 97, 0.2);
}
html.dark .stat-mint .stat-label { color: #4ade80; }

html.dark .stat-sky {
  background: rgba(59, 130, 246, 0.1);
  border-color: rgba(59, 130, 246, 0.2);
}
html.dark .stat-sky .stat-label { color: #60a5fa; }

html.dark .stat-lavender {
  background: rgba(147, 51, 234, 0.1);
  border-color: rgba(147, 51, 234, 0.2);
}
html.dark .stat-lavender .stat-label { color: #c084fc; }

@media (max-width: 900px) {
  .stat-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 540px) {
  .stat-grid {
    grid-template-columns: 1fr;
  }
}
</style>

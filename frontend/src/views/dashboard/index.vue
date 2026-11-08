<template>
  <div class="dashboard" v-loading="loading" element-loading-text="加载中...">
    <!-- Header with inline environment status -->
    <div class="dash-header">
      <div>
        <h1>{{ greeting }}</h1>
        <p class="dash-date">{{ todayDate }}</p>
      </div>
      <div class="env-status">
        <div class="env-item">
          <span class="env-dot" :class="setup?.ffmpeg_ok ? 'dot-ok' : 'dot-err'"></span>
          <span class="env-label">FFmpeg</span>
          <span class="env-val" :class="setup?.ffmpeg_ok ? 'val-ok' : 'val-err'">{{ setup?.ffmpeg_ok ? '正常' : '异常' }}</span>
        </div>
        <span class="env-divider"></span>
        <div class="env-item">
          <span class="env-dot" :class="setup?.target_dir_exists ? 'dot-ok' : 'dot-err'"></span>
          <span class="env-label">存储</span>
          <span class="env-val" :class="setup?.target_dir_exists ? 'val-ok' : 'val-err'">{{ setup?.target_dir_exists ? '正常' : '异常' }}</span>
        </div>
      </div>
    </div>

    <!-- Stat Cards with pastel tints -->
    <StatGrid
      :stats="stats"
      :yesterday-stats="yesterdayStats"
      :total-freed-stats="totalFreedStats"
      :freed-space-display="freedSpaceDisplay"
    />

    <!-- Main Grid 2x2 -->
    <div class="grid-2">
      <!-- 1. Disk Storage Status -->
      <StorageCard :detail="detail" :stats="stats" />

      <!-- 2. Runtime & Automation Monitor -->
      <ScheduleMonitorCard :detail="detail" :setup="setup" />

      <!-- 3. Trend Chart -->
      <TrendCard :stats="stats" />

      <!-- 4. Recent Operations -->
      <RecentOpsCard :recent-history="recentHistory" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onActivated } from "vue";
import { getStatusDetail, getStats } from "@/api/status";
import { getHistory } from "@/api/history";
import { setupCheck } from "@/api/setup";
import type { DetailResponse, StatsResponse, SetupCheck, HistoryRecord } from "@/api/types";

import StatGrid from "./components/StatGrid.vue";
import StorageCard from "./components/StorageCard.vue";
import ScheduleMonitorCard from "./components/ScheduleMonitorCard.vue";
import TrendCard from "./components/TrendCard.vue";
import RecentOpsCard from "./components/RecentOpsCard.vue";

const detail = ref<DetailResponse>();
const stats = ref<StatsResponse>();
const setup = ref<SetupCheck>();
const recentHistory = ref<HistoryRecord[]>([]);
const loading = ref(true);

const yesterdayStats = computed(() => {
  if (stats.value?.yesterday) {
    return {
      count: stats.value.yesterday.merge_count || 0,
      bytes: stats.value.yesterday.merge_bytes || 0
    };
  }
  const daily = stats.value?.daily;
  if (daily && daily.length >= 2) {
    const yDay = daily[daily.length - 2];
    return {
      count: yDay?.merge_count || 0,
      bytes: yDay?.merge_bytes || 0
    };
  }
  return { count: 0, bytes: 0 };
});

const totalFreedStats = computed(() => {
  if (stats.value?.total?.clean_bytes !== undefined) {
    return {
      count: stats.value.total.clean_count || 0,
      bytes: stats.value.total.clean_bytes || 0
    };
  }
  const daily = stats.value?.daily;
  if (daily && daily.length > 0) {
    const sumBytes = daily.reduce((acc, d) => acc + (d.clean_bytes || 0), 0);
    const sumCount = daily.reduce((acc, d) => acc + (d.clean_count || 0), 0);
    return {
      count: Math.max(stats.value?.month?.clean_count || 0, sumCount),
      bytes: Math.max(stats.value?.month?.clean_bytes || 0, sumBytes)
    };
  }
  return {
    count: stats.value?.month?.clean_count || 0,
    bytes: stats.value?.month?.clean_bytes || 0
  };
});

const freedSpaceDisplay = computed(() => {
  const bytes = totalFreedStats.value.bytes || 0;
  if (bytes <= 0) {
    return { val: "0.00", unit: "GB" };
  }
  const gb = bytes / (1024 ** 3);
  if (gb >= 0.01) {
    return { val: gb.toFixed(2), unit: "GB" };
  }
  const mb = bytes / (1024 ** 2);
  if (mb >= 0.1) {
    return { val: mb.toFixed(1), unit: "MB" };
  }
  return { val: (bytes / 1024).toFixed(0), unit: "KB" };
});

const greeting = computed(() => {
  const now = new Date();
  const h = now.getHours();
  const d = now.getDate();

  const pick = (list: string[]) => list[d % list.length];

  if (h < 6) {
    return pick([
      "夜深了，早点休息，明天又是新的一天",
      "夜深了，注意身体，早点睡个好觉"
    ]);
  }
  if (h < 9) {
    return pick([
      "早上好，新的一天开始啦",
      "早上好，晨光正好，元气满满",
      "早上好，迎接充满希望的新一天"
    ]);
  }
  if (h < 12) {
    return pick([
      "上午好，今天也是充实专注的一天",
      "上午好，保持专注，开启高效时光",
      "上午好，愿你有顺心顺遂的一天"
    ]);
  }
  if (h < 14) {
    return pick([
      "中午好，辛苦了，记得好好吃顿午饭",
      "中午好，午间小憩，适度放松一下"
    ]);
  }
  if (h < 18) {
    return pick([
      "下午好，喝杯水，歇歇眼睛吧",
      "下午好，来杯茶，保持好状态",
      "下午好，劳逸结合，事半功倍"
    ]);
  }
  return pick([
    "晚上好，愿你度过轻松惬意的夜晚",
    "晚上好，卸下疲惫，享受属于自己的时光",
    "晚上好，静享晚间安宁与自在"
  ]);
});

const todayDate = computed(() =>
  new Date().toLocaleDateString("zh-CN", { year: "numeric", month: "long", day: "numeric", weekday: "long" })
);

async function loadData() {
  try {
    const [d, s, h, ck] = await Promise.allSettled([
      getStatusDetail(), getStats(), getHistory({ per_page: 5 }), setupCheck()
    ]);
    if (d.status === "fulfilled") detail.value = d.value;
    if (s.status === "fulfilled") stats.value = s.value;
    if (h.status === "fulfilled") recentHistory.value = h.value.items || [];
    if (ck.status === "fulfilled") setup.value = ck.value;
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  loadData();
});

onActivated(async () => {
  try {
    const [d, s, h] = await Promise.allSettled([
      getStatusDetail(), getStats(), getHistory({ per_page: 5 })
    ]);
    if (d.status === "fulfilled") detail.value = d.value;
    if (s.status === "fulfilled") stats.value = s.value;
    if (h.status === "fulfilled") recentHistory.value = h.value.items || [];
  } catch { /* ignore */ }
});
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-height: 400px;
}

/* Header */
.dash-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}

.dash-header h1 {
  font-size: 26px;
  font-weight: 600;
  color: var(--ink);
  letter-spacing: -0.5px;
}

.dash-date {
  font-size: 12px;
  color: #888888;
  margin-top: 2px;
  font-family: var(--font-mono);
  letter-spacing: 0.2px;
}

/* Environment status inline in header */
.env-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-bottom: 4px;
}

.env-item {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
}

.env-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
  flex-shrink: 0;
}

.env-label {
  color: var(--stone);
  font-size: 12px;
}

.env-val {
  font-size: 12px;
  font-weight: 500;
}

.val-ok { color: #448361; }
.val-err { color: #e03131; }
.env-divider { width: 1px; height: 12px; background: var(--hairline); }

.grid-2 {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.grid-2 > * {
  min-width: 0;
}

@media (max-width: 900px) {
  .grid-2 {
    grid-template-columns: 1fr;
  }
}
</style>

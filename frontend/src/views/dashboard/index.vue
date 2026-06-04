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
    <div class="stat-grid">
      <div class="stat-card stat-peach">
        <span class="stat-label">今日合并</span>
        <span class="stat-value">{{ stats?.today.merge_count || 0 }}</span>
        <span class="stat-sub" :class="{ 'stat-empty': !stats?.today.merge_bytes }">{{ stats?.today.merge_bytes ? formatBytes(stats.today.merge_bytes) : '0.00 GB' }}</span>
      </div>
      <div class="stat-card stat-mint">
        <span class="stat-label">今日清理</span>
        <span class="stat-value">{{ stats?.today.clean_count || 0 }}</span>
        <span class="stat-sub" :class="{ 'stat-empty': !stats?.today.clean_bytes }">{{ stats?.today.clean_bytes ? formatBytes(stats.today.clean_bytes) : '0.00 GB' }}</span>
      </div>
      <div class="stat-card stat-sky">
        <span class="stat-label">本月合并</span>
        <span class="stat-value">{{ stats?.month.merge_count || 0 }}</span>
        <span class="stat-sub" :class="{ 'stat-empty': !stats?.month.merge_bytes }">{{ stats?.month.merge_bytes ? formatBytes(stats.month.merge_bytes) : '0.00 GB' }}</span>
      </div>
      <div class="stat-card stat-lavender">
        <span class="stat-label">本月清理</span>
        <span class="stat-value">{{ stats?.month.clean_count || 0 }}</span>
        <span class="stat-sub" :class="{ 'stat-empty': !stats?.month.clean_bytes }">{{ stats?.month.clean_bytes ? formatBytes(stats.month.clean_bytes) : '0.00 GB' }}</span>
      </div>
    </div>

    <div class="grid-2">
      <!-- Disk Usage -->
      <div class="card">
        <div class="card-title">💾 存储状态</div>
        <div class="card-body">
          <div class="disk-top">
            <span class="disk-label">磁盘使用率</span>
            <span class="disk-pct mono-val" :style="{ color: diskColor }">{{ detail?.disk?.usage_pct?.toFixed(1) || 0 }}%</span>
          </div>
          <el-progress
            :percentage="detail?.disk?.usage_pct || 0"
            :color="diskColor"
            :stroke-width="6"
            :format="() => ''"
            style="margin-bottom: 16px"
          />
          <div class="data-grid">
            <div class="data-cell"><span class="data-label">总容量</span><span class="mono-val">{{ detail?.disk?.total_gb?.toFixed(1) }} GB</span></div>
            <div class="data-cell"><span class="data-label">已用</span><span class="mono-val">{{ detail?.disk?.used_gb?.toFixed(1) }} GB</span></div>
            <div class="data-cell"><span class="data-label">剩余</span><span class="mono-val">{{ detail?.disk?.free_gb?.toFixed(1) }} GB</span></div>
            <div class="data-cell"><span class="data-label">待合并</span><span class="mono-val">{{ detail?.pending?.original_files || 0 }} 个文件</span></div>
          </div>
        </div>
      </div>

      <!-- System Status -->
      <div class="card">
        <div class="card-title">🖥️ 系统状态</div>
        <div class="card-body sys-body">
          <div class="sys-row">
            <span class="sys-label">FFmpeg</span>
            <span class="sys-val"><span class="status-dot-sm" :class="setup?.ffmpeg_ok ? 'dot-ok' : 'dot-err'" />{{ setup?.ffmpeg_ok ? '正常' : '异常' }}</span>
          </div>
          <div class="sys-row">
            <span class="sys-label">FFprobe</span>
            <span class="sys-val"><span class="status-dot-sm" :class="setup?.ffprobe_ok ? 'dot-ok' : 'dot-err'" />{{ setup?.ffprobe_ok ? '正常' : '异常' }}</span>
          </div>
          <div class="sys-row">
            <span class="sys-label">目标目录</span>
            <span class="sys-val"><span class="status-dot-sm" :class="setup?.target_dir_exists ? 'dot-ok' : 'dot-err'" />{{ setup?.target_dir_exists ? '存在' : '不存在' }}</span>
          </div>
          <div class="sys-row">
            <span class="sys-label">主播数</span>
            <span class="mono-val">{{ setup?.streamer_count || 0 }}</span>
          </div>
          <div class="sys-row sys-row-last">
            <span class="sys-label">视频数</span>
            <span class="mono-val">{{ setup?.video_count || 0 }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="grid-2">
      <!-- Trend -->
      <div class="card">
        <div class="card-title">📊 近7天趋势</div>
        <div class="card-body trend-body">
          <div class="trend-chart">
            <div v-for="day in stats?.daily" :key="day.date" class="trend-col">
              <div class="trend-bars">
                <el-tooltip :content="`合并 ${formatBytes(day.merge_bytes)}`" placement="top">
                  <div class="trend-bar bar-merge" :style="{ height: barHeight(day.merge_bytes) + 'px' }" />
                </el-tooltip>
                <el-tooltip :content="`释放 ${formatBytes(day.clean_bytes)}`" placement="top">
                  <div class="trend-bar bar-clean" :style="{ height: barHeight(day.clean_bytes) + 'px' }" />
                </el-tooltip>
              </div>
              <span class="trend-date">{{ day.date.slice(5) }}</span>
            </div>
          </div>
          <div class="trend-legend">
            <span><span class="legend-dot" style="background: #787774" />合并量</span>
            <span><span class="legend-dot" style="background: var(--success)" />释放量</span>
          </div>
        </div>
      </div>

      <!-- Recent -->
      <div class="card">
        <div class="card-title">
          <span>📋 最近操作</span>
          <button class="view-all-btn" @click="$router.push('/history')">查看全部</button>
        </div>
        <div class="card-body recent-body">
          <div v-if="recentHistory.length === 0" class="recent-empty">暂无操作记录</div>
          <div
            v-for="row in recentHistory"
            :key="row.id"
            class="recent-row"
          >
            <div class="recent-left">
              <span class="recent-icon" :class="row.status === 'fail' ? 'icon-error' : ''">
                {{ row.task === 'merge' ? '🔗' : row.task === 'clean' ? '🧹' : row.task === 'config' ? '⚙️' : '⏱️' }}
              </span>
              <span class="recent-who">{{ row.streamer || '全局' }}</span>
              <span class="recent-action">{{ row.detail || formatDetail(row) }}</span>
            </div>
            <div class="recent-right">
              <span v-if="row.status === 'fail'" class="recent-fail">失败</span>
              <span class="recent-time">{{ formatShortTime(row.time) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onActivated } from "vue";
import { getStatusDetail, getStats } from "@/api/status";
import { getHistory } from "@/api/history";
import { setupCheck } from "@/api/setup";
import { formatBytes } from "@/utils/format";
import type { DetailResponse, StatsResponse, SetupCheck, HistoryRecord } from "@/api/types";

const detail = ref<DetailResponse>();
const stats = ref<StatsResponse>();
const setup = ref<SetupCheck>();
const recentHistory = ref<HistoryRecord[]>([]);
const loading = ref(true);

const greeting = computed(() => {
  const h = new Date().getHours();
  if (h < 6) return "夜深了，注意休息";
  if (h < 12) return "早上好";
  if (h < 14) return "中午好";
  if (h < 18) return "下午好";
  return "晚上好";
});

const todayDate = computed(() =>
  new Date().toLocaleDateString("zh-CN", { year: "numeric", month: "long", day: "numeric", weekday: "long" })
);

const diskColor = computed(() => {
  const pct = detail.value?.disk?.usage_pct || 0;
  if (pct >= 90) return "#d9730d"; // Notion warm amber for critical
  return "#448361"; // Notion muted botanical green
});

const maxBytes = computed(() => {
  if (!stats.value?.daily) return 1;
  return Math.max(...stats.value.daily.map(d => Math.max(d.merge_bytes, d.clean_bytes)), 1);
});

function barHeight(b: number): number { return Math.max(4, (b / maxBytes.value) * 100); }
function formatShortTime(t?: string): string {
  if (!t || t.length < 16) return "";
  return t.slice(11, 16); // "HH:mm"
}
function formatDetail(row: HistoryRecord) {
  const p = [];
  if (row.files_count) p.push(`${row.files_count}个文件`);
  if (row.duration) p.push(`${row.duration}s`);
  return p.join(", ") || "—";
}

onMounted(async () => {
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
});

// Refresh data when component is re-activated by keep-alive router-view
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
.dashboard { display: flex; flex-direction: column; gap: 20px; min-height: 400px; }

/* Header */
.dash-header { display: flex; justify-content: space-between; align-items: flex-end; }
.dash-header h1 { font-size: 26px; font-weight: 600; color: var(--ink); letter-spacing: -0.5px; }
.dash-date { font-size: 12px; color: #888888; margin-top: 2px; font-family: var(--font-mono); letter-spacing: 0.2px; }

/* Environment status — inline in header */
.env-status { display: flex; align-items: center; gap: 8px; padding-bottom: 4px; }
.env-item { display: flex; align-items: center; gap: 5px; font-size: 12px; }
.env-dot { width: 6px; height: 6px; border-radius: 50%; display: inline-block; flex-shrink: 0; }
.dot-ok { background: #448361; }
.dot-err { background: #e03131; }
.env-label { color: var(--stone); font-size: 11px; }
.env-val { font-size: 11px; font-weight: 500; }
.val-ok { color: #448361; }
.val-err { color: #e03131; }
.env-divider { width: 1px; height: 12px; background: var(--hairline); }

/* Stat cards with Notion pastel tints */
.stat-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; }

.stat-card {
  border-radius: var(--r-md); padding: 20px;
  display: flex; flex-direction: column;
}
.stat-peach { background: var(--tint-peach); border: 1px solid #f5d4b8; }
.stat-mint { background: var(--tint-mint); border: 1px solid #b8e6c8; }
.stat-sky { background: var(--tint-sky); border: 1px solid #b8d8f5; }
.stat-lavender { background: var(--tint-lavender); border: 1px solid #d0c8e8; }

.stat-label { font-size: 13px; color: var(--slate); margin-bottom: 4px; }
.stat-value { font-size: 32px; font-weight: 600; color: var(--ink); line-height: 1; letter-spacing: -0.5px; font-family: var(--font-mono); height: 36px; display: flex; align-items: flex-end; }
.stat-sub { font-size: 13px; color: var(--steel); margin-top: 2px; font-family: var(--font-mono); }
.stat-empty { color: #c7c7cc; opacity: 0.4; }

/* Cards */
.grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

.card {
  background: var(--canvas); border: 1px solid var(--hairline);
  border-radius: var(--r-md); overflow: hidden;
  box-shadow: none;
}

.card-title {
  padding: 14px 20px; font-size: 14px; font-weight: 600; color: var(--ink);
  border-bottom: 1px solid var(--hairline);
  display: flex; justify-content: space-between; align-items: center;
}
.card-body { padding: 20px; }

/* Disk */
.disk-top { display: flex; justify-content: space-between; margin-bottom: 8px; }
.disk-label { font-size: 14px; color: var(--slate); }
.disk-pct { font-size: 20px; font-weight: 600; }
.mono-val { font-family: var(--font-mono); font-size: 13px; color: var(--charcoal); }

/* Data grid — 2×2 with hairline dividers, strict equal columns */
.data-grid {
  display: grid; grid-template-columns: repeat(2, 1fr);
  border: 1px solid var(--hairline); border-radius: var(--r-sm); overflow: hidden;
}
.data-cell {
  display: flex; align-items: center; gap: 12px;
  padding: 10px 14px;
  background: var(--canvas);
  min-width: 0; /* Prevent content from stretching column */
}
.data-cell:nth-child(odd) { border-right: 1px solid var(--hairline); }
.data-cell:nth-child(1), .data-cell:nth-child(2) { border-bottom: 1px solid var(--hairline); }
.data-label { font-size: 13px; color: var(--steel); flex-shrink: 0; }

/* System status — clean rows with dot indicators */
.sys-body { display: flex; flex-direction: column; }
.sys-row {
  display: flex; justify-content: space-between; align-items: center;
  padding: 8.5px 0;
  border-bottom: 1px solid #f1f1ef;
}
.sys-row-last { border-bottom: 1px solid #f1f1ef; }
.sys-label { font-size: 13px; color: var(--ink); min-width: 72px; }
.sys-val { display: flex; align-items: center; gap: 10px; font-size: 13px; color: var(--steel); font-weight: 500; }
.status-dot-sm { width: 6px; height: 6px; border-radius: 50%; display: inline-block; flex-shrink: 0; }
.dot-ok { background: #448361; }
.dot-err { background: #e03131; }

/* Trend */
.trend-body {
  display: flex; flex-direction: column;
  justify-content: flex-end;
  padding-top: 28px;
  min-height: 220px;
}
.trend-chart { display: flex; justify-content: space-around; align-items: flex-end; height: 120px; width: 100%; }
.trend-col { display: flex; flex-direction: column; align-items: center; gap: 6px; }
.trend-bars { display: flex; gap: 2px; align-items: flex-end; }
.trend-bar { width: 10px; border-radius: 2px 2px 0 0; min-height: 4px; transition: height 0.3s; cursor: pointer; }
.trend-bar:hover { opacity: 0.75; }
.bar-merge { background: #787774; }
.bar-clean { background: var(--success); }
.trend-date { font-size: 12px; color: var(--stone); font-family: var(--font-mono); letter-spacing: 0.2px; }
.trend-legend { display: flex; gap: 16px; justify-content: center; margin-top: 12px; font-size: 13px; color: var(--steel); }
.legend-dot { display: inline-block; width: 10px; height: 10px; border-radius: 3px; margin-right: 4px; vertical-align: middle; }

/* Recent — single-line event flow */
.view-all-btn {
  background: transparent; border: none; padding: 2px 8px;
  font-size: 13px; color: var(--steel); cursor: pointer;
  border-radius: var(--r-sm); transition: all 0.15s; font-weight: 500;
}
.view-all-btn:hover { color: var(--ink); background: var(--highlight); }

.recent-body { display: flex; flex-direction: column; gap: 2px; padding-bottom: 4px; }
.recent-empty { color: var(--stone); font-size: 13px; text-align: center; padding: 40px 0; }
.recent-row {
  display: flex; align-items: center; justify-content: space-between;
  padding: 8px 8px; border-radius: var(--r-sm);
  transition: background 0.1s;
}
.recent-row:hover { background: var(--highlight); }
.recent-left { display: flex; align-items: center; gap: 8px; font-size: 13px; min-width: 0; }
.recent-icon { font-size: 14px; flex-shrink: 0; }
.recent-icon.icon-error { filter: none; }
.recent-who { color: #8a8a8a; flex-shrink: 0; min-width: 56px; }
.recent-action { color: var(--charcoal); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.recent-right { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
.recent-fail { font-size: 12px; color: #d4433a; font-weight: 500; }
.recent-time { font-family: var(--font-mono); font-size: 11px; color: #aeaeb2; min-width: 40px; text-align: right; }

@media (max-width: 1024px) { .stat-grid { grid-template-columns: repeat(2, 1fr); } .grid-2 { grid-template-columns: 1fr; } }
@media (max-width: 600px) { .stat-grid { grid-template-columns: 1fr; } .dash-header { flex-direction: column; align-items: flex-start; gap: 8px; } }
</style>

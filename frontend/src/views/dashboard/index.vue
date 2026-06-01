<template>
  <div class="dashboard">
    <!-- Header -->
    <div class="dash-header">
      <div>
        <h1>{{ greeting }}</h1>
        <p class="dash-date">{{ todayDate }}</p>
      </div>
      <div class="dash-status">
        <span class="status-pill" :class="setup?.ffmpeg_ok ? 'status-ok' : 'status-err'">
          FFmpeg {{ setup?.ffmpeg_ok ? 'OK' : 'ERR' }}
        </span>
        <span class="status-pill" :class="setup?.target_dir_exists ? 'status-ok' : 'status-err'">
          Storage {{ setup?.target_dir_exists ? 'OK' : 'ERR' }}
        </span>
      </div>
    </div>

    <!-- Stat Cards -->
    <div class="stat-grid">
      <div v-for="card in statCards" :key="card.label" class="stat-card">
        <div class="stat-card__icon">{{ card.icon }}</div>
        <div class="stat-card__body">
          <span class="stat-card__label">{{ card.label }}</span>
          <span class="stat-card__value">{{ card.value }}</span>
          <span class="stat-card__sub">{{ card.sub }}</span>
        </div>
      </div>
    </div>

    <div class="grid-2">
      <!-- Disk Usage -->
      <div class="panel">
        <div class="panel-header">存储状态</div>
        <div class="panel-body">
          <div class="disk-row">
            <span class="disk-label">磁盘使用率</span>
            <span class="disk-pct" :style="{ color: diskColor }">{{ detail?.disk?.usage_pct?.toFixed(1) || 0 }}%</span>
          </div>
          <div class="disk-bar-bg">
            <div class="disk-bar" :style="{ width: (detail?.disk?.usage_pct || 0) + '%', background: diskColor }" />
          </div>
          <div class="disk-meta">
            <span>总容量 {{ detail?.disk?.total_gb?.toFixed(1) }} GB</span>
            <span>已用 {{ detail?.disk?.used_gb?.toFixed(1) }} GB</span>
            <span>剩余 {{ detail?.disk?.free_gb?.toFixed(1) }} GB</span>
            <span>待合并 {{ detail?.pending?.original_files || 0 }} 个文件</span>
          </div>
        </div>
      </div>

      <!-- System Status -->
      <div class="panel">
        <div class="panel-header">系统状态</div>
        <div class="panel-body">
          <div class="status-list">
            <div v-for="item in statusItems" :key="item.label" class="status-row">
              <span class="status-dot" :class="item.ok ? 'dot-ok' : 'dot-err'" />
              <span class="status-label">{{ item.label }}</span>
              <span class="status-value">{{ item.value }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="grid-2">
      <!-- Trend -->
      <div class="panel">
        <div class="panel-header">近7天趋势</div>
        <div class="panel-body">
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
            <span><span class="legend-dot bar-merge" />合并量</span>
            <span><span class="legend-dot bar-clean" />释放量</span>
          </div>
        </div>
      </div>

      <!-- Recent Operations -->
      <div class="panel">
        <div class="panel-header">
          <span>最近操作</span>
          <button class="link-btn" @click="$router.push('/history')">查看全部</button>
        </div>
        <div class="panel-body">
          <div v-for="row in recentHistory" :key="row.id" class="recent-row">
            <span class="recent-icon">{{ row.task === 'merge' ? '↗' : row.task === 'clean' ? '↙' : '⚙' }}</span>
            <div class="recent-info">
              <span class="recent-name">{{ row.streamer || '系统' }}</span>
              <span class="recent-detail">{{ row.detail || formatDetail(row) }}</span>
            </div>
            <span class="status-pill" :class="row.status === 'success' ? 'status-ok' : 'status-err'" style="font-size: 11px">
              {{ row.status === 'success' ? 'OK' : 'ERR' }}
            </span>
          </div>
          <div v-if="recentHistory.length === 0" class="empty-text">暂无操作记录</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { getStatusDetail, getStats } from "@/api/status";
import { getHistory } from "@/api/history";
import { setupCheck } from "@/api/setup";
import type { DetailResponse, StatsResponse, SetupCheck, HistoryRecord } from "@/api/types";

const detail = ref<DetailResponse>();
const stats = ref<StatsResponse>();
const setup = ref<SetupCheck>();
const recentHistory = ref<HistoryRecord[]>([]);

const greeting = computed(() => {
  const h = new Date().getHours();
  if (h < 6) return "夜深了";
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
  if (pct >= 80) return "var(--danger)";
  if (pct >= 60) return "var(--warning)";
  return "var(--success)";
});

const statCards = computed(() => {
  const s = stats.value;
  if (!s) return [];
  return [
    { label: "今日合并", value: s.today.merge_count, sub: s.today.merge_bytes > 0 ? formatBytes(s.today.merge_bytes) : "—", icon: "↗" },
    { label: "今日清理", value: s.today.clean_count, sub: s.today.clean_bytes > 0 ? formatBytes(s.today.clean_bytes) : "—", icon: "↙" },
    { label: "本月合并", value: s.month.merge_count, sub: s.month.merge_bytes > 0 ? formatBytes(s.month.merge_bytes) : "—", icon: "↗" },
    { label: "本月清理", value: s.month.clean_count, sub: s.month.clean_bytes > 0 ? formatBytes(s.month.clean_bytes) : "—", icon: "↙" }
  ];
});

const statusItems = computed(() => [
  { label: "FFmpeg", value: setup.value?.ffmpeg_ok ? "正常" : "异常", ok: setup.value?.ffmpeg_ok },
  { label: "FFprobe", value: setup.value?.ffprobe_ok ? "正常" : "异常", ok: setup.value?.ffprobe_ok },
  { label: "目标目录", value: setup.value?.target_dir_exists ? "存在" : "不存在", ok: setup.value?.target_dir_exists },
  { label: "主播数", value: `${setup.value?.streamer_count || 0}`, ok: true },
  { label: "视频数", value: `${setup.value?.video_count || 0}`, ok: true },
  { label: "总大小", value: `${(setup.value?.total_size_gb || 0).toFixed(1)} GB`, ok: true }
]);

const maxBytes = computed(() => {
  if (!stats.value?.daily) return 1;
  return Math.max(...stats.value.daily.map(d => Math.max(d.merge_bytes, d.clean_bytes)), 1);
});

function formatBytes(bytes: number): string {
  if (!bytes) return "0 B";
  const gb = bytes / 1024 ** 3;
  if (gb >= 1) return `${gb.toFixed(2)} GB`;
  return `${(bytes / 1024 ** 2).toFixed(1)} MB`;
}

function barHeight(bytes: number): number {
  return Math.max(3, (bytes / maxBytes.value) * 100);
}

function formatDetail(row: HistoryRecord) {
  const parts = [];
  if (row.files_count) parts.push(`${row.files_count}个文件`);
  if (row.duration) parts.push(`${row.duration}s`);
  return parts.join(", ") || "—";
}

onMounted(async () => {
  const [d, s, h, ck] = await Promise.allSettled([
    getStatusDetail(), getStats(), getHistory({ per_page: 5 }), setupCheck()
  ]);
  if (d.status === "fulfilled") detail.value = d.value;
  if (s.status === "fulfilled") stats.value = s.value;
  if (h.status === "fulfilled") recentHistory.value = h.value.items || [];
  if (ck.status === "fulfilled") setup.value = ck.value;
});
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Header */
.dash-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}
.dash-header h1 {
  font-family: var(--font-display);
  font-size: 24px;
  font-weight: 600;
  letter-spacing: -0.6px;
  color: var(--ink);
}
.dash-date { font-size: 13px; color: var(--ink-subtle); margin-top: 2px; }
.dash-status { display: flex; gap: 8px; }

/* Status pill */
.status-pill {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: var(--r-pill);
  font-size: 12px;
  font-weight: 500;
  background: var(--surface-2);
  color: var(--ink-muted);
}
.status-ok { color: var(--success); }
.status-err { color: var(--danger); }

/* Stat cards */
.stat-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; }

.stat-card {
  background: var(--surface-1);
  border: 1px solid var(--hairline);
  border-radius: var(--r-lg);
  padding: 18px 20px;
  display: flex;
  gap: 14px;
  align-items: center;
  transition: border-color 0.15s;
}
.stat-card:hover { border-color: var(--hairline-strong); }

.stat-card__icon {
  width: 40px;
  height: 40px;
  border-radius: var(--r-md);
  background: var(--surface-2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}

.stat-card__body { display: flex; flex-direction: column; }
.stat-card__label { font-size: 12px; color: var(--ink-subtle); }
.stat-card__value {
  font-family: var(--font-display);
  font-size: 24px;
  font-weight: 600;
  letter-spacing: -0.5px;
  color: var(--ink);
  line-height: 1.2;
}
.stat-card__sub { font-size: 12px; color: var(--ink-tertiary); margin-top: 1px; }

/* Panels */
.grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }

.panel {
  background: var(--surface-1);
  border: 1px solid var(--hairline);
  border-radius: var(--r-lg);
  overflow: hidden;
}

.panel-header {
  padding: 14px 20px;
  font-size: 13px;
  font-weight: 500;
  color: var(--ink-muted);
  border-bottom: 1px solid var(--hairline);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.panel-body { padding: 18px 20px; }

.link-btn {
  background: none;
  border: none;
  color: var(--primary);
  font-size: 13px;
  cursor: pointer;
  padding: 0;
  font-family: inherit;
}
.link-btn:hover { color: var(--primary-hover); }

/* Disk */
.disk-row { display: flex; justify-content: space-between; margin-bottom: 10px; }
.disk-label { font-size: 13px; color: var(--ink-subtle); }
.disk-pct { font-family: var(--font-display); font-size: 18px; font-weight: 600; letter-spacing: -0.3px; }

.disk-bar-bg {
  height: 6px;
  background: var(--surface-3);
  border-radius: var(--r-sm);
  overflow: hidden;
  margin-bottom: 14px;
}
.disk-bar {
  height: 100%;
  border-radius: var(--r-sm);
  transition: width 0.5s ease;
}

.disk-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  font-size: 12px;
  color: var(--ink-subtle);
}

/* Status list */
.status-list { display: flex; flex-direction: column; gap: 12px; }
.status-row { display: flex; align-items: center; gap: 10px; }
.status-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.dot-ok { background: var(--success); }
.dot-err { background: var(--danger); }
.status-label { font-size: 13px; color: var(--ink-subtle); min-width: 60px; }
.status-value { font-size: 13px; color: var(--ink-muted); font-weight: 500; }

/* Trend */
.trend-chart {
  display: flex;
  justify-content: space-around;
  align-items: flex-end;
  height: 100px;
}
.trend-col { display: flex; flex-direction: column; align-items: center; gap: 6px; }
.trend-bars { display: flex; gap: 4px; align-items: flex-end; }
.trend-bar {
  width: 14px;
  border-radius: 3px 3px 0 0;
  min-height: 3px;
  transition: height 0.3s ease;
  cursor: pointer;
}
.trend-bar:hover { opacity: 0.75; }
.bar-merge { background: var(--primary); }
.bar-clean { background: var(--success); }
.trend-date { font-size: 11px; color: var(--ink-tertiary); }

.trend-legend {
  display: flex;
  gap: 20px;
  justify-content: center;
  margin-top: 14px;
  font-size: 12px;
  color: var(--ink-subtle);
}
.legend-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 2px;
  margin-right: 6px;
  vertical-align: middle;
}

/* Recent operations */
.recent-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 1px solid var(--hairline);
}
.recent-row:last-child { border-bottom: none; }

.recent-icon {
  width: 28px;
  height: 28px;
  border-radius: var(--r-sm);
  background: var(--surface-2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  flex-shrink: 0;
}

.recent-info { flex: 1; min-width: 0; }
.recent-name { font-size: 13px; font-weight: 500; color: var(--ink); }
.recent-detail { display: block; font-size: 12px; color: var(--ink-subtle); margin-top: 1px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.empty-text { font-size: 13px; color: var(--ink-tertiary); text-align: center; padding: 24px 0; }

@media (max-width: 1024px) {
  .stat-grid { grid-template-columns: repeat(2, 1fr); }
  .grid-2 { grid-template-columns: 1fr; }
}
@media (max-width: 600px) {
  .stat-grid { grid-template-columns: 1fr; }
  .dash-header { flex-direction: column; align-items: flex-start; gap: 10px; }
}
</style>

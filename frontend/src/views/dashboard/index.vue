<template>
  <div class="dashboard">
    <!-- Header -->
    <div class="dash-header">
      <div>
        <h1>{{ greeting }}.</h1>
        <p class="dash-date">{{ todayDate }}</p>
      </div>
      <div class="dash-badges">
        <span class="badge" :class="setup?.ffmpeg_ok ? 'badge-ok' : 'badge-err'">
          <span class="badge-dot" :class="setup?.ffmpeg_ok ? 'dot-ok' : 'dot-err'" />
          FFmpeg {{ setup?.ffmpeg_ok ? 'Operational' : 'Error' }}
        </span>
        <span class="badge" :class="setup?.target_dir_exists ? 'badge-ok' : 'badge-err'">
          <span class="badge-dot" :class="setup?.target_dir_exists ? 'dot-ok' : 'dot-err'" />
          Storage {{ setup?.target_dir_exists ? 'OK' : 'Error' }}
        </span>
      </div>
    </div>

    <!-- Stat Cards -->
    <div class="stat-grid">
      <div v-for="card in statCards" :key="card.label" class="stat-card">
        <div class="stat-label">{{ card.label }}</div>
        <div class="stat-value">{{ card.value }}</div>
        <div class="stat-sub">{{ card.sub }}</div>
      </div>
    </div>

    <div class="grid-2">
      <!-- Disk Usage -->
      <div class="card">
        <div class="card-header">存储状态</div>
        <div class="card-body">
          <div class="disk-top">
            <span class="disk-label">磁盘使用率</span>
            <span class="disk-pct">{{ detail?.disk?.usage_pct?.toFixed(1) || 0 }}%</span>
          </div>
          <div class="progress-track">
            <div class="progress-bar" :style="{ width: (detail?.disk?.usage_pct || 0) + '%', background: diskColor }" />
          </div>
          <div class="meta-row">
            <div class="meta"><span class="meta-k">总容量</span><span class="meta-v">{{ detail?.disk?.total_gb?.toFixed(1) }} GB</span></div>
            <div class="meta"><span class="meta-k">已用</span><span class="meta-v">{{ detail?.disk?.used_gb?.toFixed(1) }} GB</span></div>
            <div class="meta"><span class="meta-k">剩余</span><span class="meta-v">{{ detail?.disk?.free_gb?.toFixed(1) }} GB</span></div>
            <div class="meta"><span class="meta-k">待合并</span><span class="meta-v">{{ detail?.pending?.original_files || 0 }} 文件</span></div>
          </div>
        </div>
      </div>

      <!-- System Status -->
      <div class="card">
        <div class="card-header">系统状态</div>
        <div class="card-body">
          <div v-for="item in statusItems" :key="item.label" class="status-row">
            <span class="status-dot" :class="item.ok ? 'dot-ok' : 'dot-err'" />
            <span class="status-label">{{ item.label }}</span>
            <span class="status-value">{{ item.value }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="grid-2">
      <!-- Trend -->
      <div class="card">
        <div class="card-header">近7天趋势</div>
        <div class="card-body">
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

      <!-- Recent -->
      <div class="card">
        <div class="card-header">
          <span>最近操作</span>
          <a class="view-link" @click="$router.push('/history')">查看全部</a>
        </div>
        <div class="card-body">
          <div v-for="row in recentHistory" :key="row.id" class="recent-row">
            <span class="recent-dot" :class="row.status === 'success' ? 'dot-ok' : 'dot-err'" />
            <div class="recent-info">
              <span class="recent-name">{{ row.streamer || '系统' }}</span>
              <span class="recent-detail">{{ row.detail || formatDetail(row) }}</span>
            </div>
            <span class="recent-tag">{{ row.task }}</span>
          </div>
          <div v-if="recentHistory.length === 0" class="empty">暂无操作记录</div>
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
  if (pct >= 80) return "var(--accent-red)";
  if (pct >= 60) return "var(--accent-orange)";
  return "var(--accent-green)";
});

const statCards = computed(() => {
  const s = stats.value;
  if (!s) return [];
  return [
    { label: "今日合并", value: s.today.merge_count, sub: s.today.merge_bytes > 0 ? formatBytes(s.today.merge_bytes) : "—" },
    { label: "今日清理", value: s.today.clean_count, sub: s.today.clean_bytes > 0 ? formatBytes(s.today.clean_bytes) : "—" },
    { label: "本月合并", value: s.month.merge_count, sub: s.month.merge_bytes > 0 ? formatBytes(s.month.merge_bytes) : "—" },
    { label: "本月清理", value: s.month.clean_count, sub: s.month.clean_bytes > 0 ? formatBytes(s.month.clean_bytes) : "—" }
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

function formatBytes(b: number): string {
  if (!b) return "0 B";
  if (b >= 1024 ** 3) return `${(b / 1024 ** 3).toFixed(2)} GB`;
  return `${(b / 1024 ** 2).toFixed(1)} MB`;
}
function barHeight(b: number): number { return Math.max(3, (b / maxBytes.value) * 100); }
function formatDetail(row: HistoryRecord) {
  const p = [];
  if (row.files_count) p.push(`${row.files_count}个文件`);
  if (row.duration) p.push(`${row.duration}s`);
  return p.join(", ") || "—";
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
.dashboard { display: flex; flex-direction: column; gap: 20px; }

/* Header */
.dash-header { display: flex; justify-content: space-between; align-items: flex-end; }
.dash-header h1 {
  font-family: var(--font-display); font-size: 28px; font-weight: 400;
  letter-spacing: -0.4px; color: var(--ink);
}
.dash-date { font-size: 13px; color: var(--mute); margin-top: 4px; }
.dash-badges { display: flex; gap: 8px; }

/* Badge */
.badge {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 4px 12px; border-radius: var(--r-full);
  font-size: 12px; color: var(--body-text);
  background: var(--surface-elevated);
  border: 1px solid var(--hairline-strong);
}
.badge-dot { width: 6px; height: 6px; border-radius: 50%; }
.dot-ok { background: var(--accent-green); }
.dot-err { background: var(--accent-red); }

/* Stats */
.stat-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; }

.stat-card {
  background: var(--surface-card);
  border: 1px solid var(--hairline-strong);
  border-radius: var(--r-lg);
  padding: 20px;
}
.stat-label { font-size: 13px; color: var(--mute); margin-bottom: 8px; }
.stat-value {
  font-family: var(--font-display); font-size: 32px; font-weight: 400;
  line-height: 1; color: var(--ink); margin-bottom: 4px;
}
.stat-sub { font-size: 12px; color: var(--stone); font-family: var(--font-mono); }

/* Cards */
.grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }

.card {
  background: var(--surface-card);
  border: 1px solid var(--hairline-strong);
  border-radius: var(--r-lg);
  overflow: hidden;
}

.card-header {
  padding: 14px 20px; font-size: 13px; font-weight: 500; color: var(--charcoal);
  border-bottom: 1px solid var(--hairline);
  display: flex; justify-content: space-between; align-items: center;
}
.card-body { padding: 20px; }

.view-link { font-size: 13px; color: var(--link); cursor: pointer; font-weight: 400; }

/* Disk */
.disk-top { display: flex; justify-content: space-between; margin-bottom: 8px; }
.disk-label { font-size: 13px; color: var(--mute); }
.disk-pct { font-size: 18px; font-weight: 500; letter-spacing: -0.3px; }

.progress-track { height: 4px; background: var(--surface-deep); border-radius: 2px; margin-bottom: 16px; overflow: hidden; }
.progress-bar { height: 100%; border-radius: 2px; transition: width 0.4s ease; }

.meta-row { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; }
.meta { display: flex; flex-direction: column; }
.meta-k { font-size: 12px; color: var(--stone); }
.meta-v { font-size: 14px; font-weight: 500; color: var(--body-text); }

/* Status */
.status-row { display: flex; align-items: center; gap: 10px; padding: 6px 0; }
.status-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.status-label { font-size: 13px; color: var(--mute); flex: 1; }
.status-value { font-size: 13px; font-weight: 500; color: var(--body-text); }

/* Trend */
.trend-chart { display: flex; justify-content: space-around; align-items: flex-end; height: 100px; }
.trend-col { display: flex; flex-direction: column; align-items: center; gap: 6px; }
.trend-bars { display: flex; gap: 4px; align-items: flex-end; }
.trend-bar { width: 14px; border-radius: 2px 2px 0 0; min-height: 3px; transition: height 0.3s; cursor: pointer; }
.trend-bar:hover { opacity: 0.7; }
.bar-merge { background: var(--accent-blue); }
.bar-clean { background: var(--accent-green); }
.trend-date { font-size: 11px; color: var(--stone); font-family: var(--font-mono); }
.trend-legend { display: flex; gap: 16px; justify-content: center; margin-top: 14px; font-size: 12px; color: var(--mute); }
.legend-dot { display: inline-block; width: 8px; height: 8px; border-radius: 2px; margin-right: 4px; vertical-align: middle; }

/* Recent */
.recent-row { display: flex; align-items: center; gap: 10px; padding: 8px 0; border-bottom: 1px solid var(--hairline); }
.recent-row:last-child { border-bottom: none; }
.recent-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.recent-info { flex: 1; min-width: 0; }
.recent-name { font-size: 13px; font-weight: 500; color: var(--ink); }
.recent-detail { display: block; font-size: 12px; color: var(--mute); margin-top: 1px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.recent-tag {
  font-size: 11px; font-family: var(--font-mono); color: var(--ash);
  background: var(--surface-elevated); padding: 1px 8px; border-radius: var(--r-full);
  border: 1px solid var(--hairline);
}
.empty { font-size: 13px; color: var(--stone); text-align: center; padding: 24px 0; }

@media (max-width: 1024px) { .stat-grid { grid-template-columns: repeat(2, 1fr); } .grid-2 { grid-template-columns: 1fr; } }
@media (max-width: 600px) { .stat-grid { grid-template-columns: 1fr; } .dash-header { flex-direction: column; align-items: flex-start; gap: 8px; } }
</style>

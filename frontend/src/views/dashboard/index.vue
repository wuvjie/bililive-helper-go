<template>
  <div class="dashboard">
    <!-- Header -->
    <div class="dash-header">
      <div>
        <h1>{{ greeting }}</h1>
        <p class="dash-date">{{ todayDate }}</p>
      </div>
      <div class="dash-badges">
        <span class="badge badge-green" v-if="setup?.ffmpeg_ok">FFmpeg 正常</span>
        <span class="badge badge-red" v-else>FFmpeg 异常</span>
        <span class="badge badge-green" v-if="setup?.target_dir_exists">存储正常</span>
        <span class="badge badge-red" v-else>存储异常</span>
      </div>
    </div>

    <!-- Stat Cards with pastel tints -->
    <div class="stat-grid">
      <div class="stat-card stat-peach">
        <span class="stat-label">今日合并</span>
        <span class="stat-value">{{ stats?.today.merge_count || 0 }}</span>
        <span class="stat-sub">{{ stats?.today.merge_bytes ? formatBytes(stats.today.merge_bytes) : '—' }}</span>
      </div>
      <div class="stat-card stat-mint">
        <span class="stat-label">今日清理</span>
        <span class="stat-value">{{ stats?.today.clean_count || 0 }}</span>
        <span class="stat-sub">{{ stats?.today.clean_bytes ? formatBytes(stats.today.clean_bytes) : '—' }}</span>
      </div>
      <div class="stat-card stat-sky">
        <span class="stat-label">本月合并</span>
        <span class="stat-value">{{ stats?.month.merge_count || 0 }}</span>
        <span class="stat-sub">{{ stats?.month.merge_bytes ? formatBytes(stats.month.merge_bytes) : '—' }}</span>
      </div>
      <div class="stat-card stat-lavender">
        <span class="stat-label">本月清理</span>
        <span class="stat-value">{{ stats?.month.clean_count || 0 }}</span>
        <span class="stat-sub">{{ stats?.month.clean_bytes ? formatBytes(stats.month.clean_bytes) : '—' }}</span>
      </div>
    </div>

    <div class="grid-2">
      <!-- Disk Usage -->
      <div class="card">
        <div class="card-title">💾 存储状态</div>
        <div class="card-body">
          <div class="disk-top">
            <span class="disk-label">磁盘使用率</span>
            <span class="disk-pct" :style="{ color: diskColor }">{{ detail?.disk?.usage_pct?.toFixed(1) || 0 }}%</span>
          </div>
          <el-progress
            :percentage="detail?.disk?.usage_pct || 0"
            :color="diskColor"
            :stroke-width="12"
            :format="() => ''"
            style="margin-bottom: 16px"
          />
          <el-descriptions :column="2" size="small" border>
            <el-descriptions-item label="总容量">{{ detail?.disk?.total_gb?.toFixed(1) }} GB</el-descriptions-item>
            <el-descriptions-item label="已用">{{ detail?.disk?.used_gb?.toFixed(1) }} GB</el-descriptions-item>
            <el-descriptions-item label="剩余">{{ detail?.disk?.free_gb?.toFixed(1) }} GB</el-descriptions-item>
            <el-descriptions-item label="待合并">{{ detail?.pending?.original_files || 0 }} 个文件</el-descriptions-item>
          </el-descriptions>
        </div>
      </div>

      <!-- System Status -->
      <div class="card">
        <div class="card-title">🔧 系统状态</div>
        <div class="card-body">
          <el-descriptions :column="1" size="small" border>
            <el-descriptions-item label="FFmpeg">
              <el-tag :type="setup?.ffmpeg_ok ? 'success' : 'danger'" size="small">
                {{ setup?.ffmpeg_ok ? '正常' : '异常' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="FFprobe">
              <el-tag :type="setup?.ffprobe_ok ? 'success' : 'danger'" size="small">
                {{ setup?.ffprobe_ok ? '正常' : '异常' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="目标目录">
              <el-tag :type="setup?.target_dir_exists ? 'success' : 'danger'" size="small">
                {{ setup?.target_dir_exists ? '存在' : '不存在' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="主播数">{{ setup?.streamer_count || 0 }}</el-descriptions-item>
            <el-descriptions-item label="视频数">{{ setup?.video_count || 0 }}</el-descriptions-item>
          </el-descriptions>
        </div>
      </div>
    </div>

    <div class="grid-2">
      <!-- Trend -->
      <div class="card">
        <div class="card-title">📊 近7天趋势</div>
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
            <span><span class="legend-dot" style="background: var(--primary)" />合并量</span>
            <span><span class="legend-dot" style="background: var(--success)" />释放量</span>
          </div>
        </div>
      </div>

      <!-- Recent -->
      <div class="card">
        <div class="card-title">
          <span>📋 最近操作</span>
          <el-button text type="primary" size="small" @click="$router.push('/history')">查看全部</el-button>
        </div>
        <div class="card-body">
          <el-table :data="recentHistory" size="small" :show-header="false">
            <el-table-column width="36">
              <template #default="{ row }">
                <span class="task-emoji">{{ row.task === 'merge' ? '🔄' : row.task === 'clean' ? '🧹' : '⚙️' }}</span>
              </template>
            </el-table-column>
            <el-table-column>
              <template #default="{ row }">
                <div class="recent-info">
                  <span class="recent-name">{{ row.streamer || '系统' }}</span>
                  <span class="recent-detail">{{ row.detail || formatDetail(row) }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column width="80" align="right">
              <template #default="{ row }">
                <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
                  {{ row.status === 'success' ? '成功' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
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
  if (pct >= 80) return "var(--error)";
  if (pct >= 60) return "var(--warning)";
  return "var(--success)";
});

const maxBytes = computed(() => {
  if (!stats.value?.daily) return 1;
  return Math.max(...stats.value.daily.map(d => Math.max(d.merge_bytes, d.clean_bytes)), 1);
});

function formatBytes(b: number): string {
  if (!b) return "0 B";
  if (b >= 1024 ** 3) return `${(b / 1024 ** 3).toFixed(2)} GB`;
  return `${(b / 1024 ** 2).toFixed(1)} MB`;
}
function barHeight(b: number): number { return Math.max(4, (b / maxBytes.value) * 100); }
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
.dash-header h1 { font-size: 28px; font-weight: 600; color: var(--ink); letter-spacing: -0.5px; }
.dash-date { font-size: 14px; color: var(--steel); margin-top: 2px; }
.dash-badges { display: flex; gap: 6px; }

/* Badge */
.badge {
  display: inline-flex; align-items: center; padding: 4px 12px;
  border-radius: var(--r-full); font-size: 13px; font-weight: 600;
}
.badge-green { background: var(--tint-mint); color: #0d7a2e; }
.badge-red { background: #fde8e8; color: #b91c1c; }

/* Stat cards with Notion pastel tints */
.stat-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; }

.stat-card {
  border-radius: var(--r-lg); padding: 20px;
  display: flex; flex-direction: column;
}
.stat-peach { background: var(--tint-peach); }
.stat-mint { background: var(--tint-mint); }
.stat-sky { background: var(--tint-sky); }
.stat-lavender { background: var(--tint-lavender); }

.stat-label { font-size: 13px; color: var(--slate); margin-bottom: 4px; }
.stat-value { font-size: 32px; font-weight: 600; color: var(--ink); line-height: 1.1; letter-spacing: -0.5px; }
.stat-sub { font-size: 13px; color: var(--steel); margin-top: 2px; }

/* Cards */
.grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

.card {
  background: var(--canvas); border: 1px solid var(--hairline);
  border-radius: var(--r-lg); overflow: hidden;
  box-shadow: 0 1px 2px rgba(15, 15, 15, 0.04);
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

/* Trend */
.trend-chart { display: flex; justify-content: space-around; align-items: flex-end; height: 100px; }
.trend-col { display: flex; flex-direction: column; align-items: center; gap: 6px; }
.trend-bars { display: flex; gap: 4px; align-items: flex-end; }
.trend-bar { width: 16px; border-radius: 4px 4px 0 0; min-height: 4px; transition: height 0.3s; cursor: pointer; }
.trend-bar:hover { opacity: 0.75; }
.bar-merge { background: var(--primary); }
.bar-clean { background: var(--success); }
.trend-date { font-size: 11px; color: var(--stone); }
.trend-legend { display: flex; gap: 16px; justify-content: center; margin-top: 14px; font-size: 13px; color: var(--steel); }
.legend-dot { display: inline-block; width: 10px; height: 10px; border-radius: 3px; margin-right: 4px; vertical-align: middle; }

/* Recent */
.task-emoji { font-size: 14px; }
.recent-info { display: flex; flex-direction: column; }
.recent-name { font-size: 13px; font-weight: 500; color: var(--ink); }
.recent-detail { font-size: 12px; color: var(--steel); margin-top: 1px; }

@media (max-width: 1024px) { .stat-grid { grid-template-columns: repeat(2, 1fr); } .grid-2 { grid-template-columns: 1fr; } }
@media (max-width: 600px) { .stat-grid { grid-template-columns: 1fr; } .dash-header { flex-direction: column; align-items: flex-start; gap: 8px; } }
</style>

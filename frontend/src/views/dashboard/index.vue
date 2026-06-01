<template>
  <div class="dashboard">
    <!-- Greeting Banner -->
    <div class="greeting-banner">
      <div class="greeting-left">
        <h2>{{ greeting }}，Admin 👋</h2>
        <p>{{ todayDate }} · 直播录制管理看板</p>
      </div>
      <div class="greeting-right">
        <el-tag :type="setup?.ffmpeg_ok ? 'success' : 'danger'" effect="dark" round>
          FFmpeg {{ setup?.ffmpeg_ok ? '正常' : '异常' }}
        </el-tag>
        <el-tag :type="setup?.target_dir_exists ? 'success' : 'danger'" effect="dark" round>
          存储 {{ setup?.target_dir_exists ? '正常' : '异常' }}
        </el-tag>
      </div>
    </div>

    <!-- Stat Cards -->
    <el-row :gutter="16">
      <el-col :xs="12" :sm="6" v-for="(card, i) in statCards" :key="card.label">
        <div class="stat-card" :class="`stat-card--${card.theme}`">
          <div class="stat-card__icon">{{ card.icon }}</div>
          <div class="stat-card__info">
            <div class="stat-card__label">{{ card.label }}</div>
            <div class="stat-card__value">{{ card.value }}</div>
            <div class="stat-card__sub">{{ card.sub }}</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <!-- Disk Usage -->
      <el-col :xs="24" :sm="12">
        <el-card shadow="never" class="section-card">
          <template #header>
            <div class="card-header"><span>💾 存储状态</span></div>
          </template>
          <div class="disk-info">
            <div class="disk-header">
              <span>磁盘使用率</span>
              <span class="disk-pct" :style="{ color: diskColor }">{{ detail?.disk?.usage_pct?.toFixed(1) || 0 }}%</span>
            </div>
            <el-progress
              :percentage="detail?.disk?.usage_pct || 0"
              :color="diskColor"
              :stroke-width="16"
              :format="(p: number) => ''"
            />
            <el-descriptions :column="2" size="small" border class="disk-detail">
              <el-descriptions-item label="总容量">{{ detail?.disk?.total_gb?.toFixed(1) }} GB</el-descriptions-item>
              <el-descriptions-item label="已用">{{ detail?.disk?.used_gb?.toFixed(1) }} GB</el-descriptions-item>
              <el-descriptions-item label="剩余">{{ detail?.disk?.free_gb?.toFixed(1) }} GB</el-descriptions-item>
              <el-descriptions-item label="待合并">{{ detail?.pending?.original_files || 0 }} 个文件 ({{ (detail?.pending?.original_size_gb || 0).toFixed(2) }} GB)</el-descriptions-item>
            </el-descriptions>
          </div>
        </el-card>
      </el-col>

      <!-- System Status -->
      <el-col :xs="24" :sm="12">
        <el-card shadow="never" class="section-card">
          <template #header>
            <div class="card-header"><span>🔧 系统状态</span></div>
          </template>
          <div class="status-grid">
            <div class="status-item" v-for="item in statusItems" :key="item.label">
              <div class="status-dot" :class="item.ok ? 'status-dot--ok' : 'status-dot--err'" />
              <div class="status-info">
                <span class="status-label">{{ item.label }}</span>
                <span class="status-value">{{ item.value }}</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <!-- 7-Day Trend -->
      <el-col :xs="24" :sm="12">
        <el-card shadow="never" class="section-card">
          <template #header>
            <div class="card-header"><span>📊 近7天趋势</span></div>
          </template>
          <div class="trend-chart">
            <div v-for="day in stats?.daily" :key="day.date" class="trend-col">
              <div class="trend-bars">
                <el-tooltip :content="`合并 ${formatBytes(day.merge_bytes)}`" placement="top">
                  <div class="trend-bar merge" :style="{ height: barHeight(day.merge_bytes) + 'px' }" />
                </el-tooltip>
                <el-tooltip :content="`释放 ${formatBytes(day.clean_bytes)}`" placement="top">
                  <div class="trend-bar clean" :style="{ height: barHeight(day.clean_bytes) + 'px' }" />
                </el-tooltip>
              </div>
              <div class="trend-date">{{ day.date.slice(5) }}</div>
            </div>
          </div>
          <div class="trend-legend">
            <span class="legend-item"><span class="legend-dot merge" />合并量</span>
            <span class="legend-item"><span class="legend-dot clean" />释放量</span>
          </div>
        </el-card>
      </el-col>

      <!-- Recent Operations -->
      <el-col :xs="24" :sm="12">
        <el-card shadow="never" class="section-card">
          <template #header>
            <div class="card-header">
              <span>📋 最近操作</span>
              <el-button text type="primary" size="small" @click="$router.push('/history')">查看全部</el-button>
            </div>
          </template>
          <el-table :data="recentHistory" size="small" :show-header="false" class="recent-table">
            <el-table-column width="40">
              <template #default="{ row }">
                <span class="task-icon">{{ row.task === 'merge' ? '🔄' : row.task === 'clean' ? '🧹' : '⚙️' }}</span>
              </template>
            </el-table-column>
            <el-table-column>
              <template #default="{ row }">
                <div class="recent-item">
                  <span class="recent-streamer">{{ row.streamer || '系统' }}</span>
                  <span class="recent-detail">{{ row.detail || formatDetail(row) }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column width="100" align="right">
              <template #default="{ row }">
                <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small" round>
                  {{ row.status === 'success' ? '成功' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
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
  if (h < 9) return "早上好";
  if (h < 12) return "上午好";
  if (h < 14) return "中午好";
  if (h < 18) return "下午好";
  return "晚上好";
});

const todayDate = computed(() => {
  return new Date().toLocaleDateString("zh-CN", { year: "numeric", month: "long", day: "numeric", weekday: "long" });
});

const diskColor = computed(() => {
  const pct = detail.value?.disk?.usage_pct || 0;
  if (pct >= 80) return "#f56c6c";
  if (pct >= 60) return "#e6a23c";
  return "#67c23a";
});

const statCards = computed(() => {
  const s = stats.value;
  if (!s) return [];
  return [
    { label: "今日合并次数", value: s.today.merge_count, sub: s.today.merge_bytes > 0 ? `合并 ${formatBytes(s.today.merge_bytes)}` : "暂无合并", icon: "🔄", theme: "blue" },
    { label: "今日清理次数", value: s.today.clean_count, sub: s.today.clean_bytes > 0 ? `释放 ${formatBytes(s.today.clean_bytes)}` : "暂无清理", icon: "🧹", theme: "green" },
    { label: "本月合并次数", value: s.month.merge_count, sub: s.month.merge_bytes > 0 ? `合并 ${formatBytes(s.month.merge_bytes)}` : "暂无合并", icon: "📦", theme: "purple" },
    { label: "本月清理次数", value: s.month.clean_count, sub: s.month.clean_bytes > 0 ? `释放 ${formatBytes(s.month.clean_bytes)}` : "暂无清理", icon: "💨", theme: "orange" }
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
  return Math.max(4, (bytes / maxBytes.value) * 100);
}

function formatDetail(row: HistoryRecord) {
  const parts = [];
  if (row.files_count) parts.push(`${row.files_count}个文件`);
  if (row.duration) parts.push(`${row.duration}s`);
  return parts.join(", ") || "-";
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
.dashboard { display: flex; flex-direction: column; gap: 16px; }

/* Greeting */
.greeting-banner {
  background: linear-gradient(135deg, #3370ff 0%, #5e8fff 100%);
  border-radius: var(--radius-lg);
  padding: 24px 28px;
  color: #fff;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}
.greeting-left h2 { font-size: 22px; font-weight: 700; margin-bottom: 4px; }
.greeting-left p { font-size: 14px; opacity: 0.85; }
.greeting-right { display: flex; gap: 8px; }

/* Stat Cards */
.stat-card {
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  padding: 20px;
  display: flex;
  gap: 16px;
  align-items: center;
  border: 1px solid var(--border-color);
  transition: transform 0.2s, box-shadow 0.2s, background 0.3s;
  cursor: default;
}
.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}
.stat-card__icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  flex-shrink: 0;
}
.stat-card--blue .stat-card__icon { background: #ecf2ff; }
.stat-card--green .stat-card__icon { background: #e8f5e9; }
.stat-card--purple .stat-card__icon { background: #f3e5f5; }
.stat-card--orange .stat-card__icon { background: #fff3e0; }

.stat-card__label { font-size: 13px; color: var(--text-regular); margin-bottom: 4px; }
.stat-card__value { font-size: 28px; font-weight: 700; color: var(--text-primary); line-height: 1.2; }
.stat-card__sub { font-size: 12px; color: var(--text-placeholder); margin-top: 2px; }

/* Section Cards */
.section-card { background: var(--bg-card); transition: background 0.3s; }
.section-card :deep(.el-card__header) { padding: 14px 20px; border-bottom: 1px solid var(--border-color); }
.card-header { display: flex; justify-content: space-between; align-items: center; }

/* Status Grid */
.status-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
.status-item { display: flex; align-items: center; gap: 10px; }
.status-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.status-dot--ok { background: #67c23a; box-shadow: 0 0 6px rgba(103, 194, 58, 0.4); }
.status-dot--err { background: #f56c6c; box-shadow: 0 0 6px rgba(245, 108, 108, 0.4); }
.status-info { display: flex; flex-direction: column; }
.status-label { font-size: 12px; color: var(--text-placeholder); }
.status-value { font-size: 14px; font-weight: 500; color: var(--text-primary); }

/* Disk */
.disk-info { display: flex; flex-direction: column; gap: 12px; }
.disk-header { display: flex; justify-content: space-between; font-size: 14px; color: var(--text-regular); }
.disk-pct { font-weight: 700; font-size: 16px; }
.disk-detail { margin-top: 8px; }

/* Trend */
.trend-chart {
  display: flex;
  justify-content: space-around;
  align-items: flex-end;
  height: 120px;
  padding: 0 8px;
}
.trend-col { display: flex; flex-direction: column; align-items: center; gap: 6px; }
.trend-bars { display: flex; gap: 4px; align-items: flex-end; }
.trend-bar {
  width: 16px;
  border-radius: 4px 4px 0 0;
  min-height: 4px;
  transition: height 0.4s ease;
  cursor: pointer;
}
.trend-bar.merge { background: linear-gradient(180deg, #409eff, #66b1ff); }
.trend-bar.clean { background: linear-gradient(180deg, #67c23a, #85ce61); }
.trend-bar:hover { opacity: 0.8; }
.trend-date { font-size: 11px; color: var(--text-placeholder); }
.trend-legend { display: flex; justify-content: center; gap: 20px; margin-top: 12px; }
.legend-item { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--text-regular); }
.legend-dot { width: 10px; height: 10px; border-radius: 3px; display: inline-block; }
.legend-dot.merge { background: #409eff; }
.legend-dot.clean { background: #67c23a; }

/* Recent */
.recent-table { --el-table-border: none; }
.task-icon { font-size: 16px; }
.recent-item { display: flex; flex-direction: column; }
.recent-streamer { font-size: 13px; font-weight: 500; color: var(--text-primary); }
.recent-detail { font-size: 12px; color: var(--text-placeholder); margin-top: 2px; }

@media (max-width: 768px) {
  .stat-card { padding: 14px; gap: 10px; }
  .stat-card__icon { width: 40px; height: 40px; font-size: 20px; }
  .stat-card__value { font-size: 22px; }
  .status-grid { grid-template-columns: repeat(2, 1fr); }
  .greeting-banner { padding: 16px 20px; }
  .greeting-left h2 { font-size: 18px; }
}
</style>

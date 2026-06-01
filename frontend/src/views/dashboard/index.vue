<template>
  <div class="dashboard">
    <div class="dash-header">
      <div>
        <h1>{{ greeting }}</h1>
        <p class="dash-date">{{ todayDate }}</p>
      </div>
      <div class="dash-badges">
        <span class="badge" :class="setup?.ffmpeg_ok ? 'badge-ok' : 'badge-err'">FFmpeg {{ setup?.ffmpeg_ok ? '正常' : '异常' }}</span>
        <span class="badge" :class="setup?.target_dir_exists ? 'badge-ok' : 'badge-err'">存储 {{ setup?.target_dir_exists ? '正常' : '异常' }}</span>
      </div>
    </div>

    <div class="stat-grid">
      <div v-for="card in statCards" :key="card.label" class="stat-card">
        <span class="stat-label">{{ card.label }}</span>
        <span class="stat-value">{{ card.value }}</span>
        <span class="stat-sub">{{ card.sub }}</span>
      </div>
    </div>

    <div class="grid-2">
      <div class="card">
        <div class="card-title">存储状态</div>
        <div class="card-body">
          <div class="disk-top">
            <span>磁盘使用率</span>
            <span class="disk-pct" :style="{ color: diskColor }">{{ detail?.disk?.usage_pct?.toFixed(1) || 0 }}%</span>
          </div>
          <el-progress :percentage="detail?.disk?.usage_pct || 0" :color="diskColor" :stroke-width="12" :format="() => ''" style="margin-bottom: 16px" />
          <el-descriptions :column="2" size="small" border>
            <el-descriptions-item label="总容量">{{ detail?.disk?.total_gb?.toFixed(1) }} GB</el-descriptions-item>
            <el-descriptions-item label="已用">{{ detail?.disk?.used_gb?.toFixed(1) }} GB</el-descriptions-item>
            <el-descriptions-item label="剩余">{{ detail?.disk?.free_gb?.toFixed(1) }} GB</el-descriptions-item>
            <el-descriptions-item label="待合并">{{ detail?.pending?.original_files || 0 }} 个文件</el-descriptions-item>
          </el-descriptions>
        </div>
      </div>

      <div class="card">
        <div class="card-title">系统状态</div>
        <div class="card-body">
          <el-descriptions :column="1" size="small" border>
            <el-descriptions-item label="FFmpeg"><el-tag :type="setup?.ffmpeg_ok ? 'success' : 'danger'" size="small">{{ setup?.ffmpeg_ok ? '正常' : '异常' }}</el-tag></el-descriptions-item>
            <el-descriptions-item label="FFprobe"><el-tag :type="setup?.ffprobe_ok ? 'success' : 'danger'" size="small">{{ setup?.ffprobe_ok ? '正常' : '异常' }}</el-tag></el-descriptions-item>
            <el-descriptions-item label="目标目录"><el-tag :type="setup?.target_dir_exists ? 'success' : 'danger'" size="small">{{ setup?.target_dir_exists ? '存在' : '不存在' }}</el-tag></el-descriptions-item>
            <el-descriptions-item label="主播数">{{ setup?.streamer_count || 0 }}</el-descriptions-item>
            <el-descriptions-item label="视频数">{{ setup?.video_count || 0 }}</el-descriptions-item>
          </el-descriptions>
        </div>
      </div>
    </div>

    <div class="grid-2">
      <div class="card">
        <div class="card-title">近7天趋势</div>
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
            <span><span class="legend-dot" style="background: var(--ink)" />合并量</span>
            <span><span class="legend-dot" style="background: var(--success)" />释放量</span>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-title">
          <span>最近操作</span>
          <el-button text type="primary" size="small" @click="$router.push('/history')">查看全部</el-button>
        </div>
        <div class="card-body">
          <el-table :data="recentHistory" size="small" :show-header="false">
            <el-table-column width="36"><template #default="{ row }"><span>{{ row.task === 'merge' ? '🔄' : row.task === 'clean' ? '🧹' : '⚙️' }}</span></template></el-table-column>
            <el-table-column>
              <template #default="{ row }">
                <div class="recent-info"><span class="recent-name">{{ row.streamer || '系统' }}</span><span class="recent-detail">{{ row.detail || formatDetail(row) }}</span></div>
              </template>
            </el-table-column>
            <el-table-column width="80" align="right">
              <template #default="{ row }"><el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">{{ row.status === 'success' ? '成功' : '失败' }}</el-tag></template>
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
  if (h < 6) return "夜深了"; if (h < 12) return "早上好"; if (h < 14) return "中午好"; if (h < 18) return "下午好"; return "晚上好";
});
const todayDate = computed(() => new Date().toLocaleDateString("zh-CN", { year: "numeric", month: "long", day: "numeric", weekday: "long" }));
const diskColor = computed(() => { const p = detail.value?.disk?.usage_pct || 0; return p >= 80 ? "var(--error)" : p >= 60 ? "var(--warning)" : "var(--success)"; });
const statCards = computed(() => {
  const s = stats.value; if (!s) return [];
  return [
    { label: "今日合并", value: s.today.merge_count, sub: s.today.merge_bytes > 0 ? formatBytes(s.today.merge_bytes) : "—" },
    { label: "今日清理", value: s.today.clean_count, sub: s.today.clean_bytes > 0 ? formatBytes(s.today.clean_bytes) : "—" },
    { label: "本月合并", value: s.month.merge_count, sub: s.month.merge_bytes > 0 ? formatBytes(s.month.merge_bytes) : "—" },
    { label: "本月清理", value: s.month.clean_count, sub: s.month.clean_bytes > 0 ? formatBytes(s.month.clean_bytes) : "—" }
  ];
});
const maxBytes = computed(() => { if (!stats.value?.daily) return 1; return Math.max(...stats.value.daily.map(d => Math.max(d.merge_bytes, d.clean_bytes)), 1); });
function formatBytes(b: number) { if (!b) return "0 B"; return b >= 1024 ** 3 ? `${(b / 1024 ** 3).toFixed(2)} GB` : `${(b / 1024 ** 2).toFixed(1)} MB`; }
function barHeight(b: number) { return Math.max(4, (b / maxBytes.value) * 100); }
function formatDetail(row: HistoryRecord) { const p = []; if (row.files_count) p.push(`${row.files_count}个文件`); if (row.duration) p.push(`${row.duration}s`); return p.join(", ") || "—"; }

onMounted(async () => {
  const [d, s, h, ck] = await Promise.allSettled([getStatusDetail(), getStats(), getHistory({ per_page: 5 }), setupCheck()]);
  if (d.status === "fulfilled") detail.value = d.value;
  if (s.status === "fulfilled") stats.value = s.value;
  if (h.status === "fulfilled") recentHistory.value = h.value.items || [];
  if (ck.status === "fulfilled") setup.value = ck.value;
});
</script>

<style scoped>
.dashboard { display: flex; flex-direction: column; gap: 24px; }

.dash-header { display: flex; justify-content: space-between; align-items: flex-end; }
.dash-header h1 { font-size: 40px; font-weight: 900; color: var(--ink); line-height: 34px; letter-spacing: -0.5px; }
.dash-date { font-size: 16px; color: var(--body-text); margin-top: 8px; }
.dash-badges { display: flex; gap: 8px; }

.badge { display: inline-flex; padding: 4px 14px; border-radius: var(--r-pill); font-size: 14px; font-weight: 600; }
.badge-ok { background: var(--primary-pale); color: #054d28; }
.badge-err { background: #fde8e8; color: #a72027; }

.stat-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; }
.stat-card { background: var(--canvas); border-radius: var(--r-xl); padding: 24px; display: flex; flex-direction: column; }
.stat-label { font-size: 16px; color: var(--mute); margin-bottom: 8px; }
.stat-value { font-size: 40px; font-weight: 900; color: var(--ink); line-height: 1; }
.stat-sub { font-size: 14px; color: var(--body-text); margin-top: 6px; }

.grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.card { background: var(--canvas); border-radius: var(--r-xl); overflow: hidden; }
.card-title { padding: 16px 24px; font-size: 20px; font-weight: 900; color: var(--ink); border-bottom: 1px solid var(--canvas-soft); display: flex; justify-content: space-between; align-items: center; }
.card-body { padding: 24px; }

.disk-top { display: flex; justify-content: space-between; margin-bottom: 8px; font-size: 16px; color: var(--body-text); }
.disk-pct { font-size: 24px; font-weight: 900; }

.trend-chart { display: flex; justify-content: space-around; align-items: flex-end; height: 110px; }
.trend-col { display: flex; flex-direction: column; align-items: center; gap: 6px; }
.trend-bars { display: flex; gap: 5px; align-items: flex-end; }
.trend-bar { width: 20px; border-radius: var(--r-sm) var(--r-sm) 0 0; min-height: 4px; transition: height 0.3s; cursor: pointer; }
.trend-bar:hover { opacity: 0.75; }
.bar-merge { background: var(--ink); }
.bar-clean { background: var(--success); }
.trend-date { font-size: 13px; color: var(--mute); }
.trend-legend { display: flex; gap: 20px; justify-content: center; margin-top: 16px; font-size: 14px; color: var(--body-text); }
.legend-dot { display: inline-block; width: 12px; height: 12px; border-radius: 4px; margin-right: 6px; vertical-align: middle; }

.recent-info { display: flex; flex-direction: column; }
.recent-name { font-size: 16px; font-weight: 600; color: var(--ink); }
.recent-detail { font-size: 14px; color: var(--mute); margin-top: 2px; }

@media (max-width: 1024px) { .stat-grid { grid-template-columns: repeat(2, 1fr); } .grid-2 { grid-template-columns: 1fr; } }
@media (max-width: 600px) { .stat-grid { grid-template-columns: 1fr; } .dash-header { flex-direction: column; align-items: flex-start; gap: 10px; } }
</style>

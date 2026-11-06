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

    <!-- Stat Cards with pastel tints (今日合并 · 昨日合并 · 本月合并 · 累计清理) -->
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

    <div class="grid-2">
      <!-- 1. Disk Storage Status -->
      <div class="card">
        <div class="card-title">
          <div class="card-title-left">
            <div class="ui-icon-box icon-box-blue">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12H2M5.45 5.11L2 12v6a2 2 0 002 2h16a2 2 0 002-2v-6l-3.45-6.89A2 2 0 0016.76 4H7.24a2 2 0 00-1.79 1.11z"/><line x1="6" y1="16" x2="6.01" y2="16"/><line x1="10" y1="16" x2="10.01" y2="16"/></svg>
            </div>
            <span class="card-title-main">存储空间监控</span>
          </div>
        </div>
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
            <div class="data-cell">
              <span class="data-label">总容量</span>
              <span class="mono-val">{{ detail?.disk?.total_gb?.toFixed(1) }} GB</span>
            </div>
            <div class="data-cell">
              <span class="data-label">已用</span>
              <span class="mono-val">{{ detail?.disk?.used_gb?.toFixed(1) }} GB</span>
            </div>
            <div class="data-cell">
              <span class="data-label">剩余</span>
              <span class="mono-val">{{ detail?.disk?.free_gb?.toFixed(1) }} GB</span>
            </div>
            <div class="data-cell">
              <span class="data-label">
                录播库占用
                <el-tooltip content="当前已录制的视频文件所占用的磁盘物理空间总和" placement="top">
                  <span class="label-help-ico">?</span>
                </el-tooltip>
              </span>
              <span class="mono-val">{{ totalRecordingsGB }} GB</span>
            </div>
          </div>
          <div v-if="stats?.daily_output_gb || stats?.days_until_full" class="bottom-health-bar">
            <span class="health-pill">
              近7天产出: <strong class="mono-val">{{ stats?.daily_output_gb?.toFixed(1) || 0 }}</strong> GB/天
            </span>
            <span class="health-pill" :class="{ 'health-pill-warn': (stats?.days_until_full || 999) < 7 }">
              余量预估: <strong class="mono-val">{{ (stats?.days_until_full && stats?.days_until_full > 0) ? stats.days_until_full.toFixed(0) + ' 天' : '充足' }}</strong>
            </span>
          </div>
        </div>
      </div>

      <!-- 2. Runtime & Automation Monitor -->
      <div class="card">
        <div class="card-title">
          <div class="card-title-left">
            <div class="ui-icon-box icon-box-purple">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
            </div>
            <span class="card-title-main">自动化调度监控</span>
            <span class="monitor-mode-tag" :class="isTaskRunning ? 'tag-running' : 'tag-in-service'">
              <span class="status-pulse-dot" :class="isTaskRunning ? 'dot-active' : 'dot-in-service'"></span>
              {{ isTaskRunning ? '任务执行中' : '自动调度生效中' }}
            </span>
          </div>
          <router-link to="/tasks" class="card-header-action" title="前往任务中心查看流式终端与执行记录">
            任务中心 ↗
          </router-link>
        </div>
        <div class="card-body">
          <div class="runtime-top">
            <span class="runtime-label">当前任务状态</span>
            <div class="runtime-badge" :class="isTaskRunning ? 'badge-running' : 'badge-idle'">
              <span>{{ isTaskRunning ? (detail?.schedule?.merge?.is_running ? '正在合并视频...' : '正在清理磁盘...') : '空闲待命中' }}</span>
            </div>
          </div>
          <div class="runtime-substrip">
            <span class="substrip-text">
              {{ isTaskRunning ? '后台正在处理媒体流水线任务，任务中心可查看终端流式输出' : '定时调度器与目录监听器持续常驻，主播关播分片将自动静置后合并' }}
            </span>
          </div>

          <div class="data-grid">
            <div class="data-cell">
              <span class="data-label">自动合并</span>
              <span class="mono-val">{{ detail?.schedule?.merge_enabled ? ('每 ' + (detail?.schedule?.merge_interval || 360) + ' 分钟') : '已暂停' }}</span>
            </div>
            <div class="data-cell">
              <span class="data-label">自动清理</span>
              <span class="mono-val">{{ detail?.schedule?.clean_enabled ? ('每 ' + (detail?.schedule?.clean_interval || 720) + ' 分钟') : '未开启' }}</span>
            </div>
            <div class="data-cell">
              <span class="data-label">录制主播</span>
              <span class="mono-val">{{ streamerCount }} 位</span>
            </div>
            <div class="data-cell">
              <span class="data-label">
                视频总数
                <el-tooltip content="全部主播录像文件总数（含未合并分段切片与已合并成品）" placement="top">
                  <span class="label-help-ico">?</span>
                </el-tooltip>
              </span>
              <span class="mono-val">{{ videoCount }} 个</span>
            </div>
          </div>

          <div class="bottom-health-bar">
            <span class="health-pill">
              下次合并: <strong class="mono-val">{{ formatNextTime(detail?.schedule?.merge?.next_run) }}</strong>
            </span>
            <span class="health-pill">
              下次清理: <strong class="mono-val">{{ formatNextTime(detail?.schedule?.clean?.next_run) }}</strong>
            </span>
          </div>
        </div>
      </div>
    </div>

    <div class="grid-2">
      <!-- 3. Trend -->
      <div class="card">
        <div class="card-title">
          <div class="card-title-left">
            <div class="ui-icon-box icon-box-teal">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
            </div>
            <span class="card-title-main">近 7 天产出与清理趋势</span>
          </div>
          <span class="trend-sub-summary" v-if="hasTrendData">
            累计合并 {{ totalMergeGB }} · 释放 {{ totalCleanGB }}
          </span>
        </div>
        <div class="card-body trend-body">
          <div v-if="hasTrendData" class="trend-chart">
            <div v-for="day in stats?.daily" :key="day.date" class="trend-col">
              <div class="trend-bars">
                <el-tooltip :content="getMergeTooltip(day)" placement="top">
                  <div class="trend-bar bar-merge" :style="{ height: barHeight(day.merge_bytes) + 'px' }" />
                </el-tooltip>
                <el-tooltip :content="getCleanTooltip(day)" placement="top">
                  <div class="trend-bar bar-clean" :style="{ height: barHeight(day.clean_bytes) + 'px' }" />
                </el-tooltip>
              </div>
              <span class="trend-date">{{ day.date.slice(5) }}</span>
            </div>
          </div>
          <div v-else class="trend-empty empty-illustration">
            <div class="empty-icon-wrapper" style="background: rgba(20, 184, 166, 0.09); color: #0d9488;">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
            </div>
            <span class="empty-title">近 7 天暂无自动化处理数据</span>
            <span class="empty-desc">当录像分片合并或触发磁盘清理时，每日吞吐量将自动在此展现</span>
          </div>

          <div class="trend-legend">
            <span><span class="legend-dot" style="background: #787774" />合并量</span>
            <span><span class="legend-dot" style="background: var(--success)" />释放量</span>
          </div>
        </div>
      </div>

      <!-- 4. Recent Operations -->
      <div class="card">
        <div class="card-title">
          <div class="card-title-left">
            <div class="ui-icon-box icon-box-indigo">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
            </div>
            <span class="card-title-main">最近操作记录</span>
          </div>
          <router-link to="/history" class="card-header-action" title="前往操作日志查看全部记录">
            查看全部 ↗
          </router-link>
        </div>
        <div class="card-body recent-body">
          <div v-if="recentHistory.length === 0" class="recent-empty">
            <div class="empty-icon-wrapper"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg></div>
            <span class="recent-empty-title">暂无近期操作记录</span>
            <span class="recent-empty-desc">后台自动执行或手动触发的任务记录将保存在此处</span>
          </div>
          <div v-else class="recent-list">
            <div
              v-for="row in recentHistory"
              :key="row.id"
              class="recent-row"
            >
              <div class="recent-top-line">
                <div class="recent-left">
                  <span class="recent-tag" :class="'tag-' + row.task">
                    {{ row.task === 'merge' ? '合并' : row.task === 'clean' ? '清理' : row.task === 'config' ? '配置' : '调度' }}
                  </span>
                  <span class="recent-who">{{ row.streamer || '全局' }}</span>
                  <span class="recent-action desktop-action" :title="cleanDetailText(row.detail || formatDetail(row), row.task)">
                    {{ cleanDetailText(row.detail || formatDetail(row), row.task) }}
                  </span>
                </div>
                <div class="recent-right">
                  <span class="recent-status-dot" :class="row.status === 'fail' ? 'dot-fail' : 'dot-emerald'"></span>
                  <span class="recent-time">{{ formatShortTime(row.time) }}</span>
                </div>
              </div>
              <div class="recent-action mobile-action" :title="cleanDetailText(row.detail || formatDetail(row), row.task)">
                {{ cleanDetailText(row.detail || formatDetail(row), row.task) }}
              </div>
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
import type { DetailResponse, StatsResponse, SetupCheck, HistoryRecord, DayStats } from "@/api/types";

const detail = ref<DetailResponse>();
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

const stats = ref<StatsResponse>();
const setup = ref<SetupCheck>();
const recentHistory = ref<HistoryRecord[]>([]);
const loading = ref(true);

const greeting = computed(() => {
  const now = new Date();
  const h = now.getHours();
  const d = now.getDate(); // 基于日期做轮换，当天内刷新稳定不跳跃，每日又富有新鲜感

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

const diskColor = computed(() => {
  const pct = detail.value?.disk?.usage_pct || 0;
  if (pct >= 90) return "#d9730d"; // Notion warm amber for critical
  return "#448361"; // Notion muted botanical green
});

const totalRecordingsGB = computed(() => {
  const sumBytes = detail.value?.streamers?.reduce((acc, s) => {
    if (s.size_bytes !== undefined && s.size_bytes !== null) {
      return acc + s.size_bytes;
    }
    return acc + (s.size_gb ? s.size_gb * (1024 ** 3) : 0);
  }, 0) || 0;
  return (sumBytes / (1024 ** 3)).toFixed(1);
});

const isTaskRunning = computed(() => {
  return Boolean(
    detail.value?.schedule?.running ||
    detail.value?.schedule?.merge?.is_running ||
    detail.value?.schedule?.clean?.is_running
  );
});

const streamerCount = computed(() => {
  return setup.value?.streamer_count ?? detail.value?.streamers?.length ?? 0;
});

const videoCount = computed(() => {
  return setup.value?.video_count ?? detail.value?.streamers?.reduce((acc, s) => acc + (s.files || 0), 0) ?? 0;
});

const hasTrendData = computed(() => {
  if (!stats.value?.daily || stats.value.daily.length === 0) return false;
  return stats.value.daily.some(d => (d.merge_bytes > 0 || d.clean_bytes > 0));
});

const totalMergeGB = computed(() => {
  const bytes = stats.value?.daily?.reduce((acc, d) => acc + (d.merge_bytes || 0), 0) || 0;
  return formatBytes(bytes);
});

const totalCleanGB = computed(() => {
  const bytes = stats.value?.daily?.reduce((acc, d) => acc + (d.clean_bytes || 0), 0) || 0;
  return formatBytes(bytes);
});

const maxBytes = computed(() => {
  if (!stats.value?.daily) return 1;
  return Math.max(...stats.value.daily.map(d => Math.max(d.merge_bytes || 0, d.clean_bytes || 0)), 1);
});

function barHeight(b: number): number {
  if (!b || b <= 0) return 4;
  return Math.max(6, Math.min(100, Math.round((b / maxBytes.value) * 100)));
}

function getMergeTooltip(day: DayStats): string {
  if (!day.merge_bytes || day.merge_bytes <= 0) return `${day.date} 无合并任务`;
  const count = day.merge_count ? `${day.merge_count} 次` : "";
  return `合并 ${count} (${formatBytes(day.merge_bytes)})`;
}

function getCleanTooltip(day: DayStats): string {
  if (!day.clean_bytes || day.clean_bytes <= 0) return `${day.date} 无清理操作`;
  const count = day.clean_count ? `${day.clean_count} 次` : "";
  return `释放 ${count} (${formatBytes(day.clean_bytes)})`;
}

function formatShortTime(t?: string): string {
  if (!t || t.length < 16) return "";
  return t.slice(11, 16); // "HH:mm"
}

function formatNextTime(ts?: number): string {
  if (!ts || ts <= 0) return "未安排";
  const date = new Date(ts * 1000);
  const now = new Date();
  const diffMinutes = Math.round((date.getTime() - now.getTime()) / 60000);
  if (diffMinutes <= 0) return "即将执行";
  if (diffMinutes < 60) return `${diffMinutes} 分钟后`;
  const hours = String(date.getHours()).padStart(2, "0");
  const mins = String(date.getMinutes()).padStart(2, "0");
  return `${hours}:${mins}`;
}

function cleanDetailText(raw: string, task?: string): string {
  if (!raw) return "-";
  let s = raw.trim();
  if (s.startsWith("调度已保存:") || s.startsWith("调度已保存：")) {
    const params = s.replace(/^调度已保存[:：]\s*/, "");
    return `重新应用当前调度（${params}，无参数变动）`;
  }
  if (s === "配置未改动" || s === "无变更") {
    return "重新应用当前配置 (无参数改动)";
  }
  s = s.replace(/^(调度变更|配置变更|系统配置变更|合并任务|清理任务)[:：]\s*/, "");
  s = s.replace(/([^\s]+)\s*→\s*([^\s]+)/g, "$1 → $2");
  s = s.replace(/,\s*/g, "，");
  return s;
}

function formatDetail(row: HistoryRecord): string {
  if (row.detail) return row.detail;
  const p: string[] = [];
  if (row.task === "merge") {
    if (row.files_count) p.push(`合并 ${row.files_count} 个分片`);
    if (row.merged_bytes) p.push(formatBytes(row.merged_bytes));
  } else if (row.task === "clean") {
    if (row.files_count) p.push(`清理 ${row.files_count} 个文件`);
    if (row.freed_bytes) p.push(`释放 ${formatBytes(row.freed_bytes)}`);
  } else {
    if (row.files_count) p.push(`${row.files_count} 个文件`);
  }
  if (row.duration) p.push(`${row.duration}s`);
  return p.join(" · ") || "已完成";
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
.dash-header h1 { font-size: 28px; font-weight: 600; color: var(--ink); letter-spacing: -0.5px; }
.dash-date { font-size: 12px; color: #888888; margin-top: 2px; font-family: var(--font-mono); letter-spacing: 0.2px; }

/* Environment status — inline in header */
.env-status { display: flex; align-items: center; gap: 8px; padding-bottom: 4px; }
.env-item { display: flex; align-items: center; gap: 5px; font-size: 12px; }
.env-dot { width: 6px; height: 6px; border-radius: 50%; display: inline-block; flex-shrink: 0; }
.env-label { color: var(--stone); font-size: 12px; }
.env-val {
  font-size: 12px; font-weight: 500; }
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

.stat-label { font-size: 13px; color: var(--slate); margin-bottom: 6px; }
.stat-unit {
  font-size: 14px;
  font-weight: 500;
  color: var(--steel, #64748b);
  margin-left: 4px;
  font-family: var(--font-mono);
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: var(--ink);
  line-height: 1;
  letter-spacing: -0.5px;
  font-family: var(--font-mono);
  height: 36px;
  display: flex;
  align-items: baseline;
}
.stat-sub {
  font-size: 12px;
  color: var(--steel);
  margin-top: 4px;
  font-family: var(--font-mono);
  white-space: nowrap;
}


/* Cards */
.grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

.card {
  background: var(--surface-card); border: 1px solid var(--hairline);
  border-radius: var(--r-md); overflow: hidden;
  box-shadow: none;
  display: flex; flex-direction: column;
}

.card-title {
  padding: 14px 20px; font-size: 14px; font-weight: 600; color: var(--ink);
  border-bottom: 1px solid var(--hairline);
  display: flex; justify-content: space-between; align-items: center;
}
.card-title-sub {
  font-size: 12px;
  font-weight: 400;
  color: var(--stone);
}
.card-body { padding: 20px; display: flex; flex-direction: column; flex: 1; }

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
  display: flex; align-items: center; justify-content: space-between; gap: 8px;
  padding: 10px 14px;
  background: var(--surface);
  min-width: 0;
}
.data-cell:nth-child(odd) { border-right: 1px solid var(--hairline); }
.data-cell:nth-child(1), .data-cell:nth-child(2) { border-bottom: 1px solid var(--hairline); }
.data-label {
  font-size: 13px;
  color: var(--steel);
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.label-help-ico {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border-radius: 50%; font-size: 11px;
  font-weight: 600;
  color: var(--steel);
  background: var(--surface);
  border: 1px solid var(--hairline);
  cursor: help;
  line-height: 1;
  transition: all 0.15s ease;

  &:hover {
    color: var(--charcoal);
    border-color: var(--steel);
    background: var(--highlight);
  }
}

.info-trigger {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  cursor: help;
}
.info-bubble {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: var(--surface-hover);
  color: var(--stone); font-size: 11px;
  font-weight: 600;
}

/* Title Header & Runtime Monitor Card */
.card-title-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.card-title-main {
  font-size: 14px;
  font-weight: 600;
  color: var(--ink);
  white-space: nowrap;
}
.card-header-action {
  font-size: 12px;
  font-weight: 500;
  color: var(--steel);
  text-decoration: none;
  padding: 3px 8px;
  border-radius: var(--r-sm);
  transition: all 0.15s ease;
  display: inline-flex;
  align-items: center;
  gap: 2px;
  white-space: nowrap;
}
.card-header-action:hover {
  color: var(--ink);
  background: var(--highlight);
}
.monitor-mode-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 500;
  color: var(--steel);
  background: var(--surface);
  border: 1px solid var(--hairline);
  padding: 2px 8px;
  border-radius: 12px;
  white-space: nowrap;
}
.status-pulse-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
  display: inline-block;
}
.dot-idle {
  background: #71717a;
  box-shadow: none;
}
.dot-in-service {
  background: #10b981;
  box-shadow: none;
}
.monitor-mode-tag.tag-in-service {
  background: rgba(16, 185, 129, 0.08);
  border-color: rgba(16, 185, 129, 0.20);
  color: #047857;
}
.monitor-mode-tag.tag-running {
  background: #fff6ed;
  border-color: #ffd8af;
  color: #b35600;
}
.dot-active {
  background: #b45309;
  box-shadow: 0 0 0 3px rgba(217, 115, 13, 0.25);
  animation: pulse 1.6s infinite;
}
@keyframes pulse {
  0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(217, 115, 13, 0.5); }
  70% { transform: scale(1.1); box-shadow: 0 0 0 6px rgba(217, 115, 13, 0); }
  100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(217, 115, 13, 0); }
}

.runtime-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}
.runtime-label {
  font-size: 14px;
  color: var(--slate);
}
.runtime-badge {
  display: inline-flex;
  align-items: center;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}
.badge-idle {
  background: #f4f4f5;
  color: #71717a;
  border: 1px solid #e4e4e7;
}
.badge-running {
  background: #fff6ed;
  color: #b35600;
  border: 1px solid #ffd8af;
}

.runtime-substrip {
  margin-bottom: 14px;
  min-height: 20px;
}
.substrip-text {
  font-size: 12px;
  color: var(--stone);
  line-height: 1.4;
}

.bottom-health-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px dashed var(--hairline);
  width: 100%;
}
.health-pill {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 28px;
  box-sizing: border-box;
  font-size: 12px;
  color: var(--slate);
  background: var(--surface);
  border: 1px solid var(--hairline);
  padding: 0 10px;
  border-radius: var(--r-sm);
  white-space: nowrap;
}
.health-pill-warn {
  color: var(--warning);
  border-color: rgba(217, 115, 13, 0.3);
  background: #fff8f0;
}
.quick-tasks-btn {
  font-size: 12px;
  color: var(--steel);
  text-decoration: none;
  padding: 2px 6px;
  border-radius: var(--r-sm);
  transition: all 0.15s;
  margin-left: auto;
}
.quick-tasks-btn:hover {
  color: var(--ink);
  background: var(--highlight);
}

/* Trend Card */
.trend-sub-summary {
  font-size: 12px;
  color: var(--stone);
  font-family: var(--font-mono);
  font-weight: 400;
}
.trend-body {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-height: 200px;
  height: 200px;
}
.trend-chart {
  display: flex;
  justify-content: space-around;
  align-items: flex-end;
  height: 110px;
  width: 100%;
}
.trend-col { display: flex; flex-direction: column; align-items: center; gap: 6px; }
.trend-bars { display: flex; gap: 2px; align-items: flex-end; }
.trend-bar { width: 10px; border-radius: 2px 2px 0 0; min-height: 4px; transition: height 0.3s; cursor: pointer; }
.trend-bar:hover { opacity: 0.75; }
.bar-merge { background: #787774; }
.bar-clean { background: var(--success); }
.trend-date { font-size: 12px; color: var(--stone); font-family: var(--font-mono); letter-spacing: 0.2px; }

.trend-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 110px;
  gap: 4px;
}
.trend-empty-icon { font-size: 24px; opacity: 0.6; }
.trend-empty-title { font-size: 13px; font-weight: 500; color: var(--charcoal); }
.trend-empty-desc { font-size: 12px; color: var(--stone); text-align: center; }

.trend-legend { display: flex; gap: 16px; justify-content: center; margin-top: 8px; font-size: 13px; color: var(--steel); }
.legend-dot { display: inline-block; width: 10px; height: 10px; border-radius: 3px; margin-right: 4px; vertical-align: middle; }

/* Recent Operations */
.view-all-btn {
  background: transparent; border: none; padding: 2px 8px;
  font-size: 13px; color: var(--steel); cursor: pointer;
  border-radius: var(--r-sm); transition: all 0.15s; font-weight: 500;
}
.view-all-btn:hover { color: var(--ink); background: var(--highlight); }

.recent-body {
  display: flex;
  flex-direction: column;
  min-height: 200px;
  height: 200px;
  overflow-y: auto;
}
.recent-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 4px;
}
.recent-empty-icon { font-size: 24px; opacity: 0.6; }
.recent-empty-title { font-size: 13px; font-weight: 500; color: var(--charcoal); }
.recent-empty-desc { font-size: 12px; color: var(--stone); text-align: center; }

.recent-list { display: flex; flex-direction: column; gap: 3px; }

.recent-top-line {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}
.desktop-action {
  display: inline;
}
.mobile-action {
  display: none;
}

.recent-row {
  display: flex; align-items: center; justify-content: space-between;
  padding: 7px 8px; border-radius: var(--r-sm);
  transition: background 0.1s;
}
.recent-row:hover { background: var(--highlight); }
.recent-left { display: flex; align-items: center; gap: 8px; font-size: 13px; min-width: 0; }
.recent-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 4px;
  background: var(--surface);
  font-size: 13px;
  flex-shrink: 0;
}
.recent-tag {
  font-size: 12px;
  font-weight: 500;
  padding: 1px 6px;
  border-radius: var(--r-xs);
  background: var(--surface-secondary, rgba(0, 0, 0, 0.04));
  color: var(--slate);
  border: 1px solid var(--hairline);
  flex-shrink: 0;
  line-height: 1.4;
}
.tag-merge { background: rgba(16, 185, 129, 0.08); border-color: rgba(16, 185, 129, 0.2); color: #059669; }
.tag-clean { background: rgba(245, 158, 11, 0.08); border-color: rgba(245, 158, 11, 0.2); color: #b45309; }
.tag-config { background: rgba(99, 102, 241, 0.08); border-color: rgba(99, 102, 241, 0.2); color: #4f46e5; }
.tag-schedule { background: rgba(100, 116, 139, 0.08); border-color: rgba(100, 116, 139, 0.2); color: #475569; }

html.dark {
  .tag-merge { background: rgba(16, 185, 129, 0.12); border-color: rgba(16, 185, 129, 0.25); color: #34d399; }
  .tag-clean { background: rgba(245, 158, 11, 0.12); border-color: rgba(245, 158, 11, 0.25); color: #fbbf24; }
  .tag-config { background: rgba(99, 102, 241, 0.12); border-color: rgba(99, 102, 241, 0.25); color: #818cf8; }
  .tag-schedule { background: rgba(148, 163, 184, 0.12); border-color: rgba(148, 163, 184, 0.25); color: #94a3b8; }
}

.recent-who { color: var(--steel); flex-shrink: 0; min-width: 48px; font-size: 12px; }
.recent-action { color: var(--charcoal); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; font-size: 13px; }
.recent-right { display: inline-flex; align-items: center; gap: 8px; flex-shrink: 0; }
.recent-status-dot { width: 6px; height: 6px; border-radius: 50%; display: inline-block; flex-shrink: 0; margin: 0; }
.dot-emerald { background: #10b981; }
.dot-fail { background: #ef4444; }
.recent-fail { font-size: 11px; color: #ef4444; font-weight: 500; background: rgba(239, 68, 68, 0.1); padding: 1px 6px; border-radius: var(--r-xs); line-height: 1.4; }
.recent-time { font-family: var(--font-mono); font-size: 12px; color: var(--steel); min-width: 38px; text-align: right; font-feature-settings: "tnum"; }

@media (max-width: 1024px) {
  .stat-grid { grid-template-columns: repeat(2, 1fr); }
  .grid-2 { grid-template-columns: 1fr; }
}
@media (max-width: 600px) {
  .stat-grid { grid-template-columns: repeat(2, 1fr); gap: 10px; }
  .dash-header { flex-direction: column; align-items: flex-start; gap: 8px; }
  .dash-header h1 { font-size: 20px; line-height: 1.3; }
}

@media (max-width: 640px) {
  .data-grid {
    grid-template-columns: repeat(2, 1fr) !important;
  }
  .data-cell {
    padding: 8px 10px !important;
    flex-direction: column !important;
    align-items: flex-start !important;
    justify-content: center !important;
    gap: 3px !important;
  }
  .data-label {
    font-size: 12px !important;
    color: var(--steel) !important;
  }
  .data-cell .mono-val {
    white-space: nowrap !important;
    font-size: 13px !important;
    font-weight: 600 !important;
  }
  .bottom-health-bar {
    gap: 6px !important;
  }
  .health-pill {
    font-size: 12px !important;
    padding: 0 6px !important;
    min-width: 0 !important;
  }
}

/* Modern Design Polish */
.env-dot.dot-ok {
  background: #448361;
  animation: livePulse 2.4s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

.stat-card {
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02), 0 4px 12px -2px rgba(0, 0, 0, 0.02);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.03), 0 12px 24px -6px rgba(0, 0, 0, 0.05);
}

.card-title-left {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.task-type-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 8px;
  border-radius: var(--r-full);
  font-size: 12px;
  font-weight: 500;
  line-height: 1.4;
  white-space: nowrap;
}
.task-type-pill.type-merge {
  background: #ecfdf5;
  color: #065f46;
  border: 1px solid #a7f3d0;
}
.task-type-pill.type-clean {
  background: #eff6ff;
  color: #1e40af;
  border: 1px solid #bfdbfe;
}
.task-type-pill.type-config {
  background: #fdf4ff;
  color: #86198f;
  border: 1px solid #f5d0fe;
}
.task-type-pill.type-schedule {
  background: #f3f4f6;
  color: #374151;
  border: 1px solid #e5e7eb;
}
.task-svg-sm {
  width: 12px;
  height: 12px;
  flex-shrink: 0;
}


html.dark {
  .stat-peach { border-color: rgba(251, 146, 60, 0.25) !important; }
  .stat-mint { border-color: rgba(52, 211, 153, 0.25) !important; }
  .stat-sky { border-color: rgba(96, 165, 250, 0.25) !important; }
  .stat-lavender { border-color: rgba(167, 139, 250, 0.25) !important; }

  .stat-label { color: #d4d4d8; }
  .stat-sub { color: #a1a1aa; }
  

  .card-title {
    border-bottom-color: var(--hairline);
    color: var(--ink);
  }
  .card-title-main { color: #f4f4f5; }
  .card-header-action {
    color: #a1a1aa;
    &:hover { color: #f4f4f5; background: var(--highlight); }
  }
  .monitor-mode-tag {
    background: #18181b;
    border-color: var(--hairline);
    color: #a1a1aa;
  }
  .data-grid {
    border-color: var(--hairline);
  }
  .data-cell {
    background: #141416;
    border-color: var(--hairline);
  }
  .data-label { color: #a1a1aa; }
  .disk-label { color: #a1a1aa; }
  .runtime-label { color: #a1a1aa; }
  .substrip-text { color: #a1a1aa; }

  .health-pill {
    background: #141416;
    border-color: var(--hairline);
    color: #a1a1aa;
  }
  .health-pill-warn {
    background: rgba(245, 158, 11, 0.12) !important;
    border-color: rgba(245, 158, 11, 0.25) !important;
    color: #fbbf24 !important;
  }

  .badge-idle {
    background: rgba(255, 255, 255, 0.06) !important;
    color: #a1a1aa !important;
    border-color: rgba(255, 255, 255, 0.08) !important;
  }
  .dot-idle {
    background: #a1a1aa !important;
    box-shadow: none !important;
  }
  .badge-running {
    background: rgba(245, 158, 11, 0.12) !important;
    color: #fbbf24 !important;
    border-color: rgba(245, 158, 11, 0.25) !important;
  }

  .recent-row {
    border-bottom-color: var(--hairline);
    &:hover { background: rgba(255, 255, 255, 0.03); }
  }
  .recent-who { color: #a1a1aa; }
  .recent-action { color: #e4e4e7; }
  .recent-fail {
    background: rgba(239, 68, 68, 0.15) !important;
    color: #f87171 !important;
  }
  .recent-time { color: #71717a; }

  .trend-empty {
    background: #141416;
    border-color: var(--hairline);
  }
  .trend-empty-title { color: #e4e4e7; }
  .trend-empty-desc { color: #71717a; }
  .trend-date { color: #71717a; }
  .trend-legend { color: #a1a1aa; }

  .task-type-pill.type-merge {
    background: rgba(16, 185, 129, 0.12) !important;
    color: #34d399 !important;
    border-color: rgba(16, 185, 129, 0.25) !important;
  }
  .task-type-pill.type-clean {
    background: rgba(59, 130, 246, 0.12) !important;
    color: #60a5fa !important;
    border-color: rgba(59, 130, 246, 0.25) !important;
  }
  .task-type-pill.type-config {
    background: rgba(168, 85, 247, 0.12) !important;
    color: #c084fc !important;
    border-color: rgba(168, 85, 247, 0.25) !important;
  }
  .task-type-pill.type-schedule {
    background: rgba(148, 163, 184, 0.12) !important;
    color: #94a3b8 !important;
    border-color: rgba(148, 163, 184, 0.25) !important;
  }
}

</style>

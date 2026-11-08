<template>
  <div class="card recent-card">
    <div class="card-title">
      <div class="card-title-left">
        <div class="ui-icon-box icon-box-indigo">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <line x1="16" y1="13" x2="8" y2="13"/>
            <line x1="16" y1="17" x2="8" y2="17"/>
          </svg>
        </div>
        <span class="card-title-main">最近操作记录</span>
      </div>
      <router-link to="/history" class="card-header-action" title="前往操作日志查看全部记录">
        查看全部 ↗
      </router-link>
    </div>

    <div class="card-body recent-body">
      <div v-if="recentHistory.length === 0" class="recent-empty empty-illustration">
        <div class="empty-icon-wrapper">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <line x1="16" y1="13" x2="8" y2="13"/>
            <line x1="16" y1="17" x2="8" y2="17"/>
          </svg>
        </div>
        <span class="empty-title">暂无近期操作记录</span>
        <span class="empty-desc">后台自动执行或手动触发的任务记录将保存在此处</span>
      </div>

      <!-- GitHub Flush List (通栏流式列表) -->
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
              <span class="recent-who">
                <span v-if="row.task === 'config' || row.task === 'schedule'" class="scope-global-text">系统全局</span>
                <span v-else-if="!row.streamer || row.streamer === '全局'" class="scope-global-text">所有主播</span>
                <span v-else class="streamer-text" :title="row.streamer">{{ row.streamer }}</span>
              </span>
              <span class="recent-action" :title="cleanDetailText(row.detail || formatDetail(row), row.task)">
                {{ cleanDetailText(row.detail || formatDetail(row), row.task) }}
              </span>
            </div>
            <div class="recent-right">
              <span
                class="recent-status-dot"
                :class="row.status === 'fail' ? 'dot-fail' : (row.status === 'partial' ? 'dot-warn' : 'dot-emerald')"
                :title="row.status === 'fail' ? '执行失败' : (row.status === 'partial' ? '部分完成' : '执行成功')"
              ></span>
              <span class="recent-time">{{ formatShortTime(row.time) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatBytes } from "@/utils/format";
import type { HistoryRecord } from "@/api/types";

defineProps<{
  recentHistory: HistoryRecord[];
}>();

function formatShortTime(t?: string): string {
  if (!t || t.length < 16) return "";
  return t.slice(11, 16);
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
</script>

<style scoped>
.recent-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

.card-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 32px;
  margin-bottom: 16px;
  box-sizing: border-box;
}

.card-title-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-title-main {
  font-size: 15px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.2px;
}

.card-header-action {
  font-size: 12px;
  color: var(--primary);
  text-decoration: none;
  font-weight: 500;
  transition: color 0.15s ease;
}

.card-header-action:hover {
  text-decoration: underline;
}

.recent-body {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  flex: 1;
  min-height: 180px;
}

/* GitHub Flush Ruled List (通栏发丝线列表) */
.recent-list {
  display: flex;
  flex-direction: column;
  margin: 0 -4px;
}

.recent-row {
  padding: 9px 8px;
  border-radius: var(--r-sm);
  background: transparent;
  border: none;
  border-bottom: 1px solid var(--hairline);
  transition: background 0.15s ease;
}

.recent-row:last-child {
  border-bottom: none;
}

.recent-row:hover {
  background: var(--highlight);
}

.recent-top-line {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.recent-left {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  flex: 1;
}

.recent-tag {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: var(--r-xs);
  font-weight: 500;
  flex-shrink: 0;
}

.tag-merge {
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
}

.tag-clean {
  background: rgba(245, 158, 11, 0.12);
  color: #d97706;
}

.tag-config {
  background: rgba(59, 130, 246, 0.12);
  color: #2563eb;
}

.tag-schedule {
  background: rgba(147, 51, 234, 0.12);
  color: #9333ea;
}

.recent-who {
  width: 76px;
  max-width: 76px;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
}

.scope-global-text {
  font-size: 12px;
  color: var(--charcoal);
  font-weight: 500;
}

.streamer-text {
  font-size: 12px;
  font-weight: 500;
  color: var(--charcoal);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}

.recent-action {
  font-size: 12px;
  color: var(--slate);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.recent-right {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.recent-status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.dot-emerald {
  background: #10b981;
}

.dot-warn {
  background: #f59e0b;
}

.dot-fail {
  background: var(--error);
}

.recent-time {
  font-size: 11px;
  font-family: var(--font-mono);
  color: var(--stone);
}
</style>

<template>
  <div class="streamer-table-container">
    <el-table
      :data="streamers"
      :tooltip-options="{ enterable: false, showAfter: 800, placement: 'top' }"
      v-loading="loading"
      style="width: 100%"
      class="streamer-table"
      :default-sort="{ prop: 'size_gb', order: 'descending' }"
    >
      <el-table-column prop="name" label="主播" min-width="260" show-overflow-tooltip sortable>
        <template #default="{ row }">
          <div class="streamer-name-cell">
            <span
              class="live-status-dot"
              :class="isStreamerActive(row as StreamerInfo) ? 'dot-recording' : 'dot-idle'"
              :title="isStreamerActive(row as StreamerInfo) ? '正在录制或近期有新文件' : '暂无新录像（空闲）'"
            />
            <span class="streamer-name">{{ row.name }}</span>
            <span v-if="row.rule?.protected" class="badge-pill badge-protected">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
              </svg>
              <span>免删保护</span>
            </span>
            <span v-if="row.rule?.trim_enabled" class="badge-pill badge-trim">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="6" cy="6" r="3"/><circle cx="6" cy="18" r="3"/>
                <line x1="20" y1="4" x2="8.12" y2="15.88"/>
                <line x1="14.47" y1="14.48" x2="20" y2="20"/>
                <line x1="8.12" y1="8.12" x2="12" y2="12"/>
              </svg>
              <span>{{ row.rule.trim_start || '00:00:00' }} ~ {{ row.rule.trim_end || '结束' }}</span>
            </span>
          </div>
        </template>
      </el-table-column>

      <el-table-column
        prop="files"
        label="录像文件数量"
        width="140"
        align="right"
        header-align="right"
        sortable
        :sort-orders="['descending', 'ascending']"
      >
        <template #header>
          <el-tooltip content="包含未合并的视频分片与已合并的成片文件总数" placement="top" :show-after="500">
            <span class="col-header-tip">录像文件数 ⓘ</span>
          </el-tooltip>
        </template>
        <template #default="{ row }">
          <div class="data-block">
            <span class="mono-val">{{ row.files }}</span>
            <span class="unit-lbl">个</span>
          </div>
        </template>
      </el-table-column>

      <el-table-column
        prop="size_gb"
        label="磁盘占用"
        width="125"
        align="right"
        header-align="right"
        sortable
        :sort-orders="['descending', 'ascending']"
      >
        <template #default="{ row }">
          <div class="data-block">
            <span class="mono-val">{{ row.size_gb?.toFixed(2) }}</span>
            <span class="unit-lbl">GB</span>
          </div>
        </template>
      </el-table-column>

      <el-table-column
        prop="mtime"
        label="最近录像时间"
        width="130"
        align="right"
        header-align="right"
        sortable
        :sort-orders="['descending', 'ascending']"
        :sort-method="sortMtime"
      >
        <template #header>
          <el-tooltip content="该主播名下最近一个录像文件的生成或修改时间" placement="top" :show-after="500">
            <span class="col-header-tip">最近录像时间 ⓘ</span>
          </el-tooltip>
        </template>
        <template #default="{ row }">
          <el-tooltip
            v-if="row.mtime"
            :content="`${formatExactTime(row.mtime)} (${formatRelativeTime(row.mtime, tick)})`"
            placement="top"
            :show-after="800"
            :enterable="false"
          >
            <span class="mono-val cursor-help">{{ formatRelativeTime(row.mtime, tick) }}</span>
          </el-tooltip>
          <span v-else class="mono-val">-</span>
        </template>
      </el-table-column>

      <el-table-column label="操作" width="225" align="left" header-align="left">
        <template #default="{ row }">
          <div class="action-btn-group">
            <button class="ghost-action" @click="$emit('open-drawer', row as StreamerInfo)" title="查看该主播全部录像文件">
              <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
              </svg>
              <span>查看录像</span>
            </button>
            <button class="ghost-action" @click="$emit('open-rule', row as StreamerInfo)" title="配置该主播专属剪辑与免删保护规则">
              <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="3"/>
                <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
              </svg>
              <span>个性化规则</span>
            </button>
          </div>
        </template>
      </el-table-column>

      <template #empty>
        <EmptyState
          title="未发现匹配的主播"
          desc="请确认搜索关键词或检查录像存储目录中是否存在主播子文件夹"
          :action-text="searchQuery ? '清空搜索条件' : '刷新列表'"
          @action="searchQuery ? $emit('clear-search') : $emit('refresh')"
        >
          <template #icon>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
              <circle cx="12" cy="7" r="4"/>
            </svg>
          </template>
        </EmptyState>
      </template>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import EmptyState from "@/components/common/EmptyState.vue";
import { formatExactTime, formatRelativeTime } from "@/utils/format";
import type { StreamerInfo } from "@/api/types";

defineProps<{
  streamers: StreamerInfo[];
  loading: boolean;
  searchQuery: string;
  tick: number;
}>();

defineEmits<{
  (e: "open-drawer", row: StreamerInfo): void;
  (e: "open-rule", row: StreamerInfo): void;
  (e: "clear-search"): void;
  (e: "refresh"): void;
}>();

function isStreamerActive(row: StreamerInfo): boolean {
  if (!row.mtime) return false;
  return (Date.now() / 1000 - row.mtime) < 1800;
}

function sortMtime(a: StreamerInfo, b: StreamerInfo): number {
  return (a.mtime || 0) - (b.mtime || 0);
}
</script>

<style scoped>
.streamer-table-container {
  width: 100%;
}

.streamer-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.live-status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.dot-recording {
  background: #10b981;
  box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7);
  animation: pulseRecording 1.8s infinite;
}

.dot-idle {
  background: var(--stone);
  opacity: 0.5;
}

@keyframes pulseRecording {
  0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7); }
  70% { transform: scale(1); box-shadow: 0 0 0 5px rgba(16, 185, 129, 0); }
  100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(16, 185, 129, 0); }
}

.streamer-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--charcoal);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.col-header-tip {
  cursor: help;
  color: var(--slate);
  font-weight: 500;
  white-space: nowrap;
}

.action-btn-group {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  white-space: nowrap;
  flex-wrap: nowrap;
}

.ghost-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  padding: 4px 10px;
  height: 28px;
  border-radius: var(--r-sm);
  font-size: 12px;
  font-weight: 500;
  border: 1px solid var(--hairline);
  background: var(--surface);
  color: var(--slate);
  cursor: pointer;
  transition: all 0.15s ease;
  line-height: 1.4;
  white-space: nowrap;
  flex-shrink: 0;
  box-sizing: border-box;
}

.ghost-action:hover {
  background: var(--highlight);
  border-color: var(--charcoal);
  color: var(--ink);
}

.btn-svg-sm {
  width: 13px;
  height: 13px;
  flex-shrink: 0;
}

.cursor-help {
  cursor: help;
}

.data-block {
  display: inline-flex;
  align-items: baseline;
  justify-content: flex-end;
  gap: 4px;
}

.mono-val {
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  color: var(--charcoal);
  font-size: 13px;
}

.unit-lbl {
  font-size: 11px;
  color: var(--steel);
  font-weight: 500;
  font-family: var(--font-sans);
}
</style>

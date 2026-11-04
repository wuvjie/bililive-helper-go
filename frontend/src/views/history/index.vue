<template>
  <div class="history">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>📋 操作日志</span>
          <div class="header-actions">
            <el-select v-model="filterTask" placeholder="全部类型" clearable size="small" style="width: 112px" @change="loadHistory(1)">
              <el-option label="合并" value="merge" />
              <el-option label="清理" value="clean" />
              <el-option label="配置" value="config" />
              <el-option label="调度" value="schedule" />
            </el-select>
            <el-input
              v-model="filterStreamer"
              placeholder="搜索主播名"
              clearable
              size="small"
              style="width: 140px"
              @input="onStreamerSearch"
              @clear="loadHistory(1)"
            />
            <span class="action-divider"></span>
            <button class="text-action" @click="loadHistory(currentPage)">
              <svg class="action-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" /></svg>
              刷新
            </button>
            <button class="text-action" @click="handleExport">
              <svg class="action-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3" /></svg>
              导出
            </button>
          </div>
        </div>
      </template>

      <div class="table-responsive-wrapper">
      <el-table :data="history" v-loading="loading" class="history-table">
        <el-table-column prop="time" label="时间" width="180" sortable>
          <template #default="{ row }">
            <span class="mono-time">{{ row.time }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="task" label="类型" width="90" align="center" sortable>
          <template #default="{ row }">
            <span class="type-label">
              <span class="type-icon">{{ row.task === 'merge' ? '🔗' : row.task === 'clean' ? '🧹' : row.task === 'config' ? '⚙️' : '⏱️' }}</span>
              {{ taskLabel(row.task) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="streamer" label="主播" width="120" show-overflow-tooltip sortable />
        <el-table-column label="详情" min-width="250">
          <template #default="{ row }">
            <span class="detail-text">{{ row.detail || formatDetail(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="right" header-align="right" sortable>
          <template #default="{ row }">
            <span class="status-dot-text" :class="row.status === 'success' ? 'dot-success' : 'dot-fail'">
              {{ row.status === 'success' ? '成功' : '失败' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="" width="80" align="center">
          <template #default="{ row }">
            <button
              v-if="row.log_id && (row.task === 'merge' || row.task === 'clean')"
              class="log-view-btn"
              @click="viewLog(row)"
            >
              查看
            </button>
            <button
              v-if="row.task === 'merge' && row.status === 'fail' && row.streamer"
              class="retry-btn"
              @click="retryMerge(row)"
            >
              重试
            </button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination" v-if="totalPages > 1">
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="20"
          :total="totalItems"
          layout="prev, pager, next, total"
          @current-change="loadHistory"
        />
      </div>
    </el-card>

    <!-- Log Viewer Dialog (Dark Terminal Style) -->
    <el-dialog
      v-model="logDialogVisible"
      :title="`📋 操作日志: ${activeLogTitle}`"
      width="780px"
      top="6vh"
      destroy-on-close
      class="terminal-modal-dialog"
    >
      <div class="log-modal-body" v-loading="logLoading">
        <div class="log-modal-bar">
          <span class="log-meta-tag">{{ activeLogMeta }}</span>
          <button class="btn-copy-log" @click="copyLogContent">
            📋 复制全部日志
          </button>
        </div>
        <pre class="terminal-dark-code">{{ logContent }}</pre>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onActivated } from "vue";
import router from "@/router";
import { ElMessage } from "element-plus";
import { getHistory, exportHistory, getLogContent } from "@/api/history";
import { formatBytes } from "@/utils/format";
import type { HistoryRecord } from "@/api/types";

const history = ref<HistoryRecord[]>([]);
const loading = ref(false);
const currentPage = ref(1);
const totalItems = ref(0);
const totalPages = ref(0);
const filterTask = ref("");
const filterStreamer = ref("");

const logDialogVisible = ref(false);
const logContent = ref("");
const logLoading = ref(false);
const activeLogTitle = ref("");
const activeLogMeta = ref("");

async function copyLogContent() {
  if (!logContent.value) return;
  try {
    await navigator.clipboard.writeText(logContent.value);
    ElMessage.success("已复制完整日志到剪贴板");
  } catch {
    ElMessage.error("复制失败，请手动选择复制");
  }
}


let loadSeq = 0;
let streamerDebounceTimer: ReturnType<typeof setTimeout> | null = null;

function onStreamerSearch() {
  if (streamerDebounceTimer) clearTimeout(streamerDebounceTimer);
  streamerDebounceTimer = setTimeout(() => loadHistory(1), 300);
}

async function loadHistory(page: number) {
  const seq = ++loadSeq;
  loading.value = true;
  try {
    const params: { page: number; per_page: number; task?: string; streamer?: string } = { page, per_page: 20 };
    if (filterTask.value) params.task = filterTask.value;
    if (filterStreamer.value) params.streamer = filterStreamer.value;
    const res = await getHistory(params);
    if (seq !== loadSeq) return;
    history.value = res.items || [];
    totalItems.value = res.total;
    totalPages.value = res.pages;
    currentPage.value = page;
  } finally {
    if (seq === loadSeq) loading.value = false;
  }
}

function taskLabel(task: string) {
  const map: Record<string, string> = { merge: "合并", clean: "清理", config: "配置", schedule: "调度" };
  return map[task] || task;
}

function formatDetail(row: HistoryRecord | Record<string, any>) {
  if (row.detail) return row.detail;
  const parts = [];
  if (row.task === "merge") {
    if (row.files_count) parts.push(`合并 ${row.files_count} 个分片`);
    if (row.merged_bytes) parts.push(formatBytes(row.merged_bytes));
  } else if (row.task === "clean") {
    if (row.files_count) parts.push(`清理 ${row.files_count} 个文件`);
    if (row.freed_bytes) parts.push(`释放 ${formatBytes(row.freed_bytes)}`);
  } else {
    if (row.files_count) parts.push(`${row.files_count} 个文件`);
  }
  if (row.duration) parts.push(`耗时 ${row.duration}s`);
  return parts.join(" · ") || "-";
}

async function viewLog(row: HistoryRecord | Record<string, any>) {
  if (!row.log_id) return;
  if (!/^[a-z]+_\d{8}_\d{6}_[0-9a-f]{4}$/.test(row.log_id)) {
    ElMessage.warning("该记录的日志格式不兼容，请刷新页面");
    return;
  }
  const taskName = row.task === 'merge' ? '分片合并' : '磁盘清理';
  activeLogTitle.value = `${taskName} · ${row.streamer || '全局'}`;
  activeLogMeta.value = `任务 ID: ${row.log_id} · ${row.time} (${row.status === 'success' ? '成功' : '失败'})`;
  logDialogVisible.value = true;
  logContent.value = "";
  logLoading.value = true;
  try {
    const res = await getLogContent(row.task, row.log_id);
    logContent.value = res.content || "暂无日志文本";
  } catch {
    logContent.value = "日志加载失败";
  } finally {
    logLoading.value = false;
  }
}

async function handleExport() {
  try {
    const data = await exportHistory();
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `history-${new Date().toISOString().slice(0, 10)}.json`;
    a.click();
    URL.revokeObjectURL(url);
    ElMessage.success("导出成功");
  } catch {
    // Error handled by interceptor
  }
}

function retryMerge(row: HistoryRecord | Record<string, any>) {
  if (!row.streamer) return;
  router.push({ path: "/tasks", query: { streamer: row.streamer } });
}

onMounted(() => loadHistory(1));

// Refresh data when component is re-activated by keep-alive router-view
onActivated(() => loadHistory(currentPage.value));
</script>

<style scoped>
.table-responsive-wrapper {
  width: 100%;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}
:deep(.history-table) {
  min-width: 780px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}
.header-actions { display: flex; gap: 8px; align-items: center; margin-right: -4px; }
.action-divider { width: 1px; height: 14px; background: var(--hairline); flex-shrink: 0; }
/* Pure text action buttons with inline SVG icons */
.text-action {
  background: transparent; border: none; padding: 0;
  font-size: 13px; font-weight: 500; color: var(--steel);
  cursor: pointer; transition: color 0.15s;
  display: inline-flex; align-items: center; gap: 4px;
}
.text-action:hover { color: var(--ink); }
.action-icon { width: 14px; height: 14px; flex-shrink: 0; }

/* Sort arrow alignment — compensate leftward push on sortable headers */
.history-table :deep(.el-table__header) th.is-sortable .cell {
  padding-left: 3px;
}

/* Pagination — align right edge with table border */
.pagination { margin-top: 16px; display: flex; justify-content: flex-end; padding-right: 0; }
.pagination :deep(.el-pagination) { padding: 0; }

/* Time column */
.mono-time { font-family: var(--font-mono); font-size: 12px; color: var(--slate); letter-spacing: 0.2px; }

/* Type label — pure text + icon, no background */
.type-label {
  display: inline-flex; align-items: center; gap: 4px;
  font-size: 12px; color: var(--slate); font-weight: 500;
}
.type-icon { font-size: 12px; transform: translateY(0.5px); }

/* Detail text — mono for technical params */
.detail-text { font-size: 13px; color: var(--charcoal); font-family: var(--font-mono); letter-spacing: -0.2px; }

/* Status dot — no background block */
.status-dot-text {
  display: inline-flex; align-items: center; gap: 5px;
  font-size: 12px; font-weight: 500;
}
.status-dot-text::before {
  content: ''; display: inline-block; width: 6px; height: 6px; border-radius: 50%;
}
.dot-success { color: var(--slate); }
.dot-success::before { background: #448361; }
.dot-fail { color: var(--slate); }
.dot-fail::before { background: #e03131; }

/* View log button — default invisible, hover appears */
.log-view-btn {
  background: transparent; border: none; padding: 2px 6px;
  font-size: 13px; color: transparent; cursor: pointer;
  border-radius: var(--r-xs); transition: all 0.15s;
}
.history-table :deep(.el-table__row:hover) .log-view-btn {
  color: var(--ink);
}
.log-view-btn:hover {
  text-decoration: underline;
}

/* Retry button — visible on failed merge rows */
.retry-btn {
  background: transparent; border: none; padding: 2px 6px;
  font-size: 13px; color: transparent; cursor: pointer;
  border-radius: var(--r-xs); transition: all 0.15s;
}
.history-table :deep(.el-table__row:hover) .retry-btn {
  color: #d9730d;
}
.retry-btn:hover {
  text-decoration: underline;
}

/* Log viewer dialog */
.log-content {
  background: #191919;
  color: #b0b0b0;
  padding: 16px 24px 16px 16px;
  border: 1px solid #2e2e2e;
  border-radius: var(--r-sm);
  max-height: 60vh;
  overflow-y: auto;
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

/* Dialog overrides for Notion style */
.history :deep(.el-dialog) {
  border-radius: 12px !important;
  overflow: hidden;
}
.history :deep(.el-dialog__header) {
  padding: 16px 24px;
  border-bottom: 1px solid #f1f1ef;
  margin-right: 0;
}
.history :deep(.el-dialog__body) {
  padding: 0;
}

.status-pill-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
}
.pill-success {
  background: #edf6f0;
  color: #2b6a48;
  border: 1px solid #d2ecd9;
}
.pill-fail {
  background: #fff0f0;
  color: #d4433a;
  border: 1px solid #ffdcdc;
}

.log-modal-body {
  background: #191919;
  border-radius: 6px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.log-modal-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 14px;
  background: #252526;
  border-bottom: 1px solid #333;
}
.log-meta-tag {
  color: #9cdcfe;
  font-size: 12px;
  font-family: var(--font-mono);
}
.btn-copy-log {
  background: #1e1e1e;
  border: 1px solid #3c3c3c;
  color: #ccc;
  font-size: 12px;
  padding: 3px 10px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-copy-log:hover {
  background: #2a2d2e;
  color: #fff;
  border-color: #555;
}
.terminal-dark-code {
  margin: 0;
  padding: 14px 16px;
  background: #191919;
  color: #d4d4d4;
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.7;
  max-height: 480px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>

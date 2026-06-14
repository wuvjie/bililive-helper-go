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

    <!-- Log Viewer Dialog -->
    <el-dialog v-model="logDialogVisible" title="日志查看" width="70%" top="5vh" destroy-on-close>
      <pre class="log-content" v-loading="logLoading">{{ logContent }}</pre>
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

function formatDetail(row: HistoryRecord) {
  const parts = [];
  if (row.files_count) parts.push(`${row.files_count} 个文件`);
  if (row.merged_bytes) parts.push(`合并 ${formatBytes(row.merged_bytes)}`);
  if (row.freed_bytes) parts.push(`释放 ${formatBytes(row.freed_bytes)}`);
  if (row.duration) parts.push(`耗时 ${row.duration}s`);
  return parts.join(", ") || "-";
}

async function viewLog(row: HistoryRecord) {
  if (!row.log_id) return;
  // 预校验 log_id 格式：{type}_{YYYYMMDD}_{HHMMSS}_{4位hex}
  if (!/^[a-z]+_\d{8}_\d{6}_[0-9a-f]{4}$/.test(row.log_id)) {
    ElMessage.warning("该记录的日志格式不兼容，请刷新页面");
    return;
  }
  logDialogVisible.value = true;
  logContent.value = "";
  logLoading.value = true;
  try {
    logContent.value = await getLogContent(row.task, row.log_id);
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

function retryMerge(row: HistoryRecord) {
  if (!row.streamer) return;
  router.push({ path: "/tasks", query: { streamer: row.streamer } });
}

onMounted(() => loadHistory(1));

// Refresh data when component is re-activated by keep-alive router-view
onActivated(() => loadHistory(currentPage.value));
</script>

<style scoped>
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
</style>

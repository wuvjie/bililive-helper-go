<template>
  <div class="history">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="card-title-group">
            <div class="ui-icon-box">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
            </div>
            <span class="title-text">操作审计日志</span>
          </div>
          <div class="header-actions">
            <div class="header-filters">
              <el-select
                v-model="filterTask"
                placeholder="全部任务类型"
                clearable
                size="small"
                style="width: 128px"
                @change="loadHistory(1)"
              >
                <el-option label="全部任务" value="" />
                <el-option label="分片合并" value="merge" />
                <el-option label="磁盘清理" value="clean" />
                <el-option label="配置变更" value="config" />
                <el-option label="调度事件" value="schedule" />
              </el-select>
              <div class="search-input-wrap">
                <el-input
                  v-model="filterStreamer"
                  placeholder="搜索主播名..."
                  clearable
                  size="small"
                  style="width: 140px"
                  @input="onStreamerSearch"
                  @clear="loadHistory(1)"
                />
                <span v-if="!filterStreamer" class="kbd-badge search-kbd">/</span>
              </div>
            </div>
            <div class="header-buttons">
              <button class="text-action" @click="loadHistory(currentPage)" title="刷新列表">
                <svg class="action-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" /></svg>
                刷新
              </button>
              <button class="text-action" @click="handleExport" title="导出历史记录为 JSON">
                <svg class="action-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3" /></svg>
                导出
              </button>
              <span class="action-divider"></span>
              <button class="text-action text-action-danger" @click="handleClearHistory" title="清空全部操作历史记录与日志文件">
                <svg class="action-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" /></svg>
                清空
              </button>
            </div>
          </div>
        </div>
      </template>

      <div class="table-responsive-wrapper">
      <el-table :data="history" v-loading="loading" class="history-table">
        <el-table-column prop="time" label="时间" width="175" sortable>
          <template #default="{ row }">
            <span class="mono-time">{{ row.time }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="task" label="类型" width="110" align="center" header-align="center" sortable>
          <template #default="{ row }">
            <span class="type-label">
              <span class="task-type-pill" :class="'type-' + row.task">
              <svg v-if="row.task === 'merge'" class="task-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
              <svg v-else-if="row.task === 'clean'" class="task-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>
              <svg v-else-if="row.task === 'config'" class="task-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
              <svg v-else class="task-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
              <span>{{ row.task === 'merge' ? '分片合并' : row.task === 'clean' ? '磁盘清理' : row.task === 'config' ? '配置变更' : '调度事件' }}</span>
            </span>
              {{ taskLabel(row.task) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="streamer" label="主播" width="145" show-overflow-tooltip sortable />
        <el-table-column label="详情" min-width="260">
          <template #default="{ row }">
            <span class="detail-text">{{ row.detail || formatDetail(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="95" align="center" header-align="center" sortable>
          <template #default="{ row }">
            <span class="status-dot-text" :class="row.status === 'success' ? 'dot-success' : 'dot-fail'">
              {{ row.status === 'success' ? '成功' : '失败' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" align="center" header-align="center">
          <template #default="{ row }">
            <div class="row-actions-group">
              <button
                v-if="row.task === 'merge' && row.status === 'fail' && row.streamer"
                class="row-action-btn action-retry"
                @click="retryMerge(row)"
                title="重试合并"
              >
                重试
              </button>
              <button
                v-if="row.log_id && (row.task === 'merge' || row.task === 'clean')"
                class="row-action-btn action-view"
                @click="viewLog(row)"
                title="查看详细运行日志"
              >
                查看
              </button>
              <button
                class="row-action-btn action-delete"
                @click="handleDeleteRecord(row)"
                title="删除此条记录"
              >
                删除
              </button>
            </div>
          </template>
        </el-table-column>
        <template #empty>
          <div class="history-empty-box">
            <div class="empty-icon-wrapper">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
            </div>
            <span class="empty-title">暂无操作日志</span>
            <span class="empty-desc">未找到符合当前类型或主播筛选条件的历史记录</span>
          </div>
        </template>
      </el-table>
      </div>

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

    <!-- Log Viewer Dialog (Dark Terminal Style with Line Numbers and Search) -->
    <el-dialog
      v-model="logDialogVisible"
      :title="`操作审计日志 · ${activeLogTitle}`"
      width="min(880px, 94vw)"
      top="5vh"
      destroy-on-close
      class="terminal-modal-dialog"
    >
      <div class="log-modal-body" v-loading="logLoading">
        <div class="log-modal-bar">
          <div class="log-bar-left">
            <span
              class="log-meta-tag log-meta-id"
              :title="`任务 ID: ${activeLogId || activeLogMeta} (点击复制)`"
              @click="copyToClipboard(activeLogId, '已复制任务 ID')"
            >
              <span class="meta-label">ID:</span> {{ activeLogId || activeLogMeta }}
            </span>
            <span
              v-if="activeLogStatus"
              class="log-meta-status"
              :class="activeLogStatus === 'success' ? 'status-ok' : (activeLogStatus === 'partial' ? 'status-warn' : 'status-fail')"
            >
              {{ activeLogStatus === 'success' ? '成功' : (activeLogStatus === 'partial' ? '部分' : '失败') }}
            </span>
            <span v-if="activeLogTime" class="log-meta-time" :title="`执行时间: ${activeLogTime}`">
              {{ activeLogTime }}
            </span>
            <span class="log-lines-count">{{ formattedLogLines.length }} 行</span>
          </div>
          <div class="log-modal-actions">
            <input
              v-model="logFilter"
              placeholder="过滤本日志..."
              class="log-filter-box"
            />
            <button class="btn-copy-log" @click="downloadRawLog" title="下载完整原始日志 (.log)">
              <svg class="action-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg> 下载
            </button>
            <button class="btn-copy-log" @click="copyLogContent" title="复制全部内容到剪贴板">
              <svg class="action-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg> 复制
            </button>
          </div>
        </div>

        <div class="log-terminal-lines">
          <div v-if="filteredLogLines.length === 0" class="log-empty-lines">
            {{ logFilter ? '未匹配到包含搜索词的日志行' : '日志内容为空' }}
          </div>
          <div
            v-for="(line, idx) in filteredLogLines"
            :key="idx"
            class="log-code-row"
            :class="line.typeClass"
          >
            <span class="log-line-num">{{ idx + 1 }}</span>
            <span class="log-line-text">{{ line.text }}</span>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onActivated } from "vue";
import router from "@/router";
import { ElMessage, ElMessageBox } from "element-plus";
import { getHistory, exportHistory, getLogContent, getRawLogDownloadUrl, deleteHistory, clearHistory } from "@/api/history";
import { copyToClipboard } from "@/utils/clipboard";
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
const activeLogTask = ref("");
const activeLogId = ref("");
const activeLogTime = ref("");
const activeLogStatus = ref("");
const logFilter = ref("");

interface FormattedLogLine {
  text: string;
  typeClass: string;
}

const formattedLogLines = computed<FormattedLogLine[]>(() => {
  if (!logContent.value) return [];
  const rawLines = logContent.value.split("\n");
  return rawLines.map(line => {
    let typeClass = "line-info";
    if (/❌|错误|失败|error|failed|fatal/i.test(line)) {
      typeClass = "line-error";
    } else if (/⚠️|⚠|警告|warn|warning|注意|跳过|空间不足/i.test(line)) {
      typeClass = "line-warn";
    } else if (/✅|完成|成功|success|done/i.test(line)) {
      typeClass = "line-success";
    } else if (/^[─═=—\-\s]{3,}$/.test(line.trim())) {
      typeClass = "line-divider";
    } else if (/📊 结果摘要|▶ 开始/i.test(line)) {
      typeClass = "line-header";
    }
    return { text: line, typeClass };
  });
});

const filteredLogLines = computed(() => {
  if (!logFilter.value.trim()) return formattedLogLines.value;
  const q = logFilter.value.toLowerCase();
  return formattedLogLines.value.filter(l => l.text.toLowerCase().includes(q));
});

function downloadRawLog() {
  const content = logContent.value;
  if (!content) {
    ElMessage.warning("日志内容为空或尚未加载完成");
    return;
  }
  const blob = new Blob([content], { type: "text/plain;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  const filename = `${activeLogTask.value || 'task'}_${activeLogId.value || 'log'}.log`;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
  ElMessage.success(`已开始下载日志文件: ${filename}`);
}

async function copyLogContent() {
  if (!logContent.value) return;
  await copyToClipboard(logContent.value, "已复制完整日志到剪贴板");
}


let loadSeq = 0;
let streamerDebounceTimer: ReturnType<typeof setTimeout> | null = null;

function onStreamerSearch() {
  if (streamerDebounceTimer) clearTimeout(streamerDebounceTimer);
  streamerDebounceTimer = setTimeout(() => loadHistory(1), 300);
}

async function handleDeleteRecord(row: HistoryRecord) {
  try {
    await ElMessageBox.confirm(
      `确定删除 ${row.time} 的 [${taskLabel(row.task)}] 操作记录吗？\n（注：仅清除日志记录，绝不影响主播实际录像视频）`,
      "确认删除记录",
      {
        confirmButtonText: "确认删除",
        cancelButtonText: "取消",
        confirmButtonClass: "el-button--danger",
        type: "warning"
      }
    );
  } catch {
    return;
  }

  try {
    await deleteHistory(row.id);
    ElMessage.success("已删除该条操作记录");
    await loadHistory(currentPage.value);
  } catch {
    // Handled by HTTP interceptor
  }
}

async function handleClearHistory() {
  try {
    await ElMessageBox.confirm(
      "确定要清空全部操作历史记录与关联日志吗？\n\n【安全声明】该操作仅清除后台任务记录和日志文本，绝不影响、修改或删除任何主播的录像视频文件！",
      "清空全部操作日志",
      {
        confirmButtonText: "确认清空",
        cancelButtonText: "取消",
        confirmButtonClass: "el-button--danger",
        type: "warning"
      }
    );
  } catch {
    return;
  }

  try {
    await clearHistory();
    ElMessage.success("已清空全部历史操作记录");
    await loadHistory(1);
  } catch {
    // Handled by HTTP interceptor
  }
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
  activeLogTask.value = row.task || '';
  activeLogId.value = row.log_id || '';
  activeLogTime.value = row.time || '';
  activeLogStatus.value = row.status || '';
  activeLogMeta.value = `任务 ID: ${row.log_id} · ${row.time} (${row.status === 'success' ? '成功' : (row.status === 'partial' ? '部分' : '失败')})`;
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
  padding-bottom: 6px;
}
:deep(.history-table) {
  min-width: 940px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}
.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}
.header-filters {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}
.header-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
  white-space: nowrap;
  flex-shrink: 0;
}
.action-divider {
  width: 1px;
  height: 14px;
  background: var(--hairline, #e2e8f0);
  margin: 0 2px;
  flex-shrink: 0;
}
/* Pure text action buttons with inline SVG icons */
.text-action {
  background: transparent;
  border: none;
  padding: 4px 8px;
  border-radius: var(--r-xs, 4px);
  font-size: 13px;
  font-weight: 500;
  color: var(--steel, #64748b);
  cursor: pointer;
  transition: all 0.15s ease;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  user-select: none;
}
.text-action:hover {
  color: var(--ink, #0f172a);
  background: rgba(0, 0, 0, 0.04);
}
.action-icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
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



.history-empty-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 0;
  gap: 6px;
}
.empty-emoji {
  font-size: 28px;
  line-height: 1;
  margin-bottom: 4px;
}
.empty-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--charcoal);
}
.empty-desc {
  font-size: 12px;
  color: var(--steel);
}

@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
  }
  .header-actions {
    width: 100%;
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
    margin-right: 0;
  }
  .header-filters {
    width: 100%;
    display: flex;
    gap: 8px;
  }
  .header-filters :deep(.el-select) {
    flex: 1;
    min-width: 120px;
  }
  .header-filters :deep(.el-input) {
    flex: 1;
    min-width: 130px;
  }
  .header-buttons {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;
    width: 100%;
  }
  :deep(.terminal-modal-dialog) {
    width: 95% !important;
    max-width: 95% !important;
    margin: 16px auto !important;
  }
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
.log-bar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.log-lines-count {
  font-size: 11px;
  color: #888;
  background: #1e1e1e;
  border: 1px solid #3c3c3c;
  padding: 1px 6px;
  border-radius: 10px;
  font-family: var(--font-mono);
}
.log-modal-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}
.log-filter-box {
  background: #1e1e1e;
  border: 1px solid #3c3c3c;
  color: #d4d4d4;
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 4px;
  outline: none;
  font-family: var(--font-mono);
  width: 120px;
  transition: all 0.15s ease;
}
.log-filter-box:focus {
  border-color: #007fd4;
  width: 150px;
}
.log-terminal-lines {
  background: #191919;
  max-height: 520px;
  overflow-y: auto;
  padding: 10px 0;
  font-family: var(--font-mono);
  font-size: 12px;
}
.log-empty-lines {
  text-align: center;
  padding: 36px 0;
  color: #666;
  font-size: 12px;
}
.log-code-row {
  display: flex;
  line-height: 1.7;
  padding: 1px 12px;
  transition: background 0.1s;
}
.log-code-row:hover {
  background: rgba(255, 255, 255, 0.05);
}
.log-line-num {
  width: 44px;
  flex-shrink: 0;
  text-align: right;
  padding-right: 14px;
  color: #4a4a4a;
  user-select: none;
  font-size: 11px;
}
.log-line-text {
  flex: 1;
  word-break: break-all;
  white-space: pre-wrap;
}
.line-info .log-line-text { color: #cccccc; }
.line-success .log-line-text { color: #6a9955; font-weight: 500; }
.line-warn .log-line-text { color: #d7ba7d; }
.line-error .log-line-text { color: #f44747; font-weight: 500; }
.line-divider .log-line-text { opacity: 0.35; }
.line-header .log-line-text { color: #4ec9b0; font-weight: 600; }

/* Unified row action buttons — optically centered 12px pill styling */
.row-actions-group {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  white-space: nowrap;
  flex-wrap: nowrap;
}
.row-action-btn {
  background: transparent;
  border: none;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.4;
  padding: 3px 8px;
  border-radius: var(--r-xs, 4px);
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s ease;
  user-select: none;
}
.action-view {
  color: var(--steel, #64748b);
}
.action-view:hover {
  color: #0284c7;
  background: rgba(2, 132, 199, 0.08);
}
.action-retry {
  color: #d9730d;
}
.action-retry:hover {
  color: #b45309;
  background: rgba(217, 115, 13, 0.1);
}
.action-delete {
  color: var(--slate, #94a3b8);
}
.action-delete:hover {
  color: #dc2626;
  background: rgba(220, 38, 38, 0.08);
}

/* Stable & Flexible Log Modal Bar */
.log-modal-bar {
  display: flex !important;
  justify-content: space-between !important;
  align-items: center !important;
  padding: 8px 14px !important;
  background: #252526 !important;
  border-bottom: 1px solid #333 !important;
  gap: 10px !important;
  flex-wrap: wrap !important;
}
.log-bar-left {
  display: flex !important;
  align-items: center !important;
  flex-wrap: wrap !important;
  gap: 8px !important;
  min-width: 0 !important;
  flex: 1 1 auto !important;
}
.log-meta-tag {
  color: #9cdcfe !important;
  font-size: 12px !important;
  font-family: var(--font-mono) !important;
}
.log-meta-id {
  cursor: pointer;
  background: rgba(156, 220, 254, 0.08);
  padding: 2px 7px;
  border-radius: 4px;
  border: 1px solid rgba(156, 220, 254, 0.22);
  white-space: nowrap !important;
  max-width: 280px;
  overflow: hidden !important;
  text-overflow: ellipsis !important;
  transition: all 0.15s ease;
  user-select: none;
}
.log-meta-id:hover {
  background: rgba(156, 220, 254, 0.18);
  border-color: #9cdcfe;
  color: #fff !important;
}
.log-meta-id .meta-label {
  color: #858585;
  font-weight: 500;
  margin-right: 3px;
}
.log-meta-status {
  font-size: 11px !important;
  padding: 2px 7px !important;
  border-radius: 4px !important;
  font-weight: 600 !important;
  white-space: nowrap !important;
  flex-shrink: 0 !important;
}
.log-meta-status.status-ok {
  background: rgba(46, 204, 113, 0.15) !important;
  color: #4ade80 !important;
  border: 1px solid rgba(46, 204, 113, 0.3) !important;
}
.log-meta-status.status-warn {
  background: rgba(234, 179, 8, 0.15) !important;
  color: #facc15 !important;
  border: 1px solid rgba(234, 179, 8, 0.3) !important;
}
.log-meta-status.status-fail {
  background: rgba(239, 68, 68, 0.15) !important;
  color: #f87171 !important;
  border: 1px solid rgba(239, 68, 68, 0.3) !important;
}
.log-meta-time {
  font-size: 11px !important;
  color: #a0a0a0 !important;
  font-family: var(--font-mono) !important;
  white-space: nowrap !important;
  flex-shrink: 0 !important;
}
.log-lines-count {
  font-size: 11px !important;
  color: #888 !important;
  background: #1e1e1e !important;
  border: 1px solid #3c3c3c !important;
  padding: 2px 7px !important;
  border-radius: 10px !important;
  font-family: var(--font-mono) !important;
  white-space: nowrap !important;
  flex-shrink: 0 !important;
  display: inline-flex !important;
  align-items: center !important;
}
.log-modal-actions {
  display: flex !important;
  align-items: center !important;
  gap: 6px !important;
  flex-shrink: 0 !important;
}
.log-filter-box {
  background: #1e1e1e !important;
  border: 1px solid #3c3c3c !important;
  color: #d4d4d4 !important;
  font-size: 11px !important;
  padding: 4px 8px !important;
  border-radius: 4px !important;
  outline: none !important;
  font-family: var(--font-mono) !important;
  width: 125px !important;
  box-sizing: border-box !important;
  transition: border-color 0.15s ease !important;
}
.log-filter-box:focus {
  border-color: #007fd4 !important;
}

@media (max-width: 680px) {
  :deep(.terminal-modal-dialog) {
    width: 95% !important;
    max-width: 95% !important;
    margin: 10px auto !important;
  }
  .log-modal-bar {
    flex-direction: column !important;
    align-items: stretch !important;
    gap: 10px !important;
    padding: 10px 12px !important;
  }
  .log-bar-left {
    display: flex !important;
    align-items: center !important;
    gap: 6px !important;
    justify-content: flex-start !important;
  }
  .log-meta-id {
    max-width: 160px !important;
  }
  .log-modal-actions {
    display: flex !important;
    width: 100% !important;
    gap: 6px !important;
  }
  .log-filter-box {
    flex: 1 !important;
    width: auto !important;
  }
  .btn-copy-log {
    flex-shrink: 0 !important;
  }
}

.card-title-group {
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
  font-size: 11px;
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
.empty-icon-wrapper {
  width: 44px;
  height: 44px;
  border-radius: var(--r-md);
  background: var(--surface);
  border: 1px solid var(--hairline);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 10px;
  color: var(--stone);
  svg {
    width: 24px;
    height: 24px;
  }
}


.search-input-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
}
.search-kbd {
  position: absolute;
  right: 8px;
  pointer-events: none;
  opacity: 0.7;
}
.search-input-wrap:focus-within .search-kbd {
  display: none;
}

</style>

<template>
  <div class="history">
    <el-card shadow="hover">
      <template #header>
        <div class="history-header">
          <div class="header-top-bar">
            <div class="card-title-group">
              <div class="ui-icon-box icon-box-indigo">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                  <polyline points="14 2 14 8 20 8"/>
                  <line x1="16" y1="13" x2="8" y2="13"/>
                  <line x1="16" y1="17" x2="8" y2="17"/>
                </svg>
              </div>
              <span class="title-text">操作审计日志</span>
            </div>

            <div class="header-right-tools">
              <div class="search-input-wrap">
                <el-input
                  ref="streamerSearchRef"
                  v-model="filterStreamer"
                  placeholder="搜索主播名..."
                  :prefix-icon="Search"
                  clearable
                  style="width: 220px"
                  @input="onStreamerSearch"
                  @clear="loadHistory(1)"
                />
                <span v-if="!filterStreamer" class="kbd-badge search-kbd">/</span>
              </div>
              <button class="btn-clear-history" @click="handleClearHistory" title="清空全部操作历史记录与日志文件">
                <svg class="action-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                </svg>
                <span>清空日志</span>
              </button>
            </div>
          </div>

          <div class="header-nav-bar">
            <UnderlineNav
              v-model="filterTask"
              :items="taskPillItems"
              @change="onFilterChange"
            />
          </div>
        </div>
      </template>

      <div class="table-responsive-wrapper">
        <el-table
          :data="history"
          v-loading="loading"
          class="history-table"
          :tooltip-options="{ enterable: false, showAfter: 500, placement: 'top' }"
        >
          <el-table-column prop="time" label="时间" width="175" sortable>
            <template #default="{ row }">
              <span class="mono-time">{{ row.time }}</span>
            </template>
          </el-table-column>

          <el-table-column prop="task" label="类型" width="115" align="center" header-align="center" sortable>
            <template #default="{ row }">
              <span class="type-label">
                <span class="task-type-pill" :class="'type-' + row.task">
                  <svg v-if="row.task === 'merge'" class="task-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
                    <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
                  </svg>
                  <svg v-else-if="row.task === 'clean'" class="task-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
                  </svg>
                  <svg v-else-if="row.task === 'config'" class="task-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="3"/>
                    <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
                  </svg>
                  <svg v-else class="task-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
                  </svg>
                  <span>{{ row.task === 'merge' ? '分片合并' : row.task === 'clean' ? '磁盘清理' : row.task === 'config' ? '配置变更' : '调度事件' }}</span>
                </span>
              </span>
            </template>
          </el-table-column>

          <el-table-column prop="streamer" label="主播" width="145" show-overflow-tooltip sortable>
            <template #default="{ row }">
              <span v-if="!row.streamer || row.streamer === '全局'" class="pill-scope-global">全局</span>
              <span v-else>{{ row.streamer }}</span>
            </template>
          </el-table-column>

          <el-table-column label="详情" min-width="260">
            <template #default="{ row }">
              <span class="detail-text">{{ cleanDetailText(row.detail || formatDetail(row), row.task) }}</span>
            </template>
          </el-table-column>

          <el-table-column prop="status" label="状态" width="95" align="center" header-align="center" sortable>
            <template #default="{ row }">
              <span class="status-dot-text" :class="row.status === 'success' ? 'dot-success' : 'dot-fail'">
                {{ row.status === 'success' ? '成功' : '失败' }}
              </span>
            </template>
          </el-table-column>

          <el-table-column label="操作" width="130" align="center" header-align="center">
            <template #default="{ row }">
              <div class="row-actions-group">
                <button
                  v-if="row.task === 'merge' && row.status === 'fail' && row.streamer"
                  class="row-action-btn action-retry"
                  @click="retryMerge(row)"
                  title="前往任务中心处理该主播合并任务"
                >
                  去处理 ↗
                </button>
                <button
                  v-if="row.log_id && (row.task === 'merge' || row.task === 'clean')"
                  class="row-action-btn action-view"
                  @click="viewLog(row)"
                  title="查看详细运行日志"
                >
                  查看
                </button>
              </div>
            </template>
          </el-table-column>

          <template #empty>
            <EmptyState
              title="暂无操作日志"
              desc="未找到符合当前类型或主播筛选条件的历史记录"
              :action-text="filterTask || filterStreamer ? '重置筛选条件' : '刷新日志列表'"
              @action="filterTask || filterStreamer ? resetFilters() : loadHistory(1)"
            >
              <template #icon>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                  <polyline points="14 2 14 8 20 8"/>
                  <line x1="16" y1="13" x2="8" y2="13"/>
                  <line x1="16" y1="17" x2="8" y2="17"/>
                </svg>
              </template>
            </EmptyState>
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

    <!-- Log Viewer Dialog -->
    <LogViewerModal
      v-model:visible="logDialogVisible"
      :title="activeLogTitle"
      :status="activeLogStatus"
      :log-id="activeLogId"
      :loading="logLoading"
      :content="logContent"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onActivated } from "vue";
import router from "@/router";
import { ElMessage, ElMessageBox } from "element-plus";
import { Search } from "@element-plus/icons-vue";
import { getHistory, getLogContent, clearHistory } from "@/api/history";
import { formatBytes } from "@/utils/format";
import type { HistoryRecord } from "@/api/types";

import UnderlineNav from "@/components/common/UnderlineNav.vue";
import EmptyState from "@/components/common/EmptyState.vue";
import LogViewerModal from "./components/LogViewerModal.vue";

const history = ref<HistoryRecord[]>([]);
const loading = ref(false);
const currentPage = ref(1);
const totalItems = ref(0);
const totalPages = ref(0);
const filterTask = ref<string | number>("");
const filterStreamer = ref("");
const streamerSearchRef = ref();

const taskPillItems = [
  { label: "全部", value: "" },
  { label: "分片合并", value: "merge" },
  { label: "磁盘清理", value: "clean" },
  { label: "配置变更", value: "config" },
  { label: "调度事件", value: "schedule" }
];

function handleKeydown(e: KeyboardEvent) {
  if (e.key === "/" && document.activeElement?.tagName !== "INPUT" && document.activeElement?.tagName !== "TEXTAREA") {
    e.preventDefault();
    streamerSearchRef.value?.focus();
  }
}

// Log Modal State
const logDialogVisible = ref(false);
const logContent = ref("");
const logLoading = ref(false);
const activeLogTitle = ref("");
const activeLogId = ref("");
const activeLogStatus = ref("");

function onFilterChange() {
  loadHistory(1);
}

function resetFilters() {
  filterTask.value = "";
  filterStreamer.value = "";
  loadHistory(1);
}

let searchTimer: any = null;
function onStreamerSearch() {
  clearTimeout(searchTimer);
  searchTimer = setTimeout(() => {
    loadHistory(1);
  }, 300);
}

async function loadHistory(page = 1) {
  loading.value = true;
  currentPage.value = page;
  try {
    const res = await getHistory({
      page,
      per_page: 20,
      task: (filterTask.value as string) || undefined,
      streamer: filterStreamer.value || undefined
    });
    history.value = res.items || [];
    totalItems.value = res.total || 0;
    totalPages.value = res.total_pages || 0;
  } catch {
    ElMessage.error("获取操作日志失败");
  } finally {
    loading.value = false;
  }
}

async function handleClearHistory() {
  try {
    await ElMessageBox.confirm(
      "此操作将永久清空全部历史操作日志记录和相关的任务日志文件，是否继续？",
      "清空日志确认",
      {
        type: "warning",
        confirmButtonText: "确定清空",
        cancelButtonText: "取消"
      }
    );
    await clearHistory();
    ElMessage.success("全部操作历史记录已清空");
    loadHistory(1);
  } catch (err: any) {
    if (err !== "cancel") {
      ElMessage.error(err.message || "清空历史记录失败");
    }
  }
}

function retryMerge(row: HistoryRecord) {
  if (row.streamer) {
    router.push({ path: "/tasks", query: { streamer: row.streamer } });
  } else {
    router.push("/tasks");
  }
}

async function viewLog(row: HistoryRecord) {
  if (!row.log_id) return;
  activeLogId.value = row.log_id;
  activeLogStatus.value = row.status;
  const taskName = row.task === "merge" ? "合并" : row.task === "clean" ? "清理" : row.task;
  activeLogTitle.value = `[${taskName}] ${row.streamer || "全局"} · ${row.time}`;
  logDialogVisible.value = true;
  logLoading.value = true;
  logContent.value = "";
  try {
    const res = await getLogContent(row.task, row.log_id);
    logContent.value = res.content || "";
  } catch {
    ElMessage.error("获取日志文件内容失败");
  } finally {
    logLoading.value = false;
  }
}

function cleanDetailText(raw: string, _task?: string): string {
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

onMounted(() => {
  loadHistory(1);
  window.addEventListener("keydown", handleKeydown);
});

onActivated(() => {
  loadHistory(currentPage.value);
});
</script>

<style scoped>
.history {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

:deep(.el-card__header) {
  padding: 0;
  border-bottom: 1px solid var(--hairline);
}

.history-header {
  display: flex;
  flex-direction: column;
}

.header-top-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px 12px 20px;
  gap: 16px;
  flex-wrap: wrap;
}

.header-nav-bar {
  padding: 0 20px;
  display: flex;
  align-items: center;
}

.card-title-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.title-text {
  font-size: 16px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.2px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.header-right-tools {
  display: flex;
  align-items: center;
  gap: 10px;
}

.search-input-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.search-input-wrap :deep(.el-input__wrapper) {
  height: 32px;
  line-height: 32px;
}

.search-kbd {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
}

.btn-clear-history {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 32px;
  box-sizing: border-box;
  padding: 0 12px;
  border-radius: var(--r-sm);
  font-size: 12px;
  font-weight: 500;
  background: transparent;
  border: 1px solid var(--hairline);
  color: var(--steel);
  cursor: pointer;
  transition: all 0.15s ease;
  line-height: 30px;
}

.btn-clear-history:hover {
  background: rgba(239, 68, 68, 0.08);
  border-color: rgba(239, 68, 68, 0.25);
  color: var(--error);
}

.action-icon {
  width: 13px;
  height: 13px;
}

.table-responsive-wrapper {
  width: 100%;
  overflow-x: auto;
}

.mono-time {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--slate);
}

.task-type-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 7px;
  border-radius: var(--r-full);
  font-size: 11px;
  font-weight: 500;
}

.task-svg-sm {
  width: 11px;
  height: 11px;
}

.type-merge {
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
}

.type-clean {
  background: rgba(245, 158, 11, 0.12);
  color: #d97706;
}

.type-config {
  background: rgba(59, 130, 246, 0.12);
  color: #2563eb;
}

.type-schedule {
  background: rgba(147, 51, 234, 0.12);
  color: #9333ea;
}

.detail-text {
  font-size: 12px;
  color: var(--charcoal);
  line-height: 1.4;
}

.status-dot-text {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  font-weight: 500;
}

.status-dot-text::before {
  content: "";
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.dot-success {
  color: #059669;
}
.dot-success::before {
  background: #10b981;
}

.dot-fail {
  color: var(--error);
}
.dot-fail::before {
  background: var(--error);
}

.row-actions-group {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.row-action-btn {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: var(--r-xs);
  font-size: 11px;
  font-weight: 500;
  border: 1px solid var(--hairline);
  background: var(--surface);
  color: var(--slate);
  cursor: pointer;
  transition: all 0.15s ease;
}

.row-action-btn:hover {
  background: var(--highlight);
  border-color: var(--charcoal);
  color: var(--ink);
}

.action-retry {
  background: rgba(239, 68, 68, 0.08);
  border-color: rgba(239, 68, 68, 0.2);
  color: var(--error);
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>

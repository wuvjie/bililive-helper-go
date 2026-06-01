<template>
  <div class="history">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>📋 操作日志</span>
          <div class="header-actions">
            <el-select v-model="filterTask" placeholder="全部类型" clearable style="width: 120px" @change="loadHistory(1)">
              <el-option label="合并" value="merge" />
              <el-option label="清理" value="clean" />
              <el-option label="配置" value="config" />
            </el-select>
            <el-button size="small" @click="loadHistory(currentPage)">
              <el-icon><Refresh /></el-icon>刷新
            </el-button>
            <el-button size="small" type="success" @click="handleExport">
              <el-icon><Download /></el-icon>导出
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="history" stripe v-loading="loading">
        <el-table-column prop="time" label="时间" width="170" />
        <el-table-column prop="task" label="类型" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="taskType(row.task)" size="small">{{ taskLabel(row.task) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="streamer" label="主播" width="120" show-overflow-tooltip />
        <el-table-column label="详情" min-width="250">
          <template #default="{ row }">
            {{ row.detail || formatDetail(row) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
              {{ row.status === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="日志" width="80" align="center">
          <template #default="{ row }">
            <el-button
              v-if="row.task === 'merge' || row.task === 'clean'"
              type="primary"
              link
              size="small"
              @click="viewLog(row)"
            >
              查看
            </el-button>
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
      <el-tabs v-model="activeLogFile" @tab-change="loadLogContent">
        <el-tab-pane
          v-for="lf in logFiles"
          :key="lf.filename"
          :label="lf.date"
          :name="lf.filename"
        />
      </el-tabs>
      <pre class="log-content" v-loading="logLoading">{{ logContent }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { ElMessage } from "element-plus";
import { Refresh, Download } from "@element-plus/icons-vue";
import { getHistory, exportHistory, getLogList, getLogContent } from "@/api/history";
import type { HistoryRecord, LogFile } from "@/api/types";

const history = ref<HistoryRecord[]>([]);
const loading = ref(false);
const currentPage = ref(1);
const totalItems = ref(0);
const totalPages = ref(0);
const filterTask = ref("");

const logDialogVisible = ref(false);
const logFiles = ref<LogFile[]>([]);
const activeLogFile = ref("");
const logContent = ref("");
const logLoading = ref(false);

async function loadHistory(page: number) {
  loading.value = true;
  try {
    const params: any = { page, per_page: 20 };
    if (filterTask.value) params.task = filterTask.value;
    const res = await getHistory(params);
    history.value = res.items || [];
    totalItems.value = res.total;
    totalPages.value = res.pages;
    currentPage.value = page;
  } finally {
    loading.value = false;
  }
}

function taskType(task: string) {
  if (task === "merge") return "primary" as const;
  if (task === "clean") return "success" as const;
  return "info" as const;
}

function taskLabel(task: string) {
  const map: Record<string, string> = { merge: "合并", clean: "清理", config: "配置" };
  return map[task] || task;
}

function formatDetail(row: HistoryRecord) {
  const parts = [];
  if (row.files_count) parts.push(`${row.files_count}个文件`);
  if (row.merged_bytes) parts.push(`合并${(row.merged_bytes / 1024 ** 3).toFixed(2)}GB`);
  if (row.freed_bytes) parts.push(`释放${(row.freed_bytes / 1024 ** 3).toFixed(2)}GB`);
  if (row.duration) parts.push(`耗时${row.duration}秒`);
  return parts.join(", ") || "-";
}

async function viewLog(row: HistoryRecord) {
  logDialogVisible.value = true;
  logContent.value = "";
  try {
    logFiles.value = await getLogList(row.task);
    if (logFiles.value.length > 0) {
      activeLogFile.value = logFiles.value[0].filename;
      await loadLogContent();
    }
  } catch {
    logFiles.value = [];
  }
}

async function loadLogContent() {
  if (!activeLogFile.value) return;
  logLoading.value = true;
  try {
    const task = logFiles.value[0]?.task || "merge";
    logContent.value = await getLogContent(task, activeLogFile.value);
  } catch {
    logContent.value = "无法加载日志内容";
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

onMounted(() => loadHistory(1));
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}
.header-actions { display: flex; gap: 8px; align-items: center; }
.pagination { margin-top: 16px; display: flex; justify-content: flex-end; }
.log-content {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 16px;
  border-radius: 6px;
  max-height: 60vh;
  overflow-y: auto;
  font-family: "SF Mono", Consolas, monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>

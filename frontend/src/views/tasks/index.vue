<template>
  <div class="tasks">
    <!-- 1. Top Section: 调度状态与全局控制 -->
    <div class="schedule-grid">
      <!-- Card 1: 自动合并 -->
      <div class="card">
        <div class="card-title card-title-flex">
          <div class="title-left">
            <div class="ui-icon-box"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg></div>
            <span class="title-text">自动分片合并</span>
          </div>
          <span class="badge-status" :class="schedule?.merge?.enabled ? 'badge-on' : 'badge-off'">
            <span class="status-dot-sm" :class="schedule?.merge?.enabled ? 'dot-ok' : 'dot-off'" />
            {{ schedule?.merge?.enabled ? '定时计划中' : '已暂停' }}
          </span>
        </div>
        <div class="card-body sched-body">
          <div class="sched-row">
            <span class="sched-label">调度周期</span>
            <span class="mono-val">每 {{ schedule?.merge?.interval || 360 }} <span class="unit-lbl">分钟</span></span>
          </div>
          <div class="sched-row">
            <span class="sched-label">上次运行</span>
            <span class="mono-val-sm">{{ formatDisplayTime(schedule?.merge?.last_run) }}</span>
          </div>
          <div class="sched-row">
            <span class="sched-label">下次预计</span>
            <span class="mono-val-sm">{{ formatDisplayTime(schedule?.merge?.next_run) }}</span>
          </div>
          <div class="sched-action-row">
            <button
              class="btn-task-action btn-merge-action"
              :disabled="sse.isRunning.value"
              @click="handleRunGlobal('merge')"
            >
              <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
              <span>立即全局合并</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Card 2: 自动清理 -->
      <div class="card">
        <div class="card-title card-title-flex">
          <div class="title-left">
            <div class="ui-icon-box"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg></div>
            <span class="title-text">自动磁盘清理</span>
          </div>
          <span class="badge-status" :class="schedule?.clean?.enabled ? 'badge-on' : 'badge-off'">
            <span class="status-dot-sm" :class="schedule?.clean?.enabled ? 'dot-ok' : 'dot-off'" />
            {{ schedule?.clean?.enabled ? '巡检保护中' : '未开启' }}
          </span>
        </div>
        <div class="card-body sched-body">
          <div class="sched-row">
            <span class="sched-label">巡检周期</span>
            <span class="mono-val">每 {{ schedule?.clean?.interval || 720 }} <span class="unit-lbl">分钟</span></span>
          </div>
          <div class="sched-row">
            <span class="sched-label">上次运行</span>
            <span class="mono-val-sm">{{ formatDisplayTime(schedule?.clean?.last_run) }}</span>
          </div>
          <div class="sched-row">
            <span class="sched-label">下次预计</span>
            <span class="mono-val-sm">{{ formatDisplayTime(schedule?.clean?.next_run) }}</span>
          </div>
          <div class="sched-action-row">
            <button
              class="btn-task-action btn-clean-action"
              :disabled="sse.isRunning.value"
              @click="handleRunGlobal('clean')"
            >
              <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg>
              <span>立即全局清理</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 2. Middle Section: 定向主播快捷快跑 -->
    <div class="card targeted-card">
      <div class="targeted-bar">
        <div class="targeted-left">
          <span class="targeted-label">
            <svg class="targeted-svg-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/></svg>
            定向主播快捷任务:
          </span>
          <el-select
            v-model="selectedStreamer"
            placeholder="选择特定主播..."
            filterable
            clearable
            class="targeted-select"
          >
            <el-option
              v-for="s in streamers"
              :key="s.name"
              :label="`${s.name} (${s.files} 个录像 · ${s.size_gb?.toFixed(1)} GB)`"
              :value="s.name"
            />
          </el-select>
          <div class="targeted-buttons">
            <button
              class="btn-targeted btn-targeted-merge"
              :disabled="!selectedStreamer || sse.isRunning.value"
              @click="handleTargetedMerge"
            >
              合并该主播
            </button>
            <button
              class="btn-targeted btn-targeted-clean"
              :disabled="!selectedStreamer || sse.isRunning.value"
              @click="handleTargetedClean"
            >
              清理该主播旧录像
            </button>
          </div>
        </div>
        <div class="targeted-right">
          <router-link
            :to="selectedStreamer ? ('/streamers?streamer=' + encodeURIComponent(selectedStreamer)) : '/streamers'"
            class="link-to-streamers"
          >
            前往主播管理查看切片详情 ↗
          </router-link>
        </div>
      </div>
    </div>

    <!-- 3. Bottom Section: 智能控制台与日志查看 (双模智能终端) -->
    <div class="terminal-container-card">
      <!-- Terminal Tab Switcher -->
      <div class="terminal-ctrl-bar">
        <div class="terminal-tabs">
          <button
            class="term-tab-btn"
            :class="{ active: activeLogTab === 'live' }"
            @click="activeLogTab = 'live'"
          >
            <span class="tab-indicator dot-live" :class="{ 'pulse-dot': sse.isRunning.value }"></span>
            实时终端
            <span v-if="sse.isRunning.value" class="tab-tag-live">运行中</span>
          </button>
          <button
            class="term-tab-btn"
            :class="{ active: activeLogTab === 'recent_merge' }"
            @click="switchRecentTab('merge')"
          >
            <svg class="tab-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg><span>最近合并日志</span>
          </button>
          <button
            class="term-tab-btn"
            :class="{ active: activeLogTab === 'recent_clean' }"
            @click="switchRecentTab('clean')"
          >
            <svg class="tab-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg><span>最近清理日志</span>
          </button>

          <!-- History dropdown selection -->
          <div class="history-select-wrap" v-if="historyOptions.length > 0">
            <el-select
              v-model="selectedHistoryId"
              placeholder="选择查看历史任务归档..."
              size="small"
              class="history-quick-select"
              @change="onHistorySelect"
            >
              <el-option
                v-for="h in historyOptions"
                :key="h.id"
                :label="`[${h.task === 'merge' ? '合并' : '清理'}] ${h.streamer || '全局'} (${formatShortTime(h.time)})`"
                :value="h.log_id || ''"
              />
            </el-select>
          </div>
        </div>

        <div class="terminal-meta-right">
          <span v-if="logLoading" class="meta-status meta-loading">
            ⏳ 读取日志中...
          </span>
          <span v-else-if="activeLogTab === 'live' && sse.isRunning.value" class="meta-status meta-live">
            <span class="status-pulse-dot dot-active"></span> 正在接收实时输出
          </span>
          <span v-else-if="activeLogTab !== 'live' && activeLogMeta" class="meta-status meta-archive" :title="activeLogMeta">
            <svg class="meta-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg> {{ activeLogMeta }}
          </span>
        </div>
      </div>

      <!-- ReTerminal Component -->
      <ReTerminal
        :lines="displayedLines"
        :title="currentTerminalTitle"
        :empty-text="currentEmptyText"
        height="460px"
        @clear="handleClear"
      />

      <!-- Reconnect bar if SSE errored -->
      <div v-if="sse.lastRequest.value && sse.error.value && !sse.isRunning.value && activeLogTab === 'live'" class="retry-bar">
        <button class="ops-action ops-retry" @click="sse.retryLast()">
          <svg class="ops-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
          重新连接
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onActivated } from "vue";
import { useRoute } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { getStreamers } from "@/api/status";
import { getSchedule } from "@/api/schedule";
import { getHistory, getLogContent } from "@/api/history";
import { useSSE, classifyLine, type SSELine } from "@/utils/sse";
import { formatTime } from "@/utils/format";
import ReTerminal from "@/components/ReTerminal/index.vue";
import type { ScheduleStatus, StreamerInfo, HistoryRecord } from "@/api/types";

const route = useRoute();
const schedule = ref<ScheduleStatus>();
const streamers = ref<StreamerInfo[]>([]);
const selectedStreamer = ref("");
const sse = useSSE();

// Log tabs: 'live' | 'recent_merge' | 'recent_clean' | 'history'
const activeLogTab = ref<"live" | "recent_merge" | "recent_clean" | "history">("live");
const historicalLines = ref<SSELine[]>([]);
const historyOptions = ref<HistoryRecord[]>([]);
const selectedHistoryId = ref<string>("");
const logLoading = ref(false);
const activeLogMeta = ref("");

// Displayed lines depending on active tab
const displayedLines = computed(() => {
  if (activeLogTab.value === "live") {
    return sse.lines.value;
  }
  return historicalLines.value;
});

const currentTerminalTitle = computed(() => {
  if (activeLogTab.value === "live") {
    return sse.isRunning.value ? "实时运行终端（流式输出中...）" : "实时运行终端（已完成）";
  }
  if (activeLogTab.value === "recent_merge") {
    return "最近一次合并任务完整日志";
  }
  if (activeLogTab.value === "recent_clean") {
    return "最近一次清理任务完整日志";
  }
  return "历史任务归档日志";
});

const currentEmptyText = computed(() => {
  if (activeLogTab.value === "live") {
    return "等待前台任务触发输出...";
  }
  if (logLoading.value) {
    return "正在从后台载入任务日志...";
  }
  return "该任务暂无关联的详细日志内容";
});

function formatDisplayTime(ts?: number): string {
  if (!ts || ts <= 0) return "未安排 / 待定";
  return formatTime(ts);
}

function formatShortTime(t?: string): string {
  if (!t || t.length < 16) return "";
  return t.slice(5, 16); // "MM-DD HH:mm"
}

// Parse backend op_*.log string into SSELine[]
function parseRawLog(content: string): SSELine[] {
  if (!content) return [];
  const rawLines = content.split(/\r?\n/).filter(l => l.trim().length > 0);
  const result: SSELine[] = [];
  const timestampRegex = /^\[(\d{4}-\d{2}-\d{2}\s+)?(\d{2}:\d{2}:\d{2})\]\s*(.*)$/;

  for (const line of rawLines) {
    const match = line.match(timestampRegex);
    if (match) {
      result.push({
        time: match[2],
        text: match[3],
        type: classifyLine(match[3]),
      });
    } else {
      result.push({
        time: "",
        text: line,
        type: classifyLine(line),
      });
    }
  }
  return result;
}

// Fetch historical log by type or logId
async function fetchAndShowLog(taskType: "merge" | "clean" | "any", specificLogId?: string) {
  logLoading.value = true;
  try {
    let targetRecord: HistoryRecord | undefined;
    if (specificLogId) {
      targetRecord = historyOptions.value.find(h => h.log_id === specificLogId);
    } else {
      const res = await getHistory({
        page: 1,
        per_page: 15,
        task: taskType === "any" ? undefined : taskType
      });
      const items = res.items || [];
      targetRecord = items.find(h => h.log_id && (taskType === "any" ? true : h.task === taskType));
    }

    if (!targetRecord || !targetRecord.log_id) {
      historicalLines.value = [
        {
          time: new Date().toLocaleTimeString("zh-CN"),
          text: `[系统提示] 暂无 ${taskType === "merge" ? "合并" : taskType === "clean" ? "清理" : "后台"} 操作记录`,
          type: "info"
        }
      ];
      activeLogMeta.value = "暂无后台日志";
      return;
    }

    const logRes = await getLogContent(targetRecord.task, targetRecord.log_id);
    historicalLines.value = parseRawLog(logRes.content || "");
    const taskName = targetRecord.task === "merge" ? "自动合并" : "自动清理";
    const statusText = targetRecord.status === "success" ? "成功" : "失败";
    activeLogMeta.value = `[${taskName}] ${targetRecord.streamer || "全局"} · ${targetRecord.time} (${statusText})`;
  } catch (err: any) {
    historicalLines.value = [
      {
        time: new Date().toLocaleTimeString("zh-CN"),
        text: `读取日志失败: ${err?.message || err}`,
        type: "error"
      }
    ];
    activeLogMeta.value = "日志读取失败";
  } finally {
    logLoading.value = false;
  }
}

async function switchRecentTab(task: "merge" | "clean") {
  activeLogTab.value = task === "merge" ? "recent_merge" : "recent_clean";
  await fetchAndShowLog(task);
}

async function onHistorySelect(logId: string) {
  if (!logId) return;
  activeLogTab.value = "history";
  await fetchAndShowLog("any", logId);
}

function handleClear() {
  if (activeLogTab.value === "live") {
    sse.clear();
  } else {
    historicalLines.value = [];
  }
}

// 1. 全局合并
async function handleRunGlobal(task: "merge" | "clean") {
  try {
    if (task === "clean") {
      await ElMessageBox.confirm(
        "即将根据系统配置的阈值执行全局磁盘清理，是否确认？",
        "全局清理确认",
        { type: "warning", confirmButtonText: "确认清理", cancelButtonText: "取消" }
      );
    }
    activeLogTab.value = "live";
    sse.startSSE(task === "merge" ? "/api/merge" : "/api/clean", {});
  } catch {
    // User cancelled
  }
}

// 2. 定向主播合并
async function handleTargetedMerge() {
  if (!selectedStreamer.value) return;
  activeLogTab.value = "live";
  sse.startSSE("/api/merge", { streamer: selectedStreamer.value });
}

// 3. 定向主播清理
async function handleTargetedClean() {
  if (!selectedStreamer.value) return;
  try {
    await ElMessageBox.confirm(
      `即将清理主播「${selectedStreamer.value}」的已合并及过期录像，确认执行？`,
      "清理确认",
      { type: "warning", confirmButtonText: "确认清理", cancelButtonText: "取消" }
    );
    activeLogTab.value = "live";
    sse.startSSE("/api/clean", { streamer: selectedStreamer.value });
  } catch {
    // User cancelled
  }
}

// Refresh schedules and history list
async function loadBaseData() {
  try {
    const [sc, st, hist] = await Promise.allSettled([
      getSchedule(),
      getStreamers(),
      getHistory({ page: 1, per_page: 8 })
    ]);
    if (sc.status === "fulfilled") schedule.value = sc.value;
    if (st.status === "fulfilled") streamers.value = st.value;
    if (hist.status === "fulfilled") {
      historyOptions.value = (hist.value.items || []).filter(h => !!h.log_id);
    }

    const queryStreamer = route.query.streamer as string;
    if (queryStreamer && streamers.value.some(s => s.name === queryStreamer)) {
      selectedStreamer.value = queryStreamer;
    }
  } catch {
    // ignore
  }
}

// Watch SSE running state to refresh when done
watch(() => sse.isRunning.value, async (running, wasRunning) => {
  if (wasRunning && !running) {
    await loadBaseData();
    const lastLine = sse.lines.value.length ? sse.lines.value[sse.lines.value.length - 1] : undefined;
    if (lastLine) {
      if (/✅/.test(lastLine.text)) {
        ElMessage.success(lastLine.text);
      } else if (/❌/.test(lastLine.text)) {
        ElMessage.error(lastLine.text);
      }
    }
  }
});

onMounted(async () => {
  await loadBaseData();
  // If no live task is running, automatically load the most recent task log
  if (!sse.isRunning.value) {
    activeLogTab.value = "recent_merge";
    await fetchAndShowLog("merge");
  }
});

watch(() => route.query.streamer, (val) => {
  if (val && typeof val === "string") {
    selectedStreamer.value = val;
  }
});

onActivated(async () => {
  await loadBaseData();
});
</script>

<style scoped>
.tasks {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* Schedule Grid */
.schedule-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.card {
  background: var(--canvas);
  border: 1px solid var(--hairline);
  border-radius: var(--r-md);
  overflow: hidden;
  box-shadow: none;
}

.card-title {
  padding: 12px 18px;
  font-size: 14px;
  font-weight: 600;
  color: var(--charcoal);
  border-bottom: 1px solid var(--hairline);
}

.card-title-flex {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title-left {
  display: flex;
  align-items: center;
  gap: 6px;
}

.title-icon {
  font-size: 15px;
}

.title-text {
  font-size: 14px;
  font-weight: 600;
  color: var(--ink);
}

.badge-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}
.badge-on {
  background: #edf6f0;
  color: #2b6a48;
  border: 1px solid #d2ecd9;
}
.badge-off {
  background: #f4f4f4;
  color: #888;
  border: 1px solid #e0e0e0;
}

.sched-body {
  padding: 16px 18px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.sched-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
}
.sched-label {
  color: var(--steel);
}
.mono-val {
  font-family: var(--font-mono);
  font-weight: 600;
  color: var(--ink);
}
.mono-val-sm {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--slate);
}
.unit-lbl {
  font-size: 12px;
  font-weight: 400;
  color: var(--stone);
}

.sched-action-row {
  margin-top: 8px;
  padding-top: 12px;
  border-top: 1px dashed var(--hairline);
}

.btn-task-action {
  width: 100%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 8px 14px;
  border-radius: var(--r-sm);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
  border: 1px solid transparent;
}
.btn-task-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-merge-action {
  background: #edf6f0;
  color: #2b6a48;
  border-color: #c7e6d0;
}
.btn-merge-action:not(:disabled):hover {
  background: #d8eedf;
  border-color: #aed9ba;
}

.btn-clean-action {
  background: #fff8f0;
  color: #b35600;
  border-color: #ffd8af;
}
.btn-clean-action:not(:disabled):hover {
  background: #ffeedb;
  border-color: #ffc485;
}

/* Targeted Streamer Operations */
.targeted-card {
  padding: 12px 18px;
}

.targeted-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.targeted-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.targeted-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--charcoal);
  white-space: nowrap;
}

.targeted-select {
  width: 260px;
}

.targeted-buttons {
  display: flex;
  gap: 8px;
}

.btn-targeted {
  width: 140px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  font-size: 12px;
  font-weight: 500;
  border-radius: var(--r-sm);
  cursor: pointer;
  background: var(--surface);
  border: 1px solid var(--hairline);
  color: var(--charcoal);
  transition: all 0.15s ease;
  white-space: nowrap;
}
.btn-targeted:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.btn-targeted-merge:not(:disabled):hover {
  background: #edf6f0;
  color: #2b6a48;
  border-color: #c7e6d0;
}
.btn-targeted-clean:not(:disabled):hover {
  background: #fff8f0;
  color: #b35600;
  border-color: #ffd8af;
}

.targeted-right {
  margin-left: auto;
}

.link-to-streamers {
  font-size: 12px;
  color: var(--steel);
  text-decoration: none;
  padding: 4px 8px;
  border-radius: var(--r-sm);
  transition: all 0.15s;
}
.link-to-streamers:hover {
  color: var(--ink);
  background: var(--highlight);
}

/* Terminal Card */
.terminal-container-card {
  background: #1e1e1e;
  border: 1px solid #333333;
  border-radius: var(--r-md, 8px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}
.terminal-container-card :deep(.re-terminal) {
  border: none !important;
  border-radius: 0 !important;
  background: transparent !important;
}
.terminal-container-card :deep(.terminal-body) {
  border-bottom-left-radius: var(--r-md, 8px);
  border-bottom-right-radius: var(--r-md, 8px);
}

.terminal-ctrl-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #252526;
  border-bottom: 1px solid #333;
  flex-wrap: wrap;
  gap: 8px;
}

.terminal-tabs {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.term-tab-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  width: 124px;
  height: 28px;
  box-sizing: border-box;
  background: #1e1e1e;
  border: 1px solid #3c3c3c;
  color: #cccccc;
  font-size: 12px;
  padding: 0 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
  font-family: var(--font-mono);
  white-space: nowrap;
}
.term-tab-btn:hover {
  background: #2a2d2e;
  color: #ffffff;
  border-color: #555;
}
.term-tab-btn.active {
  background: #37373d;
  color: #ffffff;
  border-color: #007fd4;
  font-weight: 500;
}

.tab-indicator {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
}
.dot-live {
  background: #448361;
}
.pulse-dot {
  background: #d9730d;
  animation: pulse-dot 1.2s infinite;
}
@keyframes pulse-dot {
  0% { transform: scale(0.9); box-shadow: 0 0 0 0 rgba(217, 115, 13, 0.7); }
  70% { transform: scale(1.2); box-shadow: 0 0 0 5px rgba(217, 115, 13, 0); }
  100% { transform: scale(0.9); box-shadow: 0 0 0 0 rgba(217, 115, 13, 0); }
}

.tab-tag-live {
  font-size: 10px;
  background: #d9730d;
  color: #fff;
  padding: 1px 4px;
  border-radius: 2px;
  margin-left: 2px;
}

.history-select-wrap {
  margin-left: 6px;
}
.history-quick-select {
  width: 210px;
}
.history-quick-select :deep(.el-input__wrapper) {
  background: #1e1e1e;
  box-shadow: 0 0 0 1px #3c3c3c inset;
}
.history-quick-select :deep(.el-input__inner) {
  color: #ccc;
  font-size: 12px;
  font-family: var(--font-mono);
}

.terminal-meta-right {
  display: flex;
  align-items: center;
}

.meta-status {
  font-size: 12px;
  font-family: var(--font-mono);
  max-width: 320px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.meta-loading {
  color: #d7ba7d;
}
.meta-live {
  color: #4ec9b0;
  font-weight: 500;
}
.meta-archive {
  color: #9cdcfe;
}

.retry-bar {
  background: #1f1f1f;
  padding: 6px 12px;
  border-top: 1px solid #2e2e2e;
}
.ops-action {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: 1px solid #444;
  color: #ccc;
  font-size: 12px;
  padding: 3px 8px;
  border-radius: 4px;
  cursor: pointer;
}
.ops-action:hover {
  border-color: #666;
  color: #fff;
}
.ops-icon {
  width: 14px;
  height: 14px;
}

@media (max-width: 640px) {
  .targeted-buttons {
    width: 100%;
    display: flex;
    gap: 8px;
  }
  .btn-targeted {
    flex: 1;
    width: 100%;
  }
}

@media (max-width: 900px) {
  .schedule-grid {
    grid-template-columns: 1fr;
  }
  .targeted-bar {
    flex-direction: column;
    align-items: flex-start;
  }
  .targeted-select {
    width: 100%;
  }
  .terminal-ctrl-bar {
    flex-direction: column;
    align-items: flex-start;
  }
  .history-quick-select {
    width: 100%;
  }
  .history-select-wrap {
    margin-left: 0;
    width: 100%;
  }
}

.title-left {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}
.btn-svg-sm {
  width: 14px;
  height: 14px;
  margin-right: 6px;
  flex-shrink: 0;
}
.targeted-svg-ico {
  width: 14px;
  height: 14px;
  margin-right: 4px;
  vertical-align: -2px;
}
.btn-task-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}


.tab-svg-sm {
  width: 13px;
  height: 13px;
  margin-right: 5px;
  vertical-align: -1px;
}
.meta-svg-sm {
  width: 12px;
  height: 12px;
  margin-right: 4px;
  vertical-align: -1px;
}


html.dark {
  .task-status-card {
    background: var(--surface-card) !important;
    border-color: var(--hairline) !important;
  }
  .task-sub-card {
    background: var(--surface) !important;
    border-color: var(--hairline) !important;
  }
  .btn-merge-action {
    background: rgba(34, 197, 94, 0.14) !important;
    color: #4ade80 !important;
    border-color: rgba(34, 197, 94, 0.3) !important;
    &:hover:not(:disabled) {
      background: rgba(34, 197, 94, 0.24) !important;
    }
  }
  .btn-clean-action {
    background: rgba(245, 158, 11, 0.14) !important;
    color: #fbbf24 !important;
    border-color: rgba(245, 158, 11, 0.3) !important;
    &:hover:not(:disabled) {
      background: rgba(245, 158, 11, 0.24) !important;
    }
  }
  .btn-targeted-merge {
    background: #18181b !important;
    color: #f4f4f5 !important;
    border-color: #27272a !important;
    &:hover:not(:disabled) {
      background: rgba(34, 197, 94, 0.14) !important;
      color: #4ade80 !important;
      border-color: rgba(34, 197, 94, 0.3) !important;
    }
  }
  .btn-targeted-clean {
    background: #18181b !important;
    color: #f4f4f5 !important;
    border-color: #27272a !important;
    &:hover:not(:disabled) {
      background: rgba(245, 158, 11, 0.14) !important;
      color: #fbbf24 !important;
      border-color: rgba(245, 158, 11, 0.3) !important;
    }
  }
}

</style>

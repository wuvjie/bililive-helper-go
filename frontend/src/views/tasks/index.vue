<template>
  <div class="tasks">
    <!-- Schedule Status -->
    <div class="schedule-grid">
      <div class="card">
        <div class="card-title">🔄 自动合并</div>
        <div class="card-body sched-body">
          <div class="sched-row">
            <span class="sched-label">状态</span>
            <span class="sched-val"><span class="status-dot-sm" :class="schedule?.merge?.enabled ? 'dot-ok' : 'dot-off'" />{{ schedule?.merge?.enabled ? '已启用' : '已禁用' }}</span>
          </div>
          <div class="sched-row">
            <span class="sched-label">间隔</span>
            <span class="mono-val">{{ schedule?.merge?.interval }} <span class="unit-lbl">分钟</span></span>
          </div>
          <div class="sched-row">
            <span class="sched-label">上次运行</span>
            <span class="mono-val-sm">{{ formatTime(schedule?.merge?.last_run) }}</span>
          </div>
          <div class="sched-row sched-row-last">
            <span class="sched-label">下次运行</span>
            <span class="mono-val-sm">{{ formatTime(schedule?.merge?.next_run) }}</span>
          </div>
        </div>
      </div>
      <div class="card">
        <div class="card-title">🧹 自动清理</div>
        <div class="card-body sched-body">
          <div class="sched-row">
            <span class="sched-label">状态</span>
            <span class="sched-val"><span class="status-dot-sm" :class="schedule?.clean?.enabled ? 'dot-ok' : 'dot-off'" />{{ schedule?.clean?.enabled ? '已启用' : '已禁用' }}</span>
          </div>
          <div class="sched-row">
            <span class="sched-label">间隔</span>
            <span class="mono-val">{{ schedule?.clean?.interval }} <span class="unit-lbl">分钟</span></span>
          </div>
          <div class="sched-row">
            <span class="sched-label">上次运行</span>
            <span class="mono-val-sm">{{ formatTime(schedule?.clean?.last_run) }}</span>
          </div>
          <div class="sched-row sched-row-last">
            <span class="sched-label">下次运行</span>
            <span class="mono-val-sm">{{ formatTime(schedule?.clean?.next_run) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Manual Operations -->
    <div class="card">
      <div class="card-title">🎮 手动操作</div>
      <div class="card-body">
        <div class="ops-bar">
          <el-select
            v-model="selectedStreamer"
            placeholder="选择主播"
            filterable
            size="small"
            style="width: 220px"
            @change="loadFiles"
          >
            <el-option
              v-for="s in streamers"
              :key="s.name"
              :label="`${s.name} (${s.files}个文件, ${s.size_gb?.toFixed(2)}GB)`"
              :value="s.name"
            />
          </el-select>
          <span class="ops-divider"></span>
          <button class="ops-action" :disabled="!selectedStreamer" @click="selectAllUnmerged">
            <svg class="ops-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
            全选未合并
          </button>
          <button
            class="ops-action ops-merge"
            :class="{ 'ops-merge-active': selectedFiles.length >= 2 }"
            :disabled="selectedFiles.length < 2 || sse.isRunning.value"
            @click="handleMerge"
          >
            <svg class="ops-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" /></svg>
            合并
            <span v-if="selectedFiles.length >= 2" class="ops-badge">{{ selectedFiles.length }}</span>
          </button>
          <button
            class="ops-action ops-clean"
            :disabled="!selectedStreamer || sse.isRunning.value"
            @click="handleClean"
          >
            <svg class="ops-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-4v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
            清理
          </button>
        </div>

        <el-table
          v-if="files.length > 0"
          ref="tableRef"
          :data="files"
          size="small"
          max-height="300"
          @selection-change="handleSelectionChange"
          class="mt-12"
        >
          <el-table-column type="selection" width="45" :selectable="() => true" />
          <el-table-column prop="name" label="文件名" min-width="300" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="file-name">{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="size_str" label="大小" width="100" align="right" header-align="right">
            <template #default="{ row }">
              <span class="mono-val">{{ row.size_str }}</span>
            </template>
          </el-table-column>
          <el-table-column label="修改时间" width="170" align="right" header-align="right">
            <template #default="{ row }">
              <span class="mono-val-sm">{{ formatTime(row.mtime) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100" align="right" header-align="right">
            <template #default="{ row }">
              <span class="status-dot-text" :class="row.is_merged ? 'dot-merged' : 'dot-pending'">
                {{ row.is_merged ? '已合并' : '未合并' }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <!-- Terminal Output -->
    <ReTerminal
      :lines="sse.lines.value"
      height="400px"
      @clear="sse.clear"
    />
    <div v-if="sse.lastRequest.value && sse.error.value && !sse.isRunning.value" class="retry-bar">
      <button class="ops-action ops-retry" @click="sse.retryLast()">
        <svg class="ops-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
        重新连接
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onUnmounted, onActivated } from "vue";
import { useRoute } from "vue-router";
import { ElMessage, ElMessageBox, type TableInstance } from "element-plus";
import { getStreamers, getStreamerFiles } from "@/api/status";
import { getSchedule } from "@/api/schedule";
import { getCleanEstimate } from "@/api/task";
import { useSSE } from "@/utils/sse";
import { formatTime } from "@/utils/format";
import ReTerminal from "@/components/ReTerminal/index.vue";
import type { ScheduleStatus, StreamerInfo, StreamerFile } from "@/api/types";

const route = useRoute();
const schedule = ref<ScheduleStatus>();
const streamers = ref<StreamerInfo[]>([]);
const selectedStreamer = ref("");
const files = ref<StreamerFile[]>([]);
const selectedFiles = ref<StreamerFile[]>([]);
const tableRef = ref<TableInstance>();
const sse = useSSE();

async function loadSchedule() {
  try {
    schedule.value = await getSchedule();
  } catch {
    // silently ignore
  }
}

async function loadFiles() {
  if (!selectedStreamer.value) {
    files.value = [];
    return;
  }
  try {
    files.value = await getStreamerFiles(selectedStreamer.value);
  } catch {
    files.value = [];
    ElMessage.error("加载文件列表失败");
  }
}

function handleSelectionChange(rows: StreamerFile[]) {
  selectedFiles.value = rows;
}

function selectAllUnmerged() {
  if (!tableRef.value) return;
  tableRef.value.clearSelection();
  nextTick(() => {
    let count = 0;
    files.value.forEach(row => {
      if (!row.is_merged) {
        tableRef.value!.toggleRowSelection(row, true);
        count++;
      }
    });
    if (count === 0) {
      ElMessage.info("没有未合并的文件");
    }
  });
}

async function handleMerge() {
  if (selectedFiles.value.length < 2) {
    ElMessage.warning("至少选择2个文件");
    return;
  }
  try {
    await ElMessageBox.confirm(
      `即将合并 ${selectedFiles.value.length} 个文件，是否继续？`,
      "确认合并",
      { type: "warning", confirmButtonText: "确认合并", cancelButtonText: "取消" }
    );
    sse.startSSE("/api/merge/manual", {
      streamer: selectedStreamer.value,
      files: selectedFiles.value.map(f => f.name)
    });
  } catch {
    // User cancelled
  }
}

async function handleClean() {
  try {
    if (selectedStreamer.value) {
      await ElMessageBox.confirm(
        `即将清理主播「${selectedStreamer.value}」的已合并文件，确认执行？`,
        "清理确认",
        { type: "warning", confirmButtonText: "确认清理", cancelButtonText: "取消" }
      );
    } else {
      const estimate = await getCleanEstimate();
      await ElMessageBox.confirm(
        `预计将清理 ${estimate.file_count} 个文件，释放 ${estimate.total_size_gb?.toFixed(2)} GB 空间。确认执行？`,
        "清理确认",
        { type: "warning", confirmButtonText: "确认清理", cancelButtonText: "取消" }
      );
    }
    sse.startSSE("/api/clean", { streamer: selectedStreamer.value });
  } catch {
    // User cancelled
  }
}

// When SSE finishes (isRunning goes from true -> false), refresh data and show notification
watch(() => sse.isRunning.value, async (running, wasRunning) => {
  if (wasRunning && !running) {
    await Promise.allSettled([loadFiles(), loadSchedule()]);
    const lastLine = sse.lines.value.at(-1);
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
  const [s, st] = await Promise.allSettled([getSchedule(), getStreamers()]);
  if (s.status === "fulfilled") schedule.value = s.value;
  if (st.status === "fulfilled") streamers.value = st.value;

  const queryStreamer = route.query.streamer as string;
  if (queryStreamer && streamers.value.some(s => s.name === queryStreamer)) {
    selectedStreamer.value = queryStreamer;
    await loadFiles();
  }
});

onUnmounted(() => {
  sse.abort();
});

// Refresh data when component is re-activated by keep-alive router-view
onActivated(async () => {
  const [s, st] = await Promise.allSettled([getSchedule(), getStreamers()]);
  if (s.status === "fulfilled") schedule.value = s.value;
  if (st.status === "fulfilled") streamers.value = st.value;
});
</script>

<style scoped>
.tasks { display: flex; flex-direction: column; gap: 16px; }
.mt-12 { margin-top: 12px; }

/* Schedule grid */
.schedule-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.card {
  background: var(--canvas); border: 1px solid var(--hairline);
  border-radius: var(--r-md); overflow: hidden;
  box-shadow: none;
}
.card-title {
  padding: 14px 20px; font-size: 14px; font-weight: 600; color: var(--ink);
  border-bottom: 1px solid var(--hairline);
}
.card-body { padding: 20px; }

/* Schedule rows */
.sched-body { display: flex; flex-direction: column; }
.sched-row {
  display: flex; justify-content: space-between; align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #f1f1ef;
  font-size: 13px;
}
.sched-row-last { border-bottom: none; }
.sched-label { color: var(--steel); min-width: 64px; }
.sched-val {
  display: flex; align-items: center; gap: 8px;
  color: var(--steel); font-weight: 500;
  width: 160px; text-align: left;
}
.mono-val { font-family: var(--font-mono); font-size: 13px; color: var(--ink); width: 160px; text-align: left; }
.mono-val-sm { font-family: var(--font-mono); font-size: 12px; color: var(--slate); width: 160px; text-align: left; }
.unit-lbl { font-size: 12px; color: var(--stone); margin-left: 4px; }
.file-name { font-family: var(--font-mono); font-size: 12px; color: var(--charcoal); }

/* Status dots */
.status-dot-sm { width: 6px; height: 6px; border-radius: 50%; display: inline-block; flex-shrink: 0; }
.dot-ok { background: #448361; }
.dot-off { background: #d3d1cb; }

/* Ops bar */
.ops-bar {
  display: flex; gap: 4px; flex-wrap: wrap; align-items: center;
}
.ops-divider { width: 1px; height: 14px; background: var(--hairline); flex-shrink: 0; margin: 0 4px; }
.ops-icon { width: 14px; height: 14px; flex-shrink: 0; }
/* Ghost text action buttons — no border, no background */
.ops-action {
  display: inline-flex; align-items: center; gap: 4px;
  background: transparent; border: none; padding: 0;
  font-size: 13px; font-weight: 500; color: var(--steel);
  cursor: pointer; transition: color 0.15s;
}
.ops-action:hover:not(:disabled) { color: var(--ink); }
.ops-action:disabled { color: #c7c7cc; cursor: not-allowed; }
/* Merge button — accent when files selected */
.ops-merge { color: #c7c7cc; }
.ops-merge-active { color: var(--ink) !important; font-weight: 600 !important; }
.ops-badge {
  display: inline-block; padding: 0 5px; margin-left: 4px;
  font-size: 11px; font-weight: 600; font-family: var(--font-mono);
  color: var(--ink); background: var(--highlight);
  border-radius: var(--r-xs); line-height: 1.6;
}
/* Clean button — hover turns red */
.ops-clean:hover:not(:disabled) { color: #c4554d !important; }

/* Retry bar — shown below terminal when SSE fails */
.retry-bar {
  display: flex; justify-content: flex-end; padding: 8px 0;
}
.ops-retry {
  color: #d9730d;
}
.ops-retry:hover:not(:disabled) { color: #b85c00 !important; }

/* Status dot text in table */
.status-dot-text {
  display: inline-flex; align-items: center; gap: 5px;
  font-size: 12px; font-weight: 500; background: transparent;
}
.status-dot-text::before {
  content: ''; display: inline-block; width: 6px; height: 6px; border-radius: 50%; margin-right: 6px;
}
.dot-merged { color: #2b593f; }
.dot-merged::before { background: #448361; }
.dot-pending { color: #9a6c00; }
.dot-pending::before { background: #d9730d; }

@media (max-width: 768px) {
  .schedule-grid { grid-template-columns: 1fr; }
}
</style>

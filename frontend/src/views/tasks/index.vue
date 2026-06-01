<template>
  <div class="tasks">
    <!-- Schedule Status -->
    <el-row :gutter="16" class="mb-16">
      <el-col :xs="24" :sm="12">
        <el-card shadow="hover">
          <template #header><span>🔄 自动合并</span></template>
          <el-descriptions :column="2" size="small" border>
            <el-descriptions-item label="状态">
              <el-tag :type="schedule?.merge?.enabled ? 'success' : 'info'" size="small">
                {{ schedule?.merge?.enabled ? '已启用' : '已禁用' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="间隔">{{ schedule?.merge?.interval }} 分钟</el-descriptions-item>
            <el-descriptions-item label="上次运行">{{ formatTime(schedule?.merge?.last_run) }}</el-descriptions-item>
            <el-descriptions-item label="下次运行">{{ formatTime(schedule?.merge?.next_run) }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12">
        <el-card shadow="hover">
          <template #header><span>🧹 自动清理</span></template>
          <el-descriptions :column="2" size="small" border>
            <el-descriptions-item label="状态">
              <el-tag :type="schedule?.clean?.enabled ? 'success' : 'info'" size="small">
                {{ schedule?.clean?.enabled ? '已启用' : '已禁用' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="间隔">{{ schedule?.clean?.interval }} 分钟</el-descriptions-item>
            <el-descriptions-item label="上次运行">{{ formatTime(schedule?.clean?.last_run) }}</el-descriptions-item>
            <el-descriptions-item label="下次运行">{{ formatTime(schedule?.clean?.next_run) }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>

    <!-- Manual Operations -->
    <el-card shadow="hover" class="mb-16">
      <template #header><span>🎮 手动操作</span></template>

      <div class="ops-bar">
        <el-select
          v-model="selectedStreamer"
          placeholder="选择主播"
          filterable
          style="width: 240px"
          @change="loadFiles"
        >
          <el-option
            v-for="s in streamers"
            :key="s.name"
            :label="`${s.name} (${s.files}个文件, ${s.size_gb?.toFixed(2)}GB)`"
            :value="s.name"
          />
        </el-select>
        <el-button @click="selectAllUnmerged" :disabled="!selectedStreamer">全选未合并</el-button>
        <el-button
          type="primary"
          :disabled="selectedFiles.length < 2 || sse.isRunning.value"
          @click="handleMerge"
        >
          合并 ({{ selectedFiles.length }} 个文件)
        </el-button>
        <el-button
          type="warning"
          :disabled="!selectedStreamer || sse.isRunning.value"
          @click="handleClean"
        >
          清理
        </el-button>
      </div>

      <el-table
        v-if="files.length > 0"
        :data="files"
        size="small"
        stripe
        max-height="300"
        @selection-change="handleSelectionChange"
        class="mt-12"
      >
        <el-table-column type="selection" width="45" :selectable="() => true" />
        <el-table-column prop="name" label="文件名" min-width="300" show-overflow-tooltip />
        <el-table-column prop="size_str" label="大小" width="100" />
        <el-table-column prop="mtime" label="修改时间" width="170">
          <template #default="{ row }">
            {{ formatMtime(row.mtime) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_merged ? 'success' : 'warning'" size="small">
              {{ row.is_merged ? '已合并' : '未合并' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Terminal Output -->
    <ReTerminal
      :lines="sse.lines.value"
      height="400px"
      @clear="sse.clear"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { getStreamers, getStreamerFiles } from "@/api/status";
import { getSchedule } from "@/api/schedule";
import { runMerge, runClean, getCleanEstimate } from "@/api/task";
import { useSSE } from "@/utils/sse";
import ReTerminal from "@/components/ReTerminal/index.vue";
import type { ScheduleStatus, StreamerInfo, StreamerFile } from "@/api/types";

const route = useRoute();
const schedule = ref<ScheduleStatus>();
const streamers = ref<StreamerInfo[]>([]);
const selectedStreamer = ref("");
const files = ref<StreamerFile[]>([]);
const selectedFiles = ref<StreamerFile[]>([]);
const sse = useSSE();

function formatTime(ts?: number) {
  if (!ts) return "-";
  return new Date(ts * 1000).toLocaleString("zh-CN");
}

function formatMtime(mtime: number | string) {
  if (!mtime) return "-";
  const ts = typeof mtime === "number" ? mtime : parseInt(mtime);
  if (isNaN(ts)) return String(mtime);
  // If it looks like seconds (10 digits), convert to ms
  const ms = ts > 1e12 ? ts : ts * 1000;
  return new Date(ms).toLocaleString("zh-CN");
}

async function loadFiles() {
  if (!selectedStreamer.value) {
    files.value = [];
    return;
  }
  files.value = await getStreamerFiles(selectedStreamer.value);
}

function handleSelectionChange(rows: StreamerFile[]) {
  selectedFiles.value = rows;
}

function selectAllUnmerged() {
  // Triggered via table ref - simplified: use the table's toggleAllSelection
  // For now, filter the unmerged files and we'll set them
  const table = document.querySelector(".el-table") as any;
  // Simple workaround: just select all via reactivity
}

async function handleMerge() {
  if (selectedFiles.value.length < 2) {
    ElMessage.warning("至少选择2个文件");
    return;
  }
  sse.startSSE("/api/merge/manual", {
    streamer: selectedStreamer.value,
    files: selectedFiles.value.map(f => f.name)
  });
}

async function handleClean() {
  try {
    const estimate = await getCleanEstimate();
    await ElMessageBox.confirm(
      `预计将清理 ${estimate.file_count} 个文件，释放 ${estimate.total_size_gb?.toFixed(2)} GB 空间。确认执行？`,
      "清理确认",
      { type: "warning", confirmButtonText: "确认清理", cancelButtonText: "取消" }
    );
    sse.startSSE("/api/clean", { streamer: selectedStreamer.value });
  } catch {
    // User cancelled
  }
}

onMounted(async () => {
  const [s, st] = await Promise.allSettled([getSchedule(), getStreamers()]);
  if (s.status === "fulfilled") schedule.value = s.value;
  if (st.status === "fulfilled") streamers.value = st.value;

  // Pre-select streamer from query param
  const queryStreamer = route.query.streamer as string;
  if (queryStreamer && streamers.value.some(s => s.name === queryStreamer)) {
    selectedStreamer.value = queryStreamer;
    await loadFiles();
  }
});
</script>

<style scoped>
.tasks { display: flex; flex-direction: column; gap: 16px; }
.mb-16 { margin-bottom: 0; }
.mt-12 { margin-top: 12px; }
.ops-bar {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}
</style>

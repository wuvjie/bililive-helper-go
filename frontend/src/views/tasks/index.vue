<template>
  <div class="tasks">
    <!-- 1. Top Section: 两个系统全局自动调度核心监控卡片 (并排 2 列) -->
    <ScheduleCards :schedule="schedule" />

    <!-- 2. Integrated Execution Workbench (手动执行录像合并与清理 + 实时终端一体化工作台) -->
    <WorkbenchConsole
      v-model:selected-streamer="selectedStreamer"
      :streamers="streamers"
      :is-running="sse.isRunning.value"
      :has-error="Boolean(sse.lastRequest.value && sse.error.value)"
      :lines="sse.lines.value"
      :title="currentTerminalTitle"
      :empty-text="currentEmptyText"
      @execute-merge="handleExecuteMerge"
      @execute-clean="handleExecuteClean"
      @clear="handleClear"
      @retry="sse.retryLast()"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onActivated, onDeactivated } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { getStreamers } from "@/api/status";
import { getSchedule } from "@/api/schedule";
import { useSSE } from "@/utils/sse";
import type { ScheduleStatus, StreamerInfo } from "@/api/types";

import ScheduleCards from "./components/ScheduleCards.vue";
import WorkbenchConsole from "./components/WorkbenchConsole.vue";

const route = useRoute();
const router = useRouter();
const schedule = ref<ScheduleStatus>();
const streamers = ref<StreamerInfo[]>([]);
const selectedStreamer = ref("");

let autoTriggered = false;

// 离开页面时安全重置未执行的临时定向主播选择
onDeactivated(() => {
  selectedStreamer.value = "";
  autoTriggered = false;
});

const sse = useSSE();

const currentTerminalTitle = computed(() => {
  return sse.isRunning.value ? "实时运行终端（流式输出中...）" : "实时运行终端（已完成）";
});

const currentEmptyText = computed(() => {
  return "等待手动触发或定时任务调度输出...";
});

function handleClear() {
  sse.clear();
}

// 1. 执行录像合并（根据 selectedStreamer 自动区分全局 vs 定向）
async function handleExecuteMerge() {
  if (sse.isRunning.value) return;
  const streamer = selectedStreamer.value?.trim();
  if (streamer) {
    sse.startSSE("/api/merge", { streamer });
  } else {
    sse.startSSE("/api/merge", {});
  }
}

// 2. 执行录像清理（根据 selectedStreamer 自动区分全局 vs 定向，带明确安全确认提示）
async function handleExecuteClean() {
  if (sse.isRunning.value) return;
  const streamer = selectedStreamer.value?.trim();
  try {
    if (streamer) {
      await ElMessageBox.confirm(
        `即将清理主播「${streamer}」的已合并及过期录像，是否确认立即执行？`,
        "清理确认",
        {
          type: "warning",
          confirmButtonText: "确认清理",
          cancelButtonText: "取消"
        }
      );
      sse.startSSE("/api/clean", { streamer });
    } else {
      await ElMessageBox.confirm(
        "即将根据系统配置的清理阈值与保留规则，对所有主播的录像目录执行磁盘清理，是否确认立即执行？",
        "全局清理确认",
        {
          type: "warning",
          confirmButtonText: "确认清理",
          cancelButtonText: "取消"
        }
      );
      sse.startSSE("/api/clean", {});
    }
  } catch {
    // User cancelled
  }
}

async function loadBaseData() {
  try {
    const [sc, st] = await Promise.allSettled([
      getSchedule(),
      getStreamers()
    ]);
    if (sc.status === "fulfilled") schedule.value = sc.value;
    if (st.status === "fulfilled") streamers.value = st.value;

    const queryStreamer = route.query.streamer as string;
    if (queryStreamer && streamers.value.some(s => s.name === queryStreamer)) {
      selectedStreamer.value = queryStreamer;
      if (route.query.auto === "merge" && !autoTriggered && !sse.isRunning.value) {
        autoTriggered = true;
        // 关键防递归修复：立即清除 URL 中的 auto 标识，防止后续 watch(sse.isRunning) 循环递归触发
        router.replace({ query: { ...route.query, auto: undefined } });
        setTimeout(() => {
          handleExecuteMerge();
        }, 100);
      }
    }
  } catch (e: any) {
    ElMessage.error(e?.message || "获取调度数据失败");
  }
}

onMounted(loadBaseData);
onActivated(loadBaseData);

// 任务完成时刷新基础数据
watch(sse.isRunning, (running, prev) => {
  if (prev && !running) {
    loadBaseData();
  }
});
</script>

<style scoped>
.tasks {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
</style>

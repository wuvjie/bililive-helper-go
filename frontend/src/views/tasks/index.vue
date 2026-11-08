<template>
  <div class="tasks">
    <!-- 1. Top Section: 两个系统全局自动调度核心卡片 (并排 2 列) -->
    <ScheduleCards
      :schedule="schedule"
      :is-running="sse.isRunning.value"
      @run-global="handleRunGlobal"
    />

    <!-- 2. Integrated Execution Workbench (定向主播维护与实时终端一体化融合控制台) -->
    <WorkbenchConsole
      v-model:selected-streamer="selectedStreamer"
      :streamers="streamers"
      :is-running="sse.isRunning.value"
      :has-error="Boolean(sse.lastRequest.value && sse.error.value)"
      :lines="sse.lines.value"
      :title="currentTerminalTitle"
      :empty-text="currentEmptyText"
      @targeted-merge="handleTargetedMerge"
      @targeted-clean="handleTargetedClean"
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
  return "等待前台任务触发输出...";
});

function handleClear() {
  sse.clear();
}

// 1. 全局合并 / 清理
async function handleRunGlobal(task: "merge" | "clean") {
  try {
    if (task === "clean") {
      await ElMessageBox.confirm(
        "即将根据系统配置的阈值执行全局磁盘清理，是否确认？",
        "全局清理确认",
        { type: "warning", confirmButtonText: "确认清理", cancelButtonText: "取消" }
      );
    }
    sse.startSSE(task === "merge" ? "/api/merge" : "/api/clean", {});
  } catch {
    // User cancelled
  }
}

// 2. 定向主播合并
async function handleTargetedMerge() {
  if (!selectedStreamer.value) return;
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
    sse.startSSE("/api/clean", { streamer: selectedStreamer.value });
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
          handleTargetedMerge();
        }, 100);
      }
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

watch(() => route.query.streamer, (val, oldVal) => {
  if (val && typeof val === "string") {
    selectedStreamer.value = val;
    if (val !== oldVal) {
      autoTriggered = false;
    }
  }
});

onMounted(async () => {
  await loadBaseData();
});

onActivated(async () => {
  await loadBaseData();
});
</script>

<style scoped>
.tasks {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
</style>

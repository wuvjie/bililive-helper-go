<template>
  <div class="streamers">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <div class="card-title-group">
              <div class="ui-icon-box icon-box-orange">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                  <circle cx="12" cy="7" r="4"/>
                </svg>
              </div>
              <span class="title-text">主播管理</span>
            </div>
            <div class="streamer-meta-stat">
              <span class="meta-item">
                <strong class="meta-num">{{ streamers.length }}</strong>
                <span class="meta-unit">位主播</span>
              </span>
              <span class="meta-dot">·</span>
              <span class="meta-item">
                <strong class="meta-num">{{ totalFiles }}</strong>
                <span class="meta-unit">个录像</span>
              </span>
              <span class="meta-dot">·</span>
              <span class="meta-item">
                <strong class="meta-num">{{ totalSizeGB }}</strong>
                <span class="meta-unit">GB</span>
              </span>
            </div>
          </div>
          <div class="header-right">
            <div class="search-input-wrap">
              <el-input
                ref="searchInputRef"
                v-model="searchQuery"
                placeholder="搜索主播名..."
                :prefix-icon="Search"
                clearable
                style="width: 230px"
              />
              <span v-if="!searchQuery" class="kbd-badge search-kbd">/</span>
            </div>
          </div>
        </div>
      </template>

      <div class="table-responsive-wrapper">
        <StreamerTable
          :streamers="filteredStreamers"
          :loading="loading"
          :search-query="searchQuery"
          :tick="tick"
          @open-drawer="openDrawer"
          @open-rule="openRuleDialog"
          @clear-search="searchQuery = ''"
          @refresh="loadData"
        />
      </div>
    </el-card>

    <!-- 主播规则配置弹窗 -->
    <StreamerRuleDialog
      v-model:visible="ruleDialogVisible"
      :streamer="currentStreamer"
      :rule="selectedRule"
      :saving="savingRule"
      @save="handleSaveRule"
    />

    <!-- 录像文件抽屉 -->
    <StreamerDrawer
      v-model:visible="drawerVisible"
      :streamer="drawerStreamer"
      :files="drawerFiles"
      :loading="drawerLoading"
      :tick="tick"
      @play="handlePlay"
      @delete="handleDeleteFile"
      @merge-streamer="handleMergeStreamer"
    />

    <!-- 在线视频播放预览弹窗 -->
    <VideoPlayerModal
      v-model:visible="playerVisible"
      :file-name="playFileName"
      :url="playUrl"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, onActivated } from "vue";
import { useRouter } from "vue-router";
import { getStreamers, getStreamerFiles, saveStreamerRule, getVideoPlayUrl, deleteStreamerFile } from "@/api/status";
import { Search } from "@element-plus/icons-vue";
import { ElMessage, ElMessageBox } from "element-plus";
import type { StreamerInfo, StreamerFile, StreamerRule } from "@/api/types";

import StreamerTable from "./components/StreamerTable.vue";
import StreamerDrawer from "./components/StreamerDrawer.vue";
import StreamerRuleDialog from "./components/StreamerRuleDialog.vue";
import VideoPlayerModal from "./components/VideoPlayerModal.vue";

const router = useRouter();
const streamers = ref<StreamerInfo[]>([]);
const loading = ref(true);
const searchInputRef = ref();
const searchQuery = ref("");

function handleKeydown(e: KeyboardEvent) {
  if (e.key === "/" && document.activeElement?.tagName !== "INPUT" && document.activeElement?.tagName !== "TEXTAREA") {
    e.preventDefault();
    searchInputRef.value?.focus();
  }
}

const totalFiles = computed(() => streamers.value.reduce((acc, s) => acc + (s.files || 0), 0));
const totalSizeGB = computed(() => (streamers.value.reduce((acc, s) => acc + (s.size_gb || 0), 0)).toFixed(1));

const filteredStreamers = computed(() => {
  if (!searchQuery.value) return streamers.value;
  const q = searchQuery.value.toLowerCase();
  return streamers.value.filter(s => s.name.toLowerCase().includes(q));
});

// 规则配置
const ruleDialogVisible = ref(false);
const currentStreamer = ref("");
const selectedRule = ref<StreamerRule>();
const savingRule = ref(false);

function openRuleDialog(row: StreamerInfo) {
  currentStreamer.value = row.name;
  selectedRule.value = row.rule;
  ruleDialogVisible.value = true;
}

async function handleSaveRule(rule: StreamerRule) {
  savingRule.value = true;
  try {
    await saveStreamerRule(currentStreamer.value, rule);
    ElMessage.success("主播个性化规则已保存");
    ruleDialogVisible.value = false;
    const target = streamers.value.find(s => s.name === currentStreamer.value);
    if (target) {
      target.rule = { ...rule };
      target.has_rule = true;
    }
  } catch (err: any) {
    ElMessage.error(err.message || "保存规则失败");
  } finally {
    savingRule.value = false;
  }
}

// 抽屉与文件
const drawerVisible = ref(false);
const drawerLoading = ref(false);
const drawerStreamer = ref("");
const drawerFiles = ref<StreamerFile[]>([]);

async function openDrawer(row: StreamerInfo) {
  drawerStreamer.value = row.name;
  drawerVisible.value = true;
  drawerLoading.value = true;
  try {
    const files = await getStreamerFiles(row.name);
    drawerFiles.value = (files || []).slice().sort((a, b) => (b.mtime || 0) - (a.mtime || 0));
  } catch {
    ElMessage.error("获取文件列表失败");
  } finally {
    drawerLoading.value = false;
  }
}

function handleMergeStreamer(streamerName: string) {
  drawerVisible.value = false;
  router.push({ path: "/tasks", query: { streamer: streamerName, auto: "merge" } });
}

async function handleDeleteFile(row: StreamerFile) {
  try {
    await ElMessageBox.confirm(
      `确定要永久删除录像文件 "${row.name}" 吗？此操作无法撤销。`,
      "确认删除",
      {
        confirmButtonText: "确定删除",
        cancelButtonText: "取消",
        type: "warning"
      }
    );
    await deleteStreamerFile(drawerStreamer.value, row.name);
    ElMessage.success("录像文件已删除");
    drawerFiles.value = drawerFiles.value.filter(f => f.name !== row.name);
    loadData();
  } catch (err: any) {
    if (err !== "cancel") {
      ElMessage.error(err.message || "删除文件失败");
    }
  }
}

// 播放预览
const playerVisible = ref(false);
const playFileName = ref("");
const playUrl = ref("");

function handlePlay(file: StreamerFile) {
  if (/\.(flv|ts)$/i.test(file.name)) {
    ElMessage.info("该录像是 FLV/TS 原始流媒体切片，浏览器暂不支持原生解码，建议通过合并任务转为 MP4 成片后预览");
    return;
  }
  playFileName.value = file.name;
  playUrl.value = getVideoPlayUrl(drawerStreamer.value, file.name);
  playerVisible.value = true;
}

// tick 相对时间驱动
const tick = ref(0);
let tickTimer: ReturnType<typeof setInterval> | null = null;

async function loadData() {
  loading.value = true;
  try {
    const list = await getStreamers();
    streamers.value = list || [];
  } catch {
    ElMessage.error("获取主播列表失败");
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  loadData();
  window.addEventListener("keydown", handleKeydown);
  tickTimer = setInterval(() => { tick.value++; }, 60_000);
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleKeydown);
  if (tickTimer) clearInterval(tickTimer);
});

onActivated(() => {
  loadData();
});
</script>

<style scoped>
.streamers {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
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

.streamer-meta-stat {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--slate);
  background: var(--surface-secondary, rgba(0, 0, 0, 0.03));
  padding: 3px 10px;
  border-radius: var(--r-full);
}

html.dark .streamer-meta-stat {
  background: rgba(255, 255, 255, 0.05);
}

.meta-num {
  font-family: var(--font-mono);
  font-weight: 600;
  color: var(--charcoal);
  margin-right: 2px;
}

.meta-dot {
  color: var(--hairline-strong);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.search-input-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.search-kbd {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
}

.table-responsive-wrapper {
  width: 100%;
  overflow-x: auto;
}
</style>

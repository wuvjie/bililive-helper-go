<template>
  <div class="streamers">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span class="card-title">👤 主播录播管理</span>
            <span class="streamer-total-stat">
              共 <strong>{{ streamers.length }}</strong> 位主播 · <strong>{{ totalFiles }}</strong> 个录像 · <strong>{{ totalSizeGB }}</strong> GB
            </span>
          </div>
          <div class="header-right">
            <el-input
              v-model="searchQuery"
              placeholder="搜索主播名..."
              :prefix-icon="Search"
              clearable
              size="small"
              style="width: 200px"
            />
            <button class="header-refresh-btn" @click="loadData" title="刷新列表">
              <svg class="refresh-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
              刷新
            </button>
          </div>
        </div>
      </template>

      <el-table
        :data="filteredStreamers"
        v-loading="loading"
        style="width: 100%"
        class="streamer-table"
        :default-sort="{ prop: 'size_gb', order: 'descending' }"
      >
        <el-table-column prop="name" label="主播" min-width="260" show-overflow-tooltip sortable>
          <template #default="{ row }">
            <div class="streamer-name-cell">
              <span class="streamer-name">{{ row.name }}</span>
              <span v-if="row.rule?.protected" class="badge-pill badge-protected">🔒 免删保护</span>
              <span v-if="row.rule?.trim_enabled" class="badge-pill badge-trim">
                ✂ {{ row.rule.trim_start || '00:00:00' }} ~ {{ row.rule.trim_end || '结束' }}
              </span>
              <span v-if="row.rule?.min_keep > 0" class="badge-pill badge-minkeep">
                留≥{{ row.rule.min_keep }}场
              </span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="files" label="录制文件数" width="130" align="right" header-align="right" sortable :sort-orders="['descending', 'ascending']">
          <template #header>
            <el-tooltip content="该主播名下的视频总数（含未合并分段与已合并成品）" placement="top">
              <span class="col-header-tip">录制文件数 ⓘ</span>
            </el-tooltip>
          </template>
          <template #default="{ row }">
            <div class="data-block">
              <span class="mono-val">{{ row.files }}</span><span class="unit-lbl">个</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="size_gb" label="磁盘占用" width="140" align="right" header-align="right" sortable :sort-orders="['descending', 'ascending']">
          <template #default="{ row }">
            <div class="data-block">
              <span class="mono-val">{{ row.size_gb?.toFixed(2) }}</span><span class="unit-lbl">GB</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column
          prop="mtime"
          label="最新视频"
          width="130"
          align="right"
          header-align="right"
          sortable
          :sort-orders="['descending', 'ascending']"
          :sort-method="sortMtime"
        >
          <template #default="{ row }">
            <el-tooltip v-if="row.mtime" :content="formatExactTime(row.mtime)" placement="top">
              <span class="mono-val cursor-help">{{ formatRelativeTime(row.mtime, tick) }}</span>
            </el-tooltip>
            <span v-else class="mono-val">-</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="200" align="right" header-align="right">
          <template #default="{ row }">
            <div class="action-btn-group">
              <button class="ghost-action" @click="openDrawer(row)">📂 录像 ({{ row.files }})</button>
              <span class="action-sep">|</span>
              <button class="ghost-action" @click="openRuleDialog(row)">⚙️ 规则</button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 主播规则配置弹窗 -->
    <el-dialog
      v-model="ruleDialogVisible"
      :title="`⚙️ ${currentStreamer} 的录像规则`"
      width="520px"
      destroy-on-close
      class="rule-dialog"
    >
      <el-form label-position="top" class="rule-form">
        <!-- 视频时段裁剪设置 -->
        <div class="rule-section">
          <div class="section-header">
            <div class="title-with-badge">
              <span class="section-title">✂ 视频时段裁剪</span>
              <span class="section-sub">无损流复制，零画质损失</span>
            </div>
            <el-switch v-model="formRule.trim_enabled" />
          </div>
          <p class="section-desc">
            启用后，每次直播合并转换完成后，系统将自动按指定起止时间截取（如去除开播暖场或只保留前半部分）。
          </p>

          <template v-if="formRule.trim_enabled">
            <div class="time-range-group">
              <div class="time-box">
                <span class="time-label">开始时间</span>
                <el-input v-model="formRule.trim_start" placeholder="00:00:00" class="mono-input" />
              </div>
              <span class="range-arrow">至</span>
              <div class="time-box">
                <span class="time-label">结束时间</span>
                <el-input v-model="formRule.trim_end" placeholder="00:30:00" class="mono-input" />
              </div>
              <button type="button" class="quick-pill-btn" @click="fillQuick30">⚡ 填入前30分钟</button>
            </div>
            <div class="form-tip">
              * 结束时间留空表示截取到最后；如需保留前30分钟，设置为 <code>00:00:00</code> 至 <code>00:30:00</code> 即可。
            </div>
          </template>
        </div>

        <el-divider style="margin: 18px 0" />

        <!-- 存储与清理偏好 -->
        <div class="rule-section">
          <div class="section-header">
            <span class="section-title">🛡️ 存储与清理保护</span>
          </div>

          <el-form-item label="免清理免删保护">
            <div class="switch-row">
              <el-switch v-model="formRule.protected" />
              <span class="switch-label">
                {{ formRule.protected ? '🔒 已开启保护（最高优先级豁免，即使磁盘满也绝不删除该主播录像）' : '未开启（遵循全局磁盘清理策略）' }}
              </span>
            </div>
            <div class="form-tip">
              <strong>【最高优先级】</strong>开启后，后台磁盘自动清理任务将直接完全跳过该主播目录，绝不删除其名下的任何录播文件。
            </div>
          </el-form-item>

          <el-form-item label="个性化最少保留场次">
            <div class="inline-unit-row">
              <el-input-number
                v-model="formRule.min_keep"
                :min="0"
                :max="100"
                class="compact-counter"
              />
              <span class="unit-lbl">场</span>
            </div>
            <div class="form-tip">
              <strong>【覆盖全局】</strong>设为 0 时遵循系统全局设置；设为大于 0 则优先覆盖全局，无论磁盘多么紧张，始终保底保留该主播最新的 N 场录像。
            </div>
          </el-form-item>
        </div>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="ruleDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="savingRule" @click="handleSaveRule">保存规则</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 录像文件抽屉 -->
    <el-drawer
      v-model="drawerVisible"
      :title="`🎬 ${drawerStreamer} 的录像文件清单`"
      size="680px"
      destroy-on-close
    >
      <div v-loading="drawerLoading" class="drawer-content">
        <div class="drawer-header-bar">
          <span class="drawer-sub">共 {{ drawerFiles.length }} 个录像文件</span>
          <button class="drawer-action-btn" @click="goToTasks(drawerStreamer)">针对该主播执行任务 ↗</button>
        </div>

        <el-table :data="drawerFiles" size="small" max-height="650" class="drawer-table">
          <el-table-column prop="name" label="文件名" min-width="280" show-overflow-tooltip sortable>
            <template #default="{ row }">
              <span class="drawer-file-name">{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="size" label="大小" width="90" align="right" sortable :sort-method="(a: any, b: any) => a.size - b.size">
            <template #default="{ row }">
              <span class="mono-val">{{ row.size_str }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="is_merged" label="状态" width="80" align="center">
            <template #default="{ row }">
              <span class="status-dot-text" :class="row.is_merged ? 'dot-merged' : 'dot-pending'">
                {{ row.is_merged ? '合并版' : '分段' }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80" align="center">
            <template #default="{ row }">
              <button class="play-btn-sm" @click="openPlayer(drawerStreamer, row.name)" title="在浏览器中播放预览">
                ▶ 播放
              </button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-drawer>

    <!-- 在线视频播放预览弹窗 -->
    <el-dialog
      v-model="playerVisible"
      :title="`▶ 播放预览: ${playFileName}`"
      width="800px"
      destroy-on-close
      class="player-dialog"
    >
      <div class="player-wrap">
        <video
          v-if="playUrl"
          :src="playUrl"
          controls
          autoplay
          preload="metadata"
          class="video-element"
        >
          您的浏览器不支持视频播放。
        </video>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, onActivated } from "vue";
import { useRouter } from "vue-router";
import { getStreamers, getStreamerFiles, saveStreamerRule, getVideoPlayUrl } from "@/api/status";
import { Search } from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import type { StreamerInfo, StreamerFile, StreamerRule } from "@/api/types";

const router = useRouter();
const streamers = ref<StreamerInfo[]>([]);
const loading = ref(true);
const searchQuery = ref("");

const totalFiles = computed(() => streamers.value.reduce((acc, s) => acc + (s.files || 0), 0));
const totalSizeGB = computed(() => (streamers.value.reduce((acc, s) => acc + (s.size_gb || 0), 0)).toFixed(1));

function formatExactTime(ts: number): string {
  const d = new Date(ts * 1000);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

const filteredStreamers = computed(() => {
  if (!searchQuery.value) return streamers.value;
  const q = searchQuery.value.toLowerCase();
  return streamers.value.filter(s => s.name.toLowerCase().includes(q));
});

function goToTasks(name: string) {
  router.push({ path: "/tasks", query: { streamer: name } });
}

function sortMtime(a: StreamerInfo, b: StreamerInfo): number {
  return (a.mtime || 0) - (b.mtime || 0);
}

// 规则配置弹窗状态
const ruleDialogVisible = ref(false);
const currentStreamer = ref("");
const savingRule = ref(false);
const formRule = ref<StreamerRule>({
  trim_enabled: false,
  trim_start: "00:00:00",
  trim_end: "00:30:00",
  min_keep: 0,
  protected: false,
});

function fillQuick30() {
  formRule.value.trim_start = "00:00:00";
  formRule.value.trim_end = "00:30:00";
}

function openRuleDialog(row: StreamerInfo) {
  currentStreamer.value = row.name;
  if (row.rule) {
    formRule.value = {
      trim_enabled: !!row.rule.trim_enabled,
      trim_start: row.rule.trim_start || row.rule.start_time || "00:00:00",
      trim_end: row.rule.trim_end || row.rule.end_time || (row.rule.duration_minutes ? `00:${String(row.rule.duration_minutes).padStart(2, '0')}:00` : "00:30:00"),
      min_keep: row.rule.min_keep || 0,
      protected: !!row.rule.protected,
    };
  } else {
    formRule.value = {
      trim_enabled: false,
      trim_start: "00:00:00",
      trim_end: "00:30:00",
      min_keep: 0,
      protected: false,
    };
  }
  ruleDialogVisible.value = true;
}

async function handleSaveRule() {
  savingRule.value = true;
  try {
    await saveStreamerRule(currentStreamer.value, formRule.value);
    ElMessage.success("主播个性化规则已保存");
    ruleDialogVisible.value = false;
    const target = streamers.value.find(s => s.name === currentStreamer.value);
    if (target) {
      target.rule = { ...formRule.value };
      target.has_rule = true;
    }
  } catch (err: any) {
    ElMessage.error(err.message || "保存规则失败");
  } finally {
    savingRule.value = false;
  }
}

// 抽屉文件列表状态
const drawerVisible = ref(false);
const drawerLoading = ref(false);
const drawerStreamer = ref("");
const drawerFiles = ref<StreamerFile[]>([]);

async function openDrawer(row: StreamerInfo) {
  drawerStreamer.value = row.name;
  drawerVisible.value = true;
  drawerLoading.value = true;
  try {
    drawerFiles.value = await getStreamerFiles(row.name);
  } catch {
    ElMessage.error("获取文件列表失败");
  } finally {
    drawerLoading.value = false;
  }
}

// 播放器状态
const playerVisible = ref(false);
const playFileName = ref("");
const playUrl = ref("");

function openPlayer(streamer: string, fileName: string) {
  playFileName.value = fileName;
  playUrl.value = getVideoPlayUrl(streamer, fileName);
  playerVisible.value = true;
}

// tick 相对时间驱动
const tick = ref(0);
let tickTimer: ReturnType<typeof setInterval> | null = null;
onMounted(() => { tickTimer = setInterval(() => { tick.value++; }, 60_000); });
onUnmounted(() => { if (tickTimer) clearInterval(tickTimer); });

function formatRelativeTime(ts: number, _tick?: number): string {
  const diff = Date.now() / 1000 - ts;
  if (diff < 0) return "刚刚";
  if (diff < 60) return "刚刚";
  if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`;
  if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`;
  if (diff < 2592000) return `${Math.floor(diff / 86400)} 天前`;
  const d = new Date(ts * 1000);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

async function loadData() {
  try {
    streamers.value = await getStreamers();
  } finally {
    loading.value = false;
  }
}

onMounted(() => loadData());
onActivated(() => loadData());
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.card-title {
  display: inline-flex;
  align-items: center;
  font-weight: 600;
  font-size: 14px;
  line-height: 1;
  color: var(--ink);
}
.header-right {
  display: flex;
  gap: 12px;
  align-items: center;
}
.goto-tasks-btn {
  color: var(--stone) !important;
  font-weight: 500 !important;
  padding-right: 0 !important;
}
.goto-tasks-btn:hover {
  color: var(--ink) !important;
}
.arrow-ico {
  font-size: 11px;
  opacity: 0.7;
  margin-left: 2px;
}

.streamer-name {
  font-weight: 500;
  color: var(--ink);
}
.streamer-name-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

/* Badge pills */
.badge-pill {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: var(--r-sm);
  font-weight: 500;
  line-height: 1.4;
}
.badge-protected {
  background: var(--tint-mint);
  color: var(--success);
  border: 1px solid rgba(68, 131, 97, 0.25);
}
.badge-trim {
  background: var(--surface);
  color: var(--charcoal);
  border: 1px solid var(--hairline-strong);
}
.badge-minkeep {
  background: var(--highlight);
  color: var(--slate);
  border: 1px solid var(--hairline);
}

.data-block {
  display: inline-flex;
  align-items: baseline;
  gap: 2px;
}
.unit-lbl {
  font-size: 11px;
  color: var(--stone);
}

.action-btn-group {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}
.action-sep {
  color: var(--hairline-strong);
  font-size: 12px;
}
.ghost-action {
  background: transparent;
  border: none;
  color: var(--slate);
  cursor: pointer;
  font-size: 13px;
  padding: 2px 4px;
  border-radius: var(--r-sm);
  transition: all 0.15s;
}
.ghost-action:hover {
  color: var(--ink);
  background: var(--highlight);
}

/* Rule dialog styling */
.rule-dialog :deep(.el-dialog__body) {
  padding: 16px 20px;
}
.rule-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}
.title-with-badge {
  display: flex;
  align-items: baseline;
  gap: 8px;
}
.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--charcoal);
}
.section-sub {
  font-size: 11px;
  color: var(--steel);
}
.section-desc {
  font-size: 12px;
  color: var(--slate);
  line-height: 1.5;
  margin: 0 0 10px 0;
}
.time-range-group {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--surface);
  padding: 10px 12px;
  border-radius: var(--r-sm);
  border: 1px solid var(--hairline);
  margin-bottom: 6px;
}
.time-box {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
}
.time-label {
  font-size: 11px;
  color: var(--steel);
}
.range-arrow {
  color: var(--stone);
  font-size: 12px;
  padding: 0 4px;
  margin-top: 14px;
}
.quick-pill-btn {
  background: #fff;
  border: 1px solid var(--hairline-strong);
  color: var(--charcoal);
  border-radius: var(--r-sm);
  padding: 6px 10px;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  margin-top: 14px;
  transition: all 0.15s;
}
.quick-pill-btn:hover {
  background: var(--primary);
  color: #fff;
  border-color: var(--primary);
}
.form-tip {
  font-size: 12px;
  color: var(--steel);
  line-height: 1.5;
  margin-top: 4px;
}
.switch-row {
  display: flex;
  align-items: center;
  gap: 10px;
}
.switch-label {
  font-size: 12px;
  color: var(--charcoal);
}
.inline-unit-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

/* Drawer & Player styling */
.drawer-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 0 12px;
}
.drawer-header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--hairline);
}
.drawer-sub {
  font-size: 12px;
  color: var(--steel);
}
.drawer-file-name {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--ink);
}
.play-btn-sm {
  background: var(--surface);
  border: 1px solid var(--hairline-strong);
  border-radius: var(--r-sm);
  padding: 2px 6px;
  font-size: 11px;
  color: var(--charcoal);
  cursor: pointer;
  transition: all 0.15s;
}
.play-btn-sm:hover {
  background: var(--primary);
  color: #fff;
  border-color: var(--primary);
}

.player-wrap {
  background: #000;
  border-radius: var(--r-md);
  overflow: hidden;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 440px;
}
.video-element {
  width: 100%;
  max-height: 540px;
  outline: none;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.streamer-total-stat {
  display: inline-flex;
  align-items: center;
  height: 24px;
  line-height: 1;
  font-size: 12px;
  color: var(--steel, #888);
  background: var(--surface, #fafafa);
  border: 1px solid var(--hairline, #eee);
  padding: 0 8px;
  border-radius: var(--r-sm, 4px);
  font-family: var(--font-mono);
  box-sizing: border-box;
}
.streamer-total-stat strong {
  color: var(--ink, #333);
}
.header-refresh-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: transparent;
  border: 1px solid var(--hairline, #dcdfe6);
  padding: 4px 10px;
  font-size: 12px;
  color: var(--slate, #606266);
  border-radius: var(--r-sm, 4px);
  cursor: pointer;
  transition: all 0.15s ease;
}
.header-refresh-btn:hover {
  background: var(--highlight, #f5f7fa);
  color: var(--ink, #333);
}
.refresh-ico {
  width: 13px;
  height: 13px;
}
.col-header-tip {
  cursor: help;
  border-bottom: 1px dashed var(--hairline);
}
.cursor-help {
  cursor: help;
}
.drawer-action-btn {
  background: #edf6f0;
  border: 1px solid #c7e6d0;
  color: #2b6a48;
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.15s;
}
.drawer-action-btn:hover {
  background: #d8eedf;
}
</style>

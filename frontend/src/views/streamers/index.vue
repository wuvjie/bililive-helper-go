<template>
  <div class="streamers">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <div class="card-title-group">
              <div class="ui-icon-box"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg></div>
              <span class="title-text">主播管理</span>
            </div>
            <span class="streamer-total-stat">
              共 <strong>{{ streamers.length }}</strong> 位主播 · <strong>{{ totalFiles }}</strong> 个录像 · <strong>{{ totalSizeGB }}</strong> GB
            </span>
          </div>
          <div class="header-right">
            <div class="search-input-wrap">
              <el-input
                v-model="searchQuery"
                placeholder="搜索主播名..."
                :prefix-icon="Search"
                clearable
                size="small"
                style="width: 230px"
              />
              <span v-if="!searchQuery" class="kbd-badge search-kbd">/</span>
            </div>
            <button class="header-refresh-btn" @click="loadData" title="刷新列表">
              <svg class="refresh-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
              刷新
            </button>
          </div>
        </div>
      </template>

      <div class="table-responsive-wrapper">
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
              <span v-if="row.rule?.protected" class="badge-pill badge-protected">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
                <span>免删保护</span>
              </span>
              <span v-if="row.rule?.trim_enabled" class="badge-pill badge-trim">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="6" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><line x1="20" y1="4" x2="8.12" y2="15.88"/><line x1="14.47" y1="14.48" x2="20" y2="20"/><line x1="8.12" y1="8.12" x2="12" y2="12"/></svg>
                <span>{{ row.rule.trim_start || '00:00:00' }} ~ {{ row.rule.trim_end || '结束' }}</span>
              </span>
              <span v-if="row.rule?.min_keep > 0" class="badge-pill badge-minkeep">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
                <span>留 ≥ {{ row.rule.min_keep }} 场</span>
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
              <button class="ghost-action" @click="openDrawer(row)">
                <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
                <span>录像 ({{ row.files }})</span>
              </button>
              <span class="action-sep">|</span>
              <button class="ghost-action" @click="openRuleDialog(row)">
                <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
                <span>规则</span>
              </button>
            </div>
          </template>
        </el-table-column>
        <template #empty>
          <div class="streamer-empty-box empty-illustration">
            <div class="empty-icon-wrapper">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
            </div>
            <span class="empty-title">未发现匹配的主播</span>
            <span class="empty-desc">请确认搜索关键词或检查录像存储目录中是否存在主播子文件夹</span>
          </div>
        </template>
      </el-table>
      </div>
    </el-card>

    <!-- 主播规则配置弹窗 (卡片式设置面板) -->
    <el-dialog
      v-model="ruleDialogVisible"
      :title="`个性化规则 · ${currentStreamer}`"
      width="560px"
      destroy-on-close
      class="rule-dialog"
    >
      <div class="rule-dialog-body">
        <!-- Card 1: 视频时段裁剪 -->
        <div class="rule-setting-card">
          <div class="setting-card-header">
            <div class="header-title-box">
              <span class="setting-card-title"><svg class="inline-card-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="6" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><line x1="20" y1="4" x2="8.12" y2="15.88"/><line x1="14.47" y1="14.48" x2="20" y2="20"/><line x1="8.12" y1="8.12" x2="12" y2="12"/></svg>自动视频时段裁剪</span>
              <span class="setting-card-sub">无损流复制 · 零画质损失</span>
            </div>
            <el-switch v-model="formRule.trim_enabled" />
          </div>
          <p class="setting-card-desc">
            启用后，该主播分片合并转换完成后，系统将自动截取指定起止时段（如去除开播暖场或只保留前半场精选）。
          </p>

          <div v-if="formRule.trim_enabled" class="trim-config-box">
            <div class="time-range-group">
              <div class="time-box">
                <span class="time-label">开始时间</span>
                <el-input v-model="formRule.trim_start" placeholder="00:00:00" class="mono-input" />
              </div>
              <span class="range-arrow">至</span>
              <div class="time-box">
                <span class="time-label">结束时间</span>
                <el-input v-model="formRule.trim_end" placeholder="留空截到结束" class="mono-input" />
              </div>
            </div>
            <div class="trim-quick-presets">
              <span class="preset-label">快捷填入:</span>
              <button type="button" class="preset-pill" @click="fillTrimPreset('00:00:00', '00:30:00')">前 30 分钟</button>
              <button type="button" class="preset-pill" @click="fillTrimPreset('00:00:00', '01:00:00')">前 1 小时</button>
              <button type="button" class="preset-pill preset-pill-clear" @click="fillTrimPreset('00:00:00', '')">清空结束</button>
            </div>
            <div class="form-tip">
              * 结束时间留空表示截取到最后；时间格式为 <code>时:分:秒</code>（如 00:30:00）。
            </div>
          </div>
        </div>

        <!-- Card 2: 存储与清理保护 -->
        <div class="rule-setting-card">
          <div class="setting-card-header">
            <div class="header-title-box">
              <span class="setting-card-title"><svg class="inline-card-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>存储与清理保护</span>
              <span class="setting-card-sub">优先于系统全局清理策略</span>
            </div>
          </div>

          <div class="rule-item-block">
            <div class="rule-item-header">
              <div class="rule-item-left">
                <span class="rule-item-name">免清理免删保护</span>
                <span class="rule-badge-pill" :class="formRule.protected ? 'pill-protected' : 'pill-default'">
                  {{ formRule.protected ? '已豁免清理' : '遵循全局' }}
                </span>
              </div>
              <el-switch v-model="formRule.protected" />
            </div>
            <div class="form-tip">
              <strong>【最高优先级】</strong>开启后，后台自动磁盘清理任务将彻底跳过该主播目录，即使磁盘耗尽也绝不删除该主播的任何视频。
            </div>
          </div>

          <div class="rule-item-block rule-item-border">
            <div class="rule-item-header">
              <div class="rule-item-left">
                <span class="rule-item-name">个性化最少保留场次</span>
                <span class="rule-badge-sub">
                  {{ formRule.min_keep > 0 ? `至少保留最新的 ${formRule.min_keep} 场` : '继承全局预设 (3 场)' }}
                </span>
              </div>
              <div class="inline-unit-row">
                <el-input-number
                  v-model="formRule.min_keep"
                  :min="0"
                  :max="50"
                  class="compact-counter"
                  style="width: 110px;"
                />
                <span class="unit-lbl">场</span>
              </div>
            </div>
            <div class="form-tip">
              <strong>【覆盖全局】</strong>设为 0 时遵循全局策略；设为大于 0 则优先覆盖全局，无论空间多么紧张始终保底保留该主播最新的 N 场录播。
            </div>
          </div>
        </div>
      </div>

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
      :title="`录像文件清单 · ${drawerStreamer}`"
      size="680px"
      destroy-on-close
    >
      <div v-loading="drawerLoading" class="drawer-content">
        <div class="drawer-header-bar">
          <span class="drawer-sub">共 {{ drawerFiles.length }} 个录像文件</span>
          <button class="drawer-action-btn" @click="goToTasks(drawerStreamer)">针对该主播执行任务 ↗</button>
        </div>

        <div class="table-responsive-wrapper">
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
            <el-table-column prop="is_merged" label="状态" width="105" align="center" sortable>
              <template #default="{ row }">
                <el-tooltip :content="row.is_merged ? '已通过系统合并处理的完整录播视频' : '录播工具生成的原始录像文件（待合并或独立单场录像）'" placement="top">
                  <span class="file-status-badge" :class="row.is_merged ? 'status-merged' : 'status-raw'">
                    <span class="badge-dot"></span>
                    {{ row.is_merged ? '已合并' : '原始录像' }}
                  </span>
                </el-tooltip>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="130" align="center">
              <template #default="{ row }">
                <div class="drawer-file-actions">
                  <button class="play-btn-sm" @click="openPlayer(drawerStreamer, row.name)" title="在浏览器中播放预览">
                    ▶ 播放
                  </button>
                  <button class="play-btn-sm btn-del-video" @click="handleDeleteFile(row)" title="从磁盘永久删除该录像文件">
                    <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg> 删除
                  </button>
                </div>
              </template>
            </el-table-column>
            <template #empty>
              <div class="drawer-empty-box">
                <div class="empty-icon-wrapper"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6"><rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"/><line x1="7" y1="2" x2="7" y2="22"/><line x1="17" y1="2" x2="17" y2="22"/><line x1="2" y1="12" x2="22" y2="12"/></svg></div>
                <span class="empty-title">暂无录像文件</span>
                <span class="empty-desc">该主播录像目录暂未生成切片或录像文件</span>
              </div>
            </template>
          </el-table>
        </div>
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
import { getStreamers, getStreamerFiles, saveStreamerRule, getVideoPlayUrl, deleteStreamerFile } from "@/api/status";
import { Search } from "@element-plus/icons-vue";
import { ElMessage, ElMessageBox } from "element-plus";
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
  if (/\.(flv|ts)$/i.test(fileName)) {
    ElMessage.info("该录像是 FLV/TS 原始流媒体切片，浏览器暂不支持原生解码，建议通过合并任务转为 MP4 成片后预览");
    return;
  }
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
.table-responsive-wrapper {
  width: 100%;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  padding-bottom: 6px;
}
:deep(.streamer-table) {
  min-width: 880px;
}
:deep(.drawer-table) {
  min-width: 580px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.card-title-group {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 26px;
  line-height: 26px;
  flex-shrink: 0;
}
.title-icon {
  font-size: 16px;
  line-height: 1;
  display: inline-flex;
  align-items: center;
}
.title-text {
  font-size: 14px;
  font-weight: 600;
  color: var(--ink);
  line-height: 1;
  white-space: nowrap;
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
  padding-right: 8px;
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
  flex-wrap: nowrap;
}
.streamer-total-stat {
  display: inline-flex;
  align-items: center;
  height: 24px;
  line-height: 22px;
  font-size: 12px;
  color: var(--steel, #888);
  background: var(--surface, #fafafa);
  border: 1px solid var(--hairline, #eee);
  padding: 0 8px;
  border-radius: var(--r-sm, 4px);
  font-family: var(--font-mono);
  box-sizing: border-box;
  white-space: nowrap;
  flex-shrink: 0;
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

.streamer-empty-box, .drawer-empty-box {
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
  :deep(.rule-dialog) {
    width: 94% !important;
    max-width: 94% !important;
    margin: 16px auto !important;
  }
  :deep(.player-dialog) {
    width: 96% !important;
    max-width: 96% !important;
    margin: 10px auto !important;
  }
  :deep(.el-drawer) {
    width: 100% !important;
    max-width: 100% !important;
  }
  .player-wrap {
    min-height: 220px !important;
  }
  .card-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  .header-left {
    width: 100%;
    overflow-x: auto;
    scrollbar-width: none;
  }
  .header-left::-webkit-scrollbar {
    display: none;
  }
  .header-right {
    width: 100%;
    display: flex;
    gap: 8px;
  }
  .header-right :deep(.el-input) {
    flex: 1;
  }
}

.file-status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  font-weight: 500;
  padding: 2px 7px;
  border-radius: 10px;
  line-height: 1.4;
  white-space: nowrap;
}
.file-status-badge .badge-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}
.status-merged {
  background: #f0fdf4;
  color: #166534;
  border: 1px solid #bbf7d0;
}
.status-merged .badge-dot {
  background: #22c55e;
}
.status-raw {
  background: #f8fafc;
  color: #475569;
  border: 1px solid #cbd5e1;
}
.status-raw .badge-dot {
  background: #94a3b8;
}


/* Modern Rule Dialog Styling */
:deep(.rule-dialog) {
  border-radius: var(--r-md, 8px);
  overflow: hidden;
}
:deep(.rule-dialog .el-dialog__body) {
  padding: 16px 20px 20px !important;
}
.rule-dialog-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.rule-setting-card {
  background: var(--surface, #fcfcfc);
  border: 1px solid var(--hairline, #e2e8f0);
  border-radius: var(--r-md, 8px);
  padding: 14px 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02);
  transition: all 0.15s ease;
}
.setting-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.header-title-box {
  display: flex;
  align-items: baseline;
  gap: 8px;
  flex-wrap: wrap;
}
.setting-card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--ink, #1e293b);
}
.setting-card-sub {
  font-size: 11px;
  color: var(--steel, #64748b);
}
.setting-card-desc {
  font-size: 12px;
  color: var(--slate, #475569);
  line-height: 1.5;
  margin: 0 0 10px 0;
  background: rgba(0, 0, 0, 0.02);
  padding: 8px 10px;
  border-radius: var(--r-sm, 4px);
  border-left: 3px solid #3b82f6;
}
.trim-config-box {
  display: flex;
  flex-direction: column;
  gap: 10px;
  background: #ffffff;
  border: 1px solid var(--hairline, #e2e8f0);
  border-radius: var(--r-sm, 6px);
  padding: 12px;
  margin-top: 6px;
}
.time-range-group {
  display: flex;
  align-items: center;
  gap: 8px;
}
.time-box {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}
.time-label {
  font-size: 11px;
  color: var(--steel, #64748b);
  font-weight: 500;
}
.range-arrow {
  color: var(--stone, #94a3b8);
  font-size: 12px;
  padding: 0 4px;
  margin-top: 18px;
}
.trim-quick-presets {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin: 4px 0;
}
.preset-label {
  font-size: 11px;
  color: var(--steel, #64748b);
  font-weight: 500;
}
.preset-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: #f8fafc;
  border: 1px solid var(--hairline-strong, #cbd5e1);
  color: var(--charcoal, #334155);
  border-radius: 20px;
  padding: 4px 12px;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;
}
.preset-pill:hover {
  background: #eff6ff;
  border-color: #3b82f6;
  color: #2563eb;
  transform: translateY(-1px);
}
.preset-pill-clear {
  color: var(--steel, #64748b);
}
.preset-pill-clear:hover {
  background: #fef2f2;
  border-color: #f87171;
  color: #dc2626;
}
.rule-item-block {
  background: #ffffff;
  border: 1px solid var(--hairline, #e2e8f0);
  border-radius: var(--r-sm, 6px);
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 8px;
}
.rule-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.rule-item-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.rule-item-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--ink, #1e293b);
}
.rule-badge-pill {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 500;
}
.pill-protected {
  background: #eef8f1;
  color: #166534;
  border: 1px solid #bbf7d0;
}
.pill-default {
  background: #f1f5f9;
  color: #64748b;
  border: 1px solid #e2e8f0;
}
.rule-badge-sub {
  font-size: 11px;
  color: var(--steel, #64748b);
}

/* Drawer File Actions */
.drawer-file-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}
.btn-del-video {
  color: #dc2626 !important;
  border-color: #fecaca !important;
}
.btn-del-video:hover {
  background: #fee2e2 !important;
  border-color: #ef4444 !important;
  color: #b91c1c !important;
}

.card-title-group {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}
.btn-svg-sm {
  width: 13px;
  height: 13px;
  margin-right: 4px;
  vertical-align: -1px;
}
.ghost-action {
  display: inline-flex;
  align-items: center;
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


.inline-card-svg {
  width: 14px;
  height: 14px;
  margin-right: 6px;
  vertical-align: -2px;
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


html.dark {
  .quick-pill-btn {
    background: #18181b;
    border-color: #27272a;
    color: #f4f4f5;
    &:hover {
      background: #27272a;
      border-color: #3b82f6;
      color: #3b82f6;
    }
  }
  .trim-config-box, .rule-item-block, .rule-setting-card {
    background: #18181b !important;
    border-color: #27272a !important;
  }
  .header-refresh-btn {
    background: #18181b;
    border-color: #27272a;
    color: #f4f4f5;
    &:hover { background: #27272a; }
  }
  .drawer-footer {
    background: #141416;
    border-top-color: #27272a;
  }
}

</style>

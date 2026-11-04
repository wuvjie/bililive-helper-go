<template>
  <div class="streamers">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="card-title">👥 主播管理</span>
          <div class="header-right">
            <el-input
              v-model="searchQuery"
              placeholder="搜索主播..."
              clearable
              size="small"
              style="width: 200px"
              :prefix-icon="Search"
            />
            <el-button text size="small" class="goto-tasks-btn" @click="$router.push('/tasks')">
              前往任务中心 <span class="arrow-ico">↗</span>
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="filteredStreamers" v-loading="loading" empty-text="暂无主播数据" class="helper-clean-table">
        <el-table-column type="index" label="#" width="64" align="right" header-align="right">
          <template #default="{ $index }">
            <span class="idx-num">{{ $index + 1 }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="name" label="主播名称" min-width="220" show-overflow-tooltip sortable>
          <template #default="{ row }">
            <div class="streamer-name-cell">
              <span class="streamer-name">{{ row.name }}</span>
              <el-tag v-if="row.rule?.trim_enabled" size="small" type="success" effect="plain" class="rule-badge">
                {{ row.rule.trim_mode === 'range' ? `时段 ${row.rule.start_time || '0'}-${row.rule.end_time}` : `前 ${row.rule.duration_minutes || 30}m` }}
              </el-tag>
              <el-tag v-if="row.rule?.protected" size="small" type="warning" effect="plain" class="rule-badge">
                免删保护
              </el-tag>
              <el-tag v-else-if="row.rule?.min_keep > 0" size="small" type="info" effect="plain" class="rule-badge">
                留存≥{{ row.rule.min_keep }}
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="files" label="录制文件数" width="150" align="right" header-align="right" sortable :sort-orders="['descending', 'ascending']">
          <template #default="{ row }">
            <div class="data-block">
              <span class="mono-val">{{ row.files }}</span><span class="unit-lbl">个</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="size_gb" label="磁盘占用" width="150" align="right" header-align="right" sortable :sort-orders="['descending', 'ascending']">
          <template #default="{ row }">
            <div class="data-block">
              <span class="mono-val">{{ row.size_gb?.toFixed(2) }}</span><span class="unit-lbl">GB</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column
          prop="mtime"
          label="最新视频"
          width="140"
          align="right"
          header-align="right"
          sortable
          :sort-orders="['descending', 'ascending']"
          :sort-method="sortMtime"
        >
          <template #default="{ row }">
            <span class="mono-val">{{ row.mtime ? formatRelativeTime(row.mtime, tick) : '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="160" align="right" header-align="right">
          <template #default="{ row }">
            <div class="action-btn-group">
              <button class="ghost-action" @click="openRuleDialog(row)">⚙️ 规则</button>
              <span class="action-sep">|</span>
              <button class="ghost-action" @click="goToTasks(row.name)">管理文件</button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 主播规则配置弹窗 -->
    <el-dialog
      v-model="ruleDialogVisible"
      :title="`⚙️ ${currentStreamer} 的录像规则`"
      width="480px"
      destroy-on-close
      class="rule-dialog"
    >
      <el-form label-position="top" class="rule-form">
        <!-- 视频裁剪设置 -->
        <div class="rule-section">
          <div class="section-header">
            <span class="section-title">✂ 视频时段裁剪</span>
            <el-switch v-model="formRule.trim_enabled" />
          </div>
          <p class="section-desc">
            合并或转换完成后，仅保留指定时段的视频（采用无损流复制，秒级完成且画质无损）。
          </p>

          <template v-if="formRule.trim_enabled">
            <el-form-item label="裁剪模式">
              <el-radio-group v-model="formRule.trim_mode">
                <el-radio value="head">仅保留开头部分</el-radio>
                <el-radio value="range">自定义时间区间</el-radio>
              </el-radio-group>
            </el-form-item>

            <el-form-item v-if="formRule.trim_mode === 'head'" label="保留时长（分钟）">
              <el-input-number
                v-model="formRule.duration_minutes"
                :min="1"
                :max="720"
                style="width: 160px"
              />
              <span class="form-tip">例如设为 30，则每场录像只保留前 30 分钟</span>
            </el-form-item>

            <template v-else>
              <el-row :gutter="12">
                <el-col :span="12">
                  <el-form-item label="起始时间 (HH:MM:SS)">
                    <el-input v-model="formRule.start_time" placeholder="00:00:00" />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="结束时间 (HH:MM:SS)">
                    <el-input v-model="formRule.end_time" placeholder="00:45:00" />
                  </el-form-item>
                </el-col>
              </el-row>
              <span class="form-tip">指定视频内部的时间点，如从 00:05:00 到 00:35:00</span>
            </template>
          </template>
        </div>

        <el-divider style="margin: 16px 0" />

        <!-- 存储与清理个性化设置 -->
        <div class="rule-section">
          <div class="section-header">
            <span class="section-title">🛡️ 存储与清理偏好</span>
          </div>

          <el-form-item label="免清理保护">
            <div class="switch-row">
              <el-switch v-model="formRule.protected" />
              <span class="switch-label">{{ formRule.protected ? '已开启免删保护（磁盘超阈值时绝不自动清理此主播）' : '未开启（遵循全局清理规则）' }}</span>
            </div>
          </el-form-item>

          <el-form-item label="独立保底保留数（场次）">
            <el-input-number
              v-model="formRule.min_keep"
              :min="0"
              :max="100"
              style="width: 160px"
            />
            <span class="form-tip">设为 0 表示遵循全局“每主播最少保留”设置</span>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, onActivated } from "vue";
import { useRouter } from "vue-router";
import { getStreamers, saveStreamerRule } from "@/api/status";
import { Search } from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import type { StreamerInfo, StreamerRule } from "@/api/types";

const router = useRouter();
const streamers = ref<StreamerInfo[]>([]);
const loading = ref(true);
const searchQuery = ref("");

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

const ruleDialogVisible = ref(false);
const currentStreamer = ref("");
const savingRule = ref(false);
const formRule = ref<StreamerRule>({
  trim_enabled: false,
  trim_mode: "head",
  duration_minutes: 30,
  start_time: "00:00:00",
  end_time: "00:30:00",
  min_keep: 0,
  protected: false,
});

function openRuleDialog(row: StreamerInfo) {
  currentStreamer.value = row.name;
  if (row.rule) {
    formRule.value = {
      trim_enabled: !!row.rule.trim_enabled,
      trim_mode: row.rule.trim_mode || "head",
      duration_minutes: row.rule.duration_minutes || 30,
      start_time: row.rule.start_time || "00:00:00",
      end_time: row.rule.end_time || "00:30:00",
      min_keep: row.rule.min_keep || 0,
      protected: !!row.rule.protected,
    };
  } else {
    formRule.value = {
      trim_enabled: false,
      trim_mode: "head",
      duration_minutes: 30,
      start_time: "00:00:00",
      end_time: "00:30:00",
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
    ElMessage.success("主播规则已保存");
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

// tick 每 60 秒更新一次，驱动相对时间重新计算
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

onMounted(async () => {
  try {
    streamers.value = await getStreamers();
  } finally {
    loading.value = false;
  }
});

// Refresh data when component is re-activated by keep-alive router-view
onActivated(async () => {
  try {
    streamers.value = await getStreamers();
  } catch { /* ignore */ }
});
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.card-title {
  font-weight: 600;
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
.rule-badge {
  font-size: 11px !important;
  height: 20px !important;
  padding: 0 4px !important;
  border-radius: 3px !important;
}
.action-btn-group {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}
.action-sep {
  color: var(--line, #e2e2e2);
  font-size: 12px;
}
.rule-section {
  padding: 4px 0;
}
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}
.section-title {
  font-weight: 600;
  font-size: 14px;
  color: var(--ink);
}
.section-desc {
  font-size: 12px;
  color: var(--stone);
  margin: 0 0 12px 0;
  line-height: 1.5;
}
.form-tip {
  display: block;
  font-size: 12px;
  color: var(--stone);
  margin-top: 4px;
}
.switch-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.switch-label {
  font-size: 12px;
  color: var(--stone);
}
.idx-num { font-family: var(--font-mono); font-size: 13px; color: var(--stone); }

.data-block {
  display: block;
  width: 100%;
  text-align: right;
}
.mono-val { font-family: var(--font-mono); font-size: 13px; color: var(--ink); font-weight: 500; }
.unit-lbl { font-size: 11px; color: var(--stone); margin-left: 2px; }

.ghost-action {
  color: var(--stone);
  background: transparent;
  border: none;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: color 0.15s;
  padding: 0;
  margin: 0;
}
.ghost-action:hover {
  color: var(--ink);
}

/* Header right-aligned columns — flex right flow, sort arrow pushed to end */
:deep(.helper-clean-table) th.el-table__cell.is-right .cell {
  display: flex !important;
  width: 100% !important;
  box-sizing: border-box !important;
  align-items: center !important;
  justify-content: flex-end !important;
  padding-right: 0 !important;
}

/* Sort arrow always on the right side of header text */
:deep(.helper-clean-table) th.el-table__cell.is-right .cell .caret-wrapper {
  order: 2 !important;
  margin-left: 4px !important;
  margin-right: 0 !important;
}

/* Header text flows right */
:deep(.helper-clean-table) th.el-table__cell.is-right .cell span:not(.caret-wrapper) {
  order: 1 !important;
  text-align: right !important;
}

/* Data cells — clean right alignment */
:deep(.helper-clean-table) td.el-table__cell.is-right .cell,
:deep(.helper-clean-table) td.el-table__cell:last-child .cell {
  padding-right: 0 !important;
  text-align: right !important;
}
</style>

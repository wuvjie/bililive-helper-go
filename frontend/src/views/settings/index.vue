<template>
  <div class="settings">
    <el-card shadow="hover">
      <template #header><span>⚙️ 全局设置</span></template>
      <el-tabs v-model="activeTab" tab-position="left" class="settings-tabs">

        <!-- Tab 1: General Config -->
        <el-tab-pane label="基本配置" name="general">
          <el-form label-width="160px" label-position="right" v-loading="configLoading">
            <el-form-item label="录制目录">
              <el-input v-model="config.TARGET_DIR" placeholder="/path/to/recordings" />
            </el-form-item>
            <el-form-item label="合并时间阈值(分钟)">
              <el-input-number v-model="config.MERGE_AGE_MINUTES" :min="5" :max="1440" />
            </el-form-item>
            <el-form-item label="片段间隔(分钟)">
              <el-input-number v-model="config.GAP_MINUTES" :min="1" :max="120" />
            </el-form-item>
            <el-form-item label="安全模式">
              <el-select v-model="config.SAFE_MODE" style="width: 200px">
                <el-option label="按小时" value="hours" />
                <el-option label="按天" value="days" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="config.SAFE_MODE === 'hours'" label="安全期(分钟)">
              <el-input-number v-model="config.SAFE_AGE_MINUTES" :min="10" :max="1440" />
            </el-form-item>
            <el-form-item v-if="config.SAFE_MODE === 'days'" label="安全期(天)">
              <el-input-number v-model="config.SAFE_DAYS" :min="1" :max="365" />
            </el-form-item>
            <el-form-item label="单次最大删除数">
              <el-input-number v-model="config.MAX_DELETE_PER_RUN" :min="1" :max="1000" />
            </el-form-item>
            <el-form-item label="白名单关键词">
              <el-input v-model="config.WHITELIST_KEYWORDS" placeholder="关键词用逗号分隔" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="saving" @click="handleSaveConfig">保存配置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- Tab 2: Storage -->
        <el-tab-pane label="存储管理" name="storage">
          <el-form label-width="180px" label-position="right">
            <el-form-item label="触发清理阈值">
              <el-slider v-model="config.TRIGGER_THRESHOLD" :min="50" :max="99" :format-tooltip="(v: number) => v + '%'" show-input />
            </el-form-item>
            <el-form-item label="目标清理阈值">
              <el-slider v-model="config.TARGET_THRESHOLD" :min="30" :max="89" :format-tooltip="(v: number) => v + '%'" show-input />
            </el-form-item>
            <el-form-item label="每主播最少保留">
              <el-input-number v-model="config.MIN_KEEP_PER_STREAMER" :min="1" :max="50" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="saving" @click="handleSaveConfig">保存配置</el-button>
            </el-form-item>
            <el-divider />
            <el-form-item label="清理预估">
              <span v-if="cleanEstimate">
                {{ cleanEstimate.file_count }} 个文件，{{ cleanEstimate.total_size_gb?.toFixed(2) }} GB
              </span>
              <span v-else class="placeholder">加载中...</span>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- Tab 3: Schedule -->
        <el-tab-pane label="定时任务" name="schedule">
          <el-form label-width="160px" label-position="right" v-loading="scheduleLoading">
            <el-divider content-position="left">自动合并</el-divider>
            <el-form-item label="启用">
              <el-switch v-model="scheduleForm.merge_enabled" />
            </el-form-item>
            <el-form-item label="执行间隔(分钟)">
              <el-input-number v-model="scheduleForm.merge_interval" :min="10" :max="1440" />
            </el-form-item>
            <el-divider content-position="left">自动清理</el-divider>
            <el-form-item label="启用">
              <el-switch v-model="scheduleForm.clean_enabled" />
            </el-form-item>
            <el-form-item label="执行间隔(分钟)">
              <el-input-number v-model="scheduleForm.clean_interval" :min="10" :max="1440" />
            </el-form-item>
            <el-divider content-position="left">备份窗口（暂停任务）</el-divider>
            <el-form-item label="开始时间">
              <el-time-picker v-model="backupStart" format="HH:mm" value-format="HH:mm" />
            </el-form-item>
            <el-form-item label="结束时间">
              <el-time-picker v-model="backupEnd" format="HH:mm" value-format="HH:mm" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="saving" @click="handleSaveSchedule">保存计划</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- Tab 4: Diagnostics -->
        <el-tab-pane label="系统诊断" name="diagnostics">
          <el-descriptions :column="2" border v-loading="diagLoading">
            <el-descriptions-item label="FFmpeg">
              <el-tag :type="setupData?.ffmpeg_ok ? 'success' : 'danger'" size="small">
                {{ setupData?.ffmpeg_ok ? '✅ 正常' : '❌ 异常' }}
              </el-tag>
              <span class="ml-8" style="color:#999; font-size:12px">{{ setupData?.ffmpeg_path }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="FFprobe">
              <el-tag :type="setupData?.ffprobe_ok ? 'success' : 'danger'" size="small">
                {{ setupData?.ffprobe_ok ? '✅ 正常' : '❌ 异常' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="目标目录存在">
              <el-tag :type="setupData?.target_dir_exists ? 'success' : 'danger'" size="small">
                {{ setupData?.target_dir_exists ? '是' : '否' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="目录可写">
              <el-tag :type="setupData?.target_dir_writable ? 'success' : 'danger'" size="small">
                {{ setupData?.target_dir_writable ? '是' : '否' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="主播数">{{ setupData?.streamer_count || 0 }}</el-descriptions-item>
            <el-descriptions-item label="视频数">{{ setupData?.video_count || 0 }}</el-descriptions-item>
            <el-descriptions-item label="总大小">{{ setupData?.total_size_gb?.toFixed(2) }} GB</el-descriptions-item>
            <el-descriptions-item label="磁盘总容量">{{ setupData?.disk_total_gb?.toFixed(1) }} GB</el-descriptions-item>
            <el-descriptions-item label="磁盘剩余">{{ setupData?.disk_free_gb?.toFixed(1) }} GB</el-descriptions-item>
            <el-descriptions-item label="磁盘使用率">
              <el-progress :percentage="setupData?.disk_usage_pct || 0" :stroke-width="14" style="width:200px" />
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>

        <!-- Tab 5: AI Recommendation -->
        <el-tab-pane label="智能推荐" name="recommend">
          <div v-loading="recommendLoading">
            <template v-if="recommend">
              <el-alert :type="riskAlertType" :closable="false" class="mb-16">
                <template #title>
                  风险等级：<strong>{{ recommend.risk_level?.toUpperCase() }}</strong>
                  <span v-if="recommend.reason" style="margin-left: 12px; font-weight: normal; opacity: 0.8">{{ recommend.reason }}</span>
                </template>
              </el-alert>
              <el-descriptions :column="3" size="small" border class="mb-16">
                <el-descriptions-item label="磁盘使用率">{{ recommend.current_usage?.toFixed(1) }}%</el-descriptions-item>
                <el-descriptions-item label="总容量">{{ recommend.total_gb?.toFixed(0) }} GB</el-descriptions-item>
                <el-descriptions-item label="剩余">{{ recommend.free_gb?.toFixed(1) }} GB</el-descriptions-item>
                <el-descriptions-item label="主播数">{{ recommend.analysis?.streamer_count || 0 }}</el-descriptions-item>
                <el-descriptions-item label="日产出">{{ (recommend.analysis?.daily_output_gb || 0).toFixed(1) }} GB</el-descriptions-item>
                <el-descriptions-item label="可维持">{{ (recommend.analysis?.days_until_full || 0).toFixed(0) }} 天</el-descriptions-item>
              </el-descriptions>
              <el-table :data="recommendTable" stripe size="small" border>
                <el-table-column prop="key" label="参数" width="200" />
                <el-table-column prop="current" label="当前值" width="120" />
                <el-table-column prop="recommended" label="推荐值" width="120" />
              </el-table>
              <el-button type="primary" class="mt-16" @click="applyRecommend">应用推荐值</el-button>
            </template>
            <el-empty v-else description="点击加载推荐" />
            <el-button v-if="!recommend" type="primary" @click="loadRecommend">加载推荐</el-button>
          </div>
        </el-tab-pane>

        <!-- Tab 6: Backup -->
        <el-tab-pane label="配置备份" name="backup">
          <el-row :gutter="16">
            <el-col :span="12">
              <h4>导出配置</h4>
              <el-button type="primary" @click="handleExport">导出为 JSON</el-button>
              <el-input v-if="exportJson" type="textarea" :rows="10" v-model="exportJson" class="mt-12" readonly />
            </el-col>
            <el-col :span="12">
              <h4>导入配置</h4>
              <el-input type="textarea" :rows="10" v-model="importJson" placeholder="粘贴配置 JSON" />
              <el-button type="warning" class="mt-12" :loading="importing" @click="handleImport">导入恢复</el-button>
            </el-col>
          </el-row>
        </el-tab-pane>

      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from "vue";
import { ElMessage } from "element-plus";
import { getConfig, saveConfig, getConfigRecommend, getConfigExport, importConfig as apiImportConfig } from "@/api/config";
import { getSchedule, saveSchedule } from "@/api/schedule";
import { setupCheck } from "@/api/setup";
import { getCleanEstimate } from "@/api/task";
import type { ScheduleStatus, SetupCheck, CleanEstimate, ConfigRecommend } from "@/api/types";

const activeTab = ref("general");
const saving = ref(false);
const configLoading = ref(false);
const scheduleLoading = ref(false);
const diagLoading = ref(false);
const recommendLoading = ref(false);
const importing = ref(false);

const config = reactive<Record<string, any>>({});
const scheduleForm = reactive({
  merge_enabled: true,
  clean_enabled: true,
  merge_interval: 360,
  clean_interval: 720
});
const backupStart = ref("");
const backupEnd = ref("");

const setupData = ref<SetupCheck>();
const cleanEstimate = ref<CleanEstimate>();
const recommend = ref<ConfigRecommend>();
const exportJson = ref("");
const importJson = ref("");

const recommendTable = computed(() => {
  if (!recommend.value) return [];
  const r = recommend.value;
  const cfg = config;
  return [
    { key: "触发清理阈值 (%)", current: cfg.TRIGGER_THRESHOLD, recommended: r.TRIGGER_THRESHOLD },
    { key: "目标清理阈值 (%)", current: cfg.TARGET_THRESHOLD, recommended: r.TARGET_THRESHOLD },
    { key: "每主播最少保留", current: cfg.MIN_KEEP_PER_STREAMER, recommended: r.MIN_KEEP_PER_STREAMER },
    { key: "安全期 (分钟)", current: cfg.SAFE_AGE_MINUTES, recommended: r.SAFE_AGE_MINUTES },
    { key: "安全模式", current: cfg.SAFE_MODE, recommended: r.SAFE_MODE },
    { key: "合并时间阈值 (分钟)", current: cfg.MERGE_AGE_MINUTES, recommended: r.MERGE_AGE_MINUTES },
    { key: "单次最大删除数", current: cfg.MAX_DELETE_PER_RUN, recommended: r.MAX_DELETE_PER_RUN },
    { key: "片段间隔 (分钟)", current: cfg.GAP_MINUTES, recommended: r.GAP_MINUTES }
  ];
});

const riskAlertType = computed(() => {
  const level = recommend.value?.risk_level;
  if (level === "critical") return "error" as const;
  if (level === "high") return "warning" as const;
  return "info" as const;
});

async function handleSaveConfig() {
  saving.value = true;
  try {
    const payload = { ...config };
    // Convert whitelist string to array
    if (typeof payload.WHITELIST_KEYWORDS === "string") {
      payload.WHITELIST_KEYWORDS = payload.WHITELIST_KEYWORDS.split(",").map((s: string) => s.trim()).filter(Boolean);
    }
    await saveConfig(payload);
    ElMessage.success("配置已保存");
  } finally {
    saving.value = false;
  }
}

async function handleSaveSchedule() {
  saving.value = true;
  try {
    const data: any = { ...scheduleForm };
    if (backupStart.value) {
      const [h, m] = backupStart.value.split(":");
      data.BACKUP_START_HOUR = parseInt(h);
      data.BACKUP_START_MINUTE = parseInt(m);
    }
    if (backupEnd.value) {
      const [h, m] = backupEnd.value.split(":");
      data.BACKUP_END_HOUR = parseInt(h);
      data.BACKUP_END_MINUTE = parseInt(m);
    }
    await saveSchedule(data);
    ElMessage.success("计划已保存");
  } finally {
    saving.value = false;
  }
}

async function loadRecommend() {
  recommendLoading.value = true;
  try {
    recommend.value = await getConfigRecommend();
  } finally {
    recommendLoading.value = false;
  }
}

function applyRecommend() {
  if (!recommend.value) return;
  const r = recommend.value;
  config.TRIGGER_THRESHOLD = r.TRIGGER_THRESHOLD;
  config.TARGET_THRESHOLD = r.TARGET_THRESHOLD;
  config.MIN_KEEP_PER_STREAMER = r.MIN_KEEP_PER_STREAMER;
  config.SAFE_AGE_MINUTES = r.SAFE_AGE_MINUTES;
  config.SAFE_MODE = r.SAFE_MODE;
  config.MERGE_AGE_MINUTES = r.MERGE_AGE_MINUTES;
  config.MAX_DELETE_PER_RUN = r.MAX_DELETE_PER_RUN;
  config.GAP_MINUTES = r.GAP_MINUTES;
  activeTab.value = "general";
  ElMessage.success("已填入推荐值，请手动保存");
}

async function handleExport() {
  try {
    const data = await getConfigExport();
    exportJson.value = JSON.stringify(data, null, 2);
    // Also download
    const blob = new Blob([exportJson.value], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `config-backup-${new Date().toISOString().slice(0, 10)}.json`;
    a.click();
    URL.revokeObjectURL(url);
    ElMessage.success("导出成功");
  } catch {}
}

async function handleImport() {
  if (!importJson.value.trim()) {
    ElMessage.warning("请粘贴配置内容");
    return;
  }
  importing.value = true;
  try {
    const data = JSON.parse(importJson.value);
    await apiImportConfig(data);
    ElMessage.success("导入成功");
    // Reload config
    const c = await getConfig();
    Object.assign(config, c);
  } catch (e: any) {
    if (e instanceof SyntaxError) {
      ElMessage.error("JSON 格式错误");
    }
  } finally {
    importing.value = false;
  }
}

onMounted(async () => {
  const [c, s, d, ce] = await Promise.allSettled([
    getConfig(),
    getSchedule(),
    setupCheck(),
    getCleanEstimate()
  ]);
  if (c.status === "fulfilled") {
    Object.assign(config, c.value);
    // Convert whitelist array to comma-separated string
    if (Array.isArray(c.value.WHITELIST_KEYWORDS)) {
      config.WHITELIST_KEYWORDS = c.value.WHITELIST_KEYWORDS.join(", ");
    }
    // Load backup window from config
    const cv = c.value;
    if (cv.BACKUP_START_HOUR != null) {
      backupStart.value = `${String(cv.BACKUP_START_HOUR).padStart(2, "0")}:${String(cv.BACKUP_START_MINUTE || 0).padStart(2, "0")}`;
    }
    if (cv.BACKUP_END_HOUR != null) {
      backupEnd.value = `${String(cv.BACKUP_END_HOUR).padStart(2, "0")}:${String(cv.BACKUP_END_MINUTE || 0).padStart(2, "0")}`;
    }
  }
  if (s.status === "fulfilled") {
    const sv = s.value;
    scheduleForm.merge_enabled = sv.merge_enabled;
    scheduleForm.clean_enabled = sv.clean_enabled;
    scheduleForm.merge_interval = sv.merge_interval;
    scheduleForm.clean_interval = sv.clean_interval;
  }
  if (d.status === "fulfilled") setupData.value = d.value;
  if (ce.status === "fulfilled") cleanEstimate.value = ce.value;
});
</script>

<style scoped>
.settings-tabs {
  min-height: 500px;
}
.mb-16 { margin-bottom: 16px; }
.mt-12 { margin-top: 12px; }
.mt-16 { margin-top: 16px; }
.ml-8 { margin-left: 8px; }
.placeholder { color: #c0c4cc; }
h4 { margin-bottom: 12px; color: #303133; }
</style>

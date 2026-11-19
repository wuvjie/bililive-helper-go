<template>
  <el-dialog
    :model-value="visible"
    :title="`单文件无损剪辑 · ${streamer}`"
    width="min(560px, 92vw)"
    destroy-on-close
    :close-on-click-modal="!submitting"
    :close-on-press-escape="!submitting"
    :show-close="!submitting"
    @update:visible="!submitting && emit('update:visible', $event)"
    @update:model-value="!submitting && emit('update:visible', $event)"
  >
    <div v-if="file" v-loading="probing" element-loading-text="正在探测视频时长与元数据..." class="trim-dialog-body">
      <!-- 目标文件信息摘要（Primer Sub-header Block） -->
      <div class="file-meta-card">
        <div class="file-name-row">
          <svg class="file-meta-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18" />
            <line x1="7" y1="2" x2="7" y2="22" />
            <line x1="17" y1="2" x2="17" y2="22" />
            <line x1="2" y1="12" x2="22" y2="12" />
          </svg>
          <span class="file-name-text" :title="file.name">{{ file.name }}</span>
        </div>
        <div class="file-meta-pills">
          <span class="meta-pill">大小 {{ probeInfo?.size_str || file.size_str }}</span>
          <span v-if="probeInfo?.probe_ok" class="meta-pill meta-pill-accent">
            总时长 {{ formatFriendlyDuration(probeInfo.duration_sec) }}
            <code>{{ formatDurationClock(probeInfo.duration_sec) }}</code>
          </span>
          <span v-else-if="!probing" class="meta-pill">总时长未知</span>
          <span v-if="probeInfo?.has_danmaku" class="meta-pill meta-pill-ok">含伴生弹幕 · 将自动对齐</span>
        </div>
      </div>

      <!-- 重复剪辑警示 Banner (检测 FFmpeg comment=bh_trimmed 或 -剪辑版 文件名) -->
      <div v-if="probeInfo?.is_trimmed || probeInfo?.is_trimmed_file" class="trim-alert-banner is-warning">
        <svg class="alert-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
          <line x1="12" y1="9" x2="12" y2="13" />
          <line x1="12" y1="17" x2="12.01" y2="17" />
        </svg>
        <span>
          检测到该视频已带有裁剪标记（此前已剪辑过）。继续操作将基于当前成片时间轴进行二次截取，建议选择「另存副本」以保留原文件。
        </span>
      </div>

      <!-- 剪辑区间与时间轴预览 -->
      <div class="trim-section">
        <div class="section-header-row">
          <span class="section-label">剪辑起止区间</span>
          <div class="quick-presets-row">
            <button
              type="button"
              class="preset-pill"
              :disabled="submitting"
              @click="applyPreset('00:05:00', '')"
            >
              跳过前 5 分钟
            </button>
            <button
              type="button"
              class="preset-pill"
              :disabled="submitting"
              @click="applyPreset('00:00:00', '00:30:00')"
            >
              保留前 30 分钟
            </button>
            <button
              type="button"
              class="preset-pill"
              :disabled="submitting"
              @click="applyPreset('00:00:00', '01:00:00')"
            >
              保留前 1 小时
            </button>
            <button
              v-if="hasStreamerRuleRange"
              type="button"
              class="preset-pill pill-rule"
              :disabled="submitting"
              @click="applyPreset(rule?.trim_start || '00:00:00', rule?.trim_end || '')"
            >
              套用主播规则
            </button>
          </div>
        </div>

        <!-- 胶片式动态时间轴可视化条 -->
        <div class="film-timeline-wrap">
          <div class="film-strip-track">
            <div
              v-if="timelineSegments.hasLeading"
              class="film-seg seg-discarded"
              :style="{ flexGrow: timelineSegments.leadingGrow }"
            >
              <span class="seg-text">剪除前段</span>
            </div>

            <div
              class="film-seg seg-retained"
              :style="{ flexGrow: timelineSegments.retainedGrow }"
            >
              <div class="film-sprockets top"></div>
              <div class="retained-content">
                <span class="seg-text-bold">保留导出片段</span>
                <span v-if="trimSemanticInfo.durationText" class="retained-badge">
                  {{ trimSemanticInfo.durationText }}
                </span>
              </div>
              <div class="film-sprockets bottom"></div>
            </div>

            <div
              v-if="timelineSegments.hasTrailing"
              class="film-seg seg-discarded"
              :style="{ flexGrow: timelineSegments.trailingGrow }"
            >
              <span class="seg-text">剪除后段</span>
            </div>
          </div>

          <div class="timeline-ticks-row">
            <span class="tick-node">00:00:00</span>
            <span v-if="timelineSegments.hasLeading" class="tick-node">
              起始 {{ trimStart || "00:00:00" }}
            </span>
            <span v-if="trimEnd.trim()" class="tick-node">截止 {{ trimEnd }}</span>
            <span class="tick-node">
              {{ probeInfo?.probe_ok ? `末尾 ${formatDurationClock(probeInfo.duration_sec)}` : "录像结尾" }}
            </span>
          </div>
        </div>

        <!-- 起止时间输入 -->
        <div class="time-inputs-row">
          <div class="time-field">
            <label class="time-field-label">截取起始时间</label>
            <el-input
              v-model="trimStart"
              :disabled="submitting"
              placeholder="00:00:00（从片头开始）"
              class="mono-input"
              @blur="normalizeField('start')"
            />
          </div>
          <span class="time-range-sep">至</span>
          <div class="time-field">
            <label class="time-field-label">截取结束时间</label>
            <el-input
              v-model="trimEnd"
              :disabled="submitting"
              placeholder="留空则截取至录像结尾"
              class="mono-input"
              @blur="normalizeField('end')"
            />
          </div>
        </div>

        <!-- 实时校验与时长边界提示 -->
        <div
          class="trim-helper-note"
          :class="{
            'is-invalid': trimSemanticInfo.isInvalid,
            'is-warn': !trimSemanticInfo.isInvalid && Boolean(trimSemanticInfo.warningText)
          }"
        >
          <span>
            {{ trimSemanticInfo.warningText || trimSemanticInfo.text }}
          </span>
        </div>
      </div>

      <!-- 输出与保存方式 -->
      <div class="trim-section">
        <div class="section-header-row">
          <span class="section-label">成片保存方式</span>
        </div>
        <div class="mode-options-grid">
          <label class="mode-card" :class="{ active: saveMode === 'copy', disabled: submitting }">
            <input v-model="saveMode" type="radio" value="copy" :disabled="submitting" />
            <div class="mode-card-body">
              <div class="mode-card-title">
                <span>另存为副本（-剪辑版.mp4）</span>
                <span class="mode-badge-safe">推荐 · 保留原片</span>
              </div>
              <p class="mode-card-desc">
                保留原视频不变，生成同前缀独立副本并同步弹幕，后台自动任务不会将其与原片二次合并。
              </p>
            </div>
          </label>

          <label class="mode-card" :class="{ active: saveMode === 'replace', 'is-danger': saveMode === 'replace', disabled: submitting }">
            <input v-model="saveMode" type="radio" value="replace" :disabled="submitting" />
            <div class="mode-card-body">
              <div class="mode-card-title">
                <span>覆盖原文件</span>
                <span class="mode-badge-warn">不可撤销 · 节省空间</span>
              </div>
              <p class="mode-card-desc">
                原子替换当前原视频与同名弹幕时间轴，保留原始录制时间，适合明确无需保留原完整录像的场景。
              </p>
            </div>
          </label>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button :disabled="submitting" @click="handleCancel">取消</el-button>
        <el-button
          type="primary"
          :loading="submitting"
          :disabled="probing || trimSemanticInfo.isInvalid"
          @click="handleConfirmTrim"
        >
          {{ submitting ? "正在无损剪辑..." : saveMode === "copy" ? "开始剪辑并另存副本" : "确认剪辑并覆盖原文件" }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { ElMessage } from "element-plus";
import { probeStreamerFile, manualTrimStreamerFile } from "@/api/status";
import type {
  StreamerFile,
  StreamerFileProbe,
  StreamerRule,
  ManualTrimMode,
  ManualTrimResult
} from "@/api/types";
import {
  computeTimelineSegments,
  describeTrimRange,
  formatDurationClock,
  formatFriendlyDuration,
  normalizeTimeString
} from "@/utils/trim";

const props = defineProps<{
  visible: boolean;
  streamer: string;
  file: StreamerFile | null;
  rule?: StreamerRule;
}>();

const emit = defineEmits<{
  (e: "update:visible", val: boolean): void;
  (e: "trimmed", result: ManualTrimResult): void;
}>();

const trimStart = ref("00:00:00");
const trimEnd = ref("");
const saveMode = ref<ManualTrimMode>("copy");
const probing = ref(false);
const submitting = ref(false);
const probeInfo = ref<StreamerFileProbe | null>(null);

let probeToken = 0;

const hasStreamerRuleRange = computed(() =>
  Boolean(props.rule && (props.rule.trim_start?.trim() || props.rule.trim_end?.trim()))
);

const probedDurationSec = computed(() =>
  probeInfo.value?.probe_ok && probeInfo.value.duration_sec > 0 ? probeInfo.value.duration_sec : null
);

const timelineSegments = computed(() =>
  computeTimelineSegments(trimStart.value, trimEnd.value, probedDurationSec.value)
);

const trimSemanticInfo = computed(() =>
  describeTrimRange(trimStart.value, trimEnd.value, {
    totalDurationSec: probedDurationSec.value,
    requireSelection: true
  })
);

watch(
  () => [props.visible, props.streamer, props.file?.name] as const,
  async ([vis, streamer, fileName]) => {
    if (!vis || !fileName || !streamer) {
      probeToken++;
      probeInfo.value = null;
      probing.value = false;
      return;
    }

    // 初始化表单：若主播配置过时段偏好且合理，默认预填方便按需微调
    trimStart.value = props.rule?.trim_start?.trim() ? normalizeTimeString(props.rule.trim_start) : "00:00:00";
    trimEnd.value = props.rule?.trim_end?.trim() ? normalizeTimeString(props.rule.trim_end) : "";
    saveMode.value = "copy";
    probeInfo.value = null;
    probing.value = true;

    const currentToken = ++probeToken;
    try {
      const res = await probeStreamerFile(streamer, fileName);
      if (currentToken === probeToken) {
        probeInfo.value = res;
      }
    } catch {
      if (currentToken === probeToken) {
        probeInfo.value = null;
      }
    } finally {
      if (currentToken === probeToken) {
        probing.value = false;
      }
    }
  }
);

function applyPreset(start: string, end: string) {
  trimStart.value = start;
  trimEnd.value = end;
}

function normalizeField(field: "start" | "end") {
  if (field === "start") {
    trimStart.value = trimStart.value.trim() ? normalizeTimeString(trimStart.value) : "00:00:00";
  } else if (trimEnd.value.trim()) {
    trimEnd.value = normalizeTimeString(trimEnd.value);
  }
}

function handleCancel() {
  if (submitting.value) return;
  emit("update:visible", false);
}

async function handleConfirmTrim() {
  if (!props.file || submitting.value) return;
  normalizeField("start");
  normalizeField("end");

  if (trimSemanticInfo.value.isInvalid) {
    ElMessage.warning(trimSemanticInfo.value.text);
    return;
  }

  submitting.value = true;
  try {
    const res = await manualTrimStreamerFile(props.streamer, {
      file: props.file.name,
      trim_start: trimStart.value.trim() || "00:00:00",
      trim_end: trimEnd.value.trim(),
      mode: saveMode.value
    });
    const danmakuTip = res.danmaku_synced ? "，伴生弹幕已同步对齐" : "";
    ElMessage.success(`剪辑完成：${res.output_file}（${res.output_size_str}${danmakuTip}）`);
    emit("trimmed", res);
    emit("update:visible", false);
  } catch (err: unknown) {
    const msg = err instanceof Error && err.message ? err.message : "手动剪辑失败，请查看操作日志";
    ElMessage.error(msg);
  } finally {
    submitting.value = false;
  }
}
</script>

<style scoped>
.trim-dialog-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.file-meta-card {
  padding: 10px 12px;
  border-radius: var(--r-md);
  background: var(--paper-alt, #f6f8fa);
  border: 1px solid var(--hairline);
  display: flex;
  flex-direction: column;
  gap: 8px;
}

html.dark .file-meta-card {
  background: rgba(255, 255, 255, 0.03);
}

.file-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.file-meta-ico {
  width: 15px;
  height: 15px;
  flex-shrink: 0;
  color: var(--color-accent-fg);
}

.file-name-text {
  font-family: var(--font-mono);
  font-size: 12.5px;
  font-weight: 600;
  color: var(--ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-meta-pills {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.meta-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 8px;
  border-radius: var(--r-sm);
  font-size: 11px;
  background: var(--surface);
  border: 1px solid var(--hairline);
  color: var(--stone);
}

.meta-pill code {
  font-family: var(--font-mono);
  font-size: 11px;
  color: inherit;
}

.meta-pill-accent {
  color: var(--color-accent-fg);
  border-color: color-mix(in srgb, var(--color-accent-fg) 30%, transparent);
  background: color-mix(in srgb, var(--color-accent-fg) 8%, transparent);
}

.meta-pill-ok {
  color: #16a34a;
  border-color: rgba(22, 163, 74, 0.3);
  background: rgba(22, 163, 74, 0.08);
}

html.dark .meta-pill-ok {
  color: #4ade80;
}

.trim-alert-banner {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 9px 12px;
  border-radius: var(--r-md);
  font-size: 12px;
  line-height: 1.5;
  background: rgba(217, 119, 6, 0.1);
  border: 1px solid rgba(217, 119, 6, 0.35);
  color: #b45309;
}

html.dark .trim-alert-banner {
  background: rgba(245, 158, 11, 0.12);
  border-color: rgba(245, 158, 11, 0.35);
  color: #fbbf24;
}

.alert-ico {
  width: 15px;
  height: 15px;
  flex-shrink: 0;
  margin-top: 2px;
}

.trim-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-top: 4px;
  border-top: 1px solid var(--hairline);
}

.section-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
  padding-top: 6px;
}

.section-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--ink);
}

.quick-presets-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.preset-pill {
  border: 1px solid var(--hairline);
  background: var(--surface);
  color: var(--stone);
  border-radius: 999px;
  padding: 2px 9px;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.preset-pill:hover:not(:disabled) {
  border-color: var(--color-accent-fg);
  color: var(--color-accent-fg);
}

.preset-pill:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pill-rule {
  color: var(--color-accent-fg);
  background: color-mix(in srgb, var(--color-accent-fg) 10%, transparent);
  border-color: color-mix(in srgb, var(--color-accent-fg) 30%, transparent);
}

/* 胶片时间轴可视化 */
.film-timeline-wrap {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.film-strip-track {
  height: 36px;
  display: flex;
  align-items: stretch;
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid var(--hairline, #e4e4e7);
  background: var(--surface-secondary, #f6f8fa);
}

html.dark .film-strip-track {
  background: #0d1117;
  border-color: #30363d;
}

.film-seg {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  position: relative;
  overflow: hidden;
  transition: all 0.2s ease;
  font-size: 11px;
}

.seg-discarded {
  background: repeating-linear-gradient(
    -45deg,
    rgba(239, 68, 68, 0.05),
    rgba(239, 68, 68, 0.05) 6px,
    rgba(239, 68, 68, 0.12) 6px,
    rgba(239, 68, 68, 0.12) 12px
  );
  color: #cf222e;
  border-right: 1px dashed rgba(239, 68, 68, 0.35);
  border-left: 1px dashed rgba(239, 68, 68, 0.35);
}

html.dark .seg-discarded {
  color: #f87171;
}

.seg-retained {
  background: linear-gradient(135deg, rgba(9, 105, 218, 0.12), rgba(16, 185, 129, 0.14));
  border: 1px solid rgba(9, 105, 218, 0.3);
  color: #0969da;
  position: relative;
}

html.dark .seg-retained {
  background: linear-gradient(135deg, rgba(56, 139, 253, 0.18), rgba(63, 185, 80, 0.18));
  border-color: rgba(56, 139, 253, 0.4);
  color: #58a6ff;
}

.retained-content {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  z-index: 2;
  padding: 0 8px;
}

.seg-text-bold {
  font-weight: 600;
  font-size: 11.5px;
  white-space: nowrap;
}

.retained-badge {
  font-size: 10.5px;
  font-family: var(--font-mono, monospace);
  padding: 1px 6px;
  border-radius: 4px;
  background: rgba(9, 105, 218, 0.15);
  color: inherit;
  white-space: nowrap;
}

.film-sprockets {
  position: absolute;
  left: 0;
  right: 0;
  height: 3px;
  background-image: radial-gradient(circle, rgba(0, 0, 0, 0.25) 1px, transparent 1px);
  background-size: 8px 3px;
}

.film-sprockets.top { top: 2px; }
.film-sprockets.bottom { bottom: 2px; }

html.dark .film-sprockets {
  background-image: radial-gradient(circle, rgba(255, 255, 255, 0.3) 1px, transparent 1px);
}

.timeline-ticks-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 4px;
  font-size: 10.5px;
  color: var(--color-fg-muted);
  font-family: var(--font-mono, monospace);
  flex-wrap: wrap;
  gap: 4px 10px;
}

.time-inputs-row {
  display: flex;
  align-items: flex-end;
  gap: 10px;
}

.time-field {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.time-field-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--charcoal);
}

.time-range-sep {
  font-size: 12px;
  color: var(--stone);
  padding-bottom: 8px;
  flex-shrink: 0;
}

.mono-input :deep(.el-input__inner) {
  font-family: var(--font-mono);
}

.trim-helper-note {
  padding: 7px 10px;
  border-radius: var(--r-sm);
  font-size: 12px;
  line-height: 1.45;
  background: var(--paper-alt, #f6f8fa);
  border: 1px solid var(--hairline);
  color: var(--charcoal);
}

html.dark .trim-helper-note {
  background: rgba(255, 255, 255, 0.03);
}

.trim-helper-note.is-invalid {
  background: rgba(239, 68, 68, 0.08);
  border-color: rgba(239, 68, 68, 0.35);
  color: #cf222e;
}

html.dark .trim-helper-note.is-invalid {
  color: #f87171;
}

.trim-helper-note.is-warn {
  background: rgba(217, 119, 6, 0.08);
  border-color: rgba(217, 119, 6, 0.35);
  color: #b45309;
}

html.dark .trim-helper-note.is-warn {
  color: #fbbf24;
}

.mode-options-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 8px;
}

.mode-card {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 12px;
  border-radius: var(--r-md);
  border: 1px solid var(--hairline);
  background: var(--surface);
  cursor: pointer;
  transition: border-color 0.15s ease, background-color 0.15s ease;
}

.mode-card input {
  margin-top: 3px;
  accent-color: var(--color-accent-fg);
}

.mode-card.active {
  border-color: var(--color-accent-fg);
  background: color-mix(in srgb, var(--color-accent-fg) 6%, transparent);
}

.mode-card.active.is-danger {
  border-color: #d97706;
  background: rgba(217, 119, 6, 0.08);
}

.mode-card.disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.mode-card-body {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
}

.mode-card-title {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  font-size: 12.5px;
  font-weight: 600;
  color: var(--ink);
}

.mode-badge-safe {
  font-size: 10.5px;
  font-weight: 500;
  padding: 1px 6px;
  border-radius: 999px;
  background: rgba(22, 163, 74, 0.12);
  color: #16a34a;
}

html.dark .mode-badge-safe {
  color: #4ade80;
}

.mode-badge-warn {
  font-size: 10.5px;
  font-weight: 500;
  padding: 1px 6px;
  border-radius: 999px;
  background: rgba(217, 119, 6, 0.14);
  color: #b45309;
}

html.dark .mode-badge-warn {
  color: #fbbf24;
}

.mode-card-desc {
  margin: 0;
  font-size: 11.5px;
  line-height: 1.45;
  color: var(--stone);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 520px) {
  .time-inputs-row {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
  .time-range-sep {
    display: none;
  }
}
</style>

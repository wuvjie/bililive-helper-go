<template>
  <el-dialog
    :model-value="visible"
    :title="`个性化规则 · ${streamer}`"
    width="560px"
    destroy-on-close
    class="rule-dialog"
    @update:model-value="$emit('update:visible', $event)"
  >
    <div class="rule-dialog-body">
      <!-- Section 1: 自动视频时段裁剪 (视音频无损裁剪) -->
      <div class="rule-section">
        <div class="rule-row-header">
          <div class="rule-title-col">
            <div class="title-with-badge">
              <svg class="rule-section-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="6" cy="6" r="3"/>
                <circle cx="6" cy="18" r="3"/>
                <line x1="20" y1="4" x2="8.12" y2="15.88"/>
                <line x1="14.47" y1="14.48" x2="20" y2="20"/>
                <line x1="8.12" y1="8.12" x2="12" y2="12"/>
              </svg>
              <span class="rule-title">
                <span>自动视频时段裁剪</span>
                <el-tooltip content="分片合并完成后，基于无损流复制截取指定起止时段保留成片" placement="top" :show-after="300">
                  <span class="label-help-ico">?</span>
                </el-tooltip>
              </span>
              <span class="status-pill" :class="formRule.trim_enabled ? 'pill-active' : 'pill-default'">
                {{ formRule.trim_enabled ? '裁剪已启用' : '未开启' }}
              </span>
            </div>
          </div>
          <el-switch v-model="formRule.trim_enabled" />
        </div>

        <transition name="rule-expand">
          <div v-if="formRule.trim_enabled" class="trim-expand-area">
            <!-- 拟真视频胶片时间轴预览条 (Film Strip Timeline Bar) -->
            <div class="film-timeline-wrap">
              <div class="film-strip-track">
                <!-- Segment 1: 起始前段剪除 -->
                <div
                  v-if="timelineSegments.hasLeading"
                  class="film-seg seg-discarded"
                  :style="{ flexGrow: timelineSegments.leadingGrow }"
                  title="起始前段剪除丢弃"
                >
                  <svg class="seg-mini-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"/>
                    <line x1="6" y1="6" x2="18" y2="18"/>
                  </svg>
                  <span class="seg-text">剪除前段</span>
                </div>

                <!-- Segment 2: 有效录像保留片段 (核心主体) -->
                <div
                  class="film-seg seg-retained"
                  :style="{ flexGrow: timelineSegments.retainedGrow }"
                  title="有效录像保留成片"
                >
                  <div class="film-sprockets top"></div>
                  <div class="retained-content">
                    <svg class="film-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"/>
                      <line x1="7" y1="2" x2="7" y2="22"/>
                      <line x1="17" y1="2" x2="17" y2="22"/>
                      <line x1="2" y1="12" x2="22" y2="12"/>
                    </svg>
                    <span class="seg-text-bold">保留片段</span>
                    <span v-if="trimSemanticInfo.durationText" class="retained-badge">
                      {{ trimSemanticInfo.durationText }}
                    </span>
                  </div>
                  <div class="film-sprockets bottom"></div>
                </div>

                <!-- Segment 3: 出点后段剪除 (若指定了 trim_end) -->
                <div
                  v-if="timelineSegments.hasTrailing"
                  class="film-seg seg-discarded"
                  :style="{ flexGrow: timelineSegments.trailingGrow }"
                  title="出点后段剪除丢弃"
                >
                  <svg class="seg-mini-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"/>
                    <line x1="6" y1="6" x2="18" y2="18"/>
                  </svg>
                  <span class="seg-text">剪除后段</span>
                </div>
                <!-- 或者未指定 trim_end，表示截取至自然结尾 -->
                <div
                  v-else-if="timelineSegments.isUntilEnd"
                  class="film-seg seg-tail-open"
                  title="截取至录播自然结束"
                >
                  <span class="seg-text">直至结尾</span>
                </div>
              </div>

              <!-- 时间轴节点指示 -->
              <div class="timeline-ticks-row">
                <span class="tick-node tick-start">00:00:00 录像起点</span>
                <span class="tick-node tick-in" v-if="timelineSegments.hasLeading">
                  入点 {{ formRule.trim_start }}
                </span>
                <span class="tick-node tick-out" v-if="timelineSegments.hasTrailing">
                  出点 {{ formRule.trim_end }}
                </span>
                <span class="tick-node tick-end">录像终点</span>
              </div>
            </div>

            <!-- 剪辑入点与剪辑出点输入行 -->
            <div class="time-inputs-row">
              <div class="time-field">
                <label class="time-field-label">
                  <span>剪辑入点（起始时间）</span>
                </label>
                <el-input
                  v-model="formRule.trim_start"
                  placeholder="00:00:00"
                  class="mono-input"
                  @blur="normalizeTrimField('trim_start')"
                />
              </div>
              <span class="time-range-sep">至</span>
              <div class="time-field">
                <label class="time-field-label">
                  <span>剪辑出点（结束时间）</span>
                </label>
                <el-input
                  v-model="formRule.trim_end"
                  placeholder="留空保留至录像末尾"
                  class="mono-input"
                  @blur="normalizeTrimField('trim_end')"
                />
              </div>
            </div>

            <!-- 语义解算轻量说明条 (Inline Helper) -->
            <div class="trim-helper-note" :class="{ 'is-invalid': trimSemanticInfo.isInvalid }">
              <svg v-if="!trimSemanticInfo.isInvalid" class="helper-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
              </svg>
              <svg v-else class="helper-ico ico-warn" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
                <line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/>
              </svg>
              <span class="helper-text">{{ trimSemanticInfo.text }}</span>
              <span v-if="trimSemanticInfo.durationText" class="helper-badge">
                {{ trimSemanticInfo.durationText }}
              </span>
            </div>
            <div class="trim-format-hint">
              * 格式支持 <code>时:分:秒</code>（如 01:30:00）或 <code>分:秒</code>（如 05:00）；出点留空自动保留至录像末尾。
            </div>
          </div>
        </transition>
      </div>

      <!-- Section 2: 存储清理保护 (最高保护特权) -->
      <div class="rule-section">
        <div class="rule-row-header">
          <div class="rule-title-col">
            <div class="title-with-badge">
              <svg class="rule-section-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
              </svg>
              <span class="rule-title">
                <span>免清理免删保护</span>
                <el-tooltip content="最高保护等级：开启后后台自动磁盘清理任务彻底跳过该主播目录，即使磁盘耗尽也绝对保留" placement="top" :show-after="300">
                  <span class="label-help-ico">?</span>
                </el-tooltip>
              </span>
              <span class="status-pill" :class="formRule.protected ? 'pill-exempt' : 'pill-default'">
                {{ formRule.protected ? '已豁免清理' : '遵循全局' }}
              </span>
            </div>
          </div>
          <el-switch v-model="formRule.protected" />
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="$emit('update:visible', false)">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { ElMessage } from "element-plus";
import type { StreamerRule } from "@/api/types";

const props = defineProps<{
  visible: boolean;
  streamer: string;
  rule?: StreamerRule;
  saving: boolean;
}>();

const emit = defineEmits<{
  (e: "update:visible", val: boolean): void;
  (e: "save", rule: StreamerRule): void;
}>();

const formRule = ref<StreamerRule>({
  trim_enabled: false,
  trim_start: "00:00:00",
  trim_end: "00:30:00",
  min_keep: 0,
  protected: false
});

const initialSnapshot = ref("");

watch(
  () => props.visible,
  (val) => {
    if (val) {
      if (props.rule) {
        formRule.value = {
          trim_enabled: !!props.rule.trim_enabled,
          trim_start: props.rule.trim_start || (props.rule as any).start_time || "00:00:00",
          trim_end: props.rule.trim_end || (props.rule as any).end_time || "00:30:00",
          min_keep: props.rule.min_keep || 0,
          protected: !!props.rule.protected
        };
      } else {
        formRule.value = {
          trim_enabled: false,
          trim_start: "00:00:00",
          trim_end: "00:30:00",
          min_keep: 0,
          protected: false
        };
      }
      initialSnapshot.value = JSON.stringify(formRule.value);
    }
  },
  { immediate: true }
);

function parseSeconds(val?: string): number | null {
  if (!val || !val.trim()) return null;
  const parts = val.trim().split(":").map(p => Number(p));
  if (parts.some(n => isNaN(n) || n < 0)) return null;
  if (parts.length === 3) return parts[0] * 3600 + parts[1] * 60 + parts[2];
  if (parts.length === 2) return parts[0] * 60 + parts[1];
  if (parts.length === 1) return parts[0];
  return null;
}

function formatFriendlyDuration(totalSec: number): string {
  if (totalSec <= 0) return "0 秒";
  const h = Math.floor(totalSec / 3600);
  const m = Math.floor((totalSec % 3600) / 60);
  const s = totalSec % 60;
  const parts: string[] = [];
  if (h > 0) parts.push(`${h} 小时`);
  if (m > 0) parts.push(`${m} 分钟`);
  if (s > 0 && h === 0) parts.push(`${s} 秒`);
  return parts.join(" ") || "0 秒";
}

function normalizeTimeString(val: string): string {
  if (!val || !val.trim()) return "";
  const cleaned = val.trim();
  const parts = cleaned.split(":").map(p => p.trim());
  if (parts.length === 3) {
    const h = String(parseInt(parts[0], 10) || 0).padStart(2, "0");
    const m = String(parseInt(parts[1], 10) || 0).padStart(2, "0");
    const s = String(parseInt(parts[2], 10) || 0).padStart(2, "0");
    return `${h}:${m}:${s}`;
  } else if (parts.length === 2) {
    const m = String(parseInt(parts[0], 10) || 0).padStart(2, "0");
    const s = String(parseInt(parts[1], 10) || 0).padStart(2, "0");
    return `00:${m}:${s}`;
  }
  return cleaned;
}

function normalizeTrimField(field: "trim_start" | "trim_end") {
  if (field === "trim_start") {
    if (!formRule.value.trim_start || !formRule.value.trim_start.trim()) {
      formRule.value.trim_start = "00:00:00";
    } else {
      formRule.value.trim_start = normalizeTimeString(formRule.value.trim_start);
    }
  } else {
    if (formRule.value.trim_end && formRule.value.trim_end.trim()) {
      formRule.value.trim_end = normalizeTimeString(formRule.value.trim_end);
    }
  }
}

const timelineSegments = computed(() => {
  const startSec = parseSeconds(formRule.value.trim_start) ?? 0;
  const endSec = parseSeconds(formRule.value.trim_end);

  const hasLeading = startSec > 0;
  const hasTrailing = endSec !== null && endSec > startSec;
  const isUntilEnd = endSec === null;

  let leadingGrow = 1;
  let retainedGrow = 3.5;
  let trailingGrow = 1;

  if (hasLeading && hasTrailing) {
    const diff = Math.max(1, endSec - startSec);
    leadingGrow = Math.max(1, Math.min(2.5, Math.round((startSec / diff) * 2)));
    retainedGrow = 4;
    trailingGrow = 1.2;
  } else if (hasLeading && isUntilEnd) {
    leadingGrow = 1.2;
    retainedGrow = 4;
  } else if (!hasLeading && hasTrailing) {
    retainedGrow = 4;
    trailingGrow = 1.8;
  }

  return {
    hasLeading,
    hasTrailing,
    isUntilEnd,
    leadingGrow,
    retainedGrow,
    trailingGrow
  };
});

const trimSemanticInfo = computed(() => {
  const startSec = parseSeconds(formRule.value.trim_start) ?? 0;
  const endSec = parseSeconds(formRule.value.trim_end);

  if (endSec !== null && endSec <= startSec) {
    return {
      isInvalid: true,
      text: `出点时间 (${formRule.value.trim_end}) 必须晚于入点时间 (${formRule.value.trim_start || "00:00:00"})`,
      durationText: "区间无效"
    };
  }

  if (startSec === 0 && endSec === null) {
    return {
      isInvalid: false,
      text: "全场保留：不执行裁剪，完整保留全场录播内容",
      durationText: "全场保留"
    };
  }

  if (startSec > 0 && endSec === null) {
    return {
      isInvalid: false,
      text: `跳过前段：自动截除前 ${formatFriendlyDuration(startSec)}，后续内容保留至录像结尾`,
      durationText: `跳过前 ${formatFriendlyDuration(startSec)}`
    };
  }

  if (startSec === 0 && endSec !== null) {
    return {
      isInvalid: false,
      text: `保留前段：仅截取前 ${formatFriendlyDuration(endSec)} 内容，后续部分自动剪除`,
      durationText: `时长 ${formatFriendlyDuration(endSec)}`
    };
  }

  if (startSec > 0 && endSec !== null) {
    const diff = endSec - startSec;
    return {
      isInvalid: false,
      text: `选定时段：从 ${formRule.value.trim_start} 截取至 ${formRule.value.trim_end}，剪除首尾多余部分`,
      durationText: `有效时长 ${formatFriendlyDuration(diff)}`
    };
  }

  return {
    isInvalid: false,
    text: "自定义裁剪时段",
    durationText: ""
  };
});

function handleSave() {
  if (formRule.value.trim_enabled && trimSemanticInfo.value.isInvalid) {
    ElMessage.warning(trimSemanticInfo.value.text);
    return;
  }
  if (JSON.stringify(formRule.value) === initialSnapshot.value) {
    ElMessage.info("未检测到规则变动，无需保存");
    emit("update:visible", false);
    return;
  }
  emit("save", formRule.value);
}
</script>

<style scoped>
.rule-dialog-body {
  display: flex;
  flex-direction: column;
}

/* GitHub Primer Dialog Section (无边框卡片，流式发丝线分隔) */
.rule-section {
  padding: 16px 0;
  border-bottom: 1px solid var(--hairline);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rule-section:last-child {
  border-bottom: none;
  padding-bottom: 4px;
}

.rule-section:first-child {
  padding-top: 4px;
}

.rule-row-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 38px;
}

.rule-title-col {
  display: flex;
  align-items: center;
  flex: 1;
}

.title-with-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.rule-title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.1px;
}

.label-help-ico {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  font-size: 10px;
  font-weight: 600;
  background: var(--hairline);
  color: var(--slate);
  cursor: help;
  transition: all 0.15s ease;
  user-select: none;
}

.label-help-ico:hover {
  background: var(--primary);
  color: #ffffff;
}

.status-pill {
  font-size: 11px;
  padding: 1px 7px;
  border-radius: var(--r-full);
  font-weight: 500;
  line-height: 1.4;
}

.pill-exempt {
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
}

.pill-default {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.05));
  color: var(--stone);
}

/* Expandable Time Trimming Area */
.trim-expand-area {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 14px;
  background: var(--surface-secondary, rgba(0, 0, 0, 0.02));
  border-radius: var(--r-sm);
  border: 1px solid var(--hairline);
  margin-top: 4px;
}

html.dark .trim-expand-area {
  background: rgba(255, 255, 255, 0.025);
  border-color: rgba(255, 255, 255, 0.08);
}

.time-inputs-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.time-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.time-field-label {
  font-size: 11px;
  color: var(--stone);
  font-weight: 500;
}

.time-range-sep {
  font-size: 12px;
  color: var(--stone);
  margin-top: 18px;
  flex-shrink: 0;
}

/* Helper Callout (优雅说明，拒绝突兀大色块) */
.trim-helper-note {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #059669;
  background: rgba(16, 185, 129, 0.08);
  padding: 6px 10px;
  border-radius: var(--r-xs);
}

html.dark .trim-helper-note {
  background: rgba(16, 185, 129, 0.15);
  color: #3fb950;
}

.trim-helper-note.is-invalid {
  color: #dc2626;
  background: rgba(239, 68, 68, 0.08);
}

html.dark .trim-helper-note.is-invalid {
  background: rgba(239, 68, 68, 0.18);
  color: #f85149;
}

.helper-ico {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}

.helper-text {
  flex: 1;
  font-size: 11.5px;
}

.helper-badge {
  font-size: 11px;
  font-family: var(--font-mono);
  font-weight: 600;
  padding: 1px 6px;
  border-radius: var(--r-xs);
  background: rgba(16, 185, 129, 0.12);
  white-space: nowrap;
}

.trim-format-hint {
  font-size: 11px;
  color: var(--stone);
  line-height: 1.4;
}

.trim-format-hint code {
  font-family: var(--font-mono);
  padding: 1px 4px;
  border-radius: var(--r-xs);
  background: var(--surface);
  border: 1px solid var(--hairline);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

/* Expand Animation */
.rule-expand-enter-active,
.rule-expand-leave-active {
  transition: all 0.2s ease;
  overflow: hidden;
}

.rule-expand-enter-from,
.rule-expand-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

.rule-section-icon {
  width: 15px;
  height: 15px;
  color: var(--color-accent-fg, #0969da);
  flex-shrink: 0;
}

html.dark .rule-section-icon {
  color: #58a6ff;
}

.pill-active {
  background: rgba(9, 105, 218, 0.1);
  color: #0969da;
}

html.dark .pill-active {
  background: rgba(56, 139, 253, 0.15);
  color: #58a6ff;
}

/* Film Strip Timeline Track */
.film-timeline-wrap {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 2px;
}

.film-strip-track {
  height: 38px;
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
}

.film-ico {
  width: 13px;
  height: 13px;
}

.seg-text-bold {
  font-weight: 600;
  font-size: 11.5px;
}

.retained-badge {
  font-size: 10.5px;
  font-family: var(--font-mono, monospace);
  padding: 1px 6px;
  border-radius: 4px;
  background: rgba(9, 105, 218, 0.15);
  color: #0969da;
  font-weight: 500;
}

html.dark .retained-badge {
  background: rgba(56, 139, 253, 0.25);
  color: #79c0ff;
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

.seg-tail-open {
  background: rgba(16, 185, 129, 0.05);
  color: #059669;
  border-left: 1px dashed rgba(16, 185, 129, 0.3);
  flex-grow: 1.5;
}

html.dark .seg-tail-open {
  color: #4ade80;
}

.seg-mini-ico {
  width: 10px;
  height: 10px;
}

.timeline-ticks-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 4px;
  font-size: 10.5px;
  color: var(--color-fg-muted, #57606a);
  font-family: var(--font-mono, monospace);
}

html.dark .timeline-ticks-row {
  color: #8b949e;
}

.tick-node {
  white-space: nowrap;
}

</style>

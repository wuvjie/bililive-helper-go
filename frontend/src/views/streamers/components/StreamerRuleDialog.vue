<template>
  <el-dialog
    :model-value="visible"
    :title="`个性化规则 · ${streamer}`"
    width="540px"
    destroy-on-close
    class="rule-dialog"
    @update:model-value="$emit('update:visible', $event)"
  >
    <div class="rule-dialog-body">
      <!-- Section 1: 自动视频时段裁剪 (GitHub Primer Row) -->
      <div class="rule-section">
        <div class="rule-row-header">
          <div class="rule-title-col">
            <span class="rule-title">
              <span>自动视频时段裁剪</span>
              <el-tooltip content="分片合并完成后，无损流复制截取指定起止时段（如去除开播暖场或只保留前半场精选）" placement="top" :show-after="300">
                <span class="label-help-ico">?</span>
              </el-tooltip>
            </span>
          </div>
          <el-switch v-model="formRule.trim_enabled" />
        </div>

        <transition name="rule-expand">
          <div v-if="formRule.trim_enabled" class="trim-expand-area">
            <div class="time-inputs-row">
              <div class="time-field">
                <label class="time-field-label">开始时间</label>
                <el-input
                  v-model="formRule.trim_start"
                  placeholder="00:00:00"
                  class="mono-input"
                  @blur="normalizeTrimField('trim_start')"
                />
              </div>
              <span class="time-range-sep">至</span>
              <div class="time-field">
                <label class="time-field-label">结束时间</label>
                <el-input
                  v-model="formRule.trim_end"
                  placeholder="留空截至录像末尾"
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
              * 格式支持 <code>时:分:秒</code>（如 00:05:00）或简写（如 5:00）；结束时间留空自动截取到结尾。
            </div>
          </div>
        </transition>
      </div>

      <!-- Section 2: 存储清理保护 (GitHub Primer Row) -->
      <div class="rule-section">
        <div class="rule-row-header">
          <div class="rule-title-col">
            <div class="title-with-badge">
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
    const s = String(parseInt(parts[2], 10) || 0).padStart(2, "0");
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

const trimSemanticInfo = computed(() => {
  const startSec = parseSeconds(formRule.value.trim_start) ?? 0;
  const endSec = parseSeconds(formRule.value.trim_end);

  if (endSec !== null && endSec <= startSec) {
    return {
      isInvalid: true,
      text: `结束时间 (${formRule.value.trim_end}) 必须晚于开始时间 (${formRule.value.trim_start || "00:00:00"})`,
      durationText: "区间无效"
    };
  }

  if (startSec === 0 && endSec === null) {
    return {
      isInvalid: false,
      text: "完整录制：不裁剪视频时段，完整保留录播内容",
      durationText: "全场保留"
    };
  }

  if (startSec > 0 && endSec === null) {
    return {
      isInvalid: false,
      text: `跳过暖场：自动跳过开播前 ${formatFriendlyDuration(startSec)}，截取至录像结束`,
      durationText: `跳过前 ${formatFriendlyDuration(startSec)}`
    };
  }

  if (startSec === 0 && endSec !== null) {
    return {
      isInvalid: false,
      text: `保留开播精选：截取开播前 ${formatFriendlyDuration(endSec)}，后续内容自动丢弃`,
      durationText: `时长 ${formatFriendlyDuration(endSec)}`
    };
  }

  if (startSec > 0 && endSec !== null) {
    const diff = endSec - startSec;
    return {
      isInvalid: false,
      text: `选定时段：从 ${formRule.value.trim_start} 截取至 ${formRule.value.trim_end}`,
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

.rule-desc {
  font-size: 12px;
  color: var(--steel);
  line-height: 1.45;
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
</style>

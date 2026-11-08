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
      <!-- Card 1: 视频时段裁剪 -->
      <div class="rule-setting-card">
        <div class="setting-card-header">
          <div class="header-title-box">
            <div class="card-title-left-inline">
              <div class="ui-icon-box icon-box-emerald" style="width:24px;height:24px;">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="6" cy="6" r="3"/><circle cx="6" cy="18" r="3"/>
                  <line x1="20" y1="4" x2="8.12" y2="15.88"/><line x1="14.47" y1="14.48" x2="20" y2="20"/><line x1="8.12" y1="8.12" x2="12" y2="12"/>
                </svg>
              </div>
              <span class="setting-card-title">自动视频时段裁剪</span>
            </div>
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
              <el-input
                v-model="formRule.trim_start"
                placeholder="00:00:00"
                class="mono-input"
                @blur="normalizeTrimField('trim_start')"
              />
            </div>
            <span class="range-arrow">至</span>
            <div class="time-box">
              <span class="time-label">结束时间</span>
              <el-input
                v-model="formRule.trim_end"
                placeholder="留空截到结束"
                class="mono-input"
                @blur="normalizeTrimField('trim_end')"
              />
            </div>
          </div>

          <!-- 实时语义解算卡片 -->
          <div class="trim-semantic-card" :class="{ 'is-invalid': trimSemanticInfo.isInvalid }">
            <div class="semantic-icon-row">
              <svg v-if="!trimSemanticInfo.isInvalid" class="semantic-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
              </svg>
              <svg v-else class="semantic-svg svg-warn" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
                <line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/>
              </svg>
              <span class="semantic-text">{{ trimSemanticInfo.text }}</span>
            </div>
            <span v-if="trimSemanticInfo.durationText" class="semantic-badge">
              {{ trimSemanticInfo.durationText }}
            </span>
          </div>

          <div class="form-tip">
            * 时间格式支持 <code>时:分:秒</code>（如 00:05:00）或简写（如 5:00 失焦自动规整）；结束时间留空表示截取到最后。
          </div>
        </div>
      </div>

      <!-- Card 2: 存储与清理保护 -->
      <div class="rule-setting-card">
        <div class="setting-card-header">
          <div class="header-title-box">
            <div class="card-title-left-inline">
              <div class="ui-icon-box icon-box-rose" style="width:24px;height:24px;">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                </svg>
              </div>
              <span class="setting-card-title">存储与清理保护</span>
            </div>
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
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="$emit('update:visible', false)">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存规则</el-button>
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
      text: "完整录制：不裁剪视频时段，从头到尾完整保留录播内容",
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
  gap: 16px;
}

.rule-setting-card {
  background: var(--surface);
  border: 1px solid var(--hairline);
  border-radius: var(--r-md);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.setting-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-title-box {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.card-title-left-inline {
  display: flex;
  align-items: center;
  gap: 8px;
}

.setting-card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--charcoal);
}

.setting-card-sub {
  font-size: 11px;
  color: var(--stone);
  margin-left: 32px;
}

.setting-card-desc {
  font-size: 12px;
  color: var(--slate);
  line-height: 1.5;
  margin: 0;
}

.trim-config-box {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-top: 10px;
  border-top: 1px dashed var(--hairline);
}

.time-range-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

.time-box {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.time-label {
  font-size: 11px;
  color: var(--stone);
}

.range-arrow {
  font-size: 12px;
  color: var(--stone);
  margin-top: 18px;
}

.trim-semantic-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-radius: var(--r-sm);
  background: rgba(16, 185, 129, 0.08);
  border: 1px solid rgba(16, 185, 129, 0.2);
}

.trim-semantic-card.is-invalid {
  background: rgba(239, 68, 68, 0.08);
  border-color: rgba(239, 68, 68, 0.2);
}

.semantic-icon-row {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--charcoal);
}

.semantic-svg {
  width: 14px;
  height: 14px;
  color: #059669;
}

.svg-warn {
  color: var(--error);
}

.semantic-badge {
  font-size: 11px;
  font-family: var(--font-mono);
  color: #047857;
  font-weight: 500;
}

.form-tip {
  font-size: 11px;
  color: var(--stone);
  line-height: 1.4;
}

.rule-item-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 8px;
}

.rule-item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.rule-item-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.rule-item-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--charcoal);
}

.rule-badge-pill {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: var(--r-full);
}

.pill-protected {
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
}

.pill-default {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.04));
  color: var(--stone);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>

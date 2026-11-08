<template>
  <div class="policy-tab">
    <el-form label-width="140px" label-position="left" class="settings-form">
      <!-- Section 1: 录制与合并规则 -->
      <div class="form-section-card">
        <div class="form-sub-header">
          <span class="sub-title-text">录制与合并规则</span>
        </div>
        <div class="section-body">
          <el-form-item label="断流重连容差">
            <div class="inline-unit-row">
              <el-input-number
                :model-value="config.GAP_MINUTES"
                :min="1" :max="120" :step="5" class="compact-counter"
                @update:model-value="updateField('GAP_MINUTES', Number($event))"
              />
              <span class="unit-text">分钟</span>
            </div>
            <div class="field-hint">同一主播在此时间差内断流重连，将被判定为同一场直播并自动合并为整段视频。</div>
          </el-form-item>

          <el-form-item label="完播静置等待">
            <div class="inline-unit-row">
              <el-input-number
                :model-value="config.MERGE_AGE_MINUTES"
                :min="5" :max="1440" :step="5" class="compact-counter"
                @update:model-value="updateField('MERGE_AGE_MINUTES', Number($event))"
              />
              <span class="unit-text">分钟</span>
            </div>
            <div class="field-hint">录制文件生成并静置满此时间后才允许合并，确保主播已彻底关播且文件写入完毕，防止误读活动切片。</div>
          </el-form-item>
        </div>
      </div>

      <!-- Section 2: 存储清理策略 -->
      <div class="form-section-card">
        <div class="form-sub-header">
          <span class="sub-title-text">磁盘自动清理策略</span>
        </div>
        <div class="section-body">
          <el-form-item label="触发清理阈值">
            <div class="slider-row">
              <el-slider
                :model-value="config.TRIGGER_THRESHOLD"
                :min="50"
                :max="99"
                :format-tooltip="formatPercent"
                class="slider-flex"
                @update:model-value="onTriggerChange"
              />
              <span class="slider-value mono-val">{{ config.TRIGGER_THRESHOLD }}%</span>
            </div>
            <div class="field-hint">
              磁盘使用率达到此警戒水位时，定时清理任务将正式启动。
              <span v-if="triggerFreeGB" class="hint-highlight-tag">
                相当于总容量 {{ totalDiskGb?.toFixed(0) }} GB 中剩余低于 <strong>{{ triggerFreeGB }} GB</strong> 时触发
              </span>
            </div>
          </el-form-item>

          <el-form-item label="目标清理阈值">
            <div class="slider-row">
              <el-slider
                :model-value="config.TARGET_THRESHOLD"
                :min="30"
                :max="89"
                :format-tooltip="formatPercent"
                class="slider-flex"
                @update:model-value="onTargetChange"
              />
              <span class="slider-value mono-val">{{ config.TARGET_THRESHOLD }}%</span>
            </div>
            <div class="field-hint">清理时将从最旧录播开始删除，直至磁盘使用率降至此目标水位。</div>
          </el-form-item>

          <!-- 空间释放冲程预期摘要标 -->
          <div class="stroke-summary-banner">
            <div class="stroke-summary-left">
              <svg class="stroke-svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
              <span class="stroke-label">单次清理冲程：</span>
              <span class="stroke-val-highlight">预计释放约 {{ bufferPct }}%</span>
              <span v-if="bufferGB" class="stroke-gb-badge">≈ {{ bufferGB }} GB</span>
            </div>
            <span class="stroke-range-text">触发警戒 {{ config.TRIGGER_THRESHOLD }}% → 清理至目标 {{ config.TARGET_THRESHOLD }}%</span>
          </div>

          <el-form-item label="新录像免删保护">
            <div class="inline-unit-row">
              <el-input-number
                :model-value="config.SAFE_DAYS || 3"
                :min="1"
                :max="365"
                :step="1"
                class="compact-counter"
                @update:model-value="onSafeDaysChange(Number($event))"
              />
              <span class="unit-text">天</span>
            </div>
            <div class="field-hint">在此保护天数内生成的最新录像享有免删保护，即使磁盘水位达到警戒线，清理任务也绝不删除保护期内的录像。</div>
          </el-form-item>

          <el-form-item label="每主播最少保留">
            <div class="inline-unit-row">
              <el-input-number
                :model-value="config.MIN_KEEP_PER_STREAMER"
                :min="1" :max="50" :step="1" class="compact-counter"
                @update:model-value="updateField('MIN_KEEP_PER_STREAMER', Number($event))"
              />
              <span class="unit-text">场</span>
            </div>
            <div class="field-hint">每个主播至少保留最近的场次数，无论磁盘空间多么紧张也不会被全删。</div>
          </el-form-item>

          <!-- 白名单关键词 Tag 交互 -->
          <el-form-item label="免清理白名单">
            <div class="whitelist-tag-group">
              <div class="tags-container">
                <el-tag
                  v-for="(tag, idx) in tagList"
                  :key="tag"
                  closable
                  effect="plain"
                  class="keyword-tag"
                  @close="removeTag(idx)"
                >
                  {{ tag }}
                </el-tag>
                <div class="add-tag-inline">
                  <el-input
                    v-if="inputTagVisible"
                    ref="tagInputRef"
                    v-model="inputTagValue"
                    size="small"
                    class="tag-inline-input"
                    placeholder="按回车添加"
                    @keyup.enter="handleTagConfirm"
                    @blur="handleTagConfirm"
                  />
                  <el-button v-else size="small" class="btn-add-tag" @click="showTagInput">
                    + 添加关键词
                  </el-button>
                </div>
              </div>
              <div class="quick-tags-row">
                <span class="quick-label">常用快捷添加:</span>
                <button
                  v-for="p in ['留存', '纪念', '高能', '生日', '勿删', '精选']"
                  :key="p"
                  type="button"
                  class="quick-preset-pill"
                  :disabled="tagList.includes(p)"
                  @click="addTag(p)"
                >
                  + {{ p }}
                </button>
              </div>
            </div>
            <div class="field-hint">包含任一白名单关键词的视频文件名或主播目录，将永远受保护不被清理。</div>
          </el-form-item>
        </div>
      </div>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick } from "vue";
import type { ConfigDTO } from "@/api/types";

const props = defineProps<{
  config: ConfigDTO;
  saving?: boolean;
  totalDiskGb?: number;
}>();

const emit = defineEmits<{
  "update:config": [val: ConfigDTO];
  save: [];
  "reset-defaults": [];
}>();

const formatPercent = (v: number | number[]) => Number(v) + "%";

function updateField<K extends keyof ConfigDTO>(field: K, val: ConfigDTO[K]) {
  emit("update:config", {
    ...props.config,
    [field]: val
  });
}

function onTriggerChange(val: number) {
  const currentTarget = props.config.TARGET_THRESHOLD || 60;
  let newTarget = currentTarget;
  if (val <= currentTarget) {
    newTarget = Math.max(30, val - 10);
  }
  emit("update:config", {
    ...props.config,
    TRIGGER_THRESHOLD: val,
    TARGET_THRESHOLD: newTarget
  });
}

function onTargetChange(val: number) {
  const currentTrigger = props.config.TRIGGER_THRESHOLD || 80;
  let newTrigger = currentTrigger;
  if (val >= currentTrigger) {
    newTrigger = Math.min(99, val + 10);
  }
  emit("update:config", {
    ...props.config,
    TARGET_THRESHOLD: val,
    TRIGGER_THRESHOLD: newTrigger
  });
}

function onSafeModeChange(val: string) {
  const patch: Partial<ConfigDTO> = { SAFE_MODE: val };
  if (val === "hours" && !props.config.SAFE_AGE_MINUTES) patch.SAFE_AGE_MINUTES = 120;
  if (val === "days" && !props.config.SAFE_DAYS) patch.SAFE_DAYS = 1;
  emit("update:config", { ...props.config, ...patch });
}

const safeAgeHours = computed(() => {
  const mins = props.config.SAFE_AGE_MINUTES ?? 120;
  return Math.max(1, Math.round(mins / 60));
});

// Buffer calculations
const triggerFreeGB = computed(() => {
  if (!props.totalDiskGb || props.totalDiskGb <= 0) return null;
  const trig = Number(props.config.TRIGGER_THRESHOLD || 80);
  return ((props.totalDiskGb * (100 - trig)) / 100).toFixed(1);
});

const bufferPct = computed(() => {
  const trig = Number(props.config.TRIGGER_THRESHOLD || 80);
  const targ = Number(props.config.TARGET_THRESHOLD || 60);
  return Math.max(0, trig - targ);
});

const bufferGB = computed(() => {
  if (!props.totalDiskGb || props.totalDiskGb <= 0) return null;
  return ((props.totalDiskGb * bufferPct.value) / 100).toFixed(1);
});

// Whitelist Tag list
const tagList = computed(() => {
  const raw = props.config.WHITELIST_KEYWORDS;
  if (!raw) return [];
  if (Array.isArray(raw)) return raw;
  return String(raw).split(/[,，]/).map(s => s.trim()).filter(Boolean);
});

const inputTagVisible = ref(false);
const inputTagValue = ref("");
const tagInputRef = ref<HTMLInputElement | null>(null);

function showTagInput() {
  inputTagVisible.value = true;
  nextTick(() => {
    tagInputRef.value?.focus();
  });
}

function handleTagConfirm() {
  const val = inputTagValue.value.trim();
  if (val && !tagList.value.includes(val)) {
    const nextList = [...tagList.value, val];
    emit("update:config", { ...props.config, WHITELIST_KEYWORDS: nextList });
  }
  inputTagVisible.value = false;
  inputTagValue.value = "";
}

function removeTag(index: number) {
  const nextList = tagList.value.filter((_, i) => i !== index);
  emit("update:config", { ...props.config, WHITELIST_KEYWORDS: nextList });
}

function addTag(keyword: string) {
  if (!tagList.value.includes(keyword)) {
    const nextList = [...tagList.value, keyword];
    emit("update:config", { ...props.config, WHITELIST_KEYWORDS: nextList });
  }
}
</script>

<style scoped>
.policy-tab {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-section-card {
  background: var(--surface);
  border: 1px solid var(--hairline);
  border-radius: var(--r-md);
  padding: 18px 20px;
  margin-bottom: 20px;
}

.section-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}
.section-title-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.btn-ghost-header {
  background: transparent;
  border: none;
  color: var(--steel, #64748b);
  font-size: 12px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: var(--r-sm);
  cursor: pointer;
  transition: all 0.15s ease;
}
.btn-ghost-header:hover {
  background: rgba(0, 0, 0, 0.04);
  color: var(--ink);
}

.old-section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--charcoal);
  padding-bottom: 12px;
  border-bottom: 1px solid var(--hairline);
  margin-bottom: 16px;
}

.section-icon {
  font-size: 16px;
}

.stroke-summary-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--surface-secondary, rgba(0, 0, 0, 0.025));
  border: 1px solid var(--hairline);
  border-radius: var(--r-sm, 6px);
  padding: 8px 14px;
  margin: 6px 0 16px 0;
  font-size: 12px;
  flex-wrap: wrap;
  gap: 8px;
}
.stroke-summary-left {
  display: flex;
  align-items: center;
  gap: 6px;
}
.stroke-svg {
  width: 14px !important;
  height: 14px !important;
  max-width: 14px !important;
  max-height: 14px !important;
  color: var(--primary, #059669);
  flex-shrink: 0;
}
.stroke-label {
  color: var(--steel);
}
.stroke-val-highlight {
  font-weight: 600;
  color: var(--charcoal);
  font-family: var(--font-mono);
}
.stroke-gb-badge {
  font-size: 11px;
  font-family: var(--font-mono);
  background: rgba(16, 185, 129, 0.1);
  color: #059669;
  padding: 1px 6px;
  border-radius: var(--r-xs);
}
.stroke-range-text {
  font-size: 11px;
  color: var(--stone);
  font-family: var(--font-mono);
}

.section-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.settings-form :deep(.el-form-item) {
  margin-bottom: 16px;
}

.settings-form :deep(.el-form-item__label) {
  color: var(--charcoal, #333);
  font-size: 13px;
  font-weight: 500;
}

.field-hint {
  width: 100%;
  font-size: 12px;
  color: var(--steel);
  line-height: 1.5;
  margin-top: 6px;
}

.compact-counter,
.compact-select {
  height: 32px !important;
  line-height: 32px !important;

  :deep(.el-input__wrapper) {
    height: 32px !important;
    box-sizing: border-box !important;
  }
}
.compact-counter {
  width: 120px;
  :deep(.el-input-number__decrease),
  :deep(.el-input-number__increase) {
    height: 30px !important;
    line-height: 30px !important;
  }
}
.compact-select {
  width: 96px;
  flex-shrink: 0;
}
.inline-unit-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  white-space: nowrap;
}

.unit-text {
  font-size: 13px;
  color: var(--slate);
  white-space: nowrap;
  flex-shrink: 0;
  display: inline-block;
}

.slider-row {
  display: flex;
  align-items: center;
  gap: 16px;
  width: 100%;
  box-sizing: border-box;
}

.slider-flex {
  flex: 1;
}

.slider-value {
  min-width: 44px;
  text-align: right;
  font-size: 14px;
}

.safe-age-group {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-wrap: nowrap;

  :deep(.el-input__wrapper),
  :deep(.el-select__wrapper) {
    height: 32px !important;
    box-sizing: border-box;
  }
  :deep(.el-input-number) {
    height: 32px;
  }
  :deep(.el-input-number__decrease),
  :deep(.el-input-number__increase) {
    height: 30px !important;
    width: 28px !important;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}

/* Buffer Zone Card */
.buffer-zone-card {
  background: var(--canvas, #fff);
  border: 1px solid var(--hairline, #eee);
  border-radius: var(--r-sm, 6px);
  padding: 14px 18px;
  margin: -4px 0 16px 0;
  width: 100%;
  box-sizing: border-box;
}
.buffer-zone-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.buffer-header-left {
  display: flex;
  align-items: center;
  gap: 6px;
}
.buffer-icon {
  font-size: 14px;
}
.buffer-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--charcoal, #333);
}
.buffer-val-pill {
  font-size: 12px;
  color: #b45309;
  background: #fffbeb;
  border: 1px solid rgba(180, 83, 9, 0.25);
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 500;
  font-family: var(--font-mono);
}
.buffer-visual-track {
  display: flex;
  height: 8px;
  background: var(--surface-hover, #f0f0f0);
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 6px;
}
.track-fill-safe {
  background: #448361;
  opacity: 0.8;
  transition: width 0.2s ease;
}
.track-fill-buffer {
  background: #b45309;
  opacity: 0.9;
  transition: width 0.2s ease;
}
.buffer-scale-labels {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--steel, #888);
}
.buffer-scale-item {
  display: inline-flex;
  align-items: baseline;
  gap: 4px;
  font-size: 12px;
}
.scale-lbl {
  font-size: 12px;
  font-weight: 500;
  color: var(--steel, #888);
}
.scale-val {
  font-size: 12px;
  font-weight: 600;
  font-family: var(--font-mono);
  color: var(--ink, #333);
}
.buffer-scale-mid .scale-val {
  color: #b45309;
}

/* Whitelist Tag Group */
.whitelist-tag-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
  box-sizing: border-box;
}
.tags-container {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  min-height: 36px;
  background: var(--canvas, #fff);
  border: 1px solid var(--hairline, #dcdfe6);
  border-radius: var(--r-sm, 4px);
  padding: 4px 8px;
}
.keyword-tag {
  font-size: 12px;
  font-weight: 500;
  background: var(--surface, #fafafa);
  border-color: var(--hairline-strong, #ddd);
  color: var(--charcoal, #333);
}
.tag-inline-input {
  width: 110px;
}
.btn-add-tag {
  border-style: dashed;
  font-size: 12px;
  color: var(--steel, #666);
}
.quick-tags-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.quick-label {
  font-size: 12px;
  color: var(--steel, #888);
}
.quick-preset-pill {
  background: var(--surface, #fafafa);
  border: 1px dashed var(--hairline, #ddd);
  color: var(--slate, #555);
  padding: 2px 7px;
  border-radius: 12px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s ease;
}
.quick-preset-pill:hover:not(:disabled) {
  border-color: #448361;
  color: #448361;
  background: #edf6f0;
}
.quick-preset-pill:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.policy-actions {
  display: flex;
  gap: 12px;
  padding-top: 8px;
  flex-wrap: wrap;
}
.btn-reset-policy {
  padding: 10px 18px;
  font-size: 13px;
  border-radius: var(--r-md);
}

@media (max-width: 768px) {
  .policy-actions {
    flex-direction: column;
    align-items: stretch;
    width: 100%;
    gap: 10px;
    padding-top: 12px;
  }
  .btn-reset-policy {
    width: 100% !important;
    height: 42px;
    font-size: 14px;
    display: flex;
    justify-content: center;
    align-items: center;
    margin-left: 0 !important;
  }
  .settings-form :deep(.el-form-item) {
    flex-direction: column;
    align-items: flex-start;
  }
  .settings-form :deep(.el-form-item__label) {
    width: 100% !important;
    margin-bottom: 6px;
    height: auto;
    line-height: 1.4;
  }
  .settings-form :deep(.el-form-item__content) {
    width: 100%;
  }
  .safe-age-group {
    width: 100%;
    overflow-x: auto;
  }
}

.section-title {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}
.buffer-svg {
  width: 14px;
  height: 14px;
  margin-right: 4px;
  vertical-align: -2px;
}
.btn-svg-sm {
  width: 13px;
  height: 13px;
  margin-right: 4px;
  vertical-align: -1px;
}


/* 阈值动态冲程指示器 (轻量集成于表单内) */
.threshold-buffer-indicator {
  margin: 4px 0 16px 140px;
  background: var(--surface, #f4f3f0);
  border: 1px solid var(--hairline, #e6e4df);
  border-radius: var(--r-sm, 6px);
  padding: 8px 12px;
  box-sizing: border-box;
}
.buffer-track-wrap {
  width: 100%;
}
.buffer-visual-track {
  width: 100%;
  height: 6px;
  background: rgba(0, 0, 0, 0.08);
  border-radius: 3px;
  display: flex;
  overflow: hidden;
}
.track-fill-safe {
  height: 100%;
  background: #10b981;
  transition: width 0.2s ease;
}
.track-fill-buffer {
  height: 100%;
  background: #f59e0b;
  transition: width 0.2s ease;
}
.buffer-track-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 6px;
  font-size: 12px;
  color: var(--steel, #787671);
}
.buffer-meta-main strong {
  color: #d97706;
  font-family: var(--font-mono);
  font-weight: 600;
}
.buffer-meta-gb {
  color: var(--slate, #5d5b54);
}
.buffer-meta-range {
  font-size: 11px;
  color: var(--stone, #a4a097);
  font-feature-settings: "tnum";
}

html.dark {
  .threshold-buffer-indicator {
    background: #141416;
    border-color: #27272a;
  }
  .buffer-visual-track {
    background: #27272a;
  }
  .track-fill-safe {
    background: #34d399;
  }
  .track-fill-buffer {
    background: #fbbf24;
  }
  .buffer-meta-main strong {
    color: #fbbf24;
  }
  .buffer-meta-gb {
    color: #d4d4d8;
  }
  .buffer-meta-range {
    color: #71717a;
  }
  .tags-container {
    background: #141416;
    border-color: #27272a;
  }
  .keyword-tag {
    background: #1f1f23;
    border-color: #27272a;
    color: #e4e4e7;
  }
  .quick-preset-pill {
    background: #18181b;
    border-color: #27272a;
    color: #a1a1aa;
    &:hover:not(:disabled) {
      background: rgba(34, 197, 94, 0.12);
      border-color: rgba(34, 197, 94, 0.3);
      color: #34d399;
    }
  }
}

.hint-highlight-tag {
  display: inline-block;
  margin-left: 6px;
  color: var(--color-warning, #d97706);
  font-feature-settings: "tnum";
}
.hint-highlight-tag strong {
  font-family: var(--font-mono);
  font-weight: 600;
}
html.dark .hint-highlight-tag {
  color: #fbbf24;
}

@media (max-width: 768px) {
  .threshold-buffer-indicator {
    margin-left: 0;
  }
}
</style>

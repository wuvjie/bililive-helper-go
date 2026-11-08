<template>
  <div class="policy-tab">
    <!-- Group 1: 录制与合并 -->
    <div class="setting-group">
      <div class="category-divider">
        <span class="category-label">录制与合并</span>
        <span class="category-line" />
      </div>

      <div class="setting-row">
        <div class="setting-info">
          <div class="setting-title">断流重连容差</div>
          <div class="setting-desc">主播在此时间差内断流重连，自动合并为整段视频</div>
        </div>
        <div class="setting-control">
          <div class="inline-unit-row">
            <el-input-number
              :model-value="config.GAP_MINUTES"
              :min="1" :max="120" :step="5" class="compact-counter"
              @update:model-value="updateField('GAP_MINUTES', Number($event))"
            />
            <span class="unit-text">分钟</span>
          </div>
        </div>
      </div>

      <div class="setting-row">
        <div class="setting-info">
          <div class="setting-title">完播静置等待</div>
          <div class="setting-desc">录播关播后的静置缓冲期，防止误读正在写入的切片</div>
        </div>
        <div class="setting-control">
          <div class="inline-unit-row">
            <el-input-number
              :model-value="config.MERGE_AGE_MINUTES"
              :min="5" :max="1440" :step="5" class="compact-counter"
              @update:model-value="updateField('MERGE_AGE_MINUTES', Number($event))"
            />
            <span class="unit-text">分钟</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Group 2: 磁盘自动清理 -->
    <div class="setting-group">
      <div class="category-divider">
        <span class="category-label">磁盘自动清理</span>
        <span class="category-line" />
      </div>

      <div class="setting-row">
        <div class="setting-info">
          <div class="setting-title">触发清理阈值</div>
          <div class="setting-desc">
            磁盘使用率达到此警戒水位时自动启动清理
            <span v-if="triggerFreeGB" class="hint-highlight-tag">
              (余量 &lt; {{ triggerFreeGB }} GB 时触发)
            </span>
          </div>
        </div>
        <div class="setting-control slider-control-wrap">
          <el-slider
            :model-value="config.TRIGGER_THRESHOLD"
            :min="50"
            :max="99"
            :format-tooltip="formatPercent"
            class="compact-slider"
            @update:model-value="onTriggerChange"
          />
          <span class="slider-value mono-val">{{ config.TRIGGER_THRESHOLD }}%</span>
        </div>
      </div>

      <div class="setting-row">
        <div class="setting-info">
          <div class="setting-title">目标清理阈值</div>
          <div class="setting-desc">按文件生成时间由旧至新清理，直至磁盘降至此目标水位</div>
        </div>
        <div class="setting-control slider-control-wrap">
          <el-slider
            :model-value="config.TARGET_THRESHOLD"
            :min="30"
            :max="89"
            :format-tooltip="formatPercent"
            class="compact-slider"
            @update:model-value="onTargetChange"
          />
          <span class="slider-value mono-val">{{ config.TARGET_THRESHOLD }}%</span>
        </div>
      </div>

      <!-- 单次冲程预期指示条 -->
      <div class="stroke-summary-banner">
        <div class="stroke-summary-left">
          <svg class="stroke-svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
          <span class="stroke-label">单次清理冲程：</span>
          <span class="stroke-val-highlight">预计释放约 {{ bufferPct }}%</span>
          <span v-if="bufferGB" class="stroke-gb-badge">≈ {{ bufferGB }} GB</span>
        </div>
        <span class="stroke-range-text">警戒水位 {{ config.TRIGGER_THRESHOLD }}% → 目标水位 {{ config.TARGET_THRESHOLD }}%</span>
      </div>

      <div class="setting-row">
        <div class="setting-info">
          <div class="setting-title">新录像免删保护</div>
          <div class="setting-desc">保护期内新录像享有免删特权，即使达到警戒线也绝对保留</div>
        </div>
        <div class="setting-control">
          <div class="inline-unit-row">
            <el-input-number
              :model-value="config.SAFE_DAYS || 3"
              :min="1" :max="365" :step="1" class="compact-counter"
              @update:model-value="onSafeDaysChange(Number($event))"
            />
            <span class="unit-text">天</span>
          </div>
        </div>
      </div>

      <div class="setting-row">
        <div class="setting-info">
          <div class="setting-title">每主播最少保留</div>
          <div class="setting-desc">每个主播强制保留的最低场次数，无论空间多么紧张绝不全删</div>
        </div>
        <div class="setting-control">
          <div class="inline-unit-row">
            <el-input-number
              :model-value="config.MIN_KEEP_PER_STREAMER"
              :min="1" :max="50" :step="1" class="compact-counter"
              @update:model-value="updateField('MIN_KEEP_PER_STREAMER', Number($event))"
            />
            <span class="unit-text">场</span>
          </div>
        </div>
      </div>

      <!-- 白名单关键词 Tag 交互 -->
      <div class="setting-row setting-row-stack">
        <div class="setting-info">
          <div class="setting-title">免清理白名单</div>
          <div class="setting-desc">匹配关键词的录像文件或主播目录受永久保护不被清理</div>
        </div>
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
      </div>
    </div>
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
}

.setting-group {
  display: flex;
  flex-direction: column;
}
.setting-group + .setting-group {
  margin-top: 24px;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 13px 0;
  border-bottom: 1px solid var(--hairline);
  gap: 20px;
}
.setting-row:last-child {
  border-bottom: none;
}

.setting-row-stack {
  flex-direction: column;
  align-items: stretch;
  gap: 12px;
}

.setting-info {
  display: flex;
  flex-direction: column;
  gap: 3px;
  flex: 1;
  min-width: 0;
}

.setting-title {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.1px;
}

.setting-desc {
  font-size: 12px;
  color: var(--steel);
  line-height: 1.45;
}

.setting-control {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  flex-shrink: 0;
}

.slider-control-wrap {
  width: 230px;
  display: inline-flex;
  align-items: center;
  gap: 12px;
  justify-content: flex-end;
}
.compact-slider {
  flex: 1;
}
.slider-value {
  font-size: 13px;
  width: 38px;
  text-align: right;
  font-weight: 600;
  color: var(--ink);
}

.hint-highlight-tag {
  display: inline-block;
  margin-left: 6px;
  padding: 0 5px;
  border-radius: var(--r-xs);
  background: color-mix(in srgb, var(--primary) 8%, transparent);
  color: var(--primary);
  font-size: 11px;
}

.stroke-summary-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--surface-secondary, rgba(0, 0, 0, 0.025));
  border: 1px solid var(--hairline);
  border-radius: var(--r-sm, 6px);
  padding: 8px 14px;
  margin: 6px 0 10px 0;
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

.compact-counter {
  width: 120px;
  height: 32px !important;
  line-height: 32px !important;
  :deep(.el-input__wrapper) {
    height: 32px !important;
    box-sizing: border-box !important;
  }
  :deep(.el-input-number__decrease),
  :deep(.el-input-number__increase) {
    height: 30px !important;
    line-height: 30px !important;
  }
}

.inline-unit-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.unit-text {
  font-size: 12px;
  color: var(--steel);
  font-weight: 500;
  white-space: nowrap;
}

/* Whitelist Tags */
.whitelist-tag-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.keyword-tag {
  font-size: 12px;
  background: var(--surface-secondary, rgba(0, 0, 0, 0.03)) !important;
  border-color: var(--hairline-strong) !important;
  color: var(--charcoal) !important;
  border-radius: var(--r-xs);
  padding: 0 8px;
  height: 26px;
  line-height: 24px;
}

.tag-inline-input {
  width: 120px;
}

.btn-add-tag {
  font-size: 12px;
  color: var(--primary) !important;
  border-color: color-mix(in srgb, var(--primary) 25%, transparent) !important;
  background: color-mix(in srgb, var(--primary) 6%, transparent) !important;
}

.quick-tags-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.quick-label {
  font-size: 11px;
  color: var(--muted);
}

.quick-preset-pill {
  padding: 2px 7px;
  font-size: 11px;
  border-radius: var(--r-xs);
  background: transparent;
  border: 1px dashed var(--hairline-strong);
  color: var(--steel);
  cursor: pointer;
  transition: all 0.15s ease;
}

.quick-preset-pill:hover:not(:disabled) {
  border-color: var(--primary);
  color: var(--primary);
  background: color-mix(in srgb, var(--primary) 5%, transparent);
}

.quick-preset-pill:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  border-style: solid;
}
</style>

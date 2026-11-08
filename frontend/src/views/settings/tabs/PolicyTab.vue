<template>
  <div class="policy-tab">
    <!-- Group 1: 录制与分片合并规则 -->
    <div class="setting-group-card">
      <div class="group-card-header">
        <div class="group-header-text">
          <h3 class="group-title">录制与分片合并规则</h3>
          <p class="group-desc">配置直播录制切片的合并容差与关播落盘静置期，确保录像完整且防并发冲突。</p>
        </div>
      </div>
      <div class="group-card-body">
        <div class="setting-row">
          <div class="setting-info">
            <div class="setting-title">
              <span>断流重连容差</span>
              <el-tooltip content="主播在此时间差内断流重连，自动合并为整段视频" placement="top" :show-after="300">
                <span class="label-help-ico">?</span>
              </el-tooltip>
            </div>
          </div>
          <div class="setting-control">
            <div class="inline-unit-row">
              <el-input-number
                :model-value="config.GAP_MINUTES"
                :min="1"
                :max="120"
                :step="5"
                class="compact-counter"
                @update:model-value="updateField('GAP_MINUTES', Number($event))"
              />
              <span class="unit-text">分钟</span>
            </div>
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-info">
            <div class="setting-title">
              <span>完播静置等待</span>
              <el-tooltip content="录播关播后的静置缓冲期，防止误读正在写入的活动切片" placement="top" :show-after="300">
                <span class="label-help-ico">?</span>
              </el-tooltip>
            </div>
          </div>
          <div class="setting-control">
            <div class="inline-unit-row">
              <el-input-number
                :model-value="config.MERGE_AGE_MINUTES"
                :min="5"
                :max="1440"
                :step="5"
                class="compact-counter"
                @update:model-value="updateField('MERGE_AGE_MINUTES', Number($event))"
              />
              <span class="unit-text">分钟</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Group 2: 磁盘自动清理与水位控制 -->
    <div class="setting-group-card">
      <div class="group-card-header">
        <div class="group-header-text">
          <h3 class="group-title">磁盘自动清理与水位控制</h3>
          <p class="group-desc">当磁盘占用触达警戒水位时自动清理旧录像，直至降至安全水位，形成有效清理冲程。</p>
        </div>
        <div class="gauge-metric-badge">
          <span class="metric-lbl">单次释放冲程:</span>
          <strong class="metric-pct">{{ bufferPct }}%</strong>
          <span v-if="bufferGB" class="metric-gb">≈ {{ bufferGB }} GB</span>
        </div>
      </div>

      <div class="group-card-body">
        <!-- Visual Watermark Track (水位深度与冲程展示条) -->
        <div class="gauge-preview-wrap">
          <div class="gauge-bar-track">
            <!-- Safe zone: 0 to Target -->
            <div class="gauge-seg seg-safe" :style="{ width: config.TARGET_THRESHOLD + '%' }">
              <span class="seg-label" v-if="config.TARGET_THRESHOLD >= 22">安全水位 (0 ~ {{ config.TARGET_THRESHOLD }}%)</span>
            </div>
            <!-- Reclaim stroke buffer zone: Target to Trigger -->
            <div class="gauge-seg seg-reclaim" :style="{ width: bufferPct + '%' }">
              <span class="seg-label" v-if="bufferPct >= 12">冲程 {{ bufferPct }}%</span>
            </div>
            <!-- Danger alert zone: Trigger to 100 -->
            <div class="gauge-seg seg-danger" :style="{ width: (100 - config.TRIGGER_THRESHOLD) + '%' }">
              <span class="seg-label" v-if="(100 - config.TRIGGER_THRESHOLD) >= 12">警戒区</span>
            </div>
          </div>
          <div class="gauge-ticks">
            <span class="tick-label tick-0">0%</span>
            <span class="tick-label tick-target" :style="{ left: config.TARGET_THRESHOLD + '%' }">
              目标 {{ config.TARGET_THRESHOLD }}%
            </span>
            <span class="tick-label tick-trigger" :style="{ left: config.TRIGGER_THRESHOLD + '%' }">
              警戒 {{ config.TRIGGER_THRESHOLD }}%
            </span>
            <span class="tick-label tick-100">100%</span>
          </div>
        </div>

        <!-- Row 1: 触发清理阈值 (警戒水位) -->
        <div class="setting-row setting-row-slider">
          <div class="setting-info">
            <div class="setting-title">
              <span class="slider-dot dot-warn"></span>
              <span>触发清理阈值</span>
              <el-tooltip content="磁盘使用率达到此警戒水位时自动启动清理" placement="top" :show-after="300">
                <span class="label-help-ico">?</span>
              </el-tooltip>
              <span v-if="triggerFreeGB" class="hint-free-gb">余量 &lt; {{ triggerFreeGB }} GB 触发</span>
            </div>
          </div>
          <div class="setting-control slider-control-col">
            <div class="slider-with-val">
              <el-slider
                :model-value="config.TRIGGER_THRESHOLD"
                :min="50"
                :max="99"
                :format-tooltip="formatPercent"
                class="compact-slider"
                @update:model-value="onTriggerChange"
              />
              <span class="slider-val mono-val warn-val">{{ config.TRIGGER_THRESHOLD }}%</span>
            </div>
          </div>
        </div>

        <!-- Row 2: 目标清理阈值 (安全水位) -->
        <div class="setting-row setting-row-slider">
          <div class="setting-info">
            <div class="setting-title">
              <span class="slider-dot dot-target"></span>
              <span>目标清理阈值</span>
              <el-tooltip content="按由旧至新淘汰录像，直至磁盘降回此安全水位" placement="top" :show-after="300">
                <span class="label-help-ico">?</span>
              </el-tooltip>
            </div>
          </div>
          <div class="setting-control slider-control-col">
            <div class="slider-with-val">
              <el-slider
                :model-value="config.TARGET_THRESHOLD"
                :min="30"
                :max="89"
                :format-tooltip="formatPercent"
                class="compact-slider"
                @update:model-value="onTargetChange"
              />
              <span class="slider-val mono-val target-val">{{ config.TARGET_THRESHOLD }}%</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Group 3: 安全保留防线 -->
    <div class="setting-group-card">
      <div class="group-card-header">
        <div class="group-header-text">
          <h3 class="group-title">安全保留防线</h3>
          <p class="group-desc">即使磁盘空间达到警戒水位，以下三项防御限制仍受硬性保护，绝不被超额清理。</p>
        </div>
      </div>
      <div class="group-card-body">
        <div class="setting-row">
          <div class="setting-info">
            <div class="setting-title">
              <span>新录像免删保护</span>
              <el-tooltip content="保护期内新录像享有免删特权，即使达到警戒线也绝对保留" placement="top" :show-after="300">
                <span class="label-help-ico">?</span>
              </el-tooltip>
            </div>
          </div>
          <div class="setting-control">
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
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-info">
            <div class="setting-title">
              <span>每主播最少保留</span>
              <el-tooltip content="每个主播强制保留的最低场次数，无论空间多么紧张绝不全删" placement="top" :show-after="300">
                <span class="label-help-ico">?</span>
              </el-tooltip>
            </div>
          </div>
          <div class="setting-control">
            <div class="inline-unit-row">
              <el-input-number
                :model-value="config.MIN_KEEP_PER_STREAMER"
                :min="1"
                :max="50"
                :step="1"
                class="compact-counter"
                @update:model-value="updateField('MIN_KEEP_PER_STREAMER', Number($event))"
              />
              <span class="unit-text">场</span>
            </div>
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-info">
            <div class="setting-title">
              <span>单次最多清理</span>
              <el-tooltip content="每次清理任务允许删除的文件上限，防止极端突发异常大量删除" placement="top" :show-after="300">
                <span class="label-help-ico">?</span>
              </el-tooltip>
            </div>
          </div>
          <div class="setting-control">
            <div class="inline-unit-row">
              <el-input-number
                :model-value="config.MAX_DELETE_PER_RUN || 50"
                :min="5"
                :max="500"
                :step="10"
                class="compact-counter"
                @update:model-value="updateField('MAX_DELETE_PER_RUN', Number($event))"
              />
              <span class="unit-text">个</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Group 4: 免清理白名单 -->
    <div class="setting-group-card">
      <div class="group-card-header">
        <div class="group-header-text">
          <h3 class="group-title">免清理白名单</h3>
          <p class="group-desc">匹配关键词的录像文件或主播目录受永久保护，在任何清理任务中绝不被删除。</p>
        </div>
      </div>
      <div class="group-card-body">
        <div class="whitelist-wrap">
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
            <span class="quick-label">常用预设:</span>
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

function onSafeDaysChange(days: number) {
  emit("update:config", {
    ...props.config,
    SAFE_DAYS: days,
    SAFE_MODE: "days",
    SAFE_AGE_MINUTES: days * 1440
  });
}

const bufferPct = computed(() => {
  const trig = props.config.TRIGGER_THRESHOLD ?? 80;
  const tgt = props.config.TARGET_THRESHOLD ?? 60;
  return Math.max(0, trig - tgt);
});

const bufferGB = computed(() => {
  if (!props.totalDiskGb || props.totalDiskGb <= 0) return null;
  const pct = bufferPct.value;
  return Math.round((props.totalDiskGb * pct) / 100);
});

const triggerFreeGB = computed(() => {
  if (!props.totalDiskGb || props.totalDiskGb <= 0) return null;
  const trig = props.config.TRIGGER_THRESHOLD ?? 80;
  const freePct = Math.max(0, 100 - trig);
  return Math.round((props.totalDiskGb * freePct) / 100);
});

const tagList = computed<string[]>(() => {
  const raw = props.config.WHITELIST_KEYWORDS;
  if (!raw) return [];
  if (Array.isArray(raw)) return raw;
  if (typeof raw === "string") {
    return (raw as string)
      .split(/[,，]/)
      .map((s: string) => s.trim())
      .filter(Boolean);
  }
  return [];
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

/* Unified Group Card (GitHub Primer Section Card) */
.setting-group-card {
  background: var(--color-canvas-default, #ffffff);
  border: 1px solid var(--color-border-subtle, #d0d7de);
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.setting-group-card:hover {
  border-color: var(--color-border-default, #afb8c1);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.06);
}

html.dark .setting-group-card {
  background: var(--color-canvas-subtle, #161b22);
  border-color: #30363d;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

html.dark .setting-group-card:hover {
  border-color: #484f58;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
}

.group-card-header {
  padding: 14px 20px;
  border-bottom: 1px solid var(--color-border-subtle, #d0d7de);
  background: var(--color-canvas-subtle, #f6f8fa);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

html.dark .group-card-header {
  border-bottom-color: #30363d;
  background: rgba(255, 255, 255, 0.02);
}

.group-header-text {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.group-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-fg-default, #24292f);
  margin: 0;
}

html.dark .group-title {
  color: #f0f6fc;
}

.group-desc {
  font-size: 12px;
  color: var(--color-fg-muted, #57606a);
  margin: 0;
  line-height: 1.4;
}

html.dark .group-desc {
  color: #8b949e;
}

.group-card-body {
  padding: 4px 20px 8px;
}

/* Setting Rows */
.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 0;
  gap: 20px;
  min-height: 44px;
  box-sizing: border-box;
}

.setting-row:not(:last-child) {
  border-bottom: 1px solid var(--color-border-subtle, #d0d7de);
}

html.dark .setting-row:not(:last-child) {
  border-bottom-color: #21262d;
}

.setting-info {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
}

.setting-title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13.5px;
  font-weight: 600;
  color: var(--color-fg-default, #24292f);
}

html.dark .setting-title {
  color: #c9d1d9;
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
  background: var(--hairline, #e4e4e7);
  color: var(--steel, #71717a);
  cursor: help;
  transition: all 0.15s ease;
  user-select: none;
}

.label-help-ico:hover {
  background: var(--color-accent-fg, #0969da);
  color: #ffffff;
}

.setting-control {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  flex-shrink: 0;
}

.compact-counter {
  width: 120px;
  height: 32px !important;
  line-height: 32px !important;
}

:deep(.compact-counter .el-input__wrapper) {
  height: 32px !important;
  box-sizing: border-box !important;
}

:deep(.compact-counter .el-input-number__decrease),
:deep(.compact-counter .el-input-number__increase) {
  height: 30px !important;
  line-height: 30px !important;
}

.inline-unit-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.unit-text {
  font-size: 12px;
  color: var(--color-fg-muted, #57606a);
  font-weight: 500;
  white-space: nowrap;
  min-width: 24px;
}

html.dark .unit-text {
  color: #8b949e;
}

/* Watermark Gauge Track */
.gauge-metric-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  padding: 3px 8px;
  border-radius: 6px;
  background: rgba(9, 105, 218, 0.08);
  border: 1px solid rgba(9, 105, 218, 0.2);
}

html.dark .gauge-metric-badge {
  background: rgba(56, 139, 253, 0.12);
  border-color: rgba(56, 139, 253, 0.3);
}

.metric-lbl {
  color: var(--color-fg-muted, #57606a);
}

html.dark .metric-lbl {
  color: #8b949e;
}

.metric-pct {
  color: var(--color-accent-fg, #0969da);
  font-family: var(--font-mono, monospace);
  font-weight: 600;
}

html.dark .metric-pct {
  color: #58a6ff;
}

.metric-gb {
  color: var(--color-fg-muted, #57606a);
  font-family: var(--font-mono, monospace);
}

.gauge-preview-wrap {
  padding: 16px 0 12px;
  border-bottom: 1px solid var(--color-border-subtle, #d0d7de);
}

html.dark .gauge-preview-wrap {
  border-bottom-color: #21262d;
}

.gauge-bar-track {
  display: flex;
  height: 24px;
  border-radius: 6px;
  overflow: hidden;
  background: rgba(0, 0, 0, 0.04);
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.08);
}

html.dark .gauge-bar-track {
  background: rgba(255, 255, 255, 0.06);
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.3);
}

.gauge-seg {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 500;
  transition: width 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  overflow: hidden;
  white-space: nowrap;
}

.seg-safe {
  background: #10b981;
  color: #ffffff;
}

.seg-reclaim {
  background: #0284c7;
  color: #ffffff;
}

.seg-danger {
  background: #f59e0b;
  color: #ffffff;
}

.gauge-ticks {
  position: relative;
  height: 18px;
  margin-top: 4px;
}

.tick-label {
  position: absolute;
  top: 0;
  font-size: 11px;
  color: var(--color-fg-muted, #57606a);
  font-family: var(--font-mono, monospace);
  transform: translateX(-50%);
  white-space: nowrap;
}

.tick-0 {
  left: 0;
  transform: translateX(0);
}

.tick-100 {
  left: 100%;
  transform: translateX(-100%);
}

.tick-target {
  color: #0284c7;
  font-weight: 600;
}

.tick-trigger {
  color: #d97706;
  font-weight: 600;
}

/* Sliders within settings row */
.slider-control-col {
  width: min(340px, 45vw);
}

.slider-with-val {
  display: flex;
  align-items: center;
  gap: 14px;
  width: 100%;
}

.compact-slider {
  flex: 1;
}

.slider-val {
  font-size: 13px;
  font-weight: 600;
  min-width: 42px;
  text-align: right;
  font-family: var(--font-mono, monospace);
}

.warn-val {
  color: #d97706;
}

.target-val {
  color: #0284c7;
}

.slider-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.dot-warn {
  background: #f59e0b;
}

.dot-target {
  background: #0284c7;
}

.hint-free-gb {
  font-size: 11px;
  color: var(--color-fg-muted, #57606a);
  font-family: var(--font-mono, monospace);
  margin-left: 4px;
  font-weight: 400;
}

/* Whitelist Wrap */
.whitelist-wrap {
  padding: 14px 0 10px;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.keyword-tag {
  font-size: 12px;
}

.add-tag-inline {
  display: inline-flex;
  align-items: center;
}

.tag-inline-input {
  width: 100px;
}

.btn-add-tag {
  font-size: 12px;
}

.quick-tags-row {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.quick-label {
  font-size: 12px;
  color: var(--color-fg-muted, #57606a);
}

.quick-preset-pill {
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 11.5px;
  color: var(--color-fg-muted, #57606a);
  background: var(--color-canvas-subtle, #f6f8fa);
  border: 1px solid var(--color-border-subtle, #d0d7de);
  cursor: pointer;
  transition: all 0.15s ease;
}

.quick-preset-pill:hover:not(:disabled) {
  color: var(--color-accent-fg, #0969da);
  border-color: var(--color-accent-fg, #0969da);
  background: rgba(9, 105, 218, 0.05);
}

.quick-preset-pill:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

html.dark .quick-preset-pill {
  background: #21262d;
  border-color: #30363d;
  color: #8b949e;
}
</style>

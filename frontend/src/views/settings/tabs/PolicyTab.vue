<template>
  <div class="policy-tab">
    <!-- Group 1: 录制与合并规则 -->
    <div class="setting-group">
      <div class="subhead">
        <h3 class="subhead-title">录制与合并规则</h3>
      </div>

      <div class="policy-two-col-grid">
        <div class="setting-grid-card">
          <div class="grid-card-main">
            <div class="grid-card-info">
              <div class="setting-title">
                <span>断流重连容差</span>
                <el-tooltip content="主播在此时间差内断流重连，自动合并为整段视频" placement="top" :show-after="300">
                  <span class="label-help-ico">?</span>
                </el-tooltip>
              </div>
              <div class="setting-desc">偶发网络抖动断流时，此容差内的切片自动无缝合并为同一场直播。</div>
            </div>
            <div class="grid-card-control">
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
        </div>

        <div class="setting-grid-card">
          <div class="grid-card-main">
            <div class="grid-card-info">
              <div class="setting-title">
                <span>完播静置等待</span>
                <el-tooltip content="录播关播后的静置缓冲期，防止误读正在写入的活动切片" placement="top" :show-after="300">
                  <span class="label-help-ico">?</span>
                </el-tooltip>
              </div>
              <div class="setting-desc">关播后的观察静置期，确保文件完全落盘，防止误读取写入中的切片。</div>
            </div>
            <div class="grid-card-control">
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
      </div>
    </div>

    <!-- Group 2: 磁盘自动清理与保护策略 -->
    <div class="setting-group">
      <div class="subhead">
        <h3 class="subhead-title">磁盘自动清理策略</h3>
      </div>

      <!-- 一体化水位深度与清理冲程控制台 (Unified Watermark & Stroke Gauge) -->
      <div class="watermark-gauge-card">
        <div class="gauge-header">
          <div class="gauge-title-col">
            <div class="gauge-card-title">
              <span>水位深度与清理冲程</span>
              <el-tooltip content="当磁盘触达警戒水位时触发清理，自动清除陈旧录像直至降至目标水位" placement="top" :show-after="300">
                <span class="label-help-ico">?</span>
              </el-tooltip>
            </div>
          </div>
          <div class="gauge-metric-badge">
            <span class="metric-lbl">单次释放冲程:</span>
            <strong class="metric-pct">{{ bufferPct }}%</strong>
            <span v-if="bufferGB" class="metric-gb">≈ {{ bufferGB }} GB</span>
          </div>
        </div>

        <!-- Multi-segment visual watermark track -->
        <div class="gauge-track-wrap">
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

        <!-- Dual Threshold Sliders (Grid) -->
        <div class="gauge-sliders-grid">
          <div class="gauge-slider-cell">
            <div class="slider-cell-header">
              <div class="slider-cell-title">
                <span class="slider-dot dot-warn"></span>
                <span>触发清理阈值</span>
                <el-tooltip content="磁盘使用率达到此警戒水位时自动启动清理" placement="top" :show-after="300">
                  <span class="label-help-ico">?</span>
                </el-tooltip>
                <span v-if="triggerFreeGB" class="hint-free-gb">余量 &lt; {{ triggerFreeGB }} GB 触发</span>
              </div>
              <span class="slider-val mono-val warn-val">{{ config.TRIGGER_THRESHOLD }}%</span>
            </div>
            <el-slider
              :model-value="config.TRIGGER_THRESHOLD"
              :min="50"
              :max="99"
              :format-tooltip="formatPercent"
              class="compact-slider"
              @update:model-value="onTriggerChange"
            />
          </div>

          <div class="gauge-slider-cell">
            <div class="slider-cell-header">
              <div class="slider-cell-title">
                <span class="slider-dot dot-target"></span>
                <span>目标清理阈值</span>
                <el-tooltip content="按由旧至新淘汰录像，直至磁盘降回此安全水位" placement="top" :show-after="300">
                  <span class="label-help-ico">?</span>
                </el-tooltip>
              </div>
              <span class="slider-val mono-val target-val">{{ config.TARGET_THRESHOLD }}%</span>
            </div>
            <el-slider
              :model-value="config.TARGET_THRESHOLD"
              :min="30"
              :max="89"
              :format-tooltip="formatPercent"
              class="compact-slider"
              @update:model-value="onTargetChange"
            />
          </div>
        </div>
      </div>

      <div class="subhead subhead-inset">
        <h4 class="subhead-subtitle">安全保留防线</h4>
        <p class="subhead-desc">即使磁盘空间达到警戒水位，以下三项防御限制仍受硬性保护。</p>
      </div>

      <div class="policy-three-col-grid">
        <div class="setting-grid-card">
          <div class="grid-card-main">
            <div class="grid-card-info">
              <div class="setting-title">
                <span>免删保护天数</span>
                <el-tooltip content="保护期内新录像享有免删特权，即使达到警戒线也绝对保留" placement="top" :show-after="300">
                  <span class="label-help-ico">?</span>
                </el-tooltip>
              </div>
              <div class="setting-desc">保护期内新录像豁免清理。</div>
            </div>
            <div class="grid-card-control">
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
        </div>

        <div class="setting-grid-card">
          <div class="grid-card-main">
            <div class="grid-card-info">
              <div class="setting-title">
                <span>每主播底线保留</span>
                <el-tooltip content="每个主播强制保留的最低场次数，无论空间多么紧张绝不全删" placement="top" :show-after="300">
                  <span class="label-help-ico">?</span>
                </el-tooltip>
              </div>
              <div class="setting-desc">每位主播强制保留场次数。</div>
            </div>
            <div class="grid-card-control">
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
        </div>

        <div class="setting-grid-card">
          <div class="grid-card-main">
            <div class="grid-card-info">
              <div class="setting-title">
                <span>单次清理上限</span>
                <el-tooltip content="每次清理任务允许删除的文件上限，防止极端突发异常大量删除" placement="top" :show-after="300">
                  <span class="label-help-ico">?</span>
                </el-tooltip>
              </div>
              <div class="setting-desc">单次巡检最多清理文件数。</div>
            </div>
            <div class="grid-card-control">
              <div class="inline-unit-row">
                <el-input-number
                  :model-value="config.MAX_DELETE_PER_RUN || 50"
                  :min="5" :max="500" :step="10" class="compact-counter"
                  @update:model-value="updateField('MAX_DELETE_PER_RUN', Number($event))"
                />
                <span class="unit-text">个</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 白名单关键词 Tag 交互 -->
      <div class="setting-row setting-row-stack">
        <div class="setting-info">
          <div class="setting-title">
            <span>免清理白名单</span>
            <el-tooltip content="匹配关键词的录像文件或主播目录受永久保护不被自动清理" placement="top" :show-after="300">
              <span class="label-help-ico">?</span>
            </el-tooltip>
          </div>
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

.setting-group {
  display: flex;
  flex-direction: column;
}
.setting-group + .setting-group {
  margin-top: 28px;
}

/* GitHub Primer Subhead (子分类标题与描述) */
.subhead {
  margin-bottom: 8px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--hairline);
}

.subhead-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.2px;
}

/* Setting Rows - Single Line Compact Layout */
.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
  gap: 20px;
  min-height: 40px;
  box-sizing: border-box;
}

.setting-row-stack {
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
}

.setting-info {
  display: flex;
  align-items: center;
  min-width: 0;
}

.setting-title {
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
  min-width: 28px;
  text-align: left;
}

/* ========================================================
   Precision Watermark Gauge Controller (一体化水位深度与冲程仪)
   ======================================================== */
.watermark-gauge-card {
  margin: 8px 0 14px 0;
  padding: 14px 16px;
  border-radius: var(--r-md);
  background: var(--surface-secondary, rgba(0, 0, 0, 0.02));
  border: 1px solid var(--hairline);
  display: flex;
  flex-direction: column;
  gap: 14px;
}

html.dark .watermark-gauge-card {
  background: rgba(255, 255, 255, 0.025);
  border-color: rgba(255, 255, 255, 0.08);
}

.gauge-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.gauge-title-col {
  display: flex;
  align-items: center;
}

.gauge-card-title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.1px;
}

.gauge-metric-badge {
  display: inline-flex;
  align-items: baseline;
  gap: 6px;
  padding: 4px 10px;
  border-radius: var(--r-sm);
  background: color-mix(in srgb, #3b82f6 10%, transparent);
  border: 1px solid color-mix(in srgb, #3b82f6 20%, transparent);
  white-space: nowrap;
  flex-shrink: 0;
}

.metric-lbl {
  font-size: 11px;
  color: #2563eb;
  font-weight: 500;
}

html.dark .metric-lbl {
  color: #58a6ff;
}

.metric-pct {
  font-size: 15px;
  font-weight: 700;
  color: #1d4ed8;
  font-family: var(--font-mono);
}

html.dark .metric-pct {
  color: #58a6ff;
}

.metric-gb {
  font-size: 11.5px;
  font-weight: 600;
  color: #2563eb;
  font-family: var(--font-mono);
}

html.dark .metric-gb {
  color: #79c0ff;
}

/* Track & Bar */
.gauge-track-wrap {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.gauge-bar-track {
  height: 20px;
  border-radius: 4px;
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--hairline);
  overflow: hidden;
  display: flex;
  position: relative;
}

html.dark .gauge-bar-track {
  background: #0d1117;
  border-color: #30363d;
}

.gauge-seg {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: width 0.15s ease;
  overflow: hidden;
  white-space: nowrap;
}

.seg-label {
  font-size: 10px;
  font-weight: 600;
  letter-spacing: -0.2px;
  user-select: none;
}

.seg-safe {
  background: color-mix(in srgb, #10b981 12%, transparent);
  color: #059669;
  border-right: 1px dashed rgba(16, 185, 129, 0.4);
}

html.dark .seg-safe {
  background: rgba(16, 185, 129, 0.18);
  color: #3fb950;
}

.seg-reclaim {
  background: #3b82f6;
  color: #ffffff;
}

html.dark .seg-reclaim {
  background: #1f6feb;
}

.seg-danger {
  background: color-mix(in srgb, #d97706 15%, transparent);
  color: #b45309;
}

html.dark .seg-danger {
  background: rgba(245, 158, 11, 0.18);
  color: #d29922;
}

.gauge-ticks {
  position: relative;
  height: 14px;
  font-size: 10px;
  font-family: var(--font-mono);
  color: var(--stone);
}

.tick-label {
  position: absolute;
  top: 0;
  line-height: 1;
  white-space: nowrap;
}

.tick-0 {
  left: 0;
}

.tick-100 {
  right: 0;
}

.tick-target {
  transform: translateX(-50%);
  color: #059669;
  font-weight: 600;
}

html.dark .tick-target {
  color: #3fb950;
}

.tick-trigger {
  transform: translateX(-50%);
  color: #b45309;
  font-weight: 600;
}

html.dark .tick-trigger {
  color: #d29922;
}

/* Sliders inside gauge */
.gauge-sliders-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  padding-top: 8px;
  border-top: 1px solid var(--hairline);
}

@media (max-width: 680px) {
  .gauge-sliders-grid {
    grid-template-columns: 1fr;
  }
}

.gauge-slider-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.slider-cell-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.slider-cell-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12.5px;
  font-weight: 600;
  color: var(--charcoal);
}

.slider-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.dot-warn {
  background: #f59e0b;
}

.dot-target {
  background: #10b981;
}

.slider-val {
  font-size: 13px;
  font-weight: 700;
}

.warn-val {
  color: #b45309;
}

html.dark .warn-val {
  color: #d29922;
}

.target-val {
  color: #059669;
}

html.dark .target-val {
  color: #3fb950;
}

.hint-free-gb {
  display: inline-block;
  margin-left: 4px;
  padding: 0 4px;
  border-radius: var(--r-xs);
  background: rgba(245, 158, 11, 0.12);
  color: #b45309;
  font-family: var(--font-mono);
  font-size: 10.5px;
}

html.dark .hint-free-gb {
  color: #d29922;
}

.compact-slider {
  margin-top: -2px;
}

/* Whitelist Tag Group */
.whitelist-tag-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 2px;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.keyword-tag {
  font-size: 12px;
  padding: 3px 8px;
  border-radius: var(--r-xs);
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

.policy-two-col-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
  margin-top: 12px;
  margin-bottom: 8px;

  @media (max-width: 640px) {
    grid-template-columns: 1fr;
  }
}

.policy-three-col-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
  margin-top: 10px;
  margin-bottom: 8px;

  @media (max-width: 768px) {
    grid-template-columns: 1fr;
  }
}

.setting-grid-card {
  background: var(--color-canvas-subtle, #f6f8fa);
  border: 1px solid var(--color-border-subtle, #d0d7de);
  border-radius: 8px;
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;

  &:hover {
    border-color: var(--color-border-default, #afb8c1);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  }
}

html.dark .setting-grid-card {
  background: rgba(255, 255, 255, 0.02);
  border-color: #30363d;

  &:hover {
    border-color: #484f58;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  }
}

.grid-card-main {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.grid-card-info {
  display: flex;
  flex-direction: column;
  gap: 4px;

  .setting-desc {
    font-size: 12px;
    color: var(--color-fg-muted, #57606a);
    line-height: 1.4;
  }
}

html.dark .grid-card-info .setting-desc {
  color: #8b949e;
}

.grid-card-control {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  padding-top: 4px;
}

.subhead-inset {
  margin-top: 22px;
  margin-bottom: 6px;
  padding-bottom: 0;
  border-bottom: none;
}

.subhead-subtitle {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-fg-default, #24292f);
  margin: 0;
}

html.dark .subhead-subtitle {
  color: #c9d1d9;
}

</style>

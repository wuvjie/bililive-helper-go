<template>
  <div class="settings-tab-pane">
    <!-- Box 1: 录播合并规则 -->
    <section class="primer-box">
      <header class="primer-box-header">
        <div class="box-header-title-row">
          <svg class="box-header-icon icon-blue" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect width="18" height="18" x="3" y="3" rx="2"/>
            <path d="M7 3v18"/>
            <path d="M17 3v18"/>
            <path d="M3 7.5h4"/>
            <path d="M3 12h18"/>
            <path d="M3 16.5h4"/>
            <path d="M17 7.5h4"/>
            <path d="M17 16.5h4"/>
          </svg>
          <h2 class="box-title">录播合并规则</h2>
        </div>
      </header>

      <div class="primer-box-body">
        <div class="fields-grid fields-grid-2">
          <!-- 1. 断流重连合并时间 -->
          <div class="field-cell">
            <label class="field-label">断流重连合并时间</label>
            <div class="field-control">
              <el-input-number
                :model-value="config.GAP_MINUTES ?? 30"
                :min="1"
                :max="1440"
                :step="5"
                size="default"
                class="stepper-compact"
                @update:model-value="updateField('GAP_MINUTES', Number($event))"
              />
              <span class="unit-text">分钟</span>
            </div>
          </div>

          <!-- 2. 关播后等待合并时长 -->
          <div class="field-cell">
            <label class="field-label">关播后等待合并时长</label>
            <div class="field-control">
              <el-input-number
                :model-value="config.MERGE_AGE_MINUTES ?? 30"
                :min="1"
                :max="1440"
                :step="5"
                size="default"
                class="stepper-compact"
                @update:model-value="updateField('MERGE_AGE_MINUTES', Number($event))"
              />
              <span class="unit-text">分钟</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Box 2: 磁盘自动清理与保护 -->
    <section class="primer-box">
      <header class="primer-box-header">
        <div class="box-header-title-row">
          <svg class="box-header-icon icon-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
            <path d="M12 8v4"/>
            <path d="M12 16h.01"/>
          </svg>
          <h2 class="box-title">磁盘自动清理与保护</h2>
        </div>
      </header>

      <div class="primer-box-body">
        <!-- 1. 水位深度与清理冲程控制台 -->
        <div class="gauge-section">
          <div class="gauge-console">
            <!-- 仪表盘顶栏读数 -->
            <div class="gauge-header">
              <span class="gauge-label">磁盘清理警戒线与安全回落线</span>
              <span class="gauge-summary">
                触发清理 <strong class="text-warn">{{ triggerVal }}%</strong>
                <span v-if="triggerFreeGB" class="text-sub">（余量 &lt; {{ triggerFreeGB }} GB 时启动）</span>
                · 降至安全线 <strong class="text-target">{{ targetVal }}%</strong>
                · 单次释放 <strong class="text-blue">{{ strokePct }}%</strong>
                <span v-if="strokeGB" class="text-sub">（≈ {{ strokeGB }} GB）</span>
              </span>
            </div>

            <!-- 三色连续标尺槽轨 -->
            <div class="gauge-track-wrap">
              <div class="gauge-track">
                <div class="track-segment seg-safe" :style="{ width: `${targetVal}%` }">
                  <span class="seg-label" v-if="targetVal > 15">安全水位 (0% ~ {{ targetVal }}%)</span>
                </div>
                <div class="track-segment seg-stroke" :style="{ width: `${strokePct}%` }">
                  <span class="seg-label" v-if="strokePct > 12">清理释放冲程 ({{ strokePct }}%)</span>
                </div>
                <div class="track-segment seg-warn" :style="{ width: `${100 - triggerVal}%` }">
                  <span class="seg-label" v-if="(100 - triggerVal) > 15">警戒水位 ({{ triggerVal }}% ~ 100%)</span>
                </div>
              </div>
            </div>

            <!-- 双滑块控制器 -->
            <div class="gauge-sliders-grid">
              <div class="slider-field">
                <div class="slider-meta">
                  <span class="slider-title">
                    <span class="dot-indicator dot-warn"></span>
                    达到警戒线触发清理
                  </span>
                  <span class="slider-val mono-val">{{ triggerVal }}%</span>
                </div>
                <el-slider
                  :model-value="triggerVal"
                  :min="40"
                  :max="95"
                  :step="1"
                  @update:model-value="onTriggerChange"
                />
              </div>

              <div class="slider-field">
                <div class="slider-meta">
                  <span class="slider-title">
                    <span class="dot-indicator dot-target"></span>
                    降回安全线停止清理
                  </span>
                  <span class="slider-val mono-val">{{ targetVal }}%</span>
                </div>
                <el-slider
                  :model-value="targetVal"
                  :min="20"
                  :max="triggerVal - 5"
                  :step="1"
                  @update:model-value="onTargetChange"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- 2. 清理与保护三参数网格 -->
        <div class="fields-grid fields-grid-3">
          <!-- 字段 1: 最近录像免删保护期 -->
          <div class="field-cell">
            <label class="field-label">最近录像免删保护期</label>
            <div class="field-control">
              <el-input-number
                :model-value="config.SAFE_DAYS ?? 3"
                :min="1"
                :max="365"
                :step="1"
                size="default"
                class="stepper-compact"
                @update:model-value="updateSafeDays(Number($event))"
              />
              <span class="unit-text">天</span>
            </div>
          </div>

          <!-- 字段 2: 每位主播保底留存场次 -->
          <div class="field-cell">
            <label class="field-label">每位主播保底留存场次</label>
            <div class="field-control">
              <el-input-number
                :model-value="config.MIN_KEEP_PER_STREAMER ?? 3"
                :min="1"
                :max="50"
                :step="1"
                size="default"
                class="stepper-compact"
                @update:model-value="updateField('MIN_KEEP_PER_STREAMER', Number($event))"
              />
              <span class="unit-text">场录像</span>
            </div>
          </div>

          <!-- 字段 3: 单次清理最多删除文件 -->
          <div class="field-cell">
            <label class="field-label">单次清理最多删除文件</label>
            <div class="field-control">
              <el-input-number
                :model-value="config.MAX_DELETE_PER_RUN ?? 50"
                :min="5"
                :max="500"
                :step="10"
                size="default"
                class="stepper-compact"
                @update:model-value="updateField('MAX_DELETE_PER_RUN', Number($event))"
              />
              <span class="unit-text">个</span>
            </div>
          </div>
        </div>

        <!-- 3. 永久免删关键词白名单 -->
        <div class="whitelist-section">
          <label class="field-label">永久免删关键词白名单</label>
          <div class="tags-flex-wrap">
            <el-tag
              v-for="(tag, idx) in tagList"
              :key="tag"
              closable
              effect="plain"
              class="primer-keyword-tag"
              @close="removeTag(idx)"
            >
              {{ tag }}
            </el-tag>

            <div class="tag-input-inline">
              <el-input
                v-if="inputTagVisible"
                ref="tagInputRef"
                v-model="inputTagValue"
                size="small"
                class="tag-editor-input"
                placeholder="回车确认"
                @keyup.enter="handleTagConfirm"
                @blur="handleTagConfirm"
              />
              <button
                v-else
                type="button"
                class="btn-add-tag"
                @click="showTagInput"
              >
                <svg class="plus-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="12" y1="5" x2="12" y2="19"/>
                  <line x1="5" y1="12" x2="19" y2="12"/>
                </svg>
                <span>添加关键词</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>
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
  (e: "update:config", val: ConfigDTO): void;
  (e: "save"): void;
}>();

function updateField<K extends keyof ConfigDTO>(key: K, value: ConfigDTO[K]) {
  emit("update:config", {
    ...props.config,
    [key]: value
  });
}

// 水位计算
const triggerVal = computed(() => Number(props.config.TRIGGER_THRESHOLD ?? 80));
const targetVal = computed(() => Number(props.config.TARGET_THRESHOLD ?? 60));
const strokePct = computed(() => Math.max(0, triggerVal.value - targetVal.value));

const triggerFreeGB = computed(() => {
  if (!props.totalDiskGb || props.totalDiskGb <= 0) return null;
  const freePct = 100 - triggerVal.value;
  return ((props.totalDiskGb * freePct) / 100).toFixed(1);
});

const strokeGB = computed(() => {
  if (!props.totalDiskGb || props.totalDiskGb <= 0) return null;
  return ((props.totalDiskGb * strokePct.value) / 100).toFixed(1);
});

function onTriggerChange(val: any) {
  const n = Number(val);
  const next: ConfigDTO = { ...props.config, TRIGGER_THRESHOLD: n };
  if (targetVal.value >= n) {
    next.TARGET_THRESHOLD = Math.max(20, n - 5);
  }
  emit("update:config", next);
}

function onTargetChange(val: any) {
  const n = Number(val);
  const next: ConfigDTO = { ...props.config, TARGET_THRESHOLD: n };
  if (n >= triggerVal.value) {
    next.TRIGGER_THRESHOLD = Math.min(95, n + 5);
  }
  emit("update:config", next);
}

function updateSafeDays(days: number) {
  emit("update:config", {
    ...props.config,
    SAFE_DAYS: days,
    SAFE_MODE: "days",
    SAFE_AGE_MINUTES: days * 1440
  });
}

// 白名单标签
const tagList = computed<string[]>(() => {
  const raw = props.config.WHITELIST_KEYWORDS;
  if (!raw) return [];
  if (Array.isArray(raw)) return raw;
  if (typeof raw === "string") {
    return (raw as string).split(/[,，]/).map(s => s.trim()).filter(Boolean);
  }
  return [];
});

const inputTagVisible = ref(false);
const inputTagValue = ref("");
const tagInputRef = ref();

function showTagInput() {
  inputTagVisible.value = true;
  nextTick(() => {
    tagInputRef.value?.input?.focus();
  });
}

function handleTagConfirm() {
  const text = inputTagValue.value.trim();
  if (text && !tagList.value.includes(text)) {
    const nextList = [...tagList.value, text];
    updateField("WHITELIST_KEYWORDS", nextList);
  }
  inputTagValue.value = "";
  inputTagVisible.value = false;
}

function removeTag(idx: number) {
  const nextList = tagList.value.filter((_, i) => i !== idx);
  updateField("WHITELIST_KEYWORDS", nextList);
}
</script>

<style scoped>
.settings-tab-pane {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* GitHub Primer Box Model */
.primer-box {
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--hairline, #e5e3df);
  border-radius: var(--r-md, 6px);
  overflow: hidden;
  box-shadow: var(--shadow-xs);
}

.primer-box-header {
  padding: 12px 18px;
  background: var(--surface-secondary, rgba(0, 0, 0, 0.02));
  border-bottom: 1px solid var(--hairline, #e5e3df);
}

.box-header-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.box-header-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.icon-blue {
  color: var(--primary, #2563eb);
}

.icon-amber {
  color: var(--warning, #d97706);
}

.box-title {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--ink, #1a1a1a);
  margin: 0;
  line-height: 1.3;
}

.primer-box-body {
  padding: 0;
}

/* Fields Grid: 模块化网格，告别单项占整行与无效空白 */
.fields-grid {
  display: grid;
  padding: 16px 18px;
  gap: 24px;
}

.fields-grid-2 {
  grid-template-columns: repeat(2, 1fr);
}

.fields-grid-3 {
  grid-template-columns: repeat(3, 1fr);
  border-bottom: 1px solid var(--hairline, #e5e3df);
}

.field-cell {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-label {
  font-size: 13.5px;
  font-weight: 500;
  color: var(--charcoal, #37352f);
  line-height: 1.35;
}

.field-control {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.control-unit-group {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.stepper-compact {
  width: 130px;
}

.unit-text {
  font-size: 13px;
  color: var(--slate, #5d5b54);
  font-weight: 500;
}

/* 水位深度控制台专有样式 */
.gauge-section {
  padding: 16px 18px;
  border-bottom: 1px solid var(--hairline, #e5e3df);
}

.gauge-console {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.gauge-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.gauge-label {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal, #37352f);
}

.gauge-summary {
  font-size: 12px;
  color: var(--steel, #787671);
}

.text-warn {
  color: var(--warning, #d97706);
  font-weight: 600;
}

.text-target {
  color: var(--success, #059669);
  font-weight: 600;
}

.text-blue {
  color: var(--primary, #2563eb);
  font-weight: 600;
}

.text-sub {
  color: var(--steel, #787671);
  font-weight: normal;
}

.gauge-track-wrap {
  width: 100%;
}

.gauge-track {
  height: 18px;
  width: 100%;
  border-radius: var(--r-sm, 4px);
  display: flex;
  overflow: hidden;
  background: var(--surface, #f4f3f0);
  border: 1px solid var(--hairline, #e5e3df);
}

.track-segment {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: width 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

.seg-safe {
  background: rgba(16, 185, 129, 0.18);
  border-right: 1px solid rgba(16, 185, 129, 0.4);
}

.seg-stroke {
  background: rgba(37, 99, 235, 0.22);
  border-right: 1px solid rgba(37, 99, 235, 0.4);
}

.seg-warn {
  background: rgba(245, 158, 11, 0.22);
}

.seg-label {
  font-size: 10.5px;
  font-weight: 600;
  white-space: nowrap;
  letter-spacing: 0.2px;
  padding: 0 4px;
}

.seg-safe .seg-label { color: var(--success, #059669); }
.seg-stroke .seg-label { color: var(--primary, #2563eb); }
.seg-warn .seg-label { color: var(--warning, #d97706); }

.gauge-sliders-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  background: var(--surface-secondary, rgba(0, 0, 0, 0.02));
  border: 1px solid var(--hairline, #e5e3df);
  border-radius: var(--r-md, 6px);
  padding: 12px 16px;
}

.slider-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.slider-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.slider-title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12.5px;
  font-weight: 500;
  color: var(--charcoal, #37352f);
}

.slider-val {
  font-size: 13px;
  font-weight: 600;
  color: var(--ink, #1a1a1a);
}

.dot-indicator {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.dot-warn {
  background: var(--warning, #d97706);
}

.dot-target {
  background: var(--success, #059669);
}

/* 白名单区域 */
.whitelist-section {
  padding: 16px 18px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tags-flex-wrap {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.primer-keyword-tag {
  border-radius: var(--r-sm, 4px) !important;
  font-size: 12px !important;
  padding: 2px 8px !important;
  background: var(--surface, #f4f3f0) !important;
  border-color: var(--hairline-strong, #c8c4be) !important;
  color: var(--ink, #1a1a1a) !important;
  font-weight: 500;
}

.tag-input-inline {
  display: inline-flex;
  align-items: center;
}

.tag-editor-input {
  width: 100px;
}

.btn-add-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 26px;
  padding: 0 10px;
  border-radius: var(--r-sm, 4px);
  border: 1px dashed var(--hairline-strong, #c8c4be);
  background: transparent;
  color: var(--slate, #5d5b54);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-add-tag:hover {
  border-color: var(--primary, #2563eb);
  color: var(--primary, #2563eb);
  background: color-mix(in srgb, var(--primary, #2563eb) 6%, transparent);
}

.plus-icon {
  width: 12px;
  height: 12px;
}

@media (max-width: 768px) {
  .fields-grid-2,
  .fields-grid-3 {
    grid-template-columns: 1fr;
    gap: 16px;
  }
  .gauge-sliders-grid {
    grid-template-columns: 1fr;
  }
}
</style>

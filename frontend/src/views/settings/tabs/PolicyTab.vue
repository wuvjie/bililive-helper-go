<template>
  <div class="settings-tab-pane">
    <!-- Box 1: 录播合并策略 -->
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
          <div class="box-header-text">
            <h2 class="box-title">录播合并策略</h2>
            <p class="box-desc">定义录播分片的断流容差与关播静置缓冲时间，确保视频无损合流</p>
          </div>
        </div>
      </header>

      <div class="primer-box-body">
        <!-- 1. 断流重连容差 -->
        <div class="primer-row">
          <div class="row-info">
            <label class="row-label">断流重连容差</label>
            <p class="row-desc">主播在此时间差内断流并重连，系统自动识别为同场直播并合并为单段完整录像</p>
          </div>
          <div class="row-control">
            <div class="control-unit-group">
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
        </div>

        <!-- 2. 完播静置等待 -->
        <div class="primer-row">
          <div class="row-info">
            <label class="row-label">完播静置等待</label>
            <p class="row-desc">录播关播后的静置缓冲期，防止误读录播工具正在写入的活动分片</p>
          </div>
          <div class="row-control">
            <div class="control-unit-group">
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

    <!-- Box 2: 磁盘清理与空间安全策略 -->
    <section class="primer-box">
      <header class="primer-box-header">
        <div class="box-header-title-row">
          <svg class="box-header-icon icon-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
            <path d="M12 8v4"/>
            <path d="M12 16h.01"/>
          </svg>
          <div class="box-header-text">
            <h2 class="box-title">磁盘清理与空间安全策略</h2>
            <p class="box-desc">自动化监控磁盘容量水位，按由旧至新淘汰录像并保护核心数据</p>
          </div>
        </div>
      </header>

      <div class="primer-box-body">
        <!-- 1. 水位深度与清理冲程控制台 -->
        <div class="primer-row gauge-row">
          <div class="gauge-console">
            <!-- 仪表盘顶栏读数 -->
            <div class="gauge-header">
              <div class="gauge-title-col">
                <span class="gauge-label">磁盘清理冲程与水位控制</span>
                <span class="gauge-summary">
                  触发警戒线 <strong class="text-warn">{{ triggerVal }}%</strong>
                  <span v-if="triggerFreeGB" class="text-sub">（余量 &lt; {{ triggerFreeGB }} GB 触发）</span>
                  · 降至安全线 <strong class="text-target">{{ targetVal }}%</strong>
                  · 单次冲程 <strong class="text-blue">{{ strokePct }}%</strong>
                  <span v-if="strokeGB" class="text-sub">（≈ {{ strokeGB }} GB）</span>
                </span>
              </div>
            </div>

            <!-- 三色连续标尺槽轨 -->
            <div class="gauge-track-wrap">
              <div class="gauge-track">
                <div class="track-segment seg-safe" :style="{ width: `${targetVal}%` }">
                  <span class="seg-label" v-if="targetVal > 15">安全水位区 (0% ~ {{ targetVal }}%)</span>
                </div>
                <div class="track-segment seg-stroke" :style="{ width: `${strokePct}%` }">
                  <span class="seg-label" v-if="strokePct > 12">清理释放冲程 ({{ strokePct }}%)</span>
                </div>
                <div class="track-segment seg-warn" :style="{ width: `${100 - triggerVal}%` }">
                  <span class="seg-label" v-if="(100 - triggerVal) > 15">警戒水位区 ({{ triggerVal }}% ~ 100%)</span>
                </div>
              </div>
            </div>

            <!-- 双滑块控制器 -->
            <div class="gauge-sliders-grid">
              <div class="slider-field">
                <div class="slider-meta">
                  <span class="slider-title">
                    <span class="dot-indicator dot-warn"></span>
                    触发清理阈值
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
                    目标安全阈值
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

        <!-- 2. 新录像免删保护 -->
        <div class="primer-row">
          <div class="row-info">
            <label class="row-label">新录像免删保护</label>
            <p class="row-desc">保护期内的新鲜录像享有免删特权，即使磁盘空间紧迫也绝不提前清理</p>
          </div>
          <div class="row-control">
            <div class="safe-mode-group">
              <el-radio-group
                :model-value="safeMode"
                size="small"
                class="mode-radio-group"
                @update:model-value="onSafeModeChange"
              >
                <el-radio-button label="days">按天保护</el-radio-button>
                <el-radio-button label="hours">按小时保护</el-radio-button>
              </el-radio-group>

              <div class="control-unit-group">
                <el-input-number
                  v-if="safeMode === 'days'"
                  :model-value="config.SAFE_DAYS ?? 3"
                  :min="1"
                  :max="365"
                  :step="1"
                  size="default"
                  class="stepper-compact"
                  @update:model-value="updateField('SAFE_DAYS', Number($event))"
                />
                <el-input-number
                  v-else
                  :model-value="safeHours"
                  :min="1"
                  :max="720"
                  :step="1"
                  size="default"
                  class="stepper-compact"
                  @update:model-value="onSafeHoursChange(Number($event))"
                />
                <span class="unit-text">{{ safeMode === 'days' ? '天' : '小时' }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 3. 每主播最少保留 -->
        <div class="primer-row">
          <div class="row-info">
            <label class="row-label">每主播最少保留</label>
            <p class="row-desc">无论磁盘空间多紧张，每位主播必定至少保留最近 N 场录播不被淘汰</p>
          </div>
          <div class="row-control">
            <div class="control-unit-group">
              <el-input-number
                :model-value="config.MIN_KEEP_PER_STREAMER ?? 3"
                :min="1"
                :max="50"
                :step="1"
                size="default"
                class="stepper-compact"
                @update:model-value="updateField('MIN_KEEP_PER_STREAMER', Number($event))"
              />
              <span class="unit-text">场</span>
            </div>
          </div>
        </div>

        <!-- 4. 单次最多清理 -->
        <div class="primer-row">
          <div class="row-info">
            <label class="row-label">单次最多清理</label>
            <p class="row-desc">每次清理任务允许删除的文件数量上限，防止极端异常情况下瞬间误删过多文件</p>
          </div>
          <div class="row-control">
            <div class="control-unit-group">
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

        <!-- 5. 免清理白名单关键词 -->
        <div class="primer-row whitelist-row">
          <div class="row-info">
            <label class="row-label">免清理永久白名单关键词</label>
            <p class="row-desc">文件名包含任一关键词的录播将享受免删豁免，不受任何自动清理规则影响</p>
          </div>
          <div class="whitelist-tags-area">
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

// 免删模式
const safeMode = computed(() => props.config.SAFE_MODE || "days");
const safeHours = computed(() => Math.round((props.config.SAFE_AGE_MINUTES ?? 120) / 60) || 1);

function onSafeModeChange(val: any) {
  updateField("SAFE_MODE", val as string);
}

function onSafeHoursChange(hours: number) {
  updateField("SAFE_AGE_MINUTES", hours * 60);
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
  gap: 20px;
}

/* GitHub Primer Box Model */
.primer-box {
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--hairline, #e5e3df);
  border-radius: var(--r-md, 6px);
  overflow: hidden;
  box-shadow: var(--shadow-xs);
  transition: border-color 0.15s ease;
}

.primer-box-header {
  padding: 14px 20px;
  background: var(--surface-secondary, rgba(0, 0, 0, 0.02));
  border-bottom: 1px solid var(--hairline, #e5e3df);
}

.box-header-title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.box-header-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.icon-blue {
  color: var(--primary, #2563eb);
}

.icon-amber {
  color: var(--warning, #d97706);
}

.box-header-text {
  flex: 1;
  min-width: 0;
}

.box-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--ink, #1a1a1a);
  margin: 0 0 2px 0;
  line-height: 1.3;
}

.box-desc {
  font-size: 12px;
  color: var(--steel, #787671);
  margin: 0;
  line-height: 1.4;
}

/* Primer Rows */
.primer-box-body {
  padding: 0;
}

.primer-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--hairline, #e5e3df);
  gap: 24px;
}

.primer-row:last-child {
  border-bottom: none;
}

.row-info {
  flex: 1;
  min-width: 0;
}

.row-label {
  display: block;
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal, #37352f);
  margin-bottom: 3px;
  line-height: 1.35;
}

.row-desc {
  font-size: 12px;
  color: var(--steel, #787671);
  margin: 0;
  line-height: 1.45;
}

.row-control {
  flex-shrink: 0;
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
  min-width: 28px;
}

/* 水位深度控制台专有样式 */
.gauge-row {
  display: block;
  padding: 18px 20px;
}

.gauge-console {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.gauge-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.gauge-title-col {
  display: flex;
  flex-direction: column;
  gap: 4px;
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
  height: 20px;
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
  padding: 14px 18px;
}

.slider-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
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

/* 免删保护模式切换 */
.safe-mode-group {
  display: inline-flex;
  align-items: center;
  gap: 12px;
}

/* 白名单区域 */
.whitelist-row {
  flex-direction: column;
  align-items: flex-start;
  gap: 12px;
}

.whitelist-tags-area {
  width: 100%;
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
  padding: 3px 10px !important;
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

@media (max-width: 640px) {
  .primer-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  .gauge-sliders-grid {
    grid-template-columns: 1fr;
  }
  .safe-mode-group {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>

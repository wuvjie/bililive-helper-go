<template>
  <div class="settings-tab-pane">
    <!-- Box 1: 录像合并规则 -->
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
          <h2 class="box-title">录像合并规则</h2>
        </div>
      </header>

      <div class="primer-box-body">
        <div class="rule-cards-grid">
          <!-- 模块 1: 断流重连合并等待 -->
          <div class="rule-card">
            <div class="rule-card-header">
              <span class="rule-card-title">断流重连合并等待</span>
            </div>
            <p class="rule-card-desc">
              直播异常中断或掉线时，在此时间内重新开播将自动合并为同一场录像
            </p>
            <div class="rule-card-control">
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
              <span v-if="gapHours" class="interval-hint mono-val">（约 {{ gapHours }} 小时）</span>
            </div>
          </div>

          <!-- 模块 2: 主播关播延迟合并 -->
          <div class="rule-card">
            <div class="rule-card-header">
              <span class="rule-card-title">主播关播延迟合并</span>
            </div>
            <p class="rule-card-desc">
              主播关播后静置等待此时间，确认录制彻底结束不再开播后再启动合并
            </p>
            <div class="rule-card-control">
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
              <span v-if="mergeAgeHours" class="interval-hint mono-val">（约 {{ mergeAgeHours }} 小时）</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Box 2: 磁盘空间自动清理与保护 -->
    <section class="primer-box">
      <header class="primer-box-header">
        <div class="box-header-title-row">
          <svg class="box-header-icon icon-amber" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
            <path d="M12 8v4"/>
            <path d="M12 16h.01"/>
          </svg>
          <h2 class="box-title">磁盘空间自动清理与保护</h2>
        </div>
      </header>

      <div class="primer-box-body">
        <!-- 区域 1: 磁盘容量清理阈值控制器 -->
        <div class="gauge-section">
          <div class="gauge-console">
            <!-- 仪表盘顶栏读数 -->
            <div class="gauge-header">
              <span class="gauge-label">磁盘空间自动清理阈值</span>
              <span class="gauge-summary">
                使用率达到 <strong class="text-warn">{{ triggerVal }}%</strong> 启动清理
                <span v-if="triggerFreeGB" class="text-sub">（剩余空间不足 {{ triggerFreeGB }} GB 时启动）</span>
                · 降至 <strong class="text-target">{{ targetVal }}%</strong> 停止清理
                · 计划释放约 <strong class="text-blue">{{ strokePct }}%</strong>
                <span v-if="strokeGB" class="text-sub">（约 {{ strokeGB }} GB 空间）</span>
              </span>
            </div>

            <!-- 三色连续标尺槽轨 -->
            <div class="gauge-track-wrap">
              <div class="gauge-track">
                <div class="track-segment seg-safe" :style="{ width: `${targetVal}%` }">
                  <span class="seg-label" v-if="targetVal > 15">安全使用区间 (0% ~ {{ targetVal }}%)</span>
                </div>
                <div class="track-segment seg-stroke" :style="{ width: `${strokePct}%` }">
                  <span class="seg-label" v-if="strokePct > 12">计划释放空间 ({{ strokePct }}%)</span>
                </div>
                <div class="track-segment seg-warn" :style="{ width: `${100 - triggerVal}%` }">
                  <span class="seg-label" v-if="(100 - triggerVal) > 15">超额清理区间 ({{ triggerVal }}% ~ 100%)</span>
                </div>
              </div>
            </div>

            <!-- 双滑块控制器 -->
            <div class="gauge-sliders-grid">
              <div class="slider-field">
                <div class="slider-meta">
                  <span class="slider-title">
                    <span class="dot-indicator dot-warn"></span>
                    磁盘使用率达到此值时启动清理
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
                    磁盘使用率降至此值时停止清理
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

        <!-- 区域 2: 录像免删与保留保护规则 -->
        <div class="protection-section">
          <div class="section-sub-header">
            <span class="sub-header-title">录像保留与免删保护规则</span>
            <span class="sub-header-desc">磁盘清理时系统将优先保留符合以下规则的录像文件</span>
          </div>

          <!-- 三微卡保护规则网格 -->
          <div class="protect-cards-grid">
            <!-- 保护卡 1: 新录像免删保护天数 -->
            <div class="protect-card">
              <span class="protect-card-title">新录像免删保护天数</span>
              <p class="protect-card-desc">生成未满此天数的新录像绝对不被删除</p>
              <div class="protect-card-control">
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

            <!-- 保护卡 2: 每位主播最少保留场次 -->
            <div class="protect-card">
              <span class="protect-card-title">每位主播最少保留场次</span>
              <p class="protect-card-desc">空间紧张时为每位主播保留的最近场次数</p>
              <div class="protect-card-control">
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

            <!-- 保护卡 3: 单次清理最多删除文件数 -->
            <div class="protect-card">
              <span class="protect-card-title">单次清理最多删除文件</span>
              <p class="protect-card-desc">限制单次清理的文件数，防止长时间卡顿</p>
              <div class="protect-card-control">
                <el-input-number
                  :model-value="config.MAX_DELETE_PER_RUN ?? 50"
                  :min="5"
                  :max="500"
                  :step="10"
                  size="default"
                  class="stepper-compact"
                  @update:model-value="updateField('MAX_DELETE_PER_RUN', Number($event))"
                />
                <span class="unit-text">个文件</span>
              </div>
            </div>
          </div>

          <!-- 白名单关键词卡片 -->
          <div class="whitelist-card">
            <div class="whitelist-header">
              <span class="whitelist-title">永久免删关键词（白名单）</span>
              <span class="whitelist-desc">当录像标题或主播名称包含以下任意关键词时，清理任务将始终跳过</span>
            </div>
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
  (e: "save"): void;
}>();

function updateField<K extends keyof ConfigDTO>(key: K, value: ConfigDTO[K]) {
  emit("update:config", {
    ...props.config,
    [key]: value
  });
}

// 录像合并小时数换算提示
const gapHours = computed(() => {
  const m = Number(props.config.GAP_MINUTES ?? 30);
  if (m >= 60) return (m / 60).toFixed(1);
  return null;
});

const mergeAgeHours = computed(() => {
  const m = Number(props.config.MERGE_AGE_MINUTES ?? 30);
  if (m >= 60) return (m / 60).toFixed(1);
  return null;
});

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

/* Box 1: 录像合并规则对称双卡网格 */
.rule-cards-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  padding: 16px 18px;
}

.rule-card {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.015));
  border: 1px solid var(--hairline, #e5e3df);
  border-radius: var(--r-md, 6px);
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  transition: all 0.15s ease;
}

.rule-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.rule-card-title {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--ink, #1a1a1a);
}

.rule-card-desc {
  font-size: 12px;
  color: var(--steel, #787671);
  line-height: 1.45;
  margin: 0;
  min-height: 34px;
  display: flex;
  align-items: center;
}

.rule-card-control {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 2px;
}

.stepper-compact {
  width: 130px;
}

.unit-text {
  font-size: 13px;
  color: var(--slate, #5d5b54);
  font-weight: 500;
}

.interval-hint {
  font-size: 12px;
  color: var(--steel, #787671);
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

/* 区域 2: 录像免删与保护规则专有样式 */
.protection-section {
  padding: 16px 18px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-sub-header {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.sub-header-title {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal, #37352f);
}

.sub-header-desc {
  font-size: 12px;
  color: var(--steel, #787671);
}

.protect-cards-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
}

.protect-card {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.015));
  border: 1px solid var(--hairline, #e5e3df);
  border-radius: var(--r-md, 6px);
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  transition: all 0.15s ease;
}

.protect-card-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--ink, #1a1a1a);
}

.protect-card-desc {
  font-size: 11.5px;
  color: var(--steel, #787671);
  line-height: 1.4;
  min-height: 32px;
  display: flex;
  align-items: center;
  margin: 0;
}

.protect-card-control {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 2px;
}

/* 保护白名单微卡 */
.whitelist-card {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.015));
  border: 1px solid var(--hairline, #e5e3df);
  border-radius: var(--r-md, 6px);
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.whitelist-header {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.whitelist-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--ink, #1a1a1a);
}

.whitelist-desc {
  font-size: 11.5px;
  color: var(--steel, #787671);
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
  .rule-cards-grid,
  .gauge-sliders-grid,
  .protect-cards-grid {
    grid-template-columns: 1fr;
  }
}
</style>

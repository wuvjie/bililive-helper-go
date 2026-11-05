<template>
  <div class="policy-tab">
    <el-form label-width="160px" label-position="left" class="settings-form">
      <!-- Section 1: 录制与合并规则 -->
      <div class="form-section-card">
        <div class="section-title">
          <div class="ui-icon-box icon-box-emerald"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg></div>
          <span>录制与合并规则</span>
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
        <div class="section-title">
          <div class="ui-icon-box icon-box-amber"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/></svg></div>
          <span>磁盘自动清理策略</span>
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
            <div class="field-hint">磁盘使用率达到此警戒水位时，定时清理任务将正式启动。</div>
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

          <!-- 动态空间释放模拟条 (Buffer Zone Live Simulator) -->
          <div class="buffer-zone-card">
            <div class="buffer-zone-header">
              <div class="buffer-header-left">
                <svg class="buffer-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/></svg>
                <span class="buffer-title">空间释放缓冲区间</span>
              </div>
              <span class="buffer-val-pill">
                {{ bufferPct }}% 冲程深度
                <template v-if="bufferGB">（单次预计释放约 {{ bufferGB }} GB）</template>
              </span>
            </div>
            <div class="buffer-visual-track">
              <div class="track-fill-safe" :style="{ width: config.TARGET_THRESHOLD + '%' }" title="目标安全基础水位"></div>
              <div class="track-fill-buffer" :style="{ width: bufferPct + '%' }" title="每次清理释放的空间缓冲"></div>
            </div>
            <div class="buffer-scale-labels">
              <span class="buffer-scale-item">
                <span class="scale-lbl">目标保底:</span>
                <span class="scale-val">{{ config.TARGET_THRESHOLD }}%</span>
              </span>
              <span class="buffer-scale-item buffer-scale-mid">
                <span class="scale-lbl">释放冲程:</span>
                <span class="scale-val">{{ bufferPct }}%</span>
              </span>
              <span class="buffer-scale-item">
                <span class="scale-lbl">触发上限:</span>
                <span class="scale-val">{{ config.TRIGGER_THRESHOLD }}%</span>
              </span>
            </div>
          </div>

          <el-form-item label="新录像免删保护">
            <div class="safe-age-group">
              <el-select
                :model-value="config.SAFE_MODE"
                class="compact-select"
                style="width: 92px; flex-shrink: 0;"
                @update:model-value="onSafeModeChange"
              >
                <el-option label="按小时" value="hours" />
                <el-option label="按天" value="days" />
              </el-select>
              <div v-if="config.SAFE_MODE === 'hours'" class="inline-unit-row">
                <el-input-number
                  :model-value="safeAgeHours"
                  :min="1" :max="168" :step="1" class="compact-counter"
                  @update:model-value="updateField('SAFE_AGE_MINUTES', (Number($event) || 1) * 60)"
                />
              </div>
              <div v-if="config.SAFE_MODE === 'days'" class="inline-unit-row">
                <el-input-number
                  :model-value="config.SAFE_DAYS"
                  :min="1"
                  :max="365"
                  class="compact-counter"
                  @update:model-value="updateField('SAFE_DAYS', Number($event))"
                />
              </div>
            </div>
            <div class="field-hint">在此保护期内生成的最新录像享有免删保护，即使磁盘水位达到警戒线，清理任务也绝不会删除保护期内的文件。</div>
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

          <el-form-item label="单次最大删除数">
            <div class="inline-unit-row">
              <el-input-number
                :model-value="config.MAX_DELETE_PER_RUN || 50"
                :min="5" :max="500" :step="10" class="compact-counter"
                @update:model-value="updateField('MAX_DELETE_PER_RUN', Number($event))"
              />
              <span class="unit-text">个</span>
            </div>
            <div class="field-hint">限制单次清理任务允许删除的文件上限，避免引起磁盘长时间高负荷 I/O 抖动（默认 50 个）。</div>
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

      <!-- 辅助操作：仅保留次级恢复预设操作，保存由外层浮动保存栏权威统领 -->
      <div class="policy-actions">
        <el-button class="btn-reset-policy" @click="emit('reset-defaults')">
          <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>
          <span>恢复推荐预设</span>
        </el-button>
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
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 32px;
  flex-wrap: nowrap;

  :deep(.el-input__wrapper) {
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


html.dark {
  .buffer-zone-card {
    background: #18181b;
    border-color: #27272a;
  }
  .buffer-val-pill {
    background: rgba(245, 158, 11, 0.12);
    color: #e5a034;
    border-color: rgba(245, 158, 11, 0.25);
  }
  .buffer-visual-track {
    background: #27272a;
  }
  .track-fill-safe {
    background: #10b981;
  }
  .track-fill-buffer {
    background: #e5a034;
  }
  .buffer-scale-labels {
    color: #a1a1aa;
    .scale-lbl {
      color: #a1a1aa !important;
    }
    .scale-val {
      color: #f4f4f5 !important;
    }
    .buffer-scale-mid .scale-val {
      color: #e5a034 !important;
    }
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
</style>

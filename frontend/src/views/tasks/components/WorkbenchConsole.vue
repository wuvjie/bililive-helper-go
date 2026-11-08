<template>
  <div class="workbench-console-card">
    <!-- Action Bar: 手动执行录像合并与清理 -->
    <div class="workbench-action-bar">
      <div class="action-bar-left">
        <div class="ui-icon-box icon-box-blue">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/>
          </svg>
        </div>
        <span class="action-bar-title">手动执行录像合并与清理</span>

        <div class="scope-selector-wrap">
          <span class="scope-label">处理范围：</span>
          <el-select
            :model-value="selectedStreamer || 'all'"
            class="scope-select"
            filterable
            placeholder="所有主播"
            @update:model-value="$emit('update:selectedStreamer', $event)"
          >
            <el-option label="所有主播" value="all" />
            <el-option
              v-for="s in streamers"
              :key="s.name"
              :label="`${s.name} (${s.files} 个录像 · ${s.size_gb?.toFixed(1)} GB)`"
              :value="s.name"
            />
          </el-select>
        </div>
      </div>

      <div class="action-buttons-group">
        <button
          type="button"
          class="btn-action btn-action-merge"
          :disabled="isRunning"
          :title="mergeButtonTooltip"
          @click="$emit('execute-merge')"
        >
          <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect width="18" height="18" x="3" y="3" rx="2"/>
            <path d="M7 3v18"/>
            <path d="M17 3v18"/>
            <path d="M3 7.5h4"/>
            <path d="M3 12h18"/>
            <path d="M3 16.5h4"/>
            <path d="M17 7.5h4"/>
            <path d="M17 16.5h4"/>
          </svg>
          <span>立即合并录像</span>
        </button>

        <button
          type="button"
          class="btn-action btn-action-clean"
          :disabled="isRunning"
          :title="cleanButtonTooltip"
          @click="$emit('execute-clean')"
        >
          <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"/>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
          </svg>
          <span>立即清理录像</span>
        </button>
      </div>
    </div>

    <!-- Integrated Terminal -->
    <div class="workbench-terminal-wrap">
      <ReTerminal
        :lines="lines"
        :title="title"
        :empty-text="emptyText"
        height="480px"
        @clear="$emit('clear')"
      >
        <template #title>
          <div class="term-title-wrapper">
            <span class="tab-indicator dot-live" :class="{ 'pulse-dot': isRunning }" />
            <span class="term-title-text">实时执行终端</span>
            <span v-if="isRunning" class="term-badge-tag tag-running">任务执行中...</span>
            <span v-else class="term-badge-tag tag-idle">就绪待命</span>
          </div>
        </template>
      </ReTerminal>

      <!-- Reconnect bar if SSE errored -->
      <div v-if="hasError && !isRunning" class="retry-bar">
        <button class="ops-action ops-retry" @click="$emit('retry')">
          <svg class="ops-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          <span>重新连接</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { StreamerInfo } from "@/api/types";
import type { SSELine } from "@/utils/sse";
import ReTerminal from "@/components/ReTerminal/index.vue";

const props = defineProps<{
  streamers: StreamerInfo[];
  selectedStreamer: string;
  isRunning: boolean;
  hasError: boolean;
  lines: SSELine[];
  title: string;
  emptyText: string;
}>();

defineEmits<{
  (e: "update:selectedStreamer", val: string): void;
  (e: "execute-merge"): void;
  (e: "execute-clean"): void;
  (e: "clear"): void;
  (e: "retry"): void;
}>();

const mergeButtonTooltip = computed(() => {
  if (props.selectedStreamer && props.selectedStreamer !== "all") {
    return `立即对主播「${props.selectedStreamer}」执行录像分片合并`;
  }
  return "立即对所有主播执行录像分片合并";
});

const cleanButtonTooltip = computed(() => {
  if (props.selectedStreamer && props.selectedStreamer !== "all") {
    return `立即清理主播「${props.selectedStreamer}」的已合并及过期录像`;
  }
  return "立即检查磁盘空间并对所有主播执行过期录像清理";
});
</script>

<style scoped>
.workbench-console-card {
  border-radius: var(--r-md);
  border: 1px solid var(--hairline);
  background: var(--surface-card, #ffffff);
  box-shadow: var(--shadow-card);
  overflow: hidden;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;
}

.workbench-console-card:hover {
  box-shadow: var(--shadow-card-hover);
}

html.dark .workbench-console-card {
  background: #161b22;
  border-color: #30363d;
}

.workbench-action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 18px;
  background: var(--surface-secondary, rgba(0, 0, 0, 0.02));
  border-bottom: 1px solid var(--hairline);
  flex-wrap: wrap;
  gap: 12px;
}

html.dark .workbench-action-bar {
  background: #161b22;
  border-bottom: 1px solid #30363d;
}

.action-bar-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  flex: 1;
}

.action-bar-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--charcoal);
  white-space: nowrap;
}

.scope-selector-wrap {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.scope-label {
  font-size: 13px;
  color: var(--steel, #787671);
  white-space: nowrap;
  font-weight: 500;
}

.scope-select {
  width: 260px;
}

.action-buttons-group {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin-left: auto;
}

.btn-action {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  padding: 0 14px;
  box-sizing: border-box;
  border-radius: var(--r-sm, 4px);
  font-size: 12.5px;
  font-weight: 500;
  border: 1px solid transparent;
  cursor: pointer;
  transition: all 0.15s ease;
  line-height: 1;
  box-shadow: var(--shadow-xs);
}

.btn-action-merge {
  background: color-mix(in srgb, #059669 12%, transparent);
  color: #047857;
  border-color: color-mix(in srgb, #059669 30%, transparent);
}

.btn-action-merge:hover:not(:disabled) {
  background: #059669;
  border-color: #059669;
  color: #ffffff;
}

.btn-action-clean {
  background: color-mix(in srgb, #d97706 12%, transparent);
  color: #b45309;
  border-color: color-mix(in srgb, #d97706 30%, transparent);
}

.btn-action-clean:hover:not(:disabled) {
  background: #d97706;
  border-color: #d97706;
  color: #ffffff;
}

.btn-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-svg-sm {
  width: 13px;
  height: 13px;
}

.workbench-terminal-wrap {
  position: relative;
  width: 100%;
}

.workbench-terminal-wrap :deep(.re-terminal) {
  border: none !important;
  border-radius: 0 0 var(--r-md) var(--r-md) !important;
}

.workbench-terminal-wrap :deep(.terminal-header) {
  border-top: none !important;
  border-radius: 0 !important;
}

.term-title-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tab-indicator {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.dot-live {
  background: #10b981;
}

.pulse-dot {
  box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7);
  animation: livePulse 1.8s infinite;
}

@keyframes livePulse {
  0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7); }
  70% { transform: scale(1); box-shadow: 0 0 0 6px rgba(16, 185, 129, 0); }
  100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(16, 185, 129, 0); }
}

.term-title-text {
  font-size: 13px;
  font-weight: 600;
  color: var(--charcoal);
}

html.dark .term-title-text {
  color: #e4e4e7;
}

.term-badge-tag {
  font-size: 11px;
  padding: 1px 7px;
  border-radius: var(--r-full);
  font-weight: 500;
}

.tag-running {
  background: rgba(245, 158, 11, 0.15);
  color: #d97706;
}

html.dark .tag-running {
  background: rgba(245, 158, 11, 0.2);
  color: #fbbf24;
}

.tag-idle {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.05));
  color: var(--steel);
  border: 1px solid var(--hairline);
}

html.dark .tag-idle {
  background: rgba(255, 255, 255, 0.08);
  color: #a1a1aa;
  border-color: rgba(255, 255, 255, 0.1);
}

.retry-bar {
  padding: 8px 16px;
  background: rgba(239, 68, 68, 0.1);
  border-top: 1px solid rgba(239, 68, 68, 0.2);
  display: flex;
  justify-content: flex-end;
}

.ops-action {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  border-radius: var(--r-sm);
  font-size: 12px;
  background: var(--surface);
  border: 1px solid var(--hairline);
  color: var(--charcoal);
  cursor: pointer;
  transition: all 0.15s ease;
}

.ops-action:hover {
  background: var(--highlight);
}

.ops-icon {
  width: 13px;
  height: 13px;
}
</style>

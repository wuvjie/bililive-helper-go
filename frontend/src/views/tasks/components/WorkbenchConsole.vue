<template>
  <div class="workbench-console-card">
    <!-- Action Bar: 定向主播维护 -->
    <div class="workbench-action-bar">
      <div class="action-bar-left">
        <div class="ui-icon-box icon-box-blue">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/>
          </svg>
        </div>
        <span class="action-bar-title">定向主播维护</span>
        <el-select
          :model-value="selectedStreamer"
          placeholder="选择特定主播发起维护..."
          filterable
          clearable
          class="targeted-select"
          @update:model-value="$emit('update:selectedStreamer', $event)"
        >
          <el-option
            v-for="s in streamers"
            :key="s.name"
            :label="`${s.name} (${s.files} 个 · ${s.size_gb?.toFixed(1)} GB)`"
            :value="s.name"
          />
        </el-select>
        <div class="targeted-buttons">
          <button
            class="btn-targeted btn-targeted-merge"
            :disabled="!selectedStreamer || isRunning"
            @click="$emit('targeted-merge')"
          >
            <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
              <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
            </svg>
            <span>合并该主播</span>
          </button>
          <button
            class="btn-targeted btn-targeted-clean"
            :disabled="!selectedStreamer || isRunning"
            @click="$emit('targeted-clean')"
          >
            <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="3 6 5 6 21 6"/>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
            </svg>
            <span>清理旧录像</span>
          </button>
        </div>
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
            <span v-if="isRunning" class="term-badge-tag tag-running">运行中</span>
            <span v-else class="term-badge-tag tag-idle">空闲待命</span>
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
import type { StreamerInfo } from "@/api/types";
import type { SSELine } from "@/utils/sse";
import ReTerminal from "@/components/ReTerminal/index.vue";

defineProps<{
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
  (e: "targeted-merge"): void;
  (e: "targeted-clean"): void;
  (e: "clear"): void;
  (e: "retry"): void;
}>();
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
  background: rgba(255, 255, 255, 0.03);
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

.targeted-select {
  width: 250px;
}

.targeted-buttons {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn-targeted {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: var(--r-sm);
  font-size: 12px;
  font-weight: 500;
  border: 1px solid transparent;
  cursor: pointer;
  transition: all 0.15s ease;
  line-height: 1.4;
}

.btn-targeted-merge {
  background: color-mix(in srgb, #059669 10%, transparent);
  color: #047857;
  border-color: color-mix(in srgb, #059669 25%, transparent);
}
.btn-targeted-merge:hover:not(:disabled) {
  background: #059669;
  color: #ffffff;
}

.btn-targeted-clean {
  background: color-mix(in srgb, #d97706 10%, transparent);
  color: #b45309;
  border-color: color-mix(in srgb, #d97706 25%, transparent);
}
.btn-targeted-clean:hover:not(:disabled) {
  background: #d97706;
  color: #ffffff;
}

.btn-targeted:disabled {
  opacity: 0.45;
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
  color: #e4e4e7;
}

.term-badge-tag {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: var(--r-full);
}

.tag-running {
  background: rgba(245, 158, 11, 0.2);
  color: #fbbf24;
}

.tag-idle {
  background: rgba(255, 255, 255, 0.1);
  color: #a1a1aa;
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

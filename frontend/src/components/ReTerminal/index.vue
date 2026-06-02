<template>
  <div class="re-terminal" :style="{ height: height || '400px' }">
    <div class="terminal-header">
      <span class="terminal-title">📺 任务输出</span>
      <button class="terminal-clear" @click="$emit('clear')">清除</button>
    </div>
    <div class="terminal-body" ref="bodyRef">
      <div v-if="lines.length === 0" class="terminal-empty">等待任务输出 ...</div>
      <div
        v-for="(line, i) in lines"
        :key="i"
        class="terminal-line"
        :class="`terminal-${line.type}`"
      >
        <span class="terminal-time">{{ line.time }}</span>
        <span class="terminal-text">{{ line.text }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from "vue";
import type { SSELine } from "@/utils/sse";

const props = defineProps<{
  lines: SSELine[];
  height?: string;
}>();

defineEmits<{
  clear: [];
}>();

const bodyRef = ref<HTMLElement>();

watch(
  () => props.lines.length,
  () => {
    nextTick(() => {
      if (bodyRef.value) {
        bodyRef.value.scrollTop = bodyRef.value.scrollHeight;
      }
    });
  }
);
</script>

<style scoped>
.re-terminal {
  background: #191919;
  border: 1px solid #2e2e2e;
  border-radius: 6px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  font-family: var(--font-mono);
}

.terminal-header {
  background: #1f1f1f;
  padding: 8px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
  border-bottom: 1px solid #2e2e2e;
}

.terminal-title {
  color: #8a8a8a;
  font-size: 12px;
  font-weight: 500;
}

.terminal-clear {
  background: transparent;
  border: none;
  color: #666;
  font-size: 12px;
  cursor: pointer;
  padding: 2px 0 2px 6px;
  border-radius: 4px;
  font-family: var(--font-mono);
  transition: color 0.15s;
}
.terminal-clear:hover {
  color: #aaa;
}

.terminal-body {
  flex: 1;
  overflow-y: auto;
  padding: 14px 24px 14px 16px;
}

.terminal-empty {
  color: #555;
  font-size: 13px;
  text-align: center;
  padding: 40px 0;
}

.terminal-line {
  display: flex;
  gap: 12px;
  line-height: 1.8;
  font-size: 13px;
}

.terminal-time {
  color: #555;
  flex-shrink: 0;
}

.terminal-text {
  word-break: break-word;
}

.terminal-info .terminal-text { color: #b0b0b0; }
.terminal-success .terminal-text { color: #6a9955; }
.terminal-error .terminal-text { color: #f44747; }
</style>

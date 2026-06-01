<template>
  <div class="re-terminal" :style="{ height: height || '400px' }">
    <div class="terminal-header">
      <span class="terminal-title">任务输出</span>
      <button class="terminal-clear" @click="$emit('clear')">清除</button>
    </div>
    <div class="terminal-body" ref="bodyRef">
      <div v-if="lines.length === 0" class="terminal-empty">等待任务输出...</div>
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
/* Mintlify code-block spec */
.re-terminal {
  background: #1c1c1e;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* code-block-header spec */
.terminal-header {
  background: #1c1c1e;
  padding: 8px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
  border-bottom: 1px solid #1f1f1f;
}

.terminal-title {
  color: #b3b3b3;
  font-size: 13px;
  font-family: "JetBrains Mono", "SF Mono", ui-monospace, Menlo, Consolas, monospace;
}

/* copy-code-button spec */
.terminal-clear {
  background: transparent;
  color: #b3b3b3;
  font-size: 13px;
  font-family: "Inter", sans-serif;
  border: 1px solid #1f1f1f;
  border-radius: 6px;
  padding: 2px 8px;
  cursor: pointer;
  transition: color 0.1s;
}
.terminal-clear:hover {
  color: #ffffff;
}

/* code-md spec */
.terminal-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.terminal-empty {
  color: #b3b3b3;
  font-size: 14px;
  font-family: "JetBrains Mono", "SF Mono", ui-monospace, Menlo, Consolas, monospace;
  text-align: center;
  padding: 40px 0;
  line-height: 1.5;
}

.terminal-line {
  display: flex;
  gap: 12px;
  font-family: "JetBrains Mono", "SF Mono", ui-monospace, Menlo, Consolas, monospace;
  font-size: 14px;
  line-height: 1.5;
}

.terminal-time {
  color: #5a5a5c;
  flex-shrink: 0;
}

.terminal-text {
  word-break: break-all;
  color: #b3b3b3;
}

.terminal-info .terminal-text { color: #b3b3b3; }
.terminal-success .terminal-text { color: #00d4a4; }
.terminal-error .terminal-text { color: #d45656; }
</style>

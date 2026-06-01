<template>
  <div class="re-terminal" :style="{ height: height || '400px' }">
    <div class="terminal-header">
      <span class="terminal-title">📺 任务输出</span>
      <el-button text size="small" @click="$emit('clear')">清除</el-button>
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
.re-terminal {
  background: #1e1e1e;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  font-family: "SF Mono", "Cascadia Code", Consolas, monospace;
}

.terminal-header {
  background: #2d2d2d;
  padding: 8px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.terminal-title {
  color: #ccc;
  font-size: 13px;
}

.terminal-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
}

.terminal-empty {
  color: #666;
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
  color: #666;
  flex-shrink: 0;
}

.terminal-text {
  word-break: break-all;
}

.terminal-info .terminal-text { color: #d4d4d4; }
.terminal-success .terminal-text { color: #6a9955; }
.terminal-error .terminal-text { color: #f44747; }
</style>

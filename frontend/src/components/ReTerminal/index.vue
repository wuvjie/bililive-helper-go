<template>
  <div class="re-terminal" :style="{ height: height || '400px' }">
    <div class="terminal-header">
      <div class="terminal-header-left">
        <slot name="title">
          <span class="terminal-title">{{ title || "📺 任务输出" }}</span>
        </slot>
        <span class="terminal-badge-count">{{ filteredLines.length }} 行</span>
        <span v-if="errorCount > 0" class="terminal-badge-stat badge-stat-err">{{ errorCount }} 错误</span>
        <span v-if="warnCount > 0" class="terminal-badge-stat badge-stat-warn">{{ warnCount }} 警告</span>
      </div>

      <div class="terminal-header-actions">
        <div class="filter-box-wrap" v-if="lines.length > 5">
          <input
            v-model="filterKeyword"
            type="text"
            placeholder="🔍 过滤日志..."
            class="terminal-filter-input"
          />
        </div>
        <button
          class="terminal-action-btn"
          :class="{ paused: !autoScroll }"
          @click="autoScroll = !autoScroll"
          :title="autoScroll ? '点击暂停自动滚动' : '点击恢复自动跟随到底部'"
        >
          {{ autoScroll ? "⬇ 自动滚动" : "⏸ 已暂停" }}
        </button>
        <button
          class="terminal-action-btn"
          @click="handleCopy"
          title="复制当前控制台全部日志"
        >
          📋 复制
        </button>
        <button
          class="terminal-action-btn terminal-clear-btn"
          @click="$emit('clear')"
          title="清空当前控制台"
        >
          清空
        </button>
      </div>
    </div>

    <div class="terminal-body" ref="bodyRef">
      <div v-if="filteredLines.length === 0" class="terminal-empty">
        {{ filterKeyword ? '未找到包含关键字的日志行' : (emptyText || '等待任务输出 ...') }}
      </div>
      <div
        v-for="(line, i) in filteredLines"
        :key="i"
        class="terminal-line"
        :class="[`terminal-${line.type}`, { 'terminal-line-divider': isDivider(line.text) }]"
      >
        <span class="terminal-time">{{ line.time }}</span>
        <span class="terminal-text">{{ line.text }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from "vue";
import type { SSELine } from "@/utils/sse";
import { copyToClipboard } from "@/utils/clipboard";

const props = defineProps<{
  lines: SSELine[];
  height?: string;
  title?: string;
  emptyText?: string;
}>();

defineEmits<{
  clear: [];
}>();

const bodyRef = ref<HTMLElement>();
const autoScroll = ref(true);
const filterKeyword = ref("");

const filteredLines = computed(() => {
  if (!filterKeyword.value.trim()) return props.lines;
  const q = filterKeyword.value.toLowerCase();
  return props.lines.filter(l => l.text.toLowerCase().includes(q));
});

const errorCount = computed(() => props.lines.filter(l => l.type === "error").length);
const warnCount = computed(() => props.lines.filter(l => l.type === "warn").length);

function isDivider(text: string): boolean {
  return /^[─═=\-—\s]{3,}$/.test((text || "").trim());
}

async function handleCopy() {
  if (props.lines.length === 0) return;
  const rawText = props.lines.map(l => `[${l.time}] ${l.text}`).join("\n");
  await copyToClipboard(rawText, "已复制控制台日志");
}

function isAtBottom(): boolean {
  if (!bodyRef.value) return true;
  const el = bodyRef.value;
  return el.scrollHeight - el.scrollTop - el.clientHeight < 30;
}

function onScroll() {
  autoScroll.value = isAtBottom();
}

watch(
  () => props.lines.length,
  () => {
    if (!autoScroll.value) return;
    nextTick(() => {
      if (bodyRef.value) {
        bodyRef.value.scrollTop = bodyRef.value.scrollHeight;
      }
    });
  }
);

onMounted(() => {
  bodyRef.value?.addEventListener("scroll", onScroll);
});

onUnmounted(() => {
  bodyRef.value?.removeEventListener("scroll", onScroll);
});
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
  background: #252526;
  padding: 8px 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
  border-bottom: 1px solid #333333;
  flex-wrap: wrap;
  gap: 8px;
}

.terminal-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.terminal-title {
  color: #cccccc;
  font-size: 12px;
  font-weight: 500;
}

.terminal-badge-count {
  font-size: 11px;
  color: #888888;
  background: #1e1e1e;
  border: 1px solid #3c3c3c;
  padding: 1px 6px;
  border-radius: 10px;
}

.terminal-badge-stat {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 10px;
  font-weight: 600;
}
.badge-stat-err {
  background: rgba(244, 71, 71, 0.2);
  color: #f44747;
  border: 1px solid rgba(244, 71, 71, 0.4);
}
.badge-stat-warn {
  background: rgba(215, 186, 125, 0.2);
  color: #d7ba7d;
  border: 1px solid rgba(215, 186, 125, 0.4);
}

.terminal-header-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.terminal-filter-input {
  background: #1e1e1e;
  border: 1px solid #3c3c3c;
  color: #d4d4d4;
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 4px;
  outline: none;
  font-family: var(--font-mono);
  width: 110px;
  transition: all 0.15s ease;
}
.terminal-filter-input:focus {
  border-color: #007fd4;
  width: 140px;
}

.terminal-action-btn {
  background: #1e1e1e;
  border: 1px solid #3c3c3c;
  color: #cccccc;
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
  font-family: var(--font-mono);
}
.terminal-action-btn:hover {
  background: #2d2d30;
  color: #ffffff;
  border-color: #555555;
}
.terminal-action-btn.paused {
  color: #d7ba7d;
  border-color: #d7ba7d;
}

.terminal-clear-btn {
  color: #888888;
}
.terminal-clear-btn:hover {
  color: #d4433a;
  border-color: #d4433a;
}

.terminal-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
  background: #191919;
}

.terminal-empty {
  color: #666666;
  font-size: 12px;
  text-align: center;
  padding: 48px 0;
}

.terminal-line {
  display: flex;
  gap: 12px;
  line-height: 1.7;
  font-size: 12px;
  padding: 1px 4px;
  border-radius: 2px;
  transition: background 0.1s;
}
.terminal-line:hover {
  background: rgba(255, 255, 255, 0.04);
}

.terminal-line-divider {
  opacity: 0.35;
  margin: 4px 0;
}

.terminal-time {
  color: #555555;
  flex-shrink: 0;
  font-size: 11px;
  user-select: none;
}

.terminal-text {
  word-break: break-all;
}

.terminal-info .terminal-text { color: #cccccc; }
.terminal-warn .terminal-text { color: #d7ba7d; }
.terminal-success .terminal-text { color: #6a9955; }
.terminal-error .terminal-text { color: #f44747; }

@media (max-width: 640px) {
  .terminal-filter-input {
    width: 80px;
  }
  .terminal-filter-input:focus {
    width: 100px;
  }
  .terminal-header {
    padding: 8px 10px;
  }
}
</style>

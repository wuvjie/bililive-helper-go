<template>
  <div class="re-terminal" :style="{ height: height || '400px' }">
    <div class="terminal-header">
      <div class="terminal-header-left">
        <slot name="title">
          <span class="terminal-title">
            <svg class="term-title-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>
            <span>{{ title || "任务输出" }}</span>
          </span>
        </slot>
        <span v-if="errorCount > 0" class="terminal-badge-stat badge-stat-err">{{ errorCount }} 错误</span>
        <span v-if="warnCount > 0" class="terminal-badge-stat badge-stat-warn">{{ warnCount }} 警告</span>
      </div>

      <div class="terminal-header-actions">
        <div class="filter-box-wrap" v-if="lines.length > 5">
          <input
            v-model="filterKeyword"
            type="text"
            placeholder="过滤控制台日志..."
            class="terminal-filter-input"
          />
        </div>
        <button
          class="terminal-action-btn"
          :class="{ paused: !autoScroll }"
          @click="autoScroll = !autoScroll"
          :title="autoScroll ? '点击暂停自动滚动' : '点击恢复自动跟随到底部'"
        >
          <svg v-if="autoScroll" class="term-btn-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="7 13 12 18 17 13"/>
            <polyline points="7 6 12 11 17 6"/>
          </svg>
          <svg v-else class="term-btn-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="6" y="4" width="4" height="16"/>
            <rect x="14" y="4" width="4" height="16"/>
          </svg>
          <span>{{ autoScroll ? "自动滚动" : "已暂停滚动" }}</span>
        </button>
        <button
          class="terminal-action-btn"
          @click="handleCopy"
          title="复制当前控制台全部日志"
        >
          <svg class="term-btn-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
          <span>复制</span>
        </button>
        <button
          class="terminal-action-btn terminal-clear-btn"
          @click="$emit('clear')"
          title="清空当前控制台"
        >
          <svg class="term-btn-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
          <span>清空</span>
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

    <!-- 浮动回到底部控制药丸 (防滚脱手保护触发时可见) -->
    <transition name="fade">
      <button
        v-if="!autoScroll && filteredLines.length > 0"
        class="terminal-scroll-bottom-btn"
        @click="scrollToBottom"
        title="恢复自动吸底跟随"
      >
        <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><polyline points="19 12 12 19 5 12"/></svg>
        <span>回到底部</span>
      </button>
    </transition>
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

function scrollToBottom() {
  autoScroll.value = true;
  nextTick(() => {
    if (bodyRef.value) {
      bodyRef.value.scrollTop = bodyRef.value.scrollHeight;
    }
  });
}

function onScroll() {
  if (!bodyRef.value) return;
  if (!isAtBottom()) {
    if (autoScroll.value) autoScroll.value = false;
  } else {
    if (!autoScroll.value) autoScroll.value = true;
  }
}

watch(
  () => props.lines.length,
  () => {
    if (autoScroll.value) {
      nextTick(() => {
        if (bodyRef.value) {
          bodyRef.value.scrollTop = bodyRef.value.scrollHeight;
        }
      });
    }
  }
);

onMounted(() => {
  if (bodyRef.value) {
    bodyRef.value.addEventListener("scroll", onScroll, { passive: true });
  }
});

onUnmounted(() => {
  if (bodyRef.value) {
    bodyRef.value.removeEventListener("scroll", onScroll);
  }
});
</script>

<style scoped>
.re-terminal {
  position: relative;
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--hairline);
  border-radius: 6px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  font-family: var(--font-mono);
}

.terminal-header {
  background: var(--surface-secondary, #f6f8fa);
  border-bottom: 1px solid var(--hairline);
  padding: 8px 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
  flex-wrap: wrap;
  gap: 8px;
}

.terminal-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.terminal-title {
  color: var(--charcoal);
  font-size: 12px;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.term-title-svg {
  width: 14px;
  height: 14px;
  color: var(--steel);
}

.terminal-badge-stat {
  font-size: 11px;
  padding: 1px 7px;
  border-radius: var(--r-full);
  font-weight: 600;
  white-space: nowrap;
  flex-shrink: 0;
}
.badge-stat-err {
  background: rgba(239, 68, 68, 0.15);
  color: #dc2626;
  border: 1px solid rgba(239, 68, 68, 0.3);
}
.badge-stat-warn {
  background: rgba(245, 158, 11, 0.15);
  color: #d97706;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.terminal-header-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.terminal-filter-input {
  background: var(--surface, #ffffff);
  border: 1px solid var(--hairline);
  color: var(--charcoal);
  font-size: 12px !important;
  height: 26px !important;
  line-height: 24px !important;
  box-sizing: border-box !important;
  padding: 0 8px;
  border-radius: var(--r-xs, 4px);
  outline: none;
  font-family: var(--font-mono);
  width: 120px;
  transition: all 0.15s ease;
}
.terminal-filter-input:focus {
  border-color: var(--primary);
  width: 150px;
}

.terminal-action-btn {
  background: var(--surface, #ffffff);
  border: 1px solid var(--hairline);
  color: var(--charcoal);
  font-size: 12px !important;
  height: 26px !important;
  line-height: 24px !important;
  box-sizing: border-box !important;
  padding: 0 8px;
  border-radius: var(--r-xs, 4px);
  cursor: pointer;
  transition: all 0.15s ease;
  font-family: var(--font-mono);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}
.terminal-action-btn:hover {
  background: var(--highlight, rgba(0, 0, 0, 0.04));
  border-color: var(--charcoal);
}
.terminal-action-btn.paused {
  color: #d97706;
  border-color: #d97706;
}

.terminal-clear-btn {
  color: var(--steel);
}
.terminal-clear-btn:hover {
  color: var(--error, #dc2626);
  border-color: var(--error, #dc2626);
}

.term-btn-svg {
  width: 12px;
  height: 12px;
}

.terminal-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
  background: #f6f8fa;
  color: #24292f;
}

html.dark .terminal-body {
  background: #0d1117;
  color: #c9d1d9;
}

html.dark .re-terminal {
  border-color: rgba(255, 255, 255, 0.1);
}

html.dark .terminal-header {
  background: #161b22;
  border-color: #30363d;
}

html.dark .terminal-title {
  color: #c9d1d9;
}

html.dark .terminal-filter-input {
  background: #0d1117;
  border-color: #30363d;
  color: #c9d1d9;
}

html.dark .terminal-action-btn {
  background: #21262d;
  border-color: #30363d;
  color: #c9d1d9;
}
html.dark .terminal-action-btn:hover {
  background: #30363d;
  color: #f0f6fc;
}

.terminal-empty {
  color: var(--stone);
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
  background: rgba(0, 0, 0, 0.03);
}
html.dark .terminal-line:hover {
  background: rgba(255, 255, 255, 0.04);
}

.terminal-line-divider {
  opacity: 0.4;
  margin: 4px 0;
  color: #d0d7de;
}
html.dark .terminal-line-divider {
  color: #30363d;
}

.terminal-time {
  color: #57606a;
  flex-shrink: 0;
  font-size: 11px;
  user-select: none;
}
html.dark .terminal-time {
  color: #8b949e;
}

.terminal-text {
  flex: 1;
  word-break: break-all;
}

.terminal-info .terminal-text { color: #0969da; }
.terminal-success .terminal-text { color: #1a7f37; }
.terminal-warn .terminal-text { color: #9a6700; }
.terminal-error .terminal-text { color: #cf222e; font-weight: 500; }

html.dark .terminal-info .terminal-text { color: #58a6ff; }
html.dark .terminal-success .terminal-text { color: #3fb950; }
html.dark .terminal-warn .terminal-text { color: #d29922; }
html.dark .terminal-error .terminal-text { color: #f85149; }

.terminal-scroll-bottom-btn {
  position: absolute;
  bottom: 16px;
  right: 24px;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  border-radius: var(--r-full);
  background: var(--surface);
  border: 1px solid var(--hairline);
  box-shadow: var(--shadow-card);
  color: var(--charcoal);
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
  z-index: 5;
}

.terminal-scroll-bottom-btn:hover {
  background: var(--highlight);
  color: var(--primary);
  border-color: var(--primary);
}
</style>

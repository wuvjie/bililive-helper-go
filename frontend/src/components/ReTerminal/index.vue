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
          {{ autoScroll ? "自动滚动" : "已暂停滚动" }}
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
  position: relative;
  color-scheme: dark;
  background: #191919;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 6px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  font-family: var(--font-mono);
}

.terminal-header {
  background: #18181b !important;
  border-bottom: 1px solid rgba(255, 255, 255, 0.14) !important;
  background: #18181b !important;
  border-bottom: 1px solid rgba(255, 255, 255, 0.14) !important;
  background: #252526;
  padding: 8px 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
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
  color: #cbd5e1 !important;
  background: rgba(255, 255, 255, 0.1) !important;
  color: #cbd5e1 !important;
  background: rgba(255, 255, 255, 0.1) !important;
  font-size: 12px;
  color: #888888;
  background: #1e1e1e;
  border: 1px solid rgba(255, 255, 255, 0.12);
  padding: 1px 6px;
  border-radius: 10px;
  white-space: nowrap;
  flex-shrink: 0;
}

.terminal-badge-stat {
  font-size: 12px;
  padding: 1px 6px;
  border-radius: 10px;
  font-weight: 600;
  white-space: nowrap;
  flex-shrink: 0;
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
  border: 1px solid rgba(255, 255, 255, 0.14);
  color: #d4d4d4;
  font-size: 12px !important;
  height: 26px !important;
  line-height: 24px !important;
  box-sizing: border-box !important;
  padding: 0 8px;
  border-radius: 4px;
  outline: none;
  font-family: var(--font-mono);
  width: 110px;
  transition: all 0.15s ease;
}
.terminal-filter-input:focus {
  border-color: #3b82f6;
  width: 140px;
}

.terminal-action-btn {
  color: #e4e4e7 !important;
  background: rgba(255, 255, 255, 0.08) !important;
  border: 1px solid rgba(255, 255, 255, 0.18) !important;
  color: #e4e4e7 !important;
  background: rgba(255, 255, 255, 0.08) !important;
  border: 1px solid rgba(255, 255, 255, 0.18) !important;
  background: #1e1e1e;
  border: 1px solid rgba(255, 255, 255, 0.14);
  color: #cccccc;
  font-size: 12px !important;
  height: 26px !important;
  line-height: 24px !important;
  box-sizing: border-box !important;
  padding: 0 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
  font-family: var(--font-mono);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
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

.terminal-title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.term-title-svg {
  width: 14px;
  height: 14px;
  color: #4ade80;
}
.term-btn-svg {
  width: 12px;
  height: 12px;
  margin-right: 4px;
  vertical-align: -1px;
}


.terminal-scroll-bottom-btn {
  position: absolute;
  bottom: 14px;
  right: 20px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: rgba(24, 24, 27, 0.92);
  border: 1px solid rgba(56, 189, 248, 0.45);
  color: #38bdf8;
  padding: 5px 12px;
  border-radius: 20px;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(8px);
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  z-index: 10;
  user-select: none;
}
.terminal-scroll-bottom-btn:hover {
  background: rgba(56, 189, 248, 0.15);
  border-color: #38bdf8;
  transform: translateY(-1px);
}
</style>

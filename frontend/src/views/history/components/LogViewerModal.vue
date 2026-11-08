<template>
  <el-dialog
    :model-value="visible"
    :title="`操作审计日志 · ${title}`"
    width="min(880px, 94vw)"
    top="5vh"
    destroy-on-close
    class="terminal-modal-dialog"
    @update:model-value="$emit('update:visible', $event)"
  >
    <div class="log-modal-body" v-loading="loading">
      <div class="log-modal-bar">
        <div class="log-bar-left">
          <span
            v-if="status"
            class="log-meta-status"
            :class="status === 'success' ? 'status-ok' : (status === 'partial' ? 'status-warn' : 'status-fail')"
          >
            {{ status === 'success' ? '成功' : (status === 'partial' ? '部分' : '失败') }}
          </span>
          <span
            class="log-meta-tag log-meta-id"
            :title="`任务 ID: ${logId} (点击复制)`"
            @click="copyToClipboard(logId, '已复制任务 ID')"
          >
            <span class="meta-label">ID:</span> {{ logId }}
          </span>
        </div>
        <div class="log-modal-actions">
          <button class="btn-copy-log" @click="downloadRawLog" title="下载完整原始日志 (.log)">
            <svg class="action-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
            <span>下载日志</span>
          </button>
          <button class="btn-copy-log" @click="handleCopyContent" title="复制全部内容到剪贴板">
            <svg class="action-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
            </svg>
            <span>复制日志</span>
          </button>
        </div>
      </div>

      <div class="log-terminal-lines">
        <div v-if="filteredLogLines.length === 0" class="log-empty-lines">
          日志内容为空
        </div>
        <div
          v-for="(line, idx) in filteredLogLines"
          :key="idx"
          class="log-code-row"
          :class="line.typeClass"
        >
          <span class="log-line-num">{{ idx + 1 }}</span>
          <span class="log-line-text">{{ line.text }}</span>
        </div>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { ElMessage } from "element-plus";
import { copyToClipboard } from "@/utils/clipboard";

const props = defineProps<{
  visible: boolean;
  title: string;
  status: string;
  logId: string;
  loading: boolean;
  content: string;
}>();

defineEmits<{
  (e: "update:visible", val: boolean): void;
}>();

interface FormattedLogLine {
  text: string;
  typeClass: string;
}

const filteredLogLines = computed<FormattedLogLine[]>(() => {
  if (!props.content) return [];
  const rawLines = props.content.split("\n");
  return rawLines.map(line => {
    let typeClass = "line-info";
    if (/❌|错误|失败|error|failed|fatal/i.test(line)) {
      typeClass = "line-error";
    } else if (/⚠️|⚠|警告|warn|warning|注意|跳过|空间不足/i.test(line)) {
      typeClass = "line-warn";
    } else if (/✅|完成|成功|success|done/i.test(line)) {
      typeClass = "line-success";
    } else if (/^[─═=—\-\s]{3,}$/.test(line.trim())) {
      typeClass = "line-divider";
    } else if (/📊 结果摘要|▶ 开始/i.test(line)) {
      typeClass = "line-header";
    }
    return { text: line, typeClass };
  });
});

function downloadRawLog() {
  const c = props.content;
  if (!c) {
    ElMessage.warning("日志内容为空或尚未加载完成");
    return;
  }
  const blob = new Blob([c], { type: "text/plain;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `${props.logId || "task"}.log`;
  a.click();
  URL.revokeObjectURL(url);
  ElMessage.success("日志文件已开始下载");
}

async function handleCopyContent() {
  if (!props.content) {
    ElMessage.warning("日志内容为空");
    return;
  }
  await copyToClipboard(props.content, "已复制全部日志内容");
}
</script>

<style scoped>
.log-modal-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.log-modal-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--surface-secondary, rgba(0, 0, 0, 0.03));
  border-radius: var(--r-sm);
  border: 1px solid var(--hairline);
  font-size: 12px;
  flex-wrap: wrap;
  gap: 8px;
}

.log-bar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.log-meta-status {
  padding: 1px 7px;
  border-radius: var(--r-full);
  font-weight: 600;
  font-size: 11px;
}

.status-ok {
  background: rgba(16, 185, 129, 0.15);
  color: #059669;
}

.status-warn {
  background: rgba(245, 158, 11, 0.15);
  color: #d97706;
}

.status-fail {
  background: rgba(239, 68, 68, 0.15);
  color: #dc2626;
}

.log-meta-tag {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--slate);
  cursor: pointer;
  padding: 2px 6px;
  border-radius: var(--r-xs);
  background: var(--surface);
  border: 1px solid var(--hairline);
}

.meta-label {
  color: var(--stone);
}

.log-modal-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-copy-log {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 9px;
  border-radius: var(--r-sm);
  font-size: 11px;
  border: 1px solid var(--hairline);
  background: var(--surface);
  color: var(--charcoal);
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-copy-log:hover {
  background: var(--highlight);
  color: var(--ink);
}

.action-icon {
  width: 12px;
  height: 12px;
}

.log-terminal-lines {
  background: #18181b;
  color: #f4f4f5;
  border-radius: var(--r-sm);
  padding: 12px;
  max-height: 520px;
  overflow-y: auto;
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.6;
}

.log-empty-lines {
  text-align: center;
  padding: 40px 0;
  color: #71717a;
}

.log-code-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  word-break: break-all;
  white-space: pre-wrap;
}

.log-line-num {
  color: #71717a;
  user-select: none;
  min-width: 30px;
  text-align: right;
  flex-shrink: 0;
  font-size: 11px;
}

.log-line-text {
  flex: 1;
}

.line-info { color: #e4e4e7; }
.line-error { color: #f87171; font-weight: 500; }
.line-warn { color: #fbbf24; }
.line-success { color: #4ade80; }
.line-divider { color: #52525b; }
.line-header { color: #60a5fa; font-weight: 600; }
</style>

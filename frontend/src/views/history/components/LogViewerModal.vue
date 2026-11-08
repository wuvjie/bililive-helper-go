<template>
  <el-dialog
    :model-value="visible"
    width="min(880px, 94vw)"
    top="5vh"
    destroy-on-close
    class="terminal-modal-dialog"
    @update:model-value="$emit('update:visible', $event)"
  >
    <template #header>
      <div class="modal-header-container">
        <div class="modal-header-top-row">
          <div class="header-title-left">
            <h3 class="modal-title-text">操作审计日志</h3>
            <span
              v-if="status"
              class="log-meta-status"
              :class="status === 'success' ? 'status-ok' : (status === 'partial' ? 'status-warn' : 'status-fail')"
            >
              {{ status === 'success' ? '成功' : (status === 'partial' ? '部分' : '失败') }}
            </span>
          </div>
          <div class="header-title-right">
            <button class="btn-ghost-action" @click="downloadRawLog" title="下载完整原始日志 (.log)">
              <svg class="action-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>
              </svg>
              <span>下载</span>
            </button>
            <button class="btn-ghost-action" @click="handleCopyContent" title="复制全部内容到剪贴板">
              <svg class="action-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
              </svg>
              <span>复制</span>
            </button>
          </div>
        </div>

        <div class="modal-header-subrow">
          <span class="subrow-task-desc" :title="title">{{ title }}</span>
          <span class="subrow-divider">·</span>
          <el-tooltip content="点击复制任务 ID" placement="top" :show-after="300">
            <span class="log-meta-id" @click="copyToClipboard(logId, '已复制任务 ID')">
              <span class="meta-hash">#</span>{{ logId }}
            </span>
          </el-tooltip>
        </div>
      </div>
    </template>

    <div class="log-modal-body" v-loading="loading">
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
.modal-header-container {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
  padding-right: 32px;
  box-sizing: border-box;
}

.modal-header-top-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.header-title-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-title-text {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.2px;
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

.header-title-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.btn-ghost-action {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 9px;
  border-radius: var(--r-sm);
  font-size: 11.5px;
  font-weight: 500;
  border: 1px solid var(--hairline);
  background: var(--surface);
  color: var(--charcoal);
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-ghost-action:hover {
  background: var(--highlight);
  border-color: var(--slate);
  color: var(--ink);
}

.action-icon {
  width: 12px;
  height: 12px;
}

.modal-header-subrow {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--steel);
}

.subrow-task-desc {
  max-width: 500px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
  color: var(--slate);
}

.subrow-divider {
  color: var(--stone);
}

.log-meta-id {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--steel);
  cursor: pointer;
  padding: 1px 6px;
  border-radius: var(--r-xs);
  background: var(--surface-secondary, rgba(0, 0, 0, 0.04));
  border: 1px solid var(--hairline);
  transition: all 0.15s ease;
}

.log-meta-id:hover {
  background: var(--highlight);
  color: var(--primary);
  border-color: var(--primary);
}

.meta-hash {
  color: var(--stone);
  margin-right: 1px;
}

.log-modal-body {
  display: flex;
  flex-direction: column;
}

.log-terminal-lines {
  background: #f6f8fa;
  color: #24292f;
  border: 1px solid var(--hairline);
  border-radius: var(--r-sm);
  padding: 12px 14px;
  max-height: 540px;
  overflow-y: auto;
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.65;
}

html.dark .log-terminal-lines {
  background: #0d1117;
  color: #c9d1d9;
  border-color: #30363d;
}

.log-empty-lines {
  text-align: center;
  padding: 48px 0;
  color: var(--stone);
}

.log-code-row {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  word-break: break-all;
  white-space: pre-wrap;
  padding: 1px 0;
}

.log-line-num {
  color: #6e7781;
  user-select: none;
  min-width: 32px;
  text-align: right;
  flex-shrink: 0;
  font-size: 11px;
}

html.dark .log-line-num {
  color: #8b949e;
}

.log-line-text {
  flex: 1;
}

.line-info { color: #24292f; }
.line-error { color: #cf222e; font-weight: 500; }
.line-warn { color: #9a6700; }
.line-success { color: #1a7f37; }
.line-divider { color: #d0d7de; }
.line-header { color: #0969da; font-weight: 600; }

html.dark .line-info { color: #c9d1d9; }
html.dark .line-error { color: #f87171; font-weight: 500; }
html.dark .line-warn { color: #fbbf24; }
html.dark .line-success { color: #4ade80; }
html.dark .line-divider { color: #30363d; }
html.dark .line-header { color: #58a6ff; font-weight: 600; }
</style>

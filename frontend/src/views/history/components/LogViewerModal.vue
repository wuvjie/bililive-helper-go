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
            <h3 class="modal-title-text">任务执行详细日志</h3>
            <span
              v-if="status"
              class="log-meta-status"
              :class="status === 'success' ? 'status-ok' : (status === 'partial' ? 'status-warn' : 'status-fail')"
            >
              {{ status === 'success' ? '成功' : (status === 'partial' ? '部分完成' : '失败') }}
            </span>
          </div>
        </div>

        <div v-if="title" class="modal-header-subrow">
          <span class="subrow-task-desc" :title="title">{{ title }}</span>
        </div>
      </div>
    </template>

    <div class="log-modal-body" v-loading="loading">
      <div class="log-terminal-lines">
        <div v-if="filteredLogLines.length === 0" class="log-empty-lines">
          暂无日志内容
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

    <template #footer>
      <div class="modal-footer-container">
        <div class="footer-actions-left">
          <button class="btn-dialog-secondary" @click="handleCopyContent" title="复制全部日志到剪贴板">
            <svg class="action-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
              <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
            </svg>
            <span>复制全部日志</span>
          </button>
          <button class="btn-dialog-secondary" @click="downloadRawLog" title="下载完整日志文件 (.log)">
            <svg class="action-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
            <span>下载日志文件</span>
          </button>
        </div>
        <div class="footer-actions-right">
          <button class="btn-dialog-primary" @click="$emit('update:visible', false)">
            关闭
          </button>
        </div>
      </div>
    </template>
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
  task?: string;
  logId?: string;
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
    const stripped = line.replace(/^\[\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}\]\s*/, "").trim();
    let typeClass = "line-info";
    if (/❌|错误|失败|error|failed|fatal/i.test(line)) {
      typeClass = "line-error";
    } else if (/⚠️|⚠|警告|warn|warning|注意|跳过|空间不足/i.test(line)) {
      typeClass = "line-warn";
    } else if (/✅|完成|成功|success|done/i.test(line)) {
      typeClass = "line-success";
    } else if (/^[─═=—\-\s]{3,}$/.test(stripped)) {
      typeClass = "line-divider";
    } else if (/📊 结果摘要|▶ 开始/i.test(line)) {
      typeClass = "line-header";
    }
    return { text: line, typeClass };
  });
});

async function downloadRawLog() {
  let content = props.content;
  if (props.task && props.logId) {
    try {
      const res = await fetch(
        `/api/logs/content/${encodeURIComponent(props.task)}?log_id=${encodeURIComponent(props.logId)}&raw=true`,
        { credentials: "same-origin" }
      );
      if (res.ok) {
        const ct = res.headers.get("content-type") || "";
        if (ct.includes("application/json")) {
          const parsed = await res.json();
          if (parsed?.content) {
            content = parsed.content;
          }
        } else {
          content = await res.text();
        }
      }
    } catch {
      // 网络失败时回退使用已加载的日志预览内容
    }
  }
  if (!content) {
    ElMessage.warning("暂无日志内容可供下载");
    return;
  }
  const blob = new Blob([content], { type: "text/plain;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `${props.logId || "audit"}.log`;
  a.click();
  URL.revokeObjectURL(url);
  ElMessage.success("日志文件已开始下载");
}

async function handleCopyContent() {
  if (!props.content) {
    ElMessage.warning("暂无日志内容可供复制");
    return;
  }
  await copyToClipboard(props.content, "日志内容已复制到剪贴板");
}
</script>

<style scoped>
.modal-header-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
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

html.dark {
  .status-ok {
    background: rgba(63, 185, 80, 0.15);
    color: #3fb950;
  }
  .status-warn {
    background: rgba(210, 153, 34, 0.15);
    color: #fbbf24;
  }
  .status-fail {
    background: rgba(248, 81, 73, 0.15);
    color: #f85149;
  }
}

.modal-header-subrow {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--steel);
}

.subrow-task-desc {
  max-width: 600px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
  color: var(--slate);
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
  color: var(--color-fg-muted);
  user-select: none;
  min-width: 32px;
  text-align: right;
  flex-shrink: 0;
  font-size: 11px;
}

.log-line-text {
  flex: 1;
}

.line-info { color: #24292f; }
.line-error { color: #cf222e; font-weight: 500; }
.line-warn { color: #9a6700; }
.line-success { color: #1a7f37; }
.line-divider { color: #d0d7de; }
.line-header { color: var(--color-accent-fg); font-weight: 600; }

html.dark {
  .line-info { color: #c9d1d9; }
  .line-error { color: #f87171; font-weight: 500; }
  .line-warn { color: #fbbf24; }
  .line-success { color: #4ade80; }
  .line-divider { color: #30363d; }
}

.modal-footer-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.footer-actions-left,
.footer-actions-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-dialog-secondary,
.btn-dialog-primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
  line-height: 20px;
}

.btn-dialog-secondary {
  padding: 5px 12px;
  color: var(--color-btn-text);
  background: var(--color-btn-bg);
  border: 1px solid var(--color-btn-border);
  box-shadow: 0 1px 0 rgba(27, 31, 36, 0.04);

  &:hover {
    background: var(--color-btn-hover-bg);
    border-color: var(--color-btn-hover-border);
    color: var(--color-accent-fg);
  }

  .action-icon {
    width: 13px;
    height: 13px;
  }
}

.btn-dialog-primary {
  padding: 5px 16px;
  color: #ffffff;
  background: var(--color-accent-emphasis);
  border: 1px solid rgba(27, 31, 36, 0.15);
  box-shadow: 0 1px 0 rgba(27, 31, 36, 0.1);

  &:hover {
    background: var(--color-accent-emphasis-hover);
  }
}
</style>

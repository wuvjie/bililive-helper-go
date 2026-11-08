<template>
  <el-dialog
    :model-value="visible"
    title="配置备份与恢复"
    width="min(560px, 94vw)"
    top="9vh"
    destroy-on-close
    class="backup-dialog"
    @update:model-value="$emit('update:visible', $event)"
  >
    <!-- 顶部分段切换器 (Segmented Control) -->
    <div class="backup-tab-nav">
      <button
        type="button"
        class="tab-nav-btn"
        :class="{ active: activeTab === 'export' }"
        @click="activeTab = 'export'"
      >
        <svg class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="7 10 12 15 17 10"/>
          <line x1="12" y1="15" x2="12" y2="3"/>
        </svg>
        <span>导出备份</span>
      </button>
      <button
        type="button"
        class="tab-nav-btn"
        :class="{ active: activeTab === 'import' }"
        @click="activeTab = 'import'"
      >
        <svg class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="17 8 12 3 7 8"/>
          <line x1="12" y1="3" x2="12" y2="15"/>
        </svg>
        <span>导入恢复</span>
      </button>
    </div>

    <div class="backup-body">
      <!-- Tab 1: 导出备份 -->
      <div v-if="activeTab === 'export'" class="tab-pane">
        <div class="pane-meta-row">
          <span class="pane-title">导出当前系统配置</span>
          <el-tooltip
            content="将当前录制合并规则、定时巡检周期、存储水位阈值与主播专属配置打包导出为标准 JSON 格式文件。"
            placement="top"
            :show-after="200"
          >
            <span class="help-question-icon">?</span>
          </el-tooltip>
        </div>

        <div class="export-asset-card">
          <div class="asset-icon-box">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
              <line x1="16" y1="13" x2="8" y2="13"/>
              <line x1="16" y1="17" x2="8" y2="17"/>
              <polyline points="10 9 9 9 8 9"/>
            </svg>
          </div>
          <div class="asset-meta-info">
            <span class="asset-filename">config-backup.json</span>
            <span class="asset-subinfo">包含：全局策略 · 定时调度 · 主播规则</span>
          </div>
        </div>

        <div class="export-actions-grid">
          <button type="button" class="btn-action-primary" @click="handleExport">
            <svg class="btn-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="7 10 12 15 17 10"/>
              <line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
            <span>下载配置文件 (.json)</span>
          </button>
          <button type="button" class="btn-action-secondary" @click="handleCopy">
            <svg v-if="!copiedRecently" class="btn-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
              <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
            </svg>
            <svg v-else class="btn-svg copied-check" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="20 6 9 17 4 12"/>
            </svg>
            <span>{{ copiedRecently ? '已复制到剪贴板' : '复制配置代码' }}</span>
          </button>
        </div>
      </div>

      <!-- Tab 2: 导入恢复 -->
      <div v-else class="tab-pane">
        <div class="pane-meta-row">
          <span class="pane-title">选择或上传配置文件</span>
          <el-tooltip
            content="导入配置将直接覆盖当前系统的合并与清理规则，操作不可逆，建议在导入前先导出备份以防配置丢失。"
            placement="top"
            :show-after="200"
          >
            <span class="help-question-icon">?</span>
          </el-tooltip>
        </div>

        <!-- 真实文件输入框 -->
        <input
          ref="importFileRef"
          type="file"
          accept=".json,application/json"
          style="display: none"
          @change="onFileChange"
        />

        <!-- 拖拽/点击上传区 (未选文件) -->
        <div
          v-if="!selectedFileName"
          class="dropzone-box"
          :class="{ 'is-dragover': isDragOver }"
          @click="importFileRef?.click()"
          @dragover.prevent="isDragOver = true"
          @dragleave.prevent="isDragOver = false"
          @drop.prevent="onDrop"
        >
          <div class="dropzone-icon-wrap">
            <svg class="dropzone-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="17 8 12 3 7 8"/>
              <line x1="12" y1="3" x2="12" y2="15"/>
            </svg>
          </div>
          <div class="dropzone-content">
            <div class="dropzone-primary-text">
              <span>拖拽备份文件至此处，或</span>
              <span class="dropzone-browse-link">浏览本地文件</span>
            </div>
            <div class="dropzone-secondary-text">仅支持导出的 .json 配置文件</div>
          </div>
        </div>

        <!-- 已选文件卡片 -->
        <div v-else class="selected-file-card">
          <div class="file-icon-badge">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
              <line x1="16" y1="13" x2="8" y2="13"/>
              <line x1="16" y1="17" x2="8" y2="17"/>
              <polyline points="10 9 9 9 8 9"/>
            </svg>
          </div>
          <div class="file-info-col">
            <span class="file-name" :title="selectedFileName">{{ selectedFileName }}</span>
            <span class="file-status">已读取待覆盖</span>
          </div>
          <div class="file-card-actions">
            <button type="button" class="btn-ghost-sm" title="重新选择" @click="importFileRef?.click()">重新选择</button>
            <button type="button" class="btn-ghost-remove" title="清除文件" @click="handleRemoveFile">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"/>
                <line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
            </button>
          </div>
        </div>

        <!-- 手动粘贴折叠区 -->
        <div class="import-manual-section">
          <button type="button" class="toggle-paste-btn" @click="showManualPaste = !showManualPaste">
            <svg class="toggle-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" :class="{ 'rotate-open': showManualPaste }">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
            <span>{{ showManualPaste ? '收起配置代码输入框' : '或手动粘贴 JSON 配置代码' }}</span>
          </button>

          <div v-if="showManualPaste" class="paste-textarea-wrap">
            <textarea
              class="backup-textarea"
              :value="importJson"
              placeholder="在此直接粘贴导出的 JSON 配置字符串..."
              spellcheck="false"
              @input="importJson = ($event.target as HTMLTextAreaElement).value"
            />
            <div v-if="importJson" class="paste-clear-row">
              <button type="button" class="btn-clear-text" @click="importJson = ''">清空内容</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 弹窗底部操作栏 -->
    <template #footer>
      <div class="backup-dialog-footer">
        <div class="footer-left-status">
          <span v-if="activeTab === 'import' && hasValidImportContent" class="status-ready-text">
            配置已就绪，点击确认应用
          </span>
        </div>
        <div class="footer-right-actions">
          <el-button @click="$emit('update:visible', false)">
            {{ activeTab === 'export' ? '完成' : '取消' }}
          </el-button>
          <el-button
            v-if="activeTab === 'import'"
            type="primary"
            :loading="importing"
            :disabled="!hasValidImportContent"
            @click="handleApplyImport"
          >
            确认导入并覆盖
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { getConfigExport, importConfig as apiImportConfig } from "@/api/config";
import { copyToClipboard } from "@/utils/clipboard";

defineProps<{ visible: boolean }>();
const emit = defineEmits<{
  "update:visible": [value: boolean];
  "imported": [];
}>();

const activeTab = ref<"export" | "import">("export");
const selectedFileName = ref("");
const importJson = ref("");
const showManualPaste = ref(false);
const importFileRef = ref<HTMLInputElement | null>(null);
const isDragOver = ref(false);
const copiedRecently = ref(false);
const importing = ref(false);

const hasValidImportContent = computed(() => {
  return Boolean(selectedFileName.value || importJson.value.trim());
});

async function handleExport() {
  try {
    const data = await getConfigExport();
    const exportStr = JSON.stringify(data, null, 2);
    const blob = new Blob([exportStr], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `config-backup-${new Date().toISOString().slice(0, 10)}.json`;
    a.click();
    URL.revokeObjectURL(url);
    ElMessage.success("配置备份已导出");
  } catch {
    ElMessage.error("导出配置备份失败");
  }
}

async function handleCopy() {
  try {
    const res = await getConfigExport();
    const jsonStr = JSON.stringify(res, null, 2);
    await copyToClipboard(jsonStr, "已复制配置 JSON 到剪贴板");
    copiedRecently.value = true;
    setTimeout(() => {
      copiedRecently.value = false;
    }, 2000);
  } catch {
    ElMessage.error("复制配置代码失败");
  }
}

function handleRemoveFile() {
  selectedFileName.value = "";
  importJson.value = "";
  if (importFileRef.value) {
    importFileRef.value.value = "";
  }
}

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (file) {
    if (!file.name.toLowerCase().endsWith(".json")) {
      ElMessage.warning("请选择 .json 格式的配置文件");
      input.value = "";
      return;
    }
    processFile(file);
  }
  input.value = "";
}

function onDrop(e: DragEvent) {
  isDragOver.value = false;
  const files = e.dataTransfer?.files;
  if (!files || files.length === 0) return;
  const file = files[0];
  if (!file.name.toLowerCase().endsWith(".json")) {
    ElMessage.warning("请选择 .json 格式的配置文件");
    return;
  }
  processFile(file);
}

function processFile(file: File) {
  selectedFileName.value = file.name;
  const reader = new FileReader();
  reader.onload = (e) => {
    const text = e.target?.result;
    if (typeof text === "string") {
      importJson.value = text;
      ElMessage.success(`已读取文件: ${file.name}`);
    }
  };
  reader.readAsText(file);
}

async function handleApplyImport() {
  if (!importJson.value.trim()) {
    ElMessage.warning("请先选择备份文件或粘贴配置代码");
    return;
  }

  let data: any;
  try {
    data = JSON.parse(importJson.value);
  } catch {
    ElMessage.error("JSON 格式错误，请检查配置内容");
    return;
  }

  try {
    await ElMessageBox.confirm(
      "导入配置将直接覆盖当前的录制合并策略、清理阈值与定时调度，是否确认覆盖？",
      "确认导入配置",
      { confirmButtonText: "确认导入", cancelButtonText: "取消", type: "warning" }
    );
  } catch {
    return;
  }

  importing.value = true;
  try {
    await apiImportConfig(data);
    ElMessage.success("系统配置导入恢复成功");
    emit("imported");
    emit("update:visible", false);
    handleRemoveFile();
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error || err?.message || "导入配置失败");
  } finally {
    importing.value = false;
  }
}
</script>

<style scoped>
/* Segmented Control 顶部分段切换器 */
.backup-tab-nav {
  display: flex;
  background: var(--color-canvas-subtle, #f6f8fa);
  border: 1px solid var(--color-border-subtle, #d0d7de);
  border-radius: 6px;
  padding: 3px;
  gap: 3px;
  margin-bottom: 16px;
}

html.dark .backup-tab-nav {
  background: #21262d;
  border-color: #30363d;
}

.tab-nav-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 6px 14px;
  font-size: 13px;
  font-weight: 500;
  border-radius: 5px;
  border: none;
  background: transparent;
  color: var(--color-fg-muted, #57606a);
  cursor: pointer;
  transition: all 0.15s ease;

  &:hover:not(.active) {
    color: var(--color-fg-default, #24292f);
  }

  &.active {
    background: var(--color-canvas-default, #ffffff);
    color: var(--color-fg-default, #24292f);
    font-weight: 600;
    box-shadow: 0 1px 2px rgba(27, 31, 36, 0.08);
  }

  .tab-icon {
    width: 14px;
    height: 14px;
  }
}

html.dark .tab-nav-btn {
  color: #8b949e;

  &:hover:not(.active) {
    color: #c9d1d9;
  }

  &.active {
    background: #30363d;
    color: #f0f6fc;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
  }
}

.backup-body {
  min-height: 160px;
}

.tab-pane {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.pane-meta-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.pane-title {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.2px;
}

.help-question-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 15px;
  height: 15px;
  border-radius: 50%;
  border: 1px solid var(--color-border-subtle, #d0d7de);
  color: var(--color-fg-muted, #57606a);
  font-size: 11px;
  font-weight: 600;
  cursor: help;
  user-select: none;
  line-height: 1;

  &:hover {
    border-color: var(--color-accent-fg, #0969da);
    color: var(--color-accent-fg, #0969da);
  }
}

html.dark .help-question-icon {
  border-color: #30363d;
  color: #8b949e;

  &:hover {
    border-color: #58a6ff;
    color: #58a6ff;
  }
}

/* 导出资产卡片 */
.export-asset-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 16px;
  border-radius: 6px;
  background: var(--surface-card);
  border: 1px solid var(--hairline);
}

.asset-icon-box {
  width: 38px;
  height: 38px;
  border-radius: 6px;
  background: color-mix(in srgb, #0969da 10%, transparent);
  color: #0969da;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;

  svg {
    width: 20px;
    height: 20px;
  }
}

html.dark .asset-icon-box {
  background: rgba(56, 139, 253, 0.15);
  color: #58a6ff;
}

.asset-meta-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.asset-filename {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal);
  font-family: var(--font-mono);
}

.asset-subinfo {
  font-size: 12px;
  color: var(--stone);
}

.export-actions-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;

  @media (max-width: 480px) {
    grid-template-columns: 1fr;
  }
}

.btn-action-primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 7px 14px;
  font-size: 12.5px;
  font-weight: 500;
  border-radius: 6px;
  color: #ffffff;
  background: var(--color-accent-emphasis, #0969da);
  border: 1px solid rgba(27, 31, 36, 0.15);
  box-shadow: 0 1px 0 rgba(27, 31, 36, 0.1);
  cursor: pointer;
  transition: all 0.15s ease;

  &:hover {
    background: #0860ca;
  }

  .btn-svg {
    width: 14px;
    height: 14px;
  }
}

html.dark .btn-action-primary {
  background: #1f6feb;
  border-color: rgba(240, 246, 252, 0.1);

  &:hover {
    background: #388bfd;
  }
}

.btn-action-secondary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 7px 14px;
  font-size: 12.5px;
  font-weight: 500;
  border-radius: 6px;
  color: var(--color-btn-text, #24292f);
  background: var(--color-btn-bg, #ffffff);
  border: 1px solid var(--color-btn-border, #d0d7de);
  box-shadow: 0 1px 0 rgba(27, 31, 36, 0.04);
  cursor: pointer;
  transition: all 0.15s ease;

  &:hover {
    background: var(--color-btn-hover-bg, #f3f4f6);
    border-color: var(--color-btn-hover-border, #d0d7de);
    color: var(--color-accent-fg, #0969da);
  }

  .btn-svg {
    width: 14px;
    height: 14px;
  }

  .copied-check {
    color: #1a7f37;
  }
}

html.dark .btn-action-secondary {
  color: #c9d1d9;
  background: #21262d;
  border-color: #30363d;
  box-shadow: 0 1px 0 rgba(0, 0, 0, 0.1);

  &:hover {
    background: #30363d;
    border-color: #8b949e;
    color: #58a6ff;
  }

  .copied-check {
    color: #3fb950;
  }
}

/* 拖拽上传区 */
.dropzone-box {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 18px;
  border: 1.5px dashed var(--color-border-subtle, #d0d7de);
  border-radius: 6px;
  background: var(--surface-card);
  cursor: pointer;
  transition: border-color 0.15s ease, background-color 0.15s ease;
  user-select: none;

  &:hover,
  &.is-dragover {
    border-color: var(--color-accent-fg, #0969da);
    background: var(--color-canvas-subtle, #f6f8fa);
  }
}

html.dark .dropzone-box {
  border-color: #30363d;

  &:hover,
  &.is-dragover {
    border-color: #58a6ff;
    background: #21262d;
  }
}

.dropzone-icon-wrap {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--color-canvas-subtle, #f6f8fa);
  border: 1px solid var(--hairline);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--stone);

  .dropzone-svg {
    width: 18px;
    height: 18px;
  }
}

html.dark .dropzone-icon-wrap {
  background: #21262d;
  border-color: #30363d;
  color: #8b949e;
}

.dropzone-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.dropzone-primary-text {
  font-size: 13px;
  font-weight: 500;
  color: var(--charcoal);
}

.dropzone-browse-link {
  color: var(--color-accent-fg, #0969da);
  font-weight: 500;
  margin-left: 4px;
}

html.dark .dropzone-browse-link {
  color: #58a6ff;
}

.dropzone-secondary-text {
  font-size: 11.5px;
  color: var(--stone);
}

/* 已选文件卡片 */
.selected-file-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border: 1px solid var(--hairline);
  border-radius: 6px;
  background: var(--surface-card);
}

.file-icon-badge {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;

  svg {
    width: 16px;
    height: 16px;
  }
}

.file-info-col {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.file-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--charcoal);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-status {
  font-size: 11px;
  color: #059669;
  font-weight: 500;
}

.file-card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-ghost-sm {
  background: transparent;
  border: 1px solid var(--hairline);
  border-radius: 4px;
  font-size: 11.5px;
  color: var(--charcoal);
  padding: 3px 8px;
  cursor: pointer;
  transition: all 0.15s ease;

  &:hover {
    background: var(--surface-card);
    border-color: var(--stone);
  }
}

.btn-ghost-remove {
  background: transparent;
  border: none;
  width: 24px;
  height: 24px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--stone);
  transition: all 0.15s ease;

  &:hover {
    background: rgba(239, 68, 68, 0.1);
    color: #dc2626;
  }

  svg {
    width: 14px;
    height: 14px;
  }
}

/* 手动粘贴输入区 */
.import-manual-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.toggle-paste-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: transparent;
  border: none;
  padding: 0;
  font-size: 12px;
  color: var(--color-fg-muted, #57606a);
  cursor: pointer;
  width: fit-content;
  user-select: none;

  &:hover {
    color: var(--color-fg-default, #24292f);
  }

  .toggle-svg {
    width: 12px;
    height: 12px;
    transition: transform 0.2s ease;

    &.rotate-open {
      transform: rotate(180deg);
    }
  }
}

html.dark .toggle-paste-btn {
  color: #8b949e;

  &:hover {
    color: #c9d1d9;
  }
}

.paste-textarea-wrap {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.backup-textarea {
  width: 100%;
  box-sizing: border-box;
  min-height: 90px;
  padding: 8px 12px;
  border-radius: 6px;
  border: 1px solid var(--color-border-subtle, #d0d7de);
  background: var(--surface-card);
  color: var(--charcoal);
  font-family: var(--font-mono);
  font-size: 11.5px;
  line-height: 1.5;
  resize: vertical;

  &:focus {
    outline: none;
    border-color: var(--color-accent-fg, #0969da);
  }
}

html.dark .backup-textarea {
  border-color: #30363d;

  &:focus {
    border-color: #58a6ff;
  }
}

.paste-clear-row {
  display: flex;
  justify-content: flex-end;
}

.btn-clear-text {
  background: transparent;
  border: none;
  font-size: 11px;
  color: var(--stone);
  cursor: pointer;
  padding: 2px 4px;

  &:hover {
    color: #dc2626;
  }
}

/* 弹窗底部操作栏 */
.backup-dialog-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.status-ready-text {
  font-size: 12px;
  color: #059669;
  font-weight: 500;
}

.footer-right-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: auto;

  :deep(.el-button) {
    min-width: 80px;
    height: 32px;
    font-size: 12.5px;
    box-sizing: border-box;
  }
}
</style>

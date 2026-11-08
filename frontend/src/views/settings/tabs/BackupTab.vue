<template>
  <div class="backup-tab">
    <!-- Group 1: 导出配置备份 -->
    <div class="setting-group">
      <div class="subhead">
        <h3 class="subhead-title">导出配置备份</h3>
      </div>
      <p class="group-desc">
        将当前系统策略、定时调度参数与主播个性化规则导出为标准 JSON 配置文件，用于跨设备迁移或灾备留底。
      </p>

      <div class="export-actions-row">
        <!-- 下载配置文件按钮 -->
        <button
          type="button"
          class="btn-action btn-export"
          @click="emit('export-file')"
        >
          <svg class="btn-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="7 10 12 15 17 10"/>
            <line x1="12" y1="15" x2="12" y2="3"/>
          </svg>
          <span>下载配置文件</span>
        </button>

        <!-- 复制配置代码按钮 -->
        <button
          type="button"
          class="btn-action btn-copy"
          @click="handleCopy"
        >
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

    <!-- Group 2: 导入配置恢复 -->
    <div class="setting-group">
      <div class="subhead">
        <h3 class="subhead-title">导入配置恢复</h3>
      </div>
      <p class="group-desc">
        选择导出的 JSON 备份文件或直接粘贴配置代码，校验通过后将覆盖并应用至系统。
      </p>

      <div class="import-form-container">
        <!-- 安全提示卡片 (GitHub Callout) -->
        <div class="backup-callout">
          <svg class="callout-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="16" x2="12" y2="12"/>
            <line x1="12" y1="8" x2="12.01" y2="8"/>
          </svg>
          <span class="callout-text">
            导入配置将直接覆盖现有的全局合并规则与定时清理调度。建议在导入前先导出备份以防配置丢失。
          </span>
        </div>

        <!-- 隐藏的真实原生文件输入框 -->
        <input
          ref="importFileRef"
          type="file"
          accept=".json,application/json"
          style="display: none"
          @change="onFileChange"
        />

        <!-- 1. 未选择文件时的拖拽与点击上传区 (Dropzone) -->
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
            <div class="dropzone-secondary-text">支持 bililive-helper 导出的 .json 配置文件</div>
          </div>
        </div>

        <!-- 2. 已选择文件卡片 (Selected File Asset Card) -->
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
            <span class="file-status">已读取待应用</span>
          </div>
          <div class="file-card-actions">
            <button
              type="button"
              class="btn-ghost-sm"
              title="重新选择其他文件"
              @click="importFileRef?.click()"
            >
              重新选择
            </button>
            <button
              type="button"
              class="btn-ghost-remove"
              title="清除已选文件"
              @click="handleRemoveFile"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"/>
                <line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
            </button>
          </div>
        </div>

        <!-- 3. 手动粘贴 JSON 折叠展开区 -->
        <div class="import-manual-section">
          <button
            type="button"
            class="toggle-paste-btn"
            @click="showManualPaste = !showManualPaste"
          >
            <svg class="toggle-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" :class="{ 'rotate-open': showManualPaste }">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
            <span>{{ showManualPaste ? '收起配置代码输入框' : '或手动粘贴 JSON 配置代码' }}</span>
          </button>

          <transition name="paste-collapse">
            <div v-if="showManualPaste" class="paste-textarea-wrap">
              <textarea
                class="backup-textarea"
                :value="importJson"
                placeholder="在此直接粘贴导出的 JSON 配置字符串..."
                spellcheck="false"
                @input="emit('update:import-json', ($event.target as HTMLTextAreaElement).value)"
              />
              <div v-if="importJson" class="paste-clear-row">
                <button type="button" class="btn-clear-text" @click="emit('update:import-json', '')">清空内容</button>
              </div>
            </div>
          </transition>
        </div>

        <!-- 4. 提交导入动作栏 (Submit Action Bar) -->
        <div class="import-submit-row">
          <button
            type="button"
            class="btn-primary-import"
            :disabled="!hasValidImportContent"
            @click="emit('apply-import')"
          >
            <svg class="btn-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="20 6 9 17 4 12"/>
            </svg>
            <span>导入并应用配置</span>
          </button>
          <span v-if="!hasValidImportContent" class="import-hint-disabled">
            请先选择备份文件或粘贴配置代码
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { ElMessage } from "element-plus";

const props = defineProps<{
  selectedFileName: string;
  importJson: string;
}>();

const emit = defineEmits<{
  "export-file": [];
  "copy-json": [];
  "file-selected": [file: File];
  "clear-file": [];
  "update:import-json": [val: string];
  "apply-import": [];
}>();

const showManualPaste = ref(false);
const importFileRef = ref<HTMLInputElement | null>(null);
const isDragOver = ref(false);
const copiedRecently = ref(false);

const hasValidImportContent = computed(() => {
  return Boolean(props.selectedFileName || props.importJson.trim());
});

function handleCopy() {
  emit("copy-json");
  copiedRecently.value = true;
  setTimeout(() => {
    copiedRecently.value = false;
  }, 2000);
}

function handleRemoveFile() {
  emit("clear-file");
  emit("update:import-json", "");
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
    emit("file-selected", file);
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
  emit("file-selected", file);
}
</script>

<style scoped>
.backup-tab {
  display: flex;
  flex-direction: column;
  gap: 28px;
  width: 100%;
}

.setting-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.subhead {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--hairline);
}

.subhead-title {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal);
  margin: 0;
  letter-spacing: -0.01em;
}

.group-desc {
  font-size: 12.5px;
  color: var(--steel);
  line-height: 1.5;
  margin: 0;
}

/* Group 1: 导出按钮行 */
.export-actions-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  margin-top: 2px;
}

.btn-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  height: 32px;
  padding: 0 14px;
  font-size: 12.5px;
  font-weight: 500;
  border-radius: var(--r-sm, 4px);
  cursor: pointer;
  transition: all 0.15s ease;
  border: 1px solid var(--hairline-strong);
  background: var(--surface);
  color: var(--charcoal);
  user-select: none;
  white-space: nowrap;
}

.btn-action:hover {
  background: var(--highlight);
  border-color: var(--charcoal);
  color: var(--ink);
}

.btn-action:active {
  transform: translateY(0.5px);
}

.btn-export {
  /* Subtle primary accent tint for main export action */
  background: color-mix(in srgb, var(--primary) 8%, var(--surface));
  color: var(--charcoal);
  border-color: color-mix(in srgb, var(--primary) 25%, var(--hairline-strong));
  font-weight: 550;
}

.btn-export:hover {
  background: color-mix(in srgb, var(--primary) 14%, var(--surface));
  border-color: var(--primary);
  color: var(--ink);
}

.btn-copy .copied-check {
  color: var(--success, #059669);
}

.btn-svg {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}

/* Group 2: 导入容器 (最大宽度受限在 520px，避免宽屏无限横向拉伸) */
.import-form-container {
  display: flex;
  flex-direction: column;
  gap: 14px;
  max-width: 520px;
  width: 100%;
}

/* 提示卡片 (Callout Banner) */
.backup-callout {
  display: flex;
  align-items: flex-start;
  gap: 9px;
  padding: 9px 12px;
  background: color-mix(in srgb, var(--primary) 5%, var(--surface));
  border: 1px solid color-mix(in srgb, var(--primary) 18%, var(--hairline));
  border-radius: var(--r-sm, 4px);
}

.callout-icon {
  width: 15px;
  height: 15px;
  color: var(--primary);
  flex-shrink: 0;
  margin-top: 1.5px;
}

.callout-text {
  font-size: 11.5px;
  color: var(--charcoal);
  line-height: 1.5;
}

/* 拖拽与点击上传区 (Dropzone) */
.dropzone-box {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 18px;
  border: 1.5px dashed var(--hairline-strong);
  border-radius: var(--r-sm, 4px);
  background: var(--surface);
  cursor: pointer;
  transition: border-color 0.15s ease, background-color 0.15s ease;
  user-select: none;
}

.dropzone-box:hover,
.dropzone-box.is-dragover {
  border-color: var(--primary);
  background: color-mix(in srgb, var(--primary) 5%, var(--surface));
}

.dropzone-icon-wrap {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--highlight);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--steel);
  transition: all 0.15s ease;
}

.dropzone-box:hover .dropzone-icon-wrap,
.dropzone-box.is-dragover .dropzone-icon-wrap {
  background: color-mix(in srgb, var(--primary) 14%, transparent);
  color: var(--primary);
}

.dropzone-svg {
  width: 18px;
  height: 18px;
}

.dropzone-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.dropzone-primary-text {
  font-size: 12.5px;
  font-weight: 500;
  color: var(--charcoal);
  line-height: 1.4;
}

.dropzone-browse-link {
  color: var(--primary);
  text-decoration: underline;
  text-underline-offset: 2px;
  margin-left: 4px;
}

.dropzone-secondary-text {
  font-size: 11px;
  color: var(--steel);
  line-height: 1.3;
}

/* 已选择文件资产卡片 (Selected File Asset Card) */
.selected-file-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border: 1px solid var(--hairline-strong);
  border-radius: var(--r-sm, 4px);
  background: var(--surface);
}

.file-icon-badge {
  width: 32px;
  height: 32px;
  border-radius: var(--r-sm, 4px);
  background: color-mix(in srgb, var(--success, #059669) 12%, transparent);
  color: var(--success, #059669);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.file-icon-badge svg {
  width: 16px;
  height: 16px;
}

.file-info-col {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1;
}

.file-name {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--charcoal);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-status {
  font-size: 11px;
  color: var(--success, #059669);
  font-weight: 500;
}

.file-card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.btn-ghost-sm {
  background: transparent;
  border: 1px solid var(--hairline);
  border-radius: var(--r-xs, 3px);
  font-size: 11.5px;
  color: var(--charcoal);
  padding: 3px 8px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-ghost-sm:hover {
  background: var(--highlight);
  border-color: var(--hairline-strong);
}

.btn-ghost-remove {
  background: transparent;
  border: none;
  width: 24px;
  height: 24px;
  border-radius: var(--r-xs, 3px);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--steel);
  transition: all 0.15s ease;
  padding: 0;
}

.btn-ghost-remove:hover {
  background: color-mix(in srgb, var(--error, #e03131) 12%, transparent);
  color: var(--error, #e03131);
}

.btn-ghost-remove svg {
  width: 14px;
  height: 14px;
}

/* 手动粘贴 JSON 折叠区 */
.import-manual-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.toggle-paste-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  padding: 0;
  font-size: 12px;
  color: var(--steel);
  cursor: pointer;
  width: fit-content;
  transition: color 0.15s ease;
  user-select: none;
}

.toggle-paste-btn:hover {
  color: var(--charcoal);
}

.toggle-svg {
  width: 13px;
  height: 13px;
  transition: transform 0.2s ease;
}

.toggle-svg.rotate-open {
  transform: rotate(180deg);
}

.paste-textarea-wrap {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.backup-textarea {
  width: 100%;
  box-sizing: border-box;
  min-height: 110px;
  padding: 10px 12px;
  border-radius: var(--r-sm, 4px);
  border: 1px solid var(--hairline-strong);
  background: var(--surface);
  color: var(--charcoal);
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.5;
  resize: vertical;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.backup-textarea:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--primary) 15%, transparent);
}

.paste-clear-row {
  display: flex;
  justify-content: flex-end;
}

.btn-clear-text {
  background: transparent;
  border: none;
  font-size: 11px;
  color: var(--steel);
  cursor: pointer;
  padding: 2px 4px;
  transition: color 0.15s ease;
}

.btn-clear-text:hover {
  color: var(--error, #e03131);
}

/* 提交导入动作栏 (Submit Action Bar) */
.import-submit-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 4px;
}

.btn-primary-import {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 32px;
  padding: 0 16px;
  font-size: 12.5px;
  font-weight: 600;
  border-radius: var(--r-sm, 4px);
  cursor: pointer;
  transition: all 0.15s ease;
  background: var(--primary);
  color: #ffffff;
  border: 1px solid transparent;
  user-select: none;
  white-space: nowrap;
}

.btn-primary-import:hover:not(:disabled) {
  background: var(--primary-pressed, #1e40af);
}

.btn-primary-import:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.import-hint-disabled {
  font-size: 11.5px;
  color: var(--steel);
}

/* 动效过渡 */
.paste-collapse-enter-active,
.paste-collapse-leave-active {
  transition: all 0.2s ease;
}

.paste-collapse-enter-from,
.paste-collapse-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>

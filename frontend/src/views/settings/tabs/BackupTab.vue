<template>
  <div class="backup-tab">
    <div class="backup-grid">
      <!-- Export card -->
      <div class="backup-card">
        <div class="backup-card-header">
          <div class="ui-icon-box icon-box-blue"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg></div>
          <span class="backup-title">导出配置备份</span>
        </div>
        <p class="backup-desc">将系统当前的阈值策略、定时调度参数以及各主播的个性化截取与保护规则导出为 JSON 备份文件。</p>
        <div class="backup-spacer"></div>
        <div class="backup-action-group">
          <button class="btn-action-half btn-export" @click="emit('export-file')">
              <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
              <span>下载配置文件 (JSON)</span>
            </button>
          <button class="btn-action-half btn-copy" @click="emit('copy-json')">
              <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
              <span>复制配置代码</span>
            </button>
        </div>
      </div>

      <!-- Import card -->
      <div class="backup-card">
        <div class="backup-card-header">
          <div class="ui-icon-box icon-box-teal"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg></div>
          <span class="backup-title">导入配置恢复</span>
        </div>
        <p class="backup-desc">选择之前导出的 JSON 备份文件，导入并校验通过后将自动覆盖应用现有设置。</p>
        <input
          ref="importFileRef"
          type="file"
          accept=".json"
          style="display: none"
          @change="onFileChange"
        />
        <div class="import-file-box" @click="importFileRef?.click()">
          <svg class="import-file-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
          <span class="import-file-text">{{ selectedFileName || '点击选择本地 JSON 配置文件...' }}</span>
        </div>
        <div class="import-manual-toggle">
          <span class="toggle-link" @click="showManualPaste = !showManualPaste">
            <svg class="toggle-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" :class="{ 'rotate-open': showManualPaste }"><polyline points="6 9 12 15 18 9"/></svg>
            {{ showManualPaste ? '收起手动粘贴' : '或手动粘贴 JSON 代码' }}
          </span>
          <textarea
            v-if="showManualPaste"
            class="backup-textarea"
            :value="importJson"
            placeholder="在此直接粘贴配置 JSON 字符串..."
            @input="emit('update:import-json', ($event.target as HTMLTextAreaElement).value)"
          />
        </div>
        <div class="backup-spacer"></div>
        <div class="backup-card-footer">
          <button
            class="btn-primary btn-full"
            :disabled="!selectedFileName && !importJson"
            @click="emit('apply-import')"
          >
            导入并应用配置
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

defineProps<{
  selectedFileName: string;
  importJson: string;
}>();

const emit = defineEmits<{
  "export-file": [];
  "copy-json": [];
  "file-selected": [file: File];
  "update:import-json": [val: string];
  "apply-import": [];
}>();

const showManualPaste = ref(false);
const importFileRef = ref<HTMLInputElement | null>(null);

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (file) {
    emit("file-selected", file);
  }
  input.value = "";
}
</script>

<style scoped>
.backup-tab {
  width: 100%;
}
.backup-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  padding: 8px 16px 20px 24px;
}
.backup-card {
  display: flex;
  flex-direction: column;
  background: var(--surface);
  border: 1px solid var(--hairline);
  border-radius: var(--r-md);
  padding: 20px;
}
.backup-card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}
.backup-icon {
  font-size: 18px;
}
.backup-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--charcoal);
}
.backup-desc {
  font-size: 13px;
  color: var(--slate);
  line-height: 1.5;
  margin-bottom: 16px;
}
.backup-spacer {
  flex: 1;
}
.backup-action-group {
  display: flex;
  gap: 12px;
  margin-top: 16px;
}
.btn-action-half {
  flex: 1;
  height: 38px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  border-radius: var(--r-md, 6px);
  box-sizing: border-box;
  cursor: pointer;
  transition: all 0.15s ease;
}
.btn-action-half.btn-export {
  background: rgba(37, 99, 235, 0.08);
  color: #1d4ed8;
  border: 1px solid rgba(37, 99, 235, 0.20);
  font-weight: 600;
}
.btn-action-half.btn-export:hover {
  background: rgba(37, 99, 235, 0.14);
  border-color: rgba(37, 99, 235, 0.35);
  color: #1e40af;
  transform: translateY(-1px);
}
.btn-action-half.btn-copy {
  background: var(--surface, #fafafa);
  border: 1px solid var(--hairline, #dcdfe6);
  color: var(--charcoal, #333);
}
.btn-action-half.btn-copy:hover {
  background: var(--highlight, #f0f0f0);
  border-color: var(--steel, #aaa);
}
.import-file-box {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--canvas, #fff);
  border: 1px dashed var(--hairline, #dcdfe6);
  padding: 12px 16px;
  border-radius: var(--r-sm, 4px);
  cursor: pointer;
  margin: 12px 0 8px;
  transition: border-color 0.15s;
}
.import-file-box:hover {
  border-color: #409eff;
}
.import-file-icon {
  font-size: 18px;
}
.import-file-text {
  font-size: 13px;
  color: var(--slate, #606266);
}
.import-manual-toggle {
  margin: 6px 0 12px;
}
.toggle-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--steel, #888);
  cursor: pointer;
  user-select: none;
}
.toggle-link:hover {
  color: var(--ink, #333);
}
.toggle-svg {
  width: 12px;
  height: 12px;
  transition: transform 0.2s ease;
  transform: rotate(-90deg);
  &.rotate-open {
    transform: rotate(0deg);
  }
}
.backup-textarea {
  width: 100%;
  height: 120px;
  padding: 10px;
  font-family: var(--font-mono);
  font-size: 12px;
  border: 1px solid var(--hairline);
  border-radius: var(--r-md);
  background: #fff;
  color: var(--ink);
  resize: vertical;
  margin-top: 8px;
  box-sizing: border-box;
}
.backup-textarea:focus {
  outline: none;
  border-color: var(--charcoal);
}
.btn-full {
  width: 100%;
  height: 38px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 500;
  border-radius: var(--r-md);
  border: none;
  cursor: pointer;
  background: var(--primary);
  color: #fff;
  transition: opacity 0.15s;
}
.btn-full:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .backup-grid {
    grid-template-columns: 1fr;
    padding: 8px 4px 16px 8px;
  }
}

.backup-card-header {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}
.btn-svg-sm {
  width: 14px;
  height: 14px;
  margin-right: 6px;
  vertical-align: -2px;
}
.import-file-svg {
  width: 16px;
  height: 16px;
  margin-right: 6px;
  vertical-align: -3px;
  color: var(--steel);
}


html.dark {
  .backup-card {
    background: #18181b;
    border-color: #27272a;
  }
  .btn-action-half.btn-copy {
    background: #1f1f23;
    border-color: #27272a;
    color: #e4e4e7;
    &:hover {
      background: #27272a;
      border-color: #3f3f46;
      color: #ffffff;
    }
  }
  .import-file-box {
    background: #141416;
    border-color: #27272a;
    &:hover {
      border-color: #3b82f6;
    }
  }
  .import-file-text {
    color: #a1a1aa;
  }
  .backup-textarea {
    background: #141416;
    border-color: #27272a;
    color: #f4f4f5;
    &:focus {
      border-color: #3b82f6;
    }
  }
  .toggle-link {
    color: #71717a;
    &:hover {
      color: #e4e4e7;
    }
  }
}
html.dark .btn-action-half.btn-export {
  background: rgba(37, 99, 235, 0.14) !important;
  border-color: rgba(59, 130, 246, 0.30) !important;
  color: #60a5fa !important;
}
html.dark .btn-action-half.btn-export:hover {
  background: rgba(37, 99, 235, 0.22) !important;
  border-color: #3b82f6 !important;
  color: #93c5fd !important;
}
</style>

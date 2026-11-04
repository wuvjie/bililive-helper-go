<template>
  <div class="backup-tab">
    <div class="backup-grid">
      <!-- Export card -->
      <div class="backup-card">
        <div class="backup-card-header">
          <span class="backup-icon">📤</span>
          <span class="backup-title">导出配置备份</span>
        </div>
        <p class="backup-desc">将系统当前的阈值策略、定时调度参数以及各主播的个性化截取与保护规则导出为 JSON 备份文件。</p>
        <div class="backup-spacer"></div>
        <div class="backup-action-group">
          <button class="btn-action-half btn-export" @click="emit('export-file')">📥 下载配置文件 (.json)</button>
          <button class="btn-action-half btn-copy" @click="emit('copy-json')">📋 复制配置代码</button>
        </div>
      </div>

      <!-- Import card -->
      <div class="backup-card">
        <div class="backup-card-header">
          <span class="backup-icon">📥</span>
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
          <span class="import-file-icon">📂</span>
          <span class="import-file-text">{{ selectedFileName || '点击选择本地 JSON 备份文件...' }}</span>
        </div>
        <div class="import-manual-toggle">
          <span class="toggle-link" @click="showManualPaste = !showManualPaste">
            {{ showManualPaste ? '▾ 收起手动粘贴' : '▸ 或手动粘贴 JSON 代码' }}
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
  font-size: 15px;
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
  background: var(--primary);
  color: var(--on-primary);
  border: 1px solid var(--primary);
}
.btn-action-half.btn-export:hover {
  background: var(--primary-pressed);
  border-color: var(--primary-pressed);
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
  font-size: 12px;
  color: var(--steel, #888);
  cursor: pointer;
  user-select: none;
}
.toggle-link:hover {
  color: var(--ink, #333);
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
</style>

<template>
  <div class="backup-tab">
    <!-- Block 1: 导出配置备份 -->
    <div class="backup-block">
      <div class="block-info">
        <span class="block-title">导出配置备份</span>
        <p class="block-desc">
          将系统策略、调度参数与主播个性化规则导出为标准 JSON 备份
        </p>
      </div>
      <div class="block-actions">
        <button type="button" class="btn-backup btn-export" @click="emit('export-file')">
          <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="7 10 12 15 17 10"/>
            <line x1="12" y1="15" x2="12" y2="3"/>
          </svg>
          <span>下载配置文件</span>
        </button>
        <button type="button" class="btn-backup btn-copy" @click="emit('copy-json')">
          <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
          </svg>
          <span>复制配置代码</span>
        </button>
      </div>
    </div>

    <div class="block-divider" />

    <!-- Block 2: 导入配置恢复 -->
    <div class="backup-block">
      <div class="block-info">
        <span class="block-title">导入配置恢复</span>
        <p class="block-desc">
          选择导出的 JSON 备份文件，校验通过后自动覆盖并应用至系统
        </p>
      </div>
      <div class="block-form">
        <input
          ref="importFileRef"
          type="file"
          accept=".json"
          style="display: none"
          @change="onFileChange"
        />
        <div class="import-file-box" @click="importFileRef?.click()">
          <svg class="import-file-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
          </svg>
          <span class="import-file-text">{{ selectedFileName || '点击选择本地 JSON 配置文件...' }}</span>
        </div>

        <div class="import-manual-toggle">
          <span class="toggle-link" @click="showManualPaste = !showManualPaste">
            <svg class="toggle-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" :class="{ 'rotate-open': showManualPaste }">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
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

        <div class="import-submit-row">
          <button
            type="button"
            class="btn-backup btn-import-submit"
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
  display: flex;
  flex-direction: column;
  gap: 20px;
  width: 100%;
}

.backup-block {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.block-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.block-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--charcoal);
}

.block-desc {
  font-size: 12.5px;
  color: var(--steel);
  line-height: 1.5;
  margin: 0;
}

.block-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.btn-backup {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 34px;
  padding: 0 16px;
  font-size: 12.5px;
  font-weight: 500;
  border-radius: var(--r-sm);
  cursor: pointer;
  transition: all 0.15s ease;
  line-height: 1.4;
  border: 1px solid var(--hairline);
  background: var(--surface);
  color: var(--charcoal);
}

.btn-backup:hover:not(:disabled) {
  background: var(--highlight);
  border-color: var(--charcoal);
  color: var(--ink);
}

.btn-export {
  background: color-mix(in srgb, #2563eb 8%, transparent);
  color: #1d4ed8;
  border-color: color-mix(in srgb, #2563eb 20%, transparent);
  font-weight: 600;
}

.btn-export:hover:not(:disabled) {
  background: #2563eb;
  color: #ffffff;
}

.block-divider {
  height: 1px;
  background: var(--hairline);
  margin: 4px 0;
}

.block-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.import-file-box {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--surface);
  border: 1px dashed var(--hairline-strong);
  padding: 12px 16px;
  border-radius: var(--r-sm);
  cursor: pointer;
  transition: border-color 0.15s ease, background 0.15s ease;
}

.import-file-box:hover {
  border-color: var(--primary);
  background: var(--highlight);
}

.import-file-svg {
  width: 18px;
  height: 18px;
  color: var(--steel);
  flex-shrink: 0;
}

.import-file-text {
  font-size: 13px;
  color: var(--slate);
}

.import-manual-toggle {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.toggle-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--primary);
  cursor: pointer;
  user-select: none;
}

.toggle-link:hover {
  text-decoration: underline;
}

.toggle-svg {
  width: 14px;
  height: 14px;
  transition: transform 0.2s ease;
}

.toggle-svg.rotate-open {
  transform: rotate(180deg);
}

.backup-textarea {
  width: 100%;
  box-sizing: border-box;
  min-height: 120px;
  padding: 10px 12px;
  border-radius: var(--r-sm);
  border: 1px solid var(--hairline);
  background: var(--surface);
  color: var(--charcoal);
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.5;
  resize: vertical;
}

.import-submit-row {
  display: flex;
  justify-content: flex-start;
  margin-top: 4px;
}

.btn-import-submit {
  background: var(--primary);
  color: #ffffff;
  border-color: var(--primary);
  font-weight: 600;
}

.btn-import-submit:hover:not(:disabled) {
  background: var(--primary-pressed, #047857);
}

.btn-import-submit:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.btn-svg-sm {
  width: 14px;
  height: 14px;
}
</style>

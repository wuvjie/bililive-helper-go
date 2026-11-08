<template>
  <div class="backup-tab">
    <div class="backup-grid">
      <!-- Export card -->
      <div class="backup-card">
        <div class="backup-card-top">
          <div class="backup-card-header">
            <span class="sub-title-text">导出配置备份</span>
          </div>
          <p class="backup-desc">将系统当前的阈值策略、定时调度参数以及各主播的个性化截取与保护规则导出为 JSON 备份文件。</p>
          <div class="backup-scope-list">
            <div class="scope-item">
              <svg class="scope-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
              <span>全局清理阈值、重连容差与静置时间</span>
            </div>
            <div class="scope-item">
              <svg class="scope-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
              <span>定时巡检周期与免打扰避让时段</span>
            </div>
            <div class="scope-item">
              <svg class="scope-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
              <span>所有主播专属规则与裁剪区间</span>
            </div>
          </div>
        </div>
        <div class="backup-action-group">
          <button type="button" class="btn-action-half btn-export" @click="emit('export-file')">
            <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            <span>下载配置文件</span>
          </button>
          <button type="button" class="btn-action-half btn-copy" @click="emit('copy-json')">
            <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
            <span>复制配置代码</span>
          </button>
        </div>
      </div>

      <!-- Import card -->
      <div class="backup-card">
        <div class="backup-card-top">
          <div class="backup-card-header">
            <span class="sub-title-text">导入配置恢复</span>
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
        </div>
        <div class="backup-action-group">
          <button
            type="button"
            class="btn-full btn-primary"
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
  align-items: stretch;
}
.backup-card {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  height: 100%;
  background: var(--surface, #f4f3f0);
  border: 1px solid var(--hairline, #e6e4df);
  border-radius: var(--r-md, 8px);
  padding: 20px;
  box-sizing: border-box;
}
.backup-card-top {
  display: flex;
  flex-direction: column;
  flex: 1;
}
.backup-scope-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin: 10px 0 16px 0;
  padding: 12px;
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--hairline);
  border-radius: var(--r-sm, 6px);
}
.scope-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--slate);
}
.scope-svg {
  width: 13px;
  height: 13px;
  color: var(--primary, #059669);
  flex-shrink: 0;
}
.backup-card-header {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--hairline, #e6e4df);
}
.sub-title-text {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal, #18181b);
}
.backup-desc {
  font-size: 12.5px;
  color: var(--steel, #787671);
  line-height: 1.5;
  margin: 0 0 14px 0;
  min-height: 38px;
}
.backup-action-group {
  display: flex;
  gap: 10px;
  margin-top: 14px;
}
.btn-action-half {
  flex: 1;
  height: 36px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 12.5px;
  font-weight: 500;
  border-radius: var(--r-sm, 6px);
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
  background: var(--surface-card, #ffffff);
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
  background: var(--surface-card, #fff);
  border: 1px dashed var(--hairline, #dcdfe6);
  padding: 10px 14px;
  border-radius: var(--r-sm, 6px);
  cursor: pointer;
  margin: 4px 0 8px;
  transition: border-color 0.15s;
}
.import-file-box:hover {
  border-color: var(--primary, #2563eb);
}
.import-file-svg {
  width: 16px;
  height: 16px;
  color: var(--steel, #787671);
}
.import-file-text {
  font-size: 12.5px;
  color: var(--slate, #606266);
}
.import-manual-toggle {
  margin: 4px 0 6px;
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
  height: 100px;
  padding: 8px 10px;
  font-family: var(--font-mono);
  font-size: 12px;
  border: 1px solid var(--hairline);
  border-radius: var(--r-sm, 6px);
  background: var(--surface-card, #ffffff);
  color: var(--ink);
  resize: vertical;
  margin-top: 8px;
  box-sizing: border-box;
}
.backup-textarea:focus {
  outline: none;
  border-color: var(--primary, #2563eb);
}
.btn-full {
  width: 100%;
  height: 36px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 12.5px;
  font-weight: 500;
  border-radius: var(--r-sm, 6px);
  border: none;
  cursor: pointer;
  background: var(--primary, #2563eb);
  color: #fff;
  transition: all 0.15s ease;
}
.btn-full:disabled {
  background: var(--surface, #f4f4f5) !important;
  color: var(--stone, #a1a1aa) !important;
  border: 1px solid var(--hairline, #e2e8f0) !important;
  opacity: 1 !important;
  cursor: not-allowed;
  box-shadow: none !important;
}
.btn-svg-sm {
  width: 14px;
  height: 14px;
}

@media (max-width: 768px) {
  .backup-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }
}

html.dark {
  .backup-card {
    background: #18181b;
    border-color: #27272a;
  }
  .backup-card-header {
    border-color: #27272a;
  }
  .sub-title-text {
    color: #f4f4f5;
  }
  .backup-desc {
    color: #a1a1aa;
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
  .btn-action-half.btn-export {
    background: rgba(37, 99, 235, 0.14) !important;
    border-color: rgba(59, 130, 246, 0.30) !important;
    color: #60a5fa !important;
  }
  .btn-action-half.btn-export:hover {
    background: rgba(37, 99, 235, 0.22) !important;
    border-color: #3b82f6 !important;
    color: #93c5fd !important;
  }
  .btn-full:disabled {
    background: #27272a !important;
    color: #71717a !important;
    border-color: #3f3f46 !important;
  }
}
</style>

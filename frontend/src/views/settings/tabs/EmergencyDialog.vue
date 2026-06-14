<template>
  <el-dialog
    :model-value="visible"
    title="紧急清理"
    width="440px"
    class="emergency-dialog"
    :close-on-click-modal="!sse.isRunning.value"
    destroy-on-close
    @close="handleClose"
  >
    <div v-if="!sse.lines.value.length" class="emergency-form">
      <div class="emergency-title-row">
        <span class="emergency-dot"></span>
        <span class="emergency-title">紧急磁盘清理</span>
      </div>
      <p class="emergency-desc">将临时降低磁盘使用率目标阈值，强制清理到指定百分比以下。清理结束后阈值自动恢复。</p>
      <div class="emergency-input-row">
        <span class="emergency-label">目标磁盘百分比</span>
        <el-input-number :model-value="targetPct" :min="10" :max="99" :step="5" @update:model-value="$emit('update:targetPct', $event)" />
        <span class="emergency-unit">%</span>
      </div>
    </div>
    <pre v-else class="emergency-terminal" v-loading="loading">{{ sse.lines.value.map((l) => l.text).join("\n") }}</pre>
    <template #footer>
      <button v-if="!sse.lines.value.length" class="emergency-confirm" :disabled="loading" @click="$emit('confirm')">确认紧急清理</button>
      <button v-else-if="sse.isRunning.value" class="emergency-confirm" @click="sse.abort()">中止</button>
      <button v-else class="btn-primary" @click="handleClose">关闭</button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import type { useSSE } from "@/utils/sse";

const props = defineProps<{
  visible: boolean;
  targetPct: number;
  loading: boolean;
  sse: ReturnType<typeof useSSE>;
}>();

const emit = defineEmits<{
  "update:visible": [value: boolean];
  "update:targetPct": [value: number];
  confirm: [];
  close: [];
}>();

function handleClose() {
  props.sse.abort();
  emit("update:visible", false);
  emit("close");
}
</script>

<style scoped>
.emergency-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.emergency-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.emergency-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #e03131;
}
.emergency-title {
  font-weight: 600;
  font-size: 15px;
}
.emergency-desc {
  color: #666;
  font-size: 13px;
  line-height: 1.6;
}
.emergency-input-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.emergency-label {
  font-size: 13px;
  color: #333;
}
.emergency-unit {
  font-size: 13px;
  color: #888;
}
.emergency-terminal {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 16px;
  border-radius: 8px;
  font-size: 12px;
  line-height: 1.6;
  max-height: 300px;
  overflow-y: auto;
  font-family: var(--font-mono, "JetBrains Mono", monospace);
}
.emergency-confirm {
  background: #e03131;
  color: #fff;
  border: none;
  padding: 8px 20px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
}
.emergency-confirm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.btn-primary {
  background: #409eff;
  color: #fff;
  border: none;
  padding: 8px 20px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
}
</style>

<style>
.emergency-dialog .el-dialog__header {
  display: none;
}
.emergency-dialog .el-dialog__body {
  padding: 20px;
}
.emergency-dialog .el-dialog__footer {
  padding: 0 20px 16px;
  text-align: right;
}
</style>

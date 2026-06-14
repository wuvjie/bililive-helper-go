<template>
  <el-form label-width="144px" label-position="left" v-loading="loading" class="settings-form">
    <div class="section-divider">🔄 自动合并</div>
    <el-form-item label="是否启用">
      <el-switch :model-value="form.merge_enabled" @update:model-value="$emit('update:form', { ...form, merge_enabled: $event })" />
    </el-form-item>
    <el-form-item label="执行间隔(分钟)">
      <el-input-number :model-value="form.merge_interval" :min="10" :max="1440" @update:model-value="$emit('update:form', { ...form, merge_interval: $event })" />
    </el-form-item>
    <div class="section-divider">🧹 自动清理</div>
    <el-form-item label="是否启用">
      <el-switch :model-value="form.clean_enabled" @update:model-value="$emit('update:form', { ...form, clean_enabled: $event })" />
    </el-form-item>
    <el-form-item label="执行间隔(分钟)">
      <el-input-number :model-value="form.clean_interval" :min="10" :max="1440" @update:model-value="$emit('update:form', { ...form, clean_interval: $event })" />
    </el-form-item>
    <div class="section-divider">⏳ 备份窗口（暂停任务）</div>
    <el-form-item label="开始时间">
      <el-time-picker :model-value="backupStart" format="HH:mm" value-format="HH:mm" style="width: 160px" @update:model-value="$emit('update:backupStart', $event)" />
    </el-form-item>
    <el-form-item label="结束时间">
      <el-time-picker :model-value="backupEnd" format="HH:mm" value-format="HH:mm" style="width: 160px" @update:model-value="$emit('update:backupEnd', $event)" />
    </el-form-item>
    <div class="backup-hint">支持跨午夜，如 23:00 - 06:00</div>
    <el-form-item>
      <el-button type="primary" :loading="saving" style="width: 128px" @click="$emit('save')">保存计划</el-button>
    </el-form-item>
    <div class="section-divider">🎮 手动触发</div>
    <el-form-item>
      <el-button :disabled="taskRunning" style="width: 148px" @click="$emit('trigger', 'merge')" class="ghost-trigger"> 🔄 立即执行合并 </el-button>
      <el-button :disabled="taskRunning" style="width: 148px" @click="$emit('trigger', 'clean')" class="ghost-trigger"> 🧹 立即执行清理 </el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
export interface ScheduleForm {
  merge_enabled: boolean;
  clean_enabled: boolean;
  merge_interval: number;
  clean_interval: number;
}

defineProps<{
  form: ScheduleForm;
  backupStart: string;
  backupEnd: string;
  saving: boolean;
  loading: boolean;
  taskRunning: boolean;
}>();

defineEmits<{
  "update:form": [value: ScheduleForm];
  "update:backupStart": [value: string];
  "update:backupEnd": [value: string];
  save: [];
  trigger: [task: "merge" | "clean"];
}>();
</script>

<style scoped>
.section-divider {
  font-size: 13px;
  font-weight: 600;
  color: #333;
  padding: 16px 0 8px;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 8px;
}
.section-divider:first-child {
  padding-top: 0;
}
.backup-hint {
  color: #aaa;
  font-size: 12px;
  margin: -4px 0 8px 144px;
}
.ghost-trigger {
  background: transparent;
  border: 1px solid #dcdfe6;
  color: #606266;
}
.ghost-trigger:hover {
  border-color: #409eff;
  color: #409eff;
}
.settings-form :deep(.el-form-item__label) {
  color: #666;
  font-size: 13px;
}
</style>

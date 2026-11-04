<template>
  <el-form label-width="144px" label-position="left" v-loading="loading" class="settings-form">
    <div class="section-divider">🔄 自动合并</div>
    <el-form-item label="是否启用">
      <el-switch :model-value="form.merge_enabled" @update:model-value="emit('update:form', { ...form, merge_enabled: !!$event })" />
    </el-form-item>
    <el-form-item label="执行间隔(分钟)">
      <el-input-number :model-value="form.merge_interval" :min="10" :max="1440" @update:model-value="emit('update:form', { ...form, merge_interval: Number($event) })" />
    </el-form-item>
    <div class="section-divider">🧹 自动清理</div>
    <el-form-item label="是否启用">
      <el-switch :model-value="form.clean_enabled" @update:model-value="emit('update:form', { ...form, clean_enabled: !!$event })" />
    </el-form-item>
    <el-form-item label="执行间隔(分钟)">
      <el-input-number :model-value="form.clean_interval" :min="10" :max="1440" @update:model-value="emit('update:form', { ...form, clean_interval: Number($event) })" />
    </el-form-item>
    <div class="section-divider">🌙 静默避让时段（免打扰）</div>
    <el-form-item label="开始时间">
      <el-time-picker :model-value="backupStart" format="HH:mm" value-format="HH:mm" style="width: 160px" @update:model-value="emit('update:backupStart', $event ?? '')" />
    </el-form-item>
    <el-form-item label="结束时间">
      <el-time-picker :model-value="backupEnd" format="HH:mm" value-format="HH:mm" style="width: 160px" @update:model-value="emit('update:backupEnd', $event ?? '')" />
    </el-form-item>
    <div class="backup-hint">在此时间段内暂停自动合并与清理任务，避免夜间或高峰期高 CPU 占用，支持跨午夜（如 23:00 - 06:00）</div>
    <el-form-item>
      <el-button type="primary" :loading="saving" style="width: 128px" @click="emit('save')">保存计划</el-button>
    </el-form-item>
    <div class="section-divider">⚡ 即时执行与终端</div>
    <div class="schedule-guide-card">
      <span class="guide-icon">💡</span>
      <div class="guide-text">
        需要立即手动触发全库合并、清理或针对特定主播执行任务，并实时查看流式终端输出？请前往
        <router-link to="/tasks" class="guide-link">任务中心 ↗</router-link>
      </div>
    </div>
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
  taskRunning?: boolean;
}>();

const emit = defineEmits<{
  "update:form": [value: ScheduleForm];
  "update:backupStart": [value: string];
  "update:backupEnd": [value: string];
  save: [];
  trigger?: [task: "merge" | "clean"];
}>();
</script>

<style scoped>
.section-divider {
  font-size: 13px;
  font-weight: 600;
  color: var(--charcoal, #333);
  padding: 16px 0 8px;
  border-bottom: 1px solid var(--hairline, #f0f0f0);
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
.schedule-guide-card {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--surface, #fafafa);
  border: 1px dashed var(--hairline, #e0e0e0);
  border-radius: var(--r-sm, 6px);
  padding: 12px 16px;
  margin-top: 4px;
  margin-bottom: 12px;
}
.guide-icon {
  font-size: 16px;
}
.guide-text {
  font-size: 13px;
  color: var(--slate, #666);
  line-height: 1.5;
}
.guide-link {
  color: #448361;
  font-weight: 600;
  text-decoration: none;
  margin-left: 4px;
}
.guide-link:hover {
  text-decoration: underline;
}
.settings-form :deep(.el-form-item__label) {
  color: #666;
  font-size: 13px;
}
</style>

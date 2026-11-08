<template>
  <div class="schedule-tab">
    <el-form label-width="140px" label-position="left" v-loading="loading" class="settings-form">
      <!-- Section 1: 自动合并计划 -->
      <div class="form-section-card">
        <div class="category-divider">
          <span class="category-label">分片合并调度</span>
          <span class="category-line" />
        </div>
        <div class="section-body">
          <el-form-item label="开启自动合并">
            <el-switch
              :model-value="form.merge_enabled"
              @update:model-value="emit('update:form', { ...form, merge_enabled: !!$event })"
            />
            <div class="field-hint">周期性自动扫描各主播目录并合并切片</div>
          </el-form-item>

          <el-form-item label="合并调度周期">
            <div class="inline-unit-row">
              <el-input-number
                :model-value="form.merge_interval"
                :min="10"
                :max="1440"
                :step="15"
                class="compact-counter"
                @update:model-value="emit('update:form', { ...form, merge_interval: Number($event) })"
              />
              <span class="unit-text">分钟</span>
            </div>
            <div class="field-hint">任务执行完毕后等待指定间隔再启动下次合并（推荐 360 分钟）</div>
          </el-form-item>
        </div>
      </div>

      <!-- Section 2: 自动清理计划 -->
      <div class="form-section-card">
        <div class="category-divider">
          <span class="category-label">磁盘清理巡检</span>
          <span class="category-line" />
        </div>
        <div class="section-body">
          <el-form-item label="开启自动清理">
            <el-switch
              :model-value="form.clean_enabled"
              @update:model-value="emit('update:form', { ...form, clean_enabled: !!$event })"
            />
            <div class="field-hint">定期巡检磁盘使用率，达警戒水位时自动触发清理</div>
          </el-form-item>

          <el-form-item label="巡检调度周期">
            <div class="inline-unit-row">
              <el-input-number
                :model-value="form.clean_interval"
                :min="10"
                :max="1440"
                :step="30"
                class="compact-counter"
                @update:model-value="emit('update:form', { ...form, clean_interval: Number($event) })"
              />
              <span class="unit-text">分钟</span>
            </div>
            <div class="field-hint">巡检器检查磁盘可用水位的轮询间隔（推荐 720 分钟）</div>
          </el-form-item>
        </div>
      </div>

      <!-- Section 3: 免打扰避让时段 -->
      <div class="form-section-card">
        <div class="category-divider">
          <span class="category-label">夜间避让时段</span>
          <span class="category-line" />
        </div>
        <div class="section-body">
          <el-form-item label="避让时间范围">
            <div class="time-range-row">
              <el-time-picker
                :model-value="backupStart"
                format="HH:mm"
                value-format="HH:mm"
                placeholder="开始时间"
                class="time-picker-box"
                @update:model-value="emit('update:backupStart', $event ?? '')"
              />
              <span class="time-sep">至</span>
              <el-time-picker
                :model-value="backupEnd"
                format="HH:mm"
                value-format="HH:mm"
                placeholder="结束时间"
                class="time-picker-box"
                @update:model-value="emit('update:backupEnd', $event ?? '')"
              />
            </div>
            <div class="field-hint">在此时间段内暂停自动合并与清理任务，避免夜间或高峰期高 CPU 占用，支持跨午夜（如 23:00 - 06:00）。</div>
          </el-form-item>
        </div>
      </div>
      <!-- 单一行动权威法则：由外层吸底浮动栏统一保存与放弃，消除表单内同义重复按钮 -->
    </el-form>
  </div>
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
  saving?: boolean;
  loading: boolean;
}>();

const emit = defineEmits<{
  "update:form": [value: ScheduleForm];
  "update:backupStart": [value: string];
  "update:backupEnd": [value: string];
  save?: [];
}>();
</script>

<style scoped>
.schedule-tab {
  width: 100%;
}
.settings-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.form-section-card {
  background: transparent;
  border: none;
  padding: 0 0 24px 0;
  margin-bottom: 24px;
  border-bottom: 1px solid var(--hairline);
}
.form-section-card:last-child {
  padding-bottom: 0;
  margin-bottom: 0;
  border-bottom: none;
}
.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--charcoal);
  padding-bottom: 12px;
  border-bottom: 1px solid var(--hairline);
  margin-bottom: 16px;
}
.section-icon {
  font-size: 16px;
}
.section-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.field-hint {
  width: 100%;
  font-size: 12px;
  color: var(--steel);
  line-height: 1.5;
  margin-top: 6px;
}
.inline-unit-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  white-space: nowrap;
}
.compact-counter {
  width: 120px;
}
.unit-text {
  font-size: 13px;
  color: var(--slate);
  white-space: nowrap;
  flex-shrink: 0;
  display: inline-block;
}
.time-range-row {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.time-picker-box {
  width: 150px !important;
}
.time-sep {
  font-size: 13px;
  color: var(--steel);
}

@media (max-width: 640px) {
  .settings-form :deep(.el-form-item) {
    flex-direction: column;
    align-items: flex-start;
  }
  .settings-form :deep(.el-form-item__label) {
    width: 100% !important;
    margin-bottom: 6px;
    height: auto;
    line-height: 1.4;
  }
  .settings-form :deep(.el-form-item__content) {
    width: 100%;
  }
  .time-picker-box {
    width: 100% !important;
  }
}
</style>

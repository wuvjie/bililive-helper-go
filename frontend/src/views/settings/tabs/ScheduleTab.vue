<template>
  <div class="schedule-tab">
    <el-form label-width="150px" label-position="left" v-loading="loading" class="settings-form">
      <!-- Section 1: 自动合并计划 -->
      <div class="form-section-card">
        <div class="section-title">
          <span class="section-icon">🔄</span>
          <span>自动分片合并计划</span>
        </div>
        <div class="section-body">
          <el-form-item label="开启自动合并">
            <el-switch
              :model-value="form.merge_enabled"
              @update:model-value="emit('update:form', { ...form, merge_enabled: !!$event })"
            />
            <div class="field-hint">开启后，后台调度器将周期性扫描各主播目录，自动合并已完播的视频分片切片。</div>
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
            <div class="field-hint">任务执行完毕后，等待指定间隔自动启动下一次扫描合并（推荐 360 分钟 / 6 小时）。</div>
          </el-form-item>
        </div>
      </div>

      <!-- Section 2: 自动清理计划 -->
      <div class="form-section-card">
        <div class="section-title">
          <span class="section-icon">🧹</span>
          <span>自动磁盘巡检与清理计划</span>
        </div>
        <div class="section-body">
          <el-form-item label="开启自动清理">
            <el-switch
              :model-value="form.clean_enabled"
              @update:model-value="emit('update:form', { ...form, clean_enabled: !!$event })"
            />
            <div class="field-hint">开启后，后台调度器将定期巡检磁盘使用率；仅当达到触发警戒水位时才执行清理。</div>
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
            <div class="field-hint">巡检器检查磁盘可用空间的轮询间隔（推荐 720 分钟 / 12 小时）。</div>
          </el-form-item>
        </div>
      </div>

      <!-- Section 3: 免打扰避让时段 -->
      <div class="form-section-card">
        <div class="section-title">
          <span class="section-icon">🌙</span>
          <span>免打扰避让时段</span>
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

      <!-- Action Button -->
      <div class="schedule-actions">
        <el-button type="primary" :loading="saving" class="btn-save-schedule" @click="emit('save')">
          保存调度计划
        </el-button>
      </div>
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
  saving: boolean;
  loading: boolean;
}>();

const emit = defineEmits<{
  "update:form": [value: ScheduleForm];
  "update:backupStart": [value: string];
  "update:backupEnd": [value: string];
  save: [];
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
  background: var(--surface);
  border: 1px solid var(--hairline);
  border-radius: var(--r-md);
  padding: 18px 20px;
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
.schedule-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0 16px 0;
}
.btn-save-schedule {
  padding: 8px 24px;
  font-size: 13px;
  font-weight: 500;
}

@media (max-width: 640px) {
  .time-picker-box {
    width: 100% !important;
  }
}
</style>

<template>
  <div class="schedule-tab" v-loading="loading">
    <!-- Group 1: 分片合并调度 -->
    <div class="setting-group">
      <div class="category-divider">
        <span class="category-label">分片合并调度</span>
        <span class="category-line" />
      </div>

      <div class="setting-row">
        <div class="setting-info">
          <div class="setting-title">开启自动合并</div>
          <div class="setting-desc">周期性自动扫描各主播目录并合并切片</div>
        </div>
        <div class="setting-control">
          <el-switch
            :model-value="form.merge_enabled"
            @update:model-value="emit('update:form', { ...form, merge_enabled: !!$event })"
          />
        </div>
      </div>

      <div class="setting-row">
        <div class="setting-info">
          <div class="setting-title">合并调度周期</div>
          <div class="setting-desc">任务执行完毕后等待指定间隔再启动下次合并（推荐 360 分钟 / 6 小时）</div>
        </div>
        <div class="setting-control">
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
        </div>
      </div>
    </div>

    <!-- Group 2: 磁盘清理巡检 -->
    <div class="setting-group">
      <div class="category-divider">
        <span class="category-label">磁盘清理巡检</span>
        <span class="category-line" />
      </div>

      <div class="setting-row">
        <div class="setting-info">
          <div class="setting-title">开启自动清理</div>
          <div class="setting-desc">定期巡检磁盘使用率，达警戒水位时自动触发清理</div>
        </div>
        <div class="setting-control">
          <el-switch
            :model-value="form.clean_enabled"
            @update:model-value="emit('update:form', { ...form, clean_enabled: !!$event })"
          />
        </div>
      </div>

      <div class="setting-row">
        <div class="setting-info">
          <div class="setting-title">巡检调度周期</div>
          <div class="setting-desc">巡检器检查磁盘可用水位的轮询间隔（推荐 720 分钟 / 12 小时）</div>
        </div>
        <div class="setting-control">
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
        </div>
      </div>
    </div>

    <!-- Group 3: 夜间避让时段 -->
    <div class="setting-group">
      <div class="category-divider">
        <span class="category-label">夜间避让时段</span>
        <span class="category-line" />
      </div>

      <div class="setting-row">
        <div class="setting-info">
          <div class="setting-title">避让时间范围</div>
          <div class="setting-desc">在此设定时段内自动暂停合并与清理，避开直播高峰或网络拥堵</div>
        </div>
        <div class="setting-control">
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
        </div>
      </div>
    </div>
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
  display: flex;
  flex-direction: column;
}

.setting-group {
  display: flex;
  flex-direction: column;
}
.setting-group + .setting-group {
  margin-top: 24px;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 13px 0;
  border-bottom: 1px solid var(--hairline);
  gap: 20px;
}
.setting-row:last-child {
  border-bottom: none;
}

.setting-info {
  display: flex;
  flex-direction: column;
  gap: 3px;
  flex: 1;
  min-width: 0;
}

.setting-title {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.1px;
}

.setting-desc {
  font-size: 12px;
  color: var(--steel);
  line-height: 1.45;
}

.setting-control {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  flex-shrink: 0;
}

.compact-counter {
  width: 120px;
  height: 32px !important;
  line-height: 32px !important;
  :deep(.el-input__wrapper) {
    height: 32px !important;
    box-sizing: border-box !important;
  }
  :deep(.el-input-number__decrease),
  :deep(.el-input-number__increase) {
    height: 30px !important;
    line-height: 30px !important;
  }
}

.inline-unit-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.unit-text {
  font-size: 12px;
  color: var(--steel);
  font-weight: 500;
  white-space: nowrap;
}

.time-range-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.time-picker-box {
  width: 120px !important;
}

.time-sep {
  font-size: 12px;
  color: var(--steel);
}
</style>

<script setup>
defineProps({
  schedule: {
    type: Object,
    required: true
  },
  config: {
    type: Object,
    required: true
  }
})

function clampInput(event, max) {
  const value = parseFloat(event.target.value)
  if (!isNaN(value) && value > max) {
    event.target.value = max
    event.target.dispatchEvent(new Event('input'))
  }
}
</script>

<template>
  <div class="setting-grid">
    <div class="setting-box" :class="{'is-active': schedule.merge_enabled, 'is-disabled': !schedule.merge_enabled}">
      <label style="display:flex;align-items:center;flex:1;cursor:pointer;margin:0">
        <div class="switch" style="margin-right:16px">
          <input type="checkbox" v-model="schedule.merge_enabled">
          <span class="switch-slider"></span>
        </div>
        <div class="setting-info">
          <div class="setting-title">自动合并任务</div>
          <div class="setting-desc">后台定期执行录像碎片拼接与转码</div>
        </div>
      </label>
      <div class="setting-control">
        <input type="number" v-model.number="schedule.merge_interval" :disabled="!schedule.merge_enabled" min="10" max="1440" @input="clampInput($event, 1440)">
        <span class="setting-unit">MIN</span>
      </div>
    </div>

    <div class="setting-box" :class="{'is-active': schedule.clean_enabled, 'is-disabled': !schedule.clean_enabled}">
      <label style="display:flex;align-items:center;flex:1;cursor:pointer;margin:0">
        <div class="switch" style="margin-right:16px">
          <input type="checkbox" v-model="schedule.clean_enabled">
          <span class="switch-slider"></span>
        </div>
        <div class="setting-info">
          <div class="setting-title">自动清理任务</div>
          <div class="setting-desc">触及空间警戒阈值时自动释放容量</div>
        </div>
      </label>
      <div class="setting-control">
        <input type="number" v-model.number="schedule.clean_interval" :disabled="!schedule.clean_enabled" min="10" max="1440" @input="clampInput($event, 1440)">
        <span class="setting-unit">MIN</span>
      </div>
    </div>

    <div class="setting-box span-full" style="cursor:default; flex-wrap:wrap; gap:24px; justify-content:space-between;">
      <div class="setting-info" style="margin:0;flex:1;min-width:280px">
        <div class="setting-title">系统静默时段</div>
        <div class="setting-desc">此时段内自动任务强制休眠，防止高峰期 IO 读写拥堵</div>
      </div>
      <div class="setting-control" style="gap:12px;background:var(--card);padding:6px 16px;border-radius:8px;border:1px solid var(--border);box-shadow:0 1px 2px rgba(0,0,0,.02)">
        <div class="time-picker">
          <input type="number" v-model.number="config.BACKUP_START_HOUR" min="0" max="23" placeholder="04">
          <span>:</span>
          <input type="number" v-model.number="config.BACKUP_START_MINUTE" min="0" max="59" placeholder="00">
        </div>
        <span style="font-size:12px;font-weight:600;color:var(--muted);padding:0 4px">至</span>
        <div class="time-picker">
          <input type="number" v-model.number="config.BACKUP_END_HOUR" min="0" max="23" placeholder="12">
          <span>:</span>
          <input type="number" v-model.number="config.BACKUP_END_MINUTE" min="0" max="59" placeholder="00">
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.setting-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.setting-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  background: var(--bg-sub);
  border: 1px solid var(--border);
  border-radius: 8px;
  transition: all 0.25s ease;
  cursor: pointer;
}

.setting-box:hover:not(.is-disabled) {
  border-color: var(--muted);
}

.setting-box:focus-within {
  border-color: var(--pri);
  background: var(--card);
  box-shadow: 0 0 0 3px var(--pri-r);
}

.setting-box.is-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.setting-box.is-active {
  border-color: var(--pri);
  box-shadow: 0 0 0 1px var(--pri);
}

.setting-info {
  flex: 1;
  margin-right: 16px;
}

.setting-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
}

.setting-desc {
  font-size: 12px;
  color: var(--muted);
  margin-top: 4px;
  line-height: 1.5;
}

.setting-control {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.setting-unit {
  font-size: 12px;
  font-weight: 600;
  color: var(--muted);
  width: 28px;
  text-align: left;
  display: inline-block;
}

.setting-control input[type="number"] {
  width: 60px;
  padding: 0 8px 0 0;
  text-align: right;
  font-family: var(--font-mono);
  font-size: 13px;
  font-weight: 500;
}

.span-full {
  grid-column: 1 / -1;
}

.time-picker {
  display: inline-flex;
  align-items: center;
  background: var(--bg-sub);
  border: 1px solid var(--border);
  border-radius: 6px;
  transition: all 0.2s;
  overflow: hidden;
}

.time-picker:focus-within {
  border-color: var(--pri);
  box-shadow: 0 0 0 3px var(--pri-r);
  background: var(--card);
}

.time-picker input {
  width: 40px;
  border: none;
  background: transparent;
  text-align: center;
  box-shadow: none !important;
  font-size: 13px;
  border-radius: 0;
  -moz-appearance: textfield;
}

.time-picker input::-webkit-outer-spin-button,
.time-picker input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
</style>

<template>
  <div v-if="show" class="disk-alert-banner" :class="level >= 2 ? 'critical' : 'warning'" role="alert" @click="$emit('navigate')">
    <span class="disk-alert-text">
      <svg class="alert-svg-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/>
        <line x1="12" y1="9" x2="12" y2="13"/>
        <line x1="12" y1="17" x2="12.01" y2="17"/>
      </svg>
      <template v-if="level >= 2">磁盘空间严重不足（{{ usedPct }}%），请立即清理</template>
      <template v-else>磁盘使用率达到 {{ usedPct }}%，建议及时清理</template>
    </span>
    <button class="disk-alert-close" @click.stop="$emit('dismiss')" title="关闭">
      <el-icon><Close /></el-icon>
    </button>
  </div>
</template>

<script setup lang="ts">
import { Close } from "@element-plus/icons-vue";

defineProps<{
  show: boolean;
  level: number;
  usedPct: number;
}>();

defineEmits<{
  dismiss: [];
  navigate: [];
}>();
</script>

<style scoped>
.disk-alert-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: background 0.2s;
  border-bottom: 1px solid transparent;
}
.disk-alert-banner.warning {
  background: #fff7ed;
  color: #c2410c;
  border-bottom-color: #ffedd5;
}
.disk-alert-banner.critical {
  background: #fef2f2;
  color: #dc2626;
  border-bottom-color: #fee2e2;
}
.disk-alert-text {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.alert-svg-ico {
  width: 15px;
  height: 15px;
  flex-shrink: 0;
}
.disk-alert-close {
  background: none;
  border: none;
  cursor: pointer;
  color: inherit;
  opacity: 0.6;
  padding: 4px;
  display: flex;
  align-items: center;
  border-radius: 4px;
  transition: opacity 0.15s, background 0.15s;
}
.disk-alert-close:hover {
  opacity: 1;
  background: rgba(0, 0, 0, 0.05);
}

html.dark .disk-alert-banner.warning {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
  border-bottom-color: rgba(245, 158, 11, 0.25);
}
html.dark .disk-alert-banner.critical {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
  border-bottom-color: rgba(239, 68, 68, 0.3);
}
html.dark .disk-alert-close:hover {
  background: rgba(255, 255, 255, 0.08);
}
</style>

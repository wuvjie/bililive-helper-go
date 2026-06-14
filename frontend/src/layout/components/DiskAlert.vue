<template>
  <div v-if="show" class="disk-alert-banner" :class="level >= 2 ? 'critical' : 'warning'" role="alert" @click="$emit('navigate')">
    <span class="disk-alert-text">
      <template v-if="level >= 2">🚨 磁盘空间严重不足（{{ usedPct }}%），请立即清理</template>
      <template v-else>⚠️ 磁盘使用率 {{ usedPct }}%，建议及时清理</template>
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
  padding: 6px 16px;
  cursor: pointer;
  font-size: 13px;
  transition: background 0.2s;
}
.disk-alert-banner.warning {
  background: #fff7e6;
  color: #d9730d;
}
.disk-alert-banner.critical {
  background: #ffebee;
  color: #e03131;
}
.disk-alert-close {
  background: none;
  border: none;
  cursor: pointer;
  color: inherit;
  opacity: 0.6;
  padding: 2px;
}
.disk-alert-close:hover {
  opacity: 1;
}
</style>

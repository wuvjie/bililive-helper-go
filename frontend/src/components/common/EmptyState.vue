<template>
  <div class="empty-state-container empty-illustration">
    <div class="empty-icon-wrapper" :style="iconBgStyle">
      <slot name="icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10" />
          <line x1="8" y1="12" x2="16" y2="12" />
        </svg>
      </slot>
    </div>
    <span class="empty-title">
      <slot name="title">{{ title }}</slot>
    </span>
    <span v-if="desc || $slots.desc" class="empty-desc">
      <slot name="desc">{{ desc }}</slot>
    </span>
    <div v-if="actionText || $slots.action" class="empty-action-wrap">
      <slot name="action">
        <button v-if="actionText" class="btn-empty-action" @click="$emit('action')">
          {{ actionText }}
        </button>
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    title?: string;
    desc?: string;
    actionText?: string;
    iconColor?: string;
  }>(),
  {
    title: "暂无数据",
    desc: "",
    actionText: "",
    iconColor: ""
  }
);

defineEmits<{
  (e: "action"): void;
}>();

const iconBgStyle = computed(() => {
  if (!props.iconColor) return undefined;
  return {
    color: props.iconColor,
    background: `color-mix(in srgb, ${props.iconColor} 12%, transparent)`
  };
});
</script>

<style scoped>
.empty-state-container {
  padding: 32px 16px;
}

.empty-action-wrap {
  margin-top: 4px;
}
</style>

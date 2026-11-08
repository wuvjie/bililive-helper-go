<template>
  <span class="bds-status-pill" :class="[`status-${type}`, { 'has-pulse': shouldPulse }]">
    <span class="status-dot" :class="{ 'pulse-dot': shouldPulse }"></span>
    <span class="status-label"><slot>{{ text }}</slot></span>
  </span>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    type?: "success" | "fail" | "running" | "idle" | "warn" | "raw" | "neutral";
    text?: string;
    pulse?: boolean;
  }>(),
  {
    type: "neutral",
    text: "",
    pulse: false
  }
);

const shouldPulse = computed(() => {
  return props.pulse || props.type === "running";
});
</script>

<style scoped>
.bds-status-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 8px;
  border-radius: var(--r-full);
  font-size: 12px;
  font-weight: 500;
  line-height: 1.4;
  white-space: nowrap;
  border: 1px solid transparent;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
  flex-shrink: 0;
}

/* Success / Normal */
.status-success {
  background: color-mix(in srgb, #059669 10%, transparent);
  color: #047857;
  border-color: color-mix(in srgb, #059669 20%, transparent);
}
.status-success .status-dot { background: #10b981; }

/* Fail / Error */
.status-fail {
  background: color-mix(in srgb, var(--error) 10%, transparent);
  color: var(--error);
  border-color: color-mix(in srgb, var(--error) 25%, transparent);
}
.status-fail .status-dot { background: var(--error); }

/* Running / Active */
.status-running {
  background: color-mix(in srgb, #d97706 10%, transparent);
  color: #b45309;
  border-color: color-mix(in srgb, #d97706 25%, transparent);
}
.status-running .status-dot { background: #f59e0b; }

/* Idle / Standing */
.status-idle {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.04));
  color: var(--slate);
  border-color: var(--hairline);
}
.status-idle .status-dot { background: var(--stone); }

/* Warn */
.status-warn {
  background: color-mix(in srgb, #f59e0b 12%, transparent);
  color: #d97706;
  border-color: color-mix(in srgb, #f59e0b 25%, transparent);
}
.status-warn .status-dot { background: #f59e0b; }

/* Raw / Subdued */
.status-raw {
  background: rgba(148, 163, 184, 0.12);
  color: #64748b;
  border-color: rgba(148, 163, 184, 0.2);
}
.status-raw .status-dot { background: #94a3b8; }

/* Neutral */
.status-neutral {
  background: var(--surface-secondary, rgba(0, 0, 0, 0.04));
  color: var(--charcoal);
  border-color: var(--hairline);
}
.status-neutral .status-dot { background: var(--slate); }

.pulse-dot {
  box-shadow: 0 0 0 0 rgba(245, 158, 11, 0.7);
  animation: statusPulse 1.8s infinite;
}

@keyframes statusPulse {
  0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(245, 158, 11, 0.7); }
  70% { transform: scale(1); box-shadow: 0 0 0 5px rgba(245, 158, 11, 0); }
  100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(245, 158, 11, 0); }
}

html.dark {
  .status-success {
    background: rgba(16, 185, 129, 0.12);
    color: #34d399;
    border-color: rgba(16, 185, 129, 0.25);
  }
  .status-fail {
    background: rgba(239, 68, 68, 0.12);
    color: #f87171;
    border-color: rgba(239, 68, 68, 0.25);
  }
  .status-running {
    background: rgba(245, 158, 11, 0.12);
    color: #fbbf24;
    border-color: rgba(245, 158, 11, 0.25);
  }
  .status-idle {
    background: rgba(255, 255, 255, 0.05);
    color: #94a3b8;
    border-color: rgba(255, 255, 255, 0.1);
  }
  .status-warn {
    background: rgba(245, 158, 11, 0.15);
    color: #fcd34d;
    border-color: rgba(245, 158, 11, 0.3);
  }
}
</style>

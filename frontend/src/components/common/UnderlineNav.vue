<template>
  <nav class="underline-nav" role="tablist" :aria-label="ariaLabel || '分类导航'">
    <button
      v-for="item in items"
      :key="item.value"
      type="button"
      role="tab"
      :aria-selected="modelValue === item.value"
      class="underline-nav-item"
      :class="{ 'is-active': modelValue === item.value }"
      @click="handleSelect(item.value)"
    >
      <slot name="icon" :item="item">
        <component v-if="item.icon" :is="item.icon" class="nav-item-svg" />
      </slot>
      <span class="nav-item-text">{{ item.label }}</span>
      <span v-if="item.count !== undefined" class="nav-item-badge">
        {{ item.count }}
      </span>
    </button>
  </nav>
</template>

<script setup lang="ts">
import type { Component } from "vue";

export interface UnderlineNavItem {
  label: string;
  value: string | number;
  icon?: Component;
  count?: number;
}

const props = defineProps<{
  modelValue: string | number;
  items: UnderlineNavItem[];
  ariaLabel?: string;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", val: string | number): void;
  (e: "change", val: string | number): void;
}>();

function handleSelect(val: string | number) {
  if (props.modelValue !== val) {
    emit("update:modelValue", val);
    emit("change", val);
  }
}
</script>

<style scoped>
.underline-nav {
  display: inline-flex;
  align-items: center;
  gap: 16px;
  height: 36px;
  user-select: none;
  background: transparent;
  border: none;
  padding: 0;
  margin: 0;
}

.underline-nav-item {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 4px;
  background: transparent;
  border: none;
  font-family: var(--font-sans);
  font-size: 13px;
  font-weight: 500;
  color: var(--steel);
  cursor: pointer;
  white-space: nowrap;
  outline: none;
  transition: color 0.15s ease;

  &:hover {
    color: var(--charcoal);
  }

  &.is-active {
    color: var(--ink);
    font-weight: 600;

    &::after {
      content: "";
      position: absolute;
      bottom: -1px;
      left: 0;
      right: 0;
      height: 2px;
      background: var(--primary);
      border-radius: 1px;
    }

    .nav-item-svg {
      color: var(--primary);
    }
  }

  &:focus-visible {
    border-radius: var(--r-xs);
    outline: 2px solid var(--hairline-focus);
    outline-offset: 2px;
  }
}

.nav-item-svg {
  width: 14px;
  height: 14px;
  color: inherit;
  flex-shrink: 0;
  transition: color 0.15s ease;
}

.nav-item-text {
  line-height: 1;
}

.nav-item-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 500;
  line-height: 1;
  color: var(--steel);
  background: rgba(0, 0, 0, 0.05);
  border-radius: 10px;
  transition: background 0.15s ease, color 0.15s ease;
}

.underline-nav-item.is-active .nav-item-badge {
  color: var(--charcoal);
  background: rgba(0, 0, 0, 0.08);
}

/* Dark Mode Tweaks (GitHub Primer) */
html.dark {
  .underline-nav-item {
    color: var(--steel);

    &:hover {
      color: var(--ink);
    }

    &.is-active {
      color: var(--ink);

      &::after {
        background: var(--primary);
      }
    }
  }

  .nav-item-badge {
    background: rgba(255, 255, 255, 0.08);
    color: var(--steel);
  }

  .underline-nav-item.is-active .nav-item-badge {
    color: var(--ink);
    background: rgba(255, 255, 255, 0.14);
  }
}
</style>

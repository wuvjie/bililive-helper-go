<template>
  <div class="filter-pills" role="tablist">
    <button
      v-for="item in items"
      :key="item.value"
      type="button"
      role="tab"
      :aria-selected="modelValue === item.value"
      class="filter-pill-btn"
      :class="{ 'is-active': modelValue === item.value }"
      @click="selectItem(item.value)"
    >
      <slot name="item-icon" :item="item" />
      <span class="pill-label">{{ item.label }}</span>
      <span v-if="item.count !== undefined" class="pill-badge-count">
        {{ item.count }}
      </span>
    </button>
  </div>
</template>

<script setup lang="ts">
export interface FilterPillItem {
  label: string;
  value: string | number;
  count?: number;
}

const props = defineProps<{
  modelValue: string | number;
  items: FilterPillItem[];
}>();

const emit = defineEmits<{
  (e: "update:modelValue", val: string | number): void;
  (e: "change", val: string | number): void;
}>();

function selectItem(val: string | number) {
  if (props.modelValue !== val) {
    emit("update:modelValue", val);
    emit("change", val);
  }
}
</script>

<style scoped>
.filter-pills {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  background: var(--surface-secondary, rgba(0, 0, 0, 0.04));
  padding: 3px;
  border-radius: var(--r-md);
  border: 1px solid var(--hairline);
  user-select: none;
}

html.dark .filter-pills {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.08);
}

.filter-pill-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 500;
  color: var(--slate);
  background: transparent;
  border: none;
  border-radius: var(--r-sm);
  cursor: pointer;
  transition: all 0.15s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
  line-height: 1.5;
}

.filter-pill-btn:hover:not(.is-active) {
  color: var(--ink);
  background: rgba(0, 0, 0, 0.03);
}

html.dark .filter-pill-btn:hover:not(.is-active) {
  background: rgba(255, 255, 255, 0.04);
  color: #f4f4f5;
}

.filter-pill-btn.is-active {
  background: var(--surface-card, #ffffff);
  color: var(--ink);
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08), 0 1px 2px rgba(0, 0, 0, 0.04);
}

html.dark .filter-pill-btn.is-active {
  background: #27272a;
  color: #ffffff;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.pill-badge-count {
  font-family: var(--font-mono);
  font-size: 11px;
  padding: 0 5px;
  border-radius: var(--r-full);
  background: rgba(0, 0, 0, 0.06);
  color: var(--stone);
}

.is-active .pill-badge-count {
  background: rgba(0, 0, 0, 0.09);
  color: var(--ink);
}

html.dark .pill-badge-count {
  background: rgba(255, 255, 255, 0.08);
  color: #a1a1aa;
}

html.dark .is-active .pill-badge-count {
  background: rgba(255, 255, 255, 0.15);
  color: #ffffff;
}
</style>

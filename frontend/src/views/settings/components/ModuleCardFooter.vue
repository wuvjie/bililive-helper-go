<template>
  <div class="module-card-footer">
    <div class="footer-left">
      <div class="module-step-indicator" role="tablist">
        <button
          v-for="(item, idx) in sections"
          :key="item.id"
          type="button"
          class="step-dot-btn"
          :class="{ active: item.id === activeId, passed: idx < currentIndex }"
          :title="`切换至：${item.label}`"
          @click="emit('switch', item.id)"
        />
      </div>
      <span class="module-step-text">
        第 {{ currentIndex + 1 }} / {{ sections.length }} 模块 · {{ currentLabel }}
      </span>
    </div>

    <div class="footer-right">
      <button
        v-if="prevSection"
        type="button"
        class="btn-module-nav btn-module-prev"
        @click="emit('prev')"
        :title="`返回上一模块「${prevSection.label}」`"
      >
        <svg class="nav-arrow-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="15 18 9 12 15 6"/>
        </svg>
        <span>上一模块：{{ prevSection.label }}</span>
      </button>

      <button
        v-if="nextSection"
        type="button"
        class="btn-module-nav btn-module-next"
        @click="emit('next')"
        :title="`前往下一模块「${nextSection.label}」`"
      >
        <span>下一模块：{{ nextSection.label }}</span>
        <svg class="nav-arrow-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="9 18 15 12 9 6"/>
        </svg>
      </button>

      <button
        v-else
        type="button"
        class="btn-module-nav btn-module-loop"
        @click="emit('loop')"
        title="已至末页，点击返回首项「全局策略」"
      >
        <svg class="nav-arrow-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/>
          <path d="M3 3v5h5"/>
        </svg>
        <span>已至末页 · 返回首项</span>
      </button>

      <div
        class="wheel-hint-pill"
        :class="{ 'pulling-active': isPullingNext || isPullingPrev }"
      >
        <div
          v-if="isPullingNext || isPullingPrev"
          class="pull-progress-fill"
          :style="{ width: pullProgress + '%' }"
        />
        <svg class="mouse-icon-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="6" y="3" width="12" height="18" rx="6" />
          <line x1="12" y1="7" x2="12" y2="11" />
        </svg>
        <span v-if="isPullingNext">继续下滑前往「{{ nextSection?.label }}」</span>
        <span v-else-if="isPullingPrev">继续上滑返回「{{ prevSection?.label }}」</span>
        <span v-else-if="!nextSection">已至末页（触顶可上滑回退）</span>
        <span v-else>触底滚轮快速切换</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
export interface NavSectionItem {
  id: string;
  label: string;
}

defineProps<{
  sections: NavSectionItem[];
  activeId: string;
  currentIndex: number;
  currentLabel: string;
  prevSection?: NavSectionItem | null;
  nextSection?: NavSectionItem | null;
  isPullingNext?: boolean;
  isPullingPrev?: boolean;
  pullProgress?: number;
}>();

const emit = defineEmits<{
  switch: [id: string];
  next: [];
  prev: [];
  loop: [];
}>();
</script>

<style scoped>
.module-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 13px 22px;
  background: var(--surface-secondary, rgba(0, 0, 0, 0.02));
  border-top: 1px solid var(--hairline, #e4e4e7);
  gap: 16px;
  flex-wrap: wrap;
}

.footer-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.module-step-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
}

.step-dot-btn {
  width: 8px;
  height: 8px;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: var(--hairline-strong, #d4d4d8);
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.step-dot-btn:hover {
  background: var(--slate, #71717a);
  transform: scale(1.2);
}

.step-dot-btn.passed {
  background: color-mix(in srgb, var(--primary, #2563eb) 40%, var(--hairline-strong, #d4d4d8));
}

.step-dot-btn.active {
  width: 20px;
  border-radius: 4px;
  background: var(--primary, #2563eb);
}

.module-step-text {
  font-size: 12px;
  font-weight: 500;
  color: var(--steel, #71717a);
  font-variant-numeric: tabular-nums;
}

.footer-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.btn-module-nav {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 30px;
  padding: 0 12px;
  border-radius: var(--r-sm, 6px);
  border: 1px solid var(--hairline, #e4e4e7);
  background: var(--surface-card, #ffffff);
  color: var(--charcoal, #27272a);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-module-nav:hover {
  border-color: var(--border-hover, #a1a1aa);
  background: var(--surface-hover, rgba(0, 0, 0, 0.03));
  color: var(--ink, #09090b);
}

.btn-module-prev {
  background: var(--surface-card, #ffffff);
  border-color: var(--hairline, #e4e4e7);
  color: var(--charcoal, #27272a);
}

.btn-module-next {
  background: var(--primary, #2563eb);
  border-color: var(--primary, #2563eb);
  color: #ffffff;
}

.btn-module-next:hover {
  background: #1d4ed8;
  border-color: #1d4ed8;
  color: #ffffff;
}

.btn-module-loop {
  background: var(--surface-card, #ffffff);
  border-color: var(--hairline, #e4e4e7);
  color: var(--steel, #71717a);
}

.btn-module-loop:hover {
  color: var(--primary, #2563eb);
  border-color: var(--primary, #2563eb);
}

.nav-arrow-svg {
  width: 13px;
  height: 13px;
}

.wheel-hint-pill {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11.5px;
  color: var(--muted, #a1a1aa);
  background: var(--surface-card, #ffffff);
  border: 1px dashed var(--hairline, #e4e4e7);
  padding: 3px 10px;
  border-radius: var(--r-xs, 4px);
  user-select: none;
  overflow: hidden;
  transition: border-color 0.15s, color 0.15s;
}

.wheel-hint-pill.pulling-active {
  border-color: var(--primary, #2563eb);
  color: var(--primary, #2563eb);
  font-weight: 600;
}

.pull-progress-fill {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  background: color-mix(in srgb, var(--primary, #2563eb) 12%, transparent);
  transition: width 0.05s linear;
  pointer-events: none;
}

.mouse-icon-svg {
  width: 12px;
  height: 12px;
  color: currentColor;
}

html.dark .module-card-footer {
  background: rgba(255, 255, 255, 0.02);
  border-top-color: rgba(255, 255, 255, 0.08);
}
html.dark .step-dot-btn {
  background: rgba(255, 255, 255, 0.2);
}
html.dark .step-dot-btn:hover {
  background: rgba(255, 255, 255, 0.4);
}
html.dark .step-dot-btn.passed {
  background: color-mix(in srgb, #60a5fa 40%, rgba(255, 255, 255, 0.2));
}
html.dark .step-dot-btn.active {
  background: #60a5fa;
}
html.dark .btn-module-nav {
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(255, 255, 255, 0.12);
  color: #e4e4e7;
}
html.dark .btn-module-nav:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.22);
  color: #ffffff;
}
html.dark .btn-module-next {
  background: #2563eb;
  border-color: #3b82f6;
  color: #ffffff;
}
html.dark .btn-module-next:hover {
  background: #1d4ed8;
}
html.dark .btn-module-loop {
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(255, 255, 255, 0.12);
  color: #a1a1aa;
}
html.dark .btn-module-loop:hover {
  color: #60a5fa;
  border-color: #60a5fa;
}
html.dark .wheel-hint-pill {
  background: rgba(255, 255, 255, 0.03);
  border-color: rgba(255, 255, 255, 0.12);
  color: #71717a;
}
html.dark .wheel-hint-pill.pulling-active {
  border-color: #60a5fa;
  color: #60a5fa;
}
html.dark .pull-progress-fill {
  background: rgba(96, 165, 250, 0.18);
}
</style>

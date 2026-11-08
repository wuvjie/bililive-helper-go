<template>
  <div class="bds-app-card" :class="[customClass, { 'bds-app-card-flush': flush }]">
    <!-- Header -->
    <div v-if="$slots.header || title || $slots.extra || $slots['header-left']" class="bds-card-header" :class="headerClass">
      <div class="bds-header-left">
        <slot name="header-left">
          <div v-if="$slots['header-icon'] || iconColor" class="ui-icon-box" :class="iconColor ? `icon-box-${iconColor}` : ''">
            <slot name="header-icon" />
          </div>
          <div class="bds-header-titles">
            <slot name="title">
              <span v-if="title" class="bds-card-title">{{ title }}</span>
            </slot>
            <slot name="subtitle">
              <span v-if="subtitle" class="bds-card-subtitle">{{ subtitle }}</span>
            </slot>
          </div>
        </slot>
      </div>

      <div v-if="$slots.extra || $slots.action" class="bds-header-right">
        <slot name="extra">
          <slot name="action" />
        </slot>
      </div>
    </div>

    <!-- Body -->
    <div class="bds-card-body" :class="[bodyClass, { 'body-flush': flush }]">
      <slot />
    </div>

    <!-- Footer -->
    <div v-if="$slots.footer" class="bds-card-footer" :class="footerClass">
      <slot name="footer" />
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    title?: string;
    subtitle?: string;
    iconColor?: string;
    flush?: boolean;
    customClass?: string;
    headerClass?: string;
    bodyClass?: string;
    footerClass?: string;
  }>(),
  {
    flush: false,
    iconColor: ""
  }
);
</script>

<style scoped>
.bds-app-card {
  background: var(--surface-card, #ffffff);
  border-radius: var(--r-md);
  border: 1px solid var(--hairline);
  box-shadow: var(--shadow-card);
  transition: box-shadow 0.2s ease, border-color 0.2s ease;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.bds-app-card:hover {
  box-shadow: var(--shadow-card-hover);
}

.bds-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  gap: 12px;
  border-bottom: 1px solid var(--hairline);
}

.bds-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1;
}

.bds-header-titles {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  flex-wrap: wrap;
}

.bds-card-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--charcoal);
  letter-spacing: -0.2px;
  line-height: 1.3;
}

.bds-card-subtitle {
  font-size: 12px;
  color: var(--steel);
}

.bds-header-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.bds-card-body {
  padding: 20px;
  flex: 1;
  min-width: 0;
}

.bds-card-body.body-flush {
  padding: 0;
}

.bds-card-footer {
  padding: 12px 20px;
  border-top: 1px solid var(--hairline);
  background: var(--surface-secondary, rgba(0, 0, 0, 0.02));
}

html.dark .bds-card-footer {
  background: rgba(255, 255, 255, 0.02);
}
</style>

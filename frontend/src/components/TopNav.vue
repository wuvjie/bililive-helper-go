<script setup>
defineProps({
  running: {
    type: Boolean,
    default: false
  }
})
</script>

<template>
  <header class="topbar">
    <div class="topbar-brand">
      <div class="topbar-logo">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/>
        </svg>
      </div>
      <h1>Bililive Helper</h1>
    </div>
    <div class="topbar-right">
      <span v-if="running" class="badge badge-warn">● 运行中</span>
      <a href="/logout">退出登录</a>
    </div>
  </header>
</template>

<style scoped>
.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  margin-bottom: 24px;
  position: sticky;
  top: 0;
  z-index: 40;
  background: color-mix(in srgb, var(--bg) 80%, transparent);
  backdrop-filter: blur(12px);
}

.topbar-brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.topbar-logo {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--pri);
  display: flex;
  align-items: center;
  justify-content: center;
}

.topbar-logo svg {
  width: 18px;
  height: 18px;
  color: var(--pri-fg);
}

.topbar h1 {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: -0.5px;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.topbar-right a {
  color: var(--muted);
  text-decoration: none;
  font-size: 13px;
  font-weight: 500;
  transition: color 0.2s;
}

.topbar-right a:hover {
  color: var(--text);
}
</style>

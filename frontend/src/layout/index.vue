<template>
  <div class="layout-container" :class="{ 'sidebar-collapsed': !appStore.sidebar.opened }">
    <aside class="layout-sidebar" :class="{ collapsed: !appStore.sidebar.opened }">
      <!-- Sidebar Header: Mintlify-grade left-aligned skeleton -->
      <div class="sidebar-header">
        <template v-if="appStore.sidebar.opened">
          <div class="sidebar-header-left">
            <!-- Physical pulsing status dot (Mintlify signature mint-green) -->
            <span class="status-dot-wrapper">
              <span class="status-dot-ping" />
              <span class="status-dot-core" />
            </span>
            <!-- Title: Inter sans, semi-bold, tight tracking -->
            <h1 class="sidebar-title">Bililive Helper</h1>
          </div>
          <!-- Version micro-badge: Geist Mono, uppercase, hairline border -->
          <span class="sidebar-version">v1.0</span>
        </template>
        <template v-else>
          <span class="sidebar-title-mini">BH</span>
        </template>
      </div>
      <el-scrollbar>
        <el-menu :default-active="activeMenu" :collapse="!appStore.sidebar.opened" :collapse-transition="false" router>
          <el-menu-item v-for="item in menuItems" :key="item.path" :index="item.path">
            <el-icon><component :is="item.icon" /></el-icon>
            <template #title>{{ item.title }}</template>
          </el-menu-item>
        </el-menu>
      </el-scrollbar>
    </aside>

    <div class="layout-main">
      <header class="layout-navbar">
        <div class="navbar-left">
          <el-icon class="collapse-btn" @click="appStore.toggleSidebar"><component :is="appStore.sidebar.opened ? 'Fold' : 'Expand'" /></el-icon>
          <el-breadcrumb separator="/"><el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item><el-breadcrumb-item v-if="currentRoute.meta.title">{{ currentRoute.meta.title }}</el-breadcrumb-item></el-breadcrumb>
        </div>
        <div class="navbar-right">
          <el-tooltip content="刷新" placement="bottom"><el-icon class="nav-icon" @click="refreshPage"><Refresh /></el-icon></el-tooltip>
          <el-tooltip content="全屏" placement="bottom"><el-icon class="nav-icon" @click="toggleFullscreen"><FullScreen /></el-icon></el-tooltip>
          <el-dropdown trigger="click" @command="handleCommand">
            <span class="user-info"><el-avatar :size="28" class="user-avatar">A</el-avatar><span class="user-name">Admin</span></span>
            <template #dropdown><el-dropdown-menu><el-dropdown-item command="logout">退出登录</el-dropdown-item></el-dropdown-menu></template>
          </el-dropdown>
        </div>
      </header>
      <main class="layout-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <keep-alive><component :is="Component" :key="currentRoute.fullPath" /></keep-alive>
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import { useAppStore } from "@/store/modules/app";
import { logout } from "@/api/auth";
import { Monitor, User, VideoPlay, Document, Setting, Fold, Expand, Refresh, FullScreen } from "@element-plus/icons-vue";

const route = useRoute();
const appStore = useAppStore();
const menuItems = [
  { path: "/dashboard", title: "系统概览", icon: Monitor },
  { path: "/streamers", title: "主播管理", icon: User },
  { path: "/tasks", title: "任务中心", icon: VideoPlay },
  { path: "/history", title: "操作日志", icon: Document },
  { path: "/settings", title: "全局设置", icon: Setting }
];
const activeMenu = computed(() => route.path);
const currentRoute = computed(() => route);
function refreshPage() { window.location.reload(); }
function toggleFullscreen() { document.fullscreenElement ? document.exitFullscreen() : document.documentElement.requestFullscreen(); }
function handleCommand(cmd: string) { if (cmd === "logout") logout(); }
</script>

<style lang="scss" scoped>
.layout-container { display: flex; height: 100vh; overflow: hidden; }

.layout-sidebar {
  width: 240px; background: var(--canvas);
  border-right: 1px solid var(--hairline-soft);
  transition: width 0.15s; display: flex; flex-direction: column; flex-shrink: 0; overflow: hidden;
  &.collapsed { width: 56px; }
}

.sidebar-logo {
  height: 52px; display: flex; align-items: center; justify-content: center;
  border-bottom: 1px solid var(--hairline-soft); flex-shrink: 0;
  .logo-text { font-size: 15px; font-weight: 600; color: var(--ink); letter-spacing: -0.3px; }
  .logo-text-mini { font-size: 16px; font-weight: 700; color: var(--ink); }
}

/* Mintlify-grade sidebar header */
.sidebar-header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  border-bottom: 1px solid var(--hairline-soft);
  flex-shrink: 0;
}

.sidebar-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* Physical pulsing status dot */
.status-dot-wrapper {
  position: relative;
  display: flex;
  height: 8px;
  width: 8px;
  flex-shrink: 0;
}

.status-dot-ping {
  position: absolute;
  display: inline-flex;
  height: 100%;
  width: 100%;
  border-radius: 9999px;
  background-color: var(--brand-green);
  opacity: 0.75;
  animation: ping 1.5s cubic-bezier(0, 0, 0.2, 1) infinite;
}

.status-dot-core {
  position: relative;
  display: inline-flex;
  border-radius: 9999px;
  height: 8px;
  width: 8px;
  background-color: var(--brand-green);
}

@keyframes ping {
  75%, 100% {
    transform: scale(2.5);
    opacity: 0;
  }
}

/* Title: Inter sans, semi-bold, tight tracking */
.sidebar-title {
  font-family: var(--font-sans);
  font-size: 15px;
  font-weight: 600;
  letter-spacing: -0.3px;
  color: var(--ink);
  line-height: 1;
  white-space: nowrap;
}

/* Version micro-badge: mono font, uppercase, hairline */
.sidebar-version {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.5px;
  color: var(--steel);
  background: var(--surface);
  padding: 2px 6px;
  border-radius: var(--r-xs);
  border: 1px solid var(--hairline);
  text-transform: uppercase;
  line-height: 1;
  flex-shrink: 0;
}

/* Collapsed state mini title */
.sidebar-title-mini {
  font-size: 16px;
  font-weight: 700;
  color: var(--ink);
}

:deep(.el-menu) {
  border-right: none; padding: 4px 8px;
  .el-menu-item {
    border-radius: var(--r-sm); margin-bottom: 1px;
    font-size: 14px; font-weight: 400; color: var(--steel);
    transition: background 0.1s, color 0.1s;
    &:hover { background: var(--surface); color: var(--ink); }
    &.is-active { background: var(--surface); color: var(--ink); font-weight: 500; }
  }
}

.layout-main { flex: 1; display: flex; flex-direction: column; overflow: hidden; background: var(--canvas); }

.layout-navbar {
  height: 52px; background: var(--canvas);
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 20px; border-bottom: 1px solid var(--hairline-soft);
  flex-shrink: 0; z-index: 10;
  .navbar-left, .navbar-right { display: flex; align-items: center; gap: 10px; }
}

.collapse-btn { font-size: 18px; cursor: pointer; color: var(--muted); &:hover { color: var(--ink); } }
.nav-icon { font-size: 17px; cursor: pointer; color: var(--muted); &:hover { color: var(--ink); } }

.user-info {
  display: flex; align-items: center; gap: 8px; cursor: pointer;
  .user-avatar { background: var(--ink); color: var(--on-primary); font-size: 12px; font-weight: 600; }
  .user-name { font-size: 14px; color: var(--ink); font-weight: 500; }
}

.layout-content { flex: 1; overflow-y: auto; padding: 24px; }

.fade-enter-active, .fade-leave-active { transition: opacity 0.12s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

@media (max-width: 768px) {
  .layout-sidebar { position: fixed; z-index: 100; height: 100vh; box-shadow: 1px 0 4px rgba(0,0,0,0.04); }
  .sidebar-collapsed .layout-sidebar { display: none; }
  .layout-content { padding: 16px; }
}
</style>

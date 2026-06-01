<template>
  <div class="layout-container" :class="{ 'sidebar-collapsed': !appStore.sidebar.opened }">
    <aside class="layout-sidebar" :class="{ collapsed: !appStore.sidebar.opened }">
      <div class="sidebar-logo">
        <span v-if="appStore.sidebar.opened" class="logo-text">Bililive Helper</span>
        <span v-else class="logo-text-mini">BH</span>
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
          <el-icon class="collapse-btn" @click="appStore.toggleSidebar">
            <component :is="appStore.sidebar.opened ? 'Fold' : 'Expand'" />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="currentRoute.meta.title">{{ currentRoute.meta.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="navbar-right">
          <el-tooltip content="刷新" placement="bottom">
            <el-icon class="nav-icon" @click="refreshPage"><Refresh /></el-icon>
          </el-tooltip>
          <el-tooltip content="全屏" placement="bottom">
            <el-icon class="nav-icon" @click="toggleFullscreen"><FullScreen /></el-icon>
          </el-tooltip>
          <el-dropdown trigger="click" @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="28" class="user-avatar">A</el-avatar>
              <span class="user-name">Admin</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu><el-dropdown-item command="logout">退出登录</el-dropdown-item></el-dropdown-menu>
            </template>
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
  border-right: 1px solid var(--canvas-soft);
  transition: width 0.15s; display: flex; flex-direction: column; flex-shrink: 0; overflow: hidden;
  &.collapsed { width: 56px; }
}

.sidebar-logo {
  height: 56px; display: flex; align-items: center; justify-content: center;
  border-bottom: 1px solid var(--canvas-soft); flex-shrink: 0;
  .logo-text { font-size: 16px; font-weight: 900; color: var(--ink); letter-spacing: -0.5px; }
  .logo-text-mini { font-size: 17px; font-weight: 900; color: var(--ink); }
}

:deep(.el-menu) {
  border-right: none; padding: 8px 8px;
  .el-menu-item {
    border-radius: var(--r-xl); margin-bottom: 2px;
    font-size: 16px; font-weight: 600; color: var(--mute);
    transition: background 0.1s, color 0.1s;
    &:hover { background: var(--canvas-soft); color: var(--ink); }
    &.is-active { background: var(--primary-pale); color: var(--ink); }
  }
}

.layout-main { flex: 1; display: flex; flex-direction: column; overflow: hidden; background: var(--canvas-soft); }

.layout-navbar {
  height: 56px; background: var(--canvas);
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 24px; border-bottom: 1px solid var(--canvas-soft);
  flex-shrink: 0; z-index: 10;
  .navbar-left, .navbar-right { display: flex; align-items: center; gap: 12px; }
}

.collapse-btn { font-size: 20px; cursor: pointer; color: var(--mute); &:hover { color: var(--ink); } }
.nav-icon { font-size: 18px; cursor: pointer; color: var(--mute); &:hover { color: var(--ink); } }

.user-info {
  display: flex; align-items: center; gap: 8px; cursor: pointer;
  .user-avatar { background: var(--ink); color: var(--primary); font-size: 13px; font-weight: 800; }
  .user-name { font-size: 16px; color: var(--ink); font-weight: 600; }
}

.layout-content { flex: 1; overflow-y: auto; padding: 24px; }

.fade-enter-active, .fade-leave-active { transition: opacity 0.12s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

@media (max-width: 768px) {
  .layout-sidebar { position: fixed; z-index: 100; height: 100vh; box-shadow: 2px 0 12px rgba(0,0,0,0.06); }
  .sidebar-collapsed .layout-sidebar { display: none; }
  .layout-content { padding: 16px; }
}
</style>

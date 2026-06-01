<template>
  <div class="layout-container" :class="{ 'sidebar-collapsed': !appStore.sidebar.opened }">
    <aside class="layout-sidebar" :class="{ collapsed: !appStore.sidebar.opened }">
      <div class="sidebar-logo">
        <span v-if="appStore.sidebar.opened" class="logo-text">Bililive Helper</span>
        <span v-else class="logo-text-mini">BH</span>
      </div>
      <el-scrollbar>
        <el-menu
          :default-active="activeMenu"
          :collapse="!appStore.sidebar.opened"
          :collapse-transition="false"
          router
        >
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
          <el-tooltip content="Fullscreen" placement="bottom">
            <el-icon class="nav-icon" @click="toggleFullscreen"><FullScreen /></el-icon>
          </el-tooltip>
          <el-dropdown trigger="click" @command="handleCommand">
            <span class="user-info">
              <div class="user-dot" />
              <span class="user-name">Admin</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <main class="layout-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <keep-alive>
              <component :is="Component" :key="currentRoute.fullPath" />
            </keep-alive>
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRoute } from "vue-router";
import { useAppStore } from "@/store/modules/app";
import { logout } from "@/api/auth";
import {
  Monitor, User, VideoPlay, Document, Setting,
  Fold, Expand, Refresh, FullScreen
} from "@element-plus/icons-vue";

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
function toggleFullscreen() {
  if (!document.fullscreenElement) document.documentElement.requestFullscreen();
  else document.exitFullscreen();
}
function handleCommand(cmd: string) { if (cmd === "logout") logout(); }
</script>

<style lang="scss" scoped>
.layout-container { display: flex; height: 100vh; overflow: hidden; }

.layout-sidebar {
  width: 220px;
  background: var(--surface-card);
  border-right: 1px solid var(--hairline);
  transition: width 0.15s ease;
  display: flex; flex-direction: column; flex-shrink: 0; overflow: hidden;
  &.collapsed { width: 52px; }
}

.sidebar-logo {
  height: 48px;
  display: flex; align-items: center; justify-content: center;
  border-bottom: 1px solid var(--hairline); flex-shrink: 0;
  .logo-text {
    font-family: var(--font-display);
    font-size: 15px; font-weight: 400; color: var(--ink);
    font-style: italic; letter-spacing: -0.3px;
  }
  .logo-text-mini { font-size: 14px; font-weight: 600; color: var(--ink); }
}

:deep(.el-menu) {
  border-right: none; padding: 4px 6px;
  .el-menu-item {
    border-radius: var(--r-md); margin-bottom: 1px;
    font-size: 14px; color: var(--mute);
    transition: background 0.1s, color 0.1s;
    &:hover { background: rgba(255, 255, 255, 0.04); color: var(--charcoal); }
    &.is-active { background: rgba(255, 255, 255, 0.06); color: var(--ink); }
  }
}

.layout-main { flex: 1; display: flex; flex-direction: column; overflow: hidden; background: var(--canvas); }

.layout-navbar {
  height: 48px;
  background: var(--canvas);
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 16px;
  border-bottom: 1px solid var(--hairline);
  flex-shrink: 0; z-index: 10;
  .navbar-left, .navbar-right { display: flex; align-items: center; gap: 10px; }
}

.collapse-btn { font-size: 16px; cursor: pointer; color: var(--mute); &:hover { color: var(--ink); } }
.nav-icon { font-size: 16px; cursor: pointer; color: var(--mute); &:hover { color: var(--ink); } }

.user-info {
  display: flex; align-items: center; gap: 8px; cursor: pointer;
  .user-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--accent-green); }
  .user-name { font-size: 13px; color: var(--charcoal); }
}

.layout-content { flex: 1; overflow-y: auto; padding: 24px; }

.fade-enter-active, .fade-leave-active { transition: opacity 0.1s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

@media (max-width: 768px) {
  .layout-sidebar { position: fixed; z-index: 100; height: 100vh; }
  .sidebar-collapsed .layout-sidebar { display: none; }
  .layout-content { padding: 12px; }
}
</style>

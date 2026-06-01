<template>
  <div class="layout-container" :class="{ 'sidebar-collapsed': !appStore.sidebar.opened }">
    <!-- Sidebar -->
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

    <!-- Main area -->
    <div class="layout-main">
      <!-- Navbar -->
      <header class="layout-navbar">
        <div class="navbar-left">
          <el-icon class="collapse-btn" @click="appStore.toggleSidebar">
            <component :is="appStore.sidebar.opened ? 'Fold' : 'Expand'" />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="currentRoute.meta.title">
              {{ currentRoute.meta.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="navbar-right">
          <el-tooltip content="刷新页面" placement="bottom">
            <el-icon class="nav-icon" @click="refreshPage"><Refresh /></el-icon>
          </el-tooltip>
          <el-tooltip :content="isDark ? '浅色模式' : '深色模式'" placement="bottom">
            <el-icon class="nav-icon" @click="toggleDark">
              <component :is="isDark ? 'Sunny' : 'Moon'" />
            </el-icon>
          </el-tooltip>
          <el-tooltip content="全屏" placement="bottom">
            <el-icon class="nav-icon" @click="toggleFullscreen">
              <FullScreen />
            </el-icon>
          </el-tooltip>
          <el-dropdown trigger="click" @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="28" class="user-avatar">A</el-avatar>
              <span class="username">Admin</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">
                  <el-icon><SwitchButton /></el-icon>退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <!-- Content -->
      <main class="layout-content">
        <router-view v-slot="{ Component }">
          <transition name="fade-transform" mode="out-in">
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
import { computed, ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { useAppStore } from "@/store/modules/app";
import { logout } from "@/api/auth";
import {
  Monitor, User, VideoPlay, Document, Setting,
  Fold, Expand, ArrowDown, SwitchButton,
  Refresh, Moon, Sunny, FullScreen
} from "@element-plus/icons-vue";

const route = useRoute();
const appStore = useAppStore();
const isDark = ref(false);

const menuItems = [
  { path: "/dashboard", title: "系统概览", icon: Monitor },
  { path: "/streamers", title: "主播管理", icon: User },
  { path: "/tasks", title: "任务中心", icon: VideoPlay },
  { path: "/history", title: "操作日志", icon: Document },
  { path: "/settings", title: "全局设置", icon: Setting }
];

const activeMenu = computed(() => route.path);
const currentRoute = computed(() => route);

function toggleDark() {
  isDark.value = !isDark.value;
  document.documentElement.classList.toggle("dark", isDark.value);
  localStorage.setItem("theme", isDark.value ? "dark" : "light");
}

function refreshPage() {
  window.location.reload();
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen();
  } else {
    document.exitFullscreen();
  }
}

function handleCommand(cmd: string) {
  if (cmd === "logout") logout();
}

onMounted(() => {
  const saved = localStorage.getItem("theme");
  if (saved === "dark" || (!saved && window.matchMedia("(prefers-color-scheme: dark)").matches)) {
    isDark.value = true;
    document.documentElement.classList.add("dark");
  }
});
</script>

<style lang="scss" scoped>
.layout-container {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.layout-sidebar {
  width: 220px;
  background: var(--bg-sidebar);
  transition: width 0.28s;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  overflow: hidden;

  &.collapsed { width: 64px; }
}

.sidebar-logo {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid rgba(255,255,255,0.06);
  flex-shrink: 0;

  .logo-text {
    color: #fff;
    font-size: 16px;
    font-weight: 600;
    white-space: nowrap;
    letter-spacing: 1px;
  }
  .logo-text-mini { font-size: 22px; }
}

:deep(.el-menu) {
  border-right: none;
  background: transparent !important;
  --el-menu-bg-color: transparent;
  --el-menu-hover-bg-color: var(--bg-sidebar-hover);
  --el-menu-text-color: #bfcbd9;
  --el-menu-active-color: #409eff;
  --el-menu-item-height: 48px;
}

.layout-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--bg-page);
  transition: background 0.3s;
}

.layout-navbar {
  height: 52px;
  background: var(--bg-card);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  box-shadow: var(--shadow-sm);
  flex-shrink: 0;
  z-index: 10;
  transition: background 0.3s;

  .navbar-left, .navbar-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }
}

.collapse-btn {
  font-size: 20px;
  cursor: pointer;
  color: var(--text-regular);
  transition: color 0.2s;
  &:hover { color: var(--primary); }
}

.nav-icon {
  font-size: 18px;
  cursor: pointer;
  color: var(--text-regular);
  transition: color 0.2s, transform 0.2s;
  &:hover { color: var(--primary); transform: scale(1.1); }
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  .user-avatar { background: var(--primary); color: #fff; font-size: 13px; }
  .username { font-size: 14px; color: var(--text-primary); }
}

.layout-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.fade-transform-enter-active, .fade-transform-leave-active {
  transition: all 0.2s;
}
.fade-transform-enter-from { opacity: 0; transform: translateX(-10px); }
.fade-transform-leave-to { opacity: 0; transform: translateX(10px); }

@media (max-width: 768px) {
  .layout-sidebar {
    position: fixed;
    z-index: 100;
    height: 100vh;
    transform: translateX(0);
    transition: transform 0.3s;
  }
  .sidebar-collapsed .layout-sidebar {
    transform: translateX(-100%);
  }
}
</style>

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
          <el-tooltip content="刷新" placement="bottom">
            <el-icon class="nav-icon" @click="refreshPage"><Refresh /></el-icon>
          </el-tooltip>
          <el-tooltip :content="isDark ? 'Light' : 'Dark'" placement="bottom">
            <el-icon class="nav-icon" @click="toggleDark">
              <component :is="isDark ? 'Sunny' : 'Moon'" />
            </el-icon>
          </el-tooltip>
          <el-tooltip content="Fullscreen" placement="bottom">
            <el-icon class="nav-icon" @click="toggleFullscreen"><FullScreen /></el-icon>
          </el-tooltip>
          <el-dropdown trigger="click" @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="26" class="user-avatar">A</el-avatar>
              <el-icon class="arrow"><ArrowDown /></el-icon>
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
const isDark = ref(true); // Linear is dark-first

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
  // Linear is always dark — light mode uses inverse surfaces
  if (!isDark.value) {
    document.documentElement.style.setProperty("--canvas", "#ffffff");
    document.documentElement.style.setProperty("--surface-1", "#f5f6f6");
    document.documentElement.style.setProperty("--surface-2", "#f0f1f1");
    document.documentElement.style.setProperty("--ink", "#000000");
    document.documentElement.style.setProperty("--ink-muted", "#4a4e54");
    document.documentElement.style.setProperty("--ink-subtle", "#6b7075");
    document.documentElement.style.setProperty("--hairline", "#e0e0e0");
  } else {
    document.documentElement.style.removeProperty("--canvas");
    document.documentElement.style.removeProperty("--surface-1");
    document.documentElement.style.removeProperty("--surface-2");
    document.documentElement.style.removeProperty("--ink");
    document.documentElement.style.removeProperty("--ink-muted");
    document.documentElement.style.removeProperty("--ink-subtle");
    document.documentElement.style.removeProperty("--hairline");
  }
}

function refreshPage() { window.location.reload(); }
function toggleFullscreen() {
  if (!document.fullscreenElement) document.documentElement.requestFullscreen();
  else document.exitFullscreen();
}
function handleCommand(cmd: string) { if (cmd === "logout") logout(); }

onMounted(() => {
  const saved = localStorage.getItem("theme");
  if (saved === "light") {
    isDark.value = false;
    toggleDark(); // apply light vars
    isDark.value = false; // reset since toggleDark flipped it
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
  background: var(--surface-1);
  border-right: 1px solid var(--hairline);
  transition: width 0.2s ease;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  overflow: hidden;

  &.collapsed { width: 52px; }
}

.sidebar-logo {
  height: 52px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid var(--hairline);
  flex-shrink: 0;

  .logo-text {
    color: var(--ink);
    font-size: 14px;
    font-weight: 600;
    letter-spacing: -0.3px;
    font-family: var(--font-display);
  }
  .logo-text-mini {
    color: var(--ink);
    font-size: 15px;
    font-weight: 700;
    font-family: var(--font-display);
  }
}

:deep(.el-menu) {
  border-right: none;
  padding: 4px 6px;

  .el-menu-item {
    border-radius: var(--r-md);
    margin-bottom: 2px;
    font-size: 14px;
    color: var(--ink-subtle);
    transition: background 0.15s, color 0.15s;

    &:hover {
      background: var(--surface-2);
      color: var(--ink-muted);
    }

    &.is-active {
      background: var(--surface-2);
      color: var(--ink);
    }
  }
}

.layout-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--canvas);
}

.layout-navbar {
  height: 52px;
  background: var(--canvas);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  border-bottom: 1px solid var(--hairline);
  flex-shrink: 0;
  z-index: 10;

  .navbar-left, .navbar-right {
    display: flex;
    align-items: center;
    gap: 10px;
  }
}

.collapse-btn {
  font-size: 18px;
  cursor: pointer;
  color: var(--ink-subtle);
  transition: color 0.15s;
  &:hover { color: var(--ink); }
}

.nav-icon {
  font-size: 17px;
  cursor: pointer;
  color: var(--ink-subtle);
  transition: color 0.15s;
  &:hover { color: var(--ink); }
}

.user-info {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  .user-avatar {
    background: var(--primary);
    color: var(--primary-text);
    font-size: 12px;
    font-weight: 600;
  }
  .arrow { font-size: 12px; color: var(--ink-subtle); }
}

.layout-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}

@media (max-width: 768px) {
  .layout-sidebar {
    position: fixed;
    z-index: 100;
    height: 100vh;
    transform: translateX(0);
    transition: transform 0.25s ease;
  }
  .sidebar-collapsed .layout-sidebar {
    transform: translateX(-100%);
  }
  .layout-content { padding: 12px; }
}
</style>

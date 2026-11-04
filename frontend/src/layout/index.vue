<template>
  <div class="layout-container" :class="{ 'sidebar-collapsed': !appStore.sidebar.opened }">
    <!-- 侧边栏 -->
    <aside
      class="layout-sidebar"
      :class="{
        collapsed: !isMobile && !appStore.sidebar.opened,
        'mobile-open': isMobile && mobileOpen
      }"
    >
      <div class="sidebar-logo">
        <span v-if="isMobile || appStore.sidebar.opened" class="logo-text">Bililive Helper Go</span>
        <span v-else class="logo-text-mini">BHG</span>
      </div>

      <el-scrollbar class="sidebar-scrollbar">
        <el-menu
          :default-active="activeMenu"
          :collapse="!isMobile && !appStore.sidebar.opened"
          :collapse-transition="false"
          router
          @select="handleMenuSelect"
        >
          <el-menu-item v-for="item in menuItems" :key="item.path" :index="item.path">
            <el-icon><component :is="item.icon" /></el-icon>
            <template #title>{{ item.title }}</template>
          </el-menu-item>
        </el-menu>
      </el-scrollbar>

      <!-- 侧边栏底部折叠切换小开关 (三条横线图标) -->
      <div
        class="sidebar-bottom-toggle"
        @click="handleToggleCollapse"
        :title="(!isMobile && !appStore.sidebar.opened) ? '展开侧边栏' : '收起侧边栏'"
      >
        <div class="toggle-icon-box">
          <svg class="hamburger-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="4" y1="6" x2="20" y2="6"></line>
            <line x1="4" y1="12" x2="20" y2="12"></line>
            <line x1="4" y1="18" x2="20" y2="18"></line>
          </svg>
        </div>
        <span v-if="isMobile || appStore.sidebar.opened" class="toggle-text">收起侧栏</span>
      </div>
    </aside>

    <!-- 移动端暗色遮罩层 -->
    <div
      class="sidebar-overlay"
      v-if="isMobile && mobileOpen"
      @click="mobileOpen = false"
    ></div>

    <div class="layout-main">
      <header class="layout-navbar">
        <div class="navbar-left">
          <!-- 导航汉堡按钮：电脑端与移动端均可切换，移动端显眼直观 -->
          <button
            class="nav-hamburger-btn"
            @click="handleToggleCollapse"
            title="菜单导航"
            :class="{ active: isMobile ? mobileOpen : appStore.sidebar.opened }"
          >
            <svg class="hamburger-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="4" y1="6" x2="20" y2="6"></line>
              <line x1="4" y1="12" x2="20" y2="12"></line>
              <line x1="4" y1="18" x2="20" y2="18"></line>
            </svg>
          </button>

          <el-breadcrumb separator="/" class="nav-breadcrumb">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="currentRoute.meta.title">{{ currentRoute.meta.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="navbar-right">
          <el-tooltip content="刷新数据" placement="bottom">
            <button class="nav-btn-icon" :class="{ spinning: isRefreshing }" @click="refreshPage">
              <svg class="nav-svg-ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            </button>
          </el-tooltip>
          <span class="nav-divider"></span>
          <el-dropdown trigger="click" @command="handleCommand" placement="bottom-end">
            <div class="user-pill-btn" title="当前登录身份：管理员">
              <span class="user-pill-dot"></span>
              <span class="user-pill-name">管理员</span>
              <svg class="user-pill-arrow" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
              </svg>
            </div>
            <template #dropdown>
              <el-dropdown-menu class="nav-dropdown-menu">
                <div class="user-dropdown-header">
                  <div class="user-dropdown-title">系统管理员</div>
                  <div class="user-dropdown-sub">版本 v2.0.0 (Release)</div>
                </div>
                <el-dropdown-item command="change-password" divided>
                  <span>🔑 修改管理密码</span>
                </el-dropdown-item>
                <el-dropdown-item command="logout" divided>
                  <span style="color: #d4433a;">🚪 退出系统</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <!-- Disk usage alert banner -->
      <DiskAlert :show="showDiskBanner" :level="diskAlertLevel" :used-pct="diskUsedPct" @dismiss="dismissDiskBanner" @navigate="goToTasks" />

      <main class="layout-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <keep-alive>
              <component :is="Component" />
            </keep-alive>
          </transition>
        </router-view>
      </main>
    </div>

    <!-- Password change dialog -->
    <PasswordDialog v-model:visible="pwDialogVisible" v-model:saving="pwSaving" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAppStore } from "@/store/modules/app";
import { logout } from "@/api/auth";
import { getStatusDetail } from "@/api/status";
import { Monitor, User, VideoPlay, Document, Setting, Refresh, FullScreen } from "@element-plus/icons-vue";
import PasswordDialog from "./components/PasswordDialog.vue";
import DiskAlert from "./components/DiskAlert.vue";

const route = useRoute();
const router = useRouter();
const appStore = useAppStore();

const menuItems = [
  { path: "/dashboard", title: "系统概览", icon: Monitor },
  { path: "/streamers", title: "主播管理", icon: User },
  { path: "/tasks", title: "任务中心", icon: VideoPlay },
  { path: "/history", title: "操作日志", icon: Document },
  { path: "/settings", title: "系统设置", icon: Setting }
];

const activeMenu = computed(() => route.path);
const currentRoute = computed(() => route);

const isMobile = ref(typeof window !== "undefined" ? window.innerWidth < 768 : false);
const mobileOpen = ref(false);

function onResize() {
  const mobile = window.innerWidth < 768;
  isMobile.value = mobile;
  if (!mobile) {
    mobileOpen.value = false;
  }
}

onMounted(() => window.addEventListener("resize", onResize));
onUnmounted(() => window.removeEventListener("resize", onResize));

function handleToggleCollapse() {
  if (isMobile.value) {
    mobileOpen.value = !mobileOpen.value;
  } else {
    appStore.toggleSidebar();
  }
}

function handleMenuSelect() {
  if (isMobile.value) {
    mobileOpen.value = false;
  }
}

// Password change dialog
const pwDialogVisible = ref(false);
const pwSaving = ref(false);

const isRefreshing = ref(false);
function refreshPage() {
  isRefreshing.value = true;
  setTimeout(() => {
    window.location.reload();
  }, 250);
}
function handleCommand(cmd: string) {
  if (cmd === "logout") logout();
  if (cmd === "change-password") pwDialogVisible.value = true;
}

// Disk usage alert
const diskUsedPct = ref(0);
const dismissedAlertLevel = ref(0);
const diskAlertLevel = computed(() => {
  if (diskUsedPct.value >= 95) return 2;
  if (diskUsedPct.value >= 90) return 1;
  return 0;
});
const showDiskBanner = computed(() => {
  const level = diskAlertLevel.value;
  if (level === 0) return false;
  if (level === 1) return dismissedAlertLevel.value < 1;
  return dismissedAlertLevel.value < 2;
});
function dismissDiskBanner() { dismissedAlertLevel.value = diskAlertLevel.value; }
function goToTasks() { router.push("/tasks"); }
let lastDiskFetch = 0;
async function fetchDiskStatus() {
  const now = Date.now();
  if (now - lastDiskFetch < 30_000) return; // 30 秒节流
  lastDiskFetch = now;
  try {
    const data = await getStatusDetail();
    diskUsedPct.value = data.disk.usage_pct;
    if (diskAlertLevel.value === 0) dismissedAlertLevel.value = 0;
  } catch { /* silently ignore */ }
}
onMounted(() => fetchDiskStatus());
router.afterEach(() => fetchDiskStatus());
</script>

<style lang="scss" scoped>
.layout-container { display: flex; height: 100vh; overflow: hidden; }

.layout-sidebar {
  width: 240px;
  background: var(--surface);
  border-right: 1px solid var(--hairline);
  transition: width 0.2s cubic-bezier(0.4, 0, 0.2, 1), transform 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  display: flex; flex-direction: column; flex-shrink: 0; overflow: hidden;
  z-index: 20;

  &.collapsed {
    width: 64px;
    .sidebar-logo {
      justify-content: center;
      padding: 0;
    }
    .sidebar-bottom-toggle {
      justify-content: center;
      padding: 0;
      .toggle-icon-box {
        margin: 0 auto;
      }
    }
  }
}

.sidebar-logo {
  height: 52px;
  display: flex; align-items: center; padding: 0 16px;
  border-bottom: 1px solid var(--hairline); flex-shrink: 0;
  .logo-text { font-size: 15px; font-weight: 600; color: var(--ink); letter-spacing: -0.3px; white-space: nowrap; }
  .logo-text-mini { font-size: 16px; font-weight: 700; color: var(--ink); }
}

.sidebar-scrollbar {
  flex: 1;
}

:deep(.el-menu) {
  border-right: none;
  padding: 6px 12px;
  transition: padding 0.2s cubic-bezier(0.4, 0, 0.2, 1);

  &.el-menu--collapse {
    width: 100% !important;
    padding: 6px 0 !important;

    .el-menu-item {
      padding: 0 !important;
      margin: 0 auto 4px auto !important;
      width: 44px !important;
      height: 44px !important;
      display: flex !important;
      align-items: center !important;
      justify-content: center !important;

      .el-icon {
        margin: 0 !important;
        font-size: 18px !important;
        display: flex !important;
        align-items: center !important;
        justify-content: center !important;
      }
    }
  }

  .el-menu-item {
    border-radius: var(--r-sm); margin-bottom: 2px;
    font-size: 14px; color: var(--steel);
    padding: 0 14px;
    transition: background 0.1s, color 0.1s;
    &:hover { background: var(--highlight); color: var(--charcoal); }
    &.is-active { background: var(--highlight); color: var(--ink); font-weight: 500; }
  }
}

/* 侧边栏底部折叠切换条 */
.sidebar-bottom-toggle {
  height: 48px;
  display: flex;
  align-items: center;
  padding: 0 16px;
  gap: 12px;
  border-top: 1px solid var(--hairline);
  background: var(--surface);
  cursor: pointer;
  color: var(--steel);
  transition: background 0.15s, color 0.15s;
  user-select: none;
  flex-shrink: 0;

  &:hover {
    background: var(--highlight);
    color: var(--ink);
  }

  .toggle-icon-box {
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: var(--r-xs);
  }

  .hamburger-svg {
    width: 18px;
    height: 18px;
  }

  .toggle-text {
    font-size: 13px;
    font-weight: 500;
    white-space: nowrap;
  }
}

.layout-main { flex: 1; display: flex; flex-direction: column; overflow: hidden; background: var(--canvas); }

.layout-navbar {
  height: 52px; background: var(--canvas);
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 20px;
  border-bottom: 1px solid var(--hairline);
  flex-shrink: 0; z-index: 10;
  .navbar-left, .navbar-right { display: flex; align-items: center; gap: 8px; }
}

/* 顶部汉堡按钮 */
.nav-hamburger-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--r-sm);
  border: 1px solid var(--hairline);
  background: var(--surface);
  color: var(--charcoal);
  cursor: pointer;
  transition: all 0.15s ease;
  margin-right: 6px;

  &:hover {
    background: var(--highlight);
    color: var(--ink);
    border-color: var(--faint);
  }

  &.active {
    background: var(--highlight);
    color: var(--primary);
  }

  .hamburger-svg {
    width: 17px;
    height: 17px;
  }
}

.nav-breadcrumb {
  display: flex;
  align-items: center;
}

.nav-icon { font-size: 17px; cursor: pointer; color: var(--steel); &:hover { color: var(--ink); } }
.nav-divider { width: 1px; height: 14px; background: var(--hairline); flex-shrink: 0; }

.nav-btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: var(--r-sm);
  background: transparent;
  border: 1px solid transparent;
  color: var(--steel);
  cursor: pointer;
  transition: all 0.15s ease;

  &:hover {
    background: var(--highlight);
    color: var(--ink);
    border-color: var(--hairline);
  }

  &.spinning .nav-svg-ico {
    animation: rotate-spin 0.6s linear infinite;
  }
}
@keyframes rotate-spin {
  100% { transform: rotate(360deg); }
}
.nav-svg-ico {
  width: 15px;
  height: 15px;
}

.user-pill-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 8px 3px 10px;
  background: var(--surface);
  border: 1px solid var(--hairline);
  border-radius: 14px;
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;

  &:hover {
    background: var(--highlight);
    border-color: var(--faint);
  }
}
.user-pill-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #448361;
}
.user-pill-name {
  font-size: 12px;
  font-weight: 500;
  color: var(--charcoal);
}
.user-pill-arrow {
  width: 12px;
  height: 12px;
  color: var(--steel);
  transition: transform 0.15s ease;
}
.user-dropdown-header {
  padding: 8px 14px 6px;
  min-width: 150px;
}
.user-dropdown-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--charcoal);
  line-height: 1.4;
}
.user-dropdown-sub {
  font-size: 11px;
  color: var(--muted);
  font-family: var(--font-mono);
  margin-top: 2px;
}

.layout-content { flex: 1; overflow-y: auto; padding: 24px; padding-bottom: 48px; }

/* 移动端遮罩层 */
.sidebar-overlay {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 99;
  backdrop-filter: blur(2px);
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.fade-enter-active, .fade-leave-active { transition: opacity 0.15s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

@media (max-width: 768px) {
  .layout-sidebar {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    width: 250px !important;
    height: 100vh;
    transform: translateX(-100%);
    box-shadow: none;
    z-index: 200 !important;

    &.mobile-open {
      transform: translateX(0);
      box-shadow: 4px 0 24px rgba(0, 0, 0, 0.4);
    }
  }

  .sidebar-overlay {
    z-index: 150 !important;
  }

  .sidebar-collapsed .layout-sidebar {
    display: flex;
  }

  .layout-content {
    padding: 14px;
    padding-bottom: 36px;
  }

  .nav-breadcrumb {
    font-size: 12px;
  }
}
</style>

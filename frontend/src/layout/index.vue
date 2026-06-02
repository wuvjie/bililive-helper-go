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
          <el-tooltip content="全屏" placement="bottom">
            <el-icon class="nav-icon" @click="toggleFullscreen"><FullScreen /></el-icon>
          </el-tooltip>
          <span class="nav-divider"></span>
          <el-dropdown trigger="click" @command="handleCommand" placement="bottom-end">
            <span class="user-info">
              <span class="user-avatar">A</span>
              <span class="user-name">Admin</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="change-password">修改密码</el-dropdown-item>
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
              <component :is="Component" />
            </keep-alive>
          </transition>
        </router-view>
      </main>
    </div>

    <!-- Password change dialog -->
    <el-dialog v-model="pwDialogVisible" title="修改密码" width="400px" destroy-on-close>
      <el-form label-position="top" @submit.prevent="handleChangePassword">
        <el-form-item label="旧密码">
          <el-input v-model="pwForm.old_password" type="password" show-password placeholder="输入当前密码" />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="pwForm.new_password" type="password" show-password placeholder="至少 6 个字符" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="pw-footer">
          <el-button @click="pwDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="pwSaving" @click="handleChangePassword">确认修改</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, reactive } from "vue";
import { useRoute } from "vue-router";
import { ElMessage } from "element-plus";
import { useAppStore } from "@/store/modules/app";
import { logout, changePassword } from "@/api/auth";
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

// Password change dialog
const pwDialogVisible = ref(false);
const pwSaving = ref(false);
const pwForm = reactive({ old_password: "", new_password: "" });

function refreshPage() { window.location.reload(); }
function toggleFullscreen() {
  if (!document.fullscreenElement) document.documentElement.requestFullscreen();
  else document.exitFullscreen();
}
function handleCommand(cmd: string) {
  if (cmd === "logout") logout();
  if (cmd === "change-password") { pwDialogVisible.value = true; }
}
async function handleChangePassword() {
  if (!pwForm.old_password || !pwForm.new_password) { ElMessage.warning("请填写旧密码和新密码"); return; }
  if (pwForm.new_password.length < 6) { ElMessage.warning("新密码至少 6 个字符"); return; }
  pwSaving.value = true;
  try {
    await changePassword(pwForm.old_password, pwForm.new_password);
    ElMessage.success("密码已更新");
    pwDialogVisible.value = false;
    pwForm.old_password = ""; pwForm.new_password = "";
  } catch { /* handled by interceptor */ }
  finally { pwSaving.value = false; }
}
</script>

<style lang="scss" scoped>
.layout-container { display: flex; height: 100vh; overflow: hidden; }

.layout-sidebar {
  width: 240px;
  background: var(--surface);
  border-right: 1px solid var(--hairline);
  transition: width 0.15s ease;
  display: flex; flex-direction: column; flex-shrink: 0; overflow: hidden;
  &.collapsed { width: 52px; }
}

.sidebar-logo {
  height: 52px;
  display: flex; align-items: center; padding: 0 16px;
  border-bottom: 1px solid var(--hairline); flex-shrink: 0;
  .logo-text { font-size: 15px; font-weight: 600; color: var(--ink); letter-spacing: -0.3px; }
  .logo-text-mini { font-size: 15px; font-weight: 700; color: var(--ink); }
}

:deep(.el-menu) {
  border-right: none; padding: 6px 16px;
  .el-menu-item {
    border-radius: var(--r-sm); margin-bottom: 1px;
    font-size: 14px; color: var(--steel);
    padding: 0 12px;
    transition: background 0.1s, color 0.1s;
    &:hover { background: var(--highlight); color: var(--charcoal); }
    &.is-active { background: var(--highlight); color: var(--ink); font-weight: 500; }
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

.collapse-btn { font-size: 18px; cursor: pointer; color: var(--steel); &:hover { color: var(--ink); } }
.nav-icon { font-size: 17px; cursor: pointer; color: var(--steel); &:hover { color: var(--ink); } }
.nav-divider { width: 1px; height: 14px; background: var(--hairline); flex-shrink: 0; }

.user-info {
  display: flex; align-items: center; gap: 8px; cursor: pointer;
  .user-avatar {
    width: 20px; height: 20px;
    background: var(--primary); color: var(--on-primary);
    font-size: 11px; font-weight: 700; font-family: var(--font-mono);
    display: flex; align-items: center; justify-content: center;
    border-radius: var(--r-xs); flex-shrink: 0;
  }
  .user-name { font-size: 13px; color: var(--steel); font-weight: 500; transition: color 0.15s; }
  &:hover .user-name { color: var(--ink); }
}

.layout-content { flex: 1; overflow-y: auto; padding: 24px; padding-bottom: 48px; }

/* Password dialog footer — right-aligned buttons */
.pw-footer { display: flex; justify-content: flex-end; gap: 8px; }

.fade-enter-active, .fade-leave-active { transition: opacity 0.15s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

@media (max-width: 768px) {
  .layout-sidebar { position: fixed; z-index: 100; height: 100vh; box-shadow: 2px 0 8px rgba(0,0,0,0.08); }
  .sidebar-collapsed .layout-sidebar { display: none; }
  .layout-content { padding: 12px; padding-bottom: 36px; }
}
</style>

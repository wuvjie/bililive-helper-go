import { createRouter, createWebHashHistory, type RouteRecordRaw } from "vue-router";
import Layout from "@/layout/index.vue";

const routes: RouteRecordRaw[] = [
  {
    path: "/login",
    name: "Login",
    component: () => import("@/views/login/index.vue"),
    meta: { title: "登录", public: true }
  },
  {
    path: "/setup",
    name: "Setup",
    component: () => import("@/views/setup/index.vue"),
    meta: { title: "首次配置", public: true }
  },
  {
    path: "/",
    component: Layout,
    redirect: "/dashboard",
    children: [
      { path: "dashboard", name: "Dashboard", component: () => import("@/views/dashboard/index.vue"), meta: { title: "系统概览", icon: "Monitor" } },
      { path: "streamers", name: "Streamers", component: () => import("@/views/streamers/index.vue"), meta: { title: "主播管理", icon: "User" } },
      { path: "tasks", name: "Tasks", component: () => import("@/views/tasks/index.vue"), meta: { title: "任务中心", icon: "VideoPlay" } },
      { path: "history", name: "History", component: () => import("@/views/history/index.vue"), meta: { title: "操作日志", icon: "Document" } },
      { path: "settings", name: "Settings", component: () => import("@/views/settings/index.vue"), meta: { title: "全局设置", icon: "Setting" } }
    ]
  },
  { path: "/:pathMatch(.*)*", redirect: "/" }
];

const router = createRouter({ history: createWebHashHistory(), routes });

let setupChecked = false;
let isFirstRun = false;
let authChecked = false;
let isAuthenticated = false;

// Call after login or setup to skip the next auth check
// but still force a fresh setup-status check on next navigation
export function markAuthenticated() {
  setupChecked = false; // re-check first-run on next navigation
  authChecked = true;
  isAuthenticated = true;
}

// Reset auth state on session expiry (called from 401 interceptor)
export function markUnauthenticated() {
  authChecked = false;
  isAuthenticated = false;
}

async function checkAuth(): Promise<boolean> {
  // Try /api/auth/check first, fall back to /api/status for older backends
  try {
    const res = await fetch("/api/auth/check", { credentials: "same-origin" });
    if (res.ok) return true;
  } catch { /* endpoint may not exist */ }
  try {
    const res = await fetch("/api/status", { credentials: "same-origin" });
    return res.ok;
  } catch { /* network error */ }
  return false;
}

router.beforeEach(async (to, _from) => {
  document.title = `${to.meta.title || "Bililive Helper"} - Bililive Helper`;

  if (to.meta.public) return true;

  // Check first-run status once per session
  if (!setupChecked) {
    try {
      const res = await fetch("/api/setup/status", { credentials: "same-origin" });
      if (res.ok) {
        const data = await res.json();
        isFirstRun = data.first_run;
      }
      // 非 OK 响应（如 500）不改变 isFirstRun，避免服务端错误时误跳转到初始化页面
    } catch {
      // Endpoint doesn't exist (old binary) — cannot determine, skip setup check
    }
    setupChecked = true;
  }

  if (isFirstRun) return "/setup";

  if (authChecked && isAuthenticated) return true;

  const ok = await checkAuth();
  if (ok) {
    authChecked = true;
    isAuthenticated = true;
    return true;
  } else {
    authChecked = false;
    isAuthenticated = false;
    return "/login";
  }
});

export default router;

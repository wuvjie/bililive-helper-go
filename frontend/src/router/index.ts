import { createRouter, createWebHashHistory, type RouteRecordRaw } from "vue-router";
import Layout from "@/layout/index.vue";
import { useAuthStore } from "@/store/modules/auth";

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

async function checkAuth(): Promise<boolean> {
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

  // 认证状态管理（从模块级变量迁移到 Pinia store）
  const auth = useAuthStore();

  // 首次运行检查（每个 session 只检查一次）
  if (!auth.setupChecked) {
    try {
      const res = await fetch("/api/setup/status", { credentials: "same-origin" });
      if (res.ok) {
        const data = await res.json();
        auth.markSetupChecked(data.first_run);
      }
    } catch {
      auth.markSetupChecked(false);
    }
  }

  if (auth.isFirstRun) return "/setup";

  if (auth.isAuthenticated) return true;

  const ok = await checkAuth();
  if (ok) {
    auth.markAuthenticated();
    return true;
  } else {
    auth.markUnauthenticated();
    return "/login";
  }
});

export default router;

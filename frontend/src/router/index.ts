import { createRouter, createWebHashHistory, type RouteRecordRaw } from "vue-router";
import Layout from "@/layout/index.vue";

const routes: RouteRecordRaw[] = [
  {
    path: "/login",
    name: "Login",
    component: () => import("@/views/login/index.vue"),
    meta: { title: "登录" }
  },
  {
    path: "/",
    component: Layout,
    redirect: "/dashboard",
    children: [
      {
        path: "dashboard",
        name: "Dashboard",
        component: () => import("@/views/dashboard/index.vue"),
        meta: { title: "系统概览", icon: "Monitor" }
      },
      {
        path: "streamers",
        name: "Streamers",
        component: () => import("@/views/streamers/index.vue"),
        meta: { title: "主播管理", icon: "User" }
      },
      {
        path: "tasks",
        name: "Tasks",
        component: () => import("@/views/tasks/index.vue"),
        meta: { title: "任务中心", icon: "VideoPlay" }
      },
      {
        path: "history",
        name: "History",
        component: () => import("@/views/history/index.vue"),
        meta: { title: "操作日志", icon: "Document" }
      },
      {
        path: "settings",
        name: "Settings",
        component: () => import("@/views/settings/index.vue"),
        meta: { title: "全局设置", icon: "Setting" }
      }
    ]
  }
];

const router = createRouter({
  history: createWebHashHistory(),
  routes
});

router.beforeEach((to, _from, next) => {
  document.title = `${to.meta.title || "Bililive Helper"} - Bililive Helper`;
  next();
});

export default router;

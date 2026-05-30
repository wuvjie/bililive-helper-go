<template>
  <div class="feishu-app">
    <aside class="sidebar">
      <div class="logo">
        <svg viewBox="0 0 24 24" width="24" height="24" stroke="currentColor" stroke-width="2" fill="none" style="margin-right: 8px; color: #3370ff;">
          <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
        </svg>
        <h2>BiliLive Helper</h2>
      </div>
      <nav class="nav-menu">
        <router-link to="/" class="nav-item">运行状态</router-link>
        <router-link to="/merge" class="nav-item">手动合并</router-link>
        <router-link to="/history" class="nav-item">任务历史</router-link>
        <router-link to="/settings" class="nav-item">系统设置</router-link>
      </nav>
    </aside>

    <main class="main-container">
      <header class="top-header">
        <div class="breadcrumb">管理后台 / <span style="color: #1f2329; font-weight: 500;">{{ currentRouteName }}</span></div>
      </header>
      <div class="content-area">
        <router-view></router-view>
      </div>
    </main>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const routeNameMap = {
  '/': '运行状态',
  '/merge': '手动合并任务',
  '/history': '任务执行历史',
  '/settings': '全局系统设置'
}
const currentRouteName = computed(() => routeNameMap[route.path] || '页面')
</script>

<style>
/* --- 全局飞书风格 CSS --- */
body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  background-color: #f5f6f7;
  color: #1f2329;
}
* { box-sizing: border-box; }

.feishu-app { display: flex; height: 100vh; overflow: hidden; }

/* 侧边栏 */
.sidebar {
  width: 240px;
  background-color: #ffffff;
  border-right: 1px solid #dee0e3;
  display: flex;
  flex-direction: column;
}
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  padding: 0 24px;
  border-bottom: 1px solid #dee0e3;
}
.logo h2 { font-size: 16px; font-weight: 600; margin: 0; }
.nav-menu { padding: 16px 8px; display: flex; flex-direction: column; gap: 4px; }
.nav-item {
  padding: 10px 16px;
  border-radius: 6px;
  text-decoration: none;
  color: #646a73;
  font-size: 14px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
}
.nav-item:hover { background-color: #f5f6f7; }
.nav-item.router-link-active { background-color: #e1eaff; color: #3370ff; font-weight: 500; }

/* 主体 */
.main-container { flex: 1; display: flex; flex-direction: column; }
.top-header {
  height: 60px;
  background-color: #ffffff;
  border-bottom: 1px solid #dee0e3;
  display: flex;
  align-items: center;
  padding: 0 24px;
}
.breadcrumb { font-size: 14px; color: #8f959e; }
.content-area { flex: 1; padding: 24px; overflow-y: auto; background-color: #f5f6f7; }

/* 飞书风格 UI 组件池 (供各个 View 直接使用) */
.feishu-card {
  background: #ffffff; border: 1px solid #dee0e3; border-radius: 8px;
  padding: 24px; margin-bottom: 24px; box-shadow: 0 1px 2px rgba(31, 35, 41, 0.04);
}
.feishu-btn {
  background-color: #3370ff; color: #fff; border: none; padding: 8px 16px;
  border-radius: 6px; font-size: 14px; cursor: pointer; transition: 0.2s;
  display: inline-flex; align-items: center; justify-content: center;
}
.feishu-btn:hover { background-color: #5384ff; }
.feishu-btn:active { background-color: #2458d3; }
.feishu-btn:disabled { background-color: #a9c4ff; cursor: not-allowed; }
.feishu-btn-outline {
  background-color: #ffffff; color: #1f2329; border: 1px solid #dee0e3;
}
.feishu-btn-outline:hover { background-color: #f5f6f7; }

.feishu-input {
  width: 100%; padding: 8px 12px; border: 1px solid #dee0e3; border-radius: 6px;
  font-size: 14px; outline: none; transition: border 0.2s; color: #1f2329;
}
.feishu-input:focus { border-color: #3370ff; box-shadow: 0 0 0 2px rgba(51,112,255,0.1); }
</style>

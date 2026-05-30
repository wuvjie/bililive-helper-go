<template>
  <div class="app-layout">
    <aside class="sidebar">
      <div class="brand">
        <div class="logo-icon">▶</div>
        <h1 class="brand-name">BiliLive Helper</h1>
      </div>

      <div class="nav-group">
        <div class="nav-title">监控大盘</div>
        <router-link to="/" class="nav-item">
          <span class="icon">📊</span> 概览视图
        </router-link>
        <router-link to="/streamers" class="nav-item">
          <span class="icon">👥</span> 主播管理
        </router-link>
      </div>

      <div class="nav-group">
        <div class="nav-title">系统与任务</div>
        <router-link to="/tasks" class="nav-item">
          <span class="icon">⚙️</span> 任务调度
        </router-link>
        <router-link to="/history" class="nav-item">
          <span class="icon">📝</span> 审计日志
        </router-link>
        <router-link to="/settings" class="nav-item">
          <span class="icon">🛠️</span> 全局设置
        </router-link>
      </div>
    </aside>

    <main class="main-wrapper">
      <header class="top-header">
        <div class="header-left">
          <div class="breadcrumb">
            管理后台 <span class="separator">/</span>
            <span class="current">{{ currentRouteName }}</span>
          </div>
        </div>
        <div class="header-right">
          <button class="icon-btn">🔔</button>
          <div class="user-profile">
            <img src="https://api.dicebear.com/7.x/avataaars/svg?seed=Admin" alt="avatar" class="avatar">
            <span class="username">Admin</span>
          </div>
        </div>
      </header>

      <div class="page-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </div>
    </main>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const routeNameMap = {
  '/': '概览视图 (Dashboard)',
  '/streamers': '主播录播管理',
  '/tasks': '任务调度中心',
  '/history': '系统审计日志',
  '/settings': '全局配置'
}
const currentRouteName = computed(() => routeNameMap[route.path] || '页面')
</script>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background-color: var(--bg-body);
}

/* --- 侧边栏 --- */
.sidebar {
  width: 260px;
  background-color: var(--bg-base);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  z-index: 10;
}
.brand {
  height: 64px;
  display: flex;
  align-items: center;
  padding: 0 24px;
  border-bottom: 1px solid var(--border-color);
  gap: 12px;
}
.logo-icon {
  width: 28px;
  height: 28px;
  background: var(--color-primary);
  color: white;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
}
.brand-name { font-size: 16px; font-weight: 600; color: var(--text-title); margin: 0; }

.nav-group { padding: 16px 12px 0; }
.nav-title {
  padding: 0 12px;
  font-size: 12px;
  color: var(--text-placeholder);
  margin-bottom: 8px;
  font-weight: 500;
}
.nav-item {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  border-radius: var(--radius-md);
  color: var(--text-regular);
  text-decoration: none;
  font-size: 14px;
  margin-bottom: 4px;
  transition: all 0.2s;
}
.nav-item .icon { margin-right: 12px; font-size: 16px; }
.nav-item:hover { background-color: var(--bg-hover); color: var(--text-title); }

/* 选中态 */
.nav-item.router-link-active {
  background-color: var(--color-primary-bg);
  color: var(--color-primary);
  font-weight: 500;
}

/* --- 右侧主体 --- */
.main-wrapper { flex: 1; display: flex; flex-direction: column; min-width: 0; }

/* Header */
.top-header {
  height: 64px;
  background-color: var(--bg-base);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.02);
  z-index: 9;
}
.breadcrumb { font-size: 14px; color: var(--text-regular); }
.breadcrumb .separator { margin: 0 8px; color: var(--text-placeholder); }
.breadcrumb .current { color: var(--text-title); font-weight: 500; }

.header-right { display: flex; align-items: center; gap: 20px; }
.icon-btn {
  background: none; border: none; font-size: 18px; cursor: pointer; color: var(--text-regular);
}
.user-profile {
  display: flex; align-items: center; gap: 10px; padding: 4px 8px; border-radius: var(--radius-md); cursor: pointer; transition: 0.2s;
}
.user-profile:hover { background-color: var(--bg-hover); }
.avatar { width: 32px; height: 32px; border-radius: 50%; background: #e1eaff; }
.username { font-size: 14px; font-weight: 500; color: var(--text-title); }

/* 核心内容区 */
.page-content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}

/* 页面切换动画 */
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease, transform 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; transform: translateY(5px); }
</style>

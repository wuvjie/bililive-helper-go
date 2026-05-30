<template>
  <div class="dashboard-page">
    <div class="page-header">
      <h2 class="page-title">概览视图</h2>
      <button class="feishu-btn feishu-btn-outline" @click="refreshData">
        <span class="icon">🔄</span> {{ isRefreshing ? '刷新中...' : '刷新状态' }}
      </button>
    </div>

    <div class="stats-grid">
      <div class="feishu-card stat-card">
        <div class="stat-icon blue"><span class="icon">⏱️</span></div>
        <div class="stat-info">
          <div class="label">系统连续运行</div>
          <div class="value">124 <span class="unit">小时</span></div>
        </div>
      </div>
      <div class="feishu-card stat-card">
        <div class="stat-icon green"><span class="icon">✅</span></div>
        <div class="stat-info">
          <div class="label">累计成功合并</div>
          <div class="value">3,204 <span class="unit">个视频</span></div>
        </div>
      </div>
      <div class="feishu-card stat-card">
        <div class="stat-icon orange"><span class="icon">🔥</span></div>
        <div class="stat-info">
          <div class="label">当前活动任务</div>
          <div class="value">2 <span class="unit">个进行中</span></div>
        </div>
      </div>
    </div>

    <div class="layout-row">
      <div class="feishu-card system-monitor">
        <h3 class="card-title">系统与存储状态</h3>

        <div class="monitor-item">
          <div class="item-header">
            <span class="item-name">/vol2 视频存档盘 (NAS)</span>
            <span class="item-value" :style="{ color: diskColor }">{{ diskUsage }}%</span>
          </div>
          <div class="progress-track">
            <div class="progress-bar" :style="{ width: diskUsage + '%', backgroundColor: diskColor }"></div>
          </div>
          <div class="item-desc">
            容量状态：{{ diskUsage >= 90 ? '即将触发自动清理' : (diskUsage >= 80 ? '逼近安全水位' : '容量充足') }}
            (剩 1.2 TB / 总 8 TB)
          </div>
        </div>

        <div class="monitor-item" style="margin-top: 24px;">
          <div class="item-header">
            <span class="item-name">代理节点 (Mihomo TUN)</span>
          </div>
          <div class="status-box">
            <div class="dot pulsing"></div>
            <span class="status-text">192.168.10.10:7890 正常连接</span>
            <span class="latency">延迟: 12ms</span>
          </div>
        </div>
      </div>

      <div class="feishu-card activity-feed">
        <h3 class="card-title">系统动态</h3>
        <ul class="feed-list">
          <li class="feed-item" v-for="(item, index) in recentActivities" :key="index">
            <div class="feed-time">{{ item.time }}</div>
            <div class="feed-content">
              <span class="feed-tag" :class="item.type">{{ item.tag }}</span>
              {{ item.msg }}
            </div>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const isRefreshing = ref(false)
const diskUsage = ref(82) // 模拟当前磁盘使用率

// 根据容量自动计算进度条颜色
const diskColor = computed(() => {
  if (diskUsage.value >= 90) return 'var(--color-danger)'
  if (diskUsage.value >= 80) return 'var(--color-warning)'
  return 'var(--color-success)'
})

const refreshData = () => {
  isRefreshing.value = true
  setTimeout(() => { isRefreshing.value = false }, 800)
}

const recentActivities = ref([
  { time: '10分钟前', type: 'success', tag: '合并', msg: 'douyin_game01 录像合并完成' },
  { time: '半小时前', type: 'info', tag: '监听', msg: '检测到 bilibili_vup 开播，开始录制' },
  { time: '2小时前', type: 'warning', tag: '清理', msg: '自动清理了 12 个过期碎片文件' },
  { time: '5小时前', type: 'success', tag: '合并', msg: 'dy_outdoor_22 录像合并完成' }
])
</script>

<style scoped>
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-title { margin: 0; font-size: 20px; font-weight: 600; color: var(--text-title); }

/* 数据看板网格 */
.stats-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; margin-bottom: 24px; }
.stat-card { display: flex; align-items: center; padding: 20px 24px; gap: 16px; }
.stat-icon { width: 48px; height: 48px; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 20px; }
.stat-icon.blue { background: var(--color-primary-bg); color: var(--color-primary); }
.stat-icon.green { background: var(--color-success-bg); color: var(--color-success); }
.stat-icon.orange { background: #fff2e8; color: var(--color-warning); }
.stat-info .label { font-size: 13px; color: var(--text-regular); margin-bottom: 4px; }
.stat-info .value { font-size: 24px; font-weight: 600; color: var(--text-title); }
.stat-info .unit { font-size: 12px; font-weight: 400; color: var(--text-placeholder); }

/* 下方两列布局 */
.layout-row { display: grid; grid-template-columns: 3fr 2fr; gap: 20px; }
.card-title { margin: 0 0 20px 0; font-size: 16px; font-weight: 600; }

/* 进度条监控 */
.item-header { display: flex; justify-content: space-between; font-size: 14px; font-weight: 500; margin-bottom: 8px; }
.progress-track { height: 8px; background: #ebedf0; border-radius: 4px; overflow: hidden; margin-bottom: 8px; }
.progress-bar { height: 100%; transition: width 0.5s ease, background-color 0.3s ease; }
.item-desc { font-size: 12px; color: var(--text-placeholder); }

/* 状态指示灯 */
.status-box { display: flex; align-items: center; gap: 10px; background: var(--bg-body); padding: 12px 16px; border-radius: var(--radius-md); border: 1px solid var(--border-color); }
.dot { width: 8px; height: 8px; background: var(--color-success); border-radius: 50%; }
.pulsing { animation: pulse 2s infinite; }
@keyframes pulse {
  0% { box-shadow: 0 0 0 0 rgba(52, 168, 83, 0.4); }
  70% { box-shadow: 0 0 0 6px rgba(52, 168, 83, 0); }
  100% { box-shadow: 0 0 0 0 rgba(52, 168, 83, 0); }
}
.status-text { font-size: 14px; color: var(--text-title); font-weight: 500; flex: 1; }
.latency { font-size: 12px; color: var(--text-regular); background: var(--bg-hover); padding: 2px 6px; border-radius: 4px; }

/* 动态列表 */
.feed-list { list-style: none; padding: 0; margin: 0; }
.feed-item { padding: 12px 0; border-bottom: 1px solid var(--border-color); display: flex; flex-direction: column; gap: 6px; }
.feed-item:last-child { border-bottom: none; }
.feed-time { font-size: 12px; color: var(--text-placeholder); }
.feed-content { font-size: 14px; color: var(--text-title); display: flex; align-items: center; gap: 8px; }
.feed-tag { font-size: 12px; padding: 2px 6px; border-radius: 4px; font-weight: 500; }
.feed-tag.success { background: var(--color-success-bg); color: var(--color-success); }
.feed-tag.info { background: var(--color-primary-bg); color: var(--color-primary); }
.feed-tag.warning { background: #fff2e8; color: var(--color-warning); }
</style>

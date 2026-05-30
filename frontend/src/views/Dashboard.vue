<template>
  <div class="feishu-card">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px;">
      <h3 style="margin: 0; font-size: 18px;">系统概览</h3>
      <button class="feishu-btn feishu-btn-outline" @click="fetchStatus" :disabled="loading">
        {{ loading ? '刷新中...' : '刷新状态' }}
      </button>
    </div>

    <div v-if="error" style="color: #f54a45; margin-bottom: 16px; padding: 12px; background: #fff0f0; border-radius: 6px;">
      {{ error }}
    </div>

    <!-- 统计卡片 -->
    <div class="grid">
      <div class="stat-box">
        <div class="label">主播总数</div>
        <div class="value">{{ streamerCount }}</div>
      </div>
      <div class="stat-box">
        <div class="label">磁盘占用</div>
        <div class="value">{{ diskUsage }}</div>
      </div>
      <div class="stat-box">
        <div class="label">总容量</div>
        <div class="value">{{ totalGB }}</div>
      </div>
      <div class="stat-box">
        <div class="label">监控状态</div>
        <div class="value" :style="{ color: '#00b578' }">正常运行</div>
      </div>
    </div>

    <!-- 主播列表预览 -->
    <div v-if="streamers.length > 0" style="margin-top: 24px;">
      <h4 style="margin: 0 0 16px 0; font-size: 15px; color: #1f2329;">主播列表</h4>
      <div class="streamer-grid">
        <div v-for="s in streamers.slice(0, 6)" :key="s.name" class="streamer-item">
          <div class="streamer-name">{{ s.name }}</div>
          <div class="streamer-info">{{ s.files }} 个文件 · {{ s.size_gb.toFixed(1) }} GB</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { loading, error, getStatus } = useApi()
const statusData = ref({})

const fetchStatus = async () => {
  try {
    const res = await getStatus()
    statusData.value = res.data || res
  } catch (err) {
    console.error(err)
  }
}

const streamers = computed(() => statusData.value.streamers || [])
const streamerCount = computed(() => streamers.value.length)
const diskUsage = computed(() => {
  const val = statusData.value.disk_usage
  return val ? val.toFixed(1) + '%' : '0%'
})
const totalGB = computed(() => {
  const val = statusData.value.total_gb
  return val ? val.toFixed(1) + ' GB' : '0 GB'
})

onMounted(fetchStatus)
</script>

<style scoped>
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 16px; }
.stat-box { background: #f8f9fa; border: 1px solid #dee0e3; border-radius: 8px; padding: 20px; }
.stat-box .label { font-size: 13px; color: #646a73; margin-bottom: 8px; }
.stat-box .value { font-size: 22px; font-weight: 600; color: #1f2329; }

.streamer-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 12px; }
.streamer-item { background: #f8f9fa; border: 1px solid #dee0e3; border-radius: 6px; padding: 12px; }
.streamer-name { font-size: 14px; font-weight: 500; color: #1f2329; margin-bottom: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.streamer-info { font-size: 12px; color: #86909c; }
</style>

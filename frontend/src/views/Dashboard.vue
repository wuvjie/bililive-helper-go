<template>
  <div class="feishu-card">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px;">
      <h3 style="margin: 0; font-size: 18px;">概览视图</h3>
      <button class="feishu-btn feishu-btn-outline" @click="fetchStatus" :disabled="loading">
        {{ loading ? '刷新中...' : '刷新状态' }}
      </button>
    </div>

    <div v-if="error" style="color: #f54a45; margin-bottom: 16px; padding: 12px; background: #fff0f0; border-radius: 6px;">
      {{ error }}
    </div>

    <div class="grid">
      <div v-for="(value, key) in statusData" :key="key" class="stat-box">
        <div class="label">{{ formatKey(key) }}</div>
        <div class="value">{{ value }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { loading, error, getStatus } = useApi()
const statusData = ref({})

const fetchStatus = async () => {
  try {
    const res = await getStatus()
    // 假设后端返回嵌套结构，提取出来
    statusData.value = res.data || res
  } catch (err) {
    console.error(err)
  }
}

// 简单的 key 格式化，把下划线转为空格等
const formatKey = (key) => key.replace(/_/g, ' ').toUpperCase()

onMounted(fetchStatus)
</script>

<style scoped>
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 16px; }
.stat-box { background: #f8f9fa; border: 1px solid #dee0e3; border-radius: 6px; padding: 20px; }
.stat-box .label { font-size: 13px; color: #646a73; margin-bottom: 8px; }
.stat-box .value { font-size: 24px; font-weight: 500; color: #1f2329; word-break: break-all; }
</style>

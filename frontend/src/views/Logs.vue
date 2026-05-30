<template>
  <div class="feishu-card">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
      <h3 style="margin: 0; font-size: 18px;">历史记录</h3>
      <button class="feishu-btn feishu-btn-outline" @click="fetchHistory" :disabled="loading">刷新列表</button>
    </div>

    <div v-if="error" style="color: #f54a45; padding-bottom: 16px;">{{ error }}</div>

    <div class="table-container">
      <table v-if="historyList.length > 0" class="feishu-table">
        <thead>
          <tr>
            <th>时间 (Time)</th>
            <th>类型 (Type)</th>
            <th>详情 (Details)</th>
            <th>状态 (Status)</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(item, index) in historyList" :key="index">
            <td>{{ item.time || item.timestamp || '-' }}</td>
            <td><span class="tag">{{ item.type || item.action || '系统' }}</span></td>
            <td>{{ item.detail || item.message || JSON.stringify(item) }}</td>
            <td :style="{ color: item.status === 'success' ? '#34a853' : (item.status === 'error' ? '#f54a45' : '#1f2329') }">
              {{ item.status || '已完成' }}
            </td>
          </tr>
        </tbody>
      </table>

      <div v-else-if="!loading" class="empty-state">
        暂无历史数据
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { loading, error, getHistory } = useApi()
const historyList = ref([])

const fetchHistory = async () => {
  try {
    const res = await getHistory()
    historyList.value = Array.isArray(res.data) ? res.data : (Array.isArray(res) ? res : [])
  } catch (err) {
    console.error(err)
  }
}

onMounted(fetchHistory)
</script>

<style scoped>
.table-container { overflow-x: auto; border: 1px solid #dee0e3; border-radius: 6px; border-bottom: none; }
.feishu-table { width: 100%; border-collapse: collapse; text-align: left; font-size: 14px; }
.feishu-table th { background: #f8f9fa; padding: 12px 16px; color: #646a73; font-weight: 500; border-bottom: 1px solid #dee0e3; }
.feishu-table td { padding: 12px 16px; border-bottom: 1px solid #dee0e3; color: #1f2329; }
.feishu-table tbody tr:hover { background-color: #f5f6f7; }
.tag { background: #e1eaff; color: #3370ff; padding: 2px 8px; border-radius: 4px; font-size: 12px; }
.empty-state { text-align: center; padding: 40px; color: #8f959e; border-bottom: 1px solid #dee0e3;}
</style>

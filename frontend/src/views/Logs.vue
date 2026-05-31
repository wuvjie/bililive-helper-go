<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const api = useApi()
const historyList = ref([])
const loading = ref(false)

async function fetchLogs() {
  loading.value = true
  try {
    const res = await api.get('/history')
    historyList.value = Array.isArray(res.data) ? res.data : (Array.isArray(res) ? res : [])
  } catch (e) {
    console.error('获取日志失败', e)
  } finally {
    loading.value = false
  }
}

function getLevelClass(level) {
  if (!level) return 'bg-gray-100 text-gray-600'
  const l = level.toLowerCase()
  if (l === 'error' || l === 'fail') return 'bg-[#fff0f0] text-[#f54a45]'
  if (l === 'warning' || l === 'warn') return 'bg-[#fff2e8] text-[#f88339]'
  if (l === 'success') return 'bg-[#e8f8f0] text-[#00b578]'
  return 'bg-[#e8f3ff] text-[#3370ff]'
}

onMounted(fetchLogs)
</script>

<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-5">
      <h1 class="text-[17px] font-semibold text-[#1f2329]">审计日志</h1>
      <button @click="fetchLogs" class="px-4 py-2 border border-[#dee0e3] hover:bg-gray-50 text-[#1f2329] text-sm rounded-lg transition-colors shadow-sm bg-white">
        {{ loading ? '刷新中...' : '刷新记录' }}
      </button>
    </div>

    <div class="bg-white border border-[#dee0e3] rounded-xl overflow-hidden shadow-sm">
      <table class="w-full text-left text-[14px]">
        <thead class="bg-[#f8f9fa] text-[#646a73] border-b border-[#dee0e3]">
          <tr>
            <th class="px-6 py-4 font-medium w-[20%]">发生时间</th>
            <th class="px-6 py-4 font-medium w-[15%]">动作/级别</th>
            <th class="px-6 py-4 font-medium w-[50%]">详细信息</th>
            <th class="px-6 py-4 font-medium w-[15%]">状态</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#dee0e3]">
          <tr v-for="(item, i) in historyList" :key="i" class="hover:bg-[#fcfcfc] transition-colors">
            <td class="px-6 py-4 text-[#646a73]">{{ item.time || item.timestamp || '-' }}</td>
            <td class="px-6 py-4">
              <span :class="['px-2 py-1 rounded text-xs font-medium', getLevelClass(item.level || item.type)]">
                {{ (item.level || item.type || 'INFO').toUpperCase() }}
              </span>
            </td>
            <td class="px-6 py-4 text-[#1f2329]">{{ item.detail || item.message || '-' }}</td>
            <td class="px-6 py-4">
              <span v-if="item.status === 'success'" class="text-[#00b578] font-medium">成功</span>
              <span v-else-if="item.status === 'error'" class="text-[#f54a45] font-medium">失败</span>
              <span v-else class="text-[#1f2329]">{{ item.status || '已完成' }}</span>
            </td>
          </tr>
          <tr v-if="historyList.length === 0 && !loading">
            <td colspan="4" class="px-6 py-12 text-center text-[#8f959e] text-sm">暂无审计日志记录</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

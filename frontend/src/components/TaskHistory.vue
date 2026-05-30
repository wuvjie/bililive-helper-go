<script setup>
import { ref, watch } from 'vue'
import { useApi } from '../composables/useApi'

const { get } = useApi()

const items = ref([])
const total = ref(0)
const pages = ref(0)
const currentPage = ref(1)
const taskFilter = ref('')
const loading = ref(false)

const perPage = 10

async function fetchHistory() {
  loading.value = true

  try {
    const params = new URLSearchParams({
      page: currentPage.value,
      per_page: perPage
    })

    if (taskFilter.value) {
      params.set('task', taskFilter.value)
    }

    const data = await get(`/api/history?${params.toString()}`)

    items.value = data.items || []
    total.value = data.total || 0
    pages.value = data.pages || 0
  } catch (err) {
    console.error('Failed to fetch history:', err)
  } finally {
    loading.value = false
  }
}

function formatHistoryTime(ts) {
  if (!ts) return '--'

  const date = new Date(ts.replace(' ', 'T'))
  if (isNaN(date.getTime())) return ts.substring(5, 16)

  const now = new Date()
  const h = date.getHours().toString().padStart(2, '0')
  const m = date.getMinutes().toString().padStart(2, '0')
  const t = `${h}:${m}`

  if (date.toDateString() === now.toDateString()) {
    return t
  }

  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)

  if (date.toDateString() === yesterday.toDateString()) {
    return `<span style="color:var(--muted)">昨天</span> ${t}`
  }

  return `${(date.getMonth() + 1).toString().padStart(2, '0')}-${date.getDate().toString().padStart(2, '0')} ${t}`
}

function prevPage() {
  if (currentPage.value > 1) {
    currentPage.value--
    fetchHistory()
  }
}

function nextPage() {
  if (currentPage.value < pages.value) {
    currentPage.value++
    fetchHistory()
  }
}

watch(taskFilter, () => {
  currentPage.value = 1
  fetchHistory()
})

// Initial fetch
fetchHistory()
</script>

<template>
  <div style="flex:1;display:flex;flex-direction:column;overflow:hidden">
    <div class="table-container">
      <table>
        <thead>
          <tr>
            <th style="width:25%">时间</th>
            <th style="width:55%">详情</th>
            <th style="width:20%;text-align:center">状态</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in items" :key="r.id">
            <td style="font-family:var(--font-mono);font-size:12px" v-html="formatHistoryTime(r.time)"></td>
            <td style="font-size:12px">{{ r.detail || r.streamer }}</td>
            <td style="text-align:center">
              <span :style="{color: r.status === 'success' ? 'var(--ok)' : 'var(--err)', fontWeight: 'bold'}">
                {{ r.status === 'success' ? '✓' : '✗' }}
              </span>
            </td>
          </tr>
          <tr v-if="items.length === 0">
            <td colspan="3" style="text-align:center;color:var(--faint);height:200px;vertical-align:middle">
              <div style="font-size:12px">暂无操作记录</div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="pages > 1" class="card-foot" style="display:flex;justify-content:center;gap:16px">
      <button class="btn btn-ghost btn-sm auto-w" @click="prevPage" :disabled="currentPage <= 1">
        上一页
      </button>
      <span style="font-size:12px;color:var(--muted);font-family:var(--font-mono);line-height:26px">
        {{ currentPage }} / {{ pages }}
      </span>
      <button class="btn btn-ghost btn-sm auto-w" @click="nextPage" :disabled="currentPage >= pages">
        下一页
      </button>
    </div>
  </div>
</template>

<style scoped>
.table-container {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  scrollbar-gutter: stable;
}
</style>

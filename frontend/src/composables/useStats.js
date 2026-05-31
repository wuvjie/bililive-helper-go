import { ref, onMounted } from 'vue'
import { useApi } from './useApi'

export function useStats() {
  const { get } = useApi()

  const stats = ref({
    today: { merge_count: 0, merge_bytes: 0, clean_count: 0, clean_bytes: 0 },
    month: { merge_count: 0, merge_bytes: 0, clean_count: 0, clean_bytes: 0 },
    daily: []
  })

  async function fetchStats() {
    try {
      const res = await get('/stats')
      if (res) stats.value = res
    } catch (e) {
      console.error('获取统计失败', e)
    }
  }

  onMounted(fetchStats)

  return { stats, fetchStats }
}

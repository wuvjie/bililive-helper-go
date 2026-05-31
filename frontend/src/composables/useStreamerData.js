import { ref, onMounted } from 'vue'
import { useApi } from './useApi'

export function useStreamerData() {
  const { get } = useApi()

  const streamers = ref([])
  const diskUsage = ref(0)
  const totalGB = ref(1)
  const detail = ref({})

  async function fetchStatus() {
    try {
      const res = await get('/status')
      streamers.value = res.streamers || []
      diskUsage.value = res.disk_usage || 0
      totalGB.value = res.total_gb || 1
    } catch (e) {
      console.error('获取状态失败', e)
    }
  }

  async function fetchDetail() {
    try {
      const res = await get('/status/detail')
      detail.value = res
    } catch (e) {
      console.error('获取详情失败', e)
    }
  }

  onMounted(() => {
    fetchStatus()
    fetchDetail()
  })

  return {
    streamers,
    diskUsage,
    totalGB,
    detail,
    fetchStatus,
    fetchDetail,
  }
}

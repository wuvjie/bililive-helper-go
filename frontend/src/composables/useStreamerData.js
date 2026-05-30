import { ref, computed } from 'vue'
import { useApi } from './useApi'

export function useStreamerData() {
  const { get } = useApi()

  const streamers = ref([])
  const diskUsage = ref(0)
  const totalGB = ref(1)
  const detail = ref({
    disk: {},
    pending: {},
    schedule: null
  })

  const d = computed(() => diskUsage.value)

  async function fetchStatus() {
    try {
      const data = await get('/api/status')
      if (data) {
        diskUsage.value = data.disk_usage || 0
        streamers.value = data.streamers || []
        totalGB.value = data.total_gb || 1
      }
    } catch (err) {
      console.error('Failed to fetch status:', err)
    }
  }

  async function fetchDetail() {
    try {
      const data = await get('/api/status/detail')
      if (data) {
        detail.value = data
      }
    } catch (err) {
      console.error('Failed to fetch detail:', err)
    }
  }

  function filtered(keyword) {
    const kw = (keyword || '').toLowerCase().trim()
    const list = streamers.value || []

    if (!kw) {
      return list.slice()
    }

    return list.filter(s => s.name.toLowerCase().includes(kw))
  }

  function sorted(list, by, asc) {
    return [...list].sort((a, b) => {
      let va, vb

      if (by === 'name') {
        va = a.name
        vb = b.name
        return asc ? va.localeCompare(vb) : vb.localeCompare(va)
      } else if (by === 'files') {
        va = a.files
        vb = b.files
      } else {
        va = a.size_gb
        vb = b.size_gb
      }

      return asc ? va - vb : vb - va
    })
  }

  return {
    streamers,
    diskUsage,
    totalGB,
    detail,
    d,
    fetchStatus,
    fetchDetail,
    filtered,
    sorted
  }
}

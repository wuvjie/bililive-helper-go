import { ref, onMounted } from 'vue'
import { useApi } from './useApi'

export function useConfig() {
  const { get, post } = useApi()

  const schedule = ref({})

  async function fetchSchedule() {
    try {
      const res = await get('/schedule')
      schedule.value = res
    } catch (e) {
      console.error('获取调度配置失败', e)
    }
  }

  onMounted(fetchSchedule)

  return {
    schedule,
    fetchSchedule,
  }
}

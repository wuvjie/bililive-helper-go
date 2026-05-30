import { ref } from 'vue'
import { useApi } from './useApi'
import { useToast } from './useToast'

export function useConfig() {
  const { get, post } = useApi()
  const { toast } = useToast()

  const config = ref({})
  const schedule = ref({
    merge_interval: 360,
    clean_interval: 720,
    merge_enabled: true,
    clean_enabled: true
  })

  async function fetchConfig() {
    try {
      const data = await get('/api/config')
      if (data) {
        config.value = data
      }
    } catch (err) {
      console.error('Failed to fetch config:', err)
    }
  }

  async function fetchSchedule() {
    try {
      const data = await get('/api/schedule')
      if (data) {
        schedule.value = data
      }
    } catch (err) {
      console.error('Failed to fetch schedule:', err)
    }
  }

  async function saveConfig() {
    try {
      await post('/api/config', config.value)
      toast('配置保存成功', 'ok')
    } catch (err) {
      toast(err.message || '保存失败', 'err')
      await fetchConfig()
    }
  }

  async function saveSchedule() {
    try {
      const data = await post('/api/schedule', {
        merge_enabled: schedule.value.merge_enabled,
        clean_enabled: schedule.value.clean_enabled,
        merge_interval: schedule.value.merge_interval || 360,
        clean_interval: schedule.value.clean_interval || 720,
        BACKUP_START_HOUR: config.value.BACKUP_START_HOUR,
        BACKUP_START_MINUTE: config.value.BACKUP_START_MINUTE,
        BACKUP_END_HOUR: config.value.BACKUP_END_HOUR,
        BACKUP_END_MINUTE: config.value.BACKUP_END_MINUTE
      })

      if (data && data.schedule) {
        schedule.value = data.schedule
      }

      await fetchConfig()
      toast('调度策略已更新', 'ok')
    } catch (err) {
      toast(err.message || '保存失败', 'err')
    }
  }

  function resetSchedule() {
    schedule.value.merge_interval = 360
    schedule.value.clean_interval = 720
    schedule.value.merge_enabled = true
    schedule.value.clean_enabled = true

    config.value.BACKUP_START_HOUR = 4
    config.value.BACKUP_START_MINUTE = 0
    config.value.BACKUP_END_HOUR = 12
    config.value.BACKUP_END_MINUTE = 0

    saveSchedule()
    toast('已恢复默认出厂设置', 'ok')
  }

  return {
    config,
    schedule,
    fetchConfig,
    fetchSchedule,
    saveConfig,
    saveSchedule,
    resetSchedule
  }
}

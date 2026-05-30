import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAppStore = defineStore('app', () => {
  const sidebarExpanded = ref(false)
  const isRunning = ref(false)
  const notifications = ref(0)

  function toggleSidebar() {
    sidebarExpanded.value = !sidebarExpanded.value
  }

  function setRunning(value) {
    isRunning.value = value
  }

  function addNotification() {
    notifications.value++
  }

  function clearNotifications() {
    notifications.value = 0
  }

  return {
    sidebarExpanded,
    isRunning,
    notifications,
    toggleSidebar,
    setRunning,
    addNotification,
    clearNotifications
  }
})

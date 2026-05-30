import { defineStore } from 'pinia'
import { ref, onMounted, onUnmounted } from 'vue'

export const useAppStore = defineStore('app', () => {
  const sidebarExpanded = ref(false)
  const isRunning = ref(false)
  const notifications = ref(0)
  const isMobile = ref(false)

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

  function checkMobile() {
    isMobile.value = window.innerWidth < 768
    if (isMobile.value) {
      sidebarExpanded.value = false
    }
  }

  onMounted(() => {
    checkMobile()
    window.addEventListener('resize', checkMobile)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', checkMobile)
  })

  return {
    sidebarExpanded,
    isRunning,
    notifications,
    isMobile,
    toggleSidebar,
    setRunning,
    addNotification,
    clearNotifications
  }
})

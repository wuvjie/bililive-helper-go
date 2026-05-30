import { ref } from 'vue'

// 如果前后端分离开发，可以改为完整的后端地址，例如 http://localhost:8080/api
const BASE_URL = '/api'

export function useApi() {
  const loading = ref(false)
  const error = ref(null)

  const fetchApi = async (endpoint, options = {}) => {
    loading.value = true
    error.value = null
    try {
      const token = localStorage.getItem('token') || ''
      const headers = {
        'Content-Type': 'application/json',
        ...(token ? { 'Authorization': `Bearer ${token}` } : {})
      }

      const response = await fetch(`${BASE_URL}${endpoint}`, {
        ...options,
        headers: { ...headers, ...options.headers }
      })

      if (response.status === 401) {
        window.location.href = '/login'
        throw new Error('未授权，请重新登录')
      }

      const data = await response.json()

      if (!response.ok) {
        throw new Error(data.message || data.error || `请求失败: ${response.status}`)
      }

      return data
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    getStatus: () => fetchApi('/status'),
    getConfig: () => fetchApi('/config'),
    saveConfig: (data) => fetchApi('/config', { method: 'POST', body: JSON.stringify(data) }),
    getHistory: () => fetchApi('/history'),
    triggerMerge: (data) => fetchApi('/merge', { method: 'POST', body: JSON.stringify(data) })
  }
}

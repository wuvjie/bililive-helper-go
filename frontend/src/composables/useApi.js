import { ref } from 'vue'
import { useApi } from './useApi'

const BASE_URL = '/api'

export function useApi() {
  const loading = ref(false)
  const error = ref(null)

  const request = async (url, options = {}) => {
    loading.value = true
    error.value = null
    try {
      const token = localStorage.getItem('token') || ''
      const headers = {
        'Content-Type': 'application/json',
        ...(token ? { 'Authorization': `Bearer ${token}` } : {})
      }

      const response = await fetch(`${BASE_URL}${url}`, {
        ...options,
        headers: { ...headers, ...options.headers }
      })

      if (response.status === 401) {
        window.location.href = '/login'
        throw new Error('未授权')
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
    get: (url) => request(url),
    post: (url, data) => request(url, { method: 'POST', body: JSON.stringify(data) }),
  }
}

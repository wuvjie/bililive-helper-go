import { ref } from 'vue'

export function useApi() {
  const error = ref(null)

  async function request(url, options = {}) {
    error.value = null

    const defaultOptions = {
      headers: {
        'Accept': 'application/json',
        ...options.headers
      },
      redirect: 'manual'
    }

    const mergedOptions = { ...defaultOptions, ...options }

    try {
      const response = await fetch(url, mergedOptions)

      // Handle 401 unauthorized
      if (response.status === 401 || response.type === 'opaqueredirect') {
        window.location.href = '/login'
        throw new Error('未登录')
      }

      // Handle 403 forbidden
      if (response.status === 403) {
        throw new Error('无权限')
      }

      // Parse response
      const contentType = response.headers.get('content-type') || ''

      if (contentType.includes('application/json')) {
        const data = await response.json()
        if (!response.ok) {
          throw new Error(data.error || '请求失败')
        }
        return data
      } else {
        const text = await response.text()
        if (!response.ok) {
          throw new Error(text || '请求失败')
        }
        return text
      }
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function get(url) {
    return request(url)
  }

  async function post(url, data) {
    return request(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    })
  }

  return {
    error,
    get,
    post,
    request
  }
}

import { ref } from 'vue'

export function useSSE() {
  const lines = ref([])
  const isRunning = ref(false)
  const error = ref(null)

  function addLine(text, type = 'info') {
    const time = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
    lines.value.push({ time, text, type })
  }

  function clear() {
    lines.value = []
    error.value = null
  }

  async function startSSE(url, body = null) {
    isRunning.value = true
    error.value = null

    try {
      const token = localStorage.getItem('token') || ''
      const opts = {
        method: body ? 'POST' : 'GET',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      }
      if (body) {
        opts.headers['Content-Type'] = 'application/json'
        opts.body = JSON.stringify(body)
      }

      const response = await fetch(url, opts)

      if (!response.ok) {
        const err = await response.json().catch(() => ({}))
        const msg = err.error || response.statusText
        addLine(`错误: ${msg}`, 'error')
        error.value = msg
        return false
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const parts = buffer.split('\n')
        buffer = parts.pop() || ''

        for (const line of parts) {
          if (line.startsWith('data: ')) {
            const text = line.slice(6).trim()
            if (text === '[END]') {
              return true
            }
            const type = text.startsWith('❌') || text.includes('错误') ? 'error'
              : text.startsWith('✅') || text.includes('完成') ? 'success' : 'info'
            addLine(text, type)
          }
        }
      }
      return true
    } catch (e) {
      addLine(`连接错误: ${e.message}`, 'error')
      error.value = e.message
      return false
    } finally {
      isRunning.value = false
    }
  }

  return { lines, isRunning, error, addLine, clear, startSSE }
}

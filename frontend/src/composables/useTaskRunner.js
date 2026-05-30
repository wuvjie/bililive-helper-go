import { ref } from 'vue'
import { useToast } from './useToast'

export function useTaskRunner() {
  const { toast } = useToast()

  const running = ref(false)
  const progress = ref(0)
  const progressLabel = ref('')
  const logs = ref('')
  const autoScroll = ref(true)

  function run(task, streamer, callbacks = {}) {
    running.value = true
    progress.value = 0
    progressLabel.value = '启动中...'
    logs.value = '启动中...\n'

    const url = `/api/run/${task}?streamer=${encodeURIComponent(streamer || '')}`

    const es = new EventSource(url)

    es.onmessage = (event) => {
      if (event.data === '[END]') {
        es.close()
        running.value = false
        progress.value = 0
        progressLabel.value = ''
        if (callbacks.onComplete) callbacks.onComplete()
      } else {
        logs.value += event.data + '\n'

        // Parse progress
        const progressMatch = event.data.match(/合并进度\s*(\d+)%/)
        if (progressMatch) {
          progress.value = parseInt(progressMatch[1])
          progressLabel.value = `合并进度 ${progressMatch[1]}%`
        } else if (event.data.includes('合并中') ||
                   event.data.includes('清理') ||
                   event.data.includes('发现')) {
          progressLabel.value = event.data.replace(/^[⏳▶ℹ✅❌📊🔍🗑]\s*/, '')
        }

        if (autoScroll.value) {
          requestAnimationFrame(() => {
            const term = document.getElementById('term')
            if (term) term.scrollTop = term.scrollHeight
          })
        }
      }
    }

    es.onerror = () => {
      es.close()
      logs.value += '\n❌ 任务执行中断\n'
      running.value = false
      progress.value = 0
      progressLabel.value = ''
      if (callbacks.onError) callbacks.onError()
    }

    toast(`${task === 'merge' ? '合并' : '清理'} 任务已下发`, 'ok')
  }

  function startManualMerge(streamer, files, callbacks = {}) {
    if (files.length < 2) return

    running.value = true
    progress.value = 0
    progressLabel.value = '启动手动合并...'
    logs.value = '启动手动合并...\n'

    fetch('/api/merge/manual', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ streamer, files: [...files] })
    })
    .then(async (response) => {
      if (!response.ok) throw new Error('请求拒绝')

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      const read = async () => {
        const { done, value } = await reader.read()

        if (done) {
          running.value = false
          progress.value = 0
          progressLabel.value = ''
          if (callbacks.onComplete) callbacks.onComplete()
          return
        }

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop()

        for (const line of lines) {
          const trimmed = line.trim()
          if (trimmed.startsWith('data: ')) {
            const content = trimmed.substring(6)

            if (content === '[END]') {
              running.value = false
              progress.value = 0
              progressLabel.value = ''
              if (callbacks.onComplete) callbacks.onComplete()
            } else {
              logs.value += content + '\n'
              const progressMatch = content.match(/合并进度\s*(\d+)%/)
              if (progressMatch) {
                progress.value = parseInt(progressMatch[1])
                progressLabel.value = `合并进度 ${progressMatch[1]}%`
              }

              if (autoScroll.value) {
                requestAnimationFrame(() => {
                  const term = document.getElementById('term')
                  if (term) term.scrollTop = term.scrollHeight
                })
              }
            }
          }
        }

        await read()
      }

      await read()
      toast('手动合并已排入队列', 'ok')
    })
    .catch((err) => {
      running.value = false
      logs.value += `\n❌ ${err.message || '请求失败'}\n`
      toast('下发失败', 'err')
    })
  }

  return {
    running,
    progress,
    progressLabel,
    logs,
    autoScroll,
    run,
    startManualMerge
  }
}

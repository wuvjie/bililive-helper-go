# Frontend Refactoring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor the frontend from a 934-line single-file Vue 3 CDN SPA to a Vite + Vue 3 component-based architecture with composables.

**Architecture:** Create 8 Vue components, 5 composables, use Vite for build, and maintain existing UI/CSS design system. Components will be organized by responsibility (dashboard, list, config, logs), composables will handle API calls, state management, and task execution.

**Tech Stack:** Vite 5, Vue 3 (Composition API + `<script setup>`), JavaScript, CSS Variables (existing design system)

---

## File Structure

```
/frontend (NEW)
├── index.html
├── vite.config.js
├── package.json
└── src/
    ├── main.js
    ├── App.vue
    ├── components/
    │   ├── TopNav.vue (~30 lines)
    │   ├── StatusDashboard.vue (~120 lines)
    │   ├── StreamerList.vue (~180 lines)
    │   ├── TaskHistory.vue (~80 lines)
    │   ├── LogViewer.vue (~100 lines)
    │   ├── ConfigPanel.vue (~250 lines)
    │   ├── TaskScheduler.vue (~80 lines)
    │   └── ManualMerge.vue (~100 lines)
    ├── composables/
    │   ├── useApi.js (~60 lines)
    │   ├── useStreamerData.js (~80 lines)
    │   ├── useConfig.js (~70 lines)
    │   ├── useTaskRunner.js (~90 lines)
    │   └── useToast.js (~30 lines)
    └── styles/
        ├── variables.css (~80 lines)
        └── base.css (~300 lines)
```

---

## Task 1: Setup Vite Project

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.js`
- Create: `frontend/index.html`
- Create: `frontend/src/main.js`

- [ ] **Step 1: Create package.json**

Create `frontend/package.json`:
```json
{
  "name": "bililive-helper-frontend",
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "vue": "^3.4.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0.0",
    "vite": "^5.0.0"
  }
}
```

- [ ] **Step 2: Create vite.config.js**

Create `frontend/vite.config.js`:
```javascript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:5000',
        changeOrigin: true
      },
      '/login': {
        target: 'http://localhost:5000',
        changeOrigin: true
      },
      '/logout': {
        target: 'http://localhost:5000',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true
  }
})
```

- [ ] **Step 3: Create index.html**

Create `frontend/index.html`:
```html
<!DOCTYPE html>
<html lang="zh">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Bililive Helper</title>
</head>
<body>
  <div id="app"></div>
  <script type="module" src="/src/main.js"></script>
</body>
</html>
```

- [ ] **Step 4: Create main.js**

Create `frontend/src/main.js`:
```javascript
import { createApp } from 'vue'
import App from './App.vue'

const app = createApp(App)
app.mount('#app')
```

- [ ] **Step 5: Create placeholder App.vue**

Create `frontend/src/App.vue`:
```vue
<script setup>
</script>

<template>
  <div>Bililive Helper - Refactored</div>
</template>

<style>
</style>
```

- [ ] **Step 6: Install dependencies and verify dev server**

Run: `cd frontend && npm install && npm run dev`
Expected: Vite dev server starts on http://localhost:5173

- [ ] **Step 7: Commit**

```bash
git add frontend/
git commit -m "feat: setup Vite project with Vue 3"
```

---

## Task 2: Migrate CSS Variables and Base Styles

**Files:**
- Create: `frontend/src/styles/variables.css`
- Create: `frontend/src/styles/base.css`
- Modify: `frontend/src/main.js`

- [ ] **Step 1: Create variables.css**

Create `frontend/src/styles/variables.css` (extract from templates/index.html lines 14-57):
```css
/* --- 全局 Design Tokens --- */
:root {
  color-scheme: light;
  --row-h: 52px;
  --gap: 24px;
  --radius: 10px;
  --card-shadow: 0 1px 3px rgba(0,0,0,0.04), 0 8px 24px -8px rgba(0,0,0,0.08);
  
  --bg: #fafafa;
  --bg-sub: #f5f5f5;
  --card: #ffffff;
  --text: #171717;
  --text2: #666666;
  --muted: #888888;
  --faint: #a3a3a3;
  --border: #eaeaea;
  --border-sub: #f0f0f0;
  --hover: #f9f9f9;
  
  --pri: #000000;
  --pri-fg: #ffffff;
  --pri-r: rgba(0, 0, 0, 0.15);
  --pri-s: rgba(0, 0, 0, 0.04);
  
  --ok: #059669; --ok-s: #ecfdf5;
  --err: #e11d48; --err-s: #fff1f2;
  --warn: #d97706; --warn-s: #fffbeb;
  --info: #0284c7; --info-s: #f0f9ff;
  
  --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  --font-mono: 'Fira Code', ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

@media (prefers-color-scheme: dark) {
  :root {
    color-scheme: dark;
    --bg: #0a0a0a; --bg-sub: #111111; --card: #141414;
    --text: #ededed; --text2: #a1a1aa; --muted: #71717a; --faint: #3f3f46;
    --border: #27272a; --border-sub: #1e1e1e; --hover: #1f1f1f;
    --pri: #ffffff; --pri-fg: #000000;
    --pri-r: rgba(255, 255, 255, 0.2); --pri-s: rgba(255, 255, 255, 0.08);
  }
}
```

- [ ] **Step 2: Create base.css**

Create `frontend/src/styles/base.css` (extract from templates/index.html lines 59-376):
```css
* { box-sizing: border-box; margin: 0; padding: 0; }
body { background: var(--bg); color: var(--text); font: 14px/1.6 var(--font-sans); -webkit-font-smoothing: antialiased; }
#app { max-width: 1400px; margin: 0 auto; padding: 24px 32px 64px; }
@media (max-width: 768px) { #app { padding: 16px; } }

/* 按钮样式 */
.btn { display: inline-flex; align-items: center; justify-content: center; gap: 6px; border: 1px solid transparent; border-radius: 6px; padding: 0 14px; height: 32px; font: 500 13px var(--font-sans); cursor: pointer; transition: all 0.2s; white-space: nowrap; user-select: none; width: 100%;}
.btn.auto-w { width: auto; }
.btn:focus-visible { outline: 2px solid var(--pri-r); outline-offset: 2px; }
.btn:active:not(:disabled) { transform: scale(0.97); }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.btn svg { width: 14px; height: 14px; }
.btn-pri { background: var(--pri); color: var(--pri-fg); }
.btn-pri:hover:not(:disabled) { opacity: 0.8; box-shadow: 0 4px 12px var(--pri-r); }
.btn-ghost { background: transparent; color: var(--text2); border-color: var(--border); }
.btn-ghost:hover:not(:disabled) { background: var(--hover); color: var(--text); }
.btn-ok { background: var(--ok); color: #fff; }
.btn-err { background: var(--err); color: #fff; }
.btn-sm { height: 26px !important; padding: 0 10px !important; font-size: 12px !important; }

/* 表单样式 */
input, select { height: 32px; border: 1px solid var(--border); border-radius: 6px; padding: 0 10px; font: 13px var(--font-sans); outline: none; background: var(--bg-sub); color: var(--text); transition: all 0.2s; width: 100%; max-width: 100%;}
input:hover, select:hover { border-color: var(--muted); }
input:focus, select:focus { border-color: var(--pri); background: var(--card); box-shadow: 0 0 0 3px var(--pri-r); }
select { appearance: none; -webkit-appearance: none; padding-right: 32px !important; background-image: url("data:image/svg+xml;charset=UTF-8,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23888888' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3e%3cpolyline points='6 9 12 15 18 9'%3e%3c/polyline%3e%3c/svg%3e") !important; background-repeat: no-repeat !important; background-position: right 10px center !important; background-size: 14px !important; cursor: pointer; }

/* 卡片样式 */
.card { background: var(--card); border: 1px solid var(--border); border-radius: var(--radius); box-shadow: var(--card-shadow); overflow: hidden; display: flex; flex-direction: column; transition: all 0.3s ease; min-width: 0; }
.card-head { height: var(--row-h); padding: 0 16px; border-bottom: 1px solid var(--border); background: color-mix(in srgb, var(--bg-sub) 50%, transparent); display: flex; align-items: center; flex-shrink: 0; }
.card-body { padding: 20px; flex: 1; overflow-y: auto; display: flex; flex-direction: column; min-height: 0; }
.card-foot { padding: 12px 20px; background: var(--bg-sub); border-top: 1px solid var(--border); }

/* 徽章样式 */
.badge { display: inline-flex; align-items: center; font-size: 11px; font-weight: 600; padding: 2px 8px; border-radius: 12px; white-space: nowrap; }
.badge-warn { background: var(--warn-s); color: var(--warn); }
.badge-ok { background: var(--ok-s); color: var(--ok); }
.badge-err { background: var(--err-s); color: var(--err); }
.badge-pri { background: var(--pri-s); color: var(--text); }

/* 开关样式 */
.switch { position: relative; display: inline-block; width: 36px; height: 20px; flex-shrink: 0; }
.switch input { opacity: 0; width: 0; height: 0; position: absolute; }
.switch-slider { position: absolute; cursor: pointer; inset: 0; background-color: var(--border-sub); border: 1px solid var(--border); transition: 0.3s ease; border-radius: 20px; }
.switch-slider:before { position: absolute; content: ""; height: 14px; width: 14px; left: 2px; bottom: 2px; background-color: var(--card); transition: 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275); border-radius: 50%; box-shadow: 0 2px 4px rgba(0,0,0,0.15); border: 1px solid var(--border); }
.switch input:checked + .switch-slider { background-color: var(--pri); border-color: var(--pri); }
.switch input:checked + .switch-slider:before { transform: translateX(16px); border-color: transparent; }

/* 滚动条 */
::-webkit-scrollbar { width: 6px; height: 6px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: var(--border); border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: var(--muted); }

/* 表格 */
table { width: 100%; border-collapse: separate; border-spacing: 0; text-align: left; table-layout: fixed; }
th { position: sticky; top: 0; z-index: 10; background: color-mix(in srgb, var(--bg) 92%, transparent); backdrop-filter: blur(8px); padding: 12px 16px; font-size: 11px; font-weight: 600; color: var(--muted); text-transform: uppercase; letter-spacing: 0.05em; border-bottom: 1px solid var(--border); }
td { padding: 10px 16px; font-size: 13px; border-bottom: 1px solid var(--border-sub); transition: background 0.2s; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;}
tr:last-child td { border-bottom: none !important; }
tr:hover td { background: var(--hover); }
```

- [ ] **Step 3: Import styles in main.js**

Modify `frontend/src/main.js`:
```javascript
import { createApp } from 'vue'
import App from './App.vue'
import './styles/variables.css'
import './styles/base.css'

const app = createApp(App)
app.mount('#app')
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/styles/ frontend/src/main.js
git commit -m "feat: migrate CSS variables and base styles"
```

---

## Task 3: Create useApi Composable

**Files:**
- Create: `frontend/src/composables/useApi.js`

- [ ] **Step 1: Create useApi.js**

Create `frontend/src/composables/useApi.js`:
```javascript
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
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/composables/useApi.js
git commit -m "feat: create useApi composable"
```

---

## Task 4: Create useToast Composable

**Files:**
- Create: `frontend/src/composables/useToast.js`

- [ ] **Step 1: Create useToast.js**

Create `frontend/src/composables/useToast.js`:
```javascript
let toastContainer = null

function getContainer() {
  if (!toastContainer) {
    toastContainer = document.createElement('div')
    toastContainer.id = 'toasts'
    toastContainer.style.cssText = 'position:fixed;top:24px;right:24px;z-index:300;display:flex;flex-direction:column;gap:12px;pointer-events:none;'
    document.body.appendChild(toastContainer)
  }
  return toastContainer
}

export function useToast() {
  function toast(message, type = 'info') {
    const container = getContainer()
    const el = document.createElement('div')
    el.className = `toast toast-${type}`
    el.textContent = message
    el.style.cssText = 'padding:12px 20px;border-radius:8px;font:500 13px var(--font-sans);color:#fff;pointer-events:auto;animation:tin 0.3s cubic-bezier(0.16, 1, 0.3, 1),tout 0.3s ease 3s forwards;box-shadow:0 8px 16px rgba(0,0,0,0.1);'
    
    if (type === 'ok') {
      el.style.background = 'var(--text)'
    } else if (type === 'err') {
      el.style.background = 'var(--err)'
    } else {
      el.style.background = 'var(--info)'
    }
    
    container.appendChild(el)
    
    setTimeout(() => {
      el.remove()
    }, 3500)
  }

  return {
    toast
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/composables/useToast.js
git commit -m "feat: create useToast composable"
```

---

## Task 5: Create useStreamerData Composable

**Files:**
- Create: `frontend/src/composables/useStreamerData.js`

- [ ] **Step 1: Create useStreamerData.js**

Create `frontend/src/composables/useStreamerData.js`:
```javascript
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
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/composables/useStreamerData.js
git commit -m "feat: create useStreamerData composable"
```

---

## Task 6: Create useConfig Composable

**Files:**
- Create: `frontend/src/composables/useConfig.js`

- [ ] **Step 1: Create useConfig.js**

Create `frontend/src/composables/useConfig.js`:
```javascript
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
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/composables/useConfig.js
git commit -m "feat: create useConfig composable"
```

---

## Task 7: Create useTaskRunner Composable

**Files:**
- Create: `frontend/src/composables/useTaskRunner.js`

- [ ] **Step 1: Create useTaskRunner.js**

Create `frontend/src/composables/useTaskRunner.js`:
```javascript
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
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/composables/useTaskRunner.js
git commit -m "feat: create useTaskRunner composable"
```

---

## Task 8: Create TopNav Component

**Files:**
- Create: `frontend/src/components/TopNav.vue`

- [ ] **Step 1: Create TopNav.vue**

Create `frontend/src/components/TopNav.vue`:
```vue
<script setup>
defineProps({
  running: {
    type: Boolean,
    default: false
  }
})
</script>

<template>
  <header class="topbar">
    <div class="topbar-brand">
      <div class="topbar-logo">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/>
        </svg>
      </div>
      <h1>Bililive Helper</h1>
    </div>
    <div class="topbar-right">
      <span v-if="running" class="badge badge-warn">● 运行中</span>
      <a href="/logout">退出登录</a>
    </div>
  </header>
</template>

<style scoped>
.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  margin-bottom: 24px;
  position: sticky;
  top: 0;
  z-index: 40;
  background: color-mix(in srgb, var(--bg) 80%, transparent);
  backdrop-filter: blur(12px);
}

.topbar-brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.topbar-logo {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--pri);
  display: flex;
  align-items: center;
  justify-content: center;
}

.topbar-logo svg {
  width: 18px;
  height: 18px;
  color: var(--pri-fg);
}

.topbar h1 {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: -0.5px;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.topbar-right a {
  color: var(--muted);
  text-decoration: none;
  font-size: 13px;
  font-weight: 500;
  transition: color 0.2s;
}

.topbar-right a:hover {
  color: var(--text);
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/TopNav.vue
git commit -m "feat: create TopNav component"
```

---

## Task 9: Create StatusDashboard Component

**Files:**
- Create: `frontend/src/components/StatusDashboard.vue`

- [ ] **Step 1: Create StatusDashboard.vue**

Create `frontend/src/components/StatusDashboard.vue`:
```vue
<script setup>
defineProps({
  diskUsage: {
    type: Number,
    default: 0
  },
  totalGB: {
    type: Number,
    default: 1
  },
  schedule: {
    type: Object,
    default: null
  },
  running: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['run'])

function formatSimpleTime(ts) {
  if (!ts || ts <= 0) return '--:--'
  
  const date = new Date(ts * 1000)
  const now = new Date()
  const diff = (date - now) / 1000 / 60
  
  const h = date.getHours().toString().padStart(2, '0')
  const m = date.getMinutes().toString().padStart(2, '0')
  const t = `${h}:${m}`
  
  if (Math.abs(diff) < 1) return '即将'
  
  if (diff < 0) {
    const ago = Math.abs(diff)
    if (ago < 60) return `${Math.round(ago)}m 前`
    return `${Math.round(ago / 60)}h 前`
  }
  
  return `${t} (${diff < 60 ? Math.round(diff) + 'm' : Math.round(diff / 60) + 'h'}后)`
}
</script>

<template>
  <section class="statusbar" aria-label="系统状态">
    <div class="gauge">
      <div class="gauge-val" :style="{color: diskUsage > 85 ? 'var(--err)' : diskUsage > 65 ? 'var(--warn)' : 'var(--text)'}">
        {{ diskUsage.toFixed(1) }}%
      </div>
      <div class="gauge-lbl">
        磁盘占用
        <span style="font-family:var(--font-mono);font-weight:500;color:var(--faint);margin-left:6px">
          · {{ (totalGB * (diskUsage / 100)).toFixed(1) }}G / {{ totalGB.toFixed(1) }}G
        </span>
      </div>
      <div class="gauge-track">
        <div class="gauge-fill" :style="{width: diskUsage + '%', background: diskUsage > 85 ? 'var(--err)' : diskUsage > 65 ? 'var(--warn)' : 'var(--text)'}"></div>
      </div>
    </div>
    
    <div class="tasks" v-if="schedule">
      <div class="task-widget">
        <div class="task-widget-head">
          <span class="task-dot" :class="schedule.merge?.enabled ? 'on' : 'off'" style="margin-right:6px"></span>
          自动合并
        </div>
        <div class="task-widget-data">
          <span class="task-widget-time" :style="{color: !schedule.merge?.enabled ? 'var(--faint)' : ''}">
            {{ schedule.merge?.enabled ? formatSimpleTime(schedule.merge?.next_run) : '-- : --' }}
          </span>
          <span class="task-widget-interval">
            {{ schedule.merge?.enabled ? schedule.merge?.interval + ' MIN' : '已暂停' }}
          </span>
        </div>
      </div>
      
      <div class="task-widget">
        <div class="task-widget-head">
          <span class="task-dot" :class="schedule.clean?.enabled ? 'on' : 'off'" style="margin-right:6px"></span>
          自动清理
        </div>
        <div class="task-widget-data">
          <span class="task-widget-time" :style="{color: !schedule.clean?.enabled ? 'var(--faint)' : ''}">
            {{ schedule.clean?.enabled ? formatSimpleTime(schedule.clean?.next_run) : '-- : --' }}
          </span>
          <span class="task-widget-interval">
            {{ schedule.clean?.enabled ? schedule.clean?.interval + ' MIN' : '已暂停' }}
          </span>
        </div>
      </div>
    </div>
    
    <div class="actions">
      <button class="btn btn-ghost auto-w" @click="emit('run', 'merge', '')" :disabled="running">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/>
          <circle cx="9" cy="7" r="4"/>
          <path d="M23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/>
        </svg>
        全局合并
      </button>
      <button class="btn btn-ghost auto-w" style="color:var(--err)" @click="emit('run', 'clean', '')" :disabled="running">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <polyline points="3 6 5 6 21 6"/>
          <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
        </svg>
        全局清理
      </button>
    </div>
  </section>
</template>

<style scoped>
.statusbar {
  display: grid;
  grid-template-columns: 200px 1fr auto;
  gap: 40px;
  align-items: center;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 24px 32px;
  margin-bottom: 24px;
  box-shadow: var(--card-shadow);
}

.gauge {
  text-align: center;
}

.gauge-val {
  font-family: var(--font-mono);
  font-size: 28px;
  font-weight: 700;
  line-height: 1;
  letter-spacing: -1px;
}

.gauge-lbl {
  font-size: 12px;
  color: var(--muted);
  font-weight: 500;
  margin-top: 6px;
}

.gauge-track {
  height: 4px;
  border-radius: 2px;
  background: var(--border-sub);
  margin-top: 10px;
  overflow: hidden;
}

.gauge-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.8s cubic-bezier(0.22, 1, 0.36, 1);
  background-image: linear-gradient(45deg, rgba(255,255,255,0.15) 25%, transparent 25%, transparent 50%, rgba(255,255,255,0.15) 50%, rgba(255,255,255,0.15) 75%, transparent 75%, transparent);
  background-size: 1rem 1rem;
  animation: barberpole 2s linear infinite;
}

@keyframes barberpole {
  100% { background-position: 100% 100%; }
}

.tasks {
  display: flex;
  align-items: center;
  gap: 48px;
  flex: 1;
  padding-left: 40px;
  border-left: 1px solid var(--border);
  height: 56px;
}

.task-widget {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
  min-width: 160px;
}

.task-widget-head {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 500;
  color: var(--muted);
}

.task-widget-data {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.task-widget-time {
  font-family: var(--font-mono);
  font-size: 16px;
  font-weight: 600;
  color: var(--text);
}

.task-widget-interval {
  font-size: 11px;
  font-weight: 600;
  color: var(--text2);
  background: var(--border-sub);
  padding: 2px 6px;
  border-radius: 4px;
}

.task-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.task-dot.on {
  background: var(--ok);
  box-shadow: 0 0 0 3px rgba(5,150,105,0.15);
}

.task-dot.off {
  background: var(--border);
}

.actions {
  display: flex;
  gap: 12px;
}

@media (max-width: 1024px) {
  .statusbar {
    grid-template-columns: 1fr;
    gap: 16px;
    padding: 20px;
  }
  
  .gauge {
    text-align: left;
    display: flex;
    align-items: center;
    gap: 12px;
  }
  
  .gauge-track {
    display: none;
  }
  
  .tasks {
    flex-wrap: wrap;
    gap: 16px;
    padding-left: 0;
    border-left: none;
    height: auto;
    border-top: 1px solid var(--border);
    padding-top: 12px;
  }
  
  .task-widget {
    min-width: 0;
    flex: 1;
  }
  
  .actions {
    width: 100%;
    flex-wrap: wrap;
    gap: 8px;
  }
  
  .actions .btn {
    flex: 1;
    min-width: 0;
  }
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/StatusDashboard.vue
git commit -m "feat: create StatusDashboard component"
```

---

## Task 10: Create StreamerList Component

**Files:**
- Create: `frontend/src/components/StreamerList.vue`

- [ ] **Step 1: Create StreamerList.vue**

Create `frontend/src/components/StreamerList.vue`:
```vue
<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  streamers: {
    type: Array,
    default: () => []
  },
  totalGB: {
    type: Number,
    default: 1
  },
  running: {
    type: Boolean,
    default: false
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['run', 'openManualMerge', 'refresh'])

const searchQuery = ref('')
const currentPage = ref(1)
const sortBy = ref('size')
const sortAsc = ref(false)
const pageSize = 10

const filtered = computed(() => {
  const kw = searchQuery.value.toLowerCase().trim()
  let list = props.streamers || []
  
  if (kw) {
    list = list.filter(s => s.name.toLowerCase().includes(kw))
  }
  
  return list
})

const sorted = computed(() => {
  return [...filtered.value].sort((a, b) => {
    let va, vb
    
    if (sortBy.value === 'name') {
      va = a.name
      vb = b.name
      return sortAsc.value ? va.localeCompare(vb) : vb.localeCompare(va)
    } else if (sortBy.value === 'files') {
      va = a.files
      vb = b.files
    } else {
      va = a.size_gb
      vb = b.size_gb
    }
    
    return sortAsc.value ? va - vb : vb - va
  })
})

const totalPages = computed(() => Math.max(1, Math.ceil(sorted.value.length / pageSize)))

const paginated = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return sorted.value.slice(start, start + pageSize)
})

watch(totalPages, (newVal) => {
  if (currentPage.value > newVal) {
    currentPage.value = Math.max(1, newVal)
  }
})

function toggleSort(col) {
  if (sortBy.value === col) {
    sortAsc.value = !sortAsc.value
  } else {
    sortBy.value = col
    sortAsc.value = col === 'name'
  }
}

function prevPage() {
  currentPage.value = Math.max(1, currentPage.value - 1)
}

function nextPage() {
  currentPage.value = Math.min(totalPages.value, currentPage.value + 1)
}
</script>

<template>
  <div class="card">
    <div class="card-head head-grid" style="grid-template-columns: auto 1fr auto;">
      <h2 style="margin:0; font-size:14px;">主播 · {{ filtered.length }}</h2>
      
      <div class="search" style="min-width: 0;">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <circle cx="11" cy="11" r="8"/>
          <path d="M21 21l-4.35-4.35"/>
        </svg>
        <input type="text" v-model="searchQuery" placeholder="搜索主播...">
      </div>
      
      <button class="btn btn-ghost auto-w" @click="emit('refresh')" :disabled="loading">
        刷新
      </button>
    </div>
    
    <div class="table-container">
      <table>
        <thead>
          <tr>
            <th style="width:35%" class="sortable" :class="{asc: sortBy==='name' && sortAsc, desc: sortBy==='name' && !sortAsc}" @click="toggleSort('name')">
              主播
            </th>
            <th style="width:35%" class="sortable" :class="{asc: sortBy==='size' && sortAsc, desc: sortBy==='size' && !sortAsc}" @click="toggleSort('size')">
              占用
            </th>
            <th style="width:12%;text-align:center" class="sortable" :class="{asc: sortBy==='files' && sortAsc, desc: sortBy==='files' && !sortAsc}" @click="toggleSort('files')">
              文件
            </th>
            <th style="width:18%;text-align:right">
              操作
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in paginated" :key="s.name">
            <td style="font-weight:500;max-width:120px;" :title="s.name">
              <span class="task-dot" style="display:inline-block;margin-right:6px" :class="s.is_running ? 'on' : 'off'"></span>
              {{ s.name }}
            </td>
            <td>
              <div style="display:flex;align-items:center;gap:12px">
                <span style="font-family:var(--font-mono);font-size:12px;min-width:56px;color:var(--text2)">
                  {{ s.size_gb.toFixed(1) }} GB
                </span>
                <div class="bar">
                  <div class="bar-fill" :style="{width: Math.min(100, s.size_gb / totalGB * 100) + '%'}"></div>
                </div>
              </div>
            </td>
            <td style="color:var(--muted);font-family:var(--font-mono);text-align:center">
              {{ s.files }}
            </td>
            <td style="text-align:right">
              <div style="display:flex;justify-content:flex-end;gap:6px">
                <button class="btn btn-ghost btn-sm auto-w" @click="emit('run', 'merge', s.name)" :disabled="running">
                  合并
                </button>
                <button class="btn btn-ghost btn-sm auto-w" @click="emit('run', 'clean', s.name)" :disabled="running">
                  清理
                </button>
                <button class="btn btn-ghost btn-sm auto-w" style="color:var(--info)" @click="emit('openManualMerge', s.name)" :disabled="running">
                  手动
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="paginated.length === 0">
            <td colspan="4" style="text-align:center;color:var(--faint);height:200px;vertical-align:middle">
              <div style="font-size:13px">未找到匹配的主播数据</div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    
    <div v-if="totalPages > 1" class="card-foot" style="display:flex;justify-content:center;align-items:center;gap:16px;">
      <button class="btn btn-ghost btn-sm auto-w" @click="prevPage" :disabled="currentPage <= 1">
        上一页
      </button>
      <span style="font-size:12px;color:var(--muted);font-family:var(--font-mono);">
        {{ currentPage }} / {{ totalPages }}
      </span>
      <button class="btn btn-ghost btn-sm auto-w" @click="nextPage" :disabled="currentPage >= totalPages">
        下一页
      </button>
    </div>
  </div>
</template>

<style scoped>
.card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--card-shadow);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 600px;
}

.card-head {
  height: var(--row-h);
  padding: 0 16px;
  border-bottom: 1px solid var(--border);
  background: color-mix(in srgb, var(--bg-sub) 50%, transparent);
  display: flex;
  align-items: center;
  flex-shrink: 0;
  gap: 16px;
}

.head-grid {
  display: grid;
  gap: 16px;
}

.search {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
  min-width: 0;
}

.search svg {
  position: absolute;
  left: 10px;
  width: 14px;
  height: 14px;
  color: var(--muted);
  pointer-events: none;
}

.search input {
  padding-left: 30px;
  width: 100%;
}

.table-container {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  scrollbar-gutter: stable;
}

th.sortable {
  cursor: pointer;
  user-select: none;
}

th.sortable:hover {
  color: var(--text);
}

th.sortable::after {
  content: '\2195';
  margin-left: 4px;
  opacity: 0.3;
}

th.asc::after {
  content: '\2191';
  opacity: 1;
  color: var(--text);
}

th.desc::after {
  content: '\2193';
  opacity: 1;
  color: var(--text);
}

.bar {
  height: 4px;
  border-radius: 2px;
  background: var(--border-sub);
  overflow: hidden;
  flex: 1;
}

.bar-fill {
  height: 100%;
  border-radius: 2px;
  background: var(--pri);
  transition: width 0.6s ease;
}

@media (max-width: 768px) {
  .card {
    min-height: auto;
  }
  
  .card-head {
    padding: 0 12px;
    height: auto;
    min-height: 40px;
    flex-wrap: wrap;
    gap: 8px;
    padding-top: 8px;
    padding-bottom: 8px;
  }
  
  .table-container {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }
  
  table {
    min-width: 480px;
  }
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/StreamerList.vue
git commit -m "feat: create StreamerList component"
```

---

## Task 11: Create TaskHistory Component

**Files:**
- Create: `frontend/src/components/TaskHistory.vue`

- [ ] **Step 1: Create TaskHistory.vue**

Create `frontend/src/components/TaskHistory.vue`:
```vue
<script setup>
import { ref, watch } from 'vue'
import { useApi } from '../composables/useApi'

const { get } = useApi()

const items = ref([])
const total = ref(0)
const pages = ref(0)
const currentPage = ref(1)
const taskFilter = ref('')
const loading = ref(false)

const perPage = 10

async function fetchHistory() {
  loading.value = true
  
  try {
    const params = new URLSearchParams({
      page: currentPage.value,
      per_page: perPage
    })
    
    if (taskFilter.value) {
      params.set('task', taskFilter.value)
    }
    
    const data = await get(`/api/history?${params.toString()}`)
    
    items.value = data.items || []
    total.value = data.total || 0
    pages.value = data.pages || 0
  } catch (err) {
    console.error('Failed to fetch history:', err)
  } finally {
    loading.value = false
  }
}

function formatHistoryTime(ts) {
  if (!ts) return '--'
  
  const date = new Date(ts.replace(' ', 'T'))
  if (isNaN(date.getTime())) return ts.substring(5, 16)
  
  const now = new Date()
  const h = date.getHours().toString().padStart(2, '0')
  const m = date.getMinutes().toString().padStart(2, '0')
  const t = `${h}:${m}`
  
  if (date.toDateString() === now.toDateString()) {
    return t
  }
  
  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)
  
  if (date.toDateString() === yesterday.toDateString()) {
    return `<span style="color:var(--muted)">昨天</span> ${t}`
  }
  
  return `${(date.getMonth() + 1).toString().padStart(2, '0')}-${date.getDate().toString().padStart(2, '0')} ${t}`
}

function prevPage() {
  if (currentPage.value > 1) {
    currentPage.value--
    fetchHistory()
  }
}

function nextPage() {
  if (currentPage.value < pages.value) {
    currentPage.value++
    fetchHistory()
  }
}

watch(taskFilter, () => {
  currentPage.value = 1
  fetchHistory()
})

// Initial fetch
fetchHistory()
</script>

<template>
  <div style="flex:1;display:flex;flex-direction:column;overflow:hidden">
    <div class="table-container">
      <table>
        <thead>
          <tr>
            <th style="width:25%">时间</th>
            <th style="width:55%">详情</th>
            <th style="width:20%;text-align:center">状态</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in items" :key="r.id">
            <td style="font-family:var(--font-mono);font-size:12px" v-html="formatHistoryTime(r.time)"></td>
            <td style="font-size:12px">{{ r.detail || r.streamer }}</td>
            <td style="text-align:center">
              <span :style="{color: r.status === 'success' ? 'var(--ok)' : 'var(--err)', fontWeight: 'bold'}">
                {{ r.status === 'success' ? '✓' : '✗' }}
              </span>
            </td>
          </tr>
          <tr v-if="items.length === 0">
            <td colspan="3" style="text-align:center;color:var(--faint);height:200px;vertical-align:middle">
              <div style="font-size:12px">暂无操作记录</div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    
    <div v-if="pages > 1" class="card-foot" style="display:flex;justify-content:center;gap:16px">
      <button class="btn btn-ghost btn-sm auto-w" @click="prevPage" :disabled="currentPage <= 1">
        上一页
      </button>
      <span style="font-size:12px;color:var(--muted);font-family:var(--font-mono);line-height:26px">
        {{ currentPage }} / {{ pages }}
      </span>
      <button class="btn btn-ghost btn-sm auto-w" @click="nextPage" :disabled="currentPage >= pages">
        下一页
      </button>
    </div>
  </div>
</template>

<style scoped>
.table-container {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  scrollbar-gutter: stable;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/TaskHistory.vue
git commit -m "feat: create TaskHistory component"
```

---

## Task 12: Create LogViewer Component

**Files:**
- Create: `frontend/src/components/LogViewer.vue`

- [ ] **Step 1: Create LogViewer.vue**

Create `frontend/src/components/LogViewer.vue`:
```vue
<script setup>
import { ref, watch } from 'vue'
import { useApi } from '../composables/useApi'

const { get } = useApi()

const logType = ref('merge')
const logFiles = ref([])
const selectedFile = ref('')
const content = ref('')
const contentHtml = ref('')
const autoScroll = ref(true)

async function loadLogFiles() {
  try {
    const data = await get(`/api/logs/list/${logType.value}`)
    logFiles.value = data || []
    
    if (data && data.length > 0) {
      selectedFile.value = data[0].filename
      await showLog()
    } else {
      content.value = '无日志数据'
      contentHtml.value = highlightLog('无日志数据')
    }
  } catch (err) {
    console.error('Failed to load log files:', err)
  }
}

async function showLog() {
  contentHtml.value = "<span style='color:var(--muted)'>请求中...</span>"
  
  try {
    const response = await fetch(`/api/logs/content/${logType.value}?file=${selectedFile.value}`)
    const text = await response.text()
    content.value = text || '(空)'
    contentHtml.value = highlightLog(content.value)
    
    if (autoScroll.value) {
      setTimeout(() => {
        const histLog = document.getElementById('histLog')
        if (histLog) histLog.scrollTop = histLog.scrollHeight
      }, 50)
    }
  } catch (err) {
    contentHtml.value = "<span style='color:var(--err)'>读取失败</span>"
  }
}

function escapeHtml(text) {
  return (text || '').toString()
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
}

function highlightLog(text) {
  return escapeHtml(text)
    .replace(/ERROR/gi, '<span class="te">ERROR</span>')
    .replace(/WARN/gi, '<span class="tw">WARN</span>')
    .replace(/✅|SUCCESS/gi, '<span class="tok">$&</span>')
    .replace(/ℹ|信息/g, '<span class="ti">$&</span>')
    .replace(/❌/g, '<span class="te">❌</span>')
}

watch(logType, () => {
  loadLogFiles()
})

// Initial load
loadLogFiles()
</script>

<template>
  <div class="term-wrapper">
    <div class="term" id="histLog" v-html="contentHtml || content || '选择日志文件'"></div>
  </div>
</template>

<style scoped>
.term-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 16px;
  min-width: 0;
}

.term {
  font-family: var(--font-mono);
  background: transparent;
  color: inherit;
  padding: 0;
  border: none;
  overflow-y: auto;
  overflow-x: hidden;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  word-break: break-all;
  font-size: 13px;
  line-height: 1.6;
  margin: 0;
  box-shadow: none;
  flex: 1;
  min-width: 0;
  width: 100%;
}

:deep(.te) { color: #ff7b72; }
:deep(.tw) { color: #d2a8ff; }
:deep(.tok) { color: #7ee787; }
:deep(.ti) { color: #79c0ff; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/LogViewer.vue
git commit -m "feat: create LogViewer component"
```

---

## Task 13: Create TaskScheduler Component

**Files:**
- Create: `frontend/src/components/TaskScheduler.vue`

- [ ] **Step 1: Create TaskScheduler.vue**

Create `frontend/src/components/TaskScheduler.vue`:
```vue
<script setup>
defineProps({
  schedule: {
    type: Object,
    required: true
  },
  config: {
    type: Object,
    required: true
  }
})

function clampInput(event, max) {
  const value = parseFloat(event.target.value)
  if (!isNaN(value) && value > max) {
    event.target.value = max
    event.target.dispatchEvent(new Event('input'))
  }
}
</script>

<template>
  <div class="setting-grid">
    <div class="setting-box" :class="{'is-active': schedule.merge_enabled, 'is-disabled': !schedule.merge_enabled}">
      <label style="display:flex;align-items:center;flex:1;cursor:pointer;margin:0">
        <div class="switch" style="margin-right:16px">
          <input type="checkbox" v-model="schedule.merge_enabled">
          <span class="switch-slider"></span>
        </div>
        <div class="setting-info">
          <div class="setting-title">自动合并任务</div>
          <div class="setting-desc">后台定期执行录像碎片拼接与转码</div>
        </div>
      </label>
      <div class="setting-control">
        <input type="number" v-model.number="schedule.merge_interval" :disabled="!schedule.merge_enabled" min="10" max="1440" @input="clampInput($event, 1440)">
        <span class="setting-unit">MIN</span>
      </div>
    </div>
    
    <div class="setting-box" :class="{'is-active': schedule.clean_enabled, 'is-disabled': !schedule.clean_enabled}">
      <label style="display:flex;align-items:center;flex:1;cursor:pointer;margin:0">
        <div class="switch" style="margin-right:16px">
          <input type="checkbox" v-model="schedule.clean_enabled">
          <span class="switch-slider"></span>
        </div>
        <div class="setting-info">
          <div class="setting-title">自动清理任务</div>
          <div class="setting-desc">触及空间警戒阈值时自动释放容量</div>
        </div>
      </label>
      <div class="setting-control">
        <input type="number" v-model.number="schedule.clean_interval" :disabled="!schedule.clean_enabled" min="10" max="1440" @input="clampInput($event, 1440)">
        <span class="setting-unit">MIN</span>
      </div>
    </div>
    
    <div class="setting-box span-full" style="cursor:default; flex-wrap:wrap; gap:24px; justify-content:space-between;">
      <div class="setting-info" style="margin:0;flex:1;min-width:280px">
        <div class="setting-title">系统静默时段</div>
        <div class="setting-desc">此时段内自动任务强制休眠，防止高峰期 IO 读写拥堵</div>
      </div>
      <div class="setting-control" style="gap:12px;background:var(--card);padding:6px 16px;border-radius:8px;border:1px solid var(--border);box-shadow:0 1px 2px rgba(0,0,0,.02)">
        <div class="time-picker">
          <input type="number" v-model.number="config.BACKUP_START_HOUR" min="0" max="23" placeholder="04">
          <span>:</span>
          <input type="number" v-model.number="config.BACKUP_START_MINUTE" min="0" max="59" placeholder="00">
        </div>
        <span style="font-size:12px;font-weight:600;color:var(--muted);padding:0 4px">至</span>
        <div class="time-picker">
          <input type="number" v-model.number="config.BACKUP_END_HOUR" min="0" max="23" placeholder="12">
          <span>:</span>
          <input type="number" v-model.number="config.BACKUP_END_MINUTE" min="0" max="59" placeholder="00">
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.setting-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.setting-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  background: var(--bg-sub);
  border: 1px solid var(--border);
  border-radius: 8px;
  transition: all 0.25s ease;
  cursor: pointer;
}

.setting-box:hover:not(.is-disabled) {
  border-color: var(--muted);
}

.setting-box:focus-within {
  border-color: var(--pri);
  background: var(--card);
  box-shadow: 0 0 0 3px var(--pri-r);
}

.setting-box.is-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.setting-box.is-active {
  border-color: var(--pri);
  box-shadow: 0 0 0 1px var(--pri);
}

.setting-info {
  flex: 1;
  margin-right: 16px;
}

.setting-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
}

.setting-desc {
  font-size: 12px;
  color: var(--muted);
  margin-top: 4px;
  line-height: 1.5;
}

.setting-control {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.setting-unit {
  font-size: 12px;
  font-weight: 600;
  color: var(--muted);
  width: 28px;
  text-align: left;
  display: inline-block;
}

.setting-control input[type="number"] {
  width: 60px;
  padding: 0 8px 0 0;
  text-align: right;
  font-family: var(--font-mono);
  font-size: 13px;
  font-weight: 500;
}

.span-full {
  grid-column: 1 / -1;
}

.time-picker {
  display: inline-flex;
  align-items: center;
  background: var(--bg-sub);
  border: 1px solid var(--border);
  border-radius: 6px;
  transition: all 0.2s;
  overflow: hidden;
}

.time-picker:focus-within {
  border-color: var(--pri);
  box-shadow: 0 0 0 3px var(--pri-r);
  background: var(--card);
}

.time-picker input {
  width: 40px;
  border: none;
  background: transparent;
  text-align: center;
  box-shadow: none !important;
  font-size: 13px;
  border-radius: 0;
  -moz-appearance: textfield;
}

.time-picker input::-webkit-outer-spin-button,
.time-picker input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/TaskScheduler.vue
git commit -m "feat: create TaskScheduler component"
```

---

## Task 14: Create ConfigPanel Component

**Files:**
- Create: `frontend/src/components/ConfigPanel.vue`

- [ ] **Step 1: Create ConfigPanel.vue**

Create `frontend/src/components/ConfigPanel.vue`:
```vue
<script setup>
import { ref, computed, watch } from 'vue'
import { useToast } from '../composables/useToast'
import TaskScheduler from './TaskScheduler.vue'

const props = defineProps({
  config: {
    type: Object,
    required: true
  },
  schedule: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['save', 'saveSchedule', 'resetSchedule', 'recommend'])

const { toast } = useToast()

const activeHelp = ref('')

const whitelistStr = computed({
  get() {
    const wl = props.config.WHITELIST_KEYWORDS || []
    return Array.isArray(wl) ? wl.join(',') : wl
  },
  set(val) {
    props.config.WHITELIST_KEYWORDS = val
      .split(/[,，]/)
      .map(s => s.trim())
      .filter(s => s)
  }
})

const safeMinutes = computed({
  get() {
    return props.config.SAFE_AGE_MINUTES != null ? props.config.SAFE_AGE_MINUTES : 120
  },
  set(v) {
    v = parseFloat(v)
    if (isNaN(v)) v = 120
    v = Math.max(1, Math.min(720, Math.round(v)))
    props.config.SAFE_AGE_MINUTES = v
    props.config.SAFE_MODE = 'hours'
    props.config.SAFE_DAYS = Math.max(1, Math.round(v / 60 / 24))
  }
})

// Watch and validate numeric fields
watch(() => props.config.BACKUP_START_HOUR, (v) => {
  if (typeof v !== 'number' || isNaN(v)) {
    props.config.BACKUP_START_HOUR = 0
  } else if (v < 0) {
    props.config.BACKUP_START_HOUR = 0
  } else if (v > 23) {
    props.config.BACKUP_START_HOUR = 23
  }
})

watch(() => props.config.BACKUP_END_HOUR, (v) => {
  if (typeof v !== 'number' || isNaN(v)) {
    props.config.BACKUP_END_HOUR = 0
  } else if (v < 0) {
    props.config.BACKUP_END_HOUR = 0
  } else if (v > 23) {
    props.config.BACKUP_END_HOUR = 23
  }
})

watch(() => props.config.BACKUP_START_MINUTE, (v) => {
  if (typeof v !== 'number' || isNaN(v)) {
    props.config.BACKUP_START_MINUTE = 0
  } else if (v < 0) {
    props.config.BACKUP_START_MINUTE = 0
  } else if (v > 59) {
    props.config.BACKUP_START_MINUTE = 59
  }
})

watch(() => props.config.BACKUP_END_MINUTE, (v) => {
  if (typeof v !== 'number' || isNaN(v)) {
    props.config.BACKUP_END_MINUTE = 0
  } else if (v < 0) {
    props.config.BACKUP_END_MINUTE = 0
  } else if (v > 59) {
    props.config.BACKUP_END_MINUTE = 59
  }
})

watch(() => props.config.TRIGGER_THRESHOLD, (v) => {
  if (typeof v === 'number' && !isNaN(v)) {
    props.config.TRIGGER_THRESHOLD = Math.max(0, Math.min(100, v))
  }
})

watch(() => props.config.TARGET_THRESHOLD, (v) => {
  if (typeof v === 'number' && !isNaN(v)) {
    props.config.TARGET_THRESHOLD = Math.max(0, Math.min(100, v))
  }
})

function validate() {
  if (props.config.TRIGGER_THRESHOLD < 0 || props.config.TRIGGER_THRESHOLD > 100) {
    toast('阈值必须在 0-100 之间', 'err')
    return false
  }
  
  if (props.config.TARGET_THRESHOLD < 0 || props.config.TARGET_THRESHOLD > 100) {
    toast('阈值必须在 0-100 之间', 'err')
    return false
  }
  
  return true
}

function handleSave() {
  if (validate()) {
    emit('save')
  }
}
</script>

<template>
  <div class="card" style="margin-bottom:24px;">
    <div class="card-head" style="display:flex; justify-content:space-between;">
      <h2 style="margin:0;">基础策略</h2>
      <div style="display:flex;gap:12px">
        <button class="btn btn-ghost auto-w" @click="activeHelp = activeHelp === 'config' ? '' : 'config'">
          {{ activeHelp === 'config' ? '收起说明' : '帮助说明' }}
        </button>
        <button class="btn btn-ghost auto-w" @click="emit('recommend')">智能推荐</button>
        <button class="btn btn-pri auto-w" @click="handleSave">保存配置</button>
      </div>
    </div>
    
    <div class="help-wrap" :class="{open: activeHelp === 'config'}">
      <div class="help">
        <dl>
          <dt>空间警戒阈值</dt><dd>自动清理的触发红线。建议设为 80% - 90%。</dd>
          <dt>安全回落水位</dt><dd>自动清理的目标终点。建议与警戒阈值保持 10% - 20% 差值。</dd>
          <dt>保底留存件数</dt><dd>无论空间多紧缺，每人强制保留的最新录像数。</dd>
          <dt>新文件保护期</dt><dd>刚结束录制的视频在此时间内绝对免删，防止误清。</dd>
          <dt>断流分割判定</dt><dd>断流超过此时长，合并时将切分为新场次。</dd>
          <dt>合并延迟缓冲</dt><dd>录制结束后推迟合并的安全等待时间，防止文件损坏。</dd>
          <dt>单次清理限额</dt><dd>每轮清理最多允许删除的文件数，防止失控误删。</dd>
        </dl>
      </div>
    </div>
    
    <div class="card-body">
      <div class="setting-grid" style="grid-template-columns: 1fr 1fr;">
        <label class="setting-box" style="align-items:flex-start;flex-direction:column;gap:12px;cursor:text">
          <div class="setting-info" style="margin:0;width:100%">
            <div class="setting-title">录像存储目录</div>
            <div class="setting-desc">扫描与清理的基础路径</div>
          </div>
          <input type="text" v-model="config.TARGET_DIR" style="font-family:var(--font-mono);width:100%;background:var(--card);box-shadow:inset 0 1px 2px rgba(0,0,0,.02)">
        </label>
        
        <label class="setting-box" style="align-items:flex-start;flex-direction:column;gap:12px;cursor:text">
          <div class="setting-info" style="margin:0;width:100%">
            <div class="setting-title">免删保护名单</div>
            <div class="setting-desc">含此词缀录像绝对免删（逗号分隔）</div>
          </div>
          <input type="text" v-model="whitelistStr" style="width:100%;background:var(--card);box-shadow:inset 0 1px 2px rgba(0,0,0,.02)" placeholder="留存,高能,勿删">
        </label>
      </div>
      
      <div class="setting-grid" style="grid-template-columns: 1fr 1fr;">
        <label class="setting-box">
          <div class="setting-info">
            <div class="setting-title">空间警戒阈值</div>
            <div class="setting-desc">触发自动清理的占用红线</div>
          </div>
          <div class="setting-control">
            <input type="number" v-model.number="config.TRIGGER_THRESHOLD" min="0" max="100">
            <span class="setting-unit">%</span>
          </div>
        </label>
        
        <label class="setting-box">
          <div class="setting-info">
            <div class="setting-title">安全回落水位</div>
            <div class="setting-desc">清理后的空间安全底线</div>
          </div>
          <div class="setting-control">
            <input type="number" v-model.number="config.TARGET_THRESHOLD" min="0" max="100">
            <span class="setting-unit">%</span>
          </div>
        </label>
      </div>
      
      <div class="setting-grid" style="grid-template-columns: 1fr 1fr;">
        <label class="setting-box">
          <div class="setting-info">
            <div class="setting-title">保底留存件数</div>
            <div class="setting-desc">主播强制保留的最新件数</div>
          </div>
          <div class="setting-control">
            <input type="number" v-model.number="config.MIN_KEEP_PER_STREAMER" min="1" max="50">
            <span class="setting-unit">个</span>
          </div>
        </label>
        
        <label class="setting-box">
          <div class="setting-info">
            <div class="setting-title">单次清理限额</div>
            <div class="setting-desc">每轮清理最多删除的文件数</div>
          </div>
          <div class="setting-control">
            <input type="number" v-model.number="config.MAX_DELETE_PER_RUN" min="1" max="100">
            <span class="setting-unit">个</span>
          </div>
        </label>
      </div>
      
      <div class="setting-grid" style="grid-template-columns: 1fr 1fr 1fr;">
        <label class="setting-box">
          <div class="setting-info">
            <div class="setting-title">新文件保护期</div>
            <div class="setting-desc">录制完的视频绝对免删时长</div>
          </div>
          <div class="setting-control">
            <input type="number" v-model.number="safeMinutes" min="1" max="720">
            <span class="setting-unit">MIN</span>
          </div>
        </label>
        
        <label class="setting-box">
          <div class="setting-info">
            <div class="setting-title">断流分割判定</div>
            <div class="setting-desc">断流超时即切分为新场次</div>
          </div>
          <div class="setting-control">
            <input type="number" v-model.number="config.GAP_MINUTES" min="1" max="1440">
            <span class="setting-unit">MIN</span>
          </div>
        </label>
        
        <label class="setting-box">
          <div class="setting-info">
            <div class="setting-title">合并延迟缓冲</div>
            <div class="setting-desc">录像后延迟合并的等待时间</div>
          </div>
          <div class="setting-control">
            <input type="number" v-model.number="config.MERGE_AGE_MINUTES" min="0" max="1440">
            <span class="setting-unit">MIN</span>
          </div>
        </label>
      </div>
      
      <!-- Task Scheduler -->
      <TaskScheduler :schedule="schedule" :config="config" />
    </div>
  </div>
</template>

<style scoped>
.card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--card-shadow);
  overflow: hidden;
}

.card-head {
  height: var(--row-h);
  padding: 0 16px;
  border-bottom: 1px solid var(--border);
  background: color-mix(in srgb, var(--bg-sub) 50%, transparent);
  display: flex;
  align-items: center;
  flex-shrink: 0;
  justify-content: space-between;
}

.card-body {
  padding: 20px;
  overflow-y: auto;
}

.setting-grid {
  display: grid;
  gap: 16px;
  margin-bottom: 16px;
}

.setting-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  background: var(--bg-sub);
  border: 1px solid var(--border);
  border-radius: 8px;
  transition: all 0.25s ease;
  cursor: pointer;
}

.setting-box:hover {
  border-color: var(--muted);
}

.setting-box:focus-within {
  border-color: var(--pri);
  background: var(--card);
  box-shadow: 0 0 0 3px var(--pri-r);
}

.setting-info {
  flex: 1;
  margin-right: 16px;
}

.setting-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
}

.setting-desc {
  font-size: 12px;
  color: var(--muted);
  margin-top: 4px;
  line-height: 1.5;
}

.setting-control {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.setting-unit {
  font-size: 12px;
  font-weight: 600;
  color: var(--muted);
  width: 28px;
  text-align: left;
  display: inline-block;
}

.help-wrap {
  overflow: hidden;
  max-height: 0;
  transition: max-height 0.4s ease;
}

.help-wrap.open {
  max-height: 600px;
}

.help {
  background: var(--bg-sub);
  border-left: 3px solid var(--muted);
  border-radius: 0 8px 8px 0;
  padding: 16px 20px;
  margin: 0 20px 20px;
  font-size: 12px;
  line-height: 1.6;
}

.help dt {
  font-weight: 600;
  color: var(--text);
  display: inline-block;
  margin-right: 6px;
}

.help dt::after {
  content: ":";
  color: var(--muted);
}

.help dd {
  display: inline;
  color: var(--text2);
}

.help dd::after {
  content: "";
  display: block;
  margin-bottom: 8px;
}

@media (max-width: 768px) {
  .setting-grid {
    grid-template-columns: 1fr !important;
    gap: 10px;
  }
  
  .setting-box {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
    padding: 12px;
  }
  
  .setting-info {
    margin-right: 0;
  }
  
  .setting-control {
    justify-content: flex-end;
  }
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ConfigPanel.vue
git commit -m "feat: create ConfigPanel component"
```

---

## Task 15: Create ManualMerge Component

**Files:**
- Create: `frontend/src/components/ManualMerge.vue`

- [ ] **Step 1: Create ManualMerge.vue**

Create `frontend/src/components/ManualMerge.vue`:
```vue
<script setup>
import { ref, computed, watch } from 'vue'
import { useApi } from '../composables/useApi'
import { useToast } from '../composables/useToast'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  streamer: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['close', 'merge'])

const { get } = useApi()
const { toast } = useToast()

const files = ref([])
const selected = ref([])
const loading = ref(false)

const allChecked = computed(() => {
  return files.value.length > 0 && selected.value.length === files.value.length
})

const selectedSize = computed(() => {
  return files.value
    .filter(f => selected.value.includes(f.name))
    .reduce((sum, f) => sum + f.size, 0)
})

async function fetchFiles() {
  if (!props.streamer) return
  
  loading.value = true
  files.value = []
  selected.value = []
  
  try {
    const data = await get(`/api/files/${encodeURIComponent(props.streamer)}`)
    files.value = data || []
  } catch (err) {
    if (err.message !== '未登录') {
      toast(err.message || '拉取失败', 'err')
      emit('close')
    }
  } finally {
    loading.value = false
  }
}

function toggleFile(name) {
  const idx = selected.value.indexOf(name)
  if (idx === -1) {
    selected.value.push(name)
  } else {
    selected.value.splice(idx, 1)
  }
}

function toggleAll() {
  if (allChecked.value) {
    selected.value = []
  } else {
    selected.value = files.value.map(f => f.name)
  }
}

function formatSize(bytes) {
  if (bytes >= 1 << 30) return (bytes / (1 << 30)).toFixed(2) + ' GB'
  if (bytes >= 1 << 20) return (bytes / (1 << 20)).toFixed(2) + ' MB'
  return (bytes / 1024).toFixed(2) + ' KB'
}

function handleMerge() {
  if (selected.value.length < 2) return
  emit('merge', [...selected.value])
}

watch(() => props.visible, (newVal) => {
  if (newVal) {
    fetchFiles()
  }
})
</script>

<template>
  <div v-if="visible" class="overlay" @click.self="emit('close')">
    <div class="modal large">
      <div class="card-head" style="display:flex; justify-content:space-between;">
        <div>
          <h2 style="margin:0;">手动合并 — {{ streamer }}</h2>
        </div>
        <button class="modal-x" @click="emit('close')">&times;</button>
      </div>
      
      <div style="padding:0;overflow-y:auto;flex:1;max-height:50vh;">
        <div v-if="loading" style="text-align:center;padding:40px;color:var(--muted)">
          加载中...
        </div>
        <div v-else-if="files.length === 0" style="text-align:center;padding:40px;color:var(--muted)">
          暂无视频文件
        </div>
        <table v-else>
          <thead>
            <tr>
              <th style="width:40px;text-align:center">
                <input type="checkbox" @change="toggleAll" :checked="allChecked">
              </th>
              <th>文件名</th>
              <th style="width:80px;text-align:right">大小</th>
              <th style="width:60px;text-align:center">类型</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="f in files" :key="f.name" @click="toggleFile(f.name)" style="cursor:pointer" :style="{background: selected.includes(f.name) ? 'var(--hover)' : ''}">
              <td style="text-align:center">
                <input type="checkbox" :checked="selected.includes(f.name)" @click.stop="toggleFile(f.name)">
              </td>
              <td style="font-family:var(--font-mono);font-size:12px" v-text="f.name"></td>
              <td style="text-align:right;color:var(--text2);font-family:var(--font-mono);font-size:12px" v-text="f.size_str"></td>
              <td style="text-align:center">
                <span class="badge" :class="f.is_merged ? 'badge-ok' : 'badge-warn'" v-text="f.is_merged ? '合并' : '原片'"></span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      
      <div class="card-foot" style="display:flex;justify-content:space-between;align-items:center">
        <div style="font-size:13px;color:var(--muted)">
          已选 {{ selected.length }} 个
          <span v-if="selected.length > 0"> ({{ formatSize(selectedSize) }})</span>
        </div>
        <div style="display:flex;gap:12px">
          <button class="btn btn-ghost auto-w" @click="selected = []" :disabled="selected.length === 0">
            清空
          </button>
          <button class="btn btn-pri auto-w" @click="handleMerge" :disabled="selected.length < 2">
            开始合并
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  animation: fade-in 0.2s ease-out;
}

.modal {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 40px -10px rgba(0,0,0,0.2);
  animation: pop-up 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal.large {
  max-width: 800px;
}

.modal-x {
  background: transparent;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: var(--muted);
  line-height: 1;
  padding: 4px;
  border-radius: 6px;
  transition: background 0.2s;
}

.modal-x:hover {
  background: var(--hover);
  color: var(--text);
}

@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes pop-up {
  0% { opacity: 0; transform: scale(0.96) translateY(10px); }
  100% { opacity: 1; transform: scale(1) translateY(0); }
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ManualMerge.vue
git commit -m "feat: create ManualMerge component"
```

---

## Task 16: Integrate Everything in App.vue

**Files:**
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Update App.vue**

Replace `frontend/src/App.vue` with:
```vue
<script setup>
import { ref, onMounted } from 'vue'
import TopNav from './components/TopNav.vue'
import StatusDashboard from './components/StatusDashboard.vue'
import StreamerList from './components/StreamerList.vue'
import TaskHistory from './components/TaskHistory.vue'
import LogViewer from './components/LogViewer.vue'
import ConfigPanel from './components/ConfigPanel.vue'
import ManualMerge from './components/ManualMerge.vue'
import { useStreamerData } from './composables/useStreamerData'
import { useConfig } from './composables/useConfig'
import { useTaskRunner } from './composables/useTaskRunner'
import { useToast } from './composables/useToast'

const {
  streamers,
  diskUsage,
  totalGB,
  detail,
  fetchStatus,
  fetchDetail
} = useStreamerData()

const {
  config,
  schedule,
  fetchConfig,
  fetchSchedule,
  saveConfig,
  saveSchedule,
  resetSchedule
} = useConfig()

const {
  running,
  progress,
  progressLabel,
  logs,
  autoScroll,
  run,
  startManualMerge
} = useTaskRunner()

const { toast } = useToast()

// Manual merge modal
const showMergeModal = ref(false)
const selectedStreamer = ref('')

// Confirm run modal
const showConfirmModal = ref(false)
const confirmTask = ref('')
const confirmTarget = ref('')
const confirmLabel = ref('')

// Log tab
const activeTab = ref('history')

function handleRun(task, streamer) {
  confirmTask.value = task
  confirmTarget.value = streamer || '全局'
  confirmLabel.value = task === 'merge' ? '合并' : '清理'
  showConfirmModal.value = true
}

function confirmRun() {
  showConfirmModal.value = false
  
  const onComplete = () => {
    fetchStatus()
    fetchDetail()
    fetchConfig()
  }
  
  run(confirmTask.value, confirmTarget.value, { onComplete })
}

function handleOpenManualMerge(streamer) {
  selectedStreamer.value = streamer
  showMergeModal.value = true
}

function handleMerge(files) {
  showMergeModal.value = false
  
  const onComplete = () => {
    fetchStatus()
    fetchDetail()
    fetchConfig()
  }
  
  startManualMerge(selectedStreamer.value, files, { onComplete })
}

async function handleSaveConfig() {
  await saveConfig()
  await saveSchedule()
}

async function handleRecommend() {
  toast('正在分析配置...', 'info')
  // TODO: Implement recommend
}

// Initialize
onMounted(async () => {
  await fetchConfig()
  await fetchStatus()
  await fetchSchedule()
  await fetchDetail()
  
  // Auto-refresh every minute
  setInterval(async () => {
    if (!running.value) {
      await fetchDetail()
    }
  }, 60000)
})
</script>

<template>
  <TopNav :running="running" />
  
  <StatusDashboard
    :disk-usage="diskUsage"
    :total-g-b="totalGB"
    :schedule="detail.schedule"
    :running="running"
    @run="handleRun"
  />
  
  <div class="main-grid">
    <StreamerList
      :streamers="streamers"
      :total-g-b="totalGB"
      :running="running"
      :loading="false"
      @run="handleRun"
      @open-manual-merge="handleOpenManualMerge"
      @refresh="fetchStatus"
    />
    
    <div class="card">
      <div class="card-head" style="gap: 12px;">
        <select v-model="activeTab" class="tab-select" style="cursor:pointer; width: auto; min-width: 90px;">
          <option value="history">操作历史</option>
          <option value="logs">系统日志</option>
        </select>
      </div>
      
      <div style="flex:1; display:flex; flex-direction:column; overflow:hidden">
        <TaskHistory v-show="activeTab === 'history'" />
        <LogViewer v-show="activeTab === 'logs'" />
      </div>
    </div>
  </div>
  
  <ConfigPanel
    :config="config"
    :schedule="schedule"
    @save="handleSaveConfig"
    @recommend="handleRecommend"
  />
  
  <ManualMerge
    :visible="showMergeModal"
    :streamer="selectedStreamer"
    @close="showMergeModal = false"
    @merge="handleMerge"
  />
  
  <!-- Confirm Run Modal -->
  <div v-if="showConfirmModal" class="overlay" @click.self="showConfirmModal = false">
    <div class="modal">
      <div class="card-head" style="display:flex; justify-content:space-between;">
        <h2 style="margin:0;">确认{{ confirmLabel }}</h2>
        <button class="modal-x" @click="showConfirmModal = false">&times;</button>
      </div>
      <div style="padding:32px 24px;text-align:center">
        <p style="font-size:15px;color:var(--text);font-weight:500;margin-bottom:8px">
          即将对 <strong>{{ confirmTarget }}</strong> 执行 <b>{{ confirmLabel }}</b>
        </p>
        <p style="font-size:13px;color:var(--muted)">此操作将立即开始，且无法中途撤销。</p>
      </div>
      <div class="card-foot" style="display:flex;gap:12px;justify-content:flex-end">
        <button class="btn btn-ghost auto-w" @click="showConfirmModal = false">取消</button>
        <button class="btn btn-pri auto-w" @click="confirmRun">确认执行</button>
      </div>
    </div>
  </div>
</template>

<style>
.main-grid {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(0, 1fr);
  gap: var(--gap);
  align-items: stretch;
  margin-bottom: var(--gap);
}

@media (max-width: 1024px) {
  .main-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

.card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--card-shadow);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.card-head {
  height: var(--row-h);
  padding: 0 16px;
  border-bottom: 1px solid var(--border);
  background: color-mix(in srgb, var(--bg-sub) 50%, transparent);
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.card-body {
  padding: 20px;
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.card-foot {
  padding: 12px 20px;
  background: var(--bg-sub);
  border-top: 1px solid var(--border);
}

.tab-select {
  border: none !important;
  background: transparent !important;
  font-weight: 600;
  font-size: 14px;
  padding-left: 0;
  box-shadow: none !important;
  color: var(--text) !important;
}

.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  animation: fade-in 0.2s ease-out;
}

.modal {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 40px -10px rgba(0,0,0,0.2);
  animation: pop-up 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-x {
  background: transparent;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: var(--muted);
  line-height: 1;
  padding: 4px;
  border-radius: 6px;
  transition: background 0.2s;
}

.modal-x:hover {
  background: var(--hover);
  color: var(--text);
}

@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes pop-up {
  0% { opacity: 0; transform: scale(0.96) translateY(10px); }
  100% { opacity: 1; transform: scale(1) translateY(0); }
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/App.vue
git commit -m "feat: integrate all components in App.vue"
```

---

## Task 17: Build and Deploy

**Files:**
- Modify: `frontend/vite.config.js` (if needed)

- [ ] **Step 1: Build for production**

Run: `cd frontend && npm run build`
Expected: Build output in `frontend/dist/` directory

- [ ] **Step 2: Copy to Go backend**

Run: `cp -r frontend/dist/* ../templates/frontend/`
Expected: Files copied to Go's static file directory

- [ ] **Step 3: Test in Docker**

Run: `docker restart bililive-helper-go`
Expected: Container restarts with new frontend

- [ ] **Step 4: Verify deployment**

Open browser and navigate to http://192.168.10.10:5689/
Expected: Application loads and functions correctly

- [ ] **Step 5: Commit build artifacts**

```bash
git add frontend/dist/
git commit -m "build: add production build"
```

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-05-30-frontend-refactoring-plan.md`.

**Two execution options:**

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?

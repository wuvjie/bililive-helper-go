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

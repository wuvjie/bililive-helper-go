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

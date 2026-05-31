<script setup>
import { ref } from 'vue'
import { useStreamerData } from '../composables/useStreamerData'
import { useApi } from '../composables/useApi'

const { streamers } = useStreamerData()
const api = useApi()

const isSubmitting = ref(false)
const form = ref({
  streamer: '',
  date: new Date().toISOString().split('T')[0]
})

async function submitManualMerge() {
  if (!form.value.streamer || !form.value.date) {
    alert('请选择主播并填写日期')
    return
  }
  isSubmitting.value = true
  try {
    await api.post('/merge/manual', form.value)
    alert('手动合并任务已触发入队！')
    form.value.streamer = ''
  } catch (e) {
    alert('触发失败: ' + e.message)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="p-6">
    <h1 class="text-[17px] font-semibold text-[#1f2329] mb-5">任务调度</h1>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-5">
      <div class="bg-white rounded-xl p-6 border border-[#dee0e3] shadow-sm">
        <h3 class="text-[16px] font-medium text-[#1f2329] mb-2">触发手动任务</h3>
        <p class="text-[13px] text-[#8f959e] mb-6">针对未成功合并的视频，可手动指定主播与日期触发合并队列。</p>

        <div class="space-y-4">
          <div>
            <label class="block text-[13px] font-medium text-[#1f2329] mb-2">选择主播 (Streamer)</label>
            <select v-model="form.streamer" class="w-full px-3 py-2 border border-[#dee0e3] rounded-lg text-[14px] outline-none focus:border-[#3370ff] focus:ring-1 focus:ring-[#3370ff] bg-white cursor-pointer">
              <option value="" disabled>请选择要合并的主播</option>
              <option v-for="s in streamers" :key="s.name" :value="s.name">
                {{ s.name }} ({{ s.platform || '未知平台' }})
              </option>
            </select>
          </div>

          <div>
            <label class="block text-[13px] font-medium text-[#1f2329] mb-2">目标日期 (Date)</label>
            <input v-model="form.date" type="date" class="w-full px-3 py-2 border border-[#dee0e3] rounded-lg text-[14px] outline-none focus:border-[#3370ff] focus:ring-1 focus:ring-[#3370ff]" />
          </div>

          <button @click="submitManualMerge" :disabled="isSubmitting || !form.streamer" class="w-full mt-2 py-2.5 bg-[#3370ff] hover:bg-[#5384ff] disabled:bg-[#a9c4ff] text-white text-[14px] font-medium rounded-lg transition-colors">
            {{ isSubmitting ? '正在提交...' : '立即执行合并' }}
          </button>
        </div>
      </div>

      <div class="bg-[#1e1e1e] rounded-xl p-4 border border-[#333] flex flex-col h-[400px]">
        <div class="flex items-center justify-between mb-4 pb-3 border-b border-[#333]">
          <div class="flex gap-2">
            <div class="w-3 h-3 rounded-full bg-[#ff5f56]"></div>
            <div class="w-3 h-3 rounded-full bg-[#ffbd2e]"></div>
            <div class="w-3 h-3 rounded-full bg-[#27c93f]"></div>
          </div>
          <span class="text-[#858585] text-xs font-mono">system_tasks.log</span>
        </div>
        <div class="flex-1 overflow-y-auto font-mono text-[13px] text-[#ccc] space-y-2">
          <div class="text-[#3b8eea]">[INFO] Task scheduler loaded.</div>
          <div class="text-[#5c6370]">[DEBUG] Waiting for manual trigger...</div>
          <div class="animate-pulse">_</div>
        </div>
      </div>
    </div>
  </div>
</template>

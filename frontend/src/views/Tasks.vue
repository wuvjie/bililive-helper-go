<template>
  <div class="feishu-card" style="max-width: 600px;">
    <h3 style="margin-top: 0; margin-bottom: 24px; font-size: 18px;">发起手动合并</h3>

    <div class="form-item">
      <label>主播标识 (Streamer)</label>
      <input v-model="form.streamer" type="text" class="feishu-input" placeholder="例如: douyin_xxx" />
    </div>

    <div class="form-item">
      <label>目标日期 (Date)</label>
      <input v-model="form.date" type="text" class="feishu-input" placeholder="例如: 2026-05-30" />
    </div>

    <div v-if="error" style="color: #f54a45; margin-bottom: 16px; font-size: 14px;">❌ {{ error }}</div>
    <div v-if="successMsg" style="color: #34a853; margin-bottom: 16px; font-size: 14px;">✅ {{ successMsg }}</div>

    <button class="feishu-btn" @click="submitMerge" :disabled="loading || !form.streamer || !form.date" style="width: 100%;">
      {{ loading ? '正在提交...' : '触发合并任务' }}
    </button>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useApi } from '../composables/useApi'

const { loading, error, triggerMerge } = useApi()
const successMsg = ref('')

const form = reactive({
  streamer: '',
  date: ''
})

const submitMerge = async () => {
  successMsg.value = ''
  try {
    await triggerMerge({ streamer: form.streamer, date: form.date })
    successMsg.value = '合并任务已成功触发入队！'
    form.streamer = ''
    form.date = ''
  } catch (err) {
    console.error(err)
  }
}
</script>

<style scoped>
.form-item { margin-bottom: 20px; }
.form-item label { display: block; margin-bottom: 8px; font-size: 14px; color: #1f2329; font-weight: 500; }
</style>

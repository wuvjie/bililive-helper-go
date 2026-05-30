<template>
  <div class="feishu-card" style="max-width: 800px;">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h3 style="margin: 0; font-size: 18px;">系统配置文件 (JSON)</h3>
      <div>
        <button class="feishu-btn feishu-btn-outline" @click="fetchConfig" :disabled="loading" style="margin-right: 12px;">重置</button>
        <button class="feishu-btn" @click="submitConfig" :disabled="loading">保存配置</button>
      </div>
    </div>

    <div v-if="error" style="color: #f54a45; margin-bottom: 16px; font-size: 14px;">❌ {{ error }}</div>
    <div v-if="successMsg" style="color: #34a853; margin-bottom: 16px; font-size: 14px;">✅ {{ successMsg }}</div>

    <p style="font-size: 13px; color: #8f959e; margin-bottom: 16px;">注意：请确保修改的内容符合标准 JSON 格式，否则会导致保存失败及系统异常。</p>

    <textarea
      v-model="configText"
      class="feishu-input json-editor"
      spellcheck="false"
      placeholder="加载配置中..."
    ></textarea>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { loading, error, getConfig, saveConfig } = useApi()
const configText = ref('')
const successMsg = ref('')

const fetchConfig = async () => {
  error.value = null
  successMsg.value = ''
  try {
    const res = await getConfig()
    const data = res.data || res
    configText.value = JSON.stringify(data, null, 2)
  } catch (err) {
    console.error('获取配置失败:', err)
  }
}

const submitConfig = async () => {
  error.value = null
  successMsg.value = ''
  try {
    const parsedData = JSON.parse(configText.value)
    await saveConfig(parsedData)
    successMsg.value = '配置已成功保存！可能需要重启服务生效。'
  } catch (err) {
    error.value = 'JSON 格式错误，请检查语法！'
    console.error(err)
  }
}

onMounted(fetchConfig)
</script>

<style scoped>
.json-editor {
  font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
  height: 500px;
  resize: vertical;
  background-color: #f8f9fa;
  font-size: 14px;
  line-height: 1.5;
  padding: 16px;
}
</style>

<template>
  <div class="settings-page feishu-card">
    <div class="settings-layout">
      <div class="settings-sidebar">
        <h3 class="nav-header">系统设置</h3>
        <ul class="settings-nav">
          <li :class="{ active: currentTab === 'general' }" @click="currentTab = 'general'">基础参数</li>
          <li :class="{ active: currentTab === 'storage' }" @click="currentTab = 'storage'">存储与清理规则</li>
          <li :class="{ active: currentTab === 'notify' }" @click="currentTab = 'notify'">消息通知</li>
          <li :class="{ active: currentTab === 'advanced' }" @click="currentTab = 'advanced'">高级 JSON 编辑</li>
        </ul>
      </div>

      <div class="settings-content">

        <div v-show="currentTab === 'general'" class="settings-section">
          <h2 class="section-title">基础参数</h2>
          <div class="form-row">
            <label>默认挂载路径</label>
            <input type="text" class="feishu-input" value="/vol2/1000/vedio/bililive-go/" />
            <div class="help-text">所有主播视频存档的根目录路径。</div>
          </div>
          <div class="form-row">
            <label>合并后输出格式</label>
            <select class="feishu-input" style="width: 200px;">
              <option value="mp4">.mp4 (H.264 / ACC)</option>
              <option value="flv">.flv (Raw Stream)</option>
              <option value="mkv">.mkv</option>
            </select>
          </div>
          <div class="form-row">
            <label class="checkbox-label">
              <input type="checkbox" checked />
              合并完成后自动删除原分段文件
            </label>
          </div>
        </div>

        <div v-show="currentTab === 'storage'" class="settings-section">
          <h2 class="section-title">存储与清理规则</h2>
          <div class="alert-box warning">
            <span class="icon">⚠️</span>
            当 NAS 挂载盘容量达到设定阈值时，系统将自动删除非白名单的最早视频文件。
          </div>

          <div class="form-row">
            <label>告警触发水位 (%)</label>
            <div class="slider-group">
              <input type="range" min="50" max="99" value="90" class="feishu-slider" />
              <span class="slider-value">90%</span>
            </div>
          </div>

          <div class="form-row">
            <label>清理目标安全水位 (%)</label>
            <div class="slider-group">
              <input type="range" min="40" max="89" value="80" class="feishu-slider" />
              <span class="slider-value">80%</span>
            </div>
            <div class="help-text">触发清理后，将持续删除文件直到容量回落至该水位。</div>
          </div>
        </div>

        <div v-show="currentTab === 'advanced'" class="settings-section">
          <h2 class="section-title">高级 JSON 模式</h2>
          <div class="help-text" style="margin-bottom: 16px;">直接编辑系统底层 <code>config.json</code> 文件。请确保 JSON 格式合法。</div>
          <textarea class="feishu-input json-editor" spellcheck="false" v-model="mockJson"></textarea>
        </div>

        <div class="form-actions">
          <button class="feishu-btn feishu-btn-outline" style="margin-right: 12px;">取消更改</button>
          <button class="feishu-btn" @click="saveConfig">保存配置并应用</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const currentTab = ref('general')

const mockJson = ref(`{
  "nas_path": "/vol2/1000/vedio/bililive-go/",
  "proxy": "192.168.10.10:7890",
  "disk_threshold": {
    "trigger_percent": 90,
    "safe_percent": 80
  },
  "merge_format": "mp4"
}`)

const saveConfig = () => {
  // 假装保存动画
  alert('✅ 系统配置已成功保存！后台服务可能需要 3 秒重载。')
}
</script>

<style scoped>
.settings-page { padding: 0; overflow: hidden; }
.settings-layout { display: flex; height: calc(100vh - 160px); min-height: 500px; }

/* 左侧导航 */
.settings-sidebar { width: 220px; background: #fafbfc; border-right: 1px solid var(--border-color); padding: 24px 0; }
.nav-header { margin: 0 0 16px 24px; font-size: 14px; color: var(--text-placeholder); font-weight: 500; text-transform: uppercase; }
.settings-nav { list-style: none; padding: 0; margin: 0; }
.settings-nav li { padding: 12px 24px; font-size: 14px; color: var(--text-regular); cursor: pointer; border-left: 3px solid transparent; transition: all 0.2s; }
.settings-nav li:hover { background: #f0f1f2; color: var(--text-title); }
.settings-nav li.active { background: var(--color-primary-bg); color: var(--color-primary); border-left-color: var(--color-primary); font-weight: 500; }

/* 右侧内容 */
.settings-content { flex: 1; padding: 32px 40px; overflow-y: auto; display: flex; flex-direction: column; position: relative; }
.section-title { margin: 0 0 24px 0; font-size: 20px; font-weight: 600; color: var(--text-title); }

.form-row { margin-bottom: 24px; max-width: 600px; }
.form-row label { display: block; font-weight: 500; margin-bottom: 8px; color: var(--text-title); font-size: 14px; }
.help-text { font-size: 12px; color: var(--text-placeholder); margin-top: 6px; }

/* 复选框与滑块 */
.checkbox-label { display: flex; align-items: center; gap: 8px; cursor: pointer; font-weight: 400 !important; }
.slider-group { display: flex; align-items: center; gap: 16px; }
.feishu-slider { flex: 1; accent-color: var(--color-primary); }
.slider-value { font-weight: 600; color: var(--color-primary); width: 40px; }

/* 提示框 */
.alert-box { padding: 12px 16px; border-radius: var(--radius-md); font-size: 14px; margin-bottom: 24px; display: flex; align-items: center; gap: 8px; }
.alert-box.warning { background: #fff2e8; border: 1px solid #ffd591; color: #d46b08; }

/* JSON 编辑器 */
.json-editor { font-family: "SFMono-Regular", Consolas, monospace; height: 300px; resize: vertical; background: #fafbfc; line-height: 1.5; }

/* 底部操作区 */
.form-actions { margin-top: auto; padding-top: 24px; border-top: 1px solid var(--border-color); display: flex; justify-content: flex-start; }
</style>

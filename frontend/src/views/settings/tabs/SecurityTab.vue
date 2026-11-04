<template>
  <div class="security-tab">
    <div class="form-section-card">
      <div class="section-title">
        <span class="section-icon">🔐</span>
        <span>修改管理密码</span>
      </div>
      <p class="section-desc">修改当前 Web 系统的登录访问密码。密码仅保存在本地安全凭据中，修改后即刻生效。</p>

      <el-form label-width="120px" label-position="left" class="security-form" @submit.prevent="handleSubmit">
        <el-form-item label="当前密码" required>
          <el-input
            v-model="oldPassword"
            type="password"
            show-password
            placeholder="请输入当前正在使用的旧密码"
            class="security-input"
          />
        </el-form-item>

        <el-form-item label="新密码" required>
          <el-input
            v-model="newPassword"
            type="password"
            show-password
            placeholder="请输入新密码（至少 6 个字符）"
            class="security-input"
          />
        </el-form-item>

        <el-form-item label="确认新密码" required>
          <el-input
            v-model="confirmPassword"
            type="password"
            show-password
            placeholder="请再次输入新密码"
            class="security-input"
          />
        </el-form-item>

        <div class="security-actions">
          <el-button
            type="primary"
            :loading="saving"
            :disabled="!canSubmit"
            class="btn-save-pwd"
            @click="handleSubmit"
          >
            确认修改密码
          </el-button>
          <span v-if="errorMsg" class="security-err">{{ errorMsg }}</span>
        </div>
      </el-form>
    </div>

    <!-- Security Tips Card -->
    <div class="security-tip-card">
      <span class="tip-icon">💡</span>
      <div class="tip-content">
        <div class="tip-title">密码管理提示</div>
        <div class="tip-text">
          修改密码后，旧的浏览器会话将自动失效并更新。如果遗忘密码，可在 Docker 容器日志中查看初始生成的密码，或在启动环境变量中指定 <code>PASSWORD</code> 参数直接重设。
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { ElMessage } from "element-plus";
import { changePassword } from "@/api/auth";

const oldPassword = ref("");
const newPassword = ref("");
const confirmPassword = ref("");
const saving = ref(false);
const errorMsg = ref("");

const canSubmit = computed(() => {
  return oldPassword.value.trim().length > 0 &&
         newPassword.value.length >= 6 &&
         confirmPassword.value.length >= 6;
});

async function handleSubmit() {
  errorMsg.value = "";
  if (newPassword.value.length < 6) {
    errorMsg.value = "新密码长度不能少于 6 位";
    return;
  }
  if (newPassword.value !== confirmPassword.value) {
    errorMsg.value = "两次输入的新密码不一致";
    return;
  }

  saving.value = true;
  try {
    await changePassword(oldPassword.value, newPassword.value);
    ElMessage.success("密码修改成功，请妥善保管新密码！");
    oldPassword.value = "";
    newPassword.value = "";
    confirmPassword.value = "";
  } catch (err: any) {
    errorMsg.value = err?.message || "密码修改失败，请检查旧密码是否正确";
  } finally {
    saving.value = false;
  }
}
</script>

<style scoped>
.security-tab {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
}
.form-section-card {
  background: var(--surface);
  border: 1px solid var(--hairline);
  border-radius: var(--r-md);
  padding: 18px 20px;
}
.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--charcoal);
  padding-bottom: 12px;
  border-bottom: 1px solid var(--hairline);
  margin-bottom: 12px;
}
.section-icon {
  font-size: 16px;
}
.section-desc {
  font-size: 12px;
  color: var(--slate);
  margin-bottom: 18px;
  line-height: 1.5;
}
.security-input {
  max-width: 380px;
}
.security-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-top: 8px;
}
.btn-save-pwd {
  padding: 8px 24px;
  font-size: 13px;
  font-weight: 500;
}
.security-err {
  font-size: 12px;
  color: #c93b2b;
}
.security-tip-card {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  background: var(--surface, #fafafa);
  border: 1px dashed var(--hairline, #e0e0e0);
  border-radius: var(--r-sm, 6px);
  padding: 12px 16px;
}
.tip-icon {
  font-size: 16px;
  line-height: 1.4;
}
.tip-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.tip-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--charcoal, #333);
}
.tip-text {
  font-size: 12px;
  color: var(--slate, #666);
  line-height: 1.5;
}
.tip-text code {
  background: var(--surface-hover, #eee);
  padding: 1px 5px;
  border-radius: 4px;
  font-family: var(--font-mono);
}

@media (max-width: 680px) {
  .security-form :deep(.el-form-item) {
    flex-direction: column;
    align-items: flex-start;
  }
  .security-form :deep(.el-form-item__label) {
    width: 100% !important;
    margin-bottom: 6px;
    height: auto;
    line-height: 1.4;
  }
  .security-form :deep(.el-form-item__content) {
    width: 100%;
  }
  .security-input {
    max-width: 100% !important;
  }
}
</style>

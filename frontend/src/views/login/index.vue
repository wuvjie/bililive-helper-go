<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-logo">BH</div>
      <h1 class="login-title">Bililive Helper</h1>
      <p class="login-sub">直播录制管理系统</p>
      <el-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleLogin" class="login-form">
        <el-form-item prop="password">
          <label class="login-label">管理员密码</label>
          <el-input
            v-model="form.password"
            type="password"
            placeholder="••••••••"
            show-password
            :prefix-icon="Lock"
            class="login-input"
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" class="login-btn" @click="handleLogin">
            进入控制台
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, type FormInstance } from "element-plus";
import { Lock } from "@element-plus/icons-vue";
import { login } from "@/api/auth";
import { useAuthStore } from "@/store/modules/auth";

const router = useRouter();
const auth = useAuthStore();
const formRef = ref<FormInstance>();
const loading = ref(false);
const form = reactive({ password: "" });
const rules = { password: [{ required: true, message: "请输入密码", trigger: "blur" }] };

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid) return;
  loading.value = true;
  try {
    await login(form.password);
    auth.markAuthenticated();
    ElMessage.success("登录成功");
    router.push("/");
  } catch { /* handled */ }
  finally { loading.value = false; }
}
</script>

<style scoped>
.login-container {
  height: 100vh; display: flex; align-items: center; justify-content: center;
  background: var(--surface);
  background-image: radial-gradient(var(--hairline) 1px, transparent 1px);
  background-size: 16px 16px;
}

.login-card {
  width: 380px; background: var(--canvas);
  border: 1px solid var(--hairline);
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.02);
  animation: slideUp 0.25s ease-out;
  display: flex; flex-direction: column; align-items: center;
}

@keyframes slideUp { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }

.login-logo {
  width: 48px; height: 48px; margin-bottom: 16px;
  background: var(--primary); color: var(--on-primary);
  border-radius: 8px; display: flex; align-items: center; justify-content: center;
  font-size: 18px; font-weight: 700;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.login-title {
  font-size: 16px; font-weight: 600; color: var(--ink);
  letter-spacing: -0.3px; margin-bottom: 4px;
}

.login-sub {
  font-size: 12px; color: #888888; margin-bottom: 28px;
}

.login-form {
  width: 100%;
  display: flex; flex-direction: column; gap: 0;
}

.login-label {
  display: block; font-size: 12px; color: #888888; font-weight: 500;
  margin-bottom: 6px; padding-left: 2px;
}

.login-input :deep(.el-input__wrapper) {
  padding-left: 14px;
  background: rgba(244, 244, 245, 0.6);
  border-radius: var(--r-md);
  box-shadow: none !important;
  border: 1px solid var(--hairline);
  transition: all 0.15s;
}
.login-input :deep(.el-input__wrapper):focus-within {
  background: var(--canvas);
  border-color: var(--ink);
}
.login-input :deep(.el-input__inner) {
  font-family: var(--font-mono);
  letter-spacing: 0.15em;
  font-size: 14px;
}
.login-input :deep(.el-input__inner)::placeholder {
  letter-spacing: 0.3em;
  color: var(--stone);
}

.login-btn {
  width: 100%; height: 40px;
  letter-spacing: 0.1em; font-size: 13px;
  border-radius: var(--r-md);
  margin-top: 8px;
}
.login-btn:active {
  transform: scale(0.99);
}
</style>

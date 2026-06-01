<template>
  <div class="login-container">
    <div class="login-left">
      <!-- Atmospheric glow -->
      <div class="glow glow-blue" />
      <div class="glow glow-orange" />
      <div class="left-content">
        <p class="mono-eyebrow">BILILIVE HELPER</p>
        <h1>Smart recording<br />management.</h1>
        <p class="lead">自动合并录制分片、智能磁盘管理、定时任务调度、硬件加速转码。</p>
      </div>
    </div>

    <div class="login-right">
      <div class="form-card">
        <h2>Welcome back.</h2>
        <p class="form-sub">输入密码以继续</p>
        <el-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleLogin">
          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="Password"
              size="large"
              show-password
              :prefix-icon="Lock"
              @keyup.enter="handleLogin"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" size="large" :loading="loading" class="login-btn" @click="handleLogin">
              Sign in →
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, type FormInstance } from "element-plus";
import { Lock } from "@element-plus/icons-vue";
import { login } from "@/api/auth";

const router = useRouter();
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
    ElMessage.success("登录成功");
    router.push("/");
  } catch { /* handled */ }
  finally { loading.value = false; }
}
</script>

<style scoped>
.login-container { display: flex; height: 100vh; overflow: hidden; background: var(--canvas); }

.login-left {
  flex: 1; position: relative; display: flex;
  align-items: center; justify-content: center; overflow: hidden;
}

/* Atmospheric glows */
.glow { position: absolute; border-radius: 50%; pointer-events: none; filter: blur(80px); }
.glow-blue {
  width: 500px; height: 500px; top: 10%; left: 20%;
  background: radial-gradient(circle, rgba(59, 158, 255, 0.18) 0%, transparent 70%);
}
.glow-orange {
  width: 400px; height: 400px; bottom: 10%; right: 10%;
  background: radial-gradient(circle, rgba(255, 128, 31, 0.12) 0%, transparent 70%);
}

.left-content { position: relative; z-index: 2; padding: 48px; max-width: 500px; }

.mono-eyebrow {
  font-family: var(--font-mono); font-size: 12px;
  letter-spacing: 0.1em; color: var(--mute); margin-bottom: 24px;
}

.left-content h1 {
  font-family: var(--font-display); font-size: 52px; font-weight: 400;
  line-height: 1.05; letter-spacing: -0.5px; color: var(--ink);
  margin-bottom: 20px;
}

.lead { font-size: 18px; line-height: 1.6; color: var(--body-text); }

.login-right {
  width: 480px; display: flex; align-items: center; justify-content: center;
  background: var(--surface-card); border-left: 1px solid var(--hairline);
}

.form-card {
  width: 320px; padding: 32px;
  background: var(--surface-elevated);
  border: 1px solid var(--hairline-strong);
  border-radius: var(--r-lg);
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }

.form-card h2 {
  font-family: var(--font-display); font-size: 24px; font-weight: 400;
  letter-spacing: -0.3px; margin-bottom: 4px;
}
.form-sub { font-size: 14px; color: var(--mute); margin-bottom: 24px; }

.login-btn { width: 100%; height: 40px; }

@media (max-width: 968px) {
  .login-left { display: none; }
  .login-right { width: 100%; border-left: none; }
}
</style>

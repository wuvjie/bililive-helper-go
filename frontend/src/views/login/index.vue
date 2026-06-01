<template>
  <div class="login-container">
    <!-- Left panel -->
    <div class="login-left">
      <div class="left-content">
        <div class="brand-badge">
          <span class="mono-label">BILILIVE HELPER</span>
        </div>
        <h1>Smart recording<br />management.</h1>
        <p class="lead">自动合并录制分片、智能磁盘管理、定时任务调度、硬件加速转码。</p>
      </div>
      <!-- Mesh gradient background -->
      <div class="mesh-gradient" />
    </div>

    <!-- Right form panel -->
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
            <el-button
              type="primary"
              size="large"
              :loading="loading"
              class="login-btn"
              @click="handleLogin"
            >
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
.login-container {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: #ffffff;
}

.login-left {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: var(--primary);
  color: var(--on-primary);
}

.mesh-gradient {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse at 20% 50%, rgba(0, 112, 243, 0.3) 0%, transparent 60%),
    radial-gradient(ellipse at 80% 20%, rgba(121, 40, 202, 0.25) 0%, transparent 50%),
    radial-gradient(ellipse at 60% 80%, rgba(255, 0, 128, 0.2) 0%, transparent 50%),
    radial-gradient(ellipse at 40% 30%, rgba(80, 227, 194, 0.15) 0%, transparent 40%);
  pointer-events: none;
}

.left-content {
  position: relative;
  z-index: 2;
  padding: 48px;
  max-width: 480px;
}

.brand-badge {
  margin-bottom: 32px;
}
.mono-label {
  font-family: var(--font-mono);
  font-size: 12px;
  letter-spacing: 0.08em;
  opacity: 0.6;
  text-transform: uppercase;
}

.left-content h1 {
  font-size: 48px;
  font-weight: 600;
  letter-spacing: -2.4px;
  line-height: 1;
  margin-bottom: 16px;
}

.lead {
  font-size: 18px;
  line-height: 28px;
  opacity: 0.7;
}

.login-right {
  width: 480px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--canvas-soft);
}

.form-card {
  width: 320px;
  background: #ffffff;
  padding: 32px;
  border-radius: var(--r-lg);
  box-shadow: var(--shadow-md);
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

.form-card h2 {
  font-size: 24px;
  font-weight: 600;
  letter-spacing: -0.96px;
  margin-bottom: 4px;
}

.form-sub {
  font-size: 14px;
  color: var(--mute);
  margin-bottom: 24px;
}

.login-btn {
  width: 100%;
  height: 40px;
  border-radius: var(--r-sm) !important;
}

@media (max-width: 968px) {
  .login-left { display: none; }
  .login-right { width: 100%; }
}
</style>

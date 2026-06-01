<template>
  <div class="login-container">
    <!-- Left decorative panel -->
    <div class="login-left">
      <div class="left-content">
        <div class="brand-mark">
          <span class="brand-icon">◆</span>
        </div>
        <h1>Bililive Helper</h1>
        <p class="subtitle">Smart livestream recording management</p>
        <div class="features">
          <div class="feature-item">
            <span class="feature-dot" />
            <span>自动合并录制分片</span>
          </div>
          <div class="feature-item">
            <span class="feature-dot" />
            <span>智能磁盘空间管理</span>
          </div>
          <div class="feature-item">
            <span class="feature-dot" />
            <span>定时任务调度</span>
          </div>
          <div class="feature-item">
            <span class="feature-dot" />
            <span>硬件加速转码</span>
          </div>
        </div>
      </div>
      <!-- Grid pattern background -->
      <div class="grid-pattern" />
    </div>

    <!-- Right form panel -->
    <div class="login-right">
      <div class="form-wrapper">
        <h2>Welcome back</h2>
        <p class="form-subtitle">输入密码以继续</p>
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
              Sign in
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
  } catch { /* handled by interceptor */ }
  finally { loading.value = false; }
}
</script>

<style scoped>
.login-container {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: var(--canvas);
}

.login-left {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-right: 1px solid var(--hairline);
}

.grid-pattern {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(var(--hairline) 1px, transparent 1px),
    linear-gradient(90deg, var(--hairline) 1px, transparent 1px);
  background-size: 48px 48px;
  opacity: 0.4;
  pointer-events: none;
}

.left-content {
  position: relative;
  z-index: 2;
  text-align: center;
  padding: 48px;
  max-width: 400px;
}

.brand-mark {
  margin-bottom: 32px;
}
.brand-icon {
  font-size: 40px;
  color: var(--primary);
  display: inline-block;
}

.left-content h1 {
  font-family: var(--font-display);
  font-size: 32px;
  font-weight: 600;
  letter-spacing: -1px;
  color: var(--ink);
  margin-bottom: 8px;
}

.subtitle {
  font-size: 16px;
  color: var(--ink-subtle);
  margin-bottom: 48px;
  letter-spacing: -0.02em;
}

.features {
  display: flex;
  flex-direction: column;
  gap: 16px;
  text-align: left;
}
.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
  color: var(--ink-muted);
}
.feature-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--primary);
  flex-shrink: 0;
  opacity: 0.7;
}

.login-right {
  width: 440px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--surface-1);
}

.form-wrapper {
  width: 300px;
  animation: slideUp 0.4s ease-out;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}

.form-wrapper h2 {
  font-family: var(--font-display);
  font-size: 22px;
  font-weight: 600;
  letter-spacing: -0.4px;
  color: var(--ink);
  margin-bottom: 4px;
}

.form-subtitle {
  font-size: 14px;
  color: var(--ink-subtle);
  margin-bottom: 28px;
}

.login-btn {
  width: 100%;
  height: 40px;
  font-weight: 500;
}

@media (max-width: 968px) {
  .login-left { display: none; }
  .login-right { width: 100%; }
}
</style>

<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-logo">BH</div>
      <h1>Bililive Helper</h1>
      <p class="login-sub">直播录制管理系统</p>
      <el-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleLogin">
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="输入密码" size="large" show-password :prefix-icon="Lock" @keyup.enter="handleLogin" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="large" :loading="loading" class="login-btn" @click="handleLogin">登录</el-button>
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

const router = useRouter();
const formRef = ref<FormInstance>();
const loading = ref(false);
const form = reactive({ password: "" });
const rules = { password: [{ required: true, message: "请输入密码", trigger: "blur" }] };

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid) return;
  loading.value = true;
  try { await login(form.password); ElMessage.success("登录成功"); router.push("/"); }
  catch { /* handled */ }
  finally { loading.value = false; }
}
</script>

<style scoped>
.login-container {
  height: 100vh; display: flex; align-items: center; justify-content: center;
  background: var(--canvas);
}

.login-card {
  width: 400px; background: var(--canvas);
  border: 1px solid var(--hairline);
  border-radius: var(--r-lg);
  padding: 48px 40px; text-align: center;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04), 0 4px 12px rgba(0,0,0,0.06);
  animation: slideUp 0.2s ease-out;
}

@keyframes slideUp { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: translateY(0); } }

.login-logo {
  width: 56px; height: 56px; margin: 0 auto 20px;
  background: var(--ink); color: var(--on-primary);
  border-radius: var(--r-md); display: flex; align-items: center; justify-content: center;
  font-size: 20px; font-weight: 700;
}

.login-card h1 { font-size: 24px; font-weight: 600; color: var(--ink); margin-bottom: 4px; letter-spacing: -0.5px; }
.login-sub { font-size: 15px; color: var(--muted); margin-bottom: 28px; }
.login-btn { width: 100%; }
</style>

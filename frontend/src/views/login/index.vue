<template>
  <div class="login-container">
    <!-- Left decorative panel -->
    <div class="login-left">
      <div class="deco-circle deco-1" />
      <div class="deco-circle deco-2" />
      <div class="deco-square deco-3" />
      <div class="deco-square deco-4" />
      <div class="deco-dots" />
      <div class="left-content">
        <div class="left-icon">📺</div>
        <h1>Bililive Helper</h1>
        <p>智能直播录制管理系统</p>
        <div class="features">
          <div class="feature-item"><span class="dot" />自动合并录制分片</div>
          <div class="feature-item"><span class="dot" />智能磁盘空间管理</div>
          <div class="feature-item"><span class="dot" />定时任务调度</div>
          <div class="feature-item"><span class="dot" />硬件加速转码</div>
        </div>
      </div>
    </div>

    <!-- Right form panel -->
    <div class="login-right">
      <div class="login-form-wrapper">
        <div class="form-header">
          <h2>欢迎回来</h2>
          <p>请输入密码登录系统</p>
        </div>
        <el-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleLogin">
          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
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
              登录
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
const rules = {
  password: [{ required: true, message: "请输入密码", trigger: "blur" }]
};

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid) return;
  loading.value = true;
  try {
    await login(form.password);
    ElMessage.success("登录成功");
    router.push("/");
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

/* Left panel */
.login-left {
  flex: 1;
  position: relative;
  background: linear-gradient(135deg, #1a237e 0%, #283593 50%, #3949ab 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.left-content {
  position: relative;
  z-index: 2;
  color: #fff;
  text-align: center;
  padding: 40px;
}
.left-icon { font-size: 64px; margin-bottom: 20px; }
.left-content h1 {
  font-size: 36px;
  font-weight: 700;
  margin-bottom: 12px;
  letter-spacing: 2px;
}
.left-content p {
  font-size: 16px;
  opacity: 0.8;
  margin-bottom: 40px;
}

.features {
  display: flex;
  flex-direction: column;
  gap: 16px;
  text-align: left;
  max-width: 260px;
  margin: 0 auto;
}
.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
  opacity: 0.9;
}
.feature-item .dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #64b5f6;
  flex-shrink: 0;
}

/* Decorative elements */
.deco-circle {
  position: absolute;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.1);
}
.deco-1 { width: 300px; height: 300px; top: -80px; right: -60px; animation: float 8s ease-in-out infinite; }
.deco-2 { width: 200px; height: 200px; bottom: -40px; left: -40px; animation: float 6s ease-in-out infinite reverse; }

.deco-square {
  position: absolute;
  border: 2px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
}
.deco-3 { width: 80px; height: 80px; top: 20%; right: 15%; transform: rotate(15deg); animation: spin 12s linear infinite; }
.deco-4 { width: 50px; height: 50px; bottom: 25%; left: 20%; transform: rotate(-20deg); animation: spin 10s linear infinite reverse; }

.deco-dots {
  position: absolute;
  width: 100%;
  height: 100%;
  background-image: radial-gradient(rgba(255,255,255,0.08) 1px, transparent 1px);
  background-size: 30px 30px;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-20px); }
}
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Right panel */
.login-right {
  width: 480px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-card);
  position: relative;
  transition: background 0.3s;
}

.login-form-wrapper {
  width: 320px;
  animation: slideUp 0.6s ease-out;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(30px); }
  to { opacity: 1; transform: translateY(0); }
}

.form-header {
  margin-bottom: 36px;
}
.form-header h2 {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 8px;
}
.form-header p {
  font-size: 14px;
  color: var(--text-placeholder);
}

.login-btn {
  width: 100%;
  height: 44px;
  font-size: 15px;
  border-radius: var(--radius-md);
}

/* Responsive */
@media (max-width: 968px) {
  .login-left { display: none; }
  .login-right { width: 100%; }
}
</style>

<template>
  <div class="setup-container">
    <div class="setup-card">
      <div class="setup-logo">BH</div>
      <h1 class="setup-title">Bililive Helper</h1>
      <p class="setup-sub">首次启动，请设置登录密码</p>

      <el-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleSubmit" class="setup-form">
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="至少 6 位"
            show-password
            size="large"
          />
        </el-form-item>
        <el-form-item prop="confirmPassword">
          <el-input
            v-model="form.confirmPassword"
            type="password"
            placeholder="确认密码"
            show-password
            size="large"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" class="setup-btn" @click="handleSubmit">
            开始使用
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, type FormInstance } from "element-plus";
import { setupInit, setupStatus } from "@/api/setup";
import { markAuthenticated } from "@/router";

const router = useRouter();
const formRef = ref<FormInstance>();
const loading = ref(false);
const form = reactive({ password: "", confirmPassword: "" });

const validateConfirm = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  if (value !== form.password) {
    callback(new Error("两次输入的密码不一致"));
  } else {
    callback();
  }
};

const rules = {
  password: [
    { required: true, message: "请设置密码", trigger: "blur" },
    { min: 6, message: "密码至少 6 个字符", trigger: "blur" }
  ],
  confirmPassword: [
    { required: true, message: "请再次输入密码", trigger: "blur" },
    { validator: validateConfirm, trigger: "blur" }
  ]
};

onMounted(async () => {
  try {
    const status = await setupStatus();
    if (!status.first_run) {
      router.replace("/login");
    }
  } catch {
    // If status check fails, allow setup to proceed
  }
});

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid) return;
  loading.value = true;
  try {
    await setupInit({ password: form.password });
    markAuthenticated();
    ElMessage.success("配置完成");
    router.push("/");
  } catch {
    // Error handled by interceptor
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.setup-container {
  height: 100vh; display: flex; align-items: center; justify-content: center;
  background: var(--surface);
  background-image: radial-gradient(var(--hairline) 1px, transparent 1px);
  background-size: 16px 16px;
}

.setup-card {
  width: 360px; background: var(--canvas);
  border: 1px solid var(--hairline);
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.02);
  animation: slideUp 0.25s ease-out;
  display: flex; flex-direction: column; align-items: center;
}

@keyframes slideUp { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }

.setup-logo {
  width: 48px; height: 48px; margin-bottom: 16px;
  background: var(--primary); color: var(--on-primary);
  border-radius: 8px; display: flex; align-items: center; justify-content: center;
  font-size: 18px; font-weight: 700;
}

.setup-title {
  font-size: 16px; font-weight: 600; color: var(--ink);
  letter-spacing: -0.3px; margin-bottom: 4px;
}

.setup-sub {
  font-size: 12px; color: #888888; margin-bottom: 28px;
}

.setup-form { width: 100%; }

.setup-form :deep(.el-form-item__error) {
  color: #c4554d;
  font-size: 12px;
  padding-top: 2px;
  position: static;
}

.setup-btn {
  width: 100%; height: 40px;
  letter-spacing: 0.05em; font-size: 14px;
  border-radius: var(--r-md);
  margin-top: 8px;
}
</style>

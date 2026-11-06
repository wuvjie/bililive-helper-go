<template>
  <div class="setup-container">
    <div class="setup-card">
      <div class="setup-header">
        <div class="brand-icon-wrapper">
          <svg class="brand-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="2" y="4" width="15" height="16" rx="3" />
            <path d="M17 9.5l4-3v11l-4-3" />
            <circle cx="9.5" cy="12" r="2.5" fill="currentColor" />
          </svg>
        </div>
        <h1 class="setup-title">Bililive Helper Go</h1>
        <p class="setup-sub">首次初始化部署，请设置管理密码</p>
      </div>

      <el-form ref="formRef" :model="form" :rules="rules" class="setup-form" @submit.prevent>
        <el-form-item prop="password" class="setup-form-item">
          <label class="field-label">设置管理员密码</label>
          <el-input
            v-model="form.password"
            type="password"
            placeholder="至少 6 个字符"
            show-password
            size="large"
            :prefix-icon="Lock"
            class="setup-input"
          />
        </el-form-item>

        <el-form-item prop="confirmPassword" class="setup-form-item">
          <label class="field-label">确认管理员密码</label>
          <el-input
            v-model="form.confirmPassword"
            type="password"
            placeholder="请再次输入新密码（至少 6 个字符）"
            show-password
            size="large"
            :prefix-icon="Lock"
            class="setup-input"
            @keydown.enter.prevent="handleSubmit"
          />
        </el-form-item>

        <el-form-item class="submit-form-item">
          <el-button type="primary" :loading="loading" class="setup-btn" @click="handleSubmit">
            <span v-if="loading">正在初始化...</span>
            <span v-else class="btn-inner">
              <span>完成配置并开始使用</span>
              <svg class="btn-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M5 12h14M13 6l6 6-6 6"/>
              </svg>
            </span>
          </el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="setup-footer">
      <span>Bililive Helper Go · 初始化向导</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, type FormInstance } from "element-plus";
import { Lock } from "@element-plus/icons-vue";
import { setupInit, setupStatus } from "@/api/setup";
import { useAuthStore } from "@/store/modules/auth";

const router = useRouter();
const auth = useAuthStore();
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
    { min: 6, message: "确认密码至少 6 个字符", trigger: "blur" },
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
  if (loading.value) return;
  loading.value = true;
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid) {
    loading.value = false;
    return;
  }
  try {
    await setupInit({ password: form.password });
    auth.markAuthenticated();
    ElMessage.success("配置完成，已自动登录");
    router.push("/");
  } catch {
    loading.value = false;
  }
}
</script>

<style scoped>
.setup-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #fbfbfa;
  background-image:
    radial-gradient(ellipse at 50% -20%, rgba(30, 30, 32, 0.05) 0%, transparent 70%),
    radial-gradient(#e5e3df 1px, transparent 1px);
  background-size: 100% 100%, 20px 20px;
  padding: 24px 16px;
  box-sizing: border-box;
}

.setup-card {
  width: 100%;
  max-width: 400px;
  background: #ffffff;
  border: 1px solid var(--hairline, #e5e3df);
  border-radius: 14px;
  padding: 40px 32px 32px;
  box-shadow:
    0 1px 3px rgba(0, 0, 0, 0.02),
    0 16px 36px -12px rgba(0, 0, 0, 0.06);
  box-sizing: border-box;
}

.setup-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  margin-bottom: 24px;
}

.brand-icon-wrapper {
  width: 52px;
  height: 52px;
  border-radius: 13px;
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  margin-bottom: 16px;
  box-shadow: 0 8px 20px -4px rgba(37, 99, 235, 0.28);
}

.brand-icon {
  width: 26px;
  height: 26px;
}

.setup-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--ink, #1c1c1e);
  letter-spacing: -0.4px;
  margin: 0 0 6px 0;
}

.setup-sub {
  font-size: 13px;
  color: #888681;
  margin: 0;
  line-height: 1.4;
}

.setup-form {
  width: 100%;
}

.setup-form-item {
  margin-bottom: 18px;
}

.field-label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  color: #5d5b54;
  letter-spacing: 0.2px;
  margin-bottom: 7px;
}

.setup-input :deep(.el-input__wrapper) {
  padding: 4px 12px;
  border-radius: 8px;
  background: #fdfdfd;
  box-shadow: 0 0 0 1px #e5e3df inset !important;
  transition: all 0.2s ease;
}

.setup-input :deep(.el-input__wrapper):hover {
  box-shadow: 0 0 0 1px #c8c4be inset !important;
}

.setup-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1.5px #2563eb inset, 0 0 0 3px rgba(37, 99, 235, 0.12) !important;
  background: #ffffff;
}

.setup-input :deep(.el-input__inner) {
  font-size: 14px;
  color: #1c1c1e;
  height: 38px;
}

.submit-form-item {
  margin-top: 24px;
  margin-bottom: 0;
}

.setup-btn {
  width: 100%;
  height: 44px;
  font-size: 14px;
  font-weight: 600;
  border-radius: 8px;
  background: var(--primary, #2563eb);
  border-color: var(--primary, #2563eb);
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  letter-spacing: 0.2px;
}

.setup-btn:hover {
  background: var(--primary-hover, #1d4ed8);
  border-color: var(--primary-hover, #1d4ed8);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.25);
}

.btn-inner {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.btn-arrow {
  width: 16px;
  height: 16px;
  transition: transform 0.2s ease;
}

.setup-btn:hover .btn-arrow {
  transform: translateX(3px);
}

.setup-footer {
  margin-top: 24px;
  font-size: 12px;
  color: #a4a097;
  text-align: center;
  user-select: none;
}

@media (max-width: 640px) {
  .setup-container {
    padding: 16px 12px;
  }
  .setup-card {
    padding: 28px 20px 24px;
    border-radius: 12px;
  }
  .setup-input :deep(.el-input__inner) {
    font-size: 16px !important; /* Prevents iOS Safari auto-zoom on input focus */
  }
}

html.dark {
  .setup-container {
    background: #09090b;
    background-image:
      radial-gradient(ellipse at 50% -20%, rgba(37, 99, 235, 0.12) 0%, transparent 70%),
      radial-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px);
  }
  .setup-card {
    background: #141416;
    border-color: #27272a;
    box-shadow:
      0 1px 3px rgba(0, 0, 0, 0.5),
      0 20px 40px -15px rgba(0, 0, 0, 0.7);
  }
  .brand-icon-wrapper {
    background: #1f1f23;
    color: #3b82f6;
    border: 1px solid #27272a;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
  }
  .setup-title {
    color: #f4f4f5;
  }
  .setup-sub {
    color: #a1a1aa;
  }
  .field-label {
    color: #d4d4d8;
  }
  .setup-input :deep(.el-input__wrapper) {
    background: #18181b;
    box-shadow: 0 0 0 1px #27272a inset !important;
  }
  .setup-input :deep(.el-input__wrapper:hover) {
    box-shadow: 0 0 0 1px #3f3f46 inset !important;
  }
  .setup-input :deep(.el-input__wrapper.is-focus) {
    background: #1f1f23;
    box-shadow: 0 0 0 1.5px #2563eb inset, 0 0 0 3px rgba(37, 99, 235, 0.2) !important;
  }
  .setup-input :deep(.el-input__inner) {
    color: #f4f4f5;
  }
  .setup-btn {
    background: #2563eb;
    border-color: #2563eb;
    color: #ffffff;
  }
  .setup-btn:hover {
    background: #3b82f6;
    border-color: #3b82f6;
    box-shadow: 0 4px 16px rgba(37, 99, 235, 0.35);
  }
  .setup-footer {
    color: #71717a;
  }
}
</style>

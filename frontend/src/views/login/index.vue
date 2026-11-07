<template>
  <div class="login-container">
    <div
      class="login-card-wrapper"
      :class="{ 'is-shaking': isShaking }"
      @animationend="isShaking = false"
    >
      <div class="login-card">
      <!-- 品牌 Logo 与标题区 -->
      <div class="login-header">
        <div class="brand-icon-wrapper">
          <svg class="brand-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="2" y="4" width="15" height="16" rx="3" />
            <path d="M17 9.5l4-3v11l-4-3" />
            <circle cx="9.5" cy="12" r="2.5" fill="currentColor" />
          </svg>
        </div>
        <h1 class="brand-title">Bililive Helper Go</h1>
        <p class="brand-sub">录播自动合并与媒体管理平台</p>
      </div>

      <!-- 登录表单 -->
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        class="login-form"
        @submit.prevent
      >
        <el-form-item prop="password" class="password-form-item">
          <div class="field-header">
            <label class="field-label">管理员密码</label>
            <!-- 大写锁定激活警示 -->
            <transition name="el-fade-in">
              <span v-if="isCapsLock" class="caps-hint">
                <svg class="caps-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 4l-7 7h4v7h6v-7h4l-7-7z"/></svg>
                大写锁定已开启
              </span>
            </transition>
          </div>

          <el-input
            ref="passwordInputRef"
            v-model="form.password"
            type="password"
            placeholder="请输入管理员密码"
            show-password
            size="large"
            :prefix-icon="Lock"
            class="password-input"
            @input="onPasswordInput"
            @keydown="checkCapsLock"
            @keyup="checkCapsLock"
            @focus="checkCapsLock"
            @keydown.enter.prevent="handleLogin"
          />
        </el-form-item>

        <!-- 错误提示横幅：平滑就地显示于密码框下方，卡片头部与输入框永不跳跃 -->
        <transition name="banner-fade">
          <div v-if="errorMessage" class="login-error-banner">
            <svg class="error-banner-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="8" x2="12" y2="12" />
              <line x1="12" y1="16" x2="12.01" y2="16" />
            </svg>
            <span class="error-banner-text">{{ errorMessage }}</span>
          </div>
        </transition>

        <el-form-item class="submit-form-item">
          <el-button
            type="primary"
            :loading="loading"
            class="login-submit-btn"
            @click="handleLogin"
          >
            <span v-if="loading">正在验证身份...</span>
            <span v-else class="btn-inner">
              <span>进入控制台</span>
              <svg class="btn-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M5 12h14M13 6l6 6-6 6"/>
              </svg>
            </span>
          </el-button>
        </el-form-item>
      </el-form>
      </div>
    </div>

    <!-- 底部微信息 -->
    <div class="login-footer">
      <span>Bililive Helper Go · 本地私有化媒体助手</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, type FormInstance } from "element-plus";
import { Lock } from "@element-plus/icons-vue";
import { login } from "@/api/auth";
import { useAuthStore } from "@/store/modules/auth";

const router = useRouter();
const auth = useAuthStore();
const formRef = ref<FormInstance>();
const passwordInputRef = ref();

const loading = ref(false);
const isShaking = ref(false);
const isCapsLock = ref(false);
const errorMessage = ref("");

const form = reactive({ password: "" });
const rules = {
  password: [{ required: true, message: "请输入密码", trigger: "blur" }]
};

function onPasswordInput() {
  if (errorMessage.value) {
    errorMessage.value = "";
  }
}

function checkCapsLock(e: KeyboardEvent) {
  if (e.getModifierState) {
    isCapsLock.value = e.getModifierState("CapsLock");
  }
}

function triggerShake() {
  isShaking.value = false;
  // Next tick to allow restarting animation if already shaking
  requestAnimationFrame(() => {
    isShaking.value = true;
  });
}

async function handleLogin() {
  if (loading.value) return;
  loading.value = true;
  errorMessage.value = "";

  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid) {
    loading.value = false;
    triggerShake();
    passwordInputRef.value?.focus();
    return;
  }

  try {
    await login(form.password);
    auth.markAuthenticated();
    ElMessage.closeAll();
    ElMessage.success("登录成功");
    router.push("/");
  } catch (err: any) {
    loading.value = false;
    const serverMsg = err.response?.data?.error || err.response?.data?.message;
    errorMessage.value = serverMsg || "密码错误，请重新输入";
    triggerShake();
    passwordInputRef.value?.focus();
  }
}

onMounted(() => {
  // 若已处于登录状态，直接重定向至控制台
  if (auth.isAuthenticated) {
    router.replace("/");
    return;
  }
  // 页面载入时自动聚焦密码框
  setTimeout(() => {
    passwordInputRef.value?.focus();
  }, 100);
});
</script>

<style scoped>
.login-container {
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

.login-card {
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

/* 独立的震颤外层容器 — 绝不影响卡片自身的入场动画 */
.login-card-wrapper {
  width: 100%;
  max-width: 400px;
  display: flex;
  justify-content: center;
  will-change: transform;
}

/* 仿 macOS/iOS 锁屏输错密码的自然微阻尼震颤 — 使用 translate3d 开启硬件加速且不触发布局重排 */
@keyframes naturalShake {
  0%, 100% { transform: translate3d(0, 0, 0); }
  20%, 60% { transform: translate3d(-5px, 0, 0); }
  40%, 80% { transform: translate3d(5px, 0, 0); }
}

.is-shaking {
  animation: naturalShake 0.32s ease-in-out;
}

/* 错误横幅平滑展开渐变，无突兀缩放跳动 */
.banner-fade-enter-active,
.banner-fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.banner-fade-enter-from,
.banner-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* 头部品牌展示 */
.login-header {
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
  transition: transform 0.2s ease;
}

.login-card:hover .brand-icon-wrapper {
  transform: scale(1.03);
}

.brand-icon {
  width: 26px;
  height: 26px;
}

.brand-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--ink, #1c1c1e);
  letter-spacing: -0.4px;
  margin: 0 0 6px 0;
}

.brand-sub {
  font-size: 13px;
  color: #888681;
  margin: 0;
  line-height: 1.4;
}

/* 错误提示条 */
.login-error-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #dc2626;
  padding: 9px 12px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  margin-top: -6px;
  margin-bottom: 16px;
  line-height: 1.4;
}

.error-banner-icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.error-banner-text {
  flex: 1;
  line-height: 1.4;
}

/* 表单与输入框 */
.login-form {
  width: 100%;
}

.password-form-item {
  margin-bottom: 20px;
}

.field-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 7px;
  width: 100%;
}

.field-label {
  font-size: 12px;
  font-weight: 600;
  color: #5d5b54;
  letter-spacing: 0.2px;
}

.caps-hint {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #b45309;
  background: #fff7ed;
  border: 1px solid #fed7aa;
  padding: 1px 6px;
  border-radius: 4px;
  font-weight: 500;
}

.caps-icon {
  width: 11px;
  height: 11px;
}

.password-input :deep(.el-input__wrapper) {
  padding: 4px 12px 4px 12px;
  border-radius: 8px;
  background: #fdfdfd;
  box-shadow: 0 0 0 1px #e5e3df inset !important;
  transition: all 0.2s ease;
}

.password-input :deep(.el-input__wrapper):hover {
  box-shadow: 0 0 0 1px #c8c4be inset !important;
}

.password-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1.5px #2563eb inset, 0 0 0 3px rgba(37, 99, 235, 0.12) !important;
  background: #ffffff;
}

.password-input :deep(.el-input__inner) {
  font-size: 14px;
  letter-spacing: 0.1em;
  color: #1c1c1e;
  height: 38px;
}

.password-input :deep(.el-input__inner)::placeholder {
  font-size: 13px;
  letter-spacing: 0;
  color: #a4a097;
}

.submit-form-item {
  margin-bottom: 0;
}

/* 提交按钮 */
.login-submit-btn {
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

.login-submit-btn:hover {
  background: var(--primary-hover, #1d4ed8);
  border-color: var(--primary-hover, #1d4ed8);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.25);
}

.login-submit-btn:active {
  transform: translateY(0);
  box-shadow: none;
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

.login-submit-btn:hover .btn-arrow {
  transform: translateX(3px);
}

/* 底部微信息 */
.login-footer {
  margin-top: 24px;
  font-size: 12px;
  color: #a4a097;
  text-align: center;
  user-select: none;
}

@media (max-width: 640px) {
  .login-container {
    padding: 16px 12px;
  }
  .login-card {
    padding: 28px 20px 24px;
    border-radius: 12px;
  }
  .password-input :deep(.el-input__inner) {
    font-size: 16px !important; /* Prevents iOS Safari auto-zoom on input focus */
  }
}

html.dark {
  .login-container {
    background: #09090b;
    background-image:
      radial-gradient(ellipse at 50% -20%, rgba(37, 99, 235, 0.12) 0%, transparent 70%),
      radial-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px);
  }
  .login-card {
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
  .brand-title {
    color: #f4f4f5;
  }
  .brand-sub {
    color: #a1a1aa;
  }
  .field-label {
    color: #d4d4d8;
  }
  .caps-hint {
    color: #fbbf24;
    background: rgba(245, 158, 11, 0.12);
    border-color: rgba(245, 158, 11, 0.25);
  }
  .password-input :deep(.el-input__wrapper) {
    background: #18181b;
    box-shadow: 0 0 0 1px #27272a inset !important;
  }
  .password-input :deep(.el-input__wrapper:hover) {
    box-shadow: 0 0 0 1px #3f3f46 inset !important;
  }
  .password-input :deep(.el-input__wrapper.is-focus) {
    background: #1f1f23 !important;
    box-shadow: 0 0 0 1.5px #2563eb inset, 0 0 0 3px rgba(37, 99, 235, 0.2) !important;
  }
  .password-input :deep(.el-input__inner) {
    color: #f4f4f5;
  }
  .password-input :deep(.el-input__inner::placeholder) {
    color: #71717a;
  }
  .login-error-banner {
    background: rgba(239, 68, 68, 0.12);
    border-color: rgba(239, 68, 68, 0.3);
    color: #f87171;
  }
  .login-submit-btn {
    background: #2563eb;
    border-color: #2563eb;
    color: #ffffff;
  }
  .login-submit-btn:hover {
    background: #3b82f6;
    border-color: #3b82f6;
    box-shadow: 0 4px 16px rgba(37, 99, 235, 0.35);
  }
  .login-footer {
    color: #71717a;
  }
}
</style>

<template>
  <div class="login-container">
    <!-- 右上角浅色 / 暗黑模式快速切换 -->
    <div class="login-top-bar">
      <button
        type="button"
        class="login-theme-btn"
        :title="appStore.isDark ? '切换至浅色模式' : '切换至深色模式'"
        @click="appStore.toggleTheme"
      >
        <svg v-if="appStore.isDark" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="theme-svg">
          <circle cx="12" cy="12" r="5"/>
          <line x1="12" y1="1" x2="12" y2="3"/>
          <line x1="12" y1="21" x2="12" y2="23"/>
          <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
          <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
          <line x1="1" y1="12" x2="3" y2="12"/>
          <line x1="21" y1="12" x2="23" y2="12"/>
          <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
          <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
        </svg>
        <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="theme-svg">
          <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
        </svg>
      </button>
    </div>

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
import { useAppStore } from "@/store/modules/app";

const router = useRouter();
const auth = useAuthStore();
const appStore = useAppStore();
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
  if (auth.isAuthenticated) {
    router.replace("/");
    return;
  }
  setTimeout(() => {
    passwordInputRef.value?.focus();
  }, 100);
});
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: var(--canvas, #fbfbfa);
  background-image:
    radial-gradient(ellipse at 50% -20%, color-mix(in srgb, var(--primary, #2563eb) 8%, transparent) 0%, transparent 70%),
    radial-gradient(var(--hairline, #e5e3df) 1px, transparent 1px);
  background-size: 100% 100%, 20px 20px;
  padding: 24px 16px;
  box-sizing: border-box;
}

/* 顶部主题模式快速切换按钮 */
.login-top-bar {
  position: absolute;
  top: 18px;
  right: 20px;
}

.login-theme-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--r-sm, 4px);
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--hairline, #e5e3df);
  color: var(--slate, #5d5b54);
  cursor: pointer;
  transition: all 0.15s ease;
  box-shadow: var(--shadow-xs);
}

.login-theme-btn:hover {
  background: var(--surface, #f4f3f0);
  color: var(--ink, #1a1a1a);
  border-color: var(--hairline-strong, #c8c4be);
}

.theme-svg {
  width: 15px;
  height: 15px;
}

.login-card {
  width: 100%;
  max-width: 380px;
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--hairline, #e5e3df);
  border-radius: var(--r-xl, 14px);
  padding: 36px 32px 32px;
  box-shadow: var(--shadow-card);
  box-sizing: border-box;
}

.login-card-wrapper {
  width: 100%;
  max-width: 380px;
  display: flex;
  justify-content: center;
  will-change: transform;
}

/* 仿 macOS/iOS 锁屏输错密码的自然微阻尼震颤 */
@keyframes naturalShake {
  0%, 100% { transform: translate3d(0, 0, 0); }
  20%, 60% { transform: translate3d(-5px, 0, 0); }
  40%, 80% { transform: translate3d(5px, 0, 0); }
}

.is-shaking {
  animation: naturalShake 0.32s ease-in-out;
}

/* 错误横幅平滑展开渐变 */
.banner-fade-enter-active,
.banner-fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.banner-fade-enter-from,
.banner-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* 头部品牌展示：微矢量直出 */
.login-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  margin-bottom: 24px;
}

.brand-icon-wrapper {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: color-mix(in srgb, var(--primary, #2563eb) 10%, transparent);
  border: 1px solid color-mix(in srgb, var(--primary, #2563eb) 22%, transparent);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary, #2563eb);
  margin-bottom: 14px;
  transition: transform 0.2s ease, border-color 0.2s ease;
}

.login-card:hover .brand-icon-wrapper {
  transform: scale(1.04);
  border-color: var(--primary, #2563eb);
}

.brand-icon {
  width: 24px;
  height: 24px;
}

.brand-title {
  font-size: 17.5px;
  font-weight: 600;
  color: var(--ink, #1a1a1a);
  letter-spacing: -0.3px;
  margin: 0 0 6px 0;
}

.brand-sub {
  font-size: 12.5px;
  color: var(--steel, #787671);
  margin: 0;
  line-height: 1.4;
}

/* 错误提示条 */
.login-error-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  background: color-mix(in srgb, var(--error, #e03131) 8%, transparent);
  border: 1px solid color-mix(in srgb, var(--error, #e03131) 22%, transparent);
  color: var(--error, #e03131);
  padding: 8px 12px;
  border-radius: var(--r-md, 6px);
  font-size: 12.5px;
  font-weight: 500;
  margin-top: -6px;
  margin-bottom: 16px;
  line-height: 1.4;
}

.error-banner-icon {
  width: 15px;
  height: 15px;
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
  color: var(--charcoal, #37352f);
  letter-spacing: 0.2px;
}

.caps-hint {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11.5px;
  color: var(--warning, #d97706);
  background: color-mix(in srgb, var(--warning, #d97706) 10%, transparent);
  border: 1px solid color-mix(in srgb, var(--warning, #d97706) 22%, transparent);
  padding: 1px 6px;
  border-radius: var(--r-xs, 3px);
  font-weight: 500;
}

.caps-icon {
  width: 11px;
  height: 11px;
}

.password-input :deep(.el-input__wrapper) {
  padding: 2px 12px;
  border-radius: var(--r-md, 6px);
  background: var(--surface-card, #ffffff);
  box-shadow: 0 0 0 1px var(--hairline, #e5e3df) inset !important;
  transition: all 0.2s ease;
}

.password-input :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--hairline-strong, #c8c4be) inset !important;
}

.password-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1.5px var(--primary, #2563eb) inset, 0 0 0 3px color-mix(in srgb, var(--primary, #2563eb) 15%, transparent) !important;
  background: var(--surface-card, #ffffff);
}

.password-input :deep(.el-input__inner) {
  font-size: 14px;
  letter-spacing: 0.08em;
  color: var(--ink, #1a1a1a);
  height: 38px;
}

.password-input :deep(.el-input__inner::placeholder) {
  font-size: 13px;
  letter-spacing: 0;
  color: var(--stone, #a4a097);
}

.submit-form-item {
  margin-bottom: 0;
}

/* 提交按钮 */
.login-submit-btn {
  width: 100%;
  height: 40px;
  font-size: 13.5px;
  font-weight: 600;
  border-radius: var(--r-md, 6px);
  background: var(--primary, #2563eb);
  border: 1px solid var(--primary, #2563eb);
  color: var(--on-primary, #ffffff);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
  letter-spacing: 0.2px;
}

.login-submit-btn:hover {
  background: var(--primary-hover, #1d4ed8);
  border-color: var(--primary-hover, #1d4ed8);
}

.login-submit-btn:active {
  background: var(--primary-pressed, #1e40af);
  border-color: var(--primary-pressed, #1e40af);
  transform: scale(0.99);
}

.btn-inner {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.btn-arrow {
  width: 15px;
  height: 15px;
  transition: transform 0.2s ease;
}

.login-submit-btn:hover .btn-arrow {
  transform: translateX(3px);
}

/* 底部微信息 */
.login-footer {
  margin-top: 24px;
  font-size: 12px;
  color: var(--steel, #787671);
  text-align: center;
  user-select: none;
}

@media (max-width: 640px) {
  .login-container {
    padding: 16px 12px;
  }
  .login-card {
    padding: 28px 20px 24px;
    border-radius: var(--r-lg, 10px);
  }
  .password-input :deep(.el-input__inner) {
    font-size: 16px !important;
  }
}
</style>

import { defineStore } from "pinia";
import { ref } from "vue";

/**
 * 认证状态 store。
 * 从 router/index.ts 的模块级变量迁移而来，
 * 消除 router 和 login/setup 视图之间的紧耦合。
 */
export const useAuthStore = defineStore("auth", () => {
  const isAuthenticated = ref(false);
  const setupChecked = ref(false);
  const isFirstRun = ref(false);

  function markAuthenticated() {
    setupChecked.value = false; // 登录后重新检查首次运行状态
    isAuthenticated.value = true;
  }

  function markUnauthenticated() {
    isAuthenticated.value = false;
  }

  function markSetupChecked(firstRun: boolean) {
    setupChecked.value = true;
    isFirstRun.value = firstRun;
  }

  function reset() {
    isAuthenticated.value = false;
    setupChecked.value = false;
    isFirstRun.value = false;
  }

  return {
    isAuthenticated,
    setupChecked,
    isFirstRun,
    markAuthenticated,
    markUnauthenticated,
    markSetupChecked,
    reset,
  };
});

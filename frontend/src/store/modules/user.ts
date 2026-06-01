import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { post } from "@/utils/http";

export const useUserStore = defineStore("user", () => {
  const token = ref(localStorage.getItem("token") || "");
  const username = ref("Admin");

  const isLoggedIn = computed(() => !!token.value || document.cookie.includes("session"));

  function setToken(t: string) {
    token.value = t;
    localStorage.setItem("token", t);
  }

  function clearToken() {
    token.value = "";
    localStorage.removeItem("token");
  }

  async function login(password: string) {
    const res = await post("/login", { password });
    return res;
  }

  function logout() {
    clearToken();
    window.location.href = "/logout";
  }

  return { token, username, isLoggedIn, setToken, clearToken, login, logout };
});

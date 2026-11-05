import { defineStore } from "pinia";
import { ref } from "vue";

export const useAppStore = defineStore("app", () => {
  const sidebar = ref({ opened: true });

  const theme = ref<"light" | "dark" | "auto">(
    (typeof localStorage !== "undefined" && (localStorage.getItem("bhg_theme") as "light" | "dark" | "auto")) || "auto"
  );

  const isDark = ref(
    typeof document !== "undefined" ? document.documentElement.classList.contains("dark") : false
  );

  function applyTheme() {
    if (typeof window === "undefined" || typeof document === "undefined") return;
    const prefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
    const active = theme.value === "dark" || (theme.value === "auto" && prefersDark);
    isDark.value = active;
    if (active) {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
  }

  function toggleTheme() {
    if (isDark.value) {
      theme.value = "light";
    } else {
      theme.value = "dark";
    }
    if (typeof localStorage !== "undefined") {
      localStorage.setItem("bhg_theme", theme.value);
    }
    applyTheme();
  }

  function toggleSidebar() {
    sidebar.value.opened = !sidebar.value.opened;
  }

  if (typeof window !== "undefined") {
    window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", () => {
      if (theme.value === "auto") {
        applyTheme();
      }
    });
  }

  return { sidebar, toggleSidebar, theme, isDark, applyTheme, toggleTheme };
});

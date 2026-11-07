// Initialize theme before mounting to avoid flash
(function() {
  const saved = localStorage.getItem("bhg_theme") || "auto";
  const prefersDark = window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
  if (saved === "dark" || (saved === "auto" && prefersDark)) {
    document.documentElement.classList.add("dark");
  } else {
    document.documentElement.classList.remove("dark");
  }
})();
import { createApp } from "vue";
import { createPinia } from "pinia";
import router from "./router";
import App from "./App.vue";
import "element-plus/dist/index.css";
import "element-plus/theme-chalk/dark/css-vars.css";
import "./style/index.scss";

const app = createApp(App);

app.use(createPinia());
app.use(router);

app.mount("#app");

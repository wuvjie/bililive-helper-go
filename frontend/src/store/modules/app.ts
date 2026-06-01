import { defineStore } from "pinia";
import { ref } from "vue";

export const useAppStore = defineStore("app", () => {
  const sidebar = ref({
    opened: true,
    withoutAnimation: false
  });

  const device = ref<"desktop" | "mobile">("desktop");

  function toggleSidebar() {
    sidebar.value.opened = !sidebar.value.opened;
    sidebar.value.withoutAnimation = false;
  }

  function closeSidebar(withoutAnimation = false) {
    sidebar.value.opened = false;
    sidebar.value.withoutAnimation = withoutAnimation;
  }

  function toggleDevice(val: "desktop" | "mobile") {
    device.value = val;
  }

  return { sidebar, device, toggleSidebar, closeSidebar, toggleDevice };
});

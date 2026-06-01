import { reactive } from "vue";
import type { DrawerOptions } from "./type";

export const drawerStore = reactive<DrawerOptions[]>([]);

export function addDrawer(options: DrawerOptions) {
  drawerStore.push({
    ...options,
    visible: true
  });
}

export function closeDrawer(index: number) {
  drawerStore.splice(index, 1);
}

export function closeAllDrawer() {
  drawerStore.splice(0, drawerStore.length);
}

import { reactive } from "vue";
import type { DialogOptions } from "./type";

export const dialogStore = reactive<DialogOptions[]>([]);

export function addDialog(options: DialogOptions) {
  dialogStore.push({
    ...options,
    visible: true
  });
}

export function closeDialog(index: number) {
  dialogStore.splice(index, 1);
}

export function closeAllDialog() {
  dialogStore.splice(0, dialogStore.length);
}

export function updateDialog(index: number, options: Partial<DialogOptions>) {
  if (dialogStore[index]) {
    Object.assign(dialogStore[index], options);
  }
}

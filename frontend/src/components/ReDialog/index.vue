<template>
  <div v-for="(dialog, index) in dialogStore" :key="dialog.id">
    <el-dialog
      v-model="dialog.visible"
      :title="dialog.title"
      :width="dialog.width || '500px'"
      :fullscreen="dialog.fullscreen"
      :close-on-click-modal="dialog.closeOnClickModal ?? true"
      :close-on-press-escape="dialog.closeOnPressEscape ?? true"
      :draggable="dialog.draggable ?? true"
      :destroy-on-close="true"
      @close="handleClose(dialog, index)"
    >
      <component
        v-if="dialog.contentComponent"
        :is="dialog.contentComponent"
        v-bind="dialog.contentProps || {}"
      />
      <div v-else-if="dialog.contentHtml" v-html="dialog.contentHtml" />

      <template #footer v-if="dialog.showFooter !== false">
        <el-button
          v-if="dialog.showCancel !== false"
          @click="handleCancel(dialog, index)"
        >
          {{ dialog.cancelText || "取消" }}
        </el-button>
        <el-button
          type="primary"
          :loading="dialog.sureLoading"
          @click="handleSure(dialog, index)"
        >
          {{ dialog.sureText || "确定" }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { dialogStore, closeDialog } from "./index";
import type { DialogOptions } from "./type";

function handleClose(dialog: DialogOptions, index: number) {
  dialog.closeCallback?.();
  closeDialog(index);
}

function handleCancel(dialog: DialogOptions, index: number) {
  dialog.beforeCancel?.();
  closeDialog(index);
}

function handleSure(dialog: DialogOptions, index: number) {
  dialog.beforeSure?.();
}
</script>

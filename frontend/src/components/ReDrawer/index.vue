<template>
  <div v-for="(drawer, index) in drawerStore" :key="drawer.id">
    <el-drawer
      v-model="drawer.visible"
      :title="drawer.title"
      :size="drawer.size || '400px'"
      :direction="drawer.direction || 'rtl'"
      :destroy-on-close="true"
      @close="handleClose(drawer, index)"
    >
      <component
        v-if="drawer.contentComponent"
        :is="drawer.contentComponent"
        v-bind="drawer.contentProps || {}"
      />
      <div v-else-if="drawer.contentHtml" v-html="drawer.contentHtml" />

      <template #footer v-if="drawer.showFooter !== false">
        <el-button @click="handleClose(drawer, index)">
          {{ drawer.cancelText || "关闭" }}
        </el-button>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { drawerStore, closeDrawer } from "./index";
import type { DrawerOptions } from "./type";

function handleClose(drawer: DrawerOptions, index: number) {
  drawer.closeCallback?.();
  closeDrawer(index);
}
</script>

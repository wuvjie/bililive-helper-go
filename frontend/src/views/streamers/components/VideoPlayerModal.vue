<template>
  <el-dialog
    :model-value="visible"
    :title="`播放预览: ${fileName}`"
    width="800px"
    destroy-on-close
    class="player-dialog"
    @update:model-value="$emit('update:visible', $event)"
  >
    <div class="player-wrap">
      <video
        v-if="url"
        :src="url"
        controls
        autoplay
        preload="metadata"
        class="video-element"
      >
        您的浏览器不支持视频播放。
      </video>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
defineProps<{
  visible: boolean;
  fileName: string;
  url: string;
}>();

defineEmits<{
  (e: "update:visible", val: boolean): void;
}>();
</script>

<style scoped>
.player-wrap {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  background: #000000;
  border-radius: var(--r-sm);
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.video-element {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
</style>

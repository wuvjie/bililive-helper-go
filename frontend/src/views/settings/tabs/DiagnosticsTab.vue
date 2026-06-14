<template>
  <div class="diag-table" v-loading="loading">
    <div class="diag-row">
      <span class="diag-label">FFmpeg</span>
      <span class="diag-val"
        ><span class="status-dot-sm" :class="data?.ffmpeg_ok ? 'dot-ok' : 'dot-err'" />{{
          data?.ffmpeg_ok ? "正常" : "异常"
        }}<span class="diag-path">{{ data?.ffmpeg_path }}</span></span
      >
    </div>
    <div class="diag-row">
      <span class="diag-label">FFprobe</span>
      <span class="diag-val"
        ><span class="status-dot-sm" :class="data?.ffprobe_ok ? 'dot-ok' : 'dot-err'" />{{ data?.ffprobe_ok ? "正常" : "异常" }}</span
      >
    </div>
    <div class="diag-row">
      <span class="diag-label">目标目录存在</span>
      <span class="diag-val"
        ><span class="status-dot-sm" :class="data?.target_dir_exists ? 'dot-ok' : 'dot-err'" />{{ data?.target_dir_exists ? "是" : "否" }}</span
      >
    </div>
    <div class="diag-row">
      <span class="diag-label">目录可写</span>
      <span class="diag-val"
        ><span class="status-dot-sm" :class="data?.target_dir_writable ? 'dot-ok' : 'dot-err'" />{{ data?.target_dir_writable ? "是" : "否" }}</span
      >
    </div>
    <div class="diag-row">
      <span class="diag-label">主播数</span>
      <span class="mono-val">{{ data?.streamer_count || 0 }}</span>
    </div>
    <div class="diag-row">
      <span class="diag-label">视频数</span>
      <span class="mono-val">{{ data?.video_count || 0 }}</span>
    </div>
    <div class="diag-row">
      <span class="diag-label">总大小</span>
      <span class="mono-val">{{ data?.total_size_gb?.toFixed(2) }} GB</span>
    </div>
    <div class="diag-row">
      <span class="diag-label">磁盘总容量</span>
      <span class="mono-val">{{ data?.disk_total_gb?.toFixed(1) }} GB</span>
    </div>
    <div class="diag-row">
      <span class="diag-label">磁盘剩余</span>
      <span class="mono-val">{{ data?.disk_free_gb?.toFixed(1) }} GB</span>
    </div>
    <div class="diag-row diag-row-last">
      <span class="diag-label">磁盘使用率</span>
      <div class="diag-progress">
        <el-progress :percentage="data?.disk_usage_pct || 0" :stroke-width="6" :format="() => ''" style="flex: 1" />
        <span class="mono-val diag-pct">{{ (data?.disk_usage_pct || 0).toFixed(1) }}%</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SetupCheck } from "@/api/types";

defineProps<{
  data: SetupCheck | undefined;
  loading: boolean;
}>();
</script>

<style scoped>
.diag-table {
  padding: 12px 16px;
}
.diag-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 0;
  border-bottom: 1px solid #f5f5f5;
}
.diag-row-last {
  border-bottom: none;
}
.diag-label {
  color: #888;
  font-size: 13px;
  flex-shrink: 0;
}
.diag-val {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}
.diag-path {
  color: #aaa;
  font-size: 12px;
  margin-left: 4px;
}
.mono-val {
  font-family: var(--font-mono, "JetBrains Mono", monospace);
  font-size: 13px;
  font-weight: 500;
}
.status-dot-sm {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
}
.dot-ok {
  background: #448361;
}
.dot-err {
  background: #e03131;
}
.diag-progress {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  max-width: 260px;
}
.diag-pct {
  width: 50px;
  text-align: right;
}
</style>

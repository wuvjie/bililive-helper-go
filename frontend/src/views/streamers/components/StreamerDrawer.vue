<template>
  <el-drawer
    :model-value="visible"
    :title="`录像文件清单 · ${streamer}`"
    size="840px"
    destroy-on-close
    @update:model-value="$emit('update:visible', $event)"
  >
    <div v-loading="loading" class="drawer-content">
      <div class="drawer-header-bar">
        <UnderlineNav
          v-model="activeTab"
          :items="navItems"
        />

        <div class="drawer-header-actions">
          <button
            class="btn-drawer-merge"
            :disabled="loading || files.length === 0 || isMerging"
            @click="handleMergeClick"
            title="一键触发合并该主播名下的全部录像分片"
          >
            <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="18" cy="18" r="3"/>
              <circle cx="6" cy="6" r="3"/>
              <path d="M6 21V9a9 9 0 0 0 9 9"/>
            </svg>
            <span>合并该主播</span>
          </button>
        </div>
      </div>

      <div class="table-responsive-wrapper">
        <el-table
          :data="displayFiles"
          :tooltip-options="{ enterable: false, showAfter: 800, placement: 'top' }"
          size="small"
          max-height="650"
          class="drawer-table"
          :default-sort="{ prop: 'name', order: 'descending' }"
        >
          <el-table-column
            prop="name"
            label="文件名"
            min-width="300"
            resizable
            align="left"
            header-align="left"
            sortable
            :sort-method="(a: any, b: any) => (a.mtime || 0) - (b.mtime || 0)"
          >
            <template #default="{ row }">
              <div class="drawer-file-col">
                <span class="drawer-file-name" :title="row.name">{{ row.name }}</span>
                <el-tooltip
                  v-if="row.mtime"
                  :content="formatExactTime(row.mtime)"
                  placement="top"
                  :show-after="400"
                  :enterable="false"
                >
                  <span class="drawer-file-time">{{ formatRelativeTime(row.mtime, tick) }}</span>
                </el-tooltip>
              </div>
            </template>
          </el-table-column>

          <el-table-column
            prop="size"
            label="大小"
            width="110"
            resizable
            align="right"
            header-align="right"
            sortable
            :sort-method="(a: any, b: any) => a.size - b.size"
          >
            <template #default="{ row }">
              <span class="mono-val">{{ row.size_str }}</span>
            </template>
          </el-table-column>

          <el-table-column
            prop="is_merged"
            label="状态"
            width="100"
            resizable
            align="center"
            header-align="center"
            sortable
          >
            <template #default="{ row }">
              <el-tooltip :content="row.is_merged ? '已通过系统合并处理的完整录播视频' : '录播工具生成的原始录像文件（待合并或独立单场录像）'" placement="top" :show-after="500">
                <span class="file-status-badge" :class="row.is_merged ? 'status-merged' : 'status-raw'">
                  <span class="badge-dot"></span>
                  {{ row.is_merged ? '已合并' : '原始录像' }}
                </span>
              </el-tooltip>
            </template>
          </el-table-column>

          <el-table-column
            label="操作"
            width="140"
            resizable
            align="center"
            header-align="center"
          >
            <template #default="{ row }">
              <div class="drawer-file-actions">
                <button class="play-btn-sm" @click="$emit('play', row)" title="在浏览器中播放预览">
                  <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg>
                  <span>播放</span>
                </button>
                <button class="play-btn-sm btn-del-video" @click="$emit('delete', row)" title="从磁盘永久删除该录像文件">
                  <svg class="btn-svg-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                  <span>删除</span>
                </button>
              </div>
            </template>
          </el-table-column>

          <template #empty>
            <div class="drawer-empty-box">
              <div class="empty-icon-wrapper">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6">
                  <rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"/>
                  <line x1="7" y1="2" x2="7" y2="22"/>
                  <line x1="17" y1="2" x2="17" y2="22"/>
                  <line x1="2" y1="12" x2="22" y2="12"/>
                </svg>
              </div>
              <span class="empty-title">暂无录像文件</span>
              <span class="empty-desc">该主播录像目录暂未生成切片或录像文件</span>
            </div>
          </template>
        </el-table>
      </div>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import type { StreamerFile } from "@/api/types";
import UnderlineNav, { type UnderlineNavItem } from "@/components/common/UnderlineNav.vue";

const activeTab = ref<string>("all");
const isMerging = ref(false);

function handleMergeClick() {
  if (isMerging.value) return;
  isMerging.value = true;
  emit("merge", props.streamer);
  setTimeout(() => {
    isMerging.value = false;
  }, 1000);
}

const props = defineProps<{
  visible: boolean;
  streamer: string;
  files: StreamerFile[];
  loading: boolean;
  tick: number;
}>();

const emit = defineEmits<{
  (e: "update:visible", val: boolean): void;
  (e: "play", file: StreamerFile): void;
  (e: "delete", file: StreamerFile): void;
  (e: "merge", streamer: string): void;
}>();

const mergedCount = computed(() => props.files.filter(f => f.is_merged).length);
const rawCount = computed(() => props.files.filter(f => !f.is_merged).length);

const navItems = computed<UnderlineNavItem[]>(() => [
  { label: "全部", value: "all", count: props.files.length },
  { label: "已合并", value: "merged", count: mergedCount.value },
  { label: "原始录像", value: "raw", count: rawCount.value }
]);

const displayFiles = computed(() => {
  if (activeTab.value === "merged") return props.files.filter(f => f.is_merged);
  if (activeTab.value === "raw") return props.files.filter(f => !f.is_merged);
  return props.files;
});

function formatExactTime(ts: number): string {
  const d = new Date(ts * 1000);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

function formatDisplayDateTime(ts: number): string {
  if (!ts || isNaN(ts)) return "-";
  const d = new Date(ts * 1000);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function formatRelativeTime(ts: number, _tick?: number): string {
  const diff = Date.now() / 1000 - ts;
  if (diff < 0) return "刚刚";
  if (diff < 60) return "刚刚";
  if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`;
  if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`;
  const days = Math.floor(diff / 86400);
  if (days < 30) return `${days} 天前`;
  const d = new Date(ts * 1000);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}
</script>

<style scoped>
.drawer-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 200px;
}

.drawer-header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 4px;
  border-bottom: 1px solid var(--hairline);
  margin-bottom: -4px;
}

.btn-drawer-merge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 28px;
  padding: 0 8px;
  border-radius: var(--r-sm);
  font-size: 12px;
  font-weight: 500;
  background: transparent;
  border: none;
  color: var(--primary);
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-drawer-merge:hover:not(:disabled) {
  background: color-mix(in srgb, var(--primary) 8%, transparent);
  color: var(--primary-hover, #1d4ed8);
}

.btn-drawer-merge:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.drawer-sub {
  font-size: 13px;
  color: var(--slate);
  font-weight: 500;
}

.drawer-file-col {
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding: 2px 0;
}

.drawer-file-name {
  font-size: 12.5px;
  font-weight: 500;
  color: var(--charcoal);
  word-break: break-all;
  line-height: 1.4;
}

.drawer-file-time {
  font-size: 11px;
  color: var(--steel);
  cursor: help;
  display: inline-block;
  width: fit-content;
}

.file-status-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: var(--r-full);
  font-size: 11px;
  font-weight: 500;
  line-height: 1.4;
}

.badge-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: currentColor;
}

.status-merged {
  background: rgba(16, 185, 129, 0.1);
  color: #059669;
}

.status-raw {
  background: rgba(148, 163, 184, 0.12);
  color: #64748b;
}

.drawer-file-actions {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.play-btn-sm {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 2px 7px;
  border-radius: var(--r-xs);
  font-size: 11px;
  font-weight: 500;
  border: 1px solid var(--hairline);
  background: var(--surface);
  color: var(--slate);
  cursor: pointer;
  transition: all 0.15s ease;
}

.play-btn-sm:hover {
  background: var(--highlight);
  border-color: var(--charcoal);
  color: var(--ink);
}

.btn-del-video:hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.4);
  color: var(--error);
}

.btn-svg-sm {
  width: 12px;
  height: 12px;
}

.drawer-empty-box {
  padding: 40px 16px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
</style>

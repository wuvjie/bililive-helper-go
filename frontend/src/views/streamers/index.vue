<template>
  <div class="streamers">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="card-title">👥 主播管理</span>
          <div class="header-right">
            <el-input
              v-model="searchQuery"
              placeholder="搜索主播..."
              clearable
              size="small"
              style="width: 200px"
              :prefix-icon="Search"
            />
            <el-button text size="small" class="goto-tasks-btn" @click="$router.push('/tasks')">
              前往任务中心 <span class="arrow-ico">↗</span>
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="filteredStreamers" v-loading="loading" empty-text="暂无主播数据" class="helper-clean-table">
        <el-table-column type="index" label="#" width="64" align="right" header-align="right">
          <template #default="{ $index }">
            <span class="idx-num">{{ $index + 1 }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="name" label="主播名称" min-width="200" show-overflow-tooltip sortable>
          <template #default="{ row }">
            <span class="streamer-name">{{ row.name }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="files" label="录制文件数" width="150" align="right" header-align="right" sortable>
          <template #default="{ row }">
            <div class="data-block">
              <span class="mono-val">{{ row.files }}</span><span class="unit-lbl">个</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="size_gb" label="磁盘占用" width="150" align="right" header-align="right" sortable>
          <template #default="{ row }">
            <div class="data-block">
              <span class="mono-val">{{ row.size_gb?.toFixed(2) }}</span><span class="unit-lbl">GB</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="100" align="right" header-align="right">
          <template #default="{ row }">
            <button class="ghost-action" @click="goToTasks(row.name)">管理文件</button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onActivated } from "vue";
import { useRouter } from "vue-router";
import { getStreamers } from "@/api/status";
import { Search } from "@element-plus/icons-vue";
import type { StreamerInfo } from "@/api/types";

const router = useRouter();
const streamers = ref<StreamerInfo[]>([]);
const loading = ref(true);
const searchQuery = ref("");

const filteredStreamers = computed(() => {
  if (!searchQuery.value) return streamers.value;
  const q = searchQuery.value.toLowerCase();
  return streamers.value.filter(s => s.name.toLowerCase().includes(q));
});

function goToTasks(name: string) {
  router.push({ path: "/tasks", query: { streamer: name } });
}

onMounted(async () => {
  try {
    streamers.value = await getStreamers();
  } finally {
    loading.value = false;
  }
});

// Refresh data when component is re-activated by keep-alive router-view
onActivated(async () => {
  try {
    streamers.value = await getStreamers();
  } catch { /* ignore */ }
});
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.card-title {
  font-weight: 600;
  color: var(--ink);
}
.header-right {
  display: flex;
  gap: 12px;
  align-items: center;
}
.goto-tasks-btn {
  color: var(--stone) !important;
  font-weight: 500 !important;
  padding-right: 0 !important;
}
.goto-tasks-btn:hover {
  color: var(--ink) !important;
}
.arrow-ico {
  font-size: 11px;
  opacity: 0.7;
  margin-left: 2px;
}

.streamer-name {
  font-weight: 500;
  color: var(--ink);
}
.idx-num { font-family: var(--font-mono); font-size: 13px; color: var(--stone); }

.data-block {
  display: block;
  width: 100%;
  text-align: right;
}
.mono-val { font-family: var(--font-mono); font-size: 13px; color: var(--ink); font-weight: 500; }
.unit-lbl { font-size: 11px; color: var(--stone); margin-left: 2px; }

.ghost-action {
  color: var(--stone);
  background: transparent;
  border: none;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: color 0.15s;
  padding: 0;
  margin: 0;
}
.ghost-action:hover {
  color: var(--ink);
}

/* Header right-aligned columns — flex right flow, sort arrow pushed to end */
:deep(.helper-clean-table) th.el-table__cell.is-right .cell {
  display: inline-flex !important;
  align-items: center !important;
  justify-content: flex-end !important;
  padding-right: 0 !important;
}

/* Sort arrow always on the right side of header text */
:deep(.helper-clean-table) th.el-table__cell.is-right .cell .caret-wrapper {
  order: 2 !important;
  margin-left: 4px !important;
  margin-right: 0 !important;
}

/* Header text flows right */
:deep(.helper-clean-table) th.el-table__cell.is-right .cell span:not(.caret-wrapper) {
  order: 1 !important;
  text-align: right !important;
}

/* Data cells — clean right alignment */
:deep(.helper-clean-table) td.el-table__cell.is-right .cell,
:deep(.helper-clean-table) td.el-table__cell:last-child .cell {
  padding-right: 0 !important;
  text-align: right !important;
}
</style>

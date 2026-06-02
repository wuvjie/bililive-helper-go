<template>
  <div class="streamers">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>👥 主播管理</span>
          <div class="header-right">
            <el-input
              v-model="searchQuery"
              placeholder="搜索主播..."
              clearable
              size="small"
              style="width: 200px"
              :prefix-icon="Search"
            />
            <el-button text size="small" @click="$router.push('/tasks')">
              <el-icon><VideoPlay /></el-icon>前往任务中心
            </el-button>
          </div>
        </div>
      </template>
      <el-table :data="filteredStreamers" v-loading="loading" empty-text="暂无主播数据">
        <el-table-column type="index" label="#" width="72" align="right" header-align="right">
          <template #default="{ $index }">
            <span class="idx-num">{{ $index + 1 }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="主播名称" min-width="200" show-overflow-tooltip sortable>
          <template #default="{ row }">
            <span class="streamer-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="files" label="录制文件数" width="120" align="center" header-align="center" sortable>
          <template #default="{ row }">
            <span class="mono-val">{{ row.files }}</span><span class="unit-lbl">&nbsp;个</span>
          </template>
        </el-table-column>
        <el-table-column prop="size_gb" label="磁盘占用" width="140" align="right" header-align="right" sortable>
          <template #default="{ row }">
            <span class="mono-val">{{ row.size_gb?.toFixed(2) }}</span><span class="unit-lbl"> GB</span>
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
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import { getStreamers } from "@/api/status";
import { VideoPlay, Search } from "@element-plus/icons-vue";
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
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header-right { display: flex; gap: 8px; align-items: center; }
.streamer-name {
  font-weight: 500; color: var(--ink);
  display: inline-flex; align-items: center;
  line-height: 1;
}
.idx-num { font-family: var(--font-mono); font-size: 13px; color: var(--stone); }
.mono-val { font-family: var(--font-mono); font-size: 13px; color: var(--ink); }
.unit-lbl { font-size: 12px; color: var(--stone); margin-left: 4px; }
.ghost-action {
  color: var(--stone);
  background: transparent;
  border: none;
  border-bottom: 1px solid transparent;
  padding: 2px 2px 3px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
  border-radius: 0;
}
.ghost-action:hover {
  color: var(--ink);
  border-bottom-color: var(--ink);
}
/* Sort indicator alignment — ensure header text stays visually centered */
:deep(.el-table__header) th .cell {
  display: flex;
  align-items: center;
  gap: 4px;
}
</style>

<template>
  <div class="streamers">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>👥 主播管理</span>
          <el-button type="primary" size="small" @click="$router.push('/tasks')">
            <el-icon><VideoPlay /></el-icon>前往任务中心
          </el-button>
        </div>
      </template>
      <el-table :data="streamers" stripe v-loading="loading" empty-text="暂无主播数据">
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="name" label="主播名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="streamer-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="files" label="录制文件数" width="120" align="center">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.files }} 个</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="size_gb" label="磁盘占用" width="120" align="center">
          <template #default="{ row }">
            {{ row.size_gb?.toFixed(2) }} GB
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" align="center">
          <template #default="{ row }">
            <el-button
              type="primary"
              link
              size="small"
              @click="goToTasks(row.name)"
            >
              管理文件
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { getStreamers } from "@/api/status";
import { VideoPlay } from "@element-plus/icons-vue";
import type { StreamerInfo } from "@/api/types";

const router = useRouter();
const streamers = ref<StreamerInfo[]>([]);
const loading = ref(true);

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
.streamer-name { font-weight: 500; }
</style>

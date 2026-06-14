import { ref, computed } from "vue";
import { getStatusDetail } from "@/api/status";

/** 磁盘使用状态（layout + dashboard 共享） */
const diskUsedPct = ref(0);
const loaded = ref(false);

export function useDiskStatus() {
  async function fetchDiskStatus() {
    try {
      const data = await getStatusDetail();
      if (data?.disk) {
        diskUsedPct.value = data.disk.used_pct ?? 0;
        loaded.value = true;
      }
    } catch {
      // 静默失败，磁盘状态非关键路径
    }
  }

  const alertLevel = computed(() => {
    if (diskUsedPct.value >= 95) return "critical";
    if (diskUsedPct.value >= 90) return "warning";
    return "none";
  });

  return {
    diskUsedPct,
    loaded,
    alertLevel,
    fetchDiskStatus,
  };
}

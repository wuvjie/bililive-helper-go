import { get } from "@/utils/http";
import type { HistoryResponse, HistoryRecord } from "./types";

export function getHistory(params: { page?: number; per_page?: number; task?: string }) {
  return get<HistoryResponse>("/history", params);
}

export function exportHistory() {
  return get<HistoryRecord[]>("/history/export");
}

export function getLogContent(task: string, logId: string) {
  return get<string>(`/logs/content/${task}`, { log_id: logId });
}

import { get, post } from "@/utils/http";
import type { HistoryResponse } from "./types";

export function getHistory(params: { page?: number; per_page?: number; task?: string; streamer?: string }) {
  return get<HistoryResponse>("/history", params);
}

export function getLogContent(task: string, logId: string) {
  return get<{ content: string }>(`/logs/content/${task}`, { log_id: logId });
}

export function clearHistory() {
  return post<{ cleared: boolean }>("/history/clear");
}

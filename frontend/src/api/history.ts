import { get, post, del } from "@/utils/http";
import type { HistoryResponse, HistoryRecord } from "./types";

export function getHistory(params: { page?: number; per_page?: number; task?: string }) {
  return get<HistoryResponse>("/history", params);
}

export function exportHistory() {
  return get<HistoryRecord[]>("/history/export");
}

export function getLogContent(task: string, logId: string) {
  return get<{ content: string }>(`/logs/content/${task}`, { log_id: logId });
}

export function getRawLogDownloadUrl(task: string, logId: string): string {
  return `/api/logs/content/${encodeURIComponent(task)}?log_id=${encodeURIComponent(logId)}&raw=true`;
}
export function deleteHistory(id: string) {
  return del<{ deleted: string }>(`/history/${encodeURIComponent(id)}`);
}

export function clearHistory() {
  return post<{ cleared: boolean }>("/history/clear");
}

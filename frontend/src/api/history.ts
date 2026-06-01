import { get } from "@/utils/http";
import type { HistoryResponse, HistoryRecord, LogFile } from "./types";

export function getHistory(params: { page?: number; per_page?: number; task?: string }) {
  return get<HistoryResponse>("/history", params);
}

export function exportHistory() {
  return get<HistoryRecord[]>("/history/export");
}

export function getLogList(task: string) {
  return get<LogFile[]>(`/logs/list/${task}`);
}

export function getLogContent(task: string, file: string) {
  return get<string>(`/logs/content/${task}`, { file });
}

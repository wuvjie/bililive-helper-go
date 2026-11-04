import { get, post, del } from "@/utils/http";
import type { DetailResponse, StatsResponse, StreamerInfo, StreamerFile, StreamerRule } from "./types";

export function getStatusDetail() {
  return get<DetailResponse>("/status/detail");
}

export function getStats() {
  return get<StatsResponse>("/stats");
}

export function getStreamers() {
  return get<StreamerInfo[]>("/streamers");
}

export function getStreamerFiles(name: string) {
  return get<StreamerFile[]>(`/streamers/${encodeURIComponent(name)}/files`, undefined, { timeout: 30000 });
}

export function getStreamerRule(name: string) {
  return get<{ streamer: string; rule: StreamerRule; exists: boolean }>(`/streamers/${encodeURIComponent(name)}/rule`);
}

export function saveStreamerRule(name: string, rule: StreamerRule) {
  return post<{ streamer: string; rule: StreamerRule }>(`/streamers/${encodeURIComponent(name)}/rule`, rule);
}

export function getVideoPlayUrl(streamer: string, file: string) {
  return `/api/streamers/${encodeURIComponent(streamer)}/play?file=${encodeURIComponent(file)}`;
}
export function deleteStreamerFile(name: string, file: string) {
  return del<{ deleted: string; freed_bytes: number; freed_str: string }>(
    `/streamers/${encodeURIComponent(name)}/files`,
    { file }
  );
}

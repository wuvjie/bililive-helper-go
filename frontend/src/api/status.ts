import { get } from "@/utils/http";
import type { DetailResponse, StatsResponse, StreamerInfo, StreamerFile } from "./types";

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
  return get<StreamerFile[]>(`/streamers/${encodeURIComponent(name)}/files`);
}

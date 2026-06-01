import { get, post } from "@/utils/http";
import type { StatusResponse, DetailResponse, StatsResponse, StreamerInfo, StreamerFile } from "./types";

export function getStatus() {
  return get<StatusResponse>("/status");
}

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

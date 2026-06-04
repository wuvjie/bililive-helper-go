import { get, post } from "@/utils/http";
import type { ConfigDTO, ConfigRecommend } from "./types";

export function getConfig() {
  return get<ConfigDTO>("/config");
}

export function saveConfig(data: Record<string, any>) {
  return post("/config", data);
}

export function getConfigRecommend() {
  return get<ConfigRecommend>("/config/recommend", undefined, { timeout: 60000 });
}

export function getConfigExport() {
  return get("/config/export");
}

export function importConfig(data: Record<string, any>) {
  return post("/config/import", data);
}

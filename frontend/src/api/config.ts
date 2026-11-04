import { get, post } from "@/utils/http";
import type { ConfigDTO } from "./types";

export function getConfig() {
  return get<ConfigDTO>("/config");
}

export function saveConfig(data: Partial<ConfigDTO>) {
  return post("/config", data);
}

export function getConfigExport() {
  return get("/config/export");
}

export function importConfig(data: { config?: unknown; schedule?: unknown; history?: unknown }) {
  return post("/config/import", data);
}

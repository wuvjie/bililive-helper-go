import { get, post } from "@/utils/http";
import type { SetupCheck } from "./types";

export function setupCheck() {
  return get<SetupCheck>("/setup/check");
}

export function setupStatus() {
  return get<{ first_run: boolean; log_dir: string }>("/setup/status");
}

export function setupInit(data: { password: string; target_dir: string; log_dir: string }) {
  return post<{ message: string }>("/setup/init", data);
}

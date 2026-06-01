import { get, post } from "@/utils/http";
import type { CleanEstimate } from "./types";

export function runMerge(body: { streamer: string; files: string[] }) {
  return post("/merge/manual", body);
}

export function runClean(body: { streamer: string }) {
  return post("/clean", body);
}

export function getCleanEstimate() {
  return get<CleanEstimate>("/clean/estimate");
}

import { get } from "@/utils/http";
import type { CleanEstimate } from "./types";

export function getCleanEstimate() {
  return get<CleanEstimate>("/clean/estimate");
}

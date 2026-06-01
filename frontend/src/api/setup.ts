import { get, post } from "@/utils/http";
import type { SetupCheck } from "./types";

export function setupCheck() {
  return get<SetupCheck>("/setup/check");
}

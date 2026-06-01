import { get, post } from "@/utils/http";
import type { ScheduleStatus, ScheduleConfig } from "./types";

export function getSchedule() {
  return get<ScheduleStatus>("/schedule");
}

export function saveSchedule(data: ScheduleConfig) {
  return post("/schedule", data);
}

export function runTask(task: string) {
  return post(`/schedule/run/${task}`);
}

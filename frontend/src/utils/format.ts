import type { HistoryRecord } from "@/api/types";

/**
 * Shared formatting utilities used across the application.
 */

/** Format bytes to human-readable string (KB, MB, GB) */
export function formatBytes(b: number): string {
  if (!b) return "0 B";
  if (b >= 1024 ** 3) return `${(b / 1024 ** 3).toFixed(2)} GB`;
  if (b >= 1024 ** 2) return `${(b / 1024 ** 2).toFixed(1)} MB`;
  return `${(b / 1024).toFixed(0)} KB`;
}

/** Format Unix timestamp (seconds) to locale string */
export function formatTime(ts?: number): string {
  if (!ts) return "-";
  return new Date(ts * 1000).toLocaleString("zh-CN");
}

/** Format Unix timestamp (seconds) to YYYY-MM-DD HH:mm */
export function formatDisplayDateTime(ts?: number): string {
  if (!ts || isNaN(ts)) return "-";
  const d = new Date(ts * 1000);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

/** Format Unix timestamp (seconds) to YYYY-MM-DD HH:mm:ss */
export function formatExactTime(ts?: number): string {
  if (!ts || isNaN(ts)) return "-";
  const d = new Date(ts * 1000);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

/** Format Unix timestamp (seconds) to relative time (e.g. "刚刚", "5 分钟前", "2 小时前") */
export function formatRelativeTime(ts?: number, _tick?: number): string {
  if (!ts || isNaN(ts)) return "-";
  const diff = Date.now() / 1000 - ts;
  if (diff < 0) return "刚刚";
  if (diff < 60) return "刚刚";
  if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`;
  if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`;
  const days = Math.floor(diff / 86400);
  if (days < 30) return `${days} 天前`;
  const d = new Date(ts * 1000);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

/** Clean raw detail text of operations log for concise display */
export function cleanDetailText(raw: string, _task?: string): string {
  if (!raw) return "-";
  let s = raw.trim();
  if (s.startsWith("调度已保存:") || s.startsWith("调度已保存：")) {
    const params = s.replace(/^调度已保存[:：]\s*/, "");
    return `重新应用当前调度（${params}，无参数变动）`;
  }
  if (s === "配置未改动" || s === "无变更") {
    return "重新应用当前配置 (无参数改动)";
  }
  s = s.replace(/^(调度变更|配置变更|系统配置变更|合并任务|清理任务)[:：]\s*/, "");
  s = s.replace(/([^\s]+)\s*→\s*([^\s]+)/g, "$1 → $2");
  s = s.replace(/,\s*/g, "，");
  return s;
}

/** Format fallback detail for a history record when raw detail is empty */
export function formatHistoryDetail(row: HistoryRecord): string {
  if (row.detail) return row.detail;
  const p: string[] = [];
  if (row.task === "merge") {
    if (row.files_count) p.push(`合并 ${row.files_count} 个分片`);
    if (row.merged_bytes) p.push(formatBytes(row.merged_bytes));
  } else if (row.task === "clean") {
    if (row.files_count) p.push(`清理 ${row.files_count} 个文件`);
    if (row.freed_bytes) p.push(`释放 ${formatBytes(row.freed_bytes)}`);
  } else {
    if (row.files_count) p.push(`${row.files_count} 个文件`);
  }
  if (row.duration) p.push(`${row.duration} 秒`);
  return p.join(" · ") || "已完成";
}

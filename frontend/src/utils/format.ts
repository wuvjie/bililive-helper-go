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

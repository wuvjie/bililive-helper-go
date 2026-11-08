import { describe, it, expect } from "vitest";
import {
  formatBytes,
  formatTime,
  formatDisplayDateTime,
  formatExactTime,
  formatRelativeTime,
  cleanDetailText,
  formatHistoryDetail
} from "../format";
import type { HistoryRecord } from "@/api/types";

describe("format utilities", () => {
  describe("formatBytes", () => {
    it("formats 0 bytes as 0 B", () => {
      expect(formatBytes(0)).toBe("0 B");
    });

    it("formats bytes, MB, and GB accurately", () => {
      expect(formatBytes(1024)).toBe("1 KB");
      expect(formatBytes(1024 * 1024 * 5)).toBe("5.0 MB");
      expect(formatBytes(1024 * 1024 * 1024 * 2.5)).toBe("2.50 GB");
    });
  });

  describe("formatDisplayDateTime", () => {
    it("returns - for invalid or empty timestamps", () => {
      expect(formatDisplayDateTime(0)).toBe("-");
      expect(formatDisplayDateTime(undefined)).toBe("-");
      expect(formatDisplayDateTime(NaN)).toBe("-");
    });

    it("formats timestamp as YYYY-MM-DD HH:mm", () => {
      const fixedTs = 1762557000;
      const formatted = formatDisplayDateTime(fixedTs);
      expect(formatted).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/);
    });
  });

  describe("formatExactTime", () => {
    it("returns - for invalid or empty timestamps", () => {
      expect(formatExactTime(0)).toBe("-");
      expect(formatExactTime(undefined)).toBe("-");
      expect(formatExactTime(NaN)).toBe("-");
    });

    it("formats timestamp as YYYY-MM-DD HH:mm:ss", () => {
      const fixedTs = 1762557000;
      const formatted = formatExactTime(fixedTs);
      expect(formatted).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/);
    });
  });

  describe("formatTime", () => {
    it("returns - for falsy timestamps", () => {
      expect(formatTime(0)).toBe("-");
      expect(formatTime(undefined)).toBe("-");
    });
  });

  describe("formatRelativeTime", () => {
    it("returns - for falsy or invalid timestamps", () => {
      expect(formatRelativeTime(0)).toBe("-");
      expect(formatRelativeTime(undefined)).toBe("-");
      expect(formatRelativeTime(NaN)).toBe("-");
    });

    it("returns '刚刚' for recent or future timestamps", () => {
      const nowSec = Math.floor(Date.now() / 1000);
      expect(formatRelativeTime(nowSec - 10)).toBe("刚刚");
      expect(formatRelativeTime(nowSec + 100)).toBe("刚刚");
    });

    it("formats relative minutes, hours, and days", () => {
      const nowSec = Math.floor(Date.now() / 1000);
      expect(formatRelativeTime(nowSec - 300)).toBe("5 分钟前");
      expect(formatRelativeTime(nowSec - 7200)).toBe("2 小时前");
      expect(formatRelativeTime(nowSec - 86400 * 3)).toBe("3 天前");
    });

    it("formats full date for timestamps older than 30 days", () => {
      const nowSec = Math.floor(Date.now() / 1000);
      const oldTs = nowSec - 86400 * 45;
      expect(formatRelativeTime(oldTs)).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    });
  });

  describe("cleanDetailText", () => {
    it("returns - for empty or falsy strings", () => {
      expect(cleanDetailText("")).toBe("-");
    });

    it("cleans scheduled save messages with no changes", () => {
      expect(cleanDetailText("调度已保存: 周期=60")).toBe("重新应用当前调度（周期=60，无参数变动）");
      expect(cleanDetailText("调度已保存：周期=120")).toBe("重新应用当前调度（周期=120，无参数变动）");
    });

    it("normalizes no-change config messages", () => {
      expect(cleanDetailText("配置未改动")).toBe("重新应用当前配置 (无参数改动)");
      expect(cleanDetailText("无变更")).toBe("重新应用当前配置 (无参数改动)");
    });

    it("strips prefixes and formats punctuation and arrows", () => {
      expect(cleanDetailText("配置变更: 阈值 80 → 85, 周期 30")).toBe("阈值 80 → 85，周期 30");
      expect(cleanDetailText("清理任务: 释放 1GB")).toBe("释放 1GB");
    });
  });

  describe("formatHistoryDetail", () => {
    it("returns row.detail directly if already populated", () => {
      const row = {
        id: "1",
        time: "2026-11-09",
        task: "merge",
        streamer: "测试主播",
        status: "success",
        files_count: 2,
        freed_bytes: 0,
        merged_bytes: 1024,
        duration: 5,
        detail: "已合并 2 个分片"
      };
      expect(formatHistoryDetail(row)).toBe("已合并 2 个分片");
    });

    it("builds fallback detail for merge tasks", () => {
      const row = {
        id: "2",
        time: "2026-11-09",
        task: "merge",
        streamer: "测试主播",
        status: "success",
        files_count: 4,
        freed_bytes: 0,
        merged_bytes: 1024 * 1024 * 50,
        duration: 12,
        detail: ""
      };
      expect(formatHistoryDetail(row)).toBe("合并 4 个分片 · 50.0 MB · 12 秒");
    });

    it("builds fallback detail for clean tasks", () => {
      const row = {
        id: "3",
        time: "2026-11-09",
        task: "clean",
        streamer: "测试主播",
        status: "success",
        files_count: 10,
        freed_bytes: 1024 * 1024 * 1024 * 2,
        merged_bytes: 0,
        duration: 3,
        detail: ""
      };
      expect(formatHistoryDetail(row)).toBe("清理 10 个文件 · 释放 2.00 GB · 3 秒");
    });

    it("returns '已完成' when no detailed metrics are present", () => {
      const row = {
        id: "4",
        time: "2026-11-09",
        task: "custom",
        streamer: "",
        status: "success",
        files_count: 0,
        freed_bytes: 0,
        merged_bytes: 0,
        duration: 0,
        detail: ""
      };
      expect(formatHistoryDetail(row)).toBe("已完成");
    });
  });
});

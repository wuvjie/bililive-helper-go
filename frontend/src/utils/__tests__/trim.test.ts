import { describe, it, expect } from "vitest";
import {
  parseSeconds,
  formatFriendlyDuration,
  formatDurationClock,
  normalizeTimeString,
  computeTimelineSegments,
  describeTrimRange
} from "../trim";

describe("trim utilities", () => {
  describe("parseSeconds", () => {
    it("parses HH:MM:SS, MM:SS, full-width colons, and blank strings", () => {
      expect(parseSeconds("")).toBeNull();
      expect(parseSeconds("   ")).toBeNull();
      expect(parseSeconds("00:05:30")).toBe(330);
      expect(parseSeconds("05：30")).toBe(330);
      expect(parseSeconds("1:02:03")).toBe(3723);
    });

    it("returns NaN for malformed inputs or out-of-range minutes/seconds", () => {
      expect(parseSeconds("10")).toBeNaN();
      expect(parseSeconds("00:75:00")).toBeNaN();
      expect(parseSeconds("05:99")).toBeNaN();
      expect(parseSeconds("abc:10")).toBeNaN();
    });
  });

  describe("formatFriendlyDuration & formatDurationClock", () => {
    it("formats friendly Chinese durations cleanly", () => {
      expect(formatFriendlyDuration(0)).toBe("0 秒");
      expect(formatFriendlyDuration(45)).toBe("45 秒");
      expect(formatFriendlyDuration(330)).toBe("5 分钟 30 秒");
      expect(formatFriendlyDuration(3661)).toBe("1 小时 1 分钟");
    });

    it("formats HH:MM:SS clocks cleanly", () => {
      expect(formatDurationClock(0)).toBe("00:00:00");
      expect(formatDurationClock(3723)).toBe("01:02:03");
    });
  });

  describe("normalizeTimeString", () => {
    it("normalizes shorthand and full-width colons into HH:MM:SS", () => {
      expect(normalizeTimeString("")).toBe("");
      expect(normalizeTimeString("90:05")).toBe("01:30:05");
      expect(normalizeTimeString("1：2：3")).toBe("01:02:03");
      expect(normalizeTimeString("bad:time")).toBe("bad:time");
    });
  });

  describe("computeTimelineSegments", () => {
    it("uses real proportions when totalDurationSec is supplied", () => {
      const segs = computeTimelineSegments("00:10:00", "00:30:00", 3600);
      expect(segs.hasLeading).toBe(true);
      expect(segs.hasTrailing).toBe(true);
      expect(segs.leadingGrow).toBe(600);
      expect(segs.retainedGrow).toBe(1200);
      expect(segs.trailingGrow).toBe(1800);
    });
  });

  describe("describeTrimRange", () => {
    it("enforces manual trim selection & boundary checks when totalDurationSec is provided", () => {
      expect(describeTrimRange("00:00:00", "", { requireSelection: true }).isInvalid).toBe(true);
      expect(describeTrimRange("00:00:00", "", { requireSelection: false }).isInvalid).toBe(false);

      const startPastEnd = describeTrimRange("00:40:00", "", { totalDurationSec: 1800, requireSelection: true });
      expect(startPastEnd.isInvalid).toBe(true);
      expect(startPastEnd.durationText).toBe("超出时长");

      const fullCovered = describeTrimRange("00:00:00", "00:35:00", { totalDurationSec: 1800, requireSelection: true });
      expect(fullCovered.isInvalid).toBe(true);
      expect(fullCovered.durationText).toBe("无需剪辑");

      const partialEndOverflow = describeTrimRange("00:05:00", "00:40:00", { totalDurationSec: 1800, requireSelection: true });
      expect(partialEndOverflow.isInvalid).toBe(false);
      expect(partialEndOverflow.warningText).toContain("超过视频总时长");
      expect(partialEndOverflow.durationText).toBe("预计 25 分钟");
    });
  });
});

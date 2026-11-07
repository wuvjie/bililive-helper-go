import { describe, it, expect } from "vitest";
import { formatBytes, formatTime, formatDisplayDateTime, formatExactTime } from "../format";

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
      const fixedTs = 1762557000; // a valid unix timestamp
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
});

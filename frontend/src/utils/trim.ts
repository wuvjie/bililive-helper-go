export interface TimelineSegments {
  hasLeading: boolean;
  hasTrailing: boolean;
  isUntilEnd: boolean;
  leadingGrow: number;
  retainedGrow: number;
  trailingGrow: number;
}

export interface TrimSemanticInfo {
  isInvalid: boolean;
  warningText: string;
  text: string;
  durationText: string;
}

export interface TrimDescribeOptions {
  /** 视频实际总时长（秒），传入后可按真实时长检验区间边界并给出预计成片时长 */
  totalDurationSec?: number | null;
  /** 手动剪辑模式下要求必须设置有效起始或结束边界（不能等同于 00:00:00 ~ 末尾） */
  requireSelection?: boolean;
}

/** 将时分秒字符串（HH:MM:SS 或 MM:SS）解析为秒数；空串返回 null，格式错误返回 NaN */
export function parseSeconds(val?: string): number | null {
  if (!val || !val.trim()) return null;
  const cleaned = val.trim().replace(/：/g, ":");
  const rawParts = cleaned.split(":");
  if (rawParts.length !== 2 && rawParts.length !== 3) return NaN;
  if (rawParts.some(p => !/^\d+$/.test(p.trim()))) return NaN;
  const parts = rawParts.map(p => Number(p.trim()));
  if (parts.length === 3) {
    const [h, m, s] = parts;
    if (m >= 60 || s >= 60) return NaN;
    return h * 3600 + m * 60 + s;
  }
  const [m, s] = parts;
  if (s >= 60) return NaN;
  return m * 60 + s;
}

/** 将秒数格式化为紧凑可读的中文时长（如 "1 小时 30 分钟"、"45 秒"） */
export function formatFriendlyDuration(totalSec: number): string {
  const rounded = Math.max(0, Math.round(totalSec));
  if (rounded <= 0) return "0 秒";
  const h = Math.floor(rounded / 3600);
  const m = Math.floor((rounded % 3600) / 60);
  const s = rounded % 60;
  const parts: string[] = [];
  if (h > 0) parts.push(`${h} 小时`);
  if (m > 0) parts.push(`${m} 分钟`);
  if (s > 0 && h === 0) parts.push(`${s} 秒`);
  return parts.join(" ") || "0 秒";
}

/** 将秒数格式化为标准 HH:MM:SS 字符串 */
export function formatDurationClock(totalSec: number): string {
  const safe = Math.max(0, Math.floor(totalSec));
  const h = Math.floor(safe / 3600);
  const m = Math.floor((safe % 3600) / 60);
  const s = safe % 60;
  return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

/** 将用户输入的时间文本规范化为 HH:MM:SS（支持全角冒号与 MM:SS 简写进位） */
export function normalizeTimeString(val: string): string {
  if (!val || !val.trim()) return "";
  const cleaned = val.trim().replace(/：/g, ":");
  const parts = cleaned.split(":").map(p => p.trim());
  if (parts.some(p => !/^\d+$/.test(p))) return cleaned;
  if (parts.length === 3) {
    const h = Number(parts[0]);
    const m = Number(parts[1]);
    const s = Number(parts[2]);
    if (m >= 60 || s >= 60) return cleaned;
    return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
  }
  if (parts.length === 2) {
    const m = Number(parts[0]);
    const s = Number(parts[1]);
    if (s >= 60) return cleaned;
    const h = Math.floor(m / 60);
    const remM = m % 60;
    return `${String(h).padStart(2, "0")}:${String(remM).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
  }
  return cleaned;
}

/** 计算胶片时间轴可视化预览三段比例；传入真实总时长时按实际比例分配 */
export function computeTimelineSegments(
  trimStart?: string,
  trimEnd?: string,
  totalDurationSec?: number | null
): TimelineSegments {
  const rawStart = parseSeconds(trimStart);
  const rawEnd = parseSeconds(trimEnd);
  const startSec = rawStart === null || Number.isNaN(rawStart) ? 0 : rawStart;
  const endSec = rawEnd === null || Number.isNaN(rawEnd) ? null : rawEnd;

  const hasLeading = startSec > 0;
  const hasRealDuration = typeof totalDurationSec === "number" && Number.isFinite(totalDurationSec) && totalDurationSec > 0;

  if (hasRealDuration) {
    const total = totalDurationSec;
    const clampedStart = Math.min(startSec, total);
    const actualEnd = endSec !== null && endSec > clampedStart ? Math.min(endSec, total) : total;
    const retained = Math.max(0.5, actualEnd - clampedStart);
    const trailing = Math.max(0, total - actualEnd);
    const hasTrailing = trailing > 0.5;
    return {
      hasLeading: clampedStart > 0,
      hasTrailing,
      isUntilEnd: !hasTrailing && endSec === null,
      leadingGrow: clampedStart,
      retainedGrow: retained,
      trailingGrow: trailing
    };
  }

  const hasTrailing = endSec !== null && endSec > startSec;
  const isUntilEnd = endSec === null;

  let leadingGrow = 1;
  let retainedGrow = 3.5;
  let trailingGrow = 1;

  if (hasLeading && hasTrailing && endSec !== null) {
    const diff = Math.max(1, endSec - startSec);
    leadingGrow = Math.max(1, Math.min(2.5, Math.round((startSec / diff) * 2)));
    retainedGrow = 4;
    trailingGrow = 1.2;
  } else if (hasLeading && isUntilEnd) {
    leadingGrow = 1.2;
    retainedGrow = 4;
  } else if (!hasLeading && hasTrailing) {
    retainedGrow = 4;
    trailingGrow = 1.8;
  }

  return {
    hasLeading,
    hasTrailing,
    isUntilEnd,
    leadingGrow,
    retainedGrow,
    trailingGrow
  };
}

/** 校验并生成剪辑起止时间区间的中文语义说明与警告 */
export function describeTrimRange(
  trimStart?: string,
  trimEnd?: string,
  options: TrimDescribeOptions = {}
): TrimSemanticInfo {
  const rawStart = parseSeconds(trimStart);
  const rawEnd = parseSeconds(trimEnd);
  const startDisplay = trimStart?.trim() || "00:00:00";

  if (Number.isNaN(rawStart) || (rawEnd !== null && Number.isNaN(rawEnd))) {
    return {
      isInvalid: true,
      warningText: "",
      text: "时间格式无效，请按「时:分:秒」（如 01:30:00）或「分:秒」（如 05:00）格式输入，且分、秒需小于 60",
      durationText: "格式错误"
    };
  }

  const startSec = rawStart ?? 0;
  const endSec = rawEnd;

  if (endSec !== null && endSec <= startSec) {
    return {
      isInvalid: true,
      warningText: "",
      text: `截取结束时间 (${trimEnd}) 必须晚于截取起始时间 (${startDisplay})`,
      durationText: "区间无效"
    };
  }

  const totalDur =
    typeof options.totalDurationSec === "number" &&
    Number.isFinite(options.totalDurationSec) &&
    options.totalDurationSec > 0
      ? options.totalDurationSec
      : null;

  if (totalDur !== null && startSec >= totalDur) {
    return {
      isInvalid: true,
      warningText: "",
      text: `起始时间 (${startDisplay}) 已超过当前视频总时长 (${formatFriendlyDuration(totalDur)})`,
      durationText: "超出时长"
    };
  }

  if (totalDur !== null && startSec === 0 && endSec !== null && endSec >= totalDur) {
    return {
      isInvalid: true,
      warningText: "",
      text: `当前视频总时长仅 ${formatFriendlyDuration(totalDur)}，已短于截止时间 (${trimEnd})，无需剪辑`,
      durationText: "无需剪辑"
    };
  }

  if (startSec === 0 && endSec === null) {
    if (options.requireSelection) {
      return {
        isInvalid: true,
        warningText: "",
        text: "请至少设置剪辑起始时间或结束时间（当前区间等同于保留完整原视频）",
        durationText: "未设置区间"
      };
    }
    return {
      isInvalid: false,
      warningText: "",
      text: "全场保留：不执行裁剪，完整保留全场录播内容",
      durationText: "全场保留"
    };
  }

  if (startSec > 0 && endSec === null) {
    const remainingText =
      totalDur !== null
        ? `预计成片 ${formatFriendlyDuration(Math.max(0, totalDur - startSec))}`
        : `跳过前 ${formatFriendlyDuration(startSec)}`;
    return {
      isInvalid: false,
      warningText: "",
      text: `跳过前段：截除前 ${formatFriendlyDuration(startSec)}，后续内容保留至录像结尾`,
      durationText: remainingText
    };
  }

  if (startSec === 0 && endSec !== null) {
    return {
      isInvalid: false,
      warningText: "",
      text: `保留前段：仅截取前 ${formatFriendlyDuration(endSec)} 内容，后续部分剪除`,
      durationText: `时长 ${formatFriendlyDuration(endSec)}`
    };
  }

  if (startSec > 0 && endSec !== null) {
    if (totalDur !== null && endSec > totalDur) {
      const actualDiff = Math.max(0, totalDur - startSec);
      return {
        isInvalid: false,
        warningText: `截止时间 (${trimEnd}) 超过视频总时长 (${formatFriendlyDuration(totalDur)})，将从 ${startDisplay} 截取至视频末尾`,
        text: `从 ${startDisplay} 截取至视频末尾，预计成片约 ${formatFriendlyDuration(actualDiff)}`,
        durationText: `预计 ${formatFriendlyDuration(actualDiff)}`
      };
    }
    const diff = endSec - startSec;
    return {
      isInvalid: false,
      warningText: "",
      text: `选定时段：从 ${startDisplay} 至 ${trimEnd}，剪除首尾多余部分`,
      durationText: `有效时长 ${formatFriendlyDuration(diff)}`
    };
  }

  return {
    isInvalid: false,
    warningText: "",
    text: "自定义裁剪时段",
    durationText: ""
  };
}

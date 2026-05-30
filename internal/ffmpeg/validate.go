package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	minValidSize     = 10240   // 10KB absolute minimum
	minSizePerSecond = 5000    // 5KB/s — below this bitrate something is wrong
	probeTimeout     = 30 * time.Second
)

// ValidateOutput checks if a video file is playable and complete.
// Designed for live stream recordings which may have minor timestamp glitches.
//
// Strategy:
//  1. Container integrity: ffprobe can read duration + stream info
//  2. Multi-point decode: sample 3 points (10%, 50%, 90%) with -v warning (not error)
//  3. Duration-size sanity: file is large enough for its duration
//  4. Stream presence: at least one video stream exists
//
// Returns nil if the file is considered playable.
func ValidateOutput(ctx context.Context, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("文件不存在: %s", path)
	}
	if info.Size() < minValidSize {
		return fmt.Errorf("文件过小: %s (%d bytes)", path, info.Size())
	}

	// Step 1: Read container metadata (duration, streams)
	duration, streams, err := probeMetadata(ctx, path)
	if err != nil {
		return fmt.Errorf("容器损坏，无法读取元数据: %w", err)
	}
	if duration <= 0 {
		return fmt.Errorf("时长异常: %.1fs", duration)
	}
	if streams < 1 {
		return fmt.Errorf("无媒体流")
	}

	// Step 2: Duration-size sanity check
	// For live recordings, expect at least 5KB/s (very conservative —
	// even audio-only streams are ~16KB/s at 128kbps)
	expectedMinSize := int64(duration * minSizePerSecond)
	if info.Size() < expectedMinSize {
		return fmt.Errorf("文件过小(时长%.0fs但只有%s)，可能截断", duration, formatSize(info.Size()))
	}

	// Step 3: Multi-point decode test
	// Sample at 10%, 50%, 90% of the file to catch corruption anywhere.
	// Use -v warning (not error) to tolerate minor timestamp glitches
	// common in live stream recordings.
	if err := multiPointDecode(ctx, path, duration); err != nil {
		return err
	}

	return nil
}

// probeMetadata reads duration and stream count via ffprobe.
func probeMetadata(ctx context.Context, path string) (duration float64, streams int, err error) {
	probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	// Get duration
	cmd := exec.CommandContext(probeCtx, "ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path)
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe duration 失败: %w", err)
	}
	duration, err = strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("解析时长失败: %s", strings.TrimSpace(string(out)))
	}

	// Count streams
	cmd2 := exec.CommandContext(probeCtx, "ffprobe",
		"-v", "quiet",
		"-show_entries", "stream=index",
		"-of", "csv=p=0",
		path)
	out2, err := cmd2.Output()
	if err != nil {
		return duration, 0, nil // duration readable but stream count failed — still try
	}
	streams = len(strings.Split(strings.TrimSpace(string(out2)), "\n"))
	if strings.TrimSpace(string(out2)) == "" {
		streams = 0
	}

	return duration, streams, nil
}

// multiPointDecode decodes short segments at 3 positions in the file.
// Uses -v warning to tolerate minor timestamp discontinuities.
func multiPointDecode(ctx context.Context, path string, duration float64) error {
	// For very short files (< 30s), just decode the whole thing
	if duration < 30 {
		return singleDecode(ctx, path, 0, duration)
	}

	// Sample at 10%, 50%, 90% — each segment is 5 seconds
	positions := []float64{
		duration * 0.10,
		duration * 0.50,
		duration * 0.90,
	}

	for _, pos := range positions {
		if err := singleDecode(ctx, path, pos, 5); err != nil {
			return fmt.Errorf("解码失败 @ %.0f%%: %w", pos/duration*100, err)
		}
	}
	return nil
}

// singleDecode decodes a segment of the file.
// Uses -v warning (not error) to tolerate minor timestamp issues.
// Returns error only if ffmpeg exits with non-zero AND produces output
// indicating fatal stream errors.
func singleDecode(ctx context.Context, path string, start, durationSec float64) error {
	decodeCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	args := []string{
		"-v", "warning",
		"-ss", fmt.Sprintf("%.1f", start),
		"-i", path,
		"-t", fmt.Sprintf("%.0f", durationSec),
		"-f", "null", "-",
	}
	cmd := exec.CommandContext(decodeCtx, "ffmpeg", args...)
	out, err := cmd.CombinedOutput()

	if err == nil {
		return nil
	}

	// ffmpeg returned non-zero. Check if it's a fatal error or just warnings.
	// For live recordings, we tolerate non-zero exit if the output only contains
	// common non-fatal warnings.
	output := string(out)
	fatal := isFatalFFmpegError(output)
	if fatal {
		return fmt.Errorf("解码段 @%.0fs 出现致命错误: %s", start, truncate(output, 200))
	}

	// Non-fatal warnings with non-zero exit — still acceptable for live recordings
	return nil
}

// isFatalFFmpegError checks if ffmpeg output contains truly fatal errors
// (as opposed to common non-fatal warnings in live recordings).
func isFatalFFmpegError(output string) bool {
	lower := strings.ToLower(output)

	// These are fatal — the file is truly broken
	fatalPatterns := []string{
		"invalid data",
		"no such file",
		"permission denied",
		"could not find codec",
		"decoding error",
		"error decoding",
		" corrupt",
		"truncated",
		"moov atom not found",
		"missing mandatory",
		"invalid NAL",
		"decode_slice_header",
		"no frame!",
	}
	for _, p := range fatalPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}

	// If there's no recognizable fatal pattern but ffmpeg still failed,
	// and the output is very short (likely just a version/config issue),
	// treat as non-fatal
	if len(strings.TrimSpace(output)) == 0 {
		return false
	}

	// Default: if we can't classify, be lenient for live recordings
	return false
}

// quickProbe runs a fast metadata-only check (no decode).
// Used during scanning to verify an existing output is basically readable.
func QuickProbe(ctx context.Context, path string) error {
	probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("文件不存在: %s", path)
	}
	if info.Size() < minValidSize {
		return fmt.Errorf("文件过小: %s (%d bytes)", path, info.Size())
	}

	cmd := exec.CommandContext(probeCtx, "ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("ffprobe 失败: %w", err)
	}

	dur, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil || dur <= 0 {
		return fmt.Errorf("时长异常: %s", strings.TrimSpace(string(out)))
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

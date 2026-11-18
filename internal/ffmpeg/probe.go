package ffmpeg

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// runFFprobe 执行 ffprobe 命令并返回去除首尾空白的标准输出。
func runFFprobe(ctx context.Context, timeout time.Duration, args ...string) (string, error) {
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(probeCtx, "ffprobe", args...)
	out, err := cmd.Output()
	if err != nil {
		if probeCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("ffprobe 超时（%v）: %w", timeout, err)
		}
		return "", fmt.Errorf("ffprobe 执行失败: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ProbeDuration 通过 ffprobe 获取视频文件的时长（秒），超时 30 秒。
// 从 utils/video.go 迁移而来，归入 ffmpeg 包（ffprobe 操作属于此包的职责）。
func ProbeDuration(ctx context.Context, path string) (float64, error) {
	s, err := runFFprobe(ctx, 30*time.Second, "-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path)
	if err != nil {
		return 0, err
	}
	dur, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("解析时长失败: %s", s)
	}
	return dur, nil
}

// ProbeHealth 通过 ffprobe 检查视频文件是否有有效的编码流。
// 用于在合并前剔除结构损坏的文件（如缺少 moov atom 的 MP4）。
// 返回 true 表示文件结构完整，false 表示文件损坏应跳过。
func ProbeHealth(ctx context.Context, path string) bool {
	return ProbeVideoCodec(ctx, path) != ""
}

// ProbeVideoCodec 探测输入视频主视频流的编码格式（小写，如 "h264"、"hevc"、"av1"）。
// 探测失败或无视频流时返回空字符串。
func ProbeVideoCodec(ctx context.Context, path string) string {
	out, err := runFFprobe(ctx, 10*time.Second, "-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path)
	if err != nil {
		return ""
	}
	codec := strings.ToLower(out)
	if idx := strings.IndexByte(codec, '\n'); idx >= 0 {
		codec = strings.TrimSpace(codec[:idx])
	}
	return codec
}

// IsVideoTrimmed 检查视频文件格式元数据中是否带有已裁剪标记（comment=bh_trimmed）。
func IsVideoTrimmed(ctx context.Context, path string) bool {
	out, err := runFFprobe(ctx, 10*time.Second, "-v", "quiet",
		"-show_entries", "format_tags=comment",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path)
	if err != nil {
		return false
	}
	return strings.Contains(out, "bh_trimmed")
}

// HasAudioStream 检查视频文件是否包含有效的音频流。
func HasAudioStream(ctx context.Context, path string) bool {
	out, err := runFFprobe(ctx, 10*time.Second, "-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=codec_name",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path)
	return err == nil && len(out) > 0
}

// ProbeVideoDimensions 探测视频主视频流的宽高像素（如 1920, 1080）。探测失败返回 (0, 0)。
func ProbeVideoDimensions(ctx context.Context, path string) (int, int) {
	out, err := runFFprobe(ctx, 10*time.Second, "-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0:s=x",
		path)
	if err != nil || out == "" {
		return 0, 0
	}
	firstLine := strings.Split(out, "\n")[0]
	parts := strings.Split(strings.TrimSpace(firstLine), "x")
	if len(parts) != 2 {
		return 0, 0
	}
	w, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	h, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || w <= 0 || h <= 0 {
		return 0, 0
	}
	return w, h
}

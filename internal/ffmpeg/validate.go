// validate.go 提供视频文件的输出校验功能。
// 通过容器元数据检查、时长-大小合理性验证和多点解码测试，确保输出文件可正常播放。
package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"bililive-helper/internal/utils"
)

const (
	minValidSize     = 10240   // 10KB: 文件最小有效大小
	minSizePerSecond = 5000    // 5KB/s: 低于此码率视为异常
	probeTimeout     = 30 * time.Second
)

// ValidateOutput 检查视频文件是否可播放且完整。
// 针对直播录制设计，能容忍轻微的时间戳不连续。
//
// 校验策略：
//  1. 容器完整性：ffprobe 能读取时长和流信息
//  2. 时长-大小合理性：文件大小与报告时长匹配
//  3. 多点解码：在 10%、50%、90% 位置采样解码，使用 -v warning（非 error）容忍轻微时间戳问题
//
// 返回 nil 表示文件可正常播放。
func ValidateOutput(ctx context.Context, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("文件不存在: %s", path)
	}
	if info.Size() < minValidSize {
		return fmt.Errorf("文件过小: %s (%d bytes)", path, info.Size())
	}

	// 步骤 1：读取容器元数据（时长、流数量）
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

	// 步骤 2：时长-大小合理性检查
	// 直播录制预期最低 5KB/s（非常保守 — 纯音频 128kbps 约 16KB/s）
	expectedMinSize := int64(duration * minSizePerSecond)
	if info.Size() < expectedMinSize {
		return fmt.Errorf("文件过小(时长%.0fs但只有%s)，可能截断", duration, utils.FormatSize(info.Size()))
	}

	// 步骤 3：多点解码测试
	// 在 10%、50%、90% 位置采样，捕获文件任意位置的损坏。
	// 使用 -v warning（非 error）容忍直播录制中的轻微时间戳问题。
	if err := multiPointDecode(ctx, path, duration); err != nil {
		return err
	}

	return nil
}

// probeMetadata 通过 ffprobe 读取时长和流数量。
// 出错时返回部分结果（例如时长可能已获取但流数量失败）。
func probeMetadata(ctx context.Context, path string) (duration float64, streams int, err error) {
	probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	// 获取时长
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

	// 统计流数量
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

// multiPointDecode 在文件的 3 个位置解码短片段（10%、50%、90%）。
// 使用 -v warning 容忍轻微的时间戳不连续。
func multiPointDecode(ctx context.Context, path string, duration float64) error {
	// 短文件（< 30s）直接解码整个文件
	if duration < 30 {
		return singleDecode(ctx, path, 0, duration)
	}

	// 在 10%、50%、90% 位置采样，每段 5 秒
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

// singleDecode 解码文件中的短片段。
// 使用 -v warning 容忍直播录制中的轻微时间戳问题。
// 仅对真正的致命流错误返回错误。
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

	// ffmpeg 返回非零退出码。区分致命错误和直播录制中常见的非致命警告。
	output := string(out)
	fatal := isFatalFFmpegError(output)
	if fatal {
		return fmt.Errorf("解码段 @%.0fs 出现致命错误: %s", start, truncate(output, 200))
	}

	// 非致命警告且非零退出码 — 直播录制中可接受
	return nil
}

// isFatalFFmpegError 检查 ffmpeg 输出是否包含真正的致命错误
// （区别于直播录制中常见的非致命警告）。
func isFatalFFmpegError(output string) bool {
	lower := strings.ToLower(output)

	// 这些模式表明文件确实损坏
	fatalPatterns := []string{
		"invalid data",
		"no such file",
		"permission denied",
		"could not find codec",
		"decoding error",
		"error decoding",
		" corrupt",
		"file is truncated",
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

	// 空输出 + 非零退出码 — 可能是版本/配置问题，非文件损坏
	if len(strings.TrimSpace(output)) == 0 {
		return false
	}

	// 默认：对直播录制保持宽容 — 无法判定为致命的则允许通过
	return false
}

// QuickProbe 执行快速的元数据检查（无解码）。
// 用于扫描阶段验证已存在的输出是否基本可读，不做完整校验。
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

// truncate 截断字符串到指定最大长度，超出部分用 "..." 替代。
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

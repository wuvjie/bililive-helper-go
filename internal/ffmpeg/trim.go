// trim.go 提供视频时段裁剪功能。
// 使用 ffmpeg -ss / -t / -to 进行无损 stream-copy 剪辑（零二次编码）。
package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// TrimVideo 按时间参数裁剪视频并输出到指定路径。
// 全程采用 stream-copy（-c copy），零二次重编码，速度极快且画质无损。
// startTime: 起始时间（如 "00:05:00" 或 "0"），为空时默认为 "00:00:00"；
// durationOrEnd: 时长或截止时间；
// isEnd: 若为 true 表示 durationOrEnd 为截止时间点（-to），否则为持续时长（-t）。
func TrimVideo(ctx context.Context, input, output, startTime, durationOrEnd string, isEnd bool) error {
	if startTime == "" {
		startTime = "00:00:00"
	}
	durationOrEnd = strings.TrimSpace(durationOrEnd)
	if durationOrEnd == "" {
		return fmt.Errorf("裁剪时长或截止时间不能为空")
	}

	args := []string{"-nostdin"}
	if isEnd {
		args = append(args, "-ss", startTime, "-to", durationOrEnd, "-i", input)
	} else {
		args = append(args, "-ss", startTime, "-i", input, "-t", durationOrEnd)
	}
	args = append(args, "-c", "copy", "-y", "-loglevel", "error", output)

	err := Run(ctx, Options{Args: args})
	if err != nil {
		os.Remove(output)
		return fmt.Errorf("视频裁剪失败: %w", err)
	}

	if err := ValidateOutput(ctx, output); err != nil {
		os.Remove(output)
		return fmt.Errorf("裁剪输出校验失败: %w", err)
	}

	return nil
}

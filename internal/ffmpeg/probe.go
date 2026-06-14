package ffmpeg

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// ProbeDuration 通过 ffprobe 获取视频文件的时长（秒），超时 30 秒。
// 从 utils/video.go 迁移而来，归入 ffmpeg 包（ffprobe 操作属于此包的职责）。
func ProbeDuration(ctx context.Context, path string) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ffprobe", "-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path)
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return 0, fmt.Errorf("ffprobe 超时（30秒）: %s", path)
		}
		return 0, fmt.Errorf("ffprobe 执行失败: %w", err)
	}
	s := strings.TrimSpace(string(out))
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
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ffprobe", "-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

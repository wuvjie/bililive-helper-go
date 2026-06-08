package utils

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var filenameRegex = regexp.MustCompile(`(?i)^\[(\d{4}-\d{2}-\d{2}) (\d{2}-\d{2}-\d{2})\](\[.+?\]\[.+?\]).*\.(mp4|flv|ts)$`)

// ParseFilename 从 bililive-go 格式的文件名中解析录制时间和主播标识。
// 期望格式：[YYYY-MM-DD HH-MM-SS][streamer_id][title].ext
// 返回主播标识方括号段、解析后的时间和是否成功。
func ParseFilename(name string) (string, time.Time, bool) {
	matches := filenameRegex.FindStringSubmatch(name)
	if matches == nil {
		return "", time.Time{}, false
	}
	dt, err := time.ParseInLocation("2006-01-02 15-04-05", matches[1]+" "+matches[2], time.Local)
	if err != nil {
		return "", time.Time{}, false
	}
	return matches[3], dt, true
}

// GetVideoDuration 通过 ffprobe 获取视频文件的时长（秒），超时 30 秒。
func GetVideoDuration(ctx context.Context, path string) (float64, error) {
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

// IsVideoHealthy 通过 ffprobe 检查视频文件是否有有效的编码流。
// 用于在合并前剔除结构损坏的文件（如缺少 moov atom 的 MP4）。
// 返回 true 表示文件结构完整，false 表示文件损坏应跳过。
func IsVideoHealthy(ctx context.Context, path string) bool {
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

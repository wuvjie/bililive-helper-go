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

var filenameRegex = regexp.MustCompile(`^\[(\d{4}-\d{2}-\d{2}) (\d{2}-\d{2}-\d{2})\](\[.+?\]\[.+?\]).*\.(mp4|flv|ts)$`)

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

func GetVideoDuration(path string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

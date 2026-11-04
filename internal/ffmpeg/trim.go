// trim.go 提供视频时段裁剪功能。
// 使用 ffmpeg -ss / -to 进行无损 stream-copy 剪辑（零二次编码）。
package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseTimeSeconds 将 "HH:MM:SS"、"MM:SS" 或纯秒数字符串解析为浮点秒数。
func ParseTimeSeconds(t string) (float64, error) {
	t = strings.TrimSpace(t)
	if t == "" {
		return 0, nil
	}
	parts := strings.Split(t, ":")
	if len(parts) == 1 {
		return strconv.ParseFloat(parts[0], 64)
	} else if len(parts) == 2 {
		m, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, err
		}
		s, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, err
		}
		return m*60 + s, nil
	} else if len(parts) == 3 {
		h, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, err
		}
		m, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, err
		}
		s, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return 0, err
		}
		return h*3600 + m*60 + s, nil
	}
	return 0, fmt.Errorf("无效时间格式: %s", t)
}

// TrimVideo 按起止时间裁剪视频并输出到指定路径。
// 全程采用 stream-copy（-c copy），零二次重编码，速度极快且画质无损。
// startTime: 起始时间（如 "00:05:00"），为空时默认为 "00:00:00"；
// endTime: 截止时间（如 "00:30:00"），为空时表示裁剪至视频结尾；
// 支持 -avoid_negative_ts make_zero 避免关键帧时间戳负偏移引起的首帧黑屏/卡顿。
func TrimVideo(ctx context.Context, input, output, startTime, endTime string) error {
	startTime = strings.TrimSpace(startTime)
	endTime = strings.TrimSpace(endTime)
	if startTime == "" {
		startTime = "00:00:00"
	}

	args := []string{"-nostdin"}
	if startTime != "" && startTime != "0" && startTime != "00:00:00" {
		args = append(args, "-ss", startTime)
	}
	if endTime != "" {
		args = append(args, "-to", endTime)
	}
	args = append(args, "-i", input, "-c", "copy", "-avoid_negative_ts", "make_zero", "-y", "-loglevel", "error", output)

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

// concat.go 提供 TS 文件的拼接合并功能。
// 使用 ffmpeg concat 协议进行直接流复制（无重编码），速度快但要求编解码器兼容。
package ffmpeg

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// ConcatTS 使用 ffmpeg concat 协议将多个 TS 文件合并为单个输出文件。
// 直接流复制（无重编码）— 速度快但要求编解码器兼容。
// onProgress 接收合并进度百分比更新（可为 nil）。
func ConcatTS(ctx context.Context, tsFiles []string, output string, onProgress func(string)) error {
	concatArg := "concat:" + strings.Join(tsFiles, "|")

	var durationUs int64
	onStdout := func(line string) {
		if onProgress == nil {
			return
		}
		if strings.HasPrefix(line, "duration_us=") {
			if d, err := strconv.ParseInt(strings.TrimPrefix(line, "duration_us="), 10, 64); err == nil {
				durationUs = d
			}
		} else if strings.HasPrefix(line, "out_time_us=") {
			if t, err := strconv.ParseInt(strings.TrimPrefix(line, "out_time_us="), 10, 64); err == nil {
				if durationUs > 0 {
					pct := float64(t) / float64(durationUs) * 100
					if pct > 100 {
						pct = 100
					}
					onProgress(fmt.Sprintf("⏳ 合并进度 %.0f%%", pct))
				}
			}
		}
	}

	return Run(ctx, Options{
		Args: []string{
			"-nostdin",
			"-i", concatArg,
			"-c", "copy", "-y",
			"-progress", "pipe:1", "-loglevel", "error",
			output,
		},
		OnStdout: onStdout,
	})
}

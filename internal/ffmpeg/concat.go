// concat.go 提供 TS 文件的拼接合并功能。
// 使用 ffmpeg concat demuxer（清单文件模式）进行直接流复制（无重编码），速度快且避免命令行超长。
package ffmpeg

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
)

// ConcatTS 使用 ffmpeg concat 分流器（concat demuxer）将多个 TS 文件合并为单个输出文件。
// 直接流复制（无重编码）— 速度快且不受分片文件数量限制。
// onProgress 接收合并进度百分比更新（可为 nil）。
func ConcatTS(ctx context.Context, tsFiles []string, output string, onProgress func(string)) error {
	if len(tsFiles) == 0 {
		return fmt.Errorf("无待合并文件")
	}

	// 创建临时清单文件（清单文件模式根治切片多时命令行过长报错）
	listFile, err := os.CreateTemp("", "bh_concat_*.txt")
	if err != nil {
		return fmt.Errorf("创建合并清单文件失败: %w", err)
	}
	listPath := listFile.Name()
	defer os.Remove(listPath)

	var buf bytes.Buffer
	for _, f := range tsFiles {
		absPath, err := filepath.Abs(f)
		if err != nil {
			absPath = f
		}
		// Concat demuxer 规范要求路径中的反斜杠替换为正斜杠，单引号转义
		formatted := strings.ReplaceAll(absPath, "\\", "/")
		formatted = strings.ReplaceAll(formatted, "'", "'\\''")
		buf.WriteString(fmt.Sprintf("file '%s'\n", formatted))
	}
	if _, err := listFile.Write(buf.Bytes()); err != nil {
		_ = listFile.Close()
		return fmt.Errorf("写入合并清单文件失败: %w", err)
	}
	_ = listFile.Close()

	var durationUs atomic.Int64
	onStdout := func(line string) {
		if onProgress == nil {
			return
		}
		if strings.HasPrefix(line, "duration_us=") {
			if d, err := strconv.ParseInt(strings.TrimPrefix(line, "duration_us="), 10, 64); err == nil {
				durationUs.Store(d)
			}
		} else if strings.HasPrefix(line, "out_time_us=") {
			if t, err := strconv.ParseInt(strings.TrimPrefix(line, "out_time_us="), 10, 64); err == nil {
				dur := durationUs.Load()
				if dur > 0 {
					pct := float64(t) / float64(dur) * 100
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
			"-fflags", "+genpts+igndts",
			"-f", "concat",
			"-safe", "0",
			"-i", listPath,
			"-c", "copy", "-y",
			"-progress", "pipe:1", "-loglevel", "error",
			output,
		},
		OnStdout: onStdout,
	})
}

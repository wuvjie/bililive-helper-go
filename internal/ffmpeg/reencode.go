package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	ReencodeTimeout = 6 * time.Hour
	MaxReencodeSize = 5 * 1024 * 1024 * 1024 // 5GB
)

// Reencode merges files using ffmpeg concat filter with re-encoding.
// Used as fallback when copy mode fails. Auto-detects h264_rkmpp hardware encoder.
// onProgress receives status updates (may be nil).
func Reencode(ctx context.Context, files []string, folder, output string, onProgress func(string)) error {
	if len(files) < 2 {
		return fmt.Errorf("重编码需要至少2个文件")
	}

	// Detect hardware encoder
	videoEncoder := detectVideoEncoder()

	args := []string{"-nostdin"}
	for _, f := range files {
		args = append(args, "-i", filepath.ToSlash(filepath.Join(folder, f)))
	}

	// Build filter_complex: [0:v][0:a][1:v][1:a]...concat=n=N:v=1:a=1[outv][outa]
	n := len(files)
	var filterParts []string
	for i := 0; i < n; i++ {
		filterParts = append(filterParts, fmt.Sprintf("[%d:v:0][%d:a:0]", i, i))
	}
	filter := fmt.Sprintf("%sconcat=n=%d:v=1:a=1[outv][outa]", strings.Join(filterParts, ""), n)

	args = append(args,
		"-filter_complex", filter,
		"-map", "[outv]", "-map", "[outa]",
		"-c:v", videoEncoder,
	)

	if videoEncoder == "h264_rkmpp" {
		args = append(args,
			"-rc_mode", "2",
			"-qp", "28",
			"-maxrate", "4M",
			"-bufsize", "8M",
			"-threads", "2",
			"-err_detect", "ignore_err",
			"-fflags", "+genpts",
			"-avoid_negative_ts", "make_zero",
		)
	} else {
		args = append(args,
			"-preset", "ultrafast",
			"-crf", "28",
			"-threads", "2",
		)
	}

	args = append(args,
		"-c:a", "aac", "-b:a", "128k",
		"-y", "-loglevel", "error",
		output)

	if onProgress != nil {
		onProgress(fmt.Sprintf("🔄 重编码合并 %d 个文件…", len(files)))
	}

	err := Run(ctx, Options{
		Args:    args,
		Timeout: ReencodeTimeout,
	})
	if err != nil {
		os.Remove(output)
		return fmt.Errorf("重编码失败: %w", err)
	}

	// Validate output
	if err := ValidateOutput(ctx, output); err != nil {
		os.Remove(output)
		return fmt.Errorf("重编码输出校验失败: %w", err)
	}

	info, err := os.Stat(output)
	if err != nil {
		os.Remove(output)
		return fmt.Errorf("重编码输出不存在: %w", err)
	}

	if onProgress != nil {
		onProgress(fmt.Sprintf("✅ 重编码完成: %s", formatSize(info.Size())))
	}
	return nil
}

// detectVideoEncoder returns "h264_rkmpp" if available, otherwise "libx264".
func detectVideoEncoder() string {
	if out, err := exec.Command("ffmpeg", "-encoders").Output(); err == nil {
		if strings.Contains(string(out), "h264_rkmpp") {
			return "h264_rkmpp"
		}
	}
	return "libx264"
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	default:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	}
}

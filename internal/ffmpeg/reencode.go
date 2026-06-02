// reencode.go 提供 ffmpeg concat filter 重编码合并功能。
// 作为 stream-copy 失败时的 fallback，自动检测并使用硬件编码器（h264_rkmpp）。
package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bililive-helper/internal/utils"
)

const (
	ReencodeTimeout = 6 * time.Hour
	MaxReencodeSize = 5 * 1024 * 1024 * 1024 // 5GB
)

// Reencode 使用 ffmpeg concat filter 重编码合并文件。
// 作为 stream-copy 合并失败时的 fallback（编解码器不兼容、头部损坏等情况）。
// 自动检测 ARM 开发板（如 Rockchip）上的 h264_rkmpp 硬件编码器。
// onProgress 接收状态更新（可为 nil）。
func Reencode(ctx context.Context, files []string, folder, output string, onProgress func(string)) error {
	if len(files) < 2 {
		return fmt.Errorf("重编码需要至少2个文件")
	}

	videoEncoder := detectVideoEncoder()

	args := []string{"-nostdin"}
	for _, f := range files {
		args = append(args, "-i", filepath.ToSlash(filepath.Join(folder, f)))
	}

	// filter_complex: [0:v][0:a][1:v][1:a]...concat=n=N:v=1:a=1[outv][outa]
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

	// 校验重编码输出：确保文件可正常播放
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
		onProgress(fmt.Sprintf("✅ 重编码完成: %s", utils.FormatSize(info.Size())))
	}
	return nil
}

// detectVideoEncoder 检测系统可用的视频编码器。
// 优先返回 h264_rkmpp（Rockchip 硬件加速），否则返回 libx264（软件编码）。
// 使用 sync.Once 缓存结果，线程安全。
var (
	cachedEncoder   string
	encoderOnce     sync.Once
)

func detectVideoEncoder() string {
	encoderOnce.Do(func() {
		if out, err := exec.Command("ffmpeg", "-encoders").Output(); err == nil {
			if strings.Contains(string(out), "h264_rkmpp") {
				cachedEncoder = "h264_rkmpp"
				return
			}
		}
		cachedEncoder = "libx264"
	})
	return cachedEncoder
}

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

	"bililive-helper-go/internal/utils"
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
	return ReencodeWithDimensions(ctx, files, folder, output, nil, nil, onProgress)
}

// ReencodeWithDimensions 与 Reencode 相同，支持复用调用方已探测的各分片宽高（长度与 files 一致时跳过重复 ffprobe）。
func ReencodeWithDimensions(ctx context.Context, files []string, folder, output string, widths, heights []int, onProgress func(string)) error {
	if len(files) < 2 {
		return fmt.Errorf("重编码需要至少2个文件")
	}

	videoEncoder := detectVideoEncoder()

	args := []string{"-nostdin"}
	for _, f := range files {
		args = append(args, "-i", filepath.ToSlash(filepath.Join(folder, f)))
	}

	// 预检每个输入文件是否包含音频流、视频分辨率是否一致，兼容混合音轨及跨分辨率/横竖屏切换场景
	n := len(files)
	hasAudioSlice := make([]bool, n)
	durations := make([]float64, n)
	hasDims := len(widths) == n && len(heights) == n
	if !hasDims {
		widths = make([]int, n)
		heights = make([]int, n)
	}
	anyHasAudio := false
	allHaveAudio := true
	for i, f := range files {
		fullPath := filepath.Join(folder, f)
		if HasAudioStream(ctx, fullPath) {
			hasAudioSlice[i] = true
			anyHasAudio = true
		} else {
			allHaveAudio = false
		}
		if !hasDims {
			widths[i], heights[i] = ProbeVideoDimensions(ctx, fullPath)
		}
	}
	if anyHasAudio && !allHaveAudio {
		for i, f := range files {
			if !hasAudioSlice[i] {
				if dur, err := ProbeDuration(ctx, filepath.Join(folder, f)); err == nil && dur > 0 {
					durations[i] = dur
				}
			}
		}
	}

	filter, useAudio := buildConcatFilterWithScale(n, hasAudioSlice, durations, widths, heights)
	args = append(args, "-filter_complex", filter, "-map", "[outv]")
	if useAudio {
		args = append(args, "-map", "[outa]")
	}
	args = append(args, "-c:v", videoEncoder)

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

	if useAudio {
		args = append(args, "-c:a", "aac", "-b:a", "128k")
	}
	args = append(args,
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

// buildConcatFilter 根据各分片的音轨存在情况与时长构造 filter_complex 表达式。
// 全有音轨时直接拼接音视频；全无音轨或缺失音轨且无法探测时长时回退为纯视频拼接；
// 部分分片缺失音轨且时长已知时，使用 anullsrc 生成等长静音流补齐音频轨。
func buildConcatFilter(n int, hasAudio []bool, durations []float64) (filter string, useAudio bool) {
	return buildConcatFilterWithScale(n, hasAudio, durations, nil, nil)
}

func buildConcatFilterWithScale(n int, hasAudio []bool, durations []float64, widths, heights []int) (filter string, useAudio bool) {
	targetW, targetH := 0, 0
	needScale := false
	for i := 0; i < n && i < len(widths) && i < len(heights); i++ {
		if widths[i] <= 0 || heights[i] <= 0 {
			continue
		}
		if targetW == 0 {
			targetW, targetH = widths[i], heights[i]
		} else if widths[i] != targetW || heights[i] != targetH {
			needScale = true
		}
	}

	anyHave := false
	canPadSilence := true
	for i := 0; i < n; i++ {
		if i < len(hasAudio) && hasAudio[i] {
			anyHave = true
		} else if i >= len(durations) || durations[i] <= 0 {
			canPadSilence = false
		}
	}
	useAudio = anyHave && canPadSilence

	var preFilters []string
	var silenceDefs []string
	var parts []string
	for i := 0; i < n; i++ {
		vRef := fmt.Sprintf("[%d:v:0]", i)
		if needScale {
			vRef = fmt.Sprintf("[v_%d]", i)
			preFilters = append(preFilters, fmt.Sprintf("[%d:v:0]scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,setsar=1%s;", i, targetW, targetH, targetW, targetH, vRef))
		}
		parts = append(parts, vRef)
		if !useAudio {
			continue
		}
		if i < len(hasAudio) && hasAudio[i] {
			parts = append(parts, fmt.Sprintf("[%d:a:0]", i))
		} else {
			silenceDefs = append(silenceDefs, fmt.Sprintf("anullsrc=r=44100:cl=stereo:d=%.3f[silent_%d];", durations[i], i))
			parts = append(parts, fmt.Sprintf("[silent_%d]", i))
		}
	}

	if useAudio {
		return fmt.Sprintf("%s%s%sconcat=n=%d:v=1:a=1[outv][outa]", strings.Join(preFilters, ""), strings.Join(silenceDefs, ""), strings.Join(parts, ""), n), true
	}
	return fmt.Sprintf("%s%sconcat=n=%d:v=1:a=0[outv]", strings.Join(preFilters, ""), strings.Join(parts, ""), n), false
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
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "ffmpeg", "-encoders")
		if out, err := cmd.Output(); err == nil {
			if strings.Contains(string(out), "h264_rkmpp") {
				cachedEncoder = "h264_rkmpp"
				return
			}
		}
		cachedEncoder = "libx264"
	})
	return cachedEncoder
}

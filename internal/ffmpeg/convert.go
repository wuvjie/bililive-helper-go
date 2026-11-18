// convert.go 提供 FLV 到 MP4 的格式转换功能。
// 通过 TS 中间格式实现无损转换（全程 stream-copy，零重编码）。
package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// VideoBSFForCodec 根据视频编码格式返回封装 MPEG-TS 所需的比特流过滤器参数。
// H.264 使用 h264_mp4toannexb，HEVC/H.265 使用 hevc_mp4toannexb，其他编码（如 AV1）不附加过滤器。
func VideoBSFForCodec(codec string) []string {
	switch strings.ToLower(strings.TrimSpace(codec)) {
	case "hevc", "h265":
		return []string{"-bsf:v", "hevc_mp4toannexb"}
	case "h264", "avc1", "avc":
		return []string{"-bsf:v", "h264_mp4toannexb"}
	default:
		return nil
	}
}

// ConvertViaTS 将输入文件无损转封装为输出文件。
// 优先尝试单遍直接流复制（-c copy），兼容 H.264/HEVC/AV1 且避免 TS 中转的双倍读写；
// 若直接转封装失败，自动降级为 TS 中转管线并根据编码动态匹配比特流过滤器。
func ConvertViaTS(ctx context.Context, input, output string) error {
	// 第 1 级：优先尝试直接流复制转封装，零中间文件
	directErr := Run(ctx, Options{
		Args: []string{"-nostdin", "-i", input, "-c", "copy", "-y", "-loglevel", "error", output},
	})
	if directErr == nil {
		if info, err := os.Stat(output); err == nil && info.Size() > 0 {
			return nil
		}
	}
	os.Remove(output) // 清理直接转封装失败的残留输出

	// 第 2 级：降级通过 TS 中间容器转封装，并按实际视频编码匹配 BSF
	tsPath := output + ".tmp.ts"
	defer os.Remove(tsPath)

	codec := ProbeVideoCodec(ctx, input)
	tsArgs := []string{"-nostdin", "-i", input, "-c", "copy"}
	tsArgs = append(tsArgs, VideoBSFForCodec(codec)...)
	tsArgs = append(tsArgs, "-y", "-loglevel", "error", tsPath)

	if err := Run(ctx, Options{Args: tsArgs}); err != nil {
		return fmt.Errorf("转换 TS 失败: %w", err)
	}

	if err := Run(ctx, Options{
		Args: []string{"-nostdin", "-i", tsPath, "-c", "copy", "-y", "-loglevel", "error", output},
	}); err != nil {
		os.Remove(output) // 清理部分写入的输出文件
		return fmt.Errorf("TS→输出 失败: %w", err)
	}

	return nil
}

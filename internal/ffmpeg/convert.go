// convert.go 提供 FLV 到 MP4 的格式转换功能。
// 通过 TS 中间格式实现无损转换（全程 stream-copy，零重编码）。
package ffmpeg

import (
	"context"
	"fmt"
	"os"
)

// ConvertViaTS 通过 TS 中间容器将输入文件转换为输出文件。
// 管线：input → TS（stream-copy + h264_mp4toannexb 比特流过滤器）→ output（stream-copy）。
// TS 中间文件在转换完成后自动清理。
func ConvertViaTS(ctx context.Context, input, output string) error {
	tsPath := output + ".tmp.ts"
	defer os.Remove(tsPath)

	// 步骤 1：输入文件 → TS
	err := Run(ctx, Options{
		Args: []string{"-nostdin", "-i", input, "-c", "copy", "-bsf:v", "h264_mp4toannexb", "-y", "-loglevel", "error", tsPath},
	})
	if err != nil {
		return fmt.Errorf("转换 TS 失败: %w", err)
	}

	// 步骤 2：TS → 输出文件
	err = Run(ctx, Options{
		Args: []string{"-nostdin", "-i", tsPath, "-c", "copy", "-y", "-loglevel", "error", output},
	})
	if err != nil {
		return fmt.Errorf("TS→输出 失败: %w", err)
	}

	return nil
}

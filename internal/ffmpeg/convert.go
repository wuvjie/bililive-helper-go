package ffmpeg

import (
	"context"
	"fmt"
	"os"
)

// ConvertViaTS converts an input file to output via the TS intermediate format.
// Pipeline: input -> TS (copy) -> output (copy).
// The TS intermediate file is automatically cleaned up.
func ConvertViaTS(ctx context.Context, input, output string) error {
	tsPath := output + ".tmp.ts"
	defer os.Remove(tsPath)

	// Step 1: input -> TS
	err := Run(ctx, Options{
		Args: []string{"-nostdin", "-i", input, "-c", "copy", "-bsf:v", "h264_mp4toannexb", "-y", "-loglevel", "error", tsPath},
	})
	if err != nil {
		return fmt.Errorf("转换 TS 失败: %w", err)
	}

	// Step 2: TS -> output
	err = Run(ctx, Options{
		Args: []string{"-nostdin", "-i", tsPath, "-c", "copy", "-y", "-loglevel", "error", output},
	})
	if err != nil {
		return fmt.Errorf("TS→输出 失败: %w", err)
	}

	return nil
}

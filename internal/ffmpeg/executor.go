package ffmpeg

import "context"

// Executor 封装所有 ffmpeg/ffprobe 操作的接口。
// 用于依赖注入：service 层通过接口调用 ffmpeg，测试时可注入 mock 实现。
type Executor interface {
	// ConcatTS 将多个 TS 文件拼接为一个输出文件（stream-copy）。
	ConcatTS(ctx context.Context, tsFiles []string, output string, onProgress func(string)) error

	// ConvertViaTS 将输入文件通过 TS 中间格式转换为 MP4（无损 stream-copy）。
	ConvertViaTS(ctx context.Context, input, output string) error

	// Reencode 使用 concat filter 重新编码多个文件为一个输出文件（fallback 方案）。
	Reencode(ctx context.Context, files []string, folder, output string, onProgress func(string)) error

	// ValidateOutput 校验输出文件的完整性（容器格式、时长、多点解码测试）。
	ValidateOutput(ctx context.Context, path string) error

	// ProbeDuration 通过 ffprobe 获取视频时长（秒）。
	ProbeDuration(ctx context.Context, path string) (float64, error)

	// ProbeHealth 通过 ffprobe 检查视频文件是否有有效的编码流。
	ProbeHealth(ctx context.Context, path string) bool
}

// DefaultExecutor 是 Executor 的默认实现，直接调用包级函数。
type DefaultExecutor struct{}

func (d *DefaultExecutor) ConcatTS(ctx context.Context, tsFiles []string, output string, onProgress func(string)) error {
	return ConcatTS(ctx, tsFiles, output, onProgress)
}

func (d *DefaultExecutor) ConvertViaTS(ctx context.Context, input, output string) error {
	return ConvertViaTS(ctx, input, output)
}

func (d *DefaultExecutor) Reencode(ctx context.Context, files []string, folder, output string, onProgress func(string)) error {
	return Reencode(ctx, files, folder, output, onProgress)
}

func (d *DefaultExecutor) ValidateOutput(ctx context.Context, path string) error {
	return ValidateOutput(ctx, path)
}

func (d *DefaultExecutor) ProbeDuration(ctx context.Context, path string) (float64, error) {
	return ProbeDuration(ctx, path)
}

func (d *DefaultExecutor) ProbeHealth(ctx context.Context, path string) bool {
	return ProbeHealth(ctx, path)
}

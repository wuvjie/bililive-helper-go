package service

import (
	"context"
)

// mockExecutor 是 ffmpeg.Executor 的测试用 mock 实现。
// 所有方法默认成功返回，可通过设置 ErrorField 模拟失败。
type mockExecutor struct {
	ErrorField error // 非 nil 时所有方法返回此错误
}

func (m *mockExecutor) ConcatTS(_ context.Context, _ []string, _ string, _ func(string)) error {
	return m.ErrorField
}

func (m *mockExecutor) ConvertViaTS(_ context.Context, _, _ string) error {
	return m.ErrorField
}

func (m *mockExecutor) Reencode(_ context.Context, _ []string, _, _ string, _ func(string)) error {
	return m.ErrorField
}

func (m *mockExecutor) ValidateOutput(_ context.Context, _ string) error {
	return m.ErrorField
}

func (m *mockExecutor) ProbeDuration(_ context.Context, _ string) (float64, error) {
	if m.ErrorField != nil {
		return 0, m.ErrorField
	}
	return 300.0, nil // 5 分钟
}

func (m *mockExecutor) ProbeHealth(_ context.Context, _ string) bool {
	return m.ErrorField == nil
}

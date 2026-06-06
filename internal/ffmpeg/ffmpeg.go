// Package ffmpeg 封装 FFmpeg/FFprobe 命令行工具。
// 提供进程管理（超时、进程组）、TS 拼接、格式转换、输出校验和重编码等功能。
package ffmpeg

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"
)

const DefaultTimeout = 2 * time.Hour

// Options 配置 ffmpeg 命令的执行参数。
type Options struct {
	Args     []string
	Timeout  time.Duration    // 0 means DefaultTimeout
	OnStdout func(line string) // called for each stdout line (for progress parsing)
}

// Run 执行 ffmpeg 命令。
// 进程管理：创建进程组（Setpgid），上下文取消或超时时杀死整个进程组，
// 防止产生孤立的 ffmpeg 子进程。
func Run(ctx context.Context, opts Options) error {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	cmd := exec.Command("ffmpeg", opts.Args...)
	setProcessGroup(cmd)

	// StdoutPipe 必须在 Start 之前创建（os/exec 要求），
	// 但 goroutine 在 Start 之后启动，避免 Start 失败时 goroutine 泄漏。
	var stdoutPipe io.Reader
	if opts.OnStdout != nil {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("创建 stdout pipe 失败: %w", err)
		}
		stdoutPipe = stdout
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 ffmpeg 失败: %w", err)
	}

	// 进程已启动，安全启动 goroutine 读取 stdout
	if opts.OnStdout != nil && stdoutPipe != nil {
		go func() {
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				opts.OnStdout(scanner.Text())
			}
		}()
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		killProcessGroup(cmd)
		<-done
		return ctx.Err()
	case <-timer.C:
		killProcessGroup(cmd)
		<-done
		return fmt.Errorf("ffmpeg 超时（%v）", timeout)
	}
}

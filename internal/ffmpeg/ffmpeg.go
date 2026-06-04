// Package ffmpeg 封装 FFmpeg/FFprobe 命令行工具。
// 提供进程管理（超时、进程组）、TS 拼接、格式转换、输出校验和重编码等功能。
package ffmpeg

import (
	"bufio"
	"context"
	"fmt"
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

	if opts.OnStdout != nil {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("创建 stdout pipe 失败: %w", err)
		}
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				opts.OnStdout(scanner.Text())
			}
		}()
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 ffmpeg 失败: %w", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		killProcessGroup(cmd)
		<-done
		return ctx.Err()
	case <-time.After(timeout):
		killProcessGroup(cmd)
		<-done
		return fmt.Errorf("ffmpeg 超时（%v）", timeout)
	}
}
